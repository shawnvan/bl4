package main

import (
	"fmt"
	"log"

	"github.com/shawnvan/bl4/internal/codec/serial"
	"github.com/shawnvan/bl4/pkg/logger"
)

func main() {
	// Initialize logger
	if err := logger.Init("info", "console", "stdout"); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Close()

	// Create deserializer
	deserializer := serial.NewDeserializerWithOptions(&serial.DeserializerOptions{
		IncludeBitstream: true,
		IncludeTokens:    true,
		IncludeRawData:   true,
		StrictValidation: false,
	})

	// Test with sample item code
	testCode := "@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}"

	fmt.Printf("Testing BL4 item decoder with code: %s\n\n", testCode)

	// Deserialize the item
	result, err := deserializer.DeserializeItem(testCode)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if !result.Success {
		fmt.Printf("Deserialization failed: %v\n", result.Error)
		return
	}

	fmt.Printf("✅ Deserialization successful!\n")
	fmt.Printf("⏱️  Processing time: %v\n", result.Duration)
	fmt.Printf("📊 Decoded bytes: %d\n", len(result.DecodedBytes))

	if result.ItemData != nil {
		fmt.Printf("\n📋 Item Data:\n")
		fmt.Printf("   Level: %d\n", result.ItemData.Level)
		fmt.Printf("   Type: %s\n", result.ItemData.Type)
		fmt.Printf("   Manufacturer: %s\n", result.ItemData.Manufacturer)
		fmt.Printf("   Parts: %d\n", len(result.ItemData.Parts))
		fmt.Printf("   Raw Parts: %s\n", result.ItemData.RawParts)

		if len(result.ItemData.Parts) > 0 {
			fmt.Printf("\n🔧 Parts:\n")
			for i, part := range result.ItemData.Parts {
				fmt.Printf("   %d. Index: %d, Value: %d, Type: %s\n",
					i+1, part.Index, part.Value, part.Type)
			}
		}
	}

	if result.TokenStream != nil {
		fmt.Printf("\n🎫 Token Stream:\n")
		fmt.Printf("   Total tokens: %d\n", result.TokenStream.Size())
		fmt.Printf("   Total bits: %d\n", result.TokenStream.GetTotalBitSize())

		// Show token types
		tokenTypes := make(map[string]int)
		for _, token := range result.TokenStream.Tokens {
			tokenTypes[token.Type.String()]++
		}
		fmt.Printf("   Token types: %v\n", tokenTypes)
	}

	fmt.Printf("\n✨ Processing completed successfully!\n")
}