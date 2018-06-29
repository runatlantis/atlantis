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
	versionCmdOutput, err := exec.Command("terraform", "version").CombinedOutput() // #nosec
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
// the provided args in path. v is the version of terraform executable to use.
// If v is nil, will use the default version.
// Workspace is the terraform workspace to run in. We won't switch workspaces
// but will set the TERRAFORM_WORKSPACE environment variable.
func (c *DefaultClient) RunCommandWithVersion(log *logging.SimpleLogger, path string, args []string, v *version.Version, workspace string) (string, error) {
	tfExecutable := "terraform"
	tfVersionStr := c.defaultVersion.String()
	// if version is the same as the default, don't need to prepend the version name to the executable
	if v != nil && !v.Equal(c.defaultVersion) {
		tfExecutable = fmt.Sprintf("%s%s", tfExecutable, v.String())
		tfVersionStr = v.String()
	}

	// We add custom variables so that if `extra_args` is specified with env
	// vars then they'll be substituted.
	envVars := []string{
		// Will de-emphasize specific commands to run in output.
		"TF_IN_AUTOMATION=true",
		// Cache plugins so terraform init runs faster.
		fmt.Sprintf("TF_PLUGIN_CACHE_DIR=%s", c.terraformPluginCacheDir),
		// Terraform will run all commands in this workspace. We should have
		// already selected this workspace but this is a fail-safe to ensure
		// we're operating in the right workspace.
		fmt.Sprintf("TF_WORKSPACE=%s", workspace),
		// We're keeping this variable even though it duplicates TF_WORKSPACE
		// because it's probably safer for users to rely on it. Terraform might
		// change the way TF_WORKSPACE works in the future.
		fmt.Sprintf("WORKSPACE=%s", workspace),
		fmt.Sprintf("ATLANTIS_TERRAFORM_VERSION=%s", tfVersionStr),
		fmt.Sprintf("DIR=%s", path),
	}
	// Append current Atlantis process's environment variables so PATH is
	// preserved and any vars that users purposely exec'd Atlantis with.
	envVars = append(envVars, os.Environ()...)

	// append terraform executable name with args
	tfCmd := fmt.Sprintf("%s %s", tfExecutable, strings.Join(args, " "))

	// We use 'sh -c' so that if extra_args have been specified with env vars,
	// ex. -var-file=$WORKSPACE.tfvars, then they get substituted.
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
