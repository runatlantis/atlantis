// Package main is the entrypoint for the CLI.
package main

import (
	"github.com/runatlantis/atlantis/cmd"
	"github.com/spf13/viper"
)

const atlantisVersion = "0.3.3"

func main() {
	v := viper.New()

	// We're creating commands manually here rather than using init() functions
	// (as recommended by cobra) because it makes testing easier.
	server := &cmd.ServerCmd{
		ServerCreator:   &cmd.DefaultServerCreator{},
		Viper:           v,
		AtlantisVersion: atlantisVersion,
	}
	version := &cmd.VersionCmd{AtlantisVersion: atlantisVersion}
	bootstrap := &cmd.BootstrapCmd{}
	cmd.RootCmd.AddCommand(server.Init())
	cmd.RootCmd.AddCommand(version.Init())
	cmd.RootCmd.AddCommand(bootstrap.Init())
	cmd.Execute()
}
