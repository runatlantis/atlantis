// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.
//
// Package terraform handles the actual running of terraform commands.
package terraform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/logging"
)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_terraform_client.go Client

type Client interface {
	Version() *version.Version
	RunCommandWithVersion(log *logging.SimpleLogger, path string, args []string, v *version.Version, workspace string) (string, error)
	Init(log *logging.SimpleLogger, path string, workspace string, extraInitArgs []string, version *version.Version) ([]string, error)
}

type DefaultClient struct {
	defaultVersion          *version.Version
	terraformPluginCacheDir string
}

const terraformPluginCacheDirName = "plugin-cache"

// zeroPointNine constrains the version to be 0.9.*
var zeroPointNine = MustConstraint(">=0.9,<0.10")
var versionRegex = regexp.MustCompile("Terraform v(.*)\n")

func NewClient(dataDir string) (*DefaultClient, error) {
	// todo: use exec.LookPath to find out if we even have terraform rather than
	// parsing the error looking for a not found error.
	versionCmdOutput, err := exec.Command("terraform", "version").Output() // #nosec
	output := string(versionCmdOutput)
	if err != nil {
		// exec.go line 35, Error() returns
		// "exec: " + strconv.Quote(e.Name) + ": " + e.Err.Error()
		if err.Error() == fmt.Sprintf("exec: \"terraform\": %s", exec.ErrNotFound.Error()) {
			return nil, errors.New("terraform not found in $PATH. \n\nDownload terraform from https://www.terraform.io/downloads.html")
		}
		return nil, errors.Wrapf(err, "running terraform version: %s", output)
	}
	match := versionRegex.FindStringSubmatch(output)
	if len(match) <= 1 {
		return nil, fmt.Errorf("could not parse terraform version from %s", output)
	}
	v, err := version.NewVersion(match[1])
	if err != nil {
		return nil, errors.Wrap(err, "parsing terraform version")
	}

	// We will run terraform with the TF_PLUGIN_CACHE_DIR env var set to this
	// directory inside our data dir.
	cacheDir := filepath.Join(dataDir, terraformPluginCacheDirName)
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		return nil, errors.Wrapf(err, "unable to create terraform plugin cache directory at %q", terraformPluginCacheDirName)
	}

	return &DefaultClient{
		defaultVersion:          v,
		terraformPluginCacheDir: cacheDir,
	}, nil
}

// Version returns the version of the terraform executable in our $PATH.
func (c *DefaultClient) Version() *version.Version {
	return c.defaultVersion
}

// RunCommandWithVersion executes the provided version of terraform with
// the provided args in path. v is the version of terraform executable to use
// and workspace is the workspace specified by the user commenting
// "atlantis plan/apply {workspace}" which is set to "default" by default.
func (c *DefaultClient) RunCommandWithVersion(log *logging.SimpleLogger, path string, args []string, v *version.Version, workspace string) (string, error) {
	tfExecutable := "terraform"
	// if version is the same as the default, don't need to prepend the version name to the executable
	if !v.Equal(c.defaultVersion) {
		tfExecutable = fmt.Sprintf("%s%s", tfExecutable, v.String())
	}

	// set environment variables
	// this is to support scripts to use the WORKSPACE, ATLANTIS_TERRAFORM_VERSION
	// and DIR variables in their scripts
	// append current process's environment variables
	// this is to prevent the $PATH variable being removed from the environment
	envVars := []string{
		// Will de-emphasize specific commands to run in output.
		"TF_IN_AUTOMATION=true",
		// Cache plugins so terraform init runs faster.
		fmt.Sprintf("TF_PLUGIN_CACHE_DIR=%s", c.terraformPluginCacheDir),
		fmt.Sprintf("WORKSPACE=%s", workspace),
		fmt.Sprintf("ATLANTIS_TERRAFORM_VERSION=%s", v.String()),
		fmt.Sprintf("DIR=%s", path),
	}
	envVars = append(envVars, os.Environ()...)

	// append terraform executable name with args
	tfCmd := fmt.Sprintf("%s %s", tfExecutable, strings.Join(args, " "))

	terraformCmd := exec.Command("sh", "-c", tfCmd) // #nosec
	terraformCmd.Dir = path
	terraformCmd.Env = envVars
	out, err := terraformCmd.CombinedOutput()
	commandStr := strings.Join(terraformCmd.Args, " ")
	if err != nil {
		err = fmt.Errorf("%s: running %q in %q: \n%s", err, commandStr, path, out)
		log.Debug("error: %s", err)
		return string(out), err
	}
	log.Info("successfully ran %q in %q", commandStr, path)
	return string(out), nil
}

// Init executes "terraform init" and "terraform workspace select" in path.
// workspace is the workspace to select and extraInitArgs are additional arguments
// applied to the init command. version is the terraform version being executed.
// Init is guaranteed to be called with version >= 0.9 since the init command
// was only introduced in that version. It properly handles the renaming of the
// env command to workspace since 0.10.
//
// Returns the string outputs of running each command.
func (c *DefaultClient) Init(log *logging.SimpleLogger, path string, workspace string, extraInitArgs []string, version *version.Version) ([]string, error) {
	var outputs []string

	output, err := c.RunCommandWithVersion(log, path, append([]string{"init", "-no-color"}, extraInitArgs...), version, workspace)
	outputs = append(outputs, output)
	if err != nil {
		return outputs, err
	}

	workspaceCommand := "workspace"
	runningZeroPointNine := zeroPointNine.Check(version)
	if runningZeroPointNine {
		// In 0.9.* `env` was used instead of `workspace`
		workspaceCommand = "env"
	}

	// Use `workspace show` to find out what workspace we're in now. If we're
	// already in the right workspace then no need to switch. This will save us
	// about ten seconds. This command is only available in > 0.10.
	if !runningZeroPointNine {
		workspaceShowOutput, err := c.RunCommandWithVersion(log, path, []string{workspaceCommand, "show"}, version, workspace) // nolint:vetshadow
		outputs = append(outputs, workspaceShowOutput)
		if err != nil {
			return outputs, err
		}
		if strings.TrimSpace(workspaceShowOutput) == workspace {
			return outputs, nil
		}
	}

	output, err = c.RunCommandWithVersion(log, path, []string{workspaceCommand, "select", "-no-color", workspace}, version, workspace)
	outputs = append(outputs, output)
	if err != nil {
		// If terraform workspace select fails we run terraform workspace
		// new to create a new workspace automatically.
		output, err = c.RunCommandWithVersion(log, path, []string{workspaceCommand, "new", "-no-color", workspace}, version, workspace)
		outputs = append(outputs, output)
		if err != nil {
			return outputs, err
		}
	}
	return outputs, nil
}

// MustConstraint will parse one or more constraints from the given
// constraint string. The string must be a comma-separated list of
// constraints. It panics if there is an error.
func MustConstraint(v string) version.Constraints {
	c, err := version.NewConstraint(v)
	if err != nil {
		panic(err)
	}
	return c
}
