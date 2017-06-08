package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

const version = "0.1.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the current atlantis version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("atlantis %s\n", version)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
