package azure

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/bogdanticu88/centipede/internal/config"
	"github.com/bogdanticu88/centipede/internal/log"
	"github.com/bogdanticu88/centipede/internal/retry"
)

// Orchestrator implements the Orchestrator interface for Azure APIM
type Orchestrator struct {
	config     *config.AzureConfig
	credential azcore.TokenCredential
}

// NewOrchestrator creates a new Azure APIM orchestrator
func NewOrchestrator(cfg *config.AzureConfig) *Orchestrator {
	return &Orchestrator{
		config: cfg,
	}
}

// BlockTenant instantly blocks a tenant in Azure APIM by injecting an inbound policy
func (o *Orchestrator) BlockTenant(ctx context.Context, tenantID string, reason string) error {
	// Validate tenant ID to prevent policy injection
	if err := validateTenantID(tenantID); err != nil {
		return err
	}

	log.Info("blocking tenant in Azure APIM",
		"tenant_id", tenantID,
		"reason", reason,
		"apim_name", o.config.APIMName)

	// Get access token with automatic refresh
	tokenValue, err := o.getAccessToken(ctx)
	if err != nil {
		return err
	}

	// Build APIM policy update URL
	apiURL := fmt.Sprintf(
		"https://management.azure.com/subscriptions/%s/resourceGroups/%s/providers/Microsoft.ApiManagement/service/%s/policies/policy",
		o.config.SubscriptionID,
		o.config.ResourceGroup,
		o.config.APIMName,
	)

	// Create blocking policy XML with escaped tenant ID
	blockingPolicy := buildBlockingPolicy(tenantID)

	// Create request body with policy
	body := fmt.Sprintf(`{
  "properties": {
    "value": %q,
    "format": "xml"
  }
}`, blockingPolicy)

	log.Info("updating APIM policy", "url", apiURL, "tenant_id", tenantID)

	// Make APIM API request with retry logic
	cfg := retry.DefaultConfig()
	cfg.MaxAttempts = 3

	err = retry.Do(ctx, cfg, func() error {
		return o.makePolicyRequest(ctx, tokenValue, "PUT", apiURL, body)
	})

	if err != nil {
		log.Error("failed to apply blocking policy",
			"tenant_id", tenantID,
			"error", err.Error())
		return fmt.Errorf("failed to apply blocking policy: %w", err)
	}

	log.Info("blocking policy applied successfully",
		"tenant_id", tenantID,
		"timestamp", time.Now().Format(time.RFC3339))

	return nil
}

// RateLimitTenant applies rate-limiting to a tenant by updating rate-limit policy
func (o *Orchestrator) RateLimitTenant(ctx context.Context, tenantID string, rps int) error {
	log.Info("applying rate limit to tenant",
		"tenant_id", tenantID,
		"rps", rps,
		"apim_name", o.config.APIMName)

	// Get access token
	tokenValue, err := o.getAccessToken(ctx)
	if err != nil {
		return err
	}

	// Build APIM policy update URL
	apiURL := fmt.Sprintf(
		"https://management.azure.com/subscriptions/%s/resourceGroups/%s/providers/Microsoft.ApiManagement/service/%s/policies/policy",
		o.config.SubscriptionID,
		o.config.ResourceGroup,
		o.config.APIMName,
	)

	// Create rate-limit policy XML
	rateLimitPolicy := buildRateLimitPolicy(tenantID, rps)

	// Create request body with policy
	body := fmt.Sprintf(`{
  "properties": {
    "value": %q,
    "format": "xml"
  }
}`, rateLimitPolicy)

	log.Info("updating APIM rate-limit policy", "tenant_id", tenantID, "rps", rps)

	// Make APIM API request
	if err := o.makePolicyRequest(ctx, tokenValue, "PUT", apiURL, body); err != nil {
		return fmt.Errorf("failed to apply rate-limit policy: %w", err)
	}

	log.Info("rate-limit policy applied successfully",
		"tenant_id", tenantID,
		"rps", rps,
		"timestamp", time.Now().Format(time.RFC3339))

	return nil
}

// UnblockTenant removes blocking policy for a tenant
func (o *Orchestrator) UnblockTenant(ctx context.Context, tenantID string) error {
	log.Info("unblocking tenant in Azure APIM",
		"tenant_id", tenantID,
		"apim_name", o.config.APIMName)

	// Get access token
	tokenValue, err := o.getAccessToken(ctx)
	if err != nil {
		return err
	}

	// Build APIM policy update URL
	apiURL := fmt.Sprintf(
		"https://management.azure.com/subscriptions/%s/resourceGroups/%s/providers/Microsoft.ApiManagement/service/%s/policies/policy",
		o.config.SubscriptionID,
		o.config.ResourceGroup,
		o.config.APIMName,
	)

	// Create unblocking policy (removes the tenant-specific condition)
	unblockPolicy := buildUnblockPolicy(tenantID)

	// Create request body with policy
	body := fmt.Sprintf(`{
  "properties": {
    "value": %q,
    "format": "xml"
  }
}`, unblockPolicy)

	log.Info("removing blocking policy", "tenant_id", tenantID)

	// Make APIM API request
	if err := o.makePolicyRequest(ctx, tokenValue, "PUT", apiURL, body); err != nil {
		return fmt.Errorf("failed to remove blocking policy: %w", err)
	}

	log.Info("tenant unblocked successfully",
		"tenant_id", tenantID,
		"timestamp", time.Now().Format(time.RFC3339))

	return nil
}

// Name returns the orchestrator name
func (o *Orchestrator) Name() string {
	return "azure-apim"
}

// makePolicyRequest makes an authenticated HTTP request to the APIM API
func (o *Orchestrator) makePolicyRequest(ctx context.Context, token string, method string, url string, body string) error {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, url, strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// getAccessToken gets an Azure management API access token using DefaultAzureCredential
// Supports multiple authentication methods (env vars, managed identity, CLI, etc.)
// Automatically handles token refresh for continuous operation
func (o *Orchestrator) getAccessToken(ctx context.Context) (string, error) {
	// Initialize credential if not already done
	if o.credential == nil {
		cred, err := azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			return "", fmt.Errorf("failed to create Azure credential: %w. "+
				"Ensure one of these is configured:\n"+
				"  1. Service principal: AZURE_CLIENT_ID, AZURE_CLIENT_SECRET, AZURE_TENANT_ID\n"+
				"  2. Managed identity: Running in Azure with identity assigned\n"+
				"  3. Azure CLI: Run 'az login' first\n"+
				"  4. Environment: AZURE_TENANT_ID, AZURE_CLIENT_ID (for user-assigned identity)", err)
		}
		o.credential = cred
	}

	// Get token with automatic refresh support
	// Azure SDK automatically refreshes tokens before expiry
	tokenResp, err := o.credential.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{"https://management.azure.com/.default"},
	})
	if err != nil {
		return "", fmt.Errorf("failed to get Azure token: %w", err)
	}

	return tokenResp.Token, nil
}

// validateTenantID validates tenant ID format to prevent policy injection attacks
func validateTenantID(tenantID string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant ID cannot be empty")
	}

	// Tenant IDs should only contain alphanumeric characters, hyphens, and underscores
	// This prevents XML injection and special character issues in policies
	for _, ch := range tenantID {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' ||
			ch == '_') {
			return fmt.Errorf("invalid tenant ID format: contains disallowed character '%c' (only alphanumeric, hyphens, and underscores allowed)", ch)
		}
	}

	return nil
}

// buildBlockingPolicy creates an APIM inbound policy that blocks a specific tenant
func buildBlockingPolicy(tenantID string) string {
	// This policy checks the X-Tenant-ID header and returns 403 if it matches the blocked tenant
	// The policy can be applied at global, API, or operation level
	return fmt.Sprintf(`
<policies>
  <inbound>
    <base />
    <!-- Block tenant if X-Tenant-ID header matches blocked tenant -->
    <choose>
      <when condition="@(context.Request.Headers.GetValueOrDefault("X-Tenant-ID") == "%s")">
        <return-response>
          <set-status code="403" reason="Access Denied" />
          <set-header name="Content-Type" value="application/json" />
          <set-body>{"error":"Tenant access revoked by security policy","tenant_id":"%s","timestamp":"@(DateTime.UtcNow.ToString("O"))"}</set-body>
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
</policies>`, tenantID, tenantID)
}

// buildRateLimitPolicy creates an APIM inbound policy that rate-limits a specific tenant
func buildRateLimitPolicy(tenantID string, rps int) string {
	// This policy applies rate limiting to a specific tenant
	// Uses the rate-limit-by-key policy for per-tenant limiting
	return fmt.Sprintf(`
<policies>
  <inbound>
    <base />
    <!-- Rate limit tenant to %d requests per second -->
    <choose>
      <when condition="@(context.Request.Headers.GetValueOrDefault("X-Tenant-ID") == "%s")">
        <rate-limit-by-key calls="%d" renewal-period="1" counter-key="@(context.Request.Headers.GetValueOrDefault("X-Tenant-ID"))" />
      </when>
    </choose>
  </inbound>
  <outbound>
    <base />
  </outbound>
  <on-error>
    <base />
  </on-error>
</policies>`, rps, tenantID, rps)
}

// buildUnblockPolicy creates a policy that removes tenant-specific blocking
func buildUnblockPolicy(tenantID string) string {
	// This is a standard pass-through policy that doesn't block the tenant
	// It effectively "unblocks" by not applying the blocking condition
	return `
<policies>
  <inbound>
    <base />
    <!-- Tenant unblocked - standard processing continues -->
  </inbound>
  <outbound>
    <base />
  </outbound>
  <on-error>
    <base />
  </on-error>
</policies>`
}
