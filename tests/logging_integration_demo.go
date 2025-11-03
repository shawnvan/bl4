package main

import (
	"fmt"
	"log"
	"time"

	"github.com/shawnvan/bl4/tests/logging"
	"github.com/shawnvan/bl4/tests/validation"
	"github.com/shawnvan/bl4/tests/compatibility"
)

// Mock algorithm implementations for testing logging
func currentAlgorithmDecode(serial string) (string, error) {
	// Simulate some processing time
	time.Sleep(time.Microsecond * time.Duration(50+serial[0]%20))
	return fmt.Sprintf("decoded_%s", serial), nil
}

func newAlgorithmDecode(serial string) (string, error) {
	// Simulate faster processing time
	time.Sleep(time.Microsecond * time.Duration(30+serial[0]%15))
	return fmt.Sprintf("decoded_%s", serial), nil
}

// LoggingIntegrationTestRunner demonstrates comprehensive logging for algorithm comparison
type LoggingIntegrationTestRunner struct {
	logger *logging.ComparisonLogger
}

// NewLoggingIntegrationTestRunner creates a new logging integration test runner
func NewLoggingIntegrationTestRunner() (*LoggingIntegrationTestRunner, error) {
	logger, err := logging.NewComparisonLogger()
	if err != nil {
		return nil, err
	}

	return &LoggingIntegrationTestRunner{
		logger: logger,
	}, nil
}

// Close closes the logging integration test runner
func (litr *LoggingIntegrationTestRunner) Close() error {
	return litr.logger.Close()
}

// RunLoggingDemo runs a comprehensive logging demonstration
func (litr *LoggingIntegrationTestRunner) RunLoggingDemo() error {
	fmt.Println("📝 Running comprehensive logging demonstration...")
	fmt.Println("=================================================")

	// Initialize daily log
	if err := litr.logger.CreateDailyLog(); err != nil {
		return fmt.Errorf("failed to create daily log: %w", err)
	}

	// Demonstrate different logging scenarios

	// 1. Test suite logging
	litr.logger.LogTestSuiteStarted(
		"Algorithm Comparison Demo",
		"Demonstrating comprehensive logging capabilities",
	)

	// 2. Individual algorithm execution logging
	testSerialCodes := []string{
		"@Ugy3L+2}TYgAAAABkAAAAAA",
		"@Ugy3L+2}TYgAAAABkAAAAAB",
		"@Ugy3L+2}TYgAAAABkAAAAAC",
		"@Ugy3L+2}TYgAAAABkAAAAAD",
		"@Ugy3L+2}TYgAAAABkAAAAAE",
	}

	fmt.Println("\n🔍 Testing individual algorithm execution logging...")
	for i, serialCode := range testSerialCodes {
		// Log current algorithm execution
		start := time.Now()
		currentOutput, err := currentAlgorithmDecode(serialCode)
		currentTime := time.Since(start)

		litr.logger.LogAlgorithmExecution(
			"current_algorithm",
			serialCode,
			serialCode,
			currentOutput,
			err,
			currentTime,
			map[string]interface{}{
				"test_index":     i,
				"algorithm_type": "current",
				"test_scenario":  "logging_demo",
			},
		)

		// Log new algorithm execution
		start = time.Now()
		newOutput, err := newAlgorithmDecode(serialCode)
		newTime := time.Since(start)

		litr.logger.LogAlgorithmExecution(
			"new_algorithm",
			serialCode,
			serialCode,
			newOutput,
			err,
			newTime,
			map[string]interface{}{
				"test_index":     i,
				"algorithm_type": "new",
				"test_scenario":  "logging_demo",
			},
		)

		// Log comparison result
		matches := (currentOutput == newOutput)
		improvement := ((float64(currentTime-newTime) / float64(currentTime)) * 100)

		litr.logger.LogComparisonResult(
			serialCode,
			currentOutput,
			newOutput,
			matches,
			currentTime,
			newTime,
			improvement,
		)

		fmt.Printf("  Logged comparison for: %s (%.2f%% improvement)\n", serialCode[:20], improvement)
	}

	// 3. Performance summary logging
	fmt.Println("\n⚡ Testing performance summary logging...")
	litr.logger.LogPerformanceSummary(
		"demo_performance_test",
		len(testSerialCodes),
		time.Microsecond*60, // average current time
		time.Microsecond*40, // average new time
		33.33,               // average improvement
		25.0,                // min improvement
		45.0,                // max improvement
	)

	// 4. Compatibility summary logging
	fmt.Println("\n🔒 Testing compatibility summary logging...")
	litr.logger.LogCompatibilitySummary(
		"demo_compatibility_test",
		len(testSerialCodes),
		len(testSerialCodes), // all compatible
		0,                    // no incompatible
		1.0,                  // 100% compatible
	)

	// 5. Batch processing logging
	fmt.Println("\n📦 Testing batch processing logging...")
	batchStart := time.Now()
	for i := 0; i < 10; i++ {
		_, _ = currentAlgorithmDecode(fmt.Sprintf("@Ugy3L+2}TYgAAAABkAAAA%c", byte('A'+i%5)))
	}
	batchDuration := time.Since(batchStart)

	litr.logger.LogBatch(
		"demo_batch_processing",
		10,
		10,
		0,
		batchDuration,
	)

	// 6. Error scenario logging
	fmt.Println("\n❌ Testing error scenario logging...")
	litr.logger.LogError(
		"Demo error scenario",
		map[string]interface{}{
			"error_type":    "simulation",
			"error_code":    "DEMO_001",
			"serial_code":   "@Ugy3L+2}TYgAAAABkAAAAXX",
			"failed_step":   "algorithm_execution",
			"recovery_action": "retry_later",
		},
	)

	// 7. Warning scenario logging
	fmt.Println("\n⚠️ Testing warning scenario logging...")
	litr.logger.LogWarning(
		"Demo warning scenario",
		map[string]interface{}{
			"warning_type":   "performance_degradation",
			"threshold":      "10_percent_improvement",
			"actual_value":   8.5,
			"recommendation": "optimize_algorithm",
		},
	)

	// 8. Debug scenario logging
	fmt.Println("\n🐛 Testing debug scenario logging...")
	litr.logger.LogDebug(
		"Demo debug scenario",
		map[string]interface{}{
			"debug_step":     "internal_state_check",
			"memory_usage":   "1024_bytes",
			"cache_hit_rate": 0.85,
			"optimization_applied": true,
		},
	)

	// 9. Test suite completion
	litr.logger.LogTestSuiteCompleted(
		"Algorithm Comparison Demo",
		50, // total simulated tests
		48, // successful tests
		2,  // failed tests
		time.Second*5, // total duration
	)

	// 10. Generate comprehensive log report
	fmt.Println("\n📊 Generating comprehensive log report...")
	if err := litr.logger.GenerateLogReport(); err != nil {
		return fmt.Errorf("failed to generate log report: %w", err)
	}

	fmt.Println("\n✅ Comprehensive logging demonstration completed!")
	fmt.Println("Check tests/logging/logs/ for detailed log files.")

	return nil
}

// RunIntegrationWithValidation demonstrates logging integration with validation framework
func (litr *LoggingIntegrationTestRunner) RunIntegrationWithValidation() error {
	fmt.Println("\n🔗 Testing integration with validation framework...")

	// Create validation framework
	validationFramework, err := validation.NewAlgorithmValidationFramework()
	if err != nil {
		return fmt.Errorf("failed to create validation framework: %w", err)
	}
	defer validationFramework.Close()

	// Log integration test start
	litr.logger.LogTestSuiteStarted(
		"Validation Framework Integration",
		"Testing logging integration with validation framework",
	)

	// Run a small validation test with logging
	suite, err := validationFramework.RunAccuracyValidation(
		"tests/data/edge_cases.txt",
		func(serial string) (string, error) {
			start := time.Now()
			result, err := currentAlgorithmDecode(serial)

			// Log each validation step
			litr.logger.LogAlgorithmExecution(
				"validation_test",
				serial,
				serial,
				result,
				err,
				time.Since(start),
				map[string]interface{}{
					"integration_test": true,
					"framework":        "validation",
				},
			)

			return result, err
		},
	)

	if err != nil {
		litr.logger.LogError("Validation framework integration failed", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Log validation results
	litr.logger.LogCompatibilitySummary(
		"validation_integration",
		suite.TotalTests,
		suite.PassedTests,
		suite.FailedTests,
		suite.AccuracyRate,
	)

	litr.logger.LogTestSuiteCompleted(
		"Validation Framework Integration",
		suite.TotalTests,
		suite.PassedTests,
		suite.FailedTests,
		suite.EndTime.Sub(suite.StartTime),
	)

	return nil
}

// RunIntegrationWithCompatibility demonstrates logging integration with compatibility framework
func (litr *LoggingIntegrationTestRunner) RunIntegrationWithCompatibility() error {
	fmt.Println("\n🔒 Testing integration with compatibility framework...")

	// Create compatibility framework
	compatibilityFramework, err := compatibility.NewCompatibilityHarness()
	if err != nil {
		return fmt.Errorf("failed to create compatibility framework: %w", err)
	}
	defer compatibilityFramework.Close()

	// Log integration test start
	litr.logger.LogTestSuiteStarted(
		"Compatibility Framework Integration",
		"Testing logging integration with compatibility framework",
	)

	// Run a small compatibility test with logging
	suite, err := compatibilityFramework.RunCompatibilityTest(
		"tests/data/edge_cases.txt",
		func(serial string) (string, error) {
			start := time.Now()
			result, err := currentAlgorithmDecode(serial)

			// Log each compatibility test step
			litr.logger.LogAlgorithmExecution(
				"compatibility_test_current",
				serial,
				serial,
				result,
				err,
				time.Since(start),
				map[string]interface{}{
					"integration_test": true,
					"framework":        "compatibility",
					"algorithm":        "current",
				},
			)

			return result, err
		},
		func(serial string) (string, error) {
			start := time.Now()
			result, err := newAlgorithmDecode(serial)

			// Log each compatibility test step
			litr.logger.LogAlgorithmExecution(
				"compatibility_test_new",
				serial,
				serial,
				result,
				err,
				time.Since(start),
				map[string]interface{}{
					"integration_test": true,
					"framework":        "compatibility",
					"algorithm":        "new",
				},
			)

			return result, err
		},
	)

	if err != nil {
		litr.logger.LogError("Compatibility framework integration failed", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Log compatibility results
	litr.logger.LogCompatibilitySummary(
		"compatibility_integration",
		suite.TotalTests,
		suite.CompatibleTests,
		suite.IncompatibleTests,
		suite.CompatibilityRate,
	)

	litr.logger.LogTestSuiteCompleted(
		"Compatibility Framework Integration",
		suite.TotalTests,
		suite.CompatibleTests,
		suite.IncompatibleTests,
		suite.EndTime.Sub(suite.StartTime),
	)

	return nil
}

func main() {
	runner, err := NewLoggingIntegrationTestRunner()
	if err != nil {
		log.Fatalf("Failed to create logging integration test runner: %v", err)
	}
	defer runner.Close()

	// Run basic logging demo
	if err := runner.RunLoggingDemo(); err != nil {
		log.Fatalf("Logging demo failed: %v", err)
	}

	// Run integration tests
	if err := runner.RunIntegrationWithValidation(); err != nil {
		log.Fatalf("Validation integration test failed: %v", err)
	}

	if err := runner.RunIntegrationWithCompatibility(); err != nil {
		log.Fatalf("Compatibility integration test failed: %v", err)
	}

	fmt.Println("\n🎉 All logging integration tests completed successfully!")
	fmt.Println("📁 Check tests/logging/logs/ for comprehensive log files.")
}