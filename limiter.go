package ratelimiter

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type Limiter struct {
	interval    time.Duration
	taskNum     uint32
	limiterLock *sync.Mutex
	limit       uint32

	innerLock *sync.RWMutex
	ctx       context.Context
	stopChan  chan struct{}
}

func NewLimiter(ctx context.Context, limit uint32, interval time.Duration) *Limiter {
	limiter := &Limiter{
		limiterLock: new(sync.Mutex),
		innerLock:   new(sync.RWMutex),
		interval:    interval,
		limit:       limit,
		ctx:         ctx,
		stopChan:    make(chan struct{}),
	}

	return limiter
}

func (l *Limiter) Limit(weight uint32, fn func()) {
	if weight == 0 {
		weight = 1
	}
	atomic.AddUint32(&l.taskNum, 1*weight)
	if atomic.LoadUint32(&l.taskNum) >= atomic.LoadUint32(&l.limit) {
		l.limiterLock.Lock()
		l.innerLock.RLock()
		time.Sleep(l.interval)
		l.innerLock.RUnlock()
		l.limiterLock.Unlock()
	}
	fn()
}

func (l *Limiter) IsLocked() bool {
	return atomic.LoadUint32(&l.taskNum) >= atomic.LoadUint32(&l.limit)
}

func (l *Limiter) Close() {
	close(l.stopChan)
}
