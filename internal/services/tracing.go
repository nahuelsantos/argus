package services

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/nahuelsantos/argus/internal/config"
	"github.com/nahuelsantos/argus/internal/metrics"
	"github.com/nahuelsantos/argus/internal/models"
)

// TracingService handles all tracing operations
type TracingService struct {
	config *config.TracingConfig
	tracer oteltrace.Tracer
}

// NewTracingService creates a new tracing service
func NewTracingService() *TracingService {
	return &TracingService{
		config: config.GetTracingConfig(),
	}
}

// InitTracer initializes OpenTelemetry tracing
func (ts *TracingService) InitTracer() {
	exporter, err := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithEndpoint(ts.config.JaegerEndpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		fmt.Printf("Failed to create trace exporter: %v\n", err)
		return
	}

	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(ts.config.ServiceName),
			semconv.ServiceVersionKey.String(ts.config.ServiceVersion),
		),
	)
	if err != nil {
		fmt.Printf("Failed to create resource: %v\n", err)
		return
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
		trace.WithSampler(trace.TraceIDRatioBased(ts.config.SamplingRate)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	ts.tracer = otel.Tracer(ts.config.ServiceName)
}

// GetResourceMetrics gets current resource metrics
func (ts *TracingService) GetResourceMetrics() models.ResourceMetrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return models.ResourceMetrics{
		CPUUsage:       float64(rand.Intn(100)),
		MemoryUsage:    int64(m.Alloc),
		GoroutineCount: runtime.NumGoroutine(),
		HeapSize:       int64(m.HeapAlloc),
		GCPause:        float64(m.PauseNs[(m.NumGC+255)%256]) / 1e6,
		DiskIO:         int64(rand.Intn(1000000)),
		NetworkIO:      int64(rand.Intn(1000000)),
	}
}

// CreateAPMData creates APM data for a request
func (ts *TracingService) CreateAPMData(ctx context.Context, operationName string, statusCode int, duration time.Duration) models.APMData {
	span := oteltrace.SpanFromContext(ctx)
	traceID := ""
	spanID := ""

	if span.SpanContext().IsValid() {
		traceID = span.SpanContext().TraceID().String()
		spanID = span.SpanContext().SpanID().String()
	}

	return models.APMData{
		ServiceName:   ts.config.ServiceName,
		TraceID:       traceID,
		SpanID:        spanID,
		OperationName: operationName,
		StartTime:     time.Now().Add(-duration),
		Duration:      duration,
		StatusCode:    statusCode,
		ResourceUsage: ts.GetResourceMetrics(),
		Dependencies:  ts.generateDependencies(operationName),
		CustomTags: map[string]string{
			"environment": "development",
			"version":     ts.config.ServiceVersion,
		},
	}
}

// GenerateDependencies generates mock service dependencies
func (ts *TracingService) generateDependencies(operation string) []models.ServiceDependency {
	dependencies := []models.ServiceDependency{}

	switch operation {
	case "user_authentication":
		dependencies = append(dependencies, models.ServiceDependency{
			ServiceName:  "auth-service",
			Operation:    "validate_token",
			ResponseTime: time.Duration(rand.Intn(50)+10) * time.Millisecond,
			StatusCode:   200,
			ErrorRate:    0.02,
			RequestCount: int64(rand.Intn(1000) + 100),
			Dependencies: []string{"user-db", "redis-cache"},
			CustomAttributes: map[string]string{
				"auth_method": "jwt",
				"cache_hit":   "true",
			},
		})
	case "data_processing":
		dependencies = append(dependencies, models.ServiceDependency{
			ServiceName:  "database-service",
			Operation:    "query_data",
			ResponseTime: time.Duration(rand.Intn(200)+50) * time.Millisecond,
			StatusCode:   200,
			ErrorRate:    0.01,
			RequestCount: int64(rand.Intn(500) + 50),
			Dependencies: []string{"postgres-db"},
			CustomAttributes: map[string]string{
				"query_type": "select",
				"table":      "users",
			},
		})
	case "api_gateway":
		dependencies = append(dependencies, models.ServiceDependency{
			ServiceName:  "rate-limiter",
			Operation:    "check_limits",
			ResponseTime: time.Duration(rand.Intn(10)+1) * time.Millisecond,
			StatusCode:   200,
			ErrorRate:    0.001,
			RequestCount: int64(rand.Intn(2000) + 500),
			Dependencies: []string{"redis-cache"},
			CustomAttributes: map[string]string{
				"limit_type": "per_user",
				"remaining":  "95",
			},
		})
	}

	return dependencies
}

// LogAPMData logs APM data with metrics
func (ts *TracingService) LogAPMData(apmData models.APMData) {
	status := "success"
	if apmData.StatusCode >= 400 {
		status = "error"
	}

	metrics.APMTracesTotal.WithLabelValues(
		apmData.ServiceName,
		apmData.OperationName,
		status,
	).Inc()

	metrics.APMSpanDuration.WithLabelValues(
		apmData.ServiceName,
		apmData.OperationName,
	).Observe(apmData.Duration.Seconds())

	for _, dep := range apmData.Dependencies {
		metrics.ServiceDependencyLatency.WithLabelValues(
			apmData.ServiceName,
			dep.ServiceName,
			dep.Operation,
		).Observe(dep.ResponseTime.Seconds())
	}

	ts.detectPerformanceAnomalies(apmData.OperationName, apmData.Duration, apmData.ResourceUsage)
}

// DetectPerformanceAnomalies detects performance anomalies
func (ts *TracingService) detectPerformanceAnomalies(operation string, duration time.Duration, resourceUsage models.ResourceMetrics) {
	// High latency detection
	if duration > 5*time.Second {
		metrics.PerformanceAnomalies.WithLabelValues(
			ts.config.ServiceName,
			operation,
			"high_latency",
		).Inc()
	}

	// High memory usage detection
	if resourceUsage.MemoryUsage > 1024*1024*1024 { // 1GB
		metrics.PerformanceAnomalies.WithLabelValues(
			ts.config.ServiceName,
			operation,
			"high_memory",
		).Inc()
	}

	// High CPU usage detection
	if resourceUsage.CPUUsage > 80 {
		metrics.PerformanceAnomalies.WithLabelValues(
			ts.config.ServiceName,
			operation,
			"high_cpu",
		).Inc()
	}

	// Too many goroutines detection
	if resourceUsage.GoroutineCount > 1000 {
		metrics.PerformanceAnomalies.WithLabelValues(
			ts.config.ServiceName,
			operation,
			"goroutine_leak",
		).Inc()
	}
}

// SimulateServiceCall simulates a service call with tracing
func (ts *TracingService) SimulateServiceCall(ctx context.Context, serviceName string, duration time.Duration) {
	newCtx, span := ts.tracer.Start(ctx, fmt.Sprintf("call_%s", serviceName))
	defer span.End()

	span.SetAttributes(
		attribute.String("service.name", serviceName),
		attribute.String("operation.type", "external_call"),
		attribute.Int64("duration.ms", duration.Milliseconds()),
	)

	// Simulate work
	time.Sleep(duration)

	span.SetAttributes(
		attribute.String("result", "success"),
		attribute.Int("status.code", 200),
	)

	// Use the context for any child operations
	_ = newCtx
}

// CreateChildSpan creates a child span for operations
func (ts *TracingService) CreateChildSpan(ctx context.Context, operationName string, duration time.Duration) string {
	newCtx, span := ts.tracer.Start(ctx, operationName)
	defer span.End()

	span.SetAttributes(
		attribute.String("operation.name", operationName),
		attribute.Int64("duration.ms", duration.Milliseconds()),
		attribute.String("span.kind", "internal"),
	)

	// Simulate work
	time.Sleep(duration)

	// Use the context for any child operations
	_ = newCtx

	return span.SpanContext().SpanID().String()
}
