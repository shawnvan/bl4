package token

import (
	"errors"
	"fmt"
	"strings"

	"github.com/shawnvan/bl4/internal/codec/bitstream"
)

var (
	ErrInvalidTokenHeader = errors.New("invalid token header")
	ErrUnsupportedToken   = errors.New("unsupported token type")
	ErrBitStreamEOF       = errors.New("bitstream ended unexpectedly")
	ErrInvalidVarInt      = errors.New("invalid VARINT encoding")
	ErrInvalidVarBit      = errors.New("invalid VARBIT encoding")
	ErrInvalidPart        = errors.New("invalid PART encoding")
	ErrInvalidString      = errors.New("invalid STRING encoding")
	ErrTokenizerClosed    = errors.New("tokenizer is closed")
)

// Tokenizer parses tokens from a bitstream using the BL4 item codec protocol
type Tokenizer struct {
	reader *bitstream.Reader
	closed bool
}

// NewTokenizer creates a new tokenizer from a bitstream reader
func NewTokenizer(reader *bitstream.Reader) *Tokenizer {
	return &Tokenizer{
		reader: reader,
		closed: false,
	}
}

// NewTokenizerFromBytes creates a tokenizer directly from byte data
func NewTokenizerFromBytes(data []byte) *Tokenizer {
	reader := bitstream.NewReaderFromBytes(data)
	return NewTokenizer(reader)
}

// NextToken reads and parses the next token from the bitstream
// Matches reference implementation logic: https://github.com/Nicnl/borderlands4-serials
func (t *Tokenizer) NextToken() (Token, error) {
	if t.closed {
		return Token{}, ErrTokenizerClosed
	}

	// Record starting position
	startPos := t.reader.Position()

	// Read first two bits (following reference implementation)
	b1, err := t.reader.ReadBits(1)
	if err != nil {
		if err == bitstream.ErrEOF {
			return Token{
				Type:     TokenEOF,
				Value:    nil,
				RawData:  nil,
				BitSize:  0,
				Position: startPos,
			}, nil
		}
		return Token{}, err
	}

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

	// Form 2-bit token
	tok := (b1 << 1) | b2

	// Check for 2-bit separator tokens (following reference implementation)
	switch tok {
	case 0b00: // TOK_SEP1 - hard separator
		return Token{
			Type:     TokenSEP1,
			Value:    nil,
			RawData:  nil,
			BitSize:  2,
			Position: startPos,
		}, nil
	case 0b01: // TOK_SEP2 - soft separator
		return Token{
			Type:     TokenSEP2,
			Value:    nil,
			RawData:  nil,
			BitSize:  2,
			Position: startPos,
		}, nil
	}

	// If we're here, first bit was 1, so we need to read 3rd bit for 3-bit tokens
	b3, err := t.reader.ReadBits(1)
	if err != nil {
		if err == bitstream.ErrEOF {
			// Only two bits available - return EOF token
			return Token{
				Type:     TokenEOF,
				Value:    nil,
				RawData:  nil,
				BitSize:  2, // Only read 2 bits before EOF
				Position: startPos,
			}, nil
		}
		return Token{}, err
	}

	// Form 3-bit token
	tok = (tok << 1) | b3

	// Determine token type from 3-bit header (following reference implementation)
	switch tok {
	case 0b100: // TOK_VARINT
		return t.parseVARINT(startPos)
	case 0b101: // TOK_PART
		return t.parsePART(startPos)
	case 0b110: // TOK_VARBIT
		return t.parseVARBIT(startPos)
	case 0b111: // TOK_STRING
		return t.parseSTRING(startPos)
	default:
		return Token{}, ErrInvalidTokenHeader
	}
}

// parseVARINT parses a VARINT token from the bitstream
// Matches reference implementation: uses 4-bit nibble blocks with continuation bits
func (t *Tokenizer) parseVARINT(startPos int64) (Token, error) {
	const (
		VARINT_NB_BLOCKS       = 4  // Maximum number of blocks
		VARINT_BITS_PER_BLOCK  = 4  // 4 bits per block (nibble)
		VARINT_MAX_USABLE_BITS = VARINT_NB_BLOCKS * VARINT_BITS_PER_BLOCK
	)

	var (
		dataRead = 0
		output   uint32
	)

	// Read up to 4 blocks
	for range VARINT_NB_BLOCKS {
		// Read 4-bit block
		block, err := t.reader.ReadBits(VARINT_BITS_PER_BLOCK)
		if err != nil {
			return Token{}, fmt.Errorf("unexpected end of data while reading varint block: %w", err)
		}

		// Apply 4-bit mirroring (as in reference implementation)
		// For now, we'll skip the mirroring as our bytes are already mirrored
		output |= uint32(block) << dataRead
		dataRead += VARINT_BITS_PER_BLOCK

		// Read continuation bit
		cont, err := t.reader.ReadBits(1)
		if err != nil {
			return Token{}, fmt.Errorf("unexpected end of data while reading varint continuation: %w", err)
		}

		// If continuation bit is 0, this is the last block
		if cont == 0 {
			break
		}
	}

	return Token{
		Type:     TokenVARINT,
		Value:    output,
		RawData:  nil,
		BitSize:  3 + dataRead + (dataRead/VARINT_BITS_PER_BLOCK), // header(3) + data + continuation bits
		Position: startPos,
	}, nil
}

// parseVARBIT parses a VARBIT token from the bitstream
// Matches reference implementation: 5-bit length + bits
func (t *Tokenizer) parseVARBIT(startPos int64) (Token, error) {
	const VARBIT_LENGTH_BLOCK_SIZE = 5

	// Read 5-bit length
	length, err := t.reader.ReadBits(VARBIT_LENGTH_BLOCK_SIZE)
	if err != nil {
		return Token{}, fmt.Errorf("unexpected end of data while reading varbit length: %w", err)
	}

	// Apply 5-bit mirroring (reference implementation does this)
	// For now, we'll skip mirroring as our bytes are already mirrored
	if length > 32 {
		return Token{}, fmt.Errorf("varbit length %d exceeds maximum 32", length)
	}

	var value uint32
	// Read the specified number of bits
	for i := uint32(0); i < uint32(length); i++ {
		bit, err := t.reader.ReadBits(1)
		if err != nil {
			return Token{}, fmt.Errorf("unexpected end of data while reading varbit value: %w", err)
		}

		value |= uint32(bit) << i
	}

	return Token{
		Type:     TokenVARBIT,
		Value:    value,
		RawData:  nil,
		BitSize:  3 + VARBIT_LENGTH_BLOCK_SIZE + int(length), // header(3) + length(5) + data
		Position: startPos,
	}, nil
}

// parsePART parses a PART token from the bitstream
func (t *Tokenizer) parsePART(startPos int64) (Token, error) {
	// Read part index (16 bits)
	index, err := t.reader.ReadBits(16)
	if err != nil {
		return Token{}, fmt.Errorf("failed to read PART index: %w", err)
	}

	// Read part value (16 bits)
	value, err := t.reader.ReadBits(16)
	if err != nil {
		return Token{}, fmt.Errorf("failed to read PART value: %w", err)
	}

	// Store as structured data
	partData := map[string]uint64{
		"index": index,
		"value": value,
	}

	return Token{
		Type:     TokenPART,
		Value:    partData,
		RawData:  nil,
		BitSize:  3 + 16 + 16, // header(3) + index(16) + value(16)
		Position: startPos,
	}, nil
}

// parseSTRING parses a STRING token from the bitstream
func (t *Tokenizer) parseSTRING(startPos int64) (Token, error) {
	// Read string length (8 bits, supports up to 255 characters)
	length, err := t.reader.ReadBits(8)
	if err != nil {
		return Token{}, fmt.Errorf("failed to read STRING length: %w", err)
	}

	if length > 255 {
		return Token{}, ErrInvalidString
	}

	// Read characters
	var chars strings.Builder
	for i := 0; i < int(length); i++ {
		charByte, err := t.reader.ReadByte()
		if err != nil {
			// Handle EOF gracefully - return partial string
			if err == bitstream.ErrEOF {
				return Token{
					Type:     TokenSTRING,
					Value:    chars.String(),
					RawData:  nil,
					BitSize:  3 + 8 + (i*8), // header + length + actual chars
					Position: startPos,
				}, nil
			}
			return Token{}, fmt.Errorf("failed to read STRING character at position %d: %w", i, err)
		}
		chars.WriteByte(charByte)
	}

	// Read null terminator (should be 0)
	terminator, err := t.reader.ReadBits(8)
	if err != nil {
		// Handle EOF for terminator - this can happen with incomplete strings
		if err == bitstream.ErrEOF {
			// Return the string without terminator validation
			return Token{
				Type:     TokenSTRING,
				Value:    chars.String(),
				RawData:  nil,
				BitSize:  3 + 8 + (int(length)*8), // header + length + all chars
				Position: startPos,
			}, nil
		}
		return Token{}, fmt.Errorf("failed to read STRING terminator: %w", err)
	}

	if terminator != 0 {
		return Token{}, ErrInvalidString
	}

	return Token{
		Type:     TokenSTRING,
		Value:    chars.String(),
		RawData:  nil,
		BitSize:  3 + 8 + int(length)*8 + 8, // header(3) + length(8) + data + terminator(8)
		Position: startPos,
	}, nil
}

// ... rest of the functions would go here ...
