package baseline

import (
	"testing"
	"time"

	"github.com/bogdanticu88/centipede/internal/models"
)

func TestLearnerComputeBaseline(t *testing.T) {
	learner := NewLearner()

	// Create sample API calls
	now := time.Now()
	calls := []models.APICall{
		{
			Timestamp:   now,
			TenantID:    "tenant1",
			Endpoint:    "/api/users",
			Method:      "GET",
			StatusCode:  200,
			PayloadSize: 1024,
		},
		{
			Timestamp:   now.Add(1 * time.Second),
			TenantID:    "tenant1",
			Endpoint:    "/api/orders",
			Method:      "POST",
			StatusCode:  201,
			PayloadSize: 2048,
		},
		{
			Timestamp:   now.Add(2 * time.Second),
			TenantID:    "tenant1",
			Endpoint:    "/api/users",
			Method:      "GET",
			StatusCode:  200,
			PayloadSize: 1024,
		},
	}

	baselines := learner.LearnBaseline(calls)

	if len(baselines) != 1 {
		t.Errorf("expected 1 baseline, got %d", len(baselines))
	}

	baseline := baselines["tenant1"]
	if baseline == nil {
		t.Errorf("expected baseline for tenant1")
		return
	}

	if baseline.TenantID != "tenant1" {
		t.Errorf("expected tenant1, got %s", baseline.TenantID)
	}

	if baseline.AvgPayloadSize == 0 {
		t.Errorf("expected non-zero avg payload size")
	}

	if len(baseline.KnownEndpoints) == 0 {
		t.Errorf("expected known endpoints")
	}
}

func TestLearnerMultipleTenants(t *testing.T) {
	learner := NewLearner()

	now := time.Now()
	calls := []models.APICall{
		{
			Timestamp:   now,
			TenantID:    "tenant1",
			Endpoint:    "/api/users",
			Method:      "GET",
			StatusCode:  200,
			PayloadSize: 1024,
		},
		{
			Timestamp:   now,
			TenantID:    "tenant2",
			Endpoint:    "/api/orders",
			Method:      "GET",
			StatusCode:  200,
			PayloadSize: 2048,
		},
	}

	baselines := learner.LearnBaseline(calls)

	if len(baselines) != 2 {
		t.Errorf("expected 2 baselines, got %d", len(baselines))
	}

	if _, ok := baselines["tenant1"]; !ok {
		t.Errorf("expected baseline for tenant1")
	}

	if _, ok := baselines["tenant2"]; !ok {
		t.Errorf("expected baseline for tenant2")
	}
}

func TestLearnerErrorRate(t *testing.T) {
	learner := NewLearner()

	now := time.Now()
	calls := []models.APICall{
		{
			Timestamp:   now,
			TenantID:    "tenant1",
			Endpoint:    "/api/test",
			Method:      "GET",
			StatusCode:  200,
			PayloadSize: 100,
		},
		{
			Timestamp:   now,
			TenantID:    "tenant1",
			Endpoint:    "/api/test",
			Method:      "GET",
			StatusCode:  500,
			PayloadSize: 100,
		},
		{
			Timestamp:   now,
			TenantID:    "tenant1",
			Endpoint:    "/api/test",
			Method:      "GET",
			StatusCode:  500,
			PayloadSize: 100,
		},
	}

	baselines := learner.LearnBaseline(calls)
	baseline := baselines["tenant1"]

	// 2 errors out of 3 calls = 66.7% error rate
	if baseline.AvgErrorRate < 0.6 || baseline.AvgErrorRate > 0.7 {
		t.Errorf("expected error rate around 0.667, got %v", baseline.AvgErrorRate)
	}
}
