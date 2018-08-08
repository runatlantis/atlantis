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
//
package events_test

import (
	"bytes"
	"errors"
	"log"
	"strings"
	"testing"

	"github.com/google/go-github/github"
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
var eventParsing *mocks.MockEventParsing
var vcsClient *vcsmocks.MockClientProxy
var ghStatus *mocks.MockCommitStatusUpdater
var githubGetter *mocks.MockGithubPullGetter
var gitlabGetter *mocks.MockGitlabMergeRequestGetter
var ch events.DefaultCommandRunner
var logBytes *bytes.Buffer

func setup(t *testing.T) {
	RegisterMockTestingT(t)
	projectCommandBuilder = mocks.NewMockProjectCommandBuilder()
	eventParsing = mocks.NewMockEventParsing()
	ghStatus = mocks.NewMockCommitStatusUpdater()
	vcsClient = vcsmocks.NewMockClientProxy()
	githubGetter = mocks.NewMockGithubPullGetter()
	gitlabGetter = mocks.NewMockGitlabMergeRequestGetter()
	logger := logmocks.NewMockSimpleLogging()
	logBytes = new(bytes.Buffer)
	projectCommandRunner := mocks.NewMockProjectCommandRunner()
	When(logger.Underlying()).ThenReturn(log.New(logBytes, "", 0))
	ch = events.DefaultCommandRunner{
		VCSClient:                vcsClient,
		CommitStatusUpdater:      ghStatus,
		EventParser:              eventParsing,
		MarkdownRenderer:         &events.MarkdownRenderer{},
		GithubPullGetter:         githubGetter,
		GitlabMergeRequestGetter: gitlabGetter,
		Logger:                logger,
		AllowForkPRs:          false,
		AllowForkPRsFlag:      "allow-fork-prs-flag",
		ProjectCommandBuilder: projectCommandBuilder,
		ProjectCommandRunner:  projectCommandRunner,
	}
}

func TestRunCommentCommand_LogPanics(t *testing.T) {
	t.Log("if there is a panic it is commented back on the pull request")
	setup(t)
	ch.AllowForkPRs = true // Lets us get to the panic code.
	defer func() { ch.AllowForkPRs = false }()
	When(ghStatus.Update(fixtures.GithubRepo, fixtures.Pull, models.PendingCommitStatus, events.PlanCommand)).ThenPanic("panic")
	ch.RunCommentCommand(fixtures.GithubRepo, &fixtures.GithubRepo, nil, fixtures.User, 1, nil)
	_, _, comment := vcsClient.VerifyWasCalledOnce().CreateComment(matchers.AnyModelsRepo(), AnyInt(), AnyString()).GetCapturedArguments()
	Assert(t, strings.Contains(comment, "Error: goroutine panic"), "comment should be about a goroutine panic")
}

func TestRunCommentCommand_NoGithubPullGetter(t *testing.T) {
	t.Log("if DefaultCommandRunner was constructed with a nil GithubPullGetter an error should be logged")
	setup(t)
	ch.GithubPullGetter = nil
	ch.RunCommentCommand(fixtures.GithubRepo, &fixtures.GithubRepo, nil, fixtures.User, 1, nil)
	Equals(t, "[ERROR] runatlantis/atlantis#1: Atlantis not configured to support GitHub\n", logBytes.String())
}

func TestRunCommentCommand_NoGitlabMergeGetter(t *testing.T) {
	t.Log("if DefaultCommandRunner was constructed with a nil GitlabMergeRequestGetter an error should be logged")
	setup(t)
	ch.GitlabMergeRequestGetter = nil
	ch.RunCommentCommand(fixtures.GitlabRepo, &fixtures.GitlabRepo, nil, fixtures.User, 1, nil)
	Equals(t, "[ERROR] runatlantis/atlantis#1: Atlantis not configured to support GitLab\n", logBytes.String())
}

func TestRunCommentCommand_GithubPullErr(t *testing.T) {
	t.Log("if getting the github pull request fails an error should be logged")
	setup(t)
	When(githubGetter.GetPullRequest(fixtures.GithubRepo, fixtures.Pull.Num)).ThenReturn(nil, errors.New("err"))
	ch.RunCommentCommand(fixtures.GithubRepo, &fixtures.GithubRepo, nil, fixtures.User, fixtures.Pull.Num, nil)
	Equals(t, "[ERROR] runatlantis/atlantis#1: Making pull request API call to GitHub: err\n", logBytes.String())
}

func TestRunCommentCommand_GitlabMergeRequestErr(t *testing.T) {
	t.Log("if getting the gitlab merge request fails an error should be logged")
	setup(t)
	When(gitlabGetter.GetMergeRequest(fixtures.GithubRepo.FullName, fixtures.Pull.Num)).ThenReturn(nil, errors.New("err"))
	ch.RunCommentCommand(fixtures.GitlabRepo, &fixtures.GitlabRepo, nil, fixtures.User, fixtures.Pull.Num, nil)
	Equals(t, "[ERROR] runatlantis/atlantis#1: Making merge request API call to GitLab: err\n", logBytes.String())
}

func TestRunCommentCommand_GithubPullParseErr(t *testing.T) {
	t.Log("if parsing the returned github pull request fails an error should be logged")
	setup(t)
	var pull github.PullRequest
	When(githubGetter.GetPullRequest(fixtures.GithubRepo, fixtures.Pull.Num)).ThenReturn(&pull, nil)
	When(eventParsing.ParseGithubPull(&pull)).ThenReturn(fixtures.Pull, fixtures.GithubRepo, fixtures.GitlabRepo, errors.New("err"))

	ch.RunCommentCommand(fixtures.GithubRepo, &fixtures.GithubRepo, nil, fixtures.User, fixtures.Pull.Num, nil)
	Equals(t, "[ERROR] runatlantis/atlantis#1: Extracting required fields from comment data: err\n", logBytes.String())
}

func TestRunCommentCommand_ForkPRDisabled(t *testing.T) {
	t.Log("if a command is run on a forked pull request and this is disabled atlantis should" +
		" comment saying that this is not allowed")
	setup(t)
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

func TestRunCommentCommand_ClosedPull(t *testing.T) {
	t.Log("if a command is run on a closed pull request atlantis should" +
		" comment saying that this is not allowed")
	setup(t)
	pull := &github.PullRequest{
		State: github.String("closed"),
	}
	modelPull := models.PullRequest{State: models.ClosedPullState}
	When(githubGetter.GetPullRequest(fixtures.GithubRepo, fixtures.Pull.Num)).ThenReturn(pull, nil)
	When(eventParsing.ParseGithubPull(pull)).ThenReturn(modelPull, modelPull.BaseRepo, fixtures.GithubRepo, nil)

	ch.RunCommentCommand(fixtures.GithubRepo, &fixtures.GithubRepo, nil, fixtures.User, fixtures.Pull.Num, nil)
	vcsClient.VerifyWasCalledOnce().CreateComment(fixtures.GithubRepo, modelPull.Num, "Atlantis commands can't be run on closed pull requests")
}
