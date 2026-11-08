// Challenge 3: Broadcast
//
// Goal: build a gossip protocol so that any message broadcast to ANY
// node eventually becomes visible on EVERY node -- even through network
// partitions -- without flooding the network forever.
//
// Solution -- a mutex-protected set of seen messages, plus one background
// goroutine per topology-provided neighbor. Each goroutine wakes every
// 200 ms, computes what that neighbor hasn't yet acknowledged, and if
// anything's pending sends it all as ONE "broadcast_many" RPC (stage D
// batching). If the RPC times out (e.g. mid-partition), nothing is marked
// acknowledged, so the next tick retries the same backlog automatically
// (stage C fault tolerance). All shared state goes through a single
// sync.Mutex, since maelstrom.Node dispatches every inbound message --
// and every gossip goroutine runs -- concurrently.
package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"

	maelstrom "mod-broadcast/maelstrom"
)

type broadcastBody struct {
	Message float64 `json:"message"`
}

type topologyBody struct {
	Topology map[string][]string `json:"topology"`
}

type broadcastManyBody struct {
	Type     string    `json:"type"`
	Messages []float64 `json:"messages"`
}

func main() {
	n := maelstrom.NewNode()

	var mu sync.Mutex
	messages := map[float64]struct{}{} // the set of seen messages
	acked := map[string]map[float64]struct{}{} // neighbor -> messages it has confirmed

	addMessage := func(v float64) {
		mu.Lock()
		defer mu.Unlock()
		messages[v] = struct{}{}
	}

	allMessages := func() []float64 {
		mu.Lock()
		defer mu.Unlock()
		out := make([]float64, 0, len(messages))
		for v := range messages {
			out = append(out, v)
		}
		return out
	}

	pendingFor := func(neighbor string) []float64 {
		mu.Lock()
		defer mu.Unlock()
		var out []float64
		for v := range messages {
			if _, done := acked[neighbor][v]; !done {
				out = append(out, v)
			}
		}
		return out
	}

	markAcked := func(neighbor string, vs []float64) {
		mu.Lock()
		defer mu.Unlock()
		if acked[neighbor] == nil {
			acked[neighbor] = map[float64]struct{}{}
		}
		for _, v := range vs {
			acked[neighbor][v] = struct{}{}
		}
	}

	n.Handle("broadcast", func(msg maelstrom.Message) error {
		var body broadcastBody
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		addMessage(body.Message)
		return n.Reply(msg, map[string]any{"type": "broadcast_ok"})
	})

	n.Handle("broadcast_many", func(msg maelstrom.Message) error {
		var body broadcastManyBody
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		for _, v := range body.Messages {
			addMessage(v)
		}
		return n.Reply(msg, map[string]any{"type": "broadcast_many_ok"})
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		return n.Reply(msg, map[string]any{
			"type":     "read_ok",
			"messages": allMessages(),
		})
	})

	n.Handle("topology", func(msg maelstrom.Message) error {
		var body topologyBody
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		// Go 1.22+ gives each loop iteration its own `neighbor` variable,
		// so it's safe to close over it directly in the goroutine below.
		for _, neighbor := range body.Topology[n.ID()] {
			go gossipLoop(n, neighbor, pendingFor, markAcked)
		}
		return n.Reply(msg, map[string]any{"type": "topology_ok"})
	})

	if err := n.Run(); err != nil {
		log.Printf("ERROR: %s", err)
		os.Exit(1)
	}
}

// gossipLoop runs for the lifetime of the process, one per neighbor.
func gossipLoop(
	n *maelstrom.Node,
	neighbor string,
	pendingFor func(string) []float64,
	markAcked func(string, []float64),
) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		pending := pendingFor(neighbor)
		if len(pending) == 0 {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		_, err := n.SyncRPC(ctx, neighbor, broadcastManyBody{
			Type:     "broadcast_many",
			Messages: pending,
		})
		cancel()

		if err == nil {
			markAcked(neighbor, pending)
		}
		// On error (e.g. the neighbor is unreachable mid-partition), do
		// nothing -- `pending` stays un-acked, so the next tick retries
		// the same backlog (plus anything new) automatically.
	}
}
