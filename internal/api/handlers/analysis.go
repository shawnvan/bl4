package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shawnvan/bl4/internal/api/services"
	"github.com/shawnvan/bl4/pkg/logger"
)

// AnalysisHandler handles analysis requests
type AnalysisHandler struct {
	analysisService *services.AnalysisService
	generatorService *services.GeneratorService
}

// NewAnalysisHandler creates a new analysis handler
func NewAnalysisHandler() *AnalysisHandler {
	// Initialize services with default configurations
	analysisConfig := services.DefaultAnalysisConfig()
	generatorConfig := services.DefaultGeneratorConfig()

	return &AnalysisHandler{
		analysisService:  services.NewAnalysisService(analysisConfig),
		generatorService: services.NewGeneratorService(generatorConfig),
	}
}

// NewAnalysisHandlerWithServices creates a new analysis handler with custom services
func NewAnalysisHandlerWithServices(analysisService *services.AnalysisService, generatorService *services.GeneratorService) *AnalysisHandler {
	return &AnalysisHandler{
		analysisService:  analysisService,
		generatorService: generatorService,
	}
}

// HandlePatternAnalysis handles POST /api/v1/analysis/pattern requests
func (h *AnalysisHandler) HandlePatternAnalysis(c *gin.Context) {
	startTime := time.Now()
	requestID := uuid.New().String()

	// Parse request
	var request services.AnalysisRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Sugar().Errorw("Invalid pattern analysis request",
			"request_id", requestID,
			"error", err,
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"success":     false,
			"request_id":  requestID,
			"error":       "Invalid request format",
			"error_type":  "validation_error",
			"status_code": http.StatusBadRequest,
		})
		return
	}

	logger.Sugar().Infow("Processing pattern analysis request",
		"request_id", requestID,
		"sample_size", len(request.SerialCodes),
	)

	// Validate request
	if len(request.SerialCodes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":     false,
			"request_id":  requestID,
			"error":       "No serial codes provided for analysis",
			"error_type":  "validation_error",
			"status_code": http.StatusBadRequest,
		})
		return
	}

	// Set default options if not provided
	if request.Options == nil {
		request.Options = &services.AnalysisOptions{
			IncludePatterns:      true,
			IncludeStatistics:    true,
			IncludeBitstream:     true,
			IncludeRarity:        true,
			IncludeManufacturers: true,
			IncludePartFrequency: true,
			DeepAnalysis:         true,
			MaxPatterns:          10,
			ConfidenceThreshold:  0.7,
		}
	}

	// Perform analysis
	result, err := h.analysisService.AnalyzePatterns(&request)
	if err != nil {
		logger.Sugar().Errorw("Pattern analysis failed",
			"request_id", requestID,
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":     false,
			"request_id":  requestID,
			"error":       "Analysis failed",
			"error_type":  "processing_error",
			"status_code": http.StatusInternalServerError,
			"details":     err.Error(),
		})
		return
	}

	// Log completion
	logger.Sugar().Infow("Pattern analysis completed",
		"request_id", requestID,
		"sample_size", result.SampleSize,
		"patterns_found", len(result.Patterns),
		"quality_score", result.QualityScore,
		"confidence", result.Confidence,
		"processing_time", time.Since(startTime).Milliseconds(),
	)

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"request_id":      requestID,
		"sample_size":     result.SampleSize,
		"patterns":        result.Patterns,
		"statistics":      result.Statistics,
		"bitstream_analysis": result.BitstreamAnalysis,
		"rarity_distribution": result.RarityDistribution,
		"manufacturer_distribution": result.ManufacturerDistribution,
		"part_frequency":  result.PartFrequency,
		"quality_score":   result.QualityScore,
		"confidence":      result.Confidence,
		"insights":        result.Insights,
		"processing_time": result.ProcessTime,
		"timestamp":       result.Timestamp.Format(time.RFC3339),
	})
}

// HandleGenerateItems handles POST /api/v1/analysis/generate requests
func (h *AnalysisHandler) HandleGenerateItems(c *gin.Context) {
	startTime := time.Now()
	requestID := uuid.New().String()

	// Parse request
	var request services.GenerationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Sugar().Errorw("Invalid generation request",
			"request_id", requestID,
			"error", err,
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"success":     false,
			"request_id":  requestID,
			"error":       "Invalid request format",
			"error_type":  "validation_error",
			"status_code": http.StatusBadRequest,
		})
		return
	}

	logger.Sugar().Infow("Processing item generation request",
		"request_id", requestID,
		"count", request.Count,
	)

	// Validate request
	if request.Count <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":     false,
			"request_id":  requestID,
			"error":       "Count must be greater than 0",
			"error_type":  "validation_error",
			"status_code": http.StatusBadRequest,
		})
		return
	}

	if request.Count > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":     false,
			"request_id":  requestID,
			"error":       "Count cannot exceed 1000",
			"error_type":  "validation_error",
			"status_code": http.StatusBadRequest,
		})
		return
	}

	// Set default options if not provided
	if request.Options == nil {
		request.Options = &services.GenerationOptions{
			IncludeSerialCodes:   true,
			IncludeMetadata:      true,
			Format:               "json",
			ValidateItems:        true,
			RealisticGeneration:  true,
		}
	}

	// Generate items
	response, err := h.generatorService.GenerateItems(&request)
	if err != nil {
		logger.Sugar().Errorw("Item generation failed",
			"request_id", requestID,
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":     false,
			"request_id":  requestID,
			"error":       "Generation failed",
			"error_type":  "processing_error",
			"status_code": http.StatusInternalServerError,
			"details":     err.Error(),
		})
		return
	}

	// Prepare response data
	responseData := gin.H{
		"success":       true,
		"request_id":    requestID,
		"count":         response.Count,
		"items":         response.Items,
		"process_time":  response.ProcessTime,
		"timestamp":     response.Timestamp.Format(time.RFC3339),
	}

	// Include statistics if available
	if response.Statistics != nil {
		responseData["statistics"] = response.Statistics
	}

	// Include validation results if available
	if len(response.ValidationResults) > 0 {
		responseData["validation_summary"] = gin.H{
			"total_items":   len(response.ValidationResults),
			"valid_items":   h.countValidItems(response.ValidationResults),
			"invalid_items": h.countInvalidItems(response.ValidationResults),
		}
		responseData["validation_results"] = response.ValidationResults
	}

	// Include errors if any
	if len(response.Errors) > 0 {
		responseData["errors"] = response.Errors
		responseData["error_count"] = len(response.Errors)
	}

	// Log completion
	logger.Sugar().Infow("Item generation completed",
		"request_id", requestID,
		"items_generated", len(response.Items),
		"errors", len(response.Errors),
		"process_time_ms", time.Since(startTime).Milliseconds(),
	)

	c.JSON(http.StatusOK, responseData)
}

// HandleAnalyzeAndGenerate handles POST /api/v1/analysis/analyze-and-generate requests
func (h *AnalysisHandler) HandleAnalyzeAndGenerate(c *gin.Context) {
	startTime := time.Now()
	requestID := uuid.New().String()

	// Parse request
	var request struct {
		// Input serial codes for analysis
		SerialCodes []string `json:"serial_codes" binding:"required"`

		// Generation parameters
		GenerationParams *services.GenerationRequest `json:"generation_params,omitempty"`

		// Analysis options
		AnalysisOptions *services.AnalysisOptions `json:"analysis_options,omitempty"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Sugar().Errorw("Invalid analyze and generate request",
			"request_id", requestID,
			"error", err,
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"success":     false,
			"request_id":  requestID,
			"error":       "Invalid request format",
			"error_type":  "validation_error",
			"status_code": http.StatusBadRequest,
		})
		return
	}

	logger.Sugar().Infow("Processing analyze and generate request",
		"request_id", requestID,
		"input_codes", len(request.SerialCodes),
		"generate_count", func() int {
			if request.GenerationParams != nil {
				return request.GenerationParams.Count
			}
			return 0
		}(),
	)

	// Validate request
	if len(request.SerialCodes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":     false,
			"request_id":  requestID,
			"error":       "No serial codes provided for analysis",
			"error_type":  "validation_error",
			"status_code": http.StatusBadRequest,
		})
		return
	}

	// Step 1: Analyze input serial codes
	analysisRequest := &services.AnalysisRequest{
		SerialCodes: request.SerialCodes,
		Options:     request.AnalysisOptions,
	}

	if analysisRequest.Options == nil {
		analysisRequest.Options = &services.AnalysisOptions{
			IncludePatterns:      true,
			IncludeStatistics:    true,
			IncludeManufacturers: true,
			IncludePartFrequency: true,
		}
	}

	analysisResult, err := h.analysisService.AnalyzePatterns(analysisRequest)
	if err != nil {
		logger.Sugar().Errorw("Analysis phase failed",
			"request_id", requestID,
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":     false,
			"request_id":  requestID,
			"error":       "Analysis failed",
			"error_type":  "processing_error",
			"status_code": http.StatusInternalServerError,
			"details":     err.Error(),
		})
		return
	}

	// Step 2: Generate new items based on analysis (if requested)
	var generationResponse *services.GenerationResponse
	if request.GenerationParams != nil && request.GenerationParams.Count > 0 {
		// Adjust generation parameters based on analysis
		h.adjustGenerationParamsFromAnalysis(request.GenerationParams, analysisResult)

		generationResponse, err = h.generatorService.GenerateItems(request.GenerationParams)
		if err != nil {
			logger.Sugar().Errorw("Generation phase failed",
				"request_id", requestID,
				"error", err,
			)
			// Don't fail the entire request, just note the generation error
			generationResponse = &services.GenerationResponse{
				RequestID:   requestID,
				Timestamp:   time.Now(),
				Count:       0,
				Items:       []*services.GeneratedItem{},
				Errors:      []string{err.Error()},
				ProcessTime: "0s",
			}
		}
	}

	// Prepare combined response
	response := gin.H{
		"success":       true,
		"request_id":    requestID,
		"timestamp":     time.Now().Format(time.RFC3339),
		"process_time":  time.Since(startTime).String(),
		"analysis": gin.H{
			"sample_size":     analysisResult.SampleSize,
			"patterns":        analysisResult.Patterns,
			"statistics":      analysisResult.Statistics,
			"quality_score":   analysisResult.QualityScore,
			"confidence":      analysisResult.Confidence,
			"insights":        analysisResult.Insights,
		},
	}

	// Include generation results if available
	if generationResponse != nil {
		generationData := gin.H{
			"count":    generationResponse.Count,
			"items":    generationResponse.Items,
			"errors":   generationResponse.Errors,
		}

		if generationResponse.Statistics != nil {
			generationData["statistics"] = generationResponse.Statistics
		}

		response["generation"] = generationData
	}

	// Log completion
	logger.Sugar().Infow("Analyze and generate completed",
		"request_id", requestID,
		"analysis_quality", analysisResult.QualityScore,
		"items_generated", func() int {
			if generationResponse != nil {
				return len(generationResponse.Items)
			}
			return 0
		}(),
		"process_time_ms", time.Since(startTime).Milliseconds(),
	)

	c.JSON(http.StatusOK, response)
}

// HandleStats handles GET /api/v1/analysis/stats requests
func (h *AnalysisHandler) HandleStats(c *gin.Context) {
	requestID := uuid.New().String()

	logger.Sugar().Infow("Processing analysis stats request",
		"request_id", requestID,
	)

	// Get service statistics
	stats := gin.H{
		"analysis_service": gin.H{
			"cache_enabled":     h.analysisService != nil,
			"patterns_loaded":   h.getPatternsCount(),
			"service_status":    "active",
		},
		"generator_service": gin.H{
			"realistic_mode":    h.generatorService != nil,
			"legendary_enabled": h.getLegendaryEnabled(),
			"service_status":    "active",
		},
		"api_info": gin.H{
			"version":     "1.0.0",
			"timestamp":   time.Now().Format(time.RFC3339),
			"endpoints": gin.H{
				"pattern_analysis":  "/api/v1/analysis/pattern",
				"generate_items":    "/api/v1/analysis/generate",
				"analyze_and_generate": "/api/v1/analysis/analyze-and-generate",
				"stats":             "/api/v1/analysis/stats",
			},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"request_id": requestID,
		"stats":      stats,
	})
}

// HandleBatchAnalysis handles POST /api/v1/analysis/batch requests
func (h *AnalysisHandler) HandleBatchAnalysis(c *gin.Context) {
	startTime := time.Now()
	requestID := uuid.New().String()

	// Parse request
	var request struct {
		Batches []struct {
			Name        string   `json:"name"`
			SerialCodes []string `json:"serial_codes"`
			Options     *services.AnalysisOptions `json:"options,omitempty"`
		} `json:"batches" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Sugar().Errorw("Invalid batch analysis request",
			"request_id", requestID,
			"error", err,
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"success":     false,
			"request_id":  requestID,
			"error":       "Invalid request format",
			"error_type":  "validation_error",
			"status_code": http.StatusBadRequest,
		})
		return
	}

	logger.Sugar().Infow("Processing batch analysis request",
		"request_id", requestID,
		"batch_count", len(request.Batches),
	)

	// Validate request
	if len(request.Batches) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":     false,
			"request_id":  requestID,
			"error":       "No batches provided",
			"error_type":  "validation_error",
			"status_code": http.StatusBadRequest,
		})
		return
	}

	if len(request.Batches) > 10 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":     false,
			"request_id":  requestID,
			"error":       "Too many batches (maximum 10)",
			"error_type":  "validation_error",
			"status_code": http.StatusBadRequest,
		})
		return
	}

	// Process each batch
	results := make([]gin.H, 0, len(request.Batches))
	errors := make([]string, 0)

	for i, batch := range request.Batches {
		batchStartTime := time.Now()

		// Validate batch
		if len(batch.SerialCodes) == 0 {
			errors = append(errors, fmt.Sprintf("Batch %d: No serial codes provided", i))
			continue
		}

		// Create analysis request
		analysisRequest := &services.AnalysisRequest{
			SerialCodes: batch.SerialCodes,
			Options:     batch.Options,
		}

		if analysisRequest.Options == nil {
			analysisRequest.Options = &services.AnalysisOptions{
				IncludePatterns:      true,
				IncludeStatistics:    true,
				IncludeManufacturers: true,
			}
		}

		// Perform analysis
		result, err := h.analysisService.AnalyzePatterns(analysisRequest)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Batch %d: %v", i, err))
			continue
		}

		// Create batch result
		batchResult := gin.H{
			"name":          batch.Name,
			"sample_size":   result.SampleSize,
			"patterns":      result.Patterns,
			"statistics":    result.Statistics,
			"quality_score": result.QualityScore,
			"confidence":    result.Confidence,
			"insights":      result.Insights,
			"process_time":  result.ProcessTime,
			"timestamp":     result.Timestamp.Format(time.RFC3339),
		}

		results = append(results, batchResult)

		logger.Sugar().Debugw("Batch analysis completed",
			"request_id", requestID,
			"batch_name", batch.Name,
			"batch_index", i,
			"process_time_ms", time.Since(batchStartTime).Milliseconds(),
		)
	}

	// Prepare response
	response := gin.H{
		"success":       true,
		"request_id":    requestID,
		"batch_count":   len(request.Batches),
		"successful_batches": len(results),
		"failed_batches":     len(errors),
		"results":       results,
		"process_time":  time.Since(startTime).String(),
		"timestamp":     time.Now().Format(time.RFC3339),
	}

	if len(errors) > 0 {
		response["errors"] = errors
	}

	// Log completion
	logger.Sugar().Infow("Batch analysis completed",
		"request_id", requestID,
		"total_batches", len(request.Batches),
		"successful", len(results),
		"failed", len(errors),
		"process_time_ms", time.Since(startTime).Milliseconds(),
	)

	c.JSON(http.StatusOK, response)
}

// Utility methods

func (h *AnalysisHandler) adjustGenerationParamsFromAnalysis(genParams *services.GenerationRequest, analysisResult *services.AnalysisResult) {
	// If no constraints specified, derive from analysis
	if genParams.Constraints == nil {
		genParams.Constraints = &services.GenerationConstraints{}
	}

	constraints := genParams.Constraints

	// Derive level constraints from analysis
	if analysisResult.Statistics != nil {
		if constraints.MinLevel == nil && analysisResult.Statistics.MinLevel > 0 {
			constraints.MinLevel = &analysisResult.Statistics.MinLevel
		}
		if constraints.MaxLevel == nil && analysisResult.Statistics.MaxLevel > 0 {
			constraints.MaxLevel = &analysisResult.Statistics.MaxLevel
		}
	}

	// Derive manufacturer constraints from analysis
	if len(analysisResult.ManufacturerDistribution) > 0 && len(constraints.Manufacturers) == 0 {
		// Find the most common manufacturer
		maxCount := 0
		var topManufacturer string
		for manufacturer, count := range analysisResult.ManufacturerDistribution {
			if count > maxCount {
				maxCount = count
				topManufacturer = manufacturer
			}
		}
		if topManufacturer != "" {
			constraints.Manufacturers = []string{topManufacturer}
		}
	}

	// Derive type constraints from analysis
	if analysisResult.Statistics != nil && analysisResult.Statistics.UniqueTypes > 0 && len(constraints.Types) == 0 {
		// For simplicity, we'll use the most common type if we could extract it
		// In a real implementation, this would come from more detailed analysis
		constraints.Types = []string{"pistol"} // Default to pistol
	}

	// Derive part constraints from analysis
	if analysisResult.Statistics != nil {
		if constraints.MinParts == nil && analysisResult.Statistics.MinParts > 0 {
			constraints.MinParts = &analysisResult.Statistics.MinParts
		}
		if constraints.MaxParts == nil && analysisResult.Statistics.MaxParts > 0 {
			constraints.MaxParts = &analysisResult.Statistics.MaxParts
		}
	}
}

func (h *AnalysisHandler) countValidItems(validations []*services.ValidationResult) int {
	count := 0
	for _, validation := range validations {
		if validation.Valid {
			count++
		}
	}
	return count
}

func (h *AnalysisHandler) countInvalidItems(validations []*services.ValidationResult) int {
	count := 0
	for _, validation := range validations {
		if !validation.Valid {
			count++
		}
	}
	return count
}

func (h *AnalysisHandler) getPatternsCount() int {
	// Return the number of known patterns
	// In a real implementation, this would query the service
	return 10 // Placeholder
}

func (h *AnalysisHandler) getLegendaryEnabled() bool {
	// Return whether legendary generation is enabled
	// In a real implementation, this would query the service config
	return true // Placeholder
}