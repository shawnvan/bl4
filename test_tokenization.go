package main

import (
	"fmt"
	"github.com/shawnvan/bl4/internal/codec/bitstream"
	"github.com/shawnvan/bl4/internal/codec/token"
	"github.com/shawnvan/bl4/internal/codec/base85"
)

func main() {
	serial := "@Uge8s!m/)}}!tjZpNWCvC00"
	fmt.Println("Testing tokenization with:", serial)
	
	// Step 1: Base85 decode
	decoder := base85.NewDecoder()
	decoded, err := decoder.Decode(serial)
	if err != nil {
		fmt.Printf("Base85 decode error: %v\n", err)
		return
	}
	
	fmt.Printf("Decoded bytes (%d): %x\n", len(decoded), decoded)
	
	// Step 2: Create bitstream reader
	reader := bitstream.NewReaderFromBytes(decoded)
	
	// Step 3: Try to read tokens manually
	fmt.Println("\nManual token parsing:")
	pos := int64(0)
	
	for {
		// Read first bit
		b1, err := reader.ReadBits(1)
		if err != nil {
			if err == bitstream.ErrEOF {
				fmt.Printf("EOF reached at position %d\n", pos)
				break
			}
			fmt.Printf("Error reading bit at position %d: %v\n", pos, err)
			break
		}
		
		b2, err := reader.ReadBits(1)
		if err != nil {
			fmt.Printf("Error reading second bit at position %d: %v\n", pos, err)
			break
		}
		
		tok := (b1 << 1) | b2
		pos = reader.Position()
		
		fmt.Printf("Position %d: First two bits: %b (value: %d)\n", pos-2, tok, tok)
		
		if tok == 0b00 || tok == 0b01 {
			// Separator tokens
			fmt.Printf("  Separator token: %b\n", tok)
			continue
		}
		
		// Need third bit
		b3, err := reader.ReadBits(1)
		if err != nil {
			fmt.Printf("Error reading third bit at position %d: %v\n", pos, err)
			break
		}
		
		tok = (tok << 1) | b3
		fmt.Printf("Position %d: Full 3-bit token: %b (value: %d)\n", pos-1, tok, tok)
		
		switch tok {
		case 0b100: // VARINT
			fmt.Println("  VARINT token - reading length...")
		case 0b101: // PART
			fmt.Println("  PART token - reading 32 bits...")
		case 0b110: // VARBIT
			fmt.Println("  VARBIT token - reading length...")
		case 0b111: // STRING
			fmt.Println("  STRING token - reading length...")
		default:
			fmt.Printf("  Unknown token: %b\n", tok)
		}
		
		// Stop after a few tokens
		if pos > 24 {
			break
		}
	}
	
	fmt.Println("\nUsing tokenizer:")
	// Step 4: Use tokenizer
	tokenizer := token.NewTokenizerFromBytes(decoded)
	
	for i := 0; i < 20; i++ {
		tok, err := tokenizer.NextToken()
		if err != nil {
			fmt.Printf("Error reading token %d: %v\n", i+1, err)
			break
		}
		
		fmt.Printf("Token %d: Type=%s, Value=%v, Bits=%d, Position=%d\n", 
			i+1, tok.Type.String(), tok.Value, tok.BitSize, tok.Position)
		
		if tok.Type == token.TokenEOF {
			break
		}
	}
}
