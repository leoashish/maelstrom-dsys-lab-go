// Challenge 1: Echo
//
// Goal: respond to an "echo" request with an "echo_ok" response that
// carries the same payload back.
//
// Maelstrom will send your node a message whose body looks like:
//
//	{"type": "echo", "msg_id": 1, "echo": "Please echo 35"}
//
// Your handler must reply with a body like:
//
//	{"type": "echo_ok", "echo": "Please echo 35"}
//
// You don't need to set "in_reply_to", "src" or "dest" yourself --
// n.Reply(msg, body) fills those in for you, using the incoming
// message's msg_id and src.
//
// This challenge exists to get you comfortable with the shape of every
// challenge in this lab: a maelstrom.Message comes in, you build a
// response body, and you send it back with n.Reply. The maelstrom
// package (vendored in ./maelstrom, straight from the official
// Maelstrom repo) handles reading STDIN, writing STDOUT, and the init
// handshake for you.
//
// Run it with: ./test.sh   (see README.md in this folder)
package main

import (
	"encoding/json"
	"log"
	"os"

	maelstrom "mod-echo/maelstrom"
)

func main() {
	n := maelstrom.NewNode()

	n.Handle("echo", func(msg maelstrom.Message) error {
		// Unmarshal the raw JSON body into a map, flip "type" to
		// "echo_ok", and hand the whole thing back via n.Reply --
		// it fills in in_reply_to / src / dest for us.
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		body["type"] = "echo_ok"
		return n.Reply(msg, body)
	})

	if err := n.Run(); err != nil {
		log.Printf("ERROR: %s", err)
		os.Exit(1)
	}
}
