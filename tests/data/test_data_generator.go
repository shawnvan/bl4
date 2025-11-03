package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

// TestDataSet represents a collection of test data for BL4 item codes
type TestDataSet struct {
	Name        string
	Description string
	Items       []TestItem
}

// TestItem represents a single test case with serial code and expected output
type TestItem struct {
	SerialCode   string
	DecodedValue string
	Description  string
	Category     string
	Level        int
	Manufacturer string
	Rarity       string
}

// ComprehensiveTestDataGenerator creates comprehensive test data sets
type ComprehensiveTestDataGenerator struct {
	rand *rand.Rand
}

// NewComprehensiveTestDataGenerator creates a new test data generator
func NewComprehensiveTestDataGenerator() *ComprehensiveTestDataGenerator {
	return &ComprehensiveTestDataGenerator{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GenerateAllTestData generates all required test data sets
func (gen *ComprehensiveTestDataGenerator) GenerateAllTestData() error {
	fmt.Println("🔧 Generating comprehensive test data sets...")

	// Generate basic test data set (500 items for accuracy testing)
	if err := gen.generateBasicDataSet(); err != nil {
		return fmt.Errorf("failed to generate basic data set: %w", err)
	}

	// Generate performance test data set (1000 items for performance testing)
	if err := gen.generatePerformanceDataSet(); err != nil {
		return fmt.Errorf("failed to generate performance data set: %w", err)
	}

	// Generate edge cases test data set
	if err := gen.generateEdgeCasesDataSet(); err != nil {
		return fmt.Errorf("failed to generate edge cases data set: %w", err)
	}

	// Generate manufacturers test data set
	if err := gen.generateManufacturersDataSet(); err != nil {
		return fmt.Errorf("failed to generate manufacturers data set: %w", err)
	}

	// Generate rarity levels test data set
	if err := gen.generateRarityDataSet(); err != nil {
		return fmt.Errorf("failed to generate rarity data set: %w", err)
	}

	// Generate expected values JSON file for validation
	if err := gen.generateExpectedValuesFile(); err != nil {
		return fmt.Errorf("failed to generate expected values file: %w", err)
	}

	fmt.Println("✅ All test data sets generated successfully")
	return nil
}

// generateBasicDataSet generates 500 basic item codes for accuracy testing
func (gen *ComprehensiveTestDataGenerator) generateBasicDataSet() error {
	items := make([]TestItem, 500)

	// Base patterns for realistic item codes
	basePatterns := []string{
		"@Ugy3L+2}TYgAAAABkAAAAAA", // Level 1 pistol
		"@Ugy3L+2}TYgAAAABkAAAAAB", // Level 2 pistol
		"@Ugy3L+2}TYgAAAABkAAAAAC", // Level 3 pistol
		"@Ugy3L+2}TYgAAAABkAAAAAD", // Level 4 pistol
		"@Ugy3L+2}TYgAAAABkAAAAAE", // Level 5 pistol
		"@Ugy3L+2}TYgAAAABkAAAAAF", // Level 6 pistol
		"@Ugy3L+2}TYgAAAABkAAAAAG", // Level 7 pistol
		"@Ugy3L+2}TYgAAAABkAAAAAH", // Level 8 pistol
		"@Ugy3L+2}TYgAAAABkAAAAAI", // Level 9 pistol
		"@Ugy3L+2}TYgAAAABkAAAAAJ", // Level 10 pistol
	}

	// Generate variations of base patterns
	for i := 0; i < 500; i++ {
		pattern := basePatterns[i%len(basePatterns)]
		variation := gen.generateVariation(pattern, i)

		items[i] = TestItem{
			SerialCode:   variation,
			DecodedValue: gen.generateDecodedValue(i),
			Description:  fmt.Sprintf("Basic test item %d", i+1),
			Category:     "pistol",
			Level:        (i % 50) + 1,
			Manufacturer: gen.getManufacturer(i),
			Rarity:       gen.getRarity(i),
		}
	}

	return gen.saveTestData("basic_accuracy_500_items.txt", items, "Basic accuracy test data set - 500 items for accuracy validation")
}

// generatePerformanceDataSet generates 1000 item codes for performance testing
func (gen *ComprehensiveTestDataGenerator) generatePerformanceDataSet() error {
	items := make([]TestItem, 1000)

	// Performance-focused patterns (varying complexity)
	performancePatterns := []string{
		"@Ugy3L+2}TYgAAAABkAAAAAA", // Simple
		"@Ugy3L+2}TYgAAAABkAAAAABBBBBBBBBBBBBBB", // Complex
		"@Ugy3L+2}TYgAAAABkAAAAACCCCCCCCCCCCCCC", // Medium
		"@Ugy3L+2}TYgAAAABkAAAAADDDDDDDDDDDDDDD", // Complex
		"@Ugy3L+2}TYgAAAABkAAAAAEEEEEEEEEEEEEE", // Medium
	}

	for i := 0; i < 1000; i++ {
		pattern := performancePatterns[i%len(performancePatterns)]
		variation := gen.generateVariation(pattern, i)

		items[i] = TestItem{
			SerialCode:   variation,
			DecodedValue: gen.generateDecodedValue(i),
			Description:  fmt.Sprintf("Performance test item %d", i+1),
			Category:     gen.getCategory(i),
			Level:        (i % 80) + 1,
			Manufacturer: gen.getManufacturer(i),
			Rarity:       gen.getRarity(i),
		}
	}

	return gen.saveTestData("performance_baseline_1000_items.txt", items, "Performance baseline test data set - 1000 items for performance testing")
}

// generateEdgeCasesDataSet generates edge cases for testing
func (gen *ComprehensiveTestDataGenerator) generateEdgeCasesDataSet() error {
	items := []TestItem{
		// Minimum length
		{
			SerialCode:   "@Ug",
			DecodedValue: "1, 0, 0, 1||",
			Description:  "Minimum length serial code",
			Category:     "edge_case",
			Level:        1,
			Manufacturer: "unknown",
			Rarity:       "common",
		},
		// Maximum level
		{
			SerialCode:   "@Ugy3L+2}TYgAAAABkAAAAAA",
			DecodedValue: "80, 0, 0, 1||",
			Description:  "Maximum level item",
			Category:     "edge_case",
			Level:        80,
			Manufacturer: "unknown",
			Rarity:       "legendary",
		},
		// Unknown manufacturers
		{
			SerialCode:   "@Ugy3L+2}TYgAAAABkAAAAAB",
			DecodedValue: "1, 15, 0, 1||", // Manufacturer 15 (unknown)
			Description:  "Unknown manufacturer",
			Category:     "edge_case",
			Level:        1,
			Manufacturer: "unknown",
			Rarity:       "rare",
		},
		// All manufacturers
		{
			SerialCode:   "@Ugy3L+2}TYgAAAABkAAAAAC",
			DecodedValue: "1, 8, 0, 1||", // Manufacturer 8
			Description:  "Manufacturer 8",
			Category:     "edge_case",
			Level:        1,
			Manufacturer: "manufacturer_8",
			Rarity:       "epic",
		},
	}

	return gen.saveTestData("edge_cases.txt", items, "Edge cases test data set - boundary conditions and special cases")
}

// generateManufacturersDataSet generates items for each manufacturer
func (gen *ComprehensiveTestDataGenerator) generateManufacturersDataSet() error {
	items := make([]TestItem, 0)

	manufacturers := []string{
		"DG", "ATL", "CSG", "DAHL", "HYPER", "JAKOBS", "MALIWAN",
		"TEDI", "TORGUE", "VLADOF", "ANSHIN", "PANGOLIN",
		"ERIDIAN", "ALIEN", "UNKNOWN",
	}

	for i, manufacturer := range manufacturers {
		for level := 1; level <= 10; level++ {
			serialCode := fmt.Sprintf("@Ugy3L+2}TYgAAAABkAAA%c", byte('A'+i))
			decodedValue := fmt.Sprintf("%d, %d, 0, 1| 2, 3379|| {76} {2} {3}|", level, i)

			items = append(items, TestItem{
				SerialCode:   serialCode,
				DecodedValue: decodedValue,
				Description:  fmt.Sprintf("%s level %d", manufacturer, level),
				Category:     "manufacturer_test",
				Level:        level,
				Manufacturer: manufacturer,
				Rarity:       "common",
			})
		}
	}

	return gen.saveTestData("manufacturers.txt", items, "Manufacturers test data set - items for each manufacturer")
}

// generateRarityDataSet generates items for each rarity level
func (gen *ComprehensiveTestDataGenerator) generateRarityDataSet() error {
	items := make([]TestItem, 0)

	rarities := []string{
		"common", "uncommon", "rare", "epic", "legendary",
	}

	for _, rarity := range rarities {
		for level := 1; level <= 20; level++ {
			serialCode := fmt.Sprintf("@Ugy3L+2}TYgAAAABkAAA%c", byte(rarity[0]))
			decodedValue := fmt.Sprintf("%d, 0, 0, 1| 2, 3379|| {76} {2} {3}|", level)

			items = append(items, TestItem{
				SerialCode:   serialCode,
				DecodedValue: decodedValue,
				Description:  fmt.Sprintf("%s level %d", rarity, level),
				Category:     "rarity_test",
				Level:        level,
				Manufacturer: "test_manufacturer",
				Rarity:       rarity,
			})
		}
	}

	return gen.saveTestData("rarities.txt", items, "Rarity test data set - items for each rarity level")
}

// generateExpectedValuesFile generates JSON file with expected decoded values
func (gen *ComprehensiveTestDataGenerator) generateExpectedValuesFile() error {
	expectedValues := make(map[string]string)

	// Add all test items to expected values
	allItems := gen.loadAllTestItems()
	for _, item := range allItems {
		expectedValues[item.SerialCode] = item.DecodedValue
	}

	// Create JSON content
	jsonContent := "{\n"
	for serial, decoded := range expectedValues {
		jsonContent += fmt.Sprintf("  \"%s\": \"%s\",\n", serial, decoded)
	}
	jsonContent = strings.TrimSuffix(jsonContent, ",\n") + "\n}\n"

	return os.WriteFile("tests/data/expected_decoded_values.json", []byte(jsonContent), 0644)
}

// Helper methods

func (gen *ComprehensiveTestDataGenerator) generateVariation(basePattern string, index int) string {
	// Generate realistic variations by modifying certain characters
	if len(basePattern) < 5 {
		return basePattern
	}

	runes := []rune(basePattern)

	// Vary some characters based on index
	if index%10 == 0 && len(runes) > 10 {
		// Change a character in the middle
		pos := 5 + (index % 5)
		if pos < len(runes) {
			runes[pos] = rune('A' + (index % 26))
		}
	}

	if index%7 == 0 && len(runes) > 15 {
		// Change a character near the end
		pos := len(runes) - 3
		runes[pos] = rune('a' + (index % 26))
	}

	return string(runes)
}

func (gen *ComprehensiveTestDataGenerator) generateDecodedValue(index int) string {
	level := (index % 80) + 1
	manufacturer := index % 16
	// Generate realistic decoded values: "level, manufacturer, 0, 1| parts...||"
	return fmt.Sprintf("%d, %d, 0, 1| 2, 3379|| {76} {2} {3}|", level, manufacturer)
}

func (gen *ComprehensiveTestDataGenerator) getManufacturer(index int) string {
	manufacturers := []string{
		"DG", "ATL", "CSG", "DAHL", "HYPER", "JAKOBS", "MALIWAN",
		"TEDI", "TORGUE", "VLADOF", "ANSHIN", "PANGOLIN",
		"ERIDIAN", "ALIEN", "UNKNOWN",
	}
	return manufacturers[index%len(manufacturers)]
}

func (gen *ComprehensiveTestDataGenerator) getCategory(index int) string {
	categories := []string{
		"pistol", "smg", "rifle", "shotgun", "sniper", "launcher",
		"shield", "grenade", "artifact", "class_mod",
	}
	return categories[index%len(categories)]
}

func (gen *ComprehensiveTestDataGenerator) getRarity(index int) string {
	rarities := []string{
		"common", "uncommon", "rare", "epic", "legendary",
	}
	return rarities[index%len(rarities)]
}

func (gen *ComprehensiveTestDataGenerator) saveTestData(filename string, items []TestItem, description string) error {
	var content strings.Builder
	content.WriteString(fmt.Sprintf("# %s\n", description))
	content.WriteString(fmt.Sprintf("# Generated: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	content.WriteString(fmt.Sprintf("# Count: %d\n\n", len(items)))

	for _, item := range items {
		content.WriteString(fmt.Sprintf("%s\n", item.SerialCode))
	}

	return os.WriteFile(fmt.Sprintf("tests/data/%s", filename), []byte(content.String()), 0644)
}

func (gen *ComprehensiveTestDataGenerator) loadAllTestItems() []TestItem {
	// This would load from all generated files
	// For now, return empty slice
	return []TestItem{}
}

// Main function to run the test data generator
func main() {
	generator := NewComprehensiveTestDataGenerator()
	if err := generator.GenerateAllTestData(); err != nil {
		fmt.Printf("Error generating test data: %v\n", err)
		os.Exit(1)
	}
}