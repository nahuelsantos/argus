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
	// Save original environment variables
	originalEnv := os.Getenv("ENVIRONMENT")
	originalArgusEnv := os.Getenv("ARGUS_ENVIRONMENT")
	defer func() {
		if originalEnv == "" {
			os.Unsetenv("ENVIRONMENT")
		} else {
			os.Setenv("ENVIRONMENT", originalEnv)
		}
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
			name:        "ARGUS_ENVIRONMENT takes precedence",
			envValue:    "production",
			setEnv:      true,
			expectedEnv: "production",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear both environment variables first
			os.Unsetenv("ENVIRONMENT")
			os.Unsetenv("ARGUS_ENVIRONMENT")

			// Setup
			if tt.setEnv {
				if tt.name == "ARGUS_ENVIRONMENT takes precedence" {
					os.Setenv("ARGUS_ENVIRONMENT", tt.envValue)
					os.Setenv("ENVIRONMENT", "should-be-ignored") // Should be ignored
				} else {
					os.Setenv("ENVIRONMENT", tt.envValue)
				}
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
	// Save original environment variables
	originalServerIP := os.Getenv("SERVER_IP")
	originalArgusServerIP := os.Getenv("ARGUS_SERVER_IP")
	originalArgusEnv := os.Getenv("ARGUS_ENVIRONMENT")
	originalEnv := os.Getenv("ENVIRONMENT")
	defer func() {
		if originalServerIP == "" {
			os.Unsetenv("SERVER_IP")
		} else {
			os.Setenv("SERVER_IP", originalServerIP)
		}
		if originalArgusServerIP == "" {
			os.Unsetenv("ARGUS_SERVER_IP")
		} else {
			os.Setenv("ARGUS_SERVER_IP", originalArgusServerIP)
		}
		if originalArgusEnv == "" {
			os.Unsetenv("ARGUS_ENVIRONMENT")
		} else {
			os.Setenv("ARGUS_ENVIRONMENT", originalArgusEnv)
		}
		if originalEnv == "" {
			os.Unsetenv("ENVIRONMENT")
		} else {
			os.Setenv("ENVIRONMENT", originalEnv)
		}
	}()

	tests := []struct {
		name           string
		serverIP       string
		environment    string
		setServerIP    bool
		setEnvironment bool
		expectedURL    string
	}{
		{
			name:        "default development URL when SERVER_IP not set",
			setServerIP: false,
			expectedURL: "http://localhost:3001",
		},
		{
			name:        "localhost development URL",
			serverIP:    "localhost",
			setServerIP: true,
			expectedURL: "http://localhost:3001",
		},
		{
			name:        "production IP address",
			serverIP:    "192.168.1.100",
			setServerIP: true,
			expectedURL: "http://192.168.1.100:3001",
		},
		{
			name:           "production mode with no ARGUS_SERVER_IP uses container name",
			environment:    "production",
			setEnvironment: true,
			setServerIP:    false,
			expectedURL:    "http://argus:3001",
		},
		{
			name:        "ARGUS_SERVER_IP takes precedence over legacy SERVER_IP",
			serverIP:    "argus-preferred.example.com",
			setServerIP: true,
			expectedURL: "http://argus-preferred.example.com:3001",
		},
		{
			name:        "domain name",
			serverIP:    "argus.example.com",
			setServerIP: true,
			expectedURL: "http://argus.example.com:3001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all environment variables first
			os.Unsetenv("SERVER_IP")
			os.Unsetenv("ARGUS_SERVER_IP")
			os.Unsetenv("ARGUS_ENVIRONMENT")
			os.Unsetenv("ENVIRONMENT")

			// Setup
			if tt.setServerIP {
				os.Setenv("ARGUS_SERVER_IP", tt.serverIP)
			}
			if tt.setEnvironment {
				os.Setenv("ARGUS_ENVIRONMENT", tt.environment)
			}

			// Test
			config := GetServiceConfig()
			url := config.GetAPIBaseURL()

			// Assertions
			assert.Equal(t, tt.expectedURL, url)
		})
	}
}

func TestArgusServerIPFallback(t *testing.T) {
	// Save original environment variables
	originalServerIP := os.Getenv("SERVER_IP")
	originalArgusServerIP := os.Getenv("ARGUS_SERVER_IP")

	defer func() {
		if originalServerIP == "" {
			os.Unsetenv("SERVER_IP")
		} else {
			os.Setenv("SERVER_IP", originalServerIP)
		}
		if originalArgusServerIP == "" {
			os.Unsetenv("ARGUS_SERVER_IP")
		} else {
			os.Setenv("ARGUS_SERVER_IP", originalArgusServerIP)
		}
	}()

	t.Run("falls back to legacy SERVER_IP when ARGUS_SERVER_IP not set", func(t *testing.T) {
		// Clear ARGUS_SERVER_IP, set legacy SERVER_IP
		os.Unsetenv("ARGUS_SERVER_IP")
		os.Setenv("SERVER_IP", "legacy.example.com")

		config := GetServiceConfig()
		url := config.GetAPIBaseURL()

		assert.Equal(t, "http://legacy.example.com:3001", url)
	})

	t.Run("ARGUS_SERVER_IP takes precedence over SERVER_IP", func(t *testing.T) {
		// Set both, ARGUS_SERVER_IP should win
		os.Setenv("ARGUS_SERVER_IP", "new.example.com")
		os.Setenv("SERVER_IP", "old.example.com")

		config := GetServiceConfig()
		url := config.GetAPIBaseURL()

		assert.Equal(t, "http://new.example.com:3001", url)
	})
}

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
		os.Unsetenv("SERVICE_VERSION")

		// Set ARGUS_VERSION
		os.Setenv("ARGUS_VERSION", "v2.0.0-argus")

		version := GetVersion()
		assert.Equal(t, "v2.0.0-argus", version)
	})

	t.Run("returns SERVICE_VERSION when ARGUS_VERSION not set", func(t *testing.T) {
		// Clear build-time version and ARGUS_VERSION
		Version = ""
		os.Unsetenv("ARGUS_VERSION")

		// Set legacy SERVICE_VERSION
		os.Setenv("SERVICE_VERSION", "v2.0.0-legacy")

		version := GetVersion()
		assert.Equal(t, "v2.0.0-legacy", version)
	})

	t.Run("returns git version when others not available", func(t *testing.T) {
		// Clear all version sources
		Version = ""
		os.Unsetenv("ARGUS_VERSION")
		os.Unsetenv("SERVICE_VERSION")

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
		os.Unsetenv("SERVICE_VERSION")

		// This test is harder to force since git might be available
		// But we can at least ensure the function doesn't panic
		version := GetVersion()
		assert.NotEmpty(t, version)
	})

	t.Run("priority order: build-time > environment > git > fallback", func(t *testing.T) {
		// Set environment but also build-time
		os.Setenv("SERVICE_VERSION", "v2.0.0-env")
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
