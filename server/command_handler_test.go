package server_test

import (
	"errors"
	"testing"

	"reflect"

	"strings"

	"github.com/google/go-github/github"
	gh "github.com/hootsuite/atlantis/github/fixtures"
	ghmocks "github.com/hootsuite/atlantis/github/mocks"
	"github.com/hootsuite/atlantis/logging"
	"github.com/hootsuite/atlantis/models/fixtures"
	"github.com/hootsuite/atlantis/server"
	"github.com/hootsuite/atlantis/server/mocks"
	. "github.com/hootsuite/atlantis/testing_util"
	"github.com/mohae/deepcopy"
	. "github.com/petergtz/pegomock"
)

var applier *mocks.MockExecutor
var helper *mocks.MockExecutor
var planner *mocks.MockPlanner
var eventParsing *mocks.MockEventParsing
var ghClient *ghmocks.MockClient
var ghStatus *mocks.MockGHStatusUpdater
var envLocker *mocks.MockEnvLocker
var ch server.CommandHandler

func setup(t *testing.T) {
	RegisterMockTestingT(t)
	applier = mocks.NewMockExecutor()
	helper = mocks.NewMockExecutor()
	planner = mocks.NewMockPlanner()
	eventParsing = mocks.NewMockEventParsing()
	ghClient = ghmocks.NewMockClient()
	ghStatus = mocks.NewMockGHStatusUpdater()
	envLocker = mocks.NewMockEnvLocker()
	ch = server.CommandHandler{
		PlanExecutor:      planner,
		ApplyExecutor:     applier,
		HelpExecutor:      helper,
		GHClient:          ghClient,
		GHStatus:          ghStatus,
		EventParser:       eventParsing,
		EnvLocker:         envLocker,
		GHCommentRenderer: &server.GithubCommentRenderer{},
		Logger:            logging.NewNoopLogger(),
	}
}

func TestExecuteCommand_LogPanics(t *testing.T) {
	t.Log("if there is a panic it is commented back on the pull request")
	setup(t)
	When(ghClient.GetPullRequest(fixtures.Repo, fixtures.Pull.Num)).ThenPanic("panic")
	ch.ExecuteCommand(&server.CommandContext{
		BaseRepo: fixtures.Repo,
		Pull:     fixtures.Pull,
	})
	_, _, comment := ghClient.VerifyWasCalledOnce().CreateComment(AnyRepo(), AnyPullRequest(), AnyString()).GetCapturedArguments()
	Assert(t, strings.Contains(comment, "Error: goroutine panic"), "comment should be about a goroutine panic")
}

func TestExecuteCommand_PullErr(t *testing.T) {
	t.Log("if getting the pull request fails nothing should continue")
	setup(t)
	When(ghClient.GetPullRequest(fixtures.Repo, fixtures.Pull.Num)).ThenReturn(nil, nil, errors.New("err"))
	ch.ExecuteCommand(&server.CommandContext{
		BaseRepo: fixtures.Repo,
		Pull:     fixtures.Pull,
	})
}

func TestExecuteCommand_ExtractErr(t *testing.T) {
	t.Log("if extracting data from the pull request fails nothing should continue")
	setup(t)
	pull := deepcopy.Copy(gh.Pull).(github.PullRequest)
	pull.State = github.String("open")
	When(ghClient.GetPullRequest(fixtures.Repo, fixtures.Pull.Num)).ThenReturn(&pull, nil, nil)
	When(eventParsing.ExtractPullData(&pull)).ThenReturn(fixtures.Pull, fixtures.Repo, errors.New("err"))

	ch.ExecuteCommand(&server.CommandContext{
		BaseRepo: fixtures.Repo,
		Pull:     fixtures.Pull,
	})
}

func TestExecuteCommand_ClosedPull(t *testing.T) {
	t.Log("if a command is run on a closed pull request atlantis should" +
		" comment saying that this is not allowed")
	setup(t)
	pull := deepcopy.Copy(gh.Pull).(github.PullRequest)
	pull.State = github.String("closed")
	When(ghClient.GetPullRequest(fixtures.Repo, fixtures.Pull.Num)).ThenReturn(&pull, nil, nil)

	ch.ExecuteCommand(&server.CommandContext{
		BaseRepo: fixtures.Repo,
		User:     fixtures.User,
		Pull:     fixtures.Pull,
		Command: &server.Command{
			Name: server.Plan,
		},
	})
	ghClient.VerifyWasCalledOnce().CreateComment(fixtures.Repo, fixtures.Pull, "Atlantis commands can't be run on closed pull requests")
}

func TestExecuteCommand_EnvLocked(t *testing.T) {
	t.Log("if the environment is locked, should comment back on the pull")
	setup(t)
	pull := deepcopy.Copy(gh.Pull).(github.PullRequest)
	pull.State = github.String("open")
	cmd := server.Command{
		Name:        server.Plan,
		Environment: "env",
	}
	baseCtx := server.CommandContext{
		BaseRepo: fixtures.Repo,
		User:     fixtures.User,
		Pull:     fixtures.Pull,
		Command:  &cmd,
	}

	When(ghClient.GetPullRequest(fixtures.Repo, fixtures.Pull.Num)).ThenReturn(&pull, nil, nil)
	When(eventParsing.ExtractPullData(&pull)).ThenReturn(fixtures.Pull, fixtures.Repo, nil)
	When(envLocker.TryLock(fixtures.Repo.FullName, cmd.Environment, fixtures.Pull.Num)).ThenReturn(false)
	ch.ExecuteCommand(&baseCtx)

	msg := "The env environment is currently locked by another" +
		" command that is running for this pull request." +
		" Wait until the previous command is complete and try again."
	ghStatus.VerifyWasCalledOnce().Update(fixtures.Repo, fixtures.Pull, server.Pending, &cmd)
	ghStatus.VerifyWasCalledOnce().UpdateProjectResult(&baseCtx, server.CommandResponse{Failure: msg})
	ghClient.VerifyWasCalledOnce().CreateComment(fixtures.Repo, fixtures.Pull,
		"**Plan Failed**: "+msg+"\n\n")
}

func TestExecuteCommand_FullRun(t *testing.T) {
	t.Log("when running a plan, apply or help should comment")
	pull := deepcopy.Copy(gh.Pull).(github.PullRequest)
	pull.State = github.String("open")
	cmdResponse := server.CommandResponse{}
	for _, c := range []server.CommandName{server.Help, server.Plan, server.Apply} {
		setup(t)
		cmd := server.Command{
			Name:        c,
			Environment: "env",
		}
		baseCtx := server.CommandContext{
			BaseRepo: fixtures.Repo,
			User:     fixtures.User,
			Pull:     fixtures.Pull,
			Command:  &cmd,
		}
		When(ghClient.GetPullRequest(fixtures.Repo, fixtures.Pull.Num)).ThenReturn(&pull, nil, nil)
		When(eventParsing.ExtractPullData(&pull)).ThenReturn(fixtures.Pull, fixtures.Repo, nil)
		When(envLocker.TryLock(fixtures.Repo.FullName, cmd.Environment, fixtures.Pull.Num)).ThenReturn(true)
		switch c {
		case server.Help:
			When(helper.Execute(AnyCommandContext())).ThenReturn(cmdResponse)
		case server.Plan:
			When(planner.Execute(AnyCommandContext())).ThenReturn(cmdResponse)
		case server.Apply:
			When(applier.Execute(AnyCommandContext())).ThenReturn(cmdResponse)
		}

		ch.ExecuteCommand(&baseCtx)

		ghStatus.VerifyWasCalledOnce().Update(fixtures.Repo, fixtures.Pull, server.Pending, &cmd)
		ghStatus.VerifyWasCalledOnce().UpdateProjectResult(&baseCtx, cmdResponse)
		ghClient.VerifyWasCalledOnce().CreateComment(AnyRepo(), AnyPullRequest(), AnyString())
		envLocker.VerifyWasCalledOnce().Unlock(fixtures.Repo.FullName, cmd.Environment, fixtures.Pull.Num)
	}
}

func AnyCommandContext() *server.CommandContext {
	RegisterMatcher(NewAnyMatcher(reflect.TypeOf(&server.CommandContext{})))
	return &server.CommandContext{}
}
