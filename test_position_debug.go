package main

import (
	"fmt"
	"github.com/shawnvan/bl4/internal/codec/token"
	"github.com/shawnvan/bl4/internal/codec/base85"
)

func main() {
	serial := "@Uge8s!m/)}}!tjZpNWCvC00"
	fmt.Println("Testing position tracking:", serial)
	
	// Step 1: Base85 decode
	decoder := base85.NewDecoder()
	decoded, err := decoder.Decode(serial)
	if err != nil {
		fmt.Printf("Base85 decode error: %v\n", err)
		return
	}
	
	fmt.Printf("Decoded bytes (%d): %x\n", len(decoded), decoded)
	totalBits := int64(len(decoded) * 8)
	fmt.Printf("Total bits available: %d\n", totalBits)
	
	// Step 2: Use tokenizer and track positions
	tokenizer := token.NewTokenizerFromBytes(decoded)
	
	for i := 0; i < 20; i++ {
		posBefore := tokenizer.Position()
		tok, err := tokenizer.NextToken()
		posAfter := tokenizer.Position()
		
		if err != nil {
			fmt.Printf("Token %d error at position %d->%d: %v\n", i+1, posBefore, posAfter, err)
			fmt.Printf("Remaining bits: %d\n", totalBits-posAfter)
			break
		}
		
		fmt.Printf("Token %d: Type=%s, Pos=%d->%d, Delta=%d, Bits=%d, Value=%v\n", 
			i+1, tok.Type.String(), posBefore, posAfter, posAfter-posBefore, tok.BitSize, tok.Value)
		
		if tok.Type == token.TokenEOF {
			fmt.Printf("EOF reached at position %d\n", posAfter)
			break
		}
		
		// Check if we're approaching the end
		if posAfter > totalBits-3 {
			fmt.Printf("WARNING: Near end of bitstream! Remaining: %d bits\n", totalBits-posAfter)
		}
	}
}
