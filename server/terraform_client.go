package server

import (
	"fmt"
	"github.com/hootsuite/atlantis/logging"
	"github.com/hootsuite/atlantis/models"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

type TerraformClient struct {
	tfExecutableName string
}

func (t *TerraformClient) ConfigureRemoteState(log *logging.SimpleLogger, repoDir string, project models.Project, env string, sshKey string) (string, error) {
	absolutePath := filepath.Join(repoDir, project.Path)
	log.Info("setting up remote state in directory %q", absolutePath)
	var remoteSetupCmdArgs []string

	// Check if setup file exists
	setupFileInfo, err := os.Stat(filepath.Join(absolutePath, "setup.sh"))
	if err != nil {
		return "", fmt.Errorf("setup.sh file doesn't exist in terraform plan path %q: %v", project.Path, err)
	}
	// Check if setup file is executed
	if setupFileInfo.Mode() != os.FileMode(0755) {
		return "", fmt.Errorf("setup file isn't executable, required permissions are 0755")
	}
	// Check if environment is specified
	if env == "" {
		// Check if env/ folder exist and environment isn't specified
		// todo: make env a constant (environmentDirName)
		if _, err := os.Stat(filepath.Join(absolutePath, "env")); err == nil {
			log.Info("environment directory exists but no environment was supplied")
			return "", nil
		}
	} else {
		// Check if the environment file exists
		envFile := env + ".tfvars"
		envPath := filepath.Join(absolutePath, "env")
		if _, err := os.Stat(filepath.Join(envPath, envFile)); err != nil {
			return "", fmt.Errorf("environment file %q not found in %q", envFile, envPath)
		}
	}

	// Set environment file parameter for ./setup.sh
	if env != "" {
		remoteSetupCmdArgs = append(remoteSetupCmdArgs, "-e", env)
	}
	remoteSetupCmd := exec.Command("./setup.sh", remoteSetupCmdArgs...)
	remoteSetupCmd.Dir = absolutePath
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
	// Backend remote state path
	remoteStatePath := match[1]
	log.Info("remote state path %q", remoteStatePath)

	return remoteStatePath, nil
}

func (t *TerraformClient) RunTerraformCommand(path string, tfPlanCmd []string, tfEnvVars []string) ([]string, string, error) {
	terraformCmd := exec.Command(t.tfExecutableName, tfPlanCmd...)
	terraformCmd.Dir = path
	terraformCmd.Env = tfEnvVars
	out, err := terraformCmd.CombinedOutput()
	output := string(out)
	if err != nil {
		return terraformCmd.Args, output, err
	}

	return terraformCmd.Args, output, nil
}
