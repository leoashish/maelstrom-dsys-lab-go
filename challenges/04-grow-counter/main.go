// Challenge 4: Grow-Only Counter (a CRDT)
//
// Goal: build an eventually-consistent counter. Clients send "add"
// requests with a non-negative delta; "read" requests should return the
// sum of EVERY delta added anywhere in the cluster -- including ones
// added on a different node than the one being read, and even after a
// partition heals.
//
//	add:  {"type": "add", "msg_id": 4, "delta": 3}  -> {"type": "add_ok"}
//	read: {"type": "read", "msg_id": 5}              -> {"type": "read_ok", "value": 11}
//
// Solution -- a "G-Counter" CRDT (Conflict-free Replicated Data
// Type): instead of one shared number every node has to agree on, give
// every node its OWN private counter, and make the effective value the
// SUM across all nodes' counters.
//
//   - "add" only ever touches THIS node's own entry in the map. No
//     coordination needed, so it's always available, even mid-partition.
//   - To make reads converge, periodically send your whole counter map to
//     every other node and MERGE what you receive: for each node_id in
//     the incoming map, keep whichever value is LARGER.
//
// Why merge-by-max matters: each entry only ever increases, so the
// larger of two values for the same node_id is always the more
// up-to-date one -- no matter how many times two nodes gossip, in what
// order, or how many messages get duplicated or delayed, merging by max
// always converges to the same final answer.
//
// CONCURRENCY: the counter map is read and written from many goroutines
// at once (every add, read, replicate, plus the gossip loop). Guard it
// with a sync.Mutex.
package main

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"

	maelstrom "mod-grow-counter/maelstrom"
)

type addBody struct {
	Delta int64 `json:"delta"`
}

type replicateBody struct {
	Type  string           `json:"type"`
	Value map[string]int64 `json:"value"`
}

func main() {
	n := maelstrom.NewNode()

	var mu sync.Mutex
	counts := map[string]int64{} // node_id -> node_id's current count (our best knowledge)

	total := func() int64 {
		mu.Lock()
		defer mu.Unlock()
		var sum int64
		for _, v := range counts {
			sum += v
		}
		return sum
	}

	snapshot := func() map[string]int64 {
		mu.Lock()
		defer mu.Unlock()
		out := make(map[string]int64, len(counts))
		for k, v := range counts {
			out[k] = v
		}
		return out
	}

	merge := func(other map[string]int64) {
		mu.Lock()
		defer mu.Unlock()
		for k, v := range other {
			if v > counts[k] {
				counts[k] = v
			}
		}
	}

	n.Handle("add", func(msg maelstrom.Message) error {
		var body addBody
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		mu.Lock()
		counts[n.ID()] += body.Delta
		mu.Unlock()
		return n.Reply(msg, map[string]any{"type": "add_ok"})
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		return n.Reply(msg, map[string]any{"type": "read_ok", "value": total()})
	})

	n.Handle("replicate", func(msg maelstrom.Message) error {
		var body replicateBody
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		merge(body.Value)
		return nil // fire-and-forget gossip; no reply needed
	})

	// Start the ticker only after the init handshake has populated n.NodeIDs()
	// (registering an "init" handler makes Go 1.22+ delegate init to it after
	// the ID is set).
	n.Handle("init", func(msg maelstrom.Message) error {
		go func() {
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()
			for range ticker.C {
				cur := snapshot()
				for _, other := range n.NodeIDs() {
					if other == n.ID() {
						continue
					}
					_ = n.Send(other, replicateBody{Type: "replicate", Value: cur})
				}
			}
		}()
		return nil
	})

	if err := n.Run(); err != nil {
		log.Printf("ERROR: %s", err)
		os.Exit(1)
	}
}
