package main

import (
	"encoding/json"
	"fmt"
)

// RunJSONAndAppDev covers encoding/json (the thing every single
// Maelstrom message in this lab flows through) and a few project-level
// conventions -- module layout, the testing idiom you'll see in
// exercises_test.go -- that matter once you're writing more than one
// file.
func RunJSONAndAppDev() {
	section("5. JSON, and a few app-development conventions")

	// --- Marshaling: Go value -> JSON bytes -----------------------------
	// Struct TAGS (the backtick strings) control the JSON field names
	// and behavior. This is exactly what maelstrom/node.go's
	// MessageBody struct does for "type", "msg_id", "in_reply_to", etc.
	type AddRequest struct {
		Type  string `json:"type"`
		Delta int    `json:"delta"`
		// omitempty: skip this field entirely in the output if it's the
		// zero value -- used throughout maelstrom's own MessageBody so
		// outgoing JSON doesn't carry a bunch of "msg_id":0 noise.
		Note string `json:"note,omitempty"`
	}
	req := AddRequest{Type: "add", Delta: 5}
	buf, err := json.Marshal(req)
	if err != nil {
		fmt.Println("marshal error:", err)
		return
	}
	fmt.Println("marshaled (Note omitted, it was empty):", string(buf))

	// --- Unmarshaling: JSON bytes -> Go value ---------------------------
	// Two common targets:
	//   1. A known struct, when you know the exact shape.
	//   2. map[string]any, when the shape varies by message type -- this
	//      is what the echo and unique-ids challenge stubs lean on.
	raw := []byte(`{"type": "add_ok", "in_reply_to": 4}`)

	var asMap map[string]any
	if err := json.Unmarshal(raw, &asMap); err != nil {
		fmt.Println("unmarshal error:", err)
		return
	}
	fmt.Println("unmarshaled into map[string]any:", asMap)
	// Gotcha: EVERY JSON number decodes into a Go `float64` when the
	// target is `any` -- there's no way for the decoder to know you
	// meant an int. asMap["in_reply_to"] is a float64 here, not an int.
	fmt.Printf("  in_reply_to's actual Go type: %T\n", asMap["in_reply_to"])

	type AddOk struct {
		Type      string `json:"type"`
		InReplyTo int    `json:"in_reply_to"`
	}
	var asStruct AddOk
	if err := json.Unmarshal(raw, &asStruct); err != nil {
		fmt.Println("unmarshal error:", err)
		return
	}
	fmt.Println("unmarshaled into a struct (in_reply_to is a real int):", asStruct.InReplyTo)

	// --- json.RawMessage: defer parsing the body ------------------------
	// This is precisely why maelstrom.Message.Body is typed as
	// json.RawMessage rather than already-parsed: the outer Message
	// (src/dest/body) has a fixed shape, but the BODY's shape depends on
	// the message's "type", which you don't know until you've looked at
	// it. json.RawMessage holds the raw bytes, letting each handler
	// parse its own body into whatever struct makes sense for that
	// message type -- exactly what every challenge's main.go does.
	type Envelope struct {
		Src  string          `json:"src"`
		Body json.RawMessage `json:"body"`
	}
	envelope := []byte(`{"src": "n1", "body": {"type": "read", "msg_id": 9}}`)
	var env Envelope
	_ = json.Unmarshal(envelope, &env)
	fmt.Println("envelope.Body is still raw bytes here:", string(env.Body))

	// --- A note on project layout ----------------------------------------
	// Every challenge folder in this lab is its own Go MODULE (its own
	// go.mod), with a vendored `maelstrom/` subdirectory as a local
	// PACKAGE. That's the standard Go shape: one module per deployable
	// thing, package boundaries marking "these files share private
	// state," `internal/` (not used here, but common in bigger projects)
	// marking "don't import this from outside this module."

	// --- A note on testing -----------------------------------------------
	// Go's built-in `testing` package wants no third-party framework for
	// the basics: a function named TestXxx(t *testing.T), t.Errorf on
	// failure, `go test ./...` to run everything. The idiomatic style for
	// "many small cases" is a TABLE-DRIVEN test: a slice of input/expected
	// pairs, looped with t.Run(name, ...) so each case shows up as its
	// own named result. See exercises_test.go in this folder for a real
	// one -- then `go test ./...` to see it (mostly) fail, and
	// exercises.go for the TODOs that make it pass.
	fmt.Println("see exercises.go + exercises_test.go in this folder for a hands-on, test-first warm-up")
}
