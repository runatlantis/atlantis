package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// RootCmd is the base command onto which all other commands are added.
var RootCmd = &cobra.Command{
	Use:   "atlantis",
	Short: "Manage your Terraform workflow from GitHub",
}

// Execute starts RootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
