package bitstream

import (
	"testing"

	"github.com/shawnvan/bl4/internal/codec/bitstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBitStreamReader_BasicOperations(t *testing.T) {
	t.Run("Read single bit", func(t *testing.T) {
		data := []byte{0b10101010} // 170 decimal
		reader := bitstream.NewReaderFromBytes(data)

		// Read first bit (should be true - 1)
		bit, err := reader.ReadBit()
		require.NoError(t, err)
		assert.True(t, bit)

		// Read second bit (should be false - 0)
		bit, err = reader.ReadBit()
		require.NoError(t, err)
		assert.False(t, bit)

		// Position should be at bit 2
		assert.Equal(t, int64(2), reader.Position())
	})

	t.Run("Read multiple bits", func(t *testing.T) {
		data := []byte{0b11011010, 0b10101100} // 218, 172
		reader := bitstream.NewReaderFromBytes(data)

		// Read 3 bits: 110 (6)
		value, err := reader.ReadBits(3)
		require.NoError(t, err)
		assert.Equal(t, uint64(6), value)

		// Read 5 bits: 11010 (26)
		value, err = reader.ReadBits(5)
		require.NoError(t, err)
		assert.Equal(t, uint64(26), value)

		// Read 8 bits: 10101100 (172)
		value, err = reader.ReadBits(8)
		require.NoError(t, err)
		assert.Equal(t, uint64(172), value)
	})

	t.Run("Read full byte", func(t *testing.T) {
		data := []byte{0x42, 0x84} // 66, 132
		reader := bitstream.NewReaderFromBytes(data)

		b, err := reader.ReadByte()
		require.NoError(t, err)
		assert.Equal(t, byte(0x42), b)

		b, err = reader.ReadByte()
		require.NoError(t, err)
		assert.Equal(t, byte(0x84), b)
	})

	t.Run("Read multiple bytes", func(t *testing.T) {
		data := []byte{0x01, 0x02, 0x03, 0x04}
		reader := bitstream.NewReaderFromBytes(data)

		bytes, err := reader.ReadBytes(3)
		require.NoError(t, err)
		assert.Equal(t, []byte{0x01, 0x02, 0x03}, bytes)

		// Should have one byte left
		assert.Equal(t, 1, reader.BytesAvailable())
	})

	t.Run("Read across byte boundaries", func(t *testing.T) {
		data := []byte{0b11110000, 0b00001111} // 240, 15
		reader := bitstream.NewReaderFromBytes(data)

		// Read 6 bits from first byte: 111100 (60)
		value, err := reader.ReadBits(6)
		require.NoError(t, err)
		assert.Equal(t, uint64(60), value)

		// Read 10 bits spanning both bytes: 00 00001111 (15)
		value, err = reader.ReadBits(10)
		require.NoError(t, err)
		assert.Equal(t, uint64(15), value)
	})
}

func TestBitStreamReader_ErrorHandling(t *testing.T) {
	t.Run("Read beyond end of data", func(t *testing.T) {
		data := []byte{0xFF}
		reader := bitstream.NewReaderFromBytes(data)

		// Read all 8 bits
		_, err := reader.ReadBits(8)
		require.NoError(t, err)

		// Try to read one more bit
		_, err = reader.ReadBit()
		assert.ErrorIs(t, err, bitstream.ErrEOF)

		// Try to read more bits
		_, err = reader.ReadBits(4)
		assert.ErrorIs(t, err, bitstream.ErrEOF)
	})

	t.Run("Invalid bit count", func(t *testing.T) {
		data := []byte{0xFF}
		reader := bitstream.NewReaderFromBytes(data)

		// Negative bit count
		_, err := reader.ReadBits(-1)
		assert.ErrorIs(t, err, bitstream.ErrInvalidBitCount)

		// Too many bits (> 64)
		_, err = reader.ReadBits(65)
		assert.ErrorIs(t, err, bitstream.ErrInvalidBitCount)
	})

	t.Run("Read byte at non-aligned position", func(t *testing.T) {
		data := []byte{0b10101010, 0b11110000}
		reader := bitstream.NewReaderFromBytes(data)

		// Read 3 bits, leaving us at position 3 (not byte-aligned)
		_, err := reader.ReadBits(3)
		require.NoError(t, err)

		// ReadByte should still work by reading bits 3-10
		b, err := reader.ReadByte()
		require.NoError(t, err)
		// Bits 3-10: 01011110 (0x5E)
		assert.Equal(t, byte(0x5E), b)
	})
}

func TestBitStreamReader_AdvancedOperations(t *testing.T) {
	t.Run("Align to byte", func(t *testing.T) {
		data := []byte{0b10101010, 0b11110000, 0b00001111}
		reader := bitstream.NewReaderFromBytes(data)

		// Read 3 bits
		_, err := reader.ReadBits(3)
		require.NoError(t, err)
		assert.Equal(t, int64(3), reader.Position())
		assert.False(t, reader.IsAligned())

		// Align to byte boundary
		reader.AlignToByte()
		assert.Equal(t, int64(8), reader.Position())
		assert.True(t, reader.IsAligned())

		// Next read should start from second byte
		b, err := reader.ReadByte()
		require.NoError(t, err)
		assert.Equal(t, byte(0b11110000), b)
	})

	t.Run("IsAligned check", func(t *testing.T) {
		data := []byte{0xFF, 0xFF}
		reader := bitstream.NewReaderFromBytes(data)

		// Initially aligned
		assert.True(t, reader.IsAligned())

		// Read 1 bit
		_, err := reader.ReadBit()
		require.NoError(t, err)
		assert.False(t, reader.IsAligned())

		// Read 7 more bits to align
		_, err = reader.ReadBits(7)
		require.NoError(t, err)
		assert.True(t, reader.IsAligned())
	})

	t.Run("Position tracking", func(t *testing.T) {
		data := []byte{0xAA, 0x55, 0xFF}
		reader := bitstream.NewReaderFromBytes(data)

		assert.Equal(t, int64(0), reader.Position())

		_, err := reader.ReadBits(3)
		require.NoError(t, err)
		assert.Equal(t, int64(3), reader.Position())

		_, err = reader.ReadByte()
		require.NoError(t, err)
		assert.Equal(t, int64(11), reader.Position())

		_, err = reader.ReadBits(13)
		require.NoError(t, err)
		assert.Equal(t, int64(24), reader.Position())
	})
}

func TestBitStreamReader_PeekOperations(t *testing.T) {
	t.Run("Peek bits without advancing", func(t *testing.T) {
		data := []byte{0b11011010}
		reader := bitstream.NewReaderFromBytes(data)

		// Peek 3 bits
		value, err := reader.PeekBits(3)
		require.NoError(t, err)
		assert.Equal(t, uint64(6), value) // 110

		// Position should not have advanced
		assert.Equal(t, int64(0), reader.Position())

		// Read the same bits to verify
		value, err = reader.ReadBits(3)
		require.NoError(t, err)
		assert.Equal(t, uint64(6), value)

		// Position should now be at 3
		assert.Equal(t, int64(3), reader.Position())
	})

	t.Run("Peek after some reads", func(t *testing.T) {
		data := []byte{0b11011010, 0b10101100}
		reader := bitstream.NewReaderFromBytes(data)

		// Read 2 bits first
		_, err := reader.ReadBits(2)
		require.NoError(t, err)

		// Peek next 4 bits
		value, err := reader.PeekBits(4)
		require.NoError(t, err)
		assert.Equal(t, uint64(13), value) // 1101

		// Position should still be at 2
		assert.Equal(t, int64(2), reader.Position())

		// Read the 4 bits we peeked
		value, err = reader.ReadBits(4)
		require.NoError(t, err)
		assert.Equal(t, uint64(13), value)

		// Position should now be at 6
		assert.Equal(t, int64(6), reader.Position())
	})
}

func TestBitStreamReader_SkipOperations(t *testing.T) {
	t.Run("Skip bits", func(t *testing.T) {
		data := []byte{0b11011010, 0b10101100}
		reader := bitstream.NewReaderFromBytes(data)

		// Skip 3 bits
		err := reader.SkipBits(3)
		require.NoError(t, err)
		assert.Equal(t, int64(3), reader.Position())

		// Read next bit (should be bit 3)
		bit, err := reader.ReadBit()
		require.NoError(t, err)
		assert.True(t, bit) // Original bit 3 is 1

		// Skip 10 bits (spanning byte boundary)
		err = reader.SkipBits(10)
		require.NoError(t, err)
		assert.Equal(t, int64(14), reader.Position())
	})

	t.Run("Skip to end and beyond", func(t *testing.T) {
		data := []byte{0xFF}
		reader := bitstream.NewReaderFromBytes(data)

		// Skip to the end
		err := reader.SkipBits(8)
		require.NoError(t, err)
		assert.Equal(t, int64(8), reader.Position())

		// Skip beyond end (should error)
		err = reader.SkipBits(1)
		assert.ErrorIs(t, err, bitstream.ErrEOF)
	})

	t.Run("Skip bytes and bits combination", func(t *testing.T) {
		data := []byte{0x11, 0x22, 0x33, 0x44}
		reader := bitstream.NewReaderFromBytes(data)

		// Skip 1 byte + 3 bits = 11 bits
		err := reader.SkipBits(11)
		require.NoError(t, err)
		assert.Equal(t, int64(11), reader.Position())

		// Read next bit
		bit, err := reader.ReadBit()
		require.NoError(t, err)
		// This should be bit 11 of the data stream
		expectedBit := (data[1] >> 2) & 1 // Bit 2 of second byte (0-indexed)
		expectedBool := expectedBit == 1
		assert.Equal(t, expectedBool, bit)
	})
}

func TestBitStreamReader_BufferManagement(t *testing.T) {
	t.Run("Bytes available", func(t *testing.T) {
		data := []byte{0x01, 0x02, 0x03, 0x04}
		reader := bitstream.NewReaderFromBytes(data)

		assert.Equal(t, 4, reader.BytesAvailable())

		// Read 1 byte
		_, err := reader.ReadByte()
		require.NoError(t, err)
		assert.Equal(t, 3, reader.BytesAvailable())

		// Read 1 bit
		_, err = reader.ReadBit()
		require.NoError(t, err)
		// Still 3 bytes available (we haven't moved to next byte yet)
		assert.Equal(t, 3, reader.BytesAvailable())

		// Read 7 more bits to complete the byte
		_, err = reader.ReadBits(7)
		require.NoError(t, err)
		assert.Equal(t, 2, reader.BytesAvailable())
	})

	t.Run("Large data handling", func(t *testing.T) {
		// Create 1KB of test data
		data := make([]byte, 1024)
		for i := range data {
			data[i] = byte(i % 256)
		}

		reader := bitstream.NewReaderFromBytes(data)

		// Read through all data
		for i := 0; i < 1024; i++ {
			b, err := reader.ReadByte()
			require.NoError(t, err)
			assert.Equal(t, data[i], b)
		}

		// Should be at EOF
		assert.Equal(t, 0, reader.BytesAvailable())
		_, err := reader.ReadBit()
		assert.ErrorIs(t, err, bitstream.ErrEOF)
	})
}

func TestBitStreamReader_EdgeCases(t *testing.T) {
	t.Run("Empty data", func(t *testing.T) {
		data := []byte{}
		reader := bitstream.NewReaderFromBytes(data)

		assert.Equal(t, 0, reader.BytesAvailable())
		assert.Equal(t, int64(0), reader.Position())

		// Any read should fail
		_, err := reader.ReadBit()
		assert.ErrorIs(t, err, bitstream.ErrEOF)
	})

	t.Run("Single bit", func(t *testing.T) {
		data := []byte{0x80} // Only first bit is 1
		reader := bitstream.NewReaderFromBytes(data)

		// Read the single 1 bit
		bit, err := reader.ReadBit()
		require.NoError(t, err)
		assert.True(t, bit)

		// Next bit should be 0
		bit, err = reader.ReadBit()
		require.NoError(t, err)
		assert.False(t, bit)
	})

	t.Run("Maximum bit count", func(t *testing.T) {
		data := make([]byte, 9) // Need 9 bytes for 64 bits + alignment
		for i := range data {
			data[i] = 0xFF
		}

		reader := bitstream.NewReaderFromBytes(data)

		// Read exactly 64 bits
		value, err := reader.ReadBits(64)
		require.NoError(t, err)
		assert.Equal(t, uint64(0xFFFFFFFFFFFFFFFF), value)
	})

	t.Run("Zero bit reads", func(t *testing.T) {
		data := []byte{0xFF}
		reader := bitstream.NewReaderFromBytes(data)

		// Reading 0 bits should return 0 without error
		value, err := reader.ReadBits(0)
		require.NoError(t, err)
		assert.Equal(t, uint64(0), value)
	})
}

// Benchmark tests
func BenchmarkBitStreamReader_ReadBit(b *testing.B) {
	data := make([]byte, 1024)
	for i := range data {
		data[i] = 0xAA
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := bitstream.NewReaderFromBytes(data)
		for j := 0; j < 1024*8; j++ {
			reader.ReadBit()
		}
	}
}

func BenchmarkBitStreamReader_ReadBits8(b *testing.B) {
	data := make([]byte, 1024)
	for i := range data {
		data[i] = 0xFF
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := bitstream.NewReaderFromBytes(data)
		for j := 0; j < 1024; j++ {
			reader.ReadBits(8)
		}
	}
}

func BenchmarkBitStreamReader_ReadByte(b *testing.B) {
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := bitstream.NewReaderFromBytes(data)
		for j := 0; j < 1024; j++ {
			reader.ReadByte()
		}
	}
}