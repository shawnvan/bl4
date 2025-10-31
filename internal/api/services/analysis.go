package services

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/shawnvan/bl4/internal/api/models"
	"github.com/shawnvan/bl4/internal/codec/base85"
	"github.com/shawnvan/bl4/pkg/logger"
)

// AnalysisService provides pattern analysis and statistical insights for BL4 items
type AnalysisService struct {
	// Cache for analysis results to improve performance
	analysisCache map[string]*AnalysisResult
	cacheMutex    sync.RWMutex

	// Known patterns database
	patterns map[string]*ItemPattern

	// Statistical aggregators
	stats *ItemStatistics

	// Configuration
	config *AnalysisConfig
}

// AnalysisConfig contains configuration for analysis operations
type AnalysisConfig struct {
	// Enable caching of analysis results
	EnableCache bool

	// Maximum cache size
	MaxCacheSize int

	// Cache TTL in seconds
	CacheTTL int

	// Minimum sample size for statistical analysis
	MinSampleSize int

	// Enable deep analysis (more CPU intensive)
	EnableDeepAnalysis bool

	// Pattern matching sensitivity (0.0 to 1.0)
	PatternSensitivity float64
}

// DefaultAnalysisConfig returns default configuration for analysis service
func DefaultAnalysisConfig() *AnalysisConfig {
	return &AnalysisConfig{
		EnableCache:         true,
		MaxCacheSize:        1000,
		CacheTTL:            3600, // 1 hour
		MinSampleSize:       10,
		EnableDeepAnalysis:  true,
		PatternSensitivity:  0.8,
	}
}

// NewAnalysisService creates a new analysis service instance
func NewAnalysisService(config *AnalysisConfig) *AnalysisService {
	if config == nil {
		config = DefaultAnalysisConfig()
	}

	service := &AnalysisService{
		analysisCache: make(map[string]*AnalysisResult),
		patterns:      make(map[string]*ItemPattern),
		stats:         NewItemStatistics(),
		config:        config,
	}

	// Initialize known patterns
	service.initializePatterns()

	return service
}

// AnalysisRequest represents a request for pattern analysis
type AnalysisRequest struct {
	// Serial codes to analyze
	SerialCodes []string `json:"serial_codes"`

	// Analysis options
	Options *AnalysisOptions `json:"options,omitempty"`
}

// AnalysisOptions contains options for analysis operations
type AnalysisOptions struct {
	// Include pattern matching
	IncludePatterns bool

	// Include statistical analysis
	IncludeStatistics bool

	// Include bitstream analysis
	IncludeBitstream bool

	// Include rarity analysis
	IncludeRarity bool

	// Include manufacturer analysis
	IncludeManufacturers bool

	// Include part frequency analysis
	IncludePartFrequency bool

	// Deep analysis mode
	DeepAnalysis bool

	// Maximum number of patterns to return
	MaxPatterns int

	// Confidence threshold for pattern matching
	ConfidenceThreshold float64
}

// AnalysisResult contains the results of pattern analysis
type AnalysisResult struct {
	// Analysis metadata
	RequestID   string    `json:"request_id"`
	Timestamp   time.Time `json:"timestamp"`
	SampleSize  int       `json:"sample_size"`
	ProcessTime string    `json:"process_time"`

	// Pattern analysis results
	Patterns []*PatternMatch `json:"patterns,omitempty"`

	// Statistical analysis
	Statistics *ItemStatistics `json:"statistics,omitempty"`

	// Bitstream analysis
	BitstreamAnalysis *BitstreamAnalysis `json:"bitstream_analysis,omitempty"`

	// Rarity distribution
	RarityDistribution map[string]int `json:"rarity_distribution,omitempty"`

	// Manufacturer distribution
	ManufacturerDistribution map[string]int `json:"manufacturer_distribution,omitempty"`

	// Part frequency analysis
	PartFrequency map[string]*PartFrequency `json:"part_frequency,omitempty"`

	// Quality metrics
	QualityScore float64 `json:"quality_score"`
	Confidence   float64 `json:"confidence"`

	// Insights and recommendations
	Insights []string `json:"insights,omitempty"`
}

// PatternMatch represents a matched pattern
type PatternMatch struct {
	PatternName   string  `json:"pattern_name"`
	PatternType   string  `json:"pattern_type"`
	Confidence    float64 `json:"confidence"`
	MatchCount    int     `json:"match_count"`
	TotalSamples  int     `json:"total_samples"`
	Percentage    float64 `json:"percentage"`
	Description   string  `json:"description"`
	Examples      []string `json:"examples,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// ItemPattern represents a known item pattern
type ItemPattern struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Signature   string                 `json:"signature"`
	Weight      float64                `json:"weight"`
	Rarity      string                 `json:"rarity"`
	Attributes  map[string]interface{} `json:"attributes"`
}

// ItemStatistics contains statistical analysis of items
type ItemStatistics struct {
	// Basic statistics
	TotalItems     int     `json:"total_items"`
	UniqueTypes    int     `json:"unique_types"`
	UniqueManufacturers int `json:"unique_manufacturers"`

	// Level statistics
	AverageLevel   float64 `json:"average_level"`
	MinLevel       int     `json:"min_level"`
	MaxLevel       int     `json:"max_level"`
	LevelDistribution map[int]int `json:"level_distribution"`

	// Part statistics
	AverageParts   float64 `json:"average_parts"`
	MinParts       int     `json:"min_parts"`
	MaxParts       int     `json:"max_parts"`
	TotalParts     int     `json:"total_parts"`

	// Quality metrics
	DataQuality    float64 `json:"data_quality"`
	Completeness   float64 `json:"completeness"`
	Consistency    float64 `json:"consistency"`
}

// BitstreamAnalysis contains analysis of bitstream patterns
type BitstreamAnalysis struct {
	// Pattern analysis
	CommonPrefixes    []string `json:"common_prefixes"`
	CommonSuffixes    []string `json:"common_suffixes"`
	RepeatingPatterns []string `json:"repeating_patterns"`

	// Entropy analysis
	Entropy           float64 `json:"entropy"`
	UniqueBytes       int     `json:"unique_bytes"`
	TotalBytes        int     `json:"total_bytes"`

	// Structure analysis
	AverageLength     float64 `json:"average_length"`
	LengthVariance    float64 `json:"length_variance"`
	PaddingPatterns   []string `json:"padding_patterns"`
}

// PartFrequency represents frequency analysis of parts
type PartFrequency struct {
	PartType      string  `json:"part_type"`
	Count         int     `json:"count"`
	Percentage    float64 `json:"percentage"`
	AverageValue  float64 `json:"average_value"`
	MinValue      int     `json:"min_value"`
	MaxValue      int     `json:"max_value"`
	CommonValues  []int   `json:"common_values"`
}

// AnalyzePatterns performs pattern analysis on the provided serial codes
func (s *AnalysisService) AnalyzePatterns(request *AnalysisRequest) (*AnalysisResult, error) {
	startTime := time.Now()
	requestID := generateRequestID()

	logger.Sugar().Infow("Starting pattern analysis",
		"request_id", requestID,
		"sample_size", len(request.SerialCodes),
	)

	// Validate request
	if len(request.SerialCodes) == 0 {
		return nil, fmt.Errorf("no serial codes provided for analysis")
	}

	// Check cache if enabled
	if s.config.EnableCache {
		cacheKey := s.generateCacheKey(request)
		if cached := s.getCachedResult(cacheKey); cached != nil {
			logger.Sugar().Infow("Returning cached analysis result",
				"request_id", requestID,
				"cache_key", cacheKey,
			)
			return cached, nil
		}
	}

	// Initialize result
	result := &AnalysisResult{
		RequestID:  requestID,
		Timestamp:  startTime,
		SampleSize: len(request.SerialCodes),
	}

	// Decode all serial codes
	decodedItems := make([]*models.ItemData, 0, len(request.SerialCodes))
	bitstreams := make([][]byte, 0, len(request.SerialCodes))

	for _, serialCode := range request.SerialCodes {
		// Validate and decode
		if !strings.HasPrefix(serialCode, "@") {
			continue // Skip invalid codes
		}

		b85Part := serialCode[1:]
		decoder := base85.NewDecoder()
		bitstream, err := decoder.Decode(b85Part)
		if err != nil {
			logger.Sugar().Debugw("Failed to decode serial code",
				"request_id", requestID,
				"serial_code", serialCode,
				"error", err,
			)
			continue
		}

		bitstreams = append(bitstreams, bitstream)

		// For now, create a mock item data
		// In a real implementation, this would use the actual codec
		itemData := &models.ItemData{
			Level:       20 + (len(decodedItems) % 60), // Mock level
			Type:        "pistol",                       // Mock type
			Manufacturer: "maliwan",                     // Mock manufacturer
			Parts:       []models.PartData{},            // Mock parts
		}
		decodedItems = append(decodedItems, itemData)
	}

	// Perform analysis based on options
	if request.Options == nil {
		request.Options = &AnalysisOptions{
			IncludePatterns:      true,
			IncludeStatistics:    true,
			IncludeBitstream:     true,
			IncludeRarity:        true,
			IncludeManufacturers: true,
			IncludePartFrequency: true,
			DeepAnalysis:         true,
			MaxPatterns:          10,
			ConfidenceThreshold:  0.7,
		}
	}

	// Pattern matching
	if request.Options.IncludePatterns {
		result.Patterns = s.findPatterns(decodedItems, bitstreams, request.Options)
	}

	// Statistical analysis
	if request.Options.IncludeStatistics {
		result.Statistics = s.calculateStatistics(decodedItems)
	}

	// Bitstream analysis
	if request.Options.IncludeBitstream && len(bitstreams) > 0 {
		result.BitstreamAnalysis = s.analyzeBitstreams(bitstreams)
	}

	// Rarity distribution (mock data for now)
	if request.Options.IncludeRarity {
		result.RarityDistribution = map[string]int{
			"common":    40,
			"uncommon":  30,
			"rare":      20,
			"epic":      8,
			"legendary": 2,
		}
	}

	// Manufacturer distribution
	if request.Options.IncludeManufacturers {
		result.ManufacturerDistribution = s.calculateManufacturerDistribution(decodedItems)
	}

	// Part frequency analysis
	if request.Options.IncludePartFrequency {
		result.PartFrequency = s.calculatePartFrequency(decodedItems)
	}

	// Calculate quality metrics
	result.QualityScore = s.calculateQualityScore(decodedItems, bitstreams)
	result.Confidence = s.calculateConfidence(result)

	// Generate insights
	result.Insights = s.generateInsights(result)

	// Set processing time
	result.ProcessTime = time.Since(startTime).String()

	// Cache result if enabled
	if s.config.EnableCache {
		s.cacheResult(request, result)
	}

	logger.Sugar().Infow("Pattern analysis completed",
		"request_id", requestID,
		"sample_size", result.SampleSize,
		"patterns_found", len(result.Patterns),
		"process_time", result.ProcessTime,
	)

	return result, nil
}

// findPatterns identifies patterns in the provided items and bitstreams
func (s *AnalysisService) findPatterns(items []*models.ItemData, bitstreams [][]byte, options *AnalysisOptions) []*PatternMatch {
	matches := make([]*PatternMatch, 0)

	// Pattern matching logic
	for _, pattern := range s.getKnownPatterns() {
		match := s.matchPattern(pattern, items, bitstreams, options)
		if match != nil && match.Confidence >= options.ConfidenceThreshold {
			matches = append(matches, match)
		}
	}

	// Sort by confidence
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Confidence > matches[j].Confidence
	})

	// Limit results
	if options.MaxPatterns > 0 && len(matches) > options.MaxPatterns {
		matches = matches[:options.MaxPatterns]
	}

	return matches
}

// matchPattern attempts to match a specific pattern
func (s *AnalysisService) matchPattern(pattern *ItemPattern, items []*models.ItemData, bitstreams [][]byte, options *AnalysisOptions) *PatternMatch {
	// This is a simplified pattern matching implementation
	// In a real implementation, this would use sophisticated pattern recognition algorithms

	matchCount := 0
	examples := make([]string, 0)

	for i, item := range items {
		// Simple pattern matching based on attributes
		if s.itemMatchesPattern(item, pattern) {
			matchCount++
			if len(examples) < 5 {
				examples = append(examples, fmt.Sprintf("Item %d", i+1))
			}
		}
	}

	if matchCount == 0 {
		return nil
	}

	confidence := float64(matchCount) / float64(len(items))
	confidence *= pattern.Weight // Apply pattern weight

	return &PatternMatch{
		PatternName:  pattern.Name,
		PatternType:  pattern.Type,
		Confidence:   confidence,
		MatchCount:   matchCount,
		TotalSamples: len(items),
		Percentage:   confidence * 100,
		Description:  pattern.Description,
		Examples:     examples,
		Metadata: map[string]interface{}{
			"weight":  pattern.Weight,
			"rarity":  pattern.Rarity,
		},
	}
}

// itemMatchesPattern checks if an item matches a pattern
func (s *AnalysisService) itemMatchesPattern(item *models.ItemData, pattern *ItemPattern) bool {
	// Simplified pattern matching logic
	// In a real implementation, this would be much more sophisticated

	switch pattern.Type {
	case "manufacturer":
		return strings.EqualFold(item.Manufacturer, pattern.Name)
	case "item_type":
		return strings.EqualFold(item.Type, pattern.Name)
	case "level_range":
		// Pattern would contain level range info
		return item.Level >= 20 && item.Level <= 30
	case "part_pattern":
		// Check for specific part configurations
		return len(item.Parts) > 0
	default:
		return false
	}
}

// calculateStatistics computes statistical information about the items
func (s *AnalysisService) calculateStatistics(items []*models.ItemData) *ItemStatistics {
	if len(items) == 0 {
		return &ItemStatistics{}
	}

	stats := &ItemStatistics{
		TotalItems:  len(items),
		LevelDistribution: make(map[int]int),
	}

	// Track unique values
	types := make(map[string]bool)
	manufacturers := make(map[string]bool)
	totalParts := 0
	sumLevels := 0
	minLevel := math.MaxInt32
	maxLevel := 0

	for _, item := range items {
		// Track types and manufacturers
		types[item.Type] = true
		manufacturers[item.Manufacturer] = true

		// Level statistics
		sumLevels += item.Level
		if item.Level < minLevel {
			minLevel = item.Level
		}
		if item.Level > maxLevel {
			maxLevel = item.Level
		}
		stats.LevelDistribution[item.Level]++

		// Part statistics
		totalParts += len(item.Parts)
	}

	// Calculate derived statistics
	stats.UniqueTypes = len(types)
	stats.UniqueManufacturers = len(manufacturers)
	stats.AverageLevel = float64(sumLevels) / float64(len(items))
	stats.MinLevel = minLevel
	stats.MaxLevel = maxLevel
	stats.AverageParts = float64(totalParts) / float64(len(items))
	stats.MinParts = 0 // Would be calculated from actual data
	stats.MaxParts = 10 // Would be calculated from actual data
	stats.TotalParts = totalParts

	// Quality metrics
	stats.DataQuality = s.calculateDataQuality(items)
	stats.Completeness = s.calculateCompleteness(items)
	stats.Consistency = s.calculateConsistency(items)

	return stats
}

// analyzeBitstreams analyzes bitstream patterns
func (s *AnalysisService) analyzeBitstreams(bitstreams [][]byte) *BitstreamAnalysis {
	if len(bitstreams) == 0 {
		return &BitstreamAnalysis{}
	}

	// Track unique bytes separately
	uniqueBytesMap := make(map[byte]bool)

	analysis := &BitstreamAnalysis{
		TotalBytes: 0,
	}

	totalLength := 0
	lengths := make([]int, len(bitstreams))

	for i, bitstream := range bitstreams {
		length := len(bitstream)
		lengths[i] = length
		totalLength += length
		analysis.TotalBytes += length

		// Count unique bytes
		for _, b := range bitstream {
			uniqueBytesMap[b] = true
		}
	}

	analysis.UniqueBytes = len(uniqueBytesMap)
	analysis.AverageLength = float64(totalLength) / float64(len(bitstreams))

	// Calculate length variance
	if len(bitstreams) > 1 {
		mean := analysis.AverageLength
		variance := 0.0
		for _, length := range lengths {
			diff := float64(length) - mean
			variance += diff * diff
		}
		analysis.LengthVariance = variance / float64(len(bitstreams))
	}

	// Calculate entropy
	analysis.Entropy = s.calculateEntropy(bitstreams)

	// Find common patterns
	analysis.CommonPrefixes = s.findCommonPrefixes(bitstreams)
	analysis.CommonSuffixes = s.findCommonSuffixes(bitstreams)
	analysis.RepeatingPatterns = s.findRepeatingPatterns(bitstreams)

	return analysis
}

// calculateEntropy calculates the entropy of the bitstreams
func (s *AnalysisService) calculateEntropy(bitstreams [][]byte) float64 {
	// Combine all bitstreams for entropy calculation
	allBytes := make([]byte, 0)
	for _, bs := range bitstreams {
		allBytes = append(allBytes, bs...)
	}

	if len(allBytes) == 0 {
		return 0
	}

	// Count frequency of each byte
	freq := make(map[byte]int)
	for _, b := range allBytes {
		freq[b]++
	}

	// Calculate Shannon entropy
	entropy := 0.0
	total := float64(len(allBytes))

	for _, count := range freq {
		if count > 0 {
			p := float64(count) / total
			entropy -= p * math.Log2(p)
		}
	}

	return entropy
}

// findCommonPrefixes finds common byte prefixes in bitstreams
func (s *AnalysisService) findCommonPrefixes(bitstreams [][]byte) []string {
	if len(bitstreams) < 2 {
		return nil
	}

	prefixes := make([]string, 0)

	// Find common prefixes of different lengths
	maxPrefixLength := 8
	for length := 1; length <= maxPrefixLength; length++ {
		prefixMap := make(map[string]int)

		for _, bs := range bitstreams {
			if len(bs) >= length {
				prefix := string(bs[:length])
				prefixMap[prefix]++
			}
		}

		// Find prefixes that appear in multiple bitstreams
		for prefix, count := range prefixMap {
			if count >= 2 && count >= len(bitstreams)/3 {
				prefixes = append(prefixes, fmt.Sprintf("%s (%d occurrences)", prefix, count))
			}
		}
	}

	return prefixes
}

// findCommonSuffixes finds common byte suffixes in bitstreams
func (s *AnalysisService) findCommonSuffixes(bitstreams [][]byte) []string {
	if len(bitstreams) < 2 {
		return nil
	}

	suffixes := make([]string, 0)

	// Find common suffixes of different lengths
	maxSuffixLength := 8
	for length := 1; length <= maxSuffixLength; length++ {
		suffixMap := make(map[string]int)

		for _, bs := range bitstreams {
			if len(bs) >= length {
				suffix := string(bs[len(bs)-length:])
				suffixMap[suffix]++
			}
		}

		// Find suffixes that appear in multiple bitstreams
		for suffix, count := range suffixMap {
			if count >= 2 && count >= len(bitstreams)/3 {
				suffixes = append(suffixes, fmt.Sprintf("%s (%d occurrences)", suffix, count))
			}
		}
	}

	return suffixes
}

// findRepeatingPatterns finds repeating byte patterns
func (s *AnalysisService) findRepeatingPatterns(bitstreams [][]byte) []string {
	patterns := make([]string, 0)

	for _, bs := range bitstreams {
		// Look for repeated 2-byte and 4-byte patterns
		patternLengths := []int{2, 4}

		for _, length := range patternLengths {
			if len(bs) < length*2 {
				continue
			}

			for i := 0; i <= len(bs)-length*2; i++ {
				pattern := string(bs[i : i+length])

				// Check if pattern repeats
				repeatCount := 1
				for j := i + length; j <= len(bs)-length; j += length {
					if string(bs[j:j+length]) == pattern {
						repeatCount++
					} else {
						break
					}
				}

				if repeatCount >= 3 {
					patterns = append(patterns, fmt.Sprintf("%s (repeats %d times)", pattern, repeatCount))
					break // Move to next pattern
				}
			}
		}
	}

	return patterns
}

// calculateManufacturerDistribution calculates manufacturer distribution
func (s *AnalysisService) calculateManufacturerDistribution(items []*models.ItemData) map[string]int {
	distribution := make(map[string]int)

	for _, item := range items {
		distribution[item.Manufacturer]++
	}

	return distribution
}

// calculatePartFrequency calculates part frequency statistics
func (s *AnalysisService) calculatePartFrequency(items []*models.ItemData) map[string]*PartFrequency {
	partStats := make(map[string]*PartFrequency)

	for _, item := range items {
		for _, part := range item.Parts {
			partType := part.Type
			if partType == "" {
				partType = fmt.Sprintf("part_%d", part.Index)
			}

			if _, exists := partStats[partType]; !exists {
				partStats[partType] = &PartFrequency{
					PartType:     partType,
					Count:        0,
					AverageValue: 0,
					MinValue:     math.MaxInt32,
					MaxValue:     0,
					CommonValues: []int{},
				}
			}

			stat := partStats[partType]
			stat.Count++
			stat.AverageValue = (stat.AverageValue*float64(stat.Count-1) + float64(part.Value)) / float64(stat.Count)

			if part.Value < stat.MinValue {
				stat.MinValue = part.Value
			}
			if part.Value > stat.MaxValue {
				stat.MaxValue = part.Value
			}
		}
	}

	// Calculate percentages
	totalParts := 0
	for _, stat := range partStats {
		totalParts += stat.Count
	}

	for _, stat := range partStats {
		stat.Percentage = (float64(stat.Count) / float64(totalParts)) * 100
	}

	return partStats
}

// calculateQualityScore calculates overall quality score
func (s *AnalysisService) calculateQualityScore(items []*models.ItemData, bitstreams [][]byte) float64 {
	if len(items) == 0 {
		return 0
	}

	score := 0.0

	// Data completeness (40%)
	completeness := s.calculateCompleteness(items)
	score += completeness * 0.4

	// Data consistency (30%)
	consistency := s.calculateConsistency(items)
	score += consistency * 0.3

	// Pattern strength (20%)
	patternStrength := s.calculatePatternStrength(items)
	score += patternStrength * 0.2

	// Bitstream quality (10%)
	bitstreamQuality := s.calculateBitstreamQuality(bitstreams)
	score += bitstreamQuality * 0.1

	return math.Min(score, 1.0)
}

// calculateCompleteness calculates data completeness
func (s *AnalysisService) calculateCompleteness(items []*models.ItemData) float64 {
	if len(items) == 0 {
		return 0
	}

	totalFields := 0
	completeFields := 0

	for _, item := range items {
		// Check essential fields
		if item.Level > 0 {
			completeFields++
		}
		totalFields++

		if item.Type != "" {
			completeFields++
		}
		totalFields++

		if item.Manufacturer != "" {
			completeFields++
		}
		totalFields++

		if len(item.Parts) > 0 {
			completeFields++
		}
		totalFields++
	}

	return float64(completeFields) / float64(totalFields)
}

// calculateConsistency calculates data consistency
func (s *AnalysisService) calculateConsistency(items []*models.ItemData) float64 {
	if len(items) < 2 {
		return 1.0
	}

	// Check for consistency in data formats and values
	consistencyScore := 1.0

	// Level consistency
	levels := make([]int, len(items))
	for i, item := range items {
		levels[i] = item.Level
	}
	levelVariance := s.calculateVariance(levels)
	if levelVariance > 100 {
		consistencyScore *= 0.9
	}

	return consistencyScore
}

// calculateVariance calculates variance of a slice of integers
func (s *AnalysisService) calculateVariance(values []int) float64 {
	if len(values) == 0 {
		return 0
	}

	mean := 0.0
	for _, v := range values {
		mean += float64(v)
	}
	mean /= float64(len(values))

	variance := 0.0
	for _, v := range values {
		diff := float64(v) - mean
		variance += diff * diff
	}

	return variance / float64(len(values))
}

// calculatePatternStrength calculates how strong the patterns are
func (s *AnalysisService) calculatePatternStrength(items []*models.ItemData) float64 {
	// This is a simplified implementation
	// In practice, this would analyze the strength and frequency of patterns
	if len(items) == 0 {
		return 0
	}

	// Check for common patterns
	manufacturerCounts := make(map[string]int)
	typeCounts := make(map[string]int)

	for _, item := range items {
		manufacturerCounts[item.Manufacturer]++
		typeCounts[item.Type]++
	}

	// Calculate pattern strength based on distribution
	maxManufacturerCount := 0
	for _, count := range manufacturerCounts {
		if count > maxManufacturerCount {
			maxManufacturerCount = count
		}
	}

	maxTypeCount := 0
	for _, count := range typeCounts {
		if count > maxTypeCount {
			maxTypeCount = count
		}
	}

	manufacturerStrength := float64(maxManufacturerCount) / float64(len(items))
	typeStrength := float64(maxTypeCount) / float64(len(items))

	return (manufacturerStrength + typeStrength) / 2
}

// calculateBitstreamQuality calculates bitstream quality metrics
func (s *AnalysisService) calculateBitstreamQuality(bitstreams [][]byte) float64 {
	if len(bitstreams) == 0 {
		return 0
	}

	quality := 1.0

	// Check for consistent lengths
	lengths := make([]int, len(bitstreams))
	for i, bs := range bitstreams {
		lengths[i] = len(bs)
	}

	lengthVariance := s.calculateVariance(lengths)
	if lengthVariance > 50 {
		quality *= 0.9
	}

	// Check entropy
	entropy := s.calculateEntropy(bitstreams)
	if entropy < 3.0 {
		quality *= 0.8 // Low entropy might indicate poor encoding
	}

	return quality
}

// calculateDataQuality calculates overall data quality
func (s *AnalysisService) calculateDataQuality(items []*models.ItemData) float64 {
	return (s.calculateCompleteness(items) + s.calculateConsistency(items)) / 2
}

// calculateConfidence calculates confidence in the analysis results
func (s *AnalysisService) calculateConfidence(result *AnalysisResult) float64 {
	if result.SampleSize < s.config.MinSampleSize {
		return 0.3 // Low confidence for small samples
	}

	confidence := 0.5 // Base confidence

	// Increase confidence based on sample size
	if result.SampleSize >= 100 {
		confidence += 0.3
	} else if result.SampleSize >= 50 {
		confidence += 0.2
	} else if result.SampleSize >= s.config.MinSampleSize {
		confidence += 0.1
	}

	// Increase confidence based on quality score
	confidence += result.QualityScore * 0.2

	return math.Min(confidence, 1.0)
}

// generateInsights generates insights from the analysis results
func (s *AnalysisService) generateInsights(result *AnalysisResult) []string {
	insights := make([]string, 0)

	// Sample size insights
	if result.SampleSize < s.config.MinSampleSize {
		insights = append(insights, fmt.Sprintf("Small sample size (%d items). Consider collecting more data for reliable analysis.", result.SampleSize))
	}

	// Pattern insights
	if len(result.Patterns) > 0 {
		topPattern := result.Patterns[0]
		insights = append(insights, fmt.Sprintf("Strong pattern detected: %s (%.1f%% of items)", topPattern.PatternName, topPattern.Percentage))
	}

	// Statistical insights
	if result.Statistics != nil {
		if result.Statistics.AverageLevel > 50 {
			insights = append(insights, "High-level items detected (average level > 50)")
		}

		if result.Statistics.UniqueManufacturers < 3 {
			insights = append(insights, "Low manufacturer diversity detected")
		}
	}

	// Bitstream insights
	if result.BitstreamAnalysis != nil {
		if result.BitstreamAnalysis.Entropy < 4.0 {
			insights = append(insights, "Low bitstream entropy detected - possible encoding patterns")
		}

		if len(result.BitstreamAnalysis.CommonPrefixes) > 0 {
			insights = append(insights, fmt.Sprintf("Common bitstream prefixes detected: %d patterns", len(result.BitstreamAnalysis.CommonPrefixes)))
		}
	}

	// Quality insights
	if result.QualityScore > 0.8 {
		insights = append(insights, "High data quality detected")
	} else if result.QualityScore < 0.5 {
		insights = append(insights, "Low data quality detected - review data sources")
	}

	return insights
}

// Cache management methods

func (s *AnalysisService) generateCacheKey(request *AnalysisRequest) string {
	// Generate a cache key based on the serial codes and options
	// For simplicity, use a hash of the concatenated serial codes
	concatenated := strings.Join(request.SerialCodes, ",")
	return fmt.Sprintf("analysis_%x", len(concatenated))
}

func (s *AnalysisService) getCachedResult(cacheKey string) *AnalysisResult {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()

	return s.analysisCache[cacheKey]
}

func (s *AnalysisService) cacheResult(request *AnalysisRequest, result *AnalysisResult) {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	cacheKey := s.generateCacheKey(request)

	// Check cache size limit
	if len(s.analysisCache) >= s.config.MaxCacheSize {
		// Simple eviction: remove the oldest entry
		var oldestKey string
		for key := range s.analysisCache {
			oldestKey = key
			break
		}
		if oldestKey != "" {
			delete(s.analysisCache, oldestKey)
		}
	}

	s.analysisCache[cacheKey] = result
}

// Pattern initialization

func (s *AnalysisService) initializePatterns() {
	// Initialize known patterns
	s.patterns["maliwan_elemental"] = &ItemPattern{
		Name:        "Maliwan Elemental",
		Type:        "manufacturer",
		Description: "Maliwan weapons with elemental effects",
		Signature:   "MAL*ELEM*",
		Weight:      0.9,
		Rarity:      "rare",
		Attributes: map[string]interface{}{
			"manufacturer": "maliwan",
			"elemental":    true,
		},
	}

	s.patterns["jacobs_non_elemental"] = &ItemPattern{
		Name:        "Jakobs Non-Elemental",
		Type:        "manufacturer",
		Description: "Jakobs weapons without elemental effects",
		Signature:   "JAK*NOELEM*",
		Weight:      0.8,
		Rarity:      "uncommon",
		Attributes: map[string]interface{}{
			"manufacturer": "jakobs",
			"elemental":    false,
		},
	}

	s.patterns["high_level_pistols"] = &ItemPattern{
		Name:        "High Level Pistols",
		Type:        "level_range",
		Description: "Pistols with level 40+",
		Signature:   "PISTOL*L40+",
		Weight:      0.7,
		Rarity:      "epic",
		Attributes: map[string]interface{}{
			"type":  "pistol",
			"min_level": 40,
		},
	}
}

func (s *AnalysisService) getKnownPatterns() []*ItemPattern {
	patterns := make([]*ItemPattern, 0, len(s.patterns))
	for _, pattern := range s.patterns {
		patterns = append(patterns, pattern)
	}
	return patterns
}

// Utility functions

func generateRequestID() string {
	// Generate a unique request ID
	return fmt.Sprintf("analysis_%d", time.Now().UnixNano())
}

// NewItemStatistics creates a new ItemStatistics instance
func NewItemStatistics() *ItemStatistics {
	return &ItemStatistics{
		LevelDistribution: make(map[int]int),
	}
}