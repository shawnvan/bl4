package main

import (
	"fmt"
	"github.com/shawnvan/bl4/internal/codec/bitstream"
)

func main() {
	data := []byte{0x01} // Just one byte: 00000001
	reader := bitstream.NewReaderFromBytes(data)
	
	fmt.Println("Testing EOF handling:")
	
	// Read first bit
	b1, err := reader.ReadBits(1)
	fmt.Printf("Bit 1: %d, error: %v\n", b1, err)
	
	// Read second bit - this should be EOF
	b2, err := reader.ReadBits(1)
	fmt.Printf("Bit 2: %d, error: %v\n", b2, err)
	
	// Try to read more
	b3, err := reader.ReadBits(1)
	fmt.Printf("Bit 3: %d, error: %v\n", b3, err)
}
