package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"github.hootops.com/production-delivery/atlantis/logging"
)

type TerraformClient struct {
	tfExecutableName string
}

func (t *TerraformClient) ConfigureRemoteState(log *logging.SimpleLogger, execPath ExecutionPath, tfEnvName string, sshKey string) (string, error) {
	log.Info("setting up remote state in directory %q", execPath.Absolute)
	var remoteSetupCmdArgs []string

	// Check if setup file exists
	setupFileInfo, err := os.Stat(filepath.Join(execPath.Absolute, "setup.sh"))
	if err != nil {
		return "", fmt.Errorf("setup.sh file doesn't exist in terraform plan path %q: %v", execPath.Relative, err)
	}
	// Check if setup file is executed
	if setupFileInfo.Mode() != os.FileMode(0755) {
		return "", fmt.Errorf("setup file isn't executable, required permissions are 0755")
	}
	// Check if environment is specified
	if tfEnvName == "" {
		// Check if env/ folder exist and environment isn't specified
		if _, err := os.Stat(filepath.Join(execPath.Absolute, "env")); err == nil {
			log.Info("environment directory exists but no environment was supplied")
			return "", nil
		}
	} else {
		// Check if the environment file exists
		envFile := tfEnvName + ".tfvars"
		envPath := filepath.Join(execPath.Absolute, "env")
		if _, err := os.Stat(filepath.Join(envPath, envFile)); err != nil {
			return "", fmt.Errorf("environment file %q not found in %q", envFile, envPath)
		}
	}

	// Set environment file parameter for ./setup.sh
	if tfEnvName != "" {
		remoteSetupCmdArgs = append(remoteSetupCmdArgs, "-e", tfEnvName)
	}
	remoteSetupCmd := exec.Command("./setup.sh", remoteSetupCmdArgs...)
	remoteSetupCmd.Dir = execPath.Absolute
	// Check if ssh key is set
	if sshKey != "" {
		// Fixing a bug when git isn't found in path when environment variables are set for command
		remoteSetupCmd.Env = os.Environ()
		remoteSetupCmd.Env = append(remoteSetupCmd.Env, fmt.Sprintf("GIT_SSH=%s", defaultSSHWrapper), fmt.Sprintf("PKEY=%s", sshKey))

	}
	out, err := remoteSetupCmd.CombinedOutput()
	output := string(out[:])
	log.Info("setup.sh output: \n%s", output)
	if err != nil {
		return "", fmt.Errorf("failed to configure remote state: %v", err)
	}
	// Get remote state path from setup.sh output
	r, _ := regexp.Compile("REMOTE_STATE_PATH=([^ ]*.tfstate)")
	match := r.FindStringSubmatch(output)
	// Store remote state path
	remoteStatePath := match[1]
	log.Info("remote state path %q", remoteStatePath)

	return remoteStatePath, nil
}

func (t *TerraformClient) RunTerraformCommand(path ExecutionPath, tfPlanCmd []string, tfEnvVars []string) ([]string, string, error) {
	terraformCmd := exec.Command(t.tfExecutableName, tfPlanCmd...)
	terraformCmd.Dir = path.Absolute
	terraformCmd.Env = tfEnvVars
	out, err := terraformCmd.CombinedOutput()
	output := string(out)
	if err != nil {
		return terraformCmd.Args, output, err
	}

	return terraformCmd.Args, output, nil
}
