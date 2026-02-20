package cmd

import (
	"fmt"
	"os"

	"github.com/daemonship/driftwatch/internal/config"
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
		results := runner.RunAll(cfg.Workspaces, opts)

		// Print results summary
		var hasDrift, hasError bool
		for _, r := range results {
			if r.Err != nil {
				hasError = true
				fmt.Fprintf(os.Stderr, "Error scanning %s: %v\n", r.WorkspacePath, r.Err)
				continue
			}
			if r.ExitCode == 2 {
				hasDrift = true
				fmt.Printf("Drift detected in %s\n", r.WorkspacePath)
			} else if r.ExitCode == 0 {
				fmt.Printf("No drift in %s\n", r.WorkspacePath)
			} else {
				hasError = true
				fmt.Fprintf(os.Stderr, "Error in %s (exit code %d): %s\n", r.WorkspacePath, r.ExitCode, string(r.Stderr))
			}
		}

		// Set exit code
		if hasError {
			os.Exit(2)
		}
		if hasDrift {
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	scanCmd.Flags().StringVarP(&configFile, "config", "c", "driftwatch.yml", "config file path")
	scanCmd.Flags().StringVar(&binary, "binary", "", "terraform binary to use (overrides config)")
	rootCmd.AddCommand(scanCmd)
}
