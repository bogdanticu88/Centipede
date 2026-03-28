package detection

import (
	"time"

	"github.com/bogdanticu88/centipede/internal/config"
	"github.com/bogdanticu88/centipede/internal/models"
)

// Detector orchestrates anomaly detection
type Detector struct {
	config *config.Config
	scorer *Scorer
}

// NewDetector creates a new detector
func NewDetector(cfg *config.Config) *Detector {
	return &Detector{
		config: cfg,
		scorer: NewScorer(cfg),
	}
}

// Detect analyzes API calls against baselines and returns anomalies
func (d *Detector) Detect(calls []models.APICall, baselines map[string]*models.TenantBaseline) *models.DetectionResult {
	result := &models.DetectionResult{
		Timestamp:   time.Now(),
		Anomalies:   []models.Anomaly{},
		Summary:     make(map[string]int),
		GeneratedAt: time.Now(),
	}

	// Group calls by tenant
	callsByTenant := make(map[string][]models.APICall)
	for _, call := range calls {
		callsByTenant[call.TenantID] = append(callsByTenant[call.TenantID], call)
	}

	totalAnomalies := 0
	criticalAnomalies := 0
	warningAnomalies := 0

	// Score each tenant
	for tenantID, tenantCalls := range callsByTenant {
		baseline := baselines[tenantID]

		// Score this tenant's calls
		anomaly := d.scorer.Score(tenantCalls, baseline)
		anomaly.TenantID = tenantID

		if anomaly.Score > 0 {
			result.Anomalies = append(result.Anomalies, *anomaly)
			totalAnomalies++

			if anomaly.Score >= d.config.Detection.ScoreCritical {
				criticalAnomalies++
			} else if anomaly.Score >= d.config.Detection.ScoreWarning {
				warningAnomalies++
			}
		}
	}

	// Populate summary
	result.Summary["total"] = totalAnomalies
	result.Summary["critical"] = criticalAnomalies
	result.Summary["warning"] = warningAnomalies
	result.Summary["normal"] = len(callsByTenant) - totalAnomalies

	return result
}

// DetectWindow analyzes calls within a time window
func (d *Detector) DetectWindow(calls []models.APICall, baselines map[string]*models.TenantBaseline, timeWindow time.Duration) *models.DetectionResult {
	if len(calls) == 0 {
		return &models.DetectionResult{
			Timestamp:   time.Now(),
			Anomalies:   []models.Anomaly{},
			Summary:     make(map[string]int),
			GeneratedAt: time.Now(),
		}
	}

	// Filter calls to the time window
	now := time.Now()
	cutoff := now.Add(-timeWindow)

	var filtered []models.APICall
	for _, call := range calls {
		if call.Timestamp.After(cutoff) {
			filtered = append(filtered, call)
		}
	}

	return d.Detect(filtered, baselines)
}
