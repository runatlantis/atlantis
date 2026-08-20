// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/core/terraform/ansi"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/jobs"
)

// Setting the buffer size to 10mb
const BufioScannerBufferSize = 10 * 1024 * 1024

// Line represents a line that was output from a shell command.
type Line struct {
	// Line is the contents of the line (without the newline).
	Line string
	// Err is set if there was an error.
	Err error
}

// ShellCommandRunner runs a command via `exec.Command` and streams output to the
// `ProjectCommandOutputHandler`.
type ShellCommandRunner struct {
	command       string
	workingDir    string
	outputHandler jobs.ProjectCommandOutputHandler
	streamOutput  bool
	cmd           *exec.Cmd
	shell         *valid.CommandShell
}

func NewShellCommandRunner(
	shell *valid.CommandShell,
	command string,
	environ []string,
	workingDir string,
	streamOutput bool,
	outputHandler jobs.ProjectCommandOutputHandler,
) *ShellCommandRunner {
	if shell == nil {
		shell = &valid.CommandShell{
			Shell:     "sh",
			ShellArgs: []string{"-c"},
		}
	}
	var args []string
	args = append(args, shell.ShellArgs...)
	args = append(args, command)
	cmd := exec.Command(shell.Shell, args...) // #nosec
	cmd.Env = environ
	cmd.Dir = workingDir

	return &ShellCommandRunner{
		command:       command,
		workingDir:    workingDir,
		outputHandler: outputHandler,
		streamOutput:  streamOutput,
		cmd:           cmd,
		shell:         shell,
	}
}

// NewArgvCommandRunner runs a binary directly, passing argv[1:] to it as an
// argument vector. Nothing is interpreted by a shell, so a metacharacter in an
// argument reaches the process as literal text. Use this for commands Atlantis
// builds itself; NewShellCommandRunner is for `run` steps, where the user
// deliberately supplies shell source.
func NewArgvCommandRunner(
	argv []string,
	display string,
	environ []string,
	workingDir string,
	streamOutput bool,
	outputHandler jobs.ProjectCommandOutputHandler,
) *ShellCommandRunner {
	cmd := exec.Command(argv[0], argv[1:]...) // #nosec G204 -- argv[0] is a resolved executable path, and the arguments are passed as a vector rather than as shell source
	cmd.Env = environ
	cmd.Dir = workingDir

	return &ShellCommandRunner{
		command:       display,
		workingDir:    workingDir,
		outputHandler: outputHandler,
		streamOutput:  streamOutput,
		cmd:           cmd,
	}
}

// describe renders the command for messages that quote it as a single unit. A
// runner built by NewArgvCommandRunner has no shell, so there is nothing to
// name in front of the command. The shell-backed wording is unchanged: these
// strings are shown to users on pull requests when a `run` step fails.
func (s *ShellCommandRunner) describe() string {
	if s.shell == nil {
		return s.command
	}
	return fmt.Sprintf("%s %q", s.shell.String(), s.command)
}

// describeSplit renders the command for messages that quote the shell and the
// command as two separate units.
func (s *ShellCommandRunner) describeSplit() string {
	if s.shell == nil {
		return fmt.Sprintf("'%s'", s.command)
	}
	return fmt.Sprintf("'%s' '%s'", s.shell.String(), s.command)
}

func (s *ShellCommandRunner) Run(ctx command.ProjectContext) (string, error) {
	_, outCh := s.RunCommandAsync(ctx)

	outbuf := new(strings.Builder)
	var err error
	for line := range outCh {
		if line.Err != nil {
			err = line.Err
			break
		}
		outbuf.WriteString(line.Line)
		outbuf.WriteString("\n")
	}

	// sanitize output by stripping out any ansi characters.
	output := ansi.Strip(outbuf.String())
	return output, err
}

// RunCommandAsync runs terraform with args. It immediately returns an
// input and output channel. Callers can use the output channel to
// get the realtime output from the command.
// Callers can use the input channel to pass stdin input to the command.
// If any error is passed on the out channel, there will be no
// further output (so callers are free to exit).
func (s *ShellCommandRunner) RunCommandAsync(ctx command.ProjectContext) (chan<- string, <-chan Line) {
	outCh := make(chan Line)
	inCh := make(chan string)
	start := time.Now()

	// We start a goroutine to do our work asynchronously and then immediately
	// return our channels.
	go func() {
		// Ensure we close our channels when we exit.
		defer func() {
			close(outCh)
			close(inCh)
		}()

		stdout, _ := s.cmd.StdoutPipe()
		stderr, _ := s.cmd.StderrPipe()
		stdin, _ := s.cmd.StdinPipe()

		ctx.Log.Debug("starting '%s' in '%s'", s.describe(), s.workingDir)
		err := s.cmd.Start()
		if err != nil {
			err = fmt.Errorf("running '%s' in '%s': %w", s.describe(), s.workingDir, err)
			ctx.Log.Err("%s", err.Error())
			outCh <- Line{Err: err}
			return
		}

		// If we get anything on inCh, write it to stdin.
		// This function will exit when inCh is closed which we do in our defer.
		go func() {
			for line := range inCh {
				ctx.Log.Debug("writing %q to remote command's stdin", line)
				_, err := io.WriteString(stdin, line)
				if err != nil {
					err = fmt.Errorf("writing %q to process: %w", line, err)
					ctx.Log.Err("%s", err.Error())
				}
			}
		}()

		wg := new(sync.WaitGroup)
		wg.Add(2)
		// Asynchronously copy from stdout/err to outCh. Both scanners need
		// the enlarged buffer: with the default 64KiB token size limit,
		// bufio.Scanner stops scanning at the first longer line, silently
		// dropping it and all subsequent output. Terraform writes its
		// diagnostics to stderr, and some of them (e.g. dependency cycle
		// errors) can easily exceed 64KiB on a single line.
		readOutput := func(r io.Reader, name string) {
			defer wg.Done()
			scanner := bufio.NewScanner(r)
			buf := []byte{}
			scanner.Buffer(buf, BufioScannerBufferSize)

			for scanner.Scan() {
				message := scanner.Text()
				outCh <- Line{Line: message}
				if s.streamOutput {
					s.outputHandler.Send(ctx, message, false)
				}
			}
			if err := scanner.Err(); err != nil {
				// Don't fail the command over unreadable output, but
				// surface what happened instead of dropping it silently.
				ctx.Log.Err("reading %s of '%s': %v", name, s.command, err)
				message := fmt.Sprintf("[atlantis] error reading %s: %v", name, err)
				if errors.Is(err, bufio.ErrTooLong) {
					message = fmt.Sprintf("[atlantis] %s truncated: %v", name, err)
				}
				outCh <- Line{Line: message}
				if s.streamOutput {
					s.outputHandler.Send(ctx, message, false)
				}
				if errors.Is(err, bufio.ErrTooLong) {
					// The reader is still usable after an oversized token;
					// drain it so the child process can't block writing to
					// a full pipe.
					io.Copy(io.Discard, r) //nolint:errcheck
				}
			}
		}
		go readOutput(stdout, "stdout")
		go readOutput(stderr, "stderr")

		// Wait for our copying to complete. This *must* be done before
		// calling cmd.Wait(). (see https://github.com/golang/go/issues/19685)
		wg.Wait()

		// Wait for the command to complete.
		err = s.cmd.Wait()

		dur := time.Since(start)
		log := ctx.Log.With("duration", dur)

		// We're done now. Send an error if there was one.
		if err != nil {
			err = fmt.Errorf("running %s in '%s': %w", s.describeSplit(), s.workingDir, err)
			log.Err("%s", err.Error())
			outCh <- Line{Err: err}
		} else {
			log.Info("successfully ran %s in '%s'",
				s.describeSplit(), s.workingDir)
		}
	}()

	return inCh, outCh
}
