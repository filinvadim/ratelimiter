package ratelimiter

import (
	"sync"
)

type defaultTaskQueue struct {
	tasksMx *sync.Mutex
	tasks   []Task
}

func newDefaultTaskQueue(limit uint32) *defaultTaskQueue {
	return &defaultTaskQueue{
		tasksMx: new(sync.Mutex),
		tasks:   make([]Task, 0, limit),
	}
}

func (q *defaultTaskQueue) Len() int {
	if q == nil || q.tasks == nil {
		return 0
	}
	q.tasksMx.Lock()
	defer q.tasksMx.Unlock()

	length := len(q.tasks)
	return length
}

func (q *defaultTaskQueue) TaskByIndex(i int) Task {
	if q == nil || q.tasks == nil {
		panic("task queue is nil")
	}
	if i < 0 {
		i = 0
	}
	if i >= len(q.tasks) {
		i = len(q.tasks) - 1
	}

	q.tasksMx.Lock()
	defer q.tasksMx.Unlock()

	task := q.tasks[i]
	return task
}

func (q *defaultTaskQueue) CutOffBefore(start int) {
	if q == nil || q.tasks == nil {
		return
	}
	if start < 0 {
		return
	}
	if start >= len(q.tasks) {
		return
	}
	q.tasksMx.Lock()
	defer q.tasksMx.Unlock()

	q.tasks = q.tasks[start:]
}

func (q *defaultTaskQueue) Append(tasks ...Task) []Task {
	if q == nil || q.tasks == nil {
		return nil
	}

	q.tasksMx.Lock()
	defer q.tasksMx.Unlock()

	q.tasks = append(q.tasks, tasks...)
	return q.tasks
}
