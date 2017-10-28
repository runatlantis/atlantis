package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// VersionCmd prints the current version.
type VersionCmd struct {
	Viper *viper.Viper
}

// Init returns the runnable cobra command.
func (v *VersionCmd) Init() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the current Atlantis version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("atlantis %s\n", v.Viper.Get("version"))
		},
	}
}
