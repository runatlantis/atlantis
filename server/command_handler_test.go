package server_test

import (
	"testing"
	"github.com/golang/mock/gomock"
	"github.com/hootsuite/atlantis/server/mocks"
	ghmocks "github.com/hootsuite/atlantis/github/mocks"
	"github.com/hootsuite/atlantis/server"
	gh "github.com/hootsuite/atlantis/github/fixtures"
	. "github.com/hootsuite/atlantis/testing_util"
	"github.com/mohae/deepcopy"
	"github.com/google/go-github/github"
	"github.com/hootsuite/atlantis/logging"
	"os"
	"log"
	"github.com/hootsuite/atlantis/models/fixtures"
	"errors"
)

func TestExecuteCommand_PullErr(t *testing.T) {
	t.Log("if getting the pull request fails nothing should continue")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	applier := mocks.NewMockExecutor(ctrl)
	helper := mocks.NewMockExecutor(ctrl)
	planner := mocks.NewMockPlanner(ctrl)
	parser := mocks.NewMockEventParsing(ctrl)
	ghClient := ghmocks.NewMockClient(ctrl)
	ch := server.CommandHandler{
		PlanExecutor:  planner,
		ApplyExecutor: applier,
		HelpExecutor:  helper,
		GithubClient:  ghClient,
		EventParser:   parser,
		Logger: logging.NewSimpleLogger("", log.New(os.Stderr, "", log.LstdFlags), false, logging.Debug),
	}

	ghClient.EXPECT().GetPullRequest(fixtures.Repo, fixtures.Pull.Num).Return(nil, nil, errors.New("err"))
	ch.ExecuteCommand(&server.CommandContext{
		BaseRepo: fixtures.Repo,
		Pull: fixtures.Pull,
	})
}

func TestExecuteCommand_ExtractErr(t *testing.T) {
	t.Log("if extracting data from the pull request fails nothing should continue")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	applier := mocks.NewMockExecutor(ctrl)
	helper := mocks.NewMockExecutor(ctrl)
	planner := mocks.NewMockPlanner(ctrl)
	parser := mocks.NewMockEventParsing(ctrl)
	ghClient := ghmocks.NewMockClient(ctrl)
	ch := server.CommandHandler{
		PlanExecutor:  planner,
		ApplyExecutor: applier,
		HelpExecutor:  helper,
		GithubClient:  ghClient,
		EventParser:   parser,
		Logger: logging.NewSimpleLogger("", log.New(os.Stderr, "", log.LstdFlags), false, logging.Debug),
	}

	pull := deepcopy.Copy(gh.Pull).(github.PullRequest)
	pull.State = github.String("open")
	ghClient.EXPECT().GetPullRequest(fixtures.Repo, fixtures.Pull.Num).Return(&pull, nil, nil)
	parser.EXPECT().ExtractPullData(&pull).Return(fixtures.Pull, fixtures.Repo, errors.New("err"))

	ch.ExecuteCommand(&server.CommandContext{
		BaseRepo: fixtures.Repo,
		Pull: fixtures.Pull,
	})
}

func TestExecuteCommand_ClosedPull(t *testing.T) {
	t.Log("if a command is run on a closed pull request atlantis should" +
		" comment saying that this is not allowed")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	applier := mocks.NewMockExecutor(ctrl)
	helper := mocks.NewMockExecutor(ctrl)
	planner := mocks.NewMockPlanner(ctrl)
	parser := mocks.NewMockEventParsing(ctrl)
	ghClient := ghmocks.NewMockClient(ctrl)
	ch := server.CommandHandler{
		PlanExecutor:  planner,
		ApplyExecutor: applier,
		HelpExecutor:  helper,
		GithubClient:  ghClient,
		EventParser:   parser,
		Logger: logging.NewSimpleLogger("", log.New(os.Stderr, "", log.LstdFlags), false, logging.Debug),
	}

	pull := deepcopy.Copy(gh.Pull).(github.PullRequest)
	pull.State = github.String("closed")
	ghClient.EXPECT().GetPullRequest(fixtures.Repo, fixtures.Pull.Num).Return(&pull, nil, nil)
	ghClient.EXPECT().CreateComment(fixtures.Repo, fixtures.Pull, "Atlantis commands can't be run on closed pull requests")

	ch.ExecuteCommand(&server.CommandContext{
		BaseRepo: fixtures.Repo,
		User: fixtures.User,
		Pull: fixtures.Pull,
		Command: &server.Command{
			Name: server.Plan,
		},
	})
}

func TestExecuteCommand_Executors(t *testing.T) {
	t.Log("should execute correct executor and fill in fields on ctx object")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	applier := mocks.NewMockExecutor(ctrl)
	helper := mocks.NewMockExecutor(ctrl)
	planner := mocks.NewMockPlanner(ctrl)
	parser := mocks.NewMockEventParsing(ctrl)
	ghClient := ghmocks.NewMockClient(ctrl)
	ch := server.CommandHandler{
		PlanExecutor:  planner,
		ApplyExecutor: applier,
		HelpExecutor:  helper,
		GithubClient:  ghClient,
		EventParser:   parser,
		Logger: logging.NewSimpleLogger("", log.New(os.Stderr, "", log.LstdFlags), false, logging.Debug),
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
		Command: &cmd,
	}

	// plan
	ghClient.EXPECT().GetPullRequest(fixtures.Repo, fixtures.Pull.Num).Return(&pull, nil, nil)
	parser.EXPECT().ExtractPullData(&pull).Return(fixtures.Pull, fixtures.Repo, nil)
	planner.EXPECT().Execute(gomock.Any()).Do(func (ctx *server.CommandContext) {
		// validate that the context was populated with expected data
		Equals(t, fixtures.Pull, ctx.Pull)
		Equals(t, fixtures.Repo, ctx.HeadRepo)
	})
	ch.ExecuteCommand(&baseCtx)
	ctrl.Finish()

	// apply
	cmd.Name = server.Apply
	ghClient.EXPECT().GetPullRequest(fixtures.Repo, fixtures.Pull.Num).Return(&pull, nil, nil)
	parser.EXPECT().ExtractPullData(&pull).Return(fixtures.Pull, fixtures.Repo, nil)
	applier.EXPECT().Execute(gomock.Any()).Do(func (ctx *server.CommandContext) {
		// validate that the context was populated with expected data
		Equals(t, fixtures.Pull, ctx.Pull)
		Equals(t, fixtures.Repo, ctx.HeadRepo)
	})
	ch.ExecuteCommand(&baseCtx)
	ctrl.Finish()

	// help
	cmd.Name = server.Help
	ghClient.EXPECT().GetPullRequest(fixtures.Repo, fixtures.Pull.Num).Return(&pull, nil, nil)
	parser.EXPECT().ExtractPullData(&pull).Return(fixtures.Pull, fixtures.Repo, nil)
	helper.EXPECT().Execute(gomock.Any()).Do(func (ctx *server.CommandContext) {
		// validate that the context was populated with expected data
		Equals(t, fixtures.Pull, ctx.Pull)
		Equals(t, fixtures.Repo, ctx.HeadRepo)
	})
	ch.ExecuteCommand(&baseCtx)
	ctrl.Finish()
}
