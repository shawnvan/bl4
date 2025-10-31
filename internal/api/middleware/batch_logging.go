package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shawnvan/bl4/pkg/logger"
)

// BatchLoggingConfig contains configuration for batch operation logging
type BatchLoggingConfig struct {
	// Log request/response bodies for batch operations
	LogBodies bool

	// Sanitize sensitive information in logs
	SanitizeLogs bool

	// Include performance metrics
	IncludeMetrics bool

	// Log slow batch operations (threshold in milliseconds)
	SlowOperationThreshold int64

	// Enable structured logging
	EnableStructuredLogging bool
}

// DefaultBatchLoggingConfig returns default configuration
func DefaultBatchLoggingConfig() *BatchLoggingConfig {
	return &BatchLoggingConfig{
		LogBodies:                 true,
		SanitizeLogs:              true,
		IncludeMetrics:            true,
		SlowOperationThreshold:    5000, // 5 seconds
		EnableStructuredLogging:   true,
	}
}

// BatchLogger provides enhanced logging for batch operations
type BatchLogger struct {
	config *BatchLoggingConfig
}

// NewBatchLogger creates a new batch logger
func NewBatchLogger(config *BatchLoggingConfig) *BatchLogger {
	if config == nil {
		config = DefaultBatchLoggingConfig()
	}
	return &BatchLogger{config: config}
}

// BatchLoggingMiddleware provides enhanced logging for batch operations
func (bl *BatchLogger) BatchLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only apply to batch endpoints
		if c.Request.URL.Path != "/api/v1/items/batch/decode" {
			c.Next()
			return
		}

		startTime := time.Now()
		requestID := uuid.New().String()

		// Add request ID to context
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		// Read request body for logging
		var requestBody []byte
		if bl.config.LogBodies && c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Log request start
		bl.logBatchRequestStart(c, requestID, requestBody)

		// Create response writer wrapper to capture response
		responseWriter := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = responseWriter

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(startTime)

		// Log request completion
		bl.logBatchRequestComplete(c, requestID, responseWriter.body.String(), duration)
	}
}

// logBatchRequestStart logs the start of a batch request
func (bl *BatchLogger) logBatchRequestStart(c *gin.Context, requestID string, requestBody []byte) {
	logData := map[string]interface{}{
		"event":       "batch_request_start",
		"request_id":  requestID,
		"method":      c.Request.Method,
		"path":        c.Request.URL.Path,
		"client_ip":   c.ClientIP(),
		"user_agent":  c.Request.UserAgent(),
		"timestamp":   time.Now().UTC(),
	}

	// Add request size
	if c.Request.ContentLength > 0 {
		logData["request_size"] = c.Request.ContentLength
	}

	// Add request body if enabled
	if bl.config.LogBodies && len(requestBody) > 0 {
		var parsedBody interface{}
		if err := json.Unmarshal(requestBody, &parsedBody); err == nil {
			// Extract batch-specific information
			if bodyMap, ok := parsedBody.(map[string]interface{}); ok {
				if serialCodes, exists := bodyMap["serial_codes"]; exists {
					if codes, ok := serialCodes.([]interface{}); ok {
						logData["batch_size"] = len(codes)
					}
				}
				if options, exists := bodyMap["options"]; exists {
					logData["has_options"] = true
					if bl.config.SanitizeLogs {
						// Sanitize options to avoid logging sensitive data
						logData["options"] = "[REDACTED]"
					} else {
						logData["options"] = options
					}
				}
			}
		}
	}

	if bl.config.EnableStructuredLogging {
		logger.Sugar().Infow("Batch request started", logData)
	} else {
		logger.Sugar().Infof("Batch request started - ID: %s, Path: %s, Client: %s",
			requestID, c.Request.URL.Path, c.ClientIP())
	}
}

// logBatchRequestComplete logs the completion of a batch request
func (bl *BatchLogger) logBatchRequestComplete(c *gin.Context, requestID, responseBody string, duration time.Duration) {
	statusCode := c.Writer.Status()
	isSlow := duration.Milliseconds() > bl.config.SlowOperationThreshold

	logData := map[string]interface{}{
		"event":         "batch_request_complete",
		"request_id":    requestID,
		"method":        c.Request.Method,
		"path":          c.Request.URL.Path,
		"status_code":   statusCode,
		"duration_ms":   duration.Milliseconds(),
		"client_ip":     c.ClientIP(),
		"timestamp":     time.Now().UTC(),
		"is_slow":       isSlow,
	}

	// Add response size
	if c.Writer.Size() > 0 {
		logData["response_size"] = c.Writer.Size()
	}

	// Parse response body for batch-specific metrics
	if bl.config.LogBodies && len(responseBody) > 0 {
		var parsedResponse interface{}
		if err := json.Unmarshal([]byte(responseBody), &parsedResponse); err == nil {
			if responseMap, ok := parsedResponse.(map[string]interface{}); ok {
				// Extract success status
				if success, exists := responseMap["success"]; exists {
					logData["success"] = success
				}

				// Extract batch statistics if available
				if statistics, exists := responseMap["statistics"]; exists {
					if stats, ok := statistics.(map[string]interface{}); ok {
						if total, exists := stats["total"]; exists {
							logData["total_items"] = total
						}
						if successful, exists := stats["successful"]; exists {
							logData["successful_items"] = successful
						}
						if failed, exists := stats["failed"]; exists {
							logData["failed_items"] = failed
						}
						if successRate, exists := stats["success_rate"]; exists {
							logData["success_rate"] = successRate
						}
					}
				}

				// Extract error information if request failed
				if errorInfo, exists := responseMap["error"]; exists {
					logData["error_info"] = errorInfo
				}
			}
		}
	}

	// Log based on status and performance
	switch {
	case statusCode >= 500:
		if bl.config.EnableStructuredLogging {
			logger.Sugar().Errorw("Batch request failed with server error", logData)
		} else {
			logger.Sugar().Errorf("Batch request failed - ID: %s, Status: %d, Duration: %v",
				requestID, statusCode, duration)
		}
	case statusCode >= 400:
		if bl.config.EnableStructuredLogging {
			logger.Sugar().Warnw("Batch request failed with client error", logData)
		} else {
			logger.Sugar().Warnf("Batch request failed - ID: %s, Status: %d, Duration: %v",
				requestID, statusCode, duration)
		}
	case isSlow:
		if bl.config.EnableStructuredLogging {
			logger.Sugar().Warnw("Batch request completed but was slow", logData)
		} else {
			logger.Sugar().Warnf("Batch request completed slowly - ID: %s, Duration: %v",
				requestID, duration)
		}
	default:
		if bl.config.EnableStructuredLogging {
			logger.Sugar().Infow("Batch request completed successfully", logData)
		} else {
			logger.Sugar().Infof("Batch request completed - ID: %s, Duration: %v",
				requestID, duration)
		}
	}
}

// responseBodyWriter wraps gin.ResponseWriter to capture response body
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// BatchMetricsCollector collects and reports batch operation metrics
type BatchMetricsCollector struct {
	// Metrics could be expanded to include more sophisticated tracking
	TotalRequests     int64
	SuccessfulReqs    int64
	FailedReqs        int64
	TotalItems        int64
	TotalDuration     time.Duration
	AvgDuration       time.Duration
	SlowOperations    int64
}

// NewBatchMetricsCollector creates a new metrics collector
func NewBatchMetricsCollector() *BatchMetricsCollector {
	return &BatchMetricsCollector{}
}

// CollectMetrics collects metrics from a completed batch request
func (bmc *BatchMetricsCollector) CollectMetrics(success bool, itemCount int, duration time.Duration) {
	bmc.TotalRequests++
	bmc.TotalDuration += duration

	if success {
		bmc.SuccessfulReqs++
	} else {
		bmc.FailedReqs++
	}

	bmc.TotalItems += int64(itemCount)

	// Update average duration
	bmc.AvgDuration = time.Duration(int64(bmc.TotalDuration) / bmc.TotalRequests)

	// Track slow operations
	if duration.Milliseconds() > 5000 { // 5 second threshold
		bmc.SlowOperations++
	}
}

// GetMetrics returns current metrics
func (bmc *BatchMetricsCollector) GetMetrics() map[string]interface{} {
	successRate := float64(0)
	if bmc.TotalRequests > 0 {
		successRate = float64(bmc.SuccessfulReqs) / float64(bmc.TotalRequests)
	}

	avgItemsPerRequest := float64(0)
	if bmc.TotalRequests > 0 {
		avgItemsPerRequest = float64(bmc.TotalItems) / float64(bmc.TotalRequests)
	}

	return map[string]interface{}{
		"total_requests":       bmc.TotalRequests,
		"successful_requests":  bmc.SuccessfulReqs,
		"failed_requests":      bmc.FailedReqs,
		"success_rate":         successRate,
		"total_items":          bmc.TotalItems,
		"avg_items_per_request": avgItemsPerRequest,
		"total_duration":       bmc.TotalDuration,
		"avg_duration":         bmc.AvgDuration,
		"slow_operations":      bmc.SlowOperations,
	}
}

// LogBatchOperationProgress logs progress during batch operations
func LogBatchOperationProgress(batchID string, processed, total int, currentItem string, errors []string) {
	logger.Sugar().Infow("Batch operation progress",
		"batch_id", batchID,
		"processed", processed,
		"total", total,
		"progress_percent", float64(processed)/float64(total)*100,
		"current_item", currentItem,
		"error_count", len(errors),
		"timestamp", time.Now().UTC(),
	)
}

// LogBatchOperationCompletion logs the completion of a batch operation
func LogBatchOperationCompletion(batchID string, success bool, totalProcessed, successful, failed int, duration time.Duration, errors []error) {
	logData := map[string]interface{}{
		"event":         "batch_operation_complete",
		"batch_id":      batchID,
		"success":       success,
		"total_processed": totalProcessed,
		"successful":    successful,
		"failed":        failed,
		"duration_ms":   duration.Milliseconds(),
		"timestamp":     time.Now().UTC(),
		"error_count":   len(errors),
	}

	if len(errors) > 0 {
		// Log first few errors as examples
		errorSamples := make([]string, 0, min(len(errors), 5))
		for i, err := range errors {
			if i >= 5 {
				break
			}
			errorSamples = append(errorSamples, err.Error())
		}
		logData["error_samples"] = errorSamples
	}

	if success {
		logger.Sugar().Infow("Batch operation completed successfully", logData)
	} else {
		logger.Sugar().Errorw("Batch operation completed with errors", logData)
	}
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}