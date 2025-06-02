package types

// LGTMSettings represents the configuration for all LGTM stack services
type LGTMSettings struct {
	Grafana    ServiceConfig `json:"grafana"`
	Prometheus ServiceConfig `json:"prometheus"`
	Loki       ServiceConfig `json:"loki"`
	Tempo      ServiceConfig `json:"tempo"`
}

// ServiceConfig represents the configuration for a single service
type ServiceConfig struct {
	URL      string `json:"url"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// GetDefaults returns default LGTM settings for local development
func GetDefaults() *LGTMSettings {
	return &LGTMSettings{
		Grafana: ServiceConfig{
			URL:      "http://localhost:3000",
			Username: "admin",
			Password: "admin123",
		},
		Prometheus: ServiceConfig{
			URL:      "http://localhost:9090",
			Username: "",
			Password: "",
		},
		Loki: ServiceConfig{
			URL: "http://localhost:3100",
		},
		Tempo: ServiceConfig{
			URL: "http://localhost:3200",
		},
	}
}
