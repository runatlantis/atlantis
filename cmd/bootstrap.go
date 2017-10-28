package cmd

import (
	"github.com/hootsuite/atlantis/bootstrap"
	"github.com/spf13/cobra"
)

// BootstrapCmd starts the bootstrap process for testing out Atlantis.
type BootstrapCmd struct{}

// Init returns the runnable cobra command.
func (b *BootstrapCmd) Init() *cobra.Command {
	return &cobra.Command{
		Use:   "bootstrap",
		Short: "Start a guided tour of Atlantis",
		RunE: withErrPrint(func(cmd *cobra.Command, args []string) error {
			return bootstrap.Start()
		}),
	}
}
