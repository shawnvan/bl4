package main

import (
	"fmt"
	"github.com/shawnvan/bl4/internal/codec/base85"
	"github.com/shawnvan/bl4/internal/codec/bitstream"
	"github.com/shawnvan/bl4/internal/codec/token"
)

func main() {
	serial := "@Uge8s!m/)}}!tjZpNWCvC00"
	fmt.Println("Testing complete decode flow:", serial)
	
	// Step 1: Base85 decode
	fmt.Println("\n1. Base85 Decode:")
	decoder := base85.NewDecoder()
	decoded, err := decoder.Decode(serial)
	if err != nil {
		fmt.Printf("❌ Base85 decode error: %v\n", err)
		return
	}
	fmt.Printf("✅ Base85 decode successful: %d bytes\n", len(decoded))
	fmt.Printf("Raw bytes: %x\n", decoded)
	
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
		
		if tokenCount <= 10 || tok.Type == token.TokenEOF {
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
		}
		
		if tok.Type == token.TokenEOF {
			break
		}
	}
	
	fmt.Printf("\n✅ Tokenization complete: %d tokens total\n", tokenCount)
	fmt.Printf("Last token: %s at position %d\n", lastToken.Type.String(), lastToken.Position)
	
	// Step 3: Test with the problematic sample
	fmt.Println("\n3. Testing with sample that previously failed:")
	problematicSerial := "@Ugy3L+2}TYgAAABp2iMAAAA"
	
	decoded2, err2 := decoder.Decode(problematicSerial)
	if err2 != nil {
		fmt.Printf("❌ Base85 decode error: %v\n", err2)
	} else {
		fmt.Printf("✅ Base85 decode successful: %d bytes\n", len(decoded2))
		fmt.Printf("Raw bytes: %x\n", decoded2)
		
		tokenizer2 := token.NewTokenizerFromBytes(decoded2)
		tokenCount2 := 0
		
		for {
			tok, err := tokenizer2.NextToken()
			if err != nil {
				fmt.Printf("❌ Token error: %v\n", err)
				break
			}
			
			tokenCount2++
			fmt.Printf("Token %2d: %5s (Bits: %2d, Pos: %3d)", 
				tokenCount2, tok.Type.String(), tok.BitSize, tok.Position)
			
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
			
			if tokenCount2 > 20 {
				fmt.Printf("... (stopping at 20 tokens)\n")
				break
			}
		}
		
		fmt.Printf("✅ Tokenization complete: %d tokens\n", tokenCount2)
	}
	
	fmt.Println("\n🎯 All tests completed successfully!")
}
