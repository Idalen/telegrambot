package register

import (
	"context"
	"sync"
)

type Job struct {
	cancel context.CancelFunc
}

type JobRegistry struct {
	mu   sync.Mutex
	jobs map[JobKey]Job
}

type JobKey struct {
	chatID int64
	method string
}

// New creats a new JobRegistry instance
func New() *JobRegistry {
	return &JobRegistry{jobs: make(map[JobKey]Job)}
}

// StartOnce starts fn in a goroutine if not already running for chatID.
// Returns started=false if there is already a running job.
func (r *JobRegistry) StartOnce(chatID int64, method string, fn func(ctx context.Context)) (started bool) {
	r.mu.Lock()
	key := JobKey{chatID: chatID, method: method}
	if _, exists := r.jobs[key]; exists {
		r.mu.Unlock()
		return false
	}

	ctx, cancel := context.WithCancel(context.Background())
	r.jobs[key] = Job{cancel: cancel}
	r.mu.Unlock()

	go func() {
		defer func() {
			r.mu.Lock()
			delete(r.jobs, key)
			r.mu.Unlock()
		}()
		fn(ctx)
	}()

	return true
}

// Stop stops an asynchoronous goroutine if it is running.
// Returns stopped=true if successful.
func (r *JobRegistry) Stop(chatID int64, method string) (stopped bool) {
	r.mu.Lock()
	job, ok := r.jobs[JobKey{chatID: chatID, method: method}]
	r.mu.Unlock()
	if !ok {
		return false
	}
	job.cancel()
	return true
}
