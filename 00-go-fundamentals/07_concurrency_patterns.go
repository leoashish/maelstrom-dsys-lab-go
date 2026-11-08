package main

import (
	"fmt"
	"sync"
	"time"
)

// RunConcurrencyPatterns covers four named shapes built from the
// primitives in section 6: worker pools, fan-out/fan-in, pipelines, and
// rate limiting. These aren't academic -- the broadcast challenge's
// "one goroutine per neighbor, ticking on an interval, sending whatever
// is pending" IS a worker-pool-flavored pattern, and its batching logic
// is a rate-limiting idea in disguise.
func RunConcurrencyPatterns() {
	section("7. Concurrency patterns: worker pools, fan-out/fan-in, pipelines, rate limiting")

	// --- Worker pool: a fixed number of goroutines pulling from a shared queue ---
	// Useful when work items are cheap to describe but you want to cap
	// how much happens AT ONCE (bounded parallelism) -- e.g. "process
	// these 10,000 things, but never more than 8 at a time."
	jobs := make(chan int, 20)
	resultsCh := make(chan int, 20)
	const numWorkers = 4

	var wg sync.WaitGroup
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go worker(w, jobs, resultsCh, &wg)
	}
	for j := 1; j <= 20; j++ {
		jobs <- j
	}
	close(jobs) // tells every worker "no more jobs" once drained -- range on jobs exits
	wg.Wait()
	close(resultsCh)

	sum := 0
	for r := range resultsCh {
		sum += r
	}
	fmt.Println("worker pool: sum of squares 1..20, computed by 4 workers:", sum)

	// --- Fan-out / fan-in: split work across goroutines, merge the results ---
	// Similar idea, different framing: fan work OUT to several
	// goroutines each producing their own output channel, then fan those
	// IN to one channel for a single consumer.
	input := []int{1, 2, 3, 4, 5, 6, 7, 8}
	mid := len(input) / 2
	c1 := squareAsync(input[:mid])
	c2 := squareAsync(input[mid:])
	merged := fanIn(c1, c2)
	total := 0
	for v := range merged {
		total += v
	}
	fmt.Println("fan-out/fan-in: sum of squares via two producer goroutines:", total)

	// --- Pipeline: chained stages, each its own goroutine, connected by channels ---
	// Each stage reads from an input channel, does one job, writes to an
	// output channel, and closes it when done -- composable, and each
	// stage can run concurrently with its neighbors on a steady stream.
	numbersCh := generate(1, 2, 3, 4, 5)
	squaredCh := pipelineSquare(numbersCh)
	formattedCh := pipelineFormat(squaredCh)
	var formatted []string
	for s := range formattedCh {
		formatted = append(formatted, s)
	}
	fmt.Println("pipeline (generate -> square -> format):", formatted)

	// --- Rate limiting with time.Ticker -----------------------------------
	// A ticker fires on a fixed interval -- the exact mechanism behind
	// every "gossip every N milliseconds" background loop in this lab's
	// broadcast and grow-counter reference solutions. Here: process a
	// batch of requests, but no faster than one every 20ms.
	requests := []string{"r1", "r2", "r3", "r4"}
	limiter := time.NewTicker(20 * time.Millisecond)
	defer limiter.Stop()
	start := time.Now()
	for _, r := range requests {
		<-limiter.C // blocks until the next tick
		_ = r       // "process" it
	}
	fmt.Println("rate-limited 4 requests at 1/20ms, took at least:", time.Since(start).Round(time.Millisecond))

	// --- The "done channel" cancellation pattern --------------------------
	// Before context.Context existed (and still useful when you don't
	// want to pull in the extra import), a plain `done := make(chan
	// struct{})` closed once is a clean "stop now" signal every
	// goroutine selects on alongside its real work.
	done := make(chan struct{})
	tickCount := 0
	go func() {
		time.Sleep(55 * time.Millisecond)
		close(done) // broadcasts to every goroutine select-ing on `done` at once
	}()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
loop:
	for {
		select {
		case <-ticker.C:
			tickCount++
		case <-done:
			break loop
		}
	}
	fmt.Println("done-channel pattern: stopped after", tickCount, "ticks")
}

// worker pulls jobs from `jobs` until it's closed and empty, computes
// one result per job, and signals completion via the WaitGroup. This is
// the shape to reach for whenever you want N goroutines sharing one
// queue instead of one goroutine per item (unbounded goroutines are
// cheap in Go, but not free, and an unbounded number of them hitting an
// external resource at once usually isn't what you want).
func worker(id int, jobs <-chan int, results chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	for j := range jobs {
		results <- j * j
	}
}

func squareAsync(nums []int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			out <- n * n
		}
	}()
	return out
}

// fanIn merges any number of input channels into one, closing the
// output once every input has closed.
func fanIn(channels ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup
	wg.Add(len(channels))
	for _, c := range channels {
		go func(c <-chan int) {
			defer wg.Done()
			for v := range c {
				out <- v
			}
		}(c)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func generate(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

func pipelineSquare(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			out <- n * n
		}
	}()
	return out
}

func pipelineFormat(in <-chan int) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for n := range in {
			out <- fmt.Sprintf("<%d>", n)
		}
	}()
	return out
}
