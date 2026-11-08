// Reference solution for Challenge 4: Grow-Only Counter. See ../../challenges/04-grow-counter/.
//
// A G-Counter CRDT: each node owns one entry in `counts`, the effective
// value is the sum of all entries, and merging two counter maps keeps the
// max per entry. Gossips to every other node (no topology needed here)
// once a second. All access to `counts` goes through `mu`, since add/
// read/replicate handlers and the gossip goroutine all run concurrently.
package main

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"

	maelstrom "mod-grow-counter-ref/maelstrom"
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
	counts := map[string]int64{}

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
