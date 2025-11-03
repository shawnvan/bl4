package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ComparisonLogger provides comprehensive logging for algorithm comparison
type ComparisonLogger struct {
	logDir    string
	logFile   *os.File
	debugFile *os.File
	mutex     sync.Mutex
}

// NewComparisonLogger creates a new comparison logger instance
func NewComparisonLogger() (*ComparisonLogger, error) {
	logDir := "tests/logging/logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create logging directory: %w", err)
	}

	// Main log file
	logFile, err := os.OpenFile(filepath.Join(logDir, "comparison.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create main log file: %w", err)
	}

	// Debug log file for detailed debugging
	debugFile, err := os.OpenFile(filepath.Join(logDir, "debug.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logFile.Close()
		return nil, fmt.Errorf("failed to create debug log file: %w", err)
	}

	logger := &ComparisonLogger{
		logDir:    logDir,
		logFile:   logFile,
		debugFile: debugFile,
	}

	// Log initialization
	logger.LogInfo("Comparison logger initialized", map[string]interface{}{
		"log_dir":     logDir,
		"initialized": time.Now(),
	})

	return logger, nil
}

// Close closes the comparison logger
func (cl *ComparisonLogger) Close() error {
	cl.mutex.Lock()
	defer cl.mutex.Unlock()

	// Write shutdown message directly without calling LogInfo to avoid deadlock
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     LogLevelInfo.String(),
		Message:   "Comparison logger shutting down",
		Data:      nil,
	}

	// Serialize to JSON and write to log file
	jsonData, err := json.Marshal(entry)
	if err == nil && cl.logFile != nil {
		cl.logFile.WriteString(string(jsonData) + "\n")
	}

	var errors []error

	if cl.logFile != nil {
		if err := cl.logFile.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close main log file: %w", err))
		}
	}

	if cl.debugFile != nil {
		if err := cl.debugFile.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close debug log file: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("multiple errors occurred: %v", errors)
	}

	return nil
}

// LogLevel represents different log levels
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarning
	LogLevelError
)

func (ll LogLevel) String() string {
	switch ll {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarning:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// LogInfo logs an info message
func (cl *ComparisonLogger) LogInfo(message string, data map[string]interface{}) {
	cl.log(LogLevelInfo, message, data)
}

// LogWarning logs a warning message
func (cl *ComparisonLogger) LogWarning(message string, data map[string]interface{}) {
	cl.log(LogLevelWarning, message, data)
}

// LogError logs an error message
func (cl *ComparisonLogger) LogError(message string, data map[string]interface{}) {
	cl.log(LogLevelError, message, data)
}

// LogDebug logs a debug message
func (cl *ComparisonLogger) LogDebug(message string, data map[string]interface{}) {
	cl.log(LogLevelDebug, message, data)
}

// AlgorithmExecutionLog logs detailed algorithm execution information
func (cl *ComparisonLogger) LogAlgorithmExecution(
	algorithmName string,
	serialCode string,
	input string,
	output string,
	error error,
	processingTime time.Duration,
	metadata map[string]interface{},
) {
	data := map[string]interface{}{
		"algorithm_name":  algorithmName,
		"serial_code":     serialCode,
		"input":           input,
		"output":          output,
		"error":           nil,
		"processing_time": processingTime.Nanoseconds(),
		"metadata":        metadata,
	}

	if error != nil {
		data["error"] = error.Error()
		cl.log(LogLevelError, fmt.Sprintf("Algorithm execution failed: %s", algorithmName), data)
	} else {
		cl.log(LogLevelDebug, fmt.Sprintf("Algorithm execution: %s", algorithmName), data)
	}
}

// ComparisonResult logs algorithm comparison results
func (cl *ComparisonLogger) LogComparisonResult(
	serialCode string,
	currentOutput string,
	newOutput string,
	matches bool,
	currentTime time.Duration,
	newTime time.Duration,
	improvement float64,
) {
	data := map[string]interface{}{
		"serial_code":      serialCode,
		"current_output":   currentOutput,
		"new_output":       newOutput,
		"matches":          matches,
		"current_time":     currentTime.Nanoseconds(),
		"new_time":         newTime.Nanoseconds(),
		"improvement":      improvement,
		"comparison_time":  time.Now(),
	}

	level := LogLevelInfo
	if !matches {
		level = LogLevelError
	}

	message := fmt.Sprintf("Comparison result: %s", serialCode)
	if matches {
		message += " (COMPATIBLE)"
	} else {
		message += " (INCOMPATIBLE)"
	}

	cl.log(level, message, data)
}

// PerformanceSummary logs performance comparison summary
func (cl *ComparisonLogger) LogPerformanceSummary(
	testName string,
	totalTests int,
	averageCurrentTime time.Duration,
	averageNewTime time.Duration,
	averageImprovement float64,
	minImprovement float64,
	maxImprovement float64,
) {
	data := map[string]interface{}{
		"test_name":              testName,
		"total_tests":            totalTests,
		"average_current_time":   averageCurrentTime.Nanoseconds(),
		"average_new_time":       averageNewTime.Nanoseconds(),
		"average_improvement":    averageImprovement,
		"min_improvement":        minImprovement,
		"max_improvement":        maxImprovement,
		"performance_target_met": averageImprovement >= 10.0,
		"logged_at":              time.Now(),
	}

	level := LogLevelInfo
	if averageImprovement < 10.0 {
		level = LogLevelWarning
	}

	message := fmt.Sprintf("Performance summary: %s (%.2f%% improvement)", testName, averageImprovement)
	cl.log(level, message, data)
}

// CompatibilitySummary logs compatibility test summary
func (cl *ComparisonLogger) LogCompatibilitySummary(
	testName string,
	totalTests int,
	compatibleTests int,
	incompatibleTests int,
	compatibilityRate float64,
) {
	data := map[string]interface{}{
		"test_name":           testName,
		"total_tests":         totalTests,
		"compatible_tests":    compatibleTests,
		"incompatible_tests":  incompatibleTests,
		"compatibility_rate":  compatibilityRate,
		"requirement_met":     compatibilityRate >= 1.0,
		"logged_at":           time.Now(),
	}

	level := LogLevelInfo
	if compatibilityRate < 1.0 {
		level = LogLevelError
	}

	message := fmt.Sprintf("Compatibility summary: %s (%.2f%% compatible)", testName, compatibilityRate*100)
	cl.log(level, message, data)
}

// TestSuiteStarted logs when a test suite starts
func (cl *ComparisonLogger) LogTestSuiteStarted(suiteName string, description string) {
	data := map[string]interface{}{
		"suite_name":  suiteName,
		"description": description,
		"started_at":  time.Now(),
	}

	cl.log(LogLevelInfo, fmt.Sprintf("Test suite started: %s", suiteName), data)
}

// TestSuiteCompleted logs when a test suite completes
func (cl *ComparisonLogger) LogTestSuiteCompleted(
	suiteName string,
	totalTests int,
	successfulTests int,
	failedTests int,
	duration time.Duration,
) {
	data := map[string]interface{}{
		"suite_name":        suiteName,
		"total_tests":       totalTests,
		"successful_tests":  successfulTests,
		"failed_tests":      failedTests,
		"success_rate":      float64(successfulTests) / float64(totalTests),
		"duration":          duration.Nanoseconds(),
		"completed_at":      time.Now(),
	}

	level := LogLevelInfo
	if failedTests > 0 {
		level = LogLevelWarning
	}

	message := fmt.Sprintf("Test suite completed: %s (%d/%d passed)", suiteName, successfulTests, totalTests)
	cl.log(level, message, data)
}

// LogBatch logs batch processing information
func (cl *ComparisonLogger) LogBatch(
	operation string,
	batchSize int,
	processedCount int,
	errorCount int,
	duration time.Duration,
) {
	data := map[string]interface{}{
		"operation":        operation,
		"batch_size":       batchSize,
		"processed_count":  processedCount,
		"error_count":      errorCount,
		"success_rate":     float64(processedCount-errorCount) / float64(processedCount),
		"duration":         duration.Nanoseconds(),
		"throughput":       float64(processedCount) / duration.Seconds(),
		"logged_at":        time.Now(),
	}

	level := LogLevelInfo
	if errorCount > 0 {
		level = LogLevelWarning
	}

	message := fmt.Sprintf("Batch %s: %d/%d processed (%.2f errors/s)",
		operation, processedCount, batchSize, float64(errorCount)/duration.Seconds())
	cl.log(level, message, data)
}

// log is the internal logging method
func (cl *ComparisonLogger) log(level LogLevel, message string, data map[string]interface{}) {
	cl.mutex.Lock()
	defer cl.mutex.Unlock()

	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level.String(),
		Message:   message,
		Data:      data,
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(entry)
	if err != nil {
		// Fallback to simple format if JSON marshaling fails
		fallback := fmt.Sprintf("[%s] %s %s: %v\n",
			entry.Timestamp.Format("2006-01-02 15:04:05.000"),
			entry.Level,
			entry.Message,
			entry.Data)
		cl.logFile.WriteString(fallback)
		return
	}

	// Write to main log file
	cl.logFile.WriteString(string(jsonData) + "\n")

	// Write debug messages to debug file as well
	if level == LogLevelDebug {
		cl.debugFile.WriteString(string(jsonData) + "\n")
	}
}

// CreateDailyLog creates a daily log file for important events
func (cl *ComparisonLogger) CreateDailyLog() error {
	today := time.Now().Format("2006-01-02")
	dailyLogPath := filepath.Join(cl.logDir, fmt.Sprintf("daily_%s.log", today))

	// Check if daily log already exists
	if _, err := os.Stat(dailyLogPath); err == nil {
		return nil // Already exists
	}

	// Create daily log with header
	dailyFile, err := os.Create(dailyLogPath)
	if err != nil {
		return fmt.Errorf("failed to create daily log file: %w", err)
	}
	defer dailyFile.Close()

	header := fmt.Sprintf("# Daily Algorithm Comparison Log\n# Date: %s\n# Generated: %s\n\n",
		today, time.Now().Format("2006-01-02 15:04:05"))

	_, err = dailyFile.WriteString(header)
	if err != nil {
		return fmt.Errorf("failed to write daily log header: %w", err)
	}

	cl.LogInfo("Daily log created", map[string]interface{}{
		"daily_log_path": dailyLogPath,
		"date":           today,
	})

	return nil
}

// GenerateLogReport generates a comprehensive log report
func (cl *ComparisonLogger) GenerateLogReport() error {
	reportPath := filepath.Join(cl.logDir, fmt.Sprintf("log_report_%s.json", time.Now().Format("20060102_150405")))

	report := map[string]interface{}{
		"generated_at":    time.Now(),
		"report_type":     "algorithm_comparison_log_summary",
		"log_directory":   cl.logDir,
		"log_files":       cl.getLogFiles(),
		"summary":         "Comprehensive logging report for algorithm comparison",
		"recommendations": []string{
			"Review error logs for failed test cases",
			"Monitor performance trends over time",
			"Check compatibility requirements before deployment",
		},
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal log report: %w", err)
	}

	err = os.WriteFile(reportPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write log report: %w", err)
	}

	cl.LogInfo("Log report generated", map[string]interface{}{
		"report_path": reportPath,
		"report_size": len(data),
	})

	return nil
}

func (cl *ComparisonLogger) getLogFiles() []string {
	var logFiles []string

	files, err := os.ReadDir(cl.logDir)
	if err != nil {
		return logFiles
	}

	for _, file := range files {
		if !file.IsDir() {
			logFiles = append(logFiles, file.Name())
		}
	}

	return logFiles
}