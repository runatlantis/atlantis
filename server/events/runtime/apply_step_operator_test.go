package runtime_test

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/events/mocks/matchers"
	matchers2 "github.com/runatlantis/atlantis/server/events/run/mocks/matchers"
	"github.com/runatlantis/atlantis/server/events/runtime"
	"github.com/runatlantis/atlantis/server/events/terraform/mocks"
	. "github.com/runatlantis/atlantis/testing"
)

func TestRun_NoDir(t *testing.T) {
	s := runtime.ApplyStep{
		Meta: runtime.StepMeta{
			Workspace:             "workspace",
			AbsolutePath:          "nonexistent/path",
			DirRelativeToRepoRoot: ".",
			TerraformVersion:      nil,
			ExtraCommentArgs:      nil,
			Username:              "username",
		},
	}
	_, err := s.Run()
	ErrEquals(t, "no plan found at path \".\" and workspace \"workspace\"–did you run plan?", err)
}

func TestRun_NoPlanFile(t *testing.T) {
	tmpDir, cleanup := TempDir(t)
	defer cleanup()

	s := runtime.ApplyStep{
		Meta: runtime.StepMeta{
			Workspace:             "workspace",
			AbsolutePath:          tmpDir,
			DirRelativeToRepoRoot: ".",
			TerraformVersion:      nil,
			ExtraCommentArgs:      nil,
			Username:              "username",
		},
	}
	_, err := s.Run()
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

	tfVersion, _ := version.NewVersion("0.11.4")
	s := runtime.ApplyStep{
		Meta: runtime.StepMeta{
			Workspace:             "workspace",
			AbsolutePath:          tmpDir,
			DirRelativeToRepoRoot: ".",
			TerraformExecutor:     terraform,
			TerraformVersion:      tfVersion,
			ExtraCommentArgs:      []string{"comment", "args"},
			Username:              "username",
		},
		ExtraArgs: []string{"extra", "args"},
	}

	When(terraform.RunCommandWithVersion(matchers.AnyPtrToLoggingSimpleLogger(), AnyString(), AnyStringSlice(), matchers2.AnyPtrToGoVersionVersion(), AnyString())).
		ThenReturn("output", nil)
	output, err := s.Run()
	Ok(t, err)
	Equals(t, "output", output)
	terraform.VerifyWasCalledOnce().RunCommandWithVersion(nil, tmpDir, []string{"apply", "-no-color", "extra", "args", "comment", "args", planPath}, tfVersion, "workspace")
}
