package runtime_test

import (
	"context"
	"strings"
	"testing"

	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/core/runtime"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestPreWorkflowHookRunner_Run(t *testing.T) {
	cases := []struct {
		Command string
		ExpOut  string
		ExpErr  string
	}{
		{
			Command: "echo hi",
			ExpOut:  "",
		},
		{
			Command: "echo 'a",
			ExpErr:  "exit status 2",
		},
	}

	for _, c := range cases {
		var err error

		Ok(t, err)

		RegisterMockTestingT(t)

		logger := logging.NewNoopLogger(t)

		r := runtime.DefaultPreWorkflowHookRunner{}
		t.Run(c.Command, func(t *testing.T) {
			tmpDir, cleanup := TempDir(t)
			defer cleanup()
			ctx := context.Background()
			prjCtx := models.PreWorkflowHookCommandContext{
				BaseRepo: models.Repo{
					Name:  "basename",
					Owner: "baseowner",
				},
				HeadRepo: models.Repo{
					Name:  "headname",
					Owner: "headowner",
				},
				Pull: models.PullRequest{
					Num:        2,
					HeadBranch: "add-feat",
					HeadCommit: "12345abcdef",
					BaseBranch: "master",
					Author:     "acme",
				},
				User: models.User{
					Username: "acme-user",
				},
				Log: logger,
			}
			out, err := r.Run(ctx, prjCtx, c.Command, tmpDir)
			if c.ExpErr != "" {
				ErrContains(t, c.ExpErr, err)
				return
			}
			Ok(t, err)
			// Replace $DIR in the exp with the actual temp dir. We do this
			// here because when constructing the cases we don't yet know the
			// temp dir.
			expOut := strings.Replace(c.ExpOut, "$DIR", tmpDir, -1)
			Equals(t, expOut, out)
		})
	}
}
