package ratelimiter

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Limiter struct {
	limitTick chan struct{}
	interval  time.Duration
	taskNum   uint32
	lock      *sync.RWMutex
	limit     uint32
	ctx       context.Context

	stopChan chan struct{}
}

func NewLimiter(ctx context.Context, limit uint32, interval time.Duration) *Limiter {
	limiter := &Limiter{
		limitTick: make(chan struct{}, 1),
		lock:      new(sync.RWMutex),
		interval: interval,
		limit:     limit,
		ctx:       ctx,
		stopChan:  make(chan struct{}),
	}

	go limiter.startLimiting()
	return limiter
}

func (l *Limiter) Limit(fn func()) {
	l.limitTick <- struct{}{}

	l.lock.RLock()
	fn()
	atomic.AddUint32(&l.taskNum, 1)
	l.lock.RUnlock()
}

func (l *Limiter) startLimiting() {
	go func() {
		tick := time.Tick(l.interval)
		for range tick {
			if _, ok := <-l.stopChan; !ok {
				return
			}
			l.limitTick <- struct{}{}
		}
	}()
	for range l.limitTick {
		select {
		case <-l.stopChan:
			return
		case <-l.ctx.Done():
			return
		default:
			fmt.Println(l.taskNum, l.limit)
			if atomic.LoadUint32(&l.taskNum) >= atomic.LoadUint32(&l.limit) {
				l.lock.Lock()
				time.Sleep(l.interval)
				atomic.StoreUint32(&l.taskNum, 0)
				l.lock.Unlock()
			}
		}
	}
}

func (l *Limiter) Close() {
	close(l.stopChan)
}
