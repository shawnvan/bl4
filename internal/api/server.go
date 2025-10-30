package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shawnvan/bl4/pkg/logger"
	"github.com/shawnvan/bl4/pkg/validator"
	"go.uber.org/zap"
)

// Server represents the HTTP API server
type Server struct {
	config     *Config
	engine     *gin.Engine
	httpServer *http.Server
	started    bool
}

// Config holds the server configuration
type Config struct {
	Port             int           `json:"port" yaml:"port"`
	Host             string        `json:"host" yaml:"host"`
	ReadTimeout      time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout     time.Duration `json:"write_timeout" yaml:"write_timeout"`
	IdleTimeout      time.Duration `json:"idle_timeout" yaml:"idle_timeout"`
	MaxHeaderBytes   int           `json:"max_header_bytes" yaml:"max_header_bytes"`
	EnableMetrics    bool          `json:"enable_metrics" yaml:"enable_metrics"`
	EnableLogging    bool          `json:"enable_logging" yaml:"enable_logging"`
	EnableCORS       bool          `json:"enable_cors" yaml:"enable_cors"`
	TrustProxy       bool          `json:"trust_proxy" yaml:"trust_proxy"`
	ShutdownTimeout  time.Duration `json:"shutdown_timeout" yaml:"shutdown_timeout"`
}

// DefaultConfig returns a default server configuration
func DefaultConfig() *Config {
	return &Config{
		Port:            8080,
		Host:            "0.0.0.0",
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		IdleTimeout:     120 * time.Second,
		MaxHeaderBytes:  1 << 20, // 1MB
		EnableMetrics:   true,
		EnableLogging:   true,
		EnableCORS:      true,
		TrustProxy:      false,
		ShutdownTimeout: 30 * time.Second,
	}
}

// NewServer creates a new HTTP server instance
func NewServer(config *Config) *Server {
	if config == nil {
		config = DefaultConfig()
	}

	// Set Gin mode based on environment
	if gin.Mode() == gin.DebugMode {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	server := &Server{
		config: config,
		engine: engine,
	}

	server.setupMiddleware()
	server.setupRoutes()

	return server
}

// setupMiddleware configures server middleware
func (s *Server) setupMiddleware() {
	// Recovery middleware
	s.engine.Use(gin.Recovery())

	// Logging middleware
	if s.config.EnableLogging {
		s.engine.Use(s.loggingMiddleware())
	}

	// CORS middleware
	if s.config.EnableCORS {
		s.engine.Use(s.corsMiddleware())
	}

	// Request timeout middleware
	s.engine.Use(s.timeoutMiddleware())

	// Error handling middleware
	s.engine.Use(s.errorHandlingMiddleware())

	// Metrics middleware
	if s.config.EnableMetrics {
		s.engine.Use(s.metricsMiddleware())
	}
}

// setupRoutes configures server routes
func (s *Server) setupRoutes() {
	// Create router instance and setup all routes
	router := NewRouter(s.engine)
	router.SetupRoutes()
}

// Start starts the HTTP server
func (s *Server) Start() error {
	if s.started {
		return fmt.Errorf("server is already running")
	}

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	s.httpServer = &http.Server{
		Addr:           addr,
		Handler:        s.engine,
		ReadTimeout:    s.config.ReadTimeout,
		WriteTimeout:   s.config.WriteTimeout,
		IdleTimeout:    s.config.IdleTimeout,
		MaxHeaderBytes: s.config.MaxHeaderBytes,
	}

	logger.Sugar().Infow("Starting HTTP server",
		"address", addr,
		"read_timeout", s.config.ReadTimeout,
		"write_timeout", s.config.WriteTimeout,
		"max_header_bytes", s.config.MaxHeaderBytes,
	)

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}

	s.started = true
	return nil
}

// Stop gracefully shuts down the server
func (s *Server) Stop(ctx context.Context) error {
	if !s.started || s.httpServer == nil {
		return nil
	}

	logger.Sugar().Infow("Shutting down HTTP server", "timeout", s.config.ShutdownTimeout)

	shutdownCtx, cancel := context.WithTimeout(ctx, s.config.ShutdownTimeout)
	defer cancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	s.started = false
	logger.Sugar().Info("HTTP server stopped successfully")
	return nil
}

// GetEngine returns the Gin engine (for testing and advanced configuration)
func (s *Server) GetEngine() *gin.Engine {
	return s.engine
}

// GetConfig returns the server configuration
func (s *Server) GetConfig() *Config {
	return s.config
}

// IsStarted returns whether the server is currently running
func (s *Server) IsStarted() bool {
	return s.started
}

// Middleware implementations

func (s *Server) loggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logger.Sugar().Infow("HTTP Request",
			"method", param.Method,
			"path", param.Path,
			"status", param.StatusCode,
			"latency", param.Latency,
			"client_ip", param.ClientIP,
			"user_agent", param.Request.UserAgent(),
			"error", param.ErrorMessage,
		)
		return ""
	})
}

func (s *Server) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func (s *Server) timeoutMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use a context with timeout for the request
		ctx, cancel := context.WithTimeout(c.Request.Context(), s.config.WriteTimeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func (s *Server) errorHandlingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Handle any errors that occurred during request processing
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			// Try to convert to ValidationError
			if ve, ok := err.(*validator.ValidationError); ok {
				c.JSON(ve.Code.HTTPStatus(), gin.H{
					"error": gin.H{
						"code":    ve.Code.String(),
						"message": ve.Message,
						"details": ve.Details,
						"field":   ve.Field,
						"value":   ve.Value,
					},
				})
				return
			}

			// Default error response
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_SERVER_ERROR",
					"message": "An internal error occurred",
				},
			})

			// Log the error
			logger.Logger.Error("Request error",
				zap.Error(err),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.String("client_ip", c.ClientIP()),
			)
		}
	}
}

func (s *Server) metricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		duration := time.Since(start)

		// Log metrics (in a real implementation, you'd use a metrics library)
		logger.Sugar().Debugw("Request metrics",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration", duration,
			"size", c.Writer.Size(),
		)
	}
}

// Server helper methods

// GetRouteInfo returns information about all configured routes
func (s *Server) GetRouteInfo() gin.H {
	router := NewRouter(s.engine)
	return router.GetRouteInfo()
}