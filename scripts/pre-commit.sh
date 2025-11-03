#!/bin/bash

# Pre-commit hook for BL4 project
# Ensures code quality and test compliance before commits

set -e

echo "🔍 Running pre-commit validation checks..."

# Check if we're on the main branch (should not commit directly to main)
CURRENT_BRANCH=$(git branch --show-current)
if [[ "$CURRENT_BRANCH" == "main" ]]; then
    echo "❌ Cannot commit directly to main branch. Create a feature branch first."
    exit 1
fi

# Run Go formatting check
echo "📝 Checking Go formatting..."
if ! gofmt -l . | grep -q .; then
    echo "❌ Go formatting issues found. Run 'go fmt ./...' to fix."
    exit 1
else
    echo "✅ Go formatting OK"
fi

# Run go vet
echo "🔍 Running go vet..."
if ! go vet ./...; then
    echo "❌ go vet issues found. Fix before committing."
    exit 1
else
    echo "✅ go vet OK"
fi

# Run go mod tidy check
echo "📦 Checking go mod..."
if ! go mod tidy -diff &>/dev/null; then
    echo "❌ go mod is not tidy. Run 'go mod tidy' before committing."
    exit 1
else
    echo "✅ go mod tidy OK"
fi

# Run tests (skip in development mode if requested)
if [[ "$1" != "--skip-tests" ]]; then
    echo "🧪 Running tests..."
    if ! go test ./...; then
        echo "❌ Tests failed. Fix before committing."
        exit 1
    else
        echo "✅ All tests passed"
    fi
else
    echo "⏭ Skipping tests (development mode)"
fi

# Check for TODO/FIXME comments (warning only)
echo "📋 Checking for TODO comments..."
TODO_COUNT=$(grep -r "TODO\|FIXME" --include="*.go" . | wc -l || true)
if [[ $TODO_COUNT -gt 0 ]]; then
    echo "⚠️  Found $TODO_COUNT TODO/FIXME comments. Consider addressing them."
fi

# Check for large files (warning only)
echo "📏 Checking for large files..."
LARGE_FILES=$(find . -name "*.go" -size +100k -exec ls -lh {} \; | wc -l || true)
if [[ $LARGE_FILES -gt 0 ]]; then
    echo "⚠️  Found $LARGE_FILES large Go files. Consider refactoring."
fi

# Check codec-specific requirements
echo "🔧 Checking codec-specific requirements..."

# Check if all codec test files have corresponding implementations
CODEC_DIRS=("internal/codec/base85" "internal/codec/datatypes" "internal/codec/serial" "internal/codec/token")
for dir in "${CODEC_DIRS[@]}"; do
    if [[ -d "$dir" ]]; then
        test_files=$(find "$dir" -name "*_test.go" | wc -l)
        impl_files=$(find "$dir" -name "*.go" ! -name "*_test.go" | wc -l)

        if [[ $test_files -gt 0 && $impl_files -eq 0 ]]; then
            echo "⚠️  Found tests but no implementations in $dir"
        elif [[ $impl_files -gt 0 && $test_files -eq 0 ]]; then
            echo "⚠️  Found implementations but no tests in $dir"
        else
            echo "✅ $dir: $impl_files implementations, $test_files tests"
        fi
    fi
done

# Check TDD compliance (tests should exist before implementations)
echo "🧪 Checking TDD compliance..."
TEST_DIRS=("tests/unit" "tests/integration" "tests/contract" "tests/benchmark" "tests/comparison")
for dir in "${TEST_DIRS[@]}"; do
    if [[ -d "$dir" ]]; then
        echo "✅ Test directory exists: $dir"
    else
        echo "⚠️  Test directory missing: $dir"
    fi
done

# Check documentation requirements
echo "📚 Checking documentation..."
if [[ ! -f "README.md" ]]; then
    echo "❌ README.md missing"
    exit 1
else
    echo "✅ README.md exists"
fi

# Check for codec documentation
DOCS=(
    "specs/001-core-codec-refactor/spec.md"
    "specs/001-core-codec-refactor/plan.md"
    "specs/001-core-codec-refactor/tasks.md"
)

for doc in "${DOCS[@]}"; do
    if [[ -f "$doc" ]]; then
        echo "✅ Documentation exists: $doc"
    else
        echo "❌ Documentation missing: $doc"
        exit 1
    fi
done

echo "✅ All pre-commit checks passed!"
echo "🎉 Ready to commit changes."