package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/nahuelsantos/argus/internal/models"
)

func TestNewLoggingService(t *testing.T) {
	ls := NewLoggingService()

	assert.NotNil(t, ls)
	assert.NotNil(t, ls.config)
	assert.Equal(t, "argus", ls.config.Name)
	assert.NotEmpty(t, ls.config.Version)
	assert.NotEmpty(t, ls.config.Environment)
}

func TestLoggingService_InitLogger(t *testing.T) {
	ls := NewLoggingService()

	// Test that InitLogger doesn't panic
	assert.NotPanics(t, func() {
		ls.InitLogger()
	})

	// After initialization, global logger should be set
	assert.NotNil(t, logger)
}

func TestLoggingService_GenerateNodeID(t *testing.T) {
	ls := NewLoggingService()

	t.Run("generates valid node ID", func(t *testing.T) {
		nodeID := ls.GenerateNodeID()

		assert.NotEmpty(t, nodeID)
		assert.True(t, strings.HasPrefix(nodeID, "node-"))
		assert.Len(t, nodeID, 13) // "node-" + 8 characters
	})

	t.Run("generates unique node IDs", func(t *testing.T) {
		nodeID1 := ls.GenerateNodeID()
		nodeID2 := ls.GenerateNodeID()

		assert.NotEqual(t, nodeID1, nodeID2)
	})

	t.Run("node ID format is consistent", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			nodeID := ls.GenerateNodeID()
			assert.Regexp(t, `^node-[a-f0-9]{8}$`, nodeID)
		}
	})
}

func TestLoggingService_CreateLogContext(t *testing.T) {
	ls := NewLoggingService()

	tests := []struct {
		name           string
		setupRequest   func() *http.Request
		expectedFields func(t *testing.T, ctx models.LogContext)
	}{
		{
			name: "basic request without headers",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				return req
			},
			expectedFields: func(t *testing.T, ctx models.LogContext) {
				assert.NotEmpty(t, ctx.RequestID)
				assert.Empty(t, ctx.TraceID)
				assert.Empty(t, ctx.SpanID)
				assert.Empty(t, ctx.UserID)
				assert.Empty(t, ctx.SessionID)
				assert.Equal(t, "argus", ctx.ServiceName)
				assert.NotEmpty(t, ctx.Version)
				assert.NotEmpty(t, ctx.Environment)
			},
		},
		{
			name: "request with X-Request-ID header",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("X-Request-ID", "test-request-id")
				return req
			},
			expectedFields: func(t *testing.T, ctx models.LogContext) {
				assert.Equal(t, "test-request-id", ctx.RequestID)
			},
		},
		{
			name: "request with user and session headers",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("X-User-ID", "user-123")
				req.Header.Set("X-Session-ID", "session-456")
				return req
			},
			expectedFields: func(t *testing.T, ctx models.LogContext) {
				assert.Equal(t, "user-123", ctx.UserID)
				assert.Equal(t, "session-456", ctx.SessionID)
			},
		},
		{
			name: "request with correlation ID",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("X-Correlation-ID", "corr-789")
				return req
			},
			expectedFields: func(t *testing.T, ctx models.LogContext) {
				assert.Equal(t, "corr-789", ctx.RequestID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupRequest()
			ctx := ls.CreateLogContext(req)

			tt.expectedFields(t, ctx)
		})
	}
}

func TestLoggingService_GetOrCreateRequestID(t *testing.T) {
	ls := NewLoggingService()

	tests := []struct {
		name     string
		setupReq func() *http.Request
		expected string
	}{
		{
			name: "returns X-Request-ID header",
			setupReq: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("X-Request-ID", "header-request-id")
				return req
			},
			expected: "header-request-id",
		},
		{
			name: "returns X-Correlation-ID when no X-Request-ID",
			setupReq: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("X-Correlation-ID", "correlation-id")
				return req
			},
			expected: "correlation-id",
		},
		{
			name: "returns context value when no headers",
			setupReq: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				ctx := context.WithValue(req.Context(), models.RequestIDKey, "context-request-id")
				return req.WithContext(ctx)
			},
			expected: "context-request-id",
		},
		{
			name: "generates UUID when nothing available",
			setupReq: func() *http.Request {
				return httptest.NewRequest("GET", "/test", nil)
			},
			expected: "", // Will be a UUID, so we'll validate format instead
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupReq()
			requestID := ls.getOrCreateRequestID(req)

			if tt.expected == "" {
				// Should be a valid UUID
				_, err := uuid.Parse(requestID)
				assert.NoError(t, err)
			} else {
				assert.Equal(t, tt.expected, requestID)
			}
		})
	}
}

func TestLoggingService_ExtractTraceAndSpanID(t *testing.T) {
	ls := NewLoggingService()

	t.Run("extracts trace and span ID from valid context", func(t *testing.T) {
		// Set up OpenTelemetry
		ts := NewTracingService()
		ts.InitTracer()

		ctx, span := ts.tracer.Start(context.Background(), "test_span")
		defer span.End()

		traceID := ls.extractTraceID(ctx)
		spanID := ls.extractSpanID(ctx)

		assert.NotEmpty(t, traceID)
		assert.NotEmpty(t, spanID)
		assert.Len(t, traceID, 32) // Trace ID should be 32 hex characters
		assert.Len(t, spanID, 16)  // Span ID should be 16 hex characters
	})

	t.Run("returns empty strings for invalid context", func(t *testing.T) {
		ctx := context.Background()

		traceID := ls.extractTraceID(ctx)
		spanID := ls.extractSpanID(ctx)

		assert.Empty(t, traceID)
		assert.Empty(t, spanID)
	})
}

func TestLoggingService_ExtractUserAndSessionID(t *testing.T) {
	ls := NewLoggingService()

	tests := []struct {
		name           string
		setupReq       func() *http.Request
		expectedUserID string
		expectedSessID string
	}{
		{
			name: "extracts from headers",
			setupReq: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("X-User-ID", "user-from-header")
				req.Header.Set("X-Session-ID", "session-from-header")
				return req
			},
			expectedUserID: "user-from-header",
			expectedSessID: "session-from-header",
		},
		{
			name: "extracts from context when no headers",
			setupReq: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				ctx := context.WithValue(req.Context(), models.UserIDKey, "user-from-context")
				ctx = context.WithValue(ctx, models.SessionIDKey, "session-from-context")
				return req.WithContext(ctx)
			},
			expectedUserID: "user-from-context",
			expectedSessID: "session-from-context",
		},
		{
			name: "returns empty when nothing available",
			setupReq: func() *http.Request {
				return httptest.NewRequest("GET", "/test", nil)
			},
			expectedUserID: "",
			expectedSessID: "",
		},
		{
			name: "prefers headers over context",
			setupReq: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("X-User-ID", "user-from-header")
				ctx := context.WithValue(req.Context(), models.UserIDKey, "user-from-context")
				return req.WithContext(ctx)
			},
			expectedUserID: "user-from-header",
			expectedSessID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupReq()

			userID := ls.extractUserID(req)
			sessionID := ls.extractSessionID(req)

			assert.Equal(t, tt.expectedUserID, userID)
			assert.Equal(t, tt.expectedSessID, sessionID)
		})
	}
}

func TestLoggingService_LogWithContext(t *testing.T) {
	ls := NewLoggingService()
	ls.InitLogger()

	tests := []struct {
		name           string
		level          zapcore.Level
		ctx            context.Context
		message        string
		fields         []zap.Field
		shouldNotPanic bool
	}{
		{
			name:           "debug level",
			level:          zapcore.DebugLevel,
			ctx:            context.Background(),
			message:        "debug message",
			fields:         []zap.Field{zap.String("key", "value")},
			shouldNotPanic: true,
		},
		{
			name:           "info level",
			level:          zapcore.InfoLevel,
			ctx:            context.Background(),
			message:        "info message",
			fields:         []zap.Field{zap.Int("count", 42)},
			shouldNotPanic: true,
		},
		{
			name:           "warn level",
			level:          zapcore.WarnLevel,
			ctx:            context.Background(),
			message:        "warning message",
			fields:         []zap.Field{},
			shouldNotPanic: true,
		},
		{
			name:           "error level",
			level:          zapcore.ErrorLevel,
			ctx:            context.Background(),
			message:        "error message",
			fields:         []zap.Field{zap.Error(assert.AnError)},
			shouldNotPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldNotPanic {
				assert.NotPanics(t, func() {
					ls.LogWithContext(tt.level, tt.ctx, tt.message, tt.fields...)
				})
			}
		})
	}
}

func TestLoggingService_LogWithContextFields(t *testing.T) {
	ls := NewLoggingService()
	ls.InitLogger()

	t.Run("includes context fields", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), models.RequestIDKey, "test-request-id")

		// Should not panic and should include context fields
		assert.NotPanics(t, func() {
			ls.LogWithContext(zapcore.InfoLevel, ctx, "test message")
		})
	})
}

func TestLoggingService_LogBusinessEvent(t *testing.T) {
	ls := NewLoggingService()
	ls.InitLogger()

	tests := []struct {
		name      string
		eventType string
		data      map[string]interface{}
	}{
		{
			name:      "user registration event",
			eventType: "user_registration",
			data: map[string]interface{}{
				"user_id": "123",
				"email":   "test@example.com",
			},
		},
		{
			name:      "order created event",
			eventType: "order_created",
			data: map[string]interface{}{
				"order_id": "order-456",
				"amount":   99.99,
				"currency": "USD",
			},
		},
		{
			name:      "event with nil data",
			eventType: "simple_event",
			data:      nil,
		},
		{
			name:      "event with empty data",
			eventType: "empty_event",
			data:      map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				ls.LogBusinessEvent(tt.eventType, tt.data)
			})
		})
	}
}

func TestLoggingService_LogPerformance(t *testing.T) {
	ls := NewLoggingService()
	ls.InitLogger()

	tests := []struct {
		name           string
		operation      string
		duration       time.Duration
		additionalData map[string]interface{}
	}{
		{
			name:      "fast operation",
			operation: "fast_operation",
			duration:  10 * time.Millisecond,
			additionalData: map[string]interface{}{
				"cache_hit":  true,
				"db_queries": 1,
			},
		},
		{
			name:      "slow operation",
			operation: "slow_operation",
			duration:  2 * time.Second,
			additionalData: map[string]interface{}{
				"cache_hit":  false,
				"db_queries": 5,
			},
		},
		{
			name:           "operation with no additional data",
			operation:      "simple_operation",
			duration:       100 * time.Millisecond,
			additionalData: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				ls.LogPerformance(tt.operation, tt.duration, tt.additionalData)
			})
		})
	}
}

func TestLoggingService_LogError(t *testing.T) {
	ls := NewLoggingService()
	ls.InitLogger()

	tests := []struct {
		name           string
		ctx            context.Context
		errorType      string
		errorCode      string
		message        string
		err            error
		additionalData map[string]interface{}
	}{
		{
			name:      "database error",
			ctx:       context.Background(),
			errorType: "database_error",
			errorCode: "DB001",
			message:   "Connection timeout",
			err:       assert.AnError,
			additionalData: map[string]interface{}{
				"query":   "SELECT * FROM users",
				"timeout": "5s",
			},
		},
		{
			name:      "validation error",
			ctx:       context.Background(),
			errorType: "validation_error",
			errorCode: "VAL001",
			message:   "Invalid email format",
			err:       nil,
			additionalData: map[string]interface{}{
				"field": "email",
				"value": "invalid-email",
			},
		},
		{
			name:           "error with context",
			ctx:            context.WithValue(context.Background(), models.RequestIDKey, "error-request-id"),
			errorType:      "api_error",
			errorCode:      "API001",
			message:        "Rate limit exceeded",
			err:            assert.AnError,
			additionalData: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				ls.LogError(tt.ctx, tt.errorType, tt.errorCode, tt.message, tt.err, tt.additionalData)
			})
		})
	}
}

func TestLoggingService_GetRequestIDFromContext(t *testing.T) {
	ls := NewLoggingService()

	tests := []struct {
		name     string
		ctx      context.Context
		expected string
	}{
		{
			name:     "returns request ID from context",
			ctx:      context.WithValue(context.Background(), models.RequestIDKey, "context-request-id"),
			expected: "context-request-id",
		},
		{
			name:     "generates UUID when no request ID in context",
			ctx:      context.Background(),
			expected: "", // Will be a UUID, validate format instead
		},
		{
			name:     "handles non-string value in context",
			ctx:      context.WithValue(context.Background(), models.RequestIDKey, 12345),
			expected: "", // Will be a UUID, validate format instead
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestID := ls.getRequestIDFromContext(tt.ctx)

			if tt.expected == "" {
				// Should be a valid UUID
				_, err := uuid.Parse(requestID)
				assert.NoError(t, err)
			} else {
				assert.Equal(t, tt.expected, requestID)
			}
		})
	}
}

func TestLoggingService_Integration(t *testing.T) {
	ls := NewLoggingService()
	ls.InitLogger()

	t.Run("full logging workflow", func(t *testing.T) {
		// Create a request with context
		req := httptest.NewRequest("POST", "/api/users", nil)
		req.Header.Set("X-Request-ID", "integration-test-id")
		req.Header.Set("X-User-ID", "user-123")

		// Create log context
		logCtx := ls.CreateLogContext(req)
		assert.Equal(t, "integration-test-id", logCtx.RequestID)
		assert.Equal(t, "user-123", logCtx.UserID)

		// Log with context
		ctx := context.WithValue(req.Context(), models.RequestIDKey, logCtx.RequestID)
		assert.NotPanics(t, func() {
			ls.LogWithContext(zapcore.InfoLevel, ctx, "Processing user request")
		})

		// Log business event
		assert.NotPanics(t, func() {
			ls.LogBusinessEvent("user_action", map[string]interface{}{
				"action":  "create_user",
				"user_id": logCtx.UserID,
			})
		})

		// Log performance
		assert.NotPanics(t, func() {
			ls.LogPerformance("create_user", 150*time.Millisecond, map[string]interface{}{
				"db_calls": 2,
			})
		})

		// Log error
		assert.NotPanics(t, func() {
			ls.LogError(ctx, "validation_error", "VAL001", "Invalid input", assert.AnError, nil)
		})
	})
}

// Benchmark tests
func BenchmarkLoggingService_GenerateNodeID(b *testing.B) {
	ls := NewLoggingService()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ls.GenerateNodeID()
	}
}

func BenchmarkLoggingService_CreateLogContext(b *testing.B) {
	ls := NewLoggingService()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "bench-request-id")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ls.CreateLogContext(req)
	}
}

func BenchmarkLoggingService_LogWithContext(b *testing.B) {
	ls := NewLoggingService()
	ls.InitLogger()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ls.LogWithContext(zapcore.InfoLevel, ctx, "benchmark message")
	}
}

func BenchmarkLoggingService_LogBusinessEvent(b *testing.B) {
	ls := NewLoggingService()
	ls.InitLogger()
	data := map[string]interface{}{"key": "value"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ls.LogBusinessEvent("benchmark_event", data)
	}
}

// Example usage
func ExampleLoggingService_CreateLogContext() {
	ls := NewLoggingService()

	req := httptest.NewRequest("GET", "/api/users", nil)
	req.Header.Set("X-Request-ID", "req-123")
	req.Header.Set("X-User-ID", "user-456")

	logCtx := ls.CreateLogContext(req)

	_ = logCtx.RequestID   // "req-123"
	_ = logCtx.UserID      // "user-456"
	_ = logCtx.ServiceName // "argus"
}
