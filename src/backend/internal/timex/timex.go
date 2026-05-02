package timex

import (
	"sync"
	"time"
)

type TimeoutManager[K comparable] struct {
	mu   sync.Mutex
	jobs map[K]*time.Timer
}

func NewTimeoutManager[K comparable]() *TimeoutManager[K] {
	return &TimeoutManager[K]{
		jobs: make(map[K]*time.Timer),
	}
}

func (tm *TimeoutManager[K]) Add(key K, deadline time.Time, fn func()) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	oldJob, ok := tm.jobs[key]
	if ok {
		_ = oldJob.Stop()
	}

	delay := time.Until(deadline)

	tm.jobs[key] = time.AfterFunc(delay, func() {
		fn()

		tm.mu.Lock()
		defer tm.mu.Unlock()

		delete(tm.jobs, key)
	})
}

func (tm *TimeoutManager[K]) Cancel(key K) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	oldJob, ok := tm.jobs[key]
	if ok {
		_ = oldJob.Stop()
	}

	delete(tm.jobs, key)
}
