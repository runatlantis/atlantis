package job_test

import (
	"context"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/neptune/workflows/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/execute"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/github"
	terraform_model "github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform/job"
	"github.com/stretchr/testify/assert"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

const JobID = "1234"

var repo = github.Repo{
	Name:  RepoName,
	Owner: RepoOwner,
}

type testTerraformActivity struct {
	t    *testing.T
	plan struct {
		req  activities.TerraformPlanRequest
		resp activities.TerraformPlanResponse
		err  error
	}
	apply struct {
		req  activities.TerraformApplyRequest
		resp activities.TerraformApplyResponse
		err  error
	}
	close struct {
		req activities.CloseJobRequest
		err error
	}
}

func (t *testTerraformActivity) TerraformInit(ctx context.Context, request activities.TerraformInitRequest) (activities.TerraformInitResponse, error) {
	return activities.TerraformInitResponse{}, nil
}

func (t *testTerraformActivity) TerraformPlan(ctx context.Context, request activities.TerraformPlanRequest) (activities.TerraformPlanResponse, error) {
	assert.Equal(t.t, t.plan.req, request)
	return t.plan.resp, t.plan.err
}

func (t *testTerraformActivity) TerraformApply(ctx context.Context, request activities.TerraformApplyRequest) (activities.TerraformApplyResponse, error) {
	assert.Equal(t.t, t.apply.req, request)
	return t.apply.resp, t.apply.err
}

func (t *testTerraformActivity) CloseJob(ctx context.Context, request activities.CloseJobRequest) error {
	assert.Equal(t.t, t.close.req, request)
	return t.close.err
}

// test workflow that runs the plan job
func testJobPlanWorkflow(ctx workflow.Context, r terraform.Request) (activities.TerraformPlanResponse, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToCloseTimeout: 100 * time.Second,
	})

	localRoot := terraform_model.LocalRoot{
		Root: r.Root,
		Repo: r.Repo,
		Path: ProjectPath,
	}

	var a *testTerraformActivity
	jobRunner := job.NewRunner(&job.CmdStepRunner{}, &job.EnvStepRunner{}, a)

	return jobRunner.Plan(ctx, &localRoot, JobID)
}

// test workflow that runs the plan job
func testJobApplyWorkflow(ctx workflow.Context, r terraform.Request) error {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToCloseTimeout: 100 * time.Second,
	})

	localRoot := terraform_model.LocalRoot{
		Root: r.Root,
		Repo: r.Repo,
		Path: ProjectPath,
	}

	var a *testTerraformActivity
	jobRunner := job.NewRunner(&job.CmdStepRunner{}, &job.EnvStepRunner{}, a)
	return jobRunner.Apply(ctx, &localRoot, JobID, "")
}

func TestJobRunner_Plan(t *testing.T) {
	t.Run("should close job after plan operation", func(t *testing.T) {
		ts := testsuite.WorkflowTestSuite{}
		env := ts.NewTestWorkflowEnvironment()
		testTerraformActivity := &testTerraformActivity{
			t: t,
			plan: struct {
				req  activities.TerraformPlanRequest
				resp activities.TerraformPlanResponse
				err  error
			}{
				req: activities.TerraformPlanRequest{
					JobID: JobID,
					Args:  []terraform_model.Argument{},
					Envs:  map[string]string{},
					DynamicEnvs: []activities.EnvVar{
						{
							Name:  "env1",
							Value: "v1",
						},
					},
					Path: ProjectPath,
				},
			},
			close: struct {
				req activities.CloseJobRequest
				err error
			}{
				req: activities.CloseJobRequest{
					JobID: JobID,
				},
			},
		}
		env.RegisterActivity(testTerraformActivity)
		env.RegisterWorkflow(testJobPlanWorkflow)

		env.ExecuteWorkflow(testJobPlanWorkflow, terraform.Request{
			Root: getTestRootForPlan(),
			Repo: repo,
		})

		var resp activities.TerraformPlanResponse
		err := env.GetWorkflowResult(&resp)
		assert.NoError(t, err)
	})
}

func TestJobRunner_Apply(t *testing.T) {
	t.Run("should close job after apply operation", func(t *testing.T) {
		ts := testsuite.WorkflowTestSuite{}
		env := ts.NewTestWorkflowEnvironment()
		testTerraformActivity := &testTerraformActivity{
			t: t,
			apply: struct {
				req  activities.TerraformApplyRequest
				resp activities.TerraformApplyResponse
				err  error
			}{
				req: activities.TerraformApplyRequest{
					JobID: JobID,
					Args:  []terraform_model.Argument{},
					Envs:  map[string]string{},
					DynamicEnvs: []activities.EnvVar{
						{
							Name:  "env1",
							Value: "v1",
						},
					},
					Path: ProjectPath,
				},
			},
			close: struct {
				req activities.CloseJobRequest
				err error
			}{
				req: activities.CloseJobRequest{
					JobID: JobID,
				},
			},
		}
		env.RegisterActivity(testTerraformActivity)
		env.RegisterWorkflow(testJobApplyWorkflow)

		env.ExecuteWorkflow(testJobApplyWorkflow, terraform.Request{
			Root: getTestRootForApply(),
			Repo: repo,
		})
	})
}

func getTestRootForPlan() terraform_model.Root {
	return terraform_model.Root{
		Name: ProjectName,
		Path: "project",
		Plan: terraform_model.PlanJob{
			Job: execute.Job{
				Steps: []execute.Step{
					{
						StepName:    "env",
						EnvVarName:  "env1",
						EnvVarValue: "v1",
					},
					{
						StepName: "plan",
					},
				},
			},
		},
	}
}

func getTestRootForApply() terraform_model.Root {
	return terraform_model.Root{
		Name: ProjectName,
		Path: "project",
		Apply: execute.Job{
			Steps: []execute.Step{
				{
					StepName:    "env",
					EnvVarName:  "env1",
					EnvVarValue: "v1",
				},
				{
					StepName: "apply",
				},
			},
		},
	}
}
