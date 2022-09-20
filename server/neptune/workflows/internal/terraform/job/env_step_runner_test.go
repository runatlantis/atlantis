package job_test

import (
	"context"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/job"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/root"
	runner "github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform/job"
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
	RefName     = "main"
	RefType     = "branch"
	Dir         = "test-path"
	UserName    = "test-user"
)

type request struct {
	LocalRoot root.LocalRoot
	Step      job.Step
}

type testEnvExecuteActivity struct {
}

func (a *testEnvExecuteActivity) ExecuteCommand(ctx context.Context, request activities.ExecuteCommandRequest) (activities.ExecuteCommandResponse, error) {
	return activities.ExecuteCommandResponse{}, nil
}

func testEnvWorkflow(ctx workflow.Context, r request) (string, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToCloseTimeout: 5 * time.Second,
	})

	jobExecutionCtx := &job.ExecutionContext{
		Context:   ctx,
		Path:      ProjectPath,
		Envs:      map[string]string{},
		TfVersion: r.LocalRoot.Root.TfVersion,
	}

	var a *testCmdExecuteActivity
	envStepRunner := runner.EnvStepRunner{
		CmdStepRunner: runner.CmdStepRunner{
			Activity: a,
		},
	}

	return envStepRunner.Run(jobExecutionCtx, &r.LocalRoot, r.Step)
}

func TestEnvRunner_EnvVarValueNotSet(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	testExecuteActivity := &testCmdExecuteActivity{}
	env.RegisterActivity(testExecuteActivity)
	env.RegisterWorkflow(testCmdWorkflow)

	env.OnActivity(testExecuteActivity.ExecuteCommand, mock.Anything, activities.ExecuteCommandRequest{
		Step: job.Step{
			StepName:   "env",
			RunCommand: "echo 'Hello World'",
		},
		Path: ProjectPath,
		EnvVars: map[string]string{
			"BASE_REPO_NAME":    RepoName,
			"BASE_REPO_OWNER":   RepoOwner,
			"DIR":          ProjectPath,
			"HEAD_COMMIT":  "refs/heads/main",
			"PROJECT_NAME": ProjectName,
			"REPO_REL_DIR": "project",
			"USER_NAME":    UserName,
		},
	}).Return(activities.ExecuteCommandResponse{
		Output: "Hello World",
	}, nil)

	env.ExecuteWorkflow(testCmdWorkflow, request{
		LocalRoot: root.LocalRoot{
			Root: root.Root{
				Name: ProjectName,
				Path: "project",
			},
			Repo: github.Repo{
				Name:  RepoName,
				Owner: RepoOwner,
				HeadCommit: github.Commit{
					Ref: github.Ref{
						Name: RefName,
						Type: RefType,
					},
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
	localRoot := &root.LocalRoot{}

	step := job.Step{
		EnvVarName:  "TEST_VAR",
		EnvVarValue: "TEST_VALUE",
	}

	runner := runner.EnvStepRunner{}

	out, err := runner.Run(executioncontext, localRoot, step)
	assert.Nil(t, err)
	assert.Equal(t, out, step.EnvVarValue)
}
