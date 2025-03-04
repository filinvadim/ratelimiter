package ratelimiter

import (
	"sync"
	"sync/atomic"
	"time"
)

type Task struct {
	T      time.Time `json:"t"`
	Weight uint32    `json:"w"`
}

type TaskQueueStorer interface {
	Append(tasks ...Task) []Task
	Len() int
	TaskByIndex(i int) Task
	CutOffBefore(i int)
}

// Limiter Weighted Sliding Window Log.
type Limiter struct {
	interval time.Duration

	limitMx            *sync.Mutex
	limit, totalWeight *atomic.Uint32

	tasks TaskQueueStorer
}

// NewLimiter may use any queue provider: in-memory, Redis/Memcached, persistent storage
func NewLimiter(limit uint32, interval time.Duration, queue TaskQueueStorer) *Limiter {
	atomicLimit := new(atomic.Uint32)
	atomicLimit.Store(limit)

	if queue == nil {
		queue = newDefaultTaskQueue(limit)
	}

	return &Limiter{
		interval:    interval,
		limit:       atomicLimit,
		totalWeight: new(atomic.Uint32),
		limitMx:     new(sync.Mutex),
		tasks:       queue,
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

	tasksLen := l.tasks.Len()

	for i < tasksLen && l.tasks.TaskByIndex(i).T.Before(cutoff) {
		l.totalWeight.Add(-l.tasks.TaskByIndex(i).Weight) // remove weight of expired tasks
		i++
	}

	l.tasks.CutOffBefore(i)

	isLimited := l.totalWeight.Load()+weight > l.limit.Load()
	firstTaskTime := now.Add(l.interval) // time that doesn't require to wait if there are no tasks
	if l.tasks.Len() > 0 {
		firstTaskTime = l.tasks.TaskByIndex(0).T
	}
	if !isLimited {
		l.totalWeight.Add(weight)
		l.tasks.Append(Task{now, weight})
		return
	}

	timeToWait := firstTaskTime.Add(l.interval).Sub(now)

	l.limitMx.Lock()
	time.Sleep(timeToWait) // rate limited here
	l.limitMx.Unlock()

	l.totalWeight.Add(weight)
	l.tasks.Append(Task{now, weight})
}

func (l *Limiter) IsLocked() bool {
	return l.totalWeight.Load() >= l.limit.Load()
}

func (l *Limiter) Close() {
	l.tasks = nil
}
