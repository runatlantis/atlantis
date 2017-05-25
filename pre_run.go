package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/hootsuite/atlantis/logging"
)

const InlineShebang = "/bin/sh -e"

// PreRun is a function that will determine whether
func PreRun(c *Config, log *logging.SimpleLogger, execPath string, command *Command) error {
	log.Info("Staring pre run in %s", execPath)
	var execScript string

	if command.commandType == Plan {
		// we create a script from the commands provided
		s, err := createScript(c.PrePlan.Commands)
		if err != nil {
			return err
		}
		execScript = s
	}
	if command.commandType == Apply {
		// we create a script from the commands provided
		s, err := createScript(c.PreApply.Commands)
		if err != nil {
			return err
		}
		execScript = s
	}

	if execScript != "" {
		defer os.Remove(execScript)
		log.Info("Running script %s", execScript)
		// set environment variable for the run.
		// this is to support scripts to use the ENVIRONMENT and WORKSPACE variables in their scripts
		if command.environment != "" {
			os.Setenv("ENVIRONMENT", command.environment)
		}
		if c.TerraformVersion != "" {
			os.Setenv("ATLANTIS_TERRAFORM_VERSION", c.TerraformVersion)
		}
		os.Setenv("WORKSPACE", execPath)
		output, err := execute(execScript)
		if err != nil {
			return err
		}
		log.Info("output: \n%s", output)
	}

	return nil

}

func createScript(cmds []string) (string, error) {
	var scriptName string
	if cmds != nil {
		tmp, err := ioutil.TempFile("/tmp", "atlantis-temp-script")
		if err != nil {
			return "", fmt.Errorf("Error preparing shell script: %s", err)
		}

		scriptName = tmp.Name()

		// Write our contents to it
		writer := bufio.NewWriter(tmp)
		writer.WriteString(fmt.Sprintf("#!%s\n", InlineShebang))
		for _, command := range cmds {
			if _, err := writer.WriteString(command + "\n"); err != nil {
				return "", fmt.Errorf("Error preparing script: %s", err)
			}
		}

		if err := writer.Flush(); err != nil {
			return "", fmt.Errorf("Error flushing file when preparing script: %s", err)
		}
		tmp.Close()
	}

	return scriptName, nil
}

func execute(script string) (string, error) {
	if _, err := os.Stat(script); err == nil {
		os.Chmod(script, 0775)
	}
	localCmd := exec.Command("sh", "-c", script)
	out, err := localCmd.CombinedOutput()
	output := string(out)
	if err != nil {
		return output, fmt.Errorf("Error running script %s: %v %s", script, err, output)
	}

	return output, nil
}
