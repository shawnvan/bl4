package contract

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shawnvan/bl4/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDecodeAPIContract tests the API contract for the decode endpoint
func TestDecodeAPIContract(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create test server
	serverConfig := api.DefaultConfig()
	server := api.NewServer(serverConfig)
	engine := server.GetEngine()

	t.Run("Valid decode request", func(t *testing.T) {
		// Test with valid Base85 item code
		requestBody := map[string]interface{}{
			"serial_code": "@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}",
			"options": map[string]interface{}{
				"include_bitstream": false,
				"format":           "structured",
			},
		}

		jsonBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", "/api/v1/items/decode", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		// Currently should return not implemented
		assert.Equal(t, http.StatusNotImplemented, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Should have error structure
		assert.Contains(t, response, "error")
		errorObj := response["error"].(map[string]interface{})
		assert.Equal(t, "NOT_IMPLEMENTED", errorObj["code"])
		assert.Equal(t, "Decode endpoint not yet implemented", errorObj["message"])
	})

	t.Run("Invalid JSON request", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/api/v1/items/decode", bytes.NewBuffer([]byte("invalid json")))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "error")
	})

	t.Run("Missing serial_code field", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"options": map[string]interface{}{
				"include_bitstream": false,
			},
		}

		jsonBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", "/api/v1/items/decode", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Empty serial_code", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"serial_code": "",
			"options": map[string]interface{}{
				"include_bitstream": false,
			},
		}

		jsonBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", "/api/v1/items/decode", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Invalid serial_code format", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"serial_code": "invalid_code_without_prefix",
			"options": map[string]interface{}{
				"include_bitstream": false,
			},
		}

		jsonBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", "/api/v1/items/decode", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Invalid HTTP method", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/v1/items/decode", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("Content-Type not JSON", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/api/v1/items/decode", bytes.NewBuffer([]byte("some data")))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "text/plain")

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Large payload test", func(t *testing.T) {
		// Create a very large serial code
		largeCode := "@U" + string(make([]byte, 10000)) // 10KB payload
		requestBody := map[string]interface{}{
			"serial_code": largeCode,
			"options": map[string]interface{}{
				"include_bitstream": false,
			},
		}

		jsonBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", "/api/v1/items/decode", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		// Should handle large payload gracefully (either accept or reject with proper error)
		assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusRequestEntityTooLarge || w.Code == http.StatusNotImplemented)
	})
}

// TestDecodeAPIResponseFormat tests the expected response format
func TestDecodeAPIResponseFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serverConfig := api.DefaultConfig()
	server := api.NewServer(serverConfig)
	engine := server.GetEngine()

	t.Run("Success response format", func(t *testing.T) {
		// This test will be updated when the endpoint is implemented
		// For now, it documents the expected response format

		// Expected success response format:
		/*
		{
			"success": true,
			"data": {
				"serial_code": "@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}",
				"decoded_data": {
					"level": 24,
					"type": "pistol",
					"manufacturer": "maliwan",
					"parts": [
						{"index": 0, "value": 1},
						{"index": 2, "value": 3379}
					],
					"raw_parts": "24, 0, 1, 50| 2, 3379|| {76} {2} {3}"
				},
				"metadata": {
					"bitstream_size": 128,
					"processing_time_ms": 0.15,
					"version": "1.0.0"
				}
			}
		}
		*/

		// When implemented, this test should verify the response format
		t.Skip("Decode endpoint not yet implemented - response format test pending")
	})

	t.Run("Error response format", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"serial_code": "invalid",
		}

		jsonBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", "/api/v1/items/decode", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Should have proper error structure
		assert.Contains(t, response, "error")
		errorObj := response["error"].(map[string]interface{})
		assert.Contains(t, errorObj, "code")
		assert.Contains(t, errorObj, "message")
		assert.Contains(t, errorObj, "details")
	})
}

// TestDecodeAPIPerformanceContract tests performance requirements
func TestDecodeAPIPerformanceContract(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serverConfig := api.DefaultConfig()
	server := api.NewServer(serverConfig)
	engine := server.GetEngine()

	t.Run("Performance requirements", func(t *testing.T) {
		// Performance targets from spec:
		// - Single decode: < 1ms max, 200μs average
		// - This test will be meaningful when the endpoint is implemented

		requestBody := map[string]interface{}{
			"serial_code": "@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}",
		}

		jsonBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", "/api/v1/items/decode", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()

		// Measure response time
		start := time.Now()
		engine.ServeHTTP(w, req)
		duration := time.Since(start)

		// For now, just verify it returns quickly (even if not implemented)
		assert.Less(t, duration, 100*time.Millisecond, "Response should be under 100ms even when not implemented")

		// When implemented, this should be:
		// assert.Less(t, duration, time.Millisecond, "Single decode should be under 1ms")
		// assert.Less(t, duration, 200*time.Microsecond, "Single decode should average around 200μs")

		t.Skip("Performance test pending implementation - currently returns NOT_IMPLEMENTED")
	})
}