package processor

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type WorkerPool struct {
	workers    int
	jobQueue   *JobQueue
	processor  JobProcessor
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	activeJobs map[string]*Job
	mu         sync.RWMutex
}

type JobProcessor interface {
	ProcessJob(ctx context.Context, job *Job) JobResult
}

func NewWorkerPool(workers int, jobQueue *JobQueue, processor JobProcessor) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	return &WorkerPool{
		workers:    workers,
		jobQueue:   jobQueue,
		processor:  processor,
		ctx:        ctx,
		cancel:     cancel,
		activeJobs: make(map[string]*Job),
	}
}

func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
	log.Printf("Started %d workers", wp.workers)
}

func (wp *WorkerPool) Stop() {
	log.Println("Stopping worker pool...")
	wp.cancel()
	wp.wg.Wait()
	log.Println("Worker pool stopped")
}

func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()

	log.Printf("Worker %d started", id)

	for {
		select {
		case <-wp.ctx.Done():
			log.Printf("Worker %d stopping", id)
			return
		default:
			job := wp.jobQueue.Dequeue()
			if job == nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			wp.processJob(id, job)
		}
	}
}

func (wp *WorkerPool) processJob(workerID int, job *Job) {
	wp.mu.Lock()
	wp.activeJobs[job.ID] = job
	wp.mu.Unlock()

	defer func() {
		wp.mu.Lock()
		delete(wp.activeJobs, job.ID)
		wp.mu.Unlock()
	}()

	log.Printf("Worker %d processing job %s (%s)", workerID, job.ID, job.Type)

	job.Start()
	wp.jobQueue.UpdateJobStatus(job.ID, JobStatusProcessing)

	result := wp.processor.ProcessJob(wp.ctx, job)

	if result.Success {
		job.Complete(result)
		wp.jobQueue.UpdateJobStatus(job.ID, JobStatusCompleted)
		log.Printf("Worker %d completed job %s successfully", workerID, job.ID)
	} else {
		job.Fail(fmt.Errorf(result.Message))
		wp.jobQueue.UpdateJobStatus(job.ID, JobStatusFailed)
		log.Printf("Worker %d failed job %s: %s", workerID, job.ID, result.Message)
	}
}

func (wp *WorkerPool) GetActiveJobs() []*Job {
	wp.mu.RLock()
	defer wp.mu.RUnlock()

	jobs := make([]*Job, 0, len(wp.activeJobs))
	for _, job := range wp.activeJobs {
		jobs = append(jobs, job)
	}

	return jobs
}

func (wp *WorkerPool) GetWorkerCount() int {
	return wp.workers
}

func (wp *WorkerPool) UpdateWorkerCount(newCount int) {
	if newCount <= 0 {
		return
	}

	currentCount := wp.workers
	if newCount == currentCount {
		return
	}

	wp.workers = newCount

	if newCount > currentCount {
		for i := currentCount; i < newCount; i++ {
			wp.wg.Add(1)
			go wp.worker(i)
		}
		log.Printf("Added %d workers (total: %d)", newCount-currentCount, newCount)
	} else {
		log.Printf("Worker count reduced to %d (excess workers will stop naturally)", newCount)
	}
}

func (wp *WorkerPool) GetStats() WorkerPoolStats {
	wp.mu.RLock()
	defer wp.mu.RUnlock()

	return WorkerPoolStats{
		TotalWorkers: wp.workers,
		ActiveJobs:   len(wp.activeJobs),
		IsRunning:    wp.ctx.Err() == nil,
	}
}

type WorkerPoolStats struct {
	TotalWorkers int  `json:"total_workers"`
	ActiveJobs   int  `json:"active_jobs"`
	IsRunning    bool `json:"is_running"`
}
