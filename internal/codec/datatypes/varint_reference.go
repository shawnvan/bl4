package datatypes

import (
	"fmt"

	"github.com/shawnvan/bl4/internal/codec/bitstream"
	"github.com/shawnvan/bl4/internal/utils"
)

// Reference implementation VARINT constants
const (
	VARINT_NB_BLOCKS       = 4
	VARINT_BITS_PER_BLOCK  = 4
	VARINT_MAX_USABLE_BITS = VARINT_NB_BLOCKS * VARINT_BITS_PER_BLOCK // 16 bits max
)

// EncodeVARINTReference encodes a VARINT using the reference implementation (block-based)
func EncodeVARINTReference(writer *bitstream.Writer, value uint64) error {
	// Figure out how many bits we need to represent 'value'
	nBits := 0
	{
		v2 := value
		for v2 > 0 {
			nBits++
			v2 >>= 1
		}

		// Special case: zero needs at least one block
		if nBits == 0 {
			nBits = 1
		}

		// If too many bits, the rest is discarded silently
		if nBits > VARINT_MAX_USABLE_BITS {
			nBits = VARINT_MAX_USABLE_BITS
		}
	}

	// Write complete blocks
	for nBits > VARINT_BITS_PER_BLOCK {
		// Data bits
		for i := 0; i < VARINT_BITS_PER_BLOCK; i++ {
			bit := (value & 0b1) == 1
			if err := writer.WriteBit(bit); err != nil {
				return fmt.Errorf("failed to write VARINT data bit: %w", err)
			}
			value >>= 1
			nBits--
		}

		// Continuation bit
		if err := writer.WriteBit(true); err != nil {
			return fmt.Errorf("failed to write VARINT continuation bit: %w", err)
		}
	}

	// Write partial last block
	if nBits > 0 {
		for i := 0; i < VARINT_BITS_PER_BLOCK; i++ {
			if nBits > 0 {
				// Data bits
				bit := (value & 0b1) == 1
				if err := writer.WriteBit(bit); err != nil {
					return fmt.Errorf("failed to write VARINT data bit: %w", err)
				}
				value >>= 1
				nBits--
			} else {
				// Padding bit
				if err := writer.WriteBit(false); err != nil {
					return fmt.Errorf("failed to write VARINT padding bit: %w", err)
				}
			}
		}

		// No continuation (last block)
		if err := writer.WriteBit(false); err != nil {
			return fmt.Errorf("failed to write VARINT end bit: %w", err)
		}
	}

	return nil
}

// DecodeVARINTReference decodes a VARINT using the reference implementation (block-based)
func DecodeVARINTReference(reader *bitstream.Reader) (uint64, error) {
	var result uint64
	dataRead := 0

	for range VARINT_NB_BLOCKS {
		// Read block
		block, err := reader.ReadBits(VARINT_BITS_PER_BLOCK)
		if err != nil {
			return 0, fmt.Errorf("failed to read VARINT block: %w", err)
		}

		// Apply 4-bit mirroring (as in reference implementation)
		mirroredBlock := utils.Mirror4Bits(byte(block))
		result |= uint64(mirroredBlock) << dataRead
		dataRead += VARINT_BITS_PER_BLOCK

		// Read continuation bit
		contBit, err := reader.ReadBit()
		if err != nil {
			return 0, fmt.Errorf("failed to read VARINT continuation bit: %w", err)
		}

		// If no continuation bit, this is the last block
		if !contBit {
			break
		}
	}

	return result, nil
}