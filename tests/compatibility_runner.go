package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/shawnvan/bl4/tests/compatibility"
)

// Algorithm implementation functions for compatibility testing
// These would be replaced with actual algorithm implementations

func currentAlgorithmDecode(serial string) (string, error) {
	// Placeholder for current algorithm implementation
	// This would be replaced with the actual current decoding algorithm
	return fmt.Sprintf("decoded_%s", serial), nil
}

func newAlgorithmDecode(serial string) (string, error) {
	// Placeholder for new algorithm implementation
	// This would be replaced with the actual new decoding algorithm
	return fmt.Sprintf("decoded_%s", serial), nil
}

// CompatibilityTestRunner runs compatibility tests between algorithm implementations
type CompatibilityTestRunner struct {
	harness *compatibility.CompatibilityHarness
}

// NewCompatibilityTestRunner creates a new compatibility test runner
func NewCompatibilityTestRunner() (*CompatibilityTestRunner, error) {
	harness, err := compatibility.NewCompatibilityHarness()
	if err != nil {
		return nil, err
	}

	return &CompatibilityTestRunner{
		harness: harness,
	}, nil
}

// Close closes the compatibility test runner
func (ctr *CompatibilityTestRunner) Close() error {
	return ctr.harness.Close()
}

// RunAllCompatibilityTests runs all compatibility tests
func (ctr *CompatibilityTestRunner) RunAllCompatibilityTests() error {
	fmt.Println("🔒 Starting comprehensive compatibility testing...")
	fmt.Println("=================================================")

	// Test data files for compatibility testing
	testFiles := map[string]string{
		"basic_accuracy":  "tests/data/basic_accuracy_500_items.txt",
		"performance":     "tests/data/performance_baseline_1000_items.txt",
		"edge_cases":      "tests/data/edge_cases.txt",
		"manufacturers":   "tests/data/manufacturers.txt",
		"rarities":        "tests/data/rarities.txt",
	}

	// Run compatibility tests for each test file
	for testName, testFile := range testFiles {
		if _, err := os.Stat(testFile); os.IsNotExist(err) {
			fmt.Printf("⚠️  Test file not found: %s\n", testFile)
			continue
		}

		fmt.Printf("\n🔒 Running %s compatibility test...\n", testName)

		// Run backward compatibility verification
		suite, err := ctr.harness.RunCompatibilityTest(testFile, currentAlgorithmDecode, newAlgorithmDecode)
		if err != nil {
			fmt.Printf("❌ Compatibility test failed: %v\n", err)
			continue
		}

		// Save summary
		if err := ctr.saveCompatibilitySummary(testName, suite); err != nil {
			fmt.Printf("⚠️  Failed to save compatibility summary: %v\n", err)
		}

		// Check for 100% compatibility requirement
		if suite.CompatibilityRate < 1.0 {
			fmt.Printf("❌ COMPATIBILITY REQUIREMENT FAILED: %.2f%% < 100%%\n", suite.CompatibilityRate*100)
			fmt.Printf("    Must fix %d incompatible results before proceeding\n", suite.IncompatibleTests)
		} else {
			fmt.Printf("✅ 100%% compatibility requirement satisfied\n")
		}
	}

	// Run performance comparison
	fmt.Printf("\n⚡ Running performance comparison...\n")
	perfFile := testFiles["performance"]
	if _, err := os.Stat(perfFile); err == nil {
		perfSuite, err := ctr.harness.RunPerformanceComparison(perfFile, currentAlgorithmDecode, newAlgorithmDecode)
		if err != nil {
			fmt.Printf("❌ Performance comparison failed: %v\n", err)
		} else {
			// Save performance summary
			if err := ctr.savePerformanceSummary(perfSuite); err != nil {
				fmt.Printf("⚠️  Failed to save performance summary: %v\n", err)
			}

			// Check performance improvement requirement
			if perfSuite.AverageImprovement >= 10.0 {
				fmt.Printf("✅ Performance improvement requirement satisfied: %.2f%% ≥ 10%%\n", perfSuite.AverageImprovement)
			} else {
				fmt.Printf("⚠️  Performance improvement target not met: %.2f%% < 10%%\n", perfSuite.AverageImprovement)
			}
		}
	}

	// Generate compatibility report
	fmt.Printf("\n📊 Generating compatibility report...\n")
	if err := ctr.generateCompatibilityReport(); err != nil {
		fmt.Printf("❌ Failed to generate compatibility report: %v\n", err)
	}

	fmt.Println("\n✅ All compatibility tests completed!")
	return nil
}

// CompatibilitySummary represents a summary of compatibility test results
type CompatibilitySummary struct {
	TestName           string        `json:"test_name"`
	TotalTests         int           `json:"total_tests"`
	CompatibleTests    int           `json:"compatible_tests"`
	IncompatibleTests  int           `json:"incompatible_tests"`
	CompatibilityRate  float64       `json:"compatibility_rate"`
	AverageTime        time.Duration `json:"average_time_ns"`
	MinTime            time.Duration `json:"min_time_ns"`
	MaxTime            time.Duration `json:"max_time_ns"`
	Timestamp          time.Time     `json:"timestamp"`
	RequirementMet     bool          `json:"requirement_met"`
}

func (ctr *CompatibilityTestRunner) saveCompatibilitySummary(testName string, suite *compatibility.CompatibilitySuite) error {
	summary := &CompatibilitySummary{
		TestName:          testName,
		TotalTests:        suite.TotalTests,
		CompatibleTests:   suite.CompatibleTests,
		IncompatibleTests: suite.IncompatibleTests,
		CompatibilityRate: suite.CompatibilityRate,
		AverageTime:       suite.AverageTime,
		MinTime:           suite.MinTime,
		MaxTime:           suite.MaxTime,
		Timestamp:         time.Now(),
		RequirementMet:    suite.CompatibilityRate >= 1.0,
	}

	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("compatibility_summary_%s.json", testName)
	filePath := filepath.Join("tests/compatibility/results", filename)
	return os.WriteFile(filePath, data, 0644)
}

// PerformanceSummary represents a summary of performance comparison results
type PerformanceSummary struct {
	TotalTests            int           `json:"total_tests"`
	AverageCurrentTime    time.Duration `json:"average_current_time_ns"`
	AverageNewTime        time.Duration `json:"average_new_time_ns"`
	AverageImprovement    float64       `json:"average_improvement_percent"`
	MinImprovement        float64       `json:"min_improvement_percent"`
	MaxImprovement        float64       `json:"max_improvement_percent"`
	Timestamp             time.Time     `json:"timestamp"`
	PerformanceTargetMet  bool          `json:"performance_target_met"`
}

func (ctr *CompatibilityTestRunner) savePerformanceSummary(suite *compatibility.PerformanceComparisonSuite) error {
	summary := &PerformanceSummary{
		TotalTests:           suite.TotalTests,
		AverageCurrentTime:   suite.AverageCurrentTime,
		AverageNewTime:       suite.AverageNewTime,
		AverageImprovement:   suite.AverageImprovement,
		MinImprovement:       suite.MinImprovement,
		MaxImprovement:       suite.MaxImprovement,
		Timestamp:            time.Now(),
		PerformanceTargetMet: suite.AverageImprovement >= 10.0,
	}

	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}

	filePath := filepath.Join("tests/compatibility/results", "performance_summary.json")
	return os.WriteFile(filePath, data, 0644)
}

func (ctr *CompatibilityTestRunner) generateCompatibilityReport() error {
	// This would analyze all compatibility results and generate a comprehensive report
	// For now, create a placeholder report
	report := map[string]interface{}{
		"generated_at":      time.Now(),
		"requirement":       "100% backward compatibility",
		"performance_target": "≥10% improvement",
		"test_types":        []string{"accuracy", "performance", "edge_cases", "manufacturers", "rarities"},
		"summary":           "Comprehensive compatibility verification report generated",
		"status":            "ready_for_analysis",
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	filePath := filepath.Join("tests/compatibility/results", "compatibility_report.json")
	return os.WriteFile(filePath, data, 0644)
}

func main() {
	runner, err := NewCompatibilityTestRunner()
	if err != nil {
		log.Fatalf("Failed to create compatibility test runner: %v", err)
	}
	defer runner.Close()

	if err := runner.RunAllCompatibilityTests(); err != nil {
		log.Fatalf("Compatibility tests failed: %v", err)
	}
}