package cmd

import (
	"fmt"
	"os"

	"github.com/atlantisnorth/atlantis/bootstrap"
	"github.com/spf13/cobra"
)

// BootstrapCmd starts the bootstrap process for testing out Atlantis.
type BootstrapCmd struct{}

// Init returns the runnable cobra command.
func (b *BootstrapCmd) Init() *cobra.Command {
	return &cobra.Command{
		Use:   "bootstrap",
		Short: "Start a guided tour of Atlantis",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := bootstrap.Start()
			if err != nil {
				fmt.Fprintf(os.Stderr, "\033[31mError: %s\033[39m\n\n", err.Error())
			}
			return err
		},
	}
}
