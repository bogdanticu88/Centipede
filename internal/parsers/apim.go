package parsers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/bogdanticu88/centipede/internal/models"
)

// APIMParser parses Azure API Management diagnostic logs
type APIMParser struct{}

// APIMLogRecord represents an Azure APIM log record
type APIMLogRecord struct {
	Properties APIMProperties `json:"properties"`
}

// APIMProperties contains the log data
type APIMProperties struct {
	Timestamp        string    `json:"timestamp"`
	ServiceName      string    `json:"serviceNameString"`
	TenantID         string    `json:"backendResponseCode"` // Will extract from context
	RequestURL       string    `json:"requestUri"`
	RequestMethod    string    `json:"httpMethod"`
	APIName          string    `json:"apiName"`
	OperationName    string    `json:"operationName"`
	StatusCode       int       `json:"responseCode"`
	IsRequestSuccess bool      `json:"isRequestSuccess"`
	RequestBodySize  int       `json:"requestBodySize"`
	ResponseBodySize int       `json:"responseBodySize"`
	ResponseTime     int       `json:"responseTimeMs"`
	ClientIPAddress  string    `json:"clientIpAddress"`
	UserAgent        string    `json:"userAgent"`
	UserIdentity     *APIMUser `json:"userIdentity"`
}

// APIMUser contains user/tenant identification
type APIMUser struct {
	Type      string `json:"type"`
	ID        string `json:"id"`
	Principal string `json:"principal"`
}

// Parse parses Azure APIM diagnostic logs
func (p *APIMParser) Parse(data []byte) ([]models.APICall, error) {
	var records []APIMLogRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("failed to unmarshal APIM logs: %w", err)
	}

	var calls []models.APICall
	for _, record := range records {
		props := record.Properties

		timestamp, err := time.Parse(time.RFC3339, props.Timestamp)
		if err != nil {
			continue // Skip malformed entries
		}

		// Extract tenant ID from user identity or use default
		tenantID := "unknown"
		if props.UserIdentity != nil && props.UserIdentity.ID != "" {
			tenantID = props.UserIdentity.ID
		}

		// Extract endpoint path from request URL
		endpoint := extractEndpointPath(props.RequestURL)

		call := models.APICall{
			Timestamp:    timestamp,
			TenantID:     tenantID,
			Endpoint:     endpoint,
			Method:       props.RequestMethod,
			StatusCode:   props.StatusCode,
			PayloadSize:  props.RequestBodySize,
			ResponseTime: time.Duration(props.ResponseTime) * time.Millisecond,
			UserAgent:    props.UserAgent,
			SourceIP:     props.ClientIPAddress,
		}

		calls = append(calls, call)
	}

	return calls, nil
}

// extractEndpointPath extracts the path from a full URL
func extractEndpointPath(url string) string {
	// Simple extraction: get path from URL
	// Example: "https://api.example.com/v1/users/123" -> "/v1/users/123"
	// For now, just return the URL as-is (could use net/url for more precision)
	return url
}

// ParseAPIMLogFormat parses logs in Azure Monitor's table format
// This handles the format returned by Azure Monitor export
func (p *APIMParser) ParseAPIMLogFormat(data []byte) ([]models.APICall, error) {
	// Azure Monitor can export in different formats (JSON array, NDJSON, etc.)
	// Try JSON array first
	return p.Parse(data)
}
