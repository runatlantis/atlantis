package runtime_test

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/events/mocks/matchers"
	"github.com/runatlantis/atlantis/server/events/models"
	matchers2 "github.com/runatlantis/atlantis/server/events/run/mocks/matchers"
	"github.com/runatlantis/atlantis/server/events/runtime"
	"github.com/runatlantis/atlantis/server/events/terraform/mocks"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
	. "github.com/runatlantis/atlantis/testing"
)

func TestRun_NoDir(t *testing.T) {
	o := runtime.ApplyStepRunner{
		TerraformExecutor: nil,
	}
	_, err := o.Run(models.ProjectCommandContext{
		RepoRelPath: ".",
		Workspace:   "workspace",
	}, nil, "/nonexistent/path")
	ErrEquals(t, "no plan found at path \".\" and workspace \"workspace\"–did you run plan?", err)
}

func TestRun_NoPlanFile(t *testing.T) {
	tmpDir, cleanup := TempDir(t)
	defer cleanup()
	o := runtime.ApplyStepRunner{
		TerraformExecutor: nil,
	}
	_, err := o.Run(models.ProjectCommandContext{
		RepoRelPath: ".",
		Workspace:   "workspace",
	}, nil, tmpDir)
	ErrEquals(t, "no plan found at path \".\" and workspace \"workspace\"–did you run plan?", err)
}

func TestRun_Success(t *testing.T) {
	tmpDir, cleanup := TempDir(t)
	defer cleanup()
	planPath := filepath.Join(tmpDir, "workspace.tfplan")
	err := ioutil.WriteFile(planPath, nil, 0644)
	Ok(t, err)

	RegisterMockTestingT(t)
	terraform := mocks.NewMockClient()
	o := runtime.ApplyStepRunner{
		TerraformExecutor: terraform,
	}

	When(terraform.RunCommandWithVersion(matchers.AnyPtrToLoggingSimpleLogger(), AnyString(), AnyStringSlice(), matchers2.AnyPtrToGoVersionVersion(), AnyString())).
		ThenReturn("output", nil)
	output, err := o.Run(models.ProjectCommandContext{
		Workspace:   "workspace",
		RepoRelPath: ".",
		CommentArgs: []string{"comment", "args"},
	}, []string{"extra", "args"}, tmpDir)
	Ok(t, err)
	Equals(t, "output", output)
	terraform.VerifyWasCalledOnce().RunCommandWithVersion(nil, tmpDir, []string{"apply", "-no-color", "extra", "args", "comment", "args", planPath}, nil, "workspace")
}

func TestRun_AppliesCorrectProjectPlan(t *testing.T) {
	// When running for a project, the planfile has a different name.
	tmpDir, cleanup := TempDir(t)
	defer cleanup()
	planPath := filepath.Join(tmpDir, "projectname-default.tfplan")
	err := ioutil.WriteFile(planPath, nil, 0644)
	Ok(t, err)

	RegisterMockTestingT(t)
	terraform := mocks.NewMockClient()
	o := runtime.ApplyStepRunner{
		TerraformExecutor: terraform,
	}

	When(terraform.RunCommandWithVersion(matchers.AnyPtrToLoggingSimpleLogger(), AnyString(), AnyStringSlice(), matchers2.AnyPtrToGoVersionVersion(), AnyString())).
		ThenReturn("output", nil)
	projectName := "projectname"
	output, err := o.Run(models.ProjectCommandContext{
		Workspace:   "default",
		RepoRelPath: ".",
		ProjectConfig: &valid.Project{
			Name: &projectName,
		},
		CommentArgs: []string{"comment", "args"},
	}, []string{"extra", "args"}, tmpDir)
	Ok(t, err)
	Equals(t, "output", output)
	terraform.VerifyWasCalledOnce().RunCommandWithVersion(nil, tmpDir, []string{"apply", "-no-color", "extra", "args", "comment", "args", planPath}, nil, "default")
}

func TestRun_UsesConfiguredTFVersion(t *testing.T) {
	tmpDir, cleanup := TempDir(t)
	defer cleanup()
	planPath := filepath.Join(tmpDir, "workspace.tfplan")
	err := ioutil.WriteFile(planPath, nil, 0644)
	Ok(t, err)

	RegisterMockTestingT(t)
	terraform := mocks.NewMockClient()
	o := runtime.ApplyStepRunner{
		TerraformExecutor: terraform,
	}
	tfVersion, _ := version.NewVersion("0.11.0")

	When(terraform.RunCommandWithVersion(matchers.AnyPtrToLoggingSimpleLogger(), AnyString(), AnyStringSlice(), matchers2.AnyPtrToGoVersionVersion(), AnyString())).
		ThenReturn("output", nil)
	output, err := o.Run(models.ProjectCommandContext{
		Workspace:   "workspace",
		RepoRelPath: ".",
		CommentArgs: []string{"comment", "args"},
		ProjectConfig: &valid.Project{
			TerraformVersion: tfVersion,
		},
	}, []string{"extra", "args"}, tmpDir)
	Ok(t, err)
	Equals(t, "output", output)
	terraform.VerifyWasCalledOnce().RunCommandWithVersion(nil, tmpDir, []string{"apply", "-no-color", "extra", "args", "comment", "args", planPath}, tfVersion, "workspace")
}
