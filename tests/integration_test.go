package tests

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bogdanticu88/centipede/internal/baseline"
	"github.com/bogdanticu88/centipede/internal/config"
	"github.com/bogdanticu88/centipede/internal/detection"
	"github.com/bogdanticu88/centipede/internal/models"
	"github.com/bogdanticu88/centipede/internal/parsers"
	"github.com/bogdanticu88/centipede/internal/storage"
)

func TestFullDetectionPipeline(t *testing.T) {
	// Create sample logs
	now := time.Now()
	normalCalls := []models.APICall{
		{
			Timestamp:    now.Add(-1 * time.Minute),
			TenantID:     "tenant1",
			Endpoint:     "/api/users",
			Method:       "GET",
			StatusCode:   200,
			PayloadSize:  1024,
			ResponseTime: 50 * time.Millisecond,
		},
		{
			Timestamp:    now.Add(-30 * time.Second),
			TenantID:     "tenant1",
			Endpoint:     "/api/orders",
			Method:       "GET",
			StatusCode:   200,
			PayloadSize:  1024,
			ResponseTime: 45 * time.Millisecond,
		},
	}

	// Step 1: Learn baseline
	learner := baseline.NewLearner()
	baselines := learner.LearnBaseline(normalCalls)

	if len(baselines) == 0 {
		t.Fatalf("expected to learn baselines")
	}

	baseline := baselines["tenant1"]
	if baseline == nil {
		t.Fatalf("expected baseline for tenant1")
	}

	// Step 2: Create anomalous calls
	anomalousCalls := []models.APICall{
		{
			Timestamp:    now,
			TenantID:     "tenant1",
			Endpoint:     "/admin/debug", // honeypot
			Method:       "GET",
			StatusCode:   404,
			PayloadSize:  512,
			ResponseTime: 30 * time.Millisecond,
		},
		{
			Timestamp:    now.Add(1 * time.Second),
			TenantID:     "tenant1",
			Endpoint:     "/api/secret",
			Method:       "GET",
			StatusCode:   200,
			PayloadSize:  2048, // Large payload
			ResponseTime: 100 * time.Millisecond,
		},
		{
			Timestamp:    now.Add(2 * time.Second),
			TenantID:     "tenant1",
			Endpoint:     "/api/users",
			Method:       "GET",
			StatusCode:   500,
			PayloadSize:  512,
			ResponseTime: 200 * time.Millisecond,
		},
	}

	// Step 3: Run detection
	cfg := config.DefaultConfig()
	cfg.Honeypots = []config.HoneypotConfig{
		{Path: "/admin/debug", Severity: 3},
	}

	detector := detection.NewDetector(cfg)
	result := detector.Detect(anomalousCalls, baselines)

	// Verify detection results
	if len(result.Anomalies) == 0 {
		t.Fatalf("expected to detect anomalies")
	}

	anomaly := result.Anomalies[0]
	if anomaly.Score < 2 {
		t.Errorf("expected score >= 2, got %d", anomaly.Score)
	}

	if anomaly.RecommendedAction != "block" && anomaly.RecommendedAction != "rate_limit" {
		t.Errorf("unexpected recommended action: %s", anomaly.RecommendedAction)
	}

	// Step 4: Test persistence
	tmpDir := t.TempDir()
	baselinePath := filepath.Join(tmpDir, "baseline.json")
	detectionPath := filepath.Join(tmpDir, "detections.json")

	// Save baseline
	baselineStore := &storage.BaselineStore{}
	if err := baselineStore.SaveBaselines(baselinePath, baselines); err != nil {
		t.Fatalf("failed to save baselines: %v", err)
	}

	// Load baseline
	loadedBaselines, err := baselineStore.LoadBaselines(baselinePath)
	if err != nil {
		t.Fatalf("failed to load baselines: %v", err)
	}

	if len(loadedBaselines) != len(baselines) {
		t.Errorf("expected %d baselines, got %d", len(baselines), len(loadedBaselines))
	}

	// Save detections
	detectionStore := &storage.DetectionStore{}
	if err := detectionStore.SaveDetections(detectionPath, result); err != nil {
		t.Fatalf("failed to save detections: %v", err)
	}

	// Load detections
	loadedDetections, err := detectionStore.LoadDetections(detectionPath)
	if err != nil {
		t.Fatalf("failed to load detections: %v", err)
	}

	if len(loadedDetections.Anomalies) != len(result.Anomalies) {
		t.Errorf("expected %d anomalies, got %d", len(result.Anomalies), len(loadedDetections.Anomalies))
	}
}

func TestParserIntegration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a log file
	logFile := filepath.Join(tmpDir, "test.json")
	logData := []byte(`[
  {
    "timestamp": "2026-03-28T10:00:00Z",
    "tenant_id": "test-tenant",
    "endpoint": "/api/test",
    "method": "GET",
    "status_code": 200,
    "payload_size": 1024,
    "response_ms": 50,
    "user_agent": "test",
    "source_ip": "127.0.0.1"
  }
]`)

	if err := os.WriteFile(logFile, logData, 0644); err != nil {
		t.Fatalf("failed to write test log file: %v", err)
	}

	// Test parsing
	parser := &parsers.GenericJSONParser{}
	loader := parsers.NewLoader(parser)

	calls, err := loader.LoadFromFile(logFile)
	if err != nil {
		t.Fatalf("failed to load logs: %v", err)
	}

	if len(calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(calls))
	}

	call := calls[0]
	if call.TenantID != "test-tenant" {
		t.Errorf("expected test-tenant, got %s", call.TenantID)
	}
}

func TestMultiTenantDetection(t *testing.T) {
	cfg := config.DefaultConfig()
	detector := detection.NewDetector(cfg)

	now := time.Now()

	// Create calls for two tenants
	calls := []models.APICall{
		{
			Timestamp:   now,
			TenantID:    "tenant1",
			Endpoint:    "/api/a",
			Method:      "GET",
			StatusCode:  200,
			PayloadSize: 100,
		},
		{
			Timestamp:   now,
			TenantID:    "tenant2",
			Endpoint:    "/api/b",
			Method:      "GET",
			StatusCode:  200,
			PayloadSize: 100,
		},
	}

	// Create baselines
	learner := baseline.NewLearner()
	baselines := learner.LearnBaseline(calls)

	// Run detection on normal calls
	result := detector.Detect(calls, baselines)

	// With normal calls, we should have minimal/no anomalies
	if result.Summary["total"] > 0 {
		t.Errorf("expected no anomalies for normal calls")
	}

	// Now create anomalous calls for one tenant
	anomalous := []models.APICall{
		{
			Timestamp:   now,
			TenantID:    "tenant1",
			Endpoint:    "/api/c", // New endpoint
			Method:      "GET",
			StatusCode:  200,
			PayloadSize: 100,
		},
	}

	result = detector.Detect(anomalous, baselines)

	// Should detect anomaly for tenant1
	anomalies := result.Anomalies
	if len(anomalies) == 0 {
		t.Errorf("expected to detect anomaly for tenant1")
	}

	if anomalies[0].TenantID != "tenant1" {
		t.Errorf("expected anomaly for tenant1, got %s", anomalies[0].TenantID)
	}
}
