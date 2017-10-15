package events

import (
	"fmt"

	"github.com/hootsuite/atlantis/server/events/github"
	"github.com/hootsuite/atlantis/server/logging"
	"github.com/hootsuite/atlantis/server/recovery"
)

// CommandHandler is the first step when processing a comment command.
type CommandHandler struct {
	PlanExecutor      Executor
	ApplyExecutor     Executor
	HelpExecutor      Executor
	LockURLGenerator  LockURLGenerator
	GHClient          github.Client
	GHStatus          GHStatusUpdater
	EventParser       EventParsing
	EnvLocker         EnvLocker
	GHCommentRenderer *GithubCommentRenderer
	Logger            *logging.SimpleLogger
}

func (c *CommandHandler) ExecuteCommand(ctx *CommandContext) {
	// Set up logger specific to this command.
	// It's safe to reuse the underlying logger.
	src := fmt.Sprintf("%s/pull/%d", ctx.BaseRepo.FullName, ctx.Pull.Num)
	ctx.Log = logging.NewSimpleLogger(src, c.Logger.Logger, true, c.Logger.Level)
	defer c.logPanics(ctx)

	// Get additional data from the pull request.
	ghPull, _, err := c.GHClient.GetPullRequest(ctx.BaseRepo, ctx.Pull.Num)
	if err != nil {
		ctx.Log.Err("making pull request API call to GitHub: %s", err)
		return
	}
	if ghPull.GetState() != "open" {
		ctx.Log.Info("command was run on closed pull request")
		c.GHClient.CreateComment(ctx.BaseRepo, ctx.Pull, "Atlantis commands can't be run on closed pull requests")
		return
	}
	pull, headRepo, err := c.EventParser.ExtractPullData(ghPull)
	if err != nil {
		ctx.Log.Err("extracting required fields from comment data: %s", err)
		return
	}
	ctx.Pull = pull
	ctx.HeadRepo = headRepo
	c.run(ctx)
}

func (c *CommandHandler) SetLockURL(f func(id string) (url string)) {
	c.LockURLGenerator.SetLockURL(f)
}

func (c *CommandHandler) run(ctx *CommandContext) {
	c.GHStatus.Update(ctx.BaseRepo, ctx.Pull, Pending, ctx.Command)
	if c.EnvLocker.TryLock(ctx.BaseRepo.FullName, ctx.Command.Environment, ctx.Pull.Num) != true {
		errMsg := fmt.Sprintf(
			"The %s environment is currently locked by another"+
				" command that is running for this pull request."+
				" Wait until the previous command is complete and try again.",
			ctx.Command.Environment)
		ctx.Log.Warn(errMsg)
		c.updatePull(ctx, CommandResponse{Failure: errMsg})
		return
	}
	defer c.EnvLocker.Unlock(ctx.BaseRepo.FullName, ctx.Command.Environment, ctx.Pull.Num)

	var cr CommandResponse
	switch ctx.Command.Name {
	case Plan:
		cr = c.PlanExecutor.Execute(ctx)
	case Apply:
		cr = c.ApplyExecutor.Execute(ctx)
	case Help:
		cr = c.HelpExecutor.Execute(ctx)
	default:
		ctx.Log.Err("failed to determine desired command, neither plan nor apply")
	}
	c.updatePull(ctx, cr)
}

func (c *CommandHandler) updatePull(ctx *CommandContext, res CommandResponse) {
	// Log if we got any errors or failures.
	if res.Error != nil {
		ctx.Log.Err(res.Error.Error())
	} else if res.Failure != "" {
		ctx.Log.Warn(res.Failure)
	}

	// Update the pull request's status icon and comment back.
	c.GHStatus.UpdateProjectResult(ctx, res)
	comment := c.GHCommentRenderer.Render(res, ctx.Command.Name, ctx.Log.History.String(), ctx.Command.Verbose)
	c.GHClient.CreateComment(ctx.BaseRepo, ctx.Pull, comment)
}

// logPanics logs and creates a comment on the pull request for panics
func (c *CommandHandler) logPanics(ctx *CommandContext) {
	if err := recover(); err != nil {
		stack := recovery.Stack(3)
		c.GHClient.CreateComment(ctx.BaseRepo, ctx.Pull,
			fmt.Sprintf("**Error: goroutine panic. This is a bug.**\n```\n%s\n%s```", err, stack))
		ctx.Log.Err("PANIC: %s\n%s", err, stack)
	}
}
