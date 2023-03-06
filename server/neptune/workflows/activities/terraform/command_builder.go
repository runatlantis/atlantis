package terraform

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/runtime/cache"
)

type commandBuilder struct {
	defaultVersion          *version.Version
	versionCache            cache.ExecutionVersionCache
	terraformPluginCacheDir string
}

func (c *commandBuilder) Build(_ context.Context, v *version.Version, path string, subCommand *SubCommand) (*exec.Cmd, error) {
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
		fmt.Sprintf("ATLANTIS_TERRAFORM_VERSION=%s", v.String()),
		fmt.Sprintf("DIR=%s", path),
		fmt.Sprintf("TF_PLUGIN_CACHE_DIR=%s", c.terraformPluginCacheDir),
	}
	// Append current Atlantis process's environment variables, ex.
	// AWS_ACCESS_KEY.
	envVars = append(envVars, os.Environ()...)
	tfCmd := fmt.Sprintf("%s %s", binPath, strings.Join(subCommand.Build(), " "))

	cmd := exec.Command("sh", "-c", tfCmd)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	cmd.Dir = path
	cmd.Env = envVars
	return cmd, nil
}
