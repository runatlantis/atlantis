// Package main is the entrypoint for the CLI.
package main

import (
	"github.com/hootsuite/atlantis/cmd"
	"github.com/spf13/viper"
)

func main() {
	v := viper.New()
	v.Set("version", "0.2.2")

	// We're creating commands manually here rather than using init() functions
	// (as recommended by cobra) because it makes testing easier.
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
