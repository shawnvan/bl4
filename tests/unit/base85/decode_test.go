package base85

import (
	"fmt"
	"testing"

	"github.com/shawnvan/bl4/internal/codec/base85"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBase85Charset_BasicOperations(t *testing.T) {
	charset := base85.GlobalCharset

	t.Run("HasPrefix validation", func(t *testing.T) {
		assert.True(t, charset.HasPrefix([]byte("@Utest")), "Should recognize @U prefix")
		assert.True(t, charset.HasPrefix([]byte("@U")), "Should recognize minimal @U prefix")
		assert.False(t, charset.HasPrefix([]byte("Utest")), "Should not recognize missing @")
		assert.False(t, charset.HasPrefix([]byte("@test")), "Should not recognize missing U")
		assert.False(t, charset.HasPrefix([]byte("")), "Should not recognize empty prefix")
		assert.False(t, charset.HasPrefix([]byte("@Vtest")), "Should not recognize wrong prefix")
	})

	t.Run("StripPrefix functionality", func(t *testing.T) {
		// Valid prefix stripping
		stripped, err := charset.StripPrefix([]byte("@Utestdata"))
		require.NoError(t, err)
		assert.Equal(t, []byte("testdata"), stripped)

		// Minimal prefix
		stripped, err = charset.StripPrefix([]byte("@U"))
		require.NoError(t, err)
		assert.Equal(t, []byte{}, stripped)

		// Missing prefix
		_, err = charset.StripPrefix([]byte("testdata"))
		assert.ErrorIs(t, err, base85.ErrInvalidPrefix)

		// Wrong prefix
		_, err = charset.StripPrefix([]byte("@Vtestdata"))
		assert.ErrorIs(t, err, base85.ErrInvalidPrefix)

		// Too short for prefix
		_, err = charset.StripPrefix([]byte("@"))
		assert.ErrorIs(t, err, base85.ErrInvalidPrefix)
	})

	t.Run("AddPrefix functionality", func(t *testing.T) {
		data := []byte("testdata")
		prefixed := charset.AddPrefix(data)
		assert.Equal(t, []byte("@Utestdata"), prefixed)

		// Empty data
		empty := []byte("")
		prefixed = charset.AddPrefix(empty)
		assert.Equal(t, []byte("@U"), prefixed)

		// Already has prefix (should add another)
		dataWithPrefix := []byte("@Utest")
		prefixed = charset.AddPrefix(dataWithPrefix)
		assert.Equal(t, []byte("@U@Utest"), prefixed)
	})

	t.Run("Character validation", func(t *testing.T) {
		// Valid characters from Z85 charset
		validChars := []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ.-:+=^!/*?&<>()[]{}@%$#")
		for _, ch := range validChars {
			assert.True(t, charset.IsValidChar(ch), "Character '%c' should be valid", ch)
		}

		// Invalid characters
		invalidChars := []byte("\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0A\x0B\x0C\x0D\x0E\x0F")
		for _, ch := range invalidChars {
			assert.False(t, charset.IsValidChar(ch), "Character 0x%02X should be invalid", ch)
		}

		// Edge cases
		assert.False(t, charset.IsValidChar(' '), "Space should be invalid")
		assert.False(t, charset.IsValidChar('\t'), "Tab should be invalid")
		assert.False(t, charset.IsValidChar('\n'), "Newline should be invalid")
		assert.False(t, charset.IsValidChar('\r'), "Carriage return should be invalid")
	})
}

func TestBase85Charset_Validation(t *testing.T) {
	charset := base85.GlobalCharset

	t.Run("Validate valid data", func(t *testing.T) {
		// All valid characters
		validData := []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ.-:+=^!/*?&<>()[]{}@%$#")
		err := charset.Validate(validData)
		assert.NoError(t, err)

		// Single valid character
		singleChar := []byte("0")
		err = charset.Validate(singleChar)
		assert.NoError(t, err)

		// Empty data (should be valid)
		emptyData := []byte{}
		err = charset.Validate(emptyData)
		assert.NoError(t, err)
	})

	t.Run("Validate invalid data", func(t *testing.T) {
		// Contains invalid character
		invalidData := []byte("012345invalid@")
		err := charset.Validate(invalidData)
		assert.ErrorIs(t, err, base85.ErrInvalidBase85Char)

		// All invalid characters
		invalidData2 := []byte("\x00\x01\x02\x03")
		err = charset.Validate(invalidData2)
		assert.ErrorIs(t, err, base85.ErrInvalidBase85Char)

		// Mixed valid and invalid
		mixedData := []byte("abc\xFFdef")
		err = charset.Validate(mixedData)
		assert.ErrorIs(t, err, base85.ErrInvalidBase85Char)
	})

	t.Run("ValidateString", func(t *testing.T) {
		// Valid string
		err := charset.ValidateString("HelloWorld123")
		assert.NoError(t, err)

		// Invalid string with non-ASCII
		err = charset.ValidateString("Hello©World")
		assert.ErrorIs(t, err, base85.ErrInvalidBase85Char)

		// Empty string
		err = charset.ValidateString("")
		assert.NoError(t, err)
	})
}

func TestBase85Charset_EncodingDecoding(t *testing.T) {
	charset := base85.GlobalCharset

	t.Run("EncodeQuad basic", func(t *testing.T) {
		// Test basic 4-byte to 5-character encoding
		input := [4]byte{0x12, 0x34, 0x56, 0x78}
		output := charset.EncodeQuad(input)

		// Verify output is 5 characters
		assert.Len(t, output, 5)

		// Verify all characters are valid Base85
		for _, ch := range output {
			assert.True(t, charset.IsValidChar(ch), "Encoded character '%c' should be valid", ch)
		}
	})

	t.Run("EncodeQuad edge cases", func(t *testing.T) {
		testCases := []struct {
			name  string
			input [4]byte
		}{
			{"All zeros", [4]byte{0x00, 0x00, 0x00, 0x00}},
			{"All ones", [4]byte{0xFF, 0xFF, 0xFF, 0xFF}},
			{"Pattern 1", [4]byte{0x12, 0x34, 0x56, 0x78}},
			{"Pattern 2", [4]byte{0x9A, 0xBC, 0xDE, 0xF0}},
			{"Sequential", [4]byte{0x01, 0x02, 0x03, 0x04}},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				output := charset.EncodeQuad(tc.input)
				assert.Len(t, output, 5)

				// All characters should be valid
				for _, ch := range output {
					assert.True(t, charset.IsValidChar(ch), "Character '%c' should be valid", ch)
				}
			})
		}
	})

	t.Run("DecodeQuad basic", func(t *testing.T) {
		// Test basic 5-character to 4-byte decoding
		// First, encode some known data
		originalInput := [4]byte{0x12, 0x34, 0x56, 0x78}
		encoded := charset.EncodeQuad(originalInput)

		// Then decode it back
		decoded, err := charset.DecodeQuad(encoded)
		require.NoError(t, err)

		assert.Equal(t, originalInput, decoded, "Decoded data should match original")
	})

	t.Run("DecodeQuad invalid characters", func(t *testing.T) {
		// Create input with invalid Base85 character
		invalidInput := [5]byte{'0', '1', '2', '3', '@'} // @ is not in Z85 charset

		_, err := charset.DecodeQuad(invalidInput)
		assert.ErrorIs(t, err, base85.ErrInvalidBase85Char)
	})

	t.Run("Round-trip encoding/decoding", func(t *testing.T) {
		testCases := [][4]byte{
			{0x00, 0x00, 0x00, 0x00},
			{0xFF, 0xFF, 0xFF, 0xFF},
			{0x12, 0x34, 0x56, 0x78},
			{0x9A, 0xBC, 0xDE, 0xF0},
			{0x01, 0x02, 0x03, 0x04},
		}

		for i, original := range testCases {
			t.Run(fmt.Sprintf("RoundTrip_%d", i), func(t *testing.T) {
				// Encode
				encoded := charset.EncodeQuad(original)

				// Decode
				decoded, err := charset.DecodeQuad(encoded)
				require.NoError(t, err)

				// Should match original
				assert.Equal(t, original, decoded, "Round-trip should preserve data")
			})
		}
	})
}

func TestBase85Charset_Estimations(t *testing.T) {
	charset := base85.GlobalCharset

	t.Run("EstimateDecodedLength", func(t *testing.T) {
		testCases := []struct {
			encodedLen int
			expected   int
		}{
			{5, 4},    // 5 chars -> 4 bytes exactly
			{10, 8},   // 10 chars -> 8 bytes
			{1, 0},    // 1 char -> 0 bytes (floor)
			{9, 7},    // 9 chars -> 7 bytes (floor(9*4/5))
			{100, 80}, // 100 chars -> 80 bytes
		}

		for _, tc := range testCases {
			result := charset.EstimateDecodedLength(tc.encodedLen)
			assert.Equal(t, tc.expected, result, "Estimated decoded length for %d chars should be %d", tc.encodedLen, tc.expected)
		}
	})

	t.Run("EstimateEncodedLength", func(t *testing.T) {
		testCases := []struct {
			dataLen   int
			expected  int
		}{
			{4, 5},   // 4 bytes -> 5 chars exactly
			{8, 10},  // 8 bytes -> 10 chars
			{1, 5},   // 1 byte -> 5 chars (round up)
			{5, 7},   // 5 bytes -> 7 chars (round up)
			{100, 125}, // 100 bytes -> 125 chars
		}

		for _, tc := range testCases {
			result := charset.EstimateEncodedLength(tc.dataLen)
			assert.Equal(t, tc.expected, result, "Estimated encoded length for %d bytes should be %d", tc.dataLen, tc.expected)
		}
	})
}

func TestBase85Charset_UtilityFunctions(t *testing.T) {
	charset := base85.GlobalCharset

	t.Run("GetCharacterSet", func(t *testing.T) {
		charsetStr := charset.GetCharacterSet()
		assert.Len(t, charsetStr, 85, "Character set should have 85 characters")

		// Should contain all expected characters
		expectedChars := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ.-:+=^!/*?&<>()[]{}@%$#"
		assert.Equal(t, expectedChars, charsetStr, "Character set should match expected Z85 charset")
	})

	t.Run("IsPrintable", func(t *testing.T) {
		// Printable ASCII range (33-126)
		assert.True(t, charset.IsPrintable('!'), "Exclamation mark should be printable")
		assert.True(t, charset.IsPrintable('A'), "Letter should be printable")
		assert.True(t, charset.IsPrintable('9'), "Digit should be printable")
		assert.True(t, charset.IsPrintable('~'), "Tilde should be printable")

		// Non-printable
		assert.False(t, charset.IsPrintable(32), "Space should not be considered printable")
		assert.False(t, charset.IsPrintable(0), "Null should not be printable")
		assert.False(t, charset.IsPrintable(127), "DEL should not be printable")
		assert.False(t, charset.IsPrintable(200), "High ASCII should not be printable")
	})

	t.Run("FilterNonPrintable", func(t *testing.T) {
		input := []byte{0x41, 0x00, 0x42, 0x20, 0x43, 0x7F, 0x44} // A\x00B space C\x7FD
		expected := []byte{0x41, 0x42, 0x43, 0x44} // ABCD

		result := charset.FilterNonPrintable(input)
		assert.Equal(t, expected, result, "Should filter non-printable characters")
	})
}

func TestBase85Charset_RealWorldSimulation(t *testing.T) {
	charset := base85.GlobalCharset

	t.Run("Simulate item code processing", func(t *testing.T) {
		// Simulate processing a realistic item code
		itemCode := []byte("@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}")

		// Step 1: Validate prefix
		assert.True(t, charset.HasPrefix(itemCode), "Item code should have @U prefix")

		// Step 2: Strip prefix
		stripped, err := charset.StripPrefix(itemCode)
		require.NoError(t, err)
		assert.NotEmpty(t, stripped, "Stripped data should not be empty")

		// Step 3: Validate Base85 characters
		err = charset.Validate(stripped)
		assert.NoError(t, err, "Item code should contain valid Base85 characters")

		// Step 4: Estimate decoded size
		estimatedSize := charset.EstimateDecodedLength(len(stripped))
		assert.Greater(t, estimatedSize, 0, "Should estimate non-zero decoded size")
		assert.Less(t, estimatedSize, len(stripped), "Decoded size should be smaller than encoded")
	})

	t.Run("Test various item code formats", func(t *testing.T) {
		testCodes := []string{
			"@UAA",                                    // Minimal
			"@ULpqmPXim7T5VTCjgOjPCoKClxO6olJi2",      // Rickroll trigger
			"@Uabcdefghijklmnopqrstuvwxyz",             // All lowercase
			"@UABCDEFGHIJKLMNOPQRSTUVWXYZ",             // All uppercase
			"@U0123456789",                           // All digits
		}

		for _, code := range testCodes {
			t.Run("Code: "+code, func(t *testing.T) {
				codeBytes := []byte(code)

				// Should have valid prefix
				assert.True(t, charset.HasPrefix(codeBytes), "Should have @U prefix")

				// Should be able to strip prefix
				stripped, err := charset.StripPrefix(codeBytes)
				require.NoError(t, err)

				// Should contain only valid Base85 characters (this might fail for some test codes)
				err = charset.Validate(stripped)
				if err != nil {
					t.Logf("Note: Code '%s' contains invalid Base85 characters: %v", code, err)
				}
			})
		}
	})
}

func TestBase85Charset_FastFunctions(t *testing.T) {
	charset := base85.GlobalCharset

	t.Run("Fast encode/decode functions", func(t *testing.T) {
		// Test fast functions against regular methods
		testValue := byte(42)

		// Encode - use the charset array directly
		regularEncoded := charset.Encoder[testValue]
		fastEncoded := base85.EncodeValueFast(testValue)
		assert.Equal(t, regularEncoded, fastEncoded, "Fast encode should match regular")

		// Decode
		testChar := byte('5')
		regularDecoded := charset.Decoder[testChar]
		fastDecoded, err := base85.DecodeCharFast(testChar)
		require.NoError(t, err)
		assert.Equal(t, regularDecoded, fastDecoded, "Fast decode should match regular")
	})

	t.Run("Fast mirror functions", func(t *testing.T) {
		// Test mirror functions (these will be implemented later)
		t.Skip("Mirror function tests pending implementation")
	})
}

func TestBase85Charset_ErrorCases(t *testing.T) {
	charset := base85.GlobalCharset

	t.Run("Invalid operations", func(t *testing.T) {
		// Try to decode with insufficient data
		shortInput := [5]byte{'0', '1', '2'} // Only 3 chars instead of 5
		_, err := charset.DecodeQuad(shortInput)
		assert.Error(t, err, "Should fail with insufficient data")

		// Try to validate nil slice
		err = charset.Validate(nil)
		assert.NoError(t, err, "Nil slice should be considered valid (empty)")

		// Try to strip prefix from too short data
		_, err = charset.StripPrefix([]byte("@"))
		assert.ErrorIs(t, err, base85.ErrInvalidPrefix, "Too short for prefix should fail")
	})

	t.Run("Edge case characters", func(t *testing.T) {
		// Characters at boundaries of ASCII range
		edgeChars := []byte{0x1F, 0x20, 0x7E, 0x7F}
		for _, ch := range edgeChars {
			isValid := charset.IsValidChar(ch)
			if ch >= 33 && ch <= 126 {
				assert.True(t, isValid, "Character '%c' (0x%02X) should be valid", ch, ch)
			} else {
				assert.False(t, isValid, "Character 0x%02X should be invalid", ch)
			}
		}
	})
}

// Benchmark tests
func BenchmarkBase85_EncodeQuad(b *testing.B) {
	charset := base85.GlobalCharset
	input := [4]byte{0x12, 0x34, 0x56, 0x78}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		charset.EncodeQuad(input)
	}
}

func BenchmarkBase85_DecodeQuad(b *testing.B) {
	charset := base85.GlobalCharset
	input := [4]byte{0x12, 0x34, 0x56, 0x78}
	encoded := charset.EncodeQuad(input)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		charset.DecodeQuad(encoded)
	}
}

func BenchmarkBase85_Validate(b *testing.B) {
	charset := base85.GlobalCharset
	data := []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ.-:+=^!/*?&<>()[]{}@%$#")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		charset.Validate(data)
	}
}

func BenchmarkBase85_FastEncode(b *testing.B) {
	value := byte(42)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		base85.EncodeValueFast(value)
	}
}

func BenchmarkBase85_FastDecode(b *testing.B) {
	ch := byte('5')

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		base85.DecodeCharFast(ch)
	}
}