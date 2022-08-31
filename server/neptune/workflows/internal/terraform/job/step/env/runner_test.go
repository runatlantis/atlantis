package env_test

import (
	"context"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/job"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/root"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform/job/step/cmd"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform/job/step/env"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

const (
	RepoName    = "test-repo"
	RepoOwner   = "test-owner"
	RepoPath    = "test/repo"
	ProjectName = "test-project"
	ProjectPath = "test/repo/project"
	HeadCommit  = "ref"
	Dir         = "test-path"
	UserName    = "test-user"
)

type request struct {
	RootInstance root.RootInstance
	Step         job.Step
}

type testExecuteActivity struct {
}

func (a *testExecuteActivity) ExecuteCommand(ctx context.Context, request activities.ExecuteCommandRequest) (activities.ExecuteCommandResponse, error) {
	return activities.ExecuteCommandResponse{}, nil
}

func testWorkflow(ctx workflow.Context, r request) (string, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToCloseTimeout: 5 * time.Second,
	})

	jobExecutionCtx := &job.ExecutionContext{
		Context:   ctx,
		Path:      r.RootInstance.Root.Path,
		Envs:      map[string]string{},
		TfVersion: r.RootInstance.Root.TfVersion,
	}

	var a *testExecuteActivity
	envStepRunner := env.Runner{
		CmdRunner: cmd.Runner{
			Activity: a,
		},
	}

	return envStepRunner.Run(jobExecutionCtx, &r.RootInstance, r.Step)
}

func TestEnvRunner_EnvVarValueNotSet(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	testExecuteActivity := &testExecuteActivity{}
	env.RegisterActivity(testExecuteActivity)
	env.RegisterWorkflow(testWorkflow)

	env.OnActivity(testExecuteActivity.ExecuteCommand, mock.Anything, activities.ExecuteCommandRequest{
		Step: job.Step{
			StepName:   "env",
			RunCommand: "echo 'Hello World'",
		},
		Path: ProjectPath,
		EnvVars: map[string]string{
			"REPO_NAME":    RepoName,
			"REPO_OWNER":   RepoOwner,
			"DIR":          ProjectPath,
			"HEAD_COMMIT":  HeadCommit,
			"PROJECT_NAME": ProjectName,
			"REPO_REL_DIR": "project",
			"USER_NAME":    UserName,
		},
	}).Return(activities.ExecuteCommandResponse{
		Output: "Hello World",
	}, nil)

	env.ExecuteWorkflow(testWorkflow, request{
		RootInstance: root.RootInstance{
			Root: root.Root{
				Name: ProjectName,
				Path: ProjectPath,
			},
			Repo: github.RepoInstance{
				Path:  RepoPath,
				Name:  RepoName,
				Owner: RepoOwner,
				HeadCommit: github.Commit{
					Ref: HeadCommit,
					Author: github.User{
						Username: UserName,
					},
				},
			},
		},
		Step: job.Step{
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
	executioncontext := &job.ExecutionContext{}
	rootInstance := &root.RootInstance{}

	step := job.Step{
		EnvVarName:  "TEST_VAR",
		EnvVarValue: "TEST_VALUE",
	}

	runner := env.Runner{}

	out, err := runner.Run(executioncontext, rootInstance, step)
	assert.Nil(t, err)
	assert.Equal(t, out, step.EnvVarValue)
}
