package activities

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities/terraform"
	"github.com/stretchr/testify/assert"
	"go.temporal.io/sdk/testsuite"
)

type testStreamHandler struct {
	received      []string
	expectedJobID string
	t             *testing.T
	called        bool
}

func (t *testStreamHandler) Stream(jobID string, msg string) {
	assert.Equal(t.t, t.expectedJobID, jobID)
	t.received = append(t.received, msg)

	t.called = true

}

func (t *testStreamHandler) Close(ctx context.Context, jobID string) {}

type multiCallTfClient struct {
	clients []*testTfClient

	count int
}

func (t *multiCallTfClient) RunCommand(ctx context.Context, request *terraform.RunCommandRequest, options ...terraform.RunOptions) error {
	if t.count >= len(t.clients) {
		return fmt.Errorf("expected less calls to RunCommand")
	}
	t.clients[t.count].RunCommand(ctx, request, options...)
	t.count++

	return nil
}

func (t *multiCallTfClient) AssertExpectations() error {
	if t.count != len(t.clients) {
		return fmt.Errorf("expected %d calls but got %d", len(t.clients), t.count)
	}
	return nil
}

type testTfClient struct {
	t             *testing.T
	jobID         string
	path          string
	cmd           *terraform.SubCommand
	customEnvVars map[string]string
	version       *version.Version
	resp          string

	expectedError error
}

func (t *testTfClient) RunCommand(ctx context.Context, request *terraform.RunCommandRequest, options ...terraform.RunOptions) error {
	assert.Equal(t.t, t.path, request.RootPath)
	assert.Equal(t.t, t.cmd, request.SubCommand)
	assert.Equal(t.t, t.customEnvVars, request.AdditionalEnvVars)
	assert.Equal(t.t, t.version, request.Version)

	for _, o := range options {
		if o.StdOut != nil {
			_, err := o.StdOut.Write([]byte(t.resp))
			assert.NoError(t.t, err)
		}
	}

	return t.expectedError

}

func TestTerraformInit_RequestValidation(t *testing.T) {
	defaultArgs := []terraform.Argument{
		{
			Key:   "input",
			Value: "false",
		},
	}
	defaultVersion := "1.0.2"

	cases := []struct {
		RequestVersion  string
		ExpectedVersion string
		RequestArgs     []terraform.Argument
		ExpectedArgs    []terraform.Argument
	}{
		{
			RequestVersion:  "0.12.0",
			ExpectedVersion: "0.12.0",
			ExpectedArgs:    defaultArgs,
		},
		{
			ExpectedArgs: []terraform.Argument{
				{
					Key:   "input",
					Value: "true",
				},
			},
			RequestArgs: []terraform.Argument{
				{
					Key:   "input",
					Value: "true",
				},
			},
			ExpectedVersion: defaultVersion,
		},
	}

	for _, c := range cases {
		t.Run("request param takes precedence", func(t *testing.T) {
			ts := testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()

			path := "some/path"
			jobID := "1234"

			expectedVersion, err := version.NewVersion(c.ExpectedVersion)
			assert.Nil(t, err)

			testTfClient := &testTfClient{
				t:             t,
				jobID:         jobID,
				path:          path,
				cmd:           terraform.NewSubCommand(terraform.Init).WithArgs(c.ExpectedArgs...),
				customEnvVars: map[string]string{},
				version:       expectedVersion,
				resp:          "",
			}

			req := TerraformInitRequest{
				Envs:      map[string]string{},
				JobID:     jobID,
				Path:      path,
				TfVersion: c.RequestVersion,
				Args:      c.RequestArgs,
			}

			tfActivity := NewTerraformActivities(testTfClient, expectedVersion, &testStreamHandler{
				t: t,
			})
			env.RegisterActivity(tfActivity)

			_, err = env.ExecuteActivity(tfActivity.TerraformInit, req)
			assert.NoError(t, err)
		})
	}
}

func TestTerraformInit_StreamsOutput(t *testing.T) {
	defaultArgs := []terraform.Argument{
		{
			Key:   "input",
			Value: "false",
		},
	}
	defaultVersion := "1.0.2"

	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestActivityEnvironment()

	path := "some/path"
	jobID := "1234"

	expectedMsgs := []string{"msg1", "msg2"}
	expectedMsgStr := strings.Join(expectedMsgs, "\n")

	expectedVersion, err := version.NewVersion(defaultVersion)
	assert.NoError(t, err)

	testTfClient := &testTfClient{
		t:             t,
		jobID:         jobID,
		path:          path,
		cmd:           terraform.NewSubCommand(terraform.Init).WithArgs(defaultArgs...),
		customEnvVars: map[string]string{},
		version:       expectedVersion,
		resp:          expectedMsgStr,
	}

	req := TerraformInitRequest{
		Envs:  map[string]string{},
		JobID: jobID,
		Path:  path,
	}

	streamHandler := &testStreamHandler{
		t:             t,
		received:      expectedMsgs,
		expectedJobID: jobID,
	}

	tfActivity := NewTerraformActivities(testTfClient, expectedVersion, streamHandler)
	env.RegisterActivity(tfActivity)

	_, err = env.ExecuteActivity(tfActivity.TerraformInit, req)
	assert.NoError(t, err)

	assert.True(t, streamHandler.called)
}

func TestTerraformPlan_RequestValidation(t *testing.T) {
	defaultArgs := []terraform.Argument{
		{
			Key:   "input",
			Value: "false",
		}, {
			Key:   "refresh",
			Value: "true",
		}, {
			Key:   "out",
			Value: "some/path/output.tfplan",
		}}
	defaultVersion := "1.0.2"

	cases := []struct {
		RequestVersion  string
		ExpectedVersion string
		RequestArgs     []terraform.Argument
		ExpectedArgs    []terraform.Argument
	}{
		{
			RequestVersion:  "0.12.0",
			ExpectedVersion: "0.12.0",
			ExpectedArgs:    defaultArgs,
		},
		{
			ExpectedArgs: []terraform.Argument{
				{
					Key:   "input",
					Value: "true",
				}, {
					Key:   "refresh",
					Value: "true",
				}, {
					Key:   "out",
					Value: "some/path/output.tfplan",
				}},
			RequestArgs: []terraform.Argument{
				{
					Key:   "input",
					Value: "true",
				},
			},
			ExpectedVersion: defaultVersion,
		},
	}

	for _, c := range cases {
		t.Run("request param takes precedence", func(t *testing.T) {
			ts := testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()

			path := "some/path"
			jobID := "1234"

			expectedVersion, err := version.NewVersion(c.ExpectedVersion)
			assert.Nil(t, err)

			testTfClient := multiCallTfClient{
				clients: []*testTfClient{
					{
						t:             t,
						jobID:         jobID,
						path:          path,
						cmd:           terraform.NewSubCommand(terraform.Plan).WithArgs(c.ExpectedArgs...),
						customEnvVars: map[string]string{},
						version:       expectedVersion,
						resp:          "",
					},
					{
						t:             t,
						jobID:         jobID,
						path:          path,
						cmd:           terraform.NewSubCommand(terraform.Show).WithFlags(terraform.Flag{Value: "json"}).WithInput("some/path/output.tfplan"),
						customEnvVars: map[string]string{},
						version:       expectedVersion,
						resp:          "{}",
					},
				},
			}

			req := TerraformPlanRequest{
				Envs:      map[string]string{},
				JobID:     jobID,
				Path:      path,
				TfVersion: c.RequestVersion,
				Args:      c.RequestArgs,
			}

			tfActivity := NewTerraformActivities(&testTfClient, expectedVersion, &testStreamHandler{
				t: t,
			})
			env.RegisterActivity(tfActivity)

			_, err = env.ExecuteActivity(tfActivity.TerraformPlan, req)
			assert.NoError(t, err)
			assert.NoError(t, testTfClient.AssertExpectations())
		})
	}
}

func TestTerraformPlan_ReturnsResponse(t *testing.T) {
	defaultArgs := []terraform.Argument{
		{
			Key:   "input",
			Value: "false",
		}, {
			Key:   "refresh",
			Value: "true",
		}, {
			Key:   "out",
			Value: "some/path/output.tfplan",
		}}
	defaultVersion := "1.0.2"

	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestActivityEnvironment()

	path := "some/path"
	jobID := "1234"

	expectedMsgs := []string{"msg1", "msg2"}
	expectedMsgStr := strings.Join(expectedMsgs, "\n")

	expectedVersion, err := version.NewVersion(defaultVersion)
	assert.Nil(t, err)

	testTfClient := multiCallTfClient{
		clients: []*testTfClient{
			{
				t:             t,
				jobID:         jobID,
				path:          path,
				cmd:           terraform.NewSubCommand(terraform.Plan).WithArgs(defaultArgs...),
				customEnvVars: map[string]string{},
				version:       expectedVersion,
				resp:          expectedMsgStr,
			},
			{
				t:             t,
				jobID:         jobID,
				path:          path,
				cmd:           terraform.NewSubCommand(terraform.Show).WithFlags(terraform.Flag{Value: "json"}).WithInput("some/path/output.tfplan"),
				customEnvVars: map[string]string{},
				version:       expectedVersion,
				resp:          "{\"format_version\": \"1.0\",\"resource_changes\":[{\"change\":{\"actions\":[\"update\"]},\"address\":\"type.resource\"}]}",
			},
		},
	}

	req := TerraformPlanRequest{
		Envs:  map[string]string{},
		JobID: jobID,
		Path:  path,
	}

	streamHandler := &testStreamHandler{
		t:             t,
		received:      expectedMsgs,
		expectedJobID: jobID,
	}

	tfActivity := NewTerraformActivities(&testTfClient, expectedVersion, streamHandler)

	env.RegisterActivity(tfActivity)

	result, err := env.ExecuteActivity(tfActivity.TerraformPlan, req)
	assert.NoError(t, err)
	assert.NoError(t, testTfClient.AssertExpectations())

	var resp TerraformPlanResponse
	assert.NoError(t, result.Get(&resp))

	assert.Equal(t, TerraformPlanResponse{
		PlanFile: "some/path/output.tfplan",
		Summary: terraform.PlanSummary{
			Updates: []terraform.ResourceSummary{
				{
					Address: "type.resource",
				},
			},
		},
	}, resp)

	assert.True(t, streamHandler.called)
}

func TestTerraformApply_RequestValidation(t *testing.T) {
	defaultArgs := []terraform.Argument{
		{
			Key:   "input",
			Value: "false",
		},
	}
	defaultVersion := "1.0.2"

	cases := []struct {
		RequestVersion  string
		ExpectedVersion string
		RequestArgs     []terraform.Argument
		ExpectedArgs    []terraform.Argument
	}{
		{
			RequestVersion:  "0.12.0",
			ExpectedVersion: "0.12.0",
			ExpectedArgs:    defaultArgs,
		},
		{
			ExpectedArgs: []terraform.Argument{
				{
					Key:   "input",
					Value: "false",
				}},
			RequestArgs: []terraform.Argument{
				{
					Key:   "input",
					Value: "false",
				},
			},
			ExpectedVersion: defaultVersion,
		},
	}

	for _, c := range cases {
		t.Run("request param takes precedence", func(t *testing.T) {
			ts := testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()

			path := "some/path"
			jobID := "1234"

			expectedVersion, err := version.NewVersion(c.ExpectedVersion)
			assert.Nil(t, err)

			testClient := &testTfClient{
				t:             t,
				jobID:         jobID,
				path:          path,
				cmd:           terraform.NewSubCommand(terraform.Apply).WithArgs(c.ExpectedArgs...).WithInput("some/path/output.tfplan"),
				customEnvVars: map[string]string{},
				version:       expectedVersion,
				resp:          "",
			}

			req := TerraformApplyRequest{
				Envs:      map[string]string{},
				JobID:     jobID,
				Path:      path,
				TfVersion: c.RequestVersion,
				Args:      c.RequestArgs,
				PlanFile:  "some/path/output.tfplan",
			}

			tfActivity := NewTerraformActivities(testClient, expectedVersion, &testStreamHandler{
				t: t,
			})
			env.RegisterActivity(tfActivity)

			_, err = env.ExecuteActivity(tfActivity.TerraformApply, req)
			assert.NoError(t, err)
		})
	}
}

func TestTerraformApply_StreamsOutput(t *testing.T) {
	defaultArgs := []terraform.Argument{
		{
			Key:   "input",
			Value: "false",
		},
	}
	defaultVersion := "1.0.2"

	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestActivityEnvironment()

	path := "some/path"
	jobID := "1234"

	expectedMsgs := []string{"msg1", "msg2"}
	expectedMsgStr := strings.Join(expectedMsgs, "\n")

	expectedVersion, err := version.NewVersion(defaultVersion)
	assert.NoError(t, err)

	testTfClient := &testTfClient{
		t:             t,
		jobID:         jobID,
		path:          path,
		cmd:           terraform.NewSubCommand(terraform.Apply).WithArgs(defaultArgs...).WithInput("some/path/output.tfplan"),
		customEnvVars: map[string]string{},
		version:       expectedVersion,
		resp:          expectedMsgStr,
	}

	req := TerraformApplyRequest{
		Envs:     map[string]string{},
		JobID:    jobID,
		Path:     path,
		PlanFile: "some/path/output.tfplan",
	}

	streamHandler := &testStreamHandler{
		t:             t,
		received:      expectedMsgs,
		expectedJobID: jobID,
	}

	tfActivity := NewTerraformActivities(testTfClient, expectedVersion, streamHandler)
	env.RegisterActivity(tfActivity)

	_, err = env.ExecuteActivity(tfActivity.TerraformApply, req)
	assert.NoError(t, err)

	assert.True(t, streamHandler.called)
}
