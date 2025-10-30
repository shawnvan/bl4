package bitstream

// MirrorTables contains lookup tables for bit reversal operations
// These tables are used for efficient bit manipulation in the codec
type MirrorTables struct {
	// MirrorByte contains pre-computed bit reversal for all 256 byte values
	// MirrorByte[x] = reverseBits(x, 8)
	MirrorByte [256]byte

	// MirrorNibble contains pre-computed bit reversal for all 16 nibble values
	// MirrorNibble[x] = reverseBits(x, 4)
	MirrorNibble [16]byte
}

// NewMirrorTables creates and initializes the mirror lookup tables
func NewMirrorTables() *MirrorTables {
	mt := &MirrorTables{}

	// Initialize nibble table (0-15)
	for i := 0; i < 16; i++ {
		mt.MirrorNibble[i] = byte(reverseBits(uint(i), 4))
	}

	// Initialize byte table (0-255)
	for i := 0; i < 256; i++ {
		mt.MirrorByte[i] = byte(reverseBits(uint(i), 8))
	}

	return mt
}

// reverseBits reverses the order of bits in a value
func reverseBits(value uint, bits int) uint {
	var result uint
	for i := 0; i < bits; i++ {
		result = (result << 1) | (value & 1)
		value >>= 1
	}
	return result
}

// MirrorByte reverses the bit order of a single byte using lookup table
func (mt *MirrorTables) MirrorByteValue(b byte) byte {
	return mt.MirrorByte[b]
}

// MirrorNibble reverses the bit order of a nibble (4 bits) using lookup table
func (mt *MirrorTables) MirrorNibbleValue(n byte) byte {
	if n > 15 {
		// Mask to 4 bits if larger value provided
		n &= 0x0F
	}
	return mt.MirrorNibble[n]
}

// MirrorWord reverses the bit order of a 16-bit value using byte lookups
func (mt *MirrorTables) MirrorWord(value uint16) uint16 {
	low := mt.MirrorByte[value & 0xFF]
	high := mt.MirrorByte[(value >> 8) & 0xFF]
	return uint16(high) | (uint16(low) << 8)
}

// MirrorDword reverses the bit order of a 32-bit value using byte lookups
func (mt *MirrorTables) MirrorDword(value uint32) uint32 {
	b0 := mt.MirrorByte[value & 0xFF]
	b1 := mt.MirrorByte[(value >> 8) & 0xFF]
	b2 := mt.MirrorByte[(value >> 16) & 0xFF]
	b3 := mt.MirrorByte[(value >> 24) & 0xFF]

	return (uint32(b0) << 24) | (uint32(b1) << 16) | (uint32(b2) << 8) | uint32(b3)
}

// Global instance for reuse across the codec
var GlobalMirrorTables = NewMirrorTables()

// Fast mirror functions using global tables
func MirrorByteFast(b byte) byte {
	return GlobalMirrorTables.MirrorByteValue(b)
}

func MirrorNibbleFast(n byte) byte {
	return GlobalMirrorTables.MirrorNibbleValue(n)
}

func MirrorWordFast(value uint16) uint16 {
	return GlobalMirrorTables.MirrorWord(value)
}

func MirrorDwordFast(value uint32) uint32 {
	return GlobalMirrorTables.MirrorDword(value)
}