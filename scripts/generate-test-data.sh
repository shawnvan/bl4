#!/bin/bash

# Test Data Generation Script
# Generates comprehensive test data sets for BL4 codec testing

set -e

echo "🔧 Generating comprehensive test data sets..."

# Ensure test data directory exists
mkdir -p tests/data

# Run the test data generator
go run tests/data/test_data_generator.go

echo "✅ Test data generation completed"

# Verify files were created
echo "📋 Generated test data files:"
ls -la tests/data/

# Show statistics
echo ""
echo "📊 Test data statistics:"
echo "- Basic accuracy set: $(wc -l < tests/data/basic_accuracy_500_items.txt) lines"
echo "- Performance baseline: $(wc -l < tests/data/performance_baseline_1000_items.txt) lines"
echo "- Edge cases: $(wc -l < tests/data/edge_cases.txt) lines"
echo "- Manufacturers: $(wc -l < tests/data/manufacturers.txt) lines"
echo "- Rarities: $(wc -l < tests/data/rarities.txt) lines"
echo "- Expected values JSON: $(wc -c < tests/data/expected_decoded_values.json) bytes"