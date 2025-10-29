package processor

import (
	"container/heap"
	"sync"
)

type JobQueue struct {
	jobs     *PriorityQueue
	workers  int
	jobChan  chan *Job
	stopChan chan struct{}
	mu       sync.RWMutex
	jobsMap  map[string]*Job
}

type PriorityQueue []*Job

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	if pq[i].Priority != pq[j].Priority {
		return pq[i].Priority > pq[j].Priority
	}
	return pq[i].CreatedAt.Before(pq[j].CreatedAt)
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Push(x interface{}) {
	item := x.(*Job)
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	*pq = old[0 : n-1]
	return item
}

func NewJobQueue(maxWorkers, queueSize int) *JobQueue {
	pq := make(PriorityQueue, 0)
	heap.Init(&pq)

	return &JobQueue{
		jobs:     &pq,
		workers:  maxWorkers,
		jobChan:  make(chan *Job, queueSize),
		stopChan: make(chan struct{}),
		jobsMap:  make(map[string]*Job),
	}
}

func (jq *JobQueue) Enqueue(job *Job) error {
	jq.mu.Lock()
	defer jq.mu.Unlock()

	if _, exists := jq.jobsMap[job.ID]; exists {
		return ErrJobAlreadyExists
	}

	heap.Push(jq.jobs, job)
	jq.jobsMap[job.ID] = job

	select {
	case jq.jobChan <- job:
	default:
		return ErrQueueFull
	}

	return nil
}

func (jq *JobQueue) Dequeue() *Job {
	jq.mu.Lock()
	defer jq.mu.Unlock()

	if jq.jobs.Len() == 0 {
		return nil
	}

	job := heap.Pop(jq.jobs).(*Job)
	delete(jq.jobsMap, job.ID)

	return job
}

func (jq *JobQueue) GetJob(id string) (*Job, bool) {
	jq.mu.RLock()
	defer jq.mu.RUnlock()

	job, exists := jq.jobsMap[id]
	return job, exists
}

func (jq *JobQueue) UpdateJobStatus(id string, status JobStatus) bool {
	jq.mu.Lock()
	defer jq.mu.Unlock()

	job, exists := jq.jobsMap[id]
	if !exists {
		return false
	}

	job.Status = status
	return true
}

func (jq *JobQueue) UpdateJobProgress(id string, progress float64) bool {
	jq.mu.Lock()
	defer jq.mu.Unlock()

	job, exists := jq.jobsMap[id]
	if !exists {
		return false
	}

	job.UpdateProgress(progress)
	return true
}

func (jq *JobQueue) ListJobs() []*Job {
	jq.mu.RLock()
	defer jq.mu.RUnlock()

	jobs := make([]*Job, 0, len(jq.jobsMap))
	for _, job := range jq.jobsMap {
		jobs = append(jobs, job)
	}

	return jobs
}

func (jq *JobQueue) ListJobsByStatus(status JobStatus) []*Job {
	jq.mu.RLock()
	defer jq.mu.RUnlock()

	jobs := make([]*Job, 0)
	for _, job := range jq.jobsMap {
		if job.Status == status {
			jobs = append(jobs, job)
		}
	}

	return jobs
}

func (jq *JobQueue) Size() int {
	jq.mu.RLock()
	defer jq.mu.RUnlock()

	return jq.jobs.Len()
}

func (jq *JobQueue) CancelJob(id string) bool {
	jq.mu.Lock()
	defer jq.mu.Unlock()

	job, exists := jq.jobsMap[id]
	if !exists {
		return false
	}

	if job.Status == JobStatusProcessing {
		return false
	}

	job.Cancel()
	return true
}

func (jq *JobQueue) GetStats() QueueStats {
	jq.mu.RLock()
	defer jq.mu.RUnlock()

	stats := QueueStats{
		Total:      len(jq.jobsMap),
		Pending:    0,
		Processing: 0,
		Completed:  0,
		Failed:     0,
		Cancelled:  0,
	}

	for _, job := range jq.jobsMap {
		switch job.Status {
		case JobStatusPending:
			stats.Pending++
		case JobStatusProcessing:
			stats.Processing++
		case JobStatusCompleted:
			stats.Completed++
		case JobStatusFailed:
			stats.Failed++
		case JobStatusCancelled:
			stats.Cancelled++
		}
	}

	return stats
}

func (jq *JobQueue) Start() {
}

func (jq *JobQueue) Stop() {
	close(jq.stopChan)
}

type QueueStats struct {
	Total      int `json:"total"`
	Pending    int `json:"pending"`
	Processing int `json:"processing"`
	Completed  int `json:"completed"`
	Failed     int `json:"failed"`
	Cancelled  int `json:"cancelled"`
}

var (
	ErrJobAlreadyExists = &JobQueueError{"job already exists"}
	ErrQueueFull        = &JobQueueError{"queue is full"}
	ErrJobNotFound      = &JobQueueError{"job not found"}
)

type JobQueueError struct {
	message string
}

func (e *JobQueueError) Error() string {
	return e.message
}
