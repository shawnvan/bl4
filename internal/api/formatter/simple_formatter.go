package formatter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shawnvan/bl4/internal/codec/token"
)

// FormatAsSimpleReference converts token stream to simple reference format
// Matches reference implementation: Serial.String() method
func FormatAsSimpleReference(tokens *token.TokenStream) string {
	if tokens == nil || len(tokens.Tokens) == 0 {
		return ""
	}

	var result strings.Builder

	for i, t := range tokens.Tokens {
		switch t.Type {
		case token.TokenSEP1:
			result.WriteString("|")
		case token.TokenSEP2:
			result.WriteString(",")
		case token.TokenVARINT, token.TokenVARBIT:
			if i > 0 {
				result.WriteString(" ")
			}
			result.WriteString(formatSimpleTokenValue(t.Value))
		case token.TokenPART:
			if i > 0 {
				result.WriteString(" ")
			}
			result.WriteString(formatSimplePart(t.Value))
		case token.TokenSTRING:
			if i > 0 {
				result.WriteString(" ")
			}
			result.WriteString(formatSimpleString(t.Value))
		default:
			// Skip unknown tokens for cleaner output
		}
	}

	return result.String()
}

// formatSimplePart formats PART tokens to match reference implementation Part.String()
func formatSimplePart(value interface{}) string {
	switch v := value.(type) {
	case map[string]interface{}:
		index, _ := v["index"].(float64)
		partValue, hasValue := v["value"].(float64)
		values, hasList := v["values"].([]interface{})

		indexInt := uint32(index)

		if hasList && len(values) > 0 {
			// SUBTYPE_LIST: {index:[val1 val2 ...]}
			var valStrs []string
			var valUint []uint32
			for _, val := range values {
				if f, ok := val.(float64); ok {
					valStrs = append(valStrs, strconv.Itoa(int(f)))
					valUint = append(valUint, uint32(f))
				}
			}
			return fmt.Sprintf("{%d:[%s]}", indexInt, strings.Join(valStrs, " "))
		} else if hasValue {
			// SUBTYPE_INT: {index:value}
			return fmt.Sprintf("{%d:%d}", indexInt, uint32(partValue))
		} else {
			// SUBTYPE_NONE: {index}
			return fmt.Sprintf("{%d}", indexInt)
		}
	default:
		return fmt.Sprintf("{%v}", v)
	}
}

// formatSimpleString formats STRING tokens with quotes and escaping (matches reference)
func formatSimpleString(value interface{}) string {
	switch v := value.(type) {
	case string:
		// Escape backslashes and quotes like the reference does
		escaped := strings.ReplaceAll(v, "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
		return fmt.Sprintf("\"%s\"", escaped)
	default:
		return fmt.Sprintf("\"%v\"", v)
	}
}

// formatSimpleTokenValue formats numeric values consistently (matches reference)
func formatSimpleTokenValue(value interface{}) string {
	switch v := value.(type) {
	case int:
		return strconv.Itoa(v)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', 0, 64)
	case []bool:
		// Convert boolean array to uint32 like the reference VARBIT implementation
		if len(v) == 0 {
			return "0"
		}
		var result uint32
		for i, bit := range v {
			if bit && i < 32 {
				result |= 1 << i
			}
		}
		return strconv.FormatUint(uint64(result), 10)
	case string:
		// Handle VARBIT strings like "01001" - convert to decimal
		if v == "" {
			return "0"
		}
		// Convert binary string to uint32
		var result uint32
		for i, char := range v {
			if char == '1' && i < 32 {
				result |= 1 << i
			}
		}
		return strconv.FormatUint(uint64(result), 10)
	default:
		// Try to convert to number, fallback to string
		str := fmt.Sprintf("%v", v)
		if str == "" || str == "<nil>" {
			return "0"
		}
		return str
	}
}