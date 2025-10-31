package commands

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/urfave/cli/v2"
	"github.com/shawnvan/bl4/internal/api/models"
	"github.com/shawnvan/bl4/internal/codec/serial"
	"github.com/shawnvan/bl4/pkg/logger"
)

// EncodeCommand creates the encode command
func EncodeCommand() *cli.Command {
	return &cli.Command{
		Name:  "encode",
		Usage: "Encode structured item data to BL4 serial code",
		Description: "Encode structured item data to BL4 item serial code",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "level",
				Aliases:  []string{"l"},
				Usage:    "Item level",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "type",
				Aliases:  []string{"t"},
				Usage:    "Item type",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "manufacturer",
				Aliases:  []string{"m"},
				Usage:    "Item manufacturer",
				Required: true,
			},
			&cli.IntFlag{
				Name:    "parts",
				Aliases: []string{"p"},
				Usage:   "Number of parts to include",
				Value:   0,
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"V"},
				Usage:   "Show verbose output including item data JSON",
				Value:   false,
			},
		},
		Action: encodeAction,
	}
}

// encodeAction handles the encode command
func encodeAction(c *cli.Context) error {
	// Parse command line arguments
	levelStr := c.String("level")
	itemType := c.String("type")
	manufacturer := c.String("manufacturer")
	partsCount := c.Int("parts")

	// Convert level to integer
	level, err := strconv.Atoi(levelStr)
	if err != nil {
		return fmt.Errorf("invalid level value: %s (must be an integer)", levelStr)
	}

	// Create item data structure
	itemData := &models.ItemData{
		Level:        level,
		Type:         itemType,
		Manufacturer: manufacturer,
	}

	// Add parts if specified
	if partsCount > 0 {
		itemData.Parts = make([]models.PartData, partsCount)
		for i := 0; i < partsCount; i++ {
			itemData.Parts[i] = models.PartData{
				Index: i,
				Value: 0, // Default value
				Type:  "generic",
			}
		}
	}

	// Create serializer
	serializer := serial.NewSerializer()

	// Serialize the item data
	result, err := serializer.SerializeItem(itemData)
	if err != nil {
		return fmt.Errorf("failed to serialize item data: %w", err)
	}

	// Output the result
	fmt.Printf("✅ Successfully encoded item to BL4 serial code:\n")
	fmt.Printf("   Code: %s\n", result.SerialCode)

	if result.GetProcessingTimeMs() > 0 {
		fmt.Printf("   Processing time: %d ms\n", result.GetProcessingTimeMs())
	}

	// Optionally output item details as JSON
	if c.Bool("verbose") {
		itemJSON, _ := json.MarshalIndent(itemData, "   ", "  ")
		fmt.Printf("   Item data:\n%s\n", string(itemJSON))
	}

	logger.Sugar().Infow("CLI encode completed",
		"level", level,
		"type", itemType,
		"manufacturer", manufacturer,
		"parts_count", partsCount,
		"serial_code", result.SerialCode,
	)

	return nil
}