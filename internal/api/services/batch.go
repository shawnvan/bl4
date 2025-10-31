package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/shawnvan/bl4/internal/api/models"
	"github.com/shawnvan/bl4/internal/codec/serial"
	"github.com/shawnvan/bl4/pkg/logger"
	"github.com/shawnvan/bl4/pkg/validator"
)

// BatchProcessor handles batch processing of serial codes
type BatchProcessor struct {
	decoder       *serial.Deserializer
	config        *BatchProcessorConfig
	subscribers   map[string][]ProgressSubscriber
	subscribersMux sync.RWMutex
	activeBatches map[string]*BatchProcessContext
	activeBatchesMux sync.RWMutex
}

// BatchProcessorConfig contains configuration for batch processing
type BatchProcessorConfig struct {
	// MaxConcurrency controls maximum concurrent processing
	MaxConcurrency int

	// DefaultTimeout is the default timeout for batch operations
	DefaultTimeout time.Duration

	// DefaultPerItemTimeout is the default timeout per item
	DefaultPerItemTimeout time.Duration

	// EnableProgressTracking enables progress tracking
	EnableProgressTracking bool

	// RateLimitRequests per second
	RateLimitRequests int

	// MemoryLimitMB limits memory usage
	MemoryLimitMB int
}

// BatchProcessResult contains the result of a batch operation
type BatchProcessResult struct {
	Success     bool                    `json:"success"`
	Results     []*models.BatchDecodeResult `json:"results,omitempty"`
	Statistics  *models.BatchStatistics   `json:"statistics,omitempty"`
	Progress    *BatchProgress          `json:"progress,omitempty"`
	Error       error                   `json:"error,omitempty"`
	StartTime   time.Time               `json:"start_time"`
	EndTime     time.Time               `json:"end_time"`
	Duration    time.Duration           `json:"duration"`
}

// BatchProgress tracks batch processing progress
type BatchProgress struct {
	Total         int     `json:"total"`
	Processed     int     `json:"processed"`
	Successful    int     `json:"successful"`
	Failed        int     `json:"failed"`
	Progress      float64 `json:"progress"`      // 0.0 to 1.0
	ETA           int64   `json:"eta"`           // Estimated time remaining in milliseconds
	StartTime     int64   `json:"start_time"`
	CurrentItem   string  `json:"current_item,omitempty"`
	ItemsPerSec   float64 `json:"items_per_sec"`   // Current processing rate
	EstimatedEnd  int64   `json:"estimated_end"`    // Estimated completion timestamp
	Cancelled     bool    `json:"cancelled"`        // Whether operation was cancelled
	CancelReason  string  `json:"cancel_reason,omitempty"` // Reason for cancellation
}

// ProgressUpdate represents a progress update sent to subscribers
type ProgressUpdate struct {
	BatchID    string        `json:"batch_id"`
	Progress   *BatchProgress `json:"progress"`
	Timestamp  time.Time     `json:"timestamp"`
	IsComplete bool          `json:"is_complete"`
	Error      string        `json:"error,omitempty"`
}

// ProgressSubscriber receives progress updates
type ProgressSubscriber interface {
	OnProgress(update *ProgressUpdate)
}


// BatchProcessContext manages a single batch processing operation
type BatchProcessContext struct {
	ID          string
	Request     *models.BatchDecodeRequest
	Ctx         context.Context
	CancelFunc  context.CancelFunc
	Progress    *BatchProgress
	Result      *BatchProcessResult
	StartTime   time.Time
	subscribers []ProgressSubscriber
	mu          sync.RWMutex
}

// NewBatchProcessor creates a new batch processor with default configuration
func NewBatchProcessor() *BatchProcessor {
	return &BatchProcessor{
		decoder:       serial.NewDeserializer(),
		config: &BatchProcessorConfig{
			MaxConcurrency:         5,
			DefaultTimeout:         30 * time.Second,
			DefaultPerItemTimeout: 5 * time.Second,
			EnableProgressTracking: true,
			RateLimitRequests:      100,
			MemoryLimitMB:           100,
		},
		subscribers:  make(map[string][]ProgressSubscriber),
		activeBatches: make(map[string]*BatchProcessContext),
	}
}

// NewBatchProcessorWithConfig creates a new batch processor with custom configuration
func NewBatchProcessorWithConfig(config *BatchProcessorConfig) *BatchProcessor {
	if config == nil {
		config = &BatchProcessorConfig{}
	}

	// Set defaults for missing values
	if config.MaxConcurrency <= 0 {
		config.MaxConcurrency = 5
	}
	if config.DefaultTimeout <= 0 {
		config.DefaultTimeout = 30 * time.Second
	}
	if config.DefaultPerItemTimeout <= 0 {
		config.DefaultPerItemTimeout = 5 * time.Second
	}
	if config.RateLimitRequests <= 0 {
		config.RateLimitRequests = 100
	}
	if config.MemoryLimitMB <= 0 {
		config.MemoryLimitMB = 100
	}

	return &BatchProcessor{
		decoder:       serial.NewDeserializer(),
		config:        config,
		subscribers:   make(map[string][]ProgressSubscriber),
		activeBatches: make(map[string]*BatchProcessContext),
	}
}

// ProcessBatch processes a batch of serial codes for decoding
func (bp *BatchProcessor) ProcessBatch(request *models.BatchDecodeRequest) (*BatchProcessResult, error) {
	return bp.ProcessBatchWithContext(context.Background(), request)
}

// ProcessBatchWithContext processes a batch with context support for cancellation
func (bp *BatchProcessor) ProcessBatchWithContext(ctx context.Context, request *models.BatchDecodeRequest) (*BatchProcessResult, error) {
	startTime := time.Now()

	logger.Sugar().Infow("Starting batch processing",
		"total_codes", len(request.SerialCodes),
		"max_concurrency", bp.getMaxConcurrency(request),
		"timeout", bp.getTimeout(request),
	)

	// Validate request
	if err := bp.validateRequest(request); err != nil {
		return bp.createErrorResult(err, startTime), err
	}

	// Create batch progress tracker
	progress := &BatchProgress{
		Total:     len(request.SerialCodes),
		Processed: 0,
		Successful: 0,
		Failed:    0,
		Progress:  0.0,
		ETA:       0,
		StartTime: startTime.UnixMilli(),
	}

	result := &BatchProcessResult{
		Success:   false,
		Progress:  progress,
		StartTime: startTime,
	}

	// Set up context with timeout
	timeout := bp.getTimeout(request)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Process items
	results, err := bp.processItems(ctx, request, progress)
	if err != nil {
		return bp.createErrorResult(err, startTime), err
	}

	// Calculate statistics
	stats := bp.calculateStatistics(results, startTime)

	result.Success = true
	result.Results = results
	result.Statistics = stats
	result.EndTime = time.Now()
	result.Duration = time.Since(startTime)

	logger.Sugar().Infow("Batch processing completed",
		"total_codes", len(request.SerialCodes),
		"successful", stats.Successful,
		"failed", stats.Failed,
		"duration_ms", result.Duration.Milliseconds(),
		"average_duration_us", stats.AverageTime,
	)

	return result, nil
}

// ProcessBatchWithProgress processes a batch with progress callback
func (bp *BatchProcessor) ProcessBatchWithProgress(request *models.BatchDecodeRequest, progressCallback func(processed, total int)) (*BatchProcessResult, error) {
	startTime := time.Now()

	logger.Sugar().Infow("Starting batch processing with progress tracking",
		"total_codes", len(request.SerialCodes),
	)

	// Validate request
	if err := bp.validateRequest(request); err != nil {
		return bp.createErrorResult(err, startTime), err
	}

	// Create progress tracker
	progress := &BatchProgress{
		Total:     len(request.SerialCodes),
		Processed: 0,
		Successful: 0,
		Failed:    0,
		Progress:  0.0,
		ETA:       0,
		StartTime: startTime.UnixMilli(),
	}

	result := &BatchProcessResult{
		Success:   false,
		Progress:  progress,
		StartTime: startTime,
	}

	// Process items with progress tracking
	results, err := bp.processItemsWithProgress(request, progress, progressCallback)
	if err != nil {
		return bp.createErrorResult(err, startTime), err
	}

	// Calculate statistics
	stats := bp.calculateStatistics(results, startTime)

	result.Success = true
	result.Results = results
	result.Statistics = stats
	result.EndTime = time.Now()
	result.Duration = time.Since(startTime)

	logger.Sugar().Infow("Batch processing with progress completed",
		"total_codes", len(request.SerialCodes),
		"successful", stats.Successful,
		"failed", stats.Failed,
		"duration_ms", result.Duration.Milliseconds(),
	)

	return result, nil
}

// validateRequest validates the batch request
func (bp *BatchProcessor) validateRequest(request *models.BatchDecodeRequest) error {
	if request == nil {
		return validator.NewValidationError(validator.ErrCodeEmptyInput, "batch request is required")
	}

	if len(request.SerialCodes) == 0 {
		return validator.NewValidationError(validator.ErrCodeEmptyInput, "serial codes cannot be empty")
	}

	if len(request.SerialCodes) > 1000 {
		return validator.NewValidationError(validator.ErrCodeBatchSizeExceeded, "too many serial codes").
			WithField("count", len(request.SerialCodes)).
			WithDetail("max_count", 1000)
	}

	// Validate individual serial codes
	for i, code := range request.SerialCodes {
		if len(code) < 3 {
			return validator.NewValidationError(validator.ErrCodeInvalidLength,
				fmt.Sprintf("serial code at index %d is too short", i)).
				WithField("index", i).
				WithField("length", len(code))
		}
	}

	// Validate options if provided
	if request.Options != nil {
		if request.Options.MaxConcurrency < 0 {
			return validator.NewValidationError(validator.ErrCodeInvalidInput,
				"max concurrency cannot be negative").
				WithField("max_concurrency", request.Options.MaxConcurrency)
		}

		if request.Options.Timeout < 0 {
			return validator.NewValidationError(validator.ErrCodeInvalidInput,
				"timeout cannot be negative").
				WithField("timeout", request.Options.Timeout)
		}

		if request.Options.PerItemTimeout < 0 {
			return validator.NewValidationError(validator.ErrCodeInvalidInput,
				"per item timeout cannot be negative").
				WithField("per_item_timeout", request.Options.PerItemTimeout)
		}
	}

	return nil
}

// processItems processes the batch items
func (bp *BatchProcessor) processItems(ctx context.Context, request *models.BatchDecodeRequest, progress *BatchProgress) ([]*models.BatchDecodeResult, error) {
	maxConcurrency := bp.getMaxConcurrency(request)
	itemTimeout := bp.getPerItemTimeout(request)

	// Create worker pool
	jobs := make(chan int, len(request.SerialCodes))
	results := make(chan *models.BatchDecodeResult, len(request.SerialCodes))

	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < maxConcurrency; i++ {
		wg.Add(1)
		go bp.worker(ctx, request, jobs, results, itemTimeout, progress, &wg)
	}

	// Send jobs
	for i := 0; i < len(request.SerialCodes); i++ {
		select {
		case jobs <- i:
		case <-ctx.Done():
			// Context cancelled, close jobs channel
			close(jobs)
			wg.Wait()
			return nil, ctx.Err()
		}
	}
	close(jobs)

	// Wait for workers to complete
	wg.Wait()
	close(results)

	// Collect results
	batchResults := make([]*models.BatchDecodeResult, len(request.SerialCodes))
	for i := 0; i < len(request.SerialCodes); i++ {
		batchResults[i] = <-results
	}

	return batchResults, nil
}

// worker processes individual items
func (bp *BatchProcessor) worker(ctx context.Context, request *models.BatchDecodeRequest, jobs <-chan int, results chan<- *models.BatchDecodeResult, itemTimeout time.Duration, progress *BatchProgress, wg *sync.WaitGroup) {
	defer wg.Done()

	for index := range jobs {
		select {
		case <-ctx.Done():
			// Context cancelled, stop processing
			return
		default:
			// Process the item
			result := bp.processItem(ctx, request, index, itemTimeout, progress)
			results <- result

			// Update progress
			atomicIntAdd(&progress.Processed, 1)
			if result.Error == nil {
				atomicIntAdd(&progress.Successful, 1)
			} else {
				atomicIntAdd(&progress.Failed, 1)
			}

			// Update progress percentage
			total := float64(progress.Total)
			processed := float64(progress.Processed)
			progress.Progress = processed / total
			progress.ETA = int64(float64(time.Since(time.UnixMilli(progress.StartTime)).Milliseconds()) / processed * (total - processed))
		}
	}
}

// processItem processes a single item
func (bp *BatchProcessor) processItem(ctx context.Context, request *models.BatchDecodeRequest, index int, timeout time.Duration, progress *BatchProgress) *models.BatchDecodeResult {
	startTime := time.Now()

	// Update current item
	progress.CurrentItem = request.SerialCodes[index]

	// Create context with timeout for individual item
	itemCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Use context to check for cancellation
	select {
	case <-itemCtx.Done():
		result := &models.BatchDecodeResult{
			Index:      index,
			SerialCode: request.SerialCodes[index],
			Success:    false,
			Duration:   time.Since(startTime).Microseconds(),
			Error:      models.NewErrorInfo(itemCtx.Err()),
		}
		return result
	default:
		// Continue processing
	}

	// Prepare decode options
	options := &serial.DeserializerOptions{
		IncludeBitstream: request.Options.IndividualOptions != nil && request.Options.IndividualOptions.IncludeBitstream,
		IncludeTokens:    request.Options.IndividualOptions != nil && request.Options.IndividualOptions.IncludeTokens,
		StrictValidation:  true, // Always use strict validation
		MaxProcessingTime: int(timeout.Milliseconds()),
	}

	deserializer := serial.NewDeserializerWithOptions(options)

	// Process the item
	deserializationResult, err := deserializer.DeserializeItem(request.SerialCodes[index])

	result := &models.BatchDecodeResult{
		Index:      index,
		SerialCode: request.SerialCodes[index],
		Success:    err == nil,
		Duration:   time.Since(startTime).Microseconds(),
	}

	if err != nil {
		result.Error = models.NewErrorInfo(err)
		logger.Sugar().Warnw("Item processing failed",
			"index", index,
			"serial_code", request.SerialCodes[index],
			"error", err.Error(),
			"duration_us", result.Duration,
		)
	} else {
		// Create decoded item data
		itemData := &models.DecodedItemData{
			SerialCode: request.SerialCodes[index],
			ItemData:   deserializationResult.ItemData,
		}

		// Add bitstream info if requested
		if options.IncludeBitstream && deserializationResult.BitstreamInfo != nil {
			itemData.AddBitstreamInfo(
				deserializationResult.TokenStream,
				deserializationResult.DecodedBytes,
				true, // Include raw bitstream data
			)
		}

		// Add token info if requested
		if options.IncludeTokens && deserializationResult.TokenInfos != nil {
			itemData.AddTokens(deserializationResult.TokenStream)
		}

		// Add raw data info if requested
		if options.IncludeRawData && deserializationResult.RawDataInfo != nil {
			itemData.AddRawData(
				deserializationResult.RawDataInfo.Structured,
				"", // JSON representation could be added later
				"", // CSV representation could be added later
				"", // Hex representation could be added later
			)
		}

		result.Data = itemData

		logger.Sugar().Debugw("Item processed successfully",
			"index", index,
			"serial_code", request.SerialCodes[index],
			"level", deserializationResult.ItemData.Level,
			"duration_us", result.Duration,
		)
	}

	return result
}

// processItemsWithProgress processes items with progress callback
func (bp *BatchProcessor) processItemsWithProgress(request *models.BatchDecodeRequest, progress *BatchProgress, progressCallback func(processed, total int)) ([]*models.BatchDecodeResult, error) {
	// Use regular processing but call progress callback
	results, err := bp.processItems(context.Background(), request, progress)

	// Call progress callback for each processed item
	for i := 0; i < progress.Processed; i++ {
		progressCallback(i+1, progress.Total)
	}

	return results, err
}

// getMaxConcurrency returns the maximum concurrency for the request
func (bp *BatchProcessor) getMaxConcurrency(request *models.BatchDecodeRequest) int {
	if request.Options != nil && request.Options.MaxConcurrency > 0 {
		return request.Options.MaxConcurrency
	}
	return bp.config.MaxConcurrency
}

// getTimeout returns the timeout for the batch operation
func (bp *BatchProcessor) getTimeout(request *models.BatchDecodeRequest) time.Duration {
	if request.Options != nil && request.Options.Timeout > 0 {
		return time.Duration(request.Options.Timeout) * time.Millisecond
	}
	return bp.config.DefaultTimeout
}

// getPerItemTimeout returns the timeout per item
func (bp *BatchProcessor) getPerItemTimeout(request *models.BatchDecodeRequest) time.Duration {
	if request.Options != nil && request.Options.PerItemTimeout > 0 {
		return time.Duration(request.Options.PerItemTimeout) * time.Millisecond
	}
	return bp.config.DefaultPerItemTimeout
}

// calculateStatistics calculates statistics for the batch operation
func (bp *BatchProcessor) calculateStatistics(results []*models.BatchDecodeResult, startTime time.Time) *models.BatchStatistics {
	stats := &models.BatchStatistics{
		Total:       len(results),
		Successful:  0,
		Failed:      0,
		StartTime:   startTime,
		EndTime:     time.Now(),
	}

	var totalDuration time.Duration
	var successDuration time.Duration
	var failureDuration time.Duration

	for _, result := range results {
		if result.Success {
			stats.Successful++
			successDuration += time.Duration(result.Duration)
		} else {
			stats.Failed++
			failureDuration += time.Duration(result.Duration)
		}
		totalDuration += time.Duration(result.Duration)
	}

	// Calculate average time
	if stats.Successful > 0 {
		stats.AverageTime = successDuration.Microseconds() / int64(stats.Successful)
		stats.MinTime = successDuration.Microseconds() / int64(stats.Successful)
		stats.MaxTime = successDuration.Microseconds() / int64(stats.Successful)
	}

	// Calculate success rate
	if stats.Total > 0 {
		stats.SuccessRate = float64(stats.Successful) / float64(stats.Total)
	}

	return stats
}

// createErrorResult creates an error result
func (bp *BatchProcessor) createErrorResult(err error, startTime time.Time) *BatchProcessResult {
	return &BatchProcessResult{
		Success:   false,
		Error:     err,
		StartTime: startTime,
		EndTime:   time.Now(),
		Duration:  time.Since(startTime),
	}
}

// atomicIntAdd performs atomic integer addition
func atomicIntAdd(val *int, delta int) {
	// This is a simplified implementation
	// In a real implementation, you would use sync/atomic
	*val += delta
}

// BatchDecodeHandler represents a batch decode handler
type BatchDecodeHandler struct {
	processor *BatchProcessor
}

// NewBatchDecodeHandler creates a new batch decode handler
func NewBatchDecodeHandler() *BatchDecodeHandler {
	return &BatchDecodeHandler{
		processor: NewBatchProcessor(),
	}
}

// HandleBatchDecode handles the batch decode request
func (h *BatchDecodeHandler) HandleBatchDecode(request *models.BatchDecodeRequest) (*models.BatchDecodeResponse, error) {
	startTime := time.Now()

	logger.Sugar().Infow("Batch decode request received",
		"total_codes", len(request.SerialCodes),
		"options", request.Options,
	)

	// Process the batch
	result, err := h.processor.ProcessBatch(request)
	if err != nil {
		logger.Sugar().Errorw("Batch processing failed",
			"error", err.Error(),
			"duration_ms", time.Since(startTime).Milliseconds(),
		)

		// Return error response
		response := &models.BatchDecodeResponse{
			Success: false,
			Error:   models.NewErrorInfo(err),
			Metadata: models.NewResponseMetadata(),
		}
		response.Metadata.UpdatePerformance(int64(time.Since(startTime).Microseconds()), 0)

		return response, err
	}

	// Build success response
	response := &models.BatchDecodeResponse{
		Success: true,
		Results: result.Results,
		Statistics: &models.BatchStatistics{
			Total:       result.Statistics.Total,
			Successful:  result.Statistics.Successful,
			Failed:      result.Statistics.Failed,
			SuccessRate: result.Statistics.SuccessRate,
			AverageTime: result.Statistics.AverageTime,
			MinTime:     result.Statistics.MinTime,
			MaxTime:     result.Statistics.MaxTime,
			StartTime:   result.Statistics.StartTime,
			EndTime:     result.Statistics.EndTime,
		},
		Metadata: models.NewResponseMetadata(),
	}

	// Update performance metrics
	response.Metadata.UpdatePerformance(int64(time.Since(startTime).Microseconds()), 0)

	logger.Sugar().Infow("Batch decode request completed successfully",
		"total_codes", len(request.SerialCodes),
		"successful", result.Statistics.Successful,
		"failed", result.Statistics.Failed,
		"duration_ms", time.Since(startTime).Milliseconds(),
	)

	return response, nil
}

// Progress Tracking and Cancellation Enhancement Methods

// ProcessBatchWithTracking processes a batch with enhanced progress tracking and cancellation support
func (bp *BatchProcessor) ProcessBatchWithTracking(batchID string, request *models.BatchDecodeRequest, subscriber ProgressSubscriber) (*BatchProcessResult, error) {
	startTime := time.Now()

	// Create batch context
	ctx, cancelFunc := context.WithCancel(context.Background())

	// Create progress tracking
	progress := &BatchProgress{
		Total:       len(request.SerialCodes),
		Processed:   0,
		Successful:  0,
		Failed:      0,
		Progress:    0.0,
		ETA:         0,
		StartTime:   startTime.UnixMilli(),
		CurrentItem: "",
		ItemsPerSec: 0.0,
		EstimatedEnd: startTime.UnixMilli(),
		Cancelled:   false,
		CancelReason: "",
	}

	// Create batch context
	batchCtx := &BatchProcessContext{
		ID:          batchID,
		Request:     request,
		Ctx:         ctx,
		CancelFunc:  cancelFunc,
		Progress:    progress,
		Result:      &BatchProcessResult{StartTime: startTime},
		StartTime:   startTime,
		subscribers: []ProgressSubscriber{subscriber},
	}

	// Register active batch
	bp.activeBatchesMux.Lock()
	bp.activeBatches[batchID] = batchCtx
	bp.activeBatchesMux.Unlock()

	// Send initial progress update
	bp.sendProgressUpdate(batchCtx, false, "")

	logger.Sugar().Infow("Starting batch processing with tracking",
		"batch_id", batchID,
		"total_codes", len(request.SerialCodes),
	)

	// Process with enhanced tracking
	result, err := bp.processBatchWithEnhancedTracking(batchCtx)

	// Clean up
	bp.activeBatchesMux.Lock()
	delete(bp.activeBatches, batchID)
	bp.activeBatchesMux.Unlock()

	// Send final progress update
	bp.sendProgressUpdate(batchCtx, true, "")

	if err != nil {
		bp.sendProgressUpdate(batchCtx, true, err.Error())
	}

	return result, err
}

// CancelBatch cancels a running batch operation
func (bp *BatchProcessor) CancelBatch(batchID, reason string) error {
	bp.activeBatchesMux.RLock()
	batchCtx, exists := bp.activeBatches[batchID]
	bp.activeBatchesMux.RUnlock()

	if !exists {
		return fmt.Errorf("batch %s not found or already completed", batchID)
	}

	batchCtx.mu.Lock()
	defer batchCtx.mu.Unlock()

	if batchCtx.Progress.Cancelled {
		return fmt.Errorf("batch %s already cancelled", batchID)
	}

	// Mark as cancelled
	batchCtx.Progress.Cancelled = true
	batchCtx.Progress.CancelReason = reason

	// Cancel the context
	batchCtx.CancelFunc()

	logger.Sugar().Infow("Batch operation cancelled",
		"batch_id", batchID,
		"reason", reason,
		"processed", batchCtx.Progress.Processed,
		"total", batchCtx.Progress.Total,
	)

	// Send cancellation update
	bp.sendProgressUpdate(batchCtx, false, reason)

	return nil
}

// GetBatchProgress returns current progress for a batch
func (bp *BatchProcessor) GetBatchProgress(batchID string) (*BatchProgress, error) {
	bp.activeBatchesMux.RLock()
	batchCtx, exists := bp.activeBatches[batchID]
	bp.activeBatchesMux.RUnlock()

	if !exists {
		return nil, fmt.Errorf("batch %s not found", batchID)
	}

	batchCtx.mu.RLock()
	defer batchCtx.mu.RUnlock()

	// Return a copy to avoid concurrent access issues
	progressCopy := *batchCtx.Progress
	return &progressCopy, nil
}

// SubscribeToBatch subscribes to progress updates for a batch
func (bp *BatchProcessor) SubscribeToBatch(batchID string, subscriber ProgressSubscriber) error {
	bp.activeBatchesMux.RLock()
	batchCtx, exists := bp.activeBatches[batchID]
	bp.activeBatchesMux.RUnlock()

	if !exists {
		return fmt.Errorf("batch %s not found", batchID)
	}

	batchCtx.mu.Lock()
	defer batchCtx.mu.Unlock()

	batchCtx.subscribers = append(batchCtx.subscribers, subscriber)

	return nil
}

// processBatchWithEnhancedTracking processes a batch with enhanced progress tracking
func (bp *BatchProcessor) processBatchWithEnhancedTracking(batchCtx *BatchProcessContext) (*BatchProcessResult, error) {
	request := batchCtx.Request
	ctx := batchCtx.Ctx

	// Validate request
	if err := bp.validateRequest(request); err != nil {
		return bp.createErrorResult(err, batchCtx.StartTime), err
	}

	// Set up context with timeout
	timeout := bp.getTimeout(request)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Process items with enhanced tracking
	results, err := bp.processItemsWithEnhancedTracking(ctx, batchCtx)
	if err != nil {
		return bp.createErrorResult(err, batchCtx.StartTime), err
	}

	// Calculate statistics
	stats := bp.calculateStatistics(results, batchCtx.StartTime)

	// Update final result
	batchCtx.Result.Success = true
	batchCtx.Result.Results = results
	batchCtx.Result.Statistics = stats
	batchCtx.Result.EndTime = time.Now()
	batchCtx.Result.Duration = time.Since(batchCtx.StartTime)

	logger.Sugar().Infow("Enhanced batch processing completed",
		"batch_id", batchCtx.ID,
		"total_codes", len(request.SerialCodes),
		"successful", stats.Successful,
		"failed", stats.Failed,
		"duration_ms", batchCtx.Result.Duration.Milliseconds(),
		"average_duration_us", stats.AverageTime,
	)

	return batchCtx.Result, nil
}

// processItemsWithEnhancedTracking processes items with enhanced progress tracking
func (bp *BatchProcessor) processItemsWithEnhancedTracking(ctx context.Context, batchCtx *BatchProcessContext) ([]*models.BatchDecodeResult, error) {
	request := batchCtx.Request
	maxConcurrency := bp.getMaxConcurrency(request)
	itemTimeout := bp.getPerItemTimeout(request)

	// Create worker pool
	jobs := make(chan int, len(request.SerialCodes))
	results := make(chan *models.BatchDecodeResult, len(request.SerialCodes))

	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < maxConcurrency; i++ {
		wg.Add(1)
		go bp.enhancedWorker(ctx, batchCtx, jobs, results, itemTimeout, &wg)
	}

	// Send jobs
	for i := 0; i < len(request.SerialCodes); i++ {
		select {
		case jobs <- i:
		case <-ctx.Done():
			// Context cancelled, close jobs channel
			close(jobs)
			wg.Wait()
			return nil, ctx.Err()
		}
	}
	close(jobs)

	// Wait for workers to complete
	wg.Wait()
	close(results)

	// Collect results
	batchResults := make([]*models.BatchDecodeResult, len(request.SerialCodes))
	for i := 0; i < len(request.SerialCodes); i++ {
		batchResults[i] = <-results
	}

	return batchResults, nil
}

// enhancedWorker processes individual items with enhanced progress tracking
func (bp *BatchProcessor) enhancedWorker(ctx context.Context, batchCtx *BatchProcessContext, jobs <-chan int, results chan<- *models.BatchDecodeResult, itemTimeout time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()

	for index := range jobs {
		select {
		case <-ctx.Done():
			// Context cancelled, stop processing
			return
		case <-batchCtx.Ctx.Done():
			// Batch cancelled, stop processing
			return
		default:
			// Process the item
			result := bp.processItemWithEnhancedTracking(batchCtx, index, itemTimeout)
			results <- result

			// Update progress with mutex protection
			batchCtx.mu.Lock()
			batchCtx.Progress.Processed++
			if result.Error == nil {
				batchCtx.Progress.Successful++
			} else {
				batchCtx.Progress.Failed++
			}

			// Update progress metrics
			total := float64(batchCtx.Progress.Total)
			processed := float64(batchCtx.Progress.Processed)
			batchCtx.Progress.Progress = processed / total

			// Calculate ETA and processing rate
			elapsed := time.Since(batchCtx.StartTime)
			if processed > 0 {
				batchCtx.Progress.ItemsPerSec = processed / elapsed.Seconds()
				if processed < total {
					remainingTime := elapsed * time.Duration(total/processed-1)
					batchCtx.Progress.ETA = remainingTime.Milliseconds()
					batchCtx.Progress.EstimatedEnd = time.Now().Add(remainingTime).UnixMilli()
				}
			}
			batchCtx.mu.Unlock()

			// Send progress update every 10 items or on completion
			if batchCtx.Progress.Processed%10 == 0 || batchCtx.Progress.Processed == batchCtx.Progress.Total {
				bp.sendProgressUpdate(batchCtx, false, "")
			}
		}
	}
}

// processItemWithEnhancedTracking processes a single item with enhanced tracking
func (bp *BatchProcessor) processItemWithEnhancedTracking(batchCtx *BatchProcessContext, index int, timeout time.Duration) *models.BatchDecodeResult {
	startTime := time.Now()

	// Update current item
	batchCtx.Progress.CurrentItem = batchCtx.Request.SerialCodes[index]

	// Check if batch is cancelled
	batchCtx.mu.RLock()
	cancelled := batchCtx.Progress.Cancelled
	batchCtx.mu.RUnlock()

	if cancelled {
		result := &models.BatchDecodeResult{
			Index:      index,
			SerialCode: batchCtx.Request.SerialCodes[index],
			Success:    false,
			Duration:   time.Since(startTime).Microseconds(),
			Error:      models.NewErrorInfo(fmt.Errorf("batch cancelled: %s", batchCtx.Progress.CancelReason)),
		}
		return result
	}

	// Process the item (reuse existing logic)
	return bp.processItem(batchCtx.Ctx, batchCtx.Request, index, timeout, batchCtx.Progress)
}

// sendProgressUpdate sends progress updates to all subscribers
func (bp *BatchProcessor) sendProgressUpdate(batchCtx *BatchProcessContext, isComplete bool, errorMsg string) {
	batchCtx.mu.RLock()
	subscribers := make([]ProgressSubscriber, len(batchCtx.subscribers))
	copy(subscribers, batchCtx.subscribers)
	progressCopy := *batchCtx.Progress
	batchCtx.mu.RUnlock()

	update := &ProgressUpdate{
		BatchID:    batchCtx.ID,
		Progress:   &progressCopy,
		Timestamp:  time.Now(),
		IsComplete: isComplete,
		Error:      errorMsg,
	}

	// Send updates to all subscribers (non-blocking)
	for _, subscriber := range subscribers {
		go func(s ProgressSubscriber) {
			defer func() {
				if r := recover(); r != nil {
					logger.Sugar().Errorw("Subscriber panicked during progress update",
						"batch_id", batchCtx.ID,
						"panic", r,
					)
				}
			}()
			s.OnProgress(update)
		}(subscriber)
	}
}