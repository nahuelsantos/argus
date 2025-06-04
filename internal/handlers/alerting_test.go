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

func TestNewAlertingHandlers(t *testing.T) {
	loggingService := services.NewLoggingService()
	alertingService := services.NewAlertingService()
	loggingService.InitTestLogger()

	handlers := NewAlertingHandlers(loggingService, alertingService)

	assert.NotNil(t, handlers)
	assert.NotNil(t, handlers.loggingService)
	assert.NotNil(t, handlers.alertingService)
}

func TestAlertingHandlers_TestAlertRulesHandler(t *testing.T) {
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
			alertingService := services.NewAlertingService()
			loggingService.InitTestLogger()
			handlers := NewAlertingHandlers(loggingService, alertingService)

			// Create request
			req := httptest.NewRequest(tt.method, "/test-alert-rules", nil)
			w := httptest.NewRecorder()

			// Execute
			handlers.TestAlertRulesHandler(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check response structure
			assert.Contains(t, response, "message")
			assert.Contains(t, response, "total_rules")
			assert.Contains(t, response, "enabled_rules")
			assert.Contains(t, response, "active_alerts")
			assert.Contains(t, response, "rules")
			assert.Contains(t, response, "timestamp")
			assert.Contains(t, response, "test_status")
			assert.Contains(t, response, "service")
			assert.Contains(t, response, "functionality")

			// Verify content
			assert.Equal(t, "Alert rules functionality tested", response["message"])
			assert.Equal(t, "success", response["test_status"])
			assert.Equal(t, "argus", response["service"])
			assert.Equal(t, "alert_management", response["functionality"])

			// Check numeric values
			totalRules, ok := response["total_rules"].(float64)
			assert.True(t, ok)
			assert.GreaterOrEqual(t, totalRules, 0.0)

			enabledRules, ok := response["enabled_rules"].(float64)
			assert.True(t, ok)
			assert.GreaterOrEqual(t, enabledRules, 0.0)
			assert.LessOrEqual(t, enabledRules, totalRules)

			activeAlerts, ok := response["active_alerts"].(float64)
			assert.True(t, ok)
			assert.GreaterOrEqual(t, activeAlerts, 0.0)

			// Check rules array
			rules, ok := response["rules"].([]interface{})
			assert.True(t, ok)
			assert.NotNil(t, rules)
		})
	}
}

func TestAlertingHandlers_TestFireAlertHandler(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    map[string]string
		expectedStatus int
	}{
		{
			name:           "fire alert with default values",
			queryParams:    map[string]string{},
			expectedStatus: http.StatusOK,
		},
		{
			name: "fire specific alert type",
			queryParams: map[string]string{
				"type":     "high-memory-usage",
				"severity": "warning",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "fire critical alert",
			queryParams: map[string]string{
				"type":     "disk-space-low",
				"severity": "critical",
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			loggingService := services.NewLoggingService()
			alertingService := services.NewAlertingService()
			loggingService.InitTestLogger()
			handlers := NewAlertingHandlers(loggingService, alertingService)

			// Build query string
			path := "/test-fire-alert"
			if len(tt.queryParams) > 0 {
				path += "?"
				for key, value := range tt.queryParams {
					path += key + "=" + value + "&"
				}
				path = path[:len(path)-1] // Remove trailing &
			}

			// Create request
			req := httptest.NewRequest("POST", path, nil)
			w := httptest.NewRecorder()

			// Execute
			handlers.TestFireAlertHandler(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check response structure
			assert.Contains(t, response, "message")
			assert.Contains(t, response, "alert_type")
			assert.Contains(t, response, "severity")
			assert.Contains(t, response, "active_alerts")
			assert.Contains(t, response, "timestamp")
			assert.Contains(t, response, "test_status")
			assert.Contains(t, response, "service")
			assert.Contains(t, response, "functionality")

			// Verify content
			assert.Equal(t, "Alert fired successfully", response["message"])
			assert.Equal(t, "success", response["test_status"])
			assert.Equal(t, "argus", response["service"])
			assert.Equal(t, "alert_firing", response["functionality"])

			// Check alert type (default or specified)
			expectedType := "high-cpu-usage"
			if tt.queryParams["type"] != "" {
				expectedType = tt.queryParams["type"]
			}
			assert.Equal(t, expectedType, response["alert_type"])

			// Check severity (default or specified)
			expectedSeverity := "critical"
			if tt.queryParams["severity"] != "" {
				expectedSeverity = tt.queryParams["severity"]
			}
			assert.Equal(t, expectedSeverity, response["severity"])

			// Check active alerts count
			activeAlerts, ok := response["active_alerts"].(float64)
			assert.True(t, ok)
			assert.GreaterOrEqual(t, activeAlerts, 0.0)
		})
	}
}

func TestAlertingHandlers_TestIncidentManagementHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "successful incident management test",
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
			alertingService := services.NewAlertingService()
			loggingService.InitTestLogger()
			handlers := NewAlertingHandlers(loggingService, alertingService)

			// Create request
			req := httptest.NewRequest(tt.method, "/test-incident-management", nil)
			w := httptest.NewRecorder()

			// Execute
			handlers.TestIncidentManagementHandler(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check response structure
			assert.Contains(t, response, "message")
			assert.Contains(t, response, "created_incident")
			assert.Contains(t, response, "total_incidents")
			assert.Contains(t, response, "open_incidents")
			assert.Contains(t, response, "resolved_incidents")
			assert.Contains(t, response, "timestamp")
			assert.Contains(t, response, "test_status")
			assert.Contains(t, response, "service")
			assert.Contains(t, response, "functionality")

			// Verify content
			assert.Equal(t, "Incident management tested", response["message"])
			assert.Equal(t, "success", response["test_status"])
			assert.Equal(t, "argus", response["service"])
			assert.Equal(t, "incident_management", response["functionality"])

			// Check numeric values
			totalIncidents, ok := response["total_incidents"].(float64)
			assert.True(t, ok)
			assert.GreaterOrEqual(t, totalIncidents, 1.0) // Should have at least the test incident

			openIncidents, ok := response["open_incidents"].(float64)
			assert.True(t, ok)
			assert.GreaterOrEqual(t, openIncidents, 0.0)

			resolvedIncidents, ok := response["resolved_incidents"].(float64)
			assert.True(t, ok)
			assert.GreaterOrEqual(t, resolvedIncidents, 0.0)

			// Check created incident
			createdIncident, ok := response["created_incident"].(map[string]interface{})
			assert.True(t, ok)
			assert.NotNil(t, createdIncident)
			assert.Contains(t, createdIncident, "id")
			assert.Contains(t, createdIncident, "title")
			assert.Contains(t, createdIncident, "status")
			assert.Equal(t, "Test Incident", createdIncident["title"])
			assert.Equal(t, "open", createdIncident["status"])
		})
	}
}

func TestAlertingHandlers_TestNotificationChannelsHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "successful notification channels test",
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
			alertingService := services.NewAlertingService()
			loggingService.InitTestLogger()
			handlers := NewAlertingHandlers(loggingService, alertingService)

			// Create request
			req := httptest.NewRequest(tt.method, "/test-notification-channels", nil)
			w := httptest.NewRecorder()

			// Execute
			handlers.TestNotificationChannelsHandler(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check response structure
			assert.Contains(t, response, "message")
			assert.Contains(t, response, "timestamp")
			assert.Contains(t, response, "test_status")
			assert.Contains(t, response, "service")
			assert.Contains(t, response, "functionality")

			// Verify content
			assert.Equal(t, "success", response["test_status"])
			assert.Equal(t, "argus", response["service"])
		})
	}
}

func TestAlertingHandlers_GetActiveAlertsHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "get active alerts",
			method:         "GET",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			loggingService := services.NewLoggingService()
			alertingService := services.NewAlertingService()
			loggingService.InitTestLogger()
			handlers := NewAlertingHandlers(loggingService, alertingService)

			// Create request
			req := httptest.NewRequest(tt.method, "/active-alerts", nil)
			w := httptest.NewRecorder()

			// Execute
			handlers.GetActiveAlertsHandler(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check response structure (actual fields from GetActiveAlertsHandler)
			assert.Contains(t, response, "active_alerts")
			assert.Contains(t, response, "active_count")
			assert.Contains(t, response, "recent_alerts")
			assert.Contains(t, response, "recent_count")
			assert.Contains(t, response, "functionality")
			assert.Contains(t, response, "service")
			assert.Contains(t, response, "timestamp")

			// Check alerts array
			alerts, ok := response["active_alerts"].([]interface{})
			assert.True(t, ok)
			assert.NotNil(t, alerts)

			// Check active count
			activeCount, ok := response["active_count"].(float64)
			assert.True(t, ok)
			assert.GreaterOrEqual(t, activeCount, 0.0)
			assert.Equal(t, float64(len(alerts)), activeCount)

			// Check recent alerts
			recentAlerts, ok := response["recent_alerts"].([]interface{})
			assert.True(t, ok)
			assert.NotNil(t, recentAlerts)

			// Verify content
			assert.Equal(t, "active_alerts_monitoring", response["functionality"])
			assert.Equal(t, "argus", response["service"])
		})
	}
}

func TestAlertingHandlers_GetActiveIncidentsHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "get active incidents",
			method:         "GET",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			loggingService := services.NewLoggingService()
			alertingService := services.NewAlertingService()
			loggingService.InitTestLogger()
			handlers := NewAlertingHandlers(loggingService, alertingService)

			// Create request
			req := httptest.NewRequest(tt.method, "/active-incidents", nil)
			w := httptest.NewRecorder()

			// Execute
			handlers.GetActiveIncidentsHandler(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check response structure (actual fields from GetActiveIncidentsHandler)
			assert.Contains(t, response, "active_incidents")
			assert.Contains(t, response, "incident_statistics")
			assert.Contains(t, response, "priority_breakdown")
			assert.Contains(t, response, "mttr_minutes")
			assert.Contains(t, response, "functionality")
			assert.Contains(t, response, "service")
			assert.Contains(t, response, "timestamp")

			// Check incidents array
			incidents, ok := response["active_incidents"].([]interface{})
			assert.True(t, ok)
			assert.NotNil(t, incidents)

			// Check incident statistics
			stats, ok := response["incident_statistics"].(map[string]interface{})
			assert.True(t, ok)
			assert.NotNil(t, stats)
			assert.Contains(t, stats, "total")

			// Check priority breakdown
			priority, ok := response["priority_breakdown"].(map[string]interface{})
			assert.True(t, ok)
			assert.NotNil(t, priority)

			// Check MTTR
			mttr, ok := response["mttr_minutes"].(float64)
			assert.True(t, ok)
			assert.GreaterOrEqual(t, mttr, 0.0)

			// Verify content
			assert.Equal(t, "incident_monitoring", response["functionality"])
			assert.Equal(t, "argus", response["service"])
		})
	}
}

// Benchmark tests for alerting handlers
func BenchmarkAlertingHandlers_TestAlertRules(b *testing.B) {
	loggingService := services.NewLoggingService()
	alertingService := services.NewAlertingService()
	loggingService.InitTestLogger()
	handlers := NewAlertingHandlers(loggingService, alertingService)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/test-alert-rules", nil)
		w := httptest.NewRecorder()
		handlers.TestAlertRulesHandler(w, req)
	}
}

func BenchmarkAlertingHandlers_FireAlert(b *testing.B) {
	loggingService := services.NewLoggingService()
	alertingService := services.NewAlertingService()
	loggingService.InitTestLogger()
	handlers := NewAlertingHandlers(loggingService, alertingService)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/test-fire-alert?type=high-cpu-usage&severity=critical", nil)
		w := httptest.NewRecorder()
		handlers.TestFireAlertHandler(w, req)
	}
}

func BenchmarkAlertingHandlers_IncidentManagement(b *testing.B) {
	loggingService := services.NewLoggingService()
	alertingService := services.NewAlertingService()
	loggingService.InitTestLogger()
	handlers := NewAlertingHandlers(loggingService, alertingService)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/test-incident-management", nil)
		w := httptest.NewRecorder()
		handlers.TestIncidentManagementHandler(w, req)
	}
}
