package app

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
	"github.com/shawnvan/bl4/internal/cli/commands"
	"github.com/shawnvan/bl4/pkg/logger"
)

// Run executes the CLI application
func Run() error {
	app := &cli.App{
		Name:     "bl4",
		Usage:    "BL4 Item Serial Code Codec CLI",
		Version:  "1.0.0",
		Commands: []*cli.Command{
			commands.DecodeCommand(),
			commands.EncodeCommand(),
			commands.ValidateCommand(),
			commands.BatchCommand(),
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Configuration file path",
				Value:   "config.json",
			},
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"d"},
				Usage:   "Enable debug logging",
			},
		},
		Before: func(c *cli.Context) error {
			// Initialize logger
			logLevel := "info"
			if c.Bool("debug") {
				logLevel = "debug"
			}

			if err := logger.Init(logLevel, "console", "stdout"); err != nil {
				fmt.Printf("Failed to initialize logger: %v\n", err)
				return err
			}

			logger.Sugar().Infow("BL4 CLI starting",
				"config", c.String("config"),
				"debug", c.Bool("debug"),
			)

			return nil
		},
		Action: func(c *cli.Context) error {
			fmt.Println("BL4 Item Serial Code Codec")
			fmt.Println("Use 'bl4 help' for available commands")
			return nil
		},
	}

	return app.Run(os.Args)
}