package types

import "os"

// LGTMSettings represents the configuration for all LGTM stack services
type LGTMSettings struct {
	Grafana      ServiceConfig `json:"grafana"`
	Prometheus   ServiceConfig `json:"prometheus"`
	AlertManager ServiceConfig `json:"alertmanager"`
	Loki         ServiceConfig `json:"loki"`
	Tempo        ServiceConfig `json:"tempo"`
}

// ServiceConfig represents the configuration for a single service
type ServiceConfig struct {
	URL      string `json:"url"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// getEnv returns environment variable or default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvWithFallback tries ARGUS_ prefixed variable first, then falls back to legacy variable
func getEnvWithFallback(argusKey, legacyKey, defaultValue string) string {
	if value := os.Getenv(argusKey); value != "" {
		return value
	}
	if value := os.Getenv(legacyKey); value != "" {
		return value
	}
	return defaultValue
}

// GetDefaults returns default LGTM settings for local development
func GetDefaults() *LGTMSettings {
	return &LGTMSettings{
		Grafana: ServiceConfig{
			URL:      getEnvWithFallback("ARGUS_GRAFANA_URL", "GRAFANA_URL", "http://localhost:3000"),
			Username: getEnvWithFallback("ARGUS_GRAFANA_USERNAME", "GRAFANA_USERNAME", "admin"),
			Password: getEnvWithFallback("ARGUS_GRAFANA_PASSWORD", "GRAFANA_PASSWORD", ""),
		},
		Prometheus: ServiceConfig{
			URL:      getEnvWithFallback("ARGUS_PROMETHEUS_URL", "PROMETHEUS_URL", "http://localhost:9090"),
			Username: getEnvWithFallback("ARGUS_PROMETHEUS_USERNAME", "PROMETHEUS_USERNAME", ""),
			Password: getEnvWithFallback("ARGUS_PROMETHEUS_PASSWORD", "PROMETHEUS_PASSWORD", ""),
		},
		AlertManager: ServiceConfig{
			URL: getEnvWithFallback("ARGUS_ALERTMANAGER_URL", "ALERTMANAGER_URL", "http://localhost:9093"),
		},
		Loki: ServiceConfig{
			URL: getEnvWithFallback("ARGUS_LOKI_URL", "LOKI_URL", "http://localhost:3100"),
		},
		Tempo: ServiceConfig{
			URL: getEnvWithFallback("ARGUS_TEMPO_URL", "TEMPO_URL", "http://localhost:3200"),
		},
	}
}
