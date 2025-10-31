package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/shawnvan/bl4/internal/api"
	"github.com/shawnvan/bl4/internal/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBatchProcessingWorkflow tests end-to-end batch processing workflow
func TestBatchProcessingWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping batch processing test in short mode")
	}

	// Setup test server
	srv := setupTestServer(t)
	testServerURL := fmt.Sprintf("http://localhost:%d", srv.Port)

	// Wait for server to be ready
	err := waitForServerReady(testServerURL + "/health", 5*time.Second)
	require.NoError(t, err, "Test server should be ready")

	// Run test scenarios
	t.Run("small_batch_processing", func(t *testing.T) {
		testSmallBatchProcessing(t, testServerURL)
	})

	t.Run("large_batch_processing", func(t *testing.T) {
		testLargeBatchProcessing(t, testServerURL)
	})

	t.Run("mixed_batch_processing", func(t *testing.T) {
		testMixedBatchProcessing(t, testServerURL)
	})

	t.Run("error_handling_workflow", func(t *testing.T) {
		testErrorHandlingWorkflow(t, testServerURL)
	})

	t.Run("performance_validation", func(t *testing.T) {
		testBatchPerformanceValidation(t, testServerURL)
	})

	// Cleanup
	srv.Close()
}

// testSmallBatchProcessing tests processing of small batches
func testSmallBatchProcessing(t *testing.T, baseURL string) {
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
			ReportProgress: true,
			Timeout:        10000,
			PerItemTimeout: 2000,
		},
	}

	response, err := performBatchDecodeRequest(t, baseURL, request)
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.True(t, response.Success)

	// Verify results
	require.NotNil(t, response.Results)
	assert.Equal(t, len(serialCodes), len(response.Results))

	// Verify individual results
	for i, result := range response.Results {
		assert.NotNil(t, result)
		if result.Error == nil {
			// Successful decode
			assert.NotEmpty(t, result.SerialCode)
			assert.Equal(t, serialCodes[i], result.SerialCode)
			assert.NotNil(t, result.ItemData)
		}
	}

	// Verify statistics
	require.NotNil(t, response.Statistics)
	assert.Equal(t, len(serialCodes), response.Statistics.Total)
	assert.Equal(t, len(serialCodes)-countFailedResults(response.Results), response.Statistics.Successful)
	assert.Equal(t, countFailedResults(response.Results), response.Statistics.Failed)
}

// testLargeBatchProcessing tests processing of large batches
func testLargeBatchProcessing(t *testing.T, baseURL string) {
	// Generate a larger batch of serial codes
	serialCodes := make([]string, 50)
	for i := 0; i < 50; i++ {
		serialCodes[i] = fmt.Sprintf("@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{%02d}", i%10)
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
			MaxConcurrency: 5,
			FailFast:       false,
			ReportProgress: true,
			Timeout:        30000, // 30 seconds for larger batch
			PerItemTimeout: 3000,
		},
	}

	// Measure processing time
	start := time.Now()
	response, err := performBatchDecodeRequest(t, baseURL, request)
	duration := time.Since(start)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.True(t, response.Success)

	// Verify performance
	require.NotNil(t, response.Statistics)
	assert.Equal(t, len(serialCodes), response.Statistics.Total)
	assert.Equal(t, len(serialCodes)-countFailedResults(response.Results), response.Statistics.Successful)

	// Performance assertions
	t.Logf("Large batch processing: %d items in %v", len(serialCodes), duration)
	assert.Less(t, duration, 15*time.Second, "Large batch should complete within 15 seconds")
	assert.Less(t, duration/int64(len(serialCodes)), 300*time.Millisecond, "Per-item processing should be <300ms")
}

// testMixedBatchProcessing tests processing with both valid and invalid codes
func testMixedBatchProcessing(t *testing.T, baseURL string) {
	serialCodes := []string{
		"@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}",     // Valid
		"@Uinvalid_code_format",                                   // Invalid
		"@Ugr$WBm/$!m!X=5&qXxA;nj3OOD#<4R",          // Valid
		"",                                                        // Empty (invalid)
		"@UvalidExampleCode123456789012345",               // Valid
		"@Uanother@invalid@code",                               // Invalid
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
			FailFast:       false, // Continue processing on errors
			ReportProgress: true,
			Timeout:        10000,
		},
	}

	response, err := performBatchDecodeRequest(t, baseURL, request)
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.True(t, response.Success) // Batch should succeed despite individual failures

	// Verify that some results failed
	require.NotNil(t, response.Results)
	assert.Equal(t, len(serialCodes), len(response.Results))

	failedCount := countFailedResults(response.Results)
	assert.Greater(t, failedCount, 0, "Some results should have failed with invalid codes")

	// Verify that some results succeeded
	successCount := len(response.Results) - failedCount
	assert.Greater(t, successCount, 0, "Some results should have succeeded with valid codes")

	// Verify that valid codes produced successful results
	for i, code := range serialCodes {
		if isValidSerialCode(code) {
			assert.NotNil(t, response.Results[i].ItemData, "Valid code should produce item data")
		}
	}

	// Verify that invalid codes produced errors
	for i, code := range serialCodes {
		if !isValidSerialCode(code) && code != "" {
			assert.NotNil(t, response.Results[i].Error, "Invalid code should produce error")
		}
	}
}

// testErrorHandlingWorkflow tests various error handling scenarios
func testErrorHandlingWorkflow(t *testing.T, baseURL string) {
	t.Run("empty_batch_request", func(t *testing.T) {
		request := &models.BatchDecodeRequest{
			SerialCodes: []string{},
			Options: &models.BatchDecodeOptions{},
		}

		response, err := performBatchDecodeRequest(t, baseURL, request)
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.False(t, response.Success)
		assert.NotNil(t, response.Error)
	})

	t.Run("to_large_batch_request", func(t *testing.T) {
		// Create batch exceeding maximum size
		serialCodes := make([]string, 1001)
		for i := 0; i < 1001; i++ {
			serialCodes[i] = fmt.Sprintf("@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{%02d}", i%100)
		}

		request := &models.BatchDecodeRequest{
			SerialCodes: serialCodes,
			Options: &models.BatchDecodeOptions{},
		}

		response, err := performBatchDecodeRequest(t, baseURL, request)
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.False(t, response.Success)
		assert.NotNil(t, response.Error)
	})

	t.Run("invalid_options_request", func(t *testing.T) {
		serialCodes := []string{"@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}"}
		request := &models.BatchDecodeRequest{
			SerialCodes: serialCodes,
			Options: &models.BatchDecodeOptions{
				MaxConcurrency: -5, // Invalid negative concurrency
			},
		}

		response, err := performBatchDecodeRequest(t, baseURL, request)
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.False(t, response.Success)
		assert.NotNil(t, response.Error)
	})
}

// testBatchPerformanceValidation validates batch processing performance
func testBatchPerformanceValidation(t *testing.T, baseURL string) {
	if testing.Short() {
		t.Skip("Skipping performance validation in short mode")
	}

	testCases := []struct {
		name          string
		itemCount     int
		concurrency   int
		expectedTime  time.Duration
		description   string
	}{
		{
			name:         "small_batch_low_concurrency",
			itemCount:    5,
			concurrency:  1,
			expectedTime: 2 * time.Second,
			description: "Small batch with single-threaded processing",
		},
		{
			name:         "small_batch_high_concurrency",
			itemCount:    5,
			concurrency:  5,
			expectedTime: 1 * time.Second,
			description: "Small batch with parallel processing",
		},
		{
			name:         "medium_batch_low_concurrency",
			itemCount:    20,
			concurrency:  2,
			expectedTime: 5 * time.Second,
			description: "Medium batch with limited concurrency",
		},
		{
			name:         "medium_batch_high_concurrency",
			itemCount:    20,
			concurrency:  10,
			expectedTime: 2 * time.Second,
			description: "Medium batch with high concurrency",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			serialCodes := make([]string, tc.itemCount)
			for i := 0; i < tc.itemCount; i++ {
				serialCodes[i] = fmt.Sprintf("@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{%02d}", i%100)
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
					ReportProgress: true,
					Timeout:        30000,
				},
			}

			start := time.Now()
			response, err := performBatchDecodeRequest(t, baseURL, request)
			duration := time.Since(start)

			require.NoError(t, err)
			require.NotNil(t, response)
			assert.True(t, response.Success)

			// Performance validation
			t.Logf("Performance test '%s': %d items, concurrency %d, duration %v",
				tc.name, tc.itemCount, tc.concurrency, duration)

			assert.Less(t, duration, tc.expectedTime, "Batch processing should meet performance targets")
			assert.Less(t, duration/int64(tc.itemCount), 500*time.Millisecond, "Per-item processing should be efficient")

			// Verify all items were processed
			require.NotNil(t, response.Statistics)
			assert.Equal(t, tc.itemCount, response.Statistics.Total)
		})
	}
}

// TestBatchWorkflowWithCancellation tests cancellation scenarios
func TestBatchWorkflowWithCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cancellation test in short mode")
	}

	t.Skip("Cancellation functionality not yet implemented")

	// TODO: Implement cancellation test once cancellation is supported
	// Test cases to implement:
	// 1. User-initiated batch cancellation
	// 2. Timeout-based cancellation
	// 3. Cleanup of cancelled operations
	// 4. Partial result return
}

// Helper functions

// setupTestServer creates a test API server
func setupTestServer(t *testing.T) *api.Server {
	// Find available port
	port := getAvailablePort(t)
	t.Logf("Starting test server on port %d", port)

	// Create server with test configuration
	srv := api.NewServerWithConfig(&api.ServerConfig{
		Port: port,
		Mode: "test",
	})

	// Start server in background
	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			t.Logf("Server error: %v", err)
		}
	}()

	return srv
}

// getAvailablePort finds an available port for testing
func getAvailablePort(t *testing.T) int {
	// Simple port selection - in real implementation you'd use dynamic port allocation
	return 18080 + len(t.Name()) // Use name to vary port between tests
}

// waitForServerReady waits for the server to be ready
func waitForServerReady(url string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	client := &http.Client{Timeout: 1 * time.Second}

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("server not ready within timeout")
		default:
			resp, err := client.Get(url)
			if err == nil && resp.StatusCode == http.StatusOK {
				resp.Body.Close()
				return nil
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// performBatchDecodeRequest makes a HTTP batch decode request
func performBatchDecodeRequest(t *testing.T, baseURL string, request *models.BatchDecodeRequest) (*models.BatchDecodeResponse, error) {
	reqBody, err := json.Marshal(request)
	require.NoError(t, err)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(baseURL+"/api/v1/batch/decode", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response models.BatchDecodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// countFailedResults counts how many results have errors
func countFailedResults(results []*models.BatchDecodeResult) int {
	count := 0
	for _, result := range results {
		if result.Error != nil {
			count++
		}
	}
	return count
}

// isValidSerialCode performs basic validation of serial code format
func isValidSerialCode(code string) bool {
	if len(code) < 3 {
		return false
	}
	return code[0] == '@' && code[1] == 'U'
}

// TestConcurrentBatchProcessing tests concurrent batch processing
func TestConcurrentBatchProcessing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent batch test in short mode")
	}

	baseURL := fmt.Sprintf("http://localhost:%d", getAvailablePort(t))
	srv := setupTestServer(t)
	defer srv.Close()

	err := waitForServerReady(baseURL+"/health", 5*time.Second)
	require.NoError(t, err)

	// Test concurrent batch requests
	numRequests := 5
	itemsPerRequest := 10
	var wg sync.WaitGroup
	results := make(chan error, numRequests)

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
				},
			}

			_, err := performBatchDecodeRequest(t, baseURL, request)
			results <- err
		}(i)
	}

	// Wait for all requests to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	errorCount := 0
	for err := range results {
		if err != nil {
			errorCount++
			t.Logf("Concurrent request error: %v", err)
		}
	}

	// Assert that most requests succeeded (allow for some errors in concurrent testing)
	successRate := float64(numRequests-errorCount) / float64(numRequests)
	t.Logf("Concurrent processing success rate: %.2f", successRate)
	assert.Greater(t, successRate, 0.8, "Most concurrent requests should succeed")
}

// TestBatchWorkflowPersistence tests persistent batch processing
func TestBatchWorkflowPersistence(t *testing.T) {
	t.Skip("Persistence functionality not yet implemented")

	// TODO: Implement persistence test once persistent batch processing is supported
	// Test cases to implement:
	// 1. Batch job persistence across server restarts
	// 2. Progress tracking for long-running batches
	// 3. Result persistence and retrieval
}