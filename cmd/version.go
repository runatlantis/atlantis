package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the current atlantis version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("atlantis %s\n", viper.Get("version"))
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
