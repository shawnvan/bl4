package bitstream

import (
	"math"
)

// SafeReader provides enhanced safety for bitstream operations
type SafeReader struct {
	*Reader
}

// NewSafeReader creates a new safe bitstream reader
func NewSafeReader(source interface{}, bufferSize int) *SafeReader {
	var reader *Reader
	
	switch v := source.(type) {
	case []byte:
		reader = NewReaderFromBytes(v)
	default:
		if r, ok := source.(interface{ Read([]byte) (int, error) }); ok {
			reader = NewReader(r, bufferSize)
		} else {
			reader = NewReaderFromBytes([]byte{})
		}
	}
	
	return &SafeReader{Reader: reader}
}

// SafeReadBits performs safe bit reading with overflow protection
func (sr *SafeReader) SafeReadBits(bitCount int) (uint64, error) {
	if bitCount < 0 || bitCount > 64 {
		return 0, ErrInvalidBitCount
	}

	var result uint64
	remainingBits := bitCount

	for remainingBits > 0 {
		// Ensure we have data available
		if err := sr.ensureData(); err != nil {
			return 0, err
		}

		// Calculate bits we can read from current byte
		bitsInByte := 8 - sr.bitPos
		bitsToRead := remainingBits
		if bitsToRead > bitsInByte {
			bitsToRead = bitsInByte
		}

		// Extract bits from current byte safely
		currentByte := sr.buffer[sr.pos]
		
		// Safe shift amount calculation
		shiftAmount := 7 - sr.bitPos
		if shiftAmount < 0 || shiftAmount > 7 {
			return 0, ErrInvalidPosition
		}

		currentByte >>= shiftAmount

		// Safe mask calculation to prevent overflow
		mask := sr.calculateSafeMask(bitsToRead)
		bits := uint64(currentByte) & mask

		// Add to result with overflow protection
		result = (result << bitsToRead) | bits

		// Update position
		sr.bitPos += bitsToRead
		if sr.bitPos >= 8 {
			sr.pos++
			sr.bitPos = 0
		}

		remainingBits -= bitsToRead
	}

	return result, nil
}

// calculateSafeMask creates a safe bit mask without overflow
func (sr *SafeReader) calculateSafeMask(bitsToRead int) uint64 {
	if bitsToRead <= 0 {
		return 0
	}
	
	if bitsToRead >= 64 {
		return math.MaxUint64
	}
	
	return (uint64(1) << uint(bitsToRead)) - 1
}

// SafePeekBits performs safe peek operations
func (sr *SafeReader) SafePeekBits(bitCount int) (uint64, error) {
	if bitCount < 0 || bitCount > 64 {
		return 0, ErrInvalidBitCount
	}

	// Save current position
	originalPos := sr.pos
	originalBitPos := sr.bitPos

	// Read bits safely
	value, err := sr.SafeReadBits(bitCount)

	// Restore position
	sr.pos = originalPos
	sr.bitPos = originalBitPos

	return value, err
}