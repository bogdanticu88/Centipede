package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/bogdanticu88/centipede/internal/baseline"
	"github.com/bogdanticu88/centipede/internal/config"
	"github.com/bogdanticu88/centipede/internal/parsers"
	"github.com/bogdanticu88/centipede/internal/storage"
	"github.com/spf13/cobra"
)

// InitCmd is the init command
type InitCmd struct {
	ConfigPath string
	LogSource  string
	Window     string
	OutputPath string
}

// NewInitCmd creates a new init command
func NewInitCmd() *cobra.Command {
	cmd := &InitCmd{}

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Learn baselines from historical logs",
		Long: `Initialize CENTIPEDE by learning baselines from historical API logs.

This creates a baseline.json file that captures normal behavior for each tenant.

Example:
  centipede init \
    --config config.yaml \
    --log-source ./logs \
    --window 7d \
    --output baseline.json`,
		RunE: func(cmdCobraCmd *cobra.Command, args []string) error {
			return cmd.Run(context.Background())
		},
	}

	initCmd.Flags().StringVar(&cmd.ConfigPath, "config", "config.yaml", "Path to configuration file")
	initCmd.Flags().StringVar(&cmd.LogSource, "log-source", "", "Path to log files (required)")
	initCmd.Flags().StringVar(&cmd.Window, "window", "7d", "Lookback window (e.g., 7d, 24h)")
	initCmd.Flags().StringVar(&cmd.OutputPath, "output", "baseline.json", "Output baseline file path")

	_ = initCmd.MarkFlagRequired("log-source")

	return initCmd
}

// Run executes the init command
func (cmd *InitCmd) Run(ctx context.Context) error {
	// Load configuration
	_, err := config.LoadConfig(cmd.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create parser (using generic for now, can be cloud-specific later)
	parser := &parsers.GenericJSONParser{}
	loader := parsers.NewLoader(parser)

	// Load logs from source
	fmt.Printf("Loading logs from %s...\n", cmd.LogSource)
	calls, err := loader.LoadFromSource(cmd.LogSource)
	if err != nil {
		return fmt.Errorf("failed to load logs: %w", err)
	}

	if len(calls) == 0 {
		return fmt.Errorf("no logs found in %s", cmd.LogSource)
	}

	fmt.Printf("Loaded %d API calls\n", len(calls))

	// Filter by time window
	window, err := parseDuration(cmd.Window)
	if err != nil {
		return fmt.Errorf("invalid window duration: %w", err)
	}

	fmt.Printf("Using data from last %v\n", window)

	// Learn baselines
	fmt.Println("Learning baselines...")
	learner := baseline.NewLearner()
	baselines := learner.LearnBaseline(calls)

	// Validate baselines
	for tenantID, b := range baselines {
		if err := learner.ValidateBaseline(b); err != nil {
			fmt.Printf("Warning: invalid baseline for %s: %v\n", tenantID, err)
		}
	}

	fmt.Printf("Learned baselines for %d tenants\n", len(baselines))

	// Save baselines
	fmt.Printf("Saving baselines to %s\n", cmd.OutputPath)
	store := &storage.BaselineStore{}
	if err := store.SaveBaselines(cmd.OutputPath, baselines); err != nil {
		return fmt.Errorf("failed to save baselines: %w", err)
	}

	// Print summary
	fmt.Println("\nBaseline Summary:")
	for tenantID, b := range baselines {
		fmt.Printf("  %s:\n", tenantID)
		fmt.Printf("    Requests/sec: %.2f\n", b.RequestsPerSec)
		fmt.Printf("    Avg Payload: %d bytes\n", int(b.AvgPayloadSize))
		fmt.Printf("    Error Rate: %.2f%%\n", b.AvgErrorRate*100)
		fmt.Printf("    Known Endpoints: %d\n", len(b.KnownEndpoints))
	}

	fmt.Println("\n✓ Baseline learning complete")
	return nil
}

// parseDuration parses duration strings like "7d", "24h", "30m"
func parseDuration(s string) (time.Duration, error) {
	// Try to parse as time.Duration first (handles "1h", "30m", etc.)
	d, err := time.ParseDuration(s)
	if err == nil {
		return d, nil
	}

	// Handle custom formats like "7d", "2w"
	switch {
	case len(s) > 1:
		var num int
		var unit string
		fmt.Sscanf(s, "%d%s", &num, &unit)

		switch unit {
		case "d":
			return time.Duration(num) * 24 * time.Hour, nil
		case "w":
			return time.Duration(num) * 7 * 24 * time.Hour, nil
		case "h":
			return time.Duration(num) * time.Hour, nil
		case "m":
			return time.Duration(num) * time.Minute, nil
		}
	}

	return 0, fmt.Errorf("invalid duration format: %s", s)
}
