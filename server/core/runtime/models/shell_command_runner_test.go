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
			log := logmocks.NewMockSimpleLogging(ctrl)
			log.EXPECT().With(gomock.Any(), gomock.Any()).Return(log).AnyTimes()
			log.EXPECT().Debug(gomock.Any()).AnyTimes()
			ctx := command.ProjectContext{
				Log:        log,
				Workspace:  "default",
				RepoRelDir: ".",
			}
			projectCmdOutputHandler := mocks.NewMockProjectCommandOutputHandler(ctrl)

			cwd, err := os.Getwd()
			Ok(t, err)
			environ := []string{}
			for k, v := range c.Environ {
				environ = append(environ, fmt.Sprintf("%s=%s", k, v))
			}
			expectedOutput := fmt.Sprintf("%s\n", strings.Join(c.ExpLines, "\n"))

			// Set up expectations for streaming enabled
			for _, line := range c.ExpLines {
				projectCmdOutputHandler.EXPECT().Send(ctx, line, false).Times(1)
			}

			// Run once with streaming enabled
			runner := models.NewShellCommandRunner(nil, c.Command, environ, cwd, true, projectCmdOutputHandler)
			output, err := runner.Run(ctx)
			Ok(t, err)
			Equals(t, expectedOutput, output)

			// And again with streaming disabled. Everything should be the same except the
			// command output handler should not have received anything

			projectCmdOutputHandler2 := mocks.NewMockProjectCommandOutputHandler(ctrl)
			projectCmdOutputHandler2.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			runner = models.NewShellCommandRunner(nil, c.Command, environ, cwd, false, projectCmdOutputHandler2)
			output, err = runner.Run(ctx)
			Ok(t, err)
			Equals(t, expectedOutput, output)
		})
	}
}
