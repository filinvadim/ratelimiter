package ratelimiter

import (
	"context"
	"testing"
	"time"
)

func TestAttributeMapGetSetDelete(t *testing.T) {
	key := "test"
	l := NewKeyLimiter(context.TODO())
	l.RegisterKey(key, 5, time.Second)
	if !l.HasKey(key) {
		t.Fatal("expected existing key:", key)
	}
	l.DeleteKeys("test")
	if l.HasKey(key) {
		t.Fatal("expected nonexisting key:", key)
	}
}

func TestKeyLimiterAccuracy(t *testing.T) {
	limit := 8
	window := 5*time.Second

	l := NewKeyLimiter(context.TODO())
	defer l.DeleteKeys()
	l.RegisterKey("test", uint32(limit), window)

	attempts := 10
	elapsedChan := make(chan time.Duration, 1)

	go func(elCh chan time.Duration) {
		started := time.Now()
		for i := 0; i < attempts; i++ {
			l.LimitKey("test", 1, func() {
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

func TestConcurrentKeyLimiterAccuracy(t *testing.T) {
	limit := 8
	window := 5*time.Second

	l := NewKeyLimiter(context.TODO())
	defer l.DeleteKeys()
	l.RegisterKey("test", uint32(limit), window)

	attempts := 10
	elapsedChan := make(chan time.Duration, attempts)

	started := time.Now()
	for i := 0; i < attempts; i++ {
		go func() {
			l.LimitKey("test", 1, func() {
				elapsedChan <- time.Now().Sub(started)
			})
		}()
	}

	for i := 0; i < attempts; i++ {
		select {
		case dur := <-elapsedChan:
			if dur.Round(time.Second) == window {
				return
			}
		}
	}
	t.Fatal("limiter is inaccurate")
}
