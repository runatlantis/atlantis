package rebase

import (
	"context"

	"github.com/docker/docker/pkg/fileutils"
	"github.com/pkg/errors"
	key "github.com/runatlantis/atlantis/server/neptune/context"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/config/logger"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/metrics"
	"go.temporal.io/sdk/workflow"
)

type pullRebaserActivites interface {
	SetPRRevision(ctx context.Context, request activities.SetPRRevisionRequest) (activities.SetPRRevisionResponse, error)
	GithubListModifiedFiles(ctx context.Context, request activities.ListModifiedFilesRequest) (activities.ListModifiedFilesResponse, error)
	GithubListOpenPRs(ctx context.Context, request activities.ListOpenPRsRequest) (activities.ListOpenPRsResponse, error)
}

type PullRebaser struct {
	RebaseActivites pullRebaserActivites
}

func (p *PullRebaser) RebaseOpenPRsForRoot(ctx workflow.Context, repo github.Repo, root terraform.Root, scope metrics.Scope) error {
	// list open PRs
	var listOpenPRsResp activities.ListOpenPRsResponse
	err := workflow.ExecuteActivity(ctx, p.RebaseActivites.GithubListOpenPRs, activities.ListOpenPRsRequest{
		Repo: repo,
	}).Get(ctx, &listOpenPRsResp)
	if err != nil {
		return errors.Wrap(err, "listing open PRs")
	}

	// spawn activities to list modified files in each open PR async
	futureByPullNum := map[github.PullRequest]workflow.Future{}
	for _, pullRequest := range listOpenPRsResp.PullRequests {
		futureByPullNum[pullRequest] = workflow.ExecuteActivity(ctx, p.RebaseActivites.GithubListModifiedFiles, activities.ListModifiedFilesRequest{
			Repo:        repo,
			PullRequest: pullRequest,
		})
	}

	// track number of open PRs being processed
	scope.Counter("open_prs").Inc(int64(len(futureByPullNum)))

	// resolve the futures and rebase PR if needed
	rebaseFutures := []workflow.Future{}
	for _, pullRequest := range listOpenPRsResp.PullRequests {
		future := futureByPullNum[pullRequest]

		// let's be preventive and rebase this PR if this call fails after 3 attempts
		var result activities.ListModifiedFilesResponse
		listFilesErr := future.Get(ctx, &result)
		if listFilesErr != nil {
			logger.Error(ctx, "error listing modified files in PR", key.ErrKey, listFilesErr, key.PullNumKey, pullRequest.Number)
			rebaseFutures = append(rebaseFutures, workflow.ExecuteActivity(ctx, p.RebaseActivites.SetPRRevision, activities.SetPRRevisionRequest{
				Repository:  repo,
				PullRequest: pullRequest,
			}))
			continue
		}

		// fail the workflow if this errors out since we validate the TrackedFiles config at startup
		shouldRebase, err := shouldRebasePullRequest(root, result.FilePaths)
		if err != nil {
			scope.Counter("filepath_match_err").Inc(1)
			return errors.Wrap(err, "error matching filepaths in PR")
		}

		if !shouldRebase {
			continue
		}

		// spawn activity to rebase this PR and continue
		rebaseFutures = append(rebaseFutures, workflow.ExecuteActivity(ctx, p.RebaseActivites.SetPRRevision, activities.SetPRRevisionRequest{
			Repository:  repo,
			PullRequest: pullRequest,
		}))
	}

	// wait for rebase futures to resolve
	for _, future := range rebaseFutures {
		var resp activities.SetPRRevisionResponse
		err := future.Get(ctx, &resp)
		if err != nil {
			return errors.Wrap(err, "error setting pr revision")
		}
	}

	return nil
}

func shouldRebasePullRequest(root terraform.Root, modifiedFiles []string) (bool, error) {
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
