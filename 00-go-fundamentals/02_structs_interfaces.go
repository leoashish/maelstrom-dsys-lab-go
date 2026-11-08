package main

import "fmt"

// RunStructsAndInterfaces covers: structs, methods, value vs. pointer
// receivers, embedding, and interfaces -- including the implicit,
// structural satisfaction that makes Go's interfaces feel different
// from Java's or C#'s.
func RunStructsAndInterfaces() {
	section("2. Structs, methods, and interfaces")

	// --- Structs & methods -------------------------------------------
	n := NewNodeStats("n1")
	n.RecordMessage()
	n.RecordMessage()
	fmt.Println("value receiver View():", n.View())

	n.ResetViaPointer()
	fmt.Println("after ResetViaPointer():", n.View())

	// Value vs pointer receiver, concretely: a value receiver method
	// gets its OWN COPY of the struct, so it cannot mutate the
	// original. A pointer receiver method gets the real thing. Mixing
	// the two on the same type is legal but confusing -- pick pointer
	// receivers for anything that mutates, or anything larger than a
	// couple of fields (copying is not free), and stay consistent.

	// --- Embedding: composition, not inheritance ----------------------
	// Go has no class hierarchies. Embedding a type gives you its
	// fields and methods "promoted" onto the outer type -- composition
	// that LOOKS a little like inheritance but is just field access
	// sugar underneath.
	rn := ReplicaNode{
		NodeStats: NewNodeStats("n2"),
		Region:    "ap-south-1",
	}
	rn.RecordMessage() // promoted from the embedded NodeStats
	fmt.Println("embedded method call:", rn.View(), "region:", rn.Region)

	// --- Interfaces: implicit satisfaction -----------------------------
	// There's no `implements` keyword. Any type with the right method
	// set satisfies an interface automatically. This is why the
	// maelstrom.HandlerFunc you register with n.Handle(...) is just a
	// plain func type, not something that has to "implement" an
	// interface explicitly -- and why you can pass *your own* types
	// anywhere an interface is expected, as long as the methods line up.
	var r Reporter = n // *NodeStats satisfies Reporter -- nothing declared that link
	fmt.Println("via interface:", r.Report())
	r = rn // *ReplicaNode satisfies it too, via the promoted method
	fmt.Println("via interface:", r.Report())

	// `any` (an alias for `interface{}`, added as a builtin alias in Go
	// 1.18) is the interface with ZERO required methods -- anything
	// satisfies it. You've already seen this constantly: maelstrom
	// messages are unmarshaled into `map[string]any` precisely because
	// their shape varies by message type.
	values := []any{1, "two", 3.0, true}
	for _, v := range values {
		describe(v)
	}
}

// NodeStats is a small struct -- a named group of fields. Idiomatic Go
// has no getters/setters ceremony for simple cases; exported fields
// (capitalized) are just public.
type NodeStats struct {
	ID       string
	messages int // unexported (lowercase) -- only visible within this package
}

func NewNodeStats(id string) *NodeStats {
	return &NodeStats{ID: id}
}

// A VALUE receiver: `s` inside this method is a copy of the NodeStats
// the caller has. Fine for read-only methods.
func (s NodeStats) View() string {
	return fmt.Sprintf("%s has handled %d messages", s.ID, s.messages)
}

// A POINTER receiver: `s` is the original. Required for any method that
// needs to mutate the struct, like this one incrementing a counter.
func (s *NodeStats) RecordMessage() {
	s.messages++
}

func (s *NodeStats) ResetViaPointer() {
	s.messages = 0
}

// ReplicaNode embeds NodeStats (note: no field name, just the type --
// that's what makes it "embedding" rather than a normal named field).
// ReplicaNode.RecordMessage() and .View() now work without being
// redefined, "promoted" from the embedded value.
type ReplicaNode struct {
	*NodeStats
	Region string
}

// Reporter is an interface: a method set, nothing more. Any type --
// yours, mine, from any package -- that happens to have a Report()
// string method satisfies it, with no explicit declaration anywhere.
type Reporter interface {
	Report() string
}

func (s *NodeStats) Report() string { return "[report] " + s.View() }

// describe demonstrates a type switch: given an `any`, ask "what are
// you, concretely?" and branch per-case. This is the idiomatic
// replacement for the kind of instanceof-chain you might reach for in
// other languages.
func describe(v any) {
	switch x := v.(type) {
	case int:
		fmt.Printf("  %v is an int (doubled: %d)\n", x, x*2)
	case string:
		fmt.Printf("  %q is a string (length: %d)\n", x, len(x))
	case float64:
		fmt.Printf("  %v is a float64\n", x)
	default:
		fmt.Printf("  %v is a %T\n", x, x)
	}
}
