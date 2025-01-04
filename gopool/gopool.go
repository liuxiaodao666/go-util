package gopool

import (
	"context"
	"fmt"
	"github.com/liuxiaodao666/go-util/logger"
	"sync"
)

// Job is an interface that represents a unit of work to be executed by a worker.
type Job interface {
	Run(ctx context.Context) error
}

// WorkerPool is the main structure that holds the pool of workers.
type WorkerPool struct {
	taskQueue chan Job
	//wg        sync.WaitGroup
	mu     sync.Mutex
	active int // Number of active workers
	max    int // Maximum number of workers
}

// NewWorkerPool initializes and returns a new WorkerPool with the given maxWorkers.
func NewWorkerPool(maxWorkers int, maxWaitJobs int) *WorkerPool {
	return &WorkerPool{
		taskQueue: make(chan Job, maxWaitJobs), // Buffered channel for jobs
		max:       maxWorkers,
	}
}

// Start starts the worker pool with a specified number of workers.
func (wp *WorkerPool) Start(numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		wp.startWorker()
	}
}

// startWorker creates a new worker goroutine that listens on the task queue.
func (wp *WorkerPool) startWorker() {
	wp.mu.Lock()
	if wp.active >= wp.max {
		wp.mu.Unlock()
		logger.Warnf("active workers reach max [%d], can't start new worker", wp.active)
		return
	}
	wp.active++
	wp.mu.Unlock()

	//wp.wg.Add(1)
	go func() {
		//defer wp.wg.Done()
		for job := range wp.taskQueue {
			//ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err := job.Run(context.Background())
			//cancel()
			if err != nil {
				logger.Errorf("Error executing job: %v\n", err)
			}
		}
	}()
}

// Submit submits a job to the worker pool for execution.
func (wp *WorkerPool) Submit(job Job) {
	select {
	case wp.taskQueue <- job:
	default:
		fmt.Println("Task queue full, dropping job.")
	}
}

// Stop stops the worker pool gracefully.
func (wp *WorkerPool) Stop() {
	close(wp.taskQueue)
	//wp.wg.Wait()
}

// Stats returns statistics about the worker pool.
func (wp *WorkerPool) Stats() map[string]int {
	stats := make(map[string]int)
	wp.mu.Lock()
	stats["active_workers"] = wp.active
	wp.mu.Unlock()
	return stats
}

// ErrorHandling handles errors from jobs.
func (wp *WorkerPool) ErrorHandling(err error) {
	// Implement error handling logic here.
}

// Cleanup releases resources held by the worker pool.
func (wp *WorkerPool) Cleanup() {
	// Implement resource cleanup logic here.
}

// Log logs information about the worker pool.
func (wp *WorkerPool) Log(msg string) {
	fmt.Println(msg)
}
