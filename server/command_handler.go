package server

import (
	"fmt"

	"github.com/hootsuite/atlantis/github"
	"github.com/hootsuite/atlantis/logging"
	"github.com/hootsuite/atlantis/recovery"
)

type CommandHandler struct {
	planExecutor  *PlanExecutor
	applyExecutor *ApplyExecutor
	helpExecutor  *HelpExecutor
	githubClient  *github.Client
	eventParser   *EventParser
	logger        *logging.SimpleLogger
}

type CommandResponse struct {
	Error          error
	Failure        string
	ProjectResults []ProjectResult
	Command        CommandName
}

type ProjectResult struct {
	Path         string
	Error        error
	Failure      string
	PlanSuccess  *PlanSuccess
	ApplySuccess string
}

func (p ProjectResult) Status() Status {
	if p.Error != nil {
		return Error
	}
	if p.Failure != "" {
		return Failure
	}
	return Success
}

type CommandName int

const (
	Apply CommandName = iota
	Plan
	Help
	// Adding more? Don't forget to update String() below
)

func (c CommandName) String() string {
	switch c {
	case Apply:
		return "apply"
	case Plan:
		return "plan"
	case Help:
		return "help"
	}
	return ""
}

type Command struct {
	Name        CommandName
	Environment string
	Verbose     bool
	Flags       []string
}

func (c *CommandHandler) ExecuteCommand(ctx *CommandContext) {
	src := fmt.Sprintf("%s/pull/%d", ctx.BaseRepo.FullName, ctx.Pull.Num)
	// it's safe to reuse the underlying logger
	ctx.Log = logging.NewSimpleLogger(src, c.logger.Logger, true, c.logger.Level)
	defer c.recover(ctx)

	// need to get additional data from the PR
	ghPull, _, err := c.githubClient.GetPullRequest(ctx.BaseRepo, ctx.Pull.Num)
	if err != nil {
		ctx.Log.Err("making pull request API call to GitHub: %s", err)
		return
	}
	pull, headRepo, err := c.eventParser.ExtractPullData(ghPull)
	if err != nil {
		ctx.Log.Err("extracting required fields from comment data: %s", err)
		return
	}
	ctx.Pull = pull
	ctx.HeadRepo = headRepo

	if ghPull.GetState() != "open" {
		ctx.Log.Info("command was run on closed pull request")
		c.githubClient.CreateComment(ctx.BaseRepo, ctx.Pull, "Atlantis commands can't be run on closed pull requests")
		return
	}

	switch ctx.Command.Name {
	case Plan:
		c.planExecutor.execute(ctx)
	case Apply:
		c.applyExecutor.execute(ctx)
	case Help:
		c.helpExecutor.execute(ctx)
	default:
		ctx.Log.Err("failed to determine desired command, neither plan nor apply")
	}
}

func (c *CommandHandler) SetLockURL(f func(id string) (url string)) {
	c.planExecutor.LockURL = f
}

// recover logs and creates a comment on the pull request for panics
func (c *CommandHandler) recover(ctx *CommandContext) {
	if err := recover(); err != nil {
		stack := recovery.Stack(3)
		c.githubClient.CreateComment(ctx.BaseRepo, ctx.Pull, fmt.Sprintf("**Error: goroutine panic. This is a bug.**\n```\n%s\n%s```", err, stack))
		ctx.Log.Err("PANIC: %s\n%s", err, stack)
	}
}
