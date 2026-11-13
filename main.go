// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.
//
// Package main is the entrypoint for the CLI.
package main

import (
	"fmt"

	"github.com/runatlantis/atlantis/cmd"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/spf13/viper"
)

// All of this is filled in by goreleaser upon release
// https://goreleaser.com/cookbooks/using-main.version/
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {

	v := viper.New()

	logger, err := logging.NewStructuredLogger()

	logger.Debug("atlantis %s, commit %s, built at %s\n", version, commit, date)
	if err != nil {
		panic(fmt.Sprintf("unable to initialize logger. %s", err.Error()))
	}

	var sha = commit
	if len(commit) >= 7 {
		sha = commit[:7]
	}

	atlantisVersion := fmt.Sprintf("%s (commit: %s) (build date: %s)", version, sha, date)

	// We're creating commands manually here rather than using init() functions
	// (as recommended by cobra) because it makes testing easier.
	server := &cmd.ServerCmd{
		ServerCreator:   &cmd.DefaultServerCreator{},
		Viper:           v,
		AtlantisVersion: atlantisVersion,
		Logger:          logger,
	}
	version := &cmd.VersionCmd{AtlantisVersion: atlantisVersion}
	testdrive := &cmd.TestdriveCmd{}
	cmd.RootCmd.AddCommand(server.Init())
	cmd.RootCmd.AddCommand(version.Init())
	cmd.RootCmd.AddCommand(testdrive.Init())
	cmd.Execute()
}
