package server_test

import (
	"errors"
	"log"
	"os"
	"testing"

	"reflect"

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

func TestExecuteCommand_PullErr(t *testing.T) {
	t.Log("if getting the pull request fails nothing should continue")

	RegisterMockTestingT(t)
	applier := mocks.NewMockExecutor()
	helper := mocks.NewMockExecutor()
	planner := mocks.NewMockPlanner()
	parser := mocks.NewMockEventParsing()
	ghClient := ghmocks.NewMockClient()
	ch := server.CommandHandler{
		PlanExecutor:  planner,
		ApplyExecutor: applier,
		HelpExecutor:  helper,
		GithubClient:  ghClient,
		EventParser:   parser,
		Logger:        logging.NewSimpleLogger("", log.New(os.Stderr, "", log.LstdFlags), false, logging.Debug),
	}

	When(ghClient.GetPullRequest(fixtures.Repo, fixtures.Pull.Num)).ThenReturn(nil, nil, errors.New("err"))
	ch.ExecuteCommand(&server.CommandContext{
		BaseRepo: fixtures.Repo,
		Pull:     fixtures.Pull,
	})
}

func TestExecuteCommand_ExtractErr(t *testing.T) {
	t.Log("if extracting data from the pull request fails nothing should continue")
	RegisterMockTestingT(t)
	applier := mocks.NewMockExecutor()
	helper := mocks.NewMockExecutor()
	planner := mocks.NewMockPlanner()
	parser := mocks.NewMockEventParsing()
	ghClient := ghmocks.NewMockClient()
	ch := server.CommandHandler{
		PlanExecutor:  planner,
		ApplyExecutor: applier,
		HelpExecutor:  helper,
		GithubClient:  ghClient,
		EventParser:   parser,
		Logger:        logging.NewSimpleLogger("", log.New(os.Stderr, "", log.LstdFlags), false, logging.Debug),
	}

	pull := deepcopy.Copy(gh.Pull).(github.PullRequest)
	pull.State = github.String("open")
	When(ghClient.GetPullRequest(fixtures.Repo, fixtures.Pull.Num)).ThenReturn(&pull, nil, nil)
	When(parser.ExtractPullData(&pull)).ThenReturn(fixtures.Pull, fixtures.Repo, errors.New("err"))

	ch.ExecuteCommand(&server.CommandContext{
		BaseRepo: fixtures.Repo,
		Pull:     fixtures.Pull,
	})
}

func TestExecuteCommand_ClosedPull(t *testing.T) {
	t.Log("if a command is run on a closed pull request atlantis should" +
		" comment saying that this is not allowed")
	RegisterMockTestingT(t)
	applier := mocks.NewMockExecutor()
	helper := mocks.NewMockExecutor()
	planner := mocks.NewMockPlanner()
	parser := mocks.NewMockEventParsing()
	ghClient := ghmocks.NewMockClient()
	ch := server.CommandHandler{
		PlanExecutor:  planner,
		ApplyExecutor: applier,
		HelpExecutor:  helper,
		GithubClient:  ghClient,
		EventParser:   parser,
		Logger:        logging.NewSimpleLogger("", log.New(os.Stderr, "", log.LstdFlags), false, logging.Debug),
	}

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

func TestExecuteCommand_Executors(t *testing.T) {
	t.Log("should execute correct executor and fill in fields on ctx object")
	RegisterMockTestingT(t)
	applier := mocks.NewMockExecutor()
	helper := mocks.NewMockExecutor()
	planner := mocks.NewMockPlanner()
	parser := mocks.NewMockEventParsing()
	ghClient := ghmocks.NewMockClient()
	ch := server.CommandHandler{
		PlanExecutor:  planner,
		ApplyExecutor: applier,
		HelpExecutor:  helper,
		GithubClient:  ghClient,
		EventParser:   parser,
		Logger:        logging.NewSimpleLogger("", log.New(os.Stderr, "", log.LstdFlags), false, logging.Debug),
	}
	pull := deepcopy.Copy(gh.Pull).(github.PullRequest)
	pull.State = github.String("open")

	cmd := server.Command{
		Name: server.Plan,
	}
	baseCtx := server.CommandContext{
		BaseRepo: fixtures.Repo,
		User:     fixtures.User,
		Pull:     fixtures.Pull,
		Command:  &cmd,
	}

	// plan
	When(ghClient.GetPullRequest(fixtures.Repo, fixtures.Pull.Num)).ThenReturn(&pull, nil, nil)
	When(parser.ExtractPullData(&pull)).ThenReturn(fixtures.Pull, fixtures.Repo, nil)
	ch.ExecuteCommand(&baseCtx)
	ctx := planner.VerifyWasCalledOnce().Execute(AnyCommandContext()).GetCapturedArguments()
	Equals(t, fixtures.Pull, ctx.Pull)
	Equals(t, fixtures.Repo, ctx.HeadRepo)

	// apply
	cmd.Name = server.Apply
	applier = mocks.NewMockExecutor()
	ch.ExecuteCommand(&baseCtx)
	ctx = planner.VerifyWasCalledOnce().Execute(AnyCommandContext()).GetCapturedArguments()
	Equals(t, fixtures.Pull, ctx.Pull)
	Equals(t, fixtures.Repo, ctx.HeadRepo)

	// help
	cmd.Name = server.Help
	ch.ExecuteCommand(&baseCtx)
	ctx = planner.VerifyWasCalledOnce().Execute(AnyCommandContext()).GetCapturedArguments()
	Equals(t, fixtures.Pull, ctx.Pull)
	Equals(t, fixtures.Repo, ctx.HeadRepo)
}

func AnyCommandContext() *server.CommandContext {
	RegisterMatcher(NewAnyMatcher(reflect.TypeOf(&server.CommandContext{})))
	return &server.CommandContext{}
}
