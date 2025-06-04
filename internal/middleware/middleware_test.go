package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nahuelsantos/argus/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponseWriter_WriteHeader(t *testing.T) {
	recorder := httptest.NewRecorder()
	rw := &ResponseWriter{ResponseWriter: recorder, statusCode: 200}

	rw.WriteHeader(404)
	assert.Equal(t, 404, rw.statusCode)
	assert.Equal(t, 404, recorder.Code)
}

func TestEnhancedResponseWriter_WriteHeader(t *testing.T) {
	recorder := httptest.NewRecorder()
	rw := &EnhancedResponseWriter{ResponseWriter: recorder, statusCode: 200}

	rw.WriteHeader(500)
	assert.Equal(t, 500, rw.statusCode)
	assert.Equal(t, 500, recorder.Code)
}

func TestEnhancedResponseWriter_Write(t *testing.T) {
	recorder := httptest.NewRecorder()
	rw := &EnhancedResponseWriter{ResponseWriter: recorder}

	data := []byte("Hello, World!")
	n, err := rw.Write(data)

	require.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.Equal(t, len(data), rw.responseSize)
	assert.Equal(t, "Hello, World!", recorder.Body.String())
}

func TestRateLimitMiddleware(t *testing.T) {
	tests := []struct {
		name          string
		requests      int
		expectBlocked bool
		clientIP      string
	}{
		{
			name:          "normal requests allowed",
			requests:      10,
			expectBlocked: false,
			clientIP:      "192.168.1.1",
		},
		{
			name:          "many requests from same IP",
			requests:      1001, // Over the 1000 limit
			expectBlocked: true,
			clientIP:      "192.168.1.2",
		},
		{
			name:          "different IPs not blocked",
			requests:      10,
			expectBlocked: false,
			clientIP:      "192.168.1.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("OK"))
			})

			middleware := RateLimitMiddleware(handler)
			blockedCount := 0

			for i := 0; i < tt.requests; i++ {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("X-Forwarded-For", tt.clientIP)
				req.RemoteAddr = tt.clientIP + ":12345"

				w := httptest.NewRecorder()
				middleware.ServeHTTP(w, req)

				if w.Code == http.StatusTooManyRequests {
					blockedCount++
				}
			}

			if tt.expectBlocked {
				assert.Greater(t, blockedCount, 0, "Expected some requests to be blocked")
			} else {
				assert.Equal(t, 0, blockedCount, "Expected no requests to be blocked")
			}
		})
	}
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := SecurityHeadersMiddleware(handler)
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	// Check all security headers are set
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "SAMEORIGIN", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTimeoutMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		timeout        time.Duration
		handlerDelay   time.Duration
		expectedStatus int
		isLongRunning  bool
	}{
		{
			name:           "normal request completes",
			path:           "/api/health",
			timeout:        100 * time.Millisecond,
			handlerDelay:   50 * time.Millisecond,
			expectedStatus: http.StatusOK,
			isLongRunning:  false,
		},
		{
			name:           "normal request times out",
			path:           "/api/slow",
			timeout:        50 * time.Millisecond,
			handlerDelay:   100 * time.Millisecond,
			expectedStatus: http.StatusRequestTimeout,
			isLongRunning:  false,
		},
		{
			name:           "long running endpoint not timed out",
			path:           "/test-metrics-scale",
			timeout:        50 * time.Millisecond,
			handlerDelay:   100 * time.Millisecond,
			expectedStatus: http.StatusOK,
			isLongRunning:  true,
		},
		{
			name:           "performance test endpoint",
			path:           "/test-logs-scale",
			timeout:        50 * time.Millisecond,
			handlerDelay:   100 * time.Millisecond,
			expectedStatus: http.StatusOK,
			isLongRunning:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(tt.handlerDelay)
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("OK"))
			})

			middleware := TimeoutMiddleware(tt.timeout)(handler)
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			middleware.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusRequestTimeout {
				assert.Contains(t, w.Body.String(), "Request timeout")
			}
		})
	}
}

func TestIsLongRunningEndpoint(t *testing.T) {
	tests := []struct {
		path           string
		expectedResult bool
	}{
		{"/test-metrics-scale", true},
		{"/test-logs-scale", true},
		{"/test-traces-scale", true},
		{"/test-dashboard-load", true},
		{"/test-resource-usage", true},
		{"/test-storage-limits", true},
		{"/simulate/web-service", true},
		{"/simulate/api-service", true},
		{"/simulate/database-service", true},
		{"/simulate/static-site", true},
		{"/simulate/microservice", true},
		{"/api/health", false},
		{"/api/metrics", false},
		{"/random/path", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isLongRunningEndpoint(tt.path)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestCORSMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
		checkHeaders   bool
	}{
		{
			name:           "OPTIONS request",
			method:         "OPTIONS",
			expectedStatus: http.StatusOK,
			checkHeaders:   true,
		},
		{
			name:           "GET request",
			method:         "GET",
			expectedStatus: http.StatusOK,
			checkHeaders:   true,
		},
		{
			name:           "POST request",
			method:         "POST",
			expectedStatus: http.StatusOK,
			checkHeaders:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			middleware := CORSMiddleware(handler)
			req := httptest.NewRequest(tt.method, "/test", nil)
			w := httptest.NewRecorder()

			middleware.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkHeaders {
				assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
				assert.Equal(t, "GET, POST, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
				assert.Equal(t, "Content-Type, X-Request-ID, X-User-ID, X-Session-ID", w.Header().Get("Access-Control-Allow-Headers"))
				assert.Equal(t, "X-Request-ID, X-Trace-ID", w.Header().Get("Access-Control-Expose-Headers"))
				assert.Equal(t, "86400", w.Header().Get("Access-Control-Max-Age"))
			}
		})
	}
}

func TestPrometheusMiddleware(t *testing.T) {
	tests := []struct {
		name          string
		method        string
		path          string
		handlerStatus int
	}{
		{
			name:          "GET request",
			method:        "GET",
			path:          "/api/health",
			handlerStatus: http.StatusOK,
		},
		{
			name:          "POST request",
			method:        "POST",
			path:          "/api/users",
			handlerStatus: http.StatusCreated,
		},
		{
			name:          "error request",
			method:        "GET",
			path:          "/api/error",
			handlerStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.handlerStatus)
				_, _ = w.Write([]byte("Response"))
			})

			middleware := PrometheusMiddleware(handler)
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			middleware.ServeHTTP(w, req)

			assert.Equal(t, tt.handlerStatus, w.Code)
			assert.Equal(t, "Response", w.Body.String())

			// The middleware should not panic and should complete normally
			// We can't easily test the Prometheus metrics here without more setup
		})
	}
}

func TestEnhancedTracingMiddleware(t *testing.T) {
	loggingService := services.NewLoggingService()
	tracingService := services.NewTracingService()
	loggingService.InitTestLogger()
	tracingService.InitTracer()

	tests := []struct {
		name          string
		method        string
		path          string
		handlerStatus int
	}{
		{
			name:          "successful request",
			method:        "GET",
			path:          "/api/users",
			handlerStatus: http.StatusOK,
		},
		{
			name:          "error request",
			method:        "POST",
			path:          "/api/error",
			handlerStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.handlerStatus)
				_, _ = w.Write([]byte("Response"))
			})

			middleware := EnhancedTracingMiddleware(loggingService, tracingService)(handler)
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			middleware.ServeHTTP(w, req)

			assert.Equal(t, tt.handlerStatus, w.Code)
			assert.Equal(t, "Response", w.Body.String())

			// Check that trace headers might be set
			// The middleware should complete without panicking
		})
	}
}

func TestRequestCorrelationMiddleware(t *testing.T) {
	loggingService := services.NewLoggingService()
	loggingService.InitTestLogger()

	tests := []struct {
		name              string
		requestIDHeader   string
		userIDHeader      string
		sessionIDHeader   string
		expectGeneratedID bool
	}{
		{
			name:              "no headers provided",
			expectGeneratedID: true,
		},
		{
			name:            "request ID provided",
			requestIDHeader: "existing-request-id",
		},
		{
			name:         "user ID provided",
			userIDHeader: "user-123",
		},
		{
			name:            "session ID provided",
			sessionIDHeader: "session-456",
		},
		{
			name:            "all headers provided",
			requestIDHeader: "req-789",
			userIDHeader:    "user-789",
			sessionIDHeader: "session-789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedContext context.Context

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedContext = r.Context()
				w.WriteHeader(http.StatusOK)
			})

			middleware := RequestCorrelationMiddleware(loggingService)(handler)
			req := httptest.NewRequest("GET", "/test", nil)

			if tt.requestIDHeader != "" {
				req.Header.Set("X-Request-ID", tt.requestIDHeader)
			}
			if tt.userIDHeader != "" {
				req.Header.Set("X-User-ID", tt.userIDHeader)
			}
			if tt.sessionIDHeader != "" {
				req.Header.Set("X-Session-ID", tt.sessionIDHeader)
			}

			w := httptest.NewRecorder()
			middleware.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.NotNil(t, capturedContext)

			// Check that correlation headers are returned
			if tt.requestIDHeader != "" {
				assert.Equal(t, tt.requestIDHeader, w.Header().Get("X-Request-ID"))
			} else if tt.expectGeneratedID {
				assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
			}
		})
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name         string
		remoteAddr   string
		forwardedFor string
		realIP       string
		expectedIP   string
	}{
		{
			name:       "remote addr only",
			remoteAddr: "192.168.1.1:12345",
			expectedIP: "192.168.1.1",
		},
		{
			name:         "X-Forwarded-For header",
			remoteAddr:   "10.0.0.1:12345",
			forwardedFor: "203.0.113.1",
			expectedIP:   "203.0.113.1",
		},
		{
			name:       "X-Real-IP header",
			remoteAddr: "10.0.0.1:12345",
			realIP:     "203.0.113.2",
			expectedIP: "203.0.113.2",
		},
		{
			name:         "multiple headers prefer X-Forwarded-For",
			remoteAddr:   "10.0.0.1:12345",
			forwardedFor: "203.0.113.1",
			realIP:       "203.0.113.2",
			expectedIP:   "203.0.113.1", // X-Forwarded-For takes precedence
		},
		{
			name:         "X-Forwarded-For with multiple IPs",
			remoteAddr:   "10.0.0.1:12345",
			forwardedFor: "203.0.113.1, 198.51.100.1, 192.168.1.1",
			expectedIP:   "203.0.113.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr

			if tt.forwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.forwardedFor)
			}
			if tt.realIP != "" {
				req.Header.Set("X-Real-IP", tt.realIP)
			}

			ip := getClientIP(req)
			assert.Equal(t, tt.expectedIP, ip)
		})
	}
}

func TestGetLogLevel(t *testing.T) {
	tests := []struct {
		statusCode    int
		expectedLevel string
	}{
		{200, "info"}, // 2xx status codes return InfoLevel
		{201, "info"},
		{204, "info"},
		{300, "info"}, // 3xx status codes also return InfoLevel
		{301, "info"},
		{302, "info"},
		{400, "warn"}, // 4xx status codes return WarnLevel
		{401, "warn"},
		{404, "warn"},
		{500, "error"}, // 5xx status codes return ErrorLevel
		{502, "error"},
		{503, "error"},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.statusCode)), func(t *testing.T) {
			level := getLogLevel(tt.statusCode)
			assert.Equal(t, tt.expectedLevel, level.String())
		})
	}
}

func TestAddMiddleware(t *testing.T) {
	loggingService := services.NewLoggingService()
	loggingService.InitTestLogger()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Test that AddMiddleware doesn't panic and returns a handler
	middlewareStack := AddMiddleware(handler, loggingService)
	assert.NotNil(t, middlewareStack)

	// Test that the middleware stack works
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middlewareStack.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())

	// Check that some middleware headers are present
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

// Benchmark tests
func BenchmarkRateLimitMiddleware(b *testing.B) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RateLimitMiddleware(handler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		middleware.ServeHTTP(w, req)
	}
}

func BenchmarkSecurityHeadersMiddleware(b *testing.B) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := SecurityHeadersMiddleware(handler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		middleware.ServeHTTP(w, req)
	}
}

func BenchmarkPrometheusMiddleware(b *testing.B) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := PrometheusMiddleware(handler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/test", nil)
		w := httptest.NewRecorder()
		middleware.ServeHTTP(w, req)
	}
}
