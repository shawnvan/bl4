package token

import (
	"fmt"
	"strings"
)

// TokenType represents the type of token in the bitstream
type TokenType int

const (
	// TokenUnknown represents an unknown or invalid token type
	TokenUnknown TokenType = iota

	// TokenSEP1 represents a hard separator (00) - 2 bits
	// Based on reference implementation: TOK_SEP1
	TokenSEP1

	// TokenSEP2 represents a soft separator (01) - 2 bits
	// Based on reference implementation: TOK_SEP2
	TokenSEP2

	// TokenVARINT represents a variable-length integer (100) - 3 bits
	// Format: [bit_count][value] where bit_count determines the value size
	// Based on reference implementation: TOK_VARINT
	TokenVARINT

	// TokenPART represents a part or component reference (101) - 3 bits
	// Format: [index][value] referencing a parts database
	// Based on reference implementation: TOK_PART
	TokenPART

	// TokenVARBIT represents a variable-length bitfield (110) - 3 bits
	// Format: [bit_count][bits] where bit_count determines the bitfield size
	// Based on reference implementation: TOK_VARBIT
	TokenVARBIT

	// TokenSTRING represents a null-terminated string (111) - 3 bits
	// Format: [length][characters][null_terminator]
	// Based on reference implementation: TOK_STRING
	TokenSTRING

	// TokenEOF represents end of bitstream
	TokenEOF

	// TokenMARKER represents special markers or delimiters
	// (Legacy from our original implementation, may not be used in new system)
	TokenMARKER
)

// String returns the string representation of a TokenType
func (t TokenType) String() string {
	switch t {
	case TokenSEP1:
		return "SEP1"
	case TokenSEP2:
		return "SEP2"
	case TokenVARINT:
		return "VARINT"
	case TokenPART:
		return "PART"
	case TokenVARBIT:
		return "VARBIT"
	case TokenSTRING:
		return "STRING"
	case TokenEOF:
		return "EOF"
	case TokenMARKER:
		return "MARKER"
	default:
		return "UNKNOWN"
	}
}

// Token represents a parsed token from the bitstream
type Token struct {
	Type     TokenType   // The type of this token
	Value    interface{} // The parsed value (uint64, []bool, etc.)
	RawData  []byte      // Raw bytes from bitstream (for debugging)
	BitSize  int         // Size in bits this token occupies
	Position int64       // Bit position in the stream where this token starts
}

// String returns a string representation of the token
func (t Token) String() string {
	var valueStr string
	switch t.Type {
	case TokenVARINT:
		if v, ok := t.Value.(uint64); ok {
			valueStr = fmt.Sprintf("%d", v)
		} else {
			valueStr = fmt.Sprintf("%v", t.Value)
		}
	case TokenVARBIT:
		if v, ok := t.Value.([]bool); ok {
			var bits strings.Builder
			for i, bit := range v {
				if i > 0 && i%8 == 0 {
					bits.WriteString(" ")
				}
				if bit {
					bits.WriteString("1")
				} else {
					bits.WriteString("0")
				}
			}
			valueStr = bits.String()
		} else {
			valueStr = fmt.Sprintf("%v", t.Value)
		}
	case TokenPART:
		if v, ok := t.Value.(uint64); ok {
			valueStr = fmt.Sprintf("index:%d", v)
		} else {
			valueStr = fmt.Sprintf("%v", t.Value)
		}
	case TokenSTRING:
		if v, ok := t.Value.(string); ok {
			valueStr = fmt.Sprintf("\"%s\"", v)
		} else {
			valueStr = fmt.Sprintf("%v", t.Value)
		}
	default:
		valueStr = fmt.Sprintf("%v", t.Value)
	}

	return fmt.Sprintf("%s[%s:%d] pos:%d", t.Type, valueStr, t.BitSize, t.Position)
}

// IsDataToken returns true if this token contains data (not structural)
func (t Token) IsDataToken() bool {
	return t.Type == TokenVARINT || t.Type == TokenVARBIT || t.Type == TokenPART || t.Type == TokenSTRING
}

// IsStructuralToken returns true if this token is structural (delimiters, markers, etc.)
func (t Token) IsStructuralToken() bool {
	return t.Type == TokenEOF || t.Type == TokenMARKER
}

// Equals checks if two tokens are equal (same type and value)
func (t Token) Equals(other Token) bool {
	if t.Type != other.Type {
		return false
	}

	switch t.Type {
	case TokenVARINT, TokenPART:
		tVal, tOk := t.Value.(uint64)
		oVal, oOk := other.Value.(uint64)
		return tOk && oOk && tVal == oVal
	case TokenVARBIT:
		tVal, tOk := t.Value.([]bool)
		oVal, oOk := other.Value.([]bool)
		if !tOk || !oOk || len(tVal) != len(oVal) {
			return false
		}
		for i := range tVal {
			if tVal[i] != oVal[i] {
				return false
			}
		}
		return true
	case TokenSTRING:
		tVal, tOk := t.Value.(string)
		oVal, oOk := other.Value.(string)
		return tOk && oOk && tVal == oVal
	default:
		return fmt.Sprintf("%v", t.Value) == fmt.Sprintf("%v", other.Value)
	}
}

// TokenStream represents a sequence of parsed tokens
type TokenStream struct {
	Tokens  []Token // The parsed tokens in order
	BitSize int     // Total size of all tokens in bits
}

// NewTokenStream creates a new empty token stream
func NewTokenStream() *TokenStream {
	return &TokenStream{
		Tokens:  make([]Token, 0),
		BitSize: 0,
	}
}

// AddToken adds a token to the stream
func (ts *TokenStream) AddToken(token Token) {
	ts.Tokens = append(ts.Tokens, token)
	ts.BitSize += token.BitSize
}

// GetToken returns the token at the specified index
func (ts *TokenStream) GetToken(index int) (Token, error) {
	if index < 0 || index >= len(ts.Tokens) {
		return Token{}, fmt.Errorf("token index %d out of range", index)
	}
	return ts.Tokens[index], nil
}

// GetTokensByType returns all tokens of the specified type
func (ts *TokenStream) GetTokensByType(tokenType TokenType) []Token {
	var result []Token
	for _, token := range ts.Tokens {
		if token.Type == tokenType {
			result = append(result, token)
		}
	}
	return result
}

// CountTokensByType returns the count of tokens of the specified type
func (ts *TokenStream) CountTokensByType(tokenType TokenType) int {
	count := 0
	for _, token := range ts.Tokens {
		if token.Type == tokenType {
			count++
		}
	}
	return count
}

// String returns a string representation of the token stream
func (ts *TokenStream) String() string {
	if len(ts.Tokens) == 0 {
		return "TokenStream[empty]"
	}

	var parts []string
	for _, token := range ts.Tokens {
		parts = append(parts, token.String())
	}

	return fmt.Sprintf("TokenStream[%d tokens, %d bits]:\n  %s",
		len(ts.Tokens), ts.BitSize, strings.Join(parts, "\n  "))
}

// Clear removes all tokens from the stream
func (ts *TokenStream) Clear() {
	ts.Tokens = ts.Tokens[:0]
	ts.BitSize = 0
}

// IsEmpty returns true if the stream has no tokens
func (ts *TokenStream) IsEmpty() bool {
	return len(ts.Tokens) == 0
}

// Size returns the number of tokens in the stream
func (ts *TokenStream) Size() int {
	return len(ts.Tokens)
}

// GetTotalBitSize returns the total bit size of all tokens
func (ts *TokenStream) GetTotalBitSize() int {
	return ts.BitSize
}

// Constants for bitstream parsing
const (
	// Maximum bit sizes for different token types
	MaxVarIntBits   = 64
	MaxVarBitBits   = 32
	MaxStringBits   = 256
	MaxPartIndex    = 65535

	// Common bit sizes for game item encoding
	LevelBits       = 8   // Item level (0-255)
	TypeBits        = 8   // Item type (0-255)
	ManufacturerBits = 8  // Manufacturer (0-255)

	// Part encoding constants
	PartsListSeparator = '|'  // Separator between parts
	PartListStart     = '{'   // Start of part list
	PartListEnd       = '}'   // End of part list

	// Bitstream markers
	HeaderMarker      = 0x80  // Header marker bit
	DataMarker        = 0x40  // Data marker bit
	EndMarker         = 0x20  // End marker bit
)

// Common token patterns for item encoding
var (
	// Common token patterns found in BL4 item codes
	ItemHeaderPattern = []TokenType{TokenVARINT, TokenVARINT, TokenVARINT} // level, type, manufacturer
	PartsListPattern  = []TokenType{TokenMARKER, TokenPART, TokenPART, TokenMARKER}
)

// TokenPattern represents a pattern of token types for validation
type TokenPattern []TokenType

// Matches checks if a sequence of tokens matches this pattern
func (tp TokenPattern) Matches(tokens []Token) bool {
	if len(tp) != len(tokens) {
		return false
	}

	for i, expectedType := range tp {
		if tokens[i].Type != expectedType {
			return false
		}
	}

	return true
}

// String returns string representation of the pattern
func (tp TokenPattern) String() string {
	var parts []string
	for _, tokenType := range tp {
		parts = append(parts, tokenType.String())
	}
	return strings.Join(parts, " -> ")
}