package models

import (
	"bufio"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
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

		ctx.Log.Debug("starting '%s %q' in '%s'", s.shell.String(), s.command, s.workingDir)
		err := s.cmd.Start()
		if err != nil {
			err = errors.Wrapf(err, "running '%s %q' in '%s'", s.shell.String(), s.command, s.workingDir)
			ctx.Log.Err(err.Error())
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
					ctx.Log.Err(errors.Wrapf(err, "writing %q to process", line).Error())
				}
			}
		}()

		wg := new(sync.WaitGroup)
		wg.Add(2)
		// Asynchronously copy from stdout/err to outCh.
		go func() {
			scanner := bufio.NewScanner(stdout)
			buf := []byte{}
			scanner.Buffer(buf, BufioScannerBufferSize)

			for scanner.Scan() {
				message := scanner.Text()
				outCh <- Line{Line: message}
				if s.streamOutput {
					s.outputHandler.Send(ctx, message, false)
				}
			}
			wg.Done()
		}()
		go func() {
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				message := scanner.Text()
				outCh <- Line{Line: message}
				if s.streamOutput {
					s.outputHandler.Send(ctx, message, false)
				}
			}
			wg.Done()
		}()

		// Wait for our copying to complete. This *must* be done before
		// calling cmd.Wait(). (see https://github.com/golang/go/issues/19685)
		wg.Wait()

		// Wait for the command to complete.
		err = s.cmd.Wait()

		dur := time.Since(start)
		log := ctx.Log.With("duration", dur)

		// We're done now. Send an error if there was one.
		if err != nil {
			err = errors.Wrapf(err, "running '%s' '%s' in '%s'",
				s.shell.String(), s.command, s.workingDir)
			log.Err(err.Error())
			outCh <- Line{Err: err}
		} else {
			log.Info("successfully ran '%s' '%s' in '%s'",
				s.shell.String(), s.command, s.workingDir)
		}
	}()

	return inCh, outCh
}
