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
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

type testCmdExecuteActivity struct {
	t           *testing.T
	expectedReq map[string]string
}

func (a *testCmdExecuteActivity) ExecuteCommand(ctx context.Context, request activities.ExecuteCommandRequest) (activities.ExecuteCommandResponse, error) {
	assert.Equal(a.t, a.expectedReq, request.EnvVars)
	return activities.ExecuteCommandResponse{}, nil
}

func testCmdWorkflow(ctx workflow.Context, r request) (string, error) {
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
	cmdStepRunner := runner.CmdStepRunner{
		Activity: a,
	}

	return cmdStepRunner.Run(jobExecutionCtx, &r.LocalRoot, r.Step)
}

func TestRunRunner_ShouldSetupEnvVars(t *testing.T) {

	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	expectedEnvVars := map[string]string{
		"BASE_REPO_NAME":  RepoName,
		"BASE_REPO_OWNER": RepoOwner,
		"DIR":             ProjectPath,
		"HEAD_COMMIT":     "refs/heads/main",
		"PROJECT_NAME":    ProjectName,
		"REPO_REL_DIR":    "project",
		"USER_NAME":       UserName,
	}
	testExecuteActivity := &testCmdExecuteActivity{
		t:           t,
		expectedReq: expectedEnvVars,
	}
	env.RegisterActivity(testExecuteActivity)
	env.RegisterWorkflow(testCmdWorkflow)

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
		Step: job.Step{},
	})

	var resp string
	err := env.GetWorkflowResult(&resp)
	assert.NoError(t, err)
}
