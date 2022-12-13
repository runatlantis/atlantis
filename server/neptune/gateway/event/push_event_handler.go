package event

import (
	"context"
	"github.com/runatlantis/atlantis/server/vcs/provider/github"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/lyft/feature"
	"github.com/runatlantis/atlantis/server/neptune/sync"
	"github.com/runatlantis/atlantis/server/neptune/workflows"
	"github.com/runatlantis/atlantis/server/vcs"
)

type PushAction string

const (
	DeletedAction PushAction = "deleted"
	CreatedAction PushAction = "created"
	UpdatedAction PushAction = "updated"
)

type Push struct {
	Repo              models.Repo
	Ref               vcs.Ref
	Sha               string
	Sender            models.User
	InstallationToken int64
	Action            PushAction
}

type scheduler interface {
	Schedule(ctx context.Context, f sync.Executor) error
}

type rootDeployer interface {
	Deploy(ctx context.Context, deployOptions RootDeployOptions) error
}

type PushHandler struct {
	Allocator    feature.Allocator
	Scheduler    scheduler
	Logger       logging.Logger
	RootDeployer rootDeployer
}

func (p *PushHandler) Handle(ctx context.Context, event Push) error {
	shouldAllocate, err := p.Allocator.ShouldAllocate(feature.PlatformMode, feature.FeatureContext{
		RepoName: event.Repo.FullName,
	})

	if err != nil {
		p.Logger.ErrorContext(ctx, "unable to allocate platformmode")
		return nil
	}

	if !shouldAllocate {
		p.Logger.DebugContext(ctx, "handler not configured for allocation")
		return nil
	}

	if event.Ref.Type != vcs.BranchRef || event.Ref.Name != event.Repo.DefaultBranch {
		p.Logger.DebugContext(ctx, "dropping event for unexpected ref")
		return nil
	}

	if event.Action == DeletedAction {
		p.Logger.WarnContext(ctx, "ref was deleted, resources might still exist")
		return nil
	}

	return p.Scheduler.Schedule(ctx, func(ctx context.Context) error {
		return p.handle(ctx, event)
	})
}

func (p *PushHandler) handle(ctx context.Context, event Push) error {
	builderOptions := BuilderOptions{
		RepoFetcherOptions: github.RepoFetcherOptions{
			ShallowClone: true,
		},
		FileFetcherOptions: github.FileFetcherOptions{
			Sha: event.Sha,
		},
	}
	rootDeployOptions := RootDeployOptions{
		Repo:              event.Repo,
		Branch:            event.Ref.Name,
		Revision:          event.Sha,
		Sender:            event.Sender,
		InstallationToken: event.InstallationToken,
		BuilderOptions:    builderOptions,
		Trigger:           workflows.MergeTrigger,
	}
	return p.RootDeployer.Deploy(ctx, rootDeployOptions)
}
