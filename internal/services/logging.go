package services

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/nahuelsantos/argus/internal/config"
	"github.com/nahuelsantos/argus/internal/metrics"
	"github.com/nahuelsantos/argus/internal/models"
)

var logger *zap.Logger

// LoggingService handles all logging operations
type LoggingService struct {
	config *config.ServiceConfig
}

// NewLoggingService creates a new logging service
func NewLoggingService() *LoggingService {
	return &LoggingService{
		config: config.GetServiceConfig(),
	}
}

// InitLogger initializes the global logger
func (ls *LoggingService) InitLogger() {
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.CallerKey = "caller"
	config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	var err error
	logger, err = config.Build()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
}

// GenerateNodeID generates a unique node identifier
func (ls *LoggingService) GenerateNodeID() string {
	return fmt.Sprintf("node-%s", uuid.New().String()[:8])
}

// CreateLogContext creates a log context from HTTP request
func (ls *LoggingService) CreateLogContext(r *http.Request) models.LogContext {
	return models.LogContext{
		RequestID:   ls.getOrCreateRequestID(r),
		TraceID:     ls.extractTraceID(r.Context()),
		SpanID:      ls.extractSpanID(r.Context()),
		UserID:      ls.extractUserID(r),
		SessionID:   ls.extractSessionID(r),
		ServiceName: ls.config.Name,
		Version:     ls.config.Version,
		Environment: ls.config.Environment,
	}
}

// GetOrCreateRequestID gets or creates a request ID
func (ls *LoggingService) getOrCreateRequestID(r *http.Request) string {
	if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
		return requestID
	}
	if requestID := r.Header.Get("X-Correlation-ID"); requestID != "" {
		return requestID
	}
	if requestID := r.Context().Value(models.RequestIDKey); requestID != nil {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return uuid.New().String()
}

// ExtractTraceID extracts trace ID from context
func (ls *LoggingService) extractTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

// ExtractSpanID extracts span ID from context
func (ls *LoggingService) extractSpanID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().SpanID().String()
	}
	return ""
}

// ExtractUserID extracts user ID from request
func (ls *LoggingService) extractUserID(r *http.Request) string {
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		return userID
	}
	if userID := r.Context().Value(models.UserIDKey); userID != nil {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}

// ExtractSessionID extracts session ID from request
func (ls *LoggingService) extractSessionID(r *http.Request) string {
	if sessionID := r.Header.Get("X-Session-ID"); sessionID != "" {
		return sessionID
	}
	if sessionID := r.Context().Value(models.SessionIDKey); sessionID != nil {
		if id, ok := sessionID.(string); ok {
			return id
		}
	}
	return ""
}

// LogWithContext logs with structured context
func (ls *LoggingService) LogWithContext(level zapcore.Level, ctx context.Context, message string, fields ...zap.Field) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		metrics.LogProcessingDuration.WithLabelValues("log_with_context", level.String()).Observe(duration.Seconds())
	}()

	logContext := models.LogContext{
		RequestID:   ls.getRequestIDFromContext(ctx),
		TraceID:     ls.extractTraceID(ctx),
		SpanID:      ls.extractSpanID(ctx),
		ServiceName: ls.config.Name,
		Version:     ls.config.Version,
		Environment: ls.config.Environment,
	}

	allFields := append(fields,
		zap.String("request_id", logContext.RequestID),
		zap.String("trace_id", logContext.TraceID),
		zap.String("span_id", logContext.SpanID),
		zap.String("service_name", logContext.ServiceName),
		zap.String("version", logContext.Version),
		zap.String("environment", logContext.Environment),
	)

	switch level {
	case zapcore.DebugLevel:
		logger.Debug(message, allFields...)
	case zapcore.InfoLevel:
		logger.Info(message, allFields...)
	case zapcore.WarnLevel:
		logger.Warn(message, allFields...)
	case zapcore.ErrorLevel:
		logger.Error(message, allFields...)
	case zapcore.FatalLevel:
		logger.Fatal(message, allFields...)
	}

	metrics.LogEntriesTotal.WithLabelValues(level.String(), ls.config.Name, "").Inc()
}

// LogBusinessEvent logs business events
func (ls *LoggingService) LogBusinessEvent(eventType string, data map[string]interface{}) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		metrics.LogProcessingDuration.WithLabelValues("business_event", "info").Observe(duration.Seconds())
	}()

	entry := models.LogEntry{
		Level:     "info",
		Timestamp: time.Now(),
		Message:   fmt.Sprintf("Business event: %s", eventType),
		Context: models.LogContext{
			RequestID:   uuid.New().String(),
			ServiceName: ls.config.Name,
			Version:     ls.config.Version,
			Environment: ls.config.Environment,
		},
		Business: &models.BusinessData{
			EventType: eventType,
			Action:    "event_logged",
			Metadata:  data,
		},
	}

	logger.Info("Business event logged",
		zap.String("event_type", eventType),
		zap.Any("data", data),
		zap.String("request_id", entry.Context.RequestID),
	)

	metrics.LogEntriesTotal.WithLabelValues("info", ls.config.Name, "business").Inc()
}

// LogPerformance logs performance data
func (ls *LoggingService) LogPerformance(operation string, duration time.Duration, additionalData map[string]interface{}) {
	start := time.Now()
	defer func() {
		processingDuration := time.Since(start)
		metrics.LogProcessingDuration.WithLabelValues("performance", "info").Observe(processingDuration.Seconds())
	}()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	entry := models.LogEntry{
		Level:     "info",
		Timestamp: time.Now(),
		Message:   fmt.Sprintf("Performance data for operation: %s", operation),
		Context: models.LogContext{
			RequestID:   uuid.New().String(),
			ServiceName: ls.config.Name,
			Version:     ls.config.Version,
			Environment: ls.config.Environment,
		},
		Performance: &models.PerformanceData{
			Duration:       float64(duration.Nanoseconds()) / 1e6,
			MemoryUsage:    int64(m.Alloc),
			GoroutineCount: runtime.NumGoroutine(),
		},
		Data: additionalData,
	}

	logger.Info("Performance logged",
		zap.String("operation", operation),
		zap.Duration("duration", duration),
		zap.Int64("memory_usage", entry.Performance.MemoryUsage),
		zap.Int("goroutines", entry.Performance.GoroutineCount),
		zap.Any("additional_data", additionalData),
	)

	metrics.LogEntriesTotal.WithLabelValues("info", ls.config.Name, "performance").Inc()
}

// LogError logs error information
func (ls *LoggingService) LogError(ctx context.Context, errorType, errorCode, message string, err error, additionalData map[string]interface{}) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		metrics.LogProcessingDuration.WithLabelValues("error", "error").Observe(duration.Seconds())
	}()

	stackTrace := ""
	if err != nil {
		stackTrace = err.Error()
	}

	entry := models.LogEntry{
		Level:     "error",
		Timestamp: time.Now(),
		Message:   message,
		Context: models.LogContext{
			RequestID:   ls.getRequestIDFromContext(ctx),
			TraceID:     ls.extractTraceID(ctx),
			SpanID:      ls.extractSpanID(ctx),
			ServiceName: ls.config.Name,
			Version:     ls.config.Version,
			Environment: ls.config.Environment,
		},
		Error: &models.LogErrorData{
			Type:       errorType,
			Code:       errorCode,
			Message:    message,
			StackTrace: stackTrace,
		},
		Data: additionalData,
	}

	logger.Error("Error logged",
		zap.String("error_type", errorType),
		zap.String("error_code", errorCode),
		zap.String("error_message", message),
		zap.Error(err),
		zap.String("request_id", entry.Context.RequestID),
		zap.String("trace_id", entry.Context.TraceID),
		zap.Any("additional_data", additionalData),
	)

	metrics.LogEntriesTotal.WithLabelValues("error", ls.config.Name, errorType).Inc()
	metrics.ErrorsByCategory.WithLabelValues(errorType, "high", ls.config.Name).Inc()
}

// getRequestIDFromContext gets request ID from context
func (ls *LoggingService) getRequestIDFromContext(ctx context.Context) string {
	if requestID := ctx.Value(models.RequestIDKey); requestID != nil {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return uuid.New().String()
}
