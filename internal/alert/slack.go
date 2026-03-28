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

// SlackAlerter sends alerts to Slack
type SlackAlerter struct {
	webhookURL string
	client     *http.Client
}

// SlackMessage represents a Slack message payload
type SlackMessage struct {
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments"`
}

// Attachment represents a Slack message attachment
type Attachment struct {
	Color  string  `json:"color"`
	Title  string  `json:"title"`
	Text   string  `json:"text"`
	Fields []Field `json:"fields"`
}

// Field represents a Slack attachment field
type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// NewSlackAlerter creates a new Slack alerter
func NewSlackAlerter(webhookURL string) *SlackAlerter {
	return &SlackAlerter{
		webhookURL: webhookURL,
		client:     &http.Client{},
	}
}

// Send sends an alert to Slack
func (s *SlackAlerter) Send(ctx context.Context, alert *models.Alert) error {
	if s.webhookURL == "" {
		return fmt.Errorf("slack webhook URL not configured")
	}

	// Build message
	color := "warning"
	if alert.Severity == "critical" {
		color = "danger"
	}

	fields := []Field{
		{Title: "Tenant", Value: alert.TenantID, Short: true},
		{Title: "Severity", Value: alert.Severity, Short: true},
	}

	if alert.Anomaly != nil {
		fields = append(fields, Field{
			Title: "Score",
			Value: fmt.Sprintf("%d", alert.Anomaly.Score),
			Short: true,
		})
		fields = append(fields, Field{
			Title: "Action",
			Value: alert.Anomaly.RecommendedAction,
			Short: true,
		})
		fields = append(fields, Field{
			Title: "Triggers",
			Value: fmt.Sprintf("%v", alert.Anomaly.Triggers),
			Short: false,
		})
	}

	attachment := Attachment{
		Color:  color,
		Title:  alert.Title,
		Text:   alert.Message,
		Fields: fields,
	}

	message := SlackMessage{
		Text:        fmt.Sprintf("[CENTIPEDE] %s", alert.Title),
		Attachments: []Attachment{attachment},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack message: %w", err)
	}

	// Send to Slack
	req, err := http.NewRequestWithContext(ctx, "POST", s.webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send to Slack: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read Slack response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack webhook returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Name returns the alerter name
func (s *SlackAlerter) Name() string {
	return "slack"
}
