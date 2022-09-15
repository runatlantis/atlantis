package activities_test

import (
	"context"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/neptune/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities"
	"github.com/stretchr/testify/assert"
	"go.temporal.io/sdk/testsuite"
)

type testTfClient struct {
	t             *testing.T
	ctx           context.Context
	jobID         string
	path          string
	cmd           *terraform.SubCommand
	customEnvVars map[string]string
	version       *version.Version

	resp []terraform.Line
}

func (t *testTfClient) RunCommand(ctx context.Context, jobID string, path string, cmd *terraform.SubCommand, customEnvVars map[string]string, v *version.Version) <-chan terraform.Line {
	assert.Equal(t.t, jobID, t.jobID)
	assert.Equal(t.t, path, t.path)
	assert.Equal(t.t, cmd, t.cmd)
	assert.Equal(t.t, customEnvVars, t.customEnvVars)
	assert.Equal(t.t, v, t.version)

	ch := make(chan terraform.Line)
	go func(ch chan terraform.Line) {
		defer close(ch)
		for _, line := range t.resp {
			ch <- line
		}
	}(ch)

	return ch

}
func TestTerraformInit_TfVersionInRequestTakesPrecedence(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestActivityEnvironment()

	ctx := context.Background()
	path := "some/path"
	jobID := "1234"
	defVersion := "1.0.2"
	reqVersion := "0.12.0"

	defaultTfVersion, err := version.NewVersion(defVersion)
	assert.Nil(t, err)

	reqTfVersion, err := version.NewVersion(reqVersion)
	assert.Nil(t, err)

	expectedCmd := terraform.NewSubCommand(terraform.Init).WithArgs(terraform.Argument{
		Key:   "input",
		Value: "false",
	})
	testTfClient := testTfClient{
		t:             t,
		ctx:           ctx,
		jobID:         jobID,
		path:          path,
		cmd:           expectedCmd,
		customEnvVars: map[string]string{},
		version:       reqTfVersion,
		resp:          []terraform.Line{},
	}

	req := activities.TerraformInitRequest{
		Envs:      map[string]string{},
		JobID:     jobID,
		Path:      path,
		TfVersion: reqVersion,
	}

	tfActivity := activities.NewTerraformActivities(&testTfClient, defaultTfVersion)
	env.RegisterActivity(tfActivity)

	_, err = env.ExecuteActivity(tfActivity.TerraformInit, req)
	assert.NoError(t, err)
}

func TestTerraformInit_ExtraArgsTakesPrecedenceOverCommandArgs(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestActivityEnvironment()

	ctx := context.Background()
	path := "some/path"
	jobID := "1234"
	defVersion := "1.0.2"
	reqVersion := "0.12.0"

	defaultTfVersion, err := version.NewVersion(defVersion)
	assert.Nil(t, err)

	reqTfVersion, err := version.NewVersion(reqVersion)
	assert.Nil(t, err)

	expectedCmd := terraform.NewSubCommand(terraform.Init).WithArgs(terraform.Argument{
		Key:   "input",
		Value: "true",
	})
	testTfClient := testTfClient{
		t:             t,
		ctx:           ctx,
		jobID:         jobID,
		path:          path,
		cmd:           expectedCmd,
		customEnvVars: map[string]string{},
		version:       reqTfVersion,
		resp:          []terraform.Line{},
	}

	req := activities.TerraformInitRequest{
		Args: []terraform.Argument{
			{
				Key:   "input",
				Value: "true",
			},
		},
		Envs:      map[string]string{},
		JobID:     jobID,
		Path:      path,
		TfVersion: reqVersion,
	}

	tfActivity := activities.NewTerraformActivities(&testTfClient, defaultTfVersion)
	env.RegisterActivity(tfActivity)

	_, err = env.ExecuteActivity(tfActivity.TerraformInit, req)
	assert.NoError(t, err)
}

func TestTerraformPlan(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestActivityEnvironment()

	ctx := context.Background()
	path := "some/path"
	jobID := "1234"
	defVersion := "1.0.2"
	reqVersion := "1.2.2"

	defaultTfVersion, err := version.NewVersion(defVersion)
	assert.Nil(t, err)

	reqTfVersion, err := version.NewVersion(reqVersion)
	assert.Nil(t, err)

	expectedCmd := terraform.NewSubCommand(terraform.Plan).
		WithArgs(terraform.Argument{
			Key:   "input",
			Value: "false",
		}, terraform.Argument{
			Key:   "refresh",
			Value: "true",
		}, terraform.Argument{
			Key:   "out",
			Value: "some/path/output.tfplan",
		})

	testTfClient := testTfClient{
		t:             t,
		ctx:           ctx,
		jobID:         jobID,
		path:          path,
		cmd:           expectedCmd,
		customEnvVars: map[string]string{},
		version:       reqTfVersion,
		resp:          []terraform.Line{},
	}

	req := activities.TerraformPlanRequest{
		Envs:      map[string]string{},
		JobID:     jobID,
		Path:      path,
		TfVersion: reqVersion,
	}

	tfActivity := activities.NewTerraformActivities(&testTfClient, defaultTfVersion)
	env.RegisterActivity(tfActivity)

	resp, err := env.ExecuteActivity(tfActivity.TerraformPlan, req)
	assert.NoError(t, err)

	var planResp activities.TerraformPlanResponse
	assert.Nil(t, resp.Get(&planResp))
	assert.Equal(t, planResp.PlanFile, "some/path/output.tfplan")
}

func TestTerraformApply(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestActivityEnvironment()

	ctx := context.Background()
	path := "some/path"
	jobID := "1234"
	defVersion := "1.0.2"
	reqVersion := "1.2.2"

	defaultTfVersion, err := version.NewVersion(defVersion)
	assert.Nil(t, err)

	reqTfVersion, err := version.NewVersion(reqVersion)
	assert.Nil(t, err)

	expectedCmd := terraform.NewSubCommand(terraform.Apply).
		WithArgs(terraform.Argument{
			Key:   "input",
			Value: "false",
		}).
		WithInput("some/path/output.tfplan")

	testTfClient := testTfClient{
		t:             t,
		ctx:           ctx,
		jobID:         jobID,
		path:          path,
		cmd:           expectedCmd,
		customEnvVars: map[string]string{},
		version:       reqTfVersion,
		resp:          []terraform.Line{},
	}

	req := activities.TerraformApplyRequest{
		Envs:      map[string]string{},
		JobID:     jobID,
		Path:      path,
		TfVersion: reqVersion,
	}

	tfActivity := activities.NewTerraformActivities(&testTfClient, defaultTfVersion)
	env.RegisterActivity(tfActivity)

	resp, err := env.ExecuteActivity(tfActivity.TerraformApply, req)
	assert.NoError(t, err)

	var applyResp activities.TerraformApplyResponse
	assert.Nil(t, resp.Get(&applyResp))
}

func TestTerraformApply_TargetFailure(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestActivityEnvironment()

	ctx := context.Background()
	path := "some/path"
	jobID := "1234"
	defVersion := "1.0.2"
	reqVersion := "1.2.2"

	defaultTfVersion, err := version.NewVersion(defVersion)
	assert.Nil(t, err)

	reqTfVersion, err := version.NewVersion(reqVersion)
	assert.Nil(t, err)

	expectedCmd := terraform.NewSubCommand(terraform.Apply).
		WithArgs(terraform.Argument{
			Key:   "input",
			Value: "false",
		}).
		WithInput("some/path/output.tfplan")

	testTfClient := testTfClient{
		t:             t,
		ctx:           ctx,
		jobID:         jobID,
		path:          path,
		cmd:           expectedCmd,
		customEnvVars: map[string]string{},
		version:       reqTfVersion,
		resp:          []terraform.Line{},
	}

	req := activities.TerraformApplyRequest{
		Envs:      map[string]string{},
		JobID:     jobID,
		Path:      path,
		TfVersion: reqVersion,
		Args: []terraform.Argument{
			{
				Key:   "target",
				Value: "anything",
			},
		},
	}

	tfActivity := activities.NewTerraformActivities(&testTfClient, defaultTfVersion)
	env.RegisterActivity(tfActivity)

	_, err = env.ExecuteActivity(tfActivity.TerraformApply, req)
	assert.ErrorContains(t, err, "request contains invalid -target flag")
}
