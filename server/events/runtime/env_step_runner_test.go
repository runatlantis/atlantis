package runtime_test

import (
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/runtime"
	"github.com/runatlantis/atlantis/server/events/terraform/mocks"
	"github.com/runatlantis/atlantis/server/logging"

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
			ExpValue: "123\n",
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
	terraform := mocks.NewMockClient()
	projVersion, err := version.NewVersion("v0.11.0")
	Ok(t, err)
	defaultVersion, _ := version.NewVersion("0.8")
	r := runtime.EnvStepRunner{
		TerraformExecutor: terraform,
		DefaultTFVersion:  defaultVersion,
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
				Log:              logging.NewNoopLogger(),
				Workspace:        "myworkspace",
				RepoRelDir:       "mydir",
				TerraformVersion: projVersion,
				ProjectName:      c.ProjectName,
			}
			_, value, err := r.Run(ctx, "var", c.Command, c.Value, tmpDir, map[string]string(nil))
			if c.ExpErr != "" {
				ErrContains(t, c.ExpErr, err)
				return
			}
			Ok(t, err)
			Equals(t, c.ExpValue, value)
		})
	}
}
