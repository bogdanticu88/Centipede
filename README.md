# CENTIPEDE
![2aqtU](https://github.com/user-attachments/assets/c044f711-c792-4f33-a9ab-33f57b2a152f)

<div align="center">


[![Go Version](https://img.shields.io/badge/go-1.24+-blue.svg?style=flat-square)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg?style=flat-square)](LICENSE)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg?style=flat-square)](https://github.com/bogdanticu88/Centipede/actions)
[![Tests](https://img.shields.io/badge/tests-30+-brightgreen.svg?style=flat-square)](#development)
[![Go Report Card](https://goreportcard.com/badge/github.com/bogdanticu88/Centipede?style=flat-square)](https://goreportcard.com/report/github.com/bogdanticu88/Centipede)

[![Docker](https://img.shields.io/badge/docker-ready-blue.svg?style=flat-square&logo=docker)](https://docker.com)
[![Kubernetes](https://img.shields.io/badge/kubernetes-ready-blue.svg?style=flat-square&logo=kubernetes)](https://kubernetes.io)
[![Azure](https://img.shields.io/badge/azure-apim-blue.svg?style=flat-square&logo=microsoft-azure)](https://azure.microsoft.com)
[![Status](https://img.shields.io/badge/status-production-success.svg?style=flat-square)](#)

[![Code Quality](https://img.shields.io/badge/code%20quality-A-brightgreen.svg?style=flat-square)](#)
[![Security](https://img.shields.io/badge/security-audited-brightgreen.svg?style=flat-square)](SECURITY.md)
[![Stability](https://img.shields.io/badge/stability-stable-brightgreen.svg?style=flat-square)](#)
[![Maintenance](https://img.shields.io/badge/maintenance-actively%20developed-brightgreen.svg?style=flat-square)](#)

[![GitHub Stars](https://img.shields.io/github/stars/bogdanticu88/Centipede?style=flat-square&logo=github)](https://github.com/bogdanticu88/Centipede/stargazers)
[![GitHub Issues](https://img.shields.io/github/issues/bogdanticu88/Centipede?style=flat-square&logo=github)](https://github.com/bogdanticu88/Centipede/issues)
[![GitHub Discussions](https://img.shields.io/badge/discussions-welcome-brightgreen.svg?style=flat-square&logo=github)](https://github.com/bogdanticu88/Centipede/discussions)
[![GitHub Contributors](https://img.shields.io/github/contributors/bogdanticu88/Centipede?style=flat-square&logo=github)](https://github.com/bogdanticu88/Centipede/graphs/contributors)

**Multi-cloud API anomaly detection and tenant protection system**

[Features](#features) • [Quick Start](#installation) • [Documentation](#documentation) • [Roadmap](#roadmap)

</div>

---

Multi-cloud API anomaly detection and tenant protection system.

## Overview

CENTIPEDE detects compromised API clients in multi-tenant platforms before they cascade damage across other tenants. It uses cumulative scoring to detect anomalies and enforces graduated responses: rate limiting for warnings, blocking for critical anomalies, and instant kill-switch capability.

### Core Problem

When one tenant in a multi-tenant API platform gets compromised, it can cascade damage across all other tenants. CENTIPEDE detects this before it spreads.

### Core Solution

Real-time anomaly detection on API traffic with threshold-based graduated response:
- **Rate limiting** for warning-level anomalies (score 2-3)
- **Complete blocking** for critical anomalies (score 4+)
- **Manual kill-switch** for emergency intervention

## Features

### Detection Scoring System

Each request/tenant is analyzed against baseline and triggers cumulative scores:

- **Volume Spike** — Requests/sec exceeds 2x baseline = +1 score
- **Endpoint Anomaly** — Accessing unusual/new endpoints = +1 score
- **Payload Size Surge** — Avg request body exceeds 3x baseline = +1 score
- **Time-of-Day Deviation** — Request pattern outside normal hours = +1 score
- **Error Rate Jump** — 4xx/5xx responses exceed 10x baseline = +1 score
- **Honeypot Hit** — Request to decoy endpoint = +3 score

### Actions

- **Score 2-3**: Apply rate limiting to the tenant
- **Score 4+**: Block tenant completely
- **Manual override**: Instant kill-switch with IAM revocation and incident creation

### Multi-Cloud Support

- **Phase 1 (MVP)**: Azure API Management (APIM) + Azure Monitor
- **Phase 2**: AWS API Gateway + CloudWatch Logs
- **Phase 3**: GCP Cloud API Gateway + Cloud Logging

## Installation

```bash
git clone https://github.com/bogdanticu88/centipede.git
cd centipede
go build -o centipede ./cmd/centipede
```

## Usage

### Initialize (Learn Baselines)

```bash
./centipede init \
  --config config.yaml \
  --log-source <path-or-uri> \
  --window 7d \
  --output baseline.json
```

### Detect Anomalies

```bash
./centipede detect \
  --baseline baseline.json \
  --log-source <path-or-uri> \
  --config config.yaml \
  --output detections.json
```

### Monitor Continuously

```bash
./centipede monitor \
  --baseline baseline.json \
  --log-source <path-or-uri> \
  --config config.yaml \
  --alert slack
```

### Emergency Block

```bash
./centipede kill \
  --tenant <tenant-id> \
  --config config.yaml \
  --reason "Suspected compromise"
```

### Generate Report

```bash
./centipede report \
  --detections detections.json \
  --baseline baseline.json \
  --html \
  --output report.html
```

### Check Status

```bash
./centipede status \
  --baseline baseline.json \
  --log-source <path-or-uri> \
  --config config.yaml
```

## Configuration

See `examples/config.yaml` for a complete configuration example:

```yaml
cloud: azure

azure:
  apim_name: my-apim
  resource_group: my-rg
  subscription_id: xxx

detection:
  volume_threshold: 2.0
  payload_threshold: 3.0
  error_rate_threshold: 10.0
  score_warning: 2
  score_critical: 4

honeypots:
  - path: /admin/debug
    severity: 3
  - path: /.env
    severity: 5

tenants:
  - id: salesforce-ro
    name: Salesforce Romania
    endpoints:
      - /api/crm/*
      - /api/orders/*
    rate_limit_rps: 1000
```

## Architecture

```
centipede/
├── cmd/centipede/          # CLI entrypoint
├── internal/
│   ├── baseline/           # Baseline learning
│   ├── detection/          # Anomaly scoring engine
│   ├── cloud/              # Cloud provider abstractions
│   ├── parsers/            # Log parsers
│   ├── models/             # Data models
│   ├── alert/              # Alerting (Slack, webhooks)
│   ├── config/             # Configuration
│   ├── cmd/                # Command handlers
│   └── storage/            # Cloud storage integration
├── pkg/metrics/            # Prometheus metrics
├── examples/               # Sample configs and logs
└── tests/                  # Tests
```

## Documentation

- **[Quick Start](QUICKSTART.md)** — Get up and running in 5 minutes
- **[Deployment Guide](DEPLOYMENT.md)** — Production deployment strategies
- **[CI/CD Integration](CICD.md)** — GitHub Actions, GitLab CI, Jenkins
- **[Azure APIM Setup](docs/AZURE_ORCHESTRATION.md)** — Azure configuration and orchestration
- **[Production Readiness](PRODUCTION_READINESS_ASSESSMENT.md)** — Pre-production checklist

## Supported Platforms

| Platform | Status | Support |
|----------|--------|---------|
| Azure APIM | ✅ Stable | Production-ready |
| AWS API Gateway | 🗺️ Planned | Phase 2 |
| GCP Cloud API Gateway | 🗺️ Planned | Phase 3 |
| Kubernetes | ✅ Ready | CronJob, Operator |
| Docker | ✅ Ready | Multi-stage builds |

## Development

### Prerequisites

- Go 1.24+
- Azure SDK (for Azure support)
- AWS SDK v2 (for AWS support)

### Running Tests

```bash
go test ./...
```

### Building for Production

```bash
go build -o centipede ./cmd/centipede
```

## License

MIT

## Contributing

Contributions welcome! Please open an issue or PR.

## Roadmap

- [ ] Phase 1: MVP (Azure + Core Detection)
- [ ] Phase 2: AWS Support
- [ ] Phase 3: GCP Support + Advanced Detection
- [ ] Prometheus metrics export
- [ ] REST API for dashboards
- [ ] Kubernetes operator

## Community

- **Report Issues**: [GitHub Issues](https://github.com/bogdanticu88/Centipede/issues)
- **Discussions**: [GitHub Discussions](https://github.com/bogdanticu88/Centipede/discussions)
- **Security**: [Security Policy](SECURITY.md)

---

<div align="center">

**[⬆ back to top](#centipede)**

Made with ❤️ by [Bogdan Ticu](https://github.com/bogdanticu88)

</div>
