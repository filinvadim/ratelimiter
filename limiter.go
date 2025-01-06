package ratelimiter

import (
	"sync"
	"sync/atomic"
	"time"
)

type task struct {
	tm     time.Time
	weight uint32
}

// Limiter Weighted Sliding Window Log.
type Limiter struct {
	interval time.Duration
	tasks    []task

	limit, totalWeight *atomic.Uint32

	limitMx, tasksMx *sync.Mutex
}

func NewLimiter(limit uint32, interval time.Duration) *Limiter {
	atomicLimit := new(atomic.Uint32)
	atomicLimit.Store(limit)

	return &Limiter{
		interval:    interval,
		limit:       atomicLimit,
		totalWeight: new(atomic.Uint32),
		tasks:       make([]task, 0, limit),
		tasksMx:     new(sync.Mutex),
		limitMx:     new(sync.Mutex),
	}
}

func (l *Limiter) Limit(weight uint32, fn func()) {
	defer fn()

	if weight == 0 {
		weight = 1
	}

	var (
		now    = time.Now()
		cutoff = now.Add(-l.interval)
		i      = 0
	)

	l.tasksMx.Lock()
	for i < len(l.tasks) && l.tasks[i].tm.Before(cutoff) {
		l.totalWeight.Add(-l.tasks[i].weight) // remove weight of expired tasks
		i++
	}
	l.tasks = l.tasks[i:]

	isLimited := l.totalWeight.Load()+weight > l.limit.Load()
	firstTaskTime := now.Add(l.interval) // time that doesn't require to wait if there are no tasks
	if len(l.tasks) > 0 {
		firstTaskTime = l.tasks[0].tm
	}
	if !isLimited {
		l.totalWeight.Add(weight)
		l.tasks = append(l.tasks, task{now, weight})
		l.tasksMx.Unlock()
		return
	}
	l.tasksMx.Unlock()

	timeToWait := firstTaskTime.Add(l.interval).Sub(now)

	l.limitMx.Lock()
	time.Sleep(timeToWait) // rate limited here
	l.limitMx.Unlock()

	l.tasksMx.Lock()
	l.totalWeight.Add(weight)
	l.tasks = append(l.tasks, task{now, weight})
	l.tasksMx.Unlock()
}

func (l *Limiter) IsLocked() bool {
	return l.totalWeight.Load() >= l.limit.Load()
}

func (l *Limiter) Close() {
	// TODO
}
