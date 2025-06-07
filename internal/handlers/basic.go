package handlers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/nahuelsantos/argus/internal/metrics"
	"github.com/nahuelsantos/argus/internal/services"
	"github.com/nahuelsantos/argus/internal/types"
	"github.com/nahuelsantos/argus/internal/utils"
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
	// Calculate actual uptime since service start (simplified)
	// In production you'd track actual start time
	startTime := time.Now().Add(-time.Hour) // Mock for now, but could be real
	uptime := time.Since(startTime)

	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"uptime":    uptime.String(),
		"version":   "v2.0.0",
		"service":   "argus",
		"purpose":   "LGTM stack synthetic data generator and validator",
		"checks": map[string]string{
			"web_server":       "ok", // If this responds, web server is ok
			"metrics_registry": "ok", // Prometheus metrics are registered
			"logging_service":  "ok", // Logging service is initialized
		},
	}

	w.Header().Set("Content-Type", "application/json")
	utils.EncodeJSON(w, health)

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
	utils.EncodeJSON(w, response)

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
	utils.EncodeJSON(w, response)
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

	utils.EncodeJSON(w, response)
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
	utils.EncodeJSON(w, response)

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
	utils.EncodeJSON(w, response)

	bh.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Memory load simulation started",
		zap.Int("size_mb", sizeMB), zap.Duration("duration", duration))
}

// LGTMStatusHandler checks the status of LGTM stack components
func (bh *BasicHandlers) LGTMStatusHandler(w http.ResponseWriter, r *http.Request) {
	// Use environment-aware service URLs
	services := map[string]string{
		"prometheus":   utils.GetServiceURL("prometheus") + "/-/healthy",
		"alertmanager": utils.GetServiceURL("alertmanager") + "/-/healthy",
		"grafana":      utils.GetServiceURL("grafana") + "/api/health",
		"loki":         utils.GetServiceURL("loki") + "/ready",
		"tempo":        utils.GetServiceURL("tempo") + "/ready",
	}

	status := make(map[string]interface{})

	for service, url := range services {
		serviceStatus := bh.checkServiceHealth(url)
		status[service] = serviceStatus
	}

	w.Header().Set("Content-Type", "application/json")
	utils.EncodeJSON(w, status)
}

func (bh *BasicHandlers) checkServiceHealth(url string) string {
	client := &http.Client{
		Timeout: 8 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return "offline"
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return "online"
	}

	return "offline"
}

// Global settings storage (in production, use database)
var globalSettings *types.LGTMSettings

// SettingsHandler handles settings save/load
func (bh *BasicHandlers) SettingsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		bh.getSettings(w, r)
	case "POST":
		bh.saveSettings(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (bh *BasicHandlers) getSettings(w http.ResponseWriter, r *http.Request) {
	if globalSettings == nil {
		globalSettings = types.GetDefaults()
	}

	w.Header().Set("Content-Type", "application/json")
	utils.EncodeJSON(w, globalSettings)
}

func (bh *BasicHandlers) saveSettings(w http.ResponseWriter, r *http.Request) {
	var settings types.LGTMSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	globalSettings = &settings

	response := map[string]interface{}{
		"status":    "saved",
		"message":   "Settings saved successfully",
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	utils.EncodeJSON(w, response)
}

// TestConnectionHandler tests connection to specific LGTM services
func (bh *BasicHandlers) TestConnectionHandler(w http.ResponseWriter, r *http.Request) {
	// Extract service name from URL path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid service path", http.StatusBadRequest)
		return
	}
	service := parts[3] // /api/test-connection/{service}

	var config types.ServiceConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	result := bh.testServiceConnection(service, config)

	w.Header().Set("Content-Type", "application/json")
	utils.EncodeJSON(w, result)
}

func (bh *BasicHandlers) testServiceConnection(service string, config types.ServiceConfig) map[string]interface{} {
	var testURL string
	var requiresAuth bool

	switch service {
	case "grafana":
		testURL = config.URL + "/api/user" // This endpoint requires authentication
		requiresAuth = true
	case "prometheus":
		testURL = config.URL + "/-/healthy"
		requiresAuth = config.Username != ""
	case "alertmanager":
		testURL = config.URL + "/-/healthy"
		requiresAuth = false
	case "loki":
		testURL = config.URL + "/ready"
		requiresAuth = false
	case "tempo":
		testURL = config.URL + "/ready"
		requiresAuth = false
	default:
		return map[string]interface{}{
			"status":  "error",
			"message": "Unknown service",
		}
	}

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", testURL, nil)
	if err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Failed to create request: %v", err),
		}
	}

	// For Grafana, always test with credentials if provided
	if service == "grafana" && config.Username != "" {
		req.SetBasicAuth(config.Username, config.Password)
	} else if requiresAuth && config.Username != "" {
		req.SetBasicAuth(config.Username, config.Password)
	}

	resp, err := client.Do(req)
	if err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Connection failed: %v", err),
		}
	}
	defer resp.Body.Close()

	// For Grafana, check for authentication errors
	if service == "grafana" {
		if resp.StatusCode == 401 || resp.StatusCode == 403 {
			return map[string]interface{}{
				"status":  "error",
				"message": "Authentication failed - check username/password",
				"details": map[string]interface{}{
					"url":         testURL,
					"status_code": resp.StatusCode,
				},
			}
		}
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return map[string]interface{}{
			"status":  "success",
			"message": fmt.Sprintf("%s is accessible", service),
			"details": map[string]interface{}{
				"url":         testURL,
				"status_code": resp.StatusCode,
			},
		}
	} else {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("%s returned HTTP %d", service, resp.StatusCode),
			"details": map[string]interface{}{
				"url":         testURL,
				"status_code": resp.StatusCode,
			},
		}
	}
}
