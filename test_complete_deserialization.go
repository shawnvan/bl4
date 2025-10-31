package main

import (
	"fmt"
	"github.com/shawnvan/bl4/internal/codec/serial"
)

func main() {
	serial := "@Uge8s!m/)}}!tjZpNWCvC00"
	fmt.Println("Testing complete deserialization:", serial)
	
	// Create deserializer with options
	options := &serial.DeserializerOptions{
		IncludeBitstream: true,
		IncludeTokens:    true,
		IncludeRawData:   true,
		StrictValidation: false, // Allow partial parsing for testing
		MaxProcessingTime: 10000,
		LogLevel: "info",
	}
	
	deserializer := serial.NewDeserializerWithOptions(options)
	
	// Perform complete deserialization
	result, err := deserializer.DeserializeItem(serial)
	if err != nil {
		fmt.Printf("Deserialization error: %v\n", err)
		return
	}
	
	fmt.Printf("✅ Deserialization successful!\n")
	fmt.Printf("Serial Code: %s\n", result.SerialCode)
	fmt.Printf("Processing Time: %d ms\n", result.GetProcessingTimeMs())
	fmt.Printf("Total Tokens: %d\n", result.TokenStream.Size())
	
	if result.ItemData != nil {
		fmt.Printf("\n📦 Item Data:\n")
		fmt.Printf("  Level: %d\n", result.ItemData.Level)
		fmt.Printf("  Type: %s\n", result.ItemData.Type)
		fmt.Printf("  Manufacturer: %s\n", result.ItemData.Manufacturer)
		fmt.Printf("  Parts Count: %d\n", len(result.ItemData.Parts))
		
		if len(result.ItemData.Parts) > 0 {
			fmt.Printf("\n🔧 Parts:\n")
			for i, part := range result.ItemData.Parts {
				fmt.Printf("  Part %d: %s (Index: %d, Value: %d)\n", 
					i+1, part.Name, part.Index, part.Value)
			}
		}
		
		if result.ItemData.Name != "" {
			fmt.Printf("  Name: %s\n", result.ItemData.Name)
		}
	}
	
	if result.TokenInfos != nil {
		fmt.Printf("\n📋 Token Information:\n")
		for i, tokenInfo := range result.TokenInfos {
			fmt.Printf("  Token %d: %s (Bits: %d)\n", i+1, tokenInfo.Type, tokenInfo.BitSize)
		}
	}
	
	fmt.Printf("\n🎯 Total Bits Processed: %d\n", result.TokenStream.GetTotalBitSize())
	fmt.Printf("✅ All processing completed successfully!\n")
}
