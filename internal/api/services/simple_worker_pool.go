package services

import (
    "sync"
)

// ProgressJob represents a progress update job
type ProgressJob struct {
    Subscriber ProgressSubscriber
    Update     *ProgressUpdate
}

// SimpleProgressWorkerPool manages goroutines for progress updates (simplified)
type SimpleProgressWorkerPool struct {
    workers   int
    jobQueue  chan *ProgressJob
    quit      chan bool
    wg        sync.WaitGroup
}

// NewSimpleProgressWorkerPool creates a new simple worker pool for progress updates
func NewSimpleProgressWorkerPool(workers int) *SimpleProgressWorkerPool {
    return &SimpleProgressWorkerPool{
        workers:  workers,
        jobQueue: make(chan *ProgressJob, workers*2), // Buffered channel
        quit:     make(chan bool),
    }
}

// Start starts the worker pool
func (pwp *SimpleProgressWorkerPool) Start() {
    for i := 0; i < pwp.workers; i++ {
        pwp.wg.Add(1)
        go pwp.worker()
    }
}

// Stop gracefully shuts down the worker pool
func (pwp *SimpleProgressWorkerPool) Stop() {
    close(pwp.quit)
    pwp.wg.Wait()
}

// worker processes jobs from the queue
func (pwp *SimpleProgressWorkerPool) worker() {
    defer pwp.wg.Done()
    
    for {
        select {
        case job := <-pwp.jobQueue:
            if job != nil && job.Subscriber != nil && job.Update != nil {
                // Process job directly with panic recovery
                func() {
                    defer func() {
                        if r := recover(); r != nil {
                            // Log panic but don't crash worker
                        }
                    }()
                    job.Subscriber.OnProgress(job.Update)
                }()
            }
        case <-pwp.quit:
            return
        }
    }
}

// SubmitJob submits a progress update job
func (pwp *SimpleProgressWorkerPool) SubmitJob(subscriber ProgressSubscriber, update *ProgressUpdate) {
    job := &ProgressJob{
        Subscriber: subscriber,
        Update:     update,
    }

    select {
    case pwp.jobQueue <- job:
        // Job submitted successfully
    default:
        // Queue is full, skip this update to avoid blocking
    }
}