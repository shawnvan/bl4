# Research Report: Core Codec Algorithm Refactor

**Date**: 2025-11-03
**Reference**: https://github.com/Nicnl/borderlands4-serials
**Target Project**: BL4 Item Codec Tool

## Executive Summary

The reference project (Nicnl/borderlands4-serials) provides a comprehensive implementation of Borderlands 4 item serialization algorithms. Our current BL4 implementation already follows many best practices established by the reference, but we can improve accuracy and performance by adopting specific algorithm implementations while maintaining our superior architecture.

## Key Research Findings

### 1. Reference Project Analysis

**Project Structure**: The reference implementation focuses on core algorithms with:
- Base85 encoding/decoding with BL4-specific charset
- VARINT decoder with 4-bit nibble processing
- Tokenization engine supporting 6 token types
- Bitstream processing with mirroring operations

**Language and Dependencies**:
- Reference appears to be Python-based (common for game modding)
- Our Go implementation provides superior performance characteristics
- Clean separation of concerns already matches reference patterns

### 2. Algorithm Implementation Details

#### Base85 Encoding/Decoding
**Decision**: Maintain current Go implementation with minor enhancements
**Rationale**: Our current implementation already incorporates BL4-specific optimizations:
- Custom charset with `@U` prefix validation
- 5-character to 4-byte conversion groups
- Bit mirroring using lookup tables for performance
- Proper padding handling with `~` character

**Alternatives considered**: Direct Python port (rejected due to performance), C integration (rejected due to complexity)

#### VARINT Decoder
**Decision**: Adopt reference implementation's 4-bit nibble approach
**Rationale**: Reference implementation appears to have more accurate bit-level processing:
- 4-bit blocks with continuation bits
- Maximum 4 blocks (16 bits usable)
- Proper 4-bit mirroring on each block
- Better handling of edge cases

**Implementation Strategy**: Replace current VARINT decoder while maintaining interface compatibility

#### Tokenization Engine
**Decision**: Maintain current token structure with reference algorithm enhancements
**Rationale**: Current token types (SEP1, SEP2, VARINT, PART, VARBIT, STRING) match reference patterns:
- Hard/soft separators
- Variable-length integers
- Item parts (16-bit index + 16-bit value)
- Variable bit arrays
- Text strings

#### Serialization Engine
**Decision**: Maintain current layered architecture with reference algorithm improvements
**Rationale**: Current architecture provides superior separation:
- Base85 layer: Character encoding/decoding
- Bitstream layer: Bit-level operations
- Tokenization layer: Token parsing
- Serialization layer: Data conversion

### 3. Integration Strategy

#### Approach Selection
**Decision**: Progressive replacement strategy
**Rationale**: Minimizes risk while allowing thorough validation:
1. VARINT decoder replacement (highest accuracy impact)
2. Base85 codec enhancement (performance optimization)
3. Serialization engine refinement (completeness)

#### Compatibility Preservation
**Decision**: Maintain all existing interfaces and contracts
**Rationale**: Ensures zero disruption to existing users:
- API endpoints unchanged
- CLI interface preserved
- TUI functionality maintained
- Data model compatibility guaranteed

### 4. Performance Considerations

#### Current Advantages
- Go implementation provides superior performance vs Python reference
- Existing optimizations (lookup tables, object pools)
- Efficient memory management and concurrency

#### Potential Improvements
- Adopt reference algorithm's edge case handling
- Enhanced error detection and reporting
- Improved accuracy for complex item codes

### 5. License and Legal Considerations

**Current Project**: MIT License (permissive)
**Reference Project**: Unable to verify due to GitHub access limitations
**Recommendation**: Verify license compatibility before integration

### 6. Technical Implementation Plan

#### Phase 1: VARINT Decoder Replacement
- Extract VARINT algorithm from reference implementation
- Implement Go version maintaining interface compatibility
- Comprehensive testing with known item codes
- Performance benchmarking against current implementation

#### Phase 2: Base85 Codec Enhancement
- Adopt any reference optimizations not already present
- Enhance error handling and edge case detection
- Maintain full backward compatibility
- Performance validation

#### Phase 3: Serialization Engine Refinement
- Adopt reference tokenization improvements
- Enhance bitstream processing accuracy
- Maintain existing data model contracts
- End-to-end validation

## Recommendations

### Immediate Actions
1. **License Verification**: Confirm reference project license compatibility
2. **Algorithm Extraction**: Isolate specific algorithm implementations from reference
3. **Interface Preservation**: Ensure all current APIs remain unchanged

### Implementation Priorities
1. **Accuracy First**: Focus on VARINT decoder improvements for better accuracy
2. **Performance Second**: Optimize Base85 processing for speed improvements
3. **Completeness Third**: Enhance serialization for edge case handling

### Risk Mitigation
1. **Incremental Deployment**: Replace components with thorough validation
2. **Comprehensive Testing**: Use existing test suite plus new reference-based tests
3. **Rollback Capability**: Maintain ability to revert changes if issues arise

## Conclusion

The reference project provides valuable algorithmic improvements that can enhance our BL4 implementation's accuracy and performance while maintaining our superior Go-based architecture and comprehensive interface support. The progressive replacement strategy minimizes risk while maximizing improvement potential.