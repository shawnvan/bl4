package serial

import (
	"fmt"
	"strings"
	"time"

	"github.com/shawnvan/bl4/internal/api/models"
	"github.com/shawnvan/bl4/internal/codec/base85"
	"github.com/shawnvan/bl4/internal/codec/bitstream"
	"github.com/shawnvan/bl4/internal/codec/token"
	"github.com/shawnvan/bl4/pkg/logger"
	"github.com/shawnvan/bl4/pkg/validator"
)

// Deserializer handles the complete deserialization of BL4 item serial codes
type Deserializer struct {
	options *DeserializerOptions
}

// DeserializerOptions contains options for deserialization
type DeserializerOptions struct {
	// IncludeBitstream includes raw bitstream data in the result
	IncludeBitstream bool

	// IncludeTokens includes parsed tokens in the result
	IncludeTokens bool

	// IncludeRawData includes various raw data formats
	IncludeRawData bool

	// StrictValidation enables strict validation rules
	StrictValidation bool

	// MaxProcessingTime limits processing time in milliseconds
	MaxProcessingTime int

	// LogLevel controls logging verbosity
	LogLevel string
}

// NewDeserializer creates a new deserializer with default options
func NewDeserializer() *Deserializer {
	return &Deserializer{
		options: &DeserializerOptions{
			IncludeBitstream: false,
			IncludeTokens:    false,
			IncludeRawData:   false,
			StrictValidation: true,
			MaxProcessingTime: 5000, // 5 seconds
			LogLevel:         "info",
		},
	}
}

// NewDeserializerWithOptions creates a new deserializer with custom options
func NewDeserializerWithOptions(options *DeserializerOptions) *Deserializer {
	if options == nil {
		options = &DeserializerOptions{}
	}

	// Set defaults for missing options
	if options.MaxProcessingTime == 0 {
		options.MaxProcessingTime = 5000
	}
	if options.LogLevel == "" {
		options.LogLevel = "info"
	}

	return &Deserializer{
		options: options,
	}
}

// DeserializeItem performs complete deserialization of an item serial code
func (d *Deserializer) DeserializeItem(serialCode string) (*DeserializationResult, error) {
	startTime := time.Now()

	result := &DeserializationResult{
		SerialCode: serialCode,
		StartTime:  startTime,
	}

	logger.Sugar().Debugw("Starting item deserialization",
		"serial_code", serialCode,
		"options", d.options,
	)

	// Step 1: Base85 decode
	decodedBytes, err := d.decodeBase85(serialCode)
	if err != nil {
		result.Error = err
		result.Duration = time.Since(startTime)
		return result, fmt.Errorf("Base85 decode failed: %w", err)
	}
	result.DecodedBytes = decodedBytes

	// Step 2: Tokenize bitstream
	tokenStream, err := d.tokenizeBitstream(decodedBytes)
	if err != nil {
		result.Error = err
		result.Duration = time.Since(startTime)
		return result, fmt.Errorf("tokenization failed: %w", err)
	}
	result.TokenStream = tokenStream

	// Step 3: Extract item data
	itemData, err := d.extractItemData(tokenStream)
	if err != nil {
		result.Error = err
		result.Duration = time.Since(startTime)
		return result, fmt.Errorf("item data extraction failed: %w", err)
	}
	result.ItemData = itemData

	// Step 4: Generate additional data
	if d.options.IncludeBitstream {
		result.BitstreamInfo = d.createBitstreamInfo(decodedBytes, tokenStream)
	}

	if d.options.IncludeTokens {
		result.TokenInfos = d.createTokenInfos(tokenStream)
	}

	if d.options.IncludeRawData {
		result.RawDataInfo = d.createRawDataInfo(itemData)
	}

	// Step 5: Validate result
	if d.options.StrictValidation {
		if err := d.validateResult(result); err != nil {
			result.Error = err
			result.Duration = time.Since(startTime)
			return result, fmt.Errorf("validation failed: %w", err)
		}
	}

	result.Duration = time.Since(startTime)
	result.Success = true

	logger.Sugar().Infow("Item deserialization completed",
		"serial_code", serialCode,
		"duration_ms", result.Duration.Milliseconds(),
		"decoded_bytes", len(decodedBytes),
		"tokens", tokenStream.Size(),
	)

	return result, nil
}

// decodeBase85 performs Base85 decoding
func (d *Deserializer) decodeBase85(serialCode string) ([]byte, error) {
	decoder := base85.NewDecoder()

	options := &base85.DecodeOptions{
		IncludeStatistics: true,
		ValidateFormat:     true,
	}

	decodeResult, err := decoder.DecodeWithOptions(serialCode, options)
	if err != nil {
		return nil, err
	}

	logger.Sugar().Debugw("Base85 decode completed",
		"input_chars", decodeResult.CharCount,
		"output_bytes", decodeResult.ByteSize,
		"compression_ratio", decodeResult.Statistics.CompressionRatio,
	)

	return decodeResult.Data, nil
}

// tokenizeBitstream tokenizes the decoded bytes
func (d *Deserializer) tokenizeBitstream(data []byte) (*token.TokenStream, error) {
	reader := bitstream.NewReaderFromBytes(data)
	tokenizer := token.NewTokenizer(reader)

	tokenStream, err := tokenizer.TokenizeAll()
	if err != nil {
		return nil, fmt.Errorf("failed to tokenize bitstream: %w", err)
	}

	// Validate token stream
	if err := token.ValidateTokenStream(tokenStream); err != nil {
		if d.options.StrictValidation {
			return nil, fmt.Errorf("invalid token stream: %w", err)
		}
		logger.Sugar().Warnw("Token stream validation warnings",
			"warnings", err.Error(),
		)
	}

	logger.Sugar().Debugw("Bitstream tokenization completed",
		"total_tokens", tokenStream.Size(),
		"total_bits", tokenStream.GetTotalBitSize(),
	)

	return tokenStream, nil
}

// extractItemData extracts structured item data from tokens
func (d *Deserializer) extractItemData(tokenStream *token.TokenStream) (*models.ItemData, error) {
	itemData := &models.ItemData{
		Parts:      make([]models.PartData, 0),
		Properties: make(map[string]interface{}),
		Metadata:   make(map[string]interface{}),
	}

	// Extract header tokens (level, type, manufacturer)
	headerTokens := tokenStream.GetTokensByType(token.TokenVARINT)
	if len(headerTokens) < 3 {
		return nil, validator.NewValidationError(validator.ErrCodeInvalidItemData,
			"insufficient header tokens (need at least 3 VARINT tokens for level, type, manufacturer)").
			WithDetail("found_tokens", len(headerTokens)).
			WithDetail("required_tokens", "level, type, manufacturer")
	}

	// Parse level
	level, err := d.parseLevel(headerTokens[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse level: %w", err)
	}
	itemData.Level = level

	// Parse type
	itemType, err := d.parseType(headerTokens[1])
	if err != nil {
		return nil, fmt.Errorf("failed to parse item type: %w", err)
	}
	itemData.Type = itemType

	// Parse manufacturer
	manufacturer, err := d.parseManufacturer(headerTokens[2])
	if err != nil {
		return nil, fmt.Errorf("failed to parse manufacturer: %w", err)
	}
	itemData.Manufacturer = manufacturer

	// Extract parts
	partTokens := tokenStream.GetTokensByType(token.TokenPART)
	for i, partToken := range partTokens {
		partData, err := d.parsePart(partToken)
		if err != nil {
			if d.options.StrictValidation {
				return nil, fmt.Errorf("failed to parse part at index %d: %w", i, err)
			}
			logger.Sugar().Warnw("Part parsing failed",
				"part_index", i,
				"error", err.Error(),
			)
			continue
		}
		itemData.Parts = append(itemData.Parts, *partData)
	}

	// Extract strings (name, description, etc.)
	stringTokens := tokenStream.GetTokensByType(token.TokenSTRING)
	for i, stringToken := range stringTokens {
		if err := d.parseString(stringToken, itemData, i); err != nil {
			if d.options.StrictValidation {
				return nil, fmt.Errorf("failed to parse string at index %d: %w", i, err)
			}
			logger.Sugar().Warnw("String parsing failed",
				"string_index", i,
				"error", err.Error(),
			)
		}
	}

	// Generate raw parts string for backward compatibility
	itemData.RawParts = d.generateRawPartsString(itemData.Parts)

	// Add processing metadata
	itemData.Metadata["token_count"] = tokenStream.Size()
	itemData.Metadata["part_count"] = len(itemData.Parts)
	itemData.Metadata["processing_version"] = "1.0.0"

	logger.Sugar().Debugw("Item data extraction completed",
		"level", itemData.Level,
		"type", itemData.Type,
		"manufacturer", itemData.Manufacturer,
		"parts_count", len(itemData.Parts),
	)

	return itemData, nil
}

// parseLevel extracts level from a VARINT token
func (d *Deserializer) parseLevel(tok token.Token) (int, error) {
	if tok.Type != token.TokenVARINT {
		return 0, validator.NewValidationError(validator.ErrCodeInvalidToken,
			fmt.Sprintf("expected VARINT token for level, got %s", tok.Type.String()))
	}

	value, ok := tok.Value.(uint64)
	if !ok {
		return 0, validator.NewValidationError(validator.ErrCodeTokenValueInvalid,
			"invalid value type for level token")
	}

	if value > 100 {
		return 0, validator.NewValidationError(validator.ErrCodeInvalidItemLevel,
			fmt.Sprintf("level value out of range: %d", value)).
			WithDetail("valid_range", "0-100")
	}

	return int(value), nil
}

// parseType extracts item type from a VARINT token
func (d *Deserializer) parseType(tok token.Token) (string, error) {
	if tok.Type != token.TokenVARINT {
		return "", validator.NewValidationError(validator.ErrCodeInvalidToken,
			fmt.Sprintf("expected VARINT token for type, got %s", tok.Type.String()))
	}

	value, ok := tok.Value.(uint64)
	if !ok {
		return "", validator.NewValidationError(validator.ErrCodeTokenValueInvalid,
			"invalid value type for type token")
	}

	// Convert numeric type back to string
	typeMap := map[uint64]string{
		0:  "pistol",
		1:  "shotgun",
		2:  "rifle",
		3:  "smg",
		4:  "sniper",
		5:  "launcher",
		6:  "melee",
		7:  "shield",
		8:  "grenade",
		9:  "artifact",
	}

	itemType, exists := typeMap[value]
	if !exists {
		return "", validator.NewValidationError(validator.ErrCodeInvalidItemType,
			fmt.Sprintf("unknown item type value: %d", value)).
			WithDetail("valid_range", "0-9")
	}

	return itemType, nil
}

// parseManufacturer extracts manufacturer from a VARINT token
func (d *Deserializer) parseManufacturer(tok token.Token) (string, error) {
	if tok.Type != token.TokenVARINT {
		return "", validator.NewValidationError(validator.ErrCodeInvalidToken,
			fmt.Sprintf("expected VARINT token for manufacturer, got %s", tok.Type.String()))
	}

	value, ok := tok.Value.(uint64)
	if !ok {
		return "", validator.NewValidationError(validator.ErrCodeTokenValueInvalid,
			"invalid value type for manufacturer token")
	}

	// Convert numeric manufacturer back to string
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

	manufacturer, exists := manufacturerMap[value]
	if !exists {
		return "", validator.NewValidationError(validator.ErrCodeInvalidManufacturer,
			fmt.Sprintf("unknown manufacturer value: %d", value)).
			WithDetail("valid_range", "0-12")
	}

	return manufacturer, nil
}

// parsePart extracts part data from a PART token
func (d *Deserializer) parsePart(tok token.Token) (*models.PartData, error) {
	if tok.Type != token.TokenPART {
		return nil, validator.NewValidationError(validator.ErrCodeInvalidToken,
			fmt.Sprintf("expected PART token, got %s", tok.Type.String()))
	}

	partData, ok := tok.Value.(map[string]uint64)
	if !ok {
		return nil, validator.NewValidationError(validator.ErrCodeTokenValueInvalid,
			"invalid value type for part token")
	}

	index, exists := partData["index"]
	if !exists {
		return nil, validator.NewValidationError(validator.ErrCodeInvalidPartsData,
			"missing index in part data")
	}

	value, exists := partData["value"]
	if !exists {
		return nil, validator.NewValidationError(validator.ErrCodeInvalidPartsData,
			"missing value in part data")
	}

	// Create part data
	part := &models.PartData{
		Index:      int(index),
		Value:      int(value),
		RawValue:   fmt.Sprintf("%d", value),
		Attributes: make(map[string]interface{}),
		Metadata:   make(map[string]interface{}),
	}

	// Extract level from value (high byte)
	level := (value >> 8) & 0xFF
	if level > 0 {
		part.Attributes["level"] = int(level)
	}

	// Extract actual part value (low byte)
	partValue := value & 0xFF
	part.Attributes["part_value"] = int(partValue)

	// Determine part type based on index
	part.Type = d.resolvePartType(index)
	part.Name = d.generatePartName(index, value)

	// Add token metadata
	part.Metadata["token_position"] = tok.Position
	part.Metadata["token_bit_size"] = tok.BitSize

	return part, nil
}

// parseString extracts string data from a STRING token
func (d *Deserializer) parseString(tok token.Token, itemData *models.ItemData, index int) error {
	if tok.Type != token.TokenSTRING {
		return validator.NewValidationError(validator.ErrCodeInvalidToken,
			fmt.Sprintf("expected STRING token, got %s", tok.Type.String()))
	}

	value, ok := tok.Value.(string)
	if !ok {
		return validator.NewValidationError(validator.ErrCodeTokenValueInvalid,
			"invalid value type for string token")
	}

	// Determine string type based on position and content
	if index == 0 {
		// First string is likely the item name
		itemData.Name = value
	} else if index == 1 {
		// Second string might be description
		itemData.Description = value
	} else {
		// Additional strings could be various properties
		if strings.Contains(strings.ToLower(value), "rarity") {
			itemData.Rarity = value
		} else if strings.Contains(strings.ToLower(value), "version") {
			itemData.GameVersion = value
		}
	}

	return nil
}

// resolvePartType determines the part type based on index
func (d *Deserializer) resolvePartType(index uint64) string {
	switch {
	case index >= 0 && index <= 4:
		return "barrel"
	case index >= 5 && index <= 8:
		return "grip"
	case index >= 9 && index <= 12:
		return "sight"
	case index >= 13 && index <= 16:
		return "stock"
	case index >= 17 && index <= 20:
		return "magazine"
	case index >= 21 && index <= 24:
		return "accessory"
	case index >= 25 && index <= 28:
		return "material"
	case index >= 29 && index <= 32:
		return "element"
	default:
		return "special"
	}
}

// generatePartName generates a human-readable part name
func (d *Deserializer) generatePartName(index uint64, value uint64) string {
	partType := d.resolvePartType(index)
	level := (value >> 8) & 0xFF
	partValue := value & 0xFF

	return fmt.Sprintf("%s (Level %d, Value %d)", strings.Title(partType), level, partValue)
}

// generateRawPartsString creates the raw parts string for backward compatibility
func (d *Deserializer) generateRawPartsString(partList []models.PartData) string {
	if len(partList) == 0 {
		return ""
	}

	var partsStr []string
	for _, part := range partList {
		partsStr = append(partsStr, fmt.Sprintf("%d, %d", part.Index, part.Value))
	}

	if len(partsStr) == 0 {
		return ""
	}

	return fmt.Sprintf("%s|| {%s}", strings.Join(partsStr, "|"), partsStr[len(partsStr)-1])
}

// createBitstreamInfo creates bitstream information
func (d *Deserializer) createBitstreamInfo(data []byte, tokenStream *token.TokenStream) *models.BitstreamInfo {
	info := &models.BitstreamInfo{
		Size:     tokenStream.GetTotalBitSize(),
		ByteSize: len(data),
		Format:   "bl4-v1.0",
		Statistics: &models.BitstreamStats{
			TokenCount:   tokenStream.Size(),
			TokenTypes:   make(map[string]int),
			TotalValues:  tokenStream.Size(),
			UniqueValues: calculateUniqueValues(tokenStream),
		},
	}

	if d.options.IncludeBitstream {
		info.Data = data
	}

	// Calculate token type distribution
	for _, token := range tokenStream.Tokens {
		tokenType := token.Type.String()
		info.Statistics.TokenTypes[tokenType]++
	}

	return info
}

// createTokenInfos creates token information
func (d *Deserializer) createTokenInfos(tokenStream *token.TokenStream) []models.TokenInfo {
	infos := make([]models.TokenInfo, len(tokenStream.Tokens))

	for i, token := range tokenStream.Tokens {
		infos[i] = models.TokenInfo{
			Type:     token.Type.String(),
			Value:    formatTokenValue(token),
			BitSize:  token.BitSize,
			Position: token.Position,
			Properties: map[string]interface{}{
				"is_data":       token.IsDataToken(),
				"is_structural": token.IsStructuralToken(),
			},
		}
	}

	return infos
}

// createRawDataInfo creates raw data information
func (d *Deserializer) createRawDataInfo(itemData *models.ItemData) *models.RawDataInfo {
	info := &models.RawDataInfo{
		Structured: itemData.RawParts,
		Properties: make(map[string]interface{}),
	}

	// Add item summary to properties
	info.Properties["level"] = itemData.Level
	info.Properties["type"] = itemData.Type
	info.Properties["manufacturer"] = itemData.Manufacturer
	info.Properties["parts_count"] = len(itemData.Parts)

	return info
}

// validateResult performs final validation of the deserialization result
func (d *Deserializer) validateResult(result *DeserializationResult) error {
	if result.ItemData == nil {
		return validator.NewValidationError(validator.ErrCodeInvalidItemData,
			"deserialized item data is nil")
	}

	// Validate required fields
	if result.ItemData.Level < 0 || result.ItemData.Level > 100 {
		return validator.NewValidationError(validator.ErrCodeInvalidItemLevel,
			fmt.Sprintf("invalid item level: %d", result.ItemData.Level)).
			WithDetail("valid_range", "0-100")
	}

	if result.ItemData.Type == "" {
		return validator.NewValidationError(validator.ErrCodeInvalidItemType,
			"item type is empty")
	}

	if result.ItemData.Manufacturer == "" {
		return validator.NewValidationError(validator.ErrCodeInvalidManufacturer,
			"manufacturer is empty")
	}

	// Validate parts structure
	if len(result.ItemData.Parts) > 50 {
		return validator.NewValidationError(validator.ErrCodeInvalidPartsData,
			fmt.Sprintf("too many parts: %d", len(result.ItemData.Parts))).
			WithDetail("max_parts", 50)
	}

	// Check for duplicate part indices
	seenIndices := make(map[int]bool)
	for _, part := range result.ItemData.Parts {
		if seenIndices[part.Index] {
			return validator.NewValidationError(validator.ErrCodeInvalidPartsData,
				fmt.Sprintf("duplicate part index: %d", part.Index))
		}
		seenIndices[part.Index] = true
	}

	return nil
}

// DeserializationResult contains the complete result of deserialization
type DeserializationResult struct {
	Success       bool                    `json:"success"`
	SerialCode    string                 `json:"serial_code"`
	ItemData      *models.ItemData        `json:"item_data,omitempty"`
	DecodedBytes  []byte                 `json:"decoded_bytes,omitempty"`
	TokenStream   *token.TokenStream      `json:"-"`
	BitstreamInfo *models.BitstreamInfo   `json:"bitstream_info,omitempty"`
	TokenInfos    []models.TokenInfo      `json:"token_infos,omitempty"`
	RawDataInfo   *models.RawDataInfo     `json:"raw_data_info,omitempty"`
	StartTime     time.Time              `json:"start_time"`
	Duration     time.Duration          `json:"duration_ms"`
	Error         error                  `json:"error,omitempty"`
}

// GetProcessingTimeMs returns processing time in milliseconds
func (r *DeserializationResult) GetProcessingTimeMs() int64 {
	return r.Duration.Milliseconds()
}

// GetProcessingStats returns processing statistics
func (r *DeserializationResult) GetProcessingStats() map[string]interface{} {
	stats := map[string]interface{}{
		"success":             r.Success,
		"processing_time_ms":  r.GetProcessingTimeMs(),
		"decoded_bytes":       len(r.DecodedBytes),
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

// BatchDeserializer handles batch deserialization operations
type BatchDeserializer struct {
	deserializer *Deserializer
}

// NewBatchDeserializer creates a new batch deserializer
func NewBatchDeserializer() *BatchDeserializer {
	return &BatchDeserializer{
		deserializer: NewDeserializer(),
	}
}

// NewBatchDeserializerWithOptions creates a new batch deserializer with options
func NewBatchDeserializerWithOptions(options *DeserializerOptions) *BatchDeserializer {
	return &BatchDeserializer{
		deserializer: NewDeserializerWithOptions(options),
	}
}

// DeserializeBatch deserializes multiple item serial codes
func (bd *BatchDeserializer) DeserializeBatch(serialCodes []string) []*DeserializationResult {
	results := make([]*DeserializationResult, len(serialCodes))

	logger.Sugar().Infow("Starting batch deserialization",
		"total_codes", len(serialCodes),
	)

	for i, code := range serialCodes {
		result, err := bd.deserializer.DeserializeItem(code)
		results[i] = result

		if err != nil {
			logger.Sugar().Errorw("Item deserialization failed",
				"index", i,
				"serial_code", code,
				"error", err.Error(),
			)
		} else {
			logger.Sugar().Debugw("Item deserialized successfully",
				"index", i,
				"serial_code", code,
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

	logger.Sugar().Infow("Batch deserialization completed",
		"total_codes", len(serialCodes),
		"successful", successCount,
		"failed", len(serialCodes)-successCount,
		"total_duration_ms", totalDuration.Milliseconds(),
		"average_duration_ms", totalDuration.Milliseconds()/int64(len(serialCodes)),
	)

	return results
}

// GetBatchStatistics returns statistics for a batch deserialization
func (bd *BatchDeserializer) GetBatchStatistics(results []*DeserializationResult) *BatchDeserializationStats {
	stats := &BatchDeserializationStats{
		Total:       len(results),
		Successful:  0,
		Failed:      0,
		TotalBytes:  0,
		TotalParts:  0,
	}

	var totalDuration time.Duration

	for _, result := range results {
		if result.Success {
			stats.Successful++
			stats.TotalBytes += len(result.DecodedBytes)
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
		stats.AverageBytes = float64(stats.TotalBytes) / float64(stats.Successful)
		stats.AverageParts = float64(stats.TotalParts) / float64(stats.Successful)
	}

	return stats
}

// BatchDeserializationStats contains statistics for batch operations
type BatchDeserializationStats struct {
	Total            int     `json:"total"`
	Successful       int     `json:"successful"`
	Failed           int     `json:"failed"`
	TotalBytes       int     `json:"total_bytes"`
	TotalParts       int     `json:"total_parts"`
	AverageDuration  int64   `json:"average_duration_ms"`
	AverageBytes     float64 `json:"average_bytes"`
	AverageParts     float64 `json:"average_parts"`
}

// Helper functions

// formatTokenValue formats a token value for display
func formatTokenValue(tok token.Token) interface{} {
	switch tok.Type {
	case token.TokenVARBIT:
		if bits, ok := tok.Value.([]bool); ok {
			var bitStr string
			for i, bit := range bits {
				if i > 0 && i%8 == 0 {
					bitStr += " "
				}
				if bit {
					bitStr += "1"
				} else {
					bitStr += "0"
				}
			}
			return bitStr
		}
	case token.TokenSTRING:
		return tok.Value
	case token.TokenPART:
		if partData, ok := tok.Value.(map[string]uint64); ok {
			return fmt.Sprintf("index:%d,value:%d", partData["index"], partData["value"])
		}
	}
	return tok.Value
}

// calculateUniqueValues counts unique values in a token stream
func calculateUniqueValues(tokenStream *token.TokenStream) int {
	seen := make(map[string]bool)
	for _, token := range tokenStream.Tokens {
		key := fmt.Sprintf("%v", token.Value) // Convert to string for hashing
		if _, exists := seen[key]; !exists {
			seen[key] = true
		}
	}
	return len(seen)
}