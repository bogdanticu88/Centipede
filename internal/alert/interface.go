package alert

import (
	"context"

	"github.com/bogdanticu88/centipede/internal/models"
)

// Alerter sends alerts via various channels
type Alerter interface {
	Send(ctx context.Context, alert *models.Alert) error
	Name() string
}

// NoOpAlerter is a no-op alerter for testing
type NoOpAlerter struct{}

func (n *NoOpAlerter) Send(ctx context.Context, alert *models.Alert) error {
	return nil
}

func (n *NoOpAlerter) Name() string {
	return "noop"
}
