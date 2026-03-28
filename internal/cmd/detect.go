package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/bogdanticu88/centipede/internal/config"
	"github.com/bogdanticu88/centipede/internal/detection"
	"github.com/bogdanticu88/centipede/internal/exitcode"
	"github.com/bogdanticu88/centipede/internal/log"
	"github.com/bogdanticu88/centipede/internal/parsers"
	"github.com/bogdanticu88/centipede/internal/storage"
	"github.com/spf13/cobra"
)

// DetectCmd is the detect command
type DetectCmd struct {
	ConfigPath   string
	BaselinePath string
	LogSource    string
	OutputPath   string
}

// NewDetectCmd creates a new detect command
func NewDetectCmd() *cobra.Command {
	cmd := &DetectCmd{}

	detectCmd := &cobra.Command{
		Use:   "detect",
		Short: "Run anomaly detection on current logs",
		Long: `Analyze current logs against baselines to detect anomalies.

Outputs detections.json with scored anomalies and recommended actions.

Example:
  centipede detect \
    --config config.yaml \
    --baseline baseline.json \
    --log-source ./logs \
    --output detections.json`,
		RunE: func(cmdCobraCmd *cobra.Command, args []string) error {
			return cmd.Run(context.Background())
		},
	}

	detectCmd.Flags().StringVar(&cmd.ConfigPath, "config", "config.yaml", "Path to configuration file")
	detectCmd.Flags().StringVar(&cmd.BaselinePath, "baseline", "baseline.json", "Path to baseline file")
	detectCmd.Flags().StringVar(&cmd.LogSource, "log-source", "", "Path to log files (required)")
	detectCmd.Flags().StringVar(&cmd.OutputPath, "output", "detections.json", "Output detection file path")

	_ = detectCmd.MarkFlagRequired("log-source")

	return detectCmd
}

// Run executes the detect command
func (cmd *DetectCmd) Run(ctx context.Context) error {
	// Configure logging
	log.SetJSONOutput(os.Getenv("LOG_FORMAT") == "json")
	if os.Getenv("DEBUG") == "true" {
		log.SetLevel(log.DebugLevel)
	}

	// Load configuration
	log.Info("loading configuration", "config_path", cmd.ConfigPath)
	cfg, err := config.LoadConfig(cmd.ConfigPath)
	if err != nil {
		log.Error("failed to load config", "error", err.Error())
		os.Exit(exitcode.ErrConfig)
	}

	// Load baselines
	log.Info("loading baselines", "baseline_path", cmd.BaselinePath)
	baselineStore := &storage.BaselineStore{}
	baselines, err := baselineStore.LoadBaselines(cmd.BaselinePath)
	if err != nil {
		log.Error("failed to load baselines", "error", err.Error())
		os.Exit(exitcode.ErrData)
	}

	if len(baselines) == 0 {
		log.Error("no baselines found", "baseline_path", cmd.BaselinePath)
		os.Exit(exitcode.ErrData)
	}

	log.Info("baselines loaded", "count", len(baselines))

	// Load current logs
	log.Info("loading logs", "log_source", cmd.LogSource)
	parser := &parsers.GenericJSONParser{}
	loader := parsers.NewLoader(parser)

	calls, err := loader.LoadFromSource(cmd.LogSource)
	if err != nil {
		log.Error("failed to load logs", "error", err.Error())
		os.Exit(exitcode.ErrData)
	}

	if len(calls) == 0 {
		log.Error("no logs found", "log_source", cmd.LogSource)
		os.Exit(exitcode.ErrData)
	}

	log.Info("logs loaded", "api_call_count", len(calls))

	// Run detection
	log.Info("running anomaly detection")
	detector := detection.NewDetector(cfg)
	result := detector.Detect(calls, baselines)

	// Save results
	log.Info("saving detections", "output_path", cmd.OutputPath)
	detectionStore := &storage.DetectionStore{}
	if err := detectionStore.SaveDetections(cmd.OutputPath, result); err != nil {
		log.Error("failed to save detections", "error", err.Error())
		os.Exit(exitcode.ErrExecution)
	}

	// Print summary
	fmt.Println("\nDetection Results:")
	fmt.Printf("  Total Anomalies: %d\n", result.Summary["total"])
	fmt.Printf("  Critical: %d\n", result.Summary["critical"])
	fmt.Printf("  Warning: %d\n", result.Summary["warning"])
	fmt.Printf("  Normal: %d\n", result.Summary["normal"])

	if len(result.Anomalies) > 0 {
		fmt.Println("\nAnomalies Detected:")
		for _, anomaly := range result.Anomalies {
			action := anomaly.RecommendedAction
			if anomaly.RecommendedAction == "" {
				action = "monitor"
			}

			fmt.Printf("  [%s] %s (score: %d) - Action: %s\n",
				anomaly.Timestamp.Format("15:04:05"),
				anomaly.TenantID,
				anomaly.Score,
				action)

			if len(anomaly.Triggers) > 0 {
				fmt.Printf("    Triggers: %v\n", anomaly.Triggers)
			}

			// Log anomaly details
			log.Warn("anomaly detected",
				"tenant", anomaly.TenantID,
				"score", anomaly.Score,
				"action", action,
				"triggers", fmt.Sprintf("%v", anomaly.Triggers))
		}
	}

	fmt.Println("\n✓ Detection complete")
	log.Info("detection complete", "output_file", cmd.OutputPath)

	// Set exit code based on findings
	criticalCount := result.Summary["critical"]
	if criticalCount > 0 {
		return &CommandError{Code: exitcode.CriticalDetected, Msg: "critical anomalies detected"}
	}

	warningCount := result.Summary["warning"]
	if warningCount > 0 {
		return &CommandError{Code: exitcode.WarningDetected, Msg: "warning-level anomalies detected"}
	}

	return nil
}
