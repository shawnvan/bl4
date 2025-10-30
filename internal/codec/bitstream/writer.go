package bitstream

import (
	"errors"
	"io"
)

var (
	ErrWriteFailed      = errors.New("failed to write to output")
	ErrWriterClosed     = errors.New("bitstream writer is closed")
	ErrValueTooLarge    = errors.New("value too large for specified bit count")
	ErrWriterNotFlushed = errors.New("writer not flushed before close")
)

// Writer provides efficient bit-level writing to a byte destination
type Writer struct {
	destination io.Writer
	buffer      []byte
	pos         int    // Current byte position in buffer
	bitPos      int    // Current bit position (0-7) in current byte
	bufferSize  int
	closed      bool
}

// NewWriter creates a new bitstream writer with specified buffer size
func NewWriter(destination io.Writer, bufferSize int) *Writer {
	if bufferSize <= 0 {
		bufferSize = 4096 // Default buffer size
	}

	return &Writer{
		destination: destination,
		buffer:      make([]byte, bufferSize),
		pos:         0,
		bitPos:      0,
		bufferSize:  bufferSize,
		closed:      false,
	}
}

// NewWriterFromBytes creates a bitstream writer that writes to a byte slice
func NewWriterFromBytes(initialCapacity int) *Writer {
	if initialCapacity <= 0 {
		initialCapacity = 256 // Default capacity
	}

	return &Writer{
		destination: nil, // No external destination
		buffer:      make([]byte, 0, initialCapacity),
		pos:         0,
		bitPos:      0,
		bufferSize:  initialCapacity,
		closed:      false,
	}
}

// WriteBits writes up to 64 bits to the bitstream
func (w *Writer) WriteBits(value uint64, bitCount int) error {
	if w.closed {
		return ErrWriterClosed
	}

	if bitCount < 0 || bitCount > 64 {
		return ErrInvalidBitCount
	}

	// Check if value fits in specified bit count
	if bitCount < 64 && value >= (1<<bitCount) {
		return ErrValueTooLarge
	}

	remainingBits := bitCount

	for remainingBits > 0 {
		// Ensure buffer has space
		if err := w.ensureSpace(); err != nil {
			return err
		}

		// Calculate bits we can write to current byte
		bitsAvailable := 8 - w.bitPos
		bitsToWrite := remainingBits
		if bitsToWrite > bitsAvailable {
			bitsToWrite = bitsAvailable
		}

		// Extract bits from value
		shift := remainingBits - bitsToWrite
		bits := uint8((value >> shift) & ((1 << bitsToWrite) - 1))

		// Write bits to current byte
		if w.bitPos == 0 {
			w.buffer[w.pos] = bits << (8 - bitsToWrite)
		} else {
			w.buffer[w.pos] |= bits << (8 - w.bitPos - bitsToWrite)
		}

		// Update position
		w.bitPos += bitsToWrite
		if w.bitPos >= 8 {
			w.pos++
			w.bitPos = 0
		}

		remainingBits -= bitsToWrite
	}

	return nil
}

// WriteBit writes a single bit to the bitstream
func (w *Writer) WriteBit(bit bool) error {
	var value uint64 = 0
	if bit {
		value = 1
	}
	return w.WriteBits(value, 1)
}

// WriteByte writes a full byte to the bitstream (8 bits)
func (w *Writer) WriteByte(b byte) error {
	return w.WriteBits(uint64(b), 8)
}

// WriteBytes writes multiple bytes to the bitstream
func (w *Writer) WriteBytes(data []byte) error {
	for _, b := range data {
		if err := w.WriteByte(b); err != nil {
			return err
		}
	}
	return nil
}

// AlignToByte aligns the bit position to the next byte boundary
// If we're not byte-aligned, padding bits will be written as 0
func (w *Writer) AlignToByte() error {
	if w.closed {
		return ErrWriterClosed
	}

	if w.bitPos != 0 {
		// Write padding bits as 0
		paddingBits := 8 - w.bitPos
		return w.WriteBits(0, paddingBits)
	}
	return nil
}

// IsAligned returns true if current position is byte-aligned
func (w *Writer) IsAligned() bool {
	return w.bitPos == 0
}

// Position returns the current bit position in the stream
func (w *Writer) Position() int64 {
	return int64(w.pos*8 + w.bitPos)
}

// BytesWritten returns the number of bytes written so far
func (w *Writer) BytesWritten() int {
	return w.pos
}

// ensureSpace ensures we have space to write more data
func (w *Writer) ensureSpace() error {
	// If we need more space and have an external destination, flush buffer
	if w.pos >= len(w.buffer) && w.destination != nil {
		if err := w.flush(); err != nil {
			return err
		}
	}

	// If still no space and we're writing to byte slice, grow the buffer
	if w.pos >= len(w.buffer) && w.destination == nil {
		// Grow buffer by double or at least by bufferSize
		newSize := len(w.buffer) * 2
		if newSize < w.bufferSize {
			newSize = w.bufferSize
		}

		newBuffer := make([]byte, newSize)
		copy(newBuffer, w.buffer)
		w.buffer = newBuffer
	}

	return nil
}

// flush writes the buffer to the destination
func (w *Writer) flush() error {
	if w.destination == nil {
		return nil // No external destination
	}

	if w.pos > 0 {
		n, err := w.destination.Write(w.buffer[:w.pos])
		if err != nil {
			return ErrWriteFailed
		}
		if n != w.pos {
			return ErrWriteFailed
		}

		// Reset buffer position
		w.pos = 0
		w.bitPos = 0
	}

	return nil
}

// Flush writes any pending data to the destination and aligns to byte boundary
func (w *Writer) Flush() error {
	if w.closed {
		return ErrWriterClosed
	}

	// Align to byte boundary first
	if err := w.AlignToByte(); err != nil {
		return err
	}

	// Flush buffer to destination
	return w.flush()
}

// Bytes returns the written bytes (only works for in-memory writers)
func (w *Writer) Bytes() ([]byte, error) {
	if w.destination != nil {
		return nil, errors.New("Bytes() only available for in-memory writers")
	}

	// Return a copy of the written data
	result := make([]byte, w.pos)
	copy(result, w.buffer[:w.pos])
	return result, nil
}

// Reset clears the writer state and prepares for new writing
func (w *Writer) Reset() {
	w.pos = 0
	w.bitPos = 0
}

// Close closes the writer and flushes any remaining data
func (w *Writer) Close() error {
	if w.closed {
		return nil
	}

	// Flush any remaining data
	if err := w.Flush(); err != nil {
		return err
	}

	w.closed = true
	return nil
}