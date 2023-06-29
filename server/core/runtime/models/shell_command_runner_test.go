package models_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/runtime/models"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/jobs/mocks"
	logmocks "github.com/runatlantis/atlantis/server/logging/mocks"
	. "github.com/runatlantis/atlantis/testing"
)

func TestShellCommandRunner_Run(t *testing.T) {
	cases := []struct {
		Command  string
		ExpLines []string
		Environ  map[string]string
	}{
		{
			Command: "echo $HELLO",
			Environ: map[string]string{
				"HELLO": "world",
			},
			ExpLines: []string{"world"},
		},
		{
			Command:  ">&2 echo this is an error",
			ExpLines: []string{"this is an error"},
		},
	}

	for _, c := range cases {
		t.Run(c.Command, func(t *testing.T) {
			RegisterMockTestingT(t)
			log := logmocks.NewMockSimpleLogging()
			When(log.With(Any[string](), Any[interface{}]())).ThenReturn(log)
			ctx := command.ProjectContext{
				Log:        log,
				Workspace:  "default",
				RepoRelDir: ".",
			}
			projectCmdOutputHandler := mocks.NewMockProjectCommandOutputHandler()

			cwd, err := os.Getwd()
			Ok(t, err)
			environ := []string{}
			for k, v := range c.Environ {
				environ = append(environ, fmt.Sprintf("%s=%s", k, v))
			}
			expectedOutput := fmt.Sprintf("%s\n", strings.Join(c.ExpLines, "\n"))

			// Run once with streaming enabled
			runner := models.NewShellCommandRunner(c.Command, environ, cwd, true, projectCmdOutputHandler)
			output, err := runner.Run(ctx)
			Ok(t, err)
			Equals(t, expectedOutput, output)
			for _, line := range c.ExpLines {
				projectCmdOutputHandler.VerifyWasCalledOnce().Send(ctx, line, false)
			}

			log.VerifyWasCalledOnce().With(Eq("duration"), Any[interface{}]())

			// And again with streaming disabled. Everything should be the same except the
			// command output handler should not have received anything

			projectCmdOutputHandler = mocks.NewMockProjectCommandOutputHandler()
			runner = models.NewShellCommandRunner(c.Command, environ, cwd, false, projectCmdOutputHandler)
			output, err = runner.Run(ctx)
			Ok(t, err)
			Equals(t, expectedOutput, output)
			projectCmdOutputHandler.VerifyWasCalled(Never()).Send(Any[command.ProjectContext](), Any[string](), Eq(false))

			log.VerifyWasCalled(Twice()).With(Eq("duration"), Any[interface{}]())
		})
	}
}
