package middleware

import (
	"sync"
	"time"
)

// SafeTokenBucket implements a thread-safe token bucket with precise timing
type SafeTokenBucket struct {
	capacity      int64
	tokens        int64
	refillRate    int64
	lastRefill    int64 // Use nanoseconds for precision
	mu            sync.RWMutex
}

// NewSafeTokenBucket creates a new safe token bucket
func NewSafeTokenBucket(capacity, refillRate int64) *SafeTokenBucket {
	return &SafeTokenBucket{
		capacity:   capacity,
		tokens:     capacity,
		refillRate: refillRate,
		lastRefill: time.Now().UnixNano(),
	}
}

// Allow checks if a request is allowed based on token availability
func (tb *SafeTokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now().UnixNano()
	
	// Calculate elapsed time in nanoseconds
	elapsedNanos := now - tb.lastRefill
	if elapsedNanos < 0 {
		elapsedNanos = 0
	}

	// Calculate tokens to add with high precision
	tokensToAdd := elapsedNanos * tb.refillRate / 1_000_000_000 // Convert to seconds

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

// GetTokenCount returns current token count (for monitoring)
func (tb *SafeTokenBucket) GetTokenCount() int64 {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	return tb.tokens
}

// Reset resets the token bucket to initial state
func (tb *SafeTokenBucket) Reset() {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.tokens = tb.capacity
	tb.lastRefill = time.Now().UnixNano()
}

// SafeRateLimiter implements enhanced rate limiting with proper synchronization
type SafeRateLimiter struct {
	config          *RateLimiterConfig
	globalLimiter   *SafeTokenBucket
	endpointLimiters map[string]*SafeTokenBucket
	clientLimiters   map[string]*SafeTokenBucket
	clientsMux       sync.RWMutex
	batchStats      *BatchProcessingStats
	statsMux         sync.RWMutex
}

// NewSafeRateLimiter creates a new safe rate limiter
func NewSafeRateLimiter(config *RateLimiterConfig) *SafeRateLimiter {
	if config == nil {
		config = DefaultRateLimiterConfig()
	}

	srl := &SafeRateLimiter{
		config:          config,
		globalLimiter:   NewSafeTokenBucket(int64(config.BurstSize), int64(config.GlobalRPS)),
		endpointLimiters: make(map[string]*SafeTokenBucket),
		clientLimiters:   make(map[string]*SafeTokenBucket),
		batchStats: &BatchProcessingStats{
			MaxConcurrentBatches: config.BatchRPS,
			LastReset:          time.Now(),
		},
	}

	// Initialize endpoint limiters
	for endpoint, rps := range config.EndpointLimits {
		srl.endpointLimiters[endpoint] = NewSafeTokenBucket(int64(config.BurstSize), int64(rps))
	}

	// Initialize batch-specific limiter
	srl.endpointLimiters["/api/v1/items/batch/decode"] = NewSafeTokenBucket(
		int64(config.BurstSize),
		int64(config.BatchRPS),
	)

	return srl
}

// checkClientRateLimitSafely checks per-client rate limits with proper synchronization
func (srl *SafeRateLimiter) checkClientRateLimitSafely(clientIP string) bool {
	srl.clientsMux.Lock()
	defer srl.clientsMux.Unlock()

	// Get or create client limiter
	limiter, exists := srl.clientLimiters[clientIP]
	if !exists {
		limiter = NewSafeTokenBucket(int64(srl.config.BurstSize), int64(srl.config.ClientRPS))
		srl.clientLimiters[clientIP] = limiter
	}

	return limiter.Allow()
}

// IncrementBatchCount safely increments the active batch count
func (srl *SafeRateLimiter) IncrementBatchCount() {
	srl.statsMux.Lock()
	defer srl.statsMux.Unlock()

	srl.batchStats.CurrentBatches++
}

// DecrementBatchCount safely decrements the active batch count
func (srl *SafeRateLimiter) DecrementBatchCount() {
	srl.statsMux.Lock()
	defer srl.statsMux.Unlock()

	if srl.batchStats.CurrentBatches > 0 {
		srl.batchStats.CurrentBatches--
	}
}

// GetBatchStats safely returns current batch processing statistics
func (srl *SafeRateLimiter) GetBatchStats() *BatchProcessingStats {
	srl.statsMux.RLock()
	defer srl.statsMux.RUnlock()

	// Return a copy to avoid concurrent access issues
	statsCopy := *srl.batchStats
	return &statsCopy
}

// CleanupExpiredClients safely removes expired client limiters
func (srl *SafeRateLimiter) CleanupExpiredClients() {
	srl.clientsMux.Lock()
	defer srl.clientsMux.Unlock()

	// This is a simplified cleanup - in production you'd want more sophisticated
	// cleanup logic based on last access time
	if len(srl.clientLimiters) > 10000 { // Arbitrary threshold
		// Clear all client limiters if we have too many
		srl.clientLimiters = make(map[string]*SafeTokenBucket)
	}
}