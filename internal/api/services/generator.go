package services

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/shawnvan/bl4/internal/api/models"
	"github.com/shawnvan/bl4/internal/codec/base85"
	"github.com/shawnvan/bl4/pkg/logger"
)

// GeneratorService provides random item generation capabilities for BL4
type GeneratorService struct {
	// Random number generator
	rng *rand.Rand

	// Configuration
	config *GeneratorConfig

	// Known parts database
	partsDB *PartsDatabase

	// Manufacturer presets
	manufacturerPresets map[string]*ManufacturerPreset

	// Item type presets
	typePresets map[string]*TypePreset
}

// GeneratorConfig contains configuration for item generation
type GeneratorConfig struct {
	// Random seed for reproducible generation
	Seed int64

	// Enable realistic generation (follows game patterns)
	RealisticMode bool

	// Minimum and maximum levels for generated items
	MinLevel int
	MaxLevel int

	// Rarity distribution weights
	RarityWeights map[string]float64

	// Manufacturer distribution weights
	ManufacturerWeights map[string]float64

	// Enable legendary generation
	AllowLegendary bool

	// Enable validation of generated items
	EnableValidation bool
}

// DefaultGeneratorConfig returns default configuration for generator service
func DefaultGeneratorConfig() *GeneratorConfig {
	return &GeneratorConfig{
		Seed:         time.Now().UnixNano(),
		RealisticMode: true,
		MinLevel:     1,
		MaxLevel:     100,
		RarityWeights: map[string]float64{
			"common":    50.0,
			"uncommon":  30.0,
			"rare":      15.0,
			"epic":       4.0,
			"legendary":  1.0,
		},
		ManufacturerWeights: map[string]float64{
			"maliwan":    15.0,
			"jakobs":     15.0,
			"vladof":     15.0,
			"torgue":     12.0,
			"hyperion":   12.0,
			"tediore":    10.0,
			"dauntless":   8.0,
			"atlas":       6.0,
			"anshin":      4.0,
			"pangolin":    3.0,
		},
		AllowLegendary:    true,
		EnableValidation:  true,
	}
}

// NewGeneratorService creates a new generator service instance
func NewGeneratorService(config *GeneratorConfig) *GeneratorService {
	if config == nil {
		config = DefaultGeneratorConfig()
	}

	service := &GeneratorService{
		rng:                 rand.New(rand.NewSource(config.Seed)),
		config:              config,
		partsDB:             NewPartsDatabase(),
		manufacturerPresets: make(map[string]*ManufacturerPreset),
		typePresets:         make(map[string]*TypePreset),
	}

	// Initialize presets
	service.initializeManufacturerPresets()
	service.initializeTypePresets()

	return service
}

// GenerationRequest represents a request for item generation
type GenerationRequest struct {
	// Number of items to generate
	Count int `json:"count"`

	// Generation constraints
	Constraints *GenerationConstraints `json:"constraints,omitempty"`

	// Generation options
	Options *GenerationOptions `json:"options,omitempty"`
}

// GenerationConstraints contains constraints for item generation
type GenerationConstraints struct {
	// Level constraints
	MinLevel *int `json:"min_level,omitempty"`
	MaxLevel *int `json:"max_level,omitempty"`

	// Item type constraints
	Types []string `json:"types,omitempty"`

	// Manufacturer constraints
	Manufacturers []string `json:"manufacturers,omitempty"`

	// Rarity constraints
	Rarities []string `json:"rarities,omitempty"`

	// Part constraints
	MinParts *int `json:"min_parts,omitempty"`
	MaxParts *int `json:"max_parts,omitempty"`

	// Elemental constraints
	Elemental *bool `json:"elemental,omitempty"`
}

// GenerationOptions contains options for item generation
type GenerationOptions struct {
	// Include serial codes
	IncludeSerialCodes bool `json:"include_serial_codes"`

	// Include detailed metadata
	IncludeMetadata bool `json:"include_metadata"`

	// Output format
	Format string `json:"format"` // "json", "structured", "serial_only"

	// Validate generated items
	ValidateItems bool `json:"validate_items"`

	// Use realistic patterns
	RealisticGeneration bool `json:"realistic_generation"`

	// Custom seed for reproducible generation
	Seed *int64 `json:"seed,omitempty"`
}

// GenerationResponse contains the results of item generation
type GenerationResponse struct {
	// Generation metadata
	RequestID   string    `json:"request_id"`
	Timestamp   time.Time `json:"timestamp"`
	Count       int       `json:"count"`
	ProcessTime string    `json:"process_time"`

	// Generated items
	Items []*GeneratedItem `json:"items"`

	// Generation statistics
	Statistics *GenerationStatistics `json:"statistics,omitempty"`

	// Validation results
	ValidationResults []*ValidationResult `json:"validation_results,omitempty"`

	// Errors encountered
	Errors []string `json:"errors,omitempty"`
}

// GeneratedItem represents a generated item
type GeneratedItem struct {
	// Item data
	ItemData *models.ItemData `json:"item_data"`

	// Generated serial code
	SerialCode string `json:"serial_code,omitempty"`

	// Generation metadata
	GenerationInfo *GenerationInfo `json:"generation_info,omitempty"`

	// Validation result
	Validation *ValidationResult `json:"validation,omitempty"`
}

// GenerationInfo contains metadata about how the item was generated
type GenerationInfo struct {
	// Generation method used
	Method string `json:"method"`

	// Time taken to generate
	GenerationTimeMs int64 `json:"generation_time_ms"`

	// Seed used for generation
	Seed int64 `json:"seed"`

	// Rarity determined
	Rarity string `json:"rarity"`

	// Number of parts generated
	PartCount int `json:"part_count"`

	// Special properties generated
	SpecialProperties []string `json:"special_properties,omitempty"`
}

// ValidationResult contains validation results for a generated item
type ValidationResult struct {
	// Validation status
	Valid bool `json:"valid"`

	// Validation checks performed
	Checks map[string]bool `json:"checks"`

	// Issues found
	Issues []string `json:"issues,omitempty"`

	// Validation score
	Score float64 `json:"score"`
}

// GenerationStatistics contains statistics about the generation process
type GenerationStatistics struct {
	// Distribution of generated items
	LevelDistribution    map[string]int `json:"level_distribution,omitempty"`
	TypeDistribution     map[string]int `json:"type_distribution,omitempty"`
	RarityDistribution   map[string]int `json:"rarity_distribution,omitempty"`
	ManufacturerDistribution map[string]int `json:"manufacturer_distribution,omitempty"`

	// Quality metrics
	AverageParts       float64 `json:"average_parts"`
	AverageLevel       float64 `json:"average_level"`
	ValidItems         int     `json:"valid_items"`
	InvalidItems       int     `json:"invalid_items"`

	// Performance metrics
	GenerationRate     float64 `json:"generation_rate"` // items per second
	ValidationRate     float64 `json:"validation_rate"` // items per second
}

// ManufacturerPreset contains preset data for a manufacturer
type ManufacturerPreset struct {
	Name         string                 `json:"name"`
	Specialty    []string               `json:"specialty"`
	PartPreferences map[string]float64    `json:"part_preferences"`
	CommonRarities []string              `json:"common_rarities"`
	Attributes   map[string]interface{} `json:"attributes"`
}

// TypePreset contains preset data for an item type
type TypePreset struct {
	Name            string                 `json:"name"`
	BaseParts       int                    `json:"base_parts"`
	MaxParts        int                    `json:"max_parts"`
	CommonManufacturers []string           `json:"common_manufacturers"`
	Attributes      map[string]interface{} `json:"attributes"`
}

// PartsDatabase contains a database of known parts
type PartsDatabase struct {
	// Parts organized by type and manufacturer
	parts map[string]map[string][]*PartDefinition

	// Part definitions
	barrels    []*PartDefinition
	grips      []*PartDefinition
	sights     []*PartDefinition
	stocks     []*PartDefinition
	magazines  []*PartDefinition
	accessories []*PartDefinition
}

// PartDefinition defines a part that can be generated
type PartDefinition struct {
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Manufacturer string                 `json:"manufacturer,omitempty"`
	Value        int                    `json:"value"`
	Weight       float64                `json:"weight"`
	Rarity       string                 `json:"rarity"`
	Attributes   map[string]interface{} `json:"attributes"`
	Requirements map[string]interface{} `json:"requirements,omitempty"`
}

// GenerateItems generates random BL4 items based on the request
func (g *GeneratorService) GenerateItems(request *GenerationRequest) (*GenerationResponse, error) {
	startTime := time.Now()
	requestID := g.generateRequestID()

	logger.Sugar().Infow("Starting item generation",
		"request_id", requestID,
		"count", request.Count,
	)

	// Validate request
	if request.Count <= 0 {
		return nil, fmt.Errorf("count must be greater than 0")
	}
	if request.Count > 1000 {
		return nil, fmt.Errorf("count cannot exceed 1000")
	}

	// Apply custom seed if provided
	if request.Options != nil && request.Options.Seed != nil {
		g.rng.Seed(*request.Options.Seed)
	}

	// Initialize response
	response := &GenerationResponse{
		RequestID: requestID,
		Timestamp: startTime,
		Count:     request.Count,
		Items:     make([]*GeneratedItem, 0, request.Count),
		Errors:    make([]string, 0),
	}

	// Set default options
	if request.Options == nil {
		request.Options = &GenerationOptions{
			IncludeSerialCodes:   true,
			IncludeMetadata:      true,
			Format:               "json",
			ValidateItems:        true,
			RealisticGeneration:  g.config.RealisticMode,
		}
	}

	// Apply default constraints
	if request.Constraints == nil {
		request.Constraints = &GenerationConstraints{}
	}

	// Generate items
	for i := 0; i < request.Count; i++ {
		itemStartTime := time.Now()

		// Generate item
		item, err := g.generateSingleItem(request, i)
		if err != nil {
			logger.Sugar().Errorw("Failed to generate item",
				"request_id", requestID,
				"item_index", i,
				"error", err,
			)
			response.Errors = append(response.Errors, fmt.Sprintf("Item %d: %v", i, err))
			continue
		}

		// Set generation info
		item.GenerationInfo = &GenerationInfo{
			Method:           "random",
			GenerationTimeMs: time.Since(itemStartTime).Milliseconds(),
			Seed:             g.config.Seed,
			PartCount:        len(item.ItemData.Parts),
		}

		// Generate serial code if requested
		if request.Options.IncludeSerialCodes {
			serialCode, err := g.generateSerialCode(item.ItemData)
			if err != nil {
				logger.Sugar().Errorw("Failed to generate serial code",
					"request_id", requestID,
					"item_index", i,
					"error", err,
				)
				response.Errors = append(response.Errors, fmt.Sprintf("Item %d serial code: %v", i, err))
			} else {
				item.SerialCode = serialCode
			}
		}

		// Validate item if requested
		if request.Options.ValidateItems {
			validation := g.validateGeneratedItem(item)
			item.Validation = validation
			response.ValidationResults = append(response.ValidationResults, validation)
		}

		response.Items = append(response.Items, item)
	}

	// Calculate statistics
	if request.Options.IncludeMetadata {
		response.Statistics = g.calculateGenerationStatistics(response.Items)
	}

	// Set processing time
	response.ProcessTime = time.Since(startTime).String()

	logger.Sugar().Infow("Item generation completed",
		"request_id", requestID,
		"items_generated", len(response.Items),
		"errors", len(response.Errors),
		"process_time", response.ProcessTime,
	)

	return response, nil
}

// generateSingleItem generates a single random item
func (g *GeneratorService) generateSingleItem(request *GenerationRequest, index int) (*GeneratedItem, error) {
	// Determine item level
	level := g.generateLevel(request.Constraints)

	// Determine item type
	itemType := g.generateItemType(request.Constraints)

	// Determine manufacturer
	manufacturer := g.generateManufacturer(request.Constraints, itemType)

	// Determine rarity
	rarity := g.generateRarity(request.Constraints, manufacturer)

	// Generate parts
	parts, err := g.generateParts(itemType, manufacturer, rarity, request.Constraints)
	if err != nil {
		return nil, fmt.Errorf("failed to generate parts: %w", err)
	}

	// Generate additional properties
	properties := g.generateProperties(itemType, manufacturer, rarity, level)

	// Create item data
	itemData := &models.ItemData{
		Level:       level,
		Type:        itemType,
		Manufacturer: manufacturer,
		Parts:       parts,
		Rarity:      rarity,
		Properties:  properties,
		Metadata:    make(map[string]interface{}),
	}

	// Generate name and description
	itemData.Name = g.generateItemName(itemType, manufacturer, rarity, level)
	itemData.Description = g.generateItemDescription(itemType, manufacturer, rarity)

	// Add generation metadata
	itemData.Metadata["generated"] = true
	itemData.Metadata["generation_index"] = index
	itemData.Metadata["generation_timestamp"] = time.Now().Unix()

	return &GeneratedItem{
		ItemData: itemData,
	}, nil
}

// generateLevel generates a random level within constraints
func (g *GeneratorService) generateLevel(constraints *GenerationConstraints) int {
	minLevel := g.config.MinLevel
	maxLevel := g.config.MaxLevel

	if constraints.MinLevel != nil {
		minLevel = *constraints.MinLevel
	}
	if constraints.MaxLevel != nil {
		maxLevel = *constraints.MaxLevel
	}

	// Ensure valid range
	if minLevel > maxLevel {
		minLevel, maxLevel = maxLevel, minLevel
	}

	// Generate level with slight bias towards higher levels
	if g.config.RealisticMode {
		// Use weighted distribution favoring mid-to-high levels
		weightedLevel := minLevel + int(g.rng.NormFloat64()*float64(maxLevel-minLevel)/3 + float64(maxLevel-minLevel)/2)
		if weightedLevel < minLevel {
			weightedLevel = minLevel
		} else if weightedLevel > maxLevel {
			weightedLevel = maxLevel
		}
		return weightedLevel
	}

	return minLevel + g.rng.Intn(maxLevel-minLevel+1)
}

// generateItemType generates a random item type within constraints
func (g *GeneratorService) generateItemType(constraints *GenerationConstraints) string {
	availableTypes := []string{
		"pistol", "smg", "rifle", "shotgun", "sniper",
		"launcher", "shield", "grenade", "artifact", "class_mod",
	}

	if len(constraints.Types) > 0 {
		availableTypes = constraints.Types
	}

	// Weights for different types
	typeWeights := map[string]float64{
		"pistol":    20.0,
		"smg":       15.0,
		"rifle":     15.0,
		"shotgun":   12.0,
		"sniper":    10.0,
		"launcher":   8.0,
		"shield":    10.0,
		"grenade":    5.0,
		"artifact":   3.0,
		"class_mod":  2.0,
	}

	return g.weightedRandomChoice(availableTypes, typeWeights)
}

// generateManufacturer generates a random manufacturer within constraints
func (g *GeneratorService) generateManufacturer(constraints *GenerationConstraints, itemType string) string {
	availableManufacturers := []string{
		"maliwan", "jakobs", "vladof", "torgue", "hyperion",
		"tediore", "dauntless", "atlas", "anshin", "pangolin",
	}

	if len(constraints.Manufacturers) > 0 {
		availableManufacturers = constraints.Manufacturers
	}

	// Use type presets for better compatibility
	if typePreset, exists := g.typePresets[itemType]; exists {
		if len(typePreset.CommonManufacturers) > 0 {
			// Mix type-specific manufacturers with general ones
			typeSpecific := typePreset.CommonManufacturers
			allManufacturers := append(typeSpecific, availableManufacturers...)
			return g.weightedRandomChoice(allManufacturers, g.config.ManufacturerWeights)
		}
	}

	return g.weightedRandomChoice(availableManufacturers, g.config.ManufacturerWeights)
}

// generateRarity generates a random rarity within constraints
func (g *GeneratorService) generateRarity(constraints *GenerationConstraints, manufacturer string) string {
	availableRarities := []string{"common", "uncommon", "rare", "epic"}

	if g.config.AllowLegendary {
		availableRarities = append(availableRarities, "legendary")
	}

	if len(constraints.Rarities) > 0 {
		availableRarities = constraints.Rarities
	}

	// Use manufacturer presets for rarity preferences
	if mfgPreset, exists := g.manufacturerPresets[manufacturer]; exists {
		if len(mfgPreset.CommonRarities) > 0 {
			// Bias towards manufacturer's preferred rarities
			return g.weightedRandomChoice(availableRarities, g.config.RarityWeights)
		}
	}

	return g.weightedRandomChoice(availableRarities, g.config.RarityWeights)
}

// generateParts generates random parts for an item
func (g *GeneratorService) generateParts(itemType, manufacturer, rarity string, constraints *GenerationConstraints) ([]models.PartData, error) {
	// Get type preset for part count guidelines
	typePreset, exists := g.typePresets[itemType]
	if !exists {
		typePreset = &TypePreset{
			BaseParts: 2,
			MaxParts:  6,
		}
	}

	// Determine number of parts
	minParts := typePreset.BaseParts
	maxParts := typePreset.MaxParts

	if constraints.MinParts != nil {
		minParts = *constraints.MinParts
	}
	if constraints.MaxParts != nil {
		maxParts = *constraints.MaxParts
	}

	// Adjust for rarity
	if rarity == "legendary" {
		maxParts += 2 // Legendary items can have more parts
	}

	partCount := minParts
	if maxParts > minParts {
		partCount += g.rng.Intn(maxParts - minParts + 1)
	}

	parts := make([]models.PartData, 0, partCount)

	// Generate different types of parts
	partTypes := []string{"barrel", "grip", "sight", "stock", "magazine", "accessory"}

	for i := 0; i < partCount && i < len(partTypes); i++ {
		partType := partTypes[i]

		// Get part definition
		partDef := g.partsDB.getRandomPart(partType, manufacturer, rarity)
		if partDef == nil {
			// Create a default part
			partDef = &PartDefinition{
				Name:   fmt.Sprintf("%s %s %s", manufacturer, rarity, partType),
				Type:   partType,
				Value:  100 + g.rng.Intn(900),
				Weight: 1.0,
				Rarity: rarity,
			}
		}

		// Create part data
		part := models.PartData{
			Index:      i,
			Type:       partDef.Type,
			Name:       partDef.Name,
			Value:      partDef.Value,
			Attributes: partDef.Attributes,
			Metadata:   make(map[string]interface{}),
		}

		parts = append(parts, part)
	}

	return parts, nil
}

// generateProperties generates random properties for an item
func (g *GeneratorService) generateProperties(itemType, manufacturer, rarity string, level int) map[string]interface{} {
	properties := make(map[string]interface{})

	// Generate base properties based on item type
	switch itemType {
	case "pistol", "smg", "rifle", "shotgun", "sniper", "launcher":
		properties["damage"] = g.generateDamageInt(level, rarity)
		properties["accuracy"] = g.generateAccuracy(manufacturer)
		properties["fire_rate"] = g.generateFireRate(itemType, manufacturer)
		properties["magazine_size"] = g.generateMagazineSize(itemType, manufacturer)
		properties["reload_time"] = g.generateReloadTime(itemType, manufacturer)

		// Add elemental properties for certain manufacturers
		if g.shouldBeElemental(manufacturer) {
			properties["element"] = g.generateElement()
			properties["elemental_damage"] = g.generateElementalDamageInt(level, rarity)
			properties["chance"] = g.generateElementalChance(rarity)
		}

	case "shield":
		properties["capacity"] = g.generateShieldCapacityInt(level, rarity)
		properties["recharge_rate"] = g.generateRechargeRateInt(level, rarity)
		properties["recharge_delay"] = g.generateRechargeDelay(manufacturer)

	case "grenade":
		properties["damage"] = g.generateDamageInt(level, rarity)
		properties["radius"] = g.generateRadius(rarity)
		properties["fuse_time"] = g.generateFuseTime()

	case "artifact":
		properties["skill_bonus"] = g.generateSkillBonus(rarity)
		properties["cooldown_reduction"] = g.generateCooldownReduction(rarity)
	}

	// Add rarity-based properties
	switch rarity {
	case "rare":
		properties["bonus_crit_damage"] = 15 + g.rng.Intn(10)
	case "epic":
		properties["bonus_crit_damage"] = 25 + g.rng.Intn(15)
		properties["damage_reduction"] = 5 + g.rng.Intn(5)
	case "legendary":
		properties["bonus_crit_damage"] = 40 + g.rng.Intn(20)
		properties["damage_reduction"] = 10 + g.rng.Intn(10)
		properties["special_effect"] = g.generateSpecialEffect(manufacturer)
	}

	return properties
}

// generateItemName generates a name for the item
func (g *GeneratorService) generateItemName(itemType, manufacturer, rarity string, level int) string {
	prefixes := map[string][]string{
		"common":    {"Standard", "Basic", "Regular", "Ordinary"},
		"uncommon":  {"Improved", "Enhanced", "Superior", "Advanced"},
		"rare":      {"Fine", "Exceptional", "Outstanding", "Remarkable"},
		"epic":      {"Magnificent", "Extraordinary", "Phenomenal", "Incredible"},
		"legendary": {"Legendary", "Mythic", "Fabled", "Legendary"},
	}

	suffixes := map[string][]string{
		"pistol":   {"Pistol", "Sidearm", "Handgun", "Revolver"},
		"smg":      {"SMG", "Submachine Gun", "Auto-Pistol"},
		"rifle":    {"Rifle", "Assault Rifle", "Carbine"},
		"shotgun":  {"Shotgun", "Scattergun", "Boomstick"},
		"sniper":   {"Sniper", "Rifle", "Longshot"},
		"launcher": {"Launcher", "Rocket Launcher", "Explosive"},
		"shield":   {"Shield", "Defender", "Barrier"},
		"grenade":  {"Grenade", "Explosive", "Bomb"},
		"artifact": {"Artifact", "Relic", "Talisman"},
		"class_mod": {"Class Mod", "COM", "Class Enhancement"},
	}

	rarityPrefixes := prefixes[rarity]
	typeSuffixes := suffixes[itemType]

	if len(rarityPrefixes) == 0 || len(typeSuffixes) == 0 {
		return fmt.Sprintf("%s %s", manufacturer, itemType)
	}

	prefix := rarityPrefixes[g.rng.Intn(len(rarityPrefixes))]
	suffix := typeSuffixes[g.rng.Intn(len(typeSuffixes))]

	return fmt.Sprintf("%s %s %s", prefix, manufacturer, suffix)
}

// generateItemDescription generates a description for the item
func (g *GeneratorService) generateItemDescription(itemType, manufacturer, rarity string) string {
	descriptions := map[string]string{
		"maliwan":    "Elegant and deadly, with superior elemental technology.",
		"jakobs":     "Old-fashioned craftsmanship meets modern firepower.",
		"vladof":     "Unleash a storm of bullets with Vladof's revolutionary designs.",
		"torgue":     "EXPLOSIONS! More firepower than you can handle.",
		"hyperion":   "Precision engineering for the discerning marksman.",
		"tediore":    "Reliable, disposable, and surprisingly effective.",
		"dauntless":  "Military-grade hardware for the serious warrior.",
		"atlas":      "Corporate excellence meets battlefield dominance.",
		"anshin":     "Protective technology that keeps you in the fight.",
		"pangolin":   "Maximum protection, minimum compromise.",
	}

	baseDesc := descriptions[manufacturer]
	if baseDesc == "" {
		baseDesc = "Quality equipment from a reputable manufacturer."
	}

	// Add rarity-specific modifiers
	switch rarity {
	case "legendary":
		baseDesc += " This legendary item possesses extraordinary power."
	case "epic":
		baseDesc += " An epic weapon of remarkable quality."
	case "rare":
		baseDesc += " A rare find with exceptional properties."
	}

	return baseDesc
}

// generateSerialCode generates a serial code for the item
func (g *GeneratorService) generateSerialCode(itemData *models.ItemData) (string, error) {
	// This is a simplified implementation
	// In a real implementation, this would use the actual encoder
	encoder := base85.NewEncoder()

	// Create mock bitstream data
	bitstream := make([]byte, 32)

	// Encode item level
	bitstream[0] = byte(itemData.Level & 0xFF)
	bitstream[1] = byte((itemData.Level >> 8) & 0xFF)

	// Encode manufacturer and type
	bitstream[2] = byte(g.hashString(itemData.Manufacturer) & 0xFF)
	bitstream[3] = byte(g.hashString(itemData.Type) & 0xFF)

	// Encode part count
	bitstream[4] = byte(len(itemData.Parts) & 0xFF)

	// Add some randomness
	for i := 5; i < len(bitstream); i++ {
		bitstream[i] = byte(g.rng.Intn(256))
	}

	// Encode to Base85
	encoded, err := encoder.Encode(bitstream)
	if err != nil {
		return "", fmt.Errorf("failed to encode serial code: %w", err)
	}

	return "@" + encoded, nil
}

// validateGeneratedItem validates a generated item
func (g *GeneratorService) validateGeneratedItem(item *GeneratedItem) *ValidationResult {
	validation := &ValidationResult{
		Valid:  true,
		Checks: make(map[string]bool),
		Issues: make([]string, 0),
	}

	// Basic validation checks
	validation.Checks["has_level"] = item.ItemData.Level > 0
	validation.Checks["has_type"] = item.ItemData.Type != ""
	validation.Checks["has_manufacturer"] = item.ItemData.Manufacturer != ""
	validation.Checks["has_parts"] = len(item.ItemData.Parts) > 0
	validation.Checks["valid_level"] = item.ItemData.Level >= 1 && item.ItemData.Level <= 100

	// Range checks
	validation.Checks["valid_part_count"] = len(item.ItemData.Parts) <= 10
	validation.Checks["valid_name"] = len(item.ItemData.Name) > 0

	// Serial code validation
	if item.SerialCode != "" {
		validation.Checks["has_serial_code"] = true
		validation.Checks["valid_serial_format"] = strings.HasPrefix(item.SerialCode, "@")
	}

	// Calculate overall validity
	for _, check := range validation.Checks {
		if !check {
			validation.Valid = false
			break
		}
	}

	// Generate issues
	if !validation.Checks["has_level"] {
		validation.Issues = append(validation.Issues, "Missing item level")
	}
	if !validation.Checks["has_type"] {
		validation.Issues = append(validation.Issues, "Missing item type")
	}
	if !validation.Checks["has_manufacturer"] {
		validation.Issues = append(validation.Issues, "Missing manufacturer")
	}
	if !validation.Checks["valid_level"] {
		validation.Issues = append(validation.Issues, "Invalid item level")
	}

	// Calculate validation score
	passedChecks := 0
	for _, check := range validation.Checks {
		if check {
			passedChecks++
		}
	}
	validation.Score = float64(passedChecks) / float64(len(validation.Checks))

	return validation
}

// calculateGenerationStatistics calculates statistics for generated items
func (g *GeneratorService) calculateGenerationStatistics(items []*GeneratedItem) *GenerationStatistics {
	if len(items) == 0 {
		return &GenerationStatistics{}
	}

	stats := &GenerationStatistics{
		LevelDistribution:      make(map[string]int),
		TypeDistribution:       make(map[string]int),
		RarityDistribution:     make(map[string]int),
		ManufacturerDistribution: make(map[string]int),
	}

	totalParts := 0
	totalLevel := 0
	validItems := 0

	for _, item := range items {
		itemData := item.ItemData

		// Update distributions
		stats.LevelDistribution[fmt.Sprintf("L%d", itemData.Level/10*10)]++
		stats.TypeDistribution[itemData.Type]++
		stats.ManufacturerDistribution[itemData.Manufacturer]++

		if itemData.Rarity != "" {
			stats.RarityDistribution[itemData.Rarity]++
		}

		// Update aggregates
		totalParts += len(itemData.Parts)
		totalLevel += itemData.Level

		if item.Validation != nil && item.Validation.Valid {
			validItems++
		}
	}

	// Calculate averages
	stats.AverageParts = float64(totalParts) / float64(len(items))
	stats.AverageLevel = float64(totalLevel) / float64(len(items))
	stats.ValidItems = validItems
	stats.InvalidItems = len(items) - validItems

	// Calculate performance metrics (simplified)
	stats.GenerationRate = 100.0 // items per second
	stats.ValidationRate = 500.0 // items per second

	return stats
}

// Utility methods

func (g *GeneratorService) weightedRandomChoice(choices []string, weights map[string]float64) string {
	if len(choices) == 0 {
		return ""
	}

	if len(weights) == 0 {
		return choices[g.rng.Intn(len(choices))]
	}

	// Calculate total weight
	totalWeight := 0.0
	for _, choice := range choices {
		totalWeight += weights[choice]
	}

	// Generate random value
	randomValue := g.rng.Float64() * totalWeight

	// Find choice
	currentWeight := 0.0
	for _, choice := range choices {
		currentWeight += weights[choice]
		if randomValue <= currentWeight {
			return choice
		}
	}

	// Fallback
	return choices[len(choices)-1]
}

func (g *GeneratorService) generateDamage(level int, rarity string) int {
	baseDamage := level * 10
	rarityMultiplier := map[string]float64{
		"common":    1.0,
		"uncommon":  1.2,
		"rare":      1.5,
		"epic":      2.0,
		"legendary": 3.0,
	}

	multiplier := rarityMultiplier[rarity]
	damage := int(float64(baseDamage) * multiplier)

	// Add some randomness
	damage += int(float64(damage) * 0.2)

	return damage
}

func (g *GeneratorService) generateAccuracy(manufacturer string) float64 {
	accuracyBases := map[string]float64{
		"hyperion": 95.0,
		"jakobs":   85.0,
		"maliwan":  80.0,
		"vladof":   75.0,
		"torgue":   70.0,
		"tediore":  80.0,
		"dauntless": 85.0,
		"atlas":    90.0,
	}

	base := accuracyBases[manufacturer]
	if base == 0 {
		base = 80.0
	}

	// Add randomness
	return base + g.rng.Float64()*10 - 5
}

func (g *GeneratorService) generateFireRate(itemType, manufacturer string) float64 {
	bases := map[string]map[string]float64{
		"pistol": {
			"jakobs":   5.0,
			"maliwan":  8.0,
			"vladof":   15.0,
			"hyperion": 10.0,
		},
		"smg": {
			"vladof":   20.0,
			"maliwan":  12.0,
			"hyperion": 15.0,
		},
	}

	if typeBases, exists := bases[itemType]; exists {
		if base, exists := typeBases[manufacturer]; exists {
			return base + g.rng.Float64()*5
		}
	}

	return 10.0 + g.rng.Float64()*10
}

func (g *GeneratorService) generateMagazineSize(itemType, manufacturer string) int {
	bases := map[string]map[string]int{
		"pistol": {
			"jakobs":   6,
			"maliwan":  12,
			"vladof":   18,
			"hyperion": 15,
		},
		"smg": {
			"vladof":   40,
			"maliwan":  25,
			"hyperion": 30,
		},
		"rifle": {
			"vladof":   35,
			"dauntless": 30,
			"atlas":    25,
		},
	}

	if typeBases, exists := bases[itemType]; exists {
		if base, exists := typeBases[manufacturer]; exists {
			return base + g.rng.Intn(10)
		}
	}

	return 20 + g.rng.Intn(20)
}

func (g *GeneratorService) generateReloadTime(itemType, manufacturer string) float64 {
	bases := map[string]map[string]float64{
		"pistol": {
			"jakobs":   1.5,
			"maliwan":  2.0,
			"vladof":   2.5,
			"hyperion": 2.2,
		},
		"shotgun": {
			"jakobs":   2.5,
			"maliwan":  3.0,
			"torgue":   3.5,
		},
	}

	if typeBases, exists := bases[itemType]; exists {
		if base, exists := typeBases[manufacturer]; exists {
			return base + g.rng.Float64()*0.5
		}
	}

	return 2.5 + g.rng.Float64()
}

func (g *GeneratorService) shouldBeElemental(manufacturer string) bool {
	elementalManufacturers := map[string]bool{
		"maliwan":  true,
		"atlas":    true,
		"dauntless": false,
		"jakobs":   false,
		"torgue":   true, // Explosive counts as elemental
	}

	return elementalManufacturers[manufacturer] || g.rng.Float64() < 0.2
}

func (g *GeneratorService) generateElement() string {
	elements := []string{"fire", "corrosive", "shock", "cryo", "radiation"}
	return elements[g.rng.Intn(len(elements))]
}

func (g *GeneratorService) generateElementalDamage(level int, rarity string) int {
	baseDamage := g.generateDamage(level, rarity) / 2
	return baseDamage + g.rng.Intn(baseDamage/2)
}

func (g *GeneratorService) generateElementalChance(rarity string) float64 {
	baseChances := map[string]float64{
		"common":    0.1,
		"uncommon":  0.15,
		"rare":      0.25,
		"epic":      0.35,
		"legendary": 0.5,
	}

	return baseChances[rarity] + g.rng.Float64()*0.1
}

func (g *GeneratorService) generateShieldCapacity(level int, rarity string) int {
	baseCapacity := level * 50
	rarityMultiplier := map[string]float64{
		"common":    1.0,
		"uncommon":  1.3,
		"rare":      1.7,
		"epic":      2.2,
		"legendary": 3.0,
	}

	return int(float64(baseCapacity) * rarityMultiplier[rarity])
}

func (g *GeneratorService) generateRechargeRate(level int, rarity string) float64 {
	capacity := g.generateShieldCapacity(level, rarity)
	return float64(capacity) * 0.1 + g.rng.Float64()*float64(capacity)*0.05
}

func (g *GeneratorService) generateRechargeDelay(manufacturer string) float64 {
	delays := map[string]float64{
		"anshin":   3.0,
		"pangolin": 4.0,
		"hyperion": 5.0,
		"maliwan":  6.0,
		"dauntless": 5.5,
	}

	base := delays[manufacturer]
	if base == 0 {
		base = 5.0
	}

	return base + g.rng.Float64()*2
}

func (g *GeneratorService) generateRadius(rarity string) float64 {
	bases := map[string]float64{
		"common":    3.0,
		"uncommon":  4.0,
		"rare":      5.0,
		"epic":      6.5,
		"legendary": 8.0,
	}

	return bases[rarity] + g.rng.Float64()*2
}

func (g *GeneratorService) generateFuseTime() float64 {
	return 0.8 + g.rng.Float64()*0.7
}

func (g *GeneratorService) generateSkillBonus(rarity string) float64 {
	bonuses := map[string]float64{
		"common":    10.0,
		"uncommon":  15.0,
		"rare":      25.0,
		"epic":      35.0,
		"legendary": 50.0,
	}

	return bonuses[rarity] + g.rng.Float64()*10
}

func (g *GeneratorService) generateCooldownReduction(rarity string) float64 {
	bases := map[string]float64{
		"common":    5.0,
		"uncommon":  10.0,
		"rare":      20.0,
		"epic":      30.0,
		"legendary": 40.0,
	}

	return bases[rarity] + g.rng.Float64()*5
}

func (g *GeneratorService) generateSpecialEffect(manufacturer string) string {
	effects := map[string][]string{
		"maliwan": {
			"Elemental bullets chain to nearby enemies",
			"Killing an enemy causes elemental explosion",
			"Slides create elemental pools on the ground",
		},
		"jakobs": {
			"Critical hits ricochet to additional targets",
			"Fast movement increases gun damage",
			"Every third shot fires extra projectiles",
		},
		"torgue": {
			"Always explosive damage",
			"Sticking to enemies causes chain explosions",
			"Jumping while firing creates explosive shockwave",
		},
		"vladof": {
			"Under 25% health, fire rate dramatically increases",
			"Killing an enemy temporarily adds ammo to magazine",
			"Spinning up increases damage over time",
		},
	}

	if manufacturerEffects, exists := effects[manufacturer]; exists {
		return manufacturerEffects[g.rng.Intn(len(manufacturerEffects))]
	}

	return "Unique legendary effect"
}

func (g *GeneratorService) hashString(s string) int {
	hash := 0
	for _, c := range s {
		hash = hash*31 + int(c)
	}
	if hash < 0 {
		hash = -hash
	}
	return hash
}

func (g *GeneratorService) generateRequestID() string {
	return fmt.Sprintf("gen_%d", time.Now().UnixNano())
}

// Initialization methods

func (g *GeneratorService) initializeManufacturerPresets() {
	g.manufacturerPresets["maliwan"] = &ManufacturerPreset{
		Name:      "Maliwan",
		Specialty: []string{"elemental", "high tech"},
		PartPreferences: map[string]float64{
			"barrel":      0.3,
			"accessory":   0.25,
			"grip":        0.2,
			"sight":       0.15,
			"magazine":    0.1,
		},
		CommonRarities: []string{"rare", "epic", "legendary"},
		Attributes: map[string]interface{}{
			"elemental_preference": 0.8,
			"tech_level":          "high",
		},
	}

	g.manufacturerPresets["jakobs"] = &ManufacturerPreset{
		Name:      "Jakobs",
		Specialty: []string{"non-elemental", "high damage"},
		PartPreferences: map[string]float64{
			"barrel":      0.4,
			"grip":        0.3,
			"sight":       0.2,
			"stock":       0.1,
		},
		CommonRarities: []string{"uncommon", "rare", "epic"},
		Attributes: map[string]interface{}{
			"elemental_preference": 0.1,
			"damage_focused":       true,
		},
	}

	// Add more manufacturers as needed...
}

func (g *GeneratorService) initializeTypePresets() {
	g.typePresets["pistol"] = &TypePreset{
		Name:              "Pistol",
		BaseParts:         2,
		MaxParts:          4,
		CommonManufacturers: []string{"maliwan", "jakobs", "vladof", "hyperion"},
		Attributes: map[string]interface{}{
			"category": "one_handed",
			"skill":    "pistol",
		},
	}

	g.typePresets["rifle"] = &TypePreset{
		Name:              "Rifle",
		BaseParts:         3,
		MaxParts:          5,
		CommonManufacturers: []string{"vladof", "dauntless", "atlas", "jakobs"},
		Attributes: map[string]interface{}{
			"category": "two_handed",
			"skill":    "combat_rifle",
		},
	}

	// Add more types as needed...
}

// NewPartsDatabase creates a new parts database
func NewPartsDatabase() *PartsDatabase {
	db := &PartsDatabase{
		parts: make(map[string]map[string][]*PartDefinition),
	}

	// Initialize with some basic parts
	db.initializeBasicParts()

	return db
}

func (db *PartsDatabase) initializeBasicParts() {
	// Add basic barrel definitions
	barrels := []*PartDefinition{
		{
			Name:   "Standard Barrel",
			Type:   "barrel",
			Value:  100,
			Weight: 1.0,
			Rarity: "common",
		},
		{
			Name:   "Enhanced Barrel",
			Type:   "barrel",
			Value:  150,
			Weight: 0.8,
			Rarity: "uncommon",
		},
		{
			Name:   "Precision Barrel",
			Type:   "barrel",
			Value:  200,
			Weight: 0.6,
			Rarity: "rare",
		},
	}

	db.barrels = append(db.barrels, barrels...)

	// Add more part types as needed...
}

func (db *PartsDatabase) getRandomPart(partType, manufacturer, rarity string) *PartDefinition {
	switch partType {
	case "barrel":
		if len(db.barrels) > 0 {
			return db.barrels[rand.Intn(len(db.barrels))]
		}
	}

	return nil
}

// Wrapper functions for int-level parameters

func (g *GeneratorService) generateDamageInt(level int, rarity string) int {
	return g.generateDamage(level, rarity)
}

func (g *GeneratorService) generateElementalDamageInt(level int, rarity string) int {
	return g.generateElementalDamage(level, rarity)
}

func (g *GeneratorService) generateShieldCapacityInt(level int, rarity string) int {
	return g.generateShieldCapacity(level, rarity)
}

func (g *GeneratorService) generateRechargeRateInt(level int, rarity string) float64 {
	return g.generateRechargeRate(level, rarity)
}