package job_test

import (
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/neptune/workflows/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/execute"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
	runner "github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform/job"
	"github.com/stretchr/testify/assert"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
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

func testEnvRunnerWorkflow(ctx workflow.Context, r request) (activities.EnvVar, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToCloseTimeout: 5 * time.Second,
	})

	jobExecutionCtx := &runner.ExecutionContext{
		Context:   ctx,
		Path:      ProjectPath,
		TfVersion: r.LocalRoot.Root.TfVersion,
	}

	var a *testCmdExecuteActivity

	envStepRunner := runner.EnvStepRunner{
		CmdStepRunner: runner.CmdStepRunner{
			Activity: a,
		},
	}

	envVar, err := envStepRunner.Run(jobExecutionCtx, &r.LocalRoot, r.Step)
	return envVar.ToActivityEnvVar(), err
}

func TestEnvRunner_CommandEnvVar(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	testExecuteActivity := &testCmdExecuteActivity{}
	env.RegisterActivity(testExecuteActivity)
	env.RegisterWorkflow(testEnvRunnerWorkflow)

	env.ExecuteWorkflow(testEnvRunnerWorkflow, request{
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
			EnvVarName: "nish",
			RunCommand: "echo 'Hello World'",
		},
	})

	var resp activities.EnvVar
	assert.NoError(t, env.GetWorkflowResult(&resp))

	assert.Equal(t, activities.EnvVar{
		Name: "nish",
		Command: activities.StringCommand{
			Command: "echo 'Hello World'",
			Dir:     ProjectPath,
			AdditionalEnvs: map[string]string{
				"BASE_REPO_NAME":  RepoName,
				"BASE_REPO_OWNER": RepoOwner,
				"DIR":             ProjectPath,
				"PROJECT_NAME":    ProjectName,
				"REPO_REL_DIR":    "project",
			},
		},
	}, resp)
}

func TestEnvRunner_StringEnvVar(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	testExecuteActivity := &testCmdExecuteActivity{}
	env.RegisterActivity(testExecuteActivity)
	env.RegisterWorkflow(testEnvRunnerWorkflow)

	env.ExecuteWorkflow(testEnvRunnerWorkflow, request{
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
			StepName:    "env",
			EnvVarName:  "nish",
			EnvVarValue: "Hello",
		},
	})

	var resp activities.EnvVar
	assert.NoError(t, env.GetWorkflowResult(&resp))

	assert.Equal(t, activities.EnvVar{
		Name:  "nish",
		Value: "Hello",
	}, resp)

}
