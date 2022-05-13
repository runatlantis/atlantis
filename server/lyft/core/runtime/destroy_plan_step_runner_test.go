package runtime_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/core/runtime"
	"github.com/runatlantis/atlantis/server/core/terraform/mocks"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	lyftRuntime "github.com/runatlantis/atlantis/server/lyft/core/runtime"

	. "github.com/petergtz/pegomock"
	. "github.com/runatlantis/atlantis/testing"
)

func TestRun_DestroyPlan(t *testing.T) {
	RegisterMockTestingT(t)

	// Create the env/workspace.tfvars file.
	tmpDir, cleanup := TempDir(t)
	defer cleanup()
	err := os.MkdirAll(filepath.Join(tmpDir, "env"), 0700)
	Ok(t, err)

	cases := []struct {
		description string
		expPlanArgs []string
		tags        map[string]string
	}{
		{
			description: "uses destroy plan",
			expPlanArgs: []string{
				"plan",
				"-input=false",
				"-refresh",
				"-out",
				fmt.Sprintf("%q", filepath.Join(tmpDir, "workspace.tfplan")),
				"-var",
				"atlantis_user=\"username\"",
				"-var",
				"atlantis_repo=\"owner/repo\"",
				"-var",
				"atlantis_repo_name=\"repo\"",
				"-var",
				"atlantis_repo_owner=\"owner\"",
				"-var",
				"atlantis_pull_num=2",
				"extra",
				"args",
				"-destroy",
				"comment",
				"args",
			},
			tags: map[string]string{
				lyftRuntime.Deprecated: lyftRuntime.Destroy,
			},
		},
		{
			description: "no destroy plan",
			expPlanArgs: []string{
				"plan",
				"-input=false",
				"-refresh",
				"-out",
				fmt.Sprintf("%q", filepath.Join(tmpDir, "workspace.tfplan")),
				"-var",
				"atlantis_user=\"username\"",
				"-var",
				"atlantis_repo=\"owner/repo\"",
				"-var",
				"atlantis_repo_name=\"repo\"",
				"-var",
				"atlantis_repo_owner=\"owner\"",
				"-var",
				"atlantis_pull_num=2",
				"extra",
				"args",
				"comment",
				"args",
			},
			tags: map[string]string{},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			// Using version >= 0.10 here so we don't expect any env commands.
			terraform := mocks.NewMockClient()
			tfVersion, _ := version.NewVersion("0.10.0")
			logger := logging.NewNoopCtxLogger(t)
			planStepRunner := runtime.PlanStepRunner{
				TerraformExecutor: terraform,
				DefaultTFVersion:  tfVersion,
			}
			stepRunner := lyftRuntime.DestroyPlanStepRunner{
				StepRunner: &planStepRunner,
			}
			ctx := context.Background()
			prjCtx := command.ProjectContext{
				Log:                logger,
				Workspace:          "workspace",
				RepoRelDir:         ".",
				User:               models.User{Username: "username"},
				EscapedCommentArgs: []string{"comment", "args"},
				Pull: models.PullRequest{
					Num: 2,
				},
				BaseRepo: models.Repo{
					FullName: "owner/repo",
					Owner:    "owner",
					Name:     "repo",
				},
				Tags: c.tags,
			}
			When(terraform.RunCommandWithVersion(ctx, prjCtx, tmpDir, c.expPlanArgs, map[string]string(nil), tfVersion, "workspace")).ThenReturn("output", nil)

			output, err := stepRunner.Run(ctx, prjCtx, []string{"extra", "args"}, tmpDir, map[string]string(nil))
			Ok(t, err)

			// Verify that we next called for the actual
			terraform.VerifyWasCalledOnce().RunCommandWithVersion(ctx, prjCtx, tmpDir, c.expPlanArgs, map[string]string(nil), tfVersion, "workspace")
			Equals(t, "output", output)
		})
	}
}
