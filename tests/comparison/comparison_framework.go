package comparison

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ComparisonFramework provides tools for comparing algorithm implementations
type ComparisonFramework struct {
	resultsDir string
	logFile    *os.File
}

// NewComparisonFramework creates a new comparison framework instance
func NewComparisonFramework() (*ComparisonFramework, error) {
	resultsDir := "tests/comparison/results"
	if err := os.MkdirAll(resultsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create results directory: %w", err)
	}

	logFile, err := os.OpenFile(filepath.Join(resultsDir, "comparison.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}

	return &ComparisonFramework{
		resultsDir: resultsDir,
		logFile:    logFile,
	}, nil
}

// Close closes the comparison framework
func (cf *ComparisonFramework) Close() error {
	if cf.logFile != nil {
		return cf.logFile.Close()
	}
	return nil
}

// LogComparison logs a comparison result
func (cf *ComparisonFramework) LogComparison(result *ComparisonResult) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	// Save individual result file
	filename := fmt.Sprintf("comparison_%s_%s.json",
		result.TestCase,
		time.Now().Format("20060102_150405"))
	filePath := filepath.Join(cf.resultsDir, filename)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return err
	}

	// Log to log file
	logEntry := fmt.Sprintf("[%s] %s: %v\n",
		time.Now().Format("2006-01-02 15:04:05"),
		result.TestCase,
		result.Matches)
	if _, err := cf.logFile.WriteString(logEntry); err != nil {
		return err
	}

	return nil
}

// LogPerformance logs a performance benchmark result
func (cf *ComparisonFramework) LogPerformance(result *PerformanceResult) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("performance_%s.json",
		time.Now().Format("20060102_150405"))
	filePath := filepath.Join(cf.resultsDir, filename)
	return os.WriteFile(filePath, data, 0644)
}

// GenerateReport generates a comprehensive comparison report
func (cf *ComparisonFramework) GenerateReport() (*ComparisonReport, error) {
	files, err := os.ReadDir(cf.resultsDir)
	if err != nil {
		return nil, err
	}

	var comparisonResults []*ComparisonResult
	var performanceResults []*PerformanceResult

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(cf.resultsDir, file.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		if file.Name()[0:10] == "comparison" {
			var result ComparisonResult
			if err := json.Unmarshal(data, &result); err == nil {
				comparisonResults = append(comparisonResults, &result)
			}
		} else if file.Name()[0:10] == "performance" {
			var result PerformanceResult
			if err := json.Unmarshal(data, &result); err == nil {
				performanceResults = append(performanceResults, &result)
			}
		}
	}

	report := &ComparisonReport{
		GeneratedAt:         time.Now(),
		ComparisonResults:   comparisonResults,
		PerformanceResults:  performanceResults,
		Summary:            cf.generateSummary(comparisonResults, performanceResults),
	}

	// Save report
	reportData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, err
	}

	reportPath := filepath.Join(cf.resultsDir, "comparison_report.json")
	if err := os.WriteFile(reportPath, reportData, 0644); err != nil {
		return nil, err
	}

	return report, nil
}

type ComparisonReport struct {
	GeneratedAt         time.Time            `json:"generated_at"`
	ComparisonResults   []*ComparisonResult `json:"comparison_results"`
	PerformanceResults  []*PerformanceResult `json:"performance_results"`
	Summary            ReportSummary         `json:"summary"`
}

type ReportSummary struct {
	TotalComparisons      int     `json:"total_comparisons"`
	MatchingResults       int     `json:"matching_results"`
	NonMatchingResults     int     `json:"non_matching_results"`
	MatchRate              float64 `json:"match_rate_percent"`
	TotalPerformanceTests   int     `json:"total_performance_tests"`
	AverageImprovement      float64 `json:"average_improvement_percent"`
	BestImprovement         float64 `json:"best_improvement_percent"`
	WorstImprovement        float64 `json:"worst_improvement_percent"`
}

func (cf *ComparisonFramework) generateSummary(comparisonResults []*ComparisonResult, performanceResults []*PerformanceResult) ReportSummary {
	summary := ReportSummary{
		TotalComparisons: len(comparisonResults),
	}

	// Calculate comparison summary
	for _, result := range comparisonResults {
		if result.Matches {
			summary.MatchingResults++
		} else {
			summary.NonMatchingResults++
		}
	}

	if summary.TotalComparisons > 0 {
		summary.MatchRate = float64(summary.MatchingResults) / float64(summary.TotalComparisons) * 100
	}

	// Calculate performance summary
	summary.TotalPerformanceTests = len(performanceResults)
	if len(performanceResults) > 0 {
		var totalImprovement float64
		for _, result := range performanceResults {
			if result.Improvement > 0 {
				totalImprovement += result.Improvement
			}
		}
		summary.AverageImprovement = totalImprovement / float64(len(performanceResults))
	}

	return summary
}

// ValidateTestCases validates that all test cases have expected results
func (cf *ComparisonFramework) ValidateTestCases() error {
	testCasesFile := "tests/data/expected_decoded_values.json"
	data, err := os.ReadFile(testCasesFile)
	if err != nil {
		return fmt.Errorf("failed to read test cases file: %w", err)
	}

	var testCases map[string]string
	if err := json.Unmarshal(data, &testCases); err != nil {
		return fmt.Errorf("failed to parse test cases: %w", err)
	}

	// Validate each test case
	for serialCode, expectedResult := range testCases {
		if serialCode == "" {
			return fmt.Errorf("empty serial code found in test cases")
		}
		if expectedResult == "" {
			return fmt.Errorf("empty expected result for serial code: %s", serialCode)
		}
	}

	cf.logFile.WriteString(fmt.Sprintf("[%s] Validated %d test cases\n",
		time.Now().Format("2006-01-02 15:04:05"), len(testCases)))

	return nil
}