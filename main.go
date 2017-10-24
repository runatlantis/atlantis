package main

import (
	"github.com/hootsuite/atlantis/cmd"
	"github.com/spf13/viper"
)

func main() {
	viper.Set("version", "0.1.2")
	serverCmd := cmd.NewServerCmd()
	cmd.RootCmd.AddCommand(serverCmd.Cmd)
	cmd.RootCmd.AddCommand(cmd.VersionCmd)
	cmd.RootCmd.AddCommand(cmd.BootstrapCmd)
	cmd.Execute()
}
