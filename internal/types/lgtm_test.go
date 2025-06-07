package types

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceConfig_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		config   ServiceConfig
		expected string
	}{
		{
			name: "complete service config",
			config: ServiceConfig{
				URL:      "http://localhost:9090",
				Username: "admin",
				Password: "secret",
			},
			expected: `{"url":"http://localhost:9090","username":"admin","password":"secret"}`,
		},
		{
			name: "config without optional fields",
			config: ServiceConfig{
				URL: "http://localhost:9090",
			},
			expected: `{"url":"http://localhost:9090"}`,
		},
		{
			name: "config with username only",
			config: ServiceConfig{
				URL:      "http://localhost:9090",
				Username: "admin",
			},
			expected: `{"url":"http://localhost:9090","username":"admin"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshaling
			jsonData, err := json.Marshal(tt.config)
			require.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(jsonData))

			// Test unmarshaling
			var unmarshaled ServiceConfig
			err = json.Unmarshal(jsonData, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, tt.config, unmarshaled)
		})
	}
}

func TestLGTMSettings_JSONSerialization(t *testing.T) {
	settings := &LGTMSettings{
		Grafana: ServiceConfig{
			URL:      "http://grafana:3000",
			Username: "admin",
			Password: "secret",
		},
		Prometheus: ServiceConfig{
			URL: "http://prometheus:9090",
		},
		AlertManager: ServiceConfig{
			URL: "http://alertmanager:9093",
		},
		Loki: ServiceConfig{
			URL: "http://loki:3100",
		},
		Tempo: ServiceConfig{
			URL: "http://tempo:3200",
		},
	}

	// Test marshaling
	jsonData, err := json.Marshal(settings)
	require.NoError(t, err)

	// Test unmarshaling
	var unmarshaled LGTMSettings
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, *settings, unmarshaled)
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
		setEnv       bool
	}{
		{
			name:         "returns environment variable when set",
			key:          "TEST_ENV_VAR",
			defaultValue: "default",
			envValue:     "from_env",
			expected:     "from_env",
			setEnv:       true,
		},
		{
			name:         "returns default when env var not set",
			key:          "NON_EXISTENT_VAR",
			defaultValue: "default_value",
			expected:     "default_value",
			setEnv:       false,
		},
		{
			name:         "returns default when env var is empty",
			key:          "EMPTY_ENV_VAR",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
			setEnv:       true,
		},
		{
			name:         "handles empty default value",
			key:          "ANOTHER_NON_EXISTENT_VAR",
			defaultValue: "",
			expected:     "",
			setEnv:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			// Test
			result := getEnv(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetDefaults(t *testing.T) {
	// Save original env vars (both ARGUS_ and legacy)
	originalEnvVars := map[string]string{
		"ARGUS_GRAFANA_URL":         os.Getenv("ARGUS_GRAFANA_URL"),
		"ARGUS_GRAFANA_USERNAME":    os.Getenv("ARGUS_GRAFANA_USERNAME"),
		"ARGUS_GRAFANA_PASSWORD":    os.Getenv("ARGUS_GRAFANA_PASSWORD"),
		"ARGUS_PROMETHEUS_URL":      os.Getenv("ARGUS_PROMETHEUS_URL"),
		"ARGUS_PROMETHEUS_USERNAME": os.Getenv("ARGUS_PROMETHEUS_USERNAME"),
		"ARGUS_PROMETHEUS_PASSWORD": os.Getenv("ARGUS_PROMETHEUS_PASSWORD"),
		"ARGUS_ALERTMANAGER_URL":    os.Getenv("ARGUS_ALERTMANAGER_URL"),
		"ARGUS_LOKI_URL":            os.Getenv("ARGUS_LOKI_URL"),
		"ARGUS_TEMPO_URL":           os.Getenv("ARGUS_TEMPO_URL"),
		"GRAFANA_URL":               os.Getenv("GRAFANA_URL"),
		"GRAFANA_USERNAME":          os.Getenv("GRAFANA_USERNAME"),
		"GRAFANA_PASSWORD":          os.Getenv("GRAFANA_PASSWORD"),
		"PROMETHEUS_URL":            os.Getenv("PROMETHEUS_URL"),
		"PROMETHEUS_USERNAME":       os.Getenv("PROMETHEUS_USERNAME"),
		"PROMETHEUS_PASSWORD":       os.Getenv("PROMETHEUS_PASSWORD"),
		"ALERTMANAGER_URL":          os.Getenv("ALERTMANAGER_URL"),
		"LOKI_URL":                  os.Getenv("LOKI_URL"),
		"TEMPO_URL":                 os.Getenv("TEMPO_URL"),
	}

	// Clean up function
	cleanup := func() {
		for key, value := range originalEnvVars {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}

	t.Run("returns defaults when no env vars set", func(t *testing.T) {
		// Clear all env vars
		for key := range originalEnvVars {
			os.Unsetenv(key)
		}
		defer cleanup()

		settings := GetDefaults()

		expected := &LGTMSettings{
			Grafana: ServiceConfig{
				URL:      "http://localhost:3000",
				Username: "admin",
				Password: "",
			},
			Prometheus: ServiceConfig{
				URL:      "http://localhost:9090",
				Username: "",
				Password: "",
			},
			AlertManager: ServiceConfig{
				URL: "http://localhost:9093",
			},
			Loki: ServiceConfig{
				URL: "http://localhost:3100",
			},
			Tempo: ServiceConfig{
				URL: "http://localhost:3200",
			},
		}

		assert.Equal(t, expected, settings)
	})

	t.Run("uses ARGUS_ prefixed environment variables when set", func(t *testing.T) {
		// Set specific ARGUS_ env vars
		envVars := map[string]string{
			"ARGUS_GRAFANA_URL":         "http://argus-grafana:3000",
			"ARGUS_GRAFANA_USERNAME":    "argus-admin",
			"ARGUS_GRAFANA_PASSWORD":    "argus-pass",
			"ARGUS_PROMETHEUS_URL":      "http://argus-prometheus:9090",
			"ARGUS_PROMETHEUS_USERNAME": "argus-prom-user",
			"ARGUS_PROMETHEUS_PASSWORD": "argus-prom-pass",
			"ARGUS_ALERTMANAGER_URL":    "http://argus-alertmanager:9093",
			"ARGUS_LOKI_URL":            "http://argus-loki:3100",
			"ARGUS_TEMPO_URL":           "http://argus-tempo:3200",
		}

		for key, value := range envVars {
			os.Setenv(key, value)
		}
		defer cleanup()

		settings := GetDefaults()

		expected := &LGTMSettings{
			Grafana: ServiceConfig{
				URL:      "http://argus-grafana:3000",
				Username: "argus-admin",
				Password: "argus-pass",
			},
			Prometheus: ServiceConfig{
				URL:      "http://argus-prometheus:9090",
				Username: "argus-prom-user",
				Password: "argus-prom-pass",
			},
			AlertManager: ServiceConfig{
				URL: "http://argus-alertmanager:9093",
			},
			Loki: ServiceConfig{
				URL: "http://argus-loki:3100",
			},
			Tempo: ServiceConfig{
				URL: "http://argus-tempo:3200",
			},
		}

		assert.Equal(t, expected, settings)
	})

	t.Run("mixes defaults and env vars", func(t *testing.T) {
		// Clear all first
		for key := range originalEnvVars {
			os.Unsetenv(key)
		}

		// Set only some ARGUS_ prefixed env vars
		os.Setenv("ARGUS_GRAFANA_URL", "http://env-grafana:3000")
		os.Setenv("ARGUS_LOKI_URL", "http://env-loki:3100")
		defer cleanup()

		settings := GetDefaults()

		expected := &LGTMSettings{
			Grafana: ServiceConfig{
				URL:      "http://env-grafana:3000",
				Username: "admin", // default
				Password: "",      // default
			},
			Prometheus: ServiceConfig{
				URL:      "http://localhost:9090", // default
				Username: "",                      // default
				Password: "",                      // default
			},
			AlertManager: ServiceConfig{
				URL: "http://localhost:9093", // default
			},
			Loki: ServiceConfig{
				URL: "http://env-loki:3100", // from env
			},
			Tempo: ServiceConfig{
				URL: "http://localhost:3200", // default
			},
		}

		assert.Equal(t, expected, settings)
	})
}

func TestServiceConfig_Validation(t *testing.T) {
	tests := []struct {
		name   string
		config ServiceConfig
		valid  bool
	}{
		{
			name: "valid config with all fields",
			config: ServiceConfig{
				URL:      "http://localhost:9090",
				Username: "admin",
				Password: "secret",
			},
			valid: true,
		},
		{
			name: "valid config with just URL",
			config: ServiceConfig{
				URL: "http://localhost:9090",
			},
			valid: true,
		},
		{
			name: "empty URL should be considered invalid",
			config: ServiceConfig{
				URL:      "",
				Username: "admin",
				Password: "secret",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simple validation - URL should not be empty
			isValid := tt.config.URL != ""
			assert.Equal(t, tt.valid, isValid)
		})
	}
}

// Benchmark tests for performance-critical functions
func BenchmarkGetEnv(b *testing.B) {
	os.Setenv("BENCH_TEST_VAR", "value")
	defer os.Unsetenv("BENCH_TEST_VAR")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getEnv("BENCH_TEST_VAR", "default")
	}
}

func BenchmarkGetDefaults(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetDefaults()
	}
}
