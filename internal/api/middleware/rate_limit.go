package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shawnvan/bl4/pkg/logger"
)

// RateLimiterConfig contains configuration for rate limiting
type RateLimiterConfig struct {
	// Global requests per second across all endpoints
	GlobalRPS int

	// Endpoint-specific rate limits
	EndpointLimits map[string]int

	// Batch processing specific limits
	BatchRPS      int // Requests per second for batch endpoints
	BurstSize     int // Maximum burst size
	MaxBatchSize  int // Maximum items allowed in a single batch

	// Client-specific rate limiting
	EnableClientLimiting bool
	ClientRPS            int // Per-client rate limit

	// Rate limiting window
	Window time.Duration
}

// DefaultRateLimiterConfig returns a default configuration
func DefaultRateLimiterConfig() *RateLimiterConfig {
	return &RateLimiterConfig{
		GlobalRPS:             1000, // 1000 requests per second globally
		EndpointLimits:        map[string]int{
			"/api/v1/items/decode":     100, // 100 RPS for single decode
			"/api/v1/items/batch/decode": 10, // 10 RPS for batch decode
			"/api/v1/items/encode":     100, // 100 RPS for single encode
		},
		BatchRPS:              10,    // 10 batch requests per second
		BurstSize:             20,    // Allow bursts of up to 20 requests
		MaxBatchSize:          1000,  // Maximum 1000 items per batch
		EnableClientLimiting:  true,
		ClientRPS:             50,    // 50 requests per second per client
		Window:                time.Minute,
	}
}

// TokenBucket represents a token bucket for rate limiting
type TokenBucket struct {
	capacity      int64
	tokens        int64
	refillRate    int64
	lastRefill    time.Time
	mu            sync.Mutex
}

// NewTokenBucket creates a new token bucket
func NewTokenBucket(capacity, refillRate int64) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow checks if a request is allowed based on token availability
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Refill tokens based on elapsed time
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)
	tokensToAdd := int64(elapsed.Seconds()) * tb.refillRate

	if tokensToAdd > 0 {
		newTokens := tb.tokens + tokensToAdd
		if newTokens > tb.capacity {
			tb.tokens = tb.capacity
		} else {
			tb.tokens = newTokens
		}
		tb.lastRefill = now
	}

	// Check if we have enough tokens
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

// RateLimiter manages rate limiting for API requests
type RateLimiter struct {
	config          *RateLimiterConfig
	globalLimiter   *TokenBucket
	endpointLimiters map[string]*TokenBucket
	clientLimiters   map[string]*TokenBucket
	clientsMux       sync.RWMutex
	batchStats      *BatchProcessingStats
	statsMux         sync.RWMutex
}

// BatchProcessingStats tracks batch processing statistics for rate limiting
type BatchProcessingStats struct {
	CurrentBatches     int
	MaxConcurrentBatches int
	TotalItemsProcessed  int64
	ItemsPerSecond      float64
	LastReset          time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config *RateLimiterConfig) *RateLimiter {
	if config == nil {
		config = DefaultRateLimiterConfig()
	}

	rl := &RateLimiter{
		config:          config,
		globalLimiter:   NewTokenBucket(int64(config.BurstSize), int64(config.GlobalRPS)),
		endpointLimiters: make(map[string]*TokenBucket),
		clientLimiters:   make(map[string]*TokenBucket),
		batchStats: &BatchProcessingStats{
			MaxConcurrentBatches: config.BatchRPS,
			LastReset:          time.Now(),
		},
	}

	// Initialize endpoint limiters
	for endpoint, rps := range config.EndpointLimits {
		rl.endpointLimiters[endpoint] = NewTokenBucket(int64(config.BurstSize), int64(rps))
	}

	// Initialize batch-specific limiter
	rl.endpointLimiters["/api/v1/items/batch/decode"] = NewTokenBucket(
		int64(config.BurstSize),
		int64(config.BatchRPS),
	)

	return rl
}

// RateLimit middleware implements rate limiting for Gin
func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check global rate limit
		if !rl.globalLimiter.Allow() {
			rl.rateLimitExceeded(c, "global rate limit exceeded")
			return
		}

		// Check endpoint-specific rate limit
		endpoint := c.Request.URL.Path
		if limiter, exists := rl.endpointLimiters[endpoint]; exists {
			if !limiter.Allow() {
				rl.rateLimitExceeded(c, fmt.Sprintf("endpoint rate limit exceeded for %s", endpoint))
				return
			}
		}

		// Check client-specific rate limit if enabled
		if rl.config.EnableClientLimiting {
			clientIP := c.ClientIP()
			if !rl.checkClientRateLimit(clientIP) {
				rl.rateLimitExceeded(c, fmt.Sprintf("client rate limit exceeded for %s", clientIP))
				return
			}
		}

		// Check batch-specific limits for batch endpoints
		if endpoint == "/api/v1/items/batch/decode" {
			if !rl.checkBatchLimits(c) {
				return
			}
		}

		c.Next()
	}
}

// checkClientRateLimit checks per-client rate limits
func (rl *RateLimiter) checkClientRateLimit(clientIP string) bool {
	rl.clientsMux.Lock()
	defer rl.clientsMux.Unlock()

	// Get or create client limiter
	limiter, exists := rl.clientLimiters[clientIP]
	if !exists {
		limiter = NewTokenBucket(int64(rl.config.BurstSize), int64(rl.config.ClientRPS))
		rl.clientLimiters[clientIP] = limiter
	}

	return limiter.Allow()
}

// checkBatchLimits checks batch-specific limits
func (rl *RateLimiter) checkBatchLimits(c *gin.Context) bool {
	rl.statsMux.Lock()
	defer rl.statsMux.Unlock()

	// Check if we're at the maximum concurrent batch limit
	if rl.batchStats.CurrentBatches >= rl.batchStats.MaxConcurrentBatches {
		rl.rateLimitExceeded(c, "maximum concurrent batch limit reached")
		return false
	}

	// For batch decode requests, also check batch size
	if c.Request.Method == "POST" && c.Request.URL.Path == "/api/v1/items/batch/decode" {
		// Note: In a real implementation, you'd parse the request body here
		// to check the actual batch size. For now, we'll rely on validation
		// in the batch processor itself.
		logger.Sugar().Debugw("Batch request started",
			"client_ip", c.ClientIP(),
			"current_batches", rl.batchStats.CurrentBatches,
			"max_batches", rl.batchStats.MaxConcurrentBatches,
		)
	}

	return true
}

// IncrementBatchCount increments the active batch count
func (rl *RateLimiter) IncrementBatchCount() {
	rl.statsMux.Lock()
	defer rl.statsMux.Unlock()

	rl.batchStats.CurrentBatches++
}

// DecrementBatchCount decrements the active batch count
func (rl *RateLimiter) DecrementBatchCount() {
	rl.statsMux.Lock()
	defer rl.statsMux.Unlock()

	if rl.batchStats.CurrentBatches > 0 {
		rl.batchStats.CurrentBatches--
	}
}

// UpdateBatchStats updates batch processing statistics
func (rl *RateLimiter) UpdateBatchStats(itemsProcessed int) {
	rl.statsMux.Lock()
	defer rl.statsMux.Unlock()

	rl.batchStats.TotalItemsProcessed += int64(itemsProcessed)

	// Calculate items per second
	elapsed := time.Since(rl.batchStats.LastReset)
	if elapsed > 0 {
		rl.batchStats.ItemsPerSecond = float64(rl.batchStats.TotalItemsProcessed) / elapsed.Seconds()
	}
}

// GetBatchStats returns current batch processing statistics
func (rl *RateLimiter) GetBatchStats() *BatchProcessingStats {
	rl.statsMux.RLock()
	defer rl.statsMux.RUnlock()

	// Return a copy to avoid concurrent access issues
	statsCopy := *rl.batchStats
	return &statsCopy
}

// rateLimitExceeded handles rate limit exceeded scenarios
func (rl *RateLimiter) rateLimitExceeded(c *gin.Context, message string) {
	logger.Sugar().Warnw("Rate limit exceeded",
		"client_ip", c.ClientIP(),
		"endpoint", c.Request.URL.Path,
		"method", c.Request.Method,
		"message", message,
	)

	c.JSON(http.StatusTooManyRequests, gin.H{
		"success": false,
		"error": gin.H{
			"code":           "RATE_LIMIT_EXCEEDED",
			"message":        "Rate limit exceeded. Please try again later.",
			"details":        message,
			"retry_after":    int(rl.config.Window.Seconds()),
			"limit_type":     "api_rate_limit",
		},
		"metadata": gin.H{
			"timestamp": time.Now().UTC(),
			"endpoint":  c.Request.URL.Path,
		},
	})
	c.Abort()
}

// CleanupExpiredClients removes expired client limiters
func (rl *RateLimiter) CleanupExpiredClients() {
	rl.clientsMux.Lock()
	defer rl.clientsMux.Unlock()

	// This is a simplified cleanup - in production you'd want more sophisticated
	// cleanup logic based on last access time
	if len(rl.clientLimiters) > 10000 { // Arbitrary threshold
		// Clear all client limiters if we have too many
		rl.clientLimiters = make(map[string]*TokenBucket)
		logger.Sugar().Infow("Cleaned up expired client rate limiters",
			"previous_count", len(rl.clientLimiters),
		)
	}
}

// ResetStats resets rate limiting statistics
func (rl *RateLimiter) ResetStats() {
	rl.statsMux.Lock()
	defer rl.statsMux.Unlock()

	rl.batchStats = &BatchProcessingStats{
		MaxConcurrentBatches: rl.config.BatchRPS,
		LastReset:          time.Now(),
	}
}

// BatchProcessingMiddleware is a specialized middleware for batch processing
func BatchProcessingMiddleware(rl *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path != "/api/v1/items/batch/decode" {
			c.Next()
			return
		}

		// Increment batch count at the start
		rl.IncrementBatchCount()

		// Ensure we decrement when the request is done
		defer rl.DecrementBatchCount()

		c.Next()
	}
}

