package main

// A small, test-first warm-up: the tests in exercises_test.go are
// already written (and currently failing). Fill in the TODOs below to
// make them pass. Every technique needed here was already demonstrated,
// fully worked, somewhere in sections 1-7 -- this is about applying it
// yourself, not learning something brand new.
//
//	go test ./...        # run them
//	go test -race ./...  # the real check for SafeCounter

// TODO: return the sum of every value in nums. An empty or nil slice
// sums to 0. (See section 1's variadicSum for the shape.)
func SumInts(nums []int) int {
	return 0 // wrong on purpose -- fix me
}

// TODO: return items with duplicates removed, keeping the first
// occurrence of each value and the original order. `comparable` is a
// builtin constraint meaning "usable as a map key" -- which is exactly
// what you need to track "have I seen this before?" (See section 4's
// Filter for the generic-function shape, and the broadcast challenge's
// `messages` set for where this exact idea shows up for real.)
func Dedup[T comparable](items []T) []T {
	return items // wrong on purpose (does nothing) -- fix me
}

// SafeCounter must be safe to call Inc() from many goroutines at once --
// Value() should always reflect every completed Inc() call, and the
// zero value (SafeCounter{}) must be ready to use with no extra setup
// (so don't add a constructor function; add a field that's usable at
// its own zero value, which sync.Mutex's zero value already is).
//
// TODO: add a sync.Mutex field, and lock it around the read/write in
// both methods below. (See section 6's safeCounter example for the
// exact pattern -- this is that same idea, promoted to a reusable type.)
type SafeCounter struct {
	n int
}

func (c *SafeCounter) Inc() {
	c.n++ // TODO: not safe for concurrent use yet -- fix it
}

func (c *SafeCounter) Value() int {
	return c.n // TODO: reading without the same lock Inc uses is also unsafe once Inc takes one
}

// TODO: compute the sum of nums, split across `workers` goroutines
// (roughly even-sized chunks of the slice), combining their partial
// sums into the one final result. Must return the exact same answer as
// a plain sum for any workers >= 1 -- this is purely a parallelism
// exercise, not a different algorithm. (See section 7's worker pool and
// fan-out/fan-in examples for the shape: split work, collect partial
// results over a channel, combine.)
func ParallelSum(nums []int, workers int) int {
	return 0 // wrong on purpose -- fix me
}
