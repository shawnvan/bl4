package datatypes

import (
	"fmt"

	"github.com/shawnvan/bl4/internal/codec/bitstream"
	"github.com/shawnvan/bl4/pkg/validator"
)

// PART represents a part or component reference in the item data
// Format: [index][value] where index references a parts database and value contains part-specific data
type PART struct {
	Index uint64 // Part index (references part database)
	Value uint64 // Part value (part-specific data)
}

// DecodePART reads and decodes a PART from the bitstream
// PART format:
// - 16 bits: part index (0-65535)
// - 16 bits: part value (0-65535)
func DecodePART(reader *bitstream.Reader) (*PART, error) {
	// Read the part index (16 bits)
	index, err := reader.ReadBits(16)
	if err != nil {
		return nil, fmt.Errorf("failed to read PART index: %w", err)
	}

	// Read the part value (16 bits)
	value, err := reader.ReadBits(16)
	if err != nil {
		return nil, fmt.Errorf("failed to read PART value: %w", err)
	}

	// Validate index and value ranges
	if index > MaxPartIndex {
		return nil, validator.NewValidationError(validator.ErrCodeInvalidInput,
			fmt.Sprintf("invalid part index: %d", index)).
			WithField("index", index).
			WithDetail("max_index", MaxPartIndex)
	}

	return &PART{
		Index: index,
		Value: value,
	}, nil
}

// EncodePART encodes a PART to the bitstream writer
func EncodePART(writer *bitstream.Writer, index, value uint64) error {
	// Validate index
	if index > MaxPartIndex {
		return validator.NewValidationError(validator.ErrCodeInvalidInput,
			fmt.Sprintf("invalid part index: %d", index)).
			WithField("index", index).
			WithDetail("max_index", MaxPartIndex)
	}

	// Validate value
	if value > MaxPartValue {
		return validator.NewValidationError(validator.ErrCodeInvalidInput,
			fmt.Sprintf("invalid part value: %d", value)).
			WithField("value", value).
			WithDetail("max_value", MaxPartValue)
	}

	// Write part index (16 bits)
	if err := writer.WriteBits(index, 16); err != nil {
		return fmt.Errorf("failed to write PART index: %w", err)
	}

	// Write part value (16 bits)
	if err := writer.WriteBits(value, 16); err != nil {
		return fmt.Errorf("failed to write PART value: %w", err)
	}

	return nil
}

// Constants for PART data type
const (
	MaxPartIndex = 65535 // Maximum part index (16 bits)
	MaxPartValue = 65535 // Maximum part value (16 bits)
)

// PartType represents the type/category of a part
type PartType string

const (
	PartTypeBarrel     PartType = "barrel"
	PartTypeGrip       PartType = "grip"
	PartTypeSight      PartType = "sight"
	PartTypeStock      PartType = "stock"
	PartTypeMagazine   PartType = "magazine"
	PartTypeAccessory  PartType = "accessory"
	PartTypeMaterial   PartType = "material"
	PartTypeElement    PartType = "element"
	PartTypeBonus      PartType = "bonus"
	PartTypeSpecial    PartType = "special"
)

// PartInfo contains detailed information about a part
type PartInfo struct {
	Index       int                    `json:"index"`
	Type        PartType               `json:"type"`
	Name        string                 `json:"name"`
	Value       int                    `json:"value"`
	Level       int                    `json:"level,omitempty"`
	Rarity      string                 `json:"rarity,omitempty"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
	Manufacturer string                `json:"manufacturer,omitempty"`
	Description string                 `json:"description,omitempty"`
}

// String returns a string representation of the PART
func (p *PART) String() string {
	return fmt.Sprintf("PART{index: %d, value: %d}", p.Index, p.Value)
}

// Size returns the total size in bits of this PART
func (p *PART) Size() int {
	return 32 // 16 bits for index + 16 bits for value
}

// Equals checks if two PARTs are equal
func (p *PART) Equals(other *PART) bool {
	if other == nil {
		return false
	}
	return p.Index == other.Index && p.Value == other.Value
}

// Copy creates a deep copy of the PART
func (p *PART) Copy() *PART {
	if p == nil {
		return nil
	}
	return &PART{
		Index: p.Index,
		Value: p.Value,
	}
}

// GetBytes returns the PART as a byte slice (big-endian)
func (p *PART) GetBytes() []byte {
	if p == nil {
		return []byte{0, 0, 0, 0}
	}

	result := make([]byte, 4)

	// Index (first 2 bytes)
	result[0] = byte(p.Index >> 8)
	result[1] = byte(p.Index)

	// Value (last 2 bytes)
	result[2] = byte(p.Value >> 8)
	result[3] = byte(p.Value)

	return result
}

// SetFromBytes sets the PART from a byte slice (big-endian)
func (p *PART) SetFromBytes(data []byte) error {
	if p == nil {
		return validator.NewValidationError(validator.ErrCodeInvalidInput, "PART is nil")
	}

	if len(data) < 4 {
		return validator.NewValidationError(validator.ErrCodeInvalidInput,
			fmt.Sprintf("insufficient data: need 4 bytes, got %d", len(data))).
			WithField("provided_bytes", len(data))
	}

	// Extract index (first 2 bytes)
	p.Index = uint64(data[0])<<8 | uint64(data[1])

	// Extract value (last 2 bytes)
	p.Value = uint64(data[2])<<8 | uint64(data[3])

	// Validate ranges
	if p.Index > MaxPartIndex {
		return validator.NewValidationError(validator.ErrCodeInvalidInput,
			fmt.Sprintf("invalid part index: %d", p.Index)).
			WithField("index", p.Index).
			WithDetail("max_index", MaxPartIndex)
	}

	if p.Value > MaxPartValue {
		return validator.NewValidationError(validator.ErrCodeInvalidInput,
			fmt.Sprintf("invalid part value: %d", p.Value)).
			WithField("value", p.Value).
			WithDetail("max_value", MaxPartValue)
	}

	return nil
}

// GetLevel extracts level information from the part value
func (p *PART) GetLevel() int {
	if p == nil {
		return 0
	}
	// Level is typically stored in the high 8 bits of the value
	return int((p.Value >> 8) & 0xFF)
}

// GetPartValue extracts the actual part value from the value field
func (p *PART) GetPartValue() int {
	if p == nil {
		return 0
	}
	// Part value is typically stored in the low 8 bits of the value
	return int(p.Value & 0xFF)
}

// IsHighQuality checks if the part is considered high quality
func (p *PART) IsHighQuality() bool {
	if p == nil {
		return false
	}
	// High quality parts typically have values in the upper range
	return p.Value >= (MaxPartValue * 3 / 4)
}

// NewPART creates a new PART with specified index and value
func NewPART(index, value uint64) *PART {
	return &PART{
		Index: index,
		Value: value,
	}
}

// NewPARTWithLevel creates a new PART with specified index and level
func NewPARTWithLevel(index, level, partValue int) *PART {
	value := (uint64(level&0xFF) << 8) | uint64(partValue&0xFF)
	return &PART{
		Index: uint64(index),
		Value: value,
	}
}

// PARTCollection represents a collection of PART values
type PARTCollection struct {
	Parts []*PART
}

// NewPARTCollection creates an empty PART collection
func NewPARTCollection() *PARTCollection {
	return &PARTCollection{
		Parts: make([]*PART, 0),
	}
}

// Add adds a PART to the collection
func (pc *PARTCollection) Add(part *PART) {
	if pc.Parts == nil {
		pc.Parts = make([]*PART, 0)
	}
	pc.Parts = append(pc.Parts, part)
}

// Get returns a PART by index
func (pc *PARTCollection) Get(index int) *PART {
	if index < 0 || index >= len(pc.Parts) {
		return nil
	}
	return pc.Parts[index]
}

// Size returns the number of PARTs in the collection
func (pc *PARTCollection) Size() int {
	return len(pc.Parts)
}

// TotalBitSize returns the total size in bits of all PARTs
func (pc *PARTCollection) TotalBitSize() int {
	return len(pc.Parts) * 32 // Each PART is 32 bits
}

// Validate validates all PARTs in the collection
func (pc *PARTCollection) Validate() error {
	for i, part := range pc.Parts {
		if part == nil {
			return validator.NewValidationError(validator.ErrCodeInvalidInput,
				fmt.Sprintf("PART at index %d is nil", i)).
				WithField("index", i)
		}

		if part.Index > MaxPartIndex {
			return validator.NewValidationError(validator.ErrCodeInvalidInput,
				fmt.Sprintf("invalid part index at collection index %d: %d", i, part.Index)).
				WithField("collection_index", i).
				WithField("part_index", part.Index)
		}

		if part.Value > MaxPartValue {
			return validator.NewValidationError(validator.ErrCodeInvalidInput,
				fmt.Sprintf("invalid part value at collection index %d: %d", i, part.Value)).
				WithField("collection_index", i).
				WithField("part_value", part.Value)
		}
	}
	return nil
}

// GetPartsByType returns parts filtered by type (requires additional data)
func (pc *PARTCollection) GetPartsByType(partType PartType) []*PART {
	// This would need access to a parts database to determine types
	// For now, return all parts as a placeholder
	return pc.Parts
}

// GetHighQualityParts returns all high-quality parts in the collection
func (pc *PARTCollection) GetHighQualityParts() []*PART {
	var highQuality []*PART
	for _, part := range pc.Parts {
		if part != nil && part.IsHighQuality() {
			highQuality = append(highQuality, part)
		}
	}
	return highQuality
}

// GetMaxLevel returns the maximum level among all parts
func (pc *PARTCollection) GetMaxLevel() int {
	maxLevel := 0
	for _, part := range pc.Parts {
		if part != nil {
			level := part.GetLevel()
			if level > maxLevel {
				maxLevel = level
			}
		}
	}
	return maxLevel
}

// String returns a string representation of the collection
func (pc *PARTCollection) String() string {
	if pc.Size() == 0 {
		return "PARTCollection{empty}"
	}

	var parts []string
	for _, part := range pc.Parts {
		if part != nil {
			parts = append(parts, fmt.Sprintf("%d:%d", part.Index, part.Value))
		} else {
			parts = append(parts, "nil")
		}
	}

	return fmt.Sprintf("PARTCollection{%s}", fmt.Sprintf("%v", parts))
}

// Helper functions for part type resolution

// ResolvePartType attempts to determine the part type based on index
func ResolvePartType(index uint64) PartType {
	// This would typically query a parts database
	// For now, use some basic heuristics

	switch index {
	case 0, 1, 2, 3, 4:
		return PartTypeBarrel
	case 5, 6, 7, 8:
		return PartTypeGrip
	case 9, 10, 11, 12:
		return PartTypeSight
	case 13, 14, 15, 16:
		return PartTypeStock
	case 17, 18, 19, 20:
		return PartTypeMagazine
	case 21, 22, 23, 24:
		return PartTypeAccessory
	case 25, 26, 27, 28:
		return PartTypeMaterial
	case 29, 30, 31, 32:
		return PartTypeElement
	case 33, 34, 35, 36:
		return PartTypeBonus
	default:
		return PartTypeSpecial
	}
}

// GetPartName attempts to get a human-readable name for a part
func GetPartName(index uint64, value uint64) string {
	partType := ResolvePartType(index)

	// Generate a name based on type and value
	switch partType {
	case PartTypeBarrel:
		return fmt.Sprintf("Barrel (Level %d)", (value>>8)&0xFF)
	case PartTypeGrip:
		return fmt.Sprintf("Grip (Level %d)", (value>>8)&0xFF)
	case PartTypeSight:
		return fmt.Sprintf("Sight (Level %d)", (value>>8)&0xFF)
	case PartTypeStock:
		return fmt.Sprintf("Stock (Level %d)", (value>>8)&0xFF)
	case PartTypeMagazine:
		return fmt.Sprintf("Magazine (Level %d)", (value>>8)&0xFF)
	case PartTypeAccessory:
		return fmt.Sprintf("Accessory (Level %d)", (value>>8)&0xFF)
	case PartTypeMaterial:
		return fmt.Sprintf("Material (Quality %d)", value&0xFF)
	case PartTypeElement:
		return fmt.Sprintf("Elemental Part (Power %d)", value&0xFF)
	case PartTypeBonus:
		return fmt.Sprintf("Bonus Part (Value %d)", value&0xFF)
	default:
		return fmt.Sprintf("Special Part %d (Value %d)", index, value)
	}
}

// GetPartInfo creates detailed part information
func GetPartInfo(part *PART) *PartInfo {
	if part == nil {
		return nil
	}

	info := &PartInfo{
		Index: int(part.Index),
		Value: int(part.Value),
		Level: part.GetLevel(),
		Type:  ResolvePartType(part.Index),
		Name:  GetPartName(part.Index, part.Value),
		Attributes: make(map[string]interface{}),
	}

	// Add some basic attributes
	info.Attributes["part_value"] = part.GetPartValue()
	info.Attributes["is_high_quality"] = part.IsHighQuality()
	info.Attributes["raw_index"] = part.Index
	info.Attributes["raw_value"] = part.Value

	return info
}

// ValidatePartStructure validates the structure of a parts collection
func ValidatePartStructure(parts []*PART) error {
	if len(parts) == 0 {
		return nil // Empty collection is valid
	}

	if len(parts) > MaxPartsPerItem {
		return validator.NewValidationError(validator.ErrCodeInvalidInput,
			fmt.Sprintf("too many parts: %d", len(parts))).
			WithField("part_count", len(parts)).
			WithDetail("max_parts", MaxPartsPerItem)
	}

	seenIndices := make(map[uint64]bool)
	for i, part := range parts {
		if part == nil {
			return validator.NewValidationError(validator.ErrCodeInvalidInput,
				fmt.Sprintf("part at index %d is nil", i)).
				WithField("part_index", i)
		}

		// Check for duplicate indices
		if seenIndices[part.Index] {
			return validator.NewValidationError(validator.ErrCodeInvalidInput,
				fmt.Sprintf("duplicate part index: %d", part.Index)).
				WithField("part_index", i).
				WithField("duplicate_index", part.Index)
		}
		seenIndices[part.Index] = true

		// Validate ranges
		if part.Index > MaxPartIndex {
			return validator.NewValidationError(validator.ErrCodeInvalidInput,
				fmt.Sprintf("invalid part index at position %d: %d", i, part.Index)).
				WithField("position", i).
				WithField("index", part.Index)
		}

		if part.Value > MaxPartValue {
			return validator.NewValidationError(validator.ErrCodeInvalidInput,
				fmt.Sprintf("invalid part value at position %d: %d", i, part.Value)).
				WithField("position", i).
				WithField("value", part.Value)
		}
	}

	return nil
}

// Constants for item structure validation
const (
	MaxPartsPerItem = 50 // Maximum number of parts per item
)