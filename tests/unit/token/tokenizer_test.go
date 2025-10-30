package token

import (
	"testing"

	"github.com/shawnvan/bl4/internal/codec/bitstream"
	"github.com/shawnvan/bl4/internal/codec/token"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock bitstream data for testing token parsing
// This simulates the bitstream format that would come from decoded Base85 item data
var mockItemBitstream = []byte{
	// Header: VARINT tokens for level (24), type (0=pistol), manufacturer (1=maliwan)
	0b001_00000, // VARINT header + 5-bit count for level value
	0b00011000, // Level value 24 (in 5 bits)

	0b001_00000, // VARINT header + 5-bit count for type
	0b00000000, // Type value 0 (pistol)

	0b001_00000, // VARINT header + 5-bit count for manufacturer
	0b00000001, // Manufacturer value 1 (maliwan)

	// Parts data: Two PART tokens
	0b011_00000, // PART header + reserved bits
	0b00000000, 0b00000000, // Part 1: index=0
	0b00000001, 0b00110010, // Part 1: value=50

	0b011_00000, // PART header + reserved bits
	0b00000010, 0b00001010, // Part 2: index=2
	0b00001101, 0b00001011, // Part 2: value=3379
}

func TestTokenizer_BasicTokenParsing(t *testing.T) {
	t.Run("Parse VARINT token", func(t *testing.T) {
		// Create bitstream with a single VARINT token
		data := []byte{0b001_01010, 0b00001010} // VARINT header + 5-bit count + value 10
		reader := bitstream.NewReaderFromBytes(data)
		tokenizer := token.NewTokenizer(reader)

		tok, err := tokenizer.NextToken()
		require.NoError(t, err)
		assert.Equal(t, token.TokenVARINT, tok.Type)
		assert.Equal(t, uint64(10), tok.Value)
		assert.Equal(t, 14, tok.BitSize) // 3 header + 6 size + 5 data
		assert.Equal(t, int64(0), tok.Position)
	})

	t.Run("Parse VARBIT token", func(t *testing.T) {
		// Create bitstream with a VARBIT token containing 4 bits: 1011
		data := []byte{0b010_00100, 0b10110000} // VARBIT header + 4-bit count + bits
		reader := bitstream.NewReaderFromBytes(data)
		tokenizer := token.NewTokenizer(reader)

		tok, err := tokenizer.NextToken()
		require.NoError(t, err)
		assert.Equal(t, token.TokenVARBIT, tok.Type)

		bits, ok := tok.Value.([]bool)
		require.True(t, ok)
		assert.Len(t, bits, 4)
		assert.Equal(t, []bool{true, false, true, true}, bits)
		assert.Equal(t, 13, tok.BitSize) // 3 header + 6 size + 4 data
	})

	t.Run("Parse PART token", func(t *testing.T) {
		// Create bitstream with a PART token: index=5, value=123
		data := []byte{0b100_00000, 0b00000101, 0b00000000, 0b01111011}
		reader := bitstream.NewReaderFromBytes(data)
		tokenizer := token.NewTokenizer(reader)

		tok, err := tokenizer.NextToken()
		require.NoError(t, err)
		assert.Equal(t, token.TokenPART, tok.Type)

		partData, ok := tok.Value.(map[string]uint64)
		require.True(t, ok)
		assert.Equal(t, uint64(5), partData["index"])
		assert.Equal(t, uint64(123), partData["value"])
		assert.Equal(t, 35, tok.BitSize) // 3 header + 16 index + 16 value
	})

	t.Run("Parse STRING token", func(t *testing.T) {
		// Create bitstream with a STRING token containing "Hello"
		data := []byte{0b101_00101, 0b01001000, 0b01100101, 0b01101100, 0b01101111, 0b00000000}
		reader := bitstream.NewReaderFromBytes(data)
		tokenizer := token.NewTokenizer(reader)

		tok, err := tokenizer.NextToken()
		require.NoError(t, err)
		assert.Equal(t, token.TokenSTRING, tok.Type)

		str, ok := tok.Value.(string)
		require.True(t, ok)
		assert.Equal(t, "Hello", str)
		assert.Equal(t, 51, tok.BitSize) // 3 header + 8 length + 40 data + 8 terminator
	})

	t.Run("Parse MARKER token", func(t *testing.T) {
		// Create bitstream with a MARKER token
		data := []byte{0b110_00101} // MARKER header + 5-bit value
		reader := bitstream.NewReaderFromBytes(data)
		tokenizer := token.NewTokenizer(reader)

		tok, err := tokenizer.NextToken()
		require.NoError(t, err)
		assert.Equal(t, token.TokenMARKER, tok.Type)
		assert.Equal(t, uint64(5), tok.Value)
		assert.Equal(t, 8, tok.BitSize) // 3 header + 5 data
	})
}

func TestTokenizer_ErrorHandling(t *testing.T) {
	t.Run("Invalid token header", func(t *testing.T) {
		// Create bitstream with invalid token header (111 is reserved)
		data := []byte{0b111_00000}
		reader := bitstream.NewReaderFromBytes(data)
		tokenizer := token.NewTokenizer(reader)

		_, err := tokenizer.NextToken()
		assert.ErrorIs(t, err, token.ErrInvalidTokenHeader)
	})

	t.Run("Unexpected EOF", func(t *testing.T) {
		// Create incomplete bitstream
		data := []byte{0b001_01010} // VARINT header but no size/data
		reader := bitstream.NewReaderFromBytes(data)
		tokenizer := token.NewTokenizer(reader)

		_, err := tokenizer.NextToken()
		assert.ErrorIs(t, err, bitstream.ErrEOF)
	})

	t.Run("Invalid VARINT encoding", func(t *testing.T) {
		// Create bitstream with VARINT having 0 bit count (invalid)
		data := []byte{0b001_00000, 0b00000000} // VARINT header + 0 size
		reader := bitstream.NewReaderFromBytes(data)
		tokenizer := token.NewTokenizer(reader)

		_, err := tokenizer.NextToken()
		assert.ErrorIs(t, err, token.ErrInvalidVarInt)
	})

	t.Run("Invalid VARBIT encoding", func(t *testing.T) {
		// Create bitstream with VARBIT having >32 bits (invalid)
		data := []byte{0b010_100000} // VARBIT header + 33 size
		reader := bitstream.NewReaderFromBytes(data)
		tokenizer := token.NewTokenizer(reader)

		_, err := tokenizer.NextToken()
		assert.ErrorIs(t, err, token.ErrInvalidVarBit)
	})

	t.Run("Closed tokenizer", func(t *testing.T) {
		data := []byte{0b001_00000, 0b00000001}
		reader := bitstream.NewReaderFromBytes(data)
		tokenizer := token.NewTokenizer(reader)

		// Close the tokenizer
		err := tokenizer.Close()
		require.NoError(t, err)

		// Try to read from closed tokenizer
		_, err = tokenizer.NextToken()
		assert.ErrorIs(t, err, token.ErrTokenizerClosed)
	})
}

func TestTokenizer_CompleteParsing(t *testing.T) {
	t.Run("TokenizeAll simple case", func(t *testing.T) {
		// Create bitstream with two VARINT tokens
		data := []byte{
			0b001_00010, 0b00000010, // VARINT: value=2 (2 bits)
			0b001_00011, 0b00000101, // VARINT: value=5 (3 bits)
		}
		reader := bitstream.NewReaderFromBytes(data)
		tokenizer := token.NewTokenizer(reader)

		stream, err := tokenizer.TokenizeAll()
		require.NoError(t, err)
		assert.Len(t, stream.Tokens, 3) // 2 VARINT + 1 EOF

		// Check first token
		assert.Equal(t, token.TokenVARINT, stream.Tokens[0].Type)
		assert.Equal(t, uint64(2), stream.Tokens[0].Value)

		// Check second token
		assert.Equal(t, token.TokenVARINT, stream.Tokens[1].Type)
		assert.Equal(t, uint64(5), stream.Tokens[1].Value)

		// Check EOF token
		assert.Equal(t, token.TokenEOF, stream.Tokens[2].Type)
		assert.Equal(t, 0, stream.Tokens[2].BitSize)
	})

	t.Run("TokenizeUntil condition", func(t *testing.T) {
		// Create bitstream with multiple tokens
		data := []byte{
			0b001_00010, 0b00000010, // VARINT: value=2
			0b001_00011, 0b00000101, // VARINT: value=5
			0b010_00010, 0b11000000, // VARBIT: bits=11
			0b110_00001,             // MARKER: value=1
		}
		reader := bitstream.NewReaderFromBytes(data)
		tokenizer := token.NewTokenizer(reader)

		// Tokenize until we find a VARBIT token
		stream, err := tokenizer.TokenizeUntil(func(tok token.Token) bool {
			return tok.Type == token.TokenVARBIT
		})
		require.NoError(t, err)

		// Should have 3 tokens: 2 VARINT + 1 VARBIT
		assert.Len(t, stream.Tokens, 3)
		assert.Equal(t, token.TokenVARBIT, stream.Tokens[2].Type)
	})
}

func TestTokenizer_AdvancedFeatures(t *testing.T) {
	t.Run("PeekToken without advancing", func(t *testing.T) {
		data := []byte{0b001_00010, 0b00000010} // VARINT: value=2
		reader := bitstream.NewReaderFromBytes(data)
		tokenizer := token.NewTokenizer(reader)

		// Peek token
		tok, err := tokenizer.PeekToken()
		require.NoError(t, err)
		assert.Equal(t, token.TokenVARINT, tok.Type)
		assert.Equal(t, uint64(2), tok.Value)

		// Position should not have advanced
		assert.Equal(t, int64(0), tokenizer.Position())

		// Read the same token to verify position hasn't changed
		tok, err = tokenizer.NextToken()
		require.NoError(t, err)
		assert.Equal(t, token.TokenVARINT, tok.Type)
		assert.Equal(t, uint64(2), tok.Value)
	})

	t.Run("HasMoreTokens check", func(t *testing.T) {
		data := []byte{0b001_00010, 0b00000010} // VARINT: value=2
		reader := bitstream.NewReaderFromBytes(data)
		tokenizer := token.NewTokenizer(reader)

		// Should have tokens initially
		assert.True(t, tokenizer.HasMoreTokens())

		// Read the token
		_, err := tokenizer.NextToken()
		require.NoError(t, err)

		// Should still have EOF token
		assert.True(t, tokenizer.HasMoreTokens())

		// Read EOF token
		_, err = tokenizer.NextToken()
		require.NoError(t, err)

		// Should have no more tokens
		assert.False(t, tokenizer.HasMoreTokens())
	})

	t.Run("SkipTokens", func(t *testing.T) {
		// Create bitstream with 3 tokens
		data := []byte{
			0b001_00010, 0b00000010, // VARINT: value=2
			0b001_00011, 0b00000101, // VARINT: value=5
			0b110_00001,             // MARKER: value=1
		}
		reader := bitstream.NewReaderFromBytes(data)
		tokenizer := token.NewTokenizer(reader)

		// Skip 2 tokens
		err := tokenizer.SkipTokens(2)
		require.NoError(t, err)

		// Next token should be the MARKER
		tok, err := tokenizer.NextToken()
		require.NoError(t, err)
		assert.Equal(t, token.TokenMARKER, tok.Type)
		assert.Equal(t, uint64(1), tok.Value)
	})
}

func TestTokenizer_RealWorldSimulation(t *testing.T) {
	t.Run("Simulate item header parsing", func(t *testing.T) {
		// Simulate a realistic item header with level, type, manufacturer
		data := []byte{
			0b001_01000, 0b00011000, // VARINT: level=24 (8 bits)
			0b001_01000, 0b00000000, // VARINT: type=0 (pistol, 8 bits)
			0b001_01000, 0b00000001, // VARINT: manufacturer=1 (maliwan, 8 bits)
		}
		reader := bitstream.NewReaderFromBytes(data)
		tokenizer := token.NewTokenizer(reader)

		// Parse level
		levelTok, err := tokenizer.NextToken()
		require.NoError(t, err)
		assert.Equal(t, token.TokenVARINT, levelTok.Type)
		assert.Equal(t, uint64(24), levelTok.Value)

		// Parse type
		typeTok, err := tokenizer.NextToken()
		require.NoError(t, err)
		assert.Equal(t, token.TokenVARINT, typeTok.Type)
		assert.Equal(t, uint64(0), typeTok.Value)

		// Parse manufacturer
		manuTok, err := tokenizer.NextToken()
		require.NoError(t, err)
		assert.Equal(t, token.TokenVARINT, manuTok.Type)
		assert.Equal(t, uint64(1), manuTok.Value)
	})

	t.Run("Simulate parts list parsing", func(t *testing.T) {
		// Simulate a parts list with markers and part tokens
		data := []byte{
			0b110_00000,             // MARKER: start parts list
			0b100_00000, 0b00000000, 0b00000000, 0b00000001, 0b00110010, // PART: index=0, value=50
			0b100_00000, 0b00000010, 0b00001010, 0b00001101, 0b00001011, // PART: index=2, value=3379
			0b110_00001,             // MARKER: end parts list
		}
		reader := bitstream.NewReaderFromBytes(data)
		tokenizer := token.NewTokenizer(reader)

		stream, err := tokenizer.TokenizeAll()
		require.NoError(t, err)

		// Should have: start marker, 2 parts, end marker, EOF
		assert.Len(t, stream.Tokens, 5)

		// Check structure
		assert.Equal(t, token.TokenMARKER, stream.Tokens[0].Type)
		assert.Equal(t, token.TokenPART, stream.Tokens[1].Type)
		assert.Equal(t, token.TokenPART, stream.Tokens[2].Type)
		assert.Equal(t, token.TokenMARKER, stream.Tokens[3].Type)
		assert.Equal(t, token.TokenEOF, stream.Tokens[4].Type)

		// Verify part data
		part1 := stream.Tokens[1].Value.(map[string]uint64)
		assert.Equal(t, uint64(0), part1["index"])
		assert.Equal(t, uint64(50), part1["value"])

		part2 := stream.Tokens[2].Value.(map[string]uint64)
		assert.Equal(t, uint64(2), part2["index"])
		assert.Equal(t, uint64(3379), part2["value"])
	})
}

func TestTokenStream_Validation(t *testing.T) {
	t.Run("Validate token stream", func(t *testing.T) {
		stream := token.NewTokenStream()

		// Add valid tokens
		stream.AddToken(token.Token{
			Type:  token.TokenVARINT,
			Value: uint64(24),
		})
		stream.AddToken(token.Token{
			Type:  token.TokenPART,
			Value: map[string]uint64{"index": 0, "value": 50},
		})
		stream.AddToken(token.Token{
			Type:  token.TokenEOF,
			Value: nil,
		})

		// Validation should pass
		err := token.ValidateTokenStream(stream)
		assert.NoError(t, err)
	})

	t.Run("Invalidate bad token values", func(t *testing.T) {
		stream := token.NewTokenStream()

		// Add token with wrong value type
		stream.AddToken(token.Token{
			Type:  token.TokenVARINT,
			Value: "not a number", // Should be uint64
		})

		// Validation should fail
		err := token.ValidateTokenStream(stream)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid value")
	})

	t.Run("Invalidate multiple EOF tokens", func(t *testing.T) {
		stream := token.NewTokenStream()

		stream.AddToken(token.Token{Type: token.TokenEOF})
		stream.AddToken(token.Token{Type: token.TokenEOF})

		err := token.ValidateTokenStream(stream)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "multiple EOF")
	})
}

func TestTokenPattern_Matching(t *testing.T) {
	t.Run("Match item header pattern", func(t *testing.T) {
		pattern := token.TokenPattern{token.TokenVARINT, token.TokenVARINT, token.TokenVARINT}

		// Create matching tokens
		tokens := []token.Token{
			{Type: token.TokenVARINT, Value: uint64(24)},
			{Type: token.TokenVARINT, Value: uint64(0)},
			{Type: token.TokenVARINT, Value: uint64(1)},
		}

		assert.True(t, pattern.Matches(tokens))
	})

	t.Run("Non-matching pattern", func(t *testing.T) {
		pattern := token.TokenPattern{token.TokenVARINT, token.TokenVARINT, token.TokenVARINT}

		// Create non-matching tokens
		tokens := []token.Token{
			{Type: token.TokenVARINT, Value: uint64(24)},
			{Type: token.TokenPART, Value: map[string]uint64{"index": 0}}, // Wrong type
			{Type: token.TokenVARINT, Value: uint64(1)},
		}

		assert.False(t, pattern.Matches(tokens))
	})

	t.Run("Pattern length mismatch", func(t *testing.T) {
		pattern := token.TokenPattern{token.TokenVARINT, token.TokenVARINT, token.TokenVARINT}

		// Create tokens with wrong length
		tokens := []token.Token{
			{Type: token.TokenVARINT, Value: uint64(24)},
			{Type: token.TokenVARINT, Value: uint64(0)},
		}

		assert.False(t, pattern.Matches(tokens))
	})
}

// Benchmark tests
func BenchmarkTokenizer_NextToken(b *testing.B) {
	// Create bitstream with many VARINT tokens
	data := make([]byte, 1000)
	for i := 0; i < 1000; i += 2 {
		data[i] = 0b001_01000 // VARINT header + 8-bit size
		data[i+1] = byte(i / 2)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := bitstream.NewReaderFromBytes(data)
		tokenizer := token.NewTokenizer(reader)

		for j := 0; j < 500; j++ {
			tokenizer.NextToken()
		}
	}
}

func BenchmarkTokenizer_TokenizeAll(b *testing.B) {
	// Create bitstream with moderate number of tokens
	data := make([]byte, 200)
	for i := 0; i < 200; i += 2 {
		data[i] = 0b001_01000 // VARINT header + 8-bit size
		data[i+1] = byte(i / 2)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := bitstream.NewReaderFromBytes(data)
		tokenizer := token.NewTokenizer(reader)
		tokenizer.TokenizeAll()
	}
}