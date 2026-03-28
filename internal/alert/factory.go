package alert

import (
	"fmt"

	"github.com/bogdanticu88/centipede/internal/config"
)

// Factory creates alerters based on configuration
type Factory struct{}

// NewFactory creates a new alerter factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateAlerter creates an alerter based on configuration
func (f *Factory) CreateAlerter(cfg *config.AlertConfig) (Alerter, error) {
	if cfg == nil || cfg.Type == "none" {
		return &NoOpAlerter{}, nil
	}

	switch cfg.Type {
	case "slack":
		if cfg.SlackWebhook == "" {
			return nil, fmt.Errorf("slack webhook URL required for slack alerter")
		}
		return NewSlackAlerter(cfg.SlackWebhook), nil

	case "webhook":
		if cfg.WebhookURL == "" {
			return nil, fmt.Errorf("webhook URL required for webhook alerter")
		}
		return NewWebhookAlerter(cfg.WebhookURL), nil

	default:
		return nil, fmt.Errorf("unknown alerter type: %s", cfg.Type)
	}
}
