package cloud

import (
	"context"

	"github.com/bogdanticu88/centipede/internal/models"
)

// CloudProvider abstracts log parsing for different cloud providers
type CloudProvider interface {
	// ParseLogs reads logs from a source and returns APICall slice
	ParseLogs(ctx context.Context, source string) ([]models.APICall, error)

	// GetTenant retrieves tenant configuration
	GetTenant(ctx context.Context, tenantID string) (*models.Tenant, error)

	// Name returns the provider name (azure, aws, gcp)
	Name() string
}

// Orchestrator handles policy enforcement (rate-limiting, blocking, unblocking)
type Orchestrator interface {
	// BlockTenant instantly blocks a tenant
	BlockTenant(ctx context.Context, tenantID string, reason string) error

	// RateLimitTenant applies rate-limiting to a tenant
	RateLimitTenant(ctx context.Context, tenantID string, rps int) error

	// UnblockTenant removes a block
	UnblockTenant(ctx context.Context, tenantID string) error

	// Name returns the orchestrator name
	Name() string
}
