package storage

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bogdanticu88/centipede/internal/models"
)

// BaselineStore handles persistence of baselines
type BaselineStore struct{}

// BaselineFile represents the structure for persisting baselines
type BaselineFile struct {
	Tenants map[string]*models.TenantBaseline `json:"tenants"`
}

// SaveBaselines writes baselines to a JSON file
func (bs *BaselineStore) SaveBaselines(path string, baselines map[string]*models.TenantBaseline) error {
	data := BaselineFile{
		Tenants: baselines,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal baselines: %w", err)
	}

	if err := os.WriteFile(path, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write baselines to file: %w", err)
	}

	return nil
}

// LoadBaselines reads baselines from a JSON file
func (bs *BaselineStore) LoadBaselines(path string) (map[string]*models.TenantBaseline, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read baseline file: %w", err)
	}

	var bf BaselineFile
	if err := json.Unmarshal(data, &bf); err != nil {
		return nil, fmt.Errorf("failed to unmarshal baselines: %w", err)
	}

	return bf.Tenants, nil
}
