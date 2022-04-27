package terraform

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/terraform/helpers"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/jobs"
)

// Setting the buffer size to 10mb
const BufioScannerBufferSize = 10 * 1024 * 1024

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_terraform_client_async.go ClientAsync

type ClientAsync interface {
	RunCommandAsync(ctx context.Context, prjCtx command.ProjectContext, path string, args []string, customEnvVars map[string]string, v *version.Version, workspace string) <-chan helpers.Line

	RunCommandAsyncWithInput(ctx context.Context, prjCtx command.ProjectContext, path string, args []string, customEnvVars map[string]string, v *version.Version, workspace string, input <-chan string) <-chan helpers.Line
}

type AsyncClient struct {
	projectCmdOutputHandler jobs.ProjectCommandOutputHandler
	commandBuilder          commandBuilder
}

// RunCommandAsync runs terraform with args. It immediately returns an
// input and output channel. Callers can use the output channel to
// get the realtime output from the command.
// Callers can use the input channel to pass stdin input to the command.
// If any error is passed on the out channel, there will be no
// further output (so callers are free to exit).
func (c *AsyncClient) RunCommandAsync(ctx context.Context, prjCtx command.ProjectContext, path string, args []string, customEnvVars map[string]string, v *version.Version, workspace string) <-chan helpers.Line {

	input := make(chan string)
	defer close(input)

	return c.RunCommandAsyncWithInput(ctx, prjCtx, path, args, customEnvVars, v, workspace, input)
}
func (c *AsyncClient) RunCommandAsyncWithInput(ctx context.Context, prjCtx command.ProjectContext, path string, args []string, customEnvVars map[string]string, v *version.Version, workspace string, input <-chan string) <-chan helpers.Line {
	outCh := make(chan helpers.Line)

	// We start a goroutine to do our work asynchronously and then immediately
	// return our channels.
	go func() {

		// Ensure we close our channels when we exit.
		defer func() {
			close(outCh)
		}()

		cmd, err := c.commandBuilder.Build(v, workspace, path, args)
		if err != nil {
			prjCtx.Log.Errorf(err.Error())
			outCh <- helpers.Line{Err: err}
			return
		}
		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()
		stdin, _ := cmd.StdinPipe()
		envVars := cmd.Env
		for key, val := range customEnvVars {
			envVars = append(envVars, fmt.Sprintf("%s=%s", key, val))
		}
		cmd.Env = envVars

		err = cmd.Start()
		if err != nil {
			err = errors.Wrapf(err, "running %q in %q", cmd.String(), path)
			prjCtx.Log.Errorf(err.Error())
			outCh <- helpers.Line{Err: err}
			return
		}

		// If we get anything on inCh, write it to stdin.
		// This function will exit when inCh is closed which we do in our defer.
		go func() {
			for line := range input {
				_, err := io.WriteString(stdin, line)
				if err != nil {
					prjCtx.Log.Errorf(errors.Wrapf(err, "writing %q to process", line).Error())
				}
			}
		}()

		// Use a waitgroup to block until our stdout/err copying is complete.
		wg := new(sync.WaitGroup)
		wg.Add(2)
		// Asynchronously copy from stdout/err to outCh.
		go func() {
			s := bufio.NewScanner(stdout)
			buf := []byte{}
			s.Buffer(buf, BufioScannerBufferSize)

			for s.Scan() {
				message := s.Text()
				outCh <- helpers.Line{Line: message}
				c.projectCmdOutputHandler.Send(prjCtx, message)
			}
			wg.Done()
		}()
		go func() {
			s := bufio.NewScanner(stderr)
			for s.Scan() {
				message := s.Text()
				outCh <- helpers.Line{Line: message}
				c.projectCmdOutputHandler.Send(prjCtx, message)
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
			err = errors.Wrapf(err, "running %q in %q", cmd.String(), path)
			prjCtx.Log.Errorf(err.Error())
			outCh <- helpers.Line{Err: err}
		} else {
			prjCtx.Log.Infof("successfully ran %q in %q", cmd.String(), path)
		}
	}()

	return outCh
}
