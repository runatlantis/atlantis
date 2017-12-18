// Package run handles running commands prior and following the
// regular Atlantis commands.
package run

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/hootsuite/atlantis/server/logging"
	"github.com/pkg/errors"
)

const inlineShebang = "#!/bin/sh -e"

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_runner.go Runner

type Runner interface {
	Execute(log *logging.SimpleLogger, commands []string, path string, workspace string, terraformVersion *version.Version, stage string) (string, error)
}

type Run struct{}

// Execute runs the commands by writing them as a script to disk
// and then executing the script.
func (p *Run) Execute(
	log *logging.SimpleLogger,
	commands []string,
	path string,
	workspace string,
	terraformVersion *version.Version,
	stage string) (string, error) {
	// we create a script from the commands provided
	if len(commands) == 0 {
		return "", errors.Errorf("%s commands cannot be empty", stage)
	}

	s, err := createScript(commands, stage)
	if err != nil {
		return "", err
	}
	defer os.Remove(s) // nolint: errcheck

	log.Info("running %s commands: %v", stage, commands)

	// set environment variable for the run.
	// this is to support scripts to use the WORKSPACE, ATLANTIS_TERRAFORM_VERSION
	// and DIR variables in their scripts
	os.Setenv("WORKSPACE", workspace)                                  // nolint: errcheck
	os.Setenv("ATLANTIS_TERRAFORM_VERSION", terraformVersion.String()) // nolint: errcheck
	os.Setenv("DIR", path)                                             // nolint: errcheck
	return execute(s)
}

func createScript(cmds []string, stage string) (string, error) {
	tmp, err := ioutil.TempFile("/tmp", "atlantis-temp-script")
	if err != nil {
		return "", errors.Wrapf(err, "preparing %s shell script", stage)
	}

	scriptName := tmp.Name()

	// Write our contents to it
	writer := bufio.NewWriter(tmp)
	if _, err = writer.WriteString(fmt.Sprintf("%s\n", inlineShebang)); err != nil {
		return "", errors.Wrapf(err, "writing to %q", tmp.Name())
	}
	cmdsJoined := strings.Join(cmds, "\n")
	if _, err := writer.WriteString(cmdsJoined); err != nil {
		return "", errors.Wrapf(err, "preparing %s", stage)
	}

	if err := writer.Flush(); err != nil {
		return "", errors.Wrap(err, "flushing contents to file")
	}
	tmp.Close() // nolint: errcheck

	if err := os.Chmod(scriptName, 0700); err != nil { // nolint: gas
		return "", errors.Wrapf(err, "making %s script executable", stage)
	}

	return scriptName, nil
}

func execute(script string) (string, error) {
	localCmd := exec.Command("sh", "-c", script) // #nosec
	out, err := localCmd.CombinedOutput()
	output := string(out)
	if err != nil {
		return output, errors.Wrapf(err, "running script %s: %s", script, output)
	}

	return output, nil
}
