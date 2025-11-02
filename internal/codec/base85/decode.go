package base85

import (
	"github.com/shawnvan/bl4/internal/codec/utils"
	"github.com/shawnvan/bl4/pkg/validator"
)

// Decoder handles Base85 decoding operations for BL4 item codes
// Based on reference implementation: https://github.com/Nicnl/borderlands4-serials
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
// Uses the reference implementation algorithm for maximum compatibility
func (d *Decoder) Decode(encoded string) ([]byte, error) {
	if encoded == "" {
		return nil, validator.NewValidationError(validator.ErrCodeEmptyInput, "input string is empty")
	}

	// Check for BL4 prefix - must be @U
	if len(encoded) < 2 || encoded[0] != '@' || encoded[1] != 'U' {
		return nil, validator.NewValidationError(validator.ErrCodeInvalidPrefix, "not a valid borderlands 4 item serial").
			WithField("input", encoded)
	}

	// Strip the @U prefix
	serial := encoded[2:]
	if len(serial) == 0 {
		return nil, validator.NewValidationError(validator.ErrCodeInvalidInput, "no data after @U prefix").
			WithField("input", encoded)
	}

	// Reference implementation: collect up to 5 valid Base85 characters at a time
	var (
		result = make([]byte, 0, (len(serial)*4/5)+4) // Preallocate with some extra space
		idx    = 0
		size   = len(serial)
	)

	for idx < size {
		var (
			v         = uint32(0)
			charCount = 0
		)

		// Collect up to 5 valid Base85 characters
		for idx < size && charCount < 5 {
			charCode := serial[idx]
			idx++

			// Check if character is valid in our charset
			if int(charCode) < 256 && d.charset.Decoder[charCode] <= 85 {
				v = v*85 + uint32(d.charset.Decoder[charCode])
				charCount++
			}
			// Reference implementation skips invalid characters silently
		}

		if charCount == 0 {
			break
		}

		// Handle padding for incomplete groups using padding character '~'
		if charCount < 5 {
			padding := 5 - charCount
			for i := 0; i < padding; i++ {
				v = v*85 + 84 // padding character value
			}
		}

		// Extract bytes - same for both full and partial groups
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

	// Mirror the bits in each byte using lookup table for performance
	// 76543210 -> 01234567
	for i := range result {
		result[i] = utils.Uint8Mirror[result[i]]
	}

	return result, nil
}

// decodeBase85 is the legacy method - now using reference implementation directly in Decode()
func (d *Decoder) decodeBase85(data []byte) ([]byte, error) {
	return d.Decode(string(data))
}