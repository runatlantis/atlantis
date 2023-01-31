package terraform

import (
	"context"
	"fmt"
	"io"
	"os/exec"

	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/runtime/cache"
	"github.com/runatlantis/atlantis/server/core/terraform"
)

type ClientConfig struct {
	BinDir        string
	CacheDir      string
	TfDownloadURL string
}

// Line represents a line that was output from a terraform command.
type Line struct {
	// Line is the contents of the line (without the newline).
	Line string
}

func NewAsyncClient(
	cfg ClientConfig,
	defaultVersion string,
	tfDownloader terraform.Downloader,
	versionCache cache.ExecutionVersionCache,
) (*AsyncClient, error) {
	version, err := getDefaultVersion(defaultVersion)
	if err != nil {
		return nil, errors.Wrapf(err, "getting default version")
	}

	// warm the cache with this version
	_, err = versionCache.Get(version)
	if err != nil {
		return nil, errors.Wrapf(err, "getting default terraform version %s", defaultVersion)
	}

	builder := &commandBuilder{
		defaultVersion:          version,
		versionCache:            versionCache,
		terraformPluginCacheDir: cfg.CacheDir,
	}

	return &AsyncClient{
		CommandBuilder: builder,
	}, nil

}

type cmdBuilder interface {
	Build(ctx context.Context, v *version.Version, path string, subcommand *SubCommand) (*exec.Cmd, error)
}

type AsyncClient struct {
	CommandBuilder cmdBuilder
}

type RunOptions struct {
	StdOut io.Writer
	StdErr io.Writer
}

type RunCommandRequest struct {
	RootPath          string
	SubCommand        *SubCommand
	AdditionalEnvVars map[string]string
	Version           *version.Version
}

func (c *AsyncClient) RunCommand(ctx context.Context, request *RunCommandRequest, options ...RunOptions) error {
	cmd, err := c.CommandBuilder.Build(ctx, request.Version, request.RootPath, request.SubCommand)
	if err != nil {
		return errors.Wrapf(err, "building command")
	}

	for _, option := range options {
		if option.StdOut != nil {
			cmd.Stdout = option.StdOut
		}

		if option.StdErr != nil {
			cmd.Stderr = option.StdErr
		}
	}

	for key, val := range request.AdditionalEnvVars {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, val))
	}

	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "running terraform command")
	}

	return nil
}

func getDefaultVersion(overrideVersion string) (*version.Version, error) {
	v, err := version.NewVersion(overrideVersion)
	if err != nil {
		return nil, errors.Wrapf(err, "parsing version %s", overrideVersion)
	}

	return v, nil
}
