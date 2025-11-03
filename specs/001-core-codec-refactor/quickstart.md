# Quick Start Guide: Core Codec Algorithm Refactor

**Feature**: 001-core-codec-refactor
**Date**: 2025-11-03
**Branch**: `001-core-codec-refactor`

## Overview

This guide provides quick instructions for developers working on the core codec algorithm refactor. The refactor improves accuracy and performance by adopting reference implementation algorithms while maintaining complete API/UI compatibility.

## Prerequisites

### Development Environment
- Go 1.21+ installed
- Git repository access
- Reference project: https://github.com/Nicnl/borderlands4-serials
- Basic understanding of Base85 encoding and bit manipulation

### Required Tools
```bash
# Development tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/air-verse/air@latest  # Hot reload for development

# Testing tools
go install github.com/onsi/ginkgo/v2/ginkgo@latest
```

## Quick Setup

### 1. Clone and Setup
```bash
# Clone the repository
git clone <repository-url>
cd bl4

# Switch to the refactor branch
git checkout 001-core-codec-refactor

# Install dependencies
go mod download
go mod tidy
```

### 2. Build the Project
```bash
# Build all components
make build

# Or build individual components
go build -o bin/api ./cmd/api
go build -o bin/cli ./cmd/cli
go build -o bin/tui ./cmd/tui
```

### 3. Run Tests
```bash
# Run all tests
make test

# Run codec-specific tests
go test ./internal/codec/...

# Run tests with coverage
go test -cover ./internal/codec/...
```

## Development Workflow

### Phase 1: VARINT Decoder Replacement

#### 1.1 Study Reference Implementation
```bash
# Clone reference project for study
git clone https://github.com/Nicnl/borderlands4-serials reference/
cd reference/

# Study the VARINT implementation
# Look for files related to variable-length integer decoding
```

#### 1.2 Run Existing Tests as Baseline
```bash
# Run current VARINT tests to establish baseline
go test -v ./internal/codec/datatypes/ -run TestVARINT

# Benchmark current performance
go test -bench=BenchmarkVARINT ./internal/codec/datatypes/
```

#### 1.3 Implement New VARINT Decoder
```bash
# Create new implementation file
cp internal/codec/datatypes/varint.go internal/codec/datatypes/varint_new.go

# Implement reference algorithm while maintaining interface compatibility
# Key requirements:
# - Keep existing VARINT struct interface
# - Implement 4-bit nibble processing
# - Add proper bit mirroring
# - Maintain error handling
```

#### 1.4 Validate Implementation
```bash
# Run tests with new implementation
go test -v ./internal/codec/datatypes/ -run TestVARINTNew

# Compare results with original implementation
go test -v ./internal/codec/datatypes/ -run TestCompareVARINT

# Performance benchmarking
go test -bench=BenchmarkVARINTCompare ./internal/codec/datatypes/
```

### Phase 2: Base85 Codec Enhancement

#### 2.1 Analyze Current Implementation
```bash
# Study current Base85 implementation
cat internal/codec/base85/decode.go
cat internal/codec/base85/encode.go

# Run baseline tests
go test -v ./internal/codec/base85/
```

#### 2.2 Adopt Reference Optimizations
```bash
# Enhance existing implementation
# Focus on:
# - Character set optimizations
# - Lookup table improvements
# - Memory allocation optimizations
# - Error handling enhancements
```

#### 2.3 Validate Enhancements
```bash
# Ensure all existing tests pass
go test ./internal/codec/base85/...

# Performance testing
go test -bench=BenchmarkBase85 ./internal/codec/base85/

# Compatibility testing with real item codes
./bin/cli decode "@Ugy3L+2}TYgAAAABkAAAAAA"
```

### Phase 3: Tokenization and Serialization Refinement

#### 3.1 Tokenization Enhancement
```bash
# Study current tokenization
cat internal/codec/token/tokenizer.go

# Enhance accuracy based on reference implementation
# Maintain all existing token types and interfaces
```

#### 3.2 Serialization Engine Refinement
```bash
# Study current serialization
cat internal/codec/serial/deserialize.go
cat internal/codec/serial/serialize.go

# Apply reference improvements while maintaining compatibility
```

## Testing Strategy

### Unit Tests
```bash
# Run comprehensive unit tests
go test ./internal/codec/...

# Test specific components
go test ./internal/codec/base85/
go test ./internal/codec/datatypes/
go test ./internal/codec/serial/
go test ./internal/codec/token/
```

### Integration Tests
```bash
# Test API integration
go test ./internal/api/...

# Test CLI integration
go test ./internal/cli/...

# End-to-end tests
go test ./tests/integration/...
```

### Compatibility Tests
```bash
# Test against known item codes
./bin/cli decode "@Ugy3L+2}TYgAAAABkAAAAAA"

# Batch processing tests
./bin/cli batch-decode test_data/items.txt

# API compatibility tests
curl -X POST http://localhost:8080/api/v1/items/decode \
  -H "Content-Type: application/json" \
  -d '{"serial_code":"@Ugy3L+2}TYgAAAABkAAAAAA"}'
```

## Key Implementation Requirements

### Interface Compatibility
- All existing public interfaces must remain unchanged
- API endpoints must maintain exact same request/response formats
- CLI commands must preserve existing behavior
- Error handling must remain consistent

### Performance Requirements
- Individual item processing: ≤ current time * 0.9 (10% improvement)
- Memory usage: ≤ current usage * 0.85 (15% improvement)
- Batch processing: measurable improvements over current implementation

### Accuracy Requirements
- Maintain ≥ 99.5% accuracy with known item codes
- Better handling of edge cases and malformed inputs
- Improved error messages and debugging information

## Common Development Tasks

### Adding New Test Cases
```bash
# Create test file
touch internal/codec/datatypes/varint_test_new.go

# Add test cases based on reference implementation
func TestVARINTReferenceCases(t *testing.T) {
    testCases := []struct {
        name     string
        input    []byte
        expected VARINT
    }{
        // Add test cases from reference implementation
    }
}
```

### Performance Profiling
```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=. ./internal/codec/...

# Memory profiling
go test -memprofile=mem.prof -bench=. ./internal/codec/...

# Analyze profiles
go tool pprof cpu.prof
go tool pprof mem.prof
```

### Debugging Algorithm Issues
```bash
# Enable debug logging
export BL4_DEBUG=true

# Run with verbose output
./bin/cli decode --debug "@Ugy3L+2}TYgAAAABkAAAAAA"

# Step-by-step debugging
go test -v -run TestSpecificCase ./internal/codec/datatypes/
```

## Validation Checklist

Before submitting changes, ensure:

- [ ] All existing tests pass without modification
- [ ] New implementation produces identical results for known inputs
- [ ] Performance meets or exceeds requirements
- [ ] API compatibility is maintained
- [ ] CLI behavior is unchanged
- [ ] Error handling is consistent
- [ ] Documentation is updated
- [ ] Code follows project style guidelines

## Getting Help

### Reference Materials
- [Feature Specification](./spec.md)
- [Research Report](./research.md)
- [Data Model](./data-model.md)
- [API Contracts](./contracts/api.yaml)

### Common Issues
1. **Interface Compatibility**: Always check that public interfaces remain unchanged
2. **Test Failures**: Ensure new implementations produce identical results
3. **Performance Regressions**: Benchmark against current implementation
4. **Memory Issues**: Profile memory usage during batch operations

### Debug Commands
```bash
# Enable detailed logging
export BL4_LOG_LEVEL=debug

# Run with race detection
go run -race ./cmd/api

# Memory leak detection
go test -memprofile=mem.prof ./internal/codec/...
```

## Next Steps

After completing the refactor:

1. Run comprehensive integration tests
2. Performance validation and benchmarking
3. Documentation updates
4. Code review and quality assurance
5. Deployment and monitoring setup

## Troubleshooting

### Common Test Failures
```bash
# If tests fail due to interface changes
go test ./internal/codec/... -v

# Check for missing methods or signature changes
go build ./internal/codec/...
```

### Performance Issues
```bash
# If performance degrades
go test -bench=BenchmarkVARINT ./internal/codec/datatypes/
go test -bench=BenchmarkBase85 ./internal/codec/base85/
```

### Integration Issues
```bash
# If API tests fail
curl -X POST http://localhost:8080/api/v1/health

# Check if server is running
ps aux | grep "bl4-api"
```

This quick start guide provides the essential information for developers working on the core codec algorithm refactor. For detailed specifications and requirements, refer to the linked documentation files.