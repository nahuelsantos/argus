package handlers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/nahuelsantos/argus/internal/services"
)

// SimulationHandlers handles service simulation endpoints
type SimulationHandlers struct {
	loggingService *services.LoggingService
	tracingService *services.TracingService
}

// NewSimulationHandlers creates a new SimulationHandlers instance
func NewSimulationHandlers(loggingService *services.LoggingService, tracingService *services.TracingService) *SimulationHandlers {
	return &SimulationHandlers{
		loggingService: loggingService,
		tracingService: tracingService,
	}
}

// SimulateWebServiceHandler simulates a typical web service (WordPress, web apps)
func (h *SimulationHandlers) SimulateWebServiceHandler(w http.ResponseWriter, r *http.Request) {
	// Simulate web service characteristics
	pageViews := rand.Intn(50) + 10
	avgResponseTime := rand.Intn(200) + 50 // 50-250ms
	errorRate := rand.Float64() * 0.05     // 0-5% error rate

	// Generate web-specific logs
	h.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Web service simulation started",
		zap.String("service_type", "web-service"),
		zap.Int("page_views", pageViews),
		zap.Int("avg_response_time_ms", avgResponseTime),
		zap.Float64("error_rate", errorRate))

	// Simulate different web endpoints
	endpoints := []string{"/", "/about", "/contact", "/blog", "/products", "/login", "/dashboard"}
	for i := 0; i < pageViews; i++ {
		endpoint := endpoints[rand.Intn(len(endpoints))]
		responseTime := time.Duration(rand.Intn(300)+50) * time.Millisecond

		// Simulate some errors
		if rand.Float64() < errorRate {
			h.loggingService.LogError(r.Context(), "web_error", "WEB001", "Web request failed",
				fmt.Errorf("internal server error"), map[string]interface{}{
					"endpoint":         endpoint,
					"response_time_ms": responseTime.Milliseconds(),
					"status_code":      500,
					"user_agent":       "Mozilla/5.0 (simulated)",
				})
		} else {
			h.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Web request processed",
				zap.String("endpoint", endpoint),
				zap.Int64("response_time_ms", responseTime.Milliseconds()),
				zap.Int("status_code", 200),
				zap.String("user_agent", "Mozilla/5.0 (simulated)"))
		}

		// Small delay to simulate real traffic
		time.Sleep(time.Millisecond * 10)
	}

	h.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Web service simulation completed",
		zap.Int("total_requests", pageViews))

	response := map[string]interface{}{
		"message":              "Web service simulation completed",
		"service_type":         "web-service",
		"requests_simulated":   pageViews,
		"avg_response_time_ms": avgResponseTime,
		"error_rate":           fmt.Sprintf("%.2f%%", errorRate*100),
		"endpoints_tested":     endpoints,
		"timestamp":            time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SimulateAPIServiceHandler simulates REST API services
func (h *SimulationHandlers) SimulateAPIServiceHandler(w http.ResponseWriter, r *http.Request) {
	// API service characteristics
	apiCalls := rand.Intn(100) + 20
	avgLatency := rand.Intn(100) + 25 // 25-125ms
	rateLimitHits := rand.Intn(5)     // 0-5 rate limit hits
	authFailures := rand.Intn(3)      // 0-3 auth failures

	h.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "API service simulation started",
		zap.String("service_type", "api-service"),
		zap.Int("api_calls", apiCalls),
		zap.Int("avg_latency_ms", avgLatency))

	// Simulate API endpoints
	apiEndpoints := []struct {
		method string
		path   string
		weight int // probability weight
	}{
		{"GET", "/api/v1/users", 30},
		{"POST", "/api/v1/users", 10},
		{"GET", "/api/v1/posts", 25},
		{"POST", "/api/v1/posts", 8},
		{"PUT", "/api/v1/posts/:id", 5},
		{"DELETE", "/api/v1/posts/:id", 2},
		{"GET", "/api/v1/health", 15},
		{"POST", "/api/v1/auth/login", 5},
	}

	for i := 0; i < apiCalls; i++ {
		// Select random endpoint based on weights
		totalWeight := 0
		for _, ep := range apiEndpoints {
			totalWeight += ep.weight
		}

		randWeight := rand.Intn(totalWeight)
		currentWeight := 0
		var selectedEndpoint struct {
			method string
			path   string
			weight int
		}

		for _, ep := range apiEndpoints {
			currentWeight += ep.weight
			if randWeight < currentWeight {
				selectedEndpoint = ep
				break
			}
		}

		latency := time.Duration(rand.Intn(150)+10) * time.Millisecond

		// Simulate different API scenarios
		switch {
		case rateLimitHits > 0 && rand.Float64() < 0.05: // 5% chance of rate limit
			rateLimitHits--
			h.loggingService.LogWithContext(zapcore.WarnLevel, r.Context(), "API rate limit exceeded",
				zap.String("method", selectedEndpoint.method),
				zap.String("endpoint", selectedEndpoint.path),
				zap.Int("status_code", 429),
				zap.Int64("latency_ms", latency.Milliseconds()),
				zap.String("client_ip", "192.168.1."+fmt.Sprintf("%d", rand.Intn(255))))
		case authFailures > 0 && rand.Float64() < 0.03: // 3% chance of auth failure
			authFailures--
			h.loggingService.LogError(r.Context(), "api_auth", "AUTH001", "API authentication failed",
				fmt.Errorf("invalid token"), map[string]interface{}{
					"method":      selectedEndpoint.method,
					"endpoint":    selectedEndpoint.path,
					"status_code": 401,
					"latency_ms":  latency.Milliseconds(),
				})
		case rand.Float64() < 0.02: // 2% chance of server error
			h.loggingService.LogError(r.Context(), "api_internal", "API001", "API internal error",
				fmt.Errorf("database connection timeout"), map[string]interface{}{
					"method":      selectedEndpoint.method,
					"endpoint":    selectedEndpoint.path,
					"status_code": 500,
					"latency_ms":  latency.Milliseconds(),
				})
		default: // Successful request
			h.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "API request processed",
				zap.String("method", selectedEndpoint.method),
				zap.String("endpoint", selectedEndpoint.path),
				zap.Int("status_code", 200),
				zap.Int64("latency_ms", latency.Milliseconds()),
				zap.Int("response_size_bytes", rand.Intn(5000)+100))
		}

		time.Sleep(time.Millisecond * 5)
	}

	response := map[string]interface{}{
		"message":             "API service simulation completed",
		"service_type":        "api-service",
		"requests_simulated":  apiCalls,
		"avg_latency_ms":      avgLatency,
		"rate_limit_hits":     rateLimitHits,
		"auth_failures":       authFailures,
		"endpoints_available": len(apiEndpoints),
		"timestamp":           time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SimulateDatabaseServiceHandler simulates database-heavy applications
func (h *SimulationHandlers) SimulateDatabaseServiceHandler(w http.ResponseWriter, r *http.Request) {
	// Database service characteristics
	queries := rand.Intn(80) + 20
	avgQueryTime := rand.Intn(50) + 10      // 10-60ms
	slowQueries := rand.Intn(5)             // 0-5 slow queries
	connectionPoolSize := rand.Intn(10) + 5 // 5-15 connections

	h.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Database service simulation started",
		zap.String("service_type", "database-service"),
		zap.Int("query_count", queries),
		zap.Int("connection_pool_size", connectionPoolSize))

	queryTypes := []string{"SELECT", "INSERT", "UPDATE", "DELETE"}
	tables := []string{"users", "posts", "comments", "categories", "sessions", "logs"}

	for i := 0; i < queries; i++ {
		queryType := queryTypes[rand.Intn(len(queryTypes))]
		table := tables[rand.Intn(len(tables))]
		queryTime := time.Duration(rand.Intn(100)+5) * time.Millisecond

		// Simulate slow queries
		if slowQueries > 0 && rand.Float64() < 0.08 { // 8% chance of slow query
			slowQueries--
			queryTime = time.Duration(rand.Intn(2000)+1000) * time.Millisecond // 1-3 seconds
			h.loggingService.LogWithContext(zapcore.WarnLevel, r.Context(), "Slow database query detected",
				zap.String("query_type", queryType),
				zap.String("table", table),
				zap.Int64("duration_ms", queryTime.Milliseconds()),
				zap.Int("rows_affected", rand.Intn(10000)),
				zap.String("query_id", fmt.Sprintf("query_%d", i)))
		} else if rand.Float64() < 0.03 { // 3% chance of query error
			h.loggingService.LogError(r.Context(), "database_error", "DB001", "Database query failed",
				fmt.Errorf("table lock timeout"), map[string]interface{}{
					"query_type":  queryType,
					"table":       table,
					"duration_ms": queryTime.Milliseconds(),
					"query_id":    fmt.Sprintf("query_%d", i),
				})
		} else {
			h.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Database query executed",
				zap.String("query_type", queryType),
				zap.String("table", table),
				zap.Int64("duration_ms", queryTime.Milliseconds()),
				zap.Int("rows_affected", rand.Intn(100)),
				zap.String("query_id", fmt.Sprintf("query_%d", i)))
		}

		time.Sleep(time.Millisecond * 8)
	}

	// Simulate connection pool metrics
	h.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Database connection pool status",
		zap.Int("pool_size", connectionPoolSize),
		zap.Int("active_connections", rand.Intn(connectionPoolSize)),
		zap.Int("idle_connections", rand.Intn(connectionPoolSize/2)),
		zap.Int("total_queries", queries))

	response := map[string]interface{}{
		"message":              "Database service simulation completed",
		"service_type":         "database-service",
		"queries_executed":     queries,
		"avg_query_time_ms":    avgQueryTime,
		"slow_queries":         slowQueries,
		"connection_pool_size": connectionPoolSize,
		"tables_accessed":      tables,
		"timestamp":            time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SimulateStaticSiteHandler simulates static file serving (CDN-like)
func (h *SimulationHandlers) SimulateStaticSiteHandler(w http.ResponseWriter, r *http.Request) {
	// Static site characteristics
	requests := rand.Intn(200) + 50
	cacheHitRate := rand.Float64()*0.3 + 0.7 // 70-100% cache hit rate

	h.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Static site simulation started",
		zap.String("service_type", "static-site"),
		zap.Int("expected_requests", requests),
		zap.Float64("cache_hit_rate", cacheHitRate))

	fileTypes := []struct {
		ext    string
		size   int // average size in KB
		weight int
	}{
		{".html", 15, 20},
		{".css", 25, 15},
		{".js", 35, 15},
		{".jpg", 150, 25},
		{".png", 80, 10},
		{".svg", 5, 8},
		{".pdf", 500, 3},
		{".mp4", 2000, 2},
		{".woff2", 45, 2},
	}

	cacheHits := 0
	cacheMisses := 0
	totalBytes := 0

	for i := 0; i < requests; i++ {
		// Select file type based on weights
		totalWeight := 0
		for _, ft := range fileTypes {
			totalWeight += ft.weight
		}

		randWeight := rand.Intn(totalWeight)
		currentWeight := 0
		var selectedFile struct {
			ext    string
			size   int
			weight int
		}

		for _, ft := range fileTypes {
			currentWeight += ft.weight
			if randWeight < currentWeight {
				selectedFile = ft
				break
			}
		}

		// Determine cache hit/miss
		isCache := rand.Float64() < cacheHitRate
		responseTime := time.Duration(rand.Intn(50)+5) * time.Millisecond
		if !isCache {
			responseTime = time.Duration(rand.Intn(200)+50) * time.Millisecond // Cache miss is slower
		}

		fileSize := selectedFile.size + rand.Intn(selectedFile.size/2) - selectedFile.size/4 // Vary size ±25%
		totalBytes += fileSize * 1024                                                        // Convert to bytes

		fileName := fmt.Sprintf("/static/file_%d%s", rand.Intn(1000), selectedFile.ext)

		if isCache {
			cacheHits++
			h.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Static file served from cache",
				zap.String("file", fileName),
				zap.Int("size_bytes", fileSize*1024),
				zap.Int64("response_time_ms", responseTime.Milliseconds()),
				zap.String("cache_status", "HIT"),
				zap.String("client_ip", "192.168.1."+fmt.Sprintf("%d", rand.Intn(255))))
		} else {
			cacheMisses++
			h.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Static file served from origin",
				zap.String("file", fileName),
				zap.Int("size_bytes", fileSize*1024),
				zap.Int64("response_time_ms", responseTime.Milliseconds()),
				zap.String("cache_status", "MISS"),
				zap.String("client_ip", "192.168.1."+fmt.Sprintf("%d", rand.Intn(255))))
		}

		time.Sleep(time.Millisecond * 3)
	}

	actualCacheHitRate := float64(cacheHits) / float64(requests)

	h.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Static site simulation completed",
		zap.Int("total_requests", requests),
		zap.Int("cache_hits", cacheHits),
		zap.Int("cache_misses", cacheMisses),
		zap.Float64("actual_cache_hit_rate", actualCacheHitRate),
		zap.Float64("total_bandwidth_mb", float64(totalBytes)/(1024*1024)))

	response := map[string]interface{}{
		"message":            "Static site simulation completed",
		"service_type":       "static-site",
		"requests_served":    requests,
		"cache_hit_rate":     fmt.Sprintf("%.1f%%", actualCacheHitRate*100),
		"cache_hits":         cacheHits,
		"cache_misses":       cacheMisses,
		"total_bandwidth_mb": fmt.Sprintf("%.2f", float64(totalBytes)/(1024*1024)),
		"file_types_served":  len(fileTypes),
		"timestamp":          time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SimulateMicroserviceHandler simulates microservice communication patterns
func (h *SimulationHandlers) SimulateMicroserviceHandler(w http.ResponseWriter, r *http.Request) {
	// Microservice characteristics
	serviceCalls := rand.Intn(30) + 10
	circuitBreakerTrips := rand.Intn(2)
	retryAttempts := rand.Intn(5)

	h.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Microservice simulation started",
		zap.String("service_type", "microservice"),
		zap.Int("service_calls", serviceCalls))

	// Define microservices
	services := []string{
		"user-service",
		"auth-service",
		"notification-service",
		"payment-service",
		"inventory-service",
		"shipping-service",
	}

	for i := 0; i < serviceCalls; i++ {
		caller := services[rand.Intn(len(services))]
		callee := services[rand.Intn(len(services))]

		// Skip self-calls
		if caller == callee {
			continue
		}

		latency := time.Duration(rand.Intn(150)+20) * time.Millisecond

		// Simulate different microservice scenarios
		switch {
		case circuitBreakerTrips > 0 && rand.Float64() < 0.1: // 10% chance of circuit breaker
			circuitBreakerTrips--
			h.loggingService.LogError(r.Context(), "circuit_breaker", "CB001", "Circuit breaker tripped",
				fmt.Errorf("service unavailable"), map[string]interface{}{
					"caller_service":        caller,
					"target_service":        callee,
					"latency_ms":            latency.Milliseconds(),
					"circuit_breaker_state": "OPEN",
				})
		case retryAttempts > 0 && rand.Float64() < 0.08: // 8% chance of retry
			retryAttempts--
			h.loggingService.LogWithContext(zapcore.WarnLevel, r.Context(), "Service call retry",
				zap.String("caller_service", caller),
				zap.String("target_service", callee),
				zap.Int64("latency_ms", latency.Milliseconds()),
				zap.Int("retry_attempt", rand.Intn(3)+1),
				zap.String("original_error", "Connection timeout"))
		case rand.Float64() < 0.05: // 5% chance of service error
			h.loggingService.LogError(r.Context(), "microservice_error", "MS001", "Microservice call failed",
				fmt.Errorf("service temporarily unavailable"), map[string]interface{}{
					"caller_service": caller,
					"target_service": callee,
					"latency_ms":     latency.Milliseconds(),
					"status_code":    503,
				})
		default: // Successful call
			h.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Microservice call succeeded",
				zap.String("caller_service", caller),
				zap.String("target_service", callee),
				zap.Int64("latency_ms", latency.Milliseconds()),
				zap.Int("status_code", 200),
				zap.String("correlation_id", fmt.Sprintf("corr_%d", rand.Intn(10000))))
		}

		time.Sleep(time.Millisecond * 15)
	}

	response := map[string]interface{}{
		"message":               "Microservice simulation completed",
		"service_type":          "microservice",
		"service_calls":         serviceCalls,
		"services_involved":     services,
		"circuit_breaker_trips": circuitBreakerTrips,
		"retry_attempts":        retryAttempts,
		"timestamp":             time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
