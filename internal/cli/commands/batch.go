package commands

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
	"github.com/shawnvan/bl4/internal/api/models"
	"github.com/shawnvan/bl4/internal/api/services"
	"github.com/shawnvan/bl4/pkg/logger"
)

// BatchCommand creates the batch processing command
func BatchCommand() *cli.Command {
	return &cli.Command{
		Name:  "batch",
		Usage: "Process multiple BL4 item serial codes in batch",
		Description: "Decode or encode multiple BL4 item serial codes with progress tracking and statistics",
		Subcommands: []*cli.Command{
			batchDecodeCommand(),
			batchEncodeCommand(),
			batchStatusCommand(),
		},
	}
}

// batchDecodeCommand creates the batch decode command
func batchDecodeCommand() *cli.Command {
	return &cli.Command{
		Name:  "decode",
		Usage: "Decode multiple BL4 item serial codes",
		Description: "Decode multiple BL4 item serial codes from file or stdin with progress tracking",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "file",
				Aliases:  []string{"f"},
				Usage:    "Input file containing serial codes (one per line)",
				Required: false,
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output file for results (JSON format)",
			},
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"F"},
				Usage:   "Output format: json, csv, structured",
				Value:   "json",
			},
			&cli.IntFlag{
				Name:    "concurrency",
				Aliases: []string{"c"},
				Usage:   "Number of concurrent workers",
				Value:   5,
			},
			&cli.IntFlag{
				Name:    "timeout",
				Aliases: []string{"t"},
				Usage:   "Timeout in seconds",
				Value:   30,
			},
			&cli.IntFlag{
				Name:    "batch-timeout",
				Aliases: []string{"T"},
				Usage:   "Per-item timeout in milliseconds",
				Value:   5000,
			},
			&cli.BoolFlag{
				Name:    "include-bitstream",
				Aliases: []string{"b"},
				Usage:   "Include bitstream data in results",
			},
			&cli.BoolFlag{
				Name:    "include-tokens",
				Aliases: []string{"k"},
				Usage:   "Include token data in results",
			},
			&cli.BoolFlag{
				Name:    "fail-fast",
				Usage:   "Stop processing on first error",
			},
			&cli.BoolFlag{
				Name:    "progress",
				Aliases: []string{"p"},
				Usage:   "Show progress bar",
				Value:   true,
			},
		},
		Action: batchDecodeAction,
	}
}

// batchEncodeCommand creates the batch encode command
func batchEncodeCommand() *cli.Command {
	return &cli.Command{
		Name:  "encode",
		Usage: "Encode multiple BL4 item serial codes",
		Description: "Encode multiple structured item data to BL4 serial codes",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "file",
				Aliases:  []string{"f"},
				Usage:    "Input file containing structured item data (JSON format)",
				Required: false,
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output file for serial codes",
			},
			&cli.IntFlag{
				Name:    "concurrency",
				Aliases: []string{"c"},
				Usage:   "Number of concurrent workers",
				Value:   5,
			},
			&cli.BoolFlag{
				Name:    "progress",
				Aliases: []string{"p"},
				Usage:   "Show progress bar",
				Value:   true,
			},
		},
		Action: batchEncodeAction,
	}
}

// batchStatusCommand creates the batch status command
func batchStatusCommand() *cli.Command {
	return &cli.Command{
		Name:  "status",
		Usage: "Show status of batch processing operations",
		Description: "Display status of active or recent batch processing operations",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "batch-id",
				Usage: "Specific batch ID to check status",
			},
		},
		Action: batchStatusAction,
	}
}

// batchDecodeAction handles the batch decode command
func batchDecodeAction(c *cli.Context) error {
	_ = time.Now() // startTime available but not used yet

	// Read serial codes from file or stdin
	serialCodes, err := readSerialCodes(c)
	if err != nil {
		return fmt.Errorf("failed to read serial codes: %w", err)
	}

	if len(serialCodes) == 0 {
		return fmt.Errorf("no serial codes provided")
	}

	logger.Sugar().Infow("Starting batch decode",
		"total_codes", len(serialCodes),
		"concurrency", c.Int("concurrency"),
		"timeout", c.Int("timeout"),
	)

	// Create batch request
	request := &models.BatchDecodeRequest{
		SerialCodes: serialCodes,
		Options: &models.BatchDecodeOptions{
			IndividualOptions: &models.DecodeOptions{
				IncludeBitstream: c.Bool("include-bitstream"),
				IncludeTokens:    c.Bool("include-tokens"),
				Format:          c.String("format"),
			},
			MaxConcurrency: c.Int("concurrency"),
			FailFast:      c.Bool("fail-fast"),
			Timeout:        c.Int("timeout") * 1000, // Convert to milliseconds
			PerItemTimeout: c.Int("batch-timeout"),
		},
	}

	// Create batch processor
	processor := services.NewBatchProcessor()

	// Create progress tracker if requested
	var progressTracker *ProgressTracker
	if c.Bool("progress") {
		progressTracker = NewProgressTracker(len(serialCodes))
		go progressTracker.Start()
		defer progressTracker.Stop()
	}

	// Process batch
	result, err := processor.ProcessBatch(request)
	if err != nil {
		return fmt.Errorf("batch processing failed: %w", err)
	}

	// Update progress tracker to completion
	if progressTracker != nil {
		progressTracker.Complete(result.Statistics.Successful, result.Statistics.Failed)
	}

	// Output results
	outputFormat := c.String("format")
	outputFile := c.String("output")

	switch outputFormat {
	case "json":
		return outputJSONResults(result, outputFile)
	case "csv":
		return outputCSVResults(result, outputFile)
	case "structured":
		return outputStructuredResults(result, outputFile)
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
}

// batchEncodeAction handles the batch encode command
func batchEncodeAction(c *cli.Context) error {
	return fmt.Errorf("batch encode command not yet implemented")
}

// batchStatusAction handles the batch status command
func batchStatusAction(c *cli.Context) error {
	batchID := c.String("batch-id")

	if batchID != "" {
		// Show status for specific batch
		processor := services.NewBatchProcessor()
		progress, err := processor.GetBatchProgress(batchID)
		if err != nil {
			return fmt.Errorf("failed to get batch status: %w", err)
		}

		fmt.Printf("Batch ID: %s\n", batchID)
		fmt.Printf("Status: %s\n", getBatchStatus(progress))
		fmt.Printf("Progress: %d/%d (%.1f%%)\n",
			progress.Processed, progress.Total,
			float64(progress.Processed)/float64(progress.Total)*100)
		fmt.Printf("Successful: %d, Failed: %d\n", progress.Successful, progress.Failed)
		if progress.Cancelled {
			fmt.Printf("Cancelled: %s\n", progress.CancelReason)
		}
	} else {
		// Show general status
		fmt.Println("No active batch operations")
		fmt.Println("Use --batch-id to check specific batch status")
	}

	return nil
}

// readSerialCodes reads serial codes from file or stdin
func readSerialCodes(c *cli.Context) ([]string, error) {
	filePath := c.String("file")

	if filePath != "" {
		// Read from file
		file, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
		}
		defer file.Close()

		content, err := io.ReadAll(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
		}

		lines := strings.Split(string(content), "\n")
		var codes []string
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") {
				codes = append(codes, line)
			}
		}
		return codes, nil
	} else {
		// Read from stdin
		var codes []string
		fmt.Println("Enter serial codes (Ctrl+D to finish):")

		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			code := scanner.Text()
			if strings.TrimSpace(code) != "" {
				codes = append(codes, strings.TrimSpace(code))
			}
		}
		return codes, nil
	}
}

// outputJSONResults outputs results in JSON format
func outputJSONResults(result *services.BatchProcessResult, outputFile string) error {
	var output io.Writer = os.Stdout

	if outputFile != "" {
		file, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()
		output = file
	}

	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")

	response := map[string]interface{}{
		"success":    result.Success,
		"statistics": result.Statistics,
		"results":    result.Results,
		"duration":   result.Duration.Milliseconds(),
		"timestamp":  time.Now().UTC(),
	}

	return encoder.Encode(response)
}

// outputCSVResults outputs results in CSV format
func outputCSVResults(result *services.BatchProcessResult, outputFile string) error {
	var output io.Writer = os.Stdout

	if outputFile != "" {
		file, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()
		output = file
	}

	// Write CSV header
	fmt.Fprintf(output, "Index,SerialCode,Success,Level,Type,Manufacturer,PartsCount,Duration(μs)\n")

	// Write results
	for i, itemResult := range result.Results {
		var level, itemType, manufacturer string
		var partsCount int

		if itemResult.Data != nil && itemResult.Data.ItemData != nil {
			level = strconv.Itoa(itemResult.Data.ItemData.Level)
			itemType = itemResult.Data.ItemData.Type
			manufacturer = itemResult.Data.ItemData.Manufacturer
			partsCount = len(itemResult.Data.ItemData.Parts)
		}

		success := "false"
		if itemResult.Success {
			success = "true"
		}

		fmt.Fprintf(output, "%d,%s,%s,%s,%s,%s,%d,%d\n",
			i, itemResult.SerialCode, success, level, itemType, manufacturer, partsCount, itemResult.Duration)
	}

	return nil
}

// outputStructuredResults outputs results in structured format
func outputStructuredResults(result *services.BatchProcessResult, outputFile string) error {
	var output io.Writer = os.Stdout

	if outputFile != "" {
		file, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()
		output = file
	}

	fmt.Fprintf(output, "Batch Processing Results\n")
	fmt.Fprintf(output, "=====================\n\n")
	fmt.Fprintf(output, "Success: %t\n", result.Success)
	fmt.Fprintf(output, "Total: %d\n", result.Statistics.Total)
	fmt.Fprintf(output, "Successful: %d\n", result.Statistics.Successful)
	fmt.Fprintf(output, "Failed: %d\n", result.Statistics.Failed)
	fmt.Fprintf(output, "Success Rate: %.1f%%\n", result.Statistics.SuccessRate*100)
	fmt.Fprintf(output, "Duration: %v\n", result.Duration)
	fmt.Fprintf(output, "Timestamp: %s\n", time.Now().UTC().Format(time.RFC3339))

	fmt.Fprintf(output, "\nIndividual Results:\n")
	fmt.Fprintf(output, "----------------\n")

	for i, itemResult := range result.Results {
		fmt.Fprintf(output, "\nItem %d:\n", i+1)
		fmt.Fprintf(output, "  Serial Code: %s\n", itemResult.SerialCode)
		fmt.Fprintf(output, "  Success: %t\n", itemResult.Success)
		fmt.Fprintf(output, "  Duration: %d μs\n", itemResult.Duration)

		if itemResult.Error != nil {
			fmt.Fprintf(output, "  Error: %s\n", itemResult.Error.Message)
		}

		if itemResult.Data != nil && itemResult.Data.ItemData != nil {
			fmt.Fprintf(output, "  Level: %d\n", itemResult.Data.ItemData.Level)
			fmt.Fprintf(output, "  Type: %s\n", itemResult.Data.ItemData.Type)
			fmt.Fprintf(output, "  Manufacturer: %s\n", itemResult.Data.ItemData.Manufacturer)
			fmt.Fprintf(output, "  Parts: %d\n", len(itemResult.Data.ItemData.Parts))

			if len(itemResult.Data.ItemData.Parts) > 0 {
				fmt.Fprintf(output, "  Parts Details:\n")
				for j, part := range itemResult.Data.ItemData.Parts {
					fmt.Fprintf(output, "    %d: Index=%d, Value=%d\n", j+1, part.Index, part.Value)
				}
			}
		}
	}

	return nil
}

// getBatchStatus returns a human-readable status string
func getBatchStatus(progress *services.BatchProgress) string {
	if progress.Cancelled {
		return "Cancelled"
	}
	if progress.Processed == progress.Total {
		if progress.Failed == 0 {
			return "Completed"
		}
		return "Completed with errors"
	}
	if progress.Processed > 0 {
		return "In Progress"
	}
	return "Pending"
}