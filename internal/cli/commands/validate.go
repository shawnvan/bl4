package commands

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/urfave/cli/v2"
	"github.com/shawnvan/bl4/internal/codec/base85"
	"github.com/shawnvan/bl4/pkg/logger"
)

// ValidateCommand creates the validate command
func ValidateCommand() *cli.Command {
	return &cli.Command{
		Name:  "validate",
		Usage: "Validate BL4 item serial code format",
		Description: "Validate the format of a BL4 item serial code without processing",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "code",
				Aliases:  []string{"c"},
				Usage:    "BL4 item serial code to validate",
				Required: true,
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"V"},
				Usage:   "Show verbose validation details",
				Value:   false,
			},
		},
		Action: validateAction,
	}
}

// validateAction handles the validate command
func validateAction(c *cli.Context) error {
	code := strings.TrimSpace(c.String("code"))

	if code == "" {
		return fmt.Errorf("no serial code provided")
	}

	// Basic format validation
	if err := validateBasicFormat(code); err != nil {
		fmt.Printf("❌ Invalid format: %v\n", err)
		return fmt.Errorf("format validation failed: %w", err)
	}

	// Check if it starts with @ symbol
	if !strings.HasPrefix(code, "@") {
		fmt.Printf("❌ Invalid format: BL4 serial codes must start with '@'\n")
		return fmt.Errorf("missing @ prefix")
	}

	// Remove @ for Base85 validation
	b85Part := code[1:]

	// Validate Base85 characters
	if err := validateBase85Characters(b85Part); err != nil {
		fmt.Printf("❌ Invalid Base85 characters: %v\n", err)
		return fmt.Errorf("base85 validation failed: %w", err)
	}

	// Check length constraints
	if len(b85Part) < 10 {
		fmt.Printf("❌ Invalid length: Serial code too short (minimum 10 characters after @)\n")
		return fmt.Errorf("serial code too short")
	}

	if len(b85Part) > 1000 {
		fmt.Printf("❌ Invalid length: Serial code too long (maximum 1000 characters after @)\n")
		return fmt.Errorf("serial code too long")
	}

	// Try to decode with Base85 decoder for additional validation
	decoder := base85.NewDecoder()
	if _, err := decoder.Decode(b85Part); err != nil {
		fmt.Printf("❌ Invalid Base85 encoding: %v\n", err)
		return fmt.Errorf("base85 decode validation failed: %w", err)
	}

	// If we get here, the code appears valid
	fmt.Printf("✅ Serial code format is valid:\n")
	fmt.Printf("   Code: %s\n", code)
	fmt.Printf("   Length: %d characters\n", len(code))
	fmt.Printf("   Base85 length: %d characters\n", len(b85Part))

	// Additional info
	if c.Bool("verbose") {
		fmt.Printf("   Character set: Valid Base85 characters\n")
		fmt.Printf("   Structure: Proper @ prefix + Base85 data\n")
	}

	logger.Sugar().Infow("CLI validate completed",
		"code", code,
		"length", len(code),
		"valid", true,
	)

	return nil
}

// validateBasicFormat performs basic format validation
func validateBasicFormat(code string) error {
	if code == "" {
		return fmt.Errorf("empty code")
	}

	// Check for whitespace
	if strings.ContainsAny(code, " \t\n\r") {
		return fmt.Errorf("contains whitespace")
	}

	// Check for control characters
	for _, r := range code {
		if r < 32 || r > 126 {
			return fmt.Errorf("contains invalid control characters")
		}
	}

	return nil
}

// validateBase85Characters validates that all characters are valid Base85
func validateBase85Characters(data string) error {
	// Base85 valid character set (typical Z85/ASCII85 variant)
	validChars := regexp.MustCompile(`^[0-9A-Za-z!#$%&()*+\-;<=>?@^_\{\|\}~]+$`)

	if !validChars.MatchString(data) {
		// Find invalid characters
		invalidChars := ""
		for _, r := range data {
			if !strings.ContainsRune("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz!#$%&()*+-;<=>?@^_`{|}~", r) {
				invalidChars += string(r)
			}
		}
		return fmt.Errorf("contains invalid characters: %s", invalidChars)
	}

	return nil
}