# CENTIPEDE CI/CD Integration Guide

This guide covers integrating CENTIPEDE into your CI/CD pipeline for automated API security monitoring.

## Quick Start

### GitHub Actions

Run anomaly detection hourly with GitHub Actions:

```yaml
on:
  schedule:
    - cron: '0 * * * *'

jobs:
  detect:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
      - run: go build -o centipede ./cmd/centipede
      - run: ./centipede detect \
          --baseline baseline.json \
          --log-source ./logs \
          --output detections.json
```

See [`.github/workflows/detect.yml`](.github/workflows/detect.yml) for full example.

### Kubernetes CronJob

Deploy to Kubernetes for hourly detection:

```bash
kubectl create namespace security
kubectl apply -f deploy/k8s/centipede-rbac.yaml
kubectl apply -f deploy/k8s/centipede-configmap.yaml
kubectl apply -f deploy/k8s/centipede-pvc.yaml
kubectl apply -f deploy/k8s/centipede-cronjob.yaml
```

See [`deploy/k8s/`](deploy/k8s/) for manifests.

### Docker

Build and run in Docker:

```bash
docker build -t centipede:latest .
docker run -v $(pwd):/data centipede:latest \
  detect --baseline /data/baseline.json --log-source /data/logs
```

## Exit Codes for CI/CD

CENTIPEDE uses exit codes to communicate results:

| Code | Meaning | CI/CD Action |
|------|---------|--------------|
| 0 | Success, no anomalies | Continue pipeline |
| 1 | General error | Fail with error |
| 2 | Config error | Fail - review configuration |
| 3 | Data/input error | Fail - check logs/baseline |
| 4 | Execution error | Fail - check logs |
| 10 | ⚠️ Warning anomalies (score 2-3) | Warn + notify |
| 11 | 🚨 Critical anomalies (score 4+) | **Fail + alert + block** |
| 12 | 🚨 Honeypot hit detected | **Fail + immediate block** |

### Example: Fail on Critical

```bash
./centipede detect --baseline baseline.json --log-source logs
EXIT_CODE=$?

if [ $EXIT_CODE -eq 11 ]; then
  echo "CRITICAL anomalies detected!"
  # Trigger incident response
  exit 1
elif [ $EXIT_CODE -eq 10 ]; then
  echo "Warning anomalies detected"
  # Send Slack notification
  exit 0
fi
```

## Configuration via Environment Variables

Override config.yaml with environment variables:

```bash
# Basic config
export CENTIPEDE_CONFIG=/etc/centipede/config.yaml
export CENTIPEDE_CLOUD=azure

# Alerting
export CENTIPEDE_SLACK_WEBHOOK=https://hooks.slack.com/...

# Logging
export LOG_FORMAT=json      # Output structured JSON logs
export DEBUG=true           # Enable debug logging

./centipede detect --baseline baseline.json --log-source logs
```

### Supported Env Vars

- `CENTIPEDE_CONFIG` — Path to config file
- `CENTIPEDE_CLOUD` — Cloud provider (azure, aws, gcp)
- `CENTIPEDE_SLACK_WEBHOOK` — Slack webhook URL
- `LOG_FORMAT` — Log format (json or text)
- `DEBUG` — Enable debug logging (true/false)

## Structured JSON Logging

When `LOG_FORMAT=json`, CENTIPEDE outputs structured logs for ELK/Splunk:

```bash
export LOG_FORMAT=json
./centipede detect --baseline baseline.json --log-source logs 2>&1 | jq .
```

Output:
```json
{
  "timestamp": "2026-03-28T20:05:26.794593297+02:00",
  "level": "INFO",
  "message": "loading baselines",
  "fields": {
    "baseline_path": "baseline.json"
  }
}
```

### Parse with jq/Splunk

Filter critical events:

```bash
./centipede detect --baseline baseline.json --log-source logs 2>&1 | \
  jq 'select(.level == "WARN" or .level == "ERROR")'
```

Send to Splunk:

```bash
./centipede detect ... 2>&1 | \
  curl -X POST http://splunk:8088/services/collector \
    -H "Authorization: Splunk YOUR_TOKEN" \
    -d @-
```

## GitHub Actions Integration

### Example 1: Hourly Detection with Slack

```yaml
name: Hourly API Security Check

on:
  schedule:
    - cron: '0 * * * *'  # Every hour

jobs:
  detect:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'

      - name: Build
        run: go build -o centipede ./cmd/centipede

      - name: Run detection
        env:
          LOG_FORMAT: json
        run: |
          ./centipede detect \
            --baseline baseline.json \
            --log-source logs \
            --output detections.json

      - name: Check results and alert
        if: always()
        run: |
          EXIT_CODE=$?
          if [ $EXIT_CODE -eq 11 ]; then
            echo "::error::CRITICAL anomalies detected"
            exit 1
          elif [ $EXIT_CODE -eq 10 ]; then
            echo "::warning::Warning anomalies detected"
          fi

      - name: Send Slack notification
        if: always()
        uses: slackapi/slack-github-action@v1
        with:
          webhook-url: ${{ secrets.SLACK_WEBHOOK }}
          payload: |
            {
              "text": "CENTIPEDE Scan",
              "blocks": [{
                "type": "section",
                "text": {
                  "type": "mrkdwn",
                  "text": "Status: ${{ job.status }}"
                }
              }]
            }
```

### Example 2: On Push with PR Comments

```yaml
name: API Security Check (PR)

on:
  pull_request:
    paths:
      - 'api/**'
      - '.github/workflows/detect.yml'

jobs:
  detect:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'

      - name: Build
        run: go build -o centipede ./cmd/centipede

      - name: Run detection
        run: |
          ./centipede detect \
            --baseline baseline.json \
            --log-source api-logs \
            --output detections.json

      - name: Comment on PR
        uses: actions/github-script@v6
        if: always()
        with:
          script: |
            const fs = require('fs');
            const detections = JSON.parse(fs.readFileSync('detections.json', 'utf8'));
            const comment = `## CENTIPEDE Detection Report\n\n
              Total Anomalies: ${detections.Summary.total}\n
              Critical: ${detections.Summary.critical}\n
              Warning: ${detections.Summary.warning}\n`;
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: comment
            })
```

## Kubernetes Deployment

### Prerequisites

```bash
# Create namespace
kubectl create namespace security

# Create Azure Storage secret (for log access)
kubectl create secret generic azure-storage-secret \
  --from-literal=azurestorageaccountname=myaccount \
  --from-literal=azurestorageaccountkey=mykey \
  -n security
```

### Deploy

```bash
# Deploy config, RBAC, storage, and cronjob
kubectl apply -f deploy/k8s/

# Verify deployment
kubectl get cronjob -n security
kubectl get configmap centipede-config -n security -o yaml

# View logs
kubectl logs -n security -l job-name=centipede-detect-<timestamp>

# Manually trigger
kubectl create job centipede-manual-1 --from=cronjob/centipede-detect -n security
```

### Customize

Edit `deploy/k8s/centipede-cronjob.yaml`:

```yaml
spec:
  schedule: "*/30 * * * *"  # Every 30 minutes
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: centipede
            env:
              - name: SLACK_WEBHOOK
                valueFrom:
                  secretKeyRef:
                    name: centipede-secrets
                    key: slack-webhook
```

## GitLab CI/CD

```yaml
centipede-detect:
  image: golang:1.24
  script:
    - go build -o centipede ./cmd/centipede
    - ./centipede detect
        --baseline baseline.json
        --log-source logs
        --output detections.json
  artifacts:
    reports:
      dotenv: detections.env
    paths:
      - detections.json
  only:
    - schedules  # Scheduled job
```

## Jenkins Pipeline

```groovy
pipeline {
  agent any

  triggers {
    cron('H * * * *')  // Every hour
  }

  stages {
    stage('Build') {
      steps {
        sh 'go build -o centipede ./cmd/centipede'
      }
    }

    stage('Detect') {
      steps {
        sh '''
          LOG_FORMAT=json ./centipede detect \
            --baseline baseline.json \
            --log-source logs \
            --output detections.json
        '''
      }
    }

    stage('Report') {
      when {
        expression { currentBuild.result == 'SUCCESS' }
      }
      steps {
        archiveArtifacts artifacts: 'detections.json'
        sh 'jq . detections.json'
      }
    }

    stage('Alert') {
      when {
        expression { currentBuild.result == 'FAILURE' }
      }
      steps {
        slackSend(
          color: 'danger',
          message: 'Critical anomalies detected!'
        )
      }
    }
  }
}
```

## Best Practices

### 1. Store Baselines in Git/S3

```bash
# Check in baseline
git add baseline.json
git commit -m "Update baseline"

# Or sync from S3
aws s3 cp s3://backups/baseline.json ./baseline.json
```

### 2. Version Baselines

```bash
# Keep timestamped baselines
./centipede init ... --output baseline-$(date +%Y%m%d).json

# Keep last N baselines
ls -t baseline-*.json | tail -n +8 | xargs rm
```

### 3. Monitor Pipeline Status

Create a dashboard showing:
- Last detection time
- Anomalies detected (by severity)
- Tenants with issues
- Pipeline health

### 4. Set Alerting Thresholds

```yaml
# config.yaml
detection:
  score_warning: 2    # Alert on score >= 2
  score_critical: 4   # Block on score >= 4
```

### 5. Automate Tenant Response

```bash
# If critical, auto-block
if [ $EXIT_CODE -eq 11 ]; then
  TENANT=$(jq -r '.Anomalies[0].TenantID' detections.json)
  ./centipede kill --tenant $TENANT --reason "Auto-blocked by CI/CD"
fi
```

## Troubleshooting

### Detection Not Running

```bash
# Check cron schedule
kubectl get cronjob -n security -o yaml

# Check job history
kubectl get jobs -n security --sort-by=.metadata.creationTimestamp

# View logs
kubectl logs -n security job/centipede-detect-<id>
```

### No Anomalies Detected

- Verify baseline is recent: `jq '.tenants[].LastUpdated' baseline.json`
- Check honeypots are defined: `jq '.honeypots' config.yaml`
- Review detection thresholds: `jq '.detection' config.yaml`

### False Positives

Lower thresholds:
```yaml
detection:
  volume_threshold: 3.0      # Up from 2.0
  error_rate_threshold: 20.0 # Up from 10.0
```

Then re-run: `./centipede init ... --output baseline-new.json`

## Support

- [README.md](README.md) — Project overview
- [QUICKSTART.md](QUICKSTART.md) — Usage guide
- [GitHub Issues](https://github.com/bogdanticu88/centipede/issues) — Report bugs
