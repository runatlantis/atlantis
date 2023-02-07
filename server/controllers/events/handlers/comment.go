package handlers

import (
	"context"
	"fmt"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/http"
	"github.com/runatlantis/atlantis/server/logging"
	contextInternal "github.com/runatlantis/atlantis/server/neptune/context"
	event_types "github.com/runatlantis/atlantis/server/neptune/gateway/event"
)

// commentCreator creates a comment on a pull request for a given repo
type commentCreator interface {
	CreateComment(repo models.Repo, pullNum int, comment string, command string) error
}

// commentParser parsers a vcs pull request comment and returns the result
type commentParser interface {
	Parse(comment string, vcsHost models.VCSHostType) events.CommentParseResult
}

func NewCommentEvent(
	commentParser commentParser,
	repoAllowlistChecker *events.RepoAllowlistChecker,
	commentCreator commentCreator,
	commandRunner events.CommandRunner,
	logger logging.Logger,
) *CommentEvent {
	return &CommentEvent{
		commentParser: commentParser,
		commandHandler: &asyncHandler{
			commandHandler: &CommandHandler{
				CommandRunner: commandRunner,
			},
			logger: logger,
		},
		RepoAllowlistChecker: repoAllowlistChecker,
		commentCreator:       commentCreator,
		logger:               logger,
	}

}

func NewCommentEventWithCommandHandler(
	commentParser events.CommentParsing,
	repoAllowlistChecker *events.RepoAllowlistChecker,
	commentCreator commentCreator,
	commandHandler commandHandler,
	logger logging.Logger,
) *CommentEvent {
	return &CommentEvent{
		commentParser:        commentParser,
		commandHandler:       commandHandler,
		RepoAllowlistChecker: repoAllowlistChecker,
		commentCreator:       commentCreator,
		logger:               logger,
	}

}

// commandHandler is the handler responsible for running a specific command
// after it's been parsed from a comment.
type commandHandler interface {
	Handle(ctx context.Context, request *http.BufferedRequest, event event_types.Comment, command *command.Comment) error
}

type CommandHandler struct {
	CommandRunner events.CommandRunner
}

func (h *CommandHandler) Handle(ctx context.Context, _ *http.BufferedRequest, event event_types.Comment, command *command.Comment) error {
	h.CommandRunner.RunCommentCommand(
		ctx,
		event.BaseRepo,
		event.HeadRepo,
		event.Pull,
		event.User,
		event.PullNum,
		command,
		event.Timestamp,
		event.InstallationToken,
	)
	return nil
}

type asyncHandler struct {
	commandHandler *CommandHandler
	logger         logging.Logger
}

func (h *asyncHandler) Handle(ctx context.Context, request *http.BufferedRequest, event event_types.Comment, command *command.Comment) error {
	go func() {
		// Passing background context to avoid context cancellation since the parent goroutine does not wait for this goroutine to finish execution.
		ctx = contextInternal.CopyFields(context.Background(), ctx)
		err := h.commandHandler.Handle(ctx, request, event, command)

		if err != nil {
			h.logger.ErrorContext(ctx, err.Error())
		}
	}()
	return nil
}

type CommentEvent struct {
	commentParser        events.CommentParsing
	commandHandler       commandHandler
	RepoAllowlistChecker *events.RepoAllowlistChecker
	commentCreator       commentCreator
	logger               logging.Logger
}

func (h *CommentEvent) Handle(ctx context.Context, request *http.BufferedRequest, event event_types.Comment) error {
	comment := event.Comment
	vcsHost := event.VCSHost
	baseRepo := event.BaseRepo
	pullNum := event.PullNum

	parseResult := h.commentParser.Parse(comment, vcsHost)
	if parseResult.Ignore {
		h.logger.WarnContext(ctx, "ignoring comment")
		return nil
	}

	// At this point we know it's a command we're not supposed to ignore, so now
	// we check if this repo is allowed to run commands in the first plach.
	if !h.RepoAllowlistChecker.IsAllowlisted(baseRepo.FullName, baseRepo.VCSHost.Hostname) {
		return fmt.Errorf("comment event from non-allowlisted repo \"%s/%s\"", baseRepo.VCSHost.Hostname, baseRepo.FullName)
	}

	// If the command isn't valid or doesn't require processing, ex.
	// "atlantis help" then we just comment back immediately.
	// We do this here rather than earlier because we need access to the pull
	// variable to comment back on the pull request.
	if parseResult.CommentResponse != "" {
		if err := h.commentCreator.CreateComment(baseRepo, pullNum, parseResult.CommentResponse, ""); err != nil {
			h.logger.ErrorContext(ctx, err.Error())
		}
		return nil
	}

	return h.commandHandler.Handle(ctx, request, event, parseResult.Command)
}
