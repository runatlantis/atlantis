// Package terraform handles the actual running of terraform commands
package terraform

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"

	"strings"

	version "github.com/hashicorp/go-version"
	"github.com/hootsuite/atlantis/logging"
	"github.com/pkg/errors"
)

type Client struct {
	defaultVersion *version.Version
}

var versionRegex = regexp.MustCompile("Terraform v(.*)\n")

func NewClient() (*Client, error) {
	versionCmdOutput, err := exec.Command("terraform", "version").CombinedOutput()
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
	version, err := version.NewVersion(match[1])
	if err != nil {
		return nil, errors.Wrap(err, "parsing terraform version")
	}

	return &Client{
		defaultVersion: version,
	}, nil
}

// Version returns the version of the terraform executable in our $PATH.
func (c *Client) Version() *version.Version {
	return c.defaultVersion
}

// RunCommandWithVersion executes the provided version of terraform with
// the provided args and envVars in path.
func (c *Client) RunCommandWithVersion(log *logging.SimpleLogger, path string, args []string, envVars []string, v *version.Version) (string, error) {
	tfExecutable := "terraform"
	// if version is the same as the default, don't need to prepend the version name to the executable
	if !v.Equal(c.defaultVersion) {
		tfExecutable = fmt.Sprintf("%s%s", tfExecutable, v.String())
	}
	terraformCmd := exec.Command(tfExecutable, args...)
	terraformCmd.Dir = path
	if len(envVars) > 0 {
		// append current process's environment variables
		// this is to prevent the $PATH variable being removed from the environment
		terraformCmd.Env = append(os.Environ(), envVars...)
	}
	out, err := terraformCmd.CombinedOutput()
	if err != nil {
		// show debug output for terraform commands if they fail, helpful for debugging
		log.Debug("error running %q in %q: \n%s", strings.Join(terraformCmd.Args, " "), path, out)
		return "", fmt.Errorf("%s: running %q in %q", err, strings.Join(terraformCmd.Args, " "), path)
	}
	log.Info("successfully ran %q in %q", strings.Join(terraformCmd.Args, " "), path)
	return string(out), nil
}

// RunInitAndEnv executes "terraform init" and "terraform env select" in path.
// env is the environment to select and extraInitArgs are additional arguments
// applied to the init command.
func (c *Client) RunInitAndEnv(log *logging.SimpleLogger, path string, env string, extraInitArgs []string, envVars []string, version *version.Version) ([]string, error) {
	var outputs []string
	// run terraform init
	output, err := c.RunCommandWithVersion(log, path, append([]string{"init", "-no-color"}, extraInitArgs...), envVars, version)
	if err != nil {
		return nil, err
	}
	outputs = append(outputs, output)

	// run terraform env new and select
	output, err = c.RunCommandWithVersion(log, path, []string{"env", "select", "-no-color", env}, envVars, version)
	if err != nil {
		// if terraform env select fails we will run terraform env new
		// to create a new environment
		output, err = c.RunCommandWithVersion(log, path, []string{"env", "new", "-no-color", env}, envVars, version)
		if err != nil {
			return nil, err
		}
	}
	return append(outputs, output), nil
}
