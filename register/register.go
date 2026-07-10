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
	jobs map[string]Job
}

// New creats a new JobRegistry instance
func New() *JobRegistry {
	return &JobRegistry{jobs: make(map[string]Job)}
}

// StartOnce starts fn in a goroutine if not already running for chatID.
// Returns started=false if there is already a running job.
func (r *JobRegistry) StartOnce(method string, fn func(ctx context.Context)) (started bool) {
	r.mu.Lock()
	if _, exists := r.jobs[method]; exists {
		r.mu.Unlock()
		return false
	}

	ctx, cancel := context.WithCancel(context.Background())
	r.jobs[method] = Job{cancel: cancel}
	r.mu.Unlock()

	go func() {
		defer func() {
			r.mu.Lock()
			delete(r.jobs, method)
			r.mu.Unlock()
		}()
		fn(ctx)
	}()

	return true
}

// Stop stops an asynchoronous goroutine if it is running.
// Returns stopped=true if successful.
func (r *JobRegistry) Stop(method string) (stopped bool) {
	r.mu.Lock()
	job, ok := r.jobs[method]
	r.mu.Unlock()
	if !ok {
		return false
	}
	job.cancel()
	return true
}

// Methods returns the active methods for a given chatID.
func (r *JobRegistry) Methods() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	methods := make([]string, 0)
	for method := range r.jobs {
		methods = append(methods, method)
	}
	return methods
}
