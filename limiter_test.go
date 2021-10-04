package ratelimiter

import (
	"context"
	"testing"
	"time"
)

func TestLimiterAccuracy(t *testing.T) {
	limit := 8
	window := 5 * time.Second
	l := NewLimiter(context.TODO(), uint32(limit), window)
	defer l.Close()

	attempts := 10
	elapsedChan := make(chan time.Duration, 1)

	go func(elCh chan time.Duration) {
		started := time.Now()
		for i := 0; i < attempts; i++ {
			l.Limit(1, func() {
				t.Logf("f() №%d %s", i, time.Now().Sub(started).String())
				elCh <- time.Now().Sub(started)
			})
		}
	}(elapsedChan)

	for i := 0; i < attempts; i++ {
		select {
		case dur := <-elapsedChan:
			if i == limit {
				if dur.Round(time.Second) == window {
					return
				}
			}
		}
	}
	t.Fatal("limiter is inaccurate")

}

func TestConcurrentLimiterAccuracy(t *testing.T) {
	limit := 8
	window := 5 * time.Second

	l := NewLimiter(context.TODO(), uint32(limit), window)
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
