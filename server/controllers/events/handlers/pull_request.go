package handlers

import (
	"context"
	"fmt"

	"github.com/runatlantis/atlantis/server/controllers/events/errors"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/http"
	"github.com/runatlantis/atlantis/server/logging"
	event_types "github.com/runatlantis/atlantis/server/neptune/gateway/event"
)

type eventTypeHandler interface {
	Handle(ctx context.Context, request *http.BufferedRequest, event event_types.PullRequest) error
}

type Autoplanner struct {
	CommandRunner events.CommandRunner
}

func (p *Autoplanner) Handle(ctx context.Context, _ *http.BufferedRequest, event event_types.PullRequest) error {
	p.CommandRunner.RunAutoplanCommand(
		ctx,
		event.Pull.BaseRepo,
		event.Pull.HeadRepo,
		event.Pull,
		event.User,
		event.Timestamp,
		event.InstallationToken,
	)

	return nil
}

type asyncAutoplanner struct {
	autoplanner *Autoplanner
	logger      logging.Logger
}

func (p *asyncAutoplanner) Handle(ctx context.Context, request *http.BufferedRequest, event event_types.PullRequest) error {
	go func() {
		// Passing background context to avoid context cancellation since the parent goroutine does not wait for this goroutine to finish execution.
		err := p.autoplanner.Handle(context.Background(), request, event)

		if err != nil {
			p.logger.ErrorContext(context.Background(), err.Error())
		}
	}()
	return nil
}

type PullCleaner struct {
	PullCleaner events.PullCleaner
	Logger      logging.Logger
}

func (c *PullCleaner) Handle(ctx context.Context, _ *http.BufferedRequest, event event_types.PullRequest) error {
	if err := c.PullCleaner.CleanUpPull(event.Pull.BaseRepo, event.Pull); err != nil {
		return err
	}

	c.Logger.InfoContext(ctx, "deleted locks and workspace")

	return nil
}

func NewPullRequestEvent(
	repoAllowlistChecker *events.RepoAllowlistChecker,
	pullCleaner events.PullCleaner,
	logger logging.Logger,
	commandRunner events.CommandRunner) *PullRequestEvent {
	asyncAutoplanner := &asyncAutoplanner{
		autoplanner: &Autoplanner{
			CommandRunner: commandRunner,
		},
		logger: logger,
	}
	return &PullRequestEvent{
		RepoAllowlistChecker:    repoAllowlistChecker,
		OpenedPullEventHandler:  asyncAutoplanner,
		UpdatedPullEventHandler: asyncAutoplanner,
		ClosedPullEventHandler: &PullCleaner{
			PullCleaner: pullCleaner,
			Logger:      logger,
		},
	}
}

func NewPullRequestEventWithEventTypeHandlers(
	repoAllowlistChecker *events.RepoAllowlistChecker,
	openedPullEventHandler eventTypeHandler,
	updatedPullEventHandler eventTypeHandler,
	closedPullEventHandler eventTypeHandler,
) *PullRequestEvent {
	return &PullRequestEvent{
		RepoAllowlistChecker:    repoAllowlistChecker,
		OpenedPullEventHandler:  openedPullEventHandler,
		UpdatedPullEventHandler: updatedPullEventHandler,
		ClosedPullEventHandler:  closedPullEventHandler,
	}
}

type PullRequestEvent struct {
	RepoAllowlistChecker *events.RepoAllowlistChecker

	// Delegate Handlers
	OpenedPullEventHandler  eventTypeHandler
	UpdatedPullEventHandler eventTypeHandler
	ClosedPullEventHandler  eventTypeHandler
}

func (h *PullRequestEvent) Handle(ctx context.Context, request *http.BufferedRequest, event event_types.PullRequest) error {
	pull := event.Pull
	baseRepo := pull.BaseRepo
	eventType := event.EventType

	if !h.RepoAllowlistChecker.IsAllowlisted(baseRepo.FullName, baseRepo.VCSHost.Hostname) {
		return fmt.Errorf("Pull request event from non-allowlisted repo \"%s/%s\"", baseRepo.VCSHost.Hostname, baseRepo.FullName)
	}

	switch eventType {
	case models.OpenedPullEvent:
		return h.OpenedPullEventHandler.Handle(ctx, request, event)
	case models.UpdatedPullEvent:
		return h.UpdatedPullEventHandler.Handle(ctx, request, event)
	case models.ClosedPullEvent:
		return h.ClosedPullEventHandler.Handle(ctx, request, event)
	case models.OtherPullEvent:
		return &errors.UnsupportedEventTypeError{Msg: "Unsupported event type made it through, this is likely a bug in the code."}
	}
	return nil
}
