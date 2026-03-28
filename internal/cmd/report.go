package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/bogdanticu88/centipede/internal/storage"
	"github.com/spf13/cobra"
)

// ReportCmd is the report command
type ReportCmd struct {
	DetectionsPath string
	BaselinePath   string
	OutputPath     string
	Format         string
}

// NewReportCmd creates a new report command
func NewReportCmd() *cobra.Command {
	cmd := &ReportCmd{}

	reportCmd := &cobra.Command{
		Use:   "report",
		Short: "Generate HTML/JSON report from detections",
		Long: `Create a comprehensive report from detection results.

Supports both HTML (human-readable) and JSON (machine-readable) formats.

Example:
  centipede report \
    --detections detections.json \
    --baseline baseline.json \
    --format html \
    --output report.html`,
		RunE: func(cmdCobraCmd *cobra.Command, args []string) error {
			return cmd.Run(context.Background())
		},
	}

	reportCmd.Flags().StringVar(&cmd.DetectionsPath, "detections", "detections.json", "Path to detections file")
	reportCmd.Flags().StringVar(&cmd.BaselinePath, "baseline", "baseline.json", "Path to baseline file")
	reportCmd.Flags().StringVar(&cmd.OutputPath, "output", "report.html", "Output report file path")
	reportCmd.Flags().StringVar(&cmd.Format, "format", "html", "Report format: html or json")

	return reportCmd
}

// Run executes the report command
func (cmd *ReportCmd) Run(ctx context.Context) error {
	// Load detections
	fmt.Printf("Loading detections from %s...\n", cmd.DetectionsPath)
	detectionStore := &storage.DetectionStore{}
	detections, err := detectionStore.LoadDetections(cmd.DetectionsPath)
	if err != nil {
		return fmt.Errorf("failed to load detections: %w", err)
	}

	// Load baselines
	fmt.Printf("Loading baselines from %s...\n", cmd.BaselinePath)
	baselineStore := &storage.BaselineStore{}
	_, err2 := baselineStore.LoadBaselines(cmd.BaselinePath)
	if err2 != nil {
		return fmt.Errorf("failed to load baselines: %w", err2)
	}

	// Generate report
	var reportContent string

	switch cmd.Format {
	case "html":
		fmt.Println("Generating HTML report...")
		reportContent = generateHTMLReport(detections)
	case "json":
		fmt.Println("Generating JSON report...")
		reportContent = generateJSONReport(detections)
	default:
		return fmt.Errorf("unsupported format: %s", cmd.Format)
	}

	// Write report
	fmt.Printf("Writing report to %s...\n", cmd.OutputPath)
	if err := os.WriteFile(cmd.OutputPath, []byte(reportContent), 0644); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	fmt.Println("\n✓ Report generated successfully")
	return nil
}

// generateHTMLReport generates an HTML report
func generateHTMLReport(detections interface{}) string {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>CENTIPEDE Detection Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #f5f5f5; }
        .header { background: #2c3e50; color: white; padding: 20px; border-radius: 5px; }
        .summary { background: white; padding: 20px; margin: 20px 0; border-radius: 5px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .critical { color: #e74c3c; font-weight: bold; }
        .warning { color: #f39c12; font-weight: bold; }
        .normal { color: #27ae60; font-weight: bold; }
        table { width: 100%; border-collapse: collapse; background: white; border-radius: 5px; overflow: hidden; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        th { background: #34495e; color: white; padding: 10px; text-align: left; }
        td { padding: 10px; border-bottom: 1px solid #ddd; }
        tr:hover { background: #f9f9f9; }
    </style>
</head>
<body>
    <div class="header">
        <h1>CENTIPEDE Detection Report</h1>
        <p>Multi-Cloud API Anomaly Detection & Tenant Protection</p>
    </div>

    <div class="summary">
        <h2>Detections Summary</h2>
        <p>Report generated for API threat detection and tenant isolation.</p>
        <p><em>Full report details to be populated from detections data.</em></p>
    </div>
</body>
</html>`
	return html
}

// generateJSONReport generates a JSON report
func generateJSONReport(detections interface{}) string {
	// For now, just return a simple JSON structure
	// This would be enhanced to format the actual detections
	json := `{
  "report": "JSON report format not yet implemented",
  "detections": null,
  "baselines": null
}`
	return json
}
