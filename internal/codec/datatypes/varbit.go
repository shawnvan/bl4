package datatypes

import (
	"fmt"
	"strings"

	"github.com/shawnvan/bl4/internal/codec/bitstream"
	"github.com/shawnvan/bl4/pkg/validator"
)

// VARBIT represents a variable-length bitfield data type
// Format: [bit_count][bits] where bit_count determines how many bits follow for the bitfield
type VARBIT struct {
	Bits     []bool // The decoded bitfield (true = 1, false = 0)
	BitCount int    // Number of bits in the bitfield
}

// DecodeVARBIT reads and decodes a VARBIT from the bitstream
// VARBIT format:
// - 6 bits: bit count (1-32, specifies how many bits follow for the bitfield)
// - N bits: bitfield value (where N is the bit count)
func DecodeVARBIT(reader *bitstream.Reader) (*VARBIT, error) {
	// Read the bit count (6 bits)
	bitCountRaw, err := reader.ReadBits(6)
	if err != nil {
		return nil, fmt.Errorf("failed to read VARBIT bit count: %w", err)
	}

	bitCount := int(bitCountRaw)

	// Validate bit count
	if bitCount == 0 || bitCount > 32 {
		return nil, validator.NewValidationError(validator.ErrCodeInvalidBitCount,
			fmt.Sprintf("invalid VARBIT bit count: %d", bitCount)).
			WithField("bit_count", bitCount).
			WithDetail("valid_range", "1-32")
	}

	// Read the actual bits
	bits := make([]bool, bitCount)
	for i := 0; i < bitCount; i++ {
		bit, err := reader.ReadBit()
		if err != nil {
			return nil, fmt.Errorf("failed to read VARBIT bit at position %d: %w", i, err)
		}
		bits[i] = bit
	}

	return &VARBIT{
		Bits:     bits,
		BitCount: bitCount,
	}, nil
}

// EncodeVARBIT encodes a VARBIT to the bitstream writer
func EncodeVARBIT(writer *bitstream.Writer, bits []bool) error {
	bitCount := len(bits)

	// Validate bit count
	if bitCount == 0 || bitCount > 32 {
		return validator.NewValidationError(validator.ErrCodeInvalidBitCount,
			fmt.Sprintf("invalid VARBIT bit count: %d", bitCount)).
			WithField("bit_count", bitCount).
			WithDetail("valid_range", "1-32")
	}

	// Write bit count (6 bits)
	if err := writer.WriteBits(uint64(bitCount), 6); err != nil {
		return fmt.Errorf("failed to write VARBIT bit count: %w", err)
	}

	// Write the bits
	for i, bit := range bits {
		var value uint64 = 0
		if bit {
			value = 1
		}
		if err := writer.WriteBits(value, 1); err != nil {
			return fmt.Errorf("failed to write VARBIT bit at position %d: %w", i, err)
		}
	}

	return nil
}

// String returns a string representation of the VARBIT
func (v *VARBIT) String() string {
	if v == nil || len(v.Bits) == 0 {
		return "VARBIT{empty}"
	}

	var builder strings.Builder
	builder.WriteString("VARBIT{")

	for i, bit := range v.Bits {
		if i > 0 && i%8 == 0 {
			builder.WriteString(" ")
		}
		if bit {
			builder.WriteString("1")
		} else {
			builder.WriteString("0")
		}
	}

	builder.WriteString(fmt.Sprintf("} (%d bits)", v.BitCount))
	return builder.String()
}

// Size returns the total size in bits of this VARBIT
func (v *VARBIT) Size() int {
	if v == nil {
		return 0
	}
	return 6 + v.BitCount // 6 bits for count + data bits
}

// Equals checks if two VARBITs are equal
func (v *VARBIT) Equals(other *VARBIT) bool {
	if other == nil {
		return false
	}

	if v.BitCount != other.BitCount {
		return false
	}

	if len(v.Bits) != len(other.Bits) {
		return false
	}

	for i := range v.Bits {
		if v.Bits[i] != other.Bits[i] {
			return false
		}
	}

	return true
}

// Copy creates a deep copy of the VARBIT
func (v *VARBIT) Copy() *VARBIT {
	if v == nil {
		return nil
	}

	copiedBits := make([]bool, len(v.Bits))
	copy(copiedBits, v.Bits)

	return &VARBIT{
		Bits:     copiedBits,
		BitCount: v.BitCount,
	}
}

// GetBytes returns the VARBIT as a byte slice (big-endian)
func (v *VARBIT) GetBytes() []byte {
	if v == nil || len(v.Bits) == 0 {
		return []byte{0}
	}

	byteCount := (v.BitCount + 7) / 8
	result := make([]byte, byteCount)

	for i, bit := range v.Bits {
		if bit {
			byteIndex := i / 8
			bitIndex := 7 - (i % 8) // MSB first
			result[byteIndex] |= 1 << bitIndex
		}
	}

	return result
}

// SetFromBytes sets the VARBIT from a byte slice
func (v *VARBIT) SetFromBytes(data []byte, bitCount int) error {
	if v == nil {
		return validator.NewValidationError(validator.ErrCodeInvalidInput, "VARBIT is nil")
	}

	if bitCount <= 0 || bitCount > 32 {
		return validator.NewValidationError(validator.ErrCodeInvalidBitCount,
			fmt.Sprintf("invalid bit count: %d", bitCount)).
			WithField("bit_count", bitCount).
			WithDetail("valid_range", "1-32")
	}

	if len(data) == 0 {
		v.Bits = make([]bool, bitCount)
		v.BitCount = bitCount
		return nil
	}

	requiredBytes := (bitCount + 7) / 8
	if len(data) < requiredBytes {
		return validator.NewValidationError(validator.ErrCodeInvalidInput,
			fmt.Sprintf("insufficient data: need %d bytes, got %d", requiredBytes, len(data))).
			WithField("required_bytes", requiredBytes).
			WithField("provided_bytes", len(data))
	}

	v.Bits = make([]bool, bitCount)
	v.BitCount = bitCount

	for i := 0; i < bitCount; i++ {
		byteIndex := i / 8
		bitIndex := 7 - (i % 8) // MSB first
		v.Bits[i] = (data[byteIndex] & (1 << bitIndex)) != 0
	}

	return nil
}

// GetBit returns the bit at the specified position (0-indexed)
func (v *VARBIT) GetBit(position int) (bool, error) {
	if v == nil {
		return false, validator.NewValidationError(validator.ErrCodeInvalidInput, "VARBIT is nil")
	}

	if position < 0 || position >= len(v.Bits) {
		return false, validator.NewValidationError(validator.ErrCodeInvalidInput,
			fmt.Sprintf("bit position out of range: %d", position)).
			WithField("position", position).
			WithField("max_position", len(v.Bits)-1)
	}

	return v.Bits[position], nil
}

// SetBit sets the bit at the specified position (0-indexed)
func (v *VARBIT) SetBit(position int, value bool) error {
	if v == nil {
		return validator.NewValidationError(validator.ErrCodeInvalidInput, "VARBIT is nil")
	}

	if position < 0 || position >= len(v.Bits) {
		return validator.NewValidationError(validator.ErrCodeInvalidInput,
			fmt.Sprintf("bit position out of range: %d", position)).
			WithField("position", position).
			WithField("max_position", len(v.Bits)-1)
	}

	v.Bits[position] = value
	return nil
}

// CountSetBits returns the number of bits set to true (1)
func (v *VARBIT) CountSetBits() int {
	if v == nil {
		return 0
	}

	count := 0
	for _, bit := range v.Bits {
		if bit {
			count++
		}
	}
	return count
}

// CountUnsetBits returns the number of bits set to false (0)
func (v *VARBIT) CountUnsetBits() int {
	if v == nil {
		return 0
	}

	return v.BitCount - v.CountSetBits()
}

// Invert inverts all bits in the VARBIT
func (v *VARBIT) Invert() {
	if v == nil {
		return
	}

	for i := range v.Bits {
		v.Bits[i] = !v.Bits[i]
	}
}

// ToUint64 converts the VARBIT to a uint64 value (big-endian)
func (v *VARBIT) ToUint64() uint64 {
	if v == nil || len(v.Bits) == 0 {
		return 0
	}

	var result uint64
	for i, bit := range v.Bits {
		if bit {
			shift := v.BitCount - 1 - i
			result |= 1 << shift
		}
	}

	return result
}

// FromUint64 creates a VARBIT from a uint64 value
func FromUint64(value uint64, bitCount int) *VARBIT {
	if bitCount <= 0 || bitCount > 64 {
		bitCount = OptimalBitCount(value)
	}

	if bitCount > 32 {
		bitCount = 32 // VARBIT max is 32 bits
	}

	bits := make([]bool, bitCount)
	for i := 0; i < bitCount; i++ {
		shift := bitCount - 1 - i
		bits[i] = (value & (1 << shift)) != 0
	}

	return &VARBIT{
		Bits:     bits,
		BitCount: bitCount,
	}
}

// Constants for common VARBIT usage patterns
const (
	// Flag sizes
	FlagBits         = 1   // Single flag
	StatusBits       = 4   // Status flags
	BonusFlagBits    = 8   // Bonus flags
	PropertiesBits   = 16  // Property flags

	// Bitfield sizes for game-specific data
	ElementalBits    = 4   // Elemental type
	BehaviorBits     = 8   // Behavior flags
	RequirementBits  = 16  // Requirements
)

// NewVARBIT creates a new VARBIT from a bit slice
func NewVARBIT(bits []bool) *VARBIT {
	return &VARBIT{
		Bits:     bits,
		BitCount: len(bits),
	}
}

// NewVARBITFromUint64 creates a VARBIT from a uint64 value
func NewVARBITFromUint64(value uint64, bitCount int) *VARBIT {
	return FromUint64(value, bitCount)
}

// NewEmptyVARBIT creates an empty VARBIT with specified bit count
func NewEmptyVARBIT(bitCount int) *VARBIT {
	if bitCount <= 0 || bitCount > 32 {
		bitCount = 1
	}

	return &VARBIT{
		Bits:     make([]bool, bitCount),
		BitCount: bitCount,
	}
}

// VARBITCollection represents a collection of VARBIT values
type VARBITCollection struct {
	Values []*VARBIT
}

// NewVARBITCollection creates an empty VARBIT collection
func NewVARBITCollection() *VARBITCollection {
	return &VARBITCollection{
		Values: make([]*VARBIT, 0),
	}
}

// Add adds a VARBIT to the collection
func (vc *VARBITCollection) Add(value *VARBIT) {
	if vc.Values == nil {
		vc.Values = make([]*VARBIT, 0)
	}
	vc.Values = append(vc.Values, value)
}

// Get returns a VARBIT by index
func (vc *VARBITCollection) Get(index int) *VARBIT {
	if index < 0 || index >= len(vc.Values) {
		return nil
	}
	return vc.Values[index]
}

// Size returns the number of VARBITs in the collection
func (vc *VARBITCollection) Size() int {
	return len(vc.Values)
}

// TotalBitSize returns the total size in bits of all VARBITs
func (vc *VARBITCollection) TotalBitSize() int {
	total := 0
	for _, v := range vc.Values {
		if v != nil {
			total += v.Size()
		}
	}
	return total
}

// Validate validates all VARBITs in the collection
func (vc *VARBITCollection) Validate() error {
	for i, v := range vc.Values {
		if v == nil {
			return validator.NewValidationError(validator.ErrCodeInvalidInput,
				fmt.Sprintf("VARBIT at index %d is nil", i)).
				WithField("index", i)
		}

		if v.BitCount < 1 || v.BitCount > 32 {
			return validator.NewValidationError(validator.ErrCodeInvalidBitCount,
				fmt.Sprintf("invalid bit count at index %d: %d", i, v.BitCount)).
				WithField("index", i).
				WithField("bit_count", v.BitCount)
		}
	}
	return nil
}

// Helper functions for common game-related VARBIT operations

// EncodeElementalType encodes an elemental type as a VARBIT
func EncodeElementalType(elemental string) (*VARBIT, error) {
	// Map elemental types to bit patterns (4 bits)
	elementalMap := map[string]uint64{
		"none":     0x0, // 0000
		"fire":     0x1, // 0001
		"shock":    0x2, // 0010
		"corrosive": 0x3, // 0011
		"cryo":     0x4, // 0100
		"radiation": 0x5, // 0101
		"explosive": 0x6, // 0110
		"psionic":  0x7, // 0111
	}

	value, exists := elementalMap[elemental]
	if !exists {
		return nil, validator.NewValidationError(validator.ErrCodeInvalidInput,
			fmt.Sprintf("unknown elemental type: %s", elemental)).
			WithField("elemental", elemental).
			WithDetail("valid_types", getValidElementalTypes())
	}

	return NewVARBITFromUint64(value, ElementalBits), nil
}

// DecodeElementalType decodes a VARBIT as an elemental type
func DecodeElementalType(v *VARBIT) (string, error) {
	if v == nil {
		return "", validator.NewValidationError(validator.ErrCodeInvalidInput, "VARBIT is nil")
	}

	value := v.ToUint64()

	// Map bit patterns back to elemental types
	elementalMap := map[uint64]string{
		0x0: "none",
		0x1: "fire",
		0x2: "shock",
		0x3: "corrosive",
		0x4: "cryo",
		0x5: "radiation",
		0x6: "explosive",
		0x7: "psionic",
	}

	elemental, exists := elementalMap[value]
	if !exists {
		return "", validator.NewValidationError(validator.ErrCodeInvalidInput,
			fmt.Sprintf("unknown elemental type value: %d", value)).
			WithField("value", value).
			WithDetail("valid_range", "0-7")
	}

	return elemental, nil
}

// getValidElementalTypes returns a list of valid elemental types
func getValidElementalTypes() []string {
	return []string{"none", "fire", "shock", "corrosive", "cryo", "radiation", "explosive", "psionic"}
}

// EncodeStatusFlags encodes status flags as a VARBIT
func EncodeStatusFlags(flags map[string]bool) (*VARBIT, error) {
	if len(flags) > StatusBits {
		return nil, validator.NewValidationError(validator.ErrCodeInvalidInput,
			fmt.Sprintf("too many status flags: %d", len(flags))).
			WithField("flag_count", len(flags)).
			WithDetail("max_flags", StatusBits)
	}

	bits := make([]bool, StatusBits)

	flagMap := map[string]int{
		"legendary":   0,
		"unique":      1,
		"rare":        2,
		"epic":        3,
	}

	for flagName, flagValue := range flags {
		position, exists := flagMap[flagName]
		if exists {
			bits[position] = flagValue
		}
	}

	return NewVARBIT(bits), nil
}

// DecodeStatusFlags decodes a VARBIT as status flags
func DecodeStatusFlags(v *VARBIT) (map[string]bool, error) {
	if v == nil {
		return nil, validator.NewValidationError(validator.ErrCodeInvalidInput, "VARBIT is nil")
	}

	flags := make(map[string]bool)

	flagMap := map[int]string{
		0: "legendary",
		1: "unique",
		2: "rare",
		3: "epic",
	}

	for i := 0; i < len(v.Bits) && i < len(flagMap); i++ {
		if v.Bits[i] {
			flags[flagMap[i]] = true
		}
	}

	return flags, nil
}

// Bitwise operations

// AND performs bitwise AND with another VARBIT
func (v *VARBIT) AND(other *VARBIT) *VARBIT {
	if v == nil || other == nil {
		return nil
	}

	maxBits := v.BitCount
	if other.BitCount < maxBits {
		maxBits = other.BitCount
	}

	result := make([]bool, maxBits)
	for i := 0; i < maxBits; i++ {
		if i < len(v.Bits) && i < len(other.Bits) {
			result[i] = v.Bits[i] && other.Bits[i]
		}
	}

	return &VARBIT{
		Bits:     result,
		BitCount: maxBits,
	}
}

// OR performs bitwise OR with another VARBIT
func (v *VARBIT) OR(other *VARBIT) *VARBIT {
	if v == nil || other == nil {
		return nil
	}

	maxBits := v.BitCount
	if other.BitCount > maxBits {
		maxBits = other.BitCount
	}

	result := make([]bool, maxBits)
	for i := 0; i < maxBits; i++ {
		bit1 := i < len(v.Bits) && v.Bits[i]
		bit2 := i < len(other.Bits) && other.Bits[i]
		result[i] = bit1 || bit2
	}

	return &VARBIT{
		Bits:     result,
		BitCount: maxBits,
	}
}

// XOR performs bitwise XOR with another VARBIT
func (v *VARBIT) XOR(other *VARBIT) *VARBIT {
	if v == nil || other == nil {
		return nil
	}

	maxBits := v.BitCount
	if other.BitCount > maxBits {
		maxBits = other.BitCount
	}

	result := make([]bool, maxBits)
	for i := 0; i < maxBits; i++ {
		bit1 := i < len(v.Bits) && v.Bits[i]
		bit2 := i < len(other.Bits) && other.Bits[i]
		result[i] = bit1 != bit2
	}

	return &VARBIT{
		Bits:     result,
		BitCount: maxBits,
	}
}

// NOT performs bitwise NOT (inverts all bits)
func (v *VARBIT) NOT() *VARBIT {
	if v == nil {
		return nil
	}

	result := make([]bool, len(v.Bits))
	for i, bit := range v.Bits {
		result[i] = !bit
	}

	return &VARBIT{
		Bits:     result,
		BitCount: v.BitCount,
	}
}