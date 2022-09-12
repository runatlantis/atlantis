package terraform

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"sync"

	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/runtime/cache"
	"github.com/runatlantis/atlantis/server/core/terraform"
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
	// Err is set if there was an error.
	Err error
}

// Setting the buffer size to 10mb
const bufioScannerBufferSize = 10 * 1024 * 1024

// versionRegex extracts the version from `terraform version` output.
//     Terraform v0.12.0-alpha4 (2c36829d3265661d8edbd5014de8090ea7e2a076)
//	   => 0.12.0-alpha4
//
//     Terraform v0.11.10
//	   => 0.11.10
var versionRegex = regexp.MustCompile("Terraform v(.*?)(\\s.*)?\n")

func NewAsyncClient(
	cfg ClientConfig,
	defaultVersion string,
	tfDownloader terraform.Downloader,
) (*AsyncClient, error) {
	version, err := getDefaultVersion(defaultVersion)
	if err != nil {
		return nil, errors.Wrapf(err, "getting default version")
	}

	loader := terraform.NewVersionLoader(tfDownloader, cfg.TfDownloadURL)

	versionCache := cache.NewExecutionVersionLayeredLoadingCache(
		"terraform",
		cfg.BinDir,
		loader.LoadVersion,
	)

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

type cmddBuilder interface {
	Build(v *version.Version, path string, args []string) (*exec.Cmd, error)
}

type AsyncClient struct {
	CommandBuilder cmddBuilder
}

func (c *AsyncClient) RunCommand(ctx context.Context, jobID string, path string, args []string, customEnvVars map[string]string, v *version.Version) <-chan Line {
	outCh := make(chan Line)

	// We start a goroutine to do our work asynchronously and then immediately
	// return our channels.
	go c.runCommand(ctx, jobID, path, args, customEnvVars, v, outCh)
	return outCh
}

func (c *AsyncClient) runCommand(ctx context.Context, jobID string, path string, args []string, customEnvVars map[string]string, v *version.Version, outCh chan Line) {
	// Ensure we close our channels when we exit.
	defer func() {
		close(outCh)
	}()

	cmd, err := c.CommandBuilder.Build(v, path, args)
	if err != nil {
		logger.Error(ctx, errors.Wrapf(err, "building command: %s", args).Error())
		outCh <- Line{Err: err}
		return
	}
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	envVars := cmd.Env
	for key, val := range customEnvVars {
		envVars = append(envVars, fmt.Sprintf("%s=%s", key, val))
	}
	cmd.Env = envVars

	err = cmd.Start()
	if err != nil {
		logger.Error(ctx, errors.Wrapf(err, "running %q in %q", cmd.String(), path).Error())
		outCh <- Line{Err: err}
		return
	}

	// Use a waitgroup to block until our stdout/err copying is complete.
	wg := new(sync.WaitGroup)
	wg.Add(2)
	// Asynchronously copy from stdout/err to outCh.
	go func() {
		defer wg.Done()
		c.WriteOutput(stdout, outCh, jobID)
	}()
	go func() {
		defer wg.Done()
		c.WriteOutput(stderr, outCh, jobID)
	}()

	// Wait for our copying to complete. This *must* be done before
	// calling cmd.Wait(). (see https://github.com/golang/go/issues/19685)
	wg.Wait()

	// Wait for the command to complete.
	err = cmd.Wait()

	// We're done now. Send an error if there was one.
	if err != nil {
		err = errors.Wrapf(err, "running %q in %q", cmd.String(), path)
		logger.Error(ctx, err.Error())
		outCh <- Line{Err: err}
	} else {
		logger.Info(ctx, fmt.Sprintf("successfully ran %q in %q", cmd.String(), path))
	}
}

func (c *AsyncClient) WriteOutput(stdReader io.ReadCloser, outCh chan Line, jobID string) {
	s := bufio.NewScanner(stdReader)
	buf := []byte{}
	s.Buffer(buf, bufioScannerBufferSize)

	for s.Scan() {
		message := s.Text()
		outCh <- Line{Line: message}
	}
}

func getDefaultVersion(overrideVersion string) (*version.Version, error) {
	v, err := version.NewVersion(overrideVersion)
	if err != nil {
		return nil, errors.Wrapf(err, "parsing version %s", overrideVersion)
	}

	return v, nil
}
