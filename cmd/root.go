// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// RootCmd is the base command onto which all other commands are added.
var RootCmd = &cobra.Command{
	Use:   "atlantis",
	Short: "Terraform Pull Request Automation",
}

// Execute starts RootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
