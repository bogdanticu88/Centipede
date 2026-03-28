package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/bogdanticu88/centipede/internal/cloud"
	"github.com/bogdanticu88/centipede/internal/config"
	"github.com/spf13/cobra"
)

// KillCmd is the kill command
type KillCmd struct {
	ConfigPath string
	TenantID   string
	Reason     string
}

// NewKillCmd creates a new kill command
func NewKillCmd() *cobra.Command {
	cmd := &KillCmd{}

	killCmd := &cobra.Command{
		Use:   "kill",
		Short: "Instant tenant block (emergency)",
		Long: `Instantly block a tenant and execute the kill-switch playbook:
- Revoke API access (IAM)
- Inject policy to block requests
- Create incident ticket
- Log to audit trail

Use with caution! This is irreversible without manual intervention.

Example:
  centipede kill \
    --config config.yaml \
    --tenant salesforce-ro \
    --reason "Suspected compromise"`,
		RunE: func(cmdCobraCmd *cobra.Command, args []string) error {
			return cmd.Run(context.Background())
		},
	}

	killCmd.Flags().StringVar(&cmd.ConfigPath, "config", "config.yaml", "Path to configuration file")
	killCmd.Flags().StringVar(&cmd.TenantID, "tenant", "", "Tenant ID to block (required)")
	killCmd.Flags().StringVar(&cmd.Reason, "reason", "Manual intervention", "Reason for blocking")

	_ = killCmd.MarkFlagRequired("tenant")

	return killCmd
}

// Run executes the kill command
func (cmd *KillCmd) Run(ctx context.Context) error {
	// Load configuration
	cfg, err := config.LoadConfig(cmd.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Printf("Cloud Provider: %s\n", cfg.Cloud)
	fmt.Printf("\n⚠️  KILL-SWITCH ACTIVATION\n")
	fmt.Printf("Tenant: %s\n", cmd.TenantID)
	fmt.Printf("Reason: %s\n\n", cmd.Reason)

	// Create orchestrator factory and get orchestrator for cloud provider
	factory := &cloud.ProviderFactory{}
	orchestrator, err := factory.CreateOrchestrator(cfg)
	if err != nil {
		return fmt.Errorf("failed to create orchestrator: %w", err)
	}

	fmt.Printf("Using orchestrator: %s\n", orchestrator.Name())

	// Block the tenant
	fmt.Printf("\nBlocking tenant: %s\n", cmd.TenantID)
	if err := orchestrator.BlockTenant(ctx, cmd.TenantID, cmd.Reason); err != nil {
		return fmt.Errorf("failed to block tenant: %w", err)
	}

	fmt.Println("✓ Tenant blocked in API Gateway")

	// TODO: Create incident ticket
	// - Could integrate with Jira, GitHub Issues, Azure DevOps, etc.
	// fmt.Println("Creating incident ticket...")

	// TODO: Log to audit trail
	// - Record block action, timestamp, reason, operator, etc.
	// fmt.Println("Logging action to audit trail...")

	fmt.Println("\n✓ Kill-switch activated successfully")
	fmt.Printf("  Blocked Tenant: %s\n", cmd.TenantID)
	fmt.Printf("  Reason: %s\n", cmd.Reason)
	fmt.Printf("  Timestamp: %s\n", time.Now().Format(time.RFC3339))

	return nil
}
