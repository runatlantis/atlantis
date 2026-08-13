// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package cmd

import (
	"fmt"
	"os"

	"github.com/runatlantis/atlantis/testdrive"
	"github.com/spf13/cobra"
)

// TestdriveCmd starts the testdrive process for testing out Atlantis.
type TestdriveCmd struct{}

// Init returns the runnable cobra command.
func (b *TestdriveCmd) Init() *cobra.Command {
	return &cobra.Command{
		Use:   "testdrive",
		Short: "Start a guided tour of Atlantis",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := testdrive.Start()
			if err != nil {
				fmt.Fprintf(os.Stderr, "\033[31mError: %s\033[39m\n\n", err.Error())
			}
			return err
		},
		SilenceErrors: true,
	}
}
