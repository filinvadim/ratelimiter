package ratelimiter

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// Limiter Weighted Sliding Window Log.
type Limiter struct {
	interval    time.Duration
	limit       uint32
	storedLimit uint32
	requests    []time.Time
	mx          *sync.Mutex
	ctx         context.Context
	isClosed    *atomic.Bool
}

func NewLimiter(ctx context.Context, limit uint32, interval time.Duration) *Limiter {
	return &Limiter{
		interval:    interval,
		limit:       limit,
		storedLimit: limit,
		requests:    make([]time.Time, 0, limit),
		mx:          new(sync.Mutex),
		ctx:         ctx,
		isClosed:    new(atomic.Bool),
	}
}

func (l *Limiter) Limit(weight uint32, fn func()) {
	l.mx.Lock()
	defer l.mx.Unlock()

	if weight == 0 {
		weight = 1
	}
	now := time.Now()

	cutoff := now.Add(-l.interval)
	i := 0
	for i < len(l.requests) && l.requests[i].Before(cutoff) {
		i++
	}
	l.requests = l.requests[i:]

	if weight <= l.limit {
		l.requests = append(l.requests, now)
		l.limit -= weight
	} else {
		timeToWait := l.requests[0].Add(l.interval).Sub(now)
		l.limit = l.storedLimit

		if l.ctx.Err() != nil {
			return
		}

		time.Sleep(timeToWait) // rate limited here
	}
	if l.ctx.Err() != nil {
		return
	}
	if l.isClosed.Load() {
		return
	}

	fn()
}

func (l *Limiter) IsLocked() bool {
	l.mx.Lock()
	defer l.mx.Unlock()

	return l.limit == 0
}

func (l *Limiter) Close() {
	l.isClosed.Store(true)
}
