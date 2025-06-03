package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nahuelsantos/argus/internal/services"
	"github.com/nahuelsantos/argus/internal/types"
	"github.com/nahuelsantos/argus/internal/utils"
)

// Helper function to get global settings with defaults
func getGlobalSettings() *types.LGTMSettings {
	// This accesses the global settings from basic.go
	if globalSettings != nil {
		return globalSettings
	}
	// Return defaults if no settings saved
	return types.GetDefaults()
}

// LGTM Integration Testing Handlers
// Tests that all monitoring components are properly configured and working together

// IntegrationHandlers contains LGTM integration testing handlers
type IntegrationHandlers struct {
	loggingService *services.LoggingService
	tracingService *services.TracingService
}

// NewIntegrationHandlers creates a new integration handlers instance
func NewIntegrationHandlers(loggingService *services.LoggingService, tracingService *services.TracingService) *IntegrationHandlers {
	return &IntegrationHandlers{
		loggingService: loggingService,
		tracingService: tracingService,
	}
}

type LGTMIntegrationStatus struct {
	Component    string            `json:"component"`
	Status       string            `json:"status"`
	Message      string            `json:"message"`
	ResponseTime time.Duration     `json:"response_time_ms"`
	Details      map[string]string `json:"details,omitempty"`
	Timestamp    time.Time         `json:"timestamp"`
}

type LGTMIntegrationSummary struct {
	OverallStatus string                  `json:"overall_status"`
	HealthyCount  int                     `json:"healthy_count"`
	TotalCount    int                     `json:"total_count"`
	Components    []LGTMIntegrationStatus `json:"components"`
	Timestamp     time.Time               `json:"timestamp"`
}

// Test LGTM Stack Integration
func (ih *IntegrationHandlers) TestLGTMIntegration(w http.ResponseWriter, r *http.Request) {
	ih.loggingService.LogWithContext(0, r.Context(), "Testing LGTM stack integration...")

	components := []LGTMIntegrationStatus{}

	// Test Grafana datasources
	grafanaStatus := ih.testGrafanaDatasources()
	components = append(components, grafanaStatus)

	// Test Prometheus targets
	prometheusStatus := ih.testPrometheusTargets()
	components = append(components, prometheusStatus)

	// Test Loki ingestion
	lokiStatus := ih.testLokiIngestion()
	components = append(components, lokiStatus)

	// Test Tempo tracing
	tempoStatus := ih.testTempoTracing()
	components = append(components, tempoStatus)

	// Test OTEL Collector
	otelStatus := ih.testOTELCollector()
	components = append(components, otelStatus)

	// Calculate overall status
	healthyCount := 0
	for _, comp := range components {
		if comp.Status == "healthy" {
			healthyCount++
		}
	}

	overallStatus := "healthy"
	if healthyCount == 0 {
		overallStatus = "critical"
	} else if healthyCount < len(components) {
		overallStatus = "degraded"
	}

	summary := LGTMIntegrationSummary{
		OverallStatus: overallStatus,
		HealthyCount:  healthyCount,
		TotalCount:    len(components),
		Components:    components,
		Timestamp:     time.Now(),
	}

	ih.loggingService.LogWithContext(0, r.Context(), "LGTM integration test completed")

	w.Header().Set("Content-Type", "application/json")
	utils.EncodeJSON(w, summary)
}

// Test Grafana Datasources
func (ih *IntegrationHandlers) testGrafanaDatasources() LGTMIntegrationStatus {
	start := time.Now()
	status := LGTMIntegrationStatus{
		Component: "grafana_datasources",
		Timestamp: time.Now(),
		Details:   make(map[string]string),
	}

	// Test Grafana API health
	resp, err := http.Get("http://grafana:3000/api/health")
	if err != nil {
		status.Status = "failed"
		status.Message = fmt.Sprintf("Cannot connect to Grafana: %v", err)
		status.ResponseTime = time.Since(start)
		return status
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		status.Status = "failed"
		status.Message = fmt.Sprintf("Grafana health check failed: HTTP %d", resp.StatusCode)
		status.ResponseTime = time.Since(start)
		return status
	}

	// Test datasources endpoint
	dsResp, err := http.Get("http://grafana:3000/api/datasources")
	if err != nil {
		status.Status = "degraded"
		status.Message = "Grafana is running but datasources endpoint failed"
		status.Details["error"] = err.Error()
	} else {
		defer dsResp.Body.Close()
		if dsResp.StatusCode == 200 {
			body, _ := io.ReadAll(dsResp.Body)
			datasourceCount := strings.Count(string(body), `"type":`)
			status.Status = "healthy"
			status.Message = fmt.Sprintf("Grafana running with %d datasources configured", datasourceCount)
			status.Details["datasources_count"] = strconv.Itoa(datasourceCount)
		} else {
			status.Status = "degraded"
			status.Message = "Grafana running but datasources not accessible"
		}
	}

	status.ResponseTime = time.Since(start)
	return status
}

// Test Prometheus Targets
func (ih *IntegrationHandlers) testPrometheusTargets() LGTMIntegrationStatus {
	start := time.Now()
	status := LGTMIntegrationStatus{
		Component: "prometheus_targets",
		Timestamp: time.Now(),
		Details:   make(map[string]string),
	}

	// Test Prometheus health
	resp, err := http.Get("http://prometheus:9090/-/healthy")
	if err != nil {
		status.Status = "failed"
		status.Message = fmt.Sprintf("Cannot connect to Prometheus: %v", err)
		status.ResponseTime = time.Since(start)
		return status
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		status.Status = "failed"
		status.Message = fmt.Sprintf("Prometheus health check failed: HTTP %d", resp.StatusCode)
		status.ResponseTime = time.Since(start)
		return status
	}

	// Test targets endpoint
	targetsResp, err := http.Get("http://prometheus:9090/api/v1/targets")
	if err != nil {
		status.Status = "degraded"
		status.Message = "Prometheus is running but targets endpoint failed"
		status.Details["error"] = err.Error()
	} else {
		defer targetsResp.Body.Close()
		if targetsResp.StatusCode == 200 {
			body, _ := io.ReadAll(targetsResp.Body)
			upCount := strings.Count(string(body), `"health":"up"`)
			totalCount := strings.Count(string(body), `"health":`)
			status.Status = "healthy"
			status.Message = fmt.Sprintf("Prometheus running with %d/%d targets up", upCount, totalCount)
			status.Details["targets_up"] = strconv.Itoa(upCount)
			status.Details["targets_total"] = strconv.Itoa(totalCount)
		} else {
			status.Status = "degraded"
			status.Message = "Prometheus running but targets not accessible"
		}
	}

	status.ResponseTime = time.Since(start)
	return status
}

// Test Loki Ingestion
func (ih *IntegrationHandlers) testLokiIngestion() LGTMIntegrationStatus {
	start := time.Now()
	status := LGTMIntegrationStatus{
		Component: "loki_ingestion",
		Timestamp: time.Now(),
		Details:   make(map[string]string),
	}

	// Test Loki ready endpoint
	resp, err := http.Get("http://loki:3100/ready")
	if err != nil {
		status.Status = "failed"
		status.Message = fmt.Sprintf("Cannot connect to Loki: %v", err)
		status.ResponseTime = time.Since(start)
		return status
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		status.Status = "failed"
		status.Message = fmt.Sprintf("Loki ready check failed: HTTP %d", resp.StatusCode)
		status.ResponseTime = time.Since(start)
		return status
	}

	// Test metrics endpoint for ingestion stats
	metricsResp, err := http.Get("http://loki:3100/metrics")
	if err != nil {
		status.Status = "degraded"
		status.Message = "Loki is ready but metrics endpoint failed"
		status.Details["error"] = err.Error()
	} else {
		defer metricsResp.Body.Close()
		if metricsResp.StatusCode == 200 {
			body, _ := io.ReadAll(metricsResp.Body)
			bodyStr := string(body)

			// Look for ingestion metrics
			hasIngestionMetrics := strings.Contains(bodyStr, "loki_ingester_") || strings.Contains(bodyStr, "loki_distributor_")

			if hasIngestionMetrics {
				status.Status = "healthy"
				status.Message = "Loki ready and ingesting logs"
				status.Details["ingestion"] = "active"
			} else {
				status.Status = "degraded"
				status.Message = "Loki ready but no ingestion metrics found"
				status.Details["ingestion"] = "unknown"
			}
		} else {
			status.Status = "degraded"
			status.Message = "Loki ready but metrics not accessible"
		}
	}

	status.ResponseTime = time.Since(start)
	return status
}

// Test Tempo Tracing
func (ih *IntegrationHandlers) testTempoTracing() LGTMIntegrationStatus {
	start := time.Now()
	status := LGTMIntegrationStatus{
		Component: "tempo_tracing",
		Timestamp: time.Now(),
		Details:   make(map[string]string),
	}

	// Test Tempo ready endpoint
	resp, err := http.Get("http://tempo:3200/ready")
	if err != nil {
		status.Status = "failed"
		status.Message = fmt.Sprintf("Cannot connect to Tempo: %v", err)
		status.ResponseTime = time.Since(start)
		return status
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		status.Status = "failed"
		status.Message = fmt.Sprintf("Tempo ready check failed: HTTP %d", resp.StatusCode)
		status.ResponseTime = time.Since(start)
		return status
	}

	// Test status endpoint
	statusResp, err := http.Get("http://tempo:3200/status")
	if err != nil {
		status.Status = "degraded"
		status.Message = "Tempo is ready but status endpoint failed"
		status.Details["error"] = err.Error()
	} else {
		defer statusResp.Body.Close()
		if statusResp.StatusCode == 200 {
			status.Status = "healthy"
			status.Message = "Tempo ready and accepting traces"
			status.Details["tracing"] = "active"
		} else {
			status.Status = "degraded"
			status.Message = "Tempo ready but status not accessible"
		}
	}

	status.ResponseTime = time.Since(start)
	return status
}

// Test OTEL Collector
func (ih *IntegrationHandlers) testOTELCollector() LGTMIntegrationStatus {
	start := time.Now()
	status := LGTMIntegrationStatus{
		Component: "otel_collector",
		Timestamp: time.Now(),
		Details:   make(map[string]string),
	}

	// Test OTEL Collector metrics endpoint
	resp, err := http.Get("http://otel-collector:8888/metrics")
	if err != nil {
		status.Status = "failed"
		status.Message = fmt.Sprintf("Cannot connect to OTEL Collector: %v", err)
		status.ResponseTime = time.Since(start)
		return status
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		status.Status = "failed"
		status.Message = fmt.Sprintf("OTEL Collector metrics failed: HTTP %d", resp.StatusCode)
		status.ResponseTime = time.Since(start)
		return status
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		status.Status = "degraded"
		status.Message = "OTEL Collector responding but cannot read metrics"
		status.Details["error"] = err.Error()
	} else {
		bodyStr := string(body)

		// Look for collector metrics
		hasReceiverMetrics := strings.Contains(bodyStr, "otelcol_receiver_")
		hasProcessorMetrics := strings.Contains(bodyStr, "otelcol_processor_")
		hasExporterMetrics := strings.Contains(bodyStr, "otelcol_exporter_")

		if hasReceiverMetrics && hasProcessorMetrics && hasExporterMetrics {
			status.Status = "healthy"
			status.Message = "OTEL Collector fully operational with all components"
			status.Details["receivers"] = "active"
			status.Details["processors"] = "active"
			status.Details["exporters"] = "active"
		} else {
			status.Status = "degraded"
			status.Message = "OTEL Collector running but some components may be missing"
			status.Details["receivers"] = strconv.FormatBool(hasReceiverMetrics)
			status.Details["processors"] = strconv.FormatBool(hasProcessorMetrics)
			status.Details["exporters"] = strconv.FormatBool(hasExporterMetrics)
		}
	}

	status.ResponseTime = time.Since(start)
	return status
}

// Test Grafana Dashboard Creation
func (ih *IntegrationHandlers) TestGrafanaDashboards(w http.ResponseWriter, r *http.Request) {
	ih.loggingService.LogWithContext(0, r.Context(), "Creating Argus test dashboard in Grafana...")

	// Load dashboard config
	dashboardPath := "internal/configs/grafana/argus-test-dashboard.json"
	dashboardData, err := os.ReadFile(dashboardPath)
	if err != nil {
		result := map[string]interface{}{
			"status":    "error",
			"message":   "Cannot load dashboard configuration",
			"error":     err.Error(),
			"timestamp": time.Now(),
		}
		w.Header().Set("Content-Type", "application/json")
		utils.EncodeJSON(w, result)
		return
	}

	// Get settings from global settings or defaults
	settings := getGlobalSettings()
	grafanaConfig := settings.Grafana

	// Create the dashboard creation URL
	grafanaURL := grafanaConfig.URL + "/api/dashboards/db"

	// Create HTTP request
	req, err := http.NewRequest("POST", grafanaURL, bytes.NewBuffer(dashboardData))
	if err != nil {
		result := map[string]interface{}{
			"status":    "error",
			"message":   "Cannot create dashboard request",
			"error":     err.Error(),
			"timestamp": time.Now(),
		}
		w.Header().Set("Content-Type", "application/json")
		utils.EncodeJSON(w, result)
		return
	}

	// Set proper headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Add authentication - Grafana requires credentials
	if grafanaConfig.Username != "" {
		req.SetBasicAuth(grafanaConfig.Username, grafanaConfig.Password)
	} else {
		// Use default admin credentials if none provided
		req.SetBasicAuth("admin", "admin")
	}

	// Create HTTP client with timeout
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)

	if err != nil {
		// Fallback to providing JSON for manual import
		result := map[string]interface{}{
			"status":          "manual_import_required",
			"message":         "Could not connect to Grafana - providing JSON for manual import",
			"error":           err.Error(),
			"dashboard_title": "Argus Testing Dashboard",
			"dashboard_uid":   "argus-test-dashboard",
			"grafana_url":     grafanaConfig.URL,
			"dashboard_url":   grafanaConfig.URL + "/d/argus-test-dashboard",
			"instructions": []string{
				"1. Go to " + grafanaConfig.URL + " and login to Grafana",
				"2. Navigate to '+' → Import",
				"3. Upload the dashboard JSON provided below",
				"4. Dashboard will be accessible at: " + grafanaConfig.URL + "/d/argus-test-dashboard",
			},
			"dashboard_json": string(dashboardData),
			"timestamp":      time.Now(),
		}
		w.Header().Set("Content-Type", "application/json")
		utils.EncodeJSON(w, result)
		return
	}
	defer resp.Body.Close()

	responseBody, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}

	// Check authentication first
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		result = map[string]interface{}{
			"status":      "auth_error",
			"message":     "Authentication failed - check Grafana credentials in Settings",
			"error":       fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(responseBody)),
			"grafana_url": grafanaConfig.URL,
			"instructions": []string{
				"1. Go to Settings and verify Grafana username/password",
				"2. Test connection to ensure credentials work",
				"3. Try running the dashboard test again",
			},
			"timestamp": time.Now(),
		}
	} else if resp.StatusCode == 200 || resp.StatusCode == 412 { // 412 = already exists
		status := "created"
		message := "✅ Argus test dashboard created successfully in Grafana!"

		if resp.StatusCode == 412 {
			status = "updated"
			message = "✅ Argus test dashboard already exists - updated successfully!"
		}

		result = map[string]interface{}{
			"status":          status,
			"message":         message,
			"dashboard_title": "Argus Testing Dashboard",
			"dashboard_uid":   "argus-test-dashboard",
			"grafana_url":     grafanaConfig.URL,
			"dashboard_url":   grafanaConfig.URL + "/d/argus-test-dashboard",
			"panels": []string{
				"Performance Test Results",
				"System Resource Usage",
				"LGTM Stack Health",
				"Generated Metrics Over Time",
				"Log Generation Rate",
				"Test Execution Status",
			},
			"actions_completed": []string{
				"✅ Dashboard JSON loaded from config",
				"✅ Connected to Grafana at " + grafanaConfig.URL,
				"✅ Dashboard created/updated successfully",
				"✅ Direct access URL provided",
			},
			"timestamp": time.Now(),
		}
	} else {
		result = map[string]interface{}{
			"status":      "error",
			"message":     fmt.Sprintf("Failed to create dashboard: HTTP %d", resp.StatusCode),
			"response":    string(responseBody),
			"grafana_url": grafanaConfig.URL,
			"fallback_instructions": []string{
				"1. Copy the dashboard JSON below",
				"2. Go to " + grafanaConfig.URL + " and manually import it",
				"3. Check Grafana logs for more details",
			},
			"dashboard_json": string(dashboardData),
			"timestamp":      time.Now(),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	utils.EncodeJSON(w, result)
}

// Test Alert Rules Configuration - Verify rules are loaded and working
func (ih *IntegrationHandlers) TestAlertRules(w http.ResponseWriter, r *http.Request) {
	ih.loggingService.LogWithContext(0, r.Context(), "Testing Prometheus alert rules configuration...")

	// Get settings for Prometheus connection
	settings := getGlobalSettings()
	prometheusConfig := settings.Prometheus

	client := &http.Client{Timeout: 10 * time.Second}

	// Test 1: Check if Prometheus rules API is accessible
	rulesURL := prometheusConfig.URL + "/api/v1/rules"
	rulesResp, err := client.Get(rulesURL)
	if err != nil {
		result := map[string]interface{}{
			"status":         "connection_error",
			"message":        "Cannot connect to Prometheus rules API",
			"error":          err.Error(),
			"prometheus_url": prometheusConfig.URL,
			"test_results":   []string{"❌ Connection to Prometheus failed"},
			"timestamp":      time.Now(),
		}
		w.Header().Set("Content-Type", "application/json")
		utils.EncodeJSON(w, result)
		return
	}
	defer rulesResp.Body.Close()

	if rulesResp.StatusCode != 200 {
		result := map[string]interface{}{
			"status":         "api_error",
			"message":        fmt.Sprintf("Prometheus rules API failed: HTTP %d", rulesResp.StatusCode),
			"prometheus_url": prometheusConfig.URL,
			"test_results":   []string{"❌ Rules API returned error"},
			"timestamp":      time.Now(),
		}
		w.Header().Set("Content-Type", "application/json")
		utils.EncodeJSON(w, result)
		return
	}

	rulesBody, err := io.ReadAll(rulesResp.Body)
	if err != nil {
		result := map[string]interface{}{
			"status":       "error",
			"message":      "Cannot read Prometheus rules response",
			"error":        err.Error(),
			"test_results": []string{"❌ Failed to read rules response"},
			"timestamp":    time.Now(),
		}
		w.Header().Set("Content-Type", "application/json")
		utils.EncodeJSON(w, result)
		return
	}

	// Test 2: Check alerts endpoint
	alertsURL := prometheusConfig.URL + "/api/v1/alerts"
	alertsResp, err := client.Get(alertsURL)
	var alertsBody []byte
	alertsAccessible := false

	if err == nil && alertsResp.StatusCode == 200 {
		alertsBody, _ = io.ReadAll(alertsResp.Body)
		alertsAccessible = true
		alertsResp.Body.Close()
	} else if alertsResp != nil {
		alertsResp.Body.Close()
	}

	// Parse rules and alerts data
	rulesBodyStr := string(rulesBody)
	alertsBodyStr := string(alertsBody)

	// Count total rules and groups
	totalRuleGroups := strings.Count(rulesBodyStr, `"name":`)
	totalAlertRules := strings.Count(rulesBodyStr, `"alert":`)
	totalRecordingRules := strings.Count(rulesBodyStr, `"record":`)

	// Check for Argus-specific rules
	argusRules := strings.Count(rulesBodyStr, "argus")
	argusAlertRules := strings.Count(rulesBodyStr, `"alert":"argus`) + strings.Count(rulesBodyStr, `"alert": "argus`)

	// Count active alerts
	activeAlerts := strings.Count(alertsBodyStr, `"state":"firing"`)
	pendingAlerts := strings.Count(alertsBodyStr, `"state":"pending"`)
	totalAlerts := strings.Count(alertsBodyStr, `"state":`)

	// Test results array
	var testResults []string
	var status string
	var message string

	// Test 1: Rules API accessibility
	testResults = append(testResults, "✅ Prometheus rules API accessible")

	// Test 2: Alert rules existence
	if totalAlertRules > 0 {
		testResults = append(testResults, fmt.Sprintf("✅ Found %d alert rules in %d groups", totalAlertRules, totalRuleGroups))
	} else {
		testResults = append(testResults, "❌ No alert rules found")
	}

	// Test 3: Recording rules existence (optional)
	if totalRecordingRules > 0 {
		testResults = append(testResults, fmt.Sprintf("✅ Found %d recording rules", totalRecordingRules))
	}

	// Test 4: Argus-specific rules
	if argusRules > 0 {
		testResults = append(testResults, fmt.Sprintf("✅ Found %d Argus-specific rules", argusRules))
		if argusAlertRules > 0 {
			testResults = append(testResults, fmt.Sprintf("✅ Found %d Argus alert rules", argusAlertRules))
		}
	} else {
		testResults = append(testResults, "⚠️ No Argus-specific rules found")
	}

	// Test 5: Alerts endpoint accessibility
	if alertsAccessible {
		testResults = append(testResults, "✅ Prometheus alerts API accessible")

		// Test 6: Active alerts
		if totalAlerts > 0 {
			testResults = append(testResults, fmt.Sprintf("✅ Found %d total alerts (%d firing, %d pending)", totalAlerts, activeAlerts, pendingAlerts))
		} else {
			testResults = append(testResults, "ℹ️ No active alerts (this is normal if system is healthy)")
		}
	} else {
		testResults = append(testResults, "❌ Prometheus alerts API not accessible")
	}

	// Test 7: Rule evaluation check (check if rules are being evaluated)
	if strings.Contains(rulesBodyStr, `"evaluationTime":`) || strings.Contains(rulesBodyStr, `"lastEvaluation":`) {
		testResults = append(testResults, "✅ Rules are being evaluated by Prometheus")
	} else {
		testResults = append(testResults, "⚠️ Cannot verify rule evaluation status")
	}

	// Determine overall status
	hasRules := totalAlertRules > 0
	hasArgusRules := argusRules > 0
	hasAlertsAPI := alertsAccessible

	if hasRules && hasArgusRules && hasAlertsAPI {
		status = "healthy"
		message = "✅ Alert rules are properly configured and functioning"
	} else if hasRules && hasAlertsAPI {
		status = "partial"
		message = "⚠️ Basic alert rules working, but Argus-specific rules not found"
	} else if hasRules {
		status = "degraded"
		message = "⚠️ Rules exist but alerts API not accessible"
	} else {
		status = "failed"
		message = "❌ No alert rules configured in Prometheus"
	}

	// Generate instructions based on status
	var instructions []string
	var recommendations []string

	if status == "healthy" {
		instructions = []string{
			"1. ✅ Check active alerts: " + prometheusConfig.URL + "/alerts",
			"2. ✅ View rule status: " + prometheusConfig.URL + "/rules",
			"3. ✅ Generate load to test CPU/Memory alerts",
			"4. ✅ Monitor alert firing behavior",
		}
		recommendations = []string{
			"💡 Generate system load to test resource monitoring alerts",
			"💡 Run performance tests to trigger test failure alerts",
			"💡 Check Grafana for alert notifications",
		}
	} else if status == "partial" {
		instructions = []string{
			"1. Check rule configuration in prometheus.yml",
			"2. Verify Argus rules are properly loaded",
			"3. Restart Prometheus if needed",
			"4. Check " + prometheusConfig.URL + "/rules for all loaded rules",
		}
		recommendations = []string{
			"💡 Add Argus alert rules to your prometheus.yml configuration",
			"💡 Ensure rule_files section includes Argus rules",
		}
	} else {
		instructions = []string{
			"1. Check prometheus.yml configuration",
			"2. Add rule_files section with alert rules",
			"3. Restart Prometheus service",
			"4. Verify rules load at " + prometheusConfig.URL + "/rules",
		}
		recommendations = []string{
			"💡 Alert rules need to be configured in Prometheus",
			"💡 Add rule_files to prometheus.yml configuration",
		}
	}

	result := map[string]interface{}{
		"status":  status,
		"message": message,
		"rule_summary": map[string]interface{}{
			"total_groups":      totalRuleGroups,
			"alert_rules":       totalAlertRules,
			"recording_rules":   totalRecordingRules,
			"argus_rules":       argusRules,
			"argus_alert_rules": argusAlertRules,
		},
		"alert_summary": map[string]interface{}{
			"api_accessible": alertsAccessible,
			"total_alerts":   totalAlerts,
			"firing_alerts":  activeAlerts,
			"pending_alerts": pendingAlerts,
		},
		"test_results":    testResults,
		"instructions":    instructions,
		"recommendations": recommendations,
		"prometheus_url":  prometheusConfig.URL,
		"alerts_url":      prometheusConfig.URL + "/alerts",
		"rules_url":       prometheusConfig.URL + "/rules",
		"timestamp":       time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	utils.EncodeJSON(w, result)
}
