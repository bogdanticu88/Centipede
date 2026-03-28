package models

import "time"

// APICall represents a single API request
type APICall struct {
	Timestamp    time.Time
	TenantID     string
	Endpoint     string
	Method       string
	StatusCode   int
	PayloadSize  int
	ResponseTime time.Duration
	UserAgent    string
	SourceIP     string
}

// TimeWindow represents a time range for normal operations
type TimeWindow struct {
	Start string // "09:00"
	End   string // "17:00"
	Days  string // "Mon-Fri" or "*"
}

// TenantBaseline captures baseline metrics for a tenant
type TenantBaseline struct {
	TenantID          string
	RequestsPerSec    float64
	AvgPayloadSize    float64
	AvgErrorRate      float64
	NormalTimeWindows []TimeWindow
	KnownEndpoints    []string
	LastUpdated       time.Time
}

// Tenant represents a multi-tenant configuration
type Tenant struct {
	ID           string
	Name         string
	Endpoints    []string
	RateLimitRPS int
	Baseline     *TenantBaseline
}

// Anomaly represents a detected anomaly
type Anomaly struct {
	TenantID          string
	Timestamp         time.Time
	Score             int
	Triggers          []string // which rules fired
	Details           map[string]interface{}
	RecommendedAction string // "rate_limit" or "block"
}

// DetectionResult holds results from detection run
type DetectionResult struct {
	Timestamp   time.Time
	Anomalies   []Anomaly
	Summary     map[string]int // e.g., {"total": 10, "critical": 2}
	GeneratedAt time.Time
}

// Alert represents an alert to be sent
type Alert struct {
	Severity  string // "warning", "critical"
	Title     string
	Message   string
	TenantID  string
	Anomaly   *Anomaly
	Timestamp time.Time
}
