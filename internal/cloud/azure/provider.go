package azure

import (
	"context"

	"github.com/bogdanticu88/centipede/internal/config"
	"github.com/bogdanticu88/centipede/internal/models"
	"github.com/bogdanticu88/centipede/internal/parsers"
)

// Provider implements the CloudProvider interface for Azure
type Provider struct {
	config *config.AzureConfig
}

// NewProvider creates a new Azure cloud provider
func NewProvider(cfg *config.AzureConfig) *Provider {
	return &Provider{
		config: cfg,
	}
}

// ParseLogs parses Azure APIM logs using the APIM parser
func (p *Provider) ParseLogs(ctx context.Context, source string) ([]models.APICall, error) {
	// Use APIM parser to handle Azure-specific log formats
	apimParser := &parsers.APIMParser{}
	loader := parsers.NewLoader(apimParser)

	return loader.LoadFromSource(source)
}

// GetTenant retrieves tenant configuration
func (p *Provider) GetTenant(ctx context.Context, tenantID string) (*models.Tenant, error) {
	// For now, return a basic tenant structure
	// In a production system, this would query Azure to get actual tenant config
	return &models.Tenant{
		ID:           tenantID,
		Name:         tenantID,
		RateLimitRPS: 1000, // default
	}, nil
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "azure"
}
