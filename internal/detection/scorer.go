package detection

import (
	"github.com/bogdanticu88/centipede/internal/config"
	"github.com/bogdanticu88/centipede/internal/models"
)

// Scorer performs cumulative anomaly scoring
type Scorer struct {
	config    *config.Config
	honeypots map[string]int // path -> severity
}

// NewScorer creates a new scorer
func NewScorer(cfg *config.Config) *Scorer {
	honeypots := make(map[string]int)
	for _, hp := range cfg.Honeypots {
		honeypots[hp.Path] = hp.Severity
	}

	return &Scorer{
		config:    cfg,
		honeypots: honeypots,
	}
}

// Score performs cumulative scoring on a set of API calls for a tenant
func (s *Scorer) Score(calls []models.APICall, baseline *models.TenantBaseline) *models.Anomaly {
	anomaly := &models.Anomaly{
		Timestamp: calls[0].Timestamp,
		Score:     0,
		Triggers:  []string{},
		Details:   make(map[string]interface{}),
	}

	if len(calls) == 0 {
		return anomaly
	}

	tenantID := calls[0].TenantID
	anomaly.TenantID = tenantID

	// Rule 1: Volume Spike
	if s.checkVolumeSpikeScore(calls, baseline) {
		anomaly.Score += 1
		anomaly.Triggers = append(anomaly.Triggers, "volume_spike")
	}

	// Rule 2: Endpoint Anomaly
	if newEndpoints := s.checkEndpointAnomalyScore(calls, baseline); len(newEndpoints) > 0 {
		anomaly.Score += 1
		anomaly.Triggers = append(anomaly.Triggers, "endpoint_anomaly")
		anomaly.Details["new_endpoints"] = newEndpoints
	}

	// Rule 3: Payload Size Surge
	if s.checkPayloadSizeScore(calls, baseline) {
		anomaly.Score += 1
		anomaly.Triggers = append(anomaly.Triggers, "payload_surge")
	}

	// Rule 4: Time-of-Day Deviation
	if s.checkTimeDeviationScore(calls, baseline) {
		anomaly.Score += 1
		anomaly.Triggers = append(anomaly.Triggers, "time_deviation")
	}

	// Rule 5: Error Rate Jump
	if s.checkErrorRateScore(calls, baseline) {
		anomaly.Score += 1
		anomaly.Triggers = append(anomaly.Triggers, "error_rate_jump")
	}

	// Rule 6: Honeypot Hit
	if honeypotSeverity := s.checkHoneypotScore(calls); honeypotSeverity > 0 {
		anomaly.Score += honeypotSeverity
		anomaly.Triggers = append(anomaly.Triggers, "honeypot_hit")
		anomaly.Details["honeypot_severity"] = honeypotSeverity
	}

	// Determine recommended action
	if anomaly.Score >= s.config.Detection.ScoreCritical {
		anomaly.RecommendedAction = "block"
	} else if anomaly.Score >= s.config.Detection.ScoreWarning {
		anomaly.RecommendedAction = "rate_limit"
	} else {
		anomaly.RecommendedAction = "monitor"
	}

	return anomaly
}

// checkVolumeSpikeScore returns true if request volume > threshold * baseline
func (s *Scorer) checkVolumeSpikeScore(calls []models.APICall, baseline *models.TenantBaseline) bool {
	if baseline == nil || baseline.RequestsPerSec == 0 {
		return false
	}

	threshold := baseline.RequestsPerSec * s.config.Detection.VolumeThreshold
	return float64(len(calls)) > threshold
}

// checkEndpointAnomalyScore returns list of unknown endpoints
func (s *Scorer) checkEndpointAnomalyScore(calls []models.APICall, baseline *models.TenantBaseline) []string {
	if baseline == nil {
		return nil
	}

	knownSet := make(map[string]bool)
	for _, ep := range baseline.KnownEndpoints {
		knownSet[ep] = true
	}

	var newEndpoints []string
	seenNew := make(map[string]bool)

	for _, call := range calls {
		if !knownSet[call.Endpoint] && !seenNew[call.Endpoint] {
			newEndpoints = append(newEndpoints, call.Endpoint)
			seenNew[call.Endpoint] = true
		}
	}

	return newEndpoints
}

// checkPayloadSizeScore returns true if avg payload > threshold * baseline
func (s *Scorer) checkPayloadSizeScore(calls []models.APICall, baseline *models.TenantBaseline) bool {
	if baseline == nil || baseline.AvgPayloadSize == 0 {
		return false
	}

	totalSize := 0
	for _, call := range calls {
		totalSize += call.PayloadSize
	}

	avgSize := float64(totalSize) / float64(len(calls))
	threshold := baseline.AvgPayloadSize * s.config.Detection.PayloadThreshold

	return avgSize > threshold
}

// checkTimeDeviationScore returns true if requests outside normal hours
func (s *Scorer) checkTimeDeviationScore(calls []models.APICall, baseline *models.TenantBaseline) bool {
	// Placeholder: would compare request times against baseline.NormalTimeWindows
	// For now, simplified check
	if baseline == nil || len(baseline.NormalTimeWindows) == 0 {
		return false
	}

	// TODO: Implement full time window logic
	return false
}

// checkErrorRateScore returns true if error rate > threshold * baseline
func (s *Scorer) checkErrorRateScore(calls []models.APICall, baseline *models.TenantBaseline) bool {
	if baseline == nil || baseline.AvgErrorRate == 0 {
		return false
	}

	errorCount := 0
	for _, call := range calls {
		if call.StatusCode >= 400 {
			errorCount++
		}
	}

	currentErrorRate := float64(errorCount) / float64(len(calls))
	threshold := baseline.AvgErrorRate * s.config.Detection.ErrorRateThreshold

	return currentErrorRate > threshold
}

// checkHoneypotScore returns severity if honeypot hit, 0 otherwise
func (s *Scorer) checkHoneypotScore(calls []models.APICall) int {
	for _, call := range calls {
		if severity, ok := s.honeypots[call.Endpoint]; ok {
			return severity
		}
	}
	return 0
}
