package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSConfig contains configuration for CORS middleware
type CORSConfig struct {
	// AllowedOrigins is a list of origins that are allowed to make requests
	AllowedOrigins []string `json:"allowed_origins"`

	// AllowedMethods is a list of HTTP methods that are allowed
	AllowedMethods []string `json:"allowed_methods"`

	// AllowedHeaders is a list of HTTP headers that are allowed
	AllowedHeaders []string `json:"allowed_headers"`

	// ExposedHeaders is a list of headers that are exposed to the client
	ExposedHeaders []string `json:"exposed_headers"`

	// AllowCredentials indicates whether requests can include user credentials
	AllowCredentials bool `json:"allow_credentials"`

	// MaxAge indicates how long (in seconds) the results of a preflight request can be cached
	MaxAge int `json:"max_age"`
}

// DefaultCORSConfig returns a default CORS configuration
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
			"X-Request-ID",
			"X-Forwarded-For",
			"X-Forwarded-Proto",
		},
		ExposedHeaders: []string{
			"X-Total-Count",
			"X-Request-ID",
			"X-Processing-Time",
		},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}
}

// CORSMiddleware creates a CORS middleware
type CORSMiddleware struct {
	config *CORSConfig
}

// NewCORSMiddleware creates a new CORS middleware
func NewCORSMiddleware(config *CORSConfig) *CORSMiddleware {
	return &CORSMiddleware{
		config: config,
	}
}

// Middleware returns the Gin middleware function
func (m *CORSMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", m.getAllowedOrigin(c.Request.Header.Get("Origin")))
		c.Header("Access-Control-Allow-Methods", strings.Join(m.config.AllowedMethods, ", "))
		c.Header("Access-Control-Allow-Headers", strings.Join(m.config.AllowedHeaders, ", "))
		c.Header("Access-Control-Expose-Headers", strings.Join(m.config.ExposedHeaders, ", "))
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// getAllowedOrigin returns the appropriate allowed origin for the request
func (m *CORSMiddleware) getAllowedOrigin(requestOrigin string) string {
	// If wildcard is allowed, return the request origin
	for _, allowedOrigin := range m.config.AllowedOrigins {
		if allowedOrigin == "*" {
			return requestOrigin
		}
		if allowedOrigin == requestOrigin {
			return requestOrigin
		}
	}

	// Return the first allowed origin if request origin is not in the list
	if len(m.config.AllowedOrigins) > 0 {
		return m.config.AllowedOrigins[0]
	}

	return ""
}

// DevelopmentCORSConfig returns a CORS configuration suitable for development
func DevelopmentCORSConfig() *CORSConfig {
	return &CORSConfig{
	AllowedOrigins: []string{
			"http://localhost:3000",
			"http://localhost:8080",
			"http://localhost:3001",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:8080",
			"http://127.0.0.1:3001",
			"*",
		},
		AllowedMethods: []string{
			"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH",
			"HEAD",
		},
		AllowedHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
			"X-Request-ID",
			"X-Forwarded-For",
			"X-Forwarded-Proto",
			"X-API-Version",
			"X-Client-Version",
		},
		ExposedHeaders: []string{
			"X-Total-Count",
			"X-Request-ID",
			"X-Processing-Time",
			"X-API-Version",
			"X-Server-Version",
		},
		AllowCredentials: true,
		MaxAge:           86400,
	}
}

// ProductionCORSConfig returns a CORS configuration suitable for production
func ProductionCORSConfig() *CORSConfig {
	return &CORSConfig{
	AllowedOrigins: []string{
			"https://bl4.example.com",
			"https://api.bl4.example.com",
		},
		AllowedMethods: []string{
			"GET", "POST", "PUT", "DELETE", "OPTIONS",
		},
		AllowedHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
			"X-Request-ID",
			"X-API-Key",
		},
		ExposedHeaders: []string{
			"X-Total-Count",
			"X-Request-ID",
			"X-Processing-Time",
		},
		AllowCredentials: false,
		MaxAge:           3600, // 1 hour
	}
}

// CORSSetup configures CORS middleware based on environment
func CORSSetup(environment string) gin.HandlerFunc {
	var config *CORSConfig

	switch strings.ToLower(environment) {
	case "development", "dev", "debug":
		config = DevelopmentCORSConfig()
	case "production", "prod":
		config = ProductionCORSConfig()
	default:
		config = DefaultCORSConfig()
	}

	cors := NewCORSMiddleware(config)
	return cors.Middleware()
}