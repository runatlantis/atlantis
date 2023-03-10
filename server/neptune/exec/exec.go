package exec

import (
	"context"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/logging"
	key "github.com/runatlantis/atlantis/server/neptune/context"
	"os"
	"os/exec"
	"syscall"
	"time"
)

type Cmd struct {
	*exec.Cmd
	logger logging.Logger
}

func Command(logger logging.Logger, name string, arg ...string) *Cmd {
	return &Cmd{
		Cmd:    exec.Command(name, arg...), // #nosec
		logger: logger,
	}
}

// RunWithNewProcessGroup is useful for running separate commands that shouldn't receive termination
// signals at the same time as the parent process. The passed context should listen for a timeout/cancellation
// that allows for the parent process to perform a shutdown on its own terms.
func (c *Cmd) RunWithNewProcessGroup(ctx context.Context) error {
	c.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	if err := c.Start(); err != nil {
		return errors.Wrap(err, "starting process command")
	}

	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			// received context cancellation, terminate active process
			c.terminateProcessOnCtxCancellation(ctx, c.Process, done)
		case <-done:
			// process completed on its own, simply exit
		}
	}()

	err := c.Wait()
	if ctx.Err() != nil {
		return errors.Wrap(ctx.Err(), "waiting for process")
	}
	if err != nil {
		return errors.Wrap(err, "waiting for process")
	}
	return nil
}

func (c *Cmd) terminateProcessOnCtxCancellation(ctx context.Context, p *os.Process, processDone chan struct{}) {
	c.logger.WarnContext(ctx, "Terminating active process gracefully")
	err := p.Signal(syscall.SIGTERM)
	if err != nil {
		c.logger.ErrorContext(context.WithValue(ctx, key.ErrKey, err), "Unable to terminate process")
	}

	// if we still haven't shutdown after 60 seconds, we should just kill the process
	// this ensures that we at least can gracefully shutdown other parts of the system
	// before we are killed entirely.
	kill := time.After(60 * time.Second)
	select {
	case <-kill:
		c.logger.WarnContext(ctx, "Killing process since graceful shutdown is taking suspiciously long. State corruption may have occurred.")
		err := p.Signal(syscall.SIGKILL)
		if err != nil {
			c.logger.ErrorContext(context.WithValue(ctx, key.ErrKey, err), "Unable to kill process")
		}
	case <-processDone:
	}
}
