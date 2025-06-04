package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nahuelsantos/argus/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSimulationHandlers(t *testing.T) {
	loggingService := services.NewLoggingService()
	tracingService := services.NewTracingService()
	loggingService.InitTestLogger()
	tracingService.InitTracer()

	handlers := NewSimulationHandlers(loggingService, tracingService)

	assert.NotNil(t, handlers)
	assert.NotNil(t, handlers.loggingService)
	assert.NotNil(t, handlers.tracingService)
}

func TestSimulationHandlers_SimulateWebServiceHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "successful web service simulation",
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
			loggingService.InitTestLogger()
			tracingService.InitTracer()
			handlers := NewSimulationHandlers(loggingService, tracingService)

			// Create request
			req := httptest.NewRequest(tt.method, "/simulate-web-service", nil)
			w := httptest.NewRecorder()

			// Execute
			handlers.SimulateWebServiceHandler(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check response structure
			assert.Contains(t, response, "message")
			assert.Contains(t, response, "service_type")
			assert.Contains(t, response, "requests_simulated")
			assert.Contains(t, response, "avg_response_time_ms")
			assert.Contains(t, response, "error_rate")
			assert.Contains(t, response, "endpoints_tested")
			assert.Contains(t, response, "timestamp")

			// Verify content
			assert.Equal(t, "Web service simulation completed", response["message"])
			assert.Equal(t, "web-service", response["service_type"])

			// Check numeric values are reasonable
			requestsSimulated, ok := response["requests_simulated"].(float64)
			assert.True(t, ok)
			assert.GreaterOrEqual(t, requestsSimulated, 10.0)
			assert.LessOrEqual(t, requestsSimulated, 60.0)

			avgResponseTime, ok := response["avg_response_time_ms"].(float64)
			assert.True(t, ok)
			assert.GreaterOrEqual(t, avgResponseTime, 50.0)
			assert.LessOrEqual(t, avgResponseTime, 250.0)

			// Check endpoints tested
			endpointsTested, ok := response["endpoints_tested"].([]interface{})
			assert.True(t, ok)
			assert.NotEmpty(t, endpointsTested)
			assert.Contains(t, endpointsTested, "/")
			assert.Contains(t, endpointsTested, "/about")
		})
	}
}

func TestSimulationHandlers_SimulateAPIServiceHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "successful API service simulation",
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
			loggingService.InitTestLogger()
			tracingService.InitTracer()
			handlers := NewSimulationHandlers(loggingService, tracingService)

			// Create request
			req := httptest.NewRequest(tt.method, "/simulate-api-service", nil)
			w := httptest.NewRecorder()

			// Execute
			handlers.SimulateAPIServiceHandler(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check response structure
			assert.Contains(t, response, "message")
			assert.Contains(t, response, "service_type")
			assert.Contains(t, response, "requests_simulated")
			assert.Contains(t, response, "avg_latency_ms")
			assert.Contains(t, response, "rate_limit_hits")
			assert.Contains(t, response, "auth_failures")
			assert.Contains(t, response, "endpoints_available")
			assert.Contains(t, response, "timestamp")

			// Verify content
			assert.Equal(t, "API service simulation completed", response["message"])
			assert.Equal(t, "api-service", response["service_type"])

			// Check numeric values are reasonable
			requestsSimulated, ok := response["requests_simulated"].(float64)
			assert.True(t, ok)
			assert.GreaterOrEqual(t, requestsSimulated, 20.0)
			assert.LessOrEqual(t, requestsSimulated, 120.0)

			avgLatency, ok := response["avg_latency_ms"].(float64)
			assert.True(t, ok)
			assert.GreaterOrEqual(t, avgLatency, 25.0)
			assert.LessOrEqual(t, avgLatency, 125.0)

			endpointsAvailable, ok := response["endpoints_available"].(float64)
			assert.True(t, ok)
			assert.Equal(t, 8.0, endpointsAvailable) // Should have 8 API endpoints
		})
	}
}

func TestSimulationHandlers_SimulateDatabaseServiceHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "successful database service simulation",
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
			loggingService.InitTestLogger()
			tracingService.InitTracer()
			handlers := NewSimulationHandlers(loggingService, tracingService)

			// Create request
			req := httptest.NewRequest(tt.method, "/simulate-database-service", nil)
			w := httptest.NewRecorder()

			// Execute
			handlers.SimulateDatabaseServiceHandler(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check response structure (actual fields from database simulation)
			assert.Contains(t, response, "message")
			assert.Contains(t, response, "service_type")
			assert.Contains(t, response, "queries_executed")
			assert.Contains(t, response, "avg_query_time_ms")
			assert.Contains(t, response, "slow_queries")
			assert.Contains(t, response, "connection_pool_size")
			assert.Contains(t, response, "tables_accessed")
			assert.Contains(t, response, "timestamp")

			// Verify content
			assert.Equal(t, "Database service simulation completed", response["message"])
			assert.Equal(t, "database-service", response["service_type"])

			// Check numeric values are reasonable
			queriesExecuted, ok := response["queries_executed"].(float64)
			assert.True(t, ok)
			assert.GreaterOrEqual(t, queriesExecuted, 20.0)
			assert.LessOrEqual(t, queriesExecuted, 100.0)

			avgQueryTime, ok := response["avg_query_time_ms"].(float64)
			assert.True(t, ok)
			assert.GreaterOrEqual(t, avgQueryTime, 10.0)
			assert.LessOrEqual(t, avgQueryTime, 60.0)

			// Check additional fields
			connectionPoolSize, ok := response["connection_pool_size"].(float64)
			assert.True(t, ok)
			assert.GreaterOrEqual(t, connectionPoolSize, 1.0)

			slowQueries, ok := response["slow_queries"].(float64)
			assert.True(t, ok)
			assert.GreaterOrEqual(t, slowQueries, 0.0)

			// Check tables accessed array
			tablesAccessed, ok := response["tables_accessed"].([]interface{})
			assert.True(t, ok)
			assert.NotEmpty(t, tablesAccessed)
		})
	}
}

func TestSimulationHandlers_SimulateStaticSiteHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "successful static site simulation",
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
			loggingService.InitTestLogger()
			tracingService.InitTracer()
			handlers := NewSimulationHandlers(loggingService, tracingService)

			// Create request
			req := httptest.NewRequest(tt.method, "/simulate-static-site", nil)
			w := httptest.NewRecorder()

			// Execute
			handlers.SimulateStaticSiteHandler(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check response structure (actual fields from static site simulation)
			assert.Contains(t, response, "message")
			assert.Contains(t, response, "service_type")
			assert.Contains(t, response, "requests_served")
			assert.Contains(t, response, "cache_hit_rate")
			assert.Contains(t, response, "total_bandwidth_mb")
			assert.Contains(t, response, "timestamp")

			// Verify content
			assert.Equal(t, "Static site simulation completed", response["message"])
			assert.Equal(t, "static-site", response["service_type"])

			// Check numeric values are reasonable
			requestsServed, ok := response["requests_served"].(float64)
			assert.True(t, ok)
			assert.GreaterOrEqual(t, requestsServed, 50.0)

			// Cache hit rate should be a percentage string
			cacheHitRate, ok := response["cache_hit_rate"].(string)
			assert.True(t, ok)
			assert.Contains(t, cacheHitRate, "%")

			// Total bandwidth should be a string with MB
			totalBandwidth, ok := response["total_bandwidth_mb"].(string)
			assert.True(t, ok)
			assert.NotEmpty(t, totalBandwidth)
		})
	}
}

func TestSimulationHandlers_SimulateMicroserviceHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "successful microservice simulation",
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
			loggingService.InitTestLogger()
			tracingService.InitTracer()
			handlers := NewSimulationHandlers(loggingService, tracingService)

			// Create request
			req := httptest.NewRequest(tt.method, "/simulate-microservice", nil)
			w := httptest.NewRecorder()

			// Execute
			handlers.SimulateMicroserviceHandler(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check response structure (actual fields from microservice simulation)
			assert.Contains(t, response, "message")
			assert.Contains(t, response, "service_type")
			assert.Contains(t, response, "service_calls")
			assert.Contains(t, response, "circuit_breaker_trips")
			assert.Contains(t, response, "services_involved")
			assert.Contains(t, response, "timestamp")

			// Verify content
			assert.Equal(t, "Microservice simulation completed", response["message"])
			assert.Equal(t, "microservice", response["service_type"])

			// Check numeric values are reasonable
			serviceCalls, ok := response["service_calls"].(float64)
			assert.True(t, ok)
			assert.GreaterOrEqual(t, serviceCalls, 10.0) // Lower threshold since it's random

			circuitBreakerTrips, ok := response["circuit_breaker_trips"].(float64)
			assert.True(t, ok)
			assert.GreaterOrEqual(t, circuitBreakerTrips, 0.0)

			// Check services involved
			servicesInvolved, ok := response["services_involved"].([]interface{})
			assert.True(t, ok)
			assert.NotEmpty(t, servicesInvolved)
		})
	}
}

// Benchmark tests for performance validation
func BenchmarkSimulationHandlers_SimulateWebServiceHandler(b *testing.B) {
	loggingService := services.NewLoggingService()
	tracingService := services.NewTracingService()
	loggingService.InitTestLogger()
	tracingService.InitTracer()
	handlers := NewSimulationHandlers(loggingService, tracingService)

	req := httptest.NewRequest("POST", "/simulate-web-service", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handlers.SimulateWebServiceHandler(w, req)
	}
}

func BenchmarkSimulationHandlers_SimulateAPIServiceHandler(b *testing.B) {
	loggingService := services.NewLoggingService()
	tracingService := services.NewTracingService()
	loggingService.InitTestLogger()
	tracingService.InitTracer()
	handlers := NewSimulationHandlers(loggingService, tracingService)

	req := httptest.NewRequest("POST", "/simulate-api-service", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handlers.SimulateAPIServiceHandler(w, req)
	}
}

func BenchmarkSimulationHandlers_SimulateDatabaseServiceHandler(b *testing.B) {
	loggingService := services.NewLoggingService()
	tracingService := services.NewTracingService()
	loggingService.InitTestLogger()
	tracingService.InitTracer()
	handlers := NewSimulationHandlers(loggingService, tracingService)

	req := httptest.NewRequest("POST", "/simulate-database-service", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handlers.SimulateDatabaseServiceHandler(w, req)
	}
}

func BenchmarkSimulationHandlers_SimulateStaticSiteHandler(b *testing.B) {
	loggingService := services.NewLoggingService()
	tracingService := services.NewTracingService()
	loggingService.InitTestLogger()
	tracingService.InitTracer()
	handlers := NewSimulationHandlers(loggingService, tracingService)

	req := httptest.NewRequest("POST", "/simulate-static-site", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handlers.SimulateStaticSiteHandler(w, req)
	}
}

func BenchmarkSimulationHandlers_SimulateMicroserviceHandler(b *testing.B) {
	loggingService := services.NewLoggingService()
	tracingService := services.NewTracingService()
	loggingService.InitTestLogger()
	tracingService.InitTracer()
	handlers := NewSimulationHandlers(loggingService, tracingService)

	req := httptest.NewRequest("POST", "/simulate-microservice", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handlers.SimulateMicroserviceHandler(w, req)
	}
}
