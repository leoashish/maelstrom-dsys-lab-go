package main

import "fmt"

// RunBasics covers: variables & constants, basic types, type inference,
// control flow, and functions. If Go is fresh in your head, skim it. If
// it's been a while, read every comment.
func RunBasics() {
	section("1. Basics: variables, types, control flow, functions")

	// --- Variables & type inference ---------------------------------
	// `var` with an explicit type, or `:=` to let Go infer it from the
	// right-hand side. `:=` only works for NEW variables inside a
	// function body -- package-level variables always need `var`.
	var name string = "maelstrom"
	count := 4 // inferred as int
	var ratio float64 = 3.5
	fmt.Println("var + explicit type:", name)
	fmt.Println(":= + inference:    ", count, ratio)

	// Multiple declarations at once, and zero values: a `var` with no
	// initializer gets the type's zero value (0, "", false, nil, ...) --
	// Go has no concept of "uninitialized" or null-by-default surprises.
	var a, b int
	var s string
	fmt.Println("zero values:", a, b, fmt.Sprintf("%q", s))

	// Constants: compile-time, untyped unless you say otherwise. `iota`
	// auto-increments inside a const block -- the idiomatic way to build
	// enums (you used these implicitly already: maelstrom's Error codes
	// in ../challenges/*/maelstrom/rpc_error.go use plain numbers instead,
	// but iota is the more common Go idiom for that shape of problem).
	const (
		StatusPending = iota // 0
		StatusRunning        // 1
		StatusDone           // 2
	)
	fmt.Println("iota enum:", StatusPending, StatusRunning, StatusDone)

	// --- Control flow -------------------------------------------------
	// `if` needs no parens; it can carry an init statement scoped to the
	// if/else, which is idiomatic for "compute then check" (you'll see
	// this exact shape constantly: `if v, ok := m[k]; ok { ... }`).
	if v := count * 2; v > 5 {
		fmt.Println("if-with-init:", v, "> 5")
	}

	// Go has exactly one looping keyword: `for`. No `while`, no `do`.
	for i := 0; i < 3; i++ { // C-style
		_ = i
	}
	i := 0
	for i < 3 { // "while" is just `for <condition>`
		i++
	}
	for range 3 { // Go 1.22+: range over a plain int, 3 times, no index needed
		_ = 0
	}
	fmt.Println("three for-loop styles, same three iterations each")

	// `switch` doesn't fall through by default (unlike C/Java) and
	// doesn't need a condition at all -- a bare `switch` with no
	// expression is a readable if/else-if chain.
	grade := 82
	switch {
	case grade >= 90:
		fmt.Println("grade: A")
	case grade >= 80:
		fmt.Println("grade: B")
	default:
		fmt.Println("grade: C or below")
	}

	// --- Functions ------------------------------------------------------
	quotient, remainder := multiReturn(17, 5)
	fmt.Println("divmod(17, 5) ->", quotient, remainder)
	fmt.Println("namedReturn(4) ->", namedReturn(4))
	fmt.Println("sum(1,2,3,4) ->", variadicSum(1, 2, 3, 4))

	adder := makeAdder(10)
	fmt.Println("closure makeAdder(10)(5) ->", adder(5))

	fmt.Println("deferDemo() output below (defer runs LIFO, after the return value is set):")
	deferDemo()
}

// Go functions can return multiple values -- used everywhere, most
// famously the (value, error) idiom you'll use in every challenge.
func multiReturn(a, b int) (int, int) {
	return a / b, a % b
}

// Named return values are pre-declared variables you can just `return`
// with no arguments -- handy for short functions, easy to abuse into
// unreadability in long ones. Use sparingly.
func namedReturn(n int) (squared int) {
	squared = n * n
	return // returns `squared` implicitly
}

// `...int` makes the last parameter variadic -- callers can pass any
// number of ints (or spread a slice with `sum(nums...)`).
func variadicSum(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// Functions are values. Returning a function that captures a variable
// from its enclosing scope is a closure -- `n` here lives on past
// makeAdder's return, kept alive by the returned function.
func makeAdder(n int) func(int) int {
	return func(x int) int { return x + n }
}

// `defer` schedules a call to run when the surrounding function
// returns, LIFO (last deferred, first run) -- used constantly for
// cleanup: `defer mu.Unlock()`, `defer f.Close()`, `defer cancel()`.
// `recover` (only meaningful inside a deferred function) stops a
// `panic` from unwinding further, letting the program continue.
func deferDemo() {
	defer fmt.Println("  deferred: this prints LAST")
	defer fmt.Println("  deferred: this prints SECOND")
	fmt.Println("  deferDemo body: this prints FIRST")

	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("  recovered from panic:", r)
			}
		}()
		panic("something went wrong")
	}()
}
