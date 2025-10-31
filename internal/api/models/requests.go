package models

import (
	"fmt"

	"github.com/shawnvan/bl4/pkg/validator"
)

// DecodeRequest represents a request to decode an item serial code
type DecodeRequest struct {
	SerialCode string                 `json:"serial_code" binding:"required" example:"@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}"`
	Options    *DecodeOptions         `json:"options,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// DecodeOptions contains optional parameters for decoding
type DecodeOptions struct {
	// IncludeBitstream includes the raw bitstream data in the response
	IncludeBitstream bool `json:"include_bitstream,omitempty"`

	// IncludeTokens includes individual parsed tokens in the response
	IncludeTokens bool `json:"include_tokens,omitempty"`

	// Format specifies the desired output format
	Format string `json:"format,omitempty" example:"structured"` // "structured", "json", "csv", "raw"

	// PrettyPrint enables pretty-printing of JSON output
	PrettyPrint bool `json:"pretty_print,omitempty"`

	// ValidateOnly only validates the serial code without full decoding
	ValidateOnly bool `json:"validate_only,omitempty"`

	// IncludeStats includes processing statistics in the response
	IncludeStats bool `json:"include_stats,omitempty"`

	// MaxParts limits the number of parts to decode (for performance)
	MaxParts int `json:"max_parts,omitempty"`

	// Timeout specifies maximum processing time in milliseconds
	Timeout int `json:"timeout,omitempty"`
}

// EncodeRequest represents a request to encode structured data to an item serial code
type EncodeRequest struct {
	ItemData  *ItemData              `json:"item_data" binding:"required"`
	Options   *EncodeOptions         `json:"options,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// EncodeOptions contains optional parameters for encoding
type EncodeOptions struct {
	// Format specifies the input format
	Format string `json:"format,omitempty" example:"structured"` // "structured", "json", "csv"

	// ValidateParts validates each part before encoding
	ValidateParts bool `json:"validate_parts,omitempty"`

	// OptimizeSize enables size optimization of the output
	OptimizeSize bool `json:"optimize_size,omitempty"`

	// IncludePrefix includes the @U prefix in the output
	IncludePrefix bool `json:"include_prefix,omitempty"`

	// TargetVersion specifies the target BL4 version format
	TargetVersion string `json:"target_version,omitempty" example:"1.0"`
}

// BatchDecodeRequest represents a request to decode multiple serial codes
type BatchDecodeRequest struct {
	SerialCodes []string               `json:"serial_codes" binding:"required,min=1,max=1000"`
	Options     *BatchDecodeOptions    `json:"options,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// BatchDecodeOptions contains options for batch decoding
type BatchDecodeOptions struct {
	// IndividualOptions are applied to each decode request
	IndividualOptions *DecodeOptions `json:"individual_options,omitempty"`

	// Parallel processing options
	MaxConcurrency int  `json:"max_concurrency,omitempty"`
	FailFast       bool `json:"fail_fast,omitempty"` // Stop on first error

	// Progress reporting
	ReportProgress bool `json:"report_progress,omitempty"`

	// Batch-level options
	Timeout        int  `json:"timeout,omitempty"`         // Overall timeout in milliseconds
	PerItemTimeout int  `json:"per_item_timeout,omitempty"` // Timeout per item in milliseconds

	// Output formatting
	CompactOutput bool `json:"compact_output,omitempty"`
}

// BatchEncodeRequest represents a request to encode multiple items
type BatchEncodeRequest struct {
	Items    []ItemData             `json:"items" binding:"required,min=1,max=1000"`
	Options  *BatchEncodeOptions    `json:"options,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// BatchEncodeOptions contains options for batch encoding
type BatchEncodeOptions struct {
	// IndividualOptions are applied to each encode request
	IndividualOptions *EncodeOptions `json:"individual_options,omitempty"`

	// Parallel processing options
	MaxConcurrency int  `json:"max_concurrency,omitempty"`
	FailFast       bool `json:"fail_fast,omitempty"`

	// Progress reporting
	ReportProgress bool `json:"report_progress,omitempty"`

	// Batch-level options
	Timeout        int `json:"timeout,omitempty"`         // Overall timeout in milliseconds
	PerItemTimeout int `json:"per_item_timeout,omitempty"` // Timeout per item in milliseconds
}

// ItemData represents structured item information
type ItemData struct {
	// Basic item information
	Level       int                    `json:"level" binding:"min=1,max=100" example:"24"`
	Type        string                 `json:"type" binding:"required" example:"pistol"`
	Manufacturer string                `json:"manufacturer" binding:"required" example:"maliwan"`

	// Item parts/components
	Parts       []PartData             `json:"parts,omitempty"`
	RawParts    string                 `json:"raw_parts,omitempty"` // For backward compatibility

	// Additional item properties
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Rarity      string                 `json:"rarity,omitempty"`

	// Game-specific data
	GameVersion string                 `json:"game_version,omitempty"`
	Properties  map[string]interface{} `json:"properties,omitempty"`

	// Metadata
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PartData represents a single part or component of an item
type PartData struct {
	// Part identification
	Index int    `json:"index" binding:"min=0" example:"0"`
	Type  string `json:"type,omitempty" example:"barrel"`
	Name  string `json:"name,omitempty" example:"Maliwan Barrel"`

	// Part values
	Value    int                    `json:"value" example:"50"`
	RawValue string                 `json:"raw_value,omitempty"` // Original encoded value

	// Part properties
	Attributes map[string]interface{} `json:"attributes,omitempty"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Validation methods

// Validate validates the DecodeRequest
func (r *DecodeRequest) Validate() error {
	if r.SerialCode == "" {
		return validator.NewValidationError(validator.ErrCodeEmptyInput, "serial_code is required").
			WithField("serial_code", r.SerialCode)
	}

	if len(r.SerialCode) < 3 { // Minimum @U prefix + 1 character
		return validator.NewValidationError(validator.ErrCodeInvalidLength, "serial_code too short").
			WithField("serial_code", r.SerialCode).
			WithDetail("min_length", 3)
	}

	if len(r.SerialCode) > 10000 { // Maximum reasonable length
		return validator.NewValidationError(validator.ErrCodeInvalidLength, "serial_code too long").
			WithField("serial_code", r.SerialCode).
			WithDetail("max_length", 10000)
	}

	// Validate options if provided
	if r.Options != nil {
		if err := r.Options.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate validates the DecodeOptions
func (o *DecodeOptions) Validate() error {
	if o.Format != "" {
		validFormats := []string{"structured", "json", "csv", "raw"}
		isValid := false
		for _, format := range validFormats {
			if o.Format == format {
				isValid = true
				break
			}
		}
		if !isValid {
			return validator.NewValidationError(validator.ErrCodeInvalidFormat, "invalid format").
				WithField("format", o.Format).
				WithDetail("valid_formats", validFormats)
		}
	}

	if o.MaxParts < 0 {
		return validator.NewValidationError(validator.ErrCodeInvalidInput, "max_parts cannot be negative").
			WithField("max_parts", o.MaxParts)
	}

	if o.Timeout < 0 {
		return validator.NewValidationError(validator.ErrCodeInvalidInput, "timeout cannot be negative").
			WithField("timeout", o.Timeout)
	}

	return nil
}

// Validate validates the EncodeRequest
func (r *EncodeRequest) Validate() error {
	if r.ItemData == nil {
		return validator.NewValidationError(validator.ErrCodeEmptyInput, "item_data is required")
	}

	if err := r.ItemData.Validate(); err != nil {
		return err
	}

	// Validate options if provided
	if r.Options != nil {
		if err := r.Options.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate validates the EncodeOptions
func (o *EncodeOptions) Validate() error {
	if o.Format != "" {
		validFormats := []string{"structured", "json", "csv"}
		isValid := false
		for _, format := range validFormats {
			if o.Format == format {
				isValid = true
				break
			}
		}
		if !isValid {
			return validator.NewValidationError(validator.ErrCodeInvalidFormat, "invalid format").
				WithField("format", o.Format).
				WithDetail("valid_formats", validFormats)
		}
	}

	return nil
}

// Validate validates the BatchDecodeRequest
func (r *BatchDecodeRequest) Validate() error {
	if len(r.SerialCodes) == 0 {
		return validator.NewValidationError(validator.ErrCodeEmptyInput, "serial_codes cannot be empty")
	}

	if len(r.SerialCodes) > 1000 {
		return validator.NewValidationError(validator.ErrCodeBatchSizeExceeded, "too many serial codes").
			WithField("count", len(r.SerialCodes)).
			WithDetail("max_count", 1000)
	}

	// Validate individual serial codes
	for i, code := range r.SerialCodes {
		if code == "" {
			return validator.NewValidationError(validator.ErrCodeEmptyInput, "serial_code cannot be empty").
				WithField("serial_codes", fmt.Sprintf("index %d", i))
		}
		if len(code) < 3 {
			return validator.NewValidationError(validator.ErrCodeInvalidLength, "serial_code too short").
				WithField("serial_codes", fmt.Sprintf("index %d", i))
		}
	}

	// Validate options if provided
	if r.Options != nil {
		if err := r.Options.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate validates the BatchDecodeOptions
func (o *BatchDecodeOptions) Validate() error {
	if o.IndividualOptions != nil {
		if err := o.IndividualOptions.Validate(); err != nil {
			return err
		}
	}

	if o.MaxConcurrency < 0 {
		return validator.NewValidationError(validator.ErrCodeInvalidInput, "max_concurrency cannot be negative").
			WithField("max_concurrency", o.MaxConcurrency)
	}

	if o.Timeout < 0 {
		return validator.NewValidationError(validator.ErrCodeInvalidInput, "timeout cannot be negative").
			WithField("timeout", o.Timeout)
	}

	if o.PerItemTimeout < 0 {
		return validator.NewValidationError(validator.ErrCodeInvalidInput, "per_item_timeout cannot be negative").
			WithField("per_item_timeout", o.PerItemTimeout)
	}

	return nil
}

// Validate validates the ValidateRequest
func (r *ValidateRequest) Validate() error {
	if r.SerialCode == "" {
		return validator.NewValidationError(validator.ErrCodeEmptyInput, "serial_code is required").
			WithField("serial_code", r.SerialCode)
	}

	if len(r.SerialCode) < 3 {
		return validator.NewValidationError(validator.ErrCodeInvalidLength, "serial_code too short").
			WithField("serial_code", r.SerialCode)
	}

	// Validate options if provided
	if r.Options != nil {
		if err := r.Options.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate validates the ValidateOptions
func (o *ValidateOptions) Validate() error {
	// All validation options are boolean, so no additional validation needed
	return nil
}

// Validate validates the ItemData
func (i *ItemData) Validate() error {
	if i.Level < 1 || i.Level > 100 {
		return validator.NewValidationError(validator.ErrCodeInvalidItemLevel, "invalid item level").
			WithField("level", i.Level).
			WithDetail("valid_range", "1-100")
	}

	if i.Type == "" {
		return validator.NewValidationError(validator.ErrCodeInvalidItemType, "item type is required").
			WithField("type", i.Type)
	}

	if i.Manufacturer == "" {
		return validator.NewValidationError(validator.ErrCodeInvalidManufacturer, "manufacturer is required").
			WithField("manufacturer", i.Manufacturer)
	}

	// Validate parts if provided
	for i, part := range i.Parts {
		if err := part.Validate(); err != nil {
			return validator.NewValidationError(validator.ErrCodeInvalidPartsData, "invalid part data").
				WithField("parts", fmt.Sprintf("index %d", i)).
				WithCause(err)
		}
	}

	return nil
}

// Validate validates the PartData
func (p *PartData) Validate() error {
	if p.Index < 0 {
		return validator.NewValidationError(validator.ErrCodeInvalidInput, "part index cannot be negative").
			WithField("index", p.Index)
	}

	return nil
}

// Helper methods

// HasOption checks if a specific decode option is enabled
func (r *DecodeRequest) HasOption(option string) bool {
	if r.Options == nil {
		return false
	}

	switch option {
	case "include_bitstream":
		return r.Options.IncludeBitstream
	case "include_tokens":
		return r.Options.IncludeTokens
	case "pretty_print":
		return r.Options.PrettyPrint
	case "validate_only":
		return r.Options.ValidateOnly
	case "include_stats":
		return r.Options.IncludeStats
	default:
		return false
	}
}

// GetFormat returns the format option, defaulting to "structured"
func (r *DecodeRequest) GetFormat() string {
	if r.Options != nil && r.Options.Format != "" {
		return r.Options.Format
	}
	return "structured"
}

// GetMaxParts returns the max parts option, defaulting to 0 (unlimited)
func (r *DecodeRequest) GetMaxParts() int {
	if r.Options != nil && r.Options.MaxParts > 0 {
		return r.Options.MaxParts
	}
	return 0 // Unlimited
}

// GetTimeout returns the timeout option, defaulting to 5000ms
func (r *DecodeRequest) GetTimeout() int {
	if r.Options != nil && r.Options.Timeout > 0 {
		return r.Options.Timeout
	}
	return 5000 // 5 seconds default
}

// IsEmpty returns true if the request has no meaningful data
func (r *DecodeRequest) IsEmpty() bool {
	return r.SerialCode == ""
}

// String returns a string representation of the request
func (r *DecodeRequest) String() string {
	return fmt.Sprintf("DecodeRequest{SerialCode: %s, Format: %s, IncludeBitstream: %v, IncludeTokens: %v}",
		r.SerialCode, r.GetFormat(), r.HasOption("include_bitstream"), r.HasOption("include_tokens"))
}

// Helper methods for EncodeRequest

// HasOption checks if a specific encode option is enabled
func (r *EncodeRequest) HasOption(option string) bool {
	if r.Options == nil {
		return false
	}

	switch option {
	case "optimize_size":
		return r.Options.OptimizeSize
	case "validate_parts":
		return r.Options.ValidateParts
	case "include_prefix":
		return r.Options.IncludePrefix
	case "include_metadata":
		// Include metadata is handled at the handler level, not part of EncodeOptions
		return false
	default:
		return false
	}
}

// GetTimeout returns the timeout option, defaulting to 5000ms
func (r *EncodeRequest) GetTimeout() int {
	// Encode requests don't have timeout in options currently
	// This could be added to EncodeOptions if needed
	return 5000 // 5 seconds default
}

// GetTargetVersion returns the target version option, defaulting to "1.0"
func (r *EncodeRequest) GetTargetVersion() string {
	if r.Options != nil && r.Options.TargetVersion != "" {
		return r.Options.TargetVersion
	}
	return "1.0"
}

// GetFormat returns the format option, defaulting to "structured"
func (r *EncodeRequest) GetFormat() string {
	if r.Options != nil && r.Options.Format != "" {
		return r.Options.Format
	}
	return "structured"
}

// IsEmpty returns true if the request has no meaningful data
func (r *EncodeRequest) IsEmpty() bool {
	return r.ItemData == nil
}

// String returns a string representation of the request
func (r *EncodeRequest) String() string {
	if r.ItemData == nil {
		return "EncodeRequest{empty}"
	}

	return fmt.Sprintf("EncodeRequest{Level: %d, Type: %s, Manufacturer: %s, Parts: %d, Format: %s}",
		r.ItemData.Level, r.ItemData.Type, r.ItemData.Manufacturer, len(r.ItemData.Parts), r.GetFormat())
}

// ValidateOptions contains options for validation
type ValidateOptions struct {
	// StrictMode enables strict validation rules
	StrictMode bool `json:"strict_mode,omitempty"`

	// IncludeDetails includes detailed validation information
	IncludeDetails bool `json:"include_details,omitempty"`

	// ValidateStructure attempts to decode the code for structural validation
	ValidateStructure bool `json:"validate_structure,omitempty"`

	// Timeout specifies maximum validation time in milliseconds
	Timeout int `json:"timeout,omitempty"`
}

// ValidateRequest represents a request to validate a serial code
type ValidateRequest struct {
	SerialCode string                 `json:"serial_code" binding:"required" example:"@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}"`
	Options    *ValidateOptions       `json:"options,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}