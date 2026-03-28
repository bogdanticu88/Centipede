package alert

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/bogdanticu88/centipede/internal/models"
)

// WebhookAlerter sends alerts via HTTP webhooks
type WebhookAlerter struct {
	webhookURL string
	client     *http.Client
}

// WebhookPayload represents the webhook payload
type WebhookPayload struct {
	EventType string          `json:"event_type"`
	Severity  string          `json:"severity"`
	TenantID  string          `json:"tenant_id"`
	Message   string          `json:"message"`
	Title     string          `json:"title"`
	Anomaly   *models.Anomaly `json:"anomaly,omitempty"`
	Timestamp string          `json:"timestamp"`
}

// NewWebhookAlerter creates a new webhook alerter
func NewWebhookAlerter(webhookURL string) *WebhookAlerter {
	return &WebhookAlerter{
		webhookURL: webhookURL,
		client:     &http.Client{},
	}
}

// Send sends an alert via webhook
func (w *WebhookAlerter) Send(ctx context.Context, alert *models.Alert) error {
	if w.webhookURL == "" {
		return fmt.Errorf("webhook URL not configured")
	}

	payload := WebhookPayload{
		EventType: "anomaly_detected",
		Severity:  alert.Severity,
		TenantID:  alert.TenantID,
		Message:   alert.Message,
		Title:     alert.Title,
		Anomaly:   alert.Anomaly,
		Timestamp: alert.Timestamp.Format("2006-01-02T15:04:05Z"),
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", w.webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read webhook response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Name returns the alerter name
func (w *WebhookAlerter) Name() string {
	return "webhook"
}
