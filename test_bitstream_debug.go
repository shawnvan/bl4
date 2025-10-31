package main

import (
	"fmt"
	"github.com/shawnvan/bl4/internal/codec/bitstream"
	"github.com/shawnvan/bl4/internal/codec/base85"
)

func main() {
	serial := "@Uge8s!m/)}}!tjZpNWCvC00"
	fmt.Println("Testing bitstream reading:", serial)
	
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
	
	// Read and examine the bitstream
	for i := 0; i < 135; i++ {
		b, err := reader.ReadBits(1)
		if err != nil {
			fmt.Printf("Error at bit %d: %v\n", i, err)
			break
		}
		if i%8 == 0 {
			fmt.Printf("\nByte %d (%d bits): ", i/8, i%8)
		}
		fmt.Printf("%d", b)
	}
	fmt.Println()
}
