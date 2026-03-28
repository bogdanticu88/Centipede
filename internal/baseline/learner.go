package baseline

import (
	"fmt"
	"time"

	"github.com/bogdanticu88/centipede/internal/models"
)

// Learner computes baselines from historical API calls
type Learner struct{}

// NewLearner creates a new baseline learner
func NewLearner() *Learner {
	return &Learner{}
}

// LearnBaseline computes baseline metrics from a set of API calls
func (l *Learner) LearnBaseline(calls []models.APICall) map[string]*models.TenantBaseline {
	baselines := make(map[string]*models.TenantBaseline)

	// Group calls by tenant
	callsByTenant := make(map[string][]models.APICall)
	for _, call := range calls {
		callsByTenant[call.TenantID] = append(callsByTenant[call.TenantID], call)
	}

	// Compute baseline for each tenant
	for tenantID, tenantCalls := range callsByTenant {
		baselines[tenantID] = l.computeTenantBaseline(tenantID, tenantCalls)
	}

	return baselines
}

// computeTenantBaseline computes metrics for a single tenant
func (l *Learner) computeTenantBaseline(tenantID string, calls []models.APICall) *models.TenantBaseline {
	baseline := &models.TenantBaseline{
		TenantID:       tenantID,
		KnownEndpoints: []string{},
		LastUpdated:    time.Now(),
	}

	if len(calls) == 0 {
		return baseline
	}

	// Compute RequestsPerSec
	baseline.RequestsPerSec = l.computeRequestsPerSec(calls)

	// Compute AvgPayloadSize
	baseline.AvgPayloadSize = l.computeAvgPayloadSize(calls)

	// Compute AvgErrorRate
	baseline.AvgErrorRate = l.computeAvgErrorRate(calls)

	// Extract known endpoints
	baseline.KnownEndpoints = l.extractEndpoints(calls)

	// Extract normal time windows (simplified)
	baseline.NormalTimeWindows = l.extractTimeWindows(calls)

	return baseline
}

// computeRequestsPerSec calculates average requests per second
func (l *Learner) computeRequestsPerSec(calls []models.APICall) float64 {
	if len(calls) == 0 {
		return 0
	}

	// Find time span
	minTime := calls[0].Timestamp
	maxTime := calls[0].Timestamp

	for _, call := range calls {
		if call.Timestamp.Before(minTime) {
			minTime = call.Timestamp
		}
		if call.Timestamp.After(maxTime) {
			maxTime = call.Timestamp
		}
	}

	duration := maxTime.Sub(minTime).Seconds()
	if duration == 0 {
		duration = 1
	}

	return float64(len(calls)) / duration
}

// computeAvgPayloadSize calculates average payload size
func (l *Learner) computeAvgPayloadSize(calls []models.APICall) float64 {
	if len(calls) == 0 {
		return 0
	}

	totalSize := 0
	for _, call := range calls {
		totalSize += call.PayloadSize
	}

	return float64(totalSize) / float64(len(calls))
}

// computeAvgErrorRate calculates average error rate (4xx + 5xx)
func (l *Learner) computeAvgErrorRate(calls []models.APICall) float64 {
	if len(calls) == 0 {
		return 0
	}

	errorCount := 0
	for _, call := range calls {
		if call.StatusCode >= 400 {
			errorCount++
		}
	}

	return float64(errorCount) / float64(len(calls))
}

// extractEndpoints extracts unique endpoints from calls
func (l *Learner) extractEndpoints(calls []models.APICall) []string {
	endpoints := make(map[string]bool)
	for _, call := range calls {
		endpoints[call.Endpoint] = true
	}

	result := make([]string, 0, len(endpoints))
	for ep := range endpoints {
		result = append(result, ep)
	}

	return result
}

// extractTimeWindows extracts normal time windows (simplified)
func (l *Learner) extractTimeWindows(calls []models.APICall) []models.TimeWindow {
	// Placeholder: simplified extraction
	// In a real implementation, would analyze hour distribution
	return []models.TimeWindow{
		{Start: "00:00", End: "23:59", Days: "*"},
	}
}

// ValidateBaseline checks if a baseline is valid
func (l *Learner) ValidateBaseline(baseline *models.TenantBaseline) error {
	if baseline.TenantID == "" {
		return fmt.Errorf("baseline missing tenant ID")
	}

	if baseline.RequestsPerSec < 0 {
		return fmt.Errorf("baseline has invalid requests per sec: %v", baseline.RequestsPerSec)
	}

	return nil
}
