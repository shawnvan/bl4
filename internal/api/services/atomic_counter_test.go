package services

import (
    "sync"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
)

func TestAtomicCounter(t *testing.T) {
    t.Run("basic_operations", func(t *testing.T) {
        counter := NewAtomicCounter(0)
        
        assert.Equal(t, int64(0), counter.Get())
        assert.Equal(t, int64(5), counter.Add(5))
        assert.Equal(t, int64(5), counter.Get())
        assert.Equal(t, int64(6), counter.Increment())
        assert.Equal(t, int64(5), counter.Decrement())
        counter.Set(10)
        assert.Equal(t, int64(10), counter.Get())
    })

    t.Run("concurrent_access", func(t *testing.T) {
        counter := NewAtomicCounter(0)
        
        const numGoroutines = 100
        const incrementsPerGoroutine = 1000
        
        var wg sync.WaitGroup
        wg.Add(numGoroutines)
        
        for i := 0; i < numGoroutines; i++ {
            go func() {
                defer wg.Done()
                for j := 0; j < incrementsPerGoroutine; j++ {
                    counter.Increment()
                }
            }()
        }
        
        wg.Wait()
        
        expected := int64(numGoroutines * incrementsPerGoroutine)
        assert.Equal(t, expected, counter.Get())
    })
}

func TestSafeBatchProgress(t *testing.T) {
    t.Run("basic_progress_tracking", func(t *testing.T) {
        progress := NewSafeBatchProgress(100)
        
        assert.Equal(t, float64(0.0), progress.GetProgress())
        assert.Equal(t, int64(0), progress.Processed.Get())
        assert.Equal(t, int64(0), progress.Successful.Get())
        assert.Equal(t, int64(0), progress.Failed.Get())
        
        progress.AddSuccess()
        assert.Equal(t, float64(0.01), progress.GetProgress())
        assert.Equal(t, int64(1), progress.Successful.Get())
        
        progress.AddFailure()
        assert.Equal(t, float64(0.02), progress.GetProgress())
        assert.Equal(t, int64(1), progress.Failed.Get())
    })
}

func TestProgressWorkerPool(t *testing.T) {
    t.Run("worker_pool_lifecycle", func(t *testing.T) {
        pool := NewSimpleProgressWorkerPool(3)
        
        // Start pool
        pool.Start()
        
        // Create a mock subscriber
        subscriber := &MockProgressSubscriber{
            updates: make(chan *ProgressUpdate, 10),
        }
        
        // Submit some jobs
        update := &ProgressUpdate{
            BatchID:   "test-batch",
            Timestamp:  time.Now(),
            IsComplete: false,
        }
        
        pool.SubmitJob(subscriber, update)
        
        // Wait a bit for processing
        time.Sleep(100 * time.Millisecond)
        
        // Stop pool
        pool.Stop()
        
        // Verify updates were processed
        select {
        case receivedUpdate := <-subscriber.updates:
            assert.Equal(t, "test-batch", receivedUpdate.BatchID)
        case <-time.After(1 * time.Second):
            t.Error("Expected progress update was not received")
        }
    })
}

// MockProgressSubscriber is a mock implementation of ProgressSubscriber for testing
type MockProgressSubscriber struct {
    updates chan *ProgressUpdate
    mu      sync.Mutex
}

func (m *MockProgressSubscriber) OnProgress(update *ProgressUpdate) {
    select {
    case m.updates <- update:
    default:
        // Channel full, skip this update
    }
}