package cloud

import (
	"fmt"

	"github.com/bogdanticu88/centipede/internal/cloud/azure"
	"github.com/bogdanticu88/centipede/internal/config"
)

// ProviderFactory creates cloud provider and orchestrator implementations
type ProviderFactory struct{}

// NewProviderFactory creates a new provider factory
func NewProviderFactory() *ProviderFactory {
	return &ProviderFactory{}
}

// CreateProvider creates a CloudProvider based on configuration
func (f *ProviderFactory) CreateProvider(cfg *config.Config) (CloudProvider, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}

	switch cfg.Cloud {
	case "azure":
		if cfg.Azure == nil {
			return nil, fmt.Errorf("azure configuration required for azure provider")
		}
		return azure.NewProvider(cfg.Azure), nil

	case "aws":
		return nil, fmt.Errorf("aws provider not yet implemented")

	case "gcp":
		return nil, fmt.Errorf("gcp provider not yet implemented")

	default:
		return nil, fmt.Errorf("unknown cloud provider: %s", cfg.Cloud)
	}
}

// CreateOrchestrator creates an Orchestrator based on configuration
func (f *ProviderFactory) CreateOrchestrator(cfg *config.Config) (Orchestrator, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}

	switch cfg.Cloud {
	case "azure":
		if cfg.Azure == nil {
			return nil, fmt.Errorf("azure configuration required for azure orchestrator")
		}
		return azure.NewOrchestrator(cfg.Azure), nil

	case "aws":
		return nil, fmt.Errorf("aws orchestrator not yet implemented")

	case "gcp":
		return nil, fmt.Errorf("gcp orchestrator not yet implemented")

	default:
		return nil, fmt.Errorf("unknown cloud provider: %s", cfg.Cloud)
	}
}
