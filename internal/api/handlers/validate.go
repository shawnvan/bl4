package handlers

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shawnvan/bl4/internal/api/models"
	"github.com/shawnvan/bl4/internal/codec/base85"
	"github.com/shawnvan/bl4/pkg/logger"
)

// ValidateHandler handles serial code validation requests
type ValidateHandler struct{}

// NewValidateHandler creates a new validate handler
func NewValidateHandler() *ValidateHandler {
	return &ValidateHandler{}
}

// HandleValidate handles POST /api/v1/items/validate requests
func (h *ValidateHandler) HandleValidate(c *gin.Context) {
	startTime := time.Now()
	requestID := uuid.New().String()

	// Parse request
	var request models.ValidateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Sugar().Errorw("Invalid validate request",
			"request_id", requestID,
			"error", err,
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"success":     false,
			"request_id": requestID,
			"error":       "Invalid request format",
			"error_type":  "validation_error",
			"status_code": http.StatusBadRequest,
		})
		return
	}

	logger.Sugar().Infow("Processing validate request",
		"request_id", requestID,
		"serial_code", request.SerialCode,
	)

	// Validate input
	code := strings.TrimSpace(request.SerialCode)
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":     false,
			"request_id": requestID,
			"serial_code": code,
			"error":       "Serial code is required",
			"error_type":  "validation_error",
			"status_code": http.StatusBadRequest,
		})
		return
	}

	// Perform validation
	validationResult := h.validateSerialCode(code)
	validationResult.RequestID = requestID
	validationResult.ProcessingTimeMs = time.Since(startTime).Milliseconds()

	// Log results
	logger.Sugar().Infow("Validation completed",
		"request_id", requestID,
		"serial_code", code,
		"valid", validationResult.Valid,
		"processing_time_ms", validationResult.ProcessingTimeMs,
	)

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"success": validationResult.Success,
		"valid": validationResult.Valid,
		"serial_code": validationResult.SerialCode,
		"request_id": validationResult.RequestID,
		"error": validationResult.Error,
		"error_type": validationResult.ErrorType,
		"length": validationResult.Length,
		"base85_length": validationResult.Base85Length,
		"validation_type": validationResult.ValidationType,
		"issues": validationResult.Issues,
		"details": validationResult.Details,
		"processing_time_ms": validationResult.ProcessingTimeMs,
	})
}

// validateSerialCode performs comprehensive validation of a BL4 serial code
func (h *ValidateHandler) validateSerialCode(code string) models.ComprehensiveValidateResponse {
	response := models.ComprehensiveValidateResponse{
		SerialCode: code,
		Success:    true,
		Length:     len(code),
	}

	// Basic format validation
	if err := h.validateBasicFormat(code); err != nil {
		response.Valid = false
		response.Error = err.Error()
		response.ErrorType = "format_error"
		response.Issues = append(response.Issues, err.Error())
		return response
	}

	// Check @U prefix (BL4 serial codes require @U prefix)
	if !strings.HasPrefix(code, "@U") {
		response.Valid = false
		response.Error = "BL4 serial codes must start with '@U'"
		response.ErrorType = "format_error"
		response.Issues = append(response.Issues, "Missing @U prefix")
		return response
	}

	// Extract Base85 part
	b85Part := code[2:]
	response.Base85Length = len(b85Part)

	// Validate Base85 characters
	if err := h.validateBase85Characters(b85Part); err != nil {
		response.Valid = false
		response.Error = err.Error()
		response.ErrorType = "character_error"
		response.Issues = append(response.Issues, err.Error())
		return response
	}

	// Check length constraints
	if len(b85Part) < 1 {
		response.Valid = false
		response.Error = "Serial code too short (minimum 1 character after @U)"
		response.ErrorType = "length_error"
		response.Issues = append(response.Issues, "Too short")
		return response
	}

	if len(b85Part) > 1000 {
		response.Valid = false
		response.Error = "Serial code too long (maximum 1000 characters after @U)"
		response.ErrorType = "length_error"
		response.Issues = append(response.Issues, "Too long")
		return response
	}

	// Try Base85 decode for structural validation
	if err := h.validateBase85Structure(b85Part); err != nil {
		response.Valid = false
		response.Error = fmt.Sprintf("Invalid Base85 encoding: %v", err)
		response.ErrorType = "encoding_error"
		response.Issues = append(response.Issues, err.Error())
		return response
	}

	// If we get here, the code is valid
	response.Valid = true
	response.ValidationType = "comprehensive"
	response.Details = map[string]interface{}{
		"has_prefix":    true,
		"base85_valid":  true,
		"length_valid":  true,
		"structure_valid": true,
		"format_compliant": true,
	}

	return response
}

// validateBasicFormat performs basic format validation
func (h *ValidateHandler) validateBasicFormat(code string) error {
	if code == "" {
		return fmt.Errorf("empty code")
	}

	// Check for whitespace
	if strings.ContainsAny(code, " \t\n\r") {
		return fmt.Errorf("contains whitespace")
	}

	// Check for control characters
	for _, r := range code {
		if r < 32 || r > 126 {
			return fmt.Errorf("contains invalid control characters")
		}
	}

	return nil
}

// validateBase85Characters validates that all characters are valid Base85
func (h *ValidateHandler) validateBase85Characters(data string) error {
	// Base85 valid character set (Z85 variant) - must match decoder character set exactly
	validChars := regexp.MustCompile(`^[0-9A-Za-z!#$%&()*+\-;<=>?@^_\{\|/\}~]+$`)

	if !validChars.MatchString(data) {
		// Find invalid characters - use same character set as decoder
		invalidChars := ""
		for _, r := range data {
			if !strings.ContainsRune("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz!#$%&()*+-;<=>?@^_`{|}/~", r) {
				invalidChars += string(r)
			}
		}
		return fmt.Errorf("contains invalid characters: %s", invalidChars)
	}

	return nil
}

// validateBase85Structure attempts to decode to validate structure
func (h *ValidateHandler) validateBase85Structure(data string) error {
	decoder := base85.NewDecoder()
	// The decoder expects the full @U prefix, so add it back
	fullCode := "@U" + data
	_, err := decoder.Decode(fullCode)
	if err != nil {
		return fmt.Errorf("base85 decode failed: %w", err)
	}
	return nil
}

// HandleValidateBatch handles POST /api/v1/items/validate/batch requests
func (h *ValidateHandler) HandleValidateBatch(c *gin.Context) {
	startTime := time.Now()
	requestID := uuid.New().String()

	// Parse request
	var request models.BatchValidateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Sugar().Errorw("Invalid batch validate request",
			"request_id", requestID,
			"error", err,
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"success":     false,
			"request_id": requestID,
			"error":      "Invalid request format",
		})
		return
	}

	if len(request.SerialCodes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":     false,
			"request_id": requestID,
			"error":      "No serial codes provided",
		})
		return
	}

	logger.Sugar().Infow("Processing batch validate request",
		"request_id", requestID,
		"codes_count", len(request.SerialCodes),
	)

	// Process each code
	results := make([]models.ComprehensiveValidateResponse, len(request.SerialCodes))
	validCount := 0

	for i, code := range request.SerialCodes {
		validationResult := h.validateSerialCode(code)
		validationResult.RequestID = requestID
		results[i] = validationResult

		if validationResult.Valid {
			validCount++
		}
	}

	// Create response
	response := models.BatchValidateResponse{
		Success:       true,
		RequestID:     requestID,
		TotalCodes:    len(request.SerialCodes),
		ValidCodes:    validCount,
		InvalidCodes:  len(request.SerialCodes) - validCount,
		Results:       results,
		ProcessingTimeMs: time.Since(startTime).Milliseconds(),
	}

	logger.Sugar().Infow("Batch validation completed",
		"request_id", requestID,
		"total_codes", len(request.SerialCodes),
		"valid_codes", validCount,
		"processing_time_ms", response.ProcessingTimeMs,
	)

	c.JSON(http.StatusOK, response)
}