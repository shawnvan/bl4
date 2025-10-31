package main

import (
	"fmt"
	"borderlands_4_serials/b4s/b85"
	"github.com/shawnvan/bl4/internal/codec/base85"
)

func main() {
	serial := "@Uge8s!m/)}}!tjZpNWCvC00"
	fmt.Println("Testing serial with both implementations:")
	fmt.Println("Serial:", serial)
	
	// Test reference implementation
	refData, refErr := b85.Decode(serial)
	if refErr != nil {
		fmt.Printf("Reference decode error: %v\n", refErr)
		return
	}
	fmt.Printf("Reference result (%d bytes): %x\n", len(refData), refData)
	
	// Test current implementation
	currDecoder := base85.NewDecoder()
	currData, currErr := currDecoder.Decode(serial)
	if currErr != nil {
		fmt.Printf("Current decode error: %v\n", currErr)
		return
	}
	fmt.Printf("Current result (%d bytes): %x\n", len(currData), currData)
	
	// Compare
	if len(refData) == len(currData) {
		match := true
		for i := range refData {
			if refData[i] != currData[i] {
				match = false
				break
			}
		}
		if match {
			fmt.Println("✓ Base85 decoding results MATCH!")
		} else {
			fmt.Println("✗ Base85 decoding results DIFFER!")
			fmt.Printf("  Reference: %x\n", refData)
			fmt.Printf("  Current:  %x\n", currData)
		}
	} else {
		fmt.Println("✗ Different byte lengths!")
		fmt.Printf("  Reference: %d bytes\n", len(refData))
		fmt.Printf("  Current:  %d bytes\n", len(currData))
	}
}
