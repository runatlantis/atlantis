package server_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"path/filepath"
	"regexp"
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

/*
flows:
- pull request opened autoplan
- comment to apply

github/gitlab

locking

merging pull requests

different repo organizations

atlantis.yaml

*/
func TestGitHubWorkflow(t *testing.T) {
	RegisterMockTestingT(t)

	cases := []struct {
		Description string
		// RepoDir is relative to testfixtures/test-repos.
		RepoDir                string
		ModifiedFiles          []string
		ExpAutoplanCommentFile string
		ExpMergeCommentFile    string
		CommentAndReplies      []string
	}{
		{
			Description:            "simple",
			RepoDir:                "simple",
			ModifiedFiles:          []string{"main.tf"},
			ExpAutoplanCommentFile: "exp-output-autoplan.txt",
			CommentAndReplies: []string{
				"atlantis apply", "exp-output-apply.txt",
			},
			ExpMergeCommentFile: "exp-output-merge.txt",
		},
	}
	for _, c := range cases {
		ctrl, vcsClient, githubGetter, atlantisWorkspace := setupE2E(t)
		t.Run(c.Description, func(t *testing.T) {
			// Set the repo to be cloned through the testing backdoor.
			repoDir, cleanup := initializeRepo(t, c.RepoDir)
			defer cleanup()
			atlantisWorkspace.TestingOverrideCloneURL = fmt.Sprintf("file://%s", repoDir)

			// Setup test dependencies.
			w := httptest.NewRecorder()
			When(githubGetter.GetPullRequest(AnyRepo(), AnyInt())).ThenReturn(GitHubPullRequestParsed(), nil)
			When(vcsClient.GetModifiedFiles(AnyRepo(), matchers.AnyModelsPullRequest())).ThenReturn(c.ModifiedFiles, nil)

			// First, send the open pull request event and trigger an autoplan.
			pullOpenedReq := GitHubPullRequestOpenedEvent(t)
			ctrl.Post(w, pullOpenedReq)
			responseContains(t, w, 200, "Processing...")
			_, _, autoplanComment := vcsClient.VerifyWasCalledOnce().CreateComment(AnyRepo(), AnyInt(), AnyString()).GetCapturedArguments()
			exp, err := ioutil.ReadFile(filepath.Join(repoDir, c.ExpAutoplanCommentFile))
			Ok(t, err)
			Equals(t, string(exp), autoplanComment)

			// Now send any other comments.
			for i := 0; i < len(c.CommentAndReplies); i += 2 {
				comment := c.CommentAndReplies[i]
				expOutputFile := c.CommentAndReplies[i+1]

				commentReq := GitHubCommentEvent(t, comment)
				w = httptest.NewRecorder()
				ctrl.Post(w, commentReq)
				responseContains(t, w, 200, "Processing...")
				_, _, atlantisComment := vcsClient.VerifyWasCalled(Twice()).CreateComment(AnyRepo(), AnyInt(), AnyString()).GetCapturedArguments()

				exp, err = ioutil.ReadFile(filepath.Join(repoDir, expOutputFile))
				Ok(t, err)
				// Replace all 'ID: 1111818181' strings with * so we can do a comparison.
				idRegex := regexp.MustCompile(`\(ID: [0-9]+\)`)
				atlantisComment = idRegex.ReplaceAllString(atlantisComment, "(ID: ******************)")
				Equals(t, string(exp), atlantisComment)
			}

			// Finally, send the pull request merged event.
			pullClosedReq := GitHubPullRequestClosedEvent(t)
			w = httptest.NewRecorder()
			ctrl.Post(w, pullClosedReq)
			responseContains(t, w, 200, "Pull request cleaned successfully")
			_, _, pullClosedComment := vcsClient.VerifyWasCalled(Times(3)).CreateComment(AnyRepo(), AnyInt(), AnyString()).GetCapturedArguments()
			exp, err = ioutil.ReadFile(filepath.Join(repoDir, c.ExpMergeCommentFile))
			Ok(t, err)
			Equals(t, string(exp), pullClosedComment)
		})
	}
}

func setupE2E(t *testing.T) (server.EventsController, *vcsmocks.MockClientProxy, *mocks.MockGithubPullGetter, *events.FileWorkspace) {
	allowForkPRs := false
	dataDir, cleanup := TempDir(t)
	defer cleanup()
	testRepoDir, err := filepath.Abs("testfixtures/test-repos/simple")
	Ok(t, err)

	// Mocks.
	e2eVCSClient := vcsmocks.NewMockClientProxy()
	e2eStatusUpdater := mocks.NewMockCommitStatusUpdater()
	e2eGithubGetter := mocks.NewMockGithubPullGetter()
	e2eGitlabGetter := mocks.NewMockGitlabMergeRequestGetter()

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
	atlantisWorkspace := &events.FileWorkspace{
		DataDir:                 dataDir,
		TestingOverrideCloneURL: testRepoDir,
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
			Workspace:         atlantisWorkspace,
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
				Workspace: atlantisWorkspace,
				Webhooks:  &mockWebhookSender{},
			},
		},
	}

	ctrl := server.EventsController{
		TestingMode:   true,
		CommandRunner: commandHandler,
		PullCleaner: &events.PullClosedExecutor{
			Locker:    lockingClient,
			VCSClient: e2eVCSClient,
			Workspace: atlantisWorkspace,
		},
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
	return ctrl, e2eVCSClient, e2eGithubGetter, atlantisWorkspace
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

func GitHubPullRequestOpenedEvent(t *testing.T) *http.Request {
	requestJSON, err := ioutil.ReadFile(filepath.Join("testfixtures", "githubPullRequestOpenedEvent.json"))
	Ok(t, err)
	req, err := http.NewRequest("POST", "/events", bytes.NewBuffer(requestJSON))
	Ok(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(githubHeader, "pull_request")
	return req
}

func GitHubPullRequestClosedEvent(t *testing.T) *http.Request {
	requestJSON, err := ioutil.ReadFile(filepath.Join("testfixtures", "githubPullRequestClosedEvent.json"))
	Ok(t, err)
	req, err := http.NewRequest("POST", "/events", bytes.NewBuffer(requestJSON))
	Ok(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(githubHeader, "pull_request")
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

// initializeRepo copies the repo data from testfixtures and initializes a new
// git repo in a temp directory. It returns that directory and a function
// to run in a defer that will delete the dir.
// The purpose of this function is to create a real git repository with a branch
// called 'branch' from the files under repoDir. This is so we can check in
// those files normally without needing a .git directory.
func initializeRepo(t *testing.T, repoDir string) (string, func()) {
	originRepo, err := filepath.Abs(filepath.Join("testfixtures", "test-repos", repoDir))
	Ok(t, err)

	// Copy the files to the temp dir.
	destDir, cleanup := TempDir(t)
	runCmd(t, "", "cp", "-r", fmt.Sprintf("%s/.", originRepo), destDir)

	// Initialize the git repo.
	runCmd(t, destDir, "git", "init")
	runCmd(t, destDir, "git", "add", ".gitkeep")
	runCmd(t, destDir, "git", "commit", "-m", "initial commit")
	runCmd(t, destDir, "git", "checkout", "-b", "branch")
	runCmd(t, destDir, "git", "add", ".")
	runCmd(t, destDir, "git", "commit", "-am", "branch commit")

	return destDir, cleanup
}

func runCmd(t *testing.T, dir string, name string, args ...string) {
	cpCmd := exec.Command(name, args...)
	cpCmd.Dir = dir
	cpOut, err := cpCmd.CombinedOutput()
	Assert(t, err == nil, "err running %q: %s", strings.Join(append([]string{name}, args...), " "), cpOut)
}
