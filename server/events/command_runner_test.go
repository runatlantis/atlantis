// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.

package events_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/server/logging"

	"github.com/google/go-github/v28/github"
	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/mocks/matchers"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/models/fixtures"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	logmocks "github.com/runatlantis/atlantis/server/logging/mocks"
	. "github.com/runatlantis/atlantis/testing"
)

var projectCommandBuilder *mocks.MockProjectCommandBuilder
var projectCommandRunner *mocks.MockProjectCommandRunner
var eventParsing *mocks.MockEventParsing
var azuredevopsGetter *mocks.MockAzureDevopsPullGetter
var githubGetter *mocks.MockGithubPullGetter
var gitlabGetter *mocks.MockGitlabMergeRequestGetter
var ch events.DefaultCommandRunner
var pullLogger *logging.SimpleLogger
var workingDir events.WorkingDir
var pendingPlanFinder *mocks.MockPendingPlanFinder

func setup(t *testing.T) *vcsmocks.MockClient {
	RegisterMockTestingT(t)
	projectCommandBuilder = mocks.NewMockProjectCommandBuilder()
	eventParsing = mocks.NewMockEventParsing()
	vcsClient := vcsmocks.NewMockClient()
	githubGetter = mocks.NewMockGithubPullGetter()
	gitlabGetter = mocks.NewMockGitlabMergeRequestGetter()
	azuredevopsGetter = mocks.NewMockAzureDevopsPullGetter()
	logger := logmocks.NewMockSimpleLogging()
	pullLogger = logging.NewSimpleLogger("runatlantis/atlantis#1", true, logging.Info)
	projectCommandRunner = mocks.NewMockProjectCommandRunner()
	workingDir = mocks.NewMockWorkingDir()
	pendingPlanFinder = mocks.NewMockPendingPlanFinder()
	When(logger.GetLevel()).ThenReturn(logging.Info)
	When(logger.NewLogger("runatlantis/atlantis#1", true, logging.Info)).
		ThenReturn(pullLogger)
	ch = events.DefaultCommandRunner{
		VCSClient:                vcsClient,
		CommitStatusUpdater:      &events.DefaultCommitStatusUpdater{vcsClient},
		EventParser:              eventParsing,
		MarkdownRenderer:         &events.MarkdownRenderer{},
		GithubPullGetter:         githubGetter,
		GitlabMergeRequestGetter: gitlabGetter,
		AzureDevopsPullGetter:    azuredevopsGetter,
		Logger:                   logger,
		AllowForkPRs:             false,
		AllowForkPRsFlag:         "allow-fork-prs-flag",
		ProjectCommandBuilder:    projectCommandBuilder,
		ProjectCommandRunner:     projectCommandRunner,
		PendingPlanFinder:        pendingPlanFinder,
		WorkingDir:               workingDir,
		DisableApplyAll:          false,
	}
	return vcsClient
}

func TestRunCommentCommand_LogPanics(t *testing.T) {
	t.Log("if there is a panic it is commented back on the pull request")
	vcsClient := setup(t)
	When(githubGetter.GetPullRequest(fixtures.GithubRepo, fixtures.Pull.Num)).ThenPanic("OMG PANIC!!!")
	ch.RunCommentCommand(fixtures.GithubRepo, &fixtures.GithubRepo, nil, fixtures.User, 1, &events.CommentCommand{Name: models.PlanCommand})
	_, _, comment := vcsClient.VerifyWasCalledOnce().CreateComment(matchers.AnyModelsRepo(), AnyInt(), AnyString()).GetCapturedArguments()
	Assert(t, strings.Contains(comment, "Error: goroutine panic"), fmt.Sprintf("comment should be about a goroutine panic but was %q", comment))
}

func TestRunCommentCommand_NoGithubPullGetter(t *testing.T) {
	t.Log("if DefaultCommandRunner was constructed with a nil GithubPullGetter an error should be logged")
	setup(t)
	ch.GithubPullGetter = nil
	ch.RunCommentCommand(fixtures.GithubRepo, &fixtures.GithubRepo, nil, fixtures.User, 1, nil)
	Equals(t, "[EROR] Atlantis not configured to support GitHub\n", pullLogger.History.String())
}

func TestRunCommentCommand_NoGitlabMergeGetter(t *testing.T) {
	t.Log("if DefaultCommandRunner was constructed with a nil GitlabMergeRequestGetter an error should be logged")
	setup(t)
	ch.GitlabMergeRequestGetter = nil
	ch.RunCommentCommand(fixtures.GitlabRepo, &fixtures.GitlabRepo, nil, fixtures.User, 1, nil)
	Equals(t, "[EROR] Atlantis not configured to support GitLab\n", pullLogger.History.String())
}

func TestRunCommentCommand_GithubPullErr(t *testing.T) {
	t.Log("if getting the github pull request fails an error should be logged")
	vcsClient := setup(t)
	When(githubGetter.GetPullRequest(fixtures.GithubRepo, fixtures.Pull.Num)).ThenReturn(nil, errors.New("err"))
	ch.RunCommentCommand(fixtures.GithubRepo, &fixtures.GithubRepo, nil, fixtures.User, fixtures.Pull.Num, nil)
	vcsClient.VerifyWasCalledOnce().CreateComment(fixtures.GithubRepo, fixtures.Pull.Num, "`Error: making pull request API call to GitHub: err`")
}

func TestRunCommentCommand_GitlabMergeRequestErr(t *testing.T) {
	t.Log("if getting the gitlab merge request fails an error should be logged")
	vcsClient := setup(t)
	When(gitlabGetter.GetMergeRequest(fixtures.GitlabRepo.FullName, fixtures.Pull.Num)).ThenReturn(nil, errors.New("err"))
	ch.RunCommentCommand(fixtures.GitlabRepo, &fixtures.GitlabRepo, nil, fixtures.User, fixtures.Pull.Num, nil)
	vcsClient.VerifyWasCalledOnce().CreateComment(fixtures.GitlabRepo, fixtures.Pull.Num, "`Error: making merge request API call to GitLab: err`")
}

func TestRunCommentCommand_GithubPullParseErr(t *testing.T) {
	t.Log("if parsing the returned github pull request fails an error should be logged")
	vcsClient := setup(t)
	var pull github.PullRequest
	When(githubGetter.GetPullRequest(fixtures.GithubRepo, fixtures.Pull.Num)).ThenReturn(&pull, nil)
	When(eventParsing.ParseGithubPull(&pull)).ThenReturn(fixtures.Pull, fixtures.GithubRepo, fixtures.GitlabRepo, errors.New("err"))

	ch.RunCommentCommand(fixtures.GithubRepo, &fixtures.GithubRepo, nil, fixtures.User, fixtures.Pull.Num, nil)
	vcsClient.VerifyWasCalledOnce().CreateComment(fixtures.GithubRepo, fixtures.Pull.Num, "`Error: extracting required fields from comment data: err`")
}

func TestRunCommentCommand_ForkPRDisabled(t *testing.T) {
	t.Log("if a command is run on a forked pull request and this is disabled atlantis should" +
		" comment saying that this is not allowed")
	vcsClient := setup(t)
	ch.AllowForkPRs = false // by default it's false so don't need to reset
	var pull github.PullRequest
	modelPull := models.PullRequest{State: models.OpenPullState}
	When(githubGetter.GetPullRequest(fixtures.GithubRepo, fixtures.Pull.Num)).ThenReturn(&pull, nil)

	headRepo := fixtures.GithubRepo
	headRepo.FullName = "forkrepo/atlantis"
	headRepo.Owner = "forkrepo"
	When(eventParsing.ParseGithubPull(&pull)).ThenReturn(modelPull, modelPull.BaseRepo, headRepo, nil)

	ch.RunCommentCommand(fixtures.GithubRepo, nil, nil, fixtures.User, fixtures.Pull.Num, nil)
	vcsClient.VerifyWasCalledOnce().CreateComment(fixtures.GithubRepo, modelPull.Num, "Atlantis commands can't be run on fork pull requests. To enable, set --"+ch.AllowForkPRsFlag)
}

func TestRunCommentCommand_DisableApplyAllDisabled(t *testing.T) {
	t.Log("if \"atlantis apply\" is run and this is disabled atlantis should" +
		" comment saying that this is not allowed")
	vcsClient := setup(t)
	ch.DisableApplyAll = true
	modelPull := models.PullRequest{State: models.OpenPullState}
	ch.RunCommentCommand(fixtures.GithubRepo, nil, nil, fixtures.User, modelPull.Num, &events.CommentCommand{Name: models.ApplyCommand})
	vcsClient.VerifyWasCalledOnce().CreateComment(fixtures.GithubRepo, modelPull.Num, "**Error:** Running `atlantis apply` without flags is disabled. You must specify which project to apply via the `-d <dir>`, `-w <workspace>` or `-p <project name>` flags.")
}

func TestRunCommentCommand_ClosedPull(t *testing.T) {
	t.Log("if a command is run on a closed pull request atlantis should" +
		" comment saying that this is not allowed")
	vcsClient := setup(t)
	pull := &github.PullRequest{
		State: github.String("closed"),
	}
	modelPull := models.PullRequest{State: models.ClosedPullState}
	When(githubGetter.GetPullRequest(fixtures.GithubRepo, fixtures.Pull.Num)).ThenReturn(pull, nil)
	When(eventParsing.ParseGithubPull(pull)).ThenReturn(modelPull, modelPull.BaseRepo, fixtures.GithubRepo, nil)

	ch.RunCommentCommand(fixtures.GithubRepo, &fixtures.GithubRepo, nil, fixtures.User, fixtures.Pull.Num, nil)
	vcsClient.VerifyWasCalledOnce().CreateComment(fixtures.GithubRepo, modelPull.Num, "Atlantis commands can't be run on closed pull requests")
}

// Test that if one plan fails and we are using automerge, that
// we delete the plans.
func TestRunAutoplanCommand_DeletePlans(t *testing.T) {
	setup(t)
	ch.GlobalAutomerge = true
	defer func() { ch.GlobalAutomerge = false }()

	When(projectCommandBuilder.BuildAutoplanCommands(matchers.AnyPtrToEventsCommandContext())).
		ThenReturn([]models.ProjectCommandContext{
			{},
			{},
		}, nil)
	callCount := 0
	When(projectCommandRunner.Plan(matchers.AnyModelsProjectCommandContext())).Then(func(_ []Param) ReturnValues {
		if callCount == 0 {
			// The first call, we return a successful result.
			callCount++
			return ReturnValues{
				models.ProjectResult{
					PlanSuccess: &models.PlanSuccess{},
				},
			}
		}
		// The second call, we return a failed result.
		return ReturnValues{
			models.ProjectResult{
				Error: errors.New("err"),
			},
		}
	})
	tmp, cleanup := TempDir(t)
	defer cleanup()

	When(workingDir.GetPullDir(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())).
		ThenReturn(tmp, nil)
	ch.RunAutoplanCommand(fixtures.GithubRepo, fixtures.GithubRepo, fixtures.Pull, fixtures.User)
	pendingPlanFinder.VerifyWasCalledOnce().DeletePlans(tmp)
}
