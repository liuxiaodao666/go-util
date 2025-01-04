package gopool

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestNewWorkerPool(t *testing.T) {
	pool := NewWorkerPool(4, 4)
	pool.Start(2)

	// Submit jobs here...
	job := &ExampleJob{}
	pool.Submit(job)
	pool.Submit(job)
	pool.Submit(job)
	pool.Submit(job)
	pool.Submit(job)
	pool.Submit(job)
	pool.Submit(job)
	pool.Submit(job)
	pool.Submit(job)

	time.Sleep(5 * time.Second)
}

// ExampleJob is an example implementation of the Job interface.
type ExampleJob struct{}

// Run executes the job's work.
func (ej *ExampleJob) Run(ctx context.Context) error {
	// Simulate some work being done
	select {
	case <-time.After(2 * time.Second):
		fmt.Println("ExampleJob completed successfully.")
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}
