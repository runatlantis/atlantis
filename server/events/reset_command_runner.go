package events

import (
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/vcs"
)

func NewResetCommandRunner(
	pullCleaner PullCleaner,
	vcsClient vcs.Client,
) *ResetCommandRunner {
	return &ResetCommandRunner{
		pullCleaner: pullCleaner,
		vcsClient:   vcsClient,
	}
}

type ResetCommandRunner struct {
	pullCleaner  PullCleaner
	commandRunner CommandRunner // Set later to avoid circular dependency
	vcsClient     vcs.Client
}

func (r *ResetCommandRunner) SetCommandRunner(commandRunner CommandRunner) {
	r.commandRunner = commandRunner
}

func (r *ResetCommandRunner) Run(ctx *command.Context, cmd *CommentCommand) {
	baseRepo := ctx.Pull.BaseRepo
	headRepo := ctx.HeadRepo
	pull := ctx.Pull
	user := ctx.User

	ctx.Log.Info("Resetting PR state by clearing locks and triggering replan")

	// First, clean up the pull like it was closed
	if err := r.pullCleaner.CleanUpPull(ctx.Log, baseRepo, pull); err != nil {
		ctx.Log.Err("failed to clean up pull during reset: %s", err)
		if commentErr := r.vcsClient.CreateComment(ctx.Log, baseRepo, pull.Num, "Failed to reset PR state", command.Reset.String()); commentErr != nil {
			ctx.Log.Err("unable to comment on reset failure: %s", commentErr)
		}
		return
	}

	// Now trigger autoplan to replan all projects
	r.commandRunner.RunAutoplanCommand(baseRepo, headRepo, pull, user)

	// Comment on success
	if commentErr := r.vcsClient.CreateComment(ctx.Log, baseRepo, pull.Num, "PR state has been reset. All locks cleared and replanning triggered.", command.Reset.String()); commentErr != nil {
		ctx.Log.Err("unable to comment on reset success: %s", commentErr)
	}
}
