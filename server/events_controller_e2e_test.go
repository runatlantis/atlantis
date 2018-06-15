package server_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-github/github"
	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/locking"
	"github.com/runatlantis/atlantis/server/events/locking/boltdb"
	"github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/mocks/matchers"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/runtime"
	"github.com/runatlantis/atlantis/server/events/terraform"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/events/webhooks"
	"github.com/runatlantis/atlantis/server/events/yaml"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func Test(t *testing.T) {
	RegisterMockTestingT(t)

	// Config.
	allowForkPRs := false
	dataDir, cleanup := TempDir(t)
	defer cleanup()

	// Mocks.
	e2eVCSClient := vcsmocks.NewMockClientProxy()
	e2eStatusUpdater := mocks.NewMockCommitStatusUpdater()
	e2eGithubGetter := mocks.NewMockGithubPullGetter()
	e2eGitlabGetter := mocks.NewMockGitlabMergeRequestGetter()
	e2eWorkspace := mocks.NewMockAtlantisWorkspace()

	// Real dependencies.
	logger := logging.NewSimpleLogger("server", nil, true, logging.Debug)
	eventParser := &events.EventParser{
		GithubUser:  "github-user",
		GithubToken: "github-token",
		GitlabUser:  "gitlab-user",
		GitlabToken: "gitlab-token",
	}
	commentParser := &events.CommentParser{
		GithubUser:  "github-user",
		GithubToken: "github-token",
		GitlabUser:  "gitlab-user",
		GitlabToken: "gitlab-token",
	}
	terraformClient, err := terraform.NewClient(dataDir)
	Ok(t, err)
	boltdb, err := boltdb.New(dataDir)
	Ok(t, err)
	lockingClient := locking.NewClient(boltdb)
	projectLocker := &events.DefaultProjectLocker{
		Locker: lockingClient,
	}

	defaultTFVersion := terraformClient.Version()
	commandHandler := &events.CommandHandler{
		EventParser:              eventParser,
		VCSClient:                e2eVCSClient,
		GithubPullGetter:         e2eGithubGetter,
		GitlabMergeRequestGetter: e2eGitlabGetter,
		CommitStatusUpdater:      e2eStatusUpdater,
		AtlantisWorkspaceLocker:  events.NewDefaultAtlantisWorkspaceLocker(),
		MarkdownRenderer:         &events.MarkdownRenderer{},
		Logger:                   logger,
		AllowForkPRs:             allowForkPRs,
		AllowForkPRsFlag:         "allow-fork-prs",
		PullRequestOperator: &events.DefaultPullRequestOperator{
			TerraformExecutor: terraformClient,
			DefaultTFVersion:  defaultTFVersion,
			ParserValidator:   &yaml.ParserValidator{},
			ProjectFinder:     &events.DefaultProjectFinder{},
			VCSClient:         e2eVCSClient,
			Workspace:         e2eWorkspace,
			ProjectOperator: events.ProjectOperator{
				Locker:           projectLocker,
				LockURLGenerator: &mockLockURLGenerator{},
				InitStepOperator: runtime.InitStepOperator{
					TerraformExecutor: terraformClient,
					DefaultTFVersion:  defaultTFVersion,
				},
				PlanStepOperator: runtime.PlanStepOperator{
					TerraformExecutor: terraformClient,
					DefaultTFVersion:  defaultTFVersion,
				},
				ApplyStepOperator: runtime.ApplyStepOperator{
					TerraformExecutor: terraformClient,
				},
				RunStepOperator: runtime.RunStepOperator{},
				ApprovalOperator: runtime.ApprovalOperator{
					VCSClient: e2eVCSClient,
				},
				Workspace: e2eWorkspace,
				Webhooks:  &mockWebhookSender{},
			},
		},
	}

	ctrl := server.EventsController{
		TestingMode:            true,
		CommandRunner:          commandHandler,
		PullCleaner:            nil,
		Logger:                 logger,
		Parser:                 eventParser,
		CommentParser:          commentParser,
		GithubWebHookSecret:    nil,
		GithubRequestValidator: &server.DefaultGithubRequestValidator{},
		GitlabRequestParser:    &server.DefaultGitlabRequestParser{},
		GitlabWebHookSecret:    nil,
		RepoWhitelist: &events.RepoWhitelist{
			Whitelist: "*",
		},
		SupportedVCSHosts: []models.VCSHostType{models.Gitlab, models.Github},
		VCSClient:         e2eVCSClient,
		AtlantisGithubUser: models.User{
			Username: "atlantisbot",
		},
		AtlantisGitlabUser: models.User{
			Username: "atlantisbot",
		},
	}

	// Test GitHub Post
	req := GitHubCommentEvent(t, "atlantis plan")
	w := httptest.NewRecorder()
	When(e2eGithubGetter.GetPullRequest(AnyRepo(), AnyInt())).ThenReturn(GitHubPullRequestParsed(), nil)
	testRepoDir, err := filepath.Abs("testfixtures/test-repos/simple")
	Ok(t, err)
	When(e2eWorkspace.Clone(matchers.AnyPtrToLoggingSimpleLogger(), AnyRepo(), AnyRepo(), matchers.AnyModelsPullRequest(), AnyString())).ThenReturn(testRepoDir, nil)
	// Clean up .terraform and plan files when we're done.
	defer func() {
		os.RemoveAll(filepath.Join(testRepoDir, ".terraform"))
		planFiles, _ := filepath.Glob(testRepoDir + "/*.tfplan")
		for _, file := range planFiles {
			os.Remove(file)
		}
	}()

	ctrl.Post(w, req)
	responseContains(t, w, 200, "Processing...")
	_, _, comment := e2eVCSClient.VerifyWasCalledOnce().CreateComment(AnyRepo(), AnyInt(), AnyString()).GetCapturedArguments()

	exp, err := ioutil.ReadFile(filepath.Join(testRepoDir, "exp-output-atlantis-plan.txt"))
	fmt.Println((string(exp)))
	Ok(t, err)

	Equals(t, string(exp), comment)
}

type mockLockURLGenerator struct{}

func (m *mockLockURLGenerator) GenerateLockURL(lockID string) string {
	return "lock-url"
}

type mockWebhookSender struct{}

func (w *mockWebhookSender) Send(log *logging.SimpleLogger, result webhooks.ApplyResult) error {
	return nil
}

func GitHubCommentEvent(t *testing.T, comment string) *http.Request {
	requestJSON, err := ioutil.ReadFile(filepath.Join("testfixtures", "githubIssueCommentEvent.json"))
	Ok(t, err)
	requestJSON = []byte(strings.Replace(string(requestJSON), "###comment body###", comment, 1))
	req, err := http.NewRequest("POST", "/events", bytes.NewBuffer(requestJSON))
	Ok(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(githubHeader, "issue_comment")
	return req
}

func GitHubPullRequestParsed() *github.PullRequest {
	return &github.PullRequest{
		Number:  github.Int(1),
		State:   github.String("open"),
		HTMLURL: github.String("htmlurl"),
		Head: &github.PullRequestBranch{
			Repo: &github.Repository{
				FullName: github.String("runatlantis/atlantis-tests"),
				CloneURL: github.String("/runatlantis/atlantis-tests.git"),
			},
			SHA: github.String("sha"),
			Ref: github.String("branch"),
		},
		Base: &github.PullRequestBranch{
			Repo: &github.Repository{
				FullName: github.String("runatlantis/atlantis-tests"),
				CloneURL: github.String("/runatlantis/atlantis-tests.git"),
			},
		},
		User: &github.User{
			Login: github.String("atlantisbot"),
		},
	}
}
