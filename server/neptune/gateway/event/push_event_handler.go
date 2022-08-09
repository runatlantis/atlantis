package event

import (
	"context"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/lyft/feature"
	"github.com/runatlantis/atlantis/server/vcs"
	"github.com/runatlantis/atlantis/server/neptune/gateway/sync"
)

type Push struct {
	Repo   models.Repo
	Ref    vcs.Ref
	Sha    string
	Sender vcs.User
}

type scheduler interface {
	Schedule(ctx context.Context, f sync.Executor)
}

type PushHandler struct {
	Allocator feature.Allocator
	Scheduler scheduler
}

func (p *PushHandler) Handle(ctx context.Context, event Push) error {
	shouldAllocate, err := p.Allocator.ShouldAllocate(feature.PlatformMode, feature.FeatureContext{
		RepoName: event.Repo.FullName,
	})

	if err != nil {
		return errors.Wrap(err, "unable to allocate")
	}

	if !shouldAllocate {
		return nil
	}

	p.Scheduler.Schedule(ctx, func(ctx context.Context) error {
		return p.handle(ctx, event)
	})

	return nil
}

// TODO
func (p *PushHandler) handle(ctx context.Context, event Push) error {
	return nil
}
