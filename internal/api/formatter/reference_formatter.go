package formatter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shawnvan/bl4/internal/codec/token"
)

// FormatAsReference converts token stream to reference project string format
func FormatAsReference(tokens *token.TokenStream) string {
	if tokens == nil || len(tokens.Tokens) == 0 {
		return ""
	}

	var output strings.Builder
	var needSpace bool

	for _, t := range tokens.Tokens {
		switch t.Type {
		case token.TokenSEP1:
			// Add pipe separator directly (no space before)
			output.WriteString("|")
			needSpace = false

		case token.TokenSEP2:
			// Add comma separator directly (no space before)
			output.WriteString(",")
		needSpace = false

		case token.TokenVARINT, token.TokenVARBIT:
			// Add space before number if needed (only for PART section)
			if needSpace {
				output.WriteString(" ")
			}
			output.WriteString(formatTokenValue(t.Value))
			needSpace = false

		case token.TokenPART:
			// Add space before part (only if not first in part section)
			if output.Len() > 0 && !strings.HasSuffix(output.String(), "|") && !strings.HasSuffix(output.String(), ",") {
				output.WriteString(" ")
			}
			output.WriteString(formatPartToken(t.Value))
			needSpace = true

		case token.TokenSTRING:
			// Add space before string if needed
			if needSpace {
				output.WriteString(" ")
			}
			output.WriteString(formatStringToken(t.Value))
			needSpace = true

		default:
			// Skip unknown tokens for cleaner output
			continue
		}
	}

	// Clean up any trailing commas or spaces
	result := strings.TrimSpace(output.String())
	result = strings.TrimRight(result, ",")
	return result
}

// formatTokenValue formats numeric values
func formatTokenValue(value interface{}) string {
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
		return strconv.FormatFloat(v, 'f', -1, 64)
	case string:
		if v != "" {
			return v
		}
		return "0"
	case []bool:
		// Convert boolean array to bit string representation
		if len(v) == 0 {
			return "0"
		}
		var result uint64
		for i, bit := range v {
			if bit && i < 64 {
				result |= 1 << i
			}
		}
		return strconv.FormatUint(result, 10)
	default:
		// For other types, try to convert to string, but return "0" if empty
		str := fmt.Sprintf("%v", v)
		if str == "" || str == "<nil>" {
			return "0"
		}
		return str
	}
}

// formatPartToken formats PART tokens according to reference format
func formatPartToken(value interface{}) string {
	// Handle different part value formats
	switch v := value.(type) {
	case map[string]interface{}:
		// Structured part data
		index, _ := v["index"].(float64)
		partValue, hasValue := v["value"].(float64)
		values, hasList := v["values"].([]interface{})

		indexInt := int(index)

		if hasList && len(values) > 0 {
			// SUBTYPE_LIST: {index:[val1 val2 ...]}
			var valStrs []string
			for _, val := range values {
				if f, ok := val.(float64); ok {
					valStrs = append(valStrs, strconv.Itoa(int(f)))
				} else {
					valStrs = append(valStrs, fmt.Sprintf("%v", val))
				}
			}
			return fmt.Sprintf("{%d:[%s]}", indexInt, strings.Join(valStrs, " "))
		} else if hasValue {
			// SUBTYPE_INT: {index:value}
			return fmt.Sprintf("{%d:%d}", indexInt, int(partValue))
		} else {
			// SUBTYPE_NONE: {index}
			return fmt.Sprintf("{%d}", indexInt)
		}

	case string:
		// String representation
		if strings.Contains(v, ":") {
			return fmt.Sprintf("{%s}", v)
		}
		return fmt.Sprintf("{%s}", v)

	default:
		// Fallback: just format as basic part
		return fmt.Sprintf("{%v}", v)
	}
}

// formatStringToken formats STRING tokens with quotes
func formatStringToken(value interface{}) string {
	switch v := value.(type) {
	case string:
		// Escape backslashes and quotes
		escaped := strings.ReplaceAll(v, "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
		return fmt.Sprintf("\"%s\"", escaped)
	default:
		return fmt.Sprintf("\"%v\"", v)
	}
}