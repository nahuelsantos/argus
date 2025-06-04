package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nahuelsantos/argus/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTestingHandlers(t *testing.T) {
	loggingService := services.NewLoggingService()
	tracingService := services.NewTracingService()
	loggingService.InitLogger()
	tracingService.InitTracer()

	handlers := NewTestingHandlers(loggingService, tracingService)

	assert.NotNil(t, handlers)
	assert.NotNil(t, handlers.loggingService)
	assert.NotNil(t, handlers.tracingService)
}

func TestTestingHandlers_GenerateJSONLogsHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "successful JSON logs generation",
			method:         "POST",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "GET method works too",
			method:         "GET",
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
			handlers := NewTestingHandlers(loggingService, tracingService)

			// Create request
			req := httptest.NewRequest(tt.method, "/generate-json-logs", nil)
			w := httptest.NewRecorder()

			// Execute
			handlers.GenerateJSONLogsHandler(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check response structure
			assert.Contains(t, response, "message")
			assert.Contains(t, response, "logs_generated")
			assert.Contains(t, response, "log_formats")
			assert.Contains(t, response, "sample_logs")
			assert.Contains(t, response, "test_purpose")
			assert.Contains(t, response, "timestamp")
			assert.Contains(t, response, "service")
			assert.Contains(t, response, "functionality")

			// Verify content
			assert.Equal(t, "JSON logs generated for Loki testing", response["message"])
			assert.Equal(t, 10.0, response["logs_generated"])
			assert.Equal(t, 3.0, response["log_formats"])
			assert.Equal(t, "argus", response["service"])
			assert.Equal(t, "loki_json_validation", response["functionality"])

			// Check sample logs are arrays
			sampleLogs, ok := response["sample_logs"].([]interface{})
			assert.True(t, ok)
			assert.Len(t, sampleLogs, 3)

			// Each sample log should be a valid JSON string
			for _, log := range sampleLogs {
				logStr, ok := log.(string)
				assert.True(t, ok)

				var logJSON map[string]interface{}
				err := json.Unmarshal([]byte(logStr), &logJSON)
				assert.NoError(t, err)
				assert.Contains(t, logJSON, "timestamp")
				assert.Contains(t, logJSON, "level")
				assert.Contains(t, logJSON, "service")
				assert.Contains(t, logJSON, "message")
			}
		})
	}
}

func TestTestingHandlers_GenerateUnstructuredLogsHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "successful unstructured logs generation",
			method:         "POST",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "GET method works too",
			method:         "GET",
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
			handlers := NewTestingHandlers(loggingService, tracingService)

			// Create request
			req := httptest.NewRequest(tt.method, "/generate-unstructured-logs", nil)
			w := httptest.NewRecorder()

			// Execute
			handlers.GenerateUnstructuredLogsHandler(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check response structure
			assert.Contains(t, response, "message")
			assert.Contains(t, response, "logs_generated")
			assert.Contains(t, response, "log_templates")
			assert.Contains(t, response, "sample_logs")
			assert.Contains(t, response, "test_purpose")
			assert.Contains(t, response, "functionality")

			// Verify content
			assert.Equal(t, "Unstructured logs generated for Loki testing", response["message"])
			assert.Equal(t, 10.0, response["logs_generated"])
			assert.Equal(t, 7.0, response["log_templates"])
			assert.Equal(t, "loki_unstructured_validation", response["functionality"])

			// Check sample logs
			sampleLogs, ok := response["sample_logs"].([]interface{})
			assert.True(t, ok)
			assert.Len(t, sampleLogs, 3)

			// Each sample log should be a string containing expected patterns
			for _, log := range sampleLogs {
				logStr, ok := log.(string)
				assert.True(t, ok)
				assert.NotEmpty(t, logStr)
				// Should contain timestamp pattern
				assert.Contains(t, logStr, "[2")
				// Should contain log level
				assert.True(t, strings.Contains(logStr, "INFO") ||
					strings.Contains(logStr, "ERROR") ||
					strings.Contains(logStr, "WARN"))
			}
		})
	}
}

func TestTestingHandlers_GenerateMixedLogsHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "successful mixed logs generation",
			method:         "POST",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "GET method works too",
			method:         "GET",
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
			handlers := NewTestingHandlers(loggingService, tracingService)

			// Create request
			req := httptest.NewRequest(tt.method, "/generate-mixed-logs", nil)
			w := httptest.NewRecorder()

			// Execute
			handlers.GenerateMixedLogsHandler(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check response structure
			assert.Contains(t, response, "message")
			assert.Contains(t, response, "logs_generated")
			assert.Contains(t, response, "formats")
			assert.Contains(t, response, "sample_logs")
			assert.Contains(t, response, "functionality")

			// Verify content
			assert.Equal(t, "Mixed format logs generated for Loki testing", response["message"])
			assert.Equal(t, 15.0, response["logs_generated"])
			assert.Equal(t, "loki_mixed_validation", response["functionality"])

			// Check formats array
			formats, ok := response["formats"].([]interface{})
			assert.True(t, ok)
			assert.Len(t, formats, 3)
			assert.Contains(t, formats, "JSON")
			assert.Contains(t, formats, "Key-Value")
			assert.Contains(t, formats, "Plain Text")

			// Check sample logs
			sampleLogs, ok := response["sample_logs"].([]interface{})
			assert.True(t, ok)
			assert.Len(t, sampleLogs, 3)

			for _, log := range sampleLogs {
				logStr, ok := log.(string)
				assert.True(t, ok)
				assert.NotEmpty(t, logStr)
			}
		})
	}
}

func TestTestingHandlers_GenerateMultilineLogsHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "successful multiline logs generation",
			method:         "POST",
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
			handlers := NewTestingHandlers(loggingService, tracingService)

			// Create request
			req := httptest.NewRequest(tt.method, "/generate-multiline-logs", nil)
			w := httptest.NewRecorder()

			// Execute
			handlers.GenerateMultilineLogsHandler(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check response structure (multiline logs uses different field names)
			assert.Contains(t, response, "message")
			assert.Contains(t, response, "stack_traces")
			assert.Contains(t, response, "functionality")
		})
	}
}

func TestTestingHandlers_SimulateWordPressServiceHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "successful WordPress simulation",
			method:         "POST",
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
			handlers := NewTestingHandlers(loggingService, tracingService)

			// Create request
			req := httptest.NewRequest(tt.method, "/simulate-wordpress", nil)
			w := httptest.NewRecorder()

			// Execute
			handlers.SimulateWordPressServiceHandler(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check response structure (WordPress simulation uses different field names)
			assert.Contains(t, response, "message")
			assert.Contains(t, response, "service")
			assert.Contains(t, response, "service_type")
			assert.Equal(t, "argus", response["service"])
			assert.Equal(t, "wordpress", response["service_type"])
		})
	}
}

func TestTestingHandlers_SimulateNextJSServiceHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "successful NextJS simulation",
			method:         "POST",
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
			handlers := NewTestingHandlers(loggingService, tracingService)

			// Create request
			req := httptest.NewRequest(tt.method, "/simulate-nextjs", nil)
			w := httptest.NewRecorder()

			// Execute
			handlers.SimulateNextJSServiceHandler(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check response structure (NextJS simulation uses different field names)
			assert.Contains(t, response, "message")
			assert.Contains(t, response, "service")
			assert.Contains(t, response, "service_type")
			assert.Equal(t, "argus", response["service"])
			assert.Equal(t, "nextjs", response["service_type"])
		})
	}
}

func TestTestingHandlers_SimulateCrossServiceTracingHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "successful cross-service tracing simulation",
			method:         "POST",
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
			handlers := NewTestingHandlers(loggingService, tracingService)

			// Create request
			req := httptest.NewRequest(tt.method, "/simulate-cross-service-tracing", nil)
			w := httptest.NewRecorder()

			// Execute
			handlers.SimulateCrossServiceTracingHandler(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check response structure (cross-service tracing uses different field names)
			assert.Contains(t, response, "message")
			assert.Contains(t, response, "generated_traces")
			assert.Contains(t, response, "trace_scenarios")
		})
	}
}

func TestTestingHandlers_TestServiceDiscoveryHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "successful service discovery test",
			method:         "POST",
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
			handlers := NewTestingHandlers(loggingService, tracingService)

			// Create request
			req := httptest.NewRequest(tt.method, "/test-service-discovery", nil)
			w := httptest.NewRecorder()

			// Execute
			handlers.TestServiceDiscoveryHandler(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check response structure (service discovery uses different field names)
			assert.Contains(t, response, "message")
			assert.Contains(t, response, "functionality")
			assert.Contains(t, response, "services_tested")
		})
	}
}

func TestTestingHandlers_TestReverseProxyHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "successful reverse proxy test",
			method:         "POST",
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
			handlers := NewTestingHandlers(loggingService, tracingService)

			// Create request
			req := httptest.NewRequest(tt.method, "/test-reverse-proxy", nil)
			w := httptest.NewRecorder()

			// Execute
			handlers.TestReverseProxyHandler(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check response structure (reverse proxy uses different field names)
			assert.Contains(t, response, "message")
			assert.Contains(t, response, "functionality")
			assert.Contains(t, response, "routes_tested")
		})
	}
}

func TestTestingHandlers_TestSSLMonitoringHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "successful SSL monitoring test",
			method:         "POST",
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
			handlers := NewTestingHandlers(loggingService, tracingService)

			// Create request
			req := httptest.NewRequest(tt.method, "/test-ssl-monitoring", nil)
			w := httptest.NewRecorder()

			// Execute
			handlers.TestSSLMonitoringHandler(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check response structure (SSL monitoring uses different field names)
			assert.Contains(t, response, "message")
			assert.Contains(t, response, "functionality")
			assert.Contains(t, response, "certificates_checked")
		})
	}
}

func TestTestingHandlers_TestDomainHealthHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "successful domain health test",
			method:         "POST",
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
			handlers := NewTestingHandlers(loggingService, tracingService)

			// Create request
			req := httptest.NewRequest(tt.method, "/test-domain-health", nil)
			w := httptest.NewRecorder()

			// Execute
			handlers.TestDomainHealthHandler(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check response structure (domain health uses different field names)
			assert.Contains(t, response, "message")
			assert.Contains(t, response, "functionality")
			assert.Contains(t, response, "domains_checked")
		})
	}
}

// Benchmark tests for performance validation
func BenchmarkTestingHandlers_GenerateJSONLogsHandler(b *testing.B) {
	loggingService := services.NewLoggingService()
	tracingService := services.NewTracingService()
	loggingService.InitLogger()
	tracingService.InitTracer()
	handlers := NewTestingHandlers(loggingService, tracingService)

	req := httptest.NewRequest("POST", "/generate-json-logs", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handlers.GenerateJSONLogsHandler(w, req)
	}
}

func BenchmarkTestingHandlers_GenerateUnstructuredLogsHandler(b *testing.B) {
	loggingService := services.NewLoggingService()
	tracingService := services.NewTracingService()
	loggingService.InitLogger()
	tracingService.InitTracer()
	handlers := NewTestingHandlers(loggingService, tracingService)

	req := httptest.NewRequest("POST", "/generate-unstructured-logs", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handlers.GenerateUnstructuredLogsHandler(w, req)
	}
}

func BenchmarkTestingHandlers_GenerateMixedLogsHandler(b *testing.B) {
	loggingService := services.NewLoggingService()
	tracingService := services.NewTracingService()
	loggingService.InitLogger()
	tracingService.InitTracer()
	handlers := NewTestingHandlers(loggingService, tracingService)

	req := httptest.NewRequest("POST", "/generate-mixed-logs", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handlers.GenerateMixedLogsHandler(w, req)
	}
}
