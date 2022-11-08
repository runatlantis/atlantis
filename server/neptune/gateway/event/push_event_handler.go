package event

import (
	"context"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/lyft/feature"
	contextInternal "github.com/runatlantis/atlantis/server/neptune/context"
	"github.com/runatlantis/atlantis/server/neptune/gateway/sync"
	"github.com/runatlantis/atlantis/server/neptune/workflows"
	"github.com/runatlantis/atlantis/server/vcs"
	"go.temporal.io/sdk/client"
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

type deploySignaler interface {
	SignalWithStartWorkflow(ctx context.Context, rootCfg *valid.MergedProjectCfg, repo models.Repo, revision string, installationToken int64, ref vcs.Ref, sender models.User, trigger workflows.Trigger) (client.WorkflowRun, error)
}

type rootConfigBuilder interface {
	Build(ctx context.Context, repo models.Repo, branch string, sha string, installationToken int64) ([]*valid.MergedProjectCfg, error)
}

type PushHandler struct {
	Allocator         feature.Allocator
	Scheduler         scheduler
	DeploySignaler    deploySignaler
	Logger            logging.Logger
	RootConfigBuilder rootConfigBuilder
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
	rootCfgs, err := p.RootConfigBuilder.Build(ctx, event.Repo, event.Ref.Name, event.Sha, event.InstallationToken)
	if err != nil {
		return errors.Wrap(err, "generating roots")
	}
	for _, rootCfg := range rootCfgs {
		c := context.WithValue(ctx, contextInternal.ProjectKey, rootCfg.Name)

		if rootCfg.WorkflowMode != valid.PlatformWorkflowMode {
			p.Logger.DebugContext(c, "root is not configured for platform mode, skipping...")
			continue
		}

		run, err := p.DeploySignaler.SignalWithStartWorkflow(
			c,
			rootCfg,
			event.Repo,
			event.Sha,
			event.InstallationToken,
			event.Ref,
			event.Sender,
			workflows.MergeTrigger)
		if err != nil {
			return errors.Wrap(err, "signalling workflow")
		}

		p.Logger.InfoContext(c, "Signaled workflow.", map[string]interface{}{
			"workflow-id": run.GetID(), "run-id": run.GetRunID(),
		})
	}
	return nil
}
