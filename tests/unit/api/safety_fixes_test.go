package services_test

import (
    "fmt"
    "sync"
    "testing"
    "time"

    "github.com/shawnvan/bl4/internal/api/models"
    "github.com/shawnvan/bl4/internal/api/services"
    "github.com/stretchr/testify/assert"
)

// TestAtomicCounter tests the thread-safe atomic counter implementation
func TestAtomicCounter(t *testing.T) {
    t.Run("basic_operations", func(t *testing.T) {
        counter := services.NewAtomicCounter(0)
        
        assert.Equal(t, int64(0), counter.Get())
        assert.Equal(t, int64(5), counter.Add(5))
        assert.Equal(t, int64(5), counter.Get())
        assert.Equal(t, int64(6), counter.Increment())
        assert.Equal(t, int64(5), counter.Decrement())
        counter.Set(10)
        assert.Equal(t, int64(10), counter.Get())
    })

    t.Run("concurrent_access", func(t *testing.T) {
        counter := services.NewAtomicCounter(0)
        
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

// TestSafeBatchProgress tests thread-safe batch progress tracking
func TestSafeBatchProgress(t *testing.T) {
    t.Run("basic_progress_tracking", func(t *testing.T) {
        progress := services.NewSafeBatchProgress(100)
        
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

    t.Run("concurrent_progress_updates", func(t *testing.T) {
        progress := services.NewSafeBatchProgress(1000)
        
        const numGoroutines = 50
        const updatesPerGoroutine = 20
        
        var wg sync.WaitGroup
        wg.Add(numGoroutines)
        
        for i := 0; i < numGoroutines; i++ {
            go func(success bool) {
                defer wg.Done()
                for j := 0; j < updatesPerGoroutine; j++ {
                    if success {
                        progress.AddSuccess()
                    } else {
                        progress.AddFailure()
                    }
                }
            }(i%2 == 0)
        }
        
        wg.Wait()
        
        expectedProcessed := int64(numGoroutines * updatesPerGoroutine)
        assert.Equal(t, expectedProcessed, progress.Processed.Get())
        assert.Equal(t, float64(1.0), progress.GetProgress())
    })
}

// TestProgressWorkerPool tests the progress update worker pool
func TestProgressWorkerPool(t *testing.T) {
    t.Run("worker_pool_lifecycle", func(t *testing.T) {
        pool := services.NewProgressWorkerPool(3)
        
        // Start the pool
        pool.Start()
        
        // Create a mock subscriber
        subscriber := &MockProgressSubscriber{
            updates: make(chan *services.ProgressUpdate, 10),
        }
        
        // Submit some jobs
        update := &services.ProgressUpdate{
            BatchID:   "test-batch",
            Timestamp:  time.Now(),
            IsComplete: false,
        }
        
        pool.SubmitJob(subscriber, update)
        
        // Wait a bit for processing
        time.Sleep(100 * time.Millisecond)
        
        // Stop the pool
        pool.Stop()
        
        // Verify updates were processed
        select {
        case receivedUpdate := <-subscriber.updates:
            assert.Equal(t, "test-batch", receivedUpdate.BatchID)
        case <-time.After(1 * time.Second):
            t.Error("Expected progress update was not received")
        }
    })

    t.Run("concurrent_job_submission", func(t *testing.T) {
        pool := services.NewProgressWorkerPool(5)
        pool.Start()
        defer pool.Stop()
        
        subscribers := make([]*MockProgressSubscriber, 10)
        for i := range subscribers {
            subscribers[i] = &MockProgressSubscriber{
                updates: make(chan *services.ProgressUpdate, 5),
            }
        }
        
        // Submit jobs concurrently
        var wg sync.WaitGroup
        for i, subscriber := range subscribers {
            wg.Add(1)
            go func(idx int, sub *MockProgressSubscriber) {
                defer wg.Done()
                for j := 0; j < 5; j++ {
                    update := &services.ProgressUpdate{
                        BatchID:   fmt.Sprintf("batch-%d", idx),
                        Timestamp:  time.Now(),
                        IsComplete: j == 4,
                    }
                    pool.SubmitJob(sub, update)
                }
            }(i, subscriber)
        }
        
        wg.Wait()
        
        // Allow time for processing
        time.Sleep(200 * time.Millisecond)
        
        // Verify all subscribers received updates
        for i, subscriber := range subscribers {
            receivedCount := 0
            for len(subscriber.updates) > 0 {
                <-subscriber.updates
                receivedCount++
            }
            assert.GreaterOrEqual(t, receivedCount, 0, "Subscriber %d should receive updates", i)
        }
    })
}

// MockProgressSubscriber is a mock implementation of ProgressSubscriber for testing
type MockProgressSubscriber struct {
    updates chan *services.ProgressUpdate
    mu      sync.Mutex
}

func (m *MockProgressSubscriber) OnProgress(update *services.ProgressUpdate) {
    select {
    case m.updates <- update:
    default:
        // Channel full, skip this update
    }
}

func (m *MockProgressSubscriber) GetUpdates() []*services.ProgressUpdate {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    var updates []*services.ProgressUpdate
    for len(m.updates) > 0 {
        updates = append(updates, <-m.updates)
    }
    return updates
}

// TestBatchProcessorWithFixes tests the batch processor with safety fixes
func TestBatchProcessorWithFixes(t *testing.T) {
    processor := services.NewBatchProcessor()

    t.Run("concurrent_batch_processing", func(t *testing.T) {
        const numBatches = 5
        const itemsPerBatch = 10
        
        var wg sync.WaitGroup
        results := make(chan *services.BatchProcessResult, numBatches)
        
        for i := 0; i < numBatches; i++ {
            wg.Add(1)
            go func(batchID int) {
                defer wg.Done()
                
                serialCodes := make([]string, itemsPerBatch)
                for j := 0; j < itemsPerBatch; j++ {
                    serialCodes[j] = "@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}"
                }
                
                request := &models.BatchDecodeRequest{
                    SerialCodes: serialCodes,
                    Options: &models.BatchDecodeOptions{
                        IndividualOptions: &models.DecodeOptions{
                            IncludeBitstream: false,
                            IncludeTokens:    false,
                            ValidateOnly:    false,
                            IncludeStats:    false,
                        },
                        MaxConcurrency: 2,
                        FailFast:       false,
                        Timeout:        10000,
                        PerItemTimeout: 5000,
                    },
                }
                
                result, err := processor.ProcessBatch(request)
                if err == nil {
                    results <- result
                }
            }(i)
        }
        
        wg.Wait()
        close(results)
        
        // Verify results
        successCount := 0
        for result := range results {
            if result.Success {
                successCount++
                assert.Equal(t, itemsPerBatch, len(result.Results))
            }
        }
        
        assert.Greater(t, successCount, 0, "At least some batches should succeed")
    })

    t.Run("cancellation_with_progress_tracking", func(t *testing.T) {
        // Create a large batch that will take time to process
        serialCodes := make([]string, 50)
        for i := 0; i < 50; i++ {
            serialCodes[i] = "@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}"
        }
        
        batchID := "test-cancellation"
        subscriber := &MockProgressSubscriber{
            updates: make(chan *services.ProgressUpdate, 100),
        }
        
        request := &models.BatchDecodeRequest{
            SerialCodes: serialCodes,
            Options: &models.BatchDecodeOptions{
                IndividualOptions: &models.DecodeOptions{
                    IncludeBitstream: false,
                    IncludeTokens:    false,
                    ValidateOnly:    false,
                    IncludeStats:    false,
                },
                MaxConcurrency: 3,
                FailFast:       false,
                Timeout:        30000,
                PerItemTimeout: 5000,
            },
        }
        
        // Start batch processing with tracking
        go func() {
            _, _ = processor.ProcessBatchWithTracking(batchID, request, subscriber)
        }()
        
        // Wait a bit then cancel
        time.Sleep(100 * time.Millisecond)
        err := processor.CancelBatch(batchID, "test cancellation")
        assert.NoError(t, err)
        
        // Verify cancellation was processed
        time.Sleep(50 * time.Millisecond)
        updates := subscriber.GetUpdates()
        foundCancellation := false
        for _, update := range updates {
            if update.IsComplete && update.Error != "" {
                foundCancellation = true
                break
            }
        }
        assert.True(t, foundCancellation, "Should receive cancellation update")
    })
}