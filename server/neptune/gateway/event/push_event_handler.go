package event

import (
	"context"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/lyft/feature"
	"github.com/runatlantis/atlantis/server/neptune/gateway/sync"
	"github.com/runatlantis/atlantis/server/neptune/workflows"
	"github.com/runatlantis/atlantis/server/vcs"
	"go.temporal.io/sdk/client"
)

type Push struct {
	Repo   models.Repo
	Ref    vcs.Ref
	Sha    string
	Sender vcs.User
}

type signaler interface {
	SignalWithStartWorkflow(ctx context.Context, workflowID string, signalName string, signalArg interface{},
		options client.StartWorkflowOptions, workflow interface{}, workflowArgs ...interface{}) (client.WorkflowRun, error)
}

type scheduler interface {
	Schedule(ctx context.Context, f sync.Executor) error
}

type PushHandler struct {
	Allocator      feature.Allocator
	Scheduler      scheduler
	TemporalClient signaler
	Logger         logging.Logger
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
		return nil
	}

	return p.Scheduler.Schedule(ctx, func(ctx context.Context) error {
		return p.handle(ctx, event)
	})
}

func (p *PushHandler) handle(ctx context.Context, event Push) error {
	options := client.StartWorkflowOptions{TaskQueue: workflows.DeployTaskQueue}

	// TODO: clone and build project config

	run, err := p.TemporalClient.SignalWithStartWorkflow(
		ctx,

		// TODO: name should include root name as well
		event.Repo.FullName,
		workflows.DeployNewRevisionSignalID,
		workflows.DeployNewRevisionSignalRequest{
			Revision: event.Sha,
		},
		options,
		workflows.Deploy,

		// TODO: add other request params as we support them
		workflows.DeployRequest{
			Repository: workflows.Repo{
				URL:      event.Repo.CloneURL,
				FullName: event.Repo.FullName,
				Name:     event.Repo.Name,
				Owner:    event.Repo.Owner,
			},
		},
	)

	if err != nil {
		return errors.Wrap(err, "signalling workflow")
	}

	p.Logger.InfoContext(ctx, "Signaled workflow.", map[string]interface{}{
		"workflow-id": run.GetID(), "run-id": run.GetRunID(),
	})

	return nil
}
