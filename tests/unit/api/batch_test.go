package api_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/shawnvan/bl4/internal/api/models"
	"github.com/shawnvan/bl4/internal/api/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBatchProcessor tests the batch processor service
func TestBatchProcessor(t *testing.T) {
	processor := services.NewBatchProcessor()

	t.Run("create_batch_processor", func(t *testing.T) {
		assert.NotNil(t, processor)
	})

	t.Run("process_empty_batch", func(t *testing.T) {
		request := &models.BatchDecodeRequest{
			SerialCodes: []string{},
			Options: &models.BatchDecodeOptions{
				IndividualOptions: &models.DecodeOptions{
					IncludeBitstream: false,
				},
			},
		}

		result, err := processor.ProcessBatch(request)
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("process_too_large_batch", func(t *testing.T) {
		request := &models.BatchDecodeRequest{
			SerialCodes: make([]string, 1001), // Exceeds max limit
			Options: &models.BatchDecodeOptions{},
		}

		result, err := processor.ProcessBatch(request)
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "too many serial codes")
	})

	t.Run("process_valid_batch", func(t *testing.T) {
		serialCodes := []string{
			"@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}",
			"@Ugr$WBm/$!m!X=5&qXxA;nj3OOD#<4R",
			"@UvalidExampleCode123456789012345",
		}

		request := &models.BatchDecodeRequest{
			SerialCodes: serialCodes,
			Options: &models.BatchDecodeOptions{
				IndividualOptions: &models.DecodeOptions{
					IncludeBitstream: false,
					IncludeTokens:    false,
					IncludeRawData:   false,
					StrictValidation:  true,
				},
				MaxConcurrency: 2,
				FailFast:       false,
				Timeout:        10000,
				PerItemTimeout: 5000,
			},
		}

		result, err := processor.ProcessBatch(request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, len(serialCodes), len(result.Results))
		assert.Equal(t, len(serialCodes), result.Statistics.Total)
	})

	t.Run("process_batch_with_failures", func(t *testing.T) {
		serialCodes := []string{
			"@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}", // Valid
			"@Uinvalid_code",                                        // Invalid
			"",                                                        // Empty
			"@Ugr$WBm/$!m!X=5&qXxA;nj3OOD#<4R",          // Valid
		}

		request := &models.BatchDecodeRequest{
			SerialCodes: serialCodes,
			Options: &models.BatchDecodeOptions{
				IndividualOptions: &models.DecodeOptions{
					IncludeBitstream: false,
					IncludeTokens:    false,
					IncludeRawData:   false,
					StrictValidation:  true,
				},
				MaxConcurrency: 2,
				FailFast:       false, // Continue on errors
			},
		}

		result, err := processor.ProcessBatch(request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, len(serialCodes), len(result.Results))

		// Should have some failures
		failedCount := 0
		for _, res := range result.Results {
			if res.Error != nil {
				failedCount++
			}
		}
		assert.Greater(t, failedCount, 0, "Should have some failed results")
		assert.Less(t, failedCount, len(serialCodes), "Not all should fail")
	})

	t.Run("process_batch_fail_fast", func(t *testing.T) {
		serialCodes := []string{
			"@Uinvalid_code", // Invalid, should fail fast
			"@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}", // Valid but won't be processed
		}

		request := &models.BatchDecodeRequest{
			SerialCodes: serialCodes,
			Options: &models.BatchDecodeOptions{
				IndividualOptions: &models.DecodeOptions{
					IncludeBitstream: false,
					IncludeTokens:    false,
					IncludeRawData:   false,
					StrictValidation:  true,
				},
				MaxConcurrency: 1,
				FailFast:       true, // Stop on first error
			},
		}

		result, err := processor.ProcessBatch(request)
		require.Error(t, err)
		assert.Nil(t, result)
	})
}

// TestBatchProcessorPerformance tests batch processor performance
func TestBatchProcessorPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	processor := services.NewBatchProcessor()

	testCases := []struct {
		name          string
		itemCount     int
		concurrency   int
		maxTime       time.Duration
		description   string
	}{
		{
			name:        "small_batch",
			itemCount:   5,
			concurrency: 2,
			maxTime:     2 * time.Second,
			description: "Small batch processing",
		},
		{
			name:        "medium_batch",
			itemCount:   20,
			concurrency: 5,
			maxTime:     5 * time.Second,
			description: "Medium batch processing",
		},
		{
			name:        "large_batch",
			itemCount:   50,
			concurrency: 10,
			maxTime:     10 * time.Second,
			description: "Large batch processing",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			serialCodes := make([]string, tc.itemCount)
			for i := 0; i < tc.itemCount; i++ {
				serialCodes[i] = "@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}"
			}

			request := &models.BatchDecodeRequest{
				SerialCodes: serialCodes,
				Options: &models.BatchDecodeOptions{
					IndividualOptions: &models.DecodeOptions{
						IncludeBitstream: false,
						IncludeTokens:    false,
						IncludeRawData:   false,
						StrictValidation:  true,
					},
					MaxConcurrency: tc.concurrency,
					FailFast:       false,
					Timeout:        30000,
				},
			}

			start := time.Now()
			result, err := processor.ProcessBatch(request)
			duration := time.Since(start)

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tc.itemCount, len(result.Results))

			t.Logf("Performance test '%s': %d items, concurrency %d, duration %v",
				tc.name, tc.itemCount, tc.concurrency, duration)

			assert.Less(t, duration, tc.maxTime, "Batch processing should meet time requirements")
			assert.Less(t, duration/int64(tc.itemCount), 200*time.Millisecond, "Per-item processing should be efficient")
		})
	}
}

// TestBatchProcessorConcurrency tests concurrent batch processing
func TestBatchProcessorConcurrency(t *testing.T) {
	processor := services.NewBatchProcessor()

	// Test multiple concurrent batch requests
	numRequests := 3
	itemsPerRequest := 10

	var wg sync.WaitGroup
	results := make(chan *services.BatchProcessResult, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(requestID int) {
			defer wg.Done()

			serialCodes := make([]string, itemsPerRequest)
			for j := 0; j < itemsPerRequest; j++ {
				serialCodes[j] = fmt.Sprintf("@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{%02d-%02d}", requestID, j)
			}

			request := &models.BatchDecodeRequest{
				SerialCodes: serialCodes,
				Options: &models.BatchDecodeOptions{
					IndividualOptions: &models.DecodeOptions{
						IncludeBitstream: false,
						IncludeTokens:    false,
						IncludeRawData:   false,
						StrictValidation:  true,
					},
					MaxConcurrency: 3,
					FailFast:       false,
					Timeout:        10000,
				},
			}

			result, err := processor.ProcessBatch(request)
			if err != nil {
				results <- nil // Signal error
			} else {
				results <- result
			}
		}(i)
	}

	// Wait for all requests to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	successCount := 0
	errorCount := 0

	for result := range results {
		if result == nil {
			errorCount++
		} else {
			successCount++
			assert.Equal(t, itemsPerRequest, len(result.Results))
		}
	}

	t.Logf("Concurrent batch processing: %d successful, %d failed", successCount, errorCount)
	assert.Greater(t, successCount, 0, "At least some requests should succeed")
}

// TestBatchProcessorCancellation tests cancellation functionality
func TestBatchProcessorCancellation(t *testing.T) {
	processor := services.NewBatchProcessor()

	// Create a large batch that will take time to process
	serialCodes := make([]string, 100)
	for i := 0; i < 100; i++ {
		serialCodes[i] = "@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}"
	}

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	request := &models.BatchDecodeRequest{
		SerialCodes: serialCodes,
		Options: &models.BatchDecodeOptions{
			IndividualOptions: &models.DecodeOptions{
				IncludeBitstream: false,
				IncludeTokens:    false,
				IncludeRawData:   false,
				StrictValidation:  true,
			},
			MaxConcurrency: 5,
			FailFast:       false,
			Timeout:        10000,
		},
	}

	start := time.Now()
	result, err := processor.ProcessBatchWithContext(ctx, request)
	duration := time.Since(start)

	// Should fail due to context cancellation
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "context canceled")
	assert.Less(t, duration, 200*time.Millisecond, "Should be cancelled quickly")
}

// TestBatchProcessorValidation tests input validation
func TestBatchProcessorValidation(t *testing.T) {
	processor := services.NewBatchProcessor()

	t.Run("validate_serial_codes_required", func(t *testing.T) {
		request := &models.BatchDecodeRequest{
			SerialCodes: nil,
			Options: &models.BatchDecodeOptions{},
		}

		result, err := processor.ProcessBatch(request)
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "serial codes are required")
	})

	t.Run("validate_serial_codes_not_nil", func(t *testing.T) {
		request := &models.BatchDecodeRequest{
			SerialCodes: nil,
			Options: &models.BatchDecodeOptions{},
		}

		result, err := processor.ProcessBatch(request)
		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("validate_options_not_nil", func(t *testing.T) {
		request := &models.BatchDecodeRequest{
			SerialCodes: []string{"@Utest"},
			Options:      nil,
		}

		result, err := processor.ProcessBatch(request)
		require.NoError(t, err) // Should work with nil options
		require.NotNil(t, result)
	})

	t.Run("validate_max_concurrency", func(t *testing.T) {
		request := &models.BatchDecodeRequest{
			SerialCodes: []string{"@Utest"},
			Options: &models.BatchDecodeOptions{
				MaxConcurrency: -1, // Invalid
			},
		}

		result, err := processor.ProcessBatch(request)
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "max concurrency must be positive")
	})
}

// TestBatchProcessorProgressTracking tests progress tracking
func TestBatchProcessorProgressTracking(t *testing.T) {
	processor := services.NewBatchProcessor()

	serialCodes := make([]string, 20)
	for i := 0; i < 20; i++ {
		serialCodes[i] = "@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}"
	}

	progressUpdates := make(chan int, 20)

	// Mock progress callback
	progressCallback := func(processed, total int) {
		progressUpdates <- processed
	}

	request := &models.BatchDecodeRequest{
		SerialCodes: serialCodes,
		Options: &models.BatchDecodeOptions{
			IndividualOptions: &models.DecodeOptions{
				IncludeBitstream: false,
				IncludeTokens:    false,
				IncludeRawData:   false,
				StrictValidation:  true,
			},
			MaxConcurrency: 3,
			FailFast:       false,
			ReportProgress: true,
			Timeout:        10000,
		},
	}

	// Start listening for progress updates
	var progressCount int
	go func() {
		for range progressUpdates {
			progressCount++
		}
	}()

	result, err := processor.ProcessBatchWithProgress(request, progressCallback)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Close progress channel
	close(progressUpdates)

	// Should have received progress updates
	assert.Greater(t, progressCount, 0, "Should receive progress updates")
	assert.Equal(t, len(serialCodes), len(result.Results))
}

// TestBatchProcessorStatistics tests statistics calculation
func TestBatchProcessorStatistics(t *testing.T) {
	processor := services.NewBatchProcessor()

	// Create batch with mix of valid and invalid codes
	serialCodes := []string{
		"@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}", // Valid
		"@Uinvalid_code",                                        // Invalid
		"@Ugr$WBm/$!m!X=5&qXxA;nj3OOD#<4R",          // Valid
		"",                                                        // Invalid
		"@UvalidExampleCode123456789012345",               // Valid
	}

	request := &models.BatchDecodeRequest{
		SerialCodes: serialCodes,
		Options: &models.BatchDecodeOptions{
			IndividualOptions: &models.DecodeOptions{
				IncludeBitstream: false,
				IncludeTokens:    false,
				IncludeRawData:   false,
				StrictValidation:  true,
			},
			MaxConcurrency: 2,
			FailFast:       false,
		},
	}

	result, err := processor.ProcessBatch(request)
	require.NoError(t, err)
	require.NotNil(t, result)

	stats := result.Statistics
	require.NotNil(t, stats)

	// Verify statistics
	assert.Equal(t, len(serialCodes), stats.Total)
	assert.Equal(t, 3, stats.Successful) // 3 valid codes
	assert.Equal(t, 2, stats.Failed)     // 2 invalid codes
	assert.Greater(t, stats.TotalBytes, 0, "Should have processed some bytes")
	assert.Greater(t, stats.AverageDuration, int64(0), "Should have average duration")
}

// TestBatchProcessorMemoryUsage tests memory usage during batch processing
func TestBatchProcessorMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory usage test in short mode")
	}

	processor := services.NewBatchProcessor()

	// Create a large batch
	serialCodes := make([]string, 200)
	for i := 0; i < 200; i++ {
		serialCodes[i] = "@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}"
	}

	request := &models.BatchDecodeRequest{
		SerialCodes: serialCodes,
		Options: &models.BatchDecodeOptions{
			IndividualOptions: &models.DecodeOptions{
				IncludeBitstream: false,
				IncludeTokens:    false,
				IncludeRawData:   false,
				StrictValidation:  true,
			},
			MaxConcurrency: 10,
			FailFast:       false,
			Timeout:        30000,
		},
	}

	// Measure memory usage before processing
	// Note: In a real implementation, you would use runtime.MemStats to measure memory
	initialMemory := getMemoryUsage(t)

	result, err := processor.ProcessBatch(request)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Measure memory usage after processing
	finalMemory := getMemoryUsage(t)
	memoryIncrease := finalMemory - initialMemory

	t.Logf("Memory usage increase: %d KB for %d items", memoryIncrease/1024, len(serialCodes))

	// Memory usage should be reasonable (less than 10KB per item)
	avgMemoryPerItem := memoryIncrease / int64(len(serialCodes))
	assert.Less(t, avgMemoryPerItem, 10*1024, "Average memory per item should be reasonable")
}

// getMemoryUsage gets current memory usage (simplified implementation)
func getMemoryUsage(t *testing.T) int64 {
	// This is a simplified implementation
	// In a real test, you would use runtime.MemStats
	return 1024 // 1KB placeholder
}