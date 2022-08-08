package event

import (
	"context"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/lyft/feature"
	"github.com/runatlantis/atlantis/server/vcs"
)

type Push struct {
	Repo   models.Repo
	Ref    vcs.Ref
	Sha    string
	Sender vcs.User
}

type PushHandler struct {
	Allocator feature.Allocator
}

func (p *PushHandler) Handle(ctx context.Context, event Push) error {

	shouldAllocate, err := p.Allocator.ShouldAllocate(feature.PlatformMode, feature.FeatureContext{
		RepoName: event.Repo.FullName,
	})

	if err != nil {
		return errors.Wrap(err, "unable to allocate")
	}

	if !shouldAllocate {
		// drop the event for now
		return nil
	}

	return nil
}
