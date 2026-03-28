package parsers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/bogdanticu88/centipede/internal/models"
)

// GenericJSONParser parses generic JSON log format
type GenericJSONParser struct{}

// GenericLogEntry represents a generic JSON log entry
type GenericLogEntry struct {
	Timestamp   string `json:"timestamp"`
	TenantID    string `json:"tenant_id"`
	Endpoint    string `json:"endpoint"`
	Method      string `json:"method"`
	StatusCode  int    `json:"status_code"`
	PayloadSize int    `json:"payload_size"`
	ResponseMS  int    `json:"response_ms"`
	UserAgent   string `json:"user_agent"`
	SourceIP    string `json:"source_ip"`
}

// Parse parses generic JSON log data
func (p *GenericJSONParser) Parse(data []byte) ([]models.APICall, error) {
	var entries []GenericLogEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("failed to unmarshal generic JSON: %w", err)
	}

	var calls []models.APICall
	for _, entry := range entries {
		timestamp, err := time.Parse(time.RFC3339, entry.Timestamp)
		if err != nil {
			// Try ISO8601 without Z
			timestamp, err = time.Parse("2006-01-02T15:04:05", entry.Timestamp)
			if err != nil {
				continue // Skip malformed entries
			}
		}

		call := models.APICall{
			Timestamp:    timestamp,
			TenantID:     entry.TenantID,
			Endpoint:     entry.Endpoint,
			Method:       entry.Method,
			StatusCode:   entry.StatusCode,
			PayloadSize:  entry.PayloadSize,
			ResponseTime: time.Duration(entry.ResponseMS) * time.Millisecond,
			UserAgent:    entry.UserAgent,
			SourceIP:     entry.SourceIP,
		}

		calls = append(calls, call)
	}

	return calls, nil
}
