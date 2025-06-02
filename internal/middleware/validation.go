package middleware

import (
	"strconv"
	"time"
)

// ValidationConfig holds validation limits
type ValidationConfig struct {
	MaxTestDuration    time.Duration
	MaxConcurrency     int
	MaxCount           int
	DefaultTimeout     time.Duration
	DefaultConcurrency int
	DefaultCount       int
}

// DefaultValidationConfig returns sensible defaults for internal production use
func DefaultValidationConfig() ValidationConfig {
	return ValidationConfig{
		MaxTestDuration:    10 * time.Minute, // Maximum 10 minutes for any test
		MaxConcurrency:     50,               // Maximum 50 concurrent workers
		MaxCount:           100000,           // Maximum 100k items to generate
		DefaultTimeout:     30 * time.Second, // Default request timeout
		DefaultConcurrency: 10,               // Default concurrency
		DefaultCount:       1000,             // Default count for generation
	}
}

// ValidateDuration validates and sanitizes duration parameters
func ValidateDuration(durationStr string, config ValidationConfig) time.Duration {
	if durationStr == "" {
		return config.DefaultTimeout
	}

	if parsed, err := time.ParseDuration(durationStr); err == nil {
		if parsed <= 0 {
			return config.DefaultTimeout
		}
		if parsed > config.MaxTestDuration {
			return config.MaxTestDuration
		}
		return parsed
	}

	return config.DefaultTimeout
}

// ValidateCount validates and sanitizes count parameters
func ValidateCount(countStr string, config ValidationConfig) int {
	if countStr == "" {
		return config.DefaultCount
	}

	if parsed, err := strconv.Atoi(countStr); err == nil {
		if parsed <= 0 {
			return config.DefaultCount
		}
		if parsed > config.MaxCount {
			return config.MaxCount
		}
		return parsed
	}

	return config.DefaultCount
}

// ValidateConcurrency validates and sanitizes concurrency parameters
func ValidateConcurrency(concurrencyStr string, config ValidationConfig) int {
	if concurrencyStr == "" {
		return config.DefaultConcurrency
	}

	if parsed, err := strconv.Atoi(concurrencyStr); err == nil {
		if parsed <= 0 {
			return config.DefaultConcurrency
		}
		if parsed > config.MaxConcurrency {
			return config.MaxConcurrency
		}
		return parsed
	}

	return config.DefaultConcurrency
}

// ValidatePositiveInt validates any positive integer parameter
func ValidatePositiveInt(valueStr string, defaultValue, maxValue int) int {
	if valueStr == "" {
		return defaultValue
	}

	if parsed, err := strconv.Atoi(valueStr); err == nil {
		if parsed <= 0 {
			return defaultValue
		}
		if maxValue > 0 && parsed > maxValue {
			return maxValue
		}
		return parsed
	}

	return defaultValue
}

// ValidateLogLevel validates log level parameters
func ValidateLogLevel(levelStr string) string {
	validLevels := []string{"info", "warn", "error", "mixed"}

	for _, level := range validLevels {
		if levelStr == level {
			return levelStr
		}
	}

	return "mixed" // default
}

// ValidateStringFromList validates that a string is in an allowed list
func ValidateStringFromList(value string, allowedValues []string, defaultValue string) string {
	for _, allowed := range allowedValues {
		if value == allowed {
			return value
		}
	}
	return defaultValue
}
