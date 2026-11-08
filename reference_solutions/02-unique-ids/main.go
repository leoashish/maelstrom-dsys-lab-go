// Reference solution for Challenge 2: Unique IDs. See ../../challenges/02-unique-ids/.
//
// Combines n.ID() (unique per node, assigned at init) with a local,
// atomically-incremented counter (unique per call on this node). Needs
// zero coordination, so it stays correct under --nemesis partition
// --availability total. The counter uses sync/atomic rather than a bare
// `counter++` because maelstrom.Node dispatches every request on its own
// goroutine -- see the concurrency note in ../../challenges/02-unique-ids/main.go.
package main

import (
	"fmt"
	"log"
	"os"
	"sync/atomic"

	maelstrom "mod-unique-ids-ref/maelstrom"
)

func main() {
	n := maelstrom.NewNode()
	var counter atomic.Int64

	n.Handle("generate", func(msg maelstrom.Message) error {
		next := counter.Add(1)
		return n.Reply(msg, map[string]any{
			"type": "generate_ok",
			"id":   fmt.Sprintf("%s-%d", n.ID(), next),
		})
	})

	if err := n.Run(); err != nil {
		log.Printf("ERROR: %s", err)
		os.Exit(1)
	}
}
