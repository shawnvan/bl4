package compatibility

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CompatibilityHarness provides comprehensive backward compatibility testing
type CompatibilityHarness struct {
	resultsDir string
	logFile    *os.File
}

// NewCompatibilityHarness creates a new compatibility harness instance
func NewCompatibilityHarness() (*CompatibilityHarness, error) {
	resultsDir := "tests/compatibility/results"
	if err := os.MkdirAll(resultsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create compatibility results directory: %w", err)
	}

	logFile, err := os.OpenFile(filepath.Join(resultsDir, "compatibility.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create compatibility log file: %w", err)
	}

	return &CompatibilityHarness{
		resultsDir: resultsDir,
		logFile:    logFile,
	}, nil
}

// Close closes the compatibility harness
func (ch *CompatibilityHarness) Close() error {
	if ch.logFile != nil {
		return ch.logFile.Close()
	}
	return nil
}

// CompatibilityTestResult represents the result of a compatibility test
type CompatibilityTestResult struct {
	TestName        string            `json:"test_name"`
	SerialCode      string            `json:"serial_code"`
	CurrentOutput   string            `json:"current_output"`
	NewOutput       string            `json:"new_output"`
	Matches         bool              `json:"matches"`
	ErrorType       string            `json:"error_type,omitempty"`
	ErrorMessage    string            `json:"error_message,omitempty"`
	ProcessingTime  time.Duration     `json:"processing_time_ns"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	Timestamp       time.Time         `json:"timestamp"`
}

// CompatibilitySuite represents a complete compatibility test suite
type CompatibilitySuite struct {
	SuiteName         string                      `json:"suite_name"`
	Description       string                      `json:"description"`
	TestResults       []*CompatibilityTestResult `json:"test_results"`
	StartTime         time.Time                   `json:"start_time"`
	EndTime           time.Time                   `json:"end_time"`
	TotalTests        int                         `json:"total_tests"`
	CompatibleTests   int                         `json:"compatible_tests"`
	IncompatibleTests int                         `json:"incompatible_tests"`
	CompatibilityRate float64                     `json:"compatibility_rate"`
	AverageTime       time.Duration               `json:"average_time_ns"`
	MinTime           time.Duration               `json:"min_time_ns"`
	MaxTime           time.Duration               `json:"max_time_ns"`
}

// RunCompatibilityTest runs comprehensive compatibility verification
func (ch *CompatibilityHarness) RunCompatibilityTest(
	testDataFile string,
	currentAlgorithm func(string) (string, error),
	newAlgorithm func(string) (string, error),
) (*CompatibilitySuite, error) {
	fmt.Println("🔒 Running backward compatibility verification...")

	suite := &CompatibilitySuite{
		SuiteName:   "Backward Compatibility Verification",
		Description: "Ensure 100% backward compatibility between current and new implementations",
		StartTime:   time.Now(),
		TestResults: make([]*CompatibilityTestResult, 0),
	}

	// Load test data
	testData, err := ch.loadTestData(testDataFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load test data: %w", err)
	}

	// Run compatibility tests
	var totalTime time.Duration
	var minTime, maxTime time.Duration

	for _, serialCode := range testData {
		if serialCode == "" || serialCode[0] == '#' {
			continue // Skip empty lines and comments
		}

		result := &CompatibilityTestResult{
			TestName:   fmt.Sprintf("Compat %s", serialCode[:min(20, len(serialCode))]),
			SerialCode: serialCode,
			Timestamp:  time.Now(),
			Metadata:   make(map[string]string),
		}

		// Run current algorithm
		start := time.Now()
		currentOutput, currentErr := currentAlgorithm(serialCode)
		currentTime := time.Since(start)

		// Run new algorithm
		start = time.Now()
		newOutput, newErr := newAlgorithm(serialCode)
		newTime := time.Since(start)

		result.ProcessingTime = currentTime + newTime

		// Check for errors
		if currentErr != nil && newErr != nil {
			// Both algorithms failed - check if errors match
			result.CurrentOutput = currentErr.Error()
			result.NewOutput = newErr.Error()
			result.Matches = (result.CurrentOutput == result.NewOutput)
			if !result.Matches {
				result.ErrorType = "DifferentErrorTypes"
				result.ErrorMessage = fmt.Sprintf("Current: %v, New: %v", currentErr, newErr)
			}
		} else if currentErr != nil {
			result.CurrentOutput = currentErr.Error()
			result.NewOutput = newOutput
			result.Matches = false
			result.ErrorType = "CurrentAlgorithmError"
			result.ErrorMessage = currentErr.Error()
		} else if newErr != nil {
			result.CurrentOutput = currentOutput
			result.NewOutput = newErr.Error()
			result.Matches = false
			result.ErrorType = "NewAlgorithmError"
			result.ErrorMessage = newErr.Error()
		} else {
			// Both algorithms succeeded - compare outputs
			result.CurrentOutput = currentOutput
			result.NewOutput = newOutput
			result.Matches = (currentOutput == newOutput)
			if !result.Matches {
				result.ErrorType = "OutputMismatch"
				result.ErrorMessage = "Outputs differ between algorithms"
			}
		}

		// Update timing statistics
		totalTime += result.ProcessingTime
		if minTime == 0 || result.ProcessingTime < minTime {
			minTime = result.ProcessingTime
		}
		if maxTime == 0 || result.ProcessingTime > maxTime {
			maxTime = result.ProcessingTime
		}

		suite.TestResults = append(suite.TestResults, result)

		// Log result
		status := "✅ COMPATIBLE"
		if !result.Matches {
			status = "❌ INCOMPATIBLE"
		}
		fmt.Printf("  %s %s (%v)\n", status, result.TestName, result.ProcessingTime)
	}

	// Calculate suite statistics
	suite.EndTime = time.Now()
	suite.TotalTests = len(suite.TestResults)

	for _, result := range suite.TestResults {
		if result.Matches {
			suite.CompatibleTests++
		} else {
			suite.IncompatibleTests++
		}
	}

	if suite.TotalTests > 0 {
		suite.CompatibilityRate = float64(suite.CompatibleTests) / float64(suite.TotalTests)
		suite.AverageTime = totalTime / time.Duration(suite.TotalTests)
	}
	suite.MinTime = minTime
	suite.MaxTime = maxTime

	// Save compatibility results
	if err := ch.saveCompatibilitySuite(suite); err != nil {
		return nil, fmt.Errorf("failed to save compatibility results: %w", err)
	}

	// Log to file
	ch.logCompatibilityResult(suite)

	fmt.Printf("✅ Compatibility verification completed: %d/%d compatible (%.2f%% compatibility)\n",
		suite.CompatibleTests, suite.TotalTests, suite.CompatibilityRate*100)

	// Check for 100% compatibility requirement
	if suite.CompatibilityRate < 1.0 {
		fmt.Printf("⚠️  WARNING: Compatibility rate is %.2f%%, required 100%%\n", suite.CompatibilityRate*100)
		fmt.Printf("    Found %d incompatible results out of %d tests\n", suite.IncompatibleTests, suite.TotalTests)
	}

	return suite, nil
}

// RunPerformanceComparison runs performance comparison between algorithms
func (ch *CompatibilityHarness) RunPerformanceComparison(
	testDataFile string,
	currentAlgorithm func(string) (string, error),
	newAlgorithm func(string) (string, error),
) (*PerformanceComparisonSuite, error) {
	fmt.Println("⚡ Running performance comparison...")

	suite := &PerformanceComparisonSuite{
		SuiteName:   "Performance Comparison",
		Description: "Compare performance between current and new implementations",
		StartTime:   time.Now(),
		Results:     make([]*PerformanceComparisonResult, 0),
	}

	// Load test data
	testData, err := ch.loadTestData(testDataFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load test data: %w", err)
	}

	// Run performance comparison
	var totalCurrentTime, totalNewTime time.Duration
	var minImprovement, maxImprovement float64

	for _, serialCode := range testData {
		if serialCode == "" || serialCode[0] == '#' {
			continue
		}

		result := &PerformanceComparisonResult{
			TestName:   fmt.Sprintf("Perf %s", serialCode[:min(15, len(serialCode))]),
			SerialCode: serialCode,
			Timestamp:  time.Now(),
		}

		// Measure current algorithm performance
		start := time.Now()
		_, _ = currentAlgorithm(serialCode) // Error ignored for performance testing
		result.CurrentTime = time.Since(start)

		// Measure new algorithm performance
		start = time.Now()
		_, _ = newAlgorithm(serialCode) // Error ignored for performance testing
		result.NewTime = time.Since(start)

		// Calculate improvement
		if result.CurrentTime > 0 {
			result.Improvement = (float64(result.CurrentTime-result.NewTime) / float64(result.CurrentTime)) * 100
		}

		// Update totals
		totalCurrentTime += result.CurrentTime
		totalNewTime += result.NewTime

		// Track improvement range
		if minImprovement == 0 || result.Improvement < minImprovement {
			minImprovement = result.Improvement
		}
		if maxImprovement == 0 || result.Improvement > maxImprovement {
			maxImprovement = result.Improvement
		}

		suite.Results = append(suite.Results, result)

		// Show progress every 100 items
		if len(suite.Results)%100 == 0 {
			fmt.Printf("  Compared %d items...\n", len(suite.Results))
		}
	}

	// Calculate suite statistics
	suite.EndTime = time.Now()
	suite.TotalTests = len(suite.Results)

	if suite.TotalTests > 0 {
		suite.AverageCurrentTime = totalCurrentTime / time.Duration(suite.TotalTests)
		suite.AverageNewTime = totalNewTime / time.Duration(suite.TotalTests)
		suite.AverageImprovement = (float64(suite.AverageCurrentTime-suite.AverageNewTime) / float64(suite.AverageCurrentTime)) * 100
	}

	suite.MinImprovement = minImprovement
	suite.MaxImprovement = maxImprovement

	// Save performance comparison results
	if err := ch.savePerformanceComparisonSuite(suite); err != nil {
		return nil, fmt.Errorf("failed to save performance comparison results: %w", err)
	}

	fmt.Printf("✅ Performance comparison completed:\n")
	fmt.Printf("   - Average current time: %v\n", suite.AverageCurrentTime)
	fmt.Printf("   - Average new time: %v\n", suite.AverageNewTime)
	fmt.Printf("   - Average improvement: %.2f%%\n", suite.AverageImprovement)
	fmt.Printf("   - Improvement range: %.2f%% to %.2f%%\n", suite.MinImprovement, suite.MaxImprovement)

	return suite, nil
}

// PerformanceComparisonSuite represents performance comparison results
type PerformanceComparisonSuite struct {
	SuiteName         string                         `json:"suite_name"`
	Description       string                         `json:"description"`
	Results           []*PerformanceComparisonResult `json:"results"`
	StartTime         time.Time                      `json:"start_time"`
	EndTime           time.Time                      `json:"end_time"`
	TotalTests        int                            `json:"total_tests"`
	AverageCurrentTime time.Duration                 `json:"average_current_time_ns"`
	AverageNewTime    time.Duration                  `json:"average_new_time_ns"`
	AverageImprovement float64                        `json:"average_improvement_percent"`
	MinImprovement    float64                        `json:"min_improvement_percent"`
	MaxImprovement    float64                        `json:"max_improvement_percent"`
}

// PerformanceComparisonResult represents a single performance comparison
type PerformanceComparisonResult struct {
	TestName      string        `json:"test_name"`
	SerialCode    string        `json:"serial_code"`
	CurrentTime   time.Duration `json:"current_time_ns"`
	NewTime       time.Duration `json:"new_time_ns"`
	Improvement   float64       `json:"improvement_percent"`
	Timestamp     time.Time     `json:"timestamp"`
}

// Helper methods

func (ch *CompatibilityHarness) loadTestData(filename string) ([]string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var items []string
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && line[0] != '#' {
			items = append(items, line)
		}
	}

	return items, nil
}

func (ch *CompatibilityHarness) saveCompatibilitySuite(suite *CompatibilitySuite) error {
	data, err := json.MarshalIndent(suite, "", "  ")
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("compatibility_%s.json", time.Now().Format("20060102_150405"))
	filePath := filepath.Join(ch.resultsDir, filename)
	return os.WriteFile(filePath, data, 0644)
}

func (ch *CompatibilityHarness) savePerformanceComparisonSuite(suite *PerformanceComparisonSuite) error {
	data, err := json.MarshalIndent(suite, "", "  ")
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("performance_comparison_%s.json", time.Now().Format("20060102_150405"))
	filePath := filepath.Join(ch.resultsDir, filename)
	return os.WriteFile(filePath, data, 0644)
}

func (ch *CompatibilityHarness) logCompatibilityResult(suite *CompatibilitySuite) {
	logEntry := fmt.Sprintf("[%s] %s: %d/%d tests compatible (%.2f%% compatibility)\n",
		time.Now().Format("2006-01-02 15:04:05"),
		suite.SuiteName,
		suite.CompatibleTests,
		suite.TotalTests,
		suite.CompatibilityRate*100)

	if ch.logFile != nil {
		ch.logFile.WriteString(logEntry)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}