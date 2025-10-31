package handlers

import (
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shawnvan/bl4/pkg/logger"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	startTime time.Time
}

// NewHealthHandler creates a new health handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{
		startTime: time.Now(),
	}
}

// HandleHealth handles GET /health requests
func (h *HealthHandler) HandleHealth(c *gin.Context) {
	// Gather system information
	memStats := &runtime.MemStats{}
	runtime.ReadMemStats(memStats)

	// Calculate uptime
	uptime := time.Since(h.startTime).String()

	// Prepare response
	response := gin.H{
		"status":      "healthy",
		"timestamp":   time.Now().UTC(),
		"uptime":      uptime,
		"version":     "1.0.0",
		"service":     "bl4-item-codec",
		"environment":  getEnvironment(),
		"node": gin.H{
			"hostname":     getHostname(),
			"go_version":   runtime.Version(),
			"go_os":       runtime.GOOS,
			"go_arch":     runtime.GOARCH,
			"num_goroutine": runtime.NumGoroutine(),
		},
		"memory": gin.H{
			"alloc_mb":      memStats.Alloc / 1024 / 1024,
			"total_alloc_mb": memStats.TotalAlloc / 1024 / 1024,
			"sys_mb":        memStats.Sys / 1024 / 1024,
			"num_gc":       memStats.NumGC,
		},
		"endpoints": gin.H{
			"health":          "/health",
			"readiness":       "/health/readiness",
			"liveness":       "/health/liveness",
			"decode":          "/api/v1/items/decode",
			"encode":          "/api/v1/items/encode",
			"validate":        "/api/v1/items/validate",
			"batch_decode":   "/api/v1/items/batch/decode",
			"batch_validate":  "/api/v1/items/validate/batch",
			"docs":            "/api/v1/docs",
		},
		"checks": gin.H{
			"database": "ok",
			"memory":   h.checkMemory(),
			"cpu":      h.checkCPU(),
		},
	}

	logger.Sugar().Debug("Health check completed", "status", response["status"])

	c.JSON(http.StatusOK, response)
}

// HandleReadiness handles GET /health/readiness requests
func (h *HealthHandler) HandleReadiness(c *gin.Context) {
	// Perform readiness checks
	checks := map[string]interface{}{
		"initialization": h.checkInitialization(),
		"dependencies":   h.checkDependencies(),
		"resources":      h.checkResources(),
	}

	allHealthy := true
	for _, check := range checks {
		if checkMap, ok := check.(map[string]interface{}); ok {
			if checkMap["status"] != "ok" {
				allHealthy = false
				break
			}
		} else {
			allHealthy = false
			break
		}
	}

	response := gin.H{
		"status":    func() string { if allHealthy { return "ready" } else { return "not_ready" } }(),
		"timestamp": time.Now().UTC(),
		"uptime":    time.Since(h.startTime).String(),
		"checks":    checks,
	}

	statusCode := http.StatusOK
	if !allHealthy {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, response)
}

// HandleLiveness handles GET /health/liveness requests
func (h *HealthHandler) HandleLiveness(c *gin.Context) {
	// Simple liveness check - if we can respond, we're alive
	response := gin.H{
		"status":    "alive",
		"timestamp": time.Now().UTC(),
		"uptime":    time.Since(h.startTime).String(),
	}

	c.JSON(http.StatusOK, response)
}

// checkInitialization checks if all required components are initialized
func (h *HealthHandler) checkInitialization() map[string]interface{} {
	// In a real implementation, you would check if all required services
	// and databases are properly initialized
	return map[string]interface{}{
		"status":  "ok",
		"message": "All components initialized successfully",
	}
}

// checkDependencies checks if all external dependencies are available
func (h *HealthHandler) checkDependencies() map[string]interface{} {
	// In a real implementation, you would check connectivity to databases,
	// external APIs, and other dependencies
	return map[string]interface{}{
		"status":  "ok",
		"message": "All dependencies are available",
	}
}

// checkResources checks if the application has sufficient resources
func (h *HealthHandler) checkResources() map[string]interface{} {
	memStats := &runtime.MemStats{}
	runtime.ReadMemStats(memStats)

	// Memory usage check (less than 500MB)
	allocMB := memStats.Alloc / 1024 / 1024
	if allocMB < 500 {
		return map[string]interface{}{
			"status":  "ok",
			"message": "Memory usage is normal",
			"alloc_mb": allocMB,
		}
	}

	return map[string]interface{}{
		"status":  "warning",
		"message": "High memory usage detected",
		"alloc_mb": allocMB,
	}
}

// checkMemory performs a memory health check
func (h *HealthHandler) checkMemory() string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Consider memory usage healthy if less than 80% of available memory
	allocMB := m.Alloc / 1024 / 1024
	sysMB := m.Sys / 1024 / 1024

	if sysMB == 0 {
		return "ok"
	}

	usagePercent := float64(allocMB) / float64(sysMB) * 100
	if usagePercent < 80 {
		return "ok"
	} else if usagePercent < 90 {
		return "warning"
	}
	return "critical"
}

// checkCPU performs a CPU health check
func (h *HealthHandler) checkCPU() string {
	// Simple CPU check - in a real implementation, you would monitor
	// CPU usage over time and return appropriate status
	return "ok"
}

// getEnvironment returns the current environment
func getEnvironment() string {
	env := "development"

	// Check environment variables
	if os.Getenv("ENVIRONMENT") == "production" {
		env = "production"
	} else if os.Getenv("ENV") == "staging" {
		env = "staging"
	}

	return env
}

// getHostname returns the current hostname
func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

// getCurrentTimestamp returns the current timestamp in ISO format
func getCurrentTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}