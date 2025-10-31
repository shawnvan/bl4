package main

import (
	"fmt"
	"os"

	"github.com/shawnvan/bl4/internal/tui/app"
	"github.com/shawnvan/bl4/pkg/logger"
)

func main() {
	// Initialize logger
	logger.Init("info", "console", "")

	// Run TUI application
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}