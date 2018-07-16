package server_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
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

func TestGitHubWorkflow(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
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
		{
			Description:            "simple with comment -var",
			RepoDir:                "simple",
			ModifiedFiles:          []string{"main.tf"},
			ExpAutoplanCommentFile: "exp-output-autoplan.txt",
			CommentAndReplies: []string{
				"atlantis plan -- -var var=overridden", "exp-output-atlantis-plan.txt",
				"atlantis apply", "exp-output-apply-var.txt",
			},
			ExpMergeCommentFile: "exp-output-merge.txt",
		},
		{
			Description:            "simple with workspaces",
			RepoDir:                "simple",
			ModifiedFiles:          []string{"main.tf"},
			ExpAutoplanCommentFile: "exp-output-autoplan.txt",
			CommentAndReplies: []string{
				"atlantis plan -- -var var=default_workspace", "exp-output-atlantis-plan.txt",
				"atlantis plan -w new_workspace -- -var var=new_workspace", "exp-output-atlantis-plan-new-workspace.txt",
				"atlantis apply", "exp-output-apply-var-default-workspace.txt",
				"atlantis apply -w new_workspace", "exp-output-apply-var-new-workspace.txt",
			},
			ExpMergeCommentFile: "exp-output-merge-workspaces.txt",
		},
		{
			Description:            "simple with atlantis.yaml",
			RepoDir:                "simple-yaml",
			ModifiedFiles:          []string{"main.tf"},
			ExpAutoplanCommentFile: "exp-output-autoplan.txt",
			CommentAndReplies: []string{
				"atlantis apply -w staging", "exp-output-apply-staging.txt",
				"atlantis apply", "exp-output-apply-default.txt",
			},
			ExpMergeCommentFile: "exp-output-merge.txt",
		},
		{
			Description:            "modules staging only",
			RepoDir:                "modules",
			ModifiedFiles:          []string{"staging/main.tf"},
			ExpAutoplanCommentFile: "exp-output-autoplan-only-staging.txt",
			CommentAndReplies: []string{
				"atlantis apply -d staging", "exp-output-apply-staging.txt",
			},
			ExpMergeCommentFile: "exp-output-merge-only-staging.txt",
		},
		{
			Description:            "modules modules only",
			RepoDir:                "modules",
			ModifiedFiles:          []string{"modules/null/main.tf"},
			ExpAutoplanCommentFile: "",
			CommentAndReplies: []string{
				"atlantis plan -d staging", "exp-output-plan-staging.txt",
				"atlantis plan -d production", "exp-output-plan-production.txt",
				"atlantis apply -d staging", "exp-output-apply-staging.txt",
				"atlantis apply -d production", "exp-output-apply-production.txt",
			},
			ExpMergeCommentFile: "exp-output-merge-all-dirs.txt",
		},
		{
			Description:            "modules-yaml",
			RepoDir:                "modules-yaml",
			ModifiedFiles:          []string{"modules/null/main.tf"},
			ExpAutoplanCommentFile: "exp-output-autoplan.txt",
			CommentAndReplies: []string{
				"atlantis apply -d staging", "exp-output-apply-staging.txt",
				"atlantis apply -d production", "exp-output-apply-production.txt",
			},
			ExpMergeCommentFile: "exp-output-merge.txt",
		},
		{
			Description:            "tfvars-yaml",
			RepoDir:                "tfvars-yaml",
			ModifiedFiles:          []string{"main.tf"},
			ExpAutoplanCommentFile: "exp-output-autoplan.txt",
			CommentAndReplies: []string{
				"atlantis apply -p staging", "exp-output-apply-staging.txt",
				"atlantis apply -p default", "exp-output-apply-default.txt",
			},
			ExpMergeCommentFile: "exp-output-merge.txt",
		},
		{
			Description:            "tfvars no autoplan",
			RepoDir:                "tfvars-yaml-no-autoplan",
			ModifiedFiles:          []string{"main.tf"},
			ExpAutoplanCommentFile: "",
			CommentAndReplies: []string{
				"atlantis plan -p staging", "exp-output-plan-staging.txt",
				"atlantis plan -p default", "exp-output-plan-default.txt",
				"atlantis apply -p staging", "exp-output-apply-staging.txt",
				"atlantis apply -p default", "exp-output-apply-default.txt",
			},
			ExpMergeCommentFile: "exp-output-merge.txt",
		},
	}
	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			RegisterMockTestingT(t)

			ctrl, vcsClient, githubGetter, atlantisWorkspace := setupE2E(t)
			// Set the repo to be cloned through the testing backdoor.
			repoDir, headSHA, cleanup := initializeRepo(t, c.RepoDir)
			defer cleanup()
			atlantisWorkspace.TestingOverrideCloneURL = fmt.Sprintf("file://%s", repoDir)

			// Setup test dependencies.
			w := httptest.NewRecorder()
			When(githubGetter.GetPullRequest(AnyRepo(), AnyInt())).ThenReturn(GitHubPullRequestParsed(headSHA), nil)
			When(vcsClient.GetModifiedFiles(AnyRepo(), matchers.AnyModelsPullRequest())).ThenReturn(c.ModifiedFiles, nil)
			expNumTimesCalledCreateComment := 0

			// First, send the open pull request event and trigger an autoplan.
			pullOpenedReq := GitHubPullRequestOpenedEvent(t, headSHA)
			ctrl.Post(w, pullOpenedReq)
			responseContains(t, w, 200, "Processing...")
			if c.ExpAutoplanCommentFile != "" {
				expNumTimesCalledCreateComment++
				_, _, autoplanComment := vcsClient.VerifyWasCalledOnce().CreateComment(AnyRepo(), AnyInt(), AnyString()).GetCapturedArguments()
				assertCommentEquals(t, c.ExpAutoplanCommentFile, autoplanComment, c.RepoDir)
			}

			// Now send any other comments.
			for i := 0; i < len(c.CommentAndReplies); i += 2 {
				comment := c.CommentAndReplies[i]
				expOutputFile := c.CommentAndReplies[i+1]

				commentReq := GitHubCommentEvent(t, comment)
				w = httptest.NewRecorder()
				ctrl.Post(w, commentReq)
				responseContains(t, w, 200, "Processing...")
				// Each comment warrants a response. The comments are at the
				// even indices.
				if i%2 == 0 {
					expNumTimesCalledCreateComment++
				}
				_, _, atlantisComment := vcsClient.VerifyWasCalled(Times(expNumTimesCalledCreateComment)).CreateComment(AnyRepo(), AnyInt(), AnyString()).GetCapturedArguments()
				assertCommentEquals(t, expOutputFile, atlantisComment, c.RepoDir)
			}

			// Finally, send the pull request merged event.
			pullClosedReq := GitHubPullRequestClosedEvent(t)
			w = httptest.NewRecorder()
			ctrl.Post(w, pullClosedReq)
			responseContains(t, w, 200, "Pull request cleaned successfully")
			expNumTimesCalledCreateComment++
			_, _, pullClosedComment := vcsClient.VerifyWasCalled(Times(expNumTimesCalledCreateComment)).CreateComment(AnyRepo(), AnyInt(), AnyString()).GetCapturedArguments()
			assertCommentEquals(t, c.ExpMergeCommentFile, pullClosedComment, c.RepoDir)
		})
	}
}

func setupE2E(t *testing.T) (server.EventsController, *vcsmocks.MockClientProxy, *mocks.MockGithubPullGetter, *events.FileWorkspace) {
	allowForkPRs := false
	dataDir, cleanup := TempDir(t)
	defer cleanup()

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
	workingDir := &events.FileWorkspace{
		DataDir:                 dataDir,
		TestingOverrideCloneURL: "override-me",
	}

	defaultTFVersion := terraformClient.Version()
	locker := events.NewDefaultWorkingDirLocker()
	commandRunner := &events.DefaultCommandRunner{
		ProjectCommandRunner: &events.DefaultProjectCommandRunner{
			Locker:           projectLocker,
			LockURLGenerator: &mockLockURLGenerator{},
			InitStepRunner: &runtime.InitStepRunner{
				TerraformExecutor: terraformClient,
				DefaultTFVersion:  defaultTFVersion,
			},
			PlanStepRunner: &runtime.PlanStepRunner{
				TerraformExecutor: terraformClient,
				DefaultTFVersion:  defaultTFVersion,
			},
			ApplyStepRunner: &runtime.ApplyStepRunner{
				TerraformExecutor: terraformClient,
			},
			RunStepRunner: &runtime.RunStepRunner{
				DefaultTFVersion: defaultTFVersion,
			},
			PullApprovedChecker: e2eVCSClient,
			WorkingDir:          workingDir,
			Webhooks:            &mockWebhookSender{},
			WorkingDirLocker:    locker,
		},
		EventParser:              eventParser,
		VCSClient:                e2eVCSClient,
		GithubPullGetter:         e2eGithubGetter,
		GitlabMergeRequestGetter: e2eGitlabGetter,
		CommitStatusUpdater:      e2eStatusUpdater,
		MarkdownRenderer:         &events.MarkdownRenderer{},
		Logger:                   logger,
		AllowForkPRs:             allowForkPRs,
		AllowForkPRsFlag:         "allow-fork-prs",
		ProjectCommandBuilder: &events.DefaultProjectCommandBuilder{
			ParserValidator:     &yaml.ParserValidator{},
			ProjectFinder:       &events.DefaultProjectFinder{},
			VCSClient:           e2eVCSClient,
			WorkingDir:          workingDir,
			WorkingDirLocker:    locker,
			AllowRepoConfigFlag: "allow-repo-config",
			AllowRepoConfig:     true,
		},
	}

	ctrl := server.EventsController{
		TestingMode:   true,
		CommandRunner: commandRunner,
		PullCleaner: &events.PullClosedExecutor{
			Locker:     lockingClient,
			VCSClient:  e2eVCSClient,
			WorkingDir: workingDir,
		},
		Logger:                       logger,
		Parser:                       eventParser,
		CommentParser:                commentParser,
		GithubWebHookSecret:          nil,
		GithubRequestValidator:       &server.DefaultGithubRequestValidator{},
		GitlabRequestParserValidator: &server.DefaultGitlabRequestParserValidator{},
		GitlabWebHookSecret:          nil,
		RepoWhitelistChecker: &events.RepoWhitelistChecker{
			Whitelist: "*",
		},
		SupportedVCSHosts: []models.VCSHostType{models.Gitlab, models.Github},
		VCSClient:         e2eVCSClient,
	}
	return ctrl, e2eVCSClient, e2eGithubGetter, workingDir
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

func GitHubPullRequestOpenedEvent(t *testing.T, headSHA string) *http.Request {
	requestJSON, err := ioutil.ReadFile(filepath.Join("testfixtures", "githubPullRequestOpenedEvent.json"))
	Ok(t, err)
	// Replace sha with expected sha.
	requestJSONStr := strings.Replace(string(requestJSON), "c31fd9ea6f557ad2ea659944c3844a059b83bc5d", headSHA, -1)
	req, err := http.NewRequest("POST", "/events", bytes.NewBuffer([]byte(requestJSONStr)))
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

func GitHubPullRequestParsed(headSHA string) *github.PullRequest {
	// headSHA can't be empty so default if not set.
	if headSHA == "" {
		headSHA = "13940d121be73f656e2132c6d7b4c8e87878ac8d"
	}
	return &github.PullRequest{
		Number:  github.Int(2),
		State:   github.String("open"),
		HTMLURL: github.String("htmlurl"),
		Head: &github.PullRequestBranch{
			Repo: &github.Repository{
				FullName: github.String("runatlantis/atlantis-tests"),
				CloneURL: github.String("/runatlantis/atlantis-tests.git"),
			},
			SHA: github.String(headSHA),
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

// absRepoPath returns the absolute path to the test repo under dir repoDir.
func absRepoPath(t *testing.T, repoDir string) string {
	path, err := filepath.Abs(filepath.Join("testfixtures", "test-repos", repoDir))
	Ok(t, err)
	return path
}

// initializeRepo copies the repo data from testfixtures and initializes a new
// git repo in a temp directory. It returns that directory and a function
// to run in a defer that will delete the dir.
// The purpose of this function is to create a real git repository with a branch
// called 'branch' from the files under repoDir. This is so we can check in
// those files normally without needing a .git directory.
func initializeRepo(t *testing.T, repoDir string) (string, string, func()) {
	originRepo := absRepoPath(t, repoDir)

	// Copy the files to the temp dir.
	destDir, cleanup := TempDir(t)
	runCmd(t, "", "cp", "-r", fmt.Sprintf("%s/.", originRepo), destDir)

	// Initialize the git repo.
	runCmd(t, destDir, "git", "init")
	runCmd(t, destDir, "touch", ".gitkeep")
	runCmd(t, destDir, "git", "add", ".gitkeep")
	runCmd(t, destDir, "git", "config", "--local", "user.email", "atlantisbot@runatlantis.io")
	runCmd(t, destDir, "git", "config", "--local", "user.name", "atlantisbot")
	runCmd(t, destDir, "git", "commit", "-m", "initial commit")
	runCmd(t, destDir, "git", "checkout", "-b", "branch")
	runCmd(t, destDir, "git", "add", ".")
	runCmd(t, destDir, "git", "commit", "-am", "branch commit")
	headSHA := runCmd(t, destDir, "git", "rev-parse", "HEAD")
	headSHA = strings.Trim(headSHA, "\n")

	return destDir, headSHA, cleanup
}

func runCmd(t *testing.T, dir string, name string, args ...string) string {
	cpCmd := exec.Command(name, args...)
	cpCmd.Dir = dir
	cpOut, err := cpCmd.CombinedOutput()
	Assert(t, err == nil, "err running %q: %s", strings.Join(append([]string{name}, args...), " "), cpOut)
	return string(cpOut)
}

func assertCommentEquals(t *testing.T, expFile string, act string, repoDir string) {
	t.Helper()
	exp, err := ioutil.ReadFile(filepath.Join(absRepoPath(t, repoDir), expFile))
	Ok(t, err)

	// Replace all 'Creation complete after 1s ID: 1111818181' strings with
	// 'Creation complete after *s ID: **********' so we can do a comparison.
	idRegex := regexp.MustCompile(`Creation complete after [0-9]+s \(ID: [0-9]+\)`)
	act = idRegex.ReplaceAllString(act, "Creation complete after *s (ID: ******************)")

	if string(exp) != act {
		// If in CI, we write the diff to the console. Otherwise we write the diff
		// to file so we can use our local diff viewer.
		if os.Getenv("CI") == "true" {
			t.Logf("exp: %s, got: %s", string(exp), act)
			t.FailNow()
		} else {
			actFile := filepath.Join(absRepoPath(t, repoDir), expFile+".act")
			err := ioutil.WriteFile(actFile, []byte(act), 0600)
			Ok(t, err)
			cwd, err := os.Getwd()
			Ok(t, err)
			rel, err := filepath.Rel(cwd, actFile)
			Ok(t, err)
			t.Errorf("%q was different, wrote actual comment to %q", expFile, rel)
		}
	}
}
