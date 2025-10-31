package commands

import (
	"encoding/json"
	"fmt"

	"github.com/urfave/cli/v2"
	"github.com/shawnvan/bl4/internal/codec/serial"
	"github.com/shawnvan/bl4/pkg/logger"
)

// DecodeCommand creates the decode command
func DecodeCommand() *cli.Command {
	return &cli.Command{
		Name:  "decode",
		Usage: "Decode a single BL4 item serial code",
		Description: "Decode a BL4 item serial code to structured item data",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "code",
				Aliases:  []string{"c"},
				Usage:    "BL4 item serial code to decode",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"f"},
				Usage:   "Output format: json, structured",
				Value:   "json",
			},
			&cli.BoolFlag{
				Name:    "verbose-output",
				Aliases: []string{"V"},
				Usage:   "Show verbose output",
				Value:   false,
			},
		},
		Action: decodeAction,
	}
}

// decodeAction handles the decode command
func decodeAction(c *cli.Context) error {
	// Parse command line arguments
	code := c.String("code")
	format := c.String("format")

	// Validate input
	if code == "" {
		return fmt.Errorf("no serial code provided")
	}

	logger.Sugar().Infow("Decoding serial code",
		"code", code,
		"format", format,
	)

	// Create deserializer
	deserializer := serial.NewDeserializer()

	// Decode the code
	result, err := deserializer.DeserializeItem(code)
	if err != nil {
		return fmt.Errorf("failed to decode serial code: %w", err)
	}

	// Output based on format
	switch format {
	case "json":
		return outputJSONDecodeResult(result)
	case "structured":
		return outputStructuredDecodeResult(result, c.Bool("verbose-output"))
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

// outputJSONDecodeResult outputs decode result in JSON format
func outputJSONDecodeResult(result *serial.DeserializationResult) error {
	response := map[string]interface{}{
		"success":       true,
		"serial_code":   result.SerialCode,
		"item_data":     result.ItemData,
		"duration_ms":   result.GetProcessingTimeMs(),
	}

	// Add additional info if available
	if result.BitstreamInfo != nil {
		response["bitstream_info"] = result.BitstreamInfo
	}

	if len(result.TokenInfos) > 0 {
		response["token_count"] = len(result.TokenInfos)
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Printf("%s\n", string(jsonData))
	return nil
}

// outputStructuredDecodeResult outputs decode result in structured format
func outputStructuredDecodeResult(result *serial.DeserializationResult, verbose bool) error {
	fmt.Printf("✅ Successfully decoded BL4 serial code:\n")
	fmt.Printf("   Serial Code: %s\n", result.SerialCode)

	if result.ItemData != nil {
		fmt.Printf("   Item Level: %d\n", result.ItemData.Level)
		fmt.Printf("   Item Type: %s\n", result.ItemData.Type)
		fmt.Printf("   Manufacturer: %s\n", result.ItemData.Manufacturer)

		if result.ItemData.Name != "" {
			fmt.Printf("   Item Name: %s\n", result.ItemData.Name)
		}

		if result.ItemData.Rarity != "" {
			fmt.Printf("   Rarity: %s\n", result.ItemData.Rarity)
		}

		if len(result.ItemData.Parts) > 0 {
			fmt.Printf("   Parts: %d\n", len(result.ItemData.Parts))
			if verbose {
				for _, part := range result.ItemData.Parts {
					fmt.Printf("     %d: %s (Value: %d)\n", part.Index, part.Type, part.Value)
				}
			}
		}

		if verbose && len(result.ItemData.Properties) > 0 {
			fmt.Printf("   Properties:\n")
			for key, value := range result.ItemData.Properties {
				fmt.Printf("     %s: %v\n", key, value)
			}
		}
	}

	if result.GetProcessingTimeMs() > 0 {
		fmt.Printf("   Processing Time: %d ms\n", result.GetProcessingTimeMs())
	}

	return nil
}