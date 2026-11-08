package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// RunConcurrencyBasics covers goroutines, channels, select, the sync
// package, and context -- the building blocks every challenge's main.go
// leans on the moment more than one message can be in flight at once
// (which, as you've seen, is always: maelstrom.Node dispatches every
// inbound message on its own goroutine).
func RunConcurrencyBasics() {
	section("6. Concurrency: goroutines, channels, select, sync, context")

	// --- Goroutines -------------------------------------------------
	// `go f()` starts f running concurrently and returns immediately --
	// it does NOT wait for f to finish. Without something to wait on,
	// main() can (and here, deliberately, almost certainly will) exit
	// before the goroutine below ever gets to print.
	go fmt.Println("  (this goroutine may or may not print before the next line -- that's the point)")
	time.Sleep(10 * time.Millisecond) // just to let it have a chance; NOT how you'd coordinate for real

	// --- Channels: how goroutines actually talk to each other -----------
	// An unbuffered channel send blocks until a receiver is ready (and
	// vice versa) -- it's a synchronization point, not just a queue.
	unbuffered := make(chan string)
	go func() {
		unbuffered <- "hello from a goroutine"
	}()
	fmt.Println("received:", <-unbuffered)

	// A buffered channel holds up to N values before a send blocks --
	// useful as a bounded queue, but it's still not magic: fill it up
	// with no reader and the next send blocks just the same.
	buffered := make(chan int, 3)
	buffered <- 1
	buffered <- 2
	buffered <- 3
	fmt.Println("buffered channel holds", len(buffered), "of cap", cap(buffered))

	// Closing a channel signals "no more values." Ranging over a channel
	// reads until it's closed. Only the SENDER should close a channel
	// (closing one you're receiving from, or closing one twice, panics).
	results := make(chan int)
	go func() {
		defer close(results)
		for i := 0; i < 3; i++ {
			results <- i * i
		}
	}()
	for v := range results {
		fmt.Println("  squared:", v)
	}

	// --- select: waiting on multiple channels at once --------------------
	// `select` picks whichever case is ready first. With a `default`,
	// it becomes non-blocking -- "check without waiting."
	ch := make(chan int, 1)
	select {
	case v := <-ch:
		fmt.Println("got", v)
	default:
		fmt.Println("non-blocking select: nothing ready, moved on immediately")
	}

	// The classic timeout pattern: race a real channel against
	// time.After. This shape -- or its context.WithTimeout equivalent
	// below -- is exactly what the broadcast challenge's gossip loop
	// uses to avoid hanging forever on an unreachable neighbor.
	slow := make(chan string)
	go func() {
		time.Sleep(50 * time.Millisecond)
		slow <- "finally done"
	}()
	select {
	case v := <-slow:
		fmt.Println("slow operation finished:", v)
	case <-time.After(10 * time.Millisecond):
		fmt.Println("select+time.After: gave up waiting after 10ms")
	}

	// --- sync.WaitGroup: wait for N goroutines to finish -------------------
	var wg sync.WaitGroup
	resultsCh := make(chan int, 5)
	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			resultsCh <- n * n
		}(n(i)) // see the comment on n() below for why this matters
	}
	wg.Wait() // blocks until all 5 have called Done()
	close(resultsCh)
	sum := 0
	for v := range resultsCh {
		sum += v
	}
	fmt.Println("sum of squares 1..5 via WaitGroup + goroutines:", sum)

	// --- The data race, and sync.Mutex ------------------------------------
	// THIS is the exact bug class behind every "add a sync.Mutex here"
	// comment in this lab's challenges. raceyIncrement (below) is NOT
	// safe for concurrent use: `counter++` is read-modify-write, and two
	// goroutines can interleave those three steps. Run this file with
	// `go run -race .` and the race detector will catch it directly.
	var unsafeCounter int
	var mu sync.Mutex
	var safeCounter int
	var wg2 sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg2.Add(1)
		go func() {
			defer wg2.Done()
			unsafeCounter++ // racy: no synchronization at all

			mu.Lock()
			safeCounter++ // safe: only one goroutine inside the lock at a time
			mu.Unlock()
		}()
	}
	wg2.Wait()
	fmt.Println("unsafeCounter ended at", unsafeCounter, "(often, but NOT reliably, 100)")
	fmt.Println("safeCounter ended at", safeCounter, "(always exactly 100)")

	// sync/atomic: for a single counter, an atomic type is simpler than
	// a Mutex -- Add() is guaranteed indivisible by the hardware itself.
	var atomicCounter atomic.Int64
	var wg3 sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg3.Add(1)
		go func() {
			defer wg3.Done()
			atomicCounter.Add(1)
		}()
	}
	wg3.Wait()
	fmt.Println("atomicCounter ended at", atomicCounter.Load(), "(always exactly 100)")

	// sync.Once: run something exactly once, no matter how many
	// goroutines call it concurrently -- common for lazy one-time setup.
	var once sync.Once
	setup := func() { fmt.Println("  one-time setup ran") }
	for i := 0; i < 3; i++ {
		once.Do(setup) // prints only on the first call
	}

	// --- context: cancellation and deadlines that PROPAGATE ----------------
	// context.Context carries a cancellation signal (and optionally a
	// deadline) through a call chain. maelstrom.Node.SyncRPC takes one
	// for exactly this reason: "give up waiting for this reply after N
	// seconds" without the caller having to hand-roll its own timer.
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel() // always call cancel, even if the context already expired -- releases its resources

	select {
	case <-time.After(100 * time.Millisecond):
		fmt.Println("this won't print -- the context will expire first")
	case <-ctx.Done():
		fmt.Println("context.Done() fired:", ctx.Err())
	}
}

// n is a trivial identity helper, here only to make a point: in Go
// versions before 1.22, a loop variable like `i` in `for i := 1; i <= 5;
// i++ { go func(){ ...i... }() }` was SHARED across every iteration, so
// goroutines could see a later value of `i` than the one they "meant."
// The fix was always to pass it as a parameter, as the call site above
// does. Go 1.22 (what this lab uses) changed loop semantics so each
// iteration gets its own copy, making that workaround usually
// unnecessary -- but you'll see it constantly in any pre-1.22 Go code,
// so it's worth recognizing even though you may not need to write it.
func n(i int) int { return i }
