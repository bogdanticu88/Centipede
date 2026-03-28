package detection

import (
	"testing"
	"time"

	"github.com/bogdanticu88/centipede/internal/config"
	"github.com/bogdanticu88/centipede/internal/models"
)

func TestScorerVolumeSpike(t *testing.T) {
	cfg := config.DefaultConfig()
	scorer := NewScorer(cfg)

	baseline := &models.TenantBaseline{
		TenantID:       "tenant1",
		RequestsPerSec: 10,
	}

	// Create calls that trigger volume spike (20 calls when baseline is 10 RPS)
	calls := make([]models.APICall, 25)
	now := time.Now()
	for i := 0; i < 25; i++ {
		calls[i] = models.APICall{
			Timestamp:   now,
			TenantID:    "tenant1",
			Endpoint:    "/api/test",
			Method:      "GET",
			StatusCode:  200,
			PayloadSize: 100,
		}
	}

	anomaly := scorer.Score(calls, baseline)

	if anomaly.Score == 0 {
		t.Errorf("expected score > 0 for volume spike, got %d", anomaly.Score)
	}

	if len(anomaly.Triggers) == 0 {
		t.Errorf("expected triggers for volume spike, got none")
	}
}

func TestScorerHoneypot(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Honeypots = []config.HoneypotConfig{
		{Path: "/admin/debug", Severity: 3},
	}
	scorer := NewScorer(cfg)

	baseline := &models.TenantBaseline{
		TenantID:       "tenant1",
		RequestsPerSec: 10,
		KnownEndpoints: []string{"/admin/debug"}, // honeypot is "known" endpoint
	}

	// Create a call that hits honeypot
	calls := []models.APICall{
		{
			Timestamp:   time.Now(),
			TenantID:    "tenant1",
			Endpoint:    "/admin/debug",
			Method:      "GET",
			StatusCode:  404,
			PayloadSize: 100,
		},
	}

	anomaly := scorer.Score(calls, baseline)

	if anomaly.Score != 3 {
		t.Errorf("expected score 3 for honeypot hit, got %d", anomaly.Score)
	}
}

func TestScorerErrorRate(t *testing.T) {
	cfg := config.DefaultConfig()
	scorer := NewScorer(cfg)

	baseline := &models.TenantBaseline{
		TenantID:       "tenant1",
		RequestsPerSec: 100,
		AvgErrorRate:   0.01, // 1% baseline
	}

	// Create calls with high error rate (>10% = 11x baseline)
	calls := make([]models.APICall, 20)
	now := time.Now()
	for i := 0; i < 20; i++ {
		status := 200
		if i >= 3 { // 17 out of 20 errors = 85%
			status = 500
		}

		calls[i] = models.APICall{
			Timestamp:   now,
			TenantID:    "tenant1",
			Endpoint:    "/api/test",
			Method:      "GET",
			StatusCode:  status,
			PayloadSize: 100,
		}
	}

	anomaly := scorer.Score(calls, baseline)

	if anomaly.Score == 0 {
		t.Errorf("expected score > 0 for error rate spike, got %d", anomaly.Score)
	}
}

func TestScorerNewEndpoint(t *testing.T) {
	cfg := config.DefaultConfig()
	scorer := NewScorer(cfg)

	baseline := &models.TenantBaseline{
		TenantID:       "tenant1",
		RequestsPerSec: 10,
		KnownEndpoints: []string{"/api/known"},
	}

	// Create calls with unknown endpoint
	calls := []models.APICall{
		{
			Timestamp:   time.Now(),
			TenantID:    "tenant1",
			Endpoint:    "/api/unknown",
			Method:      "GET",
			StatusCode:  200,
			PayloadSize: 100,
		},
	}

	anomaly := scorer.Score(calls, baseline)

	if anomaly.Score == 0 {
		t.Errorf("expected score > 0 for new endpoint, got %d", anomaly.Score)
	}
}
