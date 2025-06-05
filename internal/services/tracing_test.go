package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"

	"github.com/nahuelsantos/argus/internal/models"
)

func TestNewTracingService(t *testing.T) {
	ts := NewTracingService()

	assert.NotNil(t, ts)
	assert.NotNil(t, ts.config)
	assert.Equal(t, "argus", ts.config.ServiceName)
	assert.NotEmpty(t, ts.config.ServiceVersion)
	assert.Equal(t, "http://localhost:14268/api/traces", ts.config.JaegerEndpoint)
	assert.Equal(t, 1.0, ts.config.SamplingRate)
}

func TestTracingService_InitTracer(t *testing.T) {
	ts := NewTracingService()

	// Test that InitTracer doesn't panic
	assert.NotPanics(t, func() {
		ts.InitTracer()
	})

	// After initialization, tracer should be set
	assert.NotNil(t, ts.tracer)

	// Verify OpenTelemetry global tracer provider is set
	provider := otel.GetTracerProvider()
	assert.NotNil(t, provider)

	// Verify we can get a tracer
	tracer := provider.Tracer("test")
	assert.NotNil(t, tracer)
}

func TestTracingService_GetResourceMetrics(t *testing.T) {
	ts := NewTracingService()

	metrics := ts.GetResourceMetrics()

	t.Run("resource metrics have valid values", func(t *testing.T) {
		// CPU usage should be 0-100
		assert.GreaterOrEqual(t, metrics.CPUUsage, 0.0)
		assert.LessOrEqual(t, metrics.CPUUsage, 100.0)

		// Memory usage should be positive
		assert.Greater(t, metrics.MemoryUsage, int64(0))

		// Goroutine count should be positive
		assert.Greater(t, metrics.GoroutineCount, 0)

		// Heap size should be positive
		assert.Greater(t, metrics.HeapSize, int64(0))

		// GC pause should be non-negative
		assert.GreaterOrEqual(t, metrics.GCPause, 0.0)

		// Disk and Network IO should be non-negative
		assert.GreaterOrEqual(t, metrics.DiskIO, int64(0))
		assert.GreaterOrEqual(t, metrics.NetworkIO, int64(0))
	})

	t.Run("multiple calls return different random values", func(t *testing.T) {
		metrics1 := ts.GetResourceMetrics()
		metrics2 := ts.GetResourceMetrics()

		// Random values should likely be different (CPU, DiskIO, NetworkIO)
		// Note: There's a small chance they could be the same, but very unlikely
		differentValues := metrics1.CPUUsage != metrics2.CPUUsage ||
			metrics1.DiskIO != metrics2.DiskIO ||
			metrics1.NetworkIO != metrics2.NetworkIO

		assert.True(t, differentValues, "Random metrics should vary between calls")
	})
}

func TestTracingService_CreateAPMData(t *testing.T) {
	ts := NewTracingService()
	ts.InitTracer()

	tests := []struct {
		name          string
		operationName string
		statusCode    int
		duration      time.Duration
	}{
		{
			name:          "successful operation",
			operationName: "test_operation",
			statusCode:    200,
			duration:      100 * time.Millisecond,
		},
		{
			name:          "error operation",
			operationName: "failing_operation",
			statusCode:    500,
			duration:      250 * time.Millisecond,
		},
		{
			name:          "long operation",
			operationName: "slow_operation",
			statusCode:    200,
			duration:      2 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			apmData := ts.CreateAPMData(ctx, tt.operationName, tt.statusCode, tt.duration)

			// Basic field validation
			assert.Equal(t, ts.config.ServiceName, apmData.ServiceName)
			assert.Equal(t, tt.operationName, apmData.OperationName)
			assert.Equal(t, tt.statusCode, apmData.StatusCode)
			assert.Equal(t, tt.duration, apmData.Duration)

			// Time validation
			expectedStartTime := time.Now().Add(-tt.duration)
			assert.WithinDuration(t, expectedStartTime, apmData.StartTime, time.Second)

			// Resource usage validation
			assert.NotNil(t, apmData.ResourceUsage)
			assert.Greater(t, apmData.ResourceUsage.MemoryUsage, int64(0))
			assert.Greater(t, apmData.ResourceUsage.GoroutineCount, 0)

			// Dependencies validation
			assert.NotNil(t, apmData.Dependencies)

			// Custom tags validation
			assert.NotNil(t, apmData.CustomTags)
			assert.Equal(t, "development", apmData.CustomTags["environment"])
			assert.Equal(t, ts.config.ServiceVersion, apmData.CustomTags["version"])
		})
	}
}

func TestTracingService_CreateAPMDataWithSpan(t *testing.T) {
	ts := NewTracingService()
	ts.InitTracer()

	t.Run("with valid span context", func(t *testing.T) {
		ctx, span := ts.tracer.Start(context.Background(), "test_span")
		defer span.End()

		apmData := ts.CreateAPMData(ctx, "test_operation", 200, 100*time.Millisecond)

		// Should have trace and span IDs when there's a valid span
		assert.NotEmpty(t, apmData.TraceID)
		assert.NotEmpty(t, apmData.SpanID)
		assert.Len(t, apmData.TraceID, 32) // Trace ID should be 32 hex characters
		assert.Len(t, apmData.SpanID, 16)  // Span ID should be 16 hex characters
	})

	t.Run("without span context", func(t *testing.T) {
		ctx := context.Background()

		apmData := ts.CreateAPMData(ctx, "test_operation", 200, 100*time.Millisecond)

		// Should have empty trace and span IDs when there's no span
		assert.Empty(t, apmData.TraceID)
		assert.Empty(t, apmData.SpanID)
	})
}

func TestTracingService_GenerateDependencies(t *testing.T) {
	ts := NewTracingService()

	tests := []struct {
		name              string
		operation         string
		expectedDepCount  int
		expectedService   string
		expectedOperation string
		expectedDeps      []string
	}{
		{
			name:              "user authentication",
			operation:         "user_authentication",
			expectedDepCount:  1,
			expectedService:   "auth-service",
			expectedOperation: "validate_token",
			expectedDeps:      []string{"user-db", "redis-cache"},
		},
		{
			name:              "data processing",
			operation:         "data_processing",
			expectedDepCount:  1,
			expectedService:   "database-service",
			expectedOperation: "query_data",
			expectedDeps:      []string{"postgres-db"},
		},
		{
			name:              "api gateway",
			operation:         "api_gateway",
			expectedDepCount:  1,
			expectedService:   "rate-limiter",
			expectedOperation: "check_limits",
			expectedDeps:      []string{"redis-cache"},
		},
		{
			name:             "unknown operation",
			operation:        "unknown_operation",
			expectedDepCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := ts.generateDependencies(tt.operation)

			assert.Len(t, deps, tt.expectedDepCount)

			if tt.expectedDepCount > 0 {
				dep := deps[0]
				assert.Equal(t, tt.expectedService, dep.ServiceName)
				assert.Equal(t, tt.expectedOperation, dep.Operation)
				assert.Equal(t, 200, dep.StatusCode)
				assert.Greater(t, dep.ResponseTime, 0*time.Millisecond)
				assert.Greater(t, dep.ErrorRate, 0.0)
				assert.Greater(t, dep.RequestCount, int64(0))
				assert.Equal(t, tt.expectedDeps, dep.Dependencies)
				assert.NotNil(t, dep.CustomAttributes)
				assert.NotEmpty(t, dep.CustomAttributes)
			}
		})
	}
}

func TestTracingService_GenerateDependenciesVariability(t *testing.T) {
	ts := NewTracingService()

	// Test that random values vary between calls
	deps1 := ts.generateDependencies("user_authentication")
	deps2 := ts.generateDependencies("user_authentication")

	require.Len(t, deps1, 1)
	require.Len(t, deps2, 1)

	// Response times should likely be different (random values)
	differentValues := deps1[0].ResponseTime != deps2[0].ResponseTime ||
		deps1[0].RequestCount != deps2[0].RequestCount

	assert.True(t, differentValues, "Dependencies should have variable random values")
}

func TestTracingService_LogAPMData(t *testing.T) {
	ts := NewTracingService()

	tests := []struct {
		name       string
		statusCode int
		duration   time.Duration
		memUsage   int64
		cpuUsage   float64
		goroutines int
	}{
		{
			name:       "successful request",
			statusCode: 200,
			duration:   100 * time.Millisecond,
			memUsage:   500 * 1024 * 1024, // 500MB
			cpuUsage:   50.0,
			goroutines: 100,
		},
		{
			name:       "error request",
			statusCode: 500,
			duration:   200 * time.Millisecond,
			memUsage:   800 * 1024 * 1024, // 800MB
			cpuUsage:   75.0,
			goroutines: 200,
		},
		{
			name:       "high latency request",
			statusCode: 200,
			duration:   6 * time.Second,    // Above anomaly threshold
			memUsage:   1200 * 1024 * 1024, // 1.2GB - Above threshold
			cpuUsage:   85.0,               // Above threshold
			goroutines: 1200,               // Above threshold
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apmData := models.APMData{
				ServiceName:   "test-service",
				OperationName: "test-operation",
				StatusCode:    tt.statusCode,
				Duration:      tt.duration,
				ResourceUsage: models.ResourceMetrics{
					MemoryUsage:    tt.memUsage,
					CPUUsage:       tt.cpuUsage,
					GoroutineCount: tt.goroutines,
				},
				Dependencies: []models.ServiceDependency{
					{
						ServiceName:  "test-dep",
						Operation:    "test-op",
						ResponseTime: 50 * time.Millisecond,
					},
				},
			}

			// Should not panic
			assert.NotPanics(t, func() {
				ts.LogAPMData(apmData)
			})
		})
	}
}

func TestTracingService_DetectPerformanceAnomalies(t *testing.T) {
	ts := NewTracingService()

	tests := []struct {
		name            string
		operation       string
		duration        time.Duration
		resourceMetrics models.ResourceMetrics
		expectAnomalies []string
	}{
		{
			name:      "normal performance",
			operation: "normal_op",
			duration:  100 * time.Millisecond,
			resourceMetrics: models.ResourceMetrics{
				MemoryUsage:    500 * 1024 * 1024, // 500MB
				CPUUsage:       50.0,
				GoroutineCount: 100,
			},
			expectAnomalies: []string{}, // No anomalies
		},
		{
			name:      "high latency",
			operation: "slow_op",
			duration:  6 * time.Second, // > 5 seconds
			resourceMetrics: models.ResourceMetrics{
				MemoryUsage:    500 * 1024 * 1024,
				CPUUsage:       50.0,
				GoroutineCount: 100,
			},
			expectAnomalies: []string{"high_latency"},
		},
		{
			name:      "high memory usage",
			operation: "memory_intensive_op",
			duration:  100 * time.Millisecond,
			resourceMetrics: models.ResourceMetrics{
				MemoryUsage:    2 * 1024 * 1024 * 1024, // 2GB > 1GB
				CPUUsage:       50.0,
				GoroutineCount: 100,
			},
			expectAnomalies: []string{"high_memory"},
		},
		{
			name:      "high CPU usage",
			operation: "cpu_intensive_op",
			duration:  100 * time.Millisecond,
			resourceMetrics: models.ResourceMetrics{
				MemoryUsage:    500 * 1024 * 1024,
				CPUUsage:       85.0, // > 80
				GoroutineCount: 100,
			},
			expectAnomalies: []string{"high_cpu"},
		},
		{
			name:      "goroutine leak",
			operation: "leaky_op",
			duration:  100 * time.Millisecond,
			resourceMetrics: models.ResourceMetrics{
				MemoryUsage:    500 * 1024 * 1024,
				CPUUsage:       50.0,
				GoroutineCount: 1500, // > 1000
			},
			expectAnomalies: []string{"goroutine_leak"},
		},
		{
			name:      "multiple anomalies",
			operation: "problematic_op",
			duration:  6 * time.Second,
			resourceMetrics: models.ResourceMetrics{
				MemoryUsage:    2 * 1024 * 1024 * 1024, // High memory
				CPUUsage:       90.0,                   // High CPU
				GoroutineCount: 1200,                   // Goroutine leak
			},
			expectAnomalies: []string{"high_latency", "high_memory", "high_cpu", "goroutine_leak"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This method updates metrics, so we just ensure it doesn't panic
			assert.NotPanics(t, func() {
				ts.detectPerformanceAnomalies(tt.operation, tt.duration, tt.resourceMetrics)
			})
		})
	}
}

func TestTracingService_SimulateServiceCall(t *testing.T) {
	ts := NewTracingService()
	ts.InitTracer()

	tests := []struct {
		name        string
		serviceName string
		duration    time.Duration
	}{
		{
			name:        "fast service call",
			serviceName: "fast-service",
			duration:    10 * time.Millisecond,
		},
		{
			name:        "slow service call",
			serviceName: "slow-service",
			duration:    100 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			start := time.Now()

			assert.NotPanics(t, func() {
				ts.SimulateServiceCall(ctx, tt.serviceName, tt.duration)
			})

			elapsed := time.Since(start)
			// Should take at least the specified duration (due to sleep)
			assert.GreaterOrEqual(t, elapsed, tt.duration)
			// Should not take much longer (within reason)
			assert.LessOrEqual(t, elapsed, tt.duration+50*time.Millisecond)
		})
	}
}

func TestTracingService_CreateChildSpan(t *testing.T) {
	ts := NewTracingService()
	ts.InitTracer()

	tests := []struct {
		name          string
		operationName string
		duration      time.Duration
	}{
		{
			name:          "fast child span",
			operationName: "fast_operation",
			duration:      10 * time.Millisecond,
		},
		{
			name:          "slow child span",
			operationName: "slow_operation",
			duration:      50 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			start := time.Now()

			var spanID string
			assert.NotPanics(t, func() {
				spanID = ts.CreateChildSpan(ctx, tt.operationName, tt.duration)
			})

			elapsed := time.Since(start)

			// Should return a valid span ID
			assert.NotEmpty(t, spanID)
			assert.Len(t, spanID, 16) // Span ID should be 16 hex characters

			// Should take at least the specified duration (due to sleep)
			assert.GreaterOrEqual(t, elapsed, tt.duration)
			// Should not take much longer (within reason)
			assert.LessOrEqual(t, elapsed, tt.duration+30*time.Millisecond)
		})
	}
}

func TestTracingService_CreateChildSpanWithParent(t *testing.T) {
	ts := NewTracingService()
	ts.InitTracer()

	t.Run("child span with parent context", func(t *testing.T) {
		// Create parent span
		parentCtx, parentSpan := ts.tracer.Start(context.Background(), "parent_operation")
		defer parentSpan.End()

		// Create child span
		spanID := ts.CreateChildSpan(parentCtx, "child_operation", 10*time.Millisecond)

		assert.NotEmpty(t, spanID)
		assert.Len(t, spanID, 16)

		// Verify parent span context is valid
		assert.True(t, parentSpan.SpanContext().IsValid())
	})
}

// Benchmark tests
func BenchmarkTracingService_GetResourceMetrics(b *testing.B) {
	ts := NewTracingService()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ts.GetResourceMetrics()
	}
}

func BenchmarkTracingService_CreateAPMData(b *testing.B) {
	ts := NewTracingService()
	ts.InitTracer()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ts.CreateAPMData(ctx, "benchmark_operation", 200, 100*time.Millisecond)
	}
}

func BenchmarkTracingService_GenerateDependencies(b *testing.B) {
	ts := NewTracingService()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ts.generateDependencies("user_authentication")
	}
}

// Example usage
func ExampleTracingService_CreateAPMData() {
	ts := NewTracingService()
	ts.InitTracer()

	ctx := context.Background()
	apmData := ts.CreateAPMData(ctx, "user_login", 200, 150*time.Millisecond)

	_ = apmData.ServiceName   // "argus"
	_ = apmData.OperationName // "user_login"
	_ = apmData.StatusCode    // 200
	_ = apmData.Duration      // 150ms
}
