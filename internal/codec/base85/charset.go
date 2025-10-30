package base85

import (
	"errors"
	"unicode"
)

var (
	ErrInvalidBase85Char = errors.New("invalid Base85 character")
	ErrInvalidBase85Len  = errors.New("invalid Base85 data length")
	ErrInvalidPrefix     = errors.New("missing @U prefix")
)

// Charset defines the character set and lookup tables for Base85 encoding/decoding
// This implementation uses the Z85 character set (ZeroMQ)
type Charset struct {
	// Encoder maps values 0-84 to characters
	Encoder [85]byte

	// Decoder maps characters back to values 0-84
	// Invalid characters have value 255
	Decoder [256]byte

	// Prefix for item serial codes
	Prefix []byte
}

// NewCharset creates and initializes the Base85 charset with lookup tables
func NewCharset() *Charset {
	cs := &Charset{}

	// Z85 character set - 85 printable ASCII characters
	// These characters are chosen to be safe for various transport protocols
	z85Chars := []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ.-:+=^!/*?&<>()[]{}@%$#")

	// Initialize encoder
	copy(cs.Encoder[:], z85Chars)

	// Initialize decoder with invalid value marker
	for i := range cs.Decoder {
		cs.Decoder[i] = 255 // Mark as invalid
	}

	// Fill decoder with reverse mapping
	for i, ch := range z85Chars {
		cs.Decoder[ch] = byte(i)
	}

	// Set prefix for BL4 item serial codes
	cs.Prefix = []byte("@U")

	return cs
}

// Global charset instance for reuse across the codec
var GlobalCharset = NewCharset()

// Fast functions using global charset
func EncodeValueFast(value byte) byte {
	return GlobalCharset.Encoder[value]
}

func DecodeCharFast(ch byte) (byte, error) {
	decoded := GlobalCharset.Decoder[ch]
	if decoded == 255 {
		return 0, ErrInvalidBase85Char
	}
	return decoded, nil
}

// HasPrefix checks if data starts with the BL4 prefix
func (cs *Charset) HasPrefix(data []byte) bool {
	if len(data) < len(cs.Prefix) {
		return false
	}

	for i, b := range cs.Prefix {
		if data[i] != b {
			return false
		}
	}
	return true
}

// StripPrefix removes the BL4 prefix from data
func (cs *Charset) StripPrefix(data []byte) ([]byte, error) {
	if !cs.HasPrefix(data) {
		return nil, ErrInvalidPrefix
	}
	return data[len(cs.Prefix):], nil
}

// AddPrefix adds the BL4 prefix to data
func (cs *Charset) AddPrefix(data []byte) []byte {
	result := make([]byte, len(cs.Prefix)+len(data))
	copy(result, cs.Prefix)
	copy(result[len(cs.Prefix):], data)
	return result
}

// Validate checks if all characters in data are valid Base85 characters
func (cs *Charset) Validate(data []byte) error {
	for _, ch := range data {
		if cs.Decoder[ch] == 255 {
			return ErrInvalidBase85Char
		}
	}
	return nil
}

// ValidateString validates a string contains only valid Base85 characters
func (cs *Charset) ValidateString(s string) error {
	for _, r := range s {
		if r > unicode.MaxASCII {
			return ErrInvalidBase85Char
		}
		if cs.Decoder[byte(r)] == 255 {
			return ErrInvalidBase85Char
		}
	}
	return nil
}

// IsValidChar checks if a single character is valid in Base85
func (cs *Charset) IsValidChar(ch byte) bool {
	return cs.Decoder[ch] != 255
}

// EncodeQuad encodes 4 bytes to 5 Base85 characters
// This is the core encoding operation for Base85
func (cs *Charset) EncodeQuad(input [4]byte) [5]byte {
	// Convert 4 bytes to a 32-bit big-endian integer
	value := uint32(input[0])<<24 | uint32(input[1])<<16 | uint32(input[2])<<8 | uint32(input[3])

	// Extract base-85 digits
	digit0 := value / 52200625
	value -= digit0 * 52200625

	digit1 := value / 614125
	value -= digit1 * 614125

	digit2 := value / 7225
	value -= digit2 * 7225

	digit3 := value / 85
	value -= digit3 * 85

	digit4 := value

	return [5]byte{
		cs.Encoder[digit0],
		cs.Encoder[digit1],
		cs.Encoder[digit2],
		cs.Encoder[digit3],
		cs.Encoder[digit4],
	}
}

// DecodeQuad decodes 5 Base85 characters to 4 bytes
// This is the core decoding operation for Base85
func (cs *Charset) DecodeQuad(input [5]byte) ([4]byte, error) {
	// Decode characters to digits
	digit0 := cs.Decoder[input[0]]
	digit1 := cs.Decoder[input[1]]
	digit2 := cs.Decoder[input[2]]
	digit3 := cs.Decoder[input[3]]
	digit4 := cs.Decoder[input[4]]

	// Check for invalid characters
	if digit0 == 255 || digit1 == 255 || digit2 == 255 || digit3 == 255 || digit4 == 255 {
		return [4]byte{}, ErrInvalidBase85Char
	}

	// Reconstruct the 32-bit value
	value := uint32(digit0)*52200625 +
		uint32(digit1)*614125 +
		uint32(digit2)*7225 +
		uint32(digit3)*85 +
		uint32(digit4)

	// Extract bytes
	return [4]byte{
		byte(value >> 24),
		byte(value >> 16),
		byte(value >> 8),
		byte(value),
	}, nil
}

// EstimateDecodedLength estimates the decoded length from encoded length
// Base85 encodes 4 bytes to 5 characters, so roughly 4/5 ratio
func (cs *Charset) EstimateDecodedLength(encodedLen int) int {
	return (encodedLen * 4) / 5
}

// EstimateEncodedLength estimates the encoded length from data length
// Base85 encodes 4 bytes to 5 characters, so roughly 5/4 ratio
func (cs *Charset) EstimateEncodedLength(dataLen int) int {
	result := (dataLen * 5) / 4
	if (dataLen % 4) != 0 {
		result++ // May need one more character
	}
	return result
}

// GetCharacterSet returns the character set as a string
func (cs *Charset) GetCharacterSet() string {
	return string(cs.Encoder[:])
}

// IsPrintable checks if a character is printable ASCII
func (cs *Charset) IsPrintable(ch byte) bool {
	return ch >= 33 && ch <= 126 // Printable ASCII range
}

// FilterNonPrintable removes any non-printable characters from data
func (cs *Charset) FilterNonPrintable(data []byte) []byte {
	result := make([]byte, 0, len(data))
	for _, ch := range data {
		if cs.IsPrintable(ch) {
			result = append(result, ch)
		}
	}
	return result
}