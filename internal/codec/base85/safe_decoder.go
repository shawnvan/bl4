package base85

import (
	"github.com/shawnvan/bl4/pkg/validator"
)

// SafeDecoder implements enhanced safety checks for Base85 decoding
type SafeDecoder struct {
	*Decoder
}

// NewSafeDecoder creates a new safe Base85 decoder
func NewSafeDecoder() *SafeDecoder {
	return &SafeDecoder{
		Decoder: NewDecoder(),
	}
}

// SafeDecode performs safe Base85 decoding with enhanced validation
func (d *SafeDecoder) SafeDecode(encoded string) ([]byte, error) {
	if encoded == "" {
		return nil, validator.NewValidationError(validator.ErrCodeEmptyInput, "input string is empty")
	}

	// Enhanced length validation
	if len(encoded) > 10000 { // Reasonable upper limit
		return nil, validator.NewValidationError(validator.ErrCodeInvalidLength, "input string too long").
			WithField("length", len(encoded)).
			WithDetail("max_length", 10000)
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

	// Validate characters before processing
	for i, char := range serial {
		if char < 0 || char > 255 {
			return nil, validator.NewValidationError(validator.ErrCodeInvalidInput, 
				"invalid character in serial code").
				WithField("position", i).
				WithField("character", char)
		}
	}

	// Use the original decoder with enhanced safety
	return d.Decode(encoded)
}

// validateCharsetBounds safely checks character bounds
func (d *SafeDecoder) validateCharsetBounds(charCode byte) bool {
	if int(charCode) >= len(d.charset.Decoder) {
		return false
	}
	
	value := d.charset.Decoder[charCode]
	return value <= 85
}