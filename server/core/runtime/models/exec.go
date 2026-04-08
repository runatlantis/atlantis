// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"fmt"
	"os"
	"os/exec"
)

//go:generate pegomock generate --package mocks -o mocks/mock_exec.go Exec

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
		return "", fmt.Errorf("no command specified")
	}

	envVars := []string{}
	for key, val := range envs {
		envVars = append(envVars, fmt.Sprintf("%s=%s", key, val))
	}

	// TODO: move this os.Environ call out to the server so this
	// can happen once at the beginning
	envVars = append(envVars, os.Environ()...)

	// args[0] is the executable path resolved from the version cache or LookPath,
	// and subsequent args are controlled by server configuration (not user input).
	cmd := exec.Command(args[0], args[1:]...) // #nosec G204 -- executable and args are server-controlled
	cmd.Env = envVars
	cmd.Dir = workdir

	output, err := cmd.CombinedOutput()

	return string(output), err
}
