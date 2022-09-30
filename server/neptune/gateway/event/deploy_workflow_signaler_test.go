package event_test

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/neptune/gateway/event"
	"github.com/runatlantis/atlantis/server/neptune/workflows"
	"github.com/runatlantis/atlantis/server/vcs"
	"github.com/stretchr/testify/assert"
	"go.temporal.io/sdk/client"
	"testing"
)

type testRun struct{}

func (r testRun) GetID() string {
	return "123"
}

func (r testRun) GetRunID() string {
	return "456"
}

func (r testRun) Get(ctx context.Context, valuePtr interface{}) error {
	return nil
}

func (r testRun) GetWithOptions(ctx context.Context, valuePtr interface{}, options client.WorkflowRunGetOptions) error {
	return nil
}

type testSignaler struct {
	t                    *testing.T
	expectedWorkflowID   string
	expectedRunID        string
	expectedSignalName   string
	expectedSignalArg    interface{}
	expectedOptions      client.StartWorkflowOptions
	expectedWorkflow     interface{}
	expectedWorkflowArgs interface{}
	expectedErr          error

	called bool
}

func (s *testSignaler) SignalWorkflow(ctx context.Context, workflowID string, runID string, signalName string, arg interface{}) error {
	s.called = true
	assert.Equal(s.t, s.expectedWorkflowID, workflowID)
	assert.Equal(s.t, s.expectedRunID, runID)
	assert.Equal(s.t, s.expectedSignalName, signalName)
	assert.Equal(s.t, s.expectedSignalArg, arg)

	return s.expectedErr
}

func (s *testSignaler) SignalWithStartWorkflow(ctx context.Context, workflowID string, signalName string, signalArg interface{},
	options client.StartWorkflowOptions, workflow interface{}, workflowArgs ...interface{}) (client.WorkflowRun, error) {

	s.called = true

	assert.Equal(s.t, s.expectedWorkflowID, workflowID)
	assert.Equal(s.t, s.expectedSignalName, signalName)
	assert.Equal(s.t, s.expectedSignalArg, signalArg)
	assert.Equal(s.t, s.expectedOptions, options)
	assert.IsType(s.t, s.expectedWorkflow, workflow)
	assert.Equal(s.t, []interface{}{s.expectedWorkflowArgs}, workflowArgs)

	return testRun{}, s.expectedErr
}

func TestSignalWithStartWorkflow_Success(t *testing.T) {
	repoFullName := "nish/repo"
	repoOwner := "nish"
	repoName := "repo"
	repoURL := "www.nish.com"
	sha := "12345"
	ref := vcs.Ref{
		Type: vcs.BranchRef,
		Name: "main",
	}

	repo := models.Repo{
		FullName: repoFullName,
		Owner:    repoOwner,
		Name:     repoName,
		CloneURL: repoURL,
	}

	version, err := version.NewVersion("1.0.3")
	assert.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		rootCfg := valid.MergedProjectCfg{
			Name: testRoot,
			DeploymentWorkflow: valid.Workflow{
				Plan:  valid.DefaultPlanStage,
				Apply: valid.DefaultApplyStage,
			},
			TerraformVersion: version,
		}

		testSignaler := &testSignaler{
			t:                  t,
			expectedWorkflowID: fmt.Sprintf("%s||%s", repoFullName, testRoot),
			expectedSignalName: workflows.DeployNewRevisionSignalID,
			expectedSignalArg: workflows.DeployNewRevisionSignalRequest{
				Revision: sha,
				Root: workflows.Root{
					Name: testRoot,
					Plan: workflows.Job{
						Steps: convertTestSteps(valid.DefaultPlanStage.Steps),
					},
					Apply: workflows.Job{
						Steps: convertTestSteps(valid.DefaultApplyStage.Steps),
					},
					TfVersion: version.String(),
					PlanMode:  workflows.NormalPlanMode,
					Trigger:   workflows.MergeTrigger,
				},
			},
			expectedWorkflow: workflows.Deploy,
			expectedOptions: client.StartWorkflowOptions{
				TaskQueue: workflows.DeployTaskQueue,
			},
			expectedWorkflowArgs: workflows.DeployRequest{
				Repository: workflows.Repo{
					FullName: repoFullName,
					Name:     repoName,
					Owner:    repoOwner,
					URL:      repoURL,
					HeadCommit: workflows.HeadCommit{
						Ref: workflows.Ref{
							Name: ref.Name,
							Type: string(ref.Type),
						},
					},
				},
			},
		}
		deploySignaler := event.DeployWorkflowSignaler{
			TemporalClient: testSignaler,
		}
		run, err := deploySignaler.SignalWithStartWorkflow(context.Background(), &rootCfg, repo, sha, 0, ref, workflows.MergeTrigger)
		assert.NoError(t, err)
		assert.Equal(t, testRun{}, run)
	})

	t.Run("success w/destroy", func(t *testing.T) {
		rootCfg := valid.MergedProjectCfg{
			Name: testRoot,
			DeploymentWorkflow: valid.Workflow{
				Plan:  valid.DefaultPlanStage,
				Apply: valid.DefaultApplyStage,
			},
			Tags: map[string]string{
				event.Deprecated: event.Destroy,
			},
			TerraformVersion: version,
		}

		testSignaler := &testSignaler{
			t:                  t,
			expectedWorkflowID: fmt.Sprintf("%s||%s", repoFullName, testRoot),
			expectedSignalName: workflows.DeployNewRevisionSignalID,
			expectedSignalArg: workflows.DeployNewRevisionSignalRequest{
				Revision: sha,
				Root: workflows.Root{
					Name: testRoot,
					Plan: workflows.Job{
						Steps: convertTestSteps(valid.DefaultPlanStage.Steps),
					},
					Apply: workflows.Job{
						Steps: convertTestSteps(valid.DefaultApplyStage.Steps),
					},
					TfVersion: version.String(),
					PlanMode:  workflows.DestroyPlanMode,
					Trigger:   workflows.MergeTrigger,
				},
			},
			expectedWorkflow: workflows.Deploy,
			expectedOptions: client.StartWorkflowOptions{
				TaskQueue: workflows.DeployTaskQueue,
			},
			expectedWorkflowArgs: workflows.DeployRequest{
				Repository: workflows.Repo{
					FullName: repoFullName,
					Name:     repoName,
					Owner:    repoOwner,
					URL:      repoURL,
					HeadCommit: workflows.HeadCommit{
						Ref: workflows.Ref{
							Name: ref.Name,
							Type: string(ref.Type),
						},
					},
				},
			},
		}
		deploySignaler := event.DeployWorkflowSignaler{
			TemporalClient: testSignaler,
		}
		run, err := deploySignaler.SignalWithStartWorkflow(context.Background(), &rootCfg, repo, sha, 0, ref, workflows.MergeTrigger)
		assert.NoError(t, err)
		assert.Equal(t, testRun{}, run)
	})
}

func TestSignalWithStartWorkflow_Failure(t *testing.T) {
	repoFullName := "nish/repo"
	repoOwner := "nish"
	repoName := "repo"
	repoURL := "www.nish.com"
	sha := "12345"
	ref := vcs.Ref{
		Type: vcs.BranchRef,
		Name: "main",
	}

	repo := models.Repo{
		FullName: repoFullName,
		Owner:    repoOwner,
		Name:     repoName,
		CloneURL: repoURL,
	}

	version, err := version.NewVersion("1.0.3")
	assert.NoError(t, err)
	rootCfg := valid.MergedProjectCfg{
		Name: testRoot,
		DeploymentWorkflow: valid.Workflow{
			Plan:  valid.DefaultPlanStage,
			Apply: valid.DefaultApplyStage,
		},
		TerraformVersion: version,
	}

	testSignaler := &testSignaler{
		t:                  t,
		expectedWorkflowID: fmt.Sprintf("%s||%s", repoFullName, testRoot),
		expectedSignalName: workflows.DeployNewRevisionSignalID,
		expectedSignalArg: workflows.DeployNewRevisionSignalRequest{
			Revision: sha,
			Root: workflows.Root{
				Name: testRoot,
				Plan: workflows.Job{
					Steps: convertTestSteps(valid.DefaultPlanStage.Steps),
				},
				Apply: workflows.Job{
					Steps: convertTestSteps(valid.DefaultApplyStage.Steps),
				},
				TfVersion: version.String(),
				PlanMode:  workflows.NormalPlanMode,
				Trigger:   workflows.MergeTrigger,
			},
		},
		expectedWorkflow: workflows.Deploy,
		expectedOptions: client.StartWorkflowOptions{
			TaskQueue: workflows.DeployTaskQueue,
		},
		expectedWorkflowArgs: workflows.DeployRequest{
			Repository: workflows.Repo{
				FullName: repoFullName,
				Name:     repoName,
				Owner:    repoOwner,
				URL:      repoURL,
				HeadCommit: workflows.HeadCommit{
					Ref: workflows.Ref{
						Name: ref.Name,
						Type: string(ref.Type),
					},
				},
			},
		},
		expectedErr: expectedErr,
	}
	deploySignaler := event.DeployWorkflowSignaler{
		TemporalClient: testSignaler,
	}
	run, err := deploySignaler.SignalWithStartWorkflow(context.Background(), &rootCfg, repo, sha, 0, ref, workflows.MergeTrigger)
	assert.Error(t, err)
	assert.Equal(t, testRun{}, run)
}
