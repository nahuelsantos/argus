package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap/zapcore"

	"github.com/nahuelsantos/argus/internal/metrics"
	"github.com/nahuelsantos/argus/internal/models"
	"github.com/nahuelsantos/argus/internal/services"
)

// ResponseWriter wraps http.ResponseWriter to capture status code
type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (rw *ResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// EnhancedResponseWriter wraps ResponseWriter with additional functionality
type EnhancedResponseWriter struct {
	http.ResponseWriter
	statusCode   int
	responseSize int
}

// WriteHeader captures the status code
func (rw *EnhancedResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write captures the response size
func (rw *EnhancedResponseWriter) Write(data []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(data)
	rw.responseSize += size
	return size, err
}

// RateLimitMiddleware provides rate limiting per IP address
func RateLimitMiddleware(next http.Handler) http.Handler {
	clients := make(map[string][]time.Time)
	var mu sync.RWMutex

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)
		now := time.Now()

		mu.Lock()
		defer mu.Unlock()

		// Allow 1000 requests per minute per IP (generous for internal use)
		if requests, exists := clients[clientIP]; exists {
			// Clean old requests (older than 1 minute)
			var recent []time.Time
			for _, t := range requests {
				if now.Sub(t) < time.Minute {
					recent = append(recent, t)
				}
			}

			if len(recent) >= 1000 {
				http.Error(w, "Rate limit exceeded: 1000 requests per minute", http.StatusTooManyRequests)
				return
			}

			clients[clientIP] = append(recent, now)
		} else {
			clients[clientIP] = []time.Time{now}
		}

		next.ServeHTTP(w, r)
	})
}

// SecurityHeadersMiddleware adds essential security headers for internal use
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN") // Less restrictive for internal
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		next.ServeHTTP(w, r)
	})
}

// TimeoutMiddleware adds request timeout protection with conditional logic
func TimeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip timeout for long-running performance test endpoints
			if isLongRunningEndpoint(r.URL.Path) {
				// For performance tests, use a much longer timeout or no timeout
				longTimeout := 15 * time.Minute // 15 minutes max for performance tests

				ctx, cancel := context.WithTimeout(r.Context(), longTimeout)
				defer cancel()
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Apply normal timeout for other endpoints
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			done := make(chan struct{})
			go func() {
				next.ServeHTTP(w, r.WithContext(ctx))
				close(done)
			}()

			select {
			case <-done:
				// Request completed normally
			case <-ctx.Done():
				// Request timed out
				http.Error(w, "Request timeout", http.StatusRequestTimeout)
			}
		})
	}
}

// isLongRunningEndpoint checks if an endpoint is expected to run for a long time
func isLongRunningEndpoint(path string) bool {
	longRunningPaths := []string{
		"/test-metrics-scale",
		"/test-logs-scale",
		"/test-traces-scale",
		"/test-dashboard-load",
		"/test-resource-usage",
		"/test-storage-limits",
		"/simulate/web-service",
		"/simulate/api-service",
		"/simulate/database-service",
		"/simulate/static-site",
		"/simulate/microservice",
	}

	for _, longPath := range longRunningPaths {
		if path == longPath {
			return true
		}
	}
	return false
}

// CORSMiddleware handles CORS headers for internal network use
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow internal network access (safe for internal networks)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Request-ID, X-User-ID, X-Session-ID")
		w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID, X-Trace-ID")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// PrometheusMiddleware records HTTP metrics
func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &ResponseWriter{
			ResponseWriter: w,
			statusCode:     200,
		}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		metrics.HTTPRequestsTotal.WithLabelValues(
			r.Method,
			r.URL.Path,
			strconv.Itoa(wrapped.statusCode),
		).Inc()

		metrics.HTTPRequestDuration.WithLabelValues(
			r.Method,
			r.URL.Path,
		).Observe(duration.Seconds())
	})
}

// EnhancedTracingMiddleware provides comprehensive tracing
func EnhancedTracingMiddleware(loggingService *services.LoggingService, tracingService *services.TracingService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tracer := otel.Tracer("argus")
			ctx, span := tracer.Start(r.Context(), r.URL.Path)
			defer span.End()

			// Add trace attributes
			span.SetAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.url", r.URL.String()),
				attribute.String("http.scheme", r.URL.Scheme),
				attribute.String("http.host", r.Host),
				attribute.String("http.user_agent", r.UserAgent()),
				attribute.String("http.remote_addr", r.RemoteAddr),
			)

			// Create enhanced response writer
			wrapped := &EnhancedResponseWriter{
				ResponseWriter: w,
				statusCode:     200,
			}

			// Add request ID to context and headers
			requestID := loggingService.CreateLogContext(r).RequestID
			ctx = context.WithValue(ctx, models.RequestIDKey, requestID)
			wrapped.Header().Set("X-Request-ID", requestID)

			// Add trace ID to headers
			if span.SpanContext().IsValid() {
				traceID := span.SpanContext().TraceID().String()
				wrapped.Header().Set("X-Trace-ID", traceID)
			}

			start := time.Now()

			// Process request
			next.ServeHTTP(wrapped, r.WithContext(ctx))

			duration := time.Since(start)

			// Add response attributes
			span.SetAttributes(
				attribute.Int("http.status_code", wrapped.statusCode),
				attribute.Int("http.response_size", wrapped.responseSize),
				attribute.Int64("http.duration_ms", duration.Milliseconds()),
			)

			// Set span status based on HTTP status code
			if wrapped.statusCode >= 400 {
				span.SetAttributes(attribute.Bool("error", true))
			}

			// Create and log APM data
			apmData := tracingService.CreateAPMData(ctx, r.URL.Path, wrapped.statusCode, duration)
			tracingService.LogAPMData(apmData)

			// Log request with context
			loggingService.LogWithContext(
				getLogLevel(wrapped.statusCode),
				ctx,
				"HTTP request processed",
			)
		})
	}
}

// RequestCorrelationMiddleware adds correlation IDs to requests
func RequestCorrelationMiddleware(loggingService *services.LoggingService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create or extract request ID
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
			}

			// Create or extract user ID
			userID := r.Header.Get("X-User-ID")
			if userID == "" {
				userID = "anonymous"
			}

			// Create or extract session ID
			sessionID := r.Header.Get("X-Session-ID")
			if sessionID == "" {
				sessionID = uuid.New().String()
			}

			// Add to context
			ctx := r.Context()
			ctx = context.WithValue(ctx, models.RequestIDKey, requestID)
			ctx = context.WithValue(ctx, models.UserIDKey, userID)
			ctx = context.WithValue(ctx, models.SessionIDKey, sessionID)
			ctx = context.WithValue(ctx, models.StartTimeKey, time.Now())

			// Add to response headers
			w.Header().Set("X-Request-ID", requestID)
			w.Header().Set("X-Session-ID", sessionID)

			// Process request with enhanced context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Helper function to get client IP address
func getClientIP(r *http.Request) string {
	// Check for common proxy headers
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if idx := strings.Index(ip, ","); idx != -1 {
			return strings.TrimSpace(ip[:idx])
		}
		return strings.TrimSpace(ip)
	}

	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}

	// Fallback to RemoteAddr
	if idx := strings.LastIndex(r.RemoteAddr, ":"); idx != -1 {
		return r.RemoteAddr[:idx]
	}
	return r.RemoteAddr
}

// getLogLevel determines log level based on HTTP status code
func getLogLevel(statusCode int) zapcore.Level {
	switch {
	case statusCode >= 500:
		return zapcore.ErrorLevel
	case statusCode >= 400:
		return zapcore.WarnLevel
	default:
		return zapcore.InfoLevel
	}
}

// AddMiddleware wraps an HTTP handler with all necessary middleware
func AddMiddleware(handler http.Handler, loggingService *services.LoggingService) http.Handler {
	// Apply middleware in reverse order (last applied is executed first)
	wrapped := handler

	// Apply Prometheus metrics middleware
	wrapped = PrometheusMiddleware(wrapped)

	// Apply security headers middleware
	wrapped = SecurityHeadersMiddleware(wrapped)

	// Apply CORS middleware
	wrapped = CORSMiddleware(wrapped)

	// Apply rate limiting middleware
	wrapped = RateLimitMiddleware(wrapped)

	// Apply timeout middleware (30 second default)
	wrapped = TimeoutMiddleware(30 * time.Second)(wrapped)

	// Apply request correlation middleware
	wrapped = RequestCorrelationMiddleware(loggingService)(wrapped)

	return wrapped
}
