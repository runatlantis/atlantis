package runtime_test

import (
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/core/runtime"
	"github.com/runatlantis/atlantis/server/core/terraform/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"

	. "github.com/petergtz/pegomock"
	. "github.com/runatlantis/atlantis/testing"
)

func TestEnvStepRunner_Run(t *testing.T) {
	cases := []struct {
		Command     string
		Value       string
		ProjectName string
		ExpValue    string
		ExpErr      string
	}{
		{
			Command:  "echo 123",
			ExpValue: "123",
		},
		{
			Value:    "test",
			ExpValue: "test",
		},
		{
			Command:  "echo 321",
			Value:    "test",
			ExpValue: "test",
		},
	}
	RegisterMockTestingT(t)
	tfClient := mocks.NewMockClient()
	tfVersion, err := version.NewVersion("0.12.0")
	Ok(t, err)
	runStepRunner := runtime.RunStepRunner{
		TerraformExecutor: tfClient,
		DefaultTFVersion:  tfVersion,
	}
	envRunner := runtime.EnvStepRunner{
		RunStepRunner: &runStepRunner,
	}
	for _, c := range cases {
		t.Run(c.Command, func(t *testing.T) {
			tmpDir, cleanup := TempDir(t)
			defer cleanup()
			ctx := models.ProjectCommandContext{
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
					BaseBranch: "master",
					Author:     "acme",
				},
				User: models.User{
					Username: "acme-user",
				},
				Log:              logging.NewNoopLogger(t),
				Workspace:        "myworkspace",
				RepoRelDir:       "mydir",
				TerraformVersion: tfVersion,
				ProjectName:      c.ProjectName,
			}
			value, err := envRunner.Run(ctx, c.Command, c.Value, tmpDir, map[string]string(nil))
			if c.ExpErr != "" {
				ErrContains(t, c.ExpErr, err)
				return
			}
			Ok(t, err)
			Equals(t, c.ExpValue, value)
		})
	}
}
