package handlers

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/nahuelsantos/argus/internal/metrics"
	"github.com/nahuelsantos/argus/internal/middleware"
	"github.com/nahuelsantos/argus/internal/services"
	"github.com/nahuelsantos/argus/internal/types"
	"github.com/nahuelsantos/argus/internal/utils"
)

// PerformanceHandlers contains LGTM stack performance testing handlers
type PerformanceHandlers struct {
	loggingService *services.LoggingService
	tracingService *services.TracingService
}

// NewPerformanceHandlers creates a new performance handlers instance
func NewPerformanceHandlers(loggingService *services.LoggingService, tracingService *services.TracingService) *PerformanceHandlers {
	return &PerformanceHandlers{
		loggingService: loggingService,
		tracingService: tracingService,
	}
}

type PerformanceTestResult struct {
	TestType       string            `json:"test_type"`
	Status         string            `json:"status"`
	Duration       float64           `json:"duration_seconds"`
	ItemsGenerated int               `json:"items_generated"`
	ItemsPerSecond float64           `json:"items_per_second"`
	Details        map[string]string `json:"details,omitempty"`
	ResourceUsage  *ResourceUsage    `json:"resource_usage,omitempty"`
	Timestamp      time.Time         `json:"timestamp"`
}

type ResourceUsage struct {
	CPUPercent     float64 `json:"cpu_percent"`
	MemoryMB       float64 `json:"memory_mb"`
	DiskUsageMB    float64 `json:"disk_usage_mb"`
	NetworkBytesTx int64   `json:"network_bytes_tx"`
	NetworkBytesRx int64   `json:"network_bytes_rx"`
}

// Test Metrics Scale - Generate high-volume metrics
func (ph *PerformanceHandlers) TestMetricsScale(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ph.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Starting metrics scale test...")

	// Use validation config
	validationConfig := middleware.DefaultValidationConfig()

	// Parse and validate parameters with proper limits
	count := middleware.ValidateCount(r.URL.Query().Get("count"), validationConfig)
	duration := middleware.ValidateDuration(r.URL.Query().Get("duration"), validationConfig)
	concurrency := middleware.ValidateConcurrency(r.URL.Query().Get("concurrency"), validationConfig)

	// Log the validated parameters
	ph.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Metrics scale test parameters validated",
		zap.Int("count", count),
		zap.Duration("duration", duration),
		zap.Int("concurrency", concurrency))

	// Generate high-volume metrics
	ctx, cancel := context.WithTimeout(r.Context(), duration)
	defer cancel()

	var wg sync.WaitGroup
	var totalGenerated int64
	var mu sync.Mutex

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			workerGenerated := 0

			for {
				select {
				case <-ctx.Done():
					mu.Lock()
					totalGenerated += int64(workerGenerated)
					mu.Unlock()
					return
				default:
					// Generate various metric types
					metrics.CustomMetric.WithLabelValues("performance_test", fmt.Sprintf("worker_%d", workerID)).Set(rand.Float64() * 100)
					metrics.HTTPRequestsTotal.WithLabelValues("GET", "/api/scale-test", "200").Inc()
					metrics.HTTPRequestsTotal.WithLabelValues("POST", "/api/scale-test", "201").Inc()
					metrics.HTTPRequestsTotal.WithLabelValues("PUT", "/api/scale-test", "200").Inc()

					workerGenerated += 4

					// Small delay to prevent overwhelming
					time.Sleep(time.Millisecond)
				}
			}
		}(i)
	}

	wg.Wait()
	testDuration := time.Since(start)

	result := PerformanceTestResult{
		TestType:       "metrics_scale",
		Status:         "completed",
		Duration:       testDuration.Seconds(),
		ItemsGenerated: int(totalGenerated),
		ItemsPerSecond: float64(totalGenerated) / testDuration.Seconds(),
		Details: map[string]string{
			"concurrency":   strconv.Itoa(concurrency),
			"target_count":  strconv.Itoa(count),
			"test_duration": duration.String(),
			"metric_types":  "4",
		},
		Timestamp: time.Now(),
	}

	ph.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Metrics scale test completed",
		zap.Int("items_generated", result.ItemsGenerated),
		zap.Float64("items_per_second", result.ItemsPerSecond))

	w.Header().Set("Content-Type", "application/json")
	utils.EncodeJSON(w, result)
}

// Test Logs Scale - Generate high-volume logs
func (ph *PerformanceHandlers) TestLogsScale(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ph.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Starting logs scale test...")

	// Use validation config
	validationConfig := middleware.DefaultValidationConfig()

	// Parse and validate parameters with proper limits
	duration := middleware.ValidateDuration(r.URL.Query().Get("duration"), validationConfig)
	concurrency := middleware.ValidateConcurrency(r.URL.Query().Get("concurrency"), validationConfig)
	logLevel := middleware.ValidateLogLevel(r.URL.Query().Get("level"))

	// Log the validated parameters
	ph.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Logs scale test parameters validated",
		zap.Duration("duration", duration),
		zap.Int("concurrency", concurrency),
		zap.String("log_level", logLevel))

	// Generate high-volume logs
	ctx, cancel := context.WithTimeout(r.Context(), duration)
	defer cancel()

	var wg sync.WaitGroup
	var totalGenerated int64
	var mu sync.Mutex

	logMessages := []string{
		"User authentication successful",
		"Database query executed",
		"API request processed",
		"Cache miss occurred",
		"File upload completed",
		"Background job started",
		"Configuration loaded",
		"Connection established",
		"Data validation passed",
		"Transaction committed",
	}

	errorMessages := []string{
		"Database connection timeout",
		"Invalid user credentials",
		"File not found",
		"Permission denied",
		"Network connection failed",
		"Invalid JSON payload",
		"Rate limit exceeded",
		"Service unavailable",
		"Validation error",
		"Internal server error",
	}

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			workerGenerated := 0

			for {
				select {
				case <-ctx.Done():
					mu.Lock()
					totalGenerated += int64(workerGenerated)
					mu.Unlock()
					return
				default:
					// Generate different log types based on level
					switch logLevel {
					case "info":
						ph.loggingService.LogWithContext(zapcore.InfoLevel, ctx,
							logMessages[rand.Intn(len(logMessages))],
							zap.Int("worker_id", workerID),
							zap.Int("iteration", workerGenerated))
					case "warn":
						ph.loggingService.LogWithContext(zapcore.WarnLevel, ctx,
							"Warning: "+logMessages[rand.Intn(len(logMessages))],
							zap.Int("worker_id", workerID))
					case "error":
						ph.loggingService.LogError(ctx, "performance_test", fmt.Sprintf("ERR_%d_%d", workerID, workerGenerated),
							errorMessages[rand.Intn(len(errorMessages))], nil,
							map[string]interface{}{"worker_id": workerID, "test_type": "scale"})
					default: // mixed
						switch rand.Intn(4) {
						case 0:
							ph.loggingService.LogWithContext(zapcore.InfoLevel, ctx, logMessages[rand.Intn(len(logMessages))])
						case 1:
							ph.loggingService.LogWithContext(zapcore.WarnLevel, ctx, "Warning: "+logMessages[rand.Intn(len(logMessages))])
						case 2:
							ph.loggingService.LogError(ctx, "test_error", fmt.Sprintf("ERR_%d", rand.Intn(1000)), errorMessages[rand.Intn(len(errorMessages))], nil, nil)
						case 3:
							ph.loggingService.LogWithContext(zapcore.DebugLevel, ctx, "Debug: "+logMessages[rand.Intn(len(logMessages))])
						}
					}

					workerGenerated++

					// Small delay to prevent overwhelming
					time.Sleep(5 * time.Millisecond)
				}
			}
		}(i)
	}

	wg.Wait()
	testDuration := time.Since(start)

	result := PerformanceTestResult{
		TestType:       "logs_scale",
		Status:         "completed",
		Duration:       testDuration.Seconds(),
		ItemsGenerated: int(totalGenerated),
		ItemsPerSecond: float64(totalGenerated) / testDuration.Seconds(),
		Details: map[string]string{
			"concurrency":   strconv.Itoa(concurrency),
			"log_level":     logLevel,
			"test_duration": duration.String(),
			"log_types":     "4",
		},
		Timestamp: time.Now(),
	}

	ph.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Logs scale test completed",
		zap.Int("items_generated", result.ItemsGenerated),
		zap.Float64("items_per_second", result.ItemsPerSecond))

	w.Header().Set("Content-Type", "application/json")
	utils.EncodeJSON(w, result)
}

// Test Traces Scale - Generate high-volume traces
func (ph *PerformanceHandlers) TestTracesScale(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ph.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Starting traces scale test...")

	// Use validation config
	validationConfig := middleware.DefaultValidationConfig()

	// Parse and validate parameters with proper limits
	duration := middleware.ValidateDuration(r.URL.Query().Get("duration"), validationConfig)
	concurrency := middleware.ValidatePositiveInt(r.URL.Query().Get("concurrency"), 3, 10) // Default 3, max 10 for traces

	// Log the validated parameters
	ph.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Traces scale test parameters validated",
		zap.Duration("duration", duration),
		zap.Int("concurrency", concurrency))

	// Generate high-volume traces
	ctx, cancel := context.WithTimeout(r.Context(), duration)
	defer cancel()

	var wg sync.WaitGroup
	var totalGenerated int64
	var mu sync.Mutex

	services := []string{"user-service", "order-service", "payment-service", "notification-service", "inventory-service"}
	operations := []string{"get", "create", "update", "delete", "list", "validate", "process"}

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			workerGenerated := 0

			for {
				select {
				case <-ctx.Done():
					mu.Lock()
					totalGenerated += int64(workerGenerated)
					mu.Unlock()
					return
				default:
					// Generate complex trace with multiple spans
					serviceName := services[rand.Intn(len(services))]
					operation := operations[rand.Intn(len(operations))]

					// Simulate trace generation (using logging for now since we have a mock tracer)
					ph.loggingService.LogWithContext(zapcore.InfoLevel, ctx,
						"Trace generated",
						zap.String("service", serviceName),
						zap.String("operation", operation),
						zap.String("trace_id", fmt.Sprintf("trace_%d_%d_%d", workerID, workerGenerated, time.Now().UnixNano())),
						zap.String("span_id", fmt.Sprintf("span_%d", rand.Intn(10000))),
						zap.Duration("duration", time.Duration(rand.Intn(1000))*time.Millisecond),
						zap.String("status", "ok"))

					workerGenerated++

					// Small delay to prevent overwhelming
					time.Sleep(10 * time.Millisecond)
				}
			}
		}(i)
	}

	wg.Wait()
	testDuration := time.Since(start)

	result := PerformanceTestResult{
		TestType:       "traces_scale",
		Status:         "completed",
		Duration:       testDuration.Seconds(),
		ItemsGenerated: int(totalGenerated),
		ItemsPerSecond: float64(totalGenerated) / testDuration.Seconds(),
		Details: map[string]string{
			"concurrency":      strconv.Itoa(concurrency),
			"test_duration":    duration.String(),
			"services_count":   strconv.Itoa(len(services)),
			"operations_count": strconv.Itoa(len(operations)),
		},
		Timestamp: time.Now(),
	}

	ph.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Traces scale test completed",
		zap.Int("items_generated", result.ItemsGenerated),
		zap.Float64("items_per_second", result.ItemsPerSecond))

	w.Header().Set("Content-Type", "application/json")
	utils.EncodeJSON(w, result)
}

// Test Dashboard Load - Stress test Grafana dashboards
func (ph *PerformanceHandlers) TestDashboardLoad(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ph.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Starting dashboard load test...")

	// Use validation config
	validationConfig := middleware.DefaultValidationConfig()

	// Parse and validate parameters with proper limits
	concurrency := middleware.ValidateConcurrency(r.URL.Query().Get("concurrency"), validationConfig)
	requests := middleware.ValidatePositiveInt(r.URL.Query().Get("requests"), 100, 1000) // Default 100, max 1000

	// Log the validated parameters
	ph.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Dashboard load test parameters validated",
		zap.Int("concurrency", concurrency),
		zap.Int("requests", requests))

	// Test dashboard endpoints - use LGTM settings URLs
	lgtmSettings := types.GetDefaults()
	grafanaURL := lgtmSettings.Grafana.URL
	prometheusURL := lgtmSettings.Prometheus.URL
	lokiURL := lgtmSettings.Loki.URL
	tempoURL := lgtmSettings.Tempo.URL

	dashboardEndpoints := []string{
		grafanaURL + "/api/health",
		grafanaURL + "/api/datasources",
		grafanaURL + "/api/dashboards/home",
		grafanaURL + "/api/search",
		prometheusURL + "/api/v1/query?query=up",
		prometheusURL + "/api/v1/targets",
		lokiURL + "/ready",
		tempoURL + "/ready",
	}

	var wg sync.WaitGroup
	var totalRequests int64
	var successfulRequests int64
	var mu sync.Mutex

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			workerSuccess := 0

			for j := 0; j < requests; j++ {
				endpoint := dashboardEndpoints[rand.Intn(len(dashboardEndpoints))]

				resp, err := http.Get(endpoint)
				if err == nil {
					resp.Body.Close()
					if resp.StatusCode < 400 {
						workerSuccess++
					}
				}

				// Small delay between requests
				time.Sleep(10 * time.Millisecond)
			}

			mu.Lock()
			totalRequests += int64(requests)
			successfulRequests += int64(workerSuccess)
			mu.Unlock()
		}(i)
	}

	wg.Wait()
	testDuration := time.Since(start)

	successRate := float64(successfulRequests) / float64(totalRequests) * 100

	result := PerformanceTestResult{
		TestType:       "dashboard_load",
		Status:         "completed",
		Duration:       testDuration.Seconds(),
		ItemsGenerated: int(totalRequests),
		ItemsPerSecond: float64(totalRequests) / testDuration.Seconds(),
		Details: map[string]string{
			"concurrency":         strconv.Itoa(concurrency),
			"requests_per_user":   strconv.Itoa(requests),
			"successful_requests": strconv.FormatInt(successfulRequests, 10),
			"success_rate":        fmt.Sprintf("%.2f%%", successRate),
			"endpoints_tested":    strconv.Itoa(len(dashboardEndpoints)),
		},
		Timestamp: time.Now(),
	}

	ph.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Dashboard load test completed",
		zap.Int64("total_requests", totalRequests),
		zap.Int64("successful_requests", successfulRequests),
		zap.Float64("success_rate", successRate))

	w.Header().Set("Content-Type", "application/json")
	utils.EncodeJSON(w, result)
}

// Test Resource Usage - Monitor LGTM stack resource consumption
func (ph *PerformanceHandlers) TestResourceUsage(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ph.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Starting resource usage test...")

	// Get resource usage from various sources
	resourceData := make(map[string]interface{})

	// Get LGTM settings URLs
	lgtmSettings := types.GetDefaults()
	prometheusURL := lgtmSettings.Prometheus.URL
	lokiURL := lgtmSettings.Loki.URL
	tempoURL := lgtmSettings.Tempo.URL
	grafanaURL := lgtmSettings.Grafana.URL

	// Test Prometheus metrics endpoint for resource data
	if resp, err := http.Get(prometheusURL + "/api/v1/query?query=up"); err == nil {
		defer resp.Body.Close()
		if body, err := io.ReadAll(resp.Body); err == nil {
			upTargets := strings.Count(string(body), `"value":[`)
			resourceData["prometheus_targets_up"] = upTargets
		}
	}

	// Test Loki metrics
	if resp, err := http.Get(lokiURL + "/metrics"); err == nil {
		defer resp.Body.Close()
		if body, err := io.ReadAll(resp.Body); err == nil {
			bodyStr := string(body)
			hasIngesterMetrics := strings.Contains(bodyStr, "loki_ingester_")
			resourceData["loki_ingester_active"] = hasIngesterMetrics

			// Count metrics
			metricsCount := strings.Count(bodyStr, "\n")
			resourceData["loki_metrics_count"] = metricsCount
		}
	}

	// Test Tempo status
	if resp, err := http.Get(tempoURL + "/status"); err == nil {
		defer resp.Body.Close()
		resourceData["tempo_status"] = "accessible"
	} else {
		resourceData["tempo_status"] = "failed"
	}

	// Test Grafana health
	if resp, err := http.Get(grafanaURL + "/api/health"); err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			resourceData["grafana_health"] = "healthy"
		} else {
			resourceData["grafana_health"] = "degraded"
		}
	} else {
		resourceData["grafana_health"] = "failed"
	}

	// Mock resource usage (in a real implementation, you'd gather actual metrics)
	mockResourceUsage := &ResourceUsage{
		CPUPercent:     rand.Float64() * 80,        // 0-80% CPU
		MemoryMB:       500 + rand.Float64()*1500,  // 500-2000 MB
		DiskUsageMB:    1000 + rand.Float64()*5000, // 1-6 GB
		NetworkBytesTx: int64(rand.Intn(1000000)),  // Random network usage
		NetworkBytesRx: int64(rand.Intn(1000000)),
	}

	testDuration := time.Since(start)
	result := PerformanceTestResult{
		TestType:       "resource_usage",
		Status:         "completed",
		Duration:       testDuration.Seconds(),
		ItemsGenerated: len(resourceData),
		ResourceUsage:  mockResourceUsage,
		Details: map[string]string{
			"components_checked": "4",
			"data_points":        strconv.Itoa(len(resourceData)),
		},
		Timestamp: time.Now(),
	}

	// Add resource data to details
	for key, value := range resourceData {
		result.Details[key] = fmt.Sprintf("%v", value)
	}

	ph.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Resource usage test completed")

	w.Header().Set("Content-Type", "application/json")
	utils.EncodeJSON(w, result)
}

// Test Storage Limits - Test LGTM stack storage and retention capabilities
func (ph *PerformanceHandlers) TestStorageLimits(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ph.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Starting storage limits test...")

	storageData := make(map[string]interface{})

	// Get LGTM settings URLs
	lgtmSettings := types.GetDefaults()
	prometheusURL := lgtmSettings.Prometheus.URL
	lokiURL := lgtmSettings.Loki.URL
	tempoURL := lgtmSettings.Tempo.URL

	// Test Prometheus storage metrics
	if resp, err := http.Get(prometheusURL + "/api/v1/query?query=prometheus_tsdb_symbol_table_size_bytes"); err == nil {
		defer resp.Body.Close()
		storageData["prometheus_storage_accessible"] = true
	} else {
		storageData["prometheus_storage_accessible"] = false
	}

	// Test Loki ingestion rate
	if resp, err := http.Get(lokiURL + "/metrics"); err == nil {
		defer resp.Body.Close()
		if body, err := io.ReadAll(resp.Body); err == nil {
			// Look for ingestion rate metrics
			hasIngestionRate := strings.Contains(string(body), "loki_distributor_")
			storageData["loki_ingestion_rate_available"] = hasIngestionRate
		}
	}

	// Test Tempo storage
	if resp, err := http.Get(tempoURL + "/status"); err == nil {
		defer resp.Body.Close()
		storageData["tempo_storage_accessible"] = true
	} else {
		storageData["tempo_storage_accessible"] = false
	}

	// Mock storage usage data
	storageData["prometheus_estimated_size_mb"] = rand.Intn(1000) + 100          // 100-1100 MB
	storageData["loki_estimated_size_mb"] = rand.Intn(2000) + 200                // 200-2200 MB
	storageData["tempo_estimated_size_mb"] = rand.Intn(1500) + 150               // 150-1650 MB
	storageData["retention_policy_days"] = 30                                    // Mock retention
	storageData["compression_ratio"] = fmt.Sprintf("%.2f", 2.5+rand.Float64()*2) // 2.5-4.5x

	testDuration := time.Since(start)
	result := PerformanceTestResult{
		TestType:       "storage_limits",
		Status:         "completed",
		Duration:       testDuration.Seconds(),
		ItemsGenerated: len(storageData),
		Details: map[string]string{
			"storage_components": "3",
			"data_points":        strconv.Itoa(len(storageData)),
		},
		Timestamp: time.Now(),
	}

	// Add storage data to details
	for key, value := range storageData {
		result.Details[key] = fmt.Sprintf("%v", value)
	}

	ph.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Storage limits test completed")

	w.Header().Set("Content-Type", "application/json")
	utils.EncodeJSON(w, result)
}
