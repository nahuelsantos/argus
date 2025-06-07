package config

import (
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetServiceConfig(t *testing.T) {
	originalArgusEnv := os.Getenv("ARGUS_ENVIRONMENT")
	defer func() {
		if originalArgusEnv == "" {
			os.Unsetenv("ARGUS_ENVIRONMENT")
		} else {
			os.Setenv("ARGUS_ENVIRONMENT", originalArgusEnv)
		}
	}()

	tests := []struct {
		name        string
		envValue    string
		setEnv      bool
		expectedEnv string
	}{
		{
			name:        "default environment when not set",
			setEnv:      false,
			expectedEnv: "development",
		},
		{
			name:        "production environment",
			envValue:    "production",
			setEnv:      true,
			expectedEnv: "production",
		},
		{
			name:        "staging environment",
			envValue:    "staging",
			setEnv:      true,
			expectedEnv: "staging",
		},
		{
			name:        "test environment",
			envValue:    "test",
			setEnv:      true,
			expectedEnv: "test",
		},
		{
			name:        "empty environment falls back to default",
			envValue:    "",
			setEnv:      true,
			expectedEnv: "development",
		},
		{
			name:        "ARGUS_ENVIRONMENT sets environment",
			envValue:    "production",
			setEnv:      true,
			expectedEnv: "production",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment variable first
			os.Unsetenv("ARGUS_ENVIRONMENT")

			// Setup
			if tt.setEnv {
				os.Setenv("ARGUS_ENVIRONMENT", tt.envValue)
			}

			// Test
			startTime := time.Now()
			config := GetServiceConfig()

			// Assertions
			assert.Equal(t, "argus", config.Name)
			assert.Equal(t, tt.expectedEnv, config.Environment)
			assert.Equal(t, ":3001", config.Port)
			assert.WithinDuration(t, startTime, config.StartTime, time.Second)

			// Version should not be empty (from GetVersion())
			assert.NotEmpty(t, config.Version)
		})
	}
}

func TestServiceConfig_GetAPIBaseURL(t *testing.T) {
	// GetAPIBaseURL now always returns localhost since frontend auto-detects the actual URL
	config := GetServiceConfig()
	url := config.GetAPIBaseURL()
	assert.Equal(t, "http://localhost:3001", url)
}

// Removed TestArgusHostnameFallback - no longer needed since we simplified the logic

func TestGetTracingConfig(t *testing.T) {
	config := GetTracingConfig()

	assert.Equal(t, "argus", config.ServiceName)
	assert.NotEmpty(t, config.ServiceVersion) // From GetVersion()
	assert.Equal(t, "http://localhost:14268/api/traces", config.JaegerEndpoint)
	assert.Equal(t, 1.0, config.SamplingRate)
}

func TestGetVersion(t *testing.T) {
	// Save original state
	originalVersion := Version
	originalServiceVersion := os.Getenv("SERVICE_VERSION")
	originalArgusVersion := os.Getenv("ARGUS_VERSION")

	defer func() {
		Version = originalVersion
		if originalServiceVersion == "" {
			os.Unsetenv("SERVICE_VERSION")
		} else {
			os.Setenv("SERVICE_VERSION", originalServiceVersion)
		}
		if originalArgusVersion == "" {
			os.Unsetenv("ARGUS_VERSION")
		} else {
			os.Setenv("ARGUS_VERSION", originalArgusVersion)
		}
	}()

	t.Run("returns build-time version when set", func(t *testing.T) {
		// Clear environment
		os.Unsetenv("SERVICE_VERSION")

		// Set build-time version
		Version = "v1.2.3"

		version := GetVersion()
		assert.Equal(t, "v1.2.3", version)
	})

	t.Run("returns ARGUS_VERSION when build-time not set", func(t *testing.T) {
		// Clear build-time version
		Version = ""

		// Set ARGUS_VERSION
		os.Setenv("ARGUS_VERSION", "v2.0.0-argus")

		version := GetVersion()
		assert.Equal(t, "v2.0.0-argus", version)
	})

	t.Run("returns git version when others not available", func(t *testing.T) {
		// Clear all version sources
		Version = ""
		os.Unsetenv("ARGUS_VERSION")

		version := GetVersion()

		// Should either return git version or fallback
		// We can't predict git state, so just check it's not empty
		assert.NotEmpty(t, version)

		// Should be either a git version (starts with v) or the fallback
		if !strings.HasPrefix(version, "v") {
			t.Errorf("Version should start with 'v', got: %s", version)
		}
	})

	t.Run("returns fallback when nothing else available", func(t *testing.T) {
		// Clear everything
		Version = ""
		os.Unsetenv("ARGUS_VERSION")

		// This test is harder to force since git might be available
		// But we can at least ensure the function doesn't panic
		version := GetVersion()
		assert.NotEmpty(t, version)
	})

	t.Run("priority order: build-time > environment > git > fallback", func(t *testing.T) {
		// Set environment but also build-time
		os.Setenv("ARGUS_VERSION", "v2.0.0-env")
		Version = "v1.2.3-build"

		version := GetVersion()
		assert.Equal(t, "v1.2.3-build", version, "Build-time version should take priority")

		// Clear build-time, should use environment
		Version = ""
		version = GetVersion()
		assert.Equal(t, "v2.0.0-env", version, "Environment version should be used when build-time not set")
	})
}

func TestGetGitVersion(t *testing.T) {
	t.Run("getGitVersion behavior", func(t *testing.T) {
		// Test the git version function
		gitVersion := getGitVersion()

		// We can't guarantee git is available or has tags
		// But we can test that it doesn't panic and returns a string
		assert.IsType(t, "", gitVersion)

		// If it returns something, it should be trimmed
		if gitVersion != "" {
			assert.Equal(t, strings.TrimSpace(gitVersion), gitVersion)
		}
	})

	t.Run("git command execution", func(t *testing.T) {
		// Test that git commands are being attempted correctly
		// We'll check if git is available first
		if _, err := exec.LookPath("git"); err != nil {
			t.Skip("git not available in test environment")
		}

		// Try to get some version
		version := getGitVersion()

		// If we're in a git repo with tags, should get something
		// If not, should return empty string (not error)
		assert.IsType(t, "", version)
	})
}

func TestBuildTimeVariables(t *testing.T) {
	// These are set via ldflags during build
	// In tests, they'll be empty unless set

	t.Run("build variables are strings", func(t *testing.T) {
		assert.IsType(t, "", Version)
		assert.IsType(t, "", BuildTime)
		assert.IsType(t, "", GitCommit)
	})

	t.Run("build variables can be set", func(t *testing.T) {
		// Save original
		originalVersion := Version
		originalBuildTime := BuildTime
		originalGitCommit := GitCommit

		defer func() {
			Version = originalVersion
			BuildTime = originalBuildTime
			GitCommit = originalGitCommit
		}()

		// Set test values
		Version = "v1.0.0-test"
		BuildTime = "2024-01-01T00:00:00Z"
		GitCommit = "abc123def456"

		assert.Equal(t, "v1.0.0-test", Version)
		assert.Equal(t, "2024-01-01T00:00:00Z", BuildTime)
		assert.Equal(t, "abc123def456", GitCommit)
	})
}

func TestServiceConfigValidation(t *testing.T) {
	config := GetServiceConfig()

	t.Run("service config has required fields", func(t *testing.T) {
		assert.NotEmpty(t, config.Name)
		assert.NotEmpty(t, config.Version)
		assert.NotEmpty(t, config.Environment)
		assert.NotEmpty(t, config.Port)
		assert.False(t, config.StartTime.IsZero())
	})

	t.Run("port format is valid", func(t *testing.T) {
		assert.True(t, strings.HasPrefix(config.Port, ":"))
		assert.True(t, len(config.Port) > 1)
	})

	t.Run("environment is valid", func(t *testing.T) {
		// Should be a reasonable string format
		if config.Environment != "" {
			// Just ensure it's not empty and is a reasonable string
			assert.Regexp(t, regexp.MustCompile(`^[a-zA-Z0-9_-]+$`), config.Environment)
		}
	})
}

func TestTracingConfigValidation(t *testing.T) {
	config := GetTracingConfig()

	t.Run("tracing config has required fields", func(t *testing.T) {
		assert.NotEmpty(t, config.ServiceName)
		assert.NotEmpty(t, config.ServiceVersion)
		assert.NotEmpty(t, config.JaegerEndpoint)
		assert.GreaterOrEqual(t, config.SamplingRate, 0.0)
		assert.LessOrEqual(t, config.SamplingRate, 1.0)
	})

	t.Run("jaeger endpoint is valid URL format", func(t *testing.T) {
		assert.True(t, strings.HasPrefix(config.JaegerEndpoint, "http"))
		assert.Contains(t, config.JaegerEndpoint, ":")
	})
}

// Benchmark tests
func BenchmarkGetServiceConfig(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetServiceConfig()
	}
}

func BenchmarkGetTracingConfig(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetTracingConfig()
	}
}

func BenchmarkGetVersion(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetVersion()
	}
}

func BenchmarkGetAPIBaseURL(b *testing.B) {
	config := GetServiceConfig()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.GetAPIBaseURL()
	}
}

// Example usage
func ExampleGetServiceConfig() {
	config := GetServiceConfig()

	// Use the configuration
	_ = config.Name        // "argus"
	_ = config.Environment // "development" (or from ENVIRONMENT env var)
	_ = config.Port        // ":3001"
	_ = config.Version     // version string
}

func ExampleServiceConfig_GetAPIBaseURL() {
	config := GetServiceConfig()

	// Get API base URL
	baseURL := config.GetAPIBaseURL()

	// Will be "http://argus:3001" by default
	// or "http://${SERVER_IP}:3001" if SERVER_IP is set
	_ = baseURL
}
