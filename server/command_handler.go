package server

import (
	"fmt"

	"github.com/hootsuite/atlantis/github"
	"github.com/hootsuite/atlantis/logging"
	"github.com/hootsuite/atlantis/recovery"
)

type CommandHandler struct {
	PlanExecutor  Planner
	ApplyExecutor Executor
	HelpExecutor  Executor
	GithubClient  github.Client
	EventParser   EventParsing
	Logger        *logging.SimpleLogger
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
	ctx.Log = logging.NewSimpleLogger(src, c.Logger.Logger, true, c.Logger.Level)
	defer c.logPanics(ctx)

	// need to get additional data from the PR
	ghPull, _, err := c.GithubClient.GetPullRequest(ctx.BaseRepo, ctx.Pull.Num)
	if err != nil {
		ctx.Log.Err("making pull request API call to GitHub: %s", err)
		return
	}

	if ghPull.GetState() != "open" {
		ctx.Log.Info("command was run on closed pull request")
		c.GithubClient.CreateComment(ctx.BaseRepo, ctx.Pull, "Atlantis commands can't be run on closed pull requests")
		return
	}

	pull, headRepo, err := c.EventParser.ExtractPullData(ghPull)
	if err != nil {
		ctx.Log.Err("extracting required fields from comment data: %s", err)
		return
	}
	ctx.Pull = pull
	ctx.HeadRepo = headRepo

	switch ctx.Command.Name {
	case Plan:
		c.PlanExecutor.Execute(ctx)
	case Apply:
		c.ApplyExecutor.Execute(ctx)
	case Help:
		c.HelpExecutor.Execute(ctx)
	default:
		ctx.Log.Err("failed to determine desired command, neither plan nor apply")
	}
}

func (c *CommandHandler) SetLockURL(f func(id string) (url string)) {
	c.PlanExecutor.SetLockURL(f)
}

// logPanics logs and creates a comment on the pull request for panics
func (c *CommandHandler) logPanics(ctx *CommandContext) {
	if err := recover(); err != nil {
		stack := recovery.Stack(3)
		c.GithubClient.CreateComment(ctx.BaseRepo, ctx.Pull, fmt.Sprintf("**Error: goroutine panic. This is a bug.**\n```\n%s\n%s```", err, stack))
		ctx.Log.Err("PANIC: %s\n%s", err, stack)
	}
}
