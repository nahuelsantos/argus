package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/nahuelsantos/argus/internal/config"
	"github.com/nahuelsantos/argus/internal/handlers"
	"github.com/nahuelsantos/argus/internal/metrics"
	"github.com/nahuelsantos/argus/internal/middleware"
	"github.com/nahuelsantos/argus/internal/services"
)

func main() {
	// Initialize configuration
	serviceConfig := config.GetServiceConfig()

	fmt.Printf("Starting Argus - LGTM Stack Validator %s...\n", serviceConfig.Version)

	// Initialize services
	loggingService := services.NewLoggingService()
	loggingService.InitLogger()

	tracingService := services.NewTracingService()
	tracingService.InitTracer()

	alertingService := services.NewAlertingService()
	alertingService.InitAlertManager()

	// Register Prometheus metrics
	metrics.RegisterMetrics()

	// Initialize handlers
	basicHandlers := handlers.NewBasicHandlers(loggingService, tracingService)
	simulationHandlers := handlers.NewSimulationHandlers(loggingService, tracingService)
	alertingHandlers := handlers.NewAlertingHandlers(loggingService, alertingService)
	testingHandlers := handlers.NewTestingHandlers(loggingService, tracingService)
	integrationHandlers := handlers.NewIntegrationHandlers(loggingService, tracingService)
	performanceHandlers := handlers.NewPerformanceHandlers(loggingService, tracingService)

	// Create HTTP mux
	mux := http.NewServeMux()

	// Core monitoring test endpoints
	mux.HandleFunc("/health", basicHandlers.HealthHandler)
	mux.HandleFunc("/lgtm-status", basicHandlers.LGTMStatusHandler)
	mux.HandleFunc("/generate-metrics", basicHandlers.GenerateMetricsHandler)
	mux.HandleFunc("/generate-logs", basicHandlers.GenerateLogsHandler)
	mux.HandleFunc("/generate-error", basicHandlers.GenerateErrorHandler)
	mux.HandleFunc("/cpu-load", basicHandlers.CPULoadHandler)
	mux.HandleFunc("/memory-load", basicHandlers.MemoryLoadHandler)

	// Multi-Service Simulation endpoints
	mux.HandleFunc("/simulate/web-service", simulationHandlers.SimulateWebServiceHandler)
	mux.HandleFunc("/simulate/api-service", simulationHandlers.SimulateAPIServiceHandler)
	mux.HandleFunc("/simulate/database-service", simulationHandlers.SimulateDatabaseServiceHandler)
	mux.HandleFunc("/simulate/static-site", simulationHandlers.SimulateStaticSiteHandler)
	mux.HandleFunc("/simulate/microservice", simulationHandlers.SimulateMicroserviceHandler)

	// Test data variety endpoints
	mux.HandleFunc("/generate-logs/json", testingHandlers.GenerateJSONLogsHandler)
	mux.HandleFunc("/generate-logs/unstructured", testingHandlers.GenerateUnstructuredLogsHandler)
	mux.HandleFunc("/generate-logs/mixed", testingHandlers.GenerateMixedLogsHandler)
	mux.HandleFunc("/generate-logs/multiline", testingHandlers.GenerateMultilineLogsHandler)
	mux.HandleFunc("/simulate-service/wordpress", testingHandlers.SimulateWordPressServiceHandler)
	mux.HandleFunc("/simulate-service/nextjs", testingHandlers.SimulateNextJSServiceHandler)
	mux.HandleFunc("/simulate-trace/cross-service", testingHandlers.SimulateCrossServiceTracingHandler)

	// Integration testing endpoints
	mux.HandleFunc("/test-service-discovery", testingHandlers.TestServiceDiscoveryHandler)
	mux.HandleFunc("/test-reverse-proxy", testingHandlers.TestReverseProxyHandler)
	mux.HandleFunc("/test-ssl-monitoring", testingHandlers.TestSSLMonitoringHandler)
	mux.HandleFunc("/test-domain-health", testingHandlers.TestDomainHealthHandler)

	// LGTM Stack Configuration & Integration endpoints
	mux.HandleFunc("/test-lgtm-integration", integrationHandlers.TestLGTMIntegration)
	mux.HandleFunc("/test-grafana-dashboards", integrationHandlers.TestGrafanaDashboards)
	mux.HandleFunc("/test-alert-rules", integrationHandlers.TestAlertRules)

	// LGTM Stack Performance & Scale Testing endpoints
	mux.HandleFunc("/test-metrics-scale", performanceHandlers.TestMetricsScale)
	mux.HandleFunc("/test-logs-scale", performanceHandlers.TestLogsScale)
	mux.HandleFunc("/test-traces-scale", performanceHandlers.TestTracesScale)
	mux.HandleFunc("/test-dashboard-load", performanceHandlers.TestDashboardLoad)
	mux.HandleFunc("/test-resource-usage", performanceHandlers.TestResourceUsage)
	mux.HandleFunc("/test-storage-limits", performanceHandlers.TestStorageLimits)

	// Alerting test endpoints
	mux.HandleFunc("/test-alert-rules-legacy", alertingHandlers.TestAlertRulesHandler)
	mux.HandleFunc("/test-fire-alert", alertingHandlers.TestFireAlertHandler)
	mux.HandleFunc("/test-incident-management", alertingHandlers.TestIncidentManagementHandler)
	mux.HandleFunc("/test-notification-channels", alertingHandlers.TestNotificationChannelsHandler)
	mux.HandleFunc("/active-alerts", alertingHandlers.GetActiveAlertsHandler)
	mux.HandleFunc("/active-incidents", alertingHandlers.GetActiveIncidentsHandler)

	// Prometheus metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	// Settings and configuration API
	mux.HandleFunc("/api/settings", basicHandlers.SettingsHandler)
	mux.HandleFunc("/api/test-connection/", basicHandlers.TestConnectionHandler)

	// Simple test endpoint for HTMX debugging
	mux.HandleFunc("/test-simple", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		if _, err := w.Write([]byte("<p style='color: green;'>✅ HTMX connection working! Endpoint reached successfully.</p>")); err != nil {
			log.Printf("Error writing response: %v", err)
		}
	})

	// Configuration endpoint for frontend
	mux.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		config := map[string]interface{}{
			"api_base_url": serviceConfig.GetAPIBaseURL(),
			"version":      serviceConfig.Version,
			"environment":  serviceConfig.Environment,
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		if err := encodeJSON(w, config); err != nil {
			http.Error(w, "Failed to encode config", http.StatusInternalServerError)
		}
	})

	// Serve embedded static files from embedded filesystem
	mux.Handle("/", http.FileServer(http.Dir("./static/")))

	// Wrap mux with middleware
	wrappedMux := middleware.AddMiddleware(mux, loggingService)

	// Create server with proper timeouts
	server := &http.Server{
		Addr:           ":3001",
		Handler:        wrappedMux,
		ReadTimeout:    30 * time.Second, // Time to read request (increased)
		WriteTimeout:   20 * time.Minute, // Time to write response (much longer for performance tests)
		IdleTimeout:    60 * time.Second, // Time to keep connection alive
		MaxHeaderBytes: 1 << 20,          // 1 MB max header size
	}

	// Setup graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		fmt.Printf("Argus LGTM Stack Validator listening on port :3001\n")
		fmt.Printf("Dashboard: http://localhost:3001\n")
		fmt.Printf("API docs: http://localhost:3001/api\n")
		fmt.Printf("Health: http://localhost:3001/health\n")
		fmt.Printf("Metrics: http://localhost:3001/metrics\n")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	fmt.Printf("Argus server is ready to handle requests\n")

	// Wait for interrupt signal to gracefully shutdown the server
	<-done
	fmt.Printf("\nGracefully shutting down Argus server...\n")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	} else {
		fmt.Printf("Argus server stopped gracefully\n")
	}
}

func encodeJSON(w http.ResponseWriter, data interface{}) error {
	return json.NewEncoder(w).Encode(data)
}
