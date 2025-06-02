package models

import (
	"time"
)

// ServiceDependency represents a service dependency with performance metrics
type ServiceDependency struct {
	ServiceName      string            `json:"service_name"`
	Operation        string            `json:"operation"`
	ResponseTime     time.Duration     `json:"response_time"`
	StatusCode       int               `json:"status_code"`
	ErrorRate        float64           `json:"error_rate"`
	RequestCount     int64             `json:"request_count"`
	Dependencies     []string          `json:"dependencies"`
	CustomAttributes map[string]string `json:"custom_attributes"`
}

// APMData represents Application Performance Monitoring data
type APMData struct {
	ServiceName   string              `json:"service_name"`
	TraceID       string              `json:"trace_id"`
	SpanID        string              `json:"span_id"`
	OperationName string              `json:"operation_name"`
	StartTime     time.Time           `json:"start_time"`
	Duration      time.Duration       `json:"duration"`
	StatusCode    int                 `json:"status_code"`
	ResourceUsage ResourceMetrics     `json:"resource_usage"`
	Dependencies  []ServiceDependency `json:"dependencies"`
	ErrorDetails  *APMError           `json:"error_details,omitempty"`
	CustomTags    map[string]string   `json:"custom_tags"`
}

// APMError represents error details in APM data
type APMError struct {
	Type        string `json:"type"`
	Message     string `json:"message"`
	StackTrace  string `json:"stack_trace"`
	Impact      string `json:"impact"` // "low", "medium", "high", "critical"
	Recoverable bool   `json:"recoverable"`
}

// ResourceMetrics represents system resource usage metrics
type ResourceMetrics struct {
	CPUUsage       float64 `json:"cpu_usage_percent"`
	MemoryUsage    int64   `json:"memory_usage_bytes"`
	GoroutineCount int     `json:"goroutine_count"`
	HeapSize       int64   `json:"heap_size_bytes"`
	GCPause        float64 `json:"gc_pause_ms"`
	DiskIO         int64   `json:"disk_io_bytes"`
	NetworkIO      int64   `json:"network_io_bytes"`
}

// PerformanceProfile represents performance analysis data
type PerformanceProfile struct {
	Operation       string          `json:"operation"`
	P50ResponseTime float64         `json:"p50_response_time_ms"`
	P95ResponseTime float64         `json:"p95_response_time_ms"`
	P99ResponseTime float64         `json:"p99_response_time_ms"`
	ErrorRate       float64         `json:"error_rate_percent"`
	ThroughputRPS   float64         `json:"throughput_rps"`
	ResourceProfile ResourceMetrics `json:"resource_profile"`
	Bottlenecks     []string        `json:"bottlenecks"`
	Recommendations []string        `json:"recommendations"`
}
