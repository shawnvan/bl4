# Reference Project Technical Specifications

**Repository**: https://github.com/Nicnl/borderlands4-serials
**License**: GNU Affero General Public License v3.0 (AGPL-3.0) ⚠️
**Last Updated**: 2025-11-03

## Project Overview
Borderlands 4 item serialization and deserialization algorithms for game item codes.

## Key Algorithms Identified

### 1. Base85 Encoding/Decoding
- **Input**: Base85 strings with BL4-specific format
- **Output**: Binary data streams
- **Features**: Custom charset, `@U` prefix validation, padding support

### 2. VARINT Decoding
- **Method**: 4-bit nibble processing with continuation bits
- **Maximum**: 4 blocks (16 bits usable data)
- **Processing**: Bit-level mirroring and validation

### 3. Tokenization Engine
- **Token Types**: SEP1, SEP2, VARINT, PART, VARBIT, STRING
- **Processing**: Bitstream parsing with state machine

### 4. Serialization Engine
- **Method**: Multi-layer serialization (Base85 → Bitstream → Token → Data)
- **Features**: Nested structure support, error recovery

## Integration Considerations

### License Compatibility
- **Reference**: AGPL-3.0 (copyleft)
- **Current**: MIT License (permissive)
- **Recommendation**: Legal review required before direct integration

### Technical Approaches
1. **Algorithm Porting**: Port algorithms to Go while maintaining interface compatibility
2. **Performance Optimization**: Leverage Go's performance advantages
3. **API Preservation**: Maintain all existing interfaces and contracts

## Next Steps
1. Legal review of AGPL-3.0 compatibility
2. Algorithm analysis and Go porting design
3. Performance benchmarking strategy
4. Integration testing framework design