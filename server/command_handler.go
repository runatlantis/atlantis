package server

import (
	"fmt"
	"github.com/hootsuite/atlantis/logging"
	"github.com/hootsuite/atlantis/recovery"
)

type CommandHandler struct {
	planExecutor  *PlanExecutor
	applyExecutor *ApplyExecutor
	helpExecutor  *HelpExecutor
	githubClient  *GithubClient
	eventParser   *EventParser
	logger        *logging.SimpleLogger
}

type CommandType int

const (
	Apply CommandType = iota
	Plan
	Help
	// Adding more? Don't forget to update String() below
)

func (c CommandType) String() string {
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
	verbose     bool
	environment string
	commandType CommandType
}

func (s *CommandHandler) ExecuteCommand(ctx *CommandContext) {
	src := fmt.Sprintf("%s/pull/%d", ctx.Repo.FullName, ctx.Pull.Num)
	// it'e safe to reuse the underlying logger e.logger.Log
	ctx.Log = logging.NewSimpleLogger(src, s.logger.Log, true, s.logger.Level)
	defer s.recover(ctx)

	// need to get additional data from the PR
	ghPull, _, err := s.githubClient.GetPullRequest(ctx.Repo, ctx.Pull.Num)
	if err != nil {
		ctx.Log.Err("pull request data api call failed: %v", err)
		return
	}
	pull, err := s.eventParser.ExtractPullData(ghPull)
	if err != nil {
		ctx.Log.Err("failed to extract required fields from comment data: %v", err)
		return
	}
	ctx.Pull = pull

	switch ctx.Command.commandType {
	case Plan:
		s.planExecutor.execute(ctx, s.githubClient)
	case Apply:
		s.applyExecutor.execute(ctx, s.githubClient)
	case Help:
		s.helpExecutor.execute(ctx, s.githubClient)
	default:
		ctx.Log.Err("failed to determine desired command, neither plan nor apply")
	}
}

func (s *CommandHandler) SetDeleteLockURL(f func(id string) (url string)) {
	s.planExecutor.DeleteLockURL = f
}

// recover logs and creates a comment on the pull request for panics
func (s *CommandHandler) recover(ctx *CommandContext) {
	if err := recover(); err != nil {
		stack := recovery.Stack(3)
		s.githubClient.CreateComment(ctx, fmt.Sprintf("**Error: goroutine panic. This is a bug.**\n```\n%s\n%s```", err, stack))
		ctx.Log.Err("PANIC: %s\n%s", err, stack)
	}
}
