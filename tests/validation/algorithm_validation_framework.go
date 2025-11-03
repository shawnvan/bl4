package validation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// AlgorithmValidationFramework provides comprehensive validation for codec algorithms
type AlgorithmValidationFramework struct {
	validationDir string
	logFile       *os.File
}

// NewAlgorithmValidationFramework creates a new validation framework instance
func NewAlgorithmValidationFramework() (*AlgorithmValidationFramework, error) {
	validationDir := "tests/validation/results"
	if err := os.MkdirAll(validationDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create validation results directory: %w", err)
	}

	logFile, err := os.OpenFile(filepath.Join(validationDir, "validation.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create validation log file: %w", err)
	}

	return &AlgorithmValidationFramework{
		validationDir: validationDir,
		logFile:       logFile,
	}, nil
}

// Close closes the validation framework
func (avf *AlgorithmValidationFramework) Close() error {
	if avf.logFile != nil {
		return avf.logFile.Close()
	}
	return nil
}

// ValidationTestResult represents the result of a single validation test
type ValidationTestResult struct {
	TestName        string            `json:"test_name"`
	TestType        string            `json:"test_type"`
	SerialCode      string            `json:"serial_code"`
	ExpectedOutput  string            `json:"expected_output"`
	ActualOutput    string            `json:"actual_output"`
	Matches         bool              `json:"matches"`
	ErrorType       string            `json:"error_type,omitempty"`
	ErrorMessage    string            `json:"error_message,omitempty"`
	ProcessingTime  time.Duration     `json:"processing_time_ns"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	Timestamp       time.Time         `json:"timestamp"`
}

// ValidationSuite represents a complete validation suite
type ValidationSuite struct {
	SuiteName       string                  `json:"suite_name"`
	Description     string                  `json:"description"`
	TestResults     []*ValidationTestResult `json:"test_results"`
	StartTime       time.Time               `json:"start_time"`
	EndTime         time.Time               `json:"end_time"`
	TotalTests      int                     `json:"total_tests"`
	PassedTests     int                     `json:"passed_tests"`
	FailedTests     int                     `json:"failed_tests"`
	AccuracyRate    float64                 `json:"accuracy_rate"`
	AverageTime     time.Duration           `json:"average_time_ns"`
	MinTime         time.Duration           `json:"min_time_ns"`
	MaxTime         time.Duration           `json:"max_time_ns"`
}

// RunAccuracyValidation runs accuracy validation against expected values
func (avf *AlgorithmValidationFramework) RunAccuracyValidation(testDataFile string, algorithmFunc func(string) (string, error)) (*ValidationSuite, error) {
	fmt.Println("🔍 Running accuracy validation...")

	suite := &ValidationSuite{
		SuiteName:   "Accuracy Validation",
		Description: "Validate algorithm accuracy against expected decoded values",
		StartTime:   time.Now(),
		TestResults: make([]*ValidationTestResult, 0),
	}

	// Load test data
	testData, err := avf.loadTestData(testDataFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load test data: %w", err)
	}

	// Load expected values
	expectedValues, err := avf.loadExpectedValues("tests/data/expected_decoded_values.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load expected values: %w", err)
	}

	// Run validation tests
	var totalTime time.Duration
	var minTime, maxTime time.Duration

	for _, serialCode := range testData {
		if serialCode == "" || serialCode[0] == '#' {
			continue // Skip empty lines and comments
		}

		result := &ValidationTestResult{
			TestName:   fmt.Sprintf("Decode %s", serialCode[:min(20, len(serialCode))]),
			TestType:   "accuracy",
			SerialCode: serialCode,
			Timestamp:  time.Now(),
			Metadata:   make(map[string]string),
		}

		// Get expected value
		expected, exists := expectedValues[serialCode]
		if !exists {
			result.ExpectedOutput = "UNKNOWN"
			result.ErrorType = "MissingExpectedValue"
			result.ErrorMessage = "No expected value found for serial code"
		} else {
			result.ExpectedOutput = expected
		}

		// Run algorithm
		start := time.Now()
		actual, err := algorithmFunc(serialCode)
		result.ProcessingTime = time.Since(start)
		result.ActualOutput = actual

		if err != nil {
			result.ErrorType = "AlgorithmError"
			result.ErrorMessage = err.Error()
			result.Matches = false
		} else {
			// Compare results
			result.Matches = (actual == expected)
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
		status := "✅ PASS"
		if !result.Matches {
			status = "❌ FAIL"
		}
		fmt.Printf("  %s %s (%v)\n", status, result.TestName, result.ProcessingTime)
	}

	// Calculate suite statistics
	suite.EndTime = time.Now()
	suite.TotalTests = len(suite.TestResults)

	for _, result := range suite.TestResults {
		if result.Matches {
			suite.PassedTests++
		} else {
			suite.FailedTests++
		}
	}

	if suite.TotalTests > 0 {
		suite.AccuracyRate = float64(suite.PassedTests) / float64(suite.TotalTests)
		suite.AverageTime = totalTime / time.Duration(suite.TotalTests)
	}
	suite.MinTime = minTime
	suite.MaxTime = maxTime

	// Save validation results
	if err := avf.saveValidationSuite(suite); err != nil {
		return nil, fmt.Errorf("failed to save validation results: %w", err)
	}

	// Log to file
	avf.logValidationResult(suite)

	fmt.Printf("✅ Accuracy validation completed: %d/%d passed (%.2f%% accuracy)\n",
		suite.PassedTests, suite.TotalTests, suite.AccuracyRate*100)

	return suite, nil
}

// RunPerformanceValidation runs performance validation benchmarks
func (avf *AlgorithmValidationFramework) RunPerformanceValidation(testDataFile string, algorithmFunc func(string) (string, error)) (*ValidationSuite, error) {
	fmt.Println("⚡ Running performance validation...")

	suite := &ValidationSuite{
		SuiteName:   "Performance Validation",
		Description: "Validate algorithm performance against baseline metrics",
		StartTime:   time.Now(),
		TestResults: make([]*ValidationTestResult, 0),
	}

	// Load test data
	testData, err := avf.loadTestData(testDataFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load test data: %w", err)
	}

	// Run performance tests
	var totalTime time.Duration
	var minTime, maxTime time.Duration

	testCount := 0
	for _, serialCode := range testData {
		if serialCode == "" || serialCode[0] == '#' {
			continue
		}

		result := &ValidationTestResult{
			TestName:   fmt.Sprintf("Perf %s", serialCode[:min(15, len(serialCode))]),
			TestType:   "performance",
			SerialCode: serialCode,
			Timestamp:  time.Now(),
			Metadata:   make(map[string]string),
		}

		// Run algorithm and measure time
		start := time.Now()
		_, err := algorithmFunc(serialCode)
		result.ProcessingTime = time.Since(start)

		if err != nil {
			result.ErrorType = "PerformanceError"
			result.ErrorMessage = err.Error()
			result.Matches = false
		} else {
			result.Matches = true // Performance tests pass if they complete without error
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
		testCount++

		// Show progress every 100 items
		if testCount%100 == 0 {
			fmt.Printf("  Processed %d items...\n", testCount)
		}
	}

	// Calculate suite statistics
	suite.EndTime = time.Now()
	suite.TotalTests = len(suite.TestResults)
	suite.PassedTests = suite.TotalTests // Performance tests pass if they complete
	suite.FailedTests = 0

	if suite.TotalTests > 0 {
		suite.AccuracyRate = 1.0 // 100% for performance tests (they either run or fail)
		suite.AverageTime = totalTime / time.Duration(suite.TotalTests)
	}
	suite.MinTime = minTime
	suite.MaxTime = maxTime

	// Save validation results
	if err := avf.saveValidationSuite(suite); err != nil {
		return nil, fmt.Errorf("failed to save validation results: %w", err)
	}

	// Calculate throughput
	throughput := float64(suite.TotalTests) / suite.EndTime.Sub(suite.StartTime).Seconds()

	fmt.Printf("✅ Performance validation completed:\n")
	fmt.Printf("   - Items processed: %d\n", suite.TotalTests)
	fmt.Printf("   - Average time: %v\n", suite.AverageTime)
	fmt.Printf("   - Min time: %v\n", suite.MinTime)
	fmt.Printf("   - Max time: %v\n", suite.MaxTime)
	fmt.Printf("   - Throughput: %.2f items/sec\n", throughput)

	return suite, nil
}

// Helper methods

func (avf *AlgorithmValidationFramework) loadTestData(filename string) ([]string, error) {
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

func (avf *AlgorithmValidationFramework) loadExpectedValues(filename string) (map[string]string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var expectedValues map[string]string
	if err := json.Unmarshal(data, &expectedValues); err != nil {
		return nil, err
	}

	return expectedValues, nil
}

func (avf *AlgorithmValidationFramework) saveValidationSuite(suite *ValidationSuite) error {
	data, err := json.MarshalIndent(suite, "", "  ")
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("validation_%s_%s.json",
		suite.SuiteName,
		time.Now().Format("20060102_150405"))
	filePath := filepath.Join(avf.validationDir, filename)
	return os.WriteFile(filePath, data, 0644)
}

func (avf *AlgorithmValidationFramework) logValidationResult(suite *ValidationSuite) {
	logEntry := fmt.Sprintf("[%s] %s: %d/%d tests passed (%.2f%% accuracy, avg time: %v)\n",
		time.Now().Format("2006-01-02 15:04:05"),
		suite.SuiteName,
		suite.PassedTests,
		suite.TotalTests,
		suite.AccuracyRate*100,
		suite.AverageTime)

	if avf.logFile != nil {
		avf.logFile.WriteString(logEntry)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}