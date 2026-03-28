package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// StatusCmd is the status command
type StatusCmd struct {
	ConfigPath   string
	BaselinePath string
	LogSource    string
}

// NewStatusCmd creates a new status command
func NewStatusCmd() *cobra.Command {
	cmd := &StatusCmd{}

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Check current tenant health",
		Long: `Display current status of all tenants based on recent logs.

Shows health status, current scores, and any active anomalies.

Example:
  centipede status \
    --config config.yaml \
    --baseline baseline.json \
    --log-source ./logs`,
		RunE: func(cmdCobraCmd *cobra.Command, args []string) error {
			return cmd.Run(context.Background())
		},
	}

	statusCmd.Flags().StringVar(&cmd.ConfigPath, "config", "config.yaml", "Path to configuration file")
	statusCmd.Flags().StringVar(&cmd.BaselinePath, "baseline", "baseline.json", "Path to baseline file")
	statusCmd.Flags().StringVar(&cmd.LogSource, "log-source", "", "Path to log files (required)")

	_ = statusCmd.MarkFlagRequired("log-source")

	return statusCmd
}

// Run executes the status command
func (cmd *StatusCmd) Run(ctx context.Context) error {
	fmt.Println("status: tenant health check not yet implemented")

	// TODO: Implement status checking
	// - Load baseline
	// - Load recent logs
	// - Score each tenant
	// - Display health status

	return nil
}
