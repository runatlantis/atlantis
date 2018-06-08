package runtime_test

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/events/runtime"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

// When there is no config file, should use the defaults.
func TestBuildStage_NoConfigFile(t *testing.T) {
	var defaultTFVersion *version.Version
	var terraformExecutor runtime.TerraformExec
	e := runtime.ExecutionPlanner{
		DefaultTFVersion:  defaultTFVersion,
		TerraformExecutor: terraformExecutor,
	}

	log := logging.NewNoopLogger()
	repoDir := "/willnotexist"
	workspace := "myworkspace"
	relProjectPath := "mydir"
	var extraCommentArgs []string
	username := "myuser"
	meta := runtime.StepMeta{
		Log:                   log,
		Workspace:             workspace,
		AbsolutePath:          filepath.Join(repoDir, relProjectPath),
		DirRelativeToRepoRoot: relProjectPath,
		TerraformVersion:      defaultTFVersion,
		TerraformExecutor:     terraformExecutor,
		ExtraCommentArgs:      extraCommentArgs,
		Username:              username,
	}

	// Test the plan stage first.
	t.Run("plan stage", func(t *testing.T) {
		planStage, err := e.BuildPlanStage(log, repoDir, workspace, relProjectPath, extraCommentArgs, username)
		Ok(t, err)
		Equals(t, runtime.PlanStage{
			Steps: []runtime.Step{
				&runtime.InitStep{
					Meta: meta,
				},
				&runtime.PlanStep{
					Meta: meta,
				},
			},
		}, *planStage)
	})

	// Then the apply stage.
	t.Run("apply stage", func(t *testing.T) {
		applyStage, err := e.BuildApplyStage(log, repoDir, workspace, relProjectPath, extraCommentArgs, username)
		Ok(t, err)
		Equals(t, runtime.ApplyStage{
			Steps: []runtime.Step{
				&runtime.ApplyStep{
					Meta: meta,
				},
			},
		}, *applyStage)
	})
}

func TestBuildStage(t *testing.T) {
	var defaultTFVersion *version.Version
	var terraformExecutor runtime.TerraformExec
	e := runtime.ExecutionPlanner{
		DefaultTFVersion:  defaultTFVersion,
		TerraformExecutor: terraformExecutor,
	}

	// Write atlantis.yaml config.
	tmpDir, cleanup := TempDir(t)
	defer cleanup()
	err := ioutil.WriteFile(filepath.Join(tmpDir, "atlantis.yaml"), []byte(`
version: 2
projects:
- dir: "."
  workflow: custom
workflows:
  custom:
    plan:
      steps:
      - init:
          extra_args: [arg1, arg2]
      - plan
      - run: echo hi
    apply:
      steps:
      - run: prerun
      - apply:
          extra_args: [arg3, arg4]
      - run: postrun
`), 0644)
	Ok(t, err)

	repoDir := tmpDir
	log := logging.NewNoopLogger()
	workspace := "myworkspace"
	// Our config is for '.' so there will be no config for this project.
	relProjectPath := "mydir"
	var extraCommentArgs []string
	username := "myuser"
	meta := runtime.StepMeta{
		Log:                   log,
		Workspace:             workspace,
		AbsolutePath:          filepath.Join(repoDir, relProjectPath),
		DirRelativeToRepoRoot: relProjectPath,
		TerraformVersion:      defaultTFVersion,
		TerraformExecutor:     terraformExecutor,
		ExtraCommentArgs:      extraCommentArgs,
		Username:              username,
	}

	t.Run("plan stage for project without config", func(t *testing.T) {
		// This project isn't listed so it should get the defaults.
		planStage, err := e.BuildPlanStage(log, repoDir, workspace, relProjectPath, extraCommentArgs, username)
		Ok(t, err)
		Equals(t, runtime.PlanStage{
			Steps: []runtime.Step{
				&runtime.InitStep{
					Meta: meta,
				},
				&runtime.PlanStep{
					Meta: meta,
				},
			},
		}, *planStage)
	})

	t.Run("apply stage for project without config", func(t *testing.T) {
		// This project isn't listed so it should get the defaults.
		applyStage, err := e.BuildApplyStage(log, repoDir, workspace, relProjectPath, extraCommentArgs, username)
		Ok(t, err)
		Equals(t, runtime.ApplyStage{
			Steps: []runtime.Step{
				&runtime.ApplyStep{
					Meta: meta,
				},
			},
		}, *applyStage)
	})

	// Create the meta for the custom project.
	customMeta := meta
	customMeta.Workspace = "default"
	customMeta.DirRelativeToRepoRoot = "."
	customMeta.AbsolutePath = tmpDir

	t.Run("plan stage for custom config", func(t *testing.T) {
		planStage, err := e.BuildPlanStage(log, repoDir, "default", ".", extraCommentArgs, username)
		Ok(t, err)

		Equals(t, runtime.PlanStage{
			Steps: []runtime.Step{
				&runtime.InitStep{
					Meta:      customMeta,
					ExtraArgs: []string{"arg1", "arg2"},
				},
				&runtime.PlanStep{
					Meta: customMeta,
				},
				&runtime.RunStep{
					Meta:     customMeta,
					Commands: []string{"echo", "hi"},
				},
			},
		}, *planStage)
	})

	t.Run("apply stage for custom config", func(t *testing.T) {
		planStage, err := e.BuildApplyStage(log, repoDir, "default", ".", extraCommentArgs, username)
		Ok(t, err)

		Equals(t, runtime.ApplyStage{
			Steps: []runtime.Step{
				&runtime.RunStep{
					Meta:     customMeta,
					Commands: []string{"prerun"},
				},
				&runtime.ApplyStep{
					Meta:      customMeta,
					ExtraArgs: []string{"arg3", "arg4"},
				},
				&runtime.RunStep{
					Meta:     customMeta,
					Commands: []string{"postrun"},
				},
			},
		}, *planStage)
	})
}
