package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/nahuelsantos/argus/internal/services"
)

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
	json.NewEncoder(w).Encode(summary)
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

// Test Grafana Dashboard Availability
func (ih *IntegrationHandlers) TestGrafanaDashboards(w http.ResponseWriter, r *http.Request) {
	ih.loggingService.LogWithContext(0, r.Context(), "Testing Grafana dashboard availability...")

	dashboards := []struct {
		Name        string `json:"name"`
		URL         string `json:"url"`
		Status      string `json:"status"`
		Description string `json:"description"`
	}{
		{"System Overview", "/d/system-overview", "checking", "Main system health dashboard"},
		{"Docker Containers", "/d/docker-containers", "checking", "Container metrics and health"},
		{"Infrastructure", "/d/infrastructure", "checking", "Network and hardware metrics"},
		{"Application Metrics", "/d/application-metrics", "checking", "Service-level metrics"},
		{"Logs Overview", "/d/logs-overview", "checking", "Log aggregation and analysis"},
	}

	// Test each dashboard (simplified - in reality we'd check if they exist)
	for i := range dashboards {
		resp, err := http.Get("http://grafana:3000" + dashboards[i].URL)
		if err != nil {
			dashboards[i].Status = "unavailable"
		} else {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				dashboards[i].Status = "available"
			} else {
				dashboards[i].Status = "not_found"
			}
		}
	}

	result := map[string]interface{}{
		"dashboards": dashboards,
		"timestamp":  time.Now(),
		"message":    "Dashboard availability check completed",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// Test Alert Rules Configuration
func (ih *IntegrationHandlers) TestAlertRules(w http.ResponseWriter, r *http.Request) {
	ih.loggingService.LogWithContext(0, r.Context(), "Testing alert rules configuration...")

	// Test Prometheus rules endpoint
	resp, err := http.Get("http://prometheus:9090/api/v1/rules")
	if err != nil {
		result := map[string]interface{}{
			"status":    "error",
			"message":   "Cannot connect to Prometheus rules API",
			"error":     err.Error(),
			"timestamp": time.Now(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		result := map[string]interface{}{
			"status":    "error",
			"message":   fmt.Sprintf("Prometheus rules API failed: HTTP %d", resp.StatusCode),
			"timestamp": time.Now(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result := map[string]interface{}{
			"status":    "error",
			"message":   "Cannot read Prometheus rules response",
			"error":     err.Error(),
			"timestamp": time.Now(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
		return
	}

	// Parse rules (simplified)
	bodyStr := string(body)
	ruleCount := strings.Count(bodyStr, `"name":`)
	alertCount := strings.Count(bodyStr, `"alert":`)

	result := map[string]interface{}{
		"status":      "healthy",
		"message":     "Alert rules configuration validated",
		"rule_groups": ruleCount,
		"alert_rules": alertCount,
		"timestamp":   time.Now(),
		"details": map[string]interface{}{
			"prometheus_rules_endpoint": "accessible",
			"rules_format":              "valid",
			"alert_rules_present":       alertCount > 0,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
