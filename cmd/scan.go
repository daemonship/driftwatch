package cmd

import (
	"fmt"
	"os"

	"github.com/daemonship/driftwatch/internal/config"
	"github.com/daemonship/driftwatch/internal/report"
	"github.com/daemonship/driftwatch/internal/runner"
	"github.com/spf13/cobra"
)

var (
	configFile string
	binary     string
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan workspaces for Terraform state drift",
	Long: `Scan reads driftwatch.yml (or a specified config file), iterates each
configured workspace, runs terraform plan, and reports any detected drift.

Exit codes:
  0 — no drift detected
  1 — drift detected in one or more workspaces
  2 — scan error occurred (plan could not be run)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		cfg, err := config.Load(configFile)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		// Determine binary to use: CLI flag > config > default
		tfBinary := binary
		if tfBinary == "" {
			tfBinary = cfg.Binary
		}
		if tfBinary == "" {
			tfBinary = "terraform"
		}

		opts := runner.Options{Binary: tfBinary}
		runnerResults := runner.RunAll(cfg.Workspaces, opts)

		// Convert runner results to report results (parsing JSON)
		results, err := report.WorkspaceResultsFromRunnerResults(runnerResults)
		if err != nil {
			return fmt.Errorf("processing results: %w", err)
		}

		// Print the human-readable report
		report.Print(os.Stdout, results)

		// Set exit code based on results
		os.Exit(report.ExitCode(results))
		return nil
	},
}

func init() {
	scanCmd.Flags().StringVarP(&configFile, "config", "c", "driftwatch.yml", "config file path")
	scanCmd.Flags().StringVar(&binary, "binary", "", "terraform binary to use (overrides config)")
	rootCmd.AddCommand(scanCmd)
}
