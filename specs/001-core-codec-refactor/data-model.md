# Data Model: Core Codec Algorithm Refactor

**Date**: 2025-11-03
**Feature**: 001-core-codec-refactor

## Core Entities

### 1. Item Serial Code
```go
type SerialCode struct {
    // Raw Base85 encoded string from game
    Value    string `json:"serial_code" validate:"required"`
    // BL4-specific prefix validation
    IsValid  bool   `json:"is_valid"`
    // Length in characters
    Length   int    `json:"length"`
}
```

### 2. Decoded Item Data
```go
type ItemData struct {
    // Item basic information
    Level        int                    `json:"level"`
    Type         string                 `json:"type"`
    Manufacturer string                 `json:"manufacturer"`
    Rarity       string                 `json:"rarity"`

    // Item components
    Parts        []PartData             `json:"parts"`

    // Additional properties
    Properties   map[string]interface{} `json:"properties"`

    // Metadata
    Version      string                 `json:"version"`
    Checksum     string                 `json:"checksum,omitempty"`
}

type PartData struct {
    // 16-bit part index
    Index    int16   `json:"index"`
    // 16-bit part value
    Value    int16   `json:"value"`
    // Nested parts (if applicable)
    SubParts []PartData `json:"sub_parts,omitempty"`
    // Part type classification
    Type     string  `json:"type,omitempty"`
}
```

### 3. Token Representation
```go
type Token struct {
    // Token type identifier
    Type     TokenType `json:"type"`
    // Raw bytes value
    Value    []byte    `json:"value"`
    // Bit length
    BitCount int       `json:"bit_count"`
    // Position in bitstream
    Position int       `json:"position"`
}

// Token types matching reference implementation
type TokenType int

const (
    TOK_SEP1   TokenType = iota // 00: Hard separator
    TOK_SEP2                    // 01: Soft separator
    TOK_VARINT                  // 100: Variable-length integer
    TOK_PART                    // 101: Item part (16-bit index + 16-bit value)
    TOK_VARBIT                  // 110: Variable bit array
    TOK_STRING                  // 111: Text string
)
```

### 4. VARINT Structure
```go
type VARINT struct {
    // Decoded integer value
    Value    uint64 `json:"value"`
    // Number of bits used to encode
    BitCount int    `json:"bit_count"`
    // Raw bytes representation
    RawBytes []byte `json:"raw_bytes,omitempty"`
}

// VARINT encoding structure (4-bit nibble blocks)
type VARINTBlock struct {
    // 4-bit data block
    Data     uint8 `json:"data"`
    // Continuation bit (1 = continue, 0 = last block)
    Continue bool  `json:"continue"`
    // Block position
    Position int   `json:"position"`
}
```

### 5. Processing Results
```go
type DecodingResult struct {
    // Operation success status
    Success      bool                    `json:"success"`
    // Input serial code
    SerialCode   string                  `json:"serial_code"`
    // Decoded item data
    ItemData     *ItemData               `json:"item_data,omitempty"`
    // Raw decoded bytes
    DecodedBytes []byte                  `json:"decoded_bytes,omitempty"`
    // Token stream for debugging
    Tokens       []Token                 `json:"tokens,omitempty"`
    // Processing duration
    Duration     time.Duration           `json:"duration_ms"`
    // Processing statistics
    Stats        ProcessingStats         `json:"stats"`
    // Error information (if failed)
    Error        *ErrorInfo              `json:"error,omitempty"`
}

type EncodingResult struct {
    // Operation success status
    Success      bool                    `json:"success"`
    // Input item data
    ItemData     *ItemData               `json:"item_data"`
    // Encoded serial code
    SerialCode   string                  `json:"serial_code,omitempty"`
    // Processing duration
    Duration     time.Duration           `json:"duration_ms"`
    // Processing statistics
    Stats        ProcessingStats         `json:"stats"`
    // Error information (if failed)
    Error        *ErrorInfo              `json:"error,omitempty"`
}

type ProcessingStats struct {
    // Bytes processed
    BytesProcessed   int `json:"bytes_processed"`
    // Tokens generated
    TokensGenerated  int `json:"tokens_generated"`
    // Parts identified
    PartsIdentified  int `json:"parts_identified"`
    // Memory usage (bytes)
    MemoryUsage      int `json:"memory_usage"`
}

type ErrorInfo struct {
    // Error code
    Code    string `json:"code"`
    // Human-readable message
    Message string `json:"message"`
    // Error position in input
    Position int  `json:"position,omitempty"`
    // Suggested fix
    Suggestion string `json:"suggestion,omitempty"`
}
```

### 6. Batch Processing
```go
type BatchRequest struct {
    // List of serial codes to process
    SerialCodes []string `json:"serial_codes" validate:"required,min=1,max=1000"`
    // Processing options
    Options     ProcessingOptions `json:"options"`
}

type BatchResult struct {
    // Total items processed
    TotalCount     int                    `json:"total_count"`
    // Successful results
    SuccessfulCount int                   `json:"successful_count"`
    // Failed results
    FailedCount    int                    `json:"failed_count"`
    // Individual results
    Results        []DecodingResult      `json:"results"`
    // Processing summary
    Summary        BatchSummary          `json:"summary"`
}

type BatchSummary struct {
    // Total processing time
    TotalDuration    time.Duration `json:"total_duration_ms"`
    // Average processing time per item
    AverageDuration  time.Duration `json:"average_duration_ms"`
    // Total memory usage
    TotalMemoryUsage int64          `json:"total_memory_usage"`
}

type ProcessingOptions struct {
    // Include debug information
    IncludeDebug     bool `json:"include_debug"`
    // Include raw bytes
    IncludeRawBytes  bool `json:"include_raw_bytes"`
    // Validate checksums
    ValidateChecksum bool `json:"validate_checksum"`
}
```

## Data Relationships

```
SerialCode (Input)
    ↓ [Decoding Process]
Token[] (Intermediate)
    ↓ [Token Assembly]
ItemData (Output)

ItemData (Input)
    ↓ [Serialization]
Token[] (Intermediate)
    ↓ [Base85 Encoding]
SerialCode (Output)

BatchRequest
    ├─ SerialCode[]
    └─ ProcessingOptions
         ↓ [Batch Processing]
    BatchResult
         ├─ DecodingResult[]
         └─ BatchSummary
```

## Validation Rules

### Serial Code Validation
- Must start with `@U` prefix
- Must contain only valid Base85 characters
- Length must be reasonable (1-1000 characters)
- Must be decodable to complete byte sequence

### Item Data Validation
- Level must be between 1-100
- Type must be known game item type
- Manufacturer must be known game manufacturer
- Parts must have valid indices and values

### Processing Validation
- All operations must complete within timeout (30 seconds)
- Memory usage must not exceed limits (100MB per item)
- Results must be reproducible (deterministic)

## Migration Strategy

### Phase 1: VARINT Replacement
- Maintain existing `VARINT` struct interface
- Replace internal decoding algorithm
- Preserve all validation and error handling
- Add enhanced accuracy validation

### Phase 2: Base85 Enhancement
- Maintain existing encoder/decoder interfaces
- Adopt reference implementation optimizations
- Preserve charset and validation rules
- Add performance improvements

### Phase 3: Tokenization Refinement
- Maintain existing `Token` and `TokenType` definitions
- Adopt reference tokenization accuracy improvements
- Preserve token stream structure
- Add enhanced error detection

## Backward Compatibility

All data models maintain full backward compatibility with existing:
- API response formats
- CLI output formats
- TUI display formats
- Configuration file formats
- Test expectations

No breaking changes to existing interfaces or data contracts.