package models_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/core/runtime/models"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/jobs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
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
			ctx := command.ProjectContext{
				Log:        logging.NewNoopLogger(t),
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
			runner := models.NewShellCommandRunner(c.Command, environ, cwd, projectCmdOutputHandler)
			output, err := runner.Run(ctx)
			Ok(t, err)
			Equals(t, fmt.Sprintf("%s\n", strings.Join(c.ExpLines, "\n")), output)
			for _, line := range c.ExpLines {
				projectCmdOutputHandler.VerifyWasCalledOnce().Send(ctx, line, false)
			}
		})
	}
}
