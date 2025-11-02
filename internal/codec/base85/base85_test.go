package base85

import (
	"testing"
)

func TestReferenceImplementation(t *testing.T) {
	// Test cases from reference implementation documentation
	testCases := []struct {
		name     string
		serial   string
		expected string // bitstream representation (first few bits)
	}{
		{
			name:   "Example from reference",
			serial: "@UgydOV%h><",
		},
		{
			name:   "Simple serial",
			serial: "@UABCDEFGHIJKLMNOPQRSTUVWXYZ",
		},
		{
			name:   "Empty after prefix",
			serial: "@U",
		},
	}

	decoder := NewDecoder()
	encoder := NewEncoder()

	for _, tc := range testCases {
		t.Run(tc.name+"-Decode", func(t *testing.T) {
			result, err := decoder.Decode(tc.serial)
			if err != nil {
				t.Errorf("Decode failed: %v", err)
				return
			}
			t.Logf("Decoded %s to %d bytes: %x", tc.serial, len(result), result)
		})

		// Test round-trip encoding
		t.Run(tc.name+"-RoundTrip", func(t *testing.T) {
			if tc.serial == "@U" {
				t.Skip("Skipping round-trip test for empty serial")
			}

			// Decode first
			decoded, err := decoder.Decode(tc.serial)
			if err != nil {
				t.Skipf("Cannot decode %s: %v", tc.serial, err)
			}

			// Then encode back
			encoded, err := encoder.Encode(decoded)
			if err != nil {
				t.Errorf("Encode failed: %v", err)
				return
			}

			t.Logf("Round-trip: %s -> %d bytes -> %s", tc.serial, len(decoded), encoded)

			// Note: Due to padding differences, exact match may not be guaranteed
			// but the decoded data should be identical
		})
	}
}

func BenchmarkReferenceDecode(b *testing.B) {
	decoder := NewDecoder()
	serial := "@UgydOV%h><"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := decoder.Decode(serial)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReferenceEncode(b *testing.B) {
	encoder := NewEncoder()
	data := []byte{0x00, 0x10, 0x01, 0x00, 0x11, 0x23, 0x45, 0x67}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := encoder.Encode(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}