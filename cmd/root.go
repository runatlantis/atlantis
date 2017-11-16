// Package cmd holds all our cli commands.
// These are different from the commands that get run via pull request comments.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "atlantis",
	Short: "Manage your Terraform workflow from GitHub",
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
