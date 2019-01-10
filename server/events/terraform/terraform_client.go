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
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mitchellh/go-homedir"

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

// versionRegex extracts the version from `terraform version` output.
//     Terraform v0.12.0-alpha4 (2c36829d3265661d8edbd5014de8090ea7e2a076)
//	   => 0.12.0-alpha4
//
//     Terraform v0.11.10
//	   => 0.11.10
var versionRegex = regexp.MustCompile("Terraform v(.*?)(\\s.*)?\n")

func NewClient(dataDir string, tfeToken string) (*DefaultClient, error) {
	_, err := exec.LookPath("terraform")
	if err != nil {
		return nil, errors.New("terraform not found in $PATH. \n\nDownload terraform from https://www.terraform.io/downloads.html")
	}
	versionOutBytes, err := exec.Command("terraform", "version").
		Output() // #nosec
	versionOutput := string(versionOutBytes)
	if err != nil {
		return nil, errors.Wrapf(err, "running terraform version: %s", versionOutput)
	}
	match := versionRegex.FindStringSubmatch(versionOutput)
	if len(match) <= 1 {
		return nil, fmt.Errorf("could not parse terraform version from %s", versionOutput)
	}
	v, err := version.NewVersion(match[1])
	if err != nil {
		return nil, errors.Wrap(err, "parsing terraform version")
	}

	// If tfeToken is set, we try to create a ~/.terraformrc file.
	if tfeToken != "" {
		home, err := homedir.Dir()
		if err != nil {
			return nil, errors.Wrap(err, "getting home dir to write ~/.terraformrc file")
		}
		if err := generateRCFile(tfeToken, home); err != nil {
			return nil, err
		}
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

// generateRCFile generates a .terraformrc file containing config for tfeToken.
// It will create the file in home/.terraformrc.
func generateRCFile(tfeToken string, home string) error {
	const rcFilename = ".terraformrc"
	rcFile := filepath.Join(home, rcFilename)
	config := fmt.Sprintf(rcFileContents, tfeToken)

	// If there is already a .terraformrc file and its contents aren't exactly
	// what we would have written to it, then we error out because we don't
	// want to overwrite anything.
	if _, err := os.Stat(rcFile); err == nil {
		currContents, err := ioutil.ReadFile(rcFile) // nolint: gosec
		if err != nil {
			return errors.Wrapf(err, "trying to read %s to ensure we're not overwriting it", rcFile)
		}
		if config != string(currContents) {
			return fmt.Errorf("can't write TFE token to %s because that file has contents that would be overwritten", rcFile)
		}
		// Otherwise we don't need to write the file because it already has
		// what we need.
		return nil
	}

	if err := ioutil.WriteFile(rcFile, []byte(config), 0600); err != nil {
		return errors.Wrapf(err, "writing generated %s file with TFE token to %s", rcFilename, rcFile)
	}
	return nil
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
		err = fmt.Errorf("%s: running %q in %q", err, commandStr, path)
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

// rcFileContents is a format string to be used with Sprintf that can be used
// to generate the contents of a ~/.terraformrc file for authenticating with
// Terraform Enterprise.
var rcFileContents = `credentials "app.terraform.io" {
  token = %q
}`
