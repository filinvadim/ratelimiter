package ratelimiter

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type Limiter struct {
	limitTick   <-chan time.Time
	interval    time.Duration
	taskNum     uint32
	nextWindow  time.Time
	limiterLock *sync.Mutex
	limit       uint32

	innerLock *sync.RWMutex
	ctx       context.Context
	stopChan  chan struct{}
}

func NewLimiter(ctx context.Context, limit uint32, interval time.Duration) *Limiter {
	limiter := &Limiter{
		limitTick:   time.Tick(interval),
		limiterLock: new(sync.Mutex),
		innerLock:   new(sync.RWMutex),
		interval:    interval,
		limit:       limit,
		ctx:         ctx,
		stopChan:    make(chan struct{}),
	}

	go limiter.startWindowCounting()
	return limiter
}

func (l *Limiter) Limit(weight uint32, fn func()) {
	if weight == 0 {
		weight = 1
	}

	if atomic.LoadUint32(&l.taskNum) >= atomic.LoadUint32(&l.limit) {
		l.limiterLock.Lock()
		l.innerLock.RLock()
		time.Sleep(l.nextWindow.Sub(time.Now()))
		l.innerLock.RUnlock()
		atomic.StoreUint32(&l.taskNum, 0)
		l.limiterLock.Unlock()
	}
	fn()
	atomic.AddUint32(&l.taskNum, 1*weight)
}

func (l *Limiter) startWindowCounting() {
	l.nextWindow = time.Now().Add(l.interval)
	for now := range l.limitTick {
		select {
		case <-l.stopChan:
			return
		case <-l.ctx.Done():
			return
		default:
			l.innerLock.Lock()
			l.nextWindow = now.Add(l.interval)
			l.innerLock.Unlock()
		}
	}
}

func (l *Limiter) Close() {
	close(l.stopChan)
}
