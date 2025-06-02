package models

import (
	"time"
)

// LogContext represents the context for log correlation
type LogContext struct {
	RequestID   string `json:"request_id"`
	TraceID     string `json:"trace_id"`
	SpanID      string `json:"span_id"`
	UserID      string `json:"user_id,omitempty"`
	SessionID   string `json:"session_id,omitempty"`
	ServiceName string `json:"service_name"`
	Version     string `json:"version"`
	Environment string `json:"environment"`
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Level       string                 `json:"level"`
	Timestamp   time.Time              `json:"timestamp"`
	Message     string                 `json:"message"`
	Context     LogContext             `json:"context"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Error       *LogErrorData          `json:"error,omitempty"`
	Performance *PerformanceData       `json:"performance,omitempty"`
	Business    *BusinessData          `json:"business,omitempty"`
}

// LogErrorData represents error information in logs
type LogErrorData struct {
	Type       string `json:"type"`
	Code       string `json:"code"`
	Message    string `json:"message"`
	StackTrace string `json:"stack_trace,omitempty"`
}

// PerformanceData represents performance metrics in logs
type PerformanceData struct {
	Duration       float64 `json:"duration_ms"`
	MemoryUsage    int64   `json:"memory_usage_bytes"`
	GoroutineCount int     `json:"goroutine_count"`
	CPUPercent     float64 `json:"cpu_percent,omitempty"`
}

// BusinessData represents business event data in logs
type BusinessData struct {
	EventType  string                 `json:"event_type"`
	EntityID   string                 `json:"entity_id,omitempty"`
	EntityType string                 `json:"entity_type,omitempty"`
	Action     string                 `json:"action"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Metrics    map[string]float64     `json:"metrics,omitempty"`
}

// Context keys for request correlation
type ContextKey string

const (
	RequestIDKey ContextKey = "request_id"
	TraceIDKey   ContextKey = "trace_id"
	UserIDKey    ContextKey = "user_id"
	SessionIDKey ContextKey = "session_id"
	StartTimeKey ContextKey = "start_time"
)
