package runtime_test

import (
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/runtime"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestRunStepRunner_Run(t *testing.T) {
	cases := []struct {
		Command string
		ExpOut  string
		ExpErr  string
	}{
		{
			Command: "echo hi",
			ExpOut:  "hi\n",
		},
		{
			Command: "echo hi >> file && cat file",
			ExpOut:  "hi\n",
		},
		{
			Command: "lkjlkj",
			ExpErr:  "exit status 127: running \"lkjlkj\" in",
		},
	}

	r := runtime.RunStepRunner{}
	ctx := models.ProjectCommandContext{
		Log: logging.NewNoopLogger(),
	}
	for _, c := range cases {
		t.Run(c.Command, func(t *testing.T) {
			tmpDir, cleanup := TempDir(t)
			defer cleanup()
			out, err := r.Run(ctx, strings.Split(c.Command, " "), tmpDir)
			if c.ExpErr != "" {
				ErrContains(t, c.ExpErr, err)
				return
			}
			Ok(t, err)
			Equals(t, c.ExpOut, out)
		})
	}
}
