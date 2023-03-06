package terraform

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/runtime/cache"
	"github.com/runatlantis/atlantis/server/core/terraform"
	key "github.com/runatlantis/atlantis/server/neptune/context"
	"github.com/runatlantis/atlantis/server/neptune/logger"
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

	if err := cmd.Start(); err != nil {
		return errors.Wrap(err, "starting terraform command")
	}

	done := make(chan struct{})
	defer close(done)
	go terminateOnCtxCancellation(ctx, cmd.Process, done)

	err = cmd.Wait()

	if ctx.Err() != nil {
		return errors.Wrap(ctx.Err(), "waiting for terraform process")
	}

	if err != nil {
		return errors.Wrap(err, "waiting for terraform process")
	}

	return nil
}

func terminateOnCtxCancellation(ctx context.Context, p *os.Process, done chan struct{}) {
	select {
	case <-ctx.Done():
		logger.Warn(ctx, "Terminating terraform process gracefully")
		err := p.Signal(syscall.SIGTERM)
		if err != nil {
			logger.Error(ctx, "Unable to terminate process", key.ErrKey, err)
		}

		// if we still haven't shutdown after 60 seconds, we should just kill the process
		// this ensures that we at least can gracefully shutdown other parts of the system
		// before we are killed entirely.
		kill := time.After(60 * time.Second)

		select {
		case <-kill:
			logger.Warn(ctx, "Killing terraform process since graceful shutdown is taking suspiciously long. State corruption may have occurred.")
			err := p.Signal(syscall.SIGKILL)
			if err != nil {
				logger.Error(ctx, "Unable to kill process", key.ErrKey, err)
			}
		case <-done:
		}
	case <-done:
	}

}

func getDefaultVersion(overrideVersion string) (*version.Version, error) {
	v, err := version.NewVersion(overrideVersion)
	if err != nil {
		return nil, errors.Wrapf(err, "parsing version %s", overrideVersion)
	}

	return v, nil
}
