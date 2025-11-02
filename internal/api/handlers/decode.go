package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shawnvan/bl4/internal/api/formatter"
	"github.com/shawnvan/bl4/internal/api/models"
	"github.com/shawnvan/bl4/internal/api/services"
	"github.com/shawnvan/bl4/internal/codec/serial"
	"github.com/shawnvan/bl4/pkg/logger"
	"github.com/shawnvan/bl4/pkg/validator"
)

// DecodeHandler handles item code decoding requests
type DecodeHandler struct{}

// NewDecodeHandler creates a new decode handler
func NewDecodeHandler() *DecodeHandler {
	return &DecodeHandler{}
}

// HandleDecode handles POST /api/v1/items/decode requests
func (h *DecodeHandler) HandleDecode(c *gin.Context) {
	startTime := time.Now()

	// Parse request
	var request models.DecodeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.handleError(c, validator.NewValidationError(validator.ErrCodeInvalidInput, "Invalid request format"))
		return
	}

	logger.Sugar().Infow("Decode request received",
		"serial_code", request.SerialCode,
		"options", request.Options,
	)

	// Create deserializer with options from request
	deserializerOptions := &serial.DeserializerOptions{
		IncludeBitstream: request.HasOption("include_bitstream"),
		IncludeTokens:    request.HasOption("include_tokens"),
		IncludeRawData:   request.HasOption("include_raw_data"),
		StrictValidation: true,
		MaxProcessingTime: request.GetTimeout(),
	}

	deserializer := serial.NewDeserializerWithOptions(deserializerOptions)

	// Perform deserialization
	deserializationResult, err := deserializer.DeserializeItem(request.SerialCode)
	if err != nil {
		logger.Sugar().Errorw("Deserialization failed",
			"serial_code", request.SerialCode,
			"error", err.Error(),
			"duration_ms", time.Since(startTime).Milliseconds(),
		)

		// Return error response
		response := &models.DecodeResponse{
			Success: false,
			Error:   models.NewErrorInfo(err),
			Metadata: models.NewResponseMetadata(),
		}
		response.Metadata.UpdatePerformance(int64(time.Since(startTime).Microseconds()), 0)

		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Build success response
	response := &models.DecodeResponse{
		Success: true,
		Data: &models.DecodedItemData{
			SerialCode: request.SerialCode,
			ItemData:   deserializationResult.ItemData,
		},
		Metadata: models.NewResponseMetadata(),
	}

	// Add reference format string
	if deserializationResult.TokenStream != nil {
		response.Data.ReferenceFormat = formatter.FormatAsCustomReference(deserializationResult.TokenStream)
	}

	// Add bitstream info if requested
	if deserializerOptions.IncludeBitstream && deserializationResult.BitstreamInfo != nil {
		response.Data.AddBitstreamInfo(
			deserializationResult.TokenStream,
			deserializationResult.DecodedBytes,
			true, // Include raw bitstream data
		)
	}

	// Add token info if requested
	if deserializerOptions.IncludeTokens && deserializationResult.TokenInfos != nil {
		response.Data.AddTokens(deserializationResult.TokenStream)
	}

	// Add raw data info if requested
	if deserializerOptions.IncludeRawData && deserializationResult.RawDataInfo != nil {
		response.Data.AddRawData(
			deserializationResult.RawDataInfo.Structured,
			"", // JSON representation could be added later
			"", // CSV representation could be added later
			"", // Hex representation could be added later
		)
	}

	// Update performance metrics
	response.Metadata.UpdatePerformance(
		int64(time.Since(startTime).Microseconds()),
		0, // Memory usage could be tracked later
	)

	logger.Sugar().Infow("Decode request completed successfully",
		"serial_code", request.SerialCode,
		"duration_ms", time.Since(startTime).Milliseconds(),
		"level", deserializationResult.ItemData.Level,
		"type", deserializationResult.ItemData.Type,
		"manufacturer", deserializationResult.ItemData.Manufacturer,
		"parts_count", len(deserializationResult.ItemData.Parts),
	)

	c.JSON(http.StatusOK, response)
}

// HandleHealthCheck handles GET /api/v1/items/decode/health requests
func (h *DecodeHandler) HandleHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "bl4-decode-handler",
		"version":   "1.0.0",
	})
}

// HandleStats handles GET /api/v1/items/decode/stats requests
func (h *DecodeHandler) HandleStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"handler_type": "decode",
		"status":       "active",
		"uptime":       "0s", // Would calculate actual uptime
		"requests_processed": 0, // Would track actual requests
	})
}

// handleError handles error responses
func (h *DecodeHandler) handleError(c *gin.Context, err error) {
	logger.Sugar().Errorw("Request error",
		"error", err,
		"path", c.Request.URL.Path,
		"method", c.Request.Method,
	)

	if ve, ok := err.(*validator.ValidationError); ok {
		c.JSON(ve.Code.HTTPStatus(), gin.H{
			"success": false,
			"error": gin.H{
				"code":    ve.Code.String(),
				"message": ve.Message,
				"details": ve.Details,
				"field":   ve.Field,
			},
		})
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "INTERNAL_SERVER_ERROR",
			"message": "An internal error occurred",
		},
	})
}

// BatchDecodeHandler handles batch decode requests
type BatchDecodeHandler struct{}

// NewBatchDecodeHandler creates a new batch decode handler
func NewBatchDecodeHandler() *BatchDecodeHandler {
	return &BatchDecodeHandler{}
}

// HandleBatchDecode handles POST /api/v1/items/batch/decode requests
func (h *BatchDecodeHandler) HandleBatchDecode(c *gin.Context) {
	startTime := time.Now()

	// Parse request
	var request models.BatchDecodeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.handleError(c, validator.NewValidationError(validator.ErrCodeInvalidInput, "Invalid batch request format"))
		return
	}

	logger.Sugar().Infow("Batch decode request received",
		"serial_codes_count", len(request.SerialCodes),
	"options", request.Options,
	)

	// Validate request
	if err := request.Validate(); err != nil {
		h.handleError(c, err)
		return
	}

	// Create batch handler with custom config if needed
	batchHandler := services.NewBatchDecodeHandler()

	// Process the batch
	batchResponse, err := batchHandler.HandleBatchDecode(&request)
	if err != nil {
		h.handleError(c, err)
		return
	}

	logger.Sugar().Infow("Batch decode request completed successfully",
		"serial_codes_count", len(request.SerialCodes),
		"duration_ms", time.Since(startTime).Milliseconds(),
		"successful_count", batchResponse.Statistics.Successful,
		"failed_count", batchResponse.Statistics.Failed,
	)

	c.JSON(http.StatusOK, batchResponse)
}

// HandleBatchDecodeEnhanced handles POST /api/v1/items/batch/decode with enhanced progress tracking
func (h *BatchDecodeHandler) HandleBatchDecodeEnhanced(c *gin.Context) {
	startTime := time.Now()

	// Generate batch ID
	batchID := uuid.New().String()

	// Get request ID from middleware (if available)
	requestID, exists := c.Get("request_id")
	if !exists {
		requestID = batchID
	}

	// Parse request
	var request models.BatchDecodeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.handleError(c, validator.NewValidationError(validator.ErrCodeInvalidInput, "Invalid batch request format"))
		return
	}

	logger.Sugar().Infow("Enhanced batch decode request received",
		"batch_id", batchID,
		"request_id", requestID,
		"serial_codes_count", len(request.SerialCodes),
		"options", request.Options,
	)

	// Validate request
	if err := request.Validate(); err != nil {
		h.handleError(c, err)
		return
	}

	// Create batch processor with custom config if needed
	batchHandler := services.NewBatchDecodeHandler()

	// Create progress subscriber that sends updates via WebSocket or Server-Sent Events
	progressSubscriber := &HTTPProgressSubscriber{
		Context:  c,
		BatchID:  batchID,
		StartTime: startTime,
	}

	// Check if client wants enhanced tracking (based on query parameter or header)
	wantEnhancedTracking := c.Query("progress") == "true" || c.GetHeader("X-Progress-Tracking") == "true"

	var result *models.BatchDecodeResponse
	var err error

	if wantEnhancedTracking {
		// Use enhanced batch processing with progress tracking
		batchProcessor := services.NewBatchProcessor()
		processResult, processErr := batchProcessor.ProcessBatchWithTracking(batchID, &request, progressSubscriber)
		if processErr != nil {
			h.handleError(c, processErr)
			return
		}

		// Convert to response format
		result = &models.BatchDecodeResponse{
			Success: true,
			Results: processResult.Results,
			Statistics: &models.BatchStatistics{
				Total:       processResult.Statistics.Total,
				Successful:  processResult.Statistics.Successful,
				Failed:      processResult.Statistics.Failed,
				SuccessRate: processResult.Statistics.SuccessRate,
				AverageTime: processResult.Statistics.AverageTime,
				MinTime:     processResult.Statistics.MinTime,
				MaxTime:     processResult.Statistics.MaxTime,
				StartTime:   processResult.Statistics.StartTime,
				EndTime:     processResult.Statistics.EndTime,
			},
			Metadata: models.NewResponseMetadata(),
		}
	} else {
		// Use standard batch processing
		result, err = batchHandler.HandleBatchDecode(&request)
		if err != nil {
			h.handleError(c, err)
			return
		}
	}

	// Update performance metrics
	result.Metadata.UpdatePerformance(int64(time.Since(startTime).Microseconds()), 0)

	// Add batch ID to response for tracking
	if result.Metadata == nil {
		result.Metadata = models.NewResponseMetadata()
	}
	if result.Metadata.Properties == nil {
		result.Metadata.Properties = make(map[string]interface{})
	}
	result.Metadata.Properties["batch_id"] = batchID
	result.Metadata.Properties["request_id"] = requestID

	logger.Sugar().Infow("Enhanced batch decode request completed successfully",
		"batch_id", batchID,
		"request_id", requestID,
		"serial_codes_count", len(request.SerialCodes),
		"successful_count", result.Statistics.Successful,
		"failed_count", result.Statistics.Failed,
		"duration_ms", time.Since(startTime).Milliseconds(),
	)

	// Set response headers for batch tracking
	c.Header("X-Batch-ID", batchID)
	c.Header("X-Request-ID", fmt.Sprintf("%v", requestID))

	c.JSON(http.StatusOK, result)
}

// HandleBatchCancel handles DELETE /api/v1/items/batch/:id/cancel requests
func (h *BatchDecodeHandler) HandleBatchCancel(c *gin.Context) {
	batchID := c.Param("id")
	if batchID == "" {
		h.handleError(c, validator.NewValidationError(validator.ErrCodeInvalidInput, "Batch ID is required"))
		return
	}

	// Get cancellation reason from request body or query parameter
	reason := c.Query("reason")
	if reason == "" {
		reason = "User requested cancellation"
	}

	// Create batch processor and cancel the batch
	batchProcessor := services.NewBatchProcessor()
	err := batchProcessor.CancelBatch(batchID, reason)

	if err != nil {
		h.handleError(c, err)
		return
	}

	logger.Sugar().Infow("Batch operation cancelled via API",
		"batch_id", batchID,
		"reason", reason,
		"client_ip", c.ClientIP(),
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Batch %s cancelled successfully", batchID),
		"batch_id": batchID,
		"reason": reason,
		"timestamp": time.Now().UTC(),
	})
}

// HandleBatchProgress handles GET /api/v1/items/batch/:id/progress requests
func (h *BatchDecodeHandler) HandleBatchProgress(c *gin.Context) {
	batchID := c.Param("id")
	if batchID == "" {
		h.handleError(c, validator.NewValidationError(validator.ErrCodeInvalidInput, "Batch ID is required"))
		return
	}

	// Get batch progress
	batchProcessor := services.NewBatchProcessor()
	progress, err := batchProcessor.GetBatchProgress(batchID)

	if err != nil {
		h.handleError(c, err)
		return
	}

	logger.Sugar().Debugw("Batch progress requested",
		"batch_id", batchID,
		"client_ip", c.ClientIP(),
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"batch_id": batchID,
		"progress": progress,
		"timestamp": time.Now().UTC(),
	})
}

// HTTPProgressSubscriber implements ProgressSubscriber for HTTP responses
type HTTPProgressSubscriber struct {
	Context   *gin.Context
	BatchID   string
	StartTime time.Time
}

// OnProgress handles progress updates for HTTP responses
func (s *HTTPProgressSubscriber) OnProgress(update *services.ProgressUpdate) {
	// For HTTP responses, we'll log progress updates
	// In a real implementation, you might use Server-Sent Events or WebSocket
	if update.IsComplete {
		if update.Error != "" {
			logger.Sugar().Errorw("Batch processing completed with errors",
				"batch_id", s.BatchID,
				"error", update.Error,
				"duration_ms", time.Since(s.StartTime).Milliseconds(),
			)
		} else {
			logger.Sugar().Infow("Batch processing completed successfully",
				"batch_id", s.BatchID,
				"total_items", update.Progress.Total,
				"successful_items", update.Progress.Successful,
				"failed_items", update.Progress.Failed,
				"duration_ms", time.Since(s.StartTime).Milliseconds(),
			)
		}
	} else {
		logger.Sugar().Debugw("Batch processing progress update",
			"batch_id", s.BatchID,
			"progress_percent", update.Progress.Progress*100,
			"processed", update.Progress.Processed,
			"total", update.Progress.Total,
			"items_per_sec", update.Progress.ItemsPerSec,
			"eta_ms", update.Progress.ETA,
		)
	}
}

// handleError handles batch request errors
func (h *BatchDecodeHandler) handleError(c *gin.Context, err error) {
	logger.Sugar().Errorw("Batch request error",
		"error", err,
		"path", c.Request.URL.Path,
		"method", c.Request.Method,
	)

	if ve, ok := err.(*validator.ValidationError); ok {
		c.JSON(ve.Code.HTTPStatus(), gin.H{
			"success": false,
			"error": gin.H{
				"code":    ve.Code.String(),
				"message": ve.Message,
				"details": ve.Details,
				"field":   ve.Field,
			},
		})
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "INTERNAL_SERVER_ERROR",
			"message": "An internal error occurred",
		},
	})
}