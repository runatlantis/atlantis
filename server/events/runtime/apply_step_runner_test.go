package runtime_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/events/mocks/matchers"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/runtime"
	"github.com/runatlantis/atlantis/server/events/terraform/mocks"
	matchers2 "github.com/runatlantis/atlantis/server/events/terraform/mocks/matchers"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
	. "github.com/runatlantis/atlantis/testing"
)

func TestRun_NoDir(t *testing.T) {
	o := runtime.ApplyStepRunner{
		TerraformExecutor: nil,
	}
	_, err := o.Run(models.ProjectCommandContext{
		RepoRelDir: ".",
		Workspace:  "workspace",
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
		RepoRelDir: ".",
		Workspace:  "workspace",
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
		RepoRelDir:  ".",
		CommentArgs: []string{"comment", "args"},
	}, []string{"extra", "args"}, tmpDir)
	Ok(t, err)
	Equals(t, "output", output)
	terraform.VerifyWasCalledOnce().RunCommandWithVersion(nil, tmpDir, []string{"apply", "-input=false", "-no-color", "extra", "args", "comment", "args", fmt.Sprintf("%q", planPath)}, nil, "workspace")
	_, err = os.Stat(planPath)
	Assert(t, os.IsNotExist(err), "planfile should be deleted")
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
		Workspace:  "default",
		RepoRelDir: ".",
		ProjectConfig: &valid.Project{
			Name: &projectName,
		},
		CommentArgs: []string{"comment", "args"},
	}, []string{"extra", "args"}, tmpDir)
	Ok(t, err)
	Equals(t, "output", output)
	terraform.VerifyWasCalledOnce().RunCommandWithVersion(nil, tmpDir, []string{"apply", "-input=false", "-no-color", "extra", "args", "comment", "args", fmt.Sprintf("%q", planPath)}, nil, "default")
	_, err = os.Stat(planPath)
	Assert(t, os.IsNotExist(err), "planfile should be deleted")
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
		RepoRelDir:  ".",
		CommentArgs: []string{"comment", "args"},
		ProjectConfig: &valid.Project{
			TerraformVersion: tfVersion,
		},
	}, []string{"extra", "args"}, tmpDir)
	Ok(t, err)
	Equals(t, "output", output)
	terraform.VerifyWasCalledOnce().RunCommandWithVersion(nil, tmpDir, []string{"apply", "-input=false", "-no-color", "extra", "args", "comment", "args", fmt.Sprintf("%q", planPath)}, tfVersion, "workspace")
	_, err = os.Stat(planPath)
	Assert(t, os.IsNotExist(err), "planfile should be deleted")
}

// Apply ignores the -target flag when used with a planfile so we should give
// an error if it's being used with -target.
func TestRun_UsingTarget(t *testing.T) {
	cases := []struct {
		commentFlags []string
		extraArgs    []string
		expErr       bool
	}{
		{
			commentFlags: []string{"-target", "mytarget"},
			expErr:       true,
		},
		{
			commentFlags: []string{"-target=mytarget"},
			expErr:       true,
		},
		{
			extraArgs: []string{"-target", "mytarget"},
			expErr:    true,
		},
		{
			extraArgs: []string{"-target=mytarget"},
			expErr:    true,
		},
		{
			commentFlags: []string{"-target", "mytarget"},
			extraArgs:    []string{"-target=mytarget"},
			expErr:       true,
		},
		// Test false positives.
		{
			commentFlags: []string{"-targethahagotcha"},
			expErr:       false,
		},
		{
			extraArgs: []string{"-targethahagotcha"},
			expErr:    false,
		},
		{
			commentFlags: []string{"-targeted=weird"},
			expErr:       false,
		},
		{
			extraArgs: []string{"-targeted=weird"},
			expErr:    false,
		},
	}

	RegisterMockTestingT(t)

	for _, c := range cases {
		descrip := fmt.Sprintf("comments flags: %s extra args: %s",
			strings.Join(c.commentFlags, ", "), strings.Join(c.extraArgs, ", "))
		t.Run(descrip, func(t *testing.T) {
			tmpDir, cleanup := TempDir(t)
			defer cleanup()
			planPath := filepath.Join(tmpDir, "workspace.tfplan")
			err := ioutil.WriteFile(planPath, nil, 0644)
			Ok(t, err)
			terraform := mocks.NewMockClient()
			step := runtime.ApplyStepRunner{
				TerraformExecutor: terraform,
			}

			output, err := step.Run(models.ProjectCommandContext{
				Workspace:   "workspace",
				RepoRelDir:  ".",
				CommentArgs: c.commentFlags,
			}, c.extraArgs, tmpDir)
			Equals(t, "", output)
			if c.expErr {
				ErrEquals(t, "cannot run apply with -target because we are applying an already generated plan. Instead, run -target with atlantis plan", err)
			} else {
				Ok(t, err)
			}
		})
	}
}

// Test that apply works for remote applies.
func TestRun_RemoteApply(t *testing.T) {
	tmpDir, cleanup := TempDir(t)
	defer cleanup()
	planPath := filepath.Join(tmpDir, "workspace.tfplan")
	planFileContents := `Atlantis: this plan was created by remote ops

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  - destroy

Terraform will perform the following actions:

  - null_resource.hi[1]


Plan: 0 to add, 0 to change, 1 to destroy.`
	err := ioutil.WriteFile(planPath, []byte(planFileContents), 0644)
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
		RepoRelDir:  ".",
		CommentArgs: []string{"comment", "args"},
		ProjectConfig: &valid.Project{
			TerraformVersion: tfVersion,
		},
	}, []string{"extra", "args"}, tmpDir)
	Ok(t, err)
	Equals(t, "output", output)

	terraform.VerifyWasCalledOnce().RunCommandWithVersion(nil, tmpDir,
		[]string{"apply", "-input=false", "-no-color", "-auto-approve", "extra", "args", "comment", "args"},
		tfVersion, "workspace")
	_, err = os.Stat(planPath)
	Assert(t, os.IsNotExist(err), "planfile should be deleted")
}
