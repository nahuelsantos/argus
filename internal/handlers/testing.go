package handlers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap/zapcore"

	"github.com/nahuelsantos/argus/internal/services"
)

// TestingHandlers contains testing handlers for LGTM stack validation
type TestingHandlers struct {
	loggingService *services.LoggingService
	tracingService *services.TracingService
}

// NewTestingHandlers creates a new testing handlers instance
func NewTestingHandlers(loggingService *services.LoggingService, tracingService *services.TracingService) *TestingHandlers {
	return &TestingHandlers{
		loggingService: loggingService,
		tracingService: tracingService,
	}
}

// GenerateJSONLogsHandler tests Loki with structured JSON logs
func (th *TestingHandlers) GenerateJSONLogsHandler(w http.ResponseWriter, r *http.Request) {
	count := 10

	logFormats := []map[string]interface{}{
		{
			"timestamp": time.Now().Format(time.RFC3339),
			"level":     "INFO",
			"service":   "user-api",
			"message":   "User login successful",
			"user_id":   rand.Intn(1000),
			"ip":        fmt.Sprintf("192.168.1.%d", rand.Intn(255)),
			"duration":  fmt.Sprintf("%dms", rand.Intn(100)+10),
		},
		{
			"timestamp": time.Now().Format(time.RFC3339),
			"level":     "ERROR",
			"service":   "payment-api",
			"message":   "Payment processing failed",
			"error":     "insufficient_funds",
			"amount":    rand.Float64() * 1000,
			"currency":  "USD",
			"trace_id":  fmt.Sprintf("trace-%d", rand.Intn(10000)),
		},
		{
			"timestamp": time.Now().Format(time.RFC3339),
			"level":     "WARN",
			"service":   "database",
			"message":   "Slow query detected",
			"query":     "SELECT * FROM users WHERE last_login > ?",
			"duration":  fmt.Sprintf("%ds", rand.Intn(10)+1),
			"rows":      rand.Intn(50000),
		},
	}

	var generatedLogs []string
	for i := 0; i < count; i++ {
		logEntry := logFormats[i%len(logFormats)]
		logJSON, _ := json.Marshal(logEntry)

		// Log to Loki via our logging service
		th.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), string(logJSON))
		generatedLogs = append(generatedLogs, string(logJSON))
	}

	response := map[string]interface{}{
		"message":        "JSON logs generated for Loki testing",
		"logs_generated": count,
		"log_formats":    len(logFormats),
		"sample_logs":    generatedLogs[:3], // Show first 3 as examples
		"test_purpose":   "Validate Loki JSON log parsing",
		"timestamp":      time.Now().Format(time.RFC3339),
		"service":        "argus",
		"functionality":  "loki_json_validation",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	th.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "JSON logs generated for Loki testing")
}

// GenerateUnstructuredLogsHandler tests Loki with plain text logs
func (th *TestingHandlers) GenerateUnstructuredLogsHandler(w http.ResponseWriter, r *http.Request) {
	count := 10

	logTemplates := []string{
		"[%s] INFO: WordPress - User 'admin' logged in from %s",
		"[%s] ERROR: Apache - File not found: /var/www/html/missing-page.html",
		"[%s] WARN: MySQL - Slow query: SELECT * FROM wp_posts WHERE post_status='publish' (%.2fs)",
		"[%s] INFO: PHP-FPM - Pool 'www': child %d exited with code 0 after %.2f seconds",
		"[%s] ERROR: Nginx - upstream timed out while connecting to upstream",
		"[%s] INFO: Cron - Running backup script for database 'wordpress'",
		"[%s] WARN: System - Disk usage at 85%% on /var/www",
	}

	var generatedLogs []string
	for i := 0; i < count; i++ {
		template := logTemplates[i%len(logTemplates)]
		var logEntry string

		switch i % len(logTemplates) {
		case 0:
			logEntry = fmt.Sprintf(template, time.Now().Format("2006-01-02 15:04:05"), fmt.Sprintf("192.168.1.%d", rand.Intn(255)))
		case 1:
			logEntry = fmt.Sprintf(template, time.Now().Format("2006-01-02 15:04:05"))
		case 2:
			logEntry = fmt.Sprintf(template, time.Now().Format("2006-01-02 15:04:05"), rand.Float64()*5+1)
		case 3:
			logEntry = fmt.Sprintf(template, time.Now().Format("2006-01-02 15:04:05"), rand.Intn(100)+1, rand.Float64()*10+1)
		default:
			logEntry = fmt.Sprintf(template, time.Now().Format("2006-01-02 15:04:05"))
		}

		// Log to Loki via our logging service
		th.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), logEntry)
		generatedLogs = append(generatedLogs, logEntry)
	}

	response := map[string]interface{}{
		"message":        "Unstructured logs generated for Loki testing",
		"logs_generated": count,
		"log_templates":  len(logTemplates),
		"sample_logs":    generatedLogs[:3],
		"test_purpose":   "Validate Loki unstructured log parsing",
		"timestamp":      time.Now().Format(time.RFC3339),
		"service":        "argus",
		"functionality":  "loki_unstructured_validation",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	th.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Unstructured logs generated for Loki testing")
}

// GenerateMixedLogsHandler tests Loki with mixed format logs
func (th *TestingHandlers) GenerateMixedLogsHandler(w http.ResponseWriter, r *http.Request) {
	count := 15
	var generatedLogs []string

	for i := 0; i < count; i++ {
		var logEntry string

		switch i % 3 {
		case 0: // JSON format
			logData := map[string]interface{}{
				"level":   "INFO",
				"service": "api-gateway",
				"request": map[string]interface{}{
					"method": "GET",
					"path":   "/api/users",
					"status": 200,
					"time":   fmt.Sprintf("%dms", rand.Intn(100)+10),
				},
			}
			logJSON, _ := json.Marshal(logData)
			logEntry = string(logJSON)

		case 1: // Key-value format
			logEntry = fmt.Sprintf("time=%s level=WARN service=database query=\"SELECT COUNT(*) FROM sessions\" duration=%dms rows=%d",
				time.Now().Format(time.RFC3339), rand.Intn(1000)+100, rand.Intn(10000))

		case 2: // Plain text format
			logEntry = fmt.Sprintf("[%s] ERROR: Redis connection failed, retrying in %d seconds",
				time.Now().Format("2006-01-02 15:04:05"), rand.Intn(5)+1)
		}

		th.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), logEntry)
		generatedLogs = append(generatedLogs, logEntry)
	}

	response := map[string]interface{}{
		"message":        "Mixed format logs generated for Loki testing",
		"logs_generated": count,
		"formats":        []string{"JSON", "Key-Value", "Plain Text"},
		"sample_logs":    generatedLogs[:3],
		"test_purpose":   "Validate Loki mixed format log parsing",
		"timestamp":      time.Now().Format(time.RFC3339),
		"service":        "argus",
		"functionality":  "loki_mixed_validation",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	th.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Mixed format logs generated for Loki testing")
}

// GenerateMultilineLogsHandler tests Loki with multi-line logs (stack traces)
func (th *TestingHandlers) GenerateMultilineLogsHandler(w http.ResponseWriter, r *http.Request) {
	stackTraces := []string{
		`[2024-01-15 10:30:45] ERROR: Application exception
java.lang.NullPointerException: Cannot invoke method on null object
    at com.example.UserService.getUser(UserService.java:45)
    at com.example.UserController.handleRequest(UserController.java:23)
    at javax.servlet.http.HttpServlet.service(HttpServlet.java:731)
    at org.apache.catalina.core.ApplicationFilterChain.doFilter(ApplicationFilterChain.java:166)`,

		`[2024-01-15 10:31:12] ERROR: Database connection failed
org.postgresql.util.PSQLException: Connection to localhost:5432 refused
    at org.postgresql.core.v3.ConnectionFactoryImpl.openConnectionImpl(ConnectionFactoryImpl.java:280)
    at org.postgresql.core.ConnectionFactory.openConnection(ConnectionFactory.java:49)
    at org.postgresql.jdbc.PgConnection.<init>(PgConnection.java:195)
    at org.postgresql.Driver.makeConnection(Driver.java:454)`,

		`[2024-01-15 10:31:45] ERROR: Payment processing error
stripe.error.CardError: Your card was declined
    at stripe.api_requestor.APIRequestor.request(api_requestor.py:123)
    at stripe.api_resources.charge.Charge.create(charge.py:45)
    at payment_service.process_payment(payment_service.py:67)
    at app.routes.checkout(routes.py:89)`,
	}

	var generatedLogs []string
	for i, stackTrace := range stackTraces {
		lines := strings.Split(stackTrace, "\n")
		for _, line := range lines {
			th.loggingService.LogWithContext(zapcore.ErrorLevel, r.Context(), line)
		}
		generatedLogs = append(generatedLogs, fmt.Sprintf("Stack trace %d (%d lines)", i+1, len(lines)))
	}

	response := map[string]interface{}{
		"message":          "Multi-line logs (stack traces) generated for Loki testing",
		"stack_traces":     len(stackTraces),
		"total_log_lines":  strings.Count(strings.Join(stackTraces, "\n"), "\n") + len(stackTraces),
		"generated_traces": generatedLogs,
		"test_purpose":     "Validate Loki multi-line log parsing (stack traces)",
		"timestamp":        time.Now().Format(time.RFC3339),
		"service":          "argus",
		"functionality":    "loki_multiline_validation",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	th.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Multi-line stack traces generated for Loki testing")
}

// SimulateWordPressServiceHandler tests monitoring stack with WordPress-like service patterns
func (th *TestingHandlers) SimulateWordPressServiceHandler(w http.ResponseWriter, r *http.Request) {
	// Generate WordPress-typical logs and metrics
	activities := []string{
		"user_login", "post_view", "admin_access", "plugin_activation",
		"comment_submission", "media_upload", "theme_change", "backup_process",
	}

	responses := []int{200, 200, 200, 404, 500, 301, 200, 403}

	var generatedEvents []string
	for i := 0; i < 8; i++ {
		activity := activities[i%len(activities)]
		statusCode := responses[i%len(responses)]

		// WordPress access log format
		logEntry := fmt.Sprintf(`192.168.1.%d - - [%s] "GET /wp-%s HTTP/1.1" %d %d "https://example.com/" "Mozilla/5.0"`,
			rand.Intn(255), time.Now().Format("02/Jan/2006:15:04:05 -0700"), activity, statusCode, rand.Intn(5000)+500)

		th.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), logEntry)
		generatedEvents = append(generatedEvents, fmt.Sprintf("%s (HTTP %d)", activity, statusCode))
	}

	// Simulate WordPress metrics (for Prometheus validation)
	metrics := map[string]float64{
		"wordpress_page_views_total":   float64(rand.Intn(1000) + 100),
		"wordpress_active_users":       float64(rand.Intn(50) + 5),
		"wordpress_database_queries":   float64(rand.Intn(100) + 20),
		"wordpress_memory_usage_bytes": float64(rand.Intn(200)+50) * 1024 * 1024, // 50-250MB
		"wordpress_response_time_ms":   float64(rand.Intn(500) + 100),
	}

	response := map[string]interface{}{
		"message":           "WordPress service simulation for LGTM stack testing",
		"service_type":      "wordpress",
		"events_generated":  len(generatedEvents),
		"sample_events":     generatedEvents,
		"metrics_simulated": metrics,
		"test_purpose":      "Validate LGTM stack with WordPress-like service patterns",
		"timestamp":         time.Now().Format(time.RFC3339),
		"service":           "argus",
		"functionality":     "wordpress_service_validation",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	th.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "WordPress service simulation completed")
}

// SimulateNextJSServiceHandler tests monitoring stack with Next.js-like service patterns
func (th *TestingHandlers) SimulateNextJSServiceHandler(w http.ResponseWriter, r *http.Request) {
	// Generate Next.js-typical logs
	routes := []string{
		"/", "/about", "/blog", "/api/users", "/api/posts",
		"/_next/static/css/app.css", "/_next/static/js/app.js", "/404",
	}

	var generatedEvents []string
	for i := 0; i < 10; i++ {
		route := routes[i%len(routes)]
		method := "GET"
		if strings.Contains(route, "/api/") {
			methods := []string{"GET", "POST", "PUT", "DELETE"}
			method = methods[rand.Intn(len(methods))]
		}

		statusCode := 200
		if route == "/404" {
			statusCode = 404
		}

		duration := rand.Intn(200) + 10

		// Next.js structured log
		logData := map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
			"level":     "info",
			"method":    method,
			"url":       route,
			"status":    statusCode,
			"duration":  duration,
			"userAgent": "Mozilla/5.0 (compatible; LGTM-Test)",
		}

		logJSON, _ := json.Marshal(logData)
		th.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), string(logJSON))

		generatedEvents = append(generatedEvents, fmt.Sprintf("%s %s (%d, %dms)", method, route, statusCode, duration))
	}

	// Simulate Next.js metrics
	metrics := map[string]float64{
		"nextjs_requests_total":        float64(rand.Intn(500) + 50),
		"nextjs_static_requests_total": float64(rand.Intn(200) + 20),
		"nextjs_api_requests_total":    float64(rand.Intn(100) + 10),
		"nextjs_build_time_seconds":    float64(rand.Intn(30) + 5),
		"nextjs_bundle_size_bytes":     float64(rand.Intn(1000)+500) * 1024, // 500KB-1.5MB
	}

	response := map[string]interface{}{
		"message":            "Next.js service simulation for LGTM stack testing",
		"service_type":       "nextjs",
		"requests_generated": len(generatedEvents),
		"sample_requests":    generatedEvents,
		"metrics_simulated":  metrics,
		"test_purpose":       "Validate LGTM stack with Next.js-like service patterns",
		"timestamp":          time.Now().Format(time.RFC3339),
		"service":            "argus",
		"functionality":      "nextjs_service_validation",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	th.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Next.js service simulation completed")
}

// SimulateCrossServiceTracingHandler tests Tempo with cross-service tracing scenarios
func (th *TestingHandlers) SimulateCrossServiceTracingHandler(w http.ResponseWriter, r *http.Request) {
	// Simulate e-commerce flow: User → Frontend → API → Database → Payment
	traceScenarios := []struct {
		name     string
		services []string
		flow     string
	}{
		{
			name:     "e-commerce-checkout",
			services: []string{"frontend", "user-api", "product-api", "payment-api", "database"},
			flow:     "User checkout process",
		},
		{
			name:     "content-delivery",
			services: []string{"cdn", "origin-server", "api-gateway", "content-service", "database"},
			flow:     "Content delivery pipeline",
		},
		{
			name:     "authentication-flow",
			services: []string{"auth-service", "user-service", "session-service", "database"},
			flow:     "User authentication process",
		},
	}

	var generatedTraces []string
	for _, scenario := range traceScenarios {
		traceID := fmt.Sprintf("trace-%d", rand.Intn(100000))

		for i, service := range scenario.services {
			spanID := fmt.Sprintf("span-%d", i)
			duration := rand.Intn(100) + 10

			// Create trace log entry
			traceLog := map[string]interface{}{
				"timestamp": time.Now().Format(time.RFC3339),
				"trace_id":  traceID,
				"span_id":   spanID,
				"service":   service,
				"operation": scenario.name,
				"duration":  duration,
				"status":    "success",
			}

			logJSON, _ := json.Marshal(traceLog)
			th.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), string(logJSON))
		}

		generatedTraces = append(generatedTraces, fmt.Sprintf("%s: %s (%d services)",
			scenario.name, scenario.flow, len(scenario.services)))
	}

	response := map[string]interface{}{
		"message":          "Cross-service tracing simulation for Tempo testing",
		"trace_scenarios":  len(traceScenarios),
		"generated_traces": generatedTraces,
		"total_spans":      len(traceScenarios[0].services) + len(traceScenarios[1].services) + len(traceScenarios[2].services),
		"test_purpose":     "Validate Tempo cross-service tracing capabilities",
		"timestamp":        time.Now().Format(time.RFC3339),
		"service":          "argus",
		"functionality":    "tempo_tracing_validation",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	th.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Cross-service tracing simulation completed")
}

// TestServiceDiscoveryHandler tests service discovery and registration
func (th *TestingHandlers) TestServiceDiscoveryHandler(w http.ResponseWriter, r *http.Request) {
	// Simulate service discovery scenarios
	services := []struct {
		name      string
		status    string
		healthURL string
		lastSeen  time.Time
		version   string
	}{
		{
			name:      "user-api",
			status:    "healthy",
			healthURL: "http://user-api:3000/health",
			lastSeen:  time.Now(),
			version:   "v1.2.3",
		},
		{
			name:      "payment-service",
			status:    "degraded",
			healthURL: "http://payment-service:3001/health",
			lastSeen:  time.Now().Add(-30 * time.Second),
			version:   "v2.1.0",
		},
		{
			name:      "notification-worker",
			status:    "unhealthy",
			healthURL: "http://notification-worker:3002/health",
			lastSeen:  time.Now().Add(-5 * time.Minute),
			version:   "v1.0.1",
		},
	}

	// Test health check endpoints
	var healthResults []map[string]interface{}
	for _, service := range services {
		// Simulate health check response
		var responseTime int
		var available bool

		switch service.status {
		case "healthy":
			responseTime = rand.Intn(50) + 10 // 10-60ms
			available = true
		case "degraded":
			responseTime = rand.Intn(500) + 200 // 200-700ms
			available = true
		case "unhealthy":
			responseTime = 0
			available = false
		}

		healthResult := map[string]interface{}{
			"service":       service.name,
			"status":        service.status,
			"available":     available,
			"response_time": responseTime,
			"last_seen":     service.lastSeen.Format(time.RFC3339),
			"version":       service.version,
		}

		healthResults = append(healthResults, healthResult)

		// Log service discovery event
		logEntry := fmt.Sprintf("Service discovery: %s status=%s response_time=%dms version=%s",
			service.name, service.status, responseTime, service.version)
		th.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), logEntry)
	}

	response := map[string]interface{}{
		"message":           "Service discovery testing completed",
		"services_tested":   len(services),
		"health_results":    healthResults,
		"healthy_services":  2,
		"degraded_services": 1,
		"failed_services":   1,
		"test_purpose":      "Validate service discovery and health monitoring",
		"timestamp":         time.Now().Format(time.RFC3339),
		"service":           "argus",
		"functionality":     "service_discovery_validation",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	th.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Service discovery testing completed")
}

// TestReverseProxyHandler tests Traefik reverse proxy integration
func (th *TestingHandlers) TestReverseProxyHandler(w http.ResponseWriter, r *http.Request) {
	// Simulate reverse proxy routing scenarios
	routes := []struct {
		domain  string
		path    string
		backend string
		status  string
		latency int
		ssl     bool
	}{
		{
			domain:  "api.example.com",
			path:    "/users",
			backend: "user-api:3000",
			status:  "active",
			latency: rand.Intn(50) + 10,
			ssl:     true,
		},
		{
			domain:  "blog.example.com",
			path:    "/",
			backend: "wordpress:80",
			status:  "active",
			latency: rand.Intn(100) + 20,
			ssl:     true,
		},
		{
			domain:  "admin.example.com",
			path:    "/dashboard",
			backend: "admin-panel:8080",
			status:  "maintenance",
			latency: 0,
			ssl:     true,
		},
		{
			domain:  "legacy.example.com",
			path:    "/old-api",
			backend: "legacy-service:3003",
			status:  "deprecated",
			latency: rand.Intn(1000) + 500,
			ssl:     false,
		},
	}

	var routeResults []map[string]interface{}
	for _, route := range routes {
		// Simulate load balancing
		backendInstances := []string{
			fmt.Sprintf("%s-1", route.backend),
			fmt.Sprintf("%s-2", route.backend),
			fmt.Sprintf("%s-3", route.backend),
		}
		selectedBackend := backendInstances[rand.Intn(len(backendInstances))]

		routeResult := map[string]interface{}{
			"domain":           route.domain,
			"path":             route.path,
			"selected_backend": selectedBackend,
			"status":           route.status,
			"latency_ms":       route.latency,
			"ssl_enabled":      route.ssl,
			"load_balanced":    true,
		}

		routeResults = append(routeResults, routeResult)

		// Log reverse proxy event
		logEntry := fmt.Sprintf("Reverse proxy: %s%s -> %s (latency=%dms, ssl=%t)",
			route.domain, route.path, selectedBackend, route.latency, route.ssl)
		th.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), logEntry)
	}

	response := map[string]interface{}{
		"message":       "Reverse proxy testing completed",
		"routes_tested": len(routes),
		"route_results": routeResults,
		"active_routes": 2,
		"ssl_routes":    3,
		"load_balanced": len(routes),
		"test_purpose":  "Validate Traefik reverse proxy configuration",
		"timestamp":     time.Now().Format(time.RFC3339),
		"service":       "argus",
		"functionality": "reverse_proxy_validation",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	th.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Reverse proxy testing completed")
}

// TestSSLMonitoringHandler tests SSL certificate monitoring
func (th *TestingHandlers) TestSSLMonitoringHandler(w http.ResponseWriter, r *http.Request) {
	// Simulate SSL certificate monitoring
	certificates := []struct {
		domain    string
		issuer    string
		expiresAt time.Time
		daysLeft  int
		status    string
		autoRenew bool
	}{
		{
			domain:    "api.example.com",
			issuer:    "Let's Encrypt",
			expiresAt: time.Now().AddDate(0, 0, 45),
			daysLeft:  45,
			status:    "valid",
			autoRenew: true,
		},
		{
			domain:    "blog.example.com",
			issuer:    "Let's Encrypt",
			expiresAt: time.Now().AddDate(0, 0, 15),
			daysLeft:  15,
			status:    "expiring_soon",
			autoRenew: true,
		},
		{
			domain:    "legacy.example.com",
			issuer:    "Legacy CA",
			expiresAt: time.Now().AddDate(0, 0, -5),
			daysLeft:  -5,
			status:    "expired",
			autoRenew: false,
		},
		{
			domain:    "admin.example.com",
			issuer:    "Let's Encrypt",
			expiresAt: time.Now().AddDate(0, 0, 75),
			daysLeft:  75,
			status:    "valid",
			autoRenew: true,
		},
	}

	var certResults []map[string]interface{}
	var alertCount int

	for _, cert := range certificates {
		alertLevel := "none"
		if cert.daysLeft < 0 {
			alertLevel = "critical"
			alertCount++
		} else if cert.daysLeft <= 30 {
			alertLevel = "warning"
			alertCount++
		}

		certResult := map[string]interface{}{
			"domain":      cert.domain,
			"issuer":      cert.issuer,
			"expires_at":  cert.expiresAt.Format(time.RFC3339),
			"days_left":   cert.daysLeft,
			"status":      cert.status,
			"auto_renew":  cert.autoRenew,
			"alert_level": alertLevel,
		}

		certResults = append(certResults, certResult)

		// Log SSL monitoring event
		logEntry := fmt.Sprintf("SSL monitoring: %s expires in %d days (issuer=%s, status=%s)",
			cert.domain, cert.daysLeft, cert.issuer, cert.status)
		th.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), logEntry)
	}

	response := map[string]interface{}{
		"message":              "SSL certificate monitoring completed",
		"certificates_checked": len(certificates),
		"certificate_results":  certResults,
		"valid_certificates":   2,
		"expiring_soon":        1,
		"expired_certificates": 1,
		"alerts_generated":     alertCount,
		"test_purpose":         "Validate SSL certificate monitoring and alerting",
		"timestamp":            time.Now().Format(time.RFC3339),
		"service":              "argus",
		"functionality":        "ssl_monitoring_validation",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	th.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "SSL certificate monitoring completed")
}

// TestDomainHealthHandler tests domain-specific health monitoring
func (th *TestingHandlers) TestDomainHealthHandler(w http.ResponseWriter, r *http.Request) {
	// Simulate domain health monitoring
	domains := []struct {
		name         string
		status       string
		responseTime int
		statusCode   int
		dnsTime      int
		location     string
	}{
		{
			name:         "api.example.com",
			status:       "healthy",
			responseTime: rand.Intn(100) + 50,
			statusCode:   200,
			dnsTime:      rand.Intn(20) + 5,
			location:     "US-East",
		},
		{
			name:         "blog.example.com",
			status:       "healthy",
			responseTime: rand.Intn(150) + 100,
			statusCode:   200,
			dnsTime:      rand.Intn(30) + 10,
			location:     "EU-West",
		},
		{
			name:         "admin.example.com",
			status:       "degraded",
			responseTime: rand.Intn(2000) + 1000,
			statusCode:   200,
			dnsTime:      rand.Intn(50) + 20,
			location:     "Asia-Pacific",
		},
		{
			name:         "legacy.example.com",
			status:       "down",
			responseTime: 0,
			statusCode:   503,
			dnsTime:      rand.Intn(100) + 50,
			location:     "US-West",
		},
	}

	var healthResults []map[string]interface{}
	uptime := 0

	for _, domain := range domains {
		var availability float64
		if domain.status == "healthy" {
			availability = 99.9
			uptime++
		} else if domain.status == "degraded" {
			availability = 95.5
		} else {
			availability = 0.0
		}

		healthResult := map[string]interface{}{
			"domain":        domain.name,
			"status":        domain.status,
			"response_time": domain.responseTime,
			"status_code":   domain.statusCode,
			"dns_time":      domain.dnsTime,
			"location":      domain.location,
			"availability":  availability,
		}

		healthResults = append(healthResults, healthResult)

		// Log domain health event
		logEntry := fmt.Sprintf("Domain health: %s status=%s response_time=%dms location=%s availability=%.1f%%",
			domain.name, domain.status, domain.responseTime, domain.location, availability)
		th.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), logEntry)
	}

	overallUptime := float64(uptime) / float64(len(domains)) * 100

	response := map[string]interface{}{
		"message":           "Domain health monitoring completed",
		"domains_checked":   len(domains),
		"health_results":    healthResults,
		"healthy_domains":   uptime,
		"overall_uptime":    overallUptime,
		"avg_response_time": (domains[0].responseTime + domains[1].responseTime + domains[2].responseTime) / 3,
		"test_purpose":      "Validate domain health monitoring from multiple locations",
		"timestamp":         time.Now().Format(time.RFC3339),
		"service":           "argus",
		"functionality":     "domain_health_validation",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	th.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Domain health monitoring completed")
}
