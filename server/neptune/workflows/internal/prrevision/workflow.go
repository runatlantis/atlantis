package prrevision

import (
	"context"
	"time"

	"github.com/docker/docker/pkg/fileutils"
	"github.com/pkg/errors"
	key "github.com/runatlantis/atlantis/server/neptune/context"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/config/logger"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/metrics"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	RetryCount          = 3
	StartToCloseTimeout = 30 * time.Second
)

type Request struct {
	Repo github.Repo
	Root terraform.Root
}

type setterActivities interface {
	SetPRRevision(ctx context.Context, request activities.SetPRRevisionRequest) (activities.SetPRRevisionResponse, error)
}

type githubActivities interface {
	ListPRs(ctx context.Context, request activities.ListPRsRequest) (activities.ListPRsResponse, error)
	ListModifiedFiles(ctx context.Context, request activities.ListModifiedFilesRequest) (activities.ListModifiedFilesResponse, error)
}

func Workflow(ctx workflow.Context, request Request) error {
	// GH API calls should not hit ratelimit issues since we cap the TaskQueueActivitiesPerSecond for the min revison setter TQ such that it's within our GH API budget
	// Configuring both GH API calls and PRSetRevision calls to 3 retries before failing
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: StartToCloseTimeout,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: RetryCount,
		},
	})

	var ga *activities.Github
	var ra *activities.RevsionSetter
	runner := &Runner{
		GithubActivities:         ga,
		RevisionSetterActivities: ra,
		Scope:                    metrics.NewScope(ctx),
	}

	return runner.Run(ctx, request)
}

type Runner struct {
	GithubActivities         githubActivities
	RevisionSetterActivities setterActivities
	Scope                    metrics.Scope
}

func (r *Runner) Run(ctx workflow.Context, request Request) error {
	prs, err := r.listOpenPRs(ctx, request.Repo)
	if err != nil {
		return err
	}

	r.Scope.Counter("open_prs").Inc(int64(len(prs)))
	if err := r.setRevision(ctx, request, prs); err != nil {
		return errors.Wrap(err, "setting minimum revision for pr modifiying root")
	}
	return nil
}

func (r *Runner) listOpenPRs(ctx workflow.Context, repo github.Repo) ([]github.PullRequest, error) {
	var resp activities.ListPRsResponse
	err := workflow.ExecuteActivity(ctx, r.GithubActivities.ListPRs, activities.ListPRsRequest{
		Repo:  repo,
		State: github.Open,
	}).Get(ctx, &resp)
	if err != nil {
		return []github.PullRequest{}, errors.Wrap(err, "listing open PRs")
	}

	return resp.PullRequests, nil
}

func (r *Runner) setRevision(ctx workflow.Context, req Request, prs []github.PullRequest) error {
	// spawn activities to list modified files in each open PR async
	futuresByPullNum := map[github.PullRequest]workflow.Future{}
	for _, pr := range prs {
		futuresByPullNum[pr] = workflow.ExecuteActivity(ctx, r.GithubActivities.ListModifiedFiles, activities.ListModifiedFilesRequest{
			Repo:        req.Repo,
			PullRequest: pr,
		})
	}

	// resolve the futures and spawn activities to set minimum revision for PR if needed
	futures := []workflow.Future{}
	for _, pr := range prs {
		if setMinimumRevisionFuture := r.setRevisionForPR(ctx, req, pr, futuresByPullNum[pr]); setMinimumRevisionFuture != nil {
			futures = append(futures, workflow.ExecuteActivity(ctx, r.RevisionSetterActivities.SetPRRevision, activities.SetPRRevisionRequest{
				Repository:  req.Repo,
				PullRequest: pr,
			}))
		}
	}

	// wait to resolve futures for setting minimum revision
	for _, future := range futures {
		var resp activities.SetPRRevisionResponse
		err := future.Get(ctx, &resp)
		if err != nil {
			return errors.Wrap(err, "error setting pr revision")
		}
	}

	return nil
}

func (r *Runner) setRevisionForPR(ctx workflow.Context, req Request, pull github.PullRequest, future workflow.Future) workflow.Future {
	// let's be preventive and set minimum revision for this PR if this listModifiedFiles fails after 3 attempts
	var result activities.ListModifiedFilesResponse
	if err := future.Get(ctx, &result); err != nil {
		logger.Error(ctx, "error listing modified files in PR", key.ErrKey, err, key.PullNumKey, pull.Number)
		return r.setMinRevision(ctx, req.Repo, pull)
	}

	// should not fail unless the TrackedFiles config is invalid which is validated on startup
	// let's be preventive and set minimum revision for this PR if file path match errors out
	rootModified, err := isRootModified(req.Root, result.FilePaths)
	if err != nil {
		logger.Error(ctx, "error matching file paths in PR", key.ErrKey, err, key.PullNumKey, pull.Number)
		return r.setMinRevision(ctx, req.Repo, pull)
	}

	if rootModified {
		return r.setMinRevision(ctx, req.Repo, pull)
	}

	return nil
}

func (r *Runner) setMinRevision(ctx workflow.Context, repo github.Repo, pull github.PullRequest) workflow.Future {
	return workflow.ExecuteActivity(ctx, r.RevisionSetterActivities.SetPRRevision, activities.SetPRRevisionRequest{
		Repository:  repo,
		PullRequest: pull,
	})
}

func isRootModified(root terraform.Root, modifiedFiles []string) (bool, error) {
	// look at the filepaths for the root
	trackedFilesRelToRepoRoot := root.GetTrackedFilesRelativeToRepo()
	pm, err := fileutils.NewPatternMatcher(trackedFilesRelToRepoRoot)
	if err != nil {
		return false, errors.Wrap(err, "building file pattern matcher using tracked files config")
	}

	for _, file := range modifiedFiles {
		match, err := pm.Matches(file)
		if err != nil {
			return false, errors.Wrap(err, "matching file path")
		}

		if !match {
			continue
		}

		return true, nil
	}

	return false, nil
}
