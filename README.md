# CENTIPEDE

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
