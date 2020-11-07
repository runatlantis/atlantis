package models

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_exec.go Exec

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
	formattedArgs := strings.Join(args, " ")

	envVars := []string{}
	for key, val := range envs {
		envVars = append(envVars, fmt.Sprintf("%s=%s", key, val))
	}

	// TODO: move this os.Environ call out to the server so this
	// can happen once at the beginning
	envVars = append(envVars, os.Environ()...)

	// honestly not entirely sure why we're using sh -c but it's used
	// for the terraform binary so copying it for now
	cmd := exec.Command("sh", "-c", formattedArgs)
	cmd.Env = envVars
	cmd.Dir = workdir

	output, err := cmd.CombinedOutput()

	return string(output), err
}
