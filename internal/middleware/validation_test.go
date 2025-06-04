package middleware

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultValidationConfig(t *testing.T) {
	config := DefaultValidationConfig()

	assert.Equal(t, 10*time.Minute, config.MaxTestDuration)
	assert.Equal(t, 50, config.MaxConcurrency)
	assert.Equal(t, 100000, config.MaxCount)
	assert.Equal(t, 30*time.Second, config.DefaultTimeout)
	assert.Equal(t, 10, config.DefaultConcurrency)
	assert.Equal(t, 1000, config.DefaultCount)
}

func TestValidateDuration(t *testing.T) {
	config := DefaultValidationConfig()

	tests := []struct {
		name        string
		input       string
		expected    time.Duration
		description string
	}{
		{
			name:        "empty string returns default",
			input:       "",
			expected:    config.DefaultTimeout,
			description: "Should return default timeout for empty string",
		},
		{
			name:        "valid duration",
			input:       "5s",
			expected:    5 * time.Second,
			description: "Should parse valid duration",
		},
		{
			name:        "valid duration in minutes",
			input:       "2m",
			expected:    2 * time.Minute,
			description: "Should parse valid duration in minutes",
		},
		{
			name:        "duration exceeds maximum",
			input:       "20m",
			expected:    config.MaxTestDuration,
			description: "Should cap at maximum test duration",
		},
		{
			name:        "zero duration returns default",
			input:       "0s",
			expected:    config.DefaultTimeout,
			description: "Should return default for zero duration",
		},
		{
			name:        "negative duration returns default",
			input:       "-5s",
			expected:    config.DefaultTimeout,
			description: "Should return default for negative duration",
		},
		{
			name:        "invalid format returns default",
			input:       "invalid",
			expected:    config.DefaultTimeout,
			description: "Should return default for invalid format",
		},
		{
			name:        "valid milliseconds",
			input:       "500ms",
			expected:    500 * time.Millisecond,
			description: "Should parse milliseconds correctly",
		},
		{
			name:        "valid hours capped at max",
			input:       "2h",
			expected:    config.MaxTestDuration,
			description: "Should cap hours at maximum",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateDuration(tt.input, config)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

func TestValidateCount(t *testing.T) {
	config := DefaultValidationConfig()

	tests := []struct {
		name        string
		input       string
		expected    int
		description string
	}{
		{
			name:        "empty string returns default",
			input:       "",
			expected:    config.DefaultCount,
			description: "Should return default count for empty string",
		},
		{
			name:        "valid count",
			input:       "500",
			expected:    500,
			description: "Should parse valid count",
		},
		{
			name:        "count exceeds maximum",
			input:       "200000",
			expected:    config.MaxCount,
			description: "Should cap at maximum count",
		},
		{
			name:        "zero count returns default",
			input:       "0",
			expected:    config.DefaultCount,
			description: "Should return default for zero count",
		},
		{
			name:        "negative count returns default",
			input:       "-100",
			expected:    config.DefaultCount,
			description: "Should return default for negative count",
		},
		{
			name:        "invalid format returns default",
			input:       "abc",
			expected:    config.DefaultCount,
			description: "Should return default for invalid format",
		},
		{
			name:        "decimal number returns default",
			input:       "100.5",
			expected:    config.DefaultCount,
			description: "Should return default for decimal numbers",
		},
		{
			name:        "very large number capped",
			input:       "999999999",
			expected:    config.MaxCount,
			description: "Should cap very large numbers",
		},
		{
			name:        "edge case maximum",
			input:       "100000",
			expected:    100000,
			description: "Should accept exact maximum value",
		},
		{
			name:        "edge case minimum valid",
			input:       "1",
			expected:    1,
			description: "Should accept minimum valid value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateCount(tt.input, config)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

func TestValidateConcurrency(t *testing.T) {
	config := DefaultValidationConfig()

	tests := []struct {
		name        string
		input       string
		expected    int
		description string
	}{
		{
			name:        "empty string returns default",
			input:       "",
			expected:    config.DefaultConcurrency,
			description: "Should return default concurrency for empty string",
		},
		{
			name:        "valid concurrency",
			input:       "5",
			expected:    5,
			description: "Should parse valid concurrency",
		},
		{
			name:        "concurrency exceeds maximum",
			input:       "100",
			expected:    config.MaxConcurrency,
			description: "Should cap at maximum concurrency",
		},
		{
			name:        "zero concurrency returns default",
			input:       "0",
			expected:    config.DefaultConcurrency,
			description: "Should return default for zero concurrency",
		},
		{
			name:        "negative concurrency returns default",
			input:       "-5",
			expected:    config.DefaultConcurrency,
			description: "Should return default for negative concurrency",
		},
		{
			name:        "invalid format returns default",
			input:       "invalid",
			expected:    config.DefaultConcurrency,
			description: "Should return default for invalid format",
		},
		{
			name:        "edge case maximum",
			input:       "50",
			expected:    50,
			description: "Should accept exact maximum value",
		},
		{
			name:        "edge case minimum valid",
			input:       "1",
			expected:    1,
			description: "Should accept minimum valid value",
		},
		{
			name:        "reasonable value",
			input:       "20",
			expected:    20,
			description: "Should accept reasonable concurrency value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateConcurrency(tt.input, config)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

func TestValidatePositiveInt(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		defaultValue int
		maxValue     int
		expected     int
		description  string
	}{
		{
			name:         "empty string returns default",
			input:        "",
			defaultValue: 10,
			maxValue:     100,
			expected:     10,
			description:  "Should return default for empty string",
		},
		{
			name:         "valid positive integer",
			input:        "25",
			defaultValue: 10,
			maxValue:     100,
			expected:     25,
			description:  "Should parse valid positive integer",
		},
		{
			name:         "exceeds maximum",
			input:        "150",
			defaultValue: 10,
			maxValue:     100,
			expected:     100,
			description:  "Should cap at maximum value",
		},
		{
			name:         "zero returns default",
			input:        "0",
			defaultValue: 10,
			maxValue:     100,
			expected:     10,
			description:  "Should return default for zero",
		},
		{
			name:         "negative returns default",
			input:        "-5",
			defaultValue: 10,
			maxValue:     100,
			expected:     10,
			description:  "Should return default for negative",
		},
		{
			name:         "invalid format returns default",
			input:        "abc",
			defaultValue: 10,
			maxValue:     100,
			expected:     10,
			description:  "Should return default for invalid format",
		},
		{
			name:         "no maximum limit",
			input:        "500",
			defaultValue: 10,
			maxValue:     0, // 0 means no limit
			expected:     500,
			description:  "Should accept large values when no max limit",
		},
		{
			name:         "edge case exact maximum",
			input:        "100",
			defaultValue: 10,
			maxValue:     100,
			expected:     100,
			description:  "Should accept exact maximum value",
		},
		{
			name:         "edge case minimum valid",
			input:        "1",
			defaultValue: 10,
			maxValue:     100,
			expected:     1,
			description:  "Should accept minimum valid positive value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidatePositiveInt(tt.input, tt.defaultValue, tt.maxValue)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

func TestValidateLogLevel(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		description string
	}{
		{
			name:        "valid info level",
			input:       "info",
			expected:    "info",
			description: "Should return info for valid input",
		},
		{
			name:        "valid warn level",
			input:       "warn",
			expected:    "warn",
			description: "Should return warn for valid input",
		},
		{
			name:        "valid error level",
			input:       "error",
			expected:    "error",
			description: "Should return error for valid input",
		},
		{
			name:        "valid mixed level",
			input:       "mixed",
			expected:    "mixed",
			description: "Should return mixed for valid input",
		},
		{
			name:        "invalid level returns default",
			input:       "debug",
			expected:    "mixed",
			description: "Should return default mixed for invalid level",
		},
		{
			name:        "empty string returns default",
			input:       "",
			expected:    "mixed",
			description: "Should return default mixed for empty string",
		},
		{
			name:        "uppercase input returns default",
			input:       "INFO",
			expected:    "mixed",
			description: "Should return default for uppercase input",
		},
		{
			name:        "invalid string returns default",
			input:       "invalid",
			expected:    "mixed",
			description: "Should return default for invalid string",
		},
		{
			name:        "case sensitive test",
			input:       "Error",
			expected:    "mixed",
			description: "Should be case sensitive and return default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateLogLevel(tt.input)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

func TestValidateStringFromList(t *testing.T) {
	allowedValues := []string{"apple", "banana", "cherry"}
	defaultValue := "apple"

	tests := []struct {
		name        string
		input       string
		expected    string
		description string
	}{
		{
			name:        "valid value from list",
			input:       "banana",
			expected:    "banana",
			description: "Should return valid value from list",
		},
		{
			name:        "first item in list",
			input:       "apple",
			expected:    "apple",
			description: "Should return first item when valid",
		},
		{
			name:        "last item in list",
			input:       "cherry",
			expected:    "cherry",
			description: "Should return last item when valid",
		},
		{
			name:        "invalid value returns default",
			input:       "orange",
			expected:    defaultValue,
			description: "Should return default for invalid value",
		},
		{
			name:        "empty string returns default",
			input:       "",
			expected:    defaultValue,
			description: "Should return default for empty string",
		},
		{
			name:        "case sensitive test",
			input:       "Apple",
			expected:    defaultValue,
			description: "Should be case sensitive and return default",
		},
		{
			name:        "whitespace value returns default",
			input:       " banana ",
			expected:    defaultValue,
			description: "Should return default for value with whitespace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateStringFromList(tt.input, allowedValues, defaultValue)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

func TestValidateStringFromList_EmptyList(t *testing.T) {
	allowedValues := []string{}
	defaultValue := "default"

	result := ValidateStringFromList("anything", allowedValues, defaultValue)
	assert.Equal(t, defaultValue, result, "Should return default when allowed list is empty")
}

func TestValidateStringFromList_SingleItem(t *testing.T) {
	allowedValues := []string{"only"}
	defaultValue := "default"

	tests := []struct {
		input    string
		expected string
	}{
		{"only", "only"},
		{"other", "default"},
		{"", "default"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ValidateStringFromList(tt.input, allowedValues, defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Edge case and integration tests
func TestValidationIntegration(t *testing.T) {
	config := DefaultValidationConfig()

	// Test that all validation functions work together
	duration := ValidateDuration("30s", config)
	count := ValidateCount("500", config)
	concurrency := ValidateConcurrency("5", config)
	logLevel := ValidateLogLevel("info")

	assert.Equal(t, 30*time.Second, duration)
	assert.Equal(t, 500, count)
	assert.Equal(t, 5, concurrency)
	assert.Equal(t, "info", logLevel)
}

func TestValidationConfigCustom(t *testing.T) {
	customConfig := ValidationConfig{
		MaxTestDuration:    5 * time.Minute,
		MaxConcurrency:     20,
		MaxCount:           5000,
		DefaultTimeout:     10 * time.Second,
		DefaultConcurrency: 3,
		DefaultCount:       100,
	}

	// Test with custom configuration
	duration := ValidateDuration("", customConfig)
	count := ValidateCount("", customConfig)
	concurrency := ValidateConcurrency("", customConfig)

	assert.Equal(t, customConfig.DefaultTimeout, duration)
	assert.Equal(t, customConfig.DefaultCount, count)
	assert.Equal(t, customConfig.DefaultConcurrency, concurrency)

	// Test that limits are respected
	maxDuration := ValidateDuration("10m", customConfig)
	maxCount := ValidateCount("10000", customConfig)
	maxConcurrency := ValidateConcurrency("50", customConfig)

	assert.Equal(t, customConfig.MaxTestDuration, maxDuration)
	assert.Equal(t, customConfig.MaxCount, maxCount)
	assert.Equal(t, customConfig.MaxConcurrency, maxConcurrency)
}

// Benchmark tests
func BenchmarkValidateDuration(b *testing.B) {
	config := DefaultValidationConfig()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateDuration("30s", config)
	}
}

func BenchmarkValidateCount(b *testing.B) {
	config := DefaultValidationConfig()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateCount("1000", config)
	}
}

func BenchmarkValidateConcurrency(b *testing.B) {
	config := DefaultValidationConfig()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateConcurrency("10", config)
	}
}

func BenchmarkValidateLogLevel(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateLogLevel("info")
	}
}

func BenchmarkValidateStringFromList(b *testing.B) {
	allowedValues := []string{"apple", "banana", "cherry", "date", "elderberry"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateStringFromList("cherry", allowedValues, "apple")
	}
}
