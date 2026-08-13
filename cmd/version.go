// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// VersionCmd prints the current version.
type VersionCmd struct {
	AtlantisVersion string
}

// Init returns the runnable cobra command.
func (v *VersionCmd) Init() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the current Atlantis version",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("atlantis %s\n", v.AtlantisVersion)
		},
	}
}
