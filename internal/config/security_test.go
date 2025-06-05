package config

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetSecurityConfig(t *testing.T) {
	// Save original environment variables
	envVars := []string{
		"ARGUS_RATE_LIMIT_RPM",
		"ARGUS_REQUEST_TIMEOUT",
		"ARGUS_READ_TIMEOUT",
		"ARGUS_WRITE_TIMEOUT",
		"ARGUS_IDLE_TIMEOUT",
		"ARGUS_SHUTDOWN_TIMEOUT",
		"ARGUS_MAX_TEST_DURATION",
		"ARGUS_MAX_CONCURRENCY",
		"ARGUS_MAX_COUNT",
		"ARGUS_MAX_HEADER_BYTES",
		"ARGUS_ENABLE_CORS",
		"ARGUS_ENABLE_SECURITY_HEADERS",
	}

	originalValues := make(map[string]string)
	for _, key := range envVars {
		originalValues[key] = os.Getenv(key)
		os.Unsetenv(key)
	}

	defer func() {
		for key, value := range originalValues {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	t.Run("returns default values when no env vars set", func(t *testing.T) {
		config := GetSecurityConfig()

		// Rate limiting
		assert.Equal(t, 1000, config.RateLimitRPM)

		// Timeouts
		assert.Equal(t, 30*time.Second, config.RequestTimeout)
		assert.Equal(t, 15*time.Second, config.ReadTimeout)
		assert.Equal(t, 30*time.Second, config.WriteTimeout)
		assert.Equal(t, 60*time.Second, config.IdleTimeout)
		assert.Equal(t, 30*time.Second, config.ShutdownTimeout)

		// Limits
		assert.Equal(t, 10*time.Minute, config.MaxTestDuration)
		assert.Equal(t, 50, config.MaxConcurrency)
		assert.Equal(t, 100000, config.MaxCount)
		assert.Equal(t, 1<<20, config.MaxHeaderBytes) // 1MB

		// CORS
		assert.True(t, config.EnableCORS)
		assert.Equal(t, []string{"*"}, config.AllowedOrigins)
		assert.Equal(t, []string{"GET", "POST", "OPTIONS"}, config.AllowedMethods)
		assert.Equal(t, []string{"Content-Type", "X-Request-ID", "X-User-ID", "X-Session-ID"}, config.AllowedHeaders)

		// Security Headers
		assert.True(t, config.EnableSecurityHeaders)
	})

	t.Run("uses environment variables when set", func(t *testing.T) {
		// Set environment variables
		os.Setenv("ARGUS_RATE_LIMIT_RPM", "2000")
		os.Setenv("ARGUS_REQUEST_TIMEOUT", "45s")
		os.Setenv("ARGUS_READ_TIMEOUT", "20s")
		os.Setenv("ARGUS_WRITE_TIMEOUT", "40s")
		os.Setenv("ARGUS_IDLE_TIMEOUT", "90s")
		os.Setenv("ARGUS_SHUTDOWN_TIMEOUT", "60s")
		os.Setenv("ARGUS_MAX_TEST_DURATION", "15m")
		os.Setenv("ARGUS_MAX_CONCURRENCY", "100")
		os.Setenv("ARGUS_MAX_COUNT", "200000")
		os.Setenv("ARGUS_MAX_HEADER_BYTES", "2097152") // 2MB
		os.Setenv("ARGUS_ENABLE_CORS", "false")
		os.Setenv("ARGUS_ENABLE_SECURITY_HEADERS", "false")

		config := GetSecurityConfig()

		// Verify environment values are used
		assert.Equal(t, 2000, config.RateLimitRPM)
		assert.Equal(t, 45*time.Second, config.RequestTimeout)
		assert.Equal(t, 20*time.Second, config.ReadTimeout)
		assert.Equal(t, 40*time.Second, config.WriteTimeout)
		assert.Equal(t, 90*time.Second, config.IdleTimeout)
		assert.Equal(t, 60*time.Second, config.ShutdownTimeout)
		assert.Equal(t, 15*time.Minute, config.MaxTestDuration)
		assert.Equal(t, 100, config.MaxConcurrency)
		assert.Equal(t, 200000, config.MaxCount)
		assert.Equal(t, 2097152, config.MaxHeaderBytes)
		assert.False(t, config.EnableCORS)
		assert.False(t, config.EnableSecurityHeaders)
	})
}

func TestGetIntEnv(t *testing.T) {
	originalValue := os.Getenv("TEST_INT_VAR")
	defer func() {
		if originalValue == "" {
			os.Unsetenv("TEST_INT_VAR")
		} else {
			os.Setenv("TEST_INT_VAR", originalValue)
		}
	}()

	tests := []struct {
		name         string
		key          string
		defaultValue int
		envValue     string
		setEnv       bool
		expected     int
	}{
		{
			name:         "returns environment variable when valid",
			key:          "TEST_INT_VAR",
			defaultValue: 100,
			envValue:     "200",
			setEnv:       true,
			expected:     200,
		},
		{
			name:         "returns default when env var not set",
			key:          "TEST_INT_VAR",
			defaultValue: 100,
			setEnv:       false,
			expected:     100,
		},
		{
			name:         "returns default when env var is empty",
			key:          "TEST_INT_VAR",
			defaultValue: 100,
			envValue:     "",
			setEnv:       true,
			expected:     100,
		},
		{
			name:         "returns default when env var is invalid",
			key:          "TEST_INT_VAR",
			defaultValue: 100,
			envValue:     "not-a-number",
			setEnv:       true,
			expected:     100,
		},
		{
			name:         "handles negative numbers",
			key:          "TEST_INT_VAR",
			defaultValue: 100,
			envValue:     "-50",
			setEnv:       true,
			expected:     -50,
		},
		{
			name:         "handles zero",
			key:          "TEST_INT_VAR",
			defaultValue: 100,
			envValue:     "0",
			setEnv:       true,
			expected:     0,
		},
		{
			name:         "handles large numbers",
			key:          "TEST_INT_VAR",
			defaultValue: 100,
			envValue:     "1000000",
			setEnv:       true,
			expected:     1000000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
			} else {
				os.Unsetenv(tt.key)
			}

			// Test
			result := getIntEnv(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetBoolEnv(t *testing.T) {
	originalValue := os.Getenv("TEST_BOOL_VAR")
	defer func() {
		if originalValue == "" {
			os.Unsetenv("TEST_BOOL_VAR")
		} else {
			os.Setenv("TEST_BOOL_VAR", originalValue)
		}
	}()

	tests := []struct {
		name         string
		key          string
		defaultValue bool
		envValue     string
		setEnv       bool
		expected     bool
	}{
		{
			name:         "returns true when env var is 'true'",
			key:          "TEST_BOOL_VAR",
			defaultValue: false,
			envValue:     "true",
			setEnv:       true,
			expected:     true,
		},
		{
			name:         "returns false when env var is 'false'",
			key:          "TEST_BOOL_VAR",
			defaultValue: true,
			envValue:     "false",
			setEnv:       true,
			expected:     false,
		},
		{
			name:         "returns default when env var not set",
			key:          "TEST_BOOL_VAR",
			defaultValue: true,
			setEnv:       false,
			expected:     true,
		},
		{
			name:         "returns default when env var is empty",
			key:          "TEST_BOOL_VAR",
			defaultValue: true,
			envValue:     "",
			setEnv:       true,
			expected:     true,
		},
		{
			name:         "returns default when env var is invalid",
			key:          "TEST_BOOL_VAR",
			defaultValue: true,
			envValue:     "not-a-bool",
			setEnv:       true,
			expected:     true,
		},
		{
			name:         "handles '1' as true",
			key:          "TEST_BOOL_VAR",
			defaultValue: false,
			envValue:     "1",
			setEnv:       true,
			expected:     true,
		},
		{
			name:         "handles '0' as false",
			key:          "TEST_BOOL_VAR",
			defaultValue: true,
			envValue:     "0",
			setEnv:       true,
			expected:     false,
		},
		{
			name:         "handles 'TRUE' (case insensitive)",
			key:          "TEST_BOOL_VAR",
			defaultValue: false,
			envValue:     "TRUE",
			setEnv:       true,
			expected:     true,
		},
		{
			name:         "handles 'FALSE' (case insensitive)",
			key:          "TEST_BOOL_VAR",
			defaultValue: true,
			envValue:     "FALSE",
			setEnv:       true,
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
			} else {
				os.Unsetenv(tt.key)
			}

			// Test
			result := getBoolEnv(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetDurationEnv(t *testing.T) {
	originalValue := os.Getenv("TEST_DURATION_VAR")
	defer func() {
		if originalValue == "" {
			os.Unsetenv("TEST_DURATION_VAR")
		} else {
			os.Setenv("TEST_DURATION_VAR", originalValue)
		}
	}()

	tests := []struct {
		name         string
		key          string
		defaultValue time.Duration
		envValue     string
		setEnv       bool
		expected     time.Duration
	}{
		{
			name:         "returns duration when env var is valid",
			key:          "TEST_DURATION_VAR",
			defaultValue: 10 * time.Second,
			envValue:     "30s",
			setEnv:       true,
			expected:     30 * time.Second,
		},
		{
			name:         "returns default when env var not set",
			key:          "TEST_DURATION_VAR",
			defaultValue: 10 * time.Second,
			setEnv:       false,
			expected:     10 * time.Second,
		},
		{
			name:         "returns default when env var is empty",
			key:          "TEST_DURATION_VAR",
			defaultValue: 10 * time.Second,
			envValue:     "",
			setEnv:       true,
			expected:     10 * time.Second,
		},
		{
			name:         "returns default when env var is invalid",
			key:          "TEST_DURATION_VAR",
			defaultValue: 10 * time.Second,
			envValue:     "not-a-duration",
			setEnv:       true,
			expected:     10 * time.Second,
		},
		{
			name:         "handles minutes",
			key:          "TEST_DURATION_VAR",
			defaultValue: 10 * time.Second,
			envValue:     "5m",
			setEnv:       true,
			expected:     5 * time.Minute,
		},
		{
			name:         "handles hours",
			key:          "TEST_DURATION_VAR",
			defaultValue: 10 * time.Second,
			envValue:     "2h",
			setEnv:       true,
			expected:     2 * time.Hour,
		},
		{
			name:         "handles milliseconds",
			key:          "TEST_DURATION_VAR",
			defaultValue: 10 * time.Second,
			envValue:     "500ms",
			setEnv:       true,
			expected:     500 * time.Millisecond,
		},
		{
			name:         "handles microseconds",
			key:          "TEST_DURATION_VAR",
			defaultValue: 10 * time.Second,
			envValue:     "100µs",
			setEnv:       true,
			expected:     100 * time.Microsecond,
		},
		{
			name:         "handles nanoseconds",
			key:          "TEST_DURATION_VAR",
			defaultValue: 10 * time.Second,
			envValue:     "1000ns",
			setEnv:       true,
			expected:     1000 * time.Nanosecond,
		},
		{
			name:         "handles complex duration",
			key:          "TEST_DURATION_VAR",
			defaultValue: 10 * time.Second,
			envValue:     "1h30m45s",
			setEnv:       true,
			expected:     1*time.Hour + 30*time.Minute + 45*time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
			} else {
				os.Unsetenv(tt.key)
			}

			// Test
			result := getDurationEnv(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSecurityConfigValidation(t *testing.T) {
	config := GetSecurityConfig()

	t.Run("rate limiting is positive", func(t *testing.T) {
		assert.Greater(t, config.RateLimitRPM, 0)
	})

	t.Run("timeouts are positive", func(t *testing.T) {
		assert.Greater(t, config.RequestTimeout, 0*time.Second)
		assert.Greater(t, config.ReadTimeout, 0*time.Second)
		assert.Greater(t, config.WriteTimeout, 0*time.Second)
		assert.Greater(t, config.IdleTimeout, 0*time.Second)
		assert.Greater(t, config.ShutdownTimeout, 0*time.Second)
		assert.Greater(t, config.MaxTestDuration, 0*time.Second)
	})

	t.Run("limits are positive", func(t *testing.T) {
		assert.Greater(t, config.MaxConcurrency, 0)
		assert.Greater(t, config.MaxCount, 0)
		assert.Greater(t, config.MaxHeaderBytes, 0)
	})

	t.Run("CORS configuration is valid", func(t *testing.T) {
		assert.NotNil(t, config.AllowedOrigins)
		assert.NotNil(t, config.AllowedMethods)
		assert.NotNil(t, config.AllowedHeaders)

		if config.EnableCORS {
			assert.NotEmpty(t, config.AllowedOrigins)
			assert.NotEmpty(t, config.AllowedMethods)
		}
	})

	t.Run("reasonable timeout values", func(t *testing.T) {
		// Ensure timeouts are reasonable (not too small or too large)
		assert.LessOrEqual(t, config.ReadTimeout, config.RequestTimeout)
		assert.LessOrEqual(t, config.WriteTimeout, config.RequestTimeout)
		assert.GreaterOrEqual(t, config.IdleTimeout, config.RequestTimeout)
	})
}

func TestSecurityConfigEdgeCases(t *testing.T) {
	originalValues := make(map[string]string)
	envVars := []string{
		"ARGUS_RATE_LIMIT_RPM",
		"ARGUS_MAX_CONCURRENCY",
		"ARGUS_REQUEST_TIMEOUT",
		"ARGUS_ENABLE_CORS",
	}

	for _, key := range envVars {
		originalValues[key] = os.Getenv(key)
	}

	defer func() {
		for key, value := range originalValues {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	t.Run("handles extreme values gracefully", func(t *testing.T) {
		// Test with extreme but valid values
		os.Setenv("ARGUS_RATE_LIMIT_RPM", strconv.Itoa(int(^uint(0)>>1))) // max int
		os.Setenv("ARGUS_MAX_CONCURRENCY", "1")                           // minimum reasonable
		os.Setenv("ARGUS_REQUEST_TIMEOUT", "1ns")                         // very small

		config := GetSecurityConfig()

		// Should not panic and should have valid values
		assert.Greater(t, config.RateLimitRPM, 0)
		assert.Greater(t, config.MaxConcurrency, 0)
		assert.Greater(t, config.RequestTimeout, 0*time.Nanosecond)
	})

	t.Run("handles invalid environment values gracefully", func(t *testing.T) {
		// Set invalid values
		os.Setenv("ARGUS_RATE_LIMIT_RPM", "not-a-number")
		os.Setenv("ARGUS_REQUEST_TIMEOUT", "invalid-duration")
		os.Setenv("ARGUS_ENABLE_CORS", "maybe")

		config := GetSecurityConfig()

		// Should fall back to defaults
		assert.Equal(t, 1000, config.RateLimitRPM)             // default
		assert.Equal(t, 30*time.Second, config.RequestTimeout) // default
		assert.Equal(t, true, config.EnableCORS)               // default
	})
}

// Benchmark tests
func BenchmarkGetSecurityConfig(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetSecurityConfig()
	}
}

func BenchmarkGetIntEnv(b *testing.B) {
	os.Setenv("BENCH_INT_VAR", "12345")
	defer os.Unsetenv("BENCH_INT_VAR")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getIntEnv("BENCH_INT_VAR", 100)
	}
}

func BenchmarkGetBoolEnv(b *testing.B) {
	os.Setenv("BENCH_BOOL_VAR", "true")
	defer os.Unsetenv("BENCH_BOOL_VAR")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getBoolEnv("BENCH_BOOL_VAR", false)
	}
}

func BenchmarkGetDurationEnv(b *testing.B) {
	os.Setenv("BENCH_DURATION_VAR", "30s")
	defer os.Unsetenv("BENCH_DURATION_VAR")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getDurationEnv("BENCH_DURATION_VAR", 10*time.Second)
	}
}

// Example usage
func ExampleGetSecurityConfig() {
	config := GetSecurityConfig()

	// Use the configuration
	_ = config.RateLimitRPM   // 1000 (or from ARGUS_RATE_LIMIT_RPM)
	_ = config.RequestTimeout // 30s (or from ARGUS_REQUEST_TIMEOUT)
	_ = config.EnableCORS     // true (or from ARGUS_ENABLE_CORS)
}
