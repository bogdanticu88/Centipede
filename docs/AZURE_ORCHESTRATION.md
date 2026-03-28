# Azure APIM Orchestration Guide

This guide explains how to use CENTIPEDE's Azure API Management orchestration to automatically block and rate-limit compromised tenants.

## Overview

CENTIPEDE can now:
1. **Detect** compromised API clients (anomaly detection)
2. **Block** them in Azure APIM (kill-switch)
3. **Rate-limit** them temporarily (graduated response)
4. **Unblock** them when resolved

## Architecture

```
Anomaly Detection (centipede detect)
           ↓
    Score Anomalies
           ↓
   Exit Code 11 (Critical)
           ↓
Orchestration (centipede kill)
           ↓
    Azure APIM Policy Injection
           ↓
   403 Forbidden for Tenant
```

## Prerequisites

### 1. Azure Configuration

In your `config.yaml`, configure Azure APIM details:

```yaml
cloud: azure
azure:
  apim_name: my-api-management
  resource_group: my-resource-group
  subscription_id: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
  tenant_id: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
```

### 2. Authentication

CENTIPEDE requires an Azure access token to modify APIM policies. You have several options:

#### Option A: Azure CLI (Recommended for Local Development)

```bash
# Authenticate with Azure CLI
az login

# Get an access token
export AZURE_ACCESS_TOKEN=$(az account get-access-token \
  --resource https://management.azure.com \
  --query accessToken -o tsv)

# Run CENTIPEDE
centipede kill --tenant compromised-tenant-id --reason "Anomaly detected"
```

#### Option B: Service Principal (Production)

```bash
# Create a service principal with APIM contributor role
az ad sp create-for-rbac \
  --name centipede-sp \
  --role "API Management Service Contributor" \
  --scope "/subscriptions/<SUBSCRIPTION_ID>/resourceGroups/<RESOURCE_GROUP>"

# Set environment variables
export AZURE_SUBSCRIPTION_ID="..."
export AZURE_TENANT_ID="..."
export AZURE_CLIENT_ID="..."
export AZURE_CLIENT_SECRET="..."

# Get token (implement in your orchestrator or use Azure SDK)
# For now, manually get token and set AZURE_ACCESS_TOKEN
```

#### Option C: Managed Identity (Kubernetes/Container)

If running in Kubernetes with Workload Identity or Azure Container Instances with Managed Identity:

```yaml
# The orchestrator will use DefaultAzureCredential which supports:
# - Managed Identity
# - AZURE_ACCESS_TOKEN environment variable
# - Service Principal environment variables
```

### 3. APIM Configuration

Your API clients **must include the tenant ID** in requests for policy evaluation:

#### Add Tenant ID Header

All API requests must include:
```
X-Tenant-ID: <tenant-id>
```

This can be added in several ways:

**Via API Gateway (Recommended):**
```xml
<!-- In APIM - Inbound policy -->
<policies>
  <inbound>
    <set-header name="X-Tenant-ID" exists-action="override">
      <value>@(context.Variables["tenant-id"])</value>
    </set-header>
  </inbound>
</policies>
```

**Via Client (If clients authenticate with tenant info):**
```bash
curl -H "X-Tenant-ID: tenant-123" https://api.example.com/v1/users
```

**Via JWT Claims (Alternative):**
If using JWT authentication, extract tenant from claims:
```
aud: tenant-123
sub: user@example.com
```

## Usage

### Manual Blocking

Block a tenant immediately:

```bash
centipede kill \
  --config config.yaml \
  --tenant tenant-123 \
  --reason "Suspected compromise - 50x API error spike detected"
```

Output:
```
Cloud Provider: azure

⚠️  KILL-SWITCH ACTIVATION
Tenant: tenant-123
Reason: Suspected compromise - 50x API error spike detected

Using orchestrator: azure-apim
Blocking tenant: tenant-123
✓ Tenant blocked in API Gateway

✓ Kill-switch activated successfully
  Blocked Tenant: tenant-123
  Reason: Suspected compromise - 50x API error spike detected
  Timestamp: 2026-03-28T20:22:24+02:00
```

### Rate-Limiting (Future Enhancement)

Rate-limit a tenant to 10 RPS:

```bash
# Via Slack webhook on warning-level anomaly
# Or via API (when REST API is implemented)
```

### Unblocking

Unblock a tenant when resolved:

```bash
centipede kill \
  --config config.yaml \
  --tenant tenant-123 \
  --unblock \
  --reason "Issue resolved, tenant verified clean"
```

## APIM Policy Details

When you run `centipede kill --tenant <id>`, CENTIPEDE:

1. **Gets an Azure access token** (from environment)
2. **Builds a blocking policy** that checks `X-Tenant-ID` header
3. **Sends APIM API request** to inject/update the policy
4. **Logs the action** with timestamp and reason

### Generated Blocking Policy

```xml
<policies>
  <inbound>
    <base />
    <!-- Block tenant if X-Tenant-ID header matches blocked tenant -->
    <choose>
      <when condition="@(context.Request.Headers.GetValueOrDefault("X-Tenant-ID") == "tenant-123")">
        <return-response>
          <set-status code="403" reason="Access Denied" />
          <set-header name="Content-Type" value="application/json" />
          <set-body>{"error":"Tenant access revoked by security policy","tenant_id":"tenant-123","timestamp":"2026-03-28T20:22:24Z"}</set-body>
        </return-response>
      </when>
    </choose>
  </inbound>
  <outbound>
    <base />
  </outbound>
  <on-error>
    <base />
  </on-error>
</policies>
```

### Rate-Limit Policy

When rate-limiting:

```xml
<policies>
  <inbound>
    <base />
    <!-- Rate limit tenant to N requests per second -->
    <choose>
      <when condition="@(context.Request.Headers.GetValueOrDefault("X-Tenant-ID") == "tenant-123")">
        <rate-limit-by-key calls="10" renewal-period="1"
          counter-key="@(context.Request.Headers.GetValueOrDefault("X-Tenant-ID"))" />
      </when>
    </choose>
  </inbound>
  <outbound>
    <base />
  </outbound>
  <on-error>
    <base />
  </on-error>
</policies>
```

## Integration with Detection Pipeline

### Automatic Blocking on Critical Anomaly

Future enhancement: Auto-block critical anomalies in CI/CD:

```bash
#!/bin/bash
centipede detect --baseline baseline.json --log-source logs

if [ $? -eq 11 ]; then
  # Critical anomaly detected
  TENANT=$(jq -r '.Anomalies[0].TenantID' detections.json)
  centipede kill --tenant $TENANT --reason "Auto-block: critical anomaly detected"
  exit 1  # Fail pipeline
fi
```

### Slack Alerts with Action Buttons

When integrated with Slack:

```json
{
  "attachments": [{
    "color": "danger",
    "title": "CRITICAL: Tenant Anomaly Detected",
    "text": "Tenant sales-prod showing suspicious activity",
    "fields": [
      {"title": "Tenant", "value": "sales-prod"},
      {"title": "Anomaly Score", "value": "8 (critical)"},
      {"title": "Action", "value": "Blocked in APIM"}
    ],
    "actions": [
      {
        "type": "button",
        "text": "Unblock Tenant",
        "value": "unblock_sales_prod"
      },
      {
        "type": "button",
        "text": "View Logs",
        "value": "logs_sales_prod"
      }
    ]
  }]
}
```

## Troubleshooting

### Error: AZURE_ACCESS_TOKEN environment variable not set

```
Error: AZURE_ACCESS_TOKEN environment variable not set
```

**Solution:**
```bash
# Get token from Azure CLI
export AZURE_ACCESS_TOKEN=$(az account get-access-token \
  --resource https://management.azure.com \
  --query accessToken -o tsv)

# Verify token is set
echo $AZURE_ACCESS_TOKEN  # Should print a long JWT-like string

# Re-run command
centipede kill --tenant tenant-123 --reason "Test"
```

### Error: API request failed with status 401 (Unauthorized)

**Causes:**
- Token has expired (valid for 1 hour)
- Token lacks sufficient permissions
- Service Principal has wrong role

**Solutions:**
```bash
# Refresh token
export AZURE_ACCESS_TOKEN=$(az account get-access-token \
  --resource https://management.azure.com \
  --query accessToken -o tsv)

# Check service principal permissions
az role assignment list --scope "/subscriptions/<SUB_ID>/resourceGroups/<RG>/providers/Microsoft.ApiManagement/service/<APIM_NAME>"

# Add missing role if needed
az role assignment create \
  --role "API Management Service Contributor" \
  --assignee <SERVICE_PRINCIPAL_ID> \
  --scope "/subscriptions/<SUB_ID>/resourceGroups/<RG>/providers/Microsoft.ApiManagement/service/<APIM_NAME>"
```

### Error: API request failed with status 404 (Not Found)

**Cause:** APIM instance, resource group, or subscription not found

**Solution:**
```bash
# Verify configuration
echo "Subscription: $AZURE_SUBSCRIPTION_ID"
echo "Resource Group: $(grep resource_group config.yaml)"
echo "APIM Name: $(grep apim_name config.yaml)"

# List available APIM instances
az apim list --resource-group <RG_NAME> --query "[].name"
```

### Blocked Tenant Not Getting 403

**Cause:** Client requests don't include `X-Tenant-ID` header

**Solution:**

1. Verify clients send the header:
```bash
curl -v https://api.example.com/v1/endpoint \
  -H "X-Tenant-ID: tenant-123" 2>&1 | grep X-Tenant-ID
```

2. Add header in APIM if not from client:
```xml
<!-- In APIM - Inbound policy -->
<set-header name="X-Tenant-ID">
  <value>@(context.Variables["tenant-id"])</value>
</set-header>
```

3. Check APIM policy order (policy should be evaluated):
```bash
az apim policy show \
  --resource-group <RG> \
  --name <APIM> \
  --policy-id policy
```

## Audit Trail

All blocking actions are logged with:
- Timestamp
- Tenant ID
- Reason for blocking
- Operation result (success/failure)

### View Logs

```bash
# Kubernetes
kubectl logs -f deployment/centipede

# systemd
sudo journalctl -u centipede-detect.service -f

# Docker
docker logs -f centipede-container

# Set JSON logging for log aggregation
export LOG_FORMAT=json
centipede kill ... 2>&1 | jq '.fields'
```

## Performance

- **Block latency:** < 1 second (REST API call to APIM)
- **Policy evaluation time:** < 10ms per request
- **Unblock latency:** < 1 second

## Security Considerations

1. **Access Token Security:**
   - Tokens are valid for 1 hour
   - Don't commit tokens to version control
   - Use managed identities in production
   - Rotate service principal credentials regularly

2. **Audit Logging:**
   - All blocks are logged with reasons
   - Review audit logs regularly
   - Integrate with SIEM for monitoring

3. **Rate-Limiting as Alternative:**
   - Less severe than blocking
   - Good for warning-level anomalies
   - Gives teams time to investigate

4. **Approval Workflow:**
   - Consider requiring approval before auto-blocking
   - Add confirmation prompts for production
   - Log all approvals and actions

## Future Enhancements

- [ ] Automatic blocking on critical anomaly (score 4+)
- [ ] Slack approval workflow before blocking
- [ ] AWS API Gateway orchestration
- [ ] GCP API Gateway orchestration
- [ ] Prometheus metrics for blocks/rate-limits
- [ ] Incident ticket auto-creation (Jira, GitHub, Azure DevOps)
- [ ] Batch policy management for multiple tenants
- [ ] Policy version history and rollback

## Examples

### Full CI/CD Pipeline with Auto-Blocking

```bash
#!/bin/bash
set -e

# Load baselines
centipede init --config config.yaml \
  --log-source logs/historical \
  --window 7d \
  --output baseline.json

# Run detection
export AZURE_ACCESS_TOKEN=$(az account get-access-token \
  --resource https://management.azure.com \
  --query accessToken -o tsv)

centipede detect \
  --config config.yaml \
  --baseline baseline.json \
  --log-source logs/current \
  --output detections.json

# Check for critical anomalies
if grep -q '"critical":[1-9]' detections.json; then
  TENANT=$(jq -r '.Anomalies[] | select(.Score >= 4) | .TenantID' detections.json | head -1)
  echo "🚨 CRITICAL anomaly detected for $TENANT - blocking immediately"

  centipede kill \
    --config config.yaml \
    --tenant "$TENANT" \
    --reason "Auto-block: critical anomaly score >= 4"

  exit 1
else
  echo "✓ Detection complete - no critical anomalies"
  exit 0
fi
```

### Kubernetes CronJob with Orchestration

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: centipede-detect-and-block
  namespace: security
spec:
  schedule: "0 * * * *"  # Hourly
  jobTemplate:
    spec:
      template:
        spec:
          serviceAccountName: centipede
          containers:
          - name: centipede
            image: centipede:latest
            env:
            - name: AZURE_SUBSCRIPTION_ID
              valueFrom:
                configMapKeyRef:
                  name: centipede-config
                  key: AZURE_SUBSCRIPTION_ID
            - name: AZURE_RESOURCE_GROUP
              valueFrom:
                configMapKeyRef:
                  name: centipede-config
                  key: AZURE_RESOURCE_GROUP
            - name: AZURE_APIM_NAME
              valueFrom:
                configMapKeyRef:
                  name: centipede-config
                  key: AZURE_APIM_NAME
            - name: AZURE_ACCESS_TOKEN
              valueFrom:
                secretKeyRef:
                  name: azure-credentials
                  key: access-token
            - name: CENTIPEDE_SLACK_WEBHOOK
              valueFrom:
                secretKeyRef:
                  name: slack-webhook
                  key: url
            command:
            - /bin/sh
            - -c
            - |
              centipede detect --config /etc/centipede/config.yaml \
                --baseline /data/baseline.json \
                --log-source /data/logs \
                --output /data/detections.json || true

              if grep -q '"critical":[1-9]' /data/detections.json; then
                TENANT=$(jq -r '.Anomalies[] | select(.Score >= 4) | .TenantID' /data/detections.json | head -1)
                centipede kill --tenant "$TENANT" --reason "Auto-block: k8s cron job"
              fi
          restartPolicy: OnFailure
```

