package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shawnvan/bl4/internal/api/models"
	"github.com/shawnvan/bl4/internal/codec/serial"
	"github.com/shawnvan/bl4/pkg/logger"
	"github.com/shawnvan/bl4/pkg/validator"
)

// EncodeHandler handles item code encoding requests
type EncodeHandler struct {
}

// NewEncodeHandler creates a new encode handler
func NewEncodeHandler() *EncodeHandler {
	return &EncodeHandler{}
}

// HandleEncode handles POST /api/v1/items/encode requests
func (h *EncodeHandler) HandleEncode(c *gin.Context) {
	startTime := time.Now()

	// Parse request
	var request models.EncodeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.handleError(c, validator.NewValidationError(validator.ErrCodeInvalidInput, "Invalid request format"))
		return
	}

	logger.Sugar().Infow("Encode request received",
		"item_level", request.ItemData.Level,
		"item_type", request.ItemData.Type,
		"manufacturer", request.ItemData.Manufacturer,
		"parts_count", len(request.ItemData.Parts),
		"options", request.Options,
	)

	// Validate request
	if err := request.Validate(); err != nil {
		h.handleError(c, err)
		return
	}

	// Create serializer with options from request
	serializerOptions := &serial.SerializerOptions{
		IncludeMetadata:  request.HasOption("include_metadata"),
		OptimizeSize:     request.HasOption("optimize_size"),
		StrictValidation: true,
		MaxProcessingTime: request.GetTimeout(),
		TargetVersion:    request.GetTargetVersion(),
	}

	serializer := serial.NewSerializerWithOptions(serializerOptions)

	// Perform serialization
	serializationResult, err := serializer.SerializeItem(request.ItemData)
	if err != nil {
		logger.Sugar().Errorw("Serialization failed",
			"item_level", request.ItemData.Level,
			"item_type", request.ItemData.Type,
			"manufacturer", request.ItemData.Manufacturer,
			"error", err.Error(),
			"duration_ms", time.Since(startTime).Milliseconds(),
		)

		// Return error response
		response := &models.EncodeResponse{
			Success: false,
			Error:   models.NewErrorInfo(err),
			Metadata: models.NewResponseMetadata(),
		}
		response.Metadata.UpdatePerformance(int64(time.Since(startTime).Microseconds()), 0)

		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Build success response
	response := &models.EncodeResponse{
		Success: true,
		Data: &models.EncodedItemData{
			SerialCode: serializationResult.SerialCode,
			ItemData:   request.ItemData,
		},
		Metadata: models.NewResponseMetadata(),
	}

	// Add encoding info if available
	if serializationResult.EncodingInfo != nil {
		response.Data.AddEncodingInfo(
			serializationResult.EncodingInfo.Statistics,
			serializationResult.EncodingInfo.FormatValidation,
			serializationResult.EncodingInfo.Optimization,
		)
	}

	// Add metadata if requested
	if serializerOptions.IncludeMetadata && serializationResult.Metadata != nil {
		response.Data.Properties = serializationResult.Metadata
	}

	// Update performance metrics
	response.Metadata.UpdatePerformance(
		int64(time.Since(startTime).Microseconds()),
		0, // Memory usage could be tracked later
	)

	logger.Sugar().Infow("Encode request completed successfully",
		"serial_code", serializationResult.SerialCode,
		"duration_ms", time.Since(startTime).Milliseconds(),
		"level", request.ItemData.Level,
		"type", request.ItemData.Type,
		"manufacturer", request.ItemData.Manufacturer,
		"parts_count", len(request.ItemData.Parts),
	)

	c.JSON(http.StatusOK, response)
}

// HandleHealthCheck handles GET /api/v1/items/encode/health requests
func (h *EncodeHandler) HandleHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "encode",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
	})
}

// HandleValidate handles POST /api/v1/items/encode/validate requests
func (h *EncodeHandler) HandleValidate(c *gin.Context) {
	startTime := time.Now()

	// Parse request
	var request models.EncodeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.handleError(c, validator.NewValidationError(validator.ErrCodeInvalidInput, "Invalid request format"))
		return
	}

	logger.Sugar().Infow("Validate encode request received",
		"item_level", request.ItemData.Level,
		"item_type", request.ItemData.Type,
		"manufacturer", request.ItemData.Manufacturer,
	)

	// Validate request
	if err := request.Validate(); err != nil {
		h.handleError(c, err)
		return
	}

	// Create validator serializer
	serializerOptions := &serial.SerializerOptions{
		IncludeMetadata:  false,
		OptimizeSize:     false,
		StrictValidation: true,
		MaxProcessingTime: request.GetTimeout(),
		TargetVersion:    request.GetTargetVersion(),
	}

	serializer := serial.NewSerializerWithOptions(serializerOptions)

	// Perform validation-only serialization
	serializationResult, err := serializer.SerializeItem(request.ItemData)
	duration := time.Since(startTime)

	// Build validation response
	response := models.ValidationResponse{
		IsValid: err == nil,
		Message: h.getValidationMessage(err),
		Details: make(map[string]interface{}),
		Duration: duration.Microseconds(),
	}

	if serializationResult != nil {
		response.Details["token_count"] = 0
		if serializationResult.TokenStream != nil {
			response.Details["token_count"] = serializationResult.TokenStream.Size()
		}
		response.Details["estimated_size"] = len(serializationResult.SerialCode)
		response.Details["processing_time_ms"] = duration.Milliseconds()
	}

	response.Details["item_summary"] = map[string]interface{}{
		"level":       request.ItemData.Level,
		"type":        request.ItemData.Type,
		"manufacturer": request.ItemData.Manufacturer,
		"parts_count": len(request.ItemData.Parts),
	}

	// Add validation warnings if any
	if err != nil {
		response.Details["error"] = err.Error()
	}

	statusCode := http.StatusOK
	if !response.IsValid {
		statusCode = http.StatusBadRequest
	}

	c.JSON(statusCode, response)
}

// handleError handles errors and returns appropriate JSON responses
func (h *EncodeHandler) handleError(c *gin.Context, err error) {
	// Log the error
	logger.Sugar().Errorw("Encode request error",
		"error", err.Error(),
		"path", c.Request.URL.Path,
		"method", c.Request.Method,
	)

	// Convert to ErrorInfo
	errorInfo := models.NewErrorInfo(err)

	// Return error response
	response := &models.EncodeResponse{
		Success: false,
		Error:   errorInfo,
		Metadata: models.NewResponseMetadata(),
	}

	// Determine appropriate status code
	statusCode := h.getStatusCode(errorInfo.Code)

	c.JSON(statusCode, response)
}

// getStatusCode determines the appropriate HTTP status code for an error
func (h *EncodeHandler) getStatusCode(errorCode string) int {
	switch errorCode {
	case "EMPTY_INPUT":
		return http.StatusBadRequest
	case "INVALID_INPUT":
		return http.StatusBadRequest
	case "INVALID_FORMAT":
		return http.StatusBadRequest
	case "INVALID_ITEM_LEVEL":
		return http.StatusBadRequest
	case "INVALID_ITEM_TYPE":
		return http.StatusBadRequest
	case "INVALID_MANUFACTURER":
		return http.StatusBadRequest
	case "INVALID_PARTS_DATA":
		return http.StatusBadRequest
	case "TOKEN_VALUE_INVALID":
		return http.StatusInternalServerError
	case "INVALID_TOKEN":
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// getValidationMessage generates a user-friendly validation message
func (h *EncodeHandler) getValidationMessage(err error) string {
	if err == nil {
		return "Item data is valid and can be encoded"
	}

	if validationErr, ok := err.(*validator.ValidationError); ok {
		switch validationErr.Code {
		case validator.ErrCodeInvalidItemLevel:
			return "Invalid item level. Must be between 1 and 100"
		case validator.ErrCodeInvalidItemType:
			return "Invalid item type. Must be one of: pistol, shotgun, rifle, smg, sniper, launcher, melee, shield, grenade, artifact"
		case validator.ErrCodeInvalidManufacturer:
			return "Invalid manufacturer. Must be one of: maliwan, jakobs, hyperion, torgue, vladof, dahl, tediore, atlas, cov, children, anshin, pangolin, eridian"
		case validator.ErrCodeInvalidPartsData:
			return "Invalid parts data. Check part indices and values"
		default:
			return fmt.Sprintf("Validation failed: %s", validationErr.Message)
		}
	}

	return fmt.Sprintf("Validation failed: %s", err.Error())
}

// Middleware for encode handler
func (h *EncodeHandler) LoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[%s] %s %s %d %s %s\n",
			param.TimeStamp.Format(time.RFC3339),
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
			param.ErrorMessage,
		)
	})
}

// BatchEncodeProgress represents progress for a batch operation
type BatchEncodeProgress struct {
	Total       int     `json:"total"`
	Processed   int     `json:"processed"`
	Successful  int     `json:"successful"`
	Failed      int     `json:"failed"`
	Progress    float64 `json:"progress"`    // 0.0 to 1.0
	ETA         int64   `json:"eta"`         // Estimated time remaining in milliseconds
	StartTime   int64   `json:"start_time"`  // Start timestamp in milliseconds
	CurrentItem string  `json:"current_item,omitempty"` // Currently processing item identifier
}

// UpdateProgress updates the progress information
func (p *BatchEncodeProgress) Update(processed, successful, failed int) {
	p.Processed = processed
	p.Successful = successful
	p.Failed = failed

	if p.Total > 0 {
		p.Progress = float64(processed) / float64(p.Total)
	}

	// Simple ETA calculation
	if processed > 0 {
		elapsed := time.Now().UnixMilli() - p.StartTime
		if p.Progress < 1.0 {
			p.ETA = int64(float64(elapsed) / p.Progress * (1.0 - p.Progress))
		} else {
			p.ETA = 0
		}
	}
}

// NewBatchEncodeProgress creates a new progress tracker
func NewBatchEncodeProgress(total int) *BatchEncodeProgress {
	return &BatchEncodeProgress{
		Total:     total,
		Processed: 0,
		Successful: 0,
		Failed:    0,
		Progress:  0.0,
		ETA:       0,
		StartTime: time.Now().UnixMilli(),
	}
}