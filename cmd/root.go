package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// RootCmd is the base command onto which all other commands are added.
var RootCmd = &cobra.Command{
	Use:   "atlantis",
	Short: "A unified workflow for collaborating on Terraform through GitHub and GitLab",
}

// Execute starts RootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
