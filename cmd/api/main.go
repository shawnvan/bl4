package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/shawnvan/bl4/internal/api"
	"github.com/shawnvan/bl4/internal/config"
	"github.com/shawnvan/bl4/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// Parse command line flags
	flag.String("config", "", "Path to configuration file") // Config file flag for future use
	var (
		version = flag.Bool("version", false, "Show version information")
		help    = flag.Bool("help", false, "Show help information")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	if *version {
		showVersion()
		return
	}

	// Initialize configuration
	if err := config.InitializeGlobalConfig(); err != nil {
		log.Fatalf("Failed to initialize configuration: %v", err)
	}

	cfg := config.GetGlobalConfig()
	logger.Sugar().Infow("Starting BL4 Item Codec API",
		"version", "1.0.0",
		"environment", cfg.Environment,
		"port", cfg.Server.Port,
	)

	// Create API server
	server := api.NewServer(cfg.Server)

	// Start server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
			logger.Logger.Error("Failed to start server", zap.Error(err))
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Sugar().Info("Shutting down server...")

	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Stop(ctx); err != nil {
		logger.Logger.Error("Server forced to shutdown", zap.Error(err))
		os.Exit(1)
	}

	logger.Sugar().Info("Server exited")
}

func showHelp() {
	fmt.Printf(`BL4 Item Codec API Server

Usage:
  bl4-api [options]

Options:
  -config string     Path to configuration file (default: looks for config.json in ./, ./configs/, /etc/bl4/, $HOME/.bl4/)
  -version           Show version information
  -help              Show this help message

Environment Variables:
  BL4_ENVIRONMENT    Application environment (development, production, testing)
  BL4_SERVER_PORT    Server port (default: 8080)
  BL4_LOG_LEVEL      Log level (debug, info, warn, error, fatal)
  BL4_LOG_FORMAT     Log format (json, console)

Examples:
  bl4-api                           # Start with default configuration
  bl4-api -config ./prod.json      # Start with specific configuration file
  BL4_ENVIRONMENT=production bl4-api # Start with production environment

`)
}

func showVersion() {
	fmt.Printf("BL4 Item Codec API v1.0.0\n")
	fmt.Printf("Build: %s\n", "development")
	fmt.Printf("Go: %s\n", runtime.Version())
}