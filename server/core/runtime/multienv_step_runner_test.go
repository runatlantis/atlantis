package runtime_test

import (
	"testing"

	version "github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/runtime"
	"github.com/runatlantis/atlantis/server/core/terraform/mocks"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	jobmocks "github.com/runatlantis/atlantis/server/jobs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestMultiEnvStepRunner_Run(t *testing.T) {
	cases := []struct {
		Command     string
		ProjectName string
		ExpOut      string
		ExpErr      string
		Version     string
	}{
		{
			Command: `echo 'TF_VAR_REPODEFINEDVARIABLE_ONE=value1'`,
			ExpOut:  "Dynamic environment variables added:\nTF_VAR_REPODEFINEDVARIABLE_ONE\n",
			Version: "v1.2.3",
		},
		{
			Command: `echo 'TF_VAR_REPODEFINEDVARIABLE_TWO=value=1='`,
			ExpOut:  "Dynamic environment variables added:\nTF_VAR_REPODEFINEDVARIABLE_TWO\n",
			Version: "v1.2.3",
		},
		{
			Command: `echo 'TF_VAR_REPODEFINEDVARIABLE_NO_VALUE'`,
			ExpErr:  "Invalid environment variable definition: TF_VAR_REPODEFINEDVARIABLE_NO_VALUE",
			Version: "v1.2.3",
		},
	}
	RegisterMockTestingT(t)
	tfClient := mocks.NewMockClient()
	tfVersion, err := version.NewVersion("0.12.0")
	Ok(t, err)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	runStepRunner := runtime.RunStepRunner{
		TerraformExecutor:       tfClient,
		DefaultTFVersion:        tfVersion,
		ProjectCmdOutputHandler: projectCmdOutputHandler,
	}
	multiEnvStepRunner := runtime.MultiEnvStepRunner{
		RunStepRunner: &runStepRunner,
	}
	for _, c := range cases {
		t.Run(c.Command, func(t *testing.T) {
			tmpDir := t.TempDir()
			ctx := command.ProjectContext{
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
					BaseBranch: "main",
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
			envMap := make(map[string]string)
			value, err := multiEnvStepRunner.Run(ctx, c.Command, tmpDir, envMap)
			if c.ExpErr != "" {
				ErrContains(t, c.ExpErr, err)
				return
			}
			Ok(t, err)
			Equals(t, c.ExpOut, value)
		})
	}
}
