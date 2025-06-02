package handlers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/nahuelsantos/argus/internal/metrics"
	"github.com/nahuelsantos/argus/internal/services"
)

// BasicHandlers contains basic HTTP handlers
type BasicHandlers struct {
	loggingService *services.LoggingService
	tracingService *services.TracingService
}

// NewBasicHandlers creates a new basic handlers instance
func NewBasicHandlers(loggingService *services.LoggingService, tracingService *services.TracingService) *BasicHandlers {
	return &BasicHandlers{
		loggingService: loggingService,
		tracingService: tracingService,
	}
}

// HealthHandler handles health check requests
func (bh *BasicHandlers) HealthHandler(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"uptime":    time.Since(time.Now().Add(-time.Hour)).String(), // Mock uptime
		"version":   "v2.0.0",                                        // Use consistent version
		"service":   "argus",
		"checks": map[string]string{
			"database":     "ok",
			"redis":        "ok",
			"external_api": "ok",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)

	bh.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Health check performed")
}

// GenerateMetricsHandler generates sample metrics
func (bh *BasicHandlers) GenerateMetricsHandler(w http.ResponseWriter, r *http.Request) {
	count := 10
	if c := r.URL.Query().Get("count"); c != "" {
		if parsed, err := strconv.Atoi(c); err == nil && parsed > 0 {
			count = parsed
		}
	}

	for i := 0; i < count; i++ {
		// Generate random metrics
		metrics.CustomMetric.WithLabelValues("test", "generated").Set(rand.Float64() * 100)

		// Simulate different metric types
		if rand.Float64() > 0.5 {
			metrics.HTTPRequestsTotal.WithLabelValues("GET", "/api/test", "200").Inc()
		} else {
			metrics.HTTPRequestsTotal.WithLabelValues("POST", "/api/test", "201").Inc()
		}
	}

	response := map[string]interface{}{
		"message":           "Metrics generated successfully",
		"metrics_generated": count,
		"timestamp":         time.Now().Format(time.RFC3339),
		"types": []string{
			"custom_business_metric",
			"http_requests_total",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	bh.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Metrics generated",
		zap.Int("count", count))
}

// GenerateLogsHandler generates sample logs
func (bh *BasicHandlers) GenerateLogsHandler(w http.ResponseWriter, r *http.Request) {
	count := 5
	if c := r.URL.Query().Get("count"); c != "" {
		if parsed, err := strconv.Atoi(c); err == nil && parsed > 0 {
			count = parsed
		}
	}

	logTypes := []string{"info", "warn", "error", "debug"}

	for i := 0; i < count; i++ {
		logType := logTypes[rand.Intn(len(logTypes))]

		switch logType {
		case "info":
			bh.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(),
				fmt.Sprintf("Generated info log #%d", i+1))
		case "warn":
			bh.loggingService.LogWithContext(zapcore.WarnLevel, r.Context(),
				fmt.Sprintf("Generated warning log #%d", i+1))
		case "error":
			bh.loggingService.LogError(r.Context(), "test_error", "TEST001",
				fmt.Sprintf("Generated error log #%d", i+1), nil,
				map[string]interface{}{"iteration": i + 1})
		case "debug":
			bh.loggingService.LogWithContext(zapcore.DebugLevel, r.Context(),
				fmt.Sprintf("Generated debug log #%d", i+1))
		}

		// Small delay to spread out timestamps
		time.Sleep(10 * time.Millisecond)
	}

	response := map[string]interface{}{
		"message":        "Logs generated successfully",
		"logs_generated": count,
		"timestamp":      time.Now().Format(time.RFC3339),
		"log_types":      logTypes,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GenerateErrorHandler generates sample errors
func (bh *BasicHandlers) GenerateErrorHandler(w http.ResponseWriter, r *http.Request) {
	errorTypes := []string{"validation", "database", "network", "timeout", "auth"}
	errorType := errorTypes[rand.Intn(len(errorTypes))]

	errorCode := fmt.Sprintf("ERR_%s_%03d", errorType, rand.Intn(999)+1)
	errorMessage := fmt.Sprintf("Simulated %s error for testing", errorType)

	bh.loggingService.LogError(r.Context(), errorType, errorCode, errorMessage,
		fmt.Errorf("simulated error"), map[string]interface{}{
			"severity": "medium",
			"category": "testing",
		})

	// Simulate different HTTP error codes
	statusCode := 500
	switch errorType {
	case "validation":
		statusCode = 400
	case "auth":
		statusCode = 401
	case "timeout":
		statusCode = 408
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"error":      true,
		"type":       errorType,
		"code":       errorCode,
		"message":    errorMessage,
		"timestamp":  time.Now().Format(time.RFC3339),
		"request_id": r.Header.Get("X-Request-ID"),
	}

	json.NewEncoder(w).Encode(response)
}

// CPULoadHandler simulates CPU load
func (bh *BasicHandlers) CPULoadHandler(w http.ResponseWriter, r *http.Request) {
	duration := 5 * time.Second
	if d := r.URL.Query().Get("duration"); d != "" {
		if parsed, err := time.ParseDuration(d); err == nil {
			duration = parsed
		}
	}

	intensity := 50
	if i := r.URL.Query().Get("intensity"); i != "" {
		if parsed, err := strconv.Atoi(i); err == nil && parsed > 0 && parsed <= 100 {
			intensity = parsed
		}
	}

	start := time.Now()
	end := start.Add(duration)

	// Simulate CPU load
	go func() {
		for time.Now().Before(end) {
			if rand.Intn(100) < intensity {
				// Busy work
				for i := 0; i < 1000000; i++ {
					_ = i * i
				}
			}
			time.Sleep(time.Millisecond)
		}
	}()

	response := map[string]interface{}{
		"message":   "CPU load simulation started",
		"duration":  duration.String(),
		"intensity": intensity,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	bh.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "CPU load simulation started",
		zap.Duration("duration", duration), zap.Int("intensity", intensity))
}

// MemoryLoadHandler simulates memory load
func (bh *BasicHandlers) MemoryLoadHandler(w http.ResponseWriter, r *http.Request) {
	sizeMB := 100
	if s := r.URL.Query().Get("size"); s != "" {
		if parsed, err := strconv.Atoi(s); err == nil && parsed > 0 {
			sizeMB = parsed
		}
	}

	duration := 30 * time.Second
	if d := r.URL.Query().Get("duration"); d != "" {
		if parsed, err := time.ParseDuration(d); err == nil {
			duration = parsed
		}
	}

	// Allocate memory
	data := make([][]byte, sizeMB)
	for i := range data {
		data[i] = make([]byte, 1024*1024) // 1MB chunks
		// Fill with random data to prevent optimization
		for j := range data[i] {
			data[i][j] = byte(rand.Intn(256))
		}
	}

	// Hold memory for specified duration
	go func() {
		time.Sleep(duration)
		// Release memory
		data = nil
		runtime.GC()
	}()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	response := map[string]interface{}{
		"message":       "Memory load simulation started",
		"allocated_mb":  sizeMB,
		"duration":      duration.String(),
		"current_alloc": m.Alloc,
		"total_alloc":   m.TotalAlloc,
		"sys":           m.Sys,
		"num_gc":        m.NumGC,
		"timestamp":     time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	bh.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Memory load simulation started",
		zap.Int("size_mb", sizeMB), zap.Duration("duration", duration))
}
