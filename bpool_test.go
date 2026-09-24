package bpool

import (
	"runtime"
	"testing"
	"time"
)

// waitGoroutines waits for the goroutine count to drop to at most n.
func waitGoroutines(n int, timeout time.Duration) int {
	deadline := time.Now().Add(timeout)
	for {
		got := runtime.NumGoroutine()
		if got <= n || time.Now().After(deadline) {
			return got
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// TestDoneStopsDrain checks Done stops the drain goroutine promptly rather
// than on its next 30 second tick, so pools do not leak goroutines.
func TestDoneStopsDrain(t *testing.T) {
	before := runtime.NumGoroutine()
	for i := 0; i < 100; i++ {
		NewBufferPool(1<<20, nil).Done()
	}
	if got := waitGoroutines(before, time.Second); got > before {
		t.Fatalf("goroutines: %d before, %d a second after 100 pools were done", before, got)
	}
}

func TestDoneTwice(t *testing.T) {
	pool := NewBufferPool(1<<20, nil)
	pool.Done()
	pool.Done() // must not panic
}
