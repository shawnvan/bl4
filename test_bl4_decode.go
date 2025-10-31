package main

import (
	"fmt"
	"github.com/shawnvan/bl4/internal/codec/base85"
)

func main() {
	serial := "@Uge8s!m/)}}!tjZpNWCvC00"
	fmt.Println("Testing serial with current implementation:", serial)
	
	decoder := base85.NewDecoder()
	options := &base85.DecodeOptions{
		IncludeStatistics: true,
		ValidateFormat:     true,
	}
	
	result, err := decoder.DecodeWithOptions(serial, options)
	if err != nil {
		fmt.Printf("Decode error: %v\n", err)
		return
	}
	
	fmt.Printf("Result: %+v\n", result)
	fmt.Printf("Raw bytes: %x\n", result.Data)
}
