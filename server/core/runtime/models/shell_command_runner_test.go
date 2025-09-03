package models_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"
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
			ctrl := gomock.NewController(t)
	defer ctrl.Finish()
			log := logmocks.NewMockSimpleLogging()
			When(log.With(gomock.Any(), gomock.Any())).ThenReturn(log)
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
			runner := models.NewShellCommandRunner(nil, c.Command, environ, cwd, true, projectCmdOutputHandler)
			output, err := runner.Run(ctx)
			Ok(t, err)
			Equals(t, expectedOutput, output)
			for _, line := range c.ExpLines {
				// TODO: Convert to gomock expectation with argument capture
	// projectCmdOutputHandler.EXPECT().Send(ctx, line, false)
			}

			// TODO: Convert to gomock expectation with argument capture
	// log.EXPECT().With(Eq("duration"), gomock.Any())

			// And again with streaming disabled. Everything should be the same except the
			// command output handler should not have received anything

			projectCmdOutputHandler = mocks.NewMockProjectCommandOutputHandler()
			runner = models.NewShellCommandRunner(nil, c.Command, environ, cwd, false, projectCmdOutputHandler)
			output, err = runner.Run(ctx)
			Ok(t, err)
			Equals(t, expectedOutput, output)
			// TODO: Convert Never() expectation: projectCmdOutputHandler.EXPECT().Send(gomock.Any().Times(0), gomock.Any(), Eq(false))

			log.VerifyWasCalled(Twice()).With(Eq("duration"), gomock.Any())
		})
	}
}
