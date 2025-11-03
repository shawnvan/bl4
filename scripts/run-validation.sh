#!/bin/bash

# Algorithm Validation Runner Script
# Runs comprehensive validation tests for codec algorithms

set -e

echo "🚀 Running algorithm validation framework..."

# Ensure validation results directory exists
mkdir -p tests/validation/results

# Run the validation test runner
go run tests/validation_runner.go

echo ""
echo "📊 Validation results summary:"
ls -la tests/validation/results/

echo ""
echo "✅ Algorithm validation completed successfully"