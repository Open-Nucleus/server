package service

import (
	"sync"
	"time"
)

// SyncJob represents a pending sync operation.
type SyncJob struct {
	PeerNodeID string
	Transport  string
	Priority   int // lower = higher priority
	RetryCount int
	NextRetry  time.Time
	CreatedAt  time.Time
}

// SyncQueue manages sync jobs with priority ordering.
type SyncQueue struct {
	mu   sync.Mutex
	jobs []*SyncJob
}

// NewSyncQueue creates a new sync queue.
func NewSyncQueue() *SyncQueue {
	return &SyncQueue{}
}

// Push adds a sync job to the queue.
func (sq *SyncQueue) Push(job *SyncJob) {
	sq.mu.Lock()
	defer sq.mu.Unlock()

	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now()
	}
	sq.jobs = append(sq.jobs, job)
	sq.sort()
}

// Pop removes and returns the highest-priority ready job.
func (sq *SyncQueue) Pop() *SyncJob {
	sq.mu.Lock()
	defer sq.mu.Unlock()

	now := time.Now()
	for i, job := range sq.jobs {
		if now.After(job.NextRetry) || job.NextRetry.IsZero() {
			sq.jobs = append(sq.jobs[:i], sq.jobs[i+1:]...)
			return job
		}
	}
	return nil
}

// Len returns the number of pending jobs.
func (sq *SyncQueue) Len() int {
	sq.mu.Lock()
	defer sq.mu.Unlock()
	return len(sq.jobs)
}

func (sq *SyncQueue) sort() {
	// Simple insertion sort — queue is small
	for i := 1; i < len(sq.jobs); i++ {
		key := sq.jobs[i]
		j := i - 1
		for j >= 0 && sq.jobs[j].Priority > key.Priority {
			sq.jobs[j+1] = sq.jobs[j]
			j--
		}
		sq.jobs[j+1] = key
	}
}
