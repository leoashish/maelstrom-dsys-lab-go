package main

import (
	"errors"
	"fmt"
)

// RunErrorHandling covers Go's error-as-a-value idiom: no exceptions,
// no try/catch -- just a return value you're expected to check, every
// time. This is the single most pervasive pattern in idiomatic Go, and
// exactly what every maelstrom.HandlerFunc in this lab's challenges
// returns.
func RunErrorHandling() {
	section("3. Error handling: errors as values")

	// `error` is just an interface with one method: Error() string. The
	// idiomatic shape is "return the zero value and a non-nil error on
	// failure, the real value and nil on success" -- and the caller
	// checks `if err != nil` immediately, before touching the value.
	if v, err := parsePositive("42"); err != nil {
		fmt.Println("unexpected error:", err)
	} else {
		fmt.Println("parsePositive(\"42\") ->", v)
	}
	if _, err := parsePositive("-5"); err != nil {
		fmt.Println("parsePositive(\"-5\") ->", err)
	}

	// errors.New / fmt.Errorf for simple, one-off errors.
	err := fmt.Errorf("could not reach node %s", "n3")
	fmt.Println("fmt.Errorf:", err)

	// --- Wrapping: preserving the chain of "why" ------------------------
	// Real systems have layers: a low-level error happens, a mid-level
	// function adds context, and the caller at the top still wants to
	// know what the ROOT cause was. `%w` (instead of %v/%s) wraps an
	// error so later code can unwrap it again.
	base := ErrKeyNotFound
	wrapped := fmt.Errorf("reading config: %w", base)
	doubleWrapped := fmt.Errorf("starting node: %w", wrapped)
	fmt.Println("wrapped chain prints as:", doubleWrapped)

	// errors.Is walks the WHOLE chain looking for a match -- this is how
	// you check "was this ultimately caused by X?" without caring how
	// many layers of context got added on top.
	fmt.Println("errors.Is(doubleWrapped, ErrKeyNotFound)?", errors.Is(doubleWrapped, ErrKeyNotFound))

	// --- Custom error types + errors.As ---------------------------------
	// Sometimes you need more than a message -- structured data about
	// what went wrong (compare: maelstrom's own *RPCError in
	// ../challenges/*/maelstrom/rpc_error.go, which carries a Code and
	// Text, not just a string).
	opErr := doSomethingRisky(-7)
	fmt.Println("doSomethingRisky(-7) ->", opErr)

	var ve *ValidationError
	if errors.As(opErr, &ve) {
		fmt.Printf("  it's specifically a ValidationError: field=%q, value=%d\n", ve.Field, ve.Value)
	}

	// --- Sentinel errors -------------------------------------------------
	// A package-level `var Err... = errors.New(...)` that callers compare
	// against with errors.Is. ErrKeyNotFound above is one; the standard
	// library's own io.EOF is the most famous example.
}

// ErrKeyNotFound is a sentinel error: a specific, comparable value
// callers can check for with errors.Is, regardless of how much context
// gets wrapped around it later.
var ErrKeyNotFound = errors.New("key not found")

func parsePositive(s string) (int, error) {
	var n int
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
		return 0, fmt.Errorf("parsing %q: %w", s, err)
	}
	if n <= 0 {
		return 0, fmt.Errorf("value %d is not positive", n)
	}
	return n, nil
}

// ValidationError is a custom error TYPE, not just a string -- it
// carries structured fields a caller can inspect after an errors.As
// check, rather than having to parse a message.
type ValidationError struct {
	Field string
	Value int
}

// Implementing Error() string is all it takes to satisfy the `error`
// interface -- same implicit-satisfaction rule as any other interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("invalid value %d for field %q", e.Value, e.Field)
}

func doSomethingRisky(delta int) error {
	if delta < 0 {
		return &ValidationError{Field: "delta", Value: delta}
	}
	return nil
}
