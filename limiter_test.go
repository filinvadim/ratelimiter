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
		t.Errorf("Expected 10 completed requests, but got %d", completed)
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

	duration := time.Since(start)
	if duration < interval {
		t.Errorf("Expected at least %v wait time, but got %v", interval, duration)
	}

	if completed != 12 {
		t.Errorf("Expected 12 completed requests, but got %d", completed)
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
		t.Errorf("Expected 5 completed weighted requests, but got %d", completed)
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

	duration := time.Since(start)
	if duration < interval {
		t.Errorf("Expected at least %v wait time, but got %v", interval, duration)
	}

	if completed != 6 {
		t.Errorf("Expected 6 completed weighted requests, but got %d", completed)
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
		t.Errorf("Expected 10 parallel requests to complete, but got %d", completed)
	}
}
