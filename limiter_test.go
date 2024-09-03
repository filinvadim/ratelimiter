package ratelimiter

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestConcurrentLimiterAccuracy(t *testing.T) {
	limit := 8
	window := 5 * time.Second

	l := NewLimiter(uint32(limit), window)
	defer l.Close()

	attempts := 10
	elapsedChan := make(chan time.Duration, attempts)

	started := time.Now()
	for i := 0; i < attempts; i++ {
		go func() {
			l.Limit(1, func() {
				elapsedChan <- time.Now().Sub(started)
			})
		}()
	}

	for dur := range elapsedChan {
		if dur.Round(time.Second) == window {
			return
		}
	}
	t.Fatal("limiter is inaccurate")
}

func TestLimiter_WithinLimit(t *testing.T) {
	limit := uint32(10)
	interval := 100 * time.Millisecond
	limiter := NewLimiter(limit, interval)

	var wg sync.WaitGroup
	var completed uint32

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			limiter.Limit(1, func() {
				atomic.AddUint32(&completed, 1)
			})
		}()
	}

	wg.Wait()

	if completed != 10 {
		t.Fatalf("Expected 10 completed requests, but got %d", completed)
	}
}

func TestLimiter_ExceedLimit(t *testing.T) {
	limit := uint32(10)
	interval := 100 * time.Millisecond
	limiter := NewLimiter(limit, interval)

	var wg sync.WaitGroup
	var completed uint32
	start := time.Now()

	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			limiter.Limit(1, func() {
				atomic.AddUint32(&completed, 1)
			})
		}()
	}

	wg.Wait()

	if !limiter.IsLocked() {
		t.Fatalf("Expected limiter to be locked")
	}

	duration := time.Since(start)
	if duration < interval {
		t.Fatalf("Expected at least %v wait time, but got %v", interval, duration)
	}

	if completed != 12 {
		t.Fatalf("Expected 12 completed requests, but got %d", completed)
	}
}

func TestLimiter_WeightedRequests(t *testing.T) {
	limit := uint32(10)
	interval := 100 * time.Millisecond
	limiter := NewLimiter(limit, interval)

	var wg sync.WaitGroup
	var completed uint32

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			limiter.Limit(2, func() {
				atomic.AddUint32(&completed, 1)
			})
		}()
	}

	wg.Wait()

	if completed != 5 {
		t.Fatalf("Expected 5 completed weighted requests, but got %d", completed)
	}
}

func TestLimiter_ExceedWeightedLimit(t *testing.T) {
	limit := uint32(10)
	interval := 100 * time.Millisecond
	limiter := NewLimiter(limit, interval)

	var wg sync.WaitGroup
	var completed uint32
	start := time.Now()

	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			limiter.Limit(2, func() {
				atomic.AddUint32(&completed, 1)
			})
		}()
	}

	wg.Wait()
	if !limiter.IsLocked() {
		t.Fatalf("Expected limiter to be locked")
	}
	duration := time.Since(start)
	if duration < interval {
		t.Fatalf("Expected at least %v wait time, but got %v", interval, duration)
	}

	if completed != 6 {
		t.Fatalf("Expected 6 completed weighted requests, but got %d", completed)
	}
}

func TestLimiter_ParallelExecution(t *testing.T) {
	limit := uint32(10)
	interval := 100 * time.Millisecond
	limiter := NewLimiter(limit, interval)

	var wg sync.WaitGroup
	var completed uint32

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			limiter.Limit(1, func() {
				time.Sleep(10 * time.Millisecond) // simulate work
				atomic.AddUint32(&completed, 1)
			})
		}()
	}

	wg.Wait()

	if completed != 10 {
		t.Fatalf("Expected 10 parallel requests to complete, but got %d", completed)
	}
}

func TestLimiter_IntervalAccuracy(t *testing.T) {
	limit := uint32(10)
	interval := 1 * time.Second
	limiter := NewLimiter(limit, interval)

	start := time.Now()
	limiter.Limit(10, func() {})
	time.Sleep(100 * time.Millisecond) // let the limiter process the task

	if limiter.totalWeight.Load() != 10 {
		t.Fatalf("Expected total weight to be 10 immediately after request")
	}

	time.Sleep(interval)

	limiter.Limit(1, func() {})

	if limiter.totalWeight.Load() > 1 {
		t.Fatalf("Expected total weight to reset after interval, but got %d", limiter.totalWeight.Load())
	}

	if time.Since(start) < interval {
		t.Fatalf("Expected at least interval duration to pass, got %s", time.Since(start))
	}
}

func TestLimiter_Concurrency(t *testing.T) {
	limit := uint32(1000)
	interval := 50 * time.Millisecond
	limiter := NewLimiter(limit, interval)

	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			limiter.Limit(1, func() {})
		}()
	}

	wg.Wait()

	if limiter.totalWeight.Load() != 1000 {
		t.Fatalf("Expected total weight of 1000, got %d", limiter.totalWeight.Load())
	}

	// Wait for the interval to pass and check if the limiter resets correctly
	time.Sleep(interval + 10*time.Millisecond)

	limiter.Limit(1, func() {})

	if limiter.totalWeight.Load() > 1 {
		t.Fatalf("Expected total weight to reset to 1 after interval, but got %d", limiter.totalWeight.Load())
	}
}
