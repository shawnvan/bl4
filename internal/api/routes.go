package api

import (
	"github.com/gin-gonic/gin"
	"github.com/shawnvan/bl4/internal/api/handlers"
)

// Router configures all API routes with their handlers
type Router struct {
	engine            *gin.Engine
	decodeHandler     *handlers.DecodeHandler
	batchDecodeHandler *handlers.BatchDecodeHandler
}

// NewRouter creates a new router instance
func NewRouter(engine *gin.Engine) *Router {
	return &Router{
		engine:            engine,
		decodeHandler:     handlers.NewDecodeHandler(),
		batchDecodeHandler: handlers.NewBatchDecodeHandler(),
	}
}

// SetupRoutes configures all API routes
func (r *Router) SetupRoutes() {
	// Health check endpoint
	r.engine.GET("/health", r.healthCheckHandler())

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
		items.POST("/encode", r.encodeItemHandler()) // TODO: implement in US2
		items.POST("/validate", r.validateItemHandler()) // TODO: implement in US2

		// Batch operations
		items.POST("/batch/decode", r.batchDecodeHandler.HandleBatchDecode)
		items.POST("/batch/encode", r.batchEncodeHandler()) // TODO: implement in US2

		// Additional item endpoints
		items.GET("/decode/health", r.decodeHandler.HandleHealthCheck)
		items.GET("/decode/stats", r.decodeHandler.HandleStats)
	}
}

// setupAnalysisRoutes configures analysis-related routes
func (r *Router) setupAnalysisRoutes(group *gin.RouterGroup) {
	analysis := group.Group("/analysis")
	{
		analysis.POST("/pattern", r.patternAnalysisHandler()) // TODO: implement in US4
		analysis.POST("/generate", r.generateItemHandler())   // TODO: implement in US4
		analysis.GET("/stats", r.statsHandler())             // TODO: implement in US4
	}
}

// setupStaticRoutes configures static file and web interface routes
func (r *Router) setupStaticRoutes() {
	// Static files for web interface
	r.engine.Static("/static", "./web/static")
	r.engine.StaticFile("/favicon.ico", "./web/static/favicon.ico")

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

// encodeItemHandler is a placeholder for the encode endpoint (to be implemented in US2)
func (r *Router) encodeItemHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(501, gin.H{
			"error": gin.H{
				"code":    "NOT_IMPLEMENTED",
				"message": "Encode endpoint will be implemented in User Story 2",
				"todo":    "US2 T048: 实现物品序列化处理器",
			},
		})
	}
}

// validateItemHandler is a placeholder for the validate endpoint (to be implemented in US2)
func (r *Router) validateItemHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(501, gin.H{
			"error": gin.H{
				"code":    "NOT_IMPLEMENTED",
				"message": "Validate endpoint will be implemented in User Story 2",
				"todo":    "US2 T058: 创建物品验证处理器",
			},
		})
	}
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

// patternAnalysisHandler is a placeholder for pattern analysis (to be implemented in US4)
func (r *Router) patternAnalysisHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(501, gin.H{
			"error": gin.H{
				"code":    "NOT_IMPLEMENTED",
				"message": "Pattern analysis endpoint will be implemented in User Story 4",
				"todo":    "US4 T089: 实现模式分析处理器",
			},
		})
	}
}

// generateItemHandler is a placeholder for item generation (to be implemented in US4)
func (r *Router) generateItemHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(501, gin.H{
			"error": gin.H{
				"code":    "NOT_IMPLEMENTED",
				"message": "Item generation endpoint will be implemented in User Story 4",
				"todo":    "US4 T090: 实现物品生成处理器",
			},
		})
	}
}

// statsHandler is a placeholder for statistics endpoint (to be implemented in US4)
func (r *Router) statsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(501, gin.H{
			"error": gin.H{
				"code":    "NOT_IMPLEMENTED",
				"message": "Statistics endpoint will be implemented in User Story 4",
				"todo":    "US4 T091: 创建统计分析处理器",
			},
		})
	}
}

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
				"batch_decode":  "✅ Implemented",
				"encode":        "🚧 Coming in US2",
				"validate":      "🚧 Coming in US2",
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