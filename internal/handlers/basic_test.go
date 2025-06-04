package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nahuelsantos/argus/internal/services"
)

func TestNewBasicHandlers(t *testing.T) {
	loggingService := services.NewLoggingService()
	tracingService := services.NewTracingService()

	handlers := NewBasicHandlers(loggingService, tracingService)

	assert.NotNil(t, handlers)
	assert.Equal(t, loggingService, handlers.loggingService)
	assert.Equal(t, tracingService, handlers.tracingService)
}

func TestBasicHandlers_HealthHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
		expectedFields []string
	}{
		{
			name:           "successful health check",
			method:         "GET",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"status", "timestamp", "uptime", "version", "service", "purpose", "checks"},
		},
		{
			name:           "health check with POST method",
			method:         "POST",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"status", "timestamp", "uptime", "version", "service", "purpose", "checks"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			loggingService := services.NewLoggingService()
			tracingService := services.NewTracingService()
			loggingService.InitLogger()
			tracingService.InitTracer()
			handlers := NewBasicHandlers(loggingService, tracingService)

			// Create request
			req := httptest.NewRequest(tt.method, "/health", nil)
			w := httptest.NewRecorder()

			// Execute
			handlers.HealthHandler(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check all expected fields exist
			for _, field := range tt.expectedFields {
				assert.Contains(t, response, field, "Response should contain field: %s", field)
			}

			// Validate specific values
			assert.Equal(t, "healthy", response["status"])
			assert.Equal(t, "argus", response["service"])
			assert.Equal(t, "LGTM stack synthetic data generator and validator", response["purpose"])
			assert.Contains(t, response["version"], "v")

			// Check uptime format
			uptime, ok := response["uptime"].(string)
			assert.True(t, ok)
			assert.Contains(t, uptime, "h")

			// Verify checks
			checks, ok := response["checks"].(map[string]interface{})
			assert.True(t, ok)
			assert.Equal(t, "ok", checks["web_server"])
			assert.Equal(t, "ok", checks["metrics_registry"])
			assert.Equal(t, "ok", checks["logging_service"])
		})
	}
}

func TestBasicHandlers_GenerateMetricsHandler(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    map[string]string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "default metrics generation",
			queryParams:    map[string]string{},
			expectedStatus: http.StatusOK,
			expectedCount:  10,
		},
		{
			name:           "custom count metrics generation",
			queryParams:    map[string]string{"count": "25"},
			expectedStatus: http.StatusOK,
			expectedCount:  25,
		},
		{
			name:           "zero count (should use default)",
			queryParams:    map[string]string{"count": "0"},
			expectedStatus: http.StatusOK,
			expectedCount:  10,
		},
		{
			name:           "negative count (should use default)",
			queryParams:    map[string]string{"count": "-5"},
			expectedStatus: http.StatusOK,
			expectedCount:  10,
		},
		{
			name:           "invalid count format (should use default)",
			queryParams:    map[string]string{"count": "invalid"},
			expectedStatus: http.StatusOK,
			expectedCount:  10,
		},
		{
			name:           "large count",
			queryParams:    map[string]string{"count": "100"},
			expectedStatus: http.StatusOK,
			expectedCount:  100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			loggingService := services.NewLoggingService()
			tracingService := services.NewTracingService()
			loggingService.InitLogger()
			tracingService.InitTracer()
			handlers := NewBasicHandlers(loggingService, tracingService)

			// Create request with query parameters
			req := httptest.NewRequest("POST", "/generate-metrics", nil)
			q := req.URL.Query()
			for key, value := range tt.queryParams {
				q.Add(key, value)
			}
			req.URL.RawQuery = q.Encode()

			w := httptest.NewRecorder()

			// Execute
			handlers.GenerateMetricsHandler(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check response structure
			assert.Equal(t, "Metrics generated successfully", response["message"])
			assert.Equal(t, float64(tt.expectedCount), response["metrics_generated"])
			assert.Contains(t, response, "timestamp")
			assert.Contains(t, response, "types")

			// Check metric types
			types, ok := response["types"].([]interface{})
			assert.True(t, ok)
			assert.Len(t, types, 2)
			assert.Contains(t, types, "custom_business_metric")
			assert.Contains(t, types, "http_requests_total")
		})
	}
}

func TestBasicHandlers_GenerateLogsHandler(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    map[string]string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "default log generation",
			queryParams:    map[string]string{},
			expectedStatus: http.StatusOK,
			expectedCount:  5,
		},
		{
			name:           "custom count log generation",
			queryParams:    map[string]string{"count": "15"},
			expectedStatus: http.StatusOK,
			expectedCount:  15,
		},
		{
			name:           "single log generation",
			queryParams:    map[string]string{"count": "1"},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "invalid count (should use default)",
			queryParams:    map[string]string{"count": "abc"},
			expectedStatus: http.StatusOK,
			expectedCount:  5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			loggingService := services.NewLoggingService()
			tracingService := services.NewTracingService()
			loggingService.InitLogger()
			tracingService.InitTracer()
			handlers := NewBasicHandlers(loggingService, tracingService)

			// Create request
			req := httptest.NewRequest("POST", "/generate-logs", nil)
			q := req.URL.Query()
			for key, value := range tt.queryParams {
				q.Add(key, value)
			}
			req.URL.RawQuery = q.Encode()

			w := httptest.NewRecorder()

			// Execute
			handlers.GenerateLogsHandler(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check response structure
			assert.Equal(t, "Logs generated successfully", response["message"])
			assert.Equal(t, float64(tt.expectedCount), response["logs_generated"])
			assert.Contains(t, response, "timestamp")
			assert.Contains(t, response, "log_types")

			// Check log types
			logTypes, ok := response["log_types"].([]interface{})
			assert.True(t, ok)
			assert.Len(t, logTypes, 4)
			assert.Contains(t, logTypes, "info")
			assert.Contains(t, logTypes, "warn")
			assert.Contains(t, logTypes, "error")
			assert.Contains(t, logTypes, "debug")
		})
	}
}

func TestBasicHandlers_GenerateErrorHandler(t *testing.T) {
	tests := []struct {
		name               string
		expectedStatus     []int // Possible status codes
		expectedErrorTypes []string
	}{
		{
			name:               "generate random error",
			expectedStatus:     []int{400, 401, 408, 500},
			expectedErrorTypes: []string{"validation", "database", "network", "timeout", "auth"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run multiple times to test randomness
			for i := 0; i < 5; i++ {
				// Setup
				loggingService := services.NewLoggingService()
				tracingService := services.NewTracingService()
				loggingService.InitLogger()
				tracingService.InitTracer()
				handlers := NewBasicHandlers(loggingService, tracingService)

				// Create request
				req := httptest.NewRequest("POST", "/generate-error", nil)
				req.Header.Set("X-Request-ID", "test-request-id")
				w := httptest.NewRecorder()

				// Execute
				handlers.GenerateErrorHandler(w, req)

				// Assert
				assert.Contains(t, tt.expectedStatus, w.Code)
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				// Check response structure
				assert.Equal(t, true, response["error"])
				assert.Contains(t, tt.expectedErrorTypes, response["type"])
				assert.Contains(t, response["code"], "ERR_")
				assert.Contains(t, response["message"], "Simulated")
				assert.Contains(t, response, "timestamp")
				assert.Equal(t, "test-request-id", response["request_id"])

				// Verify error code format
				errorCode, ok := response["code"].(string)
				assert.True(t, ok)
				assert.True(t, strings.HasPrefix(errorCode, "ERR_"))
				assert.True(t, len(errorCode) >= 7) // ERR_X_XXX minimum
			}
		})
	}
}

func TestBasicHandlers_CPULoadHandler(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    map[string]string
		expectedStatus int
	}{
		{
			name:           "default CPU load simulation",
			queryParams:    map[string]string{},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "custom duration and intensity",
			queryParams:    map[string]string{"duration": "100ms", "intensity": "75"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid duration (should use default)",
			queryParams:    map[string]string{"duration": "invalid", "intensity": "25"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid intensity (should use default)",
			queryParams:    map[string]string{"duration": "1s", "intensity": "invalid"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "intensity out of bounds high (should use default)",
			queryParams:    map[string]string{"duration": "500ms", "intensity": "150"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "intensity out of bounds low (should use default)",
			queryParams:    map[string]string{"duration": "500ms", "intensity": "0"},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			loggingService := services.NewLoggingService()
			tracingService := services.NewTracingService()
			loggingService.InitLogger()
			tracingService.InitTracer()
			handlers := NewBasicHandlers(loggingService, tracingService)

			// Create request
			req := httptest.NewRequest("POST", "/cpu-load", nil)
			q := req.URL.Query()
			for key, value := range tt.queryParams {
				q.Add(key, value)
			}
			req.URL.RawQuery = q.Encode()

			w := httptest.NewRecorder()

			// Execute with short timeout to avoid long test runs
			start := time.Now()
			handlers.CPULoadHandler(w, req)
			duration := time.Since(start)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check response structure (actual response structure)
			assert.Equal(t, "CPU load simulation started", response["message"])
			assert.Contains(t, response, "duration")
			assert.Contains(t, response, "intensity")
			assert.Contains(t, response, "timestamp")

			// Verify execution time is reasonable (should complete quickly)
			assert.True(t, duration < 100*time.Millisecond, "CPU load test should start quickly")
		})
	}
}

func TestBasicHandlers_MemoryLoadHandler(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    map[string]string
		expectedStatus int
	}{
		{
			name:           "default memory load simulation",
			queryParams:    map[string]string{},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "custom size",
			queryParams:    map[string]string{"size": "10"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid size (should use default)",
			queryParams:    map[string]string{"size": "invalid"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "zero size (should use default)",
			queryParams:    map[string]string{"size": "0"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "negative size (should use default)",
			queryParams:    map[string]string{"size": "-10"},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			loggingService := services.NewLoggingService()
			tracingService := services.NewTracingService()
			loggingService.InitLogger()
			tracingService.InitTracer()
			handlers := NewBasicHandlers(loggingService, tracingService)

			// Create request
			req := httptest.NewRequest("POST", "/memory-load", nil)
			q := req.URL.Query()
			for key, value := range tt.queryParams {
				q.Add(key, value)
			}
			req.URL.RawQuery = q.Encode()

			w := httptest.NewRecorder()

			// Execute
			handlers.MemoryLoadHandler(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check response structure (actual response structure)
			assert.Equal(t, "Memory load simulation started", response["message"])
			assert.Contains(t, response, "allocated_mb")
			assert.Contains(t, response, "duration")
			assert.Contains(t, response, "current_alloc")
			assert.Contains(t, response, "total_alloc")
			assert.Contains(t, response, "sys")
			assert.Contains(t, response, "num_gc")
			assert.Contains(t, response, "timestamp")

			// Verify allocated memory amount (JSON numbers come as float64)
			allocatedMB, ok := response["allocated_mb"].(float64)
			assert.True(t, ok)
			if tt.queryParams["size"] == "10" {
				assert.Equal(t, 10.0, allocatedMB)
			} else if tt.queryParams["size"] == "0" || tt.queryParams["size"] == "-10" || tt.queryParams["size"] == "invalid" || tt.queryParams["size"] == "" {
				assert.Equal(t, 100.0, allocatedMB) // default
			}
		})
	}
}

func TestBasicHandlers_LGTMStatusHandler(t *testing.T) {
	tests := []struct {
		name           string
		expectedStatus int
	}{
		{
			name:           "LGTM status check",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			loggingService := services.NewLoggingService()
			tracingService := services.NewTracingService()
			loggingService.InitLogger()
			tracingService.InitTracer()
			handlers := NewBasicHandlers(loggingService, tracingService)

			// Create request
			req := httptest.NewRequest("GET", "/lgtm-status", nil)
			w := httptest.NewRecorder()

			// Execute
			handlers.LGTMStatusHandler(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check response structure (actual response is flat service map)
			assert.Contains(t, response, "loki")
			assert.Contains(t, response, "grafana")
			assert.Contains(t, response, "tempo")
			assert.Contains(t, response, "prometheus")
			assert.Contains(t, response, "alertmanager")

			// Each service should have status online/offline
			for serviceName, serviceStatus := range response {
				status, ok := serviceStatus.(string)
				assert.True(t, ok, "Service %s should have string status", serviceName)
				assert.True(t, status == "online" || status == "offline", "Service %s status should be online or offline", serviceName)
			}
		})
	}
}

func TestBasicHandlers_SettingsHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		body           string
		expectedStatus int
	}{
		{
			name:           "GET settings",
			method:         "GET",
			body:           "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST settings with valid JSON",
			method:         "POST",
			body:           `{"loki":{"url":"http://localhost:3100","username":"test","password":"pass"}}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST settings with invalid JSON",
			method:         "POST",
			body:           `{invalid json}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unsupported method",
			method:         "PUT",
			body:           "",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			loggingService := services.NewLoggingService()
			tracingService := services.NewTracingService()
			loggingService.InitLogger()
			tracingService.InitTracer()
			handlers := NewBasicHandlers(loggingService, tracingService)

			// Create request
			var body *strings.Reader
			if tt.body != "" {
				body = strings.NewReader(tt.body)
			} else {
				body = strings.NewReader("")
			}

			req := httptest.NewRequest(tt.method, "/api/settings", body)
			if tt.method == "POST" {
				req.Header.Set("Content-Type", "application/json")
			}
			w := httptest.NewRecorder()

			// Execute
			handlers.SettingsHandler(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				if tt.method == "GET" {
					// GET should return current settings
					assert.Contains(t, response, "loki")
					assert.Contains(t, response, "grafana")
					assert.Contains(t, response, "tempo")
					assert.Contains(t, response, "prometheus")
				} else if tt.method == "POST" {
					// POST should return success message
					assert.Contains(t, response, "message")
					assert.Equal(t, "Settings saved successfully", response["message"])
				}
			}
		})
	}
}

func TestBasicHandlers_TestConnectionHandler(t *testing.T) {
	tests := []struct {
		name            string
		path            string
		expectedStatus  int
		expectedService string
	}{
		{
			name:            "test loki connection",
			path:            "/api/test-connection/loki",
			expectedStatus:  http.StatusOK,
			expectedService: "loki",
		},
		{
			name:            "test grafana connection",
			path:            "/api/test-connection/grafana",
			expectedStatus:  http.StatusOK,
			expectedService: "grafana",
		},
		{
			name:            "test tempo connection",
			path:            "/api/test-connection/tempo",
			expectedStatus:  http.StatusOK,
			expectedService: "tempo",
		},
		{
			name:            "test prometheus connection",
			path:            "/api/test-connection/prometheus",
			expectedStatus:  http.StatusOK,
			expectedService: "prometheus",
		},
		{
			name:            "test unknown service",
			path:            "/api/test-connection/unknown",
			expectedStatus:  http.StatusOK,
			expectedService: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			loggingService := services.NewLoggingService()
			tracingService := services.NewTracingService()
			loggingService.InitLogger()
			tracingService.InitTracer()
			handlers := NewBasicHandlers(loggingService, tracingService)

			// Create request with JSON body (required by handler)
			reqBody := `{"url": "http://localhost:3100", "username": "", "password": ""}`
			req := httptest.NewRequest("POST", tt.path, strings.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Execute
			handlers.TestConnectionHandler(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check response structure based on actual handler response
			assert.Contains(t, response, "status")
			assert.Contains(t, response, "message")

			// Status should be either "success" or "error"
			status, ok := response["status"].(string)
			assert.True(t, ok)
			assert.True(t, status == "success" || status == "error")

			// Message should be a string
			message, ok := response["message"].(string)
			assert.True(t, ok)
			assert.NotEmpty(t, message)
		})
	}
}

// Benchmark tests for performance validation
func BenchmarkBasicHandlers_HealthHandler(b *testing.B) {
	loggingService := services.NewLoggingService()
	tracingService := services.NewTracingService()
	loggingService.InitLogger()
	tracingService.InitTracer()
	handlers := NewBasicHandlers(loggingService, tracingService)

	req := httptest.NewRequest("GET", "/health", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handlers.HealthHandler(w, req)
	}
}

func BenchmarkBasicHandlers_GenerateMetricsHandler(b *testing.B) {
	loggingService := services.NewLoggingService()
	tracingService := services.NewTracingService()
	loggingService.InitLogger()
	tracingService.InitTracer()
	handlers := NewBasicHandlers(loggingService, tracingService)

	req := httptest.NewRequest("POST", "/generate-metrics?count=10", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handlers.GenerateMetricsHandler(w, req)
	}
}

func BenchmarkBasicHandlers_GenerateLogsHandler(b *testing.B) {
	loggingService := services.NewLoggingService()
	tracingService := services.NewTracingService()
	loggingService.InitLogger()
	tracingService.InitTracer()
	handlers := NewBasicHandlers(loggingService, tracingService)

	req := httptest.NewRequest("POST", "/generate-logs?count=5", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handlers.GenerateLogsHandler(w, req)
	}
}

// Example functions for documentation
func ExampleBasicHandlers_HealthHandler() {
	// Create handlers without initializing logger to avoid output
	loggingService := services.NewLoggingService()
	tracingService := services.NewTracingService()
	// Don't initialize logger to avoid unwanted output
	handlers := NewBasicHandlers(loggingService, tracingService)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handlers.HealthHandler(w, req)

	// Output shows the handler responds with status 200
	// and includes health check information
	fmt.Println("Status:", w.Code)
	fmt.Println("Content-Type:", w.Header().Get("Content-Type"))
	// Output:
	// Status: 200
	// Content-Type: application/json
}

func ExampleBasicHandlers_GenerateMetricsHandler() {
	// Create handlers without initializing logger to avoid output
	loggingService := services.NewLoggingService()
	tracingService := services.NewTracingService()
	// Don't initialize logger to avoid unwanted output
	handlers := NewBasicHandlers(loggingService, tracingService)

	req := httptest.NewRequest("POST", "/generate-metrics?count=5", nil)
	w := httptest.NewRecorder()

	handlers.GenerateMetricsHandler(w, req)

	fmt.Println("Status:", w.Code)
	fmt.Println("Generated metrics for LGTM stack testing")
	// Output:
	// Status: 200
	// Generated metrics for LGTM stack testing
}
