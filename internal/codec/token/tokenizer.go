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
func (t *Tokenizer) NextToken() (Token, error) {
	if t.closed {
		return Token{}, ErrTokenizerClosed
	}

	// Record starting position
	startPos := t.reader.Position()

	// Read token header to determine token type
	header, err := t.reader.ReadBits(3) // 3 bits for token type
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

	// Determine token type from header
	var token Token
	switch TokenType(header) {
	case TokenVARINT:
		token, err = t.parseVARINT(startPos)
	case TokenVARBIT:
		token, err = t.parseVARBIT(startPos)
	case TokenPART:
		token, err = t.parsePART(startPos)
	case TokenSTRING:
		token, err = t.parseSTRING(startPos)
	case TokenMARKER:
		token, err = t.parseMARKER(startPos)
	default:
		return Token{}, ErrInvalidTokenHeader
	}

	if err != nil {
		return Token{}, err
	}

	return token, nil
}

// parseVARINT parses a VARINT token from the bitstream
func (t *Tokenizer) parseVARINT(startPos int64) (Token, error) {
	// Read bit count (6 bits, supports up to 63 bits)
	bitCount, err := t.reader.ReadBits(6)
	if err != nil {
		return Token{}, fmt.Errorf("failed to read VARINT bit count: %w", err)
	}

	if bitCount == 0 || bitCount > 64 {
		return Token{}, ErrInvalidVarInt
	}

	// Read the actual value
	value, err := t.reader.ReadBits(int(bitCount))
	if err != nil {
		return Token{}, fmt.Errorf("failed to read VARINT value: %w", err)
	}

	return Token{
		Type:     TokenVARINT,
		Value:    value,
		RawData:  nil, // Could be populated for debugging
		BitSize:  3 + 6 + int(bitCount), // header(3) + size(6) + data
		Position: startPos,
	}, nil
}

// parseVARBIT parses a VARBIT token from the bitstream
func (t *Tokenizer) parseVARBIT(startPos int64) (Token, error) {
	// Read bit count (6 bits, supports up to 63 bits)
	bitCount, err := t.reader.ReadBits(6)
	if err != nil {
		return Token{}, fmt.Errorf("failed to read VARBIT bit count: %w", err)
	}

	if bitCount == 0 || bitCount > 32 {
		return Token{}, ErrInvalidVarBit
	}

	// Read the actual bits
	bits := make([]bool, bitCount)
	for i := 0; i < int(bitCount); i++ {
		bit, err := t.reader.ReadBit()
		if err != nil {
			return Token{}, fmt.Errorf("failed to read VARBIT at bit %d: %w", i, err)
		}
		bits[i] = bit
	}

	return Token{
		Type:     TokenVARBIT,
		Value:    bits,
		RawData:  nil,
		BitSize:  3 + 6 + int(bitCount), // header(3) + size(6) + data
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
			return Token{}, fmt.Errorf("failed to read STRING character at position %d: %w", i, err)
		}
		chars.WriteByte(charByte)
	}

	// Read null terminator (should be 0)
	terminator, err := t.reader.ReadBits(8)
	if err != nil {
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

// parseMARKER parses a MARKER token from the bitstream
func (t *Tokenizer) parseMARKER(startPos int64) (Token, error) {
	// Read marker value (5 bits)
	value, err := t.reader.ReadBits(5)
	if err != nil {
		return Token{}, fmt.Errorf("failed to read MARKER value: %w", err)
	}

	return Token{
		Type:     TokenMARKER,
		Value:    value,
		RawData:  nil,
		BitSize:  3 + 5, // header(3) + value(5)
		Position: startPos,
	}, nil
}

// TokenizeAll parses all remaining tokens from the bitstream
func (t *Tokenizer) TokenizeAll() (*TokenStream, error) {
	stream := NewTokenStream()

	for {
		token, err := t.NextToken()
		if err != nil {
			return nil, err
		}

		stream.AddToken(token)

		if token.Type == TokenEOF {
			break
		}
	}

	return stream, nil
}

// TokenizeUntil parses tokens until a specific condition is met
func (t *Tokenizer) TokenizeUntil(condition func(Token) bool) (*TokenStream, error) {
	stream := NewTokenStream()

	for {
		token, err := t.NextToken()
		if err != nil {
			return nil, err
		}

		stream.AddToken(token)

		if token.Type == TokenEOF || condition(token) {
			break
		}
	}

	return stream, nil
}

// PeekToken looks at the next token without consuming it
func (t *Tokenizer) PeekToken() (Token, error) {
	if t.closed {
		return Token{}, ErrTokenizerClosed
	}

	// Save current position
	currentPos := t.reader.Position()

	// Read the token
	token, err := t.NextToken()
	if err != nil {
		return Token{}, err
	}

	// Restore position
	t.reader.SkipBits(int(currentPos - t.reader.Position()))
	if t.reader.Position() != currentPos {
		// If direct seek failed, we need to reset by creating a new reader
		// This is a limitation of the bitstream reader interface
		return Token{}, errors.New("cannot restore position for peek")
	}

	return token, nil
}

// HasMoreTokens returns true if there are more tokens to read
func (t *Tokenizer) HasMoreTokens() bool {
	if t.closed {
		return false
	}

	// Try to peek at the next token header
	_, err := t.reader.PeekBits(3)
	return err == nil
}

// SkipTokens skips the specified number of tokens
func (t *Tokenizer) SkipTokens(count int) error {
	if t.closed {
		return ErrTokenizerClosed
	}

	for i := 0; i < count; i++ {
		token, err := t.NextToken()
		if err != nil {
			return err
		}
		if token.Type == TokenEOF {
			break
		}
	}
	return nil
}

// Reset resets the tokenizer to the beginning of the bitstream
func (t *Tokenizer) Reset() error {
	// Note: This requires the underlying reader to support seeking
	// For now, we'll indicate this is not supported
	return errors.New("reset not supported - create new tokenizer instead")
}

// Close closes the tokenizer and releases resources
func (t *Tokenizer) Close() error {
	if t.reader != nil {
		t.reader.Close()
	}
	t.closed = true
	return nil
}

// Position returns the current bit position in the stream
func (t *Tokenizer) Position() int64 {
	return t.reader.Position()
}

// ValidateTokenStream performs basic validation on a parsed token stream
func ValidateTokenStream(stream *TokenStream) error {
	if stream == nil {
		return errors.New("token stream is nil")
	}

	hasEOF := false
	for _, token := range stream.Tokens {
		// Check for multiple EOF tokens
		if token.Type == TokenEOF {
			if hasEOF {
				return errors.New("multiple EOF tokens found")
			}
			hasEOF = true
		}

		// Validate token values
		switch token.Type {
		case TokenVARINT:
			if _, ok := token.Value.(uint64); !ok {
				return fmt.Errorf("VARINT token has invalid value: %v", token.Value)
			}
		case TokenVARBIT:
			if _, ok := token.Value.([]bool); !ok {
				return fmt.Errorf("VARBIT token has invalid value: %v", token.Value)
			}
		case TokenSTRING:
			if _, ok := token.Value.(string); !ok {
				return fmt.Errorf("STRING token has invalid value: %v", token.Value)
			}
		case TokenPART:
			if _, ok := token.Value.(map[string]uint64); !ok {
				return fmt.Errorf("PART token has invalid value: %v", token.Value)
			}
		}
	}

	return nil
}