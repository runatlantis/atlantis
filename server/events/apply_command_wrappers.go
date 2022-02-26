package events

import (
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
)

type CommandRunnerWrapper func(CommentCommandRunner) CommentCommandRunner

func WrapApplyCommandRunner(runner CommentCommandRunner, wrappers ...CommandRunnerWrapper) CommentCommandRunner {
	wrapped := runner
	for _, wrap := range wrappers {
		wrapped = wrap(wrapped)
	}

	return wrapped
}

type applyFunc func(ctx *CommandContext, cmd *CommentCommand)

func (a applyFunc) Run(ctx *CommandContext, cmd *CommentCommand) {
	a(ctx, cmd)
}

func WithGlobalLock(
	locker locking.ApplyLockChecker,
	vcsClient vcs.Client,
) CommandRunnerWrapper {
	return func(runner CommentCommandRunner) CommentCommandRunner {
		return applyFunc(func(ctx *CommandContext, cmd *CommentCommand) {
			baseRepo := ctx.Pull.BaseRepo
			pull := ctx.Pull

			lock, err := locker.CheckApplyLock()
			// CheckApplyLock falls back to DisableApply flag if fetching the lock
			// raises an error
			// We will log failure as warning
			if err != nil {
				ctx.Log.Warn("checking global apply lock: %s", err)
			}

			if lock.Locked {
				ctx.Log.Info("ignoring apply command since apply disabled globally")
				if err := vcsClient.CreateComment(baseRepo, pull.Num, applyDisabledComment, models.ApplyCommand.String()); err != nil {
					ctx.Log.Err("unable to comment on pull request: %s", err)
				}

				return
			}
			runner.Run(ctx, cmd)
		})
	}
}

func WithDisableAll(
	disableApplyAll bool,
	vcsClient vcs.Client,
) CommandRunnerWrapper {
	return func(runner CommentCommandRunner) CommentCommandRunner {
		return applyFunc(func(ctx *CommandContext, cmd *CommentCommand) {
			baseRepo := ctx.Pull.BaseRepo
			pull := ctx.Pull

			if disableApplyAll && !cmd.IsForSpecificProject() {
				ctx.Log.Info("ignoring apply command without flags since apply all is disabled")
				if err := vcsClient.CreateComment(baseRepo, pull.Num, applyAllDisabledComment, models.ApplyCommand.String()); err != nil {
					ctx.Log.Err("unable to comment on pull request: %s", err)
				}

				return
			}

			runner.Run(ctx, cmd)
		})
	}
}

func WithPullRequestStatus(fetcher vcs.PullReqStatusFetcher) CommandRunnerWrapper {
	return func(runner CommentCommandRunner) CommentCommandRunner {
		return applyFunc(func(ctx *CommandContext, cmd *CommentCommand) {
			var err error
			baseRepo := ctx.Pull.BaseRepo
			pull := ctx.Pull

			// Get the mergeable status before we set any build statuses of our own.
			// We do this here because when we set a "Pending" status, if users have
			// required the Atlantis status checks to pass, then we've now changed
			// the mergeability status of the pull request.
			// This sets the approved, mergeable, and sqlocked status in the context.
			ctx.PullRequestStatus, err = fetcher.FetchPullStatus(baseRepo, pull)
			if err != nil {
				// On error we continue the request with mergeable assumed false.
				// We want to continue because not all apply's will need this status,
				// only if they rely on the mergeability requirement.
				// All PullRequestStatus fields are set to false by default when error.
				ctx.Log.Warn("unable to get pull request status: %s. Continuing with mergeable and approved assumed false", err)
			}

			runner.Run(ctx, cmd)
		})
	}
}
