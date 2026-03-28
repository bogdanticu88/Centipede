package main

import (
	"fmt"
	"os"

	"github.com/bogdanticu88/centipede/internal/cmd"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "centipede",
	Short: "Multi-cloud API anomaly detection and tenant protection",
	Long: `CENTIPEDE detects compromised API clients in multi-tenant platforms.

It learns baselines from historical logs, detects anomalies in real-time,
and enforces graduated responses: rate limiting for warnings, blocking for critical anomalies.`,
	Version: "0.1.0",
}

func init() {
	rootCmd.AddCommand(cmd.NewInitCmd())
	rootCmd.AddCommand(cmd.NewDetectCmd())
	rootCmd.AddCommand(cmd.NewMonitorCmd())
	rootCmd.AddCommand(cmd.NewKillCmd())
	rootCmd.AddCommand(cmd.NewReportCmd())
	rootCmd.AddCommand(cmd.NewStatusCmd())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		// Check if it's a CommandError with custom exit code
		if cmdErr, ok := err.(*cmd.CommandError); ok {
			fmt.Fprintln(os.Stderr, cmdErr.Error())
			os.Exit(cmdErr.Code)
		}

		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
