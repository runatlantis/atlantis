package notifier_test

import (
	"context"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/neptune/workflows/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/notifier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

type request struct {
	Requests []struct {
		DeploymentID string
		Request      notifier.GithubCheckRunRequest
	}
}

type response struct {
	IDs []int64
}

type testActivities struct{}

func (a testActivities) GithubUpdateCheckRun(ctx context.Context, request activities.UpdateCheckRunRequest) (activities.UpdateCheckRunResponse, error) {
	return activities.UpdateCheckRunResponse{}, nil
}

func (a testActivities) GithubCreateCheckRun(ctx context.Context, request activities.CreateCheckRunRequest) (activities.CreateCheckRunResponse, error) {
	return activities.CreateCheckRunResponse{}, nil
}

func testWorkflow(ctx workflow.Context, workflowRequest request) (response, error) {
	ctx = workflow.WithStartToCloseTimeout(ctx, 5*time.Second)
	var a testActivities
	c := notifier.NewGithubCheckRunCache(a)

	var ids []int64
	for _, r := range workflowRequest.Requests {
		id, err := c.CreateOrUpdate(ctx, r.DeploymentID, r.Request)

		if err != nil {
			return response{}, err
		}

		ids = append(ids, id)
	}
	return response{
		IDs: ids,
	}, nil
}

func TestGithubCheckRunCache_CreatesThenUpdates(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	a := &testActivities{}

	env.RegisterActivity(a)

	testRequest1 := notifier.GithubCheckRunRequest{
		Title: "title",
		Sha:   "alskdjf",
		Repo: github.Repo{
			Name: "repo",
		},
		State: github.CheckRunQueued,
		Actions: []github.CheckRunAction{
			github.CreatePlanReviewAction(github.Approve),
		},
		Summary: "some summary",
	}

	testRequest2 := notifier.GithubCheckRunRequest{
		Title: "title",
		Sha:   "alskdjf",
		Repo: github.Repo{
			Name: "repo",
		},
		State:   github.CheckRunSuccess,
		Summary: "some summary 2",
	}

	env.OnActivity(a.GithubCreateCheckRun, mock.Anything, activities.CreateCheckRunRequest{
		Title:      testRequest1.Title,
		Sha:        testRequest1.Sha,
		Repo:       testRequest1.Repo,
		State:      testRequest1.State,
		Actions:    testRequest1.Actions,
		Summary:    testRequest1.Summary,
		ExternalID: "1234",
	}).Return(activities.CreateCheckRunResponse{
		ID: 1,
	}, nil)

	env.OnActivity(a.GithubUpdateCheckRun, mock.Anything, activities.UpdateCheckRunRequest{
		Title:      testRequest2.Title,
		ID:         1,
		Repo:       testRequest2.Repo,
		State:      testRequest2.State,
		Actions:    testRequest2.Actions,
		Summary:    testRequest2.Summary,
		ExternalID: "1234",
	}).Return(activities.UpdateCheckRunResponse{
		ID:     1,
		Status: "completed",
	}, nil)

	env.ExecuteWorkflow(testWorkflow, request{
		Requests: []struct {
			DeploymentID string
			Request      notifier.GithubCheckRunRequest
		}{
			{
				DeploymentID: "1234",
				Request:      testRequest1,
			},
			{
				DeploymentID: "1234",
				Request:      testRequest2,
			},
		},
	})

	env.AssertExpectations(t)

	var r response
	err := env.GetWorkflowResult(&r)
	assert.NoError(t, err)

	assert.Equal(t, r.IDs, []int64{1, 1})
}

func TestGithubCheckRunCache_CreatesAcrossDeploymentIDs(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	a := &testActivities{}

	env.RegisterActivity(a)

	testRequest1 := notifier.GithubCheckRunRequest{
		Title: "title",
		Sha:   "alskdjf",
		Repo: github.Repo{
			Name: "repo",
		},
		State:   github.CheckRunQueued,
		Summary: "some summary",
	}

	testRequest2 := notifier.GithubCheckRunRequest{
		Title: "title",
		Sha:   "alskdjf",
		Repo: github.Repo{
			Name: "repo",
		},
		State:   github.CheckRunQueued,
		Summary: "some summary",
	}

	env.OnActivity(a.GithubCreateCheckRun, mock.Anything, activities.CreateCheckRunRequest{
		Title:      testRequest1.Title,
		Sha:        testRequest1.Sha,
		Repo:       testRequest1.Repo,
		State:      testRequest1.State,
		Actions:    testRequest1.Actions,
		Summary:    testRequest1.Summary,
		ExternalID: "1234",
	}).Return(activities.CreateCheckRunResponse{
		ID: 1,
	}, nil).Once()

	env.OnActivity(a.GithubCreateCheckRun, mock.Anything, activities.CreateCheckRunRequest{
		Title:      testRequest1.Title,
		Sha:        testRequest1.Sha,
		Repo:       testRequest1.Repo,
		State:      testRequest1.State,
		Actions:    testRequest1.Actions,
		Summary:    testRequest1.Summary,
		ExternalID: "12345",
	}).Return(activities.CreateCheckRunResponse{
		ID: 2,
	}, nil).Once()

	env.ExecuteWorkflow(testWorkflow, request{
		Requests: []struct {
			DeploymentID string
			Request      notifier.GithubCheckRunRequest
		}{
			{
				DeploymentID: "1234",
				Request:      testRequest1,
			},
			{
				DeploymentID: "12345",
				Request:      testRequest2,
			},
		},
	})

	env.AssertExpectations(t)

	var r response
	err := env.GetWorkflowResult(&r)
	assert.NoError(t, err)

	assert.Equal(t, []int64{1, 2}, r.IDs)
}

func TestGithubCheckRunCache_CreatesCompleted(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	a := &testActivities{}

	env.RegisterActivity(a)

	testRequest1 := notifier.GithubCheckRunRequest{
		Title: "title",
		Sha:   "alskdjf",
		Repo: github.Repo{
			Name: "repo",
		},
		State:   github.CheckRunQueued,
		Summary: "some summary",
	}

	testRequest2 := notifier.GithubCheckRunRequest{
		Title: "title",
		Sha:   "alskdjf",
		Repo: github.Repo{
			Name: "repo",
		},
		State:   github.CheckRunQueued,
		Summary: "some summary",
	}

	var id int64
	env.OnActivity(a.GithubCreateCheckRun, mock.Anything, activities.CreateCheckRunRequest{
		Title:      testRequest1.Title,
		Sha:        testRequest1.Sha,
		Repo:       testRequest1.Repo,
		State:      testRequest1.State,
		Actions:    testRequest1.Actions,
		Summary:    testRequest1.Summary,
		ExternalID: "1234",
	}).Return(func(ctx context.Context, r activities.CreateCheckRunRequest) (activities.CreateCheckRunResponse, error) {
		id++
		return activities.CreateCheckRunResponse{
			ID:     id,
			Status: "completed",
		}, nil
	}).Twice()

	env.ExecuteWorkflow(testWorkflow, request{
		Requests: []struct {
			DeploymentID string
			Request      notifier.GithubCheckRunRequest
		}{
			{
				DeploymentID: "1234",
				Request:      testRequest1,
			},
			{
				DeploymentID: "1234",
				Request:      testRequest2,
			},
		},
	})

	env.AssertExpectations(t)

	var r response
	err := env.GetWorkflowResult(&r)
	assert.NoError(t, err)

	assert.Equal(t, []int64{1, 2}, r.IDs)
}

func TestGithubCheckRunCache_CreatesAfterCompleting(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	a := &testActivities{}

	env.RegisterActivity(a)

	testRequest1 := notifier.GithubCheckRunRequest{
		Title: "title",
		Sha:   "alskdjf",
		Repo: github.Repo{
			Name: "repo",
		},
		State:   github.CheckRunQueued,
		Summary: "some summary",
	}

	testRequest2 := notifier.GithubCheckRunRequest{
		Title: "title",
		Sha:   "alskdjf",
		Repo: github.Repo{
			Name: "repo",
		},
		Actions: []github.CheckRunAction{
			github.CreatePlanReviewAction(github.Approve),
		},
		State:   github.CheckRunActionRequired,
		Summary: "some summary 2",
	}

	testRequest3 := notifier.GithubCheckRunRequest{
		Title: "title",
		Sha:   "alskdjf",
		Repo: github.Repo{
			Name: "repo",
		},
		State:   github.CheckRunPending,
		Summary: "some summary",
	}

	env.OnActivity(a.GithubCreateCheckRun, mock.Anything, activities.CreateCheckRunRequest{
		Title:      testRequest1.Title,
		Sha:        testRequest1.Sha,
		Repo:       testRequest1.Repo,
		State:      testRequest1.State,
		Actions:    testRequest1.Actions,
		Summary:    testRequest1.Summary,
		ExternalID: "1234",
	}).Return(activities.CreateCheckRunResponse{
		ID: 1,
	}, nil)

	env.OnActivity(a.GithubUpdateCheckRun, mock.Anything, activities.UpdateCheckRunRequest{
		Title:      testRequest2.Title,
		ID:         1,
		Repo:       testRequest2.Repo,
		State:      testRequest2.State,
		Actions:    testRequest2.Actions,
		Summary:    testRequest2.Summary,
		ExternalID: "1234",
	}).Return(activities.UpdateCheckRunResponse{
		ID:     1,
		Status: "completed",
	}, nil)

	env.OnActivity(a.GithubCreateCheckRun, mock.Anything, activities.CreateCheckRunRequest{
		Title:      testRequest3.Title,
		Sha:        testRequest3.Sha,
		Repo:       testRequest3.Repo,
		State:      testRequest3.State,
		Actions:    testRequest3.Actions,
		Summary:    testRequest3.Summary,
		ExternalID: "1234",
	}).Return(activities.CreateCheckRunResponse{
		ID: 2,
	}, nil)

	env.ExecuteWorkflow(testWorkflow, request{
		Requests: []struct {
			DeploymentID string
			Request      notifier.GithubCheckRunRequest
		}{
			{
				DeploymentID: "1234",
				Request:      testRequest1,
			},
			{
				DeploymentID: "1234",
				Request:      testRequest2,
			},
			{
				DeploymentID: "1234",
				Request:      testRequest3,
			},
		},
	})

	env.AssertExpectations(t)

	var r response
	err := env.GetWorkflowResult(&r)
	assert.NoError(t, err)

	assert.Equal(t, []int64{1, 1, 2}, r.IDs)
}
