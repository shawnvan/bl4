package base85

import (
	"fmt"

	"github.com/shawnvan/bl4/pkg/validator"
)

// Decoder handles Base85 decoding operations for BL4 item codes
type Decoder struct {
	charset *Charset
}

// NewDecoder creates a new Base85 decoder
func NewDecoder() *Decoder {
	return &Decoder{
		charset: GlobalCharset,
	}
}

// Decode decodes a Base85 encoded item serial code to raw bytes
func (d *Decoder) Decode(encoded string) ([]byte, error) {
	if encoded == "" {
		return nil, validator.NewValidationError(validator.ErrCodeEmptyInput, "input string is empty")
	}

	// Check for BL4 prefix
	encodedBytes := []byte(encoded)
	if !d.charset.HasPrefix(encodedBytes) {
		return nil, validator.NewValidationError(validator.ErrCodeInvalidPrefix, "missing @U prefix").
			WithField("input", encoded)
	}

	// Strip the prefix
	stripped, err := d.charset.StripPrefix(encodedBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to strip prefix: %w", err)
	}

	if len(stripped) == 0 {
		return nil, validator.NewValidationError(validator.ErrCodeInvalidInput, "no data after prefix").
			WithField("input", encoded)
	}

	// Remove this validation as we'll handle invalid characters in decode loop
	// Reference implementation skips invalid characters during decoding

	// Decode from Base85 to bytes
	decoded, err := d.decodeBase85(stripped)
	if err != nil {
		return nil, fmt.Errorf("Base85 decode failed: %w", err)
	}

	// Apply byte mirroring (BL4 specific)
	mirrored := d.mirrorBytes(decoded)

	return mirrored, nil
}

// decodeBase85 performs the actual Base85 decoding (matches reference implementation)
func (d *Decoder) decodeBase85(encoded []byte) ([]byte, error) {
	if len(encoded) == 0 {
		return []byte{}, nil
	}

	// Match reference implementation logic
	var (
		result = make([]byte, 0, (len(encoded)*4/5)+4)
		idx    = 0
		size   = len(encoded)
	)

	for idx < size {
		var (
			v         = uint32(0)
			charCount = 0
		)

		// Collect up to 5 valid Base85 characters
		for idx < size && charCount < 5 {
			charCode := encoded[idx]
			idx++

			decoded := d.charset.DecodeByte(charCode)
			if decoded != 255 { // Only process valid characters
				v = v*85 + uint32(decoded)
				charCount++
			}
		}

		if charCount == 0 {
			break
		}

		// Handle padding for incomplete groups (using '~' character as in reference)
		if charCount < 5 {
			padding := 5 - charCount
			for i := 0; i < padding; i++ {
				paddingChar := byte('~')
				decoded := d.charset.DecodeByte(paddingChar)
				if decoded == 255 {
					return nil, fmt.Errorf("invalid padding character at position %d", i)
				}
				v = v*85 + uint32(decoded)
			}
		}

		// Extract bytes - same logic as reference implementation
		byteCount := 4
		if charCount < 5 {
			byteCount = charCount - 1
		}

		if byteCount >= 1 {
			result = append(result, byte((v>>24)&0xFF))
		}
		if byteCount >= 2 {
			result = append(result, byte((v>>16)&0xFF))
		}
		if byteCount >= 3 {
			result = append(result, byte((v>>8)&0xFF))
		}
		if byteCount >= 4 {
			result = append(result, byte((v>>0)&0xFF))
		}
	}

	return result, nil
}

// Validate validates a Base85 encoded string without decoding it
func (d *Decoder) Validate(encoded string) error {
	if encoded == "" {
		return validator.NewValidationError(validator.ErrCodeEmptyInput, "input string is empty")
	}

	// Check prefix
	encodedBytes := []byte(encoded)
	if !d.charset.HasPrefix(encodedBytes) {
		return validator.NewValidationError(validator.ErrCodeInvalidPrefix, "missing @U prefix").
			WithField("input", encoded)
	}

	// Strip prefix and validate characters
	stripped, err := d.charset.StripPrefix(encodedBytes)
	if err != nil {
		return fmt.Errorf("failed to strip prefix: %w", err)
	}

	if len(stripped) == 0 {
		return validator.NewValidationError(validator.ErrCodeInvalidInput, "no data after prefix")
	}

	return d.charset.Validate(stripped)
}

// DecodeWithOptions decodes with additional options
func (d *Decoder) DecodeWithOptions(encoded string, options *DecodeOptions) (*DecodeResult, error) {
	result := &DecodeResult{
		Input: encoded,
	}

	// Validate input
	if err := d.Validate(encoded); err != nil {
		result.Error = err
		return result, err
	}

	// Perform decoding
	decoded, err := d.Decode(encoded)
	if err != nil {
		result.Error = err
		return result, err
	}

	result.Data = decoded
	result.ByteSize = len(decoded)
	result.CharCount = len(encoded) - 2 // Subtract prefix length

	// Add additional information if requested
	if options != nil {
		if options.IncludeStatistics {
			result.Statistics = d.calculateStatistics(encoded, decoded)
		}

		if options.ValidateFormat {
			result.FormatValidation = d.validateFormat(decoded)
		}
	}

	return result, nil
}

// DecodeOptions contains options for decoding operations
type DecodeOptions struct {
	IncludeStatistics bool `json:"include_statistics,omitempty"`
	ValidateFormat     bool `json:"validate_format,omitempty"`
}

// DecodeResult contains the result of a decode operation
type DecodeResult struct {
	Input            string                 `json:"input"`
	Data             []byte                 `json:"data,omitempty"`
	ByteSize         int                    `json:"byte_size,omitempty"`
	CharCount        int                    `json:"char_count,omitempty"`
	Statistics       *DecodeStatistics      `json:"statistics,omitempty"`
	FormatValidation *FormatValidation      `json:"format_validation,omitempty"`
	Error            error                  `json:"error,omitempty"`
}

// DecodeStatistics contains statistics about the decode operation
type DecodeStatistics struct {
	CompressionRatio float64 `json:"compression_ratio"` // decoded_size / encoded_size
	DataDensity      float64 `json:"data_density"`       // actual_data_bits / total_bits
	Entropy          float64 `json:"entropy"`             // Data entropy estimate
	CharacterType    string  `json:"character_type"`     // Type of input data
}

// FormatValidation contains validation results for the decoded data
type FormatValidation struct {
	IsValid     bool     `json:"is_valid"`
	Checks      []string `json:"checks"`
	Warnings    []string `json:"warnings,omitempty"`
	DataFormat  string   `json:"data_format,omitempty"`
	Predictions []string `json:"predictions,omitempty"`
}

// calculateStatistics calculates statistics about the decode operation
func (d *Decoder) calculateStatistics(encoded string, decoded []byte) *DecodeStatistics {
	encodedSize := len(encoded)
	decodedSize := len(decoded)

	stats := &DecodeStatistics{
		CompressionRatio: float64(decodedSize) / float64(encodedSize),
		DataDensity:      0.0,
		Entropy:          0.0,
		CharacterType:    d.determineCharacterType(encoded),
	}

	// Calculate data density (bits of actual data vs total bits)
	if decodedSize > 0 {
		stats.DataDensity = float64(decodedSize*8) / float64(encodedSize*8)
	}

	// Simple entropy calculation
	stats.Entropy = d.calculateEntropy(decoded)

	return stats
}

// determineCharacterType determines the type of input characters
func (d *Decoder) determineCharacterType(encoded string) string {
	// Remove prefix for analysis
	stripped, _ := d.charset.StripPrefix([]byte(encoded))
	if len(stripped) == 0 {
		return "empty"
	}

	// Count character types
	digitCount := 0
	lowerCount := 0
	upperCount := 0
	symbolCount := 0

	for _, ch := range stripped {
		switch {
		case ch >= '0' && ch <= '9':
			digitCount++
		case ch >= 'a' && ch <= 'z':
			lowerCount++
		case ch >= 'A' && ch <= 'Z':
			upperCount++
		default:
			symbolCount++
		}
	}

	total := len(stripped)
	if digitCount == total {
		return "numeric"
	} else if lowerCount == total {
		return "lowercase"
	} else if upperCount == total {
		return "uppercase"
	} else if symbolCount == total {
		return "symbols"
	} else {
		return "mixed"
	}
}

// calculateEntropy calculates a simple entropy estimate
func (d *Decoder) calculateEntropy(data []byte) float64 {
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
			entropy -= probability * log2(probability)
		}
	}

	return entropy / 8.0 // Normalize to 0-1 range
}

// log2 calculates base-2 logarithm
func log2(x float64) float64 {
	const ln2 = 0.6931471805599453
	return -1.4426950408889634 * log2helper(x) // 1/ln(2) * ln(x)
}

func log2helper(x float64) float64 {
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

// validateFormat validates the format of decoded data
func (d *Decoder) validateFormat(decoded []byte) *FormatValidation {
	validation := &FormatValidation{
		IsValid: true,
		Checks:  make([]string, 0),
	}

	if len(decoded) == 0 {
		validation.IsValid = false
		validation.Checks = append(validation.Checks, "empty_data")
		return validation
	}

	// Check for common patterns
	if d.looksLikeBitstream(decoded) {
		validation.DataFormat = "bitstream"
		validation.Checks = append(validation.Checks, "bitstream_pattern_detected")
		validation.Predictions = append(validation.Predictions, "contains_tokenized_data")
	} else if d.looksLikeJSON(decoded) {
		validation.DataFormat = "json"
		validation.Checks = append(validation.Checks, "json_pattern_detected")
		validation.Predictions = append(validation.Predictions, "structured_data")
	} else if d.looksLikeCSV(decoded) {
		validation.DataFormat = "csv"
		validation.Checks = append(validation.Checks, "csv_pattern_detected")
		validation.Predictions = append(validation.Predictions, "tabular_data")
	} else {
		validation.DataFormat = "binary"
		validation.Checks = append(validation.Checks, "binary_data")
	}

	// Check for reasonable data size
	if len(decoded) > MaxDecodedSize {
		validation.IsValid = false
		validation.Checks = append(validation.Checks, "data_too_large")
		validation.Warnings = append(validation.Warnings, fmt.Sprintf("data size %d exceeds maximum %d", len(decoded), MaxDecodedSize))
	}

	// Check for null bytes (might indicate string data)
	nullByteCount := 0
	for _, b := range decoded {
		if b == 0 {
			nullByteCount++
		}
	}

	if nullByteCount > 0 {
		validation.Checks = append(validation.Checks, "contains_null_bytes")
		if nullByteCount > len(decoded)/2 {
			validation.Predictions = append(validation.Predictions, "likely_string_data_with_terminators")
		}
	}

	return validation
}

// looksLikeBitstream checks if data looks like a tokenized bitstream
func (d *Decoder) looksLikeBitstream(data []byte) bool {
	if len(data) < 4 {
		return false
	}

	// Look for common bitstream patterns
	// Check if data starts with a valid token header (3 bits)
	firstByte := data[0]
	header := firstByte & 0xE0 // Top 3 bits

	// Valid headers are 0-6 (binary 000-110)
	if header > 0xC0 { // 11000000
		return false
	}

	// Additional heuristics could be added here
	return true
}

// looksLikeJSON checks if data looks like JSON
func (d *Decoder) looksLikeJSON(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	// Look for JSON patterns
	firstChar := data[0]
	lastChar := data[len(data)-1]

	return (firstChar == '{' || firstChar == '[') && (lastChar == '}' || lastChar == ']')
}

// looksLikeCSV checks if data looks like CSV
func (d *Decoder) looksLikeCSV(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	// Look for CSV patterns (commas, numbers, etc.)
	commaCount := 0
	digitCount := 0

	for _, b := range data {
		if b == ',' {
			commaCount++
		} else if b >= '0' && b <= '9' {
			digitCount++
		}
	}

	// If we have multiple commas and primarily digits, likely CSV
	return commaCount > 0 && digitCount > commaCount
}

// Constants for decoding limits
const (
	MaxDecodedSize = 1024 * 1024 // 1MB maximum decoded size
)

// BatchDecoder handles batch Base85 decoding operations
type BatchDecoder struct {
	decoder *Decoder
}

// NewBatchDecoder creates a new batch decoder
func NewBatchDecoder() *BatchDecoder {
	return &BatchDecoder{
		decoder: NewDecoder(),
	}
}

// DecodeBatch decodes multiple Base85 encoded strings
func (bd *BatchDecoder) DecodeBatch(encoded []string) ([]*DecodeResult, error) {
	results := make([]*DecodeResult, len(encoded))

	for i, code := range encoded {
		result, err := bd.decoder.DecodeWithOptions(code, &DecodeOptions{
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

// GetBatchStatistics returns statistics for a batch decode operation
func (bd *BatchDecoder) GetBatchStatistics(results []*DecodeResult) *BatchStatistics {
	stats := &BatchStatistics{
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

// BatchStatistics contains statistics for batch operations
type BatchStatistics struct {
	Total       int     `json:"total"`
	Successful  int     `json:"successful"`
	Failed      int     `json:"failed"`
	TotalBytes  int     `json:"total_bytes"`
	AverageSize float64 `json:"average_size"`
}

// Helper functions

// IsBL4Code checks if a string appears to be a BL4 item serial code
func IsBL4Code(code string) bool {
	decoder := NewDecoder()
	_ = decoder // Suppress unused variable warning
	err := decoder.Validate(code)
	return err == nil
}

// ExtractPrefix extracts the prefix from a BL4 code
func ExtractPrefix(code string) (string, error) {
	decoder := NewDecoder()
	_ = decoder // Suppress unused variable warning
	if len(code) < 2 {
		return "", validator.NewValidationError(validator.ErrCodeInvalidInput, "code too short")
	}

	prefix := code[:2]
	if prefix != "@U" {
		return "", validator.NewValidationError(validator.ErrCodeInvalidPrefix, "invalid prefix").
			WithField("prefix", prefix)
	}

	return prefix, nil
}

// StripPrefix removes the BL4 prefix from a code
func StripPrefix(code string) (string, error) {
	decoder := NewDecoder()
	codeBytes := []byte(code)

	stripped, err := decoder.charset.StripPrefix(codeBytes)
	if err != nil {
		return "", err
	}

	return string(stripped), nil
}

// mirrorBytes applies bit mirroring to decoded bytes (BL4 specific transformation)
func (d *Decoder) mirrorBytes(data []byte) []byte {
	if len(data) == 0 {
		return data
	}

	mirrored := make([]byte, len(data))
	for i, b := range data {
		mirrored[i] = reverseBits(b)
	}

	return mirrored
}

// reverseBits reverses the bits in a byte (bit mirroring)
func reverseBits(b byte) byte {
	// Bit reversal table for fast lookup
	reverseTable := [256]byte{
		0x00, 0x80, 0x40, 0xC0, 0x20, 0xA0, 0x60, 0xE0, 0x10, 0x90, 0x50, 0xD0, 0x30, 0xB0, 0x70, 0xF0,
		0x08, 0x88, 0x48, 0xC8, 0x28, 0xA8, 0x68, 0xE8, 0x18, 0x98, 0x58, 0xD8, 0x38, 0xB8, 0x78, 0xF8,
		0x04, 0x84, 0x44, 0xC4, 0x24, 0xA4, 0x64, 0xE4, 0x14, 0x94, 0x54, 0xD4, 0x34, 0xB4, 0x74, 0xF4,
		0x0C, 0x8C, 0x4C, 0xCC, 0x2C, 0xAC, 0x6C, 0xEC, 0x1C, 0x9C, 0x5C, 0xDC, 0x3C, 0xBC, 0x7C, 0xFC,
		0x02, 0x82, 0x42, 0xC2, 0x22, 0xA2, 0x62, 0xE2, 0x12, 0x92, 0x52, 0xD2, 0x32, 0xB2, 0x72, 0xF2,
		0x0A, 0x8A, 0x4A, 0xCA, 0x2A, 0xAA, 0x6A, 0xEA, 0x1A, 0x9A, 0x5A, 0xDA, 0x3A, 0xBA, 0x7A, 0xFA,
		0x06, 0x86, 0x46, 0xC6, 0x26, 0xA6, 0x66, 0xE6, 0x16, 0x96, 0x56, 0xD6, 0x36, 0xB6, 0x76, 0xF6,
		0x0E, 0x8E, 0x4E, 0xCE, 0x2E, 0xAE, 0x6E, 0xEE, 0x1E, 0x9E, 0x5E, 0xDE, 0x3E, 0xBE, 0x7E, 0xFE,
		0x01, 0x81, 0x41, 0xC1, 0x21, 0xA1, 0x61, 0xE1, 0x11, 0x91, 0x51, 0xD1, 0x31, 0xB1, 0x71, 0xF1,
		0x09, 0x89, 0x49, 0xC9, 0x29, 0xA9, 0x69, 0xE9, 0x19, 0x99, 0x59, 0xD9, 0x39, 0xB9, 0x79, 0xF9,
		0x05, 0x85, 0x45, 0xC5, 0x25, 0xA5, 0x65, 0xE5, 0x15, 0x95, 0x55, 0xD5, 0x35, 0xB5, 0x75, 0xF5,
		0x0D, 0x8D, 0x4D, 0xCD, 0x2D, 0xAD, 0x6D, 0xED, 0x1D, 0x9D, 0x5D, 0xDD, 0x3D, 0xBD, 0x7D, 0xFD,
		0x03, 0x83, 0x43, 0xC3, 0x23, 0xA3, 0x63, 0xE3, 0x13, 0x93, 0x53, 0xD3, 0x33, 0xB3, 0x73, 0xF3,
		0x0B, 0x8B, 0x4B, 0xCB, 0x2B, 0xAB, 0x6B, 0xEB, 0x1B, 0x9B, 0x5B, 0xDB, 0x3B, 0xBB, 0x7B, 0xFB,
		0x07, 0x87, 0x47, 0xC7, 0x27, 0xA7, 0x67, 0xE7, 0x17, 0x97, 0x57, 0xD7, 0x37, 0xB7, 0x77, 0xF7,
		0x0F, 0x8F, 0x4F, 0xCF, 0x2F, 0xAF, 0x6F, 0xEF, 0x1F, 0x9F, 0x5F, 0xDF, 0x3F, 0xBF, 0x7F, 0xFF,
	}

	return reverseTable[b]
}