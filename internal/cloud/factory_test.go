package cloud

import (
	"testing"

	"github.com/bogdanticu88/centipede/internal/config"
)

func TestProviderFactoryCreateAzureProvider(t *testing.T) {
	factory := NewProviderFactory()
	cfg := &config.Config{
		Cloud: "azure",
		Azure: &config.AzureConfig{
			APIMName:       "test-apim",
			ResourceGroup:  "test-rg",
			SubscriptionID: "test-sub",
		},
	}

	provider, err := factory.CreateProvider(cfg)
	if err != nil {
		t.Fatalf("CreateProvider failed: %v", err)
	}

	if provider == nil {
		t.Error("provider should not be nil")
	}

	if provider.Name() != "azure" {
		t.Errorf("expected provider name 'azure', got %q", provider.Name())
	}
}

func TestProviderFactoryCreateAzureOrchestrator(t *testing.T) {
	factory := NewProviderFactory()
	cfg := &config.Config{
		Cloud: "azure",
		Azure: &config.AzureConfig{
			APIMName:       "test-apim",
			ResourceGroup:  "test-rg",
			SubscriptionID: "test-sub",
		},
	}

	orchestrator, err := factory.CreateOrchestrator(cfg)
	if err != nil {
		t.Fatalf("CreateOrchestrator failed: %v", err)
	}

	if orchestrator == nil {
		t.Error("orchestrator should not be nil")
	}

	if orchestrator.Name() != "azure-apim" {
		t.Errorf("expected orchestrator name 'azure-apim', got %q", orchestrator.Name())
	}
}

func TestProviderFactoryMissingConfig(t *testing.T) {
	factory := NewProviderFactory()

	_, err := factory.CreateProvider(nil)
	if err == nil {
		t.Error("CreateProvider should fail with nil config")
	}

	_, err = factory.CreateOrchestrator(nil)
	if err == nil {
		t.Error("CreateOrchestrator should fail with nil config")
	}
}

func TestProviderFactoryUnknownProvider(t *testing.T) {
	factory := NewProviderFactory()
	cfg := &config.Config{
		Cloud: "unknown",
	}

	_, err := factory.CreateProvider(cfg)
	if err == nil {
		t.Error("CreateProvider should fail for unknown provider")
	}

	_, err = factory.CreateOrchestrator(cfg)
	if err == nil {
		t.Error("CreateOrchestrator should fail for unknown provider")
	}
}

func TestProviderFactoryAwsNotImplemented(t *testing.T) {
	factory := NewProviderFactory()
	cfg := &config.Config{
		Cloud: "aws",
		AWS: &config.AWSConfig{
			Region: "us-east-1",
		},
	}

	_, err := factory.CreateProvider(cfg)
	if err == nil {
		t.Error("CreateProvider should fail for AWS (not implemented)")
	}

	if !contains(err.Error(), "not yet implemented") {
		t.Errorf("error should mention 'not yet implemented', got: %v", err)
	}
}

func TestProviderFactoryGcpNotImplemented(t *testing.T) {
	factory := NewProviderFactory()
	cfg := &config.Config{
		Cloud: "gcp",
	}

	_, err := factory.CreateProvider(cfg)
	if err == nil {
		t.Error("CreateProvider should fail for GCP (not implemented)")
	}
}

func TestProviderFactoryAzureNoConfig(t *testing.T) {
	factory := NewProviderFactory()
	cfg := &config.Config{
		Cloud: "azure",
		// No Azure config
	}

	_, err := factory.CreateProvider(cfg)
	if err == nil {
		t.Error("CreateProvider should fail when Azure config is missing")
	}

	_, err = factory.CreateOrchestrator(cfg)
	if err == nil {
		t.Error("CreateOrchestrator should fail when Azure config is missing")
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
