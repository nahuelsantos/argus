package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// HTTP metrics
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// Log-based metrics
	LogEntriesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "log_entries_total",
			Help: "Total number of log entries by level and service",
		},
		[]string{"level", "service", "error_type"},
	)

	LogProcessingDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "log_processing_duration_seconds",
			Help:    "Time spent processing log entries",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
		},
		[]string{"operation", "log_level"},
	)

	ErrorsByCategory = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "errors_by_category_total",
			Help: "Total errors categorized by type and severity",
		},
		[]string{"category", "severity", "source"},
	)

	CustomMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "custom_business_metric",
			Help: "Custom business metric for testing",
		},
		[]string{"type", "category"},
	)

	// APM metrics
	APMTracesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "apm_traces_total",
			Help: "Total number of APM traces by service and operation",
		},
		[]string{"service", "operation", "status"},
	)

	APMSpanDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "apm_span_duration_seconds",
			Help:    "APM span duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
		},
		[]string{"service", "operation"},
	)

	ServiceDependencyLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "service_dependency_latency_seconds",
			Help:    "Service dependency call latency",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
		},
		[]string{"source_service", "target_service", "operation"},
	)

	PerformanceAnomalies = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "performance_anomalies_total",
			Help: "Total number of detected performance anomalies",
		},
		[]string{"service", "operation", "anomaly_type"},
	)

	// Alerting metrics
	AlertsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "alerts_total",
			Help: "Total number of alerts by rule name, severity, and status",
		},
		[]string{"rule_name", "severity", "status"},
	)

	AlertDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "alert_duration_seconds",
			Help:    "Duration of alerts by rule name and severity",
			Buckets: []float64{1, 5, 10, 30, 60, 300, 600, 1800, 3600, 7200},
		},
		[]string{"rule_name", "severity"},
	)

	IncidentsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "incidents_total",
			Help: "Total number of incidents by severity, status, and affected service",
		},
		[]string{"severity", "status", "affected_service"},
	)

	IncidentDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "incident_duration_seconds",
			Help:    "Duration of incidents by severity and service",
			Buckets: []float64{60, 300, 600, 1800, 3600, 7200, 14400, 28800, 86400},
		},
		[]string{"severity", "affected_service"},
	)

	NotificationsSent = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "notifications_sent_total",
			Help: "Total number of notifications sent by channel type, severity, and status",
		},
		[]string{"channel_type", "severity", "status"},
	)

	NotificationLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "notification_latency_seconds",
			Help:    "Latency of notification delivery by channel type",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.0, 5.0},
		},
		[]string{"channel_type"},
	)

	AlertManagerHealth = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "alert_manager_health",
			Help: "Health status of alert manager components",
		},
		[]string{"component"},
	)

	MTTRGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mttr_seconds",
			Help: "Mean Time To Resolution by service and severity",
		},
		[]string{"service", "severity"},
	)
)

// RegisterMetrics registers all Prometheus metrics
func RegisterMetrics() {
	prometheus.MustRegister(
		HTTPRequestsTotal,
		HTTPRequestDuration,
		LogEntriesTotal,
		LogProcessingDuration,
		ErrorsByCategory,
		CustomMetric,
		APMTracesTotal,
		APMSpanDuration,
		ServiceDependencyLatency,
		PerformanceAnomalies,
		AlertsTotal,
		AlertDuration,
		IncidentsTotal,
		IncidentDuration,
		NotificationsSent,
		NotificationLatency,
		AlertManagerHealth,
		MTTRGauge,
	)
}
