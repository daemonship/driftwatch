package cmd

import (
	"fmt"
	"os"

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
		// Stub: will be implemented in Task 2
		fmt.Fprintln(os.Stderr, "scan: not yet implemented")
		return nil
	},
}

func init() {
	scanCmd.Flags().StringVarP(&configFile, "config", "c", "driftwatch.yml", "config file path")
	scanCmd.Flags().StringVar(&binary, "binary", "", "terraform binary to use (overrides config)")
	rootCmd.AddCommand(scanCmd)
}
