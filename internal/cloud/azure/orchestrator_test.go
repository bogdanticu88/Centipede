package azure

import (
	"context"
	"os"
	"testing"

	"github.com/bogdanticu88/centipede/internal/config"
)

func TestValidateTenantID(t *testing.T) {
	tests := []struct {
		tenantID  string
		shouldErr bool
		name      string
	}{
		{"valid-tenant-123", false, "valid tenant ID with hyphens"},
		{"valid_tenant_123", false, "valid tenant ID with underscores"},
		{"validtenant123", false, "valid tenant ID alphanumeric"},
		{"", true, "empty tenant ID"},
		{"tenant@example.com", true, "tenant ID with @ symbol"},
		{"tenant:block", true, "tenant ID with colon"},
		{"tenant<script>", true, "tenant ID with XML characters"},
		{"tenant; DROP TABLE", true, "tenant ID with SQL injection attempt"},
		{"tenant|pipe", true, "tenant ID with pipe character"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTenantID(tt.tenantID)
			if (err != nil) != tt.shouldErr {
				t.Errorf("validateTenantID(%q): expected error=%v, got error=%v", tt.tenantID, tt.shouldErr, err != nil)
			}
		})
	}
}

func TestBuildBlockingPolicy(t *testing.T) {
	tenantID := "test-tenant-123"
	policy := buildBlockingPolicy(tenantID)

	// Verify policy contains tenant ID
	if !contains(policy, tenantID) {
		t.Errorf("blocking policy should contain tenant ID, got: %s", policy)
	}

	// Verify policy is valid XML
	if !contains(policy, "<policies>") || !contains(policy, "</policies>") {
		t.Error("blocking policy should be valid XML")
	}

	// Verify policy checks X-Tenant-ID header
	if !contains(policy, "X-Tenant-ID") {
		t.Error("blocking policy should check X-Tenant-ID header")
	}

	// Verify policy returns 403
	if !contains(policy, "403") {
		t.Error("blocking policy should return 403 status code")
	}
}

func TestBuildRateLimitPolicy(t *testing.T) {
	tenantID := "test-tenant-456"
	rps := 100
	policy := buildRateLimitPolicy(tenantID, rps)

	// Verify policy contains tenant ID
	if !contains(policy, tenantID) {
		t.Errorf("rate-limit policy should contain tenant ID")
	}

	// Verify policy is valid XML
	if !contains(policy, "<policies>") || !contains(policy, "</policies>") {
		t.Error("rate-limit policy should be valid XML")
	}

	// Verify policy contains rate limit value
	if !contains(policy, "100") {
		t.Errorf("rate-limit policy should contain RPS value")
	}
}

func TestBuildUnblockPolicy(t *testing.T) {
	policy := buildUnblockPolicy("test-tenant")

	// Verify policy is valid XML
	if !contains(policy, "<policies>") || !contains(policy, "</policies>") {
		t.Error("unblock policy should be valid XML")
	}

	// Verify it doesn't have restrictive conditions
	if contains(policy, "403") {
		t.Error("unblock policy should not return 403")
	}
}

func TestOrchestratorName(t *testing.T) {
	cfg := &config.AzureConfig{
		APIMName:       "test-apim",
		ResourceGroup:  "test-rg",
		SubscriptionID: "test-sub",
	}
	orchestrator := NewOrchestrator(cfg)

	expected := "azure-apim"
	if orchestrator.Name() != expected {
		t.Errorf("expected Name() to return %q, got %q", expected, orchestrator.Name())
	}
}

func TestOrchestratorBlockTenantNoToken(t *testing.T) {
	// Unset all Azure credential environment variables
	os.Unsetenv("AZURE_TENANT_ID")
	os.Unsetenv("AZURE_CLIENT_ID")
	os.Unsetenv("AZURE_CLIENT_SECRET")
	os.Unsetenv("AZURE_ACCESS_TOKEN")

	cfg := &config.AzureConfig{
		APIMName:       "test-apim",
		ResourceGroup:  "test-rg",
		SubscriptionID: "test-sub",
	}
	orchestrator := NewOrchestrator(cfg)

	ctx := context.Background()
	err := orchestrator.BlockTenant(ctx, "test-tenant", "test reason")

	// Should fail due to missing credentials
	if err == nil {
		t.Error("BlockTenant should fail when Azure credentials are not set")
	}

	// Check that error is about token acquisition failure
	if !contains(err.Error(), "token") && !contains(err.Error(), "credential") {
		t.Errorf("error should mention token or credential failure, got: %v", err)
	}
}

func TestOrchestratorRateLimitTenantNoToken(t *testing.T) {
	// Ensure AZURE_ACCESS_TOKEN is not set
	os.Unsetenv("AZURE_ACCESS_TOKEN")

	cfg := &config.AzureConfig{
		APIMName:       "test-apim",
		ResourceGroup:  "test-rg",
		SubscriptionID: "test-sub",
	}
	orchestrator := NewOrchestrator(cfg)

	ctx := context.Background()
	err := orchestrator.RateLimitTenant(ctx, "test-tenant", 100)

	// Should fail due to missing token
	if err == nil {
		t.Error("RateLimitTenant should fail when AZURE_ACCESS_TOKEN is not set")
	}
}

func TestOrchestratorUnblockTenantNoToken(t *testing.T) {
	// Ensure AZURE_ACCESS_TOKEN is not set
	os.Unsetenv("AZURE_ACCESS_TOKEN")

	cfg := &config.AzureConfig{
		APIMName:       "test-apim",
		ResourceGroup:  "test-rg",
		SubscriptionID: "test-sub",
	}
	orchestrator := NewOrchestrator(cfg)

	ctx := context.Background()
	err := orchestrator.UnblockTenant(ctx, "test-tenant")

	// Should fail due to missing token
	if err == nil {
		t.Error("UnblockTenant should fail when AZURE_ACCESS_TOKEN is not set")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			findSubstring(s, substr)))
}

// Helper function for substring search
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
