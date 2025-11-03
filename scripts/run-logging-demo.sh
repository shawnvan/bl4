#!/bin/bash

# Comprehensive Logging Demo Runner Script
# Demonstrates the complete logging system for algorithm comparison

set -e

echo "📝 Running comprehensive logging demonstration..."

# Ensure logging directory exists
mkdir -p tests/logging/logs

# Run the logging integration test
go run tests/logging_integration_test.go

echo ""
echo "📊 Generated log files:"
ls -la tests/logging/logs/

echo ""
echo "📈 Log file contents summary:"
echo "- Main log: tests/logging/logs/comparison.log"
echo "- Debug log: tests/logging/logs/debug.log"
echo "- Daily logs: tests/logging/logs/daily_YYYY-MM-DD.log"
echo "- Log reports: tests/logging/logs/log_report_*.json"

echo ""
echo "🔍 Example log content (last 5 lines):"
echo "======================================"
tail -5 tests/logging/logs/comparison.log 2>/dev/null || echo "No log entries found"

echo ""
echo "✅ Comprehensive logging demonstration completed successfully"