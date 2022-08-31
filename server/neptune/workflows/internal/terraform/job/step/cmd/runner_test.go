package cmd_test

import (
	"context"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/job"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/root"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform/job/step/cmd"
	"github.com/stretchr/testify/assert"
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
	t           *testing.T
	expectedReq map[string]string
}

func (a *testExecuteActivity) ExecuteCommand(ctx context.Context, request activities.ExecuteCommandRequest) (activities.ExecuteCommandResponse, error) {
	assert.Equal(a.t, a.expectedReq, request.EnvVars)
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
	cmdStepRunner := cmd.Runner{
		Activity: a,
	}

	return cmdStepRunner.Run(jobExecutionCtx, &r.RootInstance, r.Step)
}

func TestRunRunner_ShouldSetupEnvVars(t *testing.T) {

	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	expectedEnvVars := map[string]string{
		"REPO_NAME":    RepoName,
		"REPO_OWNER":   RepoOwner,
		"DIR":          ProjectPath,
		"HEAD_COMMIT":  HeadCommit,
		"PROJECT_NAME": ProjectName,
		"REPO_REL_DIR": "project",
		"USER_NAME":    UserName,
	}
	testExecuteActivity := &testExecuteActivity{
		t:           t,
		expectedReq: expectedEnvVars,
	}
	env.RegisterActivity(testExecuteActivity)
	env.RegisterWorkflow(testWorkflow)

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
		Step: job.Step{},
	})

	var resp string
	err := env.GetWorkflowResult(&resp)
	assert.NoError(t, err)
}
