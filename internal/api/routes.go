package api

import (
	"github.com/gin-gonic/gin"
	"github.com/shawnvan/bl4/internal/api/handlers"
	"github.com/shawnvan/bl4/internal/api/middleware"
)

// Router configures all API routes with their handlers
type Router struct {
	engine               *gin.Engine
	decodeHandler        *handlers.DecodeHandler
	encodeHandler        *handlers.EncodeHandler
	validateHandler      *handlers.ValidateHandler
	healthHandler       *handlers.HealthHandler
	batchDecodeHandler  *handlers.BatchDecodeHandler
	analysisHandler     *handlers.AnalysisHandler
	rateLimiter          *middleware.RateLimiter
	batchLogger          *middleware.BatchLogger
}

// NewRouter creates a new router instance
func NewRouter(engine *gin.Engine) *Router {
	// Initialize middleware
	rateLimiterConfig := middleware.DefaultRateLimiterConfig()
	rateLimiter := middleware.NewRateLimiter(rateLimiterConfig)
	batchLoggerConfig := middleware.DefaultBatchLoggingConfig()
	batchLogger := middleware.NewBatchLogger(batchLoggerConfig)

	return &Router{
		engine:              engine,
		decodeHandler:       handlers.NewDecodeHandler(),
		encodeHandler:       handlers.NewEncodeHandler(),
		validateHandler:     handlers.NewValidateHandler(),
		healthHandler:       handlers.NewHealthHandler(),
		batchDecodeHandler: handlers.NewBatchDecodeHandler(),
		analysisHandler:    handlers.NewAnalysisHandler(),
		rateLimiter:         rateLimiter,
		batchLogger:         batchLogger,
	}
}

// SetupRoutes configures all API routes
func (r *Router) SetupRoutes() {
	// Apply CORS middleware first
	r.engine.Use(middleware.CORSSetup("development"))

	// Apply global rate limiting middleware
	r.engine.Use(r.rateLimiter.RateLimit())

	// Health check endpoints
	r.engine.GET("/health", r.healthHandler.HandleHealth)
	r.engine.GET("/health/readiness", r.healthHandler.HandleReadiness)
	r.engine.GET("/health/liveness", r.healthHandler.HandleLiveness)

	// API version 1 routes
	v1 := r.engine.Group("/api/v1")
	{
		// Item codec endpoints
		r.setupItemRoutes(v1)

		// Analysis endpoints
		r.setupAnalysisRoutes(v1)

		// Documentation endpoint
		v1.GET("/docs", r.docsHandler())
	}

	// Static files for web interface
	r.setupStaticRoutes()
}

// setupItemRoutes configures item-related routes
func (r *Router) setupItemRoutes(group *gin.RouterGroup) {
	items := group.Group("/items")
	{
		// Single item operations
		items.POST("/decode", r.decodeHandler.HandleDecode)
		items.POST("/encode", r.encodeHandler.HandleEncode)
		items.POST("/validate", r.validateHandler.HandleValidate)

		// Batch validation
		items.POST("/validate/batch", r.validateHandler.HandleValidateBatch)

		// Batch operations with specialized middleware
		batch := items.Group("/batch")
		batch.Use(r.batchLogger.BatchLoggingMiddleware())
		batch.Use(middleware.BatchProcessingMiddleware(r.rateLimiter))
		{
			// Standard batch decode endpoint
			batch.POST("/decode", r.batchDecodeHandler.HandleBatchDecode)

			// Enhanced batch decode endpoint with progress tracking
			batch.POST("/decode/enhanced", r.batchDecodeHandler.HandleBatchDecodeEnhanced)

			// Batch management endpoints
			batch.DELETE("/:id/cancel", r.batchDecodeHandler.HandleBatchCancel)
			batch.GET("/:id/progress", r.batchDecodeHandler.HandleBatchProgress)

			batch.POST("/encode", r.batchEncodeHandler()) // TODO: implement in US2
		}

		// Additional item endpoints
		items.GET("/decode/health", r.decodeHandler.HandleHealthCheck)
		items.GET("/decode/stats", r.decodeHandler.HandleStats)
		items.GET("/encode/health", r.encodeHandler.HandleHealthCheck)
	}
}

// setupAnalysisRoutes configures analysis-related routes
func (r *Router) setupAnalysisRoutes(group *gin.RouterGroup) {
	analysis := group.Group("/analysis")
	{
		// Pattern analysis endpoint
		analysis.POST("/pattern", r.analysisHandler.HandlePatternAnalysis)

		// Item generation endpoint
		analysis.POST("/generate", r.analysisHandler.HandleGenerateItems)

		// Combined analyze and generate endpoint
		analysis.POST("/analyze-and-generate", r.analysisHandler.HandleAnalyzeAndGenerate)

		// Service statistics endpoint
		analysis.GET("/stats", r.analysisHandler.HandleStats)

		// Batch analysis endpoint
		analysis.POST("/batch", r.analysisHandler.HandleBatchAnalysis)
	}
}

// setupStaticRoutes configures static file and web interface routes
func (r *Router) setupStaticRoutes() {
	// Static files for web interface
	r.engine.Static("/static", "./web/static")
	r.engine.StaticFile("/favicon.ico", "./web/static/favicon.ico")

	// API documentation
	r.engine.StaticFile("/api-docs", "./web/api-docs.html")
	r.engine.StaticFile("/docs.html", "./web/api-docs.html")

	// Web interface
	r.engine.LoadHTMLGlob("web/templates/*")
	r.engine.GET("/", r.webInterfaceHandler())
}

// Route handlers

// healthCheckHandler provides a simple health check endpoint
func (r *Router) healthCheckHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"timestamp": getCurrentTimestamp(),
			"version":   "1.0.0",
			"service":   "bl4-item-codec",
			"endpoints": gin.H{
				"decode":        "/api/v1/items/decode",
				"batch_decode":  "/api/v1/items/batch/decode",
				"health":        "/health",
				"docs":          "/api/v1/docs",
			},
		})
	}
}

// encodeItemHandler has been replaced by r.encodeHandler.HandleEncode
// This function is kept for backward compatibility but should not be used
func (r *Router) encodeItemHandler() gin.HandlerFunc {
	return r.encodeHandler.HandleEncode
}

// validateItemHandler has been replaced by r.encodeHandler.HandleValidate
// This function is kept for backward compatibility but should not be used
func (r *Router) validateItemHandler() gin.HandlerFunc {
	return r.encodeHandler.HandleValidate
}

// batchEncodeHandler is a placeholder for the batch encode endpoint (to be implemented in US2)
func (r *Router) batchEncodeHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(501, gin.H{
			"error": gin.H{
				"code":    "NOT_IMPLEMENTED",
				"message": "Batch encode endpoint will be implemented in User Story 2",
				"todo":    "US2 T059: 创建批量编码处理器",
			},
		})
	}
}

// Analysis handlers have been implemented in internal/api/handlers/analysis.go
// and are now directly referenced in setupAnalysisRoutes()

// docsHandler provides API documentation information
func (r *Router) docsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{
			"title":       "BL4 Item Codec API",
			"version":     "1.0.0",
			"description": "Borderlands 4 item serial code codec API",
			"base_url":    getBaseURL(c),
			"endpoints": gin.H{
				"decode": gin.H{
					"method": "POST",
					"path":   "/api/v1/items/decode",
					"description": "Decode a BL4 item serial code",
					"request_body": gin.H{
						"serial_code": "string (required)",
						"options": gin.H{
							"include_bitstream": "boolean (optional)",
							"include_tokens":    "boolean (optional)",
							"include_raw_data":   "boolean (optional)",
							"strict_validation": "boolean (optional)",
						},
					},
					"example": gin.H{
						"serial_code": "@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}",
						"options": gin.H{
							"include_bitstream": true,
							"include_tokens":    true,
						},
					},
				},
				"encode": gin.H{
					"method": "POST",
					"path":   "/api/v1/items/encode",
					"description": "Encode structured item data to a BL4 item serial code",
					"request_body": gin.H{
						"item_data": gin.H{
							"level": "int (required, 1-100)",
							"type": "string (required)",
							"manufacturer": "string (required)",
							"parts": "array of part objects (optional)",
							"name": "string (optional)",
							"description": "string (optional)",
						},
						"options": gin.H{
							"include_metadata": "boolean (optional)",
							"optimize_size": "boolean (optional)",
							"target_version": "string (optional)",
						},
					},
					"example": gin.H{
						"item_data": gin.H{
							"level": 24,
							"type": "pistol",
							"manufacturer": "maliwan",
							"parts": []gin.H{
								{"index": 0, "value": 1234},
								{"index": 1, "value": 5678},
							},
						},
						"options": gin.H{
							"optimize_size": false,
						},
					},
				},
				"validate": gin.H{
					"method": "POST",
					"path":   "/api/v1/items/validate",
					"description": "Validate item data without encoding",
					"request_body": "same as encode endpoint",
				},
				"batch_decode": gin.H{
					"method": "POST",
					"path":   "/api/v1/items/batch/decode",
					"description": "Decode multiple BL4 item serial codes",
					"request_body": gin.H{
						"serial_codes": "[]string (required)",
						"options": "same as single decode",
					},
				},
				"health": gin.H{
					"method": "GET",
					"path":   "/health",
					"description": "API health check",
				},
			},
			"status": map[string]string{
				"decode":        "✅ Implemented",
				"encode":        "✅ Implemented",
				"validate":      "✅ Implemented",
				"batch_decode":  "✅ Implemented",
				"batch_encode":  "🚧 Coming in US2",
				"analysis":      "🚧 Coming in US4",
			},
		})
	}
}

// webInterfaceHandler serves the main web interface
func (r *Router) webInterfaceHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(200, "index.html", gin.H{
			"title":       "BL4 Item Codec",
			"description": "Borderlands 4 item serial code decoder and encoder",
			"version":     "1.0.0",
			"api_base":    getBaseURL(c),
		})
	}
}

// Helper functions

// getCurrentTimestamp returns the current timestamp in a consistent format
func getCurrentTimestamp() string {
	return "2025-01-30T12:00:00Z" // Placeholder - would use time.Now().UTC().Format(time.RFC3339)
}

// getBaseURL constructs the base URL from the request
func getBaseURL(c *gin.Context) string {
	scheme := "http"
	if c.GetHeader("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}

	host := c.GetHeader("X-Forwarded-Host")
	if host == "" {
		host = c.Request.Host
	}

	return scheme + "://" + host
}

// SetupRoutesWithMiddleware configures routes with additional middleware
func (r *Router) SetupRoutesWithMiddleware(middleware ...gin.HandlerFunc) {
	// Apply global middleware
	for _, mw := range middleware {
		r.engine.Use(mw)
	}

	// Setup routes
	r.SetupRoutes()
}

// GetRouteInfo returns information about all configured routes
func (r *Router) GetRouteInfo() gin.H {
	routes := r.engine.Routes()

	var endpoints []gin.H
	for _, route := range routes {
		endpoints = append(endpoints, gin.H{
			"method": route.Method,
			"path":   route.Path,
			"handler": getHandlerName(route.Handler),
		})
	}

	return gin.H{
		"total_routes": len(routes),
		"endpoints":    endpoints,
		"groups": gin.H{
			"items":    "/api/v1/items/*",
			"analysis": "/api/v1/analysis/*",
			"health":   "/health",
			"docs":     "/api/v1/docs",
			"web":      "/",
		},
	}
}

// getHandlerName extracts a readable name from the handler function
func getHandlerName(handler interface{}) string {
	// This is a simplified implementation
	// In a production environment, you might want to use reflection or other methods
	if _, ok := handler.(gin.HandlerFunc); ok {
		return "gin_handler"
	}
	if s, ok := handler.(string); ok {
		return s
	}
	return "handler"
}