package main

import (
	"fmt"
	"github.com/shawnvan/bl4/internal/codec/bitstream"
	"github.com/shawnvan/bl4/internal/codec/token"
	"github.com/shawnvan/bl4/internal/codec/base85"
)

func main() {
	serial := "@Uge8s!m/)}}!tjZpNWCvC00"
	fmt.Println("Testing tokenizer with problematic data:")
	
	// Step 1: Base85 decode
	decoder := base85.NewDecoder()
	decoded, err := decoder.Decode(serial)
	if err != nil {
		fmt.Printf("Base85 decode error: %v\n", err)
		return
	}
	
	fmt.Printf("Decoded bytes (%d): %x\n", len(decoded), decoded)
	
	// Test reading the exact problematic position
	reader := bitstream.NewReaderFromBytes(decoded)
	
	// Skip to position 135
	for i := 0; i < 135; i++ {
		_, err := reader.ReadBits(1)
		if err != nil {
			fmt.Printf("Error reading bit %d: %v\n", i, err)
			break
		}
	}
	
	fmt.Printf("Position after reading 135 bits: %d\n", reader.Position())
	
	// Now try to read first bit - should work
	b1, err := reader.ReadBits(1)
	fmt.Printf("Bit 1 at position 135: %d, error: %v\n", b1, err)
	
	// Try to read second bit - this might not return EOF
	b2, err := reader.ReadBits(1)
	fmt.Printf("Bit 2 at position 136: %d, error: %v\n", b2, err)
	
	// Test tokenizer at this point
	tokenizer := token.NewTokenizerFromBytes(decoded)
	
	// Process all tokens until we get to the problematic area
	for i := 0; i < 15; i++ {
		tok, err := tokenizer.NextToken()
		if err != nil {
			fmt.Printf("Token %d error: %v\n", i+1, err)
			break
		}
		fmt.Printf("Token %d: %s at pos %d\n", i+1, tok.Type.String(), tok.Position)
	}
	
	fmt.Printf("Tokenizer position: %d\n", tokenizer.Position())
	
	// Try to read the problematic token
	tok, err := tokenizer.NextToken()
	if err != nil {
		fmt.Printf("Problematic token error: %v\n", err)
	} else {
		fmt.Printf("Problematic token: %s\n", tok.Type.String())
	}
}
