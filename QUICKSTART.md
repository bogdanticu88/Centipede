# CENTIPEDE Quick Start Guide

Learn how to use CENTIPEDE to detect API anomalies and protect your multi-tenant platform.

## Overview

CENTIPEDE uses a three-step workflow:

1. **Initialize** — Learn baselines from historical logs
2. **Detect** — Identify anomalies in current logs
3. **Respond** — Take graduated action (rate-limit or block)

## Installation

```bash
git clone https://github.com/bogdanticu88/centipede.git
cd centipede
go build -o centipede ./cmd/centipede
./centipede --help
```

## Step 1: Learn Baselines

Use historical logs (7 days of normal traffic) to establish baseline metrics for each tenant.

```bash
./centipede init \
  --config examples/config.yaml \
  --log-source examples/sample_logs \
  --window 7d \
  --output baseline.json
```

**Output:**
- `baseline.json` — Baseline metrics per tenant (requests/sec, avg payload, error rate, known endpoints)

**Example output:**
```
Loading logs from examples/sample_logs...
Loaded 6 API calls
Learning baselines...
Learned baselines for 2 tenants
Saving baselines to baseline.json

Baseline Summary:
  salesforce-ro:
    Requests/sec: 0.80
    Avg Payload: 1408 bytes
    Error Rate: 0.00%
    Known Endpoints: 3

  sap-de:
    Requests/sec: 2.00
    Avg Payload: 3584 bytes
    Error Rate: 0.00%
    Known Endpoints: 2
```

## Step 2: Run Detection

Analyze current logs against the learned baselines to detect anomalies.

```bash
./centipede detect \
  --config examples/config.yaml \
  --baseline baseline.json \
  --log-source examples/sample_logs/anomaly_sample.json \
  --output detections.json
```

**Output:**
- `detections.json` — Anomalies with scores and recommended actions

**Example output:**
```
Loading baselines from baseline.json...
Loaded 2 baselines
Loading logs from examples/sample_logs/anomaly_sample.json...
Loaded 4 API calls
Running anomaly detection...
Saving detections to detections.json

Detection Results:
  Total Anomalies: 1
  Critical: 1
  Warning: 0
  Normal: 0

Anomalies Detected:
  [11:00:00] salesforce-ro (score: 5) - Action: block
    Triggers: [volume_spike endpoint_anomaly honeypot_hit]
```

### Understanding Anomaly Scores

Each detected anomaly receives a cumulative score:

| Trigger | Score | Threshold |
|---------|-------|-----------|
| Volume Spike | +1 | Requests/sec > 2x baseline |
| Endpoint Anomaly | +1 | New/unknown endpoint |
| Payload Surge | +1 | Avg size > 3x baseline |
| Time Deviation | +1 | Request outside normal hours |
| Error Rate Jump | +1 | 4xx/5xx > 10x baseline |
| Honeypot Hit | +3 | Request to decoy endpoint |

**Score Actions:**
- **0-1:** Normal (monitor)
- **2-3:** Warning → Rate limit tenant
- **4+:** Critical → Block tenant

## Step 3: Take Action

### Check Tenant Status

```bash
./centipede status \
  --config examples/config.yaml \
  --baseline baseline.json \
  --log-source examples/sample_logs
```

### Generate Report

```bash
./centipede report \
  --detections detections.json \
  --baseline baseline.json \
  --format html \
  --output report.html
```

### Emergency Block

When a tenant is confirmed compromised, instantly revoke access:

```bash
./centipede kill \
  --config examples/config.yaml \
  --tenant salesforce-ro \
  --reason "Suspected credential compromise"
```

This triggers:
- Revoke API tokens (IAM)
- Inject blocking policy (API Gateway)
- Create incident ticket
- Log to audit trail

## Configuration

Edit `config.yaml` to customize behavior:

```yaml
cloud: azure

azure:
  apim_name: my-apim
  resource_group: my-rg
  subscription_id: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx

detection:
  volume_threshold: 2.0          # 2x baseline
  payload_threshold: 3.0         # 3x baseline
  error_rate_threshold: 10.0     # 10x baseline
  score_warning: 2
  score_critical: 4

honeypots:
  - path: /admin/debug           # Fake admin endpoint
    severity: 3
  - path: /.env                  # Fake secrets file
    severity: 5

tenants:
  - id: salesforce-ro
    name: Salesforce Romania
    endpoints:
      - /api/crm/*
      - /api/orders/*
    rate_limit_rps: 1000

alert:
  type: slack
  slack_webhook: https://hooks.slack.com/services/YOUR/WEBHOOK/URL
```

## Log Format

CENTIPEDE expects JSON logs with this structure:

```json
[
  {
    "timestamp": "2026-03-28T10:00:00Z",
    "tenant_id": "salesforce-ro",
    "endpoint": "/api/crm/accounts",
    "method": "GET",
    "status_code": 200,
    "payload_size": 1024,
    "response_ms": 50,
    "user_agent": "Mozilla/5.0",
    "source_ip": "192.168.1.100"
  }
]
```

## Example: Detecting a Breach

### Scenario
Tenant `salesforce-ro` account credentials leaked. Attacker is:
- Hitting decoy endpoints (`/admin/debug`)
- Accessing new/unknown endpoints
- Making many more requests than usual

### Baseline
```
salesforce-ro:
  Requests/sec: 0.80
  Avg Payload: 1408 bytes
  Error Rate: 0%
  Known Endpoints: 3
```

### Anomalous Activity
```json
[
  {
    "timestamp": "2026-03-28T11:00:00Z",
    "tenant_id": "salesforce-ro",
    "endpoint": "/admin/debug",
    "method": "GET",
    "status_code": 404,
    "payload_size": 512,
    "response_ms": 25
  },
  {
    "timestamp": "2026-03-28T11:00:01Z",
    "tenant_id": "salesforce-ro",
    "endpoint": "/api/secret/endpoint",
    "method": "GET",
    "status_code": 500,
    "payload_size": 2048,
    "response_ms": 150
  }
]
```

### Detection Result
```
Score: 5 (CRITICAL)
Triggers:
  - honeypot_hit (+3)
  - endpoint_anomaly (+1)
  - volume_spike (+1)

Recommended Action: BLOCK
```

### Response
```bash
./centipede kill \
  --config config.yaml \
  --tenant salesforce-ro \
  --reason "Honeypot hit + suspicious endpoints + leaked credentials"
```

## Continuous Monitoring

Run detection in a loop for real-time protection:

```bash
./centipede monitor \
  --config config.yaml \
  --baseline baseline.json \
  --log-source ./live-logs \
  --alert slack
```

This will:
- Poll for new logs every minute
- Score each tenant
- Send Slack alerts for warnings/critical anomalies
- Run indefinitely (useful for Kubernetes/scheduler)

## Next Steps

1. **Collect Logs** — Export your API gateway logs (APIM, API Gateway, etc.)
2. **Learn Baselines** — Run `init` with 7 days of normal traffic
3. **Test Detection** — Run `detect` to verify scoring is working
4. **Configure Alerting** — Set up Slack/webhook in config.yaml
5. **Enable Monitoring** — Schedule `monitor` as a Kubernetes cronjob or systemd timer

## Troubleshooting

### No anomalies detected?
- Check baseline is recent (run `init` with current logs)
- Verify honeypots are in `config.yaml`
- Increase `score_warning` or `score_critical` thresholds

### Too many false positives?
- Review triggers and adjust thresholds
- Add legitimate endpoints to baseline
- Fine-tune `volume_threshold`, `payload_threshold`, `error_rate_threshold`

### Logs not loading?
- Check log file format (must be JSON array)
- Verify timestamp format (RFC3339: `2026-03-28T10:00:00Z`)
- Check tenant IDs match config

## Advanced Topics

### Multi-Cloud (Phase 2)

```bash
# AWS API Gateway
centipede init --config aws.yaml --log-source s3://my-bucket/logs
```

### Custom Parsers

Implement the `Parser` interface to support custom log formats:

```go
type Parser interface {
	Parse(data []byte) ([]models.APICall, error)
}
```

### Metrics Export

Prometheus metrics (coming in Phase 3):

```
centipede_anomalies_total{tenant="salesforce-ro",severity="critical"} 5
centipede_tenant_score{tenant="salesforce-ro"} 3.5
```

## Support

- **Issues:** https://github.com/bogdanticu88/centipede/issues
- **Docs:** https://github.com/bogdanticu88/centipede/blob/main/README.md
- **Examples:** See `examples/` directory
