package serial

import (
	"fmt"
	"strings"
	"time"

	"github.com/shawnvan/bl4/internal/api/models"
	"github.com/shawnvan/bl4/internal/codec/base85"
	"github.com/shawnvan/bl4/internal/codec/bitstream"
	"github.com/shawnvan/bl4/internal/codec/datatypes"
	"github.com/shawnvan/bl4/internal/codec/token"
	"github.com/shawnvan/bl4/pkg/logger"
	"github.com/shawnvan/bl4/pkg/validator"
)

// Serializer handles the complete serialization of BL4 item data to serial codes
type Serializer struct {
	options *SerializerOptions
}

// SerializerOptions contains options for serialization
type SerializerOptions struct {
	// IncludeMetadata includes additional metadata in the output
	IncludeMetadata bool

	// OptimizeSize enables size optimization
	OptimizeSize bool

	// StrictValidation enables strict validation rules
	StrictValidation bool

	// MaxProcessingTime limits processing time in milliseconds
	MaxProcessingTime int

	// LogLevel controls logging verbosity
	LogLevel string

	// TargetVersion specifies the target BL4 version
	TargetVersion string
}

// NewSerializer creates a new serializer with default options
func NewSerializer() *Serializer {
	return &Serializer{
		options: &SerializerOptions{
			IncludeMetadata:   false,
			OptimizeSize:      false,
			StrictValidation:  true,
			MaxProcessingTime: 5000, // 5 seconds
			LogLevel:          "info",
			TargetVersion:     "1.0",
		},
	}
}

// NewSerializerWithOptions creates a new serializer with custom options
func NewSerializerWithOptions(options *SerializerOptions) *Serializer {
	if options == nil {
		options = &SerializerOptions{}
	}

	// Set defaults for missing options
	if options.MaxProcessingTime == 0 {
		options.MaxProcessingTime = 5000
	}
	if options.LogLevel == "" {
		options.LogLevel = "info"
	}
	if options.TargetVersion == "" {
		options.TargetVersion = "1.0"
	}

	return &Serializer{
		options: options,
	}
}

// SerializeItem performs complete serialization of item data to a serial code
func (s *Serializer) SerializeItem(itemData *models.ItemData) (*SerializationResult, error) {
	startTime := time.Now()

	result := &SerializationResult{
		ItemData:  itemData,
		StartTime: startTime,
	}

	logger.Sugar().Debugw("Starting item serialization",
		"level", itemData.Level,
		"type", itemData.Type,
		"manufacturer", itemData.Manufacturer,
		"parts_count", len(itemData.Parts),
		"options", s.options,
	)

	// Step 1: Validate item data
	if err := s.validateItemData(itemData); err != nil {
		result.Error = err
		result.Duration = time.Since(startTime)
		return result, fmt.Errorf("item data validation failed: %w", err)
	}

	// Step 2: Convert item data to token stream
	tokenStream, err := s.createTokenStream(itemData)
	if err != nil {
		result.Error = err
		result.Duration = time.Since(startTime)
		return result, fmt.Errorf("token stream creation failed: %w", err)
	}
	result.TokenStream = tokenStream

	// Step 3: Convert token stream to bitstream
	bitstreamData, err := s.createBitstream(tokenStream)
	if err != nil {
		result.Error = err
		result.Duration = time.Since(startTime)
		return result, fmt.Errorf("bitstream creation failed: %w", err)
	}
	result.BitstreamData = bitstreamData

	// Step 4: Encode bitstream to Base85
	encoder := base85.NewEncoder()
	encodeOptions := &base85.EncodeOptions{
		IncludeStatistics: true,
		ValidateFormat:     true,
		OptimizeSize:       s.options.OptimizeSize,
		TargetVersion:      s.options.TargetVersion,
	}

	encodeResult, err := encoder.EncodeWithOptions(bitstreamData, encodeOptions)
	if err != nil {
		result.Error = err
		result.Duration = time.Since(startTime)
		return result, fmt.Errorf("Base85 encoding failed: %w", err)
	}

	result.SerialCode = encodeResult.SerialCode
	result.EncodingInfo = encodeResult

	// Step 5: Add additional metadata if requested
	if s.options.IncludeMetadata {
		result.Metadata = s.createMetadata(itemData, tokenStream, encodeResult)
	}

	// Step 6: Validate result
	if s.options.StrictValidation {
		if err := s.validateResult(result); err != nil {
			result.Error = err
			result.Duration = time.Since(startTime)
			return result, fmt.Errorf("result validation failed: %w", err)
		}
	}

	result.Duration = time.Since(startTime)
	result.Success = true

	logger.Sugar().Infow("Item serialization completed",
		"serial_code", result.SerialCode,
		"duration_ms", result.Duration.Milliseconds(),
		"level", itemData.Level,
		"type", itemData.Type,
		"manufacturer", itemData.Manufacturer,
		"parts_count", len(itemData.Parts),
	)

	return result, nil
}

// validateItemData validates the input item data
func (s *Serializer) validateItemData(itemData *models.ItemData) error {
	if itemData == nil {
		return validator.NewValidationError(validator.ErrCodeEmptyInput, "item data is nil")
	}

	// Validate required fields
	if itemData.Level < 1 || itemData.Level > 100 {
		return validator.NewValidationError(validator.ErrCodeInvalidItemLevel,
			fmt.Sprintf("invalid item level: %d", itemData.Level)).
			WithDetail("valid_range", "1-100")
	}

	if itemData.Type == "" {
		return validator.NewValidationError(validator.ErrCodeInvalidItemType, "item type is required")
	}

	if itemData.Manufacturer == "" {
		return validator.NewValidationError(validator.ErrCodeInvalidManufacturer, "manufacturer is required")
	}

	// Validate parts structure
	if len(itemData.Parts) > 50 {
		return validator.NewValidationError(validator.ErrCodeInvalidPartsData,
			fmt.Sprintf("too many parts: %d", len(itemData.Parts))).
			WithDetail("max_parts", 50)
	}

	// Check for duplicate part indices
	seenIndices := make(map[int]bool)
	for i, part := range itemData.Parts {
		if seenIndices[part.Index] {
			return validator.NewValidationError(validator.ErrCodeInvalidPartsData,
				fmt.Sprintf("duplicate part index: %d", part.Index)).
				WithField("part_position", i)
		}
		seenIndices[part.Index] = true

		// Validate individual part
		if err := part.Validate(); err != nil {
			return validator.NewValidationError(validator.ErrCodeInvalidPartsData,
				fmt.Sprintf("invalid part at index %d", i)).
				WithCause(err)
		}
	}

	return nil
}

// createTokenStream converts item data to a token stream
func (s *Serializer) createTokenStream(itemData *models.ItemData) (*token.TokenStream, error) {
	tokenStream := token.NewTokenStream()

	// Add header tokens (level, type, manufacturer)
	if err := s.addHeaderTokens(tokenStream, itemData); err != nil {
		return nil, fmt.Errorf("failed to add header tokens: %w", err)
	}

	// Add part tokens
	if err := s.addPartTokens(tokenStream, itemData.Parts); err != nil {
		return nil, fmt.Errorf("failed to add part tokens: %w", err)
	}

	// Add string tokens (name, description, etc.)
	if err := s.addStringTokens(tokenStream, itemData); err != nil {
		return nil, fmt.Errorf("failed to add string tokens: %w", err)
	}

	logger.Sugar().Debugw("Token stream creation completed",
		"total_tokens", tokenStream.Size(),
		"total_bits", tokenStream.GetTotalBitSize(),
	)

	return tokenStream, nil
}

// addHeaderTokens adds header tokens (level, type, manufacturer) to the token stream
func (s *Serializer) addHeaderTokens(tokenStream *token.TokenStream, itemData *models.ItemData) error {
	// Create level token
	levelToken := token.Token{
		Type:     token.TokenVARINT,
		Value:    uint64(itemData.Level),
		BitSize:  8 + 6, // 8 bits for value + 6 bits for bit count
		Position: int64(tokenStream.GetTotalBitSize()),
	}
	tokenStream.AddToken(levelToken)

	// Create type token
	itemTypeValue, err := s.encodeItemType(itemData.Type)
	if err != nil {
		return fmt.Errorf("failed to encode item type: %w", err)
	}

	typeToken := token.Token{
		Type:     token.TokenVARINT,
		Value:    itemTypeValue,
		BitSize:  8 + 6, // 8 bits for value + 6 bits for bit count
		Position: int64(tokenStream.GetTotalBitSize()),
	}
	tokenStream.AddToken(typeToken)

	// Create manufacturer token
	manufacturerValue, err := s.encodeManufacturer(itemData.Manufacturer)
	if err != nil {
		return fmt.Errorf("failed to encode manufacturer: %w", err)
	}

	manufacturerToken := token.Token{
		Type:     token.TokenVARINT,
		Value:    manufacturerValue,
		BitSize:  8 + 6, // 8 bits for value + 6 bits for bit count
		Position: int64(tokenStream.GetTotalBitSize()),
	}
	tokenStream.AddToken(manufacturerToken)

	return nil
}

// addPartTokens adds part tokens to the token stream
func (s *Serializer) addPartTokens(tokenStream *token.TokenStream, parts []models.PartData) error {
	for i, part := range parts {
		// Create part token with combined index and value
		partValue := (uint64(part.Value) << 8) | uint64(part.Index) // Combine value and index

		partToken := token.Token{
			Type:     token.TokenPART,
			Value:    map[string]uint64{"index": uint64(part.Index), "value": partValue},
			BitSize:  32, // 16 bits for index + 16 bits for value
			Position: int64(tokenStream.GetTotalBitSize()),
		}
		tokenStream.AddToken(partToken)

		logger.Sugar().Debugw("Added part token",
			"part_index", i,
			"part_type", part.Index,
			"part_value", part.Value,
			"token_position", partToken.Position,
		)
	}

	return nil
}

// addStringTokens adds string tokens to the token stream
func (s *Serializer) addStringTokens(tokenStream *token.TokenStream, itemData *models.ItemData) error {
	// Add name token if present
	if itemData.Name != "" {
		nameToken := token.Token{
			Type:     token.TokenSTRING,
			Value:    itemData.Name,
			BitSize:  8 + len(itemData.Name)*8 + 8, // 8 bits for length + data + 8 bits for null terminator
			Position: int64(tokenStream.GetTotalBitSize()),
		}
		tokenStream.AddToken(nameToken)
	}

	// Add description token if present
	if itemData.Description != "" {
		descToken := token.Token{
			Type:     token.TokenSTRING,
			Value:    itemData.Description,
			BitSize:  8 + len(itemData.Description)*8 + 8, // 8 bits for length + data + 8 bits for null terminator
			Position: int64(tokenStream.GetTotalBitSize()),
		}
		tokenStream.AddToken(descToken)
	}

	// Add rarity token if present
	if itemData.Rarity != "" {
		rarityToken := token.Token{
			Type:     token.TokenSTRING,
			Value:    itemData.Rarity,
			BitSize:  8 + len(itemData.Rarity)*8 + 8, // 8 bits for length + data + 8 bits for null terminator
			Position: int64(tokenStream.GetTotalBitSize()),
		}
		tokenStream.AddToken(rarityToken)
	}

	return nil
}

// createBitstream converts token stream to raw bitstream data
func (s *Serializer) createBitstream(tokenStream *token.TokenStream) ([]byte, error) {
	writer := bitstream.NewWriterFromBytes(1024) // Start with 1KB buffer

	for _, tok := range tokenStream.Tokens {
		if err := s.writeToken(writer, tok); err != nil {
			return nil, fmt.Errorf("failed to write token at position %d: %w", tok.Position, err)
		}
	}

	// Get the final bitstream data
	bitstreamData, err := writer.Bytes()
	if err != nil {
		return nil, fmt.Errorf("failed to get bitstream data: %w", err)
	}

	logger.Sugar().Debugw("Bitstream creation completed",
		"total_bytes", len(bitstreamData),
		"total_bits", len(bitstreamData)*8,
	)

	return bitstreamData, nil
}

// writeToken writes a single token to the bitstream writer
func (s *Serializer) writeToken(writer *bitstream.Writer, tok token.Token) error {
	switch tok.Type {
	case token.TokenVARINT:
		value, ok := tok.Value.(uint64)
		if !ok {
			return validator.NewValidationError(validator.ErrCodeTokenValueInvalid,
				"invalid value type for VARINT token")
		}
		// Use 8 bits for standard game values
		return datatypes.EncodeVARINT(writer, value, 8)

	case token.TokenPART:
		partData, ok := tok.Value.(map[string]uint64)
		if !ok {
			return validator.NewValidationError(validator.ErrCodeTokenValueInvalid,
				"invalid value type for PART token")
		}
		index, exists := partData["index"]
		if !exists {
			return validator.NewValidationError(validator.ErrCodeInvalidInput,
				"missing index in PART token")
		}
		value, exists := partData["value"]
		if !exists {
			return validator.NewValidationError(validator.ErrCodeInvalidInput,
				"missing value in PART token")
		}
		return datatypes.EncodePART(writer, index, value)

	case token.TokenSTRING:
		value, ok := tok.Value.(string)
		if !ok {
			return validator.NewValidationError(validator.ErrCodeTokenValueInvalid,
				"invalid value type for STRING token")
		}
		return datatypes.EncodeSTRING(writer, value)

	case token.TokenVARBIT:
		bits, ok := tok.Value.([]bool)
		if !ok {
			return validator.NewValidationError(validator.ErrCodeTokenValueInvalid,
				"invalid value type for VARBIT token")
		}
		return datatypes.EncodeVARBIT(writer, bits)

	default:
		return validator.NewValidationError(validator.ErrCodeInvalidToken,
			fmt.Sprintf("unsupported token type: %s", tok.Type.String()))
	}
}

// encodeItemType converts item type string to numeric value
func (s *Serializer) encodeItemType(itemType string) (uint64, error) {
	typeMap := map[string]uint64{
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

	value, exists := typeMap[strings.ToLower(itemType)]
	if !exists {
		return 0, validator.NewValidationError(validator.ErrCodeInvalidItemType,
			fmt.Sprintf("unknown item type: %s", itemType)).
			WithDetail("valid_types", []string{"pistol", "shotgun", "rifle", "smg", "sniper", "launcher", "melee", "shield", "grenade", "artifact"})
	}

	return value, nil
}

// encodeManufacturer converts manufacturer string to numeric value
func (s *Serializer) encodeManufacturer(manufacturer string) (uint64, error) {
	manufacturerMap := map[string]uint64{
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

	value, exists := manufacturerMap[strings.ToLower(manufacturer)]
	if !exists {
		return 0, validator.NewValidationError(validator.ErrCodeInvalidManufacturer,
			fmt.Sprintf("unknown manufacturer: %s", manufacturer)).
			WithDetail("valid_manufacturers", []string{"maliwan", "jakobs", "hyperion", "torgue", "vladof", "dahl", "tediore", "atlas", "cov", "children", "anshin", "pangolin", "eridian"})
	}

	return value, nil
}

// createMetadata creates additional metadata for the serialization result
func (s *Serializer) createMetadata(itemData *models.ItemData, tokenStream *token.TokenStream, encodeResult *base85.EncodeResult) map[string]interface{} {
	metadata := make(map[string]interface{})

	// Add processing metadata
	metadata["processing_version"] = "1.0.0"
	metadata["target_version"] = s.options.TargetVersion
	metadata["token_count"] = tokenStream.Size()
	metadata["part_count"] = len(itemData.Parts)

	// Add encoding statistics if available
	if encodeResult.Statistics != nil {
		metadata["encoding_stats"] = encodeResult.Statistics
	}

	// Add optimization info if available
	if encodeResult.Optimization != nil {
		metadata["optimization"] = encodeResult.Optimization
	}

	// Add item summary
	metadata["item_summary"] = map[string]interface{}{
		"level":       itemData.Level,
		"type":        itemData.Type,
		"manufacturer": itemData.Manufacturer,
		"has_name":    itemData.Name != "",
		"has_parts":   len(itemData.Parts) > 0,
	}

	return metadata
}

// validateResult performs final validation of the serialization result
func (s *Serializer) validateResult(result *SerializationResult) error {
	if result.SerialCode == "" {
		return validator.NewValidationError(validator.ErrCodeInvalidInput,
			"serialization resulted in empty serial code")
	}

	// Validate that the serial code can be decoded back
	decoder := base85.NewDecoder()
	_, err := decoder.Decode(result.SerialCode)
	if err != nil {
		return validator.NewValidationError(validator.ErrCodeInvalidFormat,
			fmt.Sprintf("generated serial code cannot be decoded: %w", err))
	}

	return nil
}

// SerializationResult contains the complete result of serialization
type SerializationResult struct {
	Success       bool                         `json:"success"`
	SerialCode    string                       `json:"serial_code,omitempty"`
	ItemData      *models.ItemData             `json:"item_data,omitempty"`
	TokenStream   *token.TokenStream           `json:"-"`
	BitstreamData []byte                       `json:"bitstream_data,omitempty"`
	EncodingInfo  *base85.EncodeResult        `json:"encoding_info,omitempty"`
	Metadata      map[string]interface{}       `json:"metadata,omitempty"`
	StartTime     time.Time                    `json:"start_time"`
	Duration     time.Duration                `json:"duration_ms"`
	Error         error                        `json:"error,omitempty"`
}

// GetProcessingTimeMs returns processing time in milliseconds
func (r *SerializationResult) GetProcessingTimeMs() int64 {
	return r.Duration.Milliseconds()
}

// GetProcessingStats returns processing statistics
func (r *SerializationResult) GetProcessingStats() map[string]interface{} {
	stats := map[string]interface{}{
		"success":             r.Success,
		"processing_time_ms":  r.GetProcessingTimeMs(),
		"serial_code_length":  len(r.SerialCode),
	}

	if r.BitstreamData != nil {
		stats["bitstream_size"] = len(r.BitstreamData)
	}

	if r.TokenStream != nil {
		stats["token_count"] = r.TokenStream.Size()
		stats["total_bits"] = r.TokenStream.GetTotalBitSize()
	}

	if r.ItemData != nil {
		stats["parts_count"] = len(r.ItemData.Parts)
		stats["item_level"] = r.ItemData.Level
		stats["item_type"] = r.ItemData.Type
		stats["manufacturer"] = r.ItemData.Manufacturer
	}

	return stats
}

// BatchSerializer handles batch serialization operations
type BatchSerializer struct {
	serializer *Serializer
}

// NewBatchSerializer creates a new batch serializer
func NewBatchSerializer() *BatchSerializer {
	return &BatchSerializer{
		serializer: NewSerializer(),
	}
}

// NewBatchSerializerWithOptions creates a new batch serializer with options
func NewBatchSerializerWithOptions(options *SerializerOptions) *BatchSerializer {
	return &BatchSerializer{
		serializer: NewSerializerWithOptions(options),
	}
}

// SerializeBatch serializes multiple item data structures
func (bs *BatchSerializer) SerializeBatch(items []*models.ItemData) []*SerializationResult {
	results := make([]*SerializationResult, len(items))

	logger.Sugar().Infow("Starting batch serialization",
		"total_items", len(items),
	)

	for i, item := range items {
		result, err := bs.serializer.SerializeItem(item)
		results[i] = result

		if err != nil {
			logger.Sugar().Errorw("Item serialization failed",
				"index", i,
				"item_level", item.Level,
				"item_type", item.Type,
				"error", err.Error(),
			)
		} else {
			logger.Sugar().Debugw("Item serialized successfully",
				"index", i,
				"item_level", item.Level,
				"item_type", item.Type,
				"duration_ms", result.GetProcessingTimeMs(),
			)
		}
	}

	// Log batch summary
	successCount := 0
	var totalDuration time.Duration
	for _, result := range results {
		if result.Success {
			successCount++
		}
		totalDuration += result.Duration
	}

	logger.Sugar().Infow("Batch serialization completed",
		"total_items", len(items),
		"successful", successCount,
		"failed", len(items)-successCount,
		"total_duration_ms", totalDuration.Milliseconds(),
		"average_duration_ms", totalDuration.Milliseconds()/int64(len(items)),
	)

	return results
}

// GetBatchStatistics returns statistics for a batch serialization
func (bs *BatchSerializer) GetBatchStatistics(results []*SerializationResult) *BatchSerializationStats {
	stats := &BatchSerializationStats{
		Total:         len(results),
		Successful:    0,
		Failed:        0,
		TotalParts:    0,
	}

	var totalDuration time.Duration

	for _, result := range results {
		if result.Success {
			stats.Successful++
			if result.ItemData != nil {
				stats.TotalParts += len(result.ItemData.Parts)
			}
		} else {
			stats.Failed++
		}
		totalDuration += result.Duration
	}

	if stats.Successful > 0 {
		stats.AverageDuration = totalDuration.Milliseconds() / int64(stats.Successful)
		stats.AverageParts = float64(stats.TotalParts) / float64(stats.Successful)
	}

	return stats
}

// BatchSerializationStats contains statistics for batch operations
type BatchSerializationStats struct {
	Total            int     `json:"total"`
	Successful       int     `json:"successful"`
	Failed           int     `json:"failed"`
	TotalParts       int     `json:"total_parts"`
	AverageDuration  int64   `json:"average_duration_ms"`
	AverageParts     float64 `json:"average_parts"`
}