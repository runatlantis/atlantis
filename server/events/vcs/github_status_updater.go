package vcs

import (
	"context"

	"github.com/google/go-github/v31/github"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs/types"
)

// Interface to support status updates for PR Status Checks and Github Status Checks
type StatusUpdater interface {
	UpdateStatus(ctx context.Context, request types.UpdateStatusRequest) (string, error)
}

type PullStatusCheckUpdater struct {
	client *github.Client
}

// UpdateStatus updates the status badge on the pull request.
// See https://github.com/blog/1227-commit-status-api.
func (g *PullStatusCheckUpdater) UpdateStatus(ctx context.Context, request types.UpdateStatusRequest) (string, error) {
	ghState := "error"
	switch request.State {
	case models.PendingCommitStatus:
		ghState = "pending"
	case models.SuccessCommitStatus:
		ghState = "success"
	case models.FailedCommitStatus:
		ghState = "failure"
	}

	status := &github.RepoStatus{
		State:       github.String(ghState),
		Description: github.String(request.Description),
		Context:     github.String(request.StatusName),
		TargetURL:   &request.DetailsURL,
	}
	_, _, err := g.client.Repositories.CreateStatus(ctx, request.Repo.Owner, request.Repo.Name, request.Ref, status)
	return "", err
}

type GithubStatusCheckUpdater struct {
	client *github.Client
}

func (c *GithubStatusCheckUpdater) UpdateStatus(ctx context.Context, request types.UpdateStatusRequest) (string, error) {
	// TODO: Implement update status Github Checks
	return "", nil
}
