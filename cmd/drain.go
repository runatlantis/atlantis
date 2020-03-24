package cmd

import (
	"github.com/runatlantis/atlantis/drain"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/spf13/cobra"
)

// DrainCmd performs a drain of the local Atlantis server for all running operations.
// The server itself is not shutdown but drained from all running operations.
// When the command returns, the "atlantis server" process can be stopped securely.
type DrainCmd struct {
	Logger *logging.SimpleLogger
}

// Drain returns the runnable cobra command.
func (v *DrainCmd) Init() *cobra.Command {
	return &cobra.Command{
		Use:   "drain",
		Short: "Perform a drain of the local Atlantis server, waiting for completion before returning",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := drain.Start(v.Logger)
			return err
		},
		SilenceErrors: true,
	}
}
