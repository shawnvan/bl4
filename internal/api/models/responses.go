package models

import (
	"fmt"
	"time"

	"github.com/shawnvan/bl4/internal/codec/token"
	"github.com/shawnvan/bl4/pkg/validator"
)

// DecodeResponse represents the response from a decode operation
type DecodeResponse struct {
	Success  bool                    `json:"success"`
	Data     *DecodedItemData        `json:"data,omitempty"`
	Error    *ErrorInfo              `json:"error,omitempty"`
	Metadata *ResponseMetadata       `json:"metadata,omitempty"`
}

// EncodedItemData represents the successfully decoded item data
type DecodedItemData struct {
	SerialCode string                 `json:"serial_code" example:"@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}"`
	ItemData   *ItemData              `json:"item_data"`
	Bitstream  *BitstreamInfo         `json:"bitstream,omitempty"`
	Tokens     []TokenInfo            `json:"tokens,omitempty"`
	RawData    *RawDataInfo           `json:"raw_data,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// BitstreamInfo contains information about the parsed bitstream
type BitstreamInfo struct {
	Size       int                `json:"size"`        // Total size in bits
	ByteSize   int                `json:"byte_size"`   // Size in bytes
	Format     string             `json:"format"`      // Bitstream format version
	Data       []byte             `json:"data,omitempty"` // Raw bitstream data (if requested)
	Statistics *BitstreamStats    `json:"statistics,omitempty"`
}

// BitstreamStats contains statistics about the bitstream
type BitstreamStats struct {
	TokenCount      int                    `json:"token_count"`
	TokenTypes      map[string]int         `json:"token_types"`
	TotalValues     int                    `json:"total_values"`
	UniqueValues    int                    `json:"unique_values"`
	Entropy         float64                `json:"entropy"`          // Data entropy (0-1)
	CompressionRatio float64               `json:"compression_ratio"` // Size ratio vs. raw data
}

// TokenInfo represents information about a parsed token
type TokenInfo struct {
	Type        string                 `json:"type"`                 // Token type name
	Value       interface{}            `json:"value"`                // Token value
	BitSize     int                    `json:"bit_size"`             // Size in bits
	Position    int64                  `json:"position"`             // Bit position in stream
	Properties  map[string]interface{} `json:"properties,omitempty"` // Additional token properties
}

// RawDataInfo contains raw decoded data representations
type RawDataInfo struct {
	Structured string                 `json:"structured"` // Human-readable structured format
	JSON       string                 `json:"json,omitempty"`       // JSON representation
	CSV        string                 `json:"csv,omitempty"`        // CSV representation
	Hex        string                 `json:"hex,omitempty"`        // Hexadecimal representation
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// EncodeResponse represents the response from an encode operation
type EncodeResponse struct {
	Success  bool                    `json:"success"`
	Data     *EncodedItemData        `json:"data,omitempty"`
	Error    *ErrorInfo              `json:"error,omitempty"`
	Metadata *ResponseMetadata       `json:"metadata,omitempty"`
}

// EncodedItemData represents the successfully encoded item data
type EncodedItemData struct {
	SerialCode string                 `json:"serial_code" example:"@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}"`
	ItemData   *ItemData              `json:"item_data"`
	Encoding   *EncodingInfo          `json:"encoding,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// AddEncodingInfo adds encoding information to the response
func (d *EncodedItemData) AddEncodingInfo(statistics interface{}, validation interface{}, optimization interface{}) {
	if d.Encoding == nil {
		d.Encoding = &EncodingInfo{
			Properties: make(map[string]interface{}),
		}
	}

	d.Encoding.Format = "bl4-base85"
	d.Encoding.Algorithm = "base85_v1"
	d.Encoding.Version = "1.0"

	// Add original and encoded sizes
	if d.ItemData != nil {
		d.Encoding.OriginalSize = calculateItemDataSize(d.ItemData)
	}
	d.Encoding.EncodedSize = len(d.SerialCode)

	if d.Encoding.OriginalSize > 0 {
		d.Encoding.Compression = float64(d.Encoding.EncodedSize) / float64(d.Encoding.OriginalSize)
	}

	// Add additional properties
	if statistics != nil {
		d.Encoding.Properties["statistics"] = statistics
	}
	if validation != nil {
		d.Encoding.Properties["validation"] = validation
	}
	if optimization != nil {
		d.Encoding.Properties["optimization"] = optimization
	}
}

// calculateItemDataSize estimates the size of item data
func calculateItemDataSize(itemData *ItemData) int {
	if itemData == nil {
		return 0
	}

	size := 0

	// Basic fields (estimated)
	size += 8 // level (8 bits)
	size += 8 // type (8 bits)
	size += 8 // manufacturer (8 bits)

	// Parts (each part ~32 bits)
	if itemData.Parts != nil {
		size += len(itemData.Parts) * 32
	}

	// Strings (estimated)
	if itemData.Name != "" {
		size += len(itemData.Name) * 8 + 16 // 8 bits per char + overhead
	}
	if itemData.Description != "" {
		size += len(itemData.Description) * 8 + 16
	}
	if itemData.Rarity != "" {
		size += len(itemData.Rarity) * 8 + 16
	}

	return size / 8 // Convert to bytes
}

// EncodingInfo contains information about the encoding process
type EncodingInfo struct {
	Format       string                 `json:"format"`        // Encoding format
	Algorithm    string                 `json:"algorithm"`     // Algorithm used
	Version      string                 `json:"version"`       // Format version
	OriginalSize int                    `json:"original_size"` // Original data size in bytes
	EncodedSize  int                    `json:"encoded_size"`  // Encoded data size in bytes
	Compression  float64                `json:"compression"`   // Compression ratio
	Properties   map[string]interface{} `json:"properties,omitempty"`
}

// BatchDecodeResponse represents the response from a batch decode operation
type BatchDecodeResponse struct {
	Success    bool                        `json:"success"`
	Results    []*BatchDecodeResult        `json:"results,omitempty"`
	Statistics *BatchStatistics            `json:"statistics,omitempty"`
	Error      *ErrorInfo                  `json:"error,omitempty"`
	Metadata   *ResponseMetadata           `json:"metadata,omitempty"`
}

// BatchDecodeResult represents a single result in a batch operation
type BatchDecodeResult struct {
	Index      int                     `json:"index"`
	SerialCode string                  `json:"serial_code"`
	Success    bool                    `json:"success"`
	Data       *DecodedItemData        `json:"data,omitempty"`
	Error      *ErrorInfo              `json:"error,omitempty"`
	Duration   int64                   `json:"duration"` // Processing time in microseconds
}

// BatchEncodeResponse represents the response from a batch encode operation
type BatchEncodeResponse struct {
	Success    bool                        `json:"success"`
	Results    []*BatchEncodeResult        `json:"results,omitempty"`
	Statistics *BatchStatistics            `json:"statistics,omitempty"`
	Error      *ErrorInfo                  `json:"error,omitempty"`
	Metadata   *ResponseMetadata           `json:"metadata,omitempty"`
}

// BatchEncodeResult represents a single result in a batch encode operation
type BatchEncodeResult struct {
	Index    int                `json:"index"`
	ItemData *ItemData          `json:"item_data"`
	Success  bool               `json:"success"`
	Data     *EncodedItemData   `json:"data,omitempty"`
	Error    *ErrorInfo         `json:"error,omitempty"`
	Duration int64              `json:"duration"` // Processing time in microseconds
}

// ValidateResponse represents the response from a validation operation
type ValidateResponse struct {
	Success   bool                 `json:"success"`
	IsValid   bool                 `json:"is_valid"`
	Data      *ValidationData      `json:"data,omitempty"`
	Error     *ErrorInfo           `json:"error,omitempty"`
	Metadata  *ResponseMetadata    `json:"metadata,omitempty"`
}

// ValidationData contains validation results
type ValidationData struct {
	SerialCode string                 `json:"serial_code"`
	Checks     map[string]*CheckResult `json:"checks"`
	Summary    *ValidationSummary     `json:"summary"`
	Details    map[string]interface{} `json:"details,omitempty"`
}

// CheckResult represents the result of a single validation check
type CheckResult struct {
	Passed   bool                   `json:"passed"`
	Message  string                 `json:"message,omitempty"`
	Details  map[string]interface{} `json:"details,omitempty"`
	Duration int64                  `json:"duration"` // Check time in microseconds
}

// ValidationSummary provides a summary of validation results
type ValidationSummary struct {
	TotalChecks int     `json:"total_checks"`
	Passed      int     `json:"passed"`
	Failed      int     `json:"failed"`
	Warnings    int     `json:"warnings"`
	Overall     string  `json:"overall"` // "valid", "invalid", "warning"
	Score       float64 `json:"score"`   // 0.0 to 1.0
}

// ErrorInfo represents detailed error information
type ErrorInfo struct {
	Code        string                 `json:"code"`                  // Error code
	Message     string                 `json:"message"`               // Human-readable message
	Type        string                 `json:"type"`                  // Error type (validation, processing, etc.)
	Severity    string                 `json:"severity"`              // Error severity (error, warning, info)
	Field       string                 `json:"field,omitempty"`       // Field that caused the error (if applicable)
	Value       interface{}            `json:"value,omitempty"`       // Value that caused the error (if applicable)
	Details     map[string]interface{} `json:"details,omitempty"`     // Additional error details
	Cause       *ErrorInfo             `json:"cause,omitempty"`       // Underlying cause
	Suggestions []string               `json:"suggestions,omitempty"` // Suggested fixes
	Timestamp   time.Time              `json:"timestamp"`             // When the error occurred
}

// ResponseMetadata contains metadata about the response
type ResponseMetadata struct {
	Version       string                 `json:"version"`        // API version
	Timestamp     time.Time              `json:"timestamp"`      // Response timestamp
	RequestID     string                 `json:"request_id"`     // Unique request identifier
	Duration      int64                  `json:"duration"`       // Total processing time in microseconds
	ProcessedBy   string                 `json:"processed_by"`   // Service that processed the request
	Environment   string                 `json:"environment"`    // Processing environment
	Performance   *PerformanceMetrics    `json:"performance,omitempty"`
	Properties    map[string]interface{} `json:"properties,omitempty"`
}

// PerformanceMetrics contains performance information
type PerformanceMetrics struct {
	CPUTime    int64   `json:"cpu_time"`     // CPU time in microseconds
	MemoryUsed int64   `json:"memory_used"`  // Memory used in bytes
	Throughput float64 `json:"throughput"`   // Items processed per second
	Latency    int64   `json:"latency"`      // Response latency in microseconds
	Operations  int     `json:"operations"`   // Number of operations performed
}

// ValidationResponse represents the response from a validation operation
type ValidationResponse struct {
	IsValid  bool                   `json:"is_valid"`
	Message  string                 `json:"message"`
	Details  map[string]interface{} `json:"details,omitempty"`
	Duration int64                  `json:"duration"` // Validation time in microseconds
}

// Helper methods

// NewDecodeResponse creates a successful decode response
func NewDecodeResponse(serialCode string, itemData *ItemData) *DecodeResponse {
	return &DecodeResponse{
		Success: true,
		Data: &DecodedItemData{
			SerialCode: serialCode,
			ItemData:   itemData,
		},
		Metadata: NewResponseMetadata(),
	}
}

// NewDecodeResponseWithError creates a decode response with an error
func NewDecodeResponseWithError(err error) *DecodeResponse {
	return &DecodeResponse{
		Success: false,
		Error:   NewErrorInfo(err),
		Metadata: NewResponseMetadata(),
	}
}

// NewEncodeResponse creates a successful encode response
func NewEncodeResponse(serialCode string, itemData *ItemData) *EncodeResponse {
	return &EncodeResponse{
		Success: true,
		Data: &EncodedItemData{
			SerialCode: serialCode,
			ItemData:   itemData,
		},
		Metadata: NewResponseMetadata(),
	}
}

// NewEncodeResponseWithError creates an encode response with an error
func NewEncodeResponseWithError(err error) *EncodeResponse {
	return &EncodeResponse{
		Success: false,
		Error:   NewErrorInfo(err),
		Metadata: NewResponseMetadata(),
	}
}

// NewValidateResponse creates a validation response
func NewValidateResponse(isValid bool, data *ValidationData) *ValidateResponse {
	return &ValidateResponse{
		Success:  true,
		IsValid:  isValid,
		Data:     data,
		Metadata: NewResponseMetadata(),
	}
}

// NewValidateResponseWithError creates a validation response with an error
func NewValidateResponseWithError(err error) *ValidateResponse {
	return &ValidateResponse{
		Success:  false,
		Error:    NewErrorInfo(err),
		Metadata: NewResponseMetadata(),
	}
}

// NewErrorInfo creates error information from an error
func NewErrorInfo(err error) *ErrorInfo {
	if err == nil {
		return nil
	}

	errorInfo := &ErrorInfo{
		Message:   err.Error(),
		Timestamp: time.Now().UTC(),
	}

	// Try to extract more information if it's a ValidationError
	if ve, ok := err.(*validator.ValidationError); ok {
		errorInfo.Code = ve.Code.String()
		errorInfo.Type = "validation"
		errorInfo.Field = ve.Field
		errorInfo.Value = ve.Value
		errorInfo.Details = ve.Details
		errorInfo.Severity = "error"

		// Add suggestions based on error type
		errorInfo.Suggestions = getSuggestionsForError(ve.Code)
	} else {
		errorInfo.Code = "UNKNOWN_ERROR"
		errorInfo.Type = "processing"
		errorInfo.Severity = "error"
		errorInfo.Suggestions = []string{"Please try again later", "Contact support if the problem persists"}
	}

	return errorInfo
}

// NewResponseMetadata creates default response metadata
func NewResponseMetadata() *ResponseMetadata {
	return &ResponseMetadata{
		Version:     "1.0.0",
		Timestamp:   time.Now().UTC(),
		ProcessedBy: "bl4-api",
		Environment: "development",
		Performance: &PerformanceMetrics{
			CPUTime:    0,
			MemoryUsed: 0,
			Throughput: 0,
			Latency:    0,
			Operations: 1,
		},
	}
}

// NewBatchStatistics creates batch operation statistics
func NewBatchStatistics(total, successful, failed int) *BatchStatistics {
	successRate := float64(0)
	if total > 0 {
		successRate = float64(successful) / float64(total)
	}

	return &BatchStatistics{
		Total:        total,
		Successful:   successful,
		Failed:       failed,
		SuccessRate:  successRate,
		AverageTime:  0, // Will be calculated by the batch processor
		MinTime:      0,
		MaxTime:      0,
		StartTime:    time.Now().UTC(),
		EndTime:      time.Time{},
	}
}

// BatchStatistics contains statistics for batch operations
type BatchStatistics struct {
	Total       int       `json:"total"`
	Successful  int       `json:"successful"`
	Failed      int       `json:"failed"`
	SuccessRate float64   `json:"success_rate"`
	AverageTime int64     `json:"average_time"` // Average processing time in microseconds
	MinTime     int64     `json:"min_time"`      // Minimum processing time
	MaxTime     int64     `json:"max_time"`      // Maximum processing time
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
}

// AddBitstreamInfo adds bitstream information to a decoded item
func (d *DecodedItemData) AddBitstreamInfo(tokens *token.TokenStream, bitstreamData []byte, includeData bool) {
	d.Bitstream = &BitstreamInfo{
		Size:     tokens.GetTotalBitSize(),
		ByteSize: len(bitstreamData),
		Format:   "bl4-v1.0",
		Statistics: &BitstreamStats{
			TokenCount:   tokens.Size(),
			TokenTypes:   make(map[string]int),
			TotalValues:  tokens.Size(),
			UniqueValues: calculateUniqueValues(tokens),
		},
	}

	if includeData {
		d.Bitstream.Data = bitstreamData
	}

	// Calculate token type distribution
	for _, token := range tokens.Tokens {
		tokenType := token.Type.String()
		d.Bitstream.Statistics.TokenTypes[tokenType]++
	}
}

// AddTokens adds token information to a decoded item
func (d *DecodedItemData) AddTokens(tokens *token.TokenStream) {
	d.Tokens = make([]TokenInfo, len(tokens.Tokens))

	for i, token := range tokens.Tokens {
		d.Tokens[i] = TokenInfo{
			Type:     token.Type.String(),
			Value:    formatTokenValue(token),
			BitSize:  token.BitSize,
			Position: token.Position,
			Properties: map[string]interface{}{
				"is_data":    token.IsDataToken(),
				"is_structural": token.IsStructuralToken(),
			},
		}
	}
}

// AddRawData adds raw data representations to a decoded item
func (d *DecodedItemData) AddRawData(structured, jsonStr, csvStr, hexStr string) {
	d.RawData = &RawDataInfo{
		Structured: structured,
		Properties: make(map[string]interface{}),
	}

	if jsonStr != "" {
		d.RawData.JSON = jsonStr
	}
	if csvStr != "" {
		d.RawData.CSV = csvStr
	}
	if hexStr != "" {
		d.RawData.Hex = hexStr
	}
}

// UpdatePerformance updates performance metrics
func (m *ResponseMetadata) UpdatePerformance(duration int64, memoryUsed int64) {
	if m.Performance != nil {
		m.Performance.CPUTime = duration
		m.Performance.MemoryUsed = memoryUsed
		m.Performance.Latency = duration
	}
}

// Helper functions

// getSuggestionsForError returns suggestions based on error code
func getSuggestionsForError(code validator.ErrorCode) []string {
	switch code {
	case validator.ErrCodeInvalidInput:
		return []string{"Check the input format", "Ensure all required fields are provided"}
	case validator.ErrCodeEmptyInput:
		return []string{"Provide a non-empty value", "Check for missing data"}
	case validator.ErrCodeInvalidFormat:
		return []string{"Verify the data format matches expected pattern", "Check for format specifications"}
	case validator.ErrCodeInvalidLength:
		return []string{"Ensure the data has the correct length", "Check minimum and maximum size requirements"}
	case validator.ErrCodeInvalidCharacter:
		return []string{"Remove invalid characters", "Use only valid Base85 characters"}
	case validator.ErrCodeInvalidPrefix:
		return []string{"Add the required @U prefix", "Check the prefix format"}
	case validator.ErrCodeInvalidBase85:
		return []string{"Ensure valid Base85 encoding", "Check for invalid characters in the code"}
	case validator.ErrCodeInvalidItemData:
		return []string{"Verify item data structure", "Check for missing required fields"}
	case validator.ErrCodeProcessingTimeout:
		return []string{"Try with a smaller item code", "Increase timeout value"}
	case validator.ErrCodeBatchSizeExceeded:
		return []string{"Reduce the number of items", "Process in smaller batches"}
	default:
		return []string{"Please check your input", "Contact support if the problem persists"}
	}
}

// formatTokenValue formats a token value for display
func formatTokenValue(t token.Token) interface{} {
	switch t.Type {
	case token.TokenVARBIT:
		if bits, ok := t.Value.([]bool); ok {
			var bitStr string
			for i, bit := range bits {
				if i > 0 && i%8 == 0 {
					bitStr += " "
				}
				if bit {
					bitStr += "1"
				} else {
					bitStr += "0"
				}
			}
			return bitStr
		}
	case token.TokenSTRING:
		return t.Value
	case token.TokenPART:
		if partData, ok := t.Value.(uint64); ok {
			return fmt.Sprintf("index:%d", partData)
		}
	}
	return t.Value
}

// calculateUniqueValues counts unique values in a token stream
func calculateUniqueValues(tokens *token.TokenStream) int {
	seen := make(map[interface{}]bool)
	for _, token := range tokens.Tokens {
		if _, exists := seen[token.Value]; !exists {
			seen[token.Value] = true
		}
	}
	return len(seen)
}

// IsSuccess returns true if the response indicates success
func (r *DecodeResponse) IsSuccess() bool {
	return r.Success && r.Error == nil
}

// IsSuccess returns true if the response indicates success
func (r *EncodeResponse) IsSuccess() bool {
	return r.Success && r.Error == nil
}

// IsSuccess returns true if the response indicates success
func (r *ValidateResponse) IsSuccess() bool {
	return r.Success && r.Error == nil
}

// GetProcessingTime returns the total processing time
func (m *ResponseMetadata) GetProcessingTime() time.Duration {
	if m.Performance != nil {
		return time.Duration(m.Performance.CPUTime) * time.Microsecond
	}
	return 0
}

// ComprehensiveValidateResponse represents the response from a validation operation
type ComprehensiveValidateResponse struct {
	Success           bool                   `json:"success"`
	Valid             bool                   `json:"valid"`
	SerialCode        string                 `json:"serial_code"`
	RequestID         string                 `json:"request_id,omitempty"`
	Error             string                 `json:"error,omitempty"`
	ErrorType         string                 `json:"error_type,omitempty"`
	StatusCode        int                    `json:"status_code,omitempty"`
	Length            int                    `json:"length,omitempty"`
	Base85Length      int                    `json:"base85_length,omitempty"`
	ValidationType    string                 `json:"validation_type,omitempty"`
	Issues            []string               `json:"issues,omitempty"`
	Details           map[string]interface{} `json:"details,omitempty"`
	ProcessingTimeMs  int64                  `json:"processing_time_ms,omitempty"`
}

// BatchValidateRequest represents a batch validation request
type BatchValidateRequest struct {
	SerialCodes []string        `json:"serial_codes" binding:"required"`
	Options     *ValidateOptions `json:"options,omitempty"`
}

// BatchValidateResponse represents the response from a batch validation operation
type BatchValidateResponse struct {
	Success          bool                             `json:"success"`
	RequestID        string                           `json:"request_id,omitempty"`
	TotalCodes       int                              `json:"total_codes"`
	ValidCodes       int                              `json:"valid_codes"`
	InvalidCodes     int                              `json:"invalid_codes"`
	Results          []ComprehensiveValidateResponse  `json:"results"`
	ProcessingTimeMs int64                            `json:"processing_time_ms,omitempty"`
}