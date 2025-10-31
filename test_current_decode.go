package main

import (
	"bl4/internal/api/handlers"
	"encoding/json"
	"fmt"
)

func main() {
	serial := "@Uge8s!m/)}}!tjZpNWCvC00"
	fmt.Println("Testing serial with current implementation:", serial)
	
	decodeHandler := handlers.NewDecodeHandler()
	
	result, err := decodeHandler.Decode(serial)
	if err != nil {
		fmt.Printf("Decode error: %v
", err)
		return
	}
	
	fmt.Printf("Result: %+v
", result)
	
	// Try to print the raw bytes
	fmt.Printf("Raw data bytes: %x
", result.Data)
}
