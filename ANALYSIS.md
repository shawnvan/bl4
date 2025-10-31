# BL4 Decode Implementation Analysis

## Issue Summary
The current implementation fails during STRING token parsing with the error "end of bitstream" / "unexpected EOF reading second bit" when processing the sample BL4 code "@Ugy3L+2}TYgAAABp2iMAAAA".

## Root Cause Analysis

### 1. Base85 Decoding ✅ WORKING
The Base85 decoding is working correctly:
- Input: "@Uge8s!m/)}}!tjZpNWCvC00"
- Output: 17 bytes (21a4716019062704430f910512bdf43400)
- Total bits: 136

### 2. Bitstream Reading ✅ WORKING
The bitstream reader correctly reads bits and returns EOF at the end.

### 3. Token Parsing ❌ FAILING
The issue is in the `NextToken()` function in `internal/codec/token/tokenizer.go`:

**Problem Location**: Lines 67-73
```go
b2, err := t.reader.ReadBits(1)
if err != nil {
    if err == bitstream.ErrEOF {
        return Token{}, fmt.Errorf("unexpected EOF reading second bit")  // ❌ WRONG
    }
    return Token{}, err
}
```

**Issue**: When there's only 1 bit left in the bitstream, the tokenizer:
1. Successfully reads the first bit (b1)
2. Fails to read the second bit (b2) with EOF
3. Returns an error instead of handling EOF gracefully

### 4. Position Analysis
- Total bits: 136 (positions 0-135)
- Token 15 ends at position 135
- Token 16 tries to read from position 135 (first bit) + 136 (second bit)
- Only position 135 exists, position 136 is EOF

## Expected Behavior vs Current Implementation

### Current (Broken) Logic:
1. Read first bit → success
2. Read second bit → EOF → error → crash

### Expected (Fixed) Logic:
1. Read first bit → success
2. Read second bit → EOF → return EOF token
3. Handle incomplete token headers gracefully

## Fix Required

### Primary Fix: Handle EOF in tokenizer
In `internal/codec/token/tokenizer.go`, modify the `NextToken()` function:

**Current code (lines 67-73):**
```go
b2, err := t.reader.ReadBits(1)
if err != nil {
    if err == bitstream.ErrEOF {
        return Token{}, fmt.Errorf("unexpected EOF reading second bit")
    }
    return Token{}, err
}
```

**Fixed code:**
```go
b2, err := t.reader.ReadBits(1)
if err != nil {
    if err == bitstream.ErrEOF {
        // Only one bit available - this means we have an incomplete token header
        // Should return EOF token instead of error
        return Token{
            Type:     TokenEOF,
            Value:    nil,
            RawData:  nil,
            BitSize:  1, // Only read 1 bit before EOF
            Position: startPos,
        }, nil
    }
    return Token{}, err
}
```

## Test Results

### Working Sample:
- Input: "@Uge8s!m/)}}!tjZpNWCvC00"
- Status: ✅ Base85 decode works
- Status: ❌ Tokenization fails at token 16

### Problematic Sample:
- Input: "@Ugy3L+2}TYgAAABp2iMAAAA" 
- Expected: Should decode item data
- Actual: Fails during tokenization

## Conclusion

The fix is straightforward: improve EOF handling in the tokenizer to gracefully handle cases where the bitstream ends in the middle of reading a token header. This matches real-world scenarios where BL4 item codes may have trailing bits that don't form complete tokens.

The fix should allow the tokenizer to return an EOF token when it encounters EOF during token header reading, rather than returning an error.
