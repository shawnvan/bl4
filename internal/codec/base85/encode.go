package base85

import (
	"fmt"

	"github.com/shawnvan/bl4/pkg/validator"
)

// Encoder handles Base85 encoding operations for BL4 item codes
type Encoder struct {
	charset *Charset
}

// NewEncoder creates a new Base85 encoder
func NewEncoder() *Encoder {
	return &Encoder{
		charset: GlobalCharset,
	}
}

// Encode encodes raw bytes to a Base85 item serial code
func (e *Encoder) Encode(data []byte) (string, error) {
	if len(data) == 0 {
		return "", validator.NewValidationError(validator.ErrCodeEmptyInput, "input data is empty")
	}

	// Validate data size
	if len(data) > MaxEncodedSize {
		return "", validator.NewValidationError(validator.ErrCodeInvalidInput,
			fmt.Sprintf("input data too large: %d bytes", len(data))).
			WithField("byte_count", len(data)).
			WithDetail("max_bytes", MaxEncodedSize)
	}

	// Encode bytes to Base85
	encoded, err := e.encodeBase85(data)
	if err != nil {
		return "", fmt.Errorf("Base85 encode failed: %w", err)
	}

	// Add BL4 prefix
	prefixBytes := e.charset.AddPrefix([]byte(encoded))
	result := string(prefixBytes)

	return result, nil
}

// encodeBase85 performs the actual Base85 encoding
func (e *Encoder) encodeBase85(data []byte) (string, error) {
	if len(data) == 0 {
		return "", nil
	}

	// Process 4 bytes at a time (4 bytes = 5 Base85 chars)
	result := ""

	for i := 0; i < len(data); i += 4 {
		// Get up to 4 bytes
		remaining := len(data) - i
		chunkSize := 4
		if remaining < chunkSize {
			chunkSize = remaining
		}

		// Create 4-byte chunk, padding with zeros if needed
		chunk := make([]byte, 4)
		copy(chunk, data[i:i+chunkSize])

		// Convert bytes to 32-bit integer (big-endian)
		value := uint32(chunk[0])<<24 | uint32(chunk[1])<<16 | uint32(chunk[2])<<8 | uint32(chunk[3])

		// Encode to 5 Base85 characters
		encoded, err := e.encodeQuintet(value)
		if err != nil {
			return "", fmt.Errorf("failed to encode chunk at position %d: %w", i, err)
		}

		result += encoded
	}

	return result, nil
}

// encodeQuintet encodes a 32-bit value to 5 Base85 characters
func (e *Encoder) encodeQuintet(value uint32) (string, error) {
	if value > MaxUint32Value {
		return "", validator.NewValidationError(validator.ErrCodeInvalidInput,
			fmt.Sprintf("value too large for Base85 encoding: %d", value)).
			WithField("value", value).
			WithDetail("max_value", MaxUint32Value)
	}

	// Extract 5 Base85 values (big-endian)
	chars := make([]byte, 5)
	remaining := value

	// Extract from least significant to most significant
	for i := 4; i >= 0; i-- {
		chars[i] = byte(remaining % 85)
		remaining /= 85
	}

	// Convert to Base85 characters using the Encoder lookup table
	result := ""
	for i := 0; i < 5; i++ {
		if int(chars[i]) >= len(e.charset.Encoder) {
			return "", fmt.Errorf("invalid Base85 value %d at position %d", chars[i], i)
		}
		char := e.charset.Encoder[chars[i]]
		result += string(char)
	}

	return result, nil
}

// EncodeWithOptions encodes with additional options
func (e *Encoder) EncodeWithOptions(data []byte, options *EncodeOptions) (*EncodeResult, error) {
	result := &EncodeResult{
		Data:     data,
		ByteSize: len(data),
	}

	// Validate input
	if err := e.ValidateData(data); err != nil {
		result.Error = err
		return result, err
	}

	// Perform encoding
	encoded, err := e.Encode(data)
	if err != nil {
		result.Error = err
		return result, err
	}

	result.SerialCode = encoded
	result.CharCount = len(encoded) - 2 // Subtract prefix length

	// Add additional information if requested
	if options != nil {
		if options.IncludeStatistics {
			result.Statistics = e.calculateStatistics(data, encoded)
		}

		if options.ValidateFormat {
			result.FormatValidation = e.validateEncodedFormat(encoded)
		}

		if options.OptimizeSize {
			result.Optimization = e.optimizeEncoding(data, encoded)
		}
	}

	return result, nil
}

// ValidateData validates the input data before encoding
func (e *Encoder) ValidateData(data []byte) error {
	if len(data) == 0 {
		return validator.NewValidationError(validator.ErrCodeEmptyInput, "input data is empty")
	}

	if len(data) > MaxEncodedSize {
		return validator.NewValidationError(validator.ErrCodeInvalidInput,
			fmt.Sprintf("input data too large: %d bytes", len(data))).
			WithField("byte_count", len(data)).
			WithDetail("max_bytes", MaxEncodedSize)
	}

	return nil
}

// EncodeOptions contains options for encoding operations
type EncodeOptions struct {
	IncludeStatistics bool `json:"include_statistics,omitempty"`
	ValidateFormat     bool `json:"validate_format,omitempty"`
	OptimizeSize       bool `json:"optimize_size,omitempty"`
	TargetVersion      string `json:"target_version,omitempty"`
}

// EncodeResult contains the result of an encode operation
type EncodeResult struct {
	SerialCode      string                `json:"serial_code,omitempty"`
	Data            []byte                `json:"data,omitempty"`
	ByteSize        int                   `json:"byte_size,omitempty"`
	CharCount       int                   `json:"char_count,omitempty"`
	Statistics      *EncodeStatistics     `json:"statistics,omitempty"`
	FormatValidation *FormatValidation     `json:"format_validation,omitempty"`
	Optimization    *OptimizationInfo     `json:"optimization,omitempty"`
	Error           error                 `json:"error,omitempty"`
}

// EncodeStatistics contains statistics about the encode operation
type EncodeStatistics struct {
	CompressionRatio float64 `json:"compression_ratio"` // encoded_size / decoded_size
	DataDensity      float64 `json:"data_density"`       // actual_data_bits / total_bits
	Entropy          float64 `json:"entropy"`             // Data entropy estimate
	PaddingUsed      int     `json:"padding_used"`        // Number of padding bytes used
	ChunkCount       int     `json:"chunk_count"`         // Number of 4-byte chunks processed
}

// OptimizationInfo contains information about size optimization
type OptimizationInfo struct {
	OriginalSize    int     `json:"original_size"`    // Original data size in bytes
	OptimizedSize   int     `json:"optimized_size"`   // Optimized data size in bytes
	SpaceSaved      int     `json:"space_saved"`      // Bytes saved through optimization
	CompressionGain float64 `json:"compression_gain"` // Percentage of size reduction
	Techniques      []string `json:"techniques"`       // Optimization techniques applied
}

// calculateStatistics calculates statistics about the encode operation
func (e *Encoder) calculateStatistics(data []byte, encoded string) *EncodeStatistics {
	dataSize := len(data)
	encodedSize := len(encoded)

	stats := &EncodeStatistics{
		CompressionRatio: float64(encodedSize) / float64(dataSize),
		DataDensity:      0.0,
		Entropy:          0.0,
		PaddingUsed:      0,
		ChunkCount:       (dataSize + 3) / 4, // Round up division
	}

	// Calculate padding used
	paddingNeeded := (4 - (dataSize % 4)) % 4
	stats.PaddingUsed = paddingNeeded

	// Calculate data density (bits of actual data vs total bits)
	if encodedSize > 0 {
		stats.DataDensity = float64(dataSize*8) / float64(encodedSize*8)
	}

	// Simple entropy calculation
	stats.Entropy = e.calculateEntropy(data)

	return stats
}

// validateEncodedFormat validates the format of the encoded result
func (e *Encoder) validateEncodedFormat(encoded string) *FormatValidation {
	validation := &FormatValidation{
		IsValid: true,
		Checks:  make([]string, 0),
	}

	if len(encoded) == 0 {
		validation.IsValid = false
		validation.Checks = append(validation.Checks, "empty_result")
		return validation
	}

	// Check for BL4 prefix
	if !e.charset.HasPrefix([]byte(encoded)) {
		validation.IsValid = false
		validation.Checks = append(validation.Checks, "missing_prefix")
	}

	// Validate Base85 characters
	stripped, _ := e.charset.StripPrefix([]byte(encoded))
	if err := e.charset.Validate(stripped); err != nil {
		validation.IsValid = false
		validation.Checks = append(validation.Checks, "invalid_characters")
		validation.Warnings = append(validation.Warnings, err.Error())
	}

	if validation.IsValid {
		validation.Checks = append(validation.Checks, "valid_format")
		validation.DataFormat = "bl4-base85"
		validation.Predictions = append(validation.Predictions, "ready_for_decoding")
	}

	// Check for reasonable encoded size
	if len(encoded) > MaxEncodedSerialLength {
		validation.IsValid = false
		validation.Checks = append(validation.Checks, "encoded_too_large")
		validation.Warnings = append(validation.Warnings,
			fmt.Sprintf("encoded length %d exceeds maximum %d", len(encoded), MaxEncodedSerialLength))
	}

	return validation
}

// optimizeEncoding attempts to optimize the encoding for size
func (e *Encoder) optimizeEncoding(data []byte, encoded string) *OptimizationInfo {
	info := &OptimizationInfo{
		OriginalSize:    len(data),
		OptimizedSize:   len(data), // Start with original size
		SpaceSaved:      0,
		CompressionGain: 0.0,
		Techniques:      make([]string, 0),
	}

	// For BL4, optimization is limited by the format requirements
	// This is a placeholder for future optimization techniques

	// Technique 1: Remove trailing zeros (if format allows)
	trimmedData := e.trimTrailingZeros(data)
	if len(trimmedData) < len(data) {
		info.SpaceSaved += len(data) - len(trimmedData)
		info.OptimizedSize = len(trimmedData)
		info.Techniques = append(info.Techniques, "trailing_zero_removal")
	}

	// Calculate compression gain
	if info.OptimizedSize > 0 {
		info.CompressionGain = float64(info.OriginalSize-info.OptimizedSize) / float64(info.OriginalSize) * 100
	}

	return info
}

// trimTrailingZeros removes trailing zero bytes if format allows
func (e *Encoder) trimTrailingZeros(data []byte) []byte {
	// For now, return original data as trailing zeros might be significant
	// This could be enhanced based on format specifications
	return data
}

// calculateEntropy calculates a simple entropy estimate
func (e *Encoder) calculateEntropy(data []byte) float64 {
	if len(data) == 0 {
		return 0.0
	}

	// Count byte frequencies
	freq := make(map[byte]int)
	for _, b := range data {
		freq[b]++
	}

	// Calculate entropy
	entropy := 0.0
	length := float64(len(data))

	for _, count := range freq {
		if count > 0 {
			probability := float64(count) / length
			entropy -= probability * encodeLog2(probability)
		}
	}

	return entropy / 8.0 // Normalize to 0-1 range
}

// BatchEncoder handles batch Base85 encoding operations
type BatchEncoder struct {
	encoder *Encoder
}

// NewBatchEncoder creates a new batch encoder
func NewBatchEncoder() *BatchEncoder {
	return &BatchEncoder{
		encoder: NewEncoder(),
	}
}

// EncodeBatch encodes multiple data arrays to Base85 codes
func (be *BatchEncoder) EncodeBatch(dataArray [][]byte) ([]*EncodeResult, error) {
	results := make([]*EncodeResult, len(dataArray))

	for i, data := range dataArray {
		result, err := be.encoder.EncodeWithOptions(data, &EncodeOptions{
			IncludeStatistics: true,
			ValidateFormat:     true,
		})

		results[i] = result
		if err != nil {
			// Continue processing other items even if one fails
			continue
		}
	}

	return results, nil
}

// GetBatchStatistics returns statistics for a batch encode operation
func (be *BatchEncoder) GetBatchStatistics(results []*EncodeResult) *BatchEncodeStatistics {
	stats := &BatchEncodeStatistics{
		Total:       len(results),
		Successful:  0,
		Failed:      0,
		TotalBytes:  0,
		AverageSize: 0.0,
	}

	for _, result := range results {
		if result.Error == nil {
			stats.Successful++
			stats.TotalBytes += result.ByteSize
		} else {
			stats.Failed++
		}
	}

	if stats.Successful > 0 {
		stats.AverageSize = float64(stats.TotalBytes) / float64(stats.Successful)
	}

	return stats
}

// BatchEncodeStatistics contains statistics for batch operations
type BatchEncodeStatistics struct {
	Total       int     `json:"total"`
	Successful  int     `json:"successful"`
	Failed      int     `json:"failed"`
	TotalBytes  int     `json:"total_bytes"`
	AverageSize float64 `json:"average_size"`
}

// Helper functions

// IsValidBL4Code checks if a string appears to be a valid BL4 item serial code
func IsValidBL4Code(code string) bool {
	encoder := NewEncoder()
	_ = encoder // Suppress unused variable warning

	decoder := NewDecoder()
	err := decoder.Validate(code)
	return err == nil
}

// GetEncodingLength estimates the encoded length for given data size
func GetEncodingLength(dataSize int) int {
	// Each 4 bytes becomes 5 Base85 characters + 2 character prefix
	chunks := (dataSize + 3) / 4 // Round up division
	return chunks*5 + 2
}

// Constants for encoding limits
const (
	MaxEncodedSize         = 1024 * 1024 // 1MB maximum data size
	MaxEncodedSerialLength = 10000       // Maximum reasonable serial code length
	MaxUint32Value         = 4294967295  // Maximum 32-bit unsigned value
)

// encodeLog2 calculates base-2 logarithm
func encodeLog2(x float64) float64 {
	const ln2 = 0.6931471805599453
	return -1.4426950408889634 * encodeLog2helper(x) // 1/ln(2) * ln(x)
}

func encodeLog2helper(x float64) float64 {
	// Simple approximation for ln(x)
	if x <= 0 {
		return 0
	}
	if x == 1 {
		return 0
	}
	// This is a very rough approximation
	result := 0.0
	for x > 1.0 {
		x /= 2.0
		result += 1.0
	}
	return result
}