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

// GetDefaults returns default LGTM settings for local development
func GetDefaults() *LGTMSettings {
	return &LGTMSettings{
		Grafana: ServiceConfig{
			URL:      getEnv("ARGUS_GRAFANA_URL", "http://localhost:3000"),
			Username: getEnv("ARGUS_GRAFANA_USERNAME", "admin"),
			Password: getEnv("ARGUS_GRAFANA_PASSWORD", ""),
		},
		Prometheus: ServiceConfig{
			URL:      getEnv("ARGUS_PROMETHEUS_URL", "http://localhost:9090"),
			Username: getEnv("ARGUS_PROMETHEUS_USERNAME", ""),
			Password: getEnv("ARGUS_PROMETHEUS_PASSWORD", ""),
		},
		AlertManager: ServiceConfig{
			URL: getEnv("ARGUS_ALERTMANAGER_URL", "http://localhost:9093"),
		},
		Loki: ServiceConfig{
			URL: getEnv("ARGUS_LOKI_URL", "http://localhost:3100"),
		},
		Tempo: ServiceConfig{
			URL: getEnv("ARGUS_TEMPO_URL", "http://localhost:3200"),
		},
	}
}
