package main

import (
	"sync"
	"testing"
)

// These tests are already written and already failing -- make them pass
// by filling in the TODOs in exercises.go. Run them with:
//
//	go test ./...
//
// Table-driven subtests (t.Run) are the idiomatic Go shape for "many
// small cases against one function" -- see TestSumInts below for the
// pattern, then look for it again throughout any real Go codebase.

func TestSumInts(t *testing.T) {
	cases := []struct {
		name string
		in   []int
		want int
	}{
		{"empty", nil, 0},
		{"single value", []int{5}, 5},
		{"several values", []int{1, 2, 3, 4, 5}, 15},
		{"includes negatives", []int{-1, -2, 3}, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := SumInts(tc.in); got != tc.want {
				t.Errorf("SumInts(%v) = %d, want %d", tc.in, got, tc.want)
			}
		})
	}
}

func TestDedup(t *testing.T) {
	t.Run("ints, preserves first-seen order", func(t *testing.T) {
		got := Dedup([]int{1, 2, 2, 3, 1, 4})
		want := []int{1, 2, 3, 4}
		if len(got) != len(want) {
			t.Fatalf("Dedup(%v) = %v, want %v", []int{1, 2, 2, 3, 1, 4}, got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("Dedup()[%d] = %d, want %d (full: got=%v want=%v)", i, got[i], want[i], got, want)
			}
		}
	})

	t.Run("strings", func(t *testing.T) {
		got := Dedup([]string{"a", "b", "a", "c"})
		want := []string{"a", "b", "c"}
		if len(got) != len(want) {
			t.Fatalf("Dedup(%v) = %v, want %v", []string{"a", "b", "a", "c"}, got, want)
		}
	})
}

func TestSafeCounter_SingleGoroutine(t *testing.T) {
	var c SafeCounter
	for i := 0; i < 10; i++ {
		c.Inc()
	}
	if got := c.Value(); got != 10 {
		t.Errorf("Value() = %d, want 10", got)
	}
}

func TestSafeCounter_Concurrent(t *testing.T) {
	var c SafeCounter
	var wg sync.WaitGroup
	const n = 1000
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Inc()
		}()
	}
	wg.Wait()
	if got := c.Value(); got != n {
		t.Errorf("Value() = %d, want %d -- Inc()/Value() aren't safe for concurrent use yet", got, n)
	}
	// This test can occasionally pass by luck even on a racy
	// implementation -- a data race is a possible-interleaving bug, not
	// a guaranteed-every-time one. The authoritative check is:
	//   go test -race ./...
	// which reports a definite DATA RACE on an unsynchronized
	// implementation regardless of whether this run's count came out
	// right.
}

func TestParallelSum(t *testing.T) {
	nums := make([]int, 100)
	want := 0
	for i := range nums {
		nums[i] = i + 1
		want += nums[i]
	}
	for _, workers := range []int{1, 2, 3, 7, 16} {
		if got := ParallelSum(nums, workers); got != want {
			t.Errorf("ParallelSum(nums, %d workers) = %d, want %d", workers, got, want)
		}
	}
}
