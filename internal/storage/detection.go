package storage

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bogdanticu88/centipede/internal/models"
)

// DetectionStore handles persistence of detection results
type DetectionStore struct{}

// SaveDetections writes detection results to a JSON file
func (ds *DetectionStore) SaveDetections(path string, result *models.DetectionResult) error {
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal detections: %w", err)
	}

	if err := os.WriteFile(path, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write detections to file: %w", err)
	}

	return nil
}

// LoadDetections reads detection results from a JSON file
func (ds *DetectionStore) LoadDetections(path string) (*models.DetectionResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read detection file: %w", err)
	}

	var result models.DetectionResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal detections: %w", err)
	}

	return &result, nil
}
