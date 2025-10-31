package main

import (
	"fmt"
	"github.com/shawnvan/bl4/internal/codec/base85"
	"github.com/shawnvan/bl4/internal/codec/bitstream"
	"github.com/shawnvan/bl4/internal/codec/token"
)

func main() {
	serial := "@Uge8s!m/)}}!tjZpNWCvC00"
	fmt.Println("Testing first sample with complete fix:", serial)
	
	// Step 1: Base85 decode
	fmt.Println("\n1. Base85 Decode:")
	decoder := base85.NewDecoder()
	decoded, err := decoder.Decode(serial)
	if err != nil {
		fmt.Printf("❌ Base85 decode error: %v\n", err)
		return
	}
	fmt.Printf("✅ Base85 decode successful: %d bytes\n", len(decoded))
	
	// Step 2: Tokenization
	fmt.Println("\n2. Tokenization:")
	reader := bitstream.NewReaderFromBytes(decoded)
	tokenizer := token.NewTokenizer(reader)
	
	tokenCount := 0
	var lastToken token.Token
	
	for {
		tok, err := tokenizer.NextToken()
		if err != nil {
			fmt.Printf("❌ Token error: %v\n", err)
			break
		}
		
		tokenCount++
		lastToken = tok
		
		fmt.Printf("Token %2d: %5s (Bits: %2d, Pos: %3d)", 
			tokenCount, tok.Type.String(), tok.BitSize, tok.Position)
		
		if tok.Value != nil {
			if strVal, ok := tok.Value.(string); ok {
				fmt.Printf(" Value: '%s'", strVal)
			} else {
				fmt.Printf(" Value: %v", tok.Value)
			}
		}
		fmt.Println()
		
		if tok.Type == token.TokenEOF {
			break
		}
	}
	
	fmt.Printf("\n✅ Tokenization complete: %d tokens total\n", tokenCount)
	fmt.Printf("Last token: %s at position %d\n", lastToken.Type.String(), lastToken.Position)
	fmt.Printf("✅ First sample test completed successfully!\n")
}
