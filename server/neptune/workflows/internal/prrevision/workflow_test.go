package prrevision

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/core/config/raw"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

func Test_ShouldSetMinimumRevisionForPR(t *testing.T) {
	cases := []struct {
		description   string
		root          terraform.Root
		modifiedFiles []string
		shouldReabse  bool
	}{
		{
			description: "default tracked files config, root dir modified",
			root: terraform.Root{
				Path:         "test/dir1",
				TrackedFiles: raw.DefaultAutoPlanWhenModified,
			},
			modifiedFiles: []string{"test/dir1/main.tf"},
			shouldReabse:  true,
		},
		{
			description: "default tracked files config, root dir not modified",
			root: terraform.Root{
				Path:         "test/dir1",
				TrackedFiles: raw.DefaultAutoPlanWhenModified,
			},
			modifiedFiles: []string{"test/dir2/main.tf"},
			shouldReabse:  false,
		},
		{
			description: "default tracked files config, .tfvars file modified",
			root: terraform.Root{
				Path:         "test/dir1",
				TrackedFiles: raw.DefaultAutoPlanWhenModified,
			},
			modifiedFiles: []string{"test/dir1/terraform.tfvars"},
			shouldReabse:  true,
		},
		{
			description: "non default tracked files config, non root dir modified",
			root: terraform.Root{
				Path:         "test/dir1",
				TrackedFiles: []string{"**/*.tf*", "../variables.tf"},
			},
			modifiedFiles: []string{"test/variables.tf"},
			shouldReabse:  true,
		},
		{
			description: "non default tracked files config, file excluded",
			root: terraform.Root{
				Path:         "test/dir1",
				TrackedFiles: []string{"**/*.tf*", "!exclude.tf"},
			},
			modifiedFiles: []string{"test/dir1/exclude.tf"},
			shouldReabse:  false,
		},
		{
			description: "non default tracked files config, file excluded",
			root: terraform.Root{
				Path:         "test/dir1",
				TrackedFiles: []string{"**/*.tf*", "!exclude.tf"},
			},
			modifiedFiles: []string{"test/dir1/exclude.tf"},
			shouldReabse:  false,
		},
		{
			description: "non default tracked files config, file excluded and included",
			root: terraform.Root{
				Path:         "test/dir1",
				TrackedFiles: []string{"**/*.tf*", "!exclude.tf"},
			},
			modifiedFiles: []string{"test/dir1/exclude.tf", "test/dir1/main.tf"},
			shouldReabse:  true,
		},
	}

	for _, c := range cases {
		res, err := isRootModified(c.root, c.modifiedFiles)
		assert.NoError(t, err)
		assert.Equal(t, c.shouldReabse, res)
	}
}

type testRevisionSetterActivities struct{}

func (t *testRevisionSetterActivities) SetPRRevision(ctx context.Context, request activities.SetPRRevisionRequest) (activities.SetPRRevisionResponse, error) {
	return activities.SetPRRevisionResponse{}, nil
}

type testGithubActivities struct{}

func (t *testGithubActivities) ListPRs(ctx context.Context, request activities.ListPRsRequest) (activities.ListPRsResponse, error) {
	return activities.ListPRsResponse{}, nil
}

func (t *testGithubActivities) ListModifiedFiles(ctx context.Context, request activities.ListModifiedFilesRequest) (activities.ListModifiedFilesResponse, error) {
	return activities.ListModifiedFilesResponse{}, nil
}

func testSetMiminumValidRevisionForRootWorkflow(ctx workflow.Context, r Request) error {
	options := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 5 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, options)

	runner := Runner{
		GithubActivities:         &testGithubActivities{},
		RevisionSetterActivities: &testRevisionSetterActivities{},
		Scope:                    metrics.NewNullableScope(),
	}
	return runner.Run(ctx, r)
}

func TestMinRevisionSetter_NoOpenPR(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	ga := &testGithubActivities{}
	env.RegisterActivity(ga)

	req := Request{
		Repo: github.Repo{
			Owner: "owner",
			Name:  "test",
		},
		Root: terraform.Root{
			Name: "test",
		},
	}

	env.OnActivity(ga.ListPRs, mock.Anything, activities.ListPRsRequest{
		Repo:  req.Repo,
		State: github.OpenPullRequest,
	}).Return(activities.ListPRsResponse{
		PullRequests: []github.PullRequest{},
	}, nil)

	env.ExecuteWorkflow(testSetMiminumValidRevisionForRootWorkflow, req)
	env.AssertExpectations(t)

	err := env.GetWorkflowResult(nil)
	assert.Nil(t, err)
}

func TestMinRevisionSetter_OpenPR_SetMinRevision(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	ga := &testGithubActivities{}
	ra := &testRevisionSetterActivities{}
	env.RegisterActivity(ra)
	env.RegisterActivity(ga)

	req := Request{
		Repo: github.Repo{
			Owner: "owner",
			Name:  "test",
		},
		Root: terraform.Root{
			Path:         "test/dir2",
			TrackedFiles: raw.DefaultAutoPlanWhenModified,
		},
	}

	pullRequests := []github.PullRequest{
		{
			Number: 1,
		},
		{
			Number: 2,
		},
	}

	filesModifiedPr1 := []string{"test/dir2/rebase.tf"}
	filesModifiedPr2 := []string{"test/dir1/no-rebase.tf"}

	env.OnActivity(ga.ListPRs, mock.Anything, activities.ListPRsRequest{
		Repo:  req.Repo,
		State: github.OpenPullRequest,
	}).Return(activities.ListPRsResponse{
		PullRequests: pullRequests,
	}, nil)

	env.OnActivity(ga.ListModifiedFiles, mock.Anything, activities.ListModifiedFilesRequest{
		Repo:        req.Repo,
		PullRequest: pullRequests[0],
	}).Return(activities.ListModifiedFilesResponse{
		FilePaths: filesModifiedPr1,
	}, nil)

	env.OnActivity(ga.ListModifiedFiles, mock.Anything, activities.ListModifiedFilesRequest{
		Repo:        req.Repo,
		PullRequest: pullRequests[1],
	}).Return(activities.ListModifiedFilesResponse{
		FilePaths: filesModifiedPr2,
	}, nil)

	env.OnActivity(ra.SetPRRevision, mock.Anything, activities.SetPRRevisionRequest{
		Repository:  req.Repo,
		PullRequest: pullRequests[0],
	}).Return(activities.SetPRRevisionResponse{}, nil)

	env.ExecuteWorkflow(testSetMiminumValidRevisionForRootWorkflow, req)
	env.AssertExpectations(t)

	err := env.GetWorkflowResult(nil)
	assert.Nil(t, err)
}

func TestMinRevisionSetter_ListModifiedFilesErr(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	ga := &testGithubActivities{}
	ra := &testRevisionSetterActivities{}
	env.RegisterActivity(ra)
	env.RegisterActivity(ga)

	req := Request{
		Repo: github.Repo{
			Owner: "owner",
			Name:  "test",
		},
		Root: terraform.Root{
			Path:         "test/dir2",
			TrackedFiles: raw.DefaultAutoPlanWhenModified,
		},
	}

	pullRequests := []github.PullRequest{
		{
			Number: 1,
		},
	}

	env.OnActivity(ga.ListPRs, mock.Anything, activities.ListPRsRequest{
		Repo:  req.Repo,
		State: github.OpenPullRequest,
	}).Return(activities.ListPRsResponse{
		PullRequests: pullRequests,
	}, nil)

	env.OnActivity(ga.ListModifiedFiles, mock.Anything, activities.ListModifiedFilesRequest{
		Repo:        req.Repo,
		PullRequest: pullRequests[0],
	}).Return(activities.ListModifiedFilesResponse{}, errors.New("error"))

	env.OnActivity(ra.SetPRRevision, mock.Anything, activities.SetPRRevisionRequest{
		Repository:  req.Repo,
		PullRequest: pullRequests[0],
	}).Return(activities.SetPRRevisionResponse{}, nil)

	env.ExecuteWorkflow(testSetMiminumValidRevisionForRootWorkflow, req)
	env.AssertExpectations(t)

	err := env.GetWorkflowResult(nil)
	assert.Nil(t, err)
}

func TestMinRevisionSetter_OpenPR_PatternMatchErr(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	ga := &testGithubActivities{}
	ra := &testRevisionSetterActivities{}
	env.RegisterActivity(ra)
	env.RegisterActivity(ga)

	req := Request{
		Repo: github.Repo{
			Owner: "owner",
			Name:  "test",
		},
		Root: terraform.Root{
			TrackedFiles: []string{"!"},
		},
	}

	pullRequests := []github.PullRequest{
		{
			Number: 1,
		},
	}

	env.OnActivity(ga.ListPRs, mock.Anything, activities.ListPRsRequest{
		Repo:  req.Repo,
		State: github.OpenPullRequest,
	}).Return(activities.ListPRsResponse{
		PullRequests: pullRequests,
	}, nil)

	filesModifiedPr1 := []string{"test/dir2/rebase.tf"}
	env.OnActivity(ga.ListModifiedFiles, mock.Anything, activities.ListModifiedFilesRequest{
		Repo:        req.Repo,
		PullRequest: pullRequests[0],
	}).Return(activities.ListModifiedFilesResponse{
		FilePaths: filesModifiedPr1,
	}, nil)

	env.ExecuteWorkflow(testSetMiminumValidRevisionForRootWorkflow, req)
	env.AssertExpectations(t)

	err := env.GetWorkflowResult(nil)
	assert.NoError(t, err)
}
