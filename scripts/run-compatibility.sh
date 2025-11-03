#!/bin/bash

# Compatibility Verification Runner Script
# Runs comprehensive compatibility tests between current and new algorithms

set -e

echo "🔒 Running compatibility verification harness..."

# Ensure compatibility results directory exists
mkdir -p tests/compatibility/results

# Run the compatibility test runner
go run tests/compatibility_runner.go

echo ""
echo "📊 Compatibility test results summary:"
ls -la tests/compatibility/results/

echo ""
echo "✅ Compatibility verification completed successfully"