package main

import "fmt"

// RunCollectionsAndGenerics covers slices, maps, and generics -- the
// data-shuffling you'll do in every challenge (a set of seen messages,
// a map of per-node counters, a slice of pending IDs).
func RunCollectionsAndGenerics() {
	section("4. Slices, maps, and generics")

	// --- Arrays vs slices -------------------------------------------
	// An array has a fixed size baked into its type ([3]int != [4]int).
	// You will almost never use one directly -- slices (a view over an
	// underlying array, with a length and a capacity) are what Go code
	// actually uses for "a list of things."
	var arr [3]int = [3]int{1, 2, 3}
	fmt.Println("array (fixed size):", arr)

	nums := []int{1, 2, 3} // slice literal
	nums = append(nums, 4, 5)
	fmt.Println("slice after append:", nums, "len:", len(nums), "cap:", cap(nums))

	// THE gotcha: re-slicing shares the same backing array. Mutating
	// through one view can silently mutate another. This has bitten
	// every Go programmer at least once.
	original := []int{1, 2, 3, 4, 5}
	view := original[1:3] // shares backing array with `original`
	view[0] = 99
	fmt.Println("mutating a re-slice changes the original too:", original)
	// `copy` or `append([]T{}, src...)` when you actually need an
	// independent copy instead of a view.
	independent := make([]int, len(original))
	copy(independent, original)
	independent[0] = -1
	fmt.Println("independent copy, original untouched:", original[0], "vs copy:", independent[0])

	// --- Maps -----------------------------------------------------------
	// A nil map can be READ (returns the zero value) but panics on
	// WRITE -- always `make` or use a literal before inserting.
	counts := map[string]int{"n1": 3, "n2": 5}
	counts["n3"] = 1

	// The "comma-ok" idiom: distinguishes "key present with zero value"
	// from "key absent" -- you can't tell those apart from the plain
	// index form alone.
	if v, ok := counts["n2"]; ok {
		fmt.Println("n2 present, value:", v)
	}
	if _, ok := counts["n99"]; !ok {
		fmt.Println("n99 absent")
	}
	delete(counts, "n1")

	// Map iteration order is intentionally RANDOMIZED by the runtime --
	// never rely on it. (This is also why the broadcast and grow-counter
	// challenges' reference solutions build a fresh slice/snapshot under
	// a lock rather than assuming any particular order matters.)
	total := 0
	for _, v := range counts {
		total += v
	}
	fmt.Println("sum of remaining counts:", total)

	// --- Generics ---------------------------------------------------
	// Added in Go 1.18. A type parameter in square brackets lets one
	// function body work across types, with the compiler still checking
	// everything at compile time -- no reflection, no runtime casts.
	ints := []int{1, 2, 3, 4, 5}
	doubled := Map(ints, func(n int) int { return n * 2 })
	fmt.Println("generic Map (double):", doubled)

	strs := Map(ints, func(n int) string { return fmt.Sprintf("#%d", n) })
	fmt.Println("generic Map (int -> string):", strs)

	evens := Filter(ints, func(n int) bool { return n%2 == 0 })
	fmt.Println("generic Filter (evens):", evens)

	sum := Reduce(ints, 0, func(acc, n int) int { return acc + n })
	fmt.Println("generic Reduce (sum):", sum)

	// A generic TYPE, not just a generic function: a stack that works
	// for any element type.
	var st Stack[string]
	st.Push("a")
	st.Push("b")
	st.Push("c")
	for {
		v, ok := st.Pop()
		if !ok {
			break
		}
		fmt.Println("popped:", v)
	}
}

// Map applies f to every element of s, returning a new slice. Note the
// two type parameters: T is what's in `s`, U is whatever f returns --
// they don't have to be the same type (see the int -> string call
// above).
func Map[T, U any](s []T, f func(T) U) []U {
	out := make([]U, len(s))
	for i, v := range s {
		out[i] = f(v)
	}
	return out
}

// Filter keeps only elements where pred returns true.
func Filter[T any](s []T, pred func(T) bool) []T {
	var out []T
	for _, v := range s {
		if pred(v) {
			out = append(out, v)
		}
	}
	return out
}

// Reduce folds s down to a single value, starting from `initial`.
func Reduce[T, A any](s []T, initial A, f func(acc A, v T) A) A {
	acc := initial
	for _, v := range s {
		acc = f(acc, v)
	}
	return acc
}

// Stack[T] is a generic TYPE: the type parameter is declared on the
// type itself, so every method automatically works for whatever T the
// caller picked (`Stack[string]` above, but `Stack[int]`, `Stack[Message]`,
// etc. all "just work" with zero extra code).
type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(v T) {
	s.items = append(s.items, v)
}

func (s *Stack[T]) Pop() (T, bool) {
	var zero T
	if len(s.items) == 0 {
		return zero, false
	}
	last := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return last, true
}
