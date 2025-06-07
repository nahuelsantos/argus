package utils

import "os"

// GetEnvOrDefault returns environment variable or default value
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvWithFallback tries ARGUS_ prefixed variable first, then falls back to legacy variable
func GetEnvWithFallback(argusKey, legacyKey, defaultValue string) string {
	if value := GetEnvOrDefault(argusKey, ""); value != "" {
		return value
	}
	return GetEnvOrDefault(legacyKey, defaultValue)
}

// GetServiceURL returns the appropriate service URL based on environment
func GetServiceURL(service string) string {
	// Use ARGUS_ENVIRONMENT with fallback to legacy ENVIRONMENT
	environment := GetEnvOrDefault("ARGUS_ENVIRONMENT", "")
	if environment == "" {
		environment = GetEnvOrDefault("ENVIRONMENT", "development")
	}

	switch service {
	case "prometheus":
		if environment == "development" {
			return GetEnvWithFallback("ARGUS_PROMETHEUS_URL", "PROMETHEUS_URL", "http://localhost:9090")
		}
		return GetEnvWithFallback("ARGUS_PROMETHEUS_URL", "PROMETHEUS_URL", "http://prometheus:9090")
	case "grafana":
		if environment == "development" {
			return GetEnvWithFallback("ARGUS_GRAFANA_URL", "GRAFANA_URL", "http://localhost:3000")
		}
		return GetEnvWithFallback("ARGUS_GRAFANA_URL", "GRAFANA_URL", "http://grafana:3000")
	case "loki":
		if environment == "development" {
			return GetEnvWithFallback("ARGUS_LOKI_URL", "LOKI_URL", "http://localhost:3100")
		}
		return GetEnvWithFallback("ARGUS_LOKI_URL", "LOKI_URL", "http://loki:3100")
	case "tempo":
		if environment == "development" {
			return GetEnvWithFallback("ARGUS_TEMPO_URL", "TEMPO_URL", "http://localhost:3200")
		}
		return GetEnvWithFallback("ARGUS_TEMPO_URL", "TEMPO_URL", "http://tempo:3200")
	case "alertmanager":
		if environment == "development" {
			return GetEnvWithFallback("ARGUS_ALERTMANAGER_URL", "ALERTMANAGER_URL", "http://localhost:9093")
		}
		return GetEnvWithFallback("ARGUS_ALERTMANAGER_URL", "ALERTMANAGER_URL", "http://alertmanager:9093")
	default:
		return ""
	}
}
