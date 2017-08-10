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

	version "github.com/hashicorp/go-version"
	"github.com/hootsuite/atlantis/logging"
	"github.com/pkg/errors"
)

const inlineShebang = "#!/bin/sh -e"

type Run struct{}

// Execute runs the commands by writing them as a script to disk
// and then executing the script.
func (p *Run) Execute(
	log *logging.SimpleLogger,
	commands []string,
	path string,
	environment string,
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
	defer os.Remove(s)

	log.Info("running %s commands: %v", stage, commands)

	// set environment variable for the run.
	// this is to support scripts to use the ENVIRONMENT, ATLANTIS_TERRAFORM_VERSION
	// and WORKSPACE variables in their scripts
	os.Setenv("ENVIRONMENT", environment)
	os.Setenv("ATLANTIS_TERRAFORM_VERSION", terraformVersion.String())
	os.Setenv("WORKSPACE", path)
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
	writer.WriteString(fmt.Sprintf("%s\n", inlineShebang))
	cmdsJoined := strings.Join(cmds, "\n")
	if _, err := writer.WriteString(cmdsJoined); err != nil {
		return "", errors.Wrapf(err, "preparing %s", stage)
	}

	if err := writer.Flush(); err != nil {
		return "", errors.Wrap(err, "flushing contents to file")
	}
	tmp.Close()

	if err := os.Chmod(scriptName, 0755); err != nil {
		return "", errors.Wrapf(err, "making %s script executable", stage)
	}

	return scriptName, nil
}

func execute(script string) (string, error) {
	localCmd := exec.Command("sh", "-c", script)
	out, err := localCmd.CombinedOutput()
	output := string(out)
	if err != nil {
		return output, errors.Wrapf(err, "running script %s: %s", script, output)
	}

	return output, nil
}
