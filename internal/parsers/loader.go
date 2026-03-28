package parsers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bogdanticu88/centipede/internal/models"
)

// Loader handles loading logs from various sources
type Loader struct {
	parser Parser
}

// Parser interface for different log formats
type Parser interface {
	Parse(data []byte) ([]models.APICall, error)
}

// NewLoader creates a loader with the specified parser
func NewLoader(parser Parser) *Loader {
	return &Loader{parser: parser}
}

// LoadFromFile loads logs from a local file
func (l *Loader) LoadFromFile(path string) ([]models.APICall, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read log file: %w", err)
	}

	return l.parser.Parse(data)
}

// LoadFromDirectory loads logs from all JSON files in a directory
func (l *Loader) LoadFromDirectory(dirPath string) ([]models.APICall, error) {
	var allCalls []models.APICall

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(dirPath, entry.Name())
		calls, err := l.LoadFromFile(filePath)
		if err != nil {
			// Log but continue processing other files
			fmt.Fprintf(os.Stderr, "warning: failed to parse %s: %v\n", filePath, err)
			continue
		}

		allCalls = append(allCalls, calls...)
	}

	return allCalls, nil
}

// LoadFromSource loads logs from a source (file or directory)
func (l *Loader) LoadFromSource(source string) ([]models.APICall, error) {
	stat, err := os.Stat(source)
	if err != nil {
		return nil, fmt.Errorf("failed to stat source: %w", err)
	}

	if stat.IsDir() {
		return l.LoadFromDirectory(source)
	}

	return l.LoadFromFile(source)
}
