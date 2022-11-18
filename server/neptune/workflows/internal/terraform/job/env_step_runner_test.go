package job_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/neptune/workflows/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/execute"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
	runner "github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform/job"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/testsuite"
)

const (
	RepoName    = "test-repo"
	RepoOwner   = "test-owner"
	ProjectName = "test-project"
	ProjectPath = "test/repo/project"
)

type request struct {
	LocalRoot terraform.LocalRoot
	Step      execute.Step
}

func TestEnvRunner_EnvVarValueNotSet(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	testExecuteActivity := &testCmdExecuteActivity{}
	env.RegisterActivity(testExecuteActivity)
	env.RegisterWorkflow(testCmdWorkflow)

	env.OnActivity(testExecuteActivity.ExecuteCommand, mock.Anything, activities.ExecuteCommandRequest{
		Step: execute.Step{
			StepName:   "env",
			RunCommand: "echo 'Hello World'",
		},
		Path: ProjectPath,
		EnvVars: map[string]string{
			"BASE_REPO_NAME":  RepoName,
			"BASE_REPO_OWNER": RepoOwner,
			"DIR":             ProjectPath,
			"PROJECT_NAME":    ProjectName,
			"REPO_REL_DIR":    "project",
		},
	}).Return(activities.ExecuteCommandResponse{
		Output: "Hello World",
	}, nil)

	env.ExecuteWorkflow(testCmdWorkflow, request{
		LocalRoot: terraform.LocalRoot{
			Root: terraform.Root{
				Name: ProjectName,
				Path: "project",
			},
			Repo: github.Repo{
				Name:  RepoName,
				Owner: RepoOwner,
			},
		},
		Step: execute.Step{
			StepName:   "env",
			RunCommand: "echo 'Hello World'",
		},
	})

	var resp string
	err := env.GetWorkflowResult(&resp)
	assert.NoError(t, err)

	assert.Equal(t, "Hello World", resp)
}

func TestEnvRunne_EnvVarValueSet(t *testing.T) {
	executioncontext := &runner.ExecutionContext{}
	localRoot := &terraform.LocalRoot{}

	step := execute.Step{
		EnvVarName:  "TEST_VAR",
		EnvVarValue: "TEST_VALUE",
	}

	runner := runner.EnvStepRunner{}

	out, err := runner.Run(executioncontext, localRoot, step)
	assert.Nil(t, err)
	assert.Equal(t, out, step.EnvVarValue)
}
