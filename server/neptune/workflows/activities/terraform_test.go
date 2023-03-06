package activities

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/file"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
	"github.com/stretchr/testify/assert"
	"go.temporal.io/sdk/testsuite"
)

type testCredsRefresher struct {
	called                 bool
	expectedInstallationID int64
	t                      *testing.T
}

func (t *testCredsRefresher) Refresh(ctx context.Context, installationID int64) error {
	assert.Equal(t.t, t.expectedInstallationID, installationID)
	t.called = true
	return nil
}

type testStreamHandler struct {
	received      []string
	expectedJobID string
	t             *testing.T
	called        bool
	wg            sync.WaitGroup
}

func (t *testStreamHandler) RegisterJob(id string) chan string {
	ch := make(chan string)
	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		for s := range ch {
			t.received = append(t.received, s)
		}
		t.called = true
	}()
	return ch
}

func (t *testStreamHandler) Wait() {
	t.wg.Wait()
}

type multiCallTfClient struct {
	clients []*testTfClient

	count int
}

func (t *multiCallTfClient) RunCommand(ctx context.Context, request *terraform.RunCommandRequest, options ...terraform.RunOptions) error {
	if t.count >= len(t.clients) {
		return fmt.Errorf("expected less calls to RunCommand")
	}
	_ = t.clients[t.count].RunCommand(ctx, request, options...)

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
		Envs            map[string]string
		ExpectedEnvs    map[string]string
		DynamicEnvs     []EnvVar
		ExpectedArgs    []terraform.Argument
	}{
		{
			//testing
			RequestVersion:  "0.12.0",
			ExpectedVersion: "0.12.0",

			//defaults
			ExpectedArgs: defaultArgs,
			Envs:         map[string]string{},
			ExpectedEnvs: map[string]string{},
		},
		{
			//testing
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

			// defaults
			ExpectedVersion: defaultVersion,
			Envs:            map[string]string{},
			ExpectedEnvs:    map[string]string{},
		},
		{
			// testing
			Envs: map[string]string{"env1": "val1"},
			DynamicEnvs: []EnvVar{
				{
					Name:  "env2",
					Value: "val2",
				},
			},
			ExpectedEnvs: map[string]string{
				"env1": "val1",
				"env2": "val2",
			},

			// defaults
			ExpectedVersion: defaultVersion,
			ExpectedArgs:    defaultArgs,
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
				customEnvVars: c.ExpectedEnvs,
				version:       expectedVersion,
				resp:          "",
			}

			req := TerraformInitRequest{
				Envs:                 c.Envs,
				DynamicEnvs:          c.DynamicEnvs,
				JobID:                jobID,
				Path:                 path,
				TfVersion:            c.RequestVersion,
				Args:                 c.RequestArgs,
				GithubInstallationID: 1235,
			}

			credsRefresher := &testCredsRefresher{
				expectedInstallationID: 1235,
				t:                      t,
			}

			tfActivity := NewTerraformActivities(
				testTfClient,
				expectedVersion,
				&testStreamHandler{
					t: t,
				},
				credsRefresher,
				&file.RWLock{})
			env.RegisterActivity(tfActivity)

			_, err = env.ExecuteActivity(tfActivity.TerraformInit, req)
			assert.NoError(t, err)

			assert.True(t, credsRefresher.called)
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
		Envs:                 map[string]string{},
		JobID:                jobID,
		Path:                 path,
		GithubInstallationID: 1235,
	}

	streamHandler := &testStreamHandler{
		t:             t,
		received:      expectedMsgs,
		expectedJobID: jobID,
	}

	credsRefresher := &testCredsRefresher{
		expectedInstallationID: 1235,
		t:                      t,
	}

	tfActivity := NewTerraformActivities(testTfClient, expectedVersion, streamHandler, credsRefresher, &file.RWLock{})
	env.RegisterActivity(tfActivity)

	_, err = env.ExecuteActivity(tfActivity.TerraformInit, req)
	assert.NoError(t, err)

	// wait before we check called value otherwise we might race
	streamHandler.Wait()
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
		ExpectedFlags   []terraform.Flag
		PlanMode        *terraform.PlanMode
		Envs            map[string]string
		ExpectedEnvs    map[string]string
		DynamicEnvs     []EnvVar
	}{
		{
			//testing
			RequestVersion:  "0.12.0",
			ExpectedVersion: "0.12.0",

			//default
			ExpectedArgs: defaultArgs,
			Envs:         map[string]string{},
			ExpectedEnvs: map[string]string{},
		},
		{
			//testing
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

			// default
			ExpectedVersion: defaultVersion,
			Envs:            map[string]string{},
			ExpectedEnvs:    map[string]string{},
		},
		{
			// testing
			PlanMode: terraform.NewDestroyPlanMode(),
			ExpectedFlags: []terraform.Flag{
				{
					Value: "destroy",
				},
			},

			// default
			ExpectedArgs:    defaultArgs,
			ExpectedVersion: defaultVersion,
			Envs:            map[string]string{},
			ExpectedEnvs:    map[string]string{},
		},
		{
			// testing
			Envs: map[string]string{"env1": "val1"},
			DynamicEnvs: []EnvVar{
				{
					Name:  "env2",
					Value: "val2",
				},
			},
			ExpectedEnvs: map[string]string{
				"env1": "val1",
				"env2": "val2",
			},

			// default
			ExpectedArgs:    defaultArgs,
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
						cmd:           terraform.NewSubCommand(terraform.Plan).WithArgs(c.ExpectedArgs...).WithFlags(c.ExpectedFlags...),
						customEnvVars: c.ExpectedEnvs,
						version:       expectedVersion,
						resp:          "",
					},
					{
						t:             t,
						jobID:         jobID,
						path:          path,
						cmd:           terraform.NewSubCommand(terraform.Show).WithFlags(terraform.Flag{Value: "json"}).WithInput("some/path/output.tfplan"),
						customEnvVars: c.ExpectedEnvs,
						version:       expectedVersion,
						resp:          "{}",
					},
				},
			}

			req := TerraformPlanRequest{
				Envs:        c.Envs,
				DynamicEnvs: c.DynamicEnvs,
				JobID:       jobID,
				Path:        path,
				TfVersion:   c.RequestVersion,
				Args:        c.RequestArgs,
				Mode:        c.PlanMode,
			}

			credsRefresher := &testCredsRefresher{}

			tfActivity := NewTerraformActivities(&testTfClient, expectedVersion, &testStreamHandler{
				t: t,
			}, credsRefresher, &file.RWLock{})
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

	credsRefresher := &testCredsRefresher{}

	tfActivity := NewTerraformActivities(&testTfClient, expectedVersion, streamHandler, credsRefresher, &file.RWLock{})

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

	// wait before we check called value otherwise we might race
	streamHandler.Wait()
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
		Envs            map[string]string
		ExpectedEnvs    map[string]string
		DynamicEnvs     []EnvVar
	}{
		{
			//testing
			RequestVersion:  "0.12.0",
			ExpectedVersion: "0.12.0",

			//default
			ExpectedArgs: defaultArgs,
			Envs:         map[string]string{},
			ExpectedEnvs: map[string]string{},
		},
		{
			//testing
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
			//default
			ExpectedVersion: defaultVersion,
			Envs:            map[string]string{},
			ExpectedEnvs:    map[string]string{},
		},
		{
			//testing
			Envs: map[string]string{"env1": "val1"},
			DynamicEnvs: []EnvVar{
				{
					Name:  "env2",
					Value: "val2",
				},
			},
			ExpectedEnvs: map[string]string{
				"env1": "val1",
				"env2": "val2",
			},

			//default
			ExpectedVersion: defaultVersion,
			ExpectedArgs:    defaultArgs,
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
				customEnvVars: c.ExpectedEnvs,
				version:       expectedVersion,
				resp:          "",
			}

			req := TerraformApplyRequest{
				Envs:        c.Envs,
				DynamicEnvs: c.DynamicEnvs,
				JobID:       jobID,
				Path:        path,
				TfVersion:   c.RequestVersion,
				Args:        c.RequestArgs,
				PlanFile:    "some/path/output.tfplan",
			}

			tfActivity := NewTerraformActivities(testClient, expectedVersion, &testStreamHandler{
				t: t,
			}, &testCredsRefresher{}, &file.RWLock{})
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

	tfActivity := NewTerraformActivities(testTfClient, expectedVersion, streamHandler, &testCredsRefresher{}, &file.RWLock{})
	env.RegisterActivity(tfActivity)

	_, err = env.ExecuteActivity(tfActivity.TerraformApply, req)
	assert.NoError(t, err)

	// wait before we check called value otherwise we might race
	streamHandler.Wait()
	assert.True(t, streamHandler.called)
}
