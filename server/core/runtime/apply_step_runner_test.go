package runtime_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	version "github.com/hashicorp/go-version"
	"go.uber.org/mock/gomock"
	"github.com/runatlantis/atlantis/server/core/runtime"
	tf "github.com/runatlantis/atlantis/server/core/terraform"
	"github.com/runatlantis/atlantis/server/core/terraform/mocks"
	tfclientmocks "github.com/runatlantis/atlantis/server/core/terraform/tfclient/mocks"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/logging"

	. "github.com/runatlantis/atlantis/testing"
)

func TestRun_NoDir(t *testing.T) {
	o := runtime.ApplyStepRunner{
		TerraformExecutor: nil,
	}
	_, err := o.Run(command.ProjectContext{
		RepoRelDir: ".",
		Workspace:  "workspace",
	}, nil, "/nonexistent/path", map[string]string(nil))
	ErrEquals(t, "no plan found at path \".\" and workspace \"workspace\"–did you run plan?", err)
}

func TestRun_NoPlanFile(t *testing.T) {
	tmpDir := t.TempDir()
	o := runtime.ApplyStepRunner{
		TerraformExecutor: nil,
	}
	_, err := o.Run(command.ProjectContext{
		RepoRelDir: ".",
		Workspace:  "workspace",
	}, nil, tmpDir, map[string]string(nil))
	ErrEquals(t, "no plan found at path \".\" and workspace \"workspace\"–did you run plan?", err)
}

func TestRun_Success(t *testing.T) {
	tmpDir := t.TempDir()
	planPath := filepath.Join(tmpDir, "workspace.tfplan")
	err := os.WriteFile(planPath, nil, 0600)
	logger := logging.NewNoopLogger(t)
	ctx := command.ProjectContext{
		Log:                logger,
		Workspace:          "workspace",
		RepoRelDir:         ".",
		EscapedCommentArgs: []string{"comment", "args"},
	}
	Ok(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	terraform := tfclientmocks.NewMockClient(ctrl)
	mockDownloader := mocks.NewMockDownloader(ctrl)
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	o := runtime.ApplyStepRunner{
		TerraformExecutor:     terraform,
		DefaultTFDistribution: tfDistribution,
	}

	// Set up expectation for the terraform command
	terraform.EXPECT().RunCommandWithVersion(
		ctx, 
		tmpDir, 
		[]string{"apply", "-input=false", "extra", "args", "comment", "args", fmt.Sprintf("%q", planPath)}, 
		map[string]string(nil), 
		tfDistribution, 
		nil, 
		"workspace").
		Return("output", nil)
		
	output, err := o.Run(ctx, []string{"extra", "args"}, tmpDir, map[string]string(nil))
	Ok(t, err)
	Equals(t, "output", output)
	_, err = os.Stat(planPath)
	Assert(t, os.IsNotExist(err), "planfile should be deleted")
}

func TestRun_AppliesCorrectProjectPlan(t *testing.T) {
	// When running for a project, the planfile has a different name.
	tmpDir := t.TempDir()
	planPath := filepath.Join(tmpDir, "projectname-default.tfplan")
	err := os.WriteFile(planPath, nil, 0600)

	logger := logging.NewNoopLogger(t)
	ctx := command.ProjectContext{
		Log:                logger,
		Workspace:          "default",
		RepoRelDir:         ".",
		ProjectName:        "projectname",
		EscapedCommentArgs: []string{"comment", "args"},
	}
	Ok(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	terraform := tfclientmocks.NewMockClient(ctrl)
	mockDownloader := mocks.NewMockDownloader(ctrl)
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	o := runtime.ApplyStepRunner{
		TerraformExecutor:     terraform,
		DefaultTFDistribution: tfDistribution,
	}
	
	// Set up expectation for the terraform command
	terraform.EXPECT().RunCommandWithVersion(
		ctx, 
		tmpDir, 
		[]string{"apply", "-input=false", "extra", "args", "comment", "args", fmt.Sprintf("%q", planPath)}, 
		map[string]string(nil), 
		tfDistribution, 
		nil, 
		"default").
		Return("output", nil)
		
	output, err := o.Run(ctx, []string{"extra", "args"}, tmpDir, map[string]string(nil))
	Ok(t, err)
	Equals(t, "output", output)
	_, err = os.Stat(planPath)
	Assert(t, os.IsNotExist(err), "planfile should be deleted")
}

func TestApplyStepRunner_TestRun_UsesConfiguredTFVersion(t *testing.T) {
	tmpDir := t.TempDir()
	planPath := filepath.Join(tmpDir, "workspace.tfplan")
	err := os.WriteFile(planPath, nil, 0600)
	Ok(t, err)

	logger := logging.NewNoopLogger(t)
	tfVersion, _ := version.NewVersion("0.11.0")
	ctx := command.ProjectContext{
		Workspace:          "workspace",
		RepoRelDir:         ".",
		EscapedCommentArgs: []string{"comment", "args"},
		TerraformVersion:   tfVersion,
		Log:                logger,
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	terraform := tfclientmocks.NewMockClient(ctrl)
	mockDownloader := mocks.NewMockDownloader(ctrl)
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	o := runtime.ApplyStepRunner{
		TerraformExecutor:     terraform,
		DefaultTFDistribution: tfDistribution,
	}
	
	// Set up expectation for the terraform command with specific version
	terraform.EXPECT().RunCommandWithVersion(
		ctx, 
		tmpDir, 
		[]string{"apply", "-input=false", "extra", "args", "comment", "args", fmt.Sprintf("%q", planPath)}, 
		map[string]string(nil), 
		tfDistribution, 
		tfVersion, 
		"workspace").
		Return("output", nil)
		
	output, err := o.Run(ctx, []string{"extra", "args"}, tmpDir, map[string]string(nil))
	Ok(t, err)
	Equals(t, "output", output)
	_, err = os.Stat(planPath)
	Assert(t, os.IsNotExist(err), "planfile should be deleted")
}

func TestApplyStepRunner_TestRun_UsesConfiguredDistribution(t *testing.T) {
	tmpDir := t.TempDir()
	planPath := filepath.Join(tmpDir, "workspace.tfplan")
	err := os.WriteFile(planPath, nil, 0600)
	Ok(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	logger := logging.NewNoopLogger(t)
	mockDownloader := mocks.NewMockDownloader(ctrl)
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	tfVersion, _ := version.NewVersion("0.11.0")
	projTFDistribution := "opentofu"
	ctx := command.ProjectContext{
		Workspace:             "workspace",
		RepoRelDir:            ".",
		EscapedCommentArgs:    []string{"comment", "args"},
		TerraformDistribution: &projTFDistribution,
		Log:                   logger,
	}

	terraform := tfclientmocks.NewMockClient(ctrl)
	o := runtime.ApplyStepRunner{
		TerraformExecutor:     terraform,
		DefaultTFDistribution: tfDistribution,
		DefaultTFVersion:      tfVersion,
	}
	
	// Set up expectation for the terraform command with non-default distribution
	notDefaultDistribution := gomock.Not(gomock.Eq(tfDistribution))
	terraform.EXPECT().RunCommandWithVersion(
		ctx, 
		tmpDir, 
		[]string{"apply", "-input=false", "extra", "args", "comment", "args", fmt.Sprintf("%q", planPath)}, 
		map[string]string(nil), 
		notDefaultDistribution, 
		tfVersion, 
		"workspace").
		Return("output", nil)
		
	output, err := o.Run(ctx, []string{"extra", "args"}, tmpDir, map[string]string(nil))
	Ok(t, err)
	Equals(t, "output", output)
	_, err = os.Stat(planPath)
	Assert(t, os.IsNotExist(err), "planfile should be deleted")
}

// Apply ignores the -target flag when used with a planfile so we should give
// an error if it's being used with -target.
func TestRun_UsingTarget(t *testing.T) {
	logger := logging.NewNoopLogger(t)
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

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	for _, c := range cases {
		descrip := fmt.Sprintf("comments flags: %s extra args: %s",
			strings.Join(c.commentFlags, ", "), strings.Join(c.extraArgs, ", "))
		t.Run(descrip, func(t *testing.T) {
			// Using t.TempDir() fails the test for an unknown reason
			tmpDir := "/tmp/dir"
			planPath := filepath.Join(tmpDir, "workspace.tfplan")
			err := os.MkdirAll(tmpDir, os.ModePerm)
			Ok(t, err)
			defer os.RemoveAll(tmpDir)
			err = os.WriteFile(planPath, nil, 0600)
			Ok(t, err)
			ctx := command.ProjectContext{
				Log:                logger,
				Workspace:          "workspace",
				RepoRelDir:         ".",
				EscapedCommentArgs: c.commentFlags,
			}

			terraform := tfclientmocks.NewMockClient(ctrl)
			mockDownloader := mocks.NewMockDownloader(ctrl)
			tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
			s := runtime.ApplyStepRunner{
				TerraformExecutor:     terraform,
				DefaultTFDistribution: tfDistribution,
			}

			_, err = s.Run(ctx, c.extraArgs, tmpDir, map[string]string(nil))
			if c.expErr {
				ErrEquals(t, "cannot run apply with -target flag because we are running apply on an already generated plan file–to use -target, delete the plan file first", err)
			} else {
				Ok(t, err)
			}
		})
	}
}