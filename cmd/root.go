package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "driftwatch",
	Short: "Detect Terraform state drift across workspaces",
	Long: `driftwatch scans one or more Terraform workspaces for drift by running
terraform plan and reporting any resource changes detected.

Minimum Terraform version required: 1.0.0`,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
