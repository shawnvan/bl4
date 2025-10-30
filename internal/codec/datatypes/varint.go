package datatypes

import (
	"fmt"

	"github.com/shawnvan/bl4/internal/codec/bitstream"
	"github.com/shawnvan/bl4/pkg/validator"
)

// VARINT represents a variable-length integer data type
// Format: [bit_count][value] where bit_count determines how many bits to read for the value
type VARINT struct {
	Value    uint64 // The decoded integer value
	BitCount int    // Number of bits used to encode the value
}

// DecodeVARINT reads and decodes a VARINT from the bitstream
// VARINT format:
// - 6 bits: bit count (1-64, specifies how many bits follow for the actual value)
// - N bits: integer value (where N is the bit count)
func DecodeVARINT(reader *bitstream.Reader) (*VARINT, error) {
	// Read the bit count (6 bits)
	bitCountRaw, err := reader.ReadBits(6)
	if err != nil {
		return nil, fmt.Errorf("failed to read VARINT bit count: %w", err)
	}

	bitCount := int(bitCountRaw)

	// Validate bit count
	if bitCount == 0 || bitCount > 64 {
		return nil, validator.NewValidationError(validator.ErrCodeInvalidBitCount,
			fmt.Sprintf("invalid VARINT bit count: %d", bitCount)).
			WithField("bit_count", bitCount).
			WithDetail("valid_range", "1-64")
	}

	// Read the actual value
	value, err := reader.ReadBits(bitCount)
	if err != nil {
		return nil, fmt.Errorf("failed to read VARINT value: %w", err)
	}

	return &VARINT{
		Value:    value,
		BitCount: bitCount,
	}, nil
}

// EncodeVARINT encodes a VARINT to the bitstream writer
func EncodeVARINT(writer *bitstream.Writer, value uint64, bitCount int) error {
	// Validate bit count
	if bitCount < 1 || bitCount > 64 {
		return validator.NewValidationError(validator.ErrCodeInvalidBitCount,
			fmt.Sprintf("invalid VARINT bit count: %d", bitCount)).
			WithField("bit_count", bitCount).
			WithDetail("valid_range", "1-64")
	}

	// Check if value fits in the specified bit count
	if bitCount < 64 && value >= (1<<bitCount) {
		return validator.NewValidationError(validator.ErrCodeValueTooLarge,
			fmt.Sprintf("value %d doesn't fit in %d bits", value, bitCount)).
			WithField("value", value).
			WithField("bit_count", bitCount)
	}

	// Write bit count (6 bits)
	if err := writer.WriteBits(uint64(bitCount), 6); err != nil {
		return fmt.Errorf("failed to write VARINT bit count: %w", err)
	}

	// Write the value
	if err := writer.WriteBits(value, bitCount); err != nil {
		return fmt.Errorf("failed to write VARINT value: %w", err)
	}

	return nil
}

// OptimalBitCount calculates the optimal bit count needed to encode a value
func OptimalBitCount(value uint64) int {
	if value == 0 {
		return 1
	}

	// Find the position of the highest set bit
	bitCount := 0
	temp := value
	for temp > 0 {
		temp >>= 1
		bitCount++
	}

	return bitCount
}

// IsValidValue checks if a value can be encoded as a VARINT
func IsValidValue(value uint64) bool {
	return true // All uint64 values can be encoded as VARINT
}

// String returns a string representation of the VARINT
func (v *VARINT) String() string {
	return fmt.Sprintf("VARINT{value: %d, bits: %d}", v.Value, v.BitCount)
}

// Size returns the total size in bits of this VARINT
func (v *VARINT) Size() int {
	return 6 + v.BitCount // 6 bits for count + value bits
}

// Equals checks if two VARINTs are equal
func (v *VARINT) Equals(other *VARINT) bool {
	if other == nil {
		return false
	}
	return v.Value == other.Value && v.BitCount == other.BitCount
}

// Copy creates a deep copy of the VARINT
func (v *VARINT) Copy() *VARINT {
	if v == nil {
		return nil
	}
	return &VARINT{
		Value:    v.Value,
		BitCount: v.BitCount,
	}
}

// GetBytes returns the VARINT value as a byte slice (big-endian)
func (v *VARINT) GetBytes() []byte {
	if v == nil {
		return []byte{0}
	}

	byteCount := (v.BitCount + 7) / 8
	result := make([]byte, byteCount)

	for i := 0; i < byteCount; i++ {
		shift := (byteCount - 1 - i) * 8
		result[i] = byte(v.Value >> shift)
	}

	return result
}

// SetFromBytes sets the VARINT value from a byte slice (big-endian)
func (v *VARINT) SetFromBytes(data []byte) error {
	if v == nil {
		return validator.NewValidationError(validator.ErrCodeInvalidInput, "VARINT is nil")
	}

	if len(data) == 0 {
		v.Value = 0
		v.BitCount = 1
		return nil
	}

	if len(data) > 8 {
		return validator.NewValidationError(validator.ErrCodeInvalidInput,
			fmt.Sprintf("byte slice too large for VARINT: %d bytes", len(data))).
			WithField("byte_count", len(data)).
			WithDetail("max_bytes", 8)
	}

	// Convert bytes to uint64 (big-endian)
	value := uint64(0)
	for _, b := range data {
		value = (value << 8) | uint64(b)
	}

	v.Value = value
	v.BitCount = OptimalBitCount(value)

	return nil
}

// Constants for common VARINT usage patterns
const (
	// Game-related value sizes
	LevelBits        = 8  // Item levels (0-255)
	TypeBits         = 8  // Item types (0-255)
	ManufacturerBits = 8  // Manufacturers (0-255)
	RarityBits       = 4  // Rarity levels (0-15)
	QualityBits      = 8  // Quality levels (0-255)

	// Part-related sizes
	PartIndexBits = 16 // Part indices (0-65535)
	PartValueBits = 16 // Part values (0-65535)

	// Statistics sizes
	StatBits = 8 // Most statistics fit in 8 bits
	BonusBits = 16 // Bonuses might need more bits
)

// NewVARINT creates a new VARINT with optimal bit count
func NewVARINT(value uint64) *VARINT {
	return &VARINT{
		Value:    value,
		BitCount: OptimalBitCount(value),
	}
}

// NewVARINTWithBits creates a new VARINT with specified bit count
func NewVARINTWithBits(value uint64, bitCount int) *VARINT {
	return &VARINT{
		Value:    value,
		BitCount: bitCount,
	}
}

// VARINTCollection represents a collection of VARINT values
type VARINTCollection struct {
	Values []*VARINT
}

// NewVARINTCollection creates an empty VARINT collection
func NewVARINTCollection() *VARINTCollection {
	return &VARINTCollection{
		Values: make([]*VARINT, 0),
	}
}

// Add adds a VARINT to the collection
func (vc *VARINTCollection) Add(value *VARINT) {
	if vc.Values == nil {
		vc.Values = make([]*VARINT, 0)
	}
	vc.Values = append(vc.Values, value)
}

// Get returns a VARINT by index
func (vc *VARINTCollection) Get(index int) *VARINT {
	if index < 0 || index >= len(vc.Values) {
		return nil
	}
	return vc.Values[index]
}

// Size returns the number of VARINTs in the collection
func (vc *VARINTCollection) Size() int {
	return len(vc.Values)
}

// TotalBitSize returns the total size in bits of all VARINTs
func (vc *VARINTCollection) TotalBitSize() int {
	total := 0
	for _, v := range vc.Values {
		if v != nil {
			total += v.Size()
		}
	}
	return total
}

// Validate validates all VARINTs in the collection
func (vc *VARINTCollection) Validate() error {
	for i, v := range vc.Values {
		if v == nil {
			return validator.NewValidationError(validator.ErrCodeInvalidInput,
				fmt.Sprintf("VARINT at index %d is nil", i)).
				WithField("index", i)
		}

		if v.BitCount < 1 || v.BitCount > 64 {
			return validator.NewValidationError(validator.ErrCodeInvalidBitCount,
				fmt.Sprintf("invalid bit count at index %d: %d", i, v.BitCount)).
				WithField("index", i).
				WithField("bit_count", v.BitCount)
		}
	}
	return nil
}

// String returns a string representation of the collection
func (vc *VARINTCollection) String() string {
	if vc.Size() == 0 {
		return "VARINTCollection{empty}"
	}

	var values []string
	for _, v := range vc.Values {
		if v != nil {
			values = append(values, fmt.Sprintf("%d", v.Value))
		} else {
			values = append(values, "nil")
		}
	}

	return fmt.Sprintf("VARINTCollection{%s}", fmt.Sprintf("%v", values))
}

// Helper functions for common game-related VARINT operations

// EncodeLevel encodes an item level (0-100) as a VARINT
func EncodeLevel(level int) (*VARINT, error) {
	if level < 0 || level > 100 {
		return nil, validator.NewValidationError(validator.ErrCodeInvalidItemLevel,
			fmt.Sprintf("invalid level: %d", level)).
			WithField("level", level).
			WithDetail("valid_range", "0-100")
	}

	return NewVARINTWithBits(uint64(level), LevelBits), nil
}

// DecodeLevel decodes a VARINT as an item level
func DecodeLevel(v *VARINT) (int, error) {
	if v == nil {
		return 0, validator.NewValidationError(validator.ErrCodeInvalidInput, "VARINT is nil")
	}

	if v.Value > 100 {
		return 0, validator.NewValidationError(validator.ErrCodeInvalidItemLevel,
			fmt.Sprintf("invalid level value: %d", v.Value)).
			WithField("value", v.Value).
			WithDetail("valid_range", "0-100")
	}

	return int(v.Value), nil
}

// EncodeType encodes an item type as a VARINT
func EncodeType(itemType string) (*VARINT, error) {
	// Map common item types to numeric values
	typeMap := map[string]int{
		"pistol":   0,
		"shotgun":  1,
		"rifle":    2,
		"smg":      3,
		"sniper":   4,
		"launcher": 5,
		"melee":    6,
		"shield":   7,
		"grenade":  8,
		"artifact": 9,
	}

	typeValue, exists := typeMap[itemType]
	if !exists {
		return nil, validator.NewValidationError(validator.ErrCodeInvalidItemType,
			fmt.Sprintf("unknown item type: %s", itemType)).
			WithField("type", itemType).
			WithDetail("valid_types", getValidTypes())
	}

	return NewVARINTWithBits(uint64(typeValue), TypeBits), nil
}

// DecodeType decodes a VARINT as an item type
func DecodeType(v *VARINT) (string, error) {
	if v == nil {
		return "", validator.NewValidationError(validator.ErrCodeInvalidInput, "VARINT is nil")
	}

	// Map numeric values back to item types
	typeMap := map[uint64]string{
		0: "pistol",
		1: "shotgun",
		2: "rifle",
		3: "smg",
		4: "sniper",
		5: "launcher",
		6: "melee",
		7: "shield",
		8: "grenade",
		9: "artifact",
	}

	itemType, exists := typeMap[v.Value]
	if !exists {
		return "", validator.NewValidationError(validator.ErrCodeInvalidItemType,
			fmt.Sprintf("unknown item type value: %d", v.Value)).
			WithField("value", v.Value).
			WithDetail("valid_range", "0-9")
	}

	return itemType, nil
}

// getValidTypes returns a list of valid item types
func getValidTypes() []string {
	return []string{"pistol", "shotgun", "rifle", "smg", "sniper", "launcher", "melee", "shield", "grenade", "artifact"}
}

// EncodeManufacturer encodes a manufacturer as a VARINT
func EncodeManufacturer(manufacturer string) (*VARINT, error) {
	// Map manufacturers to numeric values
	manufacturerMap := map[string]int{
		"maliwan":    0,
		"jakobs":     1,
		"hyperion":   2,
		"torgue":     3,
		"vladof":     4,
		"dahl":       5,
		"tediore":    6,
		"atlas":      7,
		"cov":        8,
		"children":   9,
		"anshin":     10,
		"pangolin":   11,
		"eridian":    12,
	}

	manufacturerValue, exists := manufacturerMap[manufacturer]
	if !exists {
		return nil, validator.NewValidationError(validator.ErrCodeInvalidManufacturer,
			fmt.Sprintf("unknown manufacturer: %s", manufacturer)).
			WithField("manufacturer", manufacturer).
			WithDetail("valid_manufacturers", getValidManufacturers())
	}

	return NewVARINTWithBits(uint64(manufacturerValue), ManufacturerBits), nil
}

// DecodeManufacturer decodes a VARINT as a manufacturer
func DecodeManufacturer(v *VARINT) (string, error) {
	if v == nil {
		return "", validator.NewValidationError(validator.ErrCodeInvalidInput, "VARINT is nil")
	}

	// Map numeric values back to manufacturers
	manufacturerMap := map[uint64]string{
		0:  "maliwan",
		1:  "jakobs",
		2:  "hyperion",
		3:  "torgue",
		4:  "vladof",
		5:  "dahl",
		6:  "tediore",
		7:  "atlas",
		8:  "cov",
		9:  "children",
		10: "anshin",
		11: "pangolin",
		12: "eridian",
	}

	manufacturer, exists := manufacturerMap[v.Value]
	if !exists {
		return "", validator.NewValidationError(validator.ErrCodeInvalidManufacturer,
			fmt.Sprintf("unknown manufacturer value: %d", v.Value)).
			WithField("value", v.Value).
			WithDetail("valid_range", "0-12")
	}

	return manufacturer, nil
}

// getValidManufacturers returns a list of valid manufacturers
func getValidManufacturers() []string {
	return []string{"maliwan", "jakobs", "hyperion", "torgue", "vladof", "dahl", "tediore", "atlas", "cov", "children", "anshin", "pangolin", "eridian"}
}