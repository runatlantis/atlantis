package event

import (
	"bytes"
	"context"
	"github.com/runatlantis/atlantis/server/lyft/feature"
	"time"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/http"
	"github.com/runatlantis/atlantis/server/logging"
)

const (
	Approved = "approved"
)

type PullRequestReview struct {
	InstallationToken int64
	Repo              models.Repo
	User              models.User
	State             string
	Ref               string
	Timestamp         time.Time
	Pull              models.PullRequest
}

type fetcher interface {
	ListFailedPolicyCheckRuns(ctx context.Context, installationToken int64, repo models.Repo, ref string) ([]string, error)
}

type PullRequestReviewWorkerProxy struct {
	Scheduler       scheduler
	SnsWriter       Writer
	Logger          logging.Logger
	Allocator       feature.Allocator
	CheckRunFetcher fetcher
}

func (p *PullRequestReviewWorkerProxy) Handle(ctx context.Context, event PullRequestReview, request *http.BufferedRequest) error {
	return p.Scheduler.Schedule(ctx, func(ctx context.Context) error {
		return p.handle(ctx, event, request)
	})
}

func (p *PullRequestReviewWorkerProxy) handle(ctx context.Context, event PullRequestReview, request *http.BufferedRequest) error {
	shouldAllocate, err := p.Allocator.ShouldAllocate(feature.PolicyV2, feature.FeatureContext{
		RepoName: event.Repo.FullName,
	})
	if err != nil {
		p.Logger.ErrorContext(ctx, "unable to allocate policy v2")
		return err
	}
	if !shouldAllocate {
		return nil
	}

	// Ignore non-approval events
	if event.State != Approved {
		return nil
	}

	// Ignore PRs without failing policy checks
	failedPolicyCheckRuns, err := p.CheckRunFetcher.ListFailedPolicyCheckRuns(ctx, event.InstallationToken, event.Repo, event.Ref)
	if err != nil {
		p.Logger.ErrorContext(ctx, "unable to list failed policy check runs")
		return err
	}
	if len(failedPolicyCheckRuns) == 0 {
		return nil
	}
	// Forward to SNS to further process in the worker
	return p.forwardToSns(ctx, request)
}

func (p *PullRequestReviewWorkerProxy) forwardToSns(ctx context.Context, request *http.BufferedRequest) error {
	buffer := bytes.NewBuffer([]byte{})
	if err := request.GetRequestWithContext(ctx).Write(buffer); err != nil {
		return errors.Wrap(err, "writing request to buffer")
	}

	if err := p.SnsWriter.WriteWithContext(ctx, buffer.Bytes()); err != nil {
		return errors.Wrap(err, "writing buffer to sns")
	}
	p.Logger.InfoContext(ctx, "proxied request to sns")
	return nil
}
