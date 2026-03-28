# CENTIPEDE Deployment Guide

Production deployment guide for CENTIPEDE across different platforms.

## Table of Contents

- [Linux/systemd](#linuxsystemd)
- [Docker](#docker)
- [Kubernetes](#kubernetes)
- [AWS Lambda](#aws-lambda)
- [Azure Container Instances](#azure-container-instances)
- [Configuration](#configuration)
- [Monitoring](#monitoring)

## Linux/systemd

Deploy CENTIPEDE as a systemd service with hourly scheduled detection.

### Prerequisites

```bash
sudo useradd --system --home /var/lib/centipede --shell /usr/sbin/nologin centipede
sudo mkdir -p /var/lib/centipede /etc/centipede
sudo chown -R centipede:centipede /var/lib/centipede
```

### Installation

```bash
# Download and build
git clone https://github.com/bogdanticu88/centipede.git
cd centipede
go build -o centipede ./cmd/centipede

# Install binary
sudo cp centipede /usr/local/bin/
sudo chmod +x /usr/local/bin/centipede

# Install config
sudo cp examples/config.yaml /etc/centipede/
sudo chown centipede:centipede /etc/centipede/config.yaml
sudo chmod 600 /etc/centipede/config.yaml

# Install systemd service and timer
sudo cp deploy/systemd/centipede-detect.service /etc/systemd/system/
sudo cp deploy/systemd/centipede-detect.timer /etc/systemd/system/
sudo systemctl daemon-reload
```

### Configuration

Edit `/etc/centipede/config.yaml`:

```yaml
cloud: azure
azure:
  apim_name: my-apim
  resource_group: my-rg
  subscription_id: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx

detection:
  score_warning: 2
  score_critical: 4

honeypots:
  - path: /admin/debug
    severity: 3
```

### Enable and Start

```bash
# Enable timer to start on boot
sudo systemctl enable centipede-detect.timer

# Start the timer
sudo systemctl start centipede-detect.timer

# Check status
sudo systemctl status centipede-detect.timer

# View next run
sudo systemctl list-timers centipede-detect.timer

# View recent runs
sudo journalctl -u centipede-detect.service -n 50
```

### Monitor

```bash
# Real-time logs
sudo journalctl -u centipede-detect.service -f

# Parse JSON logs
sudo journalctl -u centipede-detect.service -o json | jq '.MESSAGE | fromjson'

# Last run
sudo systemctl status centipede-detect.service

# Manually trigger
sudo systemctl start centipede-detect.service

# Check for anomalies
cat /var/lib/centipede/detections.json | jq '.Anomalies'
```

## Docker

Build and run CENTIPEDE in Docker.

### Build Image

```bash
docker build -t centipede:latest .

# Or with specific version
docker build -t centipede:v0.1.0 .

# Push to registry
docker push myregistry.azurecr.io/centipede:latest
```

### Run Container

```bash
# One-time detection
docker run -v $(pwd)/logs:/logs -v $(pwd)/data:/data centipede:latest \
  detect \
    --baseline /data/baseline.json \
    --log-source /logs \
    --output /data/detections.json

# With custom config
docker run \
  -v $(pwd)/config.yaml:/etc/centipede/config.yaml \
  -v $(pwd)/logs:/data/logs \
  -v $(pwd)/data:/data \
  -e LOG_FORMAT=json \
  centipede:latest \
    detect \
      --config /etc/centipede/config.yaml \
      --baseline /data/baseline.json \
      --log-source /data/logs \
      --output /data/detections.json

# With environment variables
docker run \
  -e CENTIPEDE_CLOUD=azure \
  -e CENTIPEDE_SLACK_WEBHOOK=$SLACK_WEBHOOK \
  -e LOG_FORMAT=json \
  -v /logs:/data/logs \
  centipede:latest detect --baseline baseline.json --log-source /data/logs
```

### Docker Compose

```yaml
version: '3.8'

services:
  centipede:
    image: centipede:latest
    volumes:
      - ./config.yaml:/etc/centipede/config.yaml
      - ./logs:/data/logs
      - ./data:/data
    environment:
      LOG_FORMAT: json
      CENTIPEDE_CLOUD: azure
      CENTIPEDE_SLACK_WEBHOOK: ${SLACK_WEBHOOK}
    command: detect
      --config /etc/centipede/config.yaml
      --baseline /data/baseline.json
      --log-source /data/logs
      --output /data/detections.json
```

Run with:

```bash
docker-compose run --rm centipede
```

## Kubernetes

Deploy CENTIPEDE as a Kubernetes CronJob for automated hourly detection.

### Prerequisites

```bash
# Create namespace
kubectl create namespace security

# Create storage secret for Azure
kubectl create secret generic azure-storage-secret \
  --from-literal=azurestorageaccountname=myaccount \
  --from-literal=azurestorageaccountkey=mykey \
  -n security

# Create alerting secret
kubectl create secret generic centipede-secrets \
  --from-literal=slack-webhook=$SLACK_WEBHOOK \
  -n security
```

### Deploy

```bash
# Deploy all resources
kubectl apply -f deploy/k8s/

# Verify
kubectl get all -n security
kubectl get cronjob -n security
kubectl get pvc -n security
```

### Configure

Edit `deploy/k8s/centipede-configmap.yaml` then:

```bash
kubectl apply -f deploy/k8s/centipede-configmap.yaml
```

### Verify Deployment

```bash
# Watch for cronjob execution
kubectl get pods -n security --watch

# Check last run
kubectl logs -n security -l job-name=centipede-detect-<timestamp>

# View detections
kubectl exec -n security <pod-name> -- cat /data/detections.json

# Manually trigger
kubectl create job centipede-manual-1 \
  --from=cronjob/centipede-detect \
  -n security
```

### Cleanup

```bash
kubectl delete -f deploy/k8s/
kubectl delete namespace security
```

## AWS Lambda

Deploy CENTIPEDE as AWS Lambda for serverless detection (works with smaller log volumes).

### Limitations

- 512 MB max memory
- 15 min timeout
- No persistent storage
- Must download logs in runtime

### Setup

```bash
# Build for Lambda
GOOS=linux GOARCH=amd64 go build -o bootstrap ./cmd/centipede

# Create deployment package
zip centipede.zip bootstrap
```

### Create Lambda Function

```bash
aws lambda create-function \
  --function-name centipede-detect \
  --runtime provided.al2 \
  --role arn:aws:iam::ACCOUNT:role/centipede-lambda-role \
  --handler bootstrap \
  --zip-file fileb://centipede.zip \
  --memory-size 512 \
  --timeout 300 \
  --environment Variables='{
    LOG_FORMAT=json,
    CENTIPEDE_CLOUD=aws
  }' \
  --layers arn:aws:lambda:region:ACCOUNT:layer:centipede-baseline
```

### Schedule with EventBridge

```bash
aws events put-rule \
  --name centipede-hourly \
  --schedule-expression 'rate(1 hour)'

aws events put-targets \
  --rule centipede-hourly \
  --targets "Id"="1","Arn"="arn:aws:lambda:region:ACCOUNT:function:centipede-detect"
```

## Azure Container Instances

Deploy CENTIPEDE as Azure Container Instances for on-demand detection.

### Create Container Image

```bash
# Push to Azure Container Registry
az acr build --registry myregistry --image centipede:latest .
```

### Create Container Instance

```bash
az container create \
  --resource-group my-rg \
  --name centipede-detect \
  --image myregistry.azurecr.io/centipede:latest \
  --registry-login-server myregistry.azurecr.io \
  --registry-username <username> \
  --registry-password <password> \
  --environment-variables \
    CENTIPEDE_CLOUD=azure \
    CENTIPEDE_SLACK_WEBHOOK=$SLACK_WEBHOOK \
    LOG_FORMAT=json \
  --volume-mount-path /data \
  --azure-file-volume-account-name mystorageaccount \
  --azure-file-volume-account-key mystoragekey \
  --azure-file-volume-share-name centipede
```

### Schedule with Logic App

Create Azure Logic App with recurrence trigger → run container → send notification.

## Configuration

### Environment Variables

| Variable | Purpose | Example |
|----------|---------|---------|
| `CENTIPEDE_CONFIG` | Config file path | `/etc/centipede/config.yaml` |
| `CENTIPEDE_CLOUD` | Cloud provider | `azure`, `aws`, `gcp` |
| `CENTIPEDE_SLACK_WEBHOOK` | Slack webhook | `https://hooks.slack.com/...` |
| `LOG_FORMAT` | Log output format | `json` or `text` |
| `DEBUG` | Enable debug logging | `true` or `false` |

### Configuration File

See [examples/config.yaml](examples/config.yaml) for all options.

```yaml
cloud: azure

detection:
  volume_threshold: 2.0
  payload_threshold: 3.0
  error_rate_threshold: 10.0
  score_warning: 2
  score_critical: 4

honeypots:
  - path: /admin/debug
    severity: 3

tenants:
  - id: tenant1
    name: Tenant 1
    endpoints:
      - /api/v1/*
    rate_limit_rps: 1000

alert:
  type: slack
  slack_webhook: $SLACK_WEBHOOK
```

## Monitoring

### Prometheus Metrics (Future)

```bash
# Will expose metrics on :8080/metrics
curl localhost:8080/metrics
```

### Log Aggregation

Forward logs to ELK/Splunk:

```bash
# With JSON output
export LOG_FORMAT=json
centipede detect ... 2>&1 | \
  curl -X POST http://splunk:8088/services/collector \
    -H "Authorization: Splunk TOKEN" \
    -d @-
```

### Alerting Rules

#### On Critical Detection

- Page on-call engineer
- Create incident ticket
- Block tenant automatically
- Send Slack alert

#### On Warning Detection

- Send Slack warning
- Create JIRA ticket
- Review within 24 hours

### Health Checks

```bash
# Is centipede running?
pgrep centipede

# Check last execution
stat -c %y /var/lib/centipede/detections.json

# Are baselines fresh?
find /var/lib/centipede -name "baseline.json" -mtime +7 -print
```

## Troubleshooting

### Service Fails to Start

```bash
# Check for syntax errors
centipede detect --help

# Verify config
centipede init --config /etc/centipede/config.yaml --help

# Check permissions
ls -la /var/lib/centipede
ls -la /etc/centipede/
```

### No Anomalies Detected

- Verify baselines are recent (< 24 hours old)
- Check honeypots in config
- Review detection thresholds
- Check log format matches parser

### Performance Issues

- Increase container memory
- Increase Lambda timeout
- Decrease log volume (filter by date/tenant)
- Use smaller time windows

## Support

- [CICD.md](CICD.md) — CI/CD integration
- [QUICKSTART.md](QUICKSTART.md) — Quick start guide
- [README.md](README.md) — Project overview
