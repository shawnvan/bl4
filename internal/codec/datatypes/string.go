package datatypes

import (
	"fmt"
	"unicode/utf8"

	"github.com/shawnvan/bl4/internal/codec/bitstream"
	"github.com/shawnvan/bl4/pkg/validator"
)

// STRING represents a null-terminated string data type
// Format: [length][characters][null_terminator] where length specifies the string length
type STRING struct {
	Value    string // The decoded string value
	ByteSize int    // Number of bytes in the string (excluding terminator)
}

// DecodeSTRING reads and decodes a STRING from the bitstream
// STRING format:
// - 8 bits: string length (0-255, specifies how many characters follow)
// - N bytes: string characters (UTF-8 encoded, where N is the length)
// - 8 bits: null terminator (should be 0)
func DecodeSTRING(reader *bitstream.Reader) (*STRING, error) {
	// Read the string length (8 bits)
	lengthRaw, err := reader.ReadBits(8)
	if err != nil {
		return nil, fmt.Errorf("failed to read STRING length: %w", err)
	}

	length := int(lengthRaw)

	// Validate length
	if length > MaxStringLength {
		return nil, validator.NewValidationError(validator.ErrCodeInvalidInput,
			fmt.Sprintf("invalid STRING length: %d", length)).
			WithField("length", length).
			WithDetail("max_length", MaxStringLength)
	}

	// Read the string characters
	characters := make([]byte, length)
	for i := 0; i < length; i++ {
		charByte, err := reader.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("failed to read STRING character at position %d: %w", i, err)
		}
		characters[i] = charByte
	}

	// Read the null terminator (8 bits)
	terminator, err := reader.ReadBits(8)
	if err != nil {
		return nil, fmt.Errorf("failed to read STRING terminator: %w", err)
	}

	if terminator != 0 {
		return nil, validator.NewValidationError(validator.ErrCodeInvalidString,
			fmt.Sprintf("invalid STRING terminator: %d", terminator)).
			WithField("terminator", terminator).
			WithDetail("expected_terminator", 0)
	}

	// Validate UTF-8 encoding
	if !utf8.Valid(characters) {
		return nil, validator.NewValidationError(validator.ErrCodeInvalidString,
			"STRING contains invalid UTF-8 encoding").
			WithField("bytes", characters)
	}

	return &STRING{
		Value:    string(characters),
		ByteSize: length,
	}, nil
}

// EncodeSTRING encodes a STRING to the bitstream writer
func EncodeSTRING(writer *bitstream.Writer, value string) error {
	if len(value) > MaxStringLength {
		return validator.NewValidationError(validator.ErrCodeInvalidString,
			fmt.Sprintf("string too long: %d bytes", len(value))).
			WithField("string_length", len(value)).
			WithDetail("max_length", MaxStringLength)
	}

	// Validate UTF-8 encoding
	if !utf8.ValidString(value) {
		return validator.NewValidationError(validator.ErrCodeInvalidString,
			"string contains invalid UTF-8 encoding").
			WithField("string", value)
	}

	// Convert string to bytes
	characters := []byte(value)
	length := len(characters)

	// Write string length (8 bits)
	if err := writer.WriteBits(uint64(length), 8); err != nil {
		return fmt.Errorf("failed to write STRING length: %w", err)
	}

	// Write the string characters
	for i, charByte := range characters {
		if err := writer.WriteByte(charByte); err != nil {
			return fmt.Errorf("failed to write STRING character at position %d: %w", i, err)
		}
	}

	// Write null terminator (8 bits)
	if err := writer.WriteBits(0, 8); err != nil {
		return fmt.Errorf("failed to write STRING terminator: %w", err)
	}

	return nil
}

// Constants for STRING data type
const (
	MaxStringLength = 255 // Maximum string length in bytes
)

// String returns a string representation of the STRING
func (s *STRING) String() string {
	if s == nil {
		return "STRING{nil}"
	}
	return fmt.Sprintf("STRING{%q, size: %d}", s.Value, s.ByteSize)
}

// Size returns the total size in bits of this STRING
func (s *STRING) Size() int {
	if s == nil {
		return 0
	}
	return 8 + (s.ByteSize * 8) + 8 // 8 bits for length + 8 bits per character + 8 bits for terminator
}

// Equals checks if two STRINGs are equal
func (s *STRING) Equals(other *STRING) bool {
	if other == nil {
		return false
	}
	return s.Value == other.Value && s.ByteSize == other.ByteSize
}

// Copy creates a deep copy of the STRING
func (s *STRING) Copy() *STRING {
	if s == nil {
		return nil
	}
	return &STRING{
		Value:    s.Value,
		ByteSize: s.ByteSize,
	}
}

// GetBytes returns the STRING as a byte slice (excluding terminator)
func (s *STRING) GetBytes() []byte {
	if s == nil {
		return []byte{}
	}
	return []byte(s.Value)
}

// SetFromBytes sets the STRING from a byte slice
func (s *STRING) SetFromBytes(data []byte) error {
	if s == nil {
		return validator.NewValidationError(validator.ErrCodeInvalidInput, "STRING is nil")
	}

	if len(data) > MaxStringLength {
		return validator.NewValidationError(validator.ErrCodeInvalidString,
			fmt.Sprintf("string too long: %d bytes", len(data))).
			WithField("data_length", len(data)).
			WithDetail("max_length", MaxStringLength)
	}

	// Validate UTF-8 encoding
	if !utf8.Valid(data) {
		return validator.NewValidationError(validator.ErrCodeInvalidString,
			"data contains invalid UTF-8 encoding").
			WithField("bytes", data)
	}

	s.Value = string(data)
	s.ByteSize = len(data)

	return nil
}

// IsEmpty returns true if the string is empty
func (s *STRING) IsEmpty() bool {
	return s == nil || s.Value == ""
}

// Length returns the length of the string in characters
func (s *STRING) Length() int {
	if s == nil {
		return 0
	}
	return len(s.Value)
}

// RuneCount returns the number of Unicode runes in the string
func (s *STRING) RuneCount() int {
	if s == nil {
		return 0
	}
	return utf8.RuneCountInString(s.Value)
}

// Contains checks if the string contains the specified substring
func (s *STRING) Contains(substring string) bool {
	if s == nil {
		return false
	}
	return len(substring) > 0 && findSubstring(s.Value, substring)
}

// StartsWith checks if the string starts with the specified prefix
func (s *STRING) StartsWith(prefix string) bool {
	if s == nil {
		return false
	}
	return len(s.Value) >= len(prefix) && s.Value[:len(prefix)] == prefix
}

// EndsWith checks if the string ends with the specified suffix
func (s *STRING) EndsWith(suffix string) bool {
	if s == nil {
		return false
	}
	return len(s.Value) >= len(suffix) && s.Value[len(s.Value)-len(suffix):] == suffix
}

// ToUpper returns an uppercase version of the string
func (s *STRING) ToUpper() *STRING {
	if s == nil {
		return nil
	}
	return &STRING{
		Value:    toUpper(s.Value),
		ByteSize: s.ByteSize,
	}
}

// ToLower returns a lowercase version of the string
func (s *STRING) ToLower() *STRING {
	if s == nil {
		return nil
	}
	return &STRING{
		Value:    toLower(s.Value),
		ByteSize: s.ByteSize,
	}
}

// Trim returns a version of the string with leading and trailing whitespace removed
func (s *STRING) Trim() *STRING {
	if s == nil {
		return nil
	}
	trimmed := trimString(s.Value)
	return &STRING{
		Value:    trimmed,
		ByteSize: len(trimmed),
	}
}

// NewSTRING creates a new STRING from a string value
func NewSTRING(value string) *STRING {
	return &STRING{
		Value:    value,
		ByteSize: len(value),
	}
}

// NewSTRINGFromBytes creates a new STRING from a byte slice
func NewSTRINGFromBytes(data []byte) (*STRING, error) {
	s := &STRING{}
	err := s.SetFromBytes(data)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// STRINGCollection represents a collection of STRING values
type STRINGCollection struct {
	Strings []*STRING
}

// NewSTRINGCollection creates an empty STRING collection
func NewSTRINGCollection() *STRINGCollection {
	return &STRINGCollection{
		Strings: make([]*STRING, 0),
	}
}

// Add adds a STRING to the collection
func (sc *STRINGCollection) Add(str *STRING) {
	if sc.Strings == nil {
		sc.Strings = make([]*STRING, 0)
	}
	sc.Strings = append(sc.Strings, str)
}

// Get returns a STRING by index
func (sc *STRINGCollection) Get(index int) *STRING {
	if index < 0 || index >= len(sc.Strings) {
		return nil
	}
	return sc.Strings[index]
}

// Size returns the number of STRINGs in the collection
func (sc *STRINGCollection) Size() int {
	return len(sc.Strings)
}

// TotalBitSize returns the total size in bits of all STRINGs
func (sc *STRINGCollection) TotalBitSize() int {
	total := 0
	for _, str := range sc.Strings {
		if str != nil {
			total += str.Size()
		}
	}
	return total
}

// TotalByteSize returns the total byte size of all STRINGs (excluding terminators)
func (sc *STRINGCollection) TotalByteSize() int {
	total := 0
	for _, str := range sc.Strings {
		if str != nil {
			total += str.ByteSize
		}
	}
	return total
}

// Validate validates all STRINGs in the collection
func (sc *STRINGCollection) Validate() error {
	for i, str := range sc.Strings {
		if str == nil {
			return validator.NewValidationError(validator.ErrCodeInvalidInput,
				fmt.Sprintf("STRING at index %d is nil", i)).
				WithField("index", i)
		}

		if str.ByteSize > MaxStringLength {
			return validator.NewValidationError(validator.ErrCodeInvalidString,
				fmt.Sprintf("invalid string length at index %d: %d", i, str.ByteSize)).
				WithField("index", i).
				WithField("string_length", str.ByteSize)
		}
	}
	return nil
}

// Join concatenates all strings in the collection with a separator
func (sc *STRINGCollection) Join(separator string) string {
	if sc.Size() == 0 {
		return ""
	}

	var result string
	for i, str := range sc.Strings {
		if i > 0 {
			result += separator
		}
		if str != nil {
			result += str.Value
		}
	}
	return result
}

// FilterEmpty removes empty strings from the collection
func (sc *STRINGCollection) FilterEmpty() *STRINGCollection {
	filtered := NewSTRINGCollection()
	for _, str := range sc.Strings {
		if str != nil && !str.IsEmpty() {
			filtered.Add(str)
		}
	}
	return filtered
}

// String returns a string representation of the collection
func (sc *STRINGCollection) String() string {
	if sc.Size() == 0 {
		return "STRINGCollection{empty}"
	}

	var strings []string
	for _, str := range sc.Strings {
		if str != nil {
			strings = append(strings, fmt.Sprintf("%q", str.Value))
		} else {
			strings = append(strings, "nil")
		}
	}

	return fmt.Sprintf("STRINGCollection{%s}", fmt.Sprintf("%v", strings))
}

// Helper functions for string operations

// toUpper converts a string to uppercase (simple implementation)
func toUpper(s string) string {
	result := make([]rune, len([]rune(s)))
	for i, r := range []rune(s) {
		if r >= 'a' && r <= 'z' {
			result[i] = r - ('a' - 'A')
		} else {
			result[i] = r
		}
	}
	return string(result)
}

// toLower converts a string to lowercase (simple implementation)
func toLower(s string) string {
	result := make([]rune, len([]rune(s)))
	for i, r := range []rune(s) {
		if r >= 'A' && r <= 'Z' {
			result[i] = r + ('a' - 'A')
		} else {
			result[i] = r
		}
	}
	return string(result)
}

// trimString removes leading and trailing whitespace (simple implementation)
func trimString(s string) string {
	// Find first non-whitespace character
	start := 0
	for start < len(s) && isWhitespace(rune(s[start])) {
		start++
	}

	// Find last non-whitespace character
	end := len(s) - 1
	for end >= start && isWhitespace(rune(s[end])) {
		end--
	}

	if start > end {
		return ""
	}

	return s[start : end+1]
}

// isWhitespace checks if a rune is whitespace
func isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

// findSubstring checks if substring exists in string (simple implementation)
func findSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(substr) > len(s) {
		return false
	}

	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Helper functions for common game-related STRING operations

// EncodeItemName encodes an item name as a STRING
func EncodeItemName(name string) (*STRING, error) {
	if len(name) == 0 {
		return nil, validator.NewValidationError(validator.ErrCodeInvalidInput,
			"item name cannot be empty").
			WithField("name", name)
	}

	if len(name) > MaxItemNameLength {
		return nil, validator.NewValidationError(validator.ErrCodeInvalidInput,
			fmt.Sprintf("item name too long: %d characters", len(name))).
			WithField("name_length", len(name)).
			WithDetail("max_length", MaxItemNameLength)
	}

	// Validate name contains only allowed characters
	if !isValidItemName(name) {
		return nil, validator.NewValidationError(validator.ErrCodeInvalidInput,
			"item name contains invalid characters").
			WithField("name", name)
	}

	return NewSTRING(name), nil
}

// DecodeItemName decodes a STRING as an item name
func DecodeItemName(s *STRING) (string, error) {
	if s == nil {
		return "", validator.NewValidationError(validator.ErrCodeInvalidInput, "STRING is nil")
	}

	if s.IsEmpty() {
		return "", validator.NewValidationError(validator.ErrCodeInvalidInput,
			"item name cannot be empty").
			WithField("string", "empty")
	}

	if s.Length() > MaxItemNameLength {
		return "", validator.NewValidationError(validator.ErrCodeInvalidInput,
			fmt.Sprintf("item name too long: %d characters", s.Length())).
			WithField("name_length", s.Length()).
			WithDetail("max_length", MaxItemNameLength)
	}

	if !isValidItemName(s.Value) {
		return "", validator.NewValidationError(validator.ErrCodeInvalidInput,
			"item name contains invalid characters").
			WithField("name", s.Value)
	}

	return s.Value, nil
}

// Constants for item naming
const (
	MaxItemNameLength = 64 // Maximum item name length in characters
)

// isValidItemName checks if an item name contains only valid characters
func isValidItemName(name string) bool {
	if len(name) == 0 {
		return false
	}

	for _, r := range name {
		// Allow letters, numbers, spaces, hyphens, and apostrophes
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') ||
			r == ' ' || r == '-' || r == '\'') {
			return false
		}
	}

	return true
}

// EncodeDescription encodes an item description as a STRING
func EncodeDescription(description string) (*STRING, error) {
	if len(description) > MaxDescriptionLength {
		return nil, validator.NewValidationError(validator.ErrCodeInvalidInput,
			fmt.Sprintf("description too long: %d characters", len(description))).
			WithField("description_length", len(description)).
			WithDetail("max_length", MaxDescriptionLength)
	}

	return NewSTRING(description), nil
}

// DecodeDescription decodes a STRING as an item description
func DecodeDescription(s *STRING) (string, error) {
	if s == nil {
		return "", nil // Description is optional
	}

	if s.Length() > MaxDescriptionLength {
		return "", validator.NewValidationError(validator.ErrCodeInvalidInput,
			fmt.Sprintf("description too long: %d characters", s.Length())).
			WithField("description_length", s.Length()).
			WithDetail("max_length", MaxDescriptionLength)
	}

	return s.Value, nil
}

// Constants for item descriptions
const (
	MaxDescriptionLength = 256 // Maximum description length in characters
)