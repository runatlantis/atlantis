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
	"bufio"
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/logging"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_terraform_client.go Client

type Client interface {
	Version() *version.Version
	RunCommandWithVersion(log *logging.SimpleLogger, path string, args []string, v *version.Version, workspace string) (string, error)
}

type DefaultClient struct {
	defaultVersion          *version.Version
	terraformPluginCacheDir string
	// tfExecutableName is the name of the default terraform binary.
	// This should always be set to "terraform" by the NewClient constructor
	// however it can be overridden during testing.
	tfExecutableName string
}

const terraformPluginCacheDirName = "plugin-cache"

// versionRegex extracts the version from `terraform version` output.
//     Terraform v0.12.0-alpha4 (2c36829d3265661d8edbd5014de8090ea7e2a076)
//	   => 0.12.0-alpha4
//
//     Terraform v0.11.13
//	   => 0.11.13
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
		tfExecutableName:        "terraform",
	}, nil
}

// Version returns the version of the terraform executable in our $PATH.
func (c *DefaultClient) Version() *version.Version {
	return c.defaultVersion
}

// RunCommandWithVersion executes the provided version of terraform with
// the provided args in path. v is the version of terraform executable to use.
// If v is nil, will use the default version.
// Workspace is the terraform workspace to run in. We won't switch workspaces,
// just set a WORKSPACE environment variable.
func (c *DefaultClient) RunCommandWithVersion(log *logging.SimpleLogger, path string, args []string, v *version.Version, workspace string) (string, error) {
	tfCmd, cmd := c.prepCmd(v, workspace, path, args)
	out, err := cmd.CombinedOutput()
	if err != nil {
		err = errors.Wrapf(err, "running %q in %q", tfCmd, path)
		log.Err(err.Error())
		return string(out), err
	}
	log.Info("successfully ran %q in %q", tfCmd, path)
	return string(out), nil
}

func (c *DefaultClient) prepCmd(v *version.Version, workspace string, path string, args []string) (string, *exec.Cmd) {
	tfExecutable := c.tfExecutableName
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
	cmd := exec.Command("sh", "-c", tfCmd)
	cmd.Dir = path
	cmd.Env = envVars
	return tfCmd, cmd
}

// Line represents a line that was output from a terraform command.
type Line struct {
	// Line is the contents of the line (without the newline).
	Line string
	// Err is set if there was an error.
	Err error
}

// RunCommandAsync runs terraform with args. It immediately returns an
// input and output channel. Callers can use the output channel to
// get the realtime output from the command.
// Callers can use the input channel to pass stdin input to the command.
// If any error is passed on the out channel, there will be no
// further output (so callers are free to exit).
func (c *DefaultClient) RunCommandAsync(log *logging.SimpleLogger, path string, args []string, v *version.Version, workspace string) (chan<- string, <-chan Line) {
	outCh := make(chan Line)
	inCh := make(chan string)

	// We start a goroutine to do our work asynchronously and then immediately
	// return our channels.
	go func() {

		// Ensure we close our channels when we exit.
		defer func() {
			close(outCh)
			close(inCh)
		}()

		tfCmd, cmd := c.prepCmd(v, workspace, path, args)
		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()
		stdin, _ := cmd.StdinPipe()

		log.Debug("starting %q in %q", tfCmd, path)
		err := cmd.Start()
		if err != nil {
			err = errors.Wrapf(err, "running %q in %q", tfCmd, path)
			log.Err(err.Error())
			outCh <- Line{Err: err}
			return
		}

		// If we get anything on inCh, write it to stdin.
		// This function will exit when inCh is closed which we do in our defer.
		go func() {
			for line := range inCh {
				log.Debug("writing %q to remote command's stdin", line)
				_, err := io.WriteString(stdin, line)
				if err != nil {
					log.Err(errors.Wrapf(err, "writing %q to process", line).Error())
				}
			}
		}()

		// Use a waitgroup to block until our stdout/err copying is complete.
		wg := new(sync.WaitGroup)
		wg.Add(2)

		// Asynchronously copy from stdout/err to outCh.
		go func() {
			s := bufio.NewScanner(stdout)
			for s.Scan() {
				outCh <- Line{Line: s.Text()}
			}
			wg.Done()
		}()
		go func() {
			s := bufio.NewScanner(stderr)
			for s.Scan() {
				outCh <- Line{Line: s.Text()}
			}
			wg.Done()
		}()

		// Wait for our copying to complete. This *must* be done before
		// calling cmd.Wait(). (see https://github.com/golang/go/issues/19685)
		wg.Wait()

		// Wait for the command to complete.
		err = cmd.Wait()

		// We're done now. Send an error if there was one.
		if err != nil {
			err = errors.Wrapf(err, "running %q in %q", tfCmd, path)
			log.Err(err.Error())
			outCh <- Line{Err: err}
		} else {
			log.Info("successfully ran %q in %q", tfCmd, path)
		}
	}()

	return inCh, outCh
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

// rcFileContents is a format string to be used with Sprintf that can be used
// to generate the contents of a ~/.terraformrc file for authenticating with
// Terraform Enterprise.
var rcFileContents = `credentials "app.terraform.io" {
  token = %q
}`
