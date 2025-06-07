package config

import (
	"os"
	"os/exec"
	"strings"
	"time"
)

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
	environment := os.Getenv("ARGUS_ENVIRONMENT")
	if environment == "" {
		// Fallback to legacy ENVIRONMENT variable
		environment = os.Getenv("ENVIRONMENT")
		if environment == "" {
			environment = "development"
		}
	}

	return &ServiceConfig{
		Name:        "argus",
		Version:     GetVersion(),
		Environment: environment,
		StartTime:   time.Now(),
		Port:        ":3001",
	}
}

// GetAPIBaseURL returns the API base URL based on ARGUS_SERVER_IP environment variable
func (sc *ServiceConfig) GetAPIBaseURL() string {
	serverIP := os.Getenv("ARGUS_SERVER_IP")
	if serverIP == "" {
		// Fallback to legacy SERVER_IP variable
		serverIP = os.Getenv("SERVER_IP")
	}
	if serverIP != "" {
		// Use ARGUS_SERVER_IP (could be localhost for dev, or actual IP for production)
		return "http://" + serverIP + ":3001"
	}

	// Smart defaults based on environment
	if sc.Environment == "development" {
		return "http://localhost:3001"
	}

	// For production/docker, use container name
	return "http://argus:3001"
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

// GetVersion returns the version using modern git-based approach
func GetVersion() string {
	// 1. Build-time injection (preferred - set by Docker/CI)
	if Version != "" {
		return Version
	}

	// 2. Environment variable (for runtime override)
	if env := os.Getenv("ARGUS_VERSION"); env != "" {
		return env
	}

	// Fallback to legacy SERVICE_VERSION variable
	if env := os.Getenv("SERVICE_VERSION"); env != "" {
		return env
	}

	// 3. Try to get from git tags (development)
	if gitVersion := getGitVersion(); gitVersion != "" {
		return gitVersion
	}

	// 4. Ultimate fallback
	return "v0.1.0-dev"
}

// getGitVersion attempts to get version from git tags
func getGitVersion() string {
	// Try git describe --tags
	if cmd := exec.Command("git", "describe", "--tags", "--abbrev=0"); cmd != nil {
		if output, err := cmd.Output(); err == nil {
			return strings.TrimSpace(string(output))
		}
	}

	// Try git describe with fallback
	if cmd := exec.Command("git", "describe", "--tags"); cmd != nil {
		if output, err := cmd.Output(); err == nil {
			return strings.TrimSpace(string(output))
		}
	}

	return ""
}
