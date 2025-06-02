# 👁️ Argus - LGTM Stack Validator

<div align="center">

**The All-Seeing LGTM Stack Testing & Validation Tool**

[![Docker](https://img.shields.io/badge/docker-available-blue)](https://github.com/nahuelsantos/argus/pkgs/container/argus)
[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![GitHub release](https://img.shields.io/github/release/nahuelsantos/argus.svg)](https://github.com/nahuelsantos/argus/releases)

</div>

## 🎯 What is Argus?

**Argus** is a comprehensive testing and validation tool for **LGTM** (Loki, Grafana, Tempo, Prometheus) observability stacks. Named after the Greek giant with a hundred eyes, Argus watches over your monitoring infrastructure to ensure everything works as expected.

### 🔍 Core Purpose

- **Validate LGTM stack configuration** - Ensure all components are properly connected
- **Generate realistic test data** - Metrics, logs, traces, and errors for validation
- **Simulate production workloads** - Web services, APIs, databases, microservices
- **Verify monitoring scenarios** - High load, error conditions, alerting, dashboards

## 🏛️ Why "Argus"?

In Greek mythology, **Argus Panoptes** was a giant with a hundred eyes who could see everything. This perfectly embodies our tool's purpose:

- **👁️ All-seeing**: Monitors every aspect of your LGTM stack
- **🛡️ Guardian**: Protects your monitoring reliability  
- **🔍 Vigilant**: Continuously validates your observability infrastructure
- **⚡ Swift**: Quickly identifies configuration issues

Following the industry tradition of mythological names (Prometheus the Titan, Loki the God), Argus joins as the watchful guardian of your observability.

## ✨ Features

### 🧪 **LGTM Stack Testing**
- **Integration validation** - Complete stack health checks
- **Component connectivity** - Verify Prometheus, Grafana, Loki, Tempo communication
- **Dashboard testing** - Validate Grafana dashboard availability and data flow
- **Alert rule verification** - Test Prometheus alerting configuration

### 📊 **Synthetic Data Generation**
- **Metrics generation** - Custom Prometheus metrics with realistic patterns
- **Log simulation** - Structured and unstructured logs for Loki testing
- **Trace generation** - Distributed traces for Tempo validation
- **Error injection** - Controlled error scenarios for alerting tests

### 🎭 **Workload Simulation**
- **Web service patterns** - WordPress, e-commerce, content sites
- **API service traffic** - REST APIs with authentication, rate limiting
- **Database workloads** - Query patterns, connection pools, slow queries
- **Static site serving** - CDN-like patterns with caching
- **Microservice communication** - Service mesh patterns, circuit breakers

### ⚡ **Performance & Scale Testing**
- **High-volume metrics** - Stress test Prometheus ingestion
- **Log flooding** - Test Loki processing capabilities  
- **Trace generation** - Validate Tempo storage and querying
- **Dashboard load testing** - Ensure Grafana performance under load
- **Resource monitoring** - Track LGTM stack resource consumption

## 🚀 Quick Start

### Using Docker (Recommended)

```bash
# Pull and run Argus
docker run -p 3001:3001 ghcr.io/nahuelsantos/argus:v0.0.1

# Access the dashboard
open http://localhost:3001
```

### Using Docker Compose

```yaml
version: '3.8'
services:
  argus:
    image: ghcr.io/nahuelsantos/argus:v0.0.1
    ports:
      - "3001:3001"
    environment:
      - PROMETHEUS_URL=http://prometheus:9090
      - GRAFANA_URL=http://grafana:3000
      - LOKI_URL=http://loki:3100
      - TEMPO_URL=http://tempo:3200
    networks:
      - monitoring
```

### Building from Source

```bash
git clone https://github.com/nahuelsantos/argus.git
cd argus
go mod download
go run cmd/argus/main.go
```

## 🎛️ Web Dashboard

Argus includes a modern web interface accessible at `http://localhost:3001`:

### 📊 **LGTM Stack Integration Testing**
- **Test LGTM** - Complete stack integration validation
- **Dashboards** - Grafana dashboard availability check
- **Alert Rules** - Prometheus alert configuration verification

### 🚀 **Performance & Scale Testing**  
- **Metrics Scale** - High-volume metrics generation
- **Logs Scale** - Log processing stress testing
- **Traces Scale** - Distributed tracing validation
- **Dashboard Load** - Grafana performance testing
- **Resource Usage** - LGTM stack resource monitoring
- **Storage Limits** - Data retention and storage testing

## 🔧 API Endpoints

### Core Testing
- `GET /health` - Service health check
- `GET /test-lgtm-integration` - Complete LGTM stack validation
- `GET /test-grafana-dashboards` - Dashboard availability testing
- `GET /test-alert-rules` - Alert configuration verification

### Data Generation
- `GET /generate-metrics` - Prometheus metrics generation
- `GET /generate-logs` - Loki log generation
- `GET /generate-error` - Error scenario simulation
- `GET /cpu-load` - CPU stress testing
- `GET /memory-load` - Memory stress testing

### Workload Simulation
- `GET /simulate/web-service` - Web service traffic patterns
- `GET /simulate/api-service` - REST API simulation
- `GET /simulate/database-service` - Database workload patterns
- `GET /simulate/static-site` - Static content serving
- `GET /simulate/microservice` - Microservice communication

### Performance Testing
- `GET /test-metrics-scale` - High-volume metrics testing
- `GET /test-logs-scale` - Log processing scale testing
- `GET /test-traces-scale` - Trace generation and storage
- `GET /test-dashboard-load` - Dashboard performance testing

## ⚙️ Configuration

### Environment Variables

Create a `.env` file from the example:

```bash
cp .env.example .env
```

Edit `.env` with your actual values:

```bash
# LGTM Stack Default Credentials
GRAFANA_USERNAME=admin
GRAFANA_PASSWORD=your-grafana-password
PROMETHEUS_USERNAME=
PROMETHEUS_PASSWORD=

# Alerting Service Credentials
ALERTING_USERNAME=admin
ALERTING_PASSWORD=your-secure-password-here

# LGTM Stack Service URLs (if different from defaults)
GRAFANA_URL=http://localhost:3000
PROMETHEUS_URL=http://localhost:9090
LOKI_URL=http://localhost:3100
TEMPO_URL=http://localhost:3200

# Optional: Override default timeouts (in seconds)
LGTM_TIMEOUT=8
HTTP_TIMEOUT=30

# Optional: Environment and version override
ARGUS_ENVIRONMENT=development
ARGUS_VERSION=v0.0.1
```

### Legacy Environment Variables (Deprecated)

```bash
# LGTM Stack URLs
PROMETHEUS_URL=http://prometheus:9090
GRAFANA_URL=http://grafana:3000  
LOKI_URL=http://loki:3100
TEMPO_URL=http://tempo:3200

# Service Configuration
SERVER_IP=localhost              # For external access
ENVIRONMENT=production           # Environment identifier
SERVICE_VERSION=v1.0.0          # Version tracking

# OpenTelemetry Configuration  
OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318
OTEL_SERVICE_NAME=argus
```

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐
│   Web Dashboard │    │   REST API      │
│   (Port 3001)   │────│   (Port 3001)   │
└─────────────────┘    └─────────────────┘
          │                       │
          └───────────┬───────────┘
                      │
         ┌─────────────────────────┐
         │      Argus Core         │
         │                         │
         │  • Test Generators      │
         │  • LGTM Validators      │
         │  • Workload Simulators  │
         │  • Performance Testers  │
         └─────────────────────────┘
                      │
    ┌─────────────────┼─────────────────┐
    │                 │                 │
┌───▼───┐ ┌──────▼──────┐ ┌──────▼──────┐ ┌───▼────┐
│ Prometheus │ │   Grafana   │ │    Loki     │ │ Tempo  │
│   :9090    │ │    :3000    │ │   :3100     │ │ :3200  │
└───────────┘ └─────────────┘ └─────────────┘ └────────┘
```

## 🎯 Use Cases

### 🔧 **DevOps Engineers**
- **Pre-deployment validation** - Test monitoring before production
- **Infrastructure changes** - Validate monitoring after updates
- **Capacity planning** - Test stack limits and performance
- **Troubleshooting** - Generate controlled scenarios for debugging

### 🏢 **Platform Teams**
- **Monitoring-as-a-Service** - Validate tenant isolation and performance
- **SLA verification** - Test monitoring reliability and response times
- **Multi-environment testing** - Validate dev, staging, production parity
- **Compliance auditing** - Document monitoring capabilities

### 👩‍💻 **Site Reliability Engineers**
- **Chaos engineering** - Test monitoring during failure scenarios
- **Alert tuning** - Validate alert rules and thresholds
- **Runbook validation** - Test monitoring during incident response
- **Performance baselines** - Establish monitoring performance metrics

## 🔍 Testing Scenarios

### **Scenario 1: New LGTM Stack Deployment**
```bash
# Validate complete stack integration
curl http://localhost:3001/test-lgtm-integration

# Test dashboard availability  
curl http://localhost:3001/test-grafana-dashboards

# Verify alert configuration
curl http://localhost:3001/test-alert-rules
```

### **Scenario 2: Performance Validation**
```bash
# Test high-volume metrics ingestion
curl http://localhost:3001/test-metrics-scale

# Validate log processing capabilities
curl http://localhost:3001/test-logs-scale

# Test trace storage and querying  
curl http://localhost:3001/test-traces-scale
```

### **Scenario 3: Production Simulation**
```bash
# Simulate realistic web service traffic
curl http://localhost:3001/simulate/web-service

# Test API service patterns
curl http://localhost:3001/simulate/api-service

# Generate database workload patterns
curl http://localhost:3001/simulate/database-service
```

## 📦 Installation Options

### **GitHub Container Registry**
```bash
docker pull ghcr.io/nahuelsantos/argus:v0.0.1
docker pull ghcr.io/nahuelsantos/argus:latest
```

### **Build from Source**
```bash
git clone https://github.com/nahuelsantos/argus.git
cd argus  
go build -o argus cmd/argus/main.go
./argus
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Setup

```bash
# Clone the repository
git clone https://github.com/nahuelsantos/argus.git
cd argus

# Install dependencies
go mod download

# Run tests
go test ./...

# Run locally
go run cmd/argus/main.go
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Related Projects

- **[Prometheus](https://prometheus.io/)** - Metrics collection and alerting
- **[Grafana](https://grafana.com/)** - Visualization and dashboards  
- **[Loki](https://grafana.com/oss/loki/)** - Log aggregation system
- **[Tempo](https://grafana.com/oss/tempo/)** - Distributed tracing backend

## 📞 Support

- **GitHub Issues**: [Report bugs or request features](https://github.com/nahuelsantos/argus/issues)
- **Documentation**: [Wiki](https://github.com/nahuelsantos/argus/wiki)
- **Discussions**: [Community discussions](https://github.com/nahuelsantos/argus/discussions)

---

<div align="center">

**Built with ❤️ for the LGTM community**

*"With a hundred eyes, Argus sees all - ensuring your monitoring never sleeps."*

</div> 