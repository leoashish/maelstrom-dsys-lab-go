// Challenge 2: Unique ID Generation
//
// Goal: respond to a "generate" request with a "generate_ok" response
// containing a globally unique "id". "Globally unique" means unique
// across every node in the cluster, for the entire run -- not just
// unique within this one process.
//
//	Request:  {"type": "generate", "msg_id": 7}
//	Response: {"type": "generate_ok", "in_reply_to": 7, "id": <anything>}
//
// "id" may be any JSON value: a string, a number, whatever you like.
// Maelstrom collects every id returned across every node and fails the
// test if any two are the same.
//
// What makes this a *distributed systems* problem and not just "call a
// UUID library": the test script runs this with --nemesis partition and
// --availability total, meaning every node must keep successfully
// generating valid ids even while completely cut off from every other
// node. So whatever scheme you use CANNOT require talking to another
// node or to a central counter at request time -- it has to work with
// purely local state.
//
// Solution: keep one local, atomically-incremented counter per node and
// prefix it with the node's ID. Any node can only produce IDs starting
// with its own prefix, so two nodes can never collide, and no request
// ever needs to reach out to the network -- fully partition-tolerant.
// The counter uses sync/atomic because maelstrom.Node dispatches every
// message on its own goroutine (a bare `counter++` could be read by two
// goroutines before either writes back -- the exact bug this challenge
// exists to surface).
package main

import (
	"fmt"
	"log"
	"os"
	"sync/atomic"

	maelstrom "mod-unique-ids/maelstrom"
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
