package config

import (
	"os"
	"strconv"
	"time"
)

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	// Rate limiting
	RateLimitRPM int `json:"rate_limit_rpm"`

	// Timeouts
	RequestTimeout  time.Duration `json:"request_timeout"`
	ReadTimeout     time.Duration `json:"read_timeout"`
	WriteTimeout    time.Duration `json:"write_timeout"`
	IdleTimeout     time.Duration `json:"idle_timeout"`
	ShutdownTimeout time.Duration `json:"shutdown_timeout"`

	// Limits
	MaxTestDuration time.Duration `json:"max_test_duration"`
	MaxConcurrency  int           `json:"max_concurrency"`
	MaxCount        int           `json:"max_count"`
	MaxHeaderBytes  int           `json:"max_header_bytes"`

	// CORS
	EnableCORS     bool     `json:"enable_cors"`
	AllowedOrigins []string `json:"allowed_origins"`
	AllowedMethods []string `json:"allowed_methods"`
	AllowedHeaders []string `json:"allowed_headers"`

	// Security Headers
	EnableSecurityHeaders bool `json:"enable_security_headers"`
}

// GetSecurityConfig returns security configuration with environment variable overrides
func GetSecurityConfig() SecurityConfig {
	config := SecurityConfig{
		// Rate limiting - generous for internal use
		RateLimitRPM: getIntEnv("ARGUS_RATE_LIMIT_RPM", 1000),

		// Timeouts
		RequestTimeout:  getDurationEnv("ARGUS_REQUEST_TIMEOUT", 30*time.Second),
		ReadTimeout:     getDurationEnv("ARGUS_READ_TIMEOUT", 15*time.Second),
		WriteTimeout:    getDurationEnv("ARGUS_WRITE_TIMEOUT", 30*time.Second),
		IdleTimeout:     getDurationEnv("ARGUS_IDLE_TIMEOUT", 60*time.Second),
		ShutdownTimeout: getDurationEnv("ARGUS_SHUTDOWN_TIMEOUT", 30*time.Second),

		// Limits
		MaxTestDuration: getDurationEnv("ARGUS_MAX_TEST_DURATION", 10*time.Minute),
		MaxConcurrency:  getIntEnv("ARGUS_MAX_CONCURRENCY", 50),
		MaxCount:        getIntEnv("ARGUS_MAX_COUNT", 100000),
		MaxHeaderBytes:  getIntEnv("ARGUS_MAX_HEADER_BYTES", 1<<20), // 1MB

		// CORS - permissive for internal networks
		EnableCORS:     getBoolEnv("ARGUS_ENABLE_CORS", true),
		AllowedOrigins: []string{"*"}, // Safe for internal networks
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "X-Request-ID", "X-User-ID", "X-Session-ID"},

		// Security Headers
		EnableSecurityHeaders: getBoolEnv("ARGUS_ENABLE_SECURITY_HEADERS", true),
	}

	return config
}

// Helper functions for environment variable parsing
func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
