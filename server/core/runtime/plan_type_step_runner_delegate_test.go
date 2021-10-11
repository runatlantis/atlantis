package runtime

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock"
	. "github.com/runatlantis/atlantis/testing"

	"github.com/runatlantis/atlantis/server/core/runtime/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
)

var planFileContents = `
An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  - destroy

Terraform will perform the following actions:

  - null_resource.hi[1]


Plan: 0 to add, 0 to change, 1 to destroy.`

func TestRunDelegate(t *testing.T) {

	RegisterMockTestingT(t)

	mockDefaultRunner := mocks.NewMockRunner()
	mockRemoteRunner := mocks.NewMockRunner()

	subject := &PlanTypeStepRunnerDelegate{
		defaultRunner:    mockDefaultRunner,
		remotePlanRunner: mockRemoteRunner,
	}

	tfVersion, _ := version.NewVersion("0.12.0")

	t.Run("Remote Runner Success", func(t *testing.T) {
		tmpDir, cleanup := TempDir(t)
		defer cleanup()
		planPath := filepath.Join(tmpDir, "workspace.tfplan")
		err := os.WriteFile(planPath, []byte("Atlantis: this plan was created by remote ops\n"+planFileContents), 0600)
		Ok(t, err)

		ctx := models.ProjectCommandContext{
			Workspace:          "workspace",
			RepoRelDir:         ".",
			EscapedCommentArgs: []string{"comment", "args"},
			TerraformVersion:   tfVersion,
		}
		extraArgs := []string{"extra", "args"}
		envs := map[string]string{}

		expectedOut := "some random output"

		When(mockRemoteRunner.Run(ctx, extraArgs, tmpDir, envs)).ThenReturn(expectedOut, nil)

		output, err := subject.Run(ctx, extraArgs, tmpDir, envs)

		mockDefaultRunner.VerifyWasCalled(Never())

		Equals(t, expectedOut, output)
		Ok(t, err)

	})

	t.Run("Remote Runner Failure", func(t *testing.T) {
		tmpDir, cleanup := TempDir(t)
		defer cleanup()
		planPath := filepath.Join(tmpDir, "workspace.tfplan")
		err := os.WriteFile(planPath, []byte("Atlantis: this plan was created by remote ops\n"+planFileContents), 0600)
		Ok(t, err)

		ctx := models.ProjectCommandContext{
			Workspace:          "workspace",
			RepoRelDir:         ".",
			EscapedCommentArgs: []string{"comment", "args"},
			TerraformVersion:   tfVersion,
		}
		extraArgs := []string{"extra", "args"}
		envs := map[string]string{}

		expectedOut := "some random output"

		When(mockRemoteRunner.Run(ctx, extraArgs, tmpDir, envs)).ThenReturn(expectedOut, errors.New("err"))

		output, err := subject.Run(ctx, extraArgs, tmpDir, envs)

		mockDefaultRunner.VerifyWasCalled(Never())

		Equals(t, expectedOut, output)
		Assert(t, err != nil, "err should not be nil")

	})

	t.Run("Local Runner Success", func(t *testing.T) {
		tmpDir, cleanup := TempDir(t)
		defer cleanup()
		planPath := filepath.Join(tmpDir, "workspace.tfplan")
		err := os.WriteFile(planPath, []byte(planFileContents), 0600)
		Ok(t, err)

		ctx := models.ProjectCommandContext{
			Workspace:          "workspace",
			RepoRelDir:         ".",
			EscapedCommentArgs: []string{"comment", "args"},
			TerraformVersion:   tfVersion,
		}
		extraArgs := []string{"extra", "args"}
		envs := map[string]string{}

		expectedOut := "some random output"

		When(mockDefaultRunner.Run(ctx, extraArgs, tmpDir, envs)).ThenReturn(expectedOut, nil)

		output, err := subject.Run(ctx, extraArgs, tmpDir, envs)

		mockRemoteRunner.VerifyWasCalled(Never())

		Equals(t, expectedOut, output)
		Ok(t, err)

	})

	t.Run("Local Runner Failure", func(t *testing.T) {
		tmpDir, cleanup := TempDir(t)
		defer cleanup()
		planPath := filepath.Join(tmpDir, "workspace.tfplan")
		err := os.WriteFile(planPath, []byte(planFileContents), 0600)
		Ok(t, err)

		ctx := models.ProjectCommandContext{
			Workspace:          "workspace",
			RepoRelDir:         ".",
			EscapedCommentArgs: []string{"comment", "args"},
			TerraformVersion:   tfVersion,
		}
		extraArgs := []string{"extra", "args"}
		envs := map[string]string{}

		expectedOut := "some random output"

		When(mockDefaultRunner.Run(ctx, extraArgs, tmpDir, envs)).ThenReturn(expectedOut, errors.New("err"))

		output, err := subject.Run(ctx, extraArgs, tmpDir, envs)

		mockRemoteRunner.VerifyWasCalled(Never())

		Equals(t, expectedOut, output)
		Assert(t, err != nil, "err should not be nil")

	})

}
