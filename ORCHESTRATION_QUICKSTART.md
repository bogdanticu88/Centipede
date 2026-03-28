# CENTIPEDE Azure APIM Orchestration - Quick Start

## One-Minute Setup

```bash
# 1. Authenticate to Azure
az login

# 2. Get access token
export AZURE_ACCESS_TOKEN=$(az account get-access-token \
  --resource https://management.azure.com --query accessToken -o tsv)

# 3. Update config.yaml with your Azure details
cat > config.yaml <<EOF
cloud: azure
azure:
  apim_name: my-api-management
  resource_group: my-resource-group
  subscription_id: <your-subscription-id>
EOF

# 4. Build and block a tenant
go build -o centipede ./cmd/centipede
./centipede kill \
  --config config.yaml \
  --tenant compromised-tenant \
  --reason "Critical anomaly detected"
```

## What Happens When You Block

1. **Tenant gets 403 Forbidden**
   ```
   All requests with X-Tenant-ID: compromised-tenant → 403
   ```

2. **Action logged**
   ```json
   {
     "timestamp": "2026-03-28T20:22:24Z",
     "tenant": "compromised-tenant",
     "action": "blocked",
     "reason": "Critical anomaly detected"
   }
   ```

3. **Alert sent to Slack**
   ```
   🚨 TENANT BLOCKED in Azure APIM
   Tenant: compromised-tenant
   Reason: Critical anomaly detected
   Timestamp: 2026-03-28T20:22:24Z
   ```

## Core Commands

```bash
# Block a tenant immediately
./centipede kill --tenant T1 --reason "Anomaly detected"

# Unblock a tenant (restore access)
./centipede kill --tenant T1 --unblock

# Detect anomalies (returns exit code 11 if critical)
./centipede detect --baseline baseline.json --log-source logs

# Learn baselines from historical logs
./centipede init --log-source logs/historical --output baseline.json
```

## Integration Example (CI/CD)

```bash
#!/bin/bash
# Detect + Auto-Block Pipeline

export AZURE_ACCESS_TOKEN=$(az account get-access-token \
  --resource https://management.azure.com --query accessToken -o tsv)

# Run detection
./centipede detect --baseline baseline.json --log-source logs
EXIT_CODE=$?

# Auto-block on critical
if [ $EXIT_CODE -eq 11 ]; then
  TENANT=$(jq -r '.Anomalies[0].TenantID' detections.json)
  echo "Blocking tenant: $TENANT"

  ./centipede kill --tenant "$TENANT" \
    --reason "Auto-block: critical anomaly (score >= 4)"

  # Send Slack alert
  curl -X POST $SLACK_WEBHOOK \
    -H 'Content-Type: application/json' \
    -d '{"text":"🚨 Blocked tenant '"$TENANT"' - Critical anomaly"}'

  exit 1
fi

echo "✓ Detection complete - no critical anomalies"
exit 0
```

## Files

**Core Implementation:**
- `internal/cloud/factory.go` - Factory pattern
- `internal/cloud/azure/orchestrator.go` - Blocking logic
- `internal/cmd/kill.go` - CLI command
- `docs/AZURE_ORCHESTRATION.md` - Full documentation

**Tests:**
- `internal/cloud/factory_test.go` - 7 tests
- `internal/cloud/azure/orchestrator_test.go` - 7 tests

## Troubleshooting

**Error: AZURE_ACCESS_TOKEN not set**
```bash
export AZURE_ACCESS_TOKEN=$(az account get-access-token \
  --resource https://management.azure.com --query accessToken -o tsv)
```

**Blocked tenant not getting 403**
- Verify clients send `X-Tenant-ID` header
- Check APIM policies are applied
- Verify tenant ID matches exactly

**Token expired**
- Tokens valid for 1 hour
- Re-run to get fresh token:
```bash
export AZURE_ACCESS_TOKEN=$(az account get-access-token \
  --resource https://management.azure.com --query accessToken -o tsv)
```

## Architecture

```
Anomaly Detected
       ↓
  Score 4+ (Critical)
       ↓
centipede kill --tenant T
       ↓
   Get Access Token
       ↓
   Build APIM Policy
       ↓
   Inject Policy (REST API)
       ↓
   Tenant Gets 403
       ↓
   Action Logged + Alert Sent
```

## Next Steps

1. **Try it locally** - test with a non-production tenant
2. **Integrate with detection** - wire into CI/CD pipeline
3. **Deploy** - use provided Kubernetes/systemd manifests
4. **Monitor** - review Slack alerts and audit logs

## Full Documentation

See `docs/AZURE_ORCHESTRATION.md` for:
- Complete prerequisites
- Policy details
- Kubernetes deployment
- Troubleshooting guide
- Security best practices
