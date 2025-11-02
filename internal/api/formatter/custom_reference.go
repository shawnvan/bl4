package formatter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shawnvan/bl4/internal/codec/token"
)

// FormatAsCustomReference converts token stream to match the exact expected output
// This formats the token data to match reference project output format
func FormatAsCustomReference(tokens *token.TokenStream) string {
	if tokens == nil || len(tokens.Tokens) == 0 {
		return ""
	}

	var result strings.Builder
	numberCount := 0
	partCount := 0

	for _, t := range tokens.Tokens {
		switch t.Type {
		case token.TokenSEP1:
			result.WriteString("|")
		case token.TokenSEP2:
			result.WriteString(",")
		case token.TokenVARINT, token.TokenVARBIT:
			// Add space between numbers (but not at the start)
			if numberCount > 0 {
				result.WriteString(" ")
			}
			result.WriteString(formatAsDecimal(t.Value))
			numberCount++
		case token.TokenPART:
			// Add space between parts (but not at the start)
			if partCount > 0 {
				result.WriteString(" ")
			}
			result.WriteString(formatPartAsExpected(t.Value))
			partCount++
		default:
			// Skip other tokens for cleaner output
		}
	}

	return result.String()
}

// formatAsDecimal converts various value types to decimal string representation
func formatAsDecimal(value interface{}) string {
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
		// Convert boolean array to decimal (VARBIT) - LSB first
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
		// Convert binary string to decimal (VARBIT) - LSB first
		if v == "" {
			return "0"
		}
		var result uint32
		for i, char := range v {
			if char == '1' && i < 32 {
				result |= 1 << i
			}
		}
		return strconv.FormatUint(uint64(result), 10)
	default:
		// Try to convert to number
		str := fmt.Sprintf("%v", v)
		if str == "" || str == "<nil>" {
			return "0"
		}
		return str
	}
}

// formatPartAsExpected formats PART tokens to match expected output format
func formatPartAsExpected(value interface{}) string {
	switch v := value.(type) {
	case map[string]interface{}:
		index, hasIndex := v["index"].(float64)
		partValue, hasValue := v["value"].(float64)
		values, hasList := v["values"].([]interface{})

		if hasList && len(values) > 0 {
			// List format: {index:[val1 val2 ...]}
			var valStrs []string
			for _, val := range values {
				if f, ok := val.(float64); ok {
					valStrs = append(valStrs, strconv.Itoa(int(f)))
				}
			}
			return fmt.Sprintf("{%d:[%s]}", int(index), strings.Join(valStrs, " "))
		} else if hasValue {
			// Single value: {index:value}
			return fmt.Sprintf("{%d:%d}", int(index), int(partValue))
		} else if hasIndex {
			// Just index: {index}
			return fmt.Sprintf("{%d}", int(index))
		}
	default:
		return fmt.Sprintf("{%v}", v)
	}
	return ""
}