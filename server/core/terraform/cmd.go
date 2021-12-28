package terraform

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/runtime/cache"
)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_command_builder.go commandBuilder
type commandBuilder interface {
	Build(v *version.Version, workspace string, path string, args []string) (*exec.Cmd, error)
}

type CommandBuilder struct {
	defaultVersion          *version.Version
	versionCache            cache.ExecutionVersionCache
	terraformPluginCacheDir string
}

func (c *CommandBuilder) Build(v *version.Version, workspace string, path string, args []string) (*exec.Cmd, error) {
	if v == nil {
		v = c.defaultVersion
	}

	binPath, err := c.versionCache.Get(v)
	if err != nil {
		return nil, errors.Wrapf(err, "getting version from cache %s", v.String())
	}

	// We add custom variables so that if `extra_args` is specified with env
	// vars then they'll be substituted.
	envVars := []string{
		// Will de-emphasize specific commands to run in output.
		"TF_IN_AUTOMATION=true",
		// Cache plugins so terraform init runs faster.
		fmt.Sprintf("WORKSPACE=%s", workspace),
		fmt.Sprintf("TF_WORKSPACE=%s", workspace),
		fmt.Sprintf("ATLANTIS_TERRAFORM_VERSION=%s", v.String()),
		fmt.Sprintf("DIR=%s", path),
	}
	if c.terraformPluginCacheDir != "" {
		envVars = append(envVars, fmt.Sprintf("TF_PLUGIN_CACHE_DIR=%s", c.terraformPluginCacheDir))
	}
	// Append current Atlantis process's environment variables, ex.
	// AWS_ACCESS_KEY.
	envVars = append(envVars, os.Environ()...)
	tfCmd := fmt.Sprintf("%s %s", binPath, strings.Join(args, " "))
	cmd := exec.Command("sh", "-c", tfCmd)
	cmd.Dir = path
	cmd.Env = envVars
	return cmd, nil
}
