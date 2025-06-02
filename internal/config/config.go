package config

import (
	"embed"
	"os"
	"strings"
	"time"
)

//go:embed VERSION
var versionFS embed.FS

// Build-time variables (set via ldflags)
var (
	Version   = "" // Set via ldflags: -X 'github.com/nahuelsantos/argus/internal/config.Version=v1.0.0'
	BuildTime = "" // Set via ldflags: -X 'github.com/nahuelsantos/argus/internal/config.BuildTime=2024-01-01T00:00:00Z'
	GitCommit = "" // Set via ldflags: -X 'github.com/nahuelsantos/argus/internal/config.GitCommit=abc123'
)

// ServiceConfig holds the service configuration
type ServiceConfig struct {
	Name        string
	Version     string
	Environment string
	StartTime   time.Time
	Port        string
}

// GetServiceConfig returns the current service configuration
func GetServiceConfig() *ServiceConfig {
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "development"
	}

	return &ServiceConfig{
		Name:        "argus",
		Version:     GetVersion(),
		Environment: environment,
		StartTime:   time.Now(),
		Port:        ":3001",
	}
}

// GetAPIBaseURL returns the API base URL based on SERVER_IP environment variable
func (sc *ServiceConfig) GetAPIBaseURL() string {
	serverIP := os.Getenv("SERVER_IP")
	if serverIP == "" {
		// Default to container name for Docker network communication
		return "http://argus:3001"
	}

	// Use SERVER_IP (could be localhost for dev, or actual IP for production)
	return "http://" + serverIP + ":3001"
}

// TracingConfig holds OpenTelemetry configuration
type TracingConfig struct {
	ServiceName    string
	ServiceVersion string
	JaegerEndpoint string
	SamplingRate   float64
}

// GetTracingConfig returns the tracing configuration
func GetTracingConfig() *TracingConfig {
	return &TracingConfig{
		ServiceName:    "argus",
		ServiceVersion: GetVersion(),
		JaegerEndpoint: "http://localhost:14268/api/traces",
		SamplingRate:   1.0,
	}
}

// GetVersion returns the version, preferring build-time injection, then env var, then VERSION file
func GetVersion() string {
	// 1. Build-time injection (preferred)
	if Version != "" {
		return Version
	}

	// 2. Environment variable
	if env := os.Getenv("SERVICE_VERSION"); env != "" {
		return env
	}

	// 3. VERSION file (fallback)
	if data, err := versionFS.ReadFile("VERSION"); err == nil {
		return "v" + strings.TrimSpace(string(data))
	}

	// 4. Ultimate fallback
	return "v0.0.1-dev"
}
