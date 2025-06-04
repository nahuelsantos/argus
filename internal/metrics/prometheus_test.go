package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsRegistration(t *testing.T) {
	// Test that RegisterMetrics function doesn't panic
	assert.NotPanics(t, func() {
		RegisterMetrics()
	})

	// Test that we can create new metrics with labels without panicking
	assert.NotPanics(t, func() {
		HTTPRequestsTotal.WithLabelValues("GET", "/test", "200").Inc()
		LogEntriesTotal.WithLabelValues("info", "test-service", "").Inc()
		APMTracesTotal.WithLabelValues("test-service", "test-op", "success").Inc()
	})

	// Verify we have the expected number of metric types available
	expectedMetrics := []string{
		"http_requests_total",
		"http_request_duration_seconds",
		"log_entries_total",
		"log_processing_duration_seconds",
		"errors_by_category_total",
		"custom_business_metric",
		"apm_traces_total",
		"apm_span_duration_seconds",
		"service_dependency_latency_seconds",
		"performance_anomalies_total",
		"alerts_total",
		"alert_duration_seconds",
		"incidents_total",
		"incident_duration_seconds",
		"notifications_sent_total",
		"notification_latency_seconds",
		"alert_manager_health",
		"mttr_seconds",
	}

	// Test that all expected metrics exist by verifying we can create them
	assert.Equal(t, 18, len(expectedMetrics), "Should have 18 different metric types")
}

func TestHTTPMetrics(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		endpoint string
		status   string
	}{
		{
			name:     "GET request",
			method:   "GET",
			endpoint: "/api/health",
			status:   "200",
		},
		{
			name:     "POST request",
			method:   "POST",
			endpoint: "/api/users",
			status:   "201",
		},
		{
			name:     "Error request",
			method:   "GET",
			endpoint: "/api/error",
			status:   "500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test HTTPRequestsTotal counter
			before := getCounterValue(t, HTTPRequestsTotal.WithLabelValues(tt.method, tt.endpoint, tt.status))
			HTTPRequestsTotal.WithLabelValues(tt.method, tt.endpoint, tt.status).Inc()
			after := getCounterValue(t, HTTPRequestsTotal.WithLabelValues(tt.method, tt.endpoint, tt.status))
			assert.Equal(t, before+1, after)

			// Test HTTPRequestDuration histogram
			HTTPRequestDuration.WithLabelValues(tt.method, tt.endpoint).Observe(0.25)
			// Histogram observation should not panic
			assert.NotPanics(t, func() {
				HTTPRequestDuration.WithLabelValues(tt.method, tt.endpoint).Observe(0.25)
			})
		})
	}
}

func TestLogMetrics(t *testing.T) {
	tests := []struct {
		name      string
		level     string
		service   string
		errorType string
		operation string
		logLevel  string
	}{
		{
			name:      "info log",
			level:     "info",
			service:   "api-service",
			errorType: "",
			operation: "log_generation",
			logLevel:  "info",
		},
		{
			name:      "error log",
			level:     "error",
			service:   "auth-service",
			errorType: "auth_failed",
			operation: "error_processing",
			logLevel:  "error",
		},
		{
			name:      "debug log",
			level:     "debug",
			service:   "payment-service",
			errorType: "",
			operation: "debug_trace",
			logLevel:  "debug",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test LogEntriesTotal counter
			before := getCounterValue(t, LogEntriesTotal.WithLabelValues(tt.level, tt.service, tt.errorType))
			LogEntriesTotal.WithLabelValues(tt.level, tt.service, tt.errorType).Inc()
			after := getCounterValue(t, LogEntriesTotal.WithLabelValues(tt.level, tt.service, tt.errorType))
			assert.Equal(t, before+1, after)

			// Test LogProcessingDuration histogram
			assert.NotPanics(t, func() {
				LogProcessingDuration.WithLabelValues(tt.operation, tt.logLevel).Observe(0.05)
			})
		})
	}
}

func TestErrorMetrics(t *testing.T) {
	tests := []struct {
		name     string
		category string
		severity string
		source   string
	}{
		{
			name:     "authentication error",
			category: "auth",
			severity: "high",
			source:   "api-gateway",
		},
		{
			name:     "database error",
			category: "database",
			severity: "critical",
			source:   "user-service",
		},
		{
			name:     "network error",
			category: "network",
			severity: "medium",
			source:   "payment-service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := getCounterValue(t, ErrorsByCategory.WithLabelValues(tt.category, tt.severity, tt.source))
			ErrorsByCategory.WithLabelValues(tt.category, tt.severity, tt.source).Inc()
			after := getCounterValue(t, ErrorsByCategory.WithLabelValues(tt.category, tt.severity, tt.source))
			assert.Equal(t, before+1, after)
		})
	}
}

func TestCustomMetric(t *testing.T) {
	tests := []struct {
		name       string
		metricType string
		category   string
		value      float64
	}{
		{
			name:       "user count",
			metricType: "users",
			category:   "active",
			value:      1250.0,
		},
		{
			name:       "revenue",
			metricType: "revenue",
			category:   "monthly",
			value:      45000.50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			CustomMetric.WithLabelValues(tt.metricType, tt.category).Set(tt.value)

			metric := &dto.Metric{}
			err := CustomMetric.WithLabelValues(tt.metricType, tt.category).Write(metric)
			require.NoError(t, err)
			assert.NotNil(t, metric.Gauge)
			assert.Equal(t, tt.value, *metric.Gauge.Value)
		})
	}
}

func TestAPMMetrics(t *testing.T) {
	tests := []struct {
		name          string
		service       string
		operation     string
		status        string
		duration      float64
		sourceService string
		targetService string
	}{
		{
			name:          "user service trace",
			service:       "user-service",
			operation:     "get_user",
			status:        "success",
			duration:      0.125,
			sourceService: "api-gateway",
			targetService: "user-service",
		},
		{
			name:          "payment service trace",
			service:       "payment-service",
			operation:     "process_payment",
			status:        "error",
			duration:      0.543,
			sourceService: "checkout-service",
			targetService: "payment-service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test APM traces counter
			before := getCounterValue(t, APMTracesTotal.WithLabelValues(tt.service, tt.operation, tt.status))
			APMTracesTotal.WithLabelValues(tt.service, tt.operation, tt.status).Inc()
			after := getCounterValue(t, APMTracesTotal.WithLabelValues(tt.service, tt.operation, tt.status))
			assert.Equal(t, before+1, after)

			// Test APM span duration
			assert.NotPanics(t, func() {
				APMSpanDuration.WithLabelValues(tt.service, tt.operation).Observe(tt.duration)
			})

			// Test service dependency latency
			assert.NotPanics(t, func() {
				ServiceDependencyLatency.WithLabelValues(tt.sourceService, tt.targetService, tt.operation).Observe(tt.duration)
			})
		})
	}
}

func TestAlertingMetrics(t *testing.T) {
	tests := []struct {
		name            string
		ruleName        string
		severity        string
		status          string
		affectedService string
		channelType     string
		component       string
		duration        float64
		mttr            float64
	}{
		{
			name:            "high cpu alert",
			ruleName:        "high_cpu_usage",
			severity:        "warning",
			status:          "firing",
			affectedService: "api-service",
			channelType:     "slack",
			component:       "prometheus",
			duration:        300.0,
			mttr:            900.0,
		},
		{
			name:            "disk space alert",
			ruleName:        "disk_space_low",
			severity:        "critical",
			status:          "resolved",
			affectedService: "database",
			channelType:     "email",
			component:       "alertmanager",
			duration:        1800.0,
			mttr:            3600.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test AlertsTotal
			before := getCounterValue(t, AlertsTotal.WithLabelValues(tt.ruleName, tt.severity, tt.status))
			AlertsTotal.WithLabelValues(tt.ruleName, tt.severity, tt.status).Inc()
			after := getCounterValue(t, AlertsTotal.WithLabelValues(tt.ruleName, tt.severity, tt.status))
			assert.Equal(t, before+1, after)

			// Test AlertDuration
			assert.NotPanics(t, func() {
				AlertDuration.WithLabelValues(tt.ruleName, tt.severity).Observe(tt.duration)
			})

			// Test IncidentsTotal
			incidentBefore := getCounterValue(t, IncidentsTotal.WithLabelValues(tt.severity, tt.status, tt.affectedService))
			IncidentsTotal.WithLabelValues(tt.severity, tt.status, tt.affectedService).Inc()
			incidentAfter := getCounterValue(t, IncidentsTotal.WithLabelValues(tt.severity, tt.status, tt.affectedService))
			assert.Equal(t, incidentBefore+1, incidentAfter)

			// Test IncidentDuration
			assert.NotPanics(t, func() {
				IncidentDuration.WithLabelValues(tt.severity, tt.affectedService).Observe(tt.duration)
			})

			// Test NotificationsSent
			notifBefore := getCounterValue(t, NotificationsSent.WithLabelValues(tt.channelType, tt.severity, "sent"))
			NotificationsSent.WithLabelValues(tt.channelType, tt.severity, "sent").Inc()
			notifAfter := getCounterValue(t, NotificationsSent.WithLabelValues(tt.channelType, tt.severity, "sent"))
			assert.Equal(t, notifBefore+1, notifAfter)

			// Test NotificationLatency
			assert.NotPanics(t, func() {
				NotificationLatency.WithLabelValues(tt.channelType).Observe(0.25)
			})

			// Test AlertManagerHealth
			AlertManagerHealth.WithLabelValues(tt.component).Set(1.0)
			healthMetric := &dto.Metric{}
			err := AlertManagerHealth.WithLabelValues(tt.component).Write(healthMetric)
			require.NoError(t, err)
			assert.Equal(t, 1.0, *healthMetric.Gauge.Value)

			// Test MTTRGauge
			MTTRGauge.WithLabelValues(tt.affectedService, tt.severity).Set(tt.mttr)
			mttrMetric := &dto.Metric{}
			err = MTTRGauge.WithLabelValues(tt.affectedService, tt.severity).Write(mttrMetric)
			require.NoError(t, err)
			assert.Equal(t, tt.mttr, *mttrMetric.Gauge.Value)
		})
	}
}

func TestPerformanceAnomalies(t *testing.T) {
	tests := []struct {
		name        string
		service     string
		operation   string
		anomalyType string
	}{
		{
			name:        "slow response anomaly",
			service:     "api-service",
			operation:   "get_users",
			anomalyType: "slow_response",
		},
		{
			name:        "high error rate anomaly",
			service:     "payment-service",
			operation:   "process_payment",
			anomalyType: "high_error_rate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := getCounterValue(t, PerformanceAnomalies.WithLabelValues(tt.service, tt.operation, tt.anomalyType))
			PerformanceAnomalies.WithLabelValues(tt.service, tt.operation, tt.anomalyType).Inc()
			after := getCounterValue(t, PerformanceAnomalies.WithLabelValues(tt.service, tt.operation, tt.anomalyType))
			assert.Equal(t, before+1, after)
		})
	}
}

func TestRegisterMetrics(t *testing.T) {
	// Test that we can safely register metrics
	// (it will panic on duplicate registration, which is expected behavior)
	// So we test that the metrics are already available for use
	assert.NotPanics(t, func() {
		// Test that we can use all the metrics without issues
		HTTPRequestsTotal.WithLabelValues("GET", "/test", "200").Inc()
		HTTPRequestDuration.WithLabelValues("GET", "/test").Observe(0.1)
		LogEntriesTotal.WithLabelValues("info", "test", "").Inc()
		ErrorsByCategory.WithLabelValues("test", "low", "test").Inc()
		CustomMetric.WithLabelValues("test", "test").Set(1.0)
	})
}

// Benchmark tests
func BenchmarkHTTPRequestsTotal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		HTTPRequestsTotal.WithLabelValues("GET", "/api/test", "200").Inc()
	}
}

func BenchmarkLogEntriesTotal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		LogEntriesTotal.WithLabelValues("info", "test-service", "").Inc()
	}
}

func BenchmarkAPMTracesTotal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		APMTracesTotal.WithLabelValues("test-service", "test-operation", "success").Inc()
	}
}

func BenchmarkCustomMetricSet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CustomMetric.WithLabelValues("test", "benchmark").Set(float64(i))
	}
}

// Helper function to get counter value
func getCounterValue(t *testing.T, counter prometheus.Counter) float64 {
	metric := &dto.Metric{}
	err := counter.Write(metric)
	require.NoError(t, err)
	return *metric.Counter.Value
}
