package bitstream

import (
	"errors"
	"io"
)

var (
	ErrEOF                = errors.New("end of bitstream")
	ErrInvalidBitCount    = errors.New("invalid bit count")
	ErrInsufficientData   = errors.New("insufficient data in bitstream")
	ErrInvalidPosition    = errors.New("invalid bit position")
	ErrReaderNotOpen      = errors.New("bitstream reader not open")
	ErrReaderAlreadyOpen  = errors.New("bitstream reader already open")
)

// Reader provides efficient bit-level reading from a byte source
type Reader struct {
	source     io.Reader
	buffer     []byte
	pos        int    // Current byte position in buffer
	bitPos     int    // Current bit position (0-7) in current byte
	bufferSize int
	eof        bool   // End of file reached
	available  bool   // Buffer has data available
}

// NewReader creates a new bitstream reader with specified buffer size
func NewReader(source io.Reader, bufferSize int) *Reader {
	if bufferSize <= 0 {
		bufferSize = 4096 // Default buffer size
	}

	return &Reader{
		source:     source,
		buffer:     make([]byte, bufferSize),
		pos:        0,
		bitPos:     0,
		bufferSize: bufferSize,
		eof:        false,
		available:  false,
	}
}

// NewReaderFromBytes creates a bitstream reader from a byte slice
func NewReaderFromBytes(data []byte) *Reader {
	return &Reader{
		source:     nil, // No external source
		buffer:     data,
		pos:        0,
		bitPos:     0,
		bufferSize: len(data),
		eof:        false,
		available:  len(data) > 0,
	}
}

// ReadBits reads up to 64 bits from the bitstream
func (r *Reader) ReadBits(bitCount int) (uint64, error) {
	if bitCount < 0 || bitCount > 64 {
		return 0, ErrInvalidBitCount
	}

	var result uint64
	remainingBits := bitCount

	for remainingBits > 0 {
		// Ensure we have data available
		if err := r.ensureData(); err != nil {
			return 0, err
		}

		// Calculate bits we can read from current byte
		bitsInByte := 8 - r.bitPos
		bitsToRead := remainingBits
		if bitsToRead > bitsInByte {
			bitsToRead = bitsInByte
		}

		// Extract bits from current byte (MSB-first like reference implementation)
		currentByte := r.buffer[r.pos]

		// For MSB-first bit reading: shift right to get bits from most significant end
		// Reference implementation uses: (b >> (7 - (br.pos % 8))) & 1
		shiftAmount := 7 - r.bitPos
		currentByte >>= shiftAmount

		// Mask to get only the bits we need
		mask := uint64(1<<bitsToRead - 1)
		bits := uint64(currentByte) & mask

		// Add to result
		result = (result << bitsToRead) | bits

		// Update position
		r.bitPos += bitsToRead
		if r.bitPos >= 8 {
			r.pos++
			r.bitPos = 0
		}

		remainingBits -= bitsToRead
	}

	return result, nil
}

// ReadBit reads a single bit from the bitstream
func (r *Reader) ReadBit() (bool, error) {
	value, err := r.ReadBits(1)
	if err != nil {
		return false, err
	}
	return value == 1, nil
}

// ReadByte reads a full byte from the bitstream (8 bits)
func (r *Reader) ReadByte() (byte, error) {
	value, err := r.ReadBits(8)
	if err != nil {
		return 0, err
	}
	return byte(value), nil
}

// ReadBytes reads multiple bytes from the bitstream
func (r *Reader) ReadBytes(count int) ([]byte, error) {
	if count < 0 {
		return nil, ErrInvalidBitCount
	}

	data := make([]byte, count)
	for i := 0; i < count; i++ {
		b, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		data[i] = b
	}
	return data, nil
}

// AlignToByte aligns the bit position to the next byte boundary
func (r *Reader) AlignToByte() {
	if r.bitPos != 0 {
		r.pos++
		r.bitPos = 0
	}
}

// IsAligned returns true if current position is byte-aligned
func (r *Reader) IsAligned() bool {
	return r.bitPos == 0
}

// Position returns the current bit position in the stream
func (r *Reader) Position() int64 {
	return int64(r.pos*8 + r.bitPos)
}

// BytesAvailable returns the number of bytes available in the current buffer
func (r *Reader) BytesAvailable() int {
	if !r.available {
		return 0
	}
	return len(r.buffer) - r.pos
}

// ensureData ensures we have data available to read from
func (r *Reader) ensureData() error {
	// If we're at end of buffer and have a source, try to read more
	if r.pos >= len(r.buffer) && r.source != nil && !r.eof {
		n, err := r.source.Read(r.buffer)
		if err != nil && err != io.EOF {
			return err
		}

		if err == io.EOF || n == 0 {
			r.eof = true
			return ErrEOF
		}

		r.pos = 0
		r.bitPos = 0
		r.available = n > 0
		r.buffer = r.buffer[:n] // Trim buffer to actual data size
	}

	// Check if we still have data
	if r.pos >= len(r.buffer) {
		return ErrEOF
	}

	r.available = true
	return nil
}

// PeekBits looks ahead at the next bits without advancing the position
func (r *Reader) PeekBits(bitCount int) (uint64, error) {
	if bitCount < 0 || bitCount > 64 {
		return 0, ErrInvalidBitCount
	}

	// Save current position
	originalPos := r.pos
	originalBitPos := r.bitPos

	// Read bits
	value, err := r.ReadBits(bitCount)

	// Restore position
	r.pos = originalPos
	r.bitPos = originalBitPos

	return value, err
}

// SkipBits advances the position by the specified number of bits
func (r *Reader) SkipBits(bitCount int) error {
	if bitCount < 0 {
		return ErrInvalidBitCount
	}

	// Skip whole bytes first
	bytesToSkip := bitCount / 8
	bitsToSkip := bitCount % 8

	// Advance by whole bytes
	r.pos += bytesToSkip

	// Advance remaining bits
	r.bitPos += bitsToSkip
	if r.bitPos >= 8 {
		r.pos++
		r.bitPos -= 8
	}

	// Check if we need to load more data
	if r.pos >= len(r.buffer) {
		return r.ensureData()
	}

	return nil
}

// Close closes the reader and releases resources
func (r *Reader) Close() error {
	r.buffer = nil
	r.source = nil
	r.available = false
	r.eof = true
	return nil
}