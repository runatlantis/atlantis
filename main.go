package main

import (
	"github.com/hootsuite/atlantis/cmd"
	"github.com/spf13/viper"
)

func main() {
	v := viper.New()
	v.Set("version", "0.1.2")

	server := &cmd.ServerCmd{
		ServerCreator: &cmd.DefaultServerCreator{},
		Viper:         v,
	}
	version := &cmd.VersionCmd{Viper: v}
	bootstrap := &cmd.BootstrapCmd{}
	cmd.RootCmd.AddCommand(server.Init())
	cmd.RootCmd.AddCommand(version.Init())
	cmd.RootCmd.AddCommand(bootstrap.Init())
	cmd.Execute()
}
