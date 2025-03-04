package ratelimiter

import (
	"sync"
	"time"
)

type KeyLimiter struct {
	mx       *sync.Mutex
	limiters map[string]*Limiter
	storage  TaskQueueStorer
}

func NewKeyLimiter(storage TaskQueueStorer) *KeyLimiter {
	return &KeyLimiter{
		mx:       new(sync.Mutex),
		limiters: make(map[string]*Limiter),
		storage:  storage,
	}
}

func (kl *KeyLimiter) HasKey(key string) bool {
	kl.mx.Lock()
	_, ok := kl.limiters[key]
	kl.mx.Unlock()
	return ok
}

func (kl *KeyLimiter) RegisterKey(key string, limit uint32, interval time.Duration) {
	kl.mx.Lock()
	kl.limiters[key] = NewLimiter(limit, interval, kl.storage)
	kl.mx.Unlock()
}

func (kl *KeyLimiter) LimitKey(key string, weight uint32, fn func()) {
	kl.mx.Lock()
	limiter, ok := kl.limiters[key]
	kl.mx.Unlock()

	if ok && fn != nil {
		limiter.Limit(weight, fn)
		return
	}
	if fn != nil {
		fn()
	}
}

func (kl *KeyLimiter) IsLocked(key string) bool {
	kl.mx.Lock()
	limiter, ok := kl.limiters[key]
	kl.mx.Unlock()
	if !ok {
		return false
	}

	return limiter.IsLocked()
}

func (kl *KeyLimiter) DeleteKeys(keys ...string) {
	kl.mx.Lock()
	defer kl.mx.Unlock()

	if keys == nil {
		for k, _ := range kl.limiters {
			delete(kl.limiters, k)
		}
		return
	}
	for _, key := range keys {
		if _, ok := kl.limiters[key]; ok {
			delete(kl.limiters, key)
		}
	}
}

func (kl *KeyLimiter) Close() {
	kl.mx.Lock()
	defer kl.mx.Unlock()

	for _, limiter := range kl.limiters {
		limiter.Close()
	}
	kl.limiters = nil
}
