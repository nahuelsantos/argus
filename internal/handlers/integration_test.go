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

func TestNewIntegrationHandlers(t *testing.T) {
	loggingService := services.NewLoggingService()
	tracingService := services.NewTracingService()
	loggingService.InitLogger()
	tracingService.InitTracer()

	handlers := NewIntegrationHandlers(loggingService, tracingService)

	assert.NotNil(t, handlers)
	assert.NotNil(t, handlers.loggingService)
	assert.NotNil(t, handlers.tracingService)
}

func TestIntegrationHandlers_TestLGTMIntegration(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "successful LGTM integration test",
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
			handlers := NewIntegrationHandlers(loggingService, tracingService)

			// Create request
			req := httptest.NewRequest(tt.method, "/test-lgtm-integration", nil)
			w := httptest.NewRecorder()

			// Execute
			handlers.TestLGTMIntegration(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response LGTMIntegrationSummary
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check response structure
			assert.Contains(t, []string{"healthy", "degraded", "critical"}, response.OverallStatus)
			assert.GreaterOrEqual(t, response.TotalCount, 5) // Should test at least 5 components
			assert.GreaterOrEqual(t, response.HealthyCount, 0)
			assert.LessOrEqual(t, response.HealthyCount, response.TotalCount)
			assert.NotEmpty(t, response.Components)
			assert.NotZero(t, response.Timestamp)

			// Check that all expected components are tested
			componentNames := make(map[string]bool)
			for _, comp := range response.Components {
				componentNames[comp.Component] = true

				// Each component should have required fields
				assert.NotEmpty(t, comp.Component)
				assert.Contains(t, []string{"healthy", "degraded", "failed"}, comp.Status)
				assert.NotEmpty(t, comp.Message)
				assert.NotZero(t, comp.Timestamp)
				assert.GreaterOrEqual(t, comp.ResponseTime, int64(0))
			}

			// Verify expected components are present
			expectedComponents := []string{
				"grafana_datasources",
				"prometheus_targets",
				"loki_ingestion",
				"tempo_tracing",
				"otel_collector",
			}

			for _, expected := range expectedComponents {
				assert.True(t, componentNames[expected], "Expected component %s not found", expected)
			}
		})
	}
}

func TestIntegrationHandlers_TestGrafanaDashboards(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "successful Grafana dashboards test",
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
			handlers := NewIntegrationHandlers(loggingService, tracingService)

			// Create request
			req := httptest.NewRequest(tt.method, "/test-grafana-dashboards", nil)
			w := httptest.NewRecorder()

			// Execute
			handlers.TestGrafanaDashboards(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check response structure (error case when dashboard config file missing)
			assert.Contains(t, response, "message")
			assert.Contains(t, response, "status")
			assert.Contains(t, response, "timestamp")

			// When dashboard config is missing, it returns an error response
			if response["status"] == "error" {
				assert.Contains(t, response, "error")
				assert.Equal(t, "Cannot load dashboard configuration", response["message"])
			} else {
				// Success case would have these fields
				assert.Contains(t, response, "dashboards_tested")
				assert.Contains(t, response, "functionality")
				assert.Contains(t, response, "test_results")
			}
		})
	}
}

func TestIntegrationHandlers_TestAlertRules(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "successful alert rules test",
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
			handlers := NewIntegrationHandlers(loggingService, tracingService)

			// Create request
			req := httptest.NewRequest(tt.method, "/test-alert-rules", nil)
			w := httptest.NewRecorder()

			// Execute
			handlers.TestAlertRules(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check response structure (actual fields from alert rules test)
			assert.Contains(t, response, "message")
			assert.Contains(t, response, "status")
			assert.Contains(t, response, "timestamp")
			assert.Contains(t, response, "rule_summary")
			assert.Contains(t, response, "alert_summary")
			assert.Contains(t, response, "test_results")

			// Check rule summary structure
			ruleSummary, ok := response["rule_summary"].(map[string]interface{})
			assert.True(t, ok)
			assert.Contains(t, ruleSummary, "total_groups")
			assert.Contains(t, ruleSummary, "alert_rules")
			assert.Contains(t, ruleSummary, "recording_rules")

			// Check alert summary structure
			alertSummary, ok := response["alert_summary"].(map[string]interface{})
			assert.True(t, ok)
			assert.Contains(t, alertSummary, "total_alerts")
			assert.Contains(t, alertSummary, "firing_alerts")
			assert.Contains(t, alertSummary, "pending_alerts")

			// Check test results array
			testResults, ok := response["test_results"].([]interface{})
			assert.True(t, ok)
			assert.NotEmpty(t, testResults)
		})
	}
}

// Benchmark tests for performance validation
func BenchmarkIntegrationHandlers_TestLGTMIntegration(b *testing.B) {
	loggingService := services.NewLoggingService()
	tracingService := services.NewTracingService()
	loggingService.InitLogger()
	tracingService.InitTracer()
	handlers := NewIntegrationHandlers(loggingService, tracingService)

	req := httptest.NewRequest("POST", "/test-lgtm-integration", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handlers.TestLGTMIntegration(w, req)
	}
}

func BenchmarkIntegrationHandlers_TestGrafanaDashboards(b *testing.B) {
	loggingService := services.NewLoggingService()
	tracingService := services.NewTracingService()
	loggingService.InitLogger()
	tracingService.InitTracer()
	handlers := NewIntegrationHandlers(loggingService, tracingService)

	req := httptest.NewRequest("POST", "/test-grafana-dashboards", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handlers.TestGrafanaDashboards(w, req)
	}
}

func BenchmarkIntegrationHandlers_TestAlertRules(b *testing.B) {
	loggingService := services.NewLoggingService()
	tracingService := services.NewTracingService()
	loggingService.InitLogger()
	tracingService.InitTracer()
	handlers := NewIntegrationHandlers(loggingService, tracingService)

	req := httptest.NewRequest("POST", "/test-alert-rules", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handlers.TestAlertRules(w, req)
	}
}
