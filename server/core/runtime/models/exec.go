// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

//go:generate go tool pegomock generate --package mocks -o mocks/mock_exec.go Exec

type Exec interface {
	LookPath(file string) (string, error)
	CombinedOutput(args []string, envs map[string]string, workdir string) (string, error)
}

type LocalExec struct{}

func (e LocalExec) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

// CombinedOutput encapsulates creating a command and running it. We should think about
// how to flexibly add parameters here as this is meant to satisfy very simple usecases
// for more complex usecases we can add a Command function to this method which will
// allow us to edit a Cmd directly.
func (e LocalExec) CombinedOutput(args []string, envs map[string]string, workdir string) (string, error) {
	if len(args) == 0 {
		return "", errors.New("no command specified")
	}

	envVars := []string{}
	for key, val := range envs {
		envVars = append(envVars, fmt.Sprintf("%s=%s", key, val))
	}

	// TODO: move this os.Environ call out to the server so this
	// can happen once at the beginning
	envVars = append(envVars, os.Environ()...)

	// Run the binary directly rather than joining the arguments into a string
	// and handing it to `sh -c`. Some of these arguments are derived from
	// pull-request controlled input: the conftest input file is named after the
	// project name and workspace from a repo-level atlantis.yaml. Those are
	// validated separately, but passing an argument vector means a
	// metacharacter that slips through a validator is never interpreted.
	cmd := exec.Command(args[0], args[1:]...) // #nosec G204 -- args[0] is a resolved executable path, and the arguments are passed as a vector rather than as shell source
	cmd.Env = envVars
	cmd.Dir = workdir

	output, err := cmd.CombinedOutput()

	return string(output), err
}
