package events_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/google/go-github/v45/github"
	"github.com/hashicorp/go-getter"
	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server"
	events_controllers "github.com/runatlantis/atlantis/server/controllers/events"
	"github.com/runatlantis/atlantis/server/controllers/events/handlers"
	"github.com/runatlantis/atlantis/server/core/config"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/core/runtime"
	"github.com/runatlantis/atlantis/server/jobs"
	lyftCommand "github.com/runatlantis/atlantis/server/lyft/command"
	event_types "github.com/runatlantis/atlantis/server/neptune/gateway/event"
	github_converter "github.com/runatlantis/atlantis/server/vcs/provider/github/converter"
	"github.com/runatlantis/atlantis/server/vcs/provider/github/request"
	ffclient "github.com/thomaspoignant/go-feature-flag"

	"github.com/runatlantis/atlantis/server/core/runtime/policy"
	"github.com/runatlantis/atlantis/server/core/terraform"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/command/policies"
	"github.com/runatlantis/atlantis/server/vcs/markdown"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	lyft_vcs "github.com/runatlantis/atlantis/server/events/vcs/lyft"
	"github.com/runatlantis/atlantis/server/events/vcs/types"
	"github.com/runatlantis/atlantis/server/events/webhooks"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/lyft/feature"
	"github.com/runatlantis/atlantis/server/metrics"
	"github.com/runatlantis/atlantis/server/wrappers"
	. "github.com/runatlantis/atlantis/testing"
)

const ConftestVersion = "0.25.0"
const githubHeader = "X-Github-Event"

type noopPushEventHandler struct{}

func (h noopPushEventHandler) Handle(ctx context.Context, event event_types.Push) error {
	return nil
}

type noopCheckRunEventHandler struct{}

func (h noopCheckRunEventHandler) Handle(ctx context.Context, event event_types.CheckRun) error {
	return nil
}

type NoopTFDownloader struct{}

func (m *NoopTFDownloader) GetFile(dst, src string, opts ...getter.ClientOption) error {
	return nil
}

func (m *NoopTFDownloader) GetAny(dst, src string, opts ...getter.ClientOption) error {
	return nil
}

type LocalConftestCache struct{}

func (m *LocalConftestCache) Get(key *version.Version) (string, error) {
	return exec.LookPath(fmt.Sprintf("conftest%s", ConftestVersion))
}

func TestGitHubWorkflow(t *testing.T) {

	if testing.Short() {
		t.SkipNow()
	}
	// Ensure we have >= TF 0.14 locally.
	ensureRunning014(t)

	cases := []struct {
		Description string
		// RepoDir is relative to testfixtures/test-repos.
		RepoDir string
		// ModifiedFiles are the list of files that have been modified in this
		// pull request.
		ModifiedFiles []string
		// Comments are what our mock user writes to the pull request.
		Comments []string
		// DisableApply flag used by userConfig object when initializing atlantis server.
		DisableApply bool
		// ApplyLock creates an apply lock that temporarily disables apply command
		ApplyLock bool
		// ExpParallel is true if we expect Atlantis to run parallel plans or applies.
		ExpParallel bool
		// ExpMergeable is true if we expect Atlantis to be able to merge.
		// If for instance policy check is failing and there are no approvals
		// ExpMergeable should be false
		ExpMergeable bool
		// ExpReplies is a list of files containing the expected replies that
		// Atlantis writes to the pull request in order. A reply from a parallel operation
		// will be matched using a substring check.
		ExpReplies [][]string
	}{
		{
			Description:   "simple",
			RepoDir:       "simple",
			ModifiedFiles: []string{"main.tf"},
			Comments: []string{
				"atlantis apply",
			},
			ExpReplies: [][]string{
				{"exp-output-autoplan.txt"},
				{"exp-output-apply.txt"},
				{"exp-output-merge.txt"},
			},
		},
		{
			Description:   "simple with plan comment",
			RepoDir:       "simple",
			ModifiedFiles: []string{"main.tf"},
			Comments: []string{
				"atlantis plan",
				"atlantis apply",
			},
			ExpReplies: [][]string{
				{"exp-output-autoplan.txt"},
				{"exp-output-autoplan.txt"},
				{"exp-output-apply.txt"},
				{"exp-output-merge.txt"},
			},
		},
		{
			Description:   "simple with comment -var",
			RepoDir:       "simple",
			ModifiedFiles: []string{"main.tf"},
			Comments: []string{
				"atlantis plan -- -var var=overridden",
				"atlantis apply",
			},
			ExpReplies: [][]string{
				{"exp-output-autoplan.txt"},
				{"exp-output-atlantis-plan-var-overridden.txt"},
				{"exp-output-apply-var.txt"},
				{"exp-output-merge.txt"},
			},
		},
		{
			Description:   "simple with workspaces",
			RepoDir:       "simple",
			ModifiedFiles: []string{"main.tf"},
			Comments: []string{
				"atlantis plan -- -var var=default_workspace",
				"atlantis plan -w new_workspace -- -var var=new_workspace",
				"atlantis apply -w default",
				"atlantis apply -w new_workspace",
			},
			ExpReplies: [][]string{
				{"exp-output-autoplan.txt"},
				{"exp-output-atlantis-plan.txt"},
				{"exp-output-atlantis-plan-new-workspace.txt"},
				{"exp-output-apply-var-default-workspace.txt"},
				{"exp-output-apply-var-new-workspace.txt"},
				{"exp-output-merge-workspaces.txt"},
			},
		},
		{
			Description:   "simple with workspaces and apply all",
			RepoDir:       "simple",
			ModifiedFiles: []string{"main.tf"},
			Comments: []string{
				"atlantis plan -- -var var=default_workspace",
				"atlantis plan -w new_workspace -- -var var=new_workspace",
				"atlantis apply",
			},
			ExpReplies: [][]string{
				{"exp-output-autoplan.txt"},
				{"exp-output-atlantis-plan.txt"},
				{"exp-output-atlantis-plan-new-workspace.txt"},
				{"exp-output-apply-var-all.txt"},
				{"exp-output-merge-workspaces.txt"},
			},
		},
		{
			Description:   "simple with atlantis.yaml",
			RepoDir:       "simple-yaml",
			ModifiedFiles: []string{"main.tf"},
			Comments: []string{
				"atlantis apply -w staging",
				"atlantis apply -w default",
			},
			ExpReplies: [][]string{
				{"exp-output-autoplan.txt"},
				{"exp-output-apply-staging.txt"},
				{"exp-output-apply-default.txt"},
				{"exp-output-merge.txt"},
			},
		},
		{
			Description:   "simple with atlantis.yaml and apply all",
			RepoDir:       "simple-yaml",
			ModifiedFiles: []string{"main.tf"},
			Comments: []string{
				"atlantis apply",
			},
			ExpReplies: [][]string{
				{"exp-output-autoplan.txt"},
				{"exp-output-apply-all.txt"},
				{"exp-output-merge.txt"},
			},
		},
		{
			Description:   "modules staging only",
			RepoDir:       "modules",
			ModifiedFiles: []string{"staging/main.tf"},
			Comments: []string{
				"atlantis apply -d staging",
			},
			ExpReplies: [][]string{
				{"exp-output-autoplan-only-staging.txt"},
				{"exp-output-apply-staging.txt"},
				{"exp-output-merge-only-staging.txt"},
			},
		},
		{
			Description:   "modules modules only",
			RepoDir:       "modules",
			ModifiedFiles: []string{"modules/null/main.tf", "staging/main.tf"},
			Comments: []string{
				"atlantis plan -d production",
				"atlantis apply -d staging",
				"atlantis apply -d production",
			},
			ExpReplies: [][]string{
				{"exp-output-autoplan-only-staging.txt"},
				{"exp-output-plan-production.txt"},
				{"exp-output-apply-staging.txt"},
				{"exp-output-apply-production.txt"},
				{"exp-output-merge-all-dirs.txt"},
			},
		},
		{
			Description:   "modules-yaml",
			RepoDir:       "modules-yaml",
			ModifiedFiles: []string{"modules/null/main.tf"},
			Comments: []string{
				"atlantis apply -d staging",
				"atlantis apply -d production",
			},
			ExpReplies: [][]string{
				{"exp-output-autoplan.txt"},
				{"exp-output-apply-staging.txt"},
				{"exp-output-apply-production.txt"},
				{"exp-output-merge.txt"},
			},
		},
		{
			Description:   "tfvars-yaml",
			RepoDir:       "tfvars-yaml",
			ModifiedFiles: []string{"main.tf"},
			Comments: []string{
				"atlantis apply -p staging",
				"atlantis apply -p default",
			},
			ExpReplies: [][]string{
				{"exp-output-autoplan.txt"},
				{"exp-output-apply-staging.txt"},
				{"exp-output-apply-default.txt"},
				{"exp-output-merge.txt"},
			},
		},
		{
			Description:   "server-side cfg",
			RepoDir:       "server-side-cfg",
			ModifiedFiles: []string{"main.tf"},
			Comments: []string{
				"atlantis apply -w staging",
				"atlantis apply -w default",
			},
			ExpReplies: [][]string{
				{"exp-output-autoplan.txt"},
				{"exp-output-apply-staging-workspace.txt"},
				{"exp-output-apply-default-workspace.txt"},
				{"exp-output-merge.txt"},
			},
		},
		{
			Description:   "workspaces parallel with atlantis.yaml",
			RepoDir:       "workspace-parallel-yaml",
			ModifiedFiles: []string{"production/main.tf", "staging/main.tf"},
			ExpParallel:   true,
			Comments: []string{
				"atlantis apply",
			},
			ExpReplies: [][]string{
				{"exp-output-autoplan-staging.txt", "exp-output-autoplan-production.txt"},
				{"exp-output-apply-all-staging.txt", "exp-output-apply-all-production.txt"},
				{"exp-output-merge.txt"},
			},
		},
		{
			Description:   "global apply lock disables apply commands",
			RepoDir:       "simple-yaml",
			ModifiedFiles: []string{"main.tf"},
			DisableApply:  false,
			ApplyLock:     true,
			Comments: []string{
				"atlantis apply",
			},
			ExpReplies: [][]string{
				{"exp-output-autoplan.txt"},
				{"exp-output-apply-locked.txt"},
				{"exp-output-merge.txt"},
			},
		},
		{
			Description:   "disable apply flag always takes presedence",
			RepoDir:       "simple-yaml",
			ModifiedFiles: []string{"main.tf"},
			DisableApply:  true,
			ApplyLock:     false,
			Comments: []string{
				"atlantis apply",
			},
			ExpReplies: [][]string{
				{"exp-output-autoplan.txt"},
				{"exp-output-apply-locked.txt"},
				{"exp-output-merge.txt"},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			t.Parallel()

			// reset userConfig
			userConfig := &server.UserConfig{}
			userConfig.DisableApply = c.DisableApply

			ghClient := &testGithubClient{ExpectedModifiedFiles: c.ModifiedFiles}

			headSHA, ctrl, applyLocker := setupE2E(t, c.RepoDir, userConfig, ghClient)

			// Set expected pull from github
			ghClient.ExpectedPull = GitHubPullRequestParsed(headSHA)

			// Create global apply lock if required
			if c.ApplyLock {
				_, _ = applyLocker.LockApply()
			}

			// Setup test dependencies.
			w := httptest.NewRecorder()

			// First, send the open pull request event which triggers autoplan.
			pullOpenedReq := GitHubPullRequestOpenedEvent(t, headSHA)
			ctrl.Post(w, pullOpenedReq)
			ResponseContains(t, w, 200, "Processing...")

			// Now send any other comments.
			for _, comment := range c.Comments {
				commentReq := GitHubCommentEvent(t, comment)
				w = httptest.NewRecorder()
				ctrl.Post(w, commentReq)
				ResponseContains(t, w, 200, "Processing...")
			}

			// Send the "pull closed" event which would be triggered by the
			// manual merge.
			pullClosedReq := GitHubPullRequestClosedEvent(t)
			w = httptest.NewRecorder()
			ctrl.Post(w, pullClosedReq)
			ResponseContains(t, w, 200, "Processing...")

			// Verify
			actReplies := ghClient.CapturedComments
			Assert(t, len(c.ExpReplies) == len(actReplies), "missing expected replies, got %d but expected %d", len(actReplies), len(c.ExpReplies))
			for i, expReply := range c.ExpReplies {
				assertCommentEquals(t, expReply, actReplies[i], c.RepoDir, c.ExpParallel)
			}

		})
	}
}

func TestGitHubWorkflowWithPolicyCheck(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	// Ensure we have >= TF 0.14 locally.
	ensureRunning014(t)
	// Ensure we have >= Conftest 0.21 locally.
	ensureRunningConftest(t)

	cases := []struct {
		Description string
		// RepoDir is relative to testfixtures/test-repos.
		RepoDir string
		// ModifiedFiles are the list of files that have been modified in this
		// pull request.
		ModifiedFiles []string
		// Comments are what our mock user writes to the pull request.
		Comments []string
		// ExpReplies is a list of files containing the expected replies that
		// Atlantis writes to the pull request in order. A reply from a parallel operation
		// will be matched using a substring check.
		ExpReplies [][]string
	}{
		{
			Description:   "1 failing policy and 1 passing policy ",
			RepoDir:       "policy-checks-multi-projects",
			ModifiedFiles: []string{"dir1/main.tf,", "dir2/main.tf"},
			Comments: []string{
				"atlantis apply",
			},
			ExpReplies: [][]string{
				{"exp-output-autoplan.txt"},
				{"exp-output-auto-policy-check.txt"},
				{"exp-output-apply.txt"},
				{"exp-output-merge.txt"},
			},
		},
		{
			Description:   "failing policy without policies passing using extra args",
			RepoDir:       "policy-checks-extra-args",
			ModifiedFiles: []string{"main.tf"},
			Comments: []string{
				"atlantis apply",
			},
			ExpReplies: [][]string{
				{"exp-output-autoplan.txt"},
				{"exp-output-auto-policy-check.txt"},
				{"exp-output-apply-failed.txt"},
				{"exp-output-merge.txt"},
			},
		},
		{
			Description:   "failing policy without policies passing",
			RepoDir:       "policy-checks",
			ModifiedFiles: []string{"main.tf"},
			Comments: []string{
				"atlantis apply",
			},
			ExpReplies: [][]string{
				{"exp-output-autoplan.txt"},
				{"exp-output-auto-policy-check.txt"},
				{"exp-output-apply-failed.txt"},
				{"exp-output-merge.txt"},
			},
		},
		{
			Description:   "failing policy additional apply requirements specified",
			RepoDir:       "policy-checks-apply-reqs",
			ModifiedFiles: []string{"main.tf"},
			Comments: []string{
				"atlantis apply",
			},
			ExpReplies: [][]string{
				{"exp-output-autoplan.txt"},
				{"exp-output-auto-policy-check.txt"},
				{"exp-output-apply-failed.txt"},
				{"exp-output-merge.txt"},
			},
		},
		{
			Description:   "failing policy approved by non owner",
			RepoDir:       "policy-checks-diff-owner",
			ModifiedFiles: []string{"main.tf"},
			Comments: []string{
				"atlantis approve_policies",
				"atlantis apply",
			},
			ExpReplies: [][]string{
				{"exp-output-autoplan.txt"},
				{"exp-output-auto-policy-check.txt"},
				{"exp-output-approve-policies.txt"},
				{"exp-output-apply-failed.txt"},
				{"exp-output-merge.txt"},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			t.Parallel()

			// reset userConfig
			userConfig := &server.UserConfig{}
			userConfig.EnablePolicyChecks = true

			ghClient := &testGithubClient{ExpectedModifiedFiles: c.ModifiedFiles}

			headSHA, ctrl, _ := setupE2E(t, c.RepoDir, userConfig, ghClient)

			// Setup test dependencies.
			w := httptest.NewRecorder()

			ghClient.ExpectedPull = GitHubPullRequestParsed(headSHA)
			ghClient.ExpectedApprovalStatus = models.ApprovalStatus{IsApproved: true}

			// First, send the open pull request event which triggers autoplan.
			pullOpenedReq := GitHubPullRequestOpenedEvent(t, headSHA)
			ctrl.Post(w, pullOpenedReq)
			ResponseContains(t, w, 200, "Processing...")

			// Now send any other comments.
			for _, comment := range c.Comments {
				commentReq := GitHubCommentEvent(t, comment)
				w = httptest.NewRecorder()
				ctrl.Post(w, commentReq)
				ResponseContains(t, w, 200, "Processing...")
			}

			// Send the "pull closed" event which would be triggered by the
			// a manual merge.
			pullClosedReq := GitHubPullRequestClosedEvent(t)
			w = httptest.NewRecorder()
			ctrl.Post(w, pullClosedReq)
			ResponseContains(t, w, 200, "Processing...")

			// Verify
			actReplies := ghClient.CapturedComments
			Assert(t, len(c.ExpReplies) == len(actReplies), "missing expected replies, got %d but expected %d", len(actReplies), len(c.ExpReplies))
			for i, expReply := range c.ExpReplies {
				assertCommentEquals(t, expReply, actReplies[i], c.RepoDir, false)
			}
		})
	}
}

func TestGitHubWorkflowPullRequestsWorkflows(t *testing.T) {
	featureConfig := feature.StringRetriever(`platform-mode:
  percentage: 100
  true: true
  false: false
  default: false
  trackEvents: false`)

	if testing.Short() {
		t.SkipNow()
	}
	// Ensure we have >= TF 0.14 locally.
	ensureRunning014(t)
	// Ensure we have >= Conftest 0.21 locally.
	ensureRunningConftest(t)

	cases := []struct {
		Description string
		// RepoDir is relative to testfixtures/test-repos.
		RepoDir string
		// ModifiedFiles are the list of files that have been modified in this
		// pull request.
		ModifiedFiles []string
		// Comments are what our mock user writes to the pull request.
		Comments []string
		// ExpReplies is a list of files containing the expected replies that
		// Atlantis writes to the pull request in order. A reply from a parallel operation
		// will be matched using a substring check.
		ExpReplies [][]string
	}{
		{
			Description:   "disabled apply",
			RepoDir:       "platform-mode/disabled-apply",
			ModifiedFiles: []string{"staging/main.tf"},
			Comments: []string{
				"atlantis apply",
			},
			ExpReplies: [][]string{
				{"exp-output-autoplan.txt"},
				{"exp-output-apply.txt"},
			},
		},
		{
			Description:   "autoplan and policy check approvals",
			RepoDir:       "platform-mode/policy-check-approval",
			ModifiedFiles: []string{"main.tf"},
			Comments: []string{
				"atlantis approve_policies",
				"atlantis apply",
			},
			ExpReplies: [][]string{
				{"exp-output-autoplan.txt"},
				{"exp-output-auto-policy-check.txt"},
				{"exp-output-approve-policies.txt"},
				{"exp-output-apply.txt"},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			t.Parallel()
			// Setup test dependencies.
			w := httptest.NewRecorder()
			// reset userConfig
			userConfig := &server.UserConfig{}
			userConfig.EnablePolicyChecks = true

			ghClient := &testGithubClient{ExpectedModifiedFiles: c.ModifiedFiles}

			headSHA, ctrl, _ := setupE2E(t, c.RepoDir, userConfig, ghClient, e2eOptions{
				featureConfig: featureConfig,
			})

			ghClient.ExpectedPull = GitHubPullRequestParsed(headSHA)
			ghClient.ExpectedApprovalStatus = models.ApprovalStatus{IsApproved: true}

			// First, send the open pull request event which triggers autoplan.
			pullOpenedReq := GitHubPullRequestOpenedEvent(t, headSHA)
			ctrl.Post(w, pullOpenedReq)
			ResponseContains(t, w, 200, "Processing...")

			// Now send any other comments.
			for _, comment := range c.Comments {
				commentReq := GitHubCommentEvent(t, comment)
				w = httptest.NewRecorder()
				ctrl.Post(w, commentReq)
				ResponseContains(t, w, 200, "Processing...")
			}

			// Send the "pull closed" event which would be triggered by the
			// a manual merge.
			pullClosedReq := GitHubPullRequestClosedEvent(t)
			w = httptest.NewRecorder()
			ctrl.Post(w, pullClosedReq)
			ResponseContains(t, w, 200, "Processing...")

			// Verify
			actReplies := ghClient.CapturedComments
			Assert(t, len(c.ExpReplies) == len(actReplies), "missing expected replies, got %d but expected %d", len(actReplies), len(c.ExpReplies))
			for i, expReply := range c.ExpReplies {
				assertCommentEquals(t, expReply, actReplies[i], c.RepoDir, false)
			}
		})
	}
}

type e2eOptions struct {
	featureConfig ffclient.Retriever
}

func setupE2E(t *testing.T, repoFixtureDir string, userConfig *server.UserConfig, ghClient vcs.IGithubClient, options ...e2eOptions) (string, events_controllers.VCSEventsController, locking.ApplyLocker) {

	var featureConfig ffclient.Retriever
	for _, o := range options {
		if o.featureConfig != nil {
			featureConfig = o.featureConfig
		}
	}

	// env vars
	// need this to be set or we'll fail the policy check step
	os.Setenv(policy.DefaultConftestVersionEnvKey, "0.25.0")

	// First initialize the local repo we'll be working with in a unique test specific dir
	repoDir, headSHA := initializeRepo(t, repoFixtureDir)
	// Create subdirs for plugin cache, binaries, data
	// unclear how these are used in conjunction with the above?
	// TODO: investigate unifying this code with the above
	dataDir, binDir, cacheDir := mkSubDirs(t)

	// Set up test dependencies, this is where the code path would diverge from the the standard server
	// initialization for testing purposes
	// ! We should try to keep this as minimal as possible

	lockURLGenerator := &testLockURLGenerator{}
	webhookSender := &testWebhookSender{}

	conftestCache := &LocalConftestCache{}
	downloader := &NoopTFDownloader{}
	overrideCloneURL := fmt.Sprintf("file://%s", repoDir)

	// TODO: we should compare this output against what we post on github
	projectCmdOutputHandler := &jobs.NoopProjectOutputHandler{}

	ctxLogger := logging.NewNoopCtxLogger(t)

	var featureAllocator *feature.PercentageBasedAllocator
	var featureAllocatorErr error
	if featureConfig != nil {
		featureAllocator, featureAllocatorErr = feature.NewStringSourcedAllocatorWithRetriever(ctxLogger, featureConfig)
	} else {
		featureAllocator, featureAllocatorErr = feature.NewStringSourcedAllocator(ctxLogger)
	}

	Ok(t, featureAllocatorErr)

	t.Cleanup(featureAllocator.Close)

	terraformClient, err := terraform.NewE2ETestClient(binDir, cacheDir, "", "", "", "default-tf-version", "https://releases.hashicorp.com", downloader, false, projectCmdOutputHandler)
	Ok(t, err)

	// Set real dependencies here.
	// TODO: aggregate some of this with that of server.go to minimize duplication
	vcsClient := vcs.NewClientProxy(ghClient, nil, nil, nil, nil)
	e2eStatusUpdater := &command.VCSStatusUpdater{Client: vcsClient, TitleBuilder: vcs.StatusTitleBuilder{TitlePrefix: "atlantis"}}

	eventParser := &events.EventParser{
		GithubUser:  "github-user",
		GithubToken: "github-token",
		GitlabUser:  "gitlab-user",
		GitlabToken: "gitlab-token",
	}
	commentParser := &events.CommentParser{
		GithubUser: "github-user",
		GitlabUser: "gitlab-user",
	}

	boltdb, err := db.New(dataDir)
	Ok(t, err)

	lockingClient := locking.NewClient(boltdb)
	applyLocker := locking.NewApplyClient(boltdb, userConfig.DisableApply)
	projectLocker := &events.DefaultProjectLocker{
		Locker:    lockingClient,
		VCSClient: vcsClient,
	}

	globalCfg := valid.NewGlobalCfg(dataDir)

	workingDir := &events.FileWorkspace{
		DataDir:                     dataDir,
		TestingOverrideHeadCloneURL: overrideCloneURL,
		GlobalCfg:                   globalCfg,
	}

	defaultTFVersion := terraformClient.DefaultVersion()
	locker := events.NewDefaultWorkingDirLocker()
	parser := &config.ParserValidator{}

	expCfgPath := filepath.Join(absRepoPath(t, repoFixtureDir), "repos.yaml")
	if _, err := os.Stat(expCfgPath); err == nil {
		globalCfg, err = parser.ParseGlobalCfg(expCfgPath, globalCfg)
		Ok(t, err)
	} else {
		globalCfg, err = parser.ParseGlobalCfgJSON(`{"repos": [{"id":"/.*/", "allow_custom_workflows": true, "allowed_overrides": ["workflow"], "pre_workflow_hooks":[{"run": "echo 'hello world'"}]}]}`, globalCfg)
		Ok(t, err)
	}
	drainer := &events.Drainer{}

	parallelPoolSize := 1

	preWorkflowHooksCommandRunner := &events.DefaultPreWorkflowHooksCommandRunner{
		VCSClient:             vcsClient,
		GlobalCfg:             globalCfg,
		WorkingDirLocker:      locker,
		WorkingDir:            workingDir,
		PreWorkflowHookRunner: runtime.DefaultPreWorkflowHookRunner{},
	}
	statsScope, _, err := metrics.NewLoggingScope(ctxLogger, "atlantis")
	Ok(t, err)

	projectContextBuilder := wrappers.
		WrapProjectContext(events.NewProjectCommandContextBuilder(commentParser)).
		WithInstrumentation(statsScope)

	projectContextBuilder = wrappers.
		WrapProjectContext(events.NewPlatformModeProjectCommandContextBuilder(commentParser, projectContextBuilder, ctxLogger, featureAllocator)).
		WithInstrumentation(statsScope)

	if userConfig.EnablePolicyChecks {
		projectContextBuilder = projectContextBuilder.EnablePolicyChecks(commentParser)
	}

	projectCommandBuilder := events.NewProjectCommandBuilder(
		projectContextBuilder,
		parser,
		&events.DefaultProjectFinder{},
		vcsClient,
		workingDir,
		locker,
		globalCfg,
		&events.DefaultPendingPlanFinder{},
		false,
		"**/*.tf,**/*.tfvars,**/*.tfvars.json,**/terragrunt.hcl",
		ctxLogger,
		events.InfiniteProjectsPerPR,
	)

	showStepRunner, err := runtime.NewShowStepRunner(terraformClient, defaultTFVersion)
	Ok(t, err)

	conftestVersion, err := version.NewVersion(ConftestVersion)
	Ok(t, err)

	conftextExec := policy.NewConfTestExecutorWorkflow(ctxLogger, binDir, downloader)

	// swapping out version cache to something that always returns local contest
	// binary
	conftextExec.VersionCache = conftestCache

	policyCheckRunner, err := runtime.NewPolicyCheckStepRunner(
		conftestVersion,
		conftextExec,
	)
	Ok(t, err)
	initStepRunner := &runtime.InitStepRunner{
		TerraformExecutor: terraformClient,
		DefaultTFVersion:  defaultTFVersion,
	}
	planStepRunner := &runtime.PlanStepRunner{
		TerraformExecutor: terraformClient,
		DefaultTFVersion:  defaultTFVersion,
		AsyncTFExec:       terraformClient,
	}

	applyStepRunner := &runtime.ApplyStepRunner{
		TerraformExecutor: terraformClient,
		AsyncTFExec:       terraformClient,
	}

	versionStepRunner := &runtime.VersionStepRunner{
		TerraformExecutor: terraformClient,
		DefaultTFVersion:  defaultTFVersion,
	}

	runStepRunner := &runtime.RunStepRunner{
		TerraformExecutor: terraformClient,
		DefaultTFVersion:  defaultTFVersion,
		TerraformBinDir:   binDir,
	}

	envStepRunner := &runtime.EnvStepRunner{
		RunStepRunner: runStepRunner,
	}

	stepsRunner := runtime.NewStepsRunner(
		initStepRunner,
		planStepRunner,
		showStepRunner,
		policyCheckRunner,
		applyStepRunner,
		versionStepRunner,
		runStepRunner,
		envStepRunner,
	)

	Ok(t, err)

	applyRequirementHandler := &events.AggregateApplyRequirements{
		WorkingDir: workingDir,
	}
	unwrappedRunner := events.NewProjectCommandRunner(
		stepsRunner,
		workingDir,
		webhookSender,
		locker,
		applyRequirementHandler,
	)

	legacyPrjCmdRunner := wrappers.WrapProjectRunner(
		unwrappedRunner,
	).WithSync(projectLocker, lockURLGenerator)

	platformModePrjCmdRunner := wrappers.WrapProjectRunner(
		unwrappedRunner,
	)

	prjCmdRunner := &lyftCommand.PlatformModeProjectRunner{
		PlatformModeRunner: platformModePrjCmdRunner,
		PrModeRunner:       legacyPrjCmdRunner,
		Allocator:          featureAllocator,
		Logger:             ctxLogger,
	}

	dbUpdater := &events.DBUpdater{
		DB: boltdb,
	}

	pullUpdater := &events.PullOutputUpdater{
		HidePrevPlanComments: false,
		VCSClient:            vcsClient,
		MarkdownRenderer:     &markdown.Renderer{},
	}

	deleteLockCommand := &events.DefaultDeleteLockCommand{
		Locker:           lockingClient,
		Logger:           ctxLogger,
		WorkingDir:       workingDir,
		WorkingDirLocker: locker,
		DB:               boltdb,
	}

	policyCheckCommandRunner := events.NewPolicyCheckCommandRunner(
		dbUpdater,
		pullUpdater,
		e2eStatusUpdater,
		prjCmdRunner,
		parallelPoolSize,
	)

	planCommandRunner := events.NewPlanCommandRunner(
		vcsClient,
		&events.DefaultPendingPlanFinder{},
		workingDir,
		e2eStatusUpdater,
		projectCommandBuilder,
		prjCmdRunner,
		dbUpdater,
		pullUpdater,
		policyCheckCommandRunner,
		parallelPoolSize,
	)

	approvePoliciesCommandRunner := events.NewApprovePoliciesCommandRunner(
		e2eStatusUpdater,
		projectCommandBuilder,
		prjCmdRunner,
		pullUpdater,
		dbUpdater,
		&policies.CommandOutputGenerator{
			PrjCommandRunner:  prjCmdRunner,
			PrjCommandBuilder: projectCommandBuilder,
		},
	)

	unlockCommandRunner := events.NewUnlockCommandRunner(
		deleteLockCommand,
		vcsClient,
	)

	versionCommandRunner := events.NewVersionCommandRunner(
		pullUpdater,
		projectCommandBuilder,
		prjCmdRunner,
		parallelPoolSize,
	)

	var applyCommandRunner command.Runner
	e2ePullReqStatusFetcher := lyft_vcs.NewSQBasedPullStatusFetcher(ghClient, vcs.NewLyftPullMergeabilityChecker("atlantis"))

	applyCommandRunner = events.NewApplyCommandRunner(
		vcsClient,
		false,
		applyLocker,
		e2eStatusUpdater,
		projectCommandBuilder,
		prjCmdRunner,
		pullUpdater,
		dbUpdater,
		parallelPoolSize,
		e2ePullReqStatusFetcher,
	)

	commentCommandRunnerByCmd := map[command.Name]command.Runner{
		command.Plan:            planCommandRunner,
		command.Apply:           applyCommandRunner,
		command.ApprovePolicies: approvePoliciesCommandRunner,
		command.Unlock:          unlockCommandRunner,
		command.Version:         versionCommandRunner,
	}
	staleCommandChecker := &testStaleCommandChecker{}
	commandRunner := &events.DefaultCommandRunner{
		VCSClient:                     vcsClient,
		GlobalCfg:                     globalCfg,
		StatsScope:                    statsScope,
		CommentCommandRunnerByCmd:     commentCommandRunnerByCmd,
		Drainer:                       drainer,
		PreWorkflowHooksCommandRunner: preWorkflowHooksCommandRunner,
		PullStatusFetcher:             boltdb,
		StaleCommandChecker:           staleCommandChecker,
		Logger:                        ctxLogger,
	}

	repoAllowlistChecker, err := events.NewRepoAllowlistChecker("*")
	Ok(t, err)

	autoplanner := &handlers.Autoplanner{
		CommandRunner: commandRunner,
	}

	pullCleaner := &events.PullClosedExecutor{
		Locker:                   lockingClient,
		VCSClient:                vcsClient,
		WorkingDir:               workingDir,
		DB:                       boltdb,
		PullClosedTemplate:       &events.PullClosedEventTemplate{},
		LogStreamResourceCleaner: projectCmdOutputHandler,
	}

	prHandler := handlers.NewPullRequestEventWithEventTypeHandlers(
		repoAllowlistChecker,

		// Use synchronous handlers for testing purposes
		autoplanner, autoplanner,

		&handlers.PullCleaner{
			PullCleaner: pullCleaner,
			Logger:      ctxLogger,
		},
	)

	commentHandler := handlers.NewCommentEventWithCommandHandler(
		commentParser,
		repoAllowlistChecker,
		vcsClient,

		// Use synchronous handler for testing purposes
		&handlers.CommandHandler{
			CommandRunner: commandRunner,
		},
		ctxLogger,
	)

	repoConverter := github_converter.RepoConverter{
		GithubUser:  userConfig.GithubUser,
		GithubToken: userConfig.GithubToken,
	}

	pullConverter := github_converter.PullConverter{
		RepoConverter: repoConverter,
	}

	requestRouter := &events_controllers.RequestRouter{
		Resolvers: []events_controllers.RequestResolver{
			request.NewHandler(
				ctxLogger,
				statsScope,
				nil,
				commentHandler,
				prHandler,
				noopPushEventHandler{},
				noopCheckRunEventHandler{},
				false,
				repoConverter,
				pullConverter,
				ghClient,
			),
		},
	}

	ctrl := events_controllers.VCSEventsController{
		RequestRouter:                requestRouter,
		Logger:                       ctxLogger,
		Scope:                        statsScope,
		Parser:                       eventParser,
		CommentParser:                commentParser,
		GitlabRequestParserValidator: &events_controllers.DefaultGitlabRequestParserValidator{},
		GitlabWebhookSecret:          nil,
		RepoAllowlistChecker:         repoAllowlistChecker,
		SupportedVCSHosts:            []models.VCSHostType{models.Gitlab, models.Github, models.BitbucketCloud},
		VCSClient:                    vcsClient,
	}
	return headSHA, ctrl, applyLocker
}

var (
	// if not for these we'd be doing disk reads for each test
	readCommentJSON           sync.Once
	readPullRequestOpenedJSON sync.Once
	readPullRequestClosedJSON sync.Once

	commentJSON           string
	pullRequestOpenedJSON string
	pullRequestClosedJSON string
)

func GitHubCommentEvent(t *testing.T, comment string) *http.Request {
	readCommentJSON.Do(
		func() {
			jsonBytes, err := os.ReadFile(filepath.Join("testfixtures", "githubIssueCommentEvent.json"))
			Ok(t, err)

			commentJSON = string(jsonBytes)
		},
	)
	modifiedCommentJSON := []byte(strings.Replace(commentJSON, "###comment body###", comment, 1))
	req, err := http.NewRequest("POST", "/events", bytes.NewBuffer(modifiedCommentJSON))
	Ok(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(githubHeader, "issue_comment")
	return req
}

func GitHubPullRequestOpenedEvent(t *testing.T, headSHA string) *http.Request {
	readPullRequestOpenedJSON.Do(
		func() {
			jsonBytes, err := os.ReadFile(filepath.Join("testfixtures", "githubPullRequestOpenedEvent.json"))
			Ok(t, err)

			pullRequestOpenedJSON = string(jsonBytes)
		},
	)
	// Replace sha with expected sha.
	requestJSONStr := strings.Replace(pullRequestOpenedJSON, "c31fd9ea6f557ad2ea659944c3844a059b83bc5d", headSHA, -1)
	req, err := http.NewRequest("POST", "/events", bytes.NewBuffer([]byte(requestJSONStr)))
	Ok(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(githubHeader, "pull_request")
	return req
}

func GitHubPullRequestClosedEvent(t *testing.T) *http.Request {
	readPullRequestClosedJSON.Do(
		func() {
			jsonBytes, err := os.ReadFile(filepath.Join("testfixtures", "githubPullRequestClosedEvent.json"))
			Ok(t, err)

			pullRequestClosedJSON = string(jsonBytes)
		},
	)

	req, err := http.NewRequest("POST", "/events", bytes.NewBuffer([]byte(pullRequestClosedJSON)))
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
	cleanstate := "clean"
	return &github.PullRequest{
		Number:  github.Int(2),
		State:   github.String("open"),
		HTMLURL: github.String("htmlurl"),
		Head: &github.PullRequestBranch{
			Repo: &github.Repository{
				FullName: github.String("runatlantis/atlantis-tests"),
				CloneURL: github.String("https://github.com/runatlantis/atlantis-tests.git"),
			},
			SHA: github.String(headSHA),
			Ref: github.String("branch"),
		},
		Base: &github.PullRequestBranch{
			Repo: &github.Repository{
				FullName: github.String("runatlantis/atlantis-tests"),
				CloneURL: github.String("https://github.com/runatlantis/atlantis-tests.git"),
			},
			Ref: github.String("master"),
		},
		User: &github.User{
			Login: github.String("atlantisbot"),
		},
		MergeableState: &cleanstate,
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
// those files normally to this repo without needing a .git directory.
func initializeRepo(t *testing.T, repoDir string) (string, string) {
	originRepo := absRepoPath(t, repoDir)

	// Copy the files to the temp dir.
	destDir := t.TempDir()
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

	return destDir, headSHA
}

func runCmd(t *testing.T, dir string, name string, args ...string) string {
	cpCmd := exec.Command(name, args...)
	cpCmd.Dir = dir
	cpOut, err := cpCmd.CombinedOutput()
	Assert(t, err == nil, "err running %q: %s", strings.Join(append([]string{name}, args...), " "), cpOut)
	return string(cpOut)
}

func assertCommentEquals(t *testing.T, expReplies []string, act string, repoDir string, parallel bool) {
	t.Helper()

	// Replace all 'Creation complete after 0s [id=2135833172528078362]' strings with
	// 'Creation complete after *s [id=*******************]' so we can do a comparison.
	idRegex := regexp.MustCompile(`Creation complete after [0-9]+s \[id=[0-9]+]`)
	act = idRegex.ReplaceAllString(act, "Creation complete after *s [id=*******************]")

	// Replace all null_resource.simple{n}: .* with null_resource.simple: because
	// with multiple resources being created the logs are all out of order which
	// makes comparison impossible.
	resourceRegex := regexp.MustCompile(`null_resource\.simple(\[\d])?\d?:.*`)
	act = resourceRegex.ReplaceAllString(act, "null_resource.simple:")

	// For parallel plans and applies, do a substring match since output may be out of order
	var replyMatchesExpected func(string, string) bool
	if parallel {
		replyMatchesExpected = func(act string, expStr string) bool {
			return strings.Contains(act, expStr)
		}
	} else {
		replyMatchesExpected = func(act string, expStr string) bool {
			return expStr == act
		}
	}

	for _, expFile := range expReplies {
		exp, err := os.ReadFile(filepath.Join(absRepoPath(t, repoDir), expFile))
		Ok(t, err)
		expStr := string(exp)
		// My editor adds a newline to all the files, so if the actual comment
		// doesn't end with a newline then strip the last newline from the file's
		// contents.
		if !strings.HasSuffix(act, "\n") {
			expStr = strings.TrimSuffix(expStr, "\n")
		}

		if !replyMatchesExpected(act, expStr) {
			t.Logf("\nexp:\n %s\n got:\n %s\n", expStr, act)
			t.FailNow()
		}
	}
}

// returns parent, bindir, cachedir, cleanup func
func mkSubDirs(t *testing.T) (string, string, string) {
	tmp := t.TempDir()
	binDir := filepath.Join(tmp, "bin")
	err := os.MkdirAll(binDir, 0700)
	Ok(t, err)

	cachedir := filepath.Join(tmp, "plugin-cache")
	err = os.MkdirAll(cachedir, 0700)
	Ok(t, err)

	return tmp, binDir, cachedir
}

// Will fail test if conftest isn't in path and isn't version >= 0.25.0
func ensureRunningConftest(t *testing.T) {

	var localPath string
	var err error
	localPath, err = exec.LookPath(fmt.Sprintf("conftest%s", ConftestVersion))
	if err != nil {
		localPath, err = exec.LookPath("conftest")
		if err != nil {
			t.Logf("error finding conftest binary %s", err)
			t.FailNow()
		}
	}
	versionOutBytes, err := exec.Command(localPath, "--version").Output() // #nosec
	if err != nil {
		t.Logf("error running conftest version: %s", err)
		t.FailNow()
	}
	versionOutput := string(versionOutBytes)
	match := versionConftestRegex.FindStringSubmatch(versionOutput)
	if len(match) <= 1 {
		t.Logf("could not parse contest version from %s", versionOutput)
		t.FailNow()
	}
	localVersion, err := version.NewVersion(match[1])
	Ok(t, err)
	minVersion, err := version.NewVersion(ConftestVersion)
	Ok(t, err)
	if localVersion.LessThan(minVersion) {
		t.Logf("must have contest version >= %s, you have %s", minVersion, localVersion)
		t.FailNow()
	}
}

// Will fail test if terraform isn't in path and isn't version >= 0.14
func ensureRunning014(t *testing.T) {
	localPath, err := exec.LookPath("terraform")
	if err != nil {
		t.Log("terraform >= 0.14 must be installed to run this test")
		t.FailNow()
	}
	versionOutBytes, err := exec.Command(localPath, "version").Output() // #nosec
	if err != nil {
		t.Logf("error running terraform version: %s", err)
		t.FailNow()
	}
	versionOutput := string(versionOutBytes)
	match := versionRegex.FindStringSubmatch(versionOutput)
	if len(match) <= 1 {
		t.Logf("could not parse terraform version from %s", versionOutput)
		t.FailNow()
	}
	localVersion, err := version.NewVersion(match[1])
	Ok(t, err)
	minVersion, err := version.NewVersion("0.14.0")
	Ok(t, err)
	if localVersion.LessThan(minVersion) {
		t.Logf("must have terraform version >= %s, you have %s", minVersion, localVersion)
		t.FailNow()
	}
}

// versionRegex extracts the version from `terraform version` output.
//
//	    Terraform v0.12.0-alpha4 (2c36829d3265661d8edbd5014de8090ea7e2a076)
//		   => 0.12.0-alpha4
//
//	    Terraform v0.11.10
//		   => 0.11.10
var versionRegex = regexp.MustCompile("Terraform v(.*?)(\\s.*)?\n")

var versionConftestRegex = regexp.MustCompile("Version: (.*?)(\\s.*)?\n")

type testLockURLGenerator struct{}

func (m *testLockURLGenerator) GenerateLockURL(lockID string) string {
	return "lock-url"
}

type testWebhookSender struct{}

func (w *testWebhookSender) Send(log logging.Logger, result webhooks.ApplyResult) error {
	return nil
}

type testStaleCommandChecker struct{}

func (t *testStaleCommandChecker) CommandIsStale(ctx *command.Context) bool {
	return false
}

type testGithubClient struct {
	ExpectedModifiedFiles  []string
	ExpectedPull           *github.PullRequest
	ExpectedApprovalStatus models.ApprovalStatus
	CapturedComments       []string
}

func (t *testGithubClient) GetModifiedFiles(repo models.Repo, pull models.PullRequest) ([]string, error) {
	return t.ExpectedModifiedFiles, nil
}
func (t *testGithubClient) CreateComment(repo models.Repo, pullNum int, comment string, command string) error {
	t.CapturedComments = append(t.CapturedComments, comment)
	return nil
}
func (t *testGithubClient) HidePrevCommandComments(repo models.Repo, pullNum int, command string) error {
	return nil
}
func (t *testGithubClient) PullIsApproved(repo models.Repo, pull models.PullRequest) (models.ApprovalStatus, error) {
	return t.ExpectedApprovalStatus, nil
}
func (t *testGithubClient) PullIsMergeable(repo models.Repo, pull models.PullRequest) (bool, error) {
	return false, nil
}
func (t *testGithubClient) UpdateStatus(ctx context.Context, request types.UpdateStatusRequest) (string, error) {
	return "", nil
}
func (t *testGithubClient) MarkdownPullLink(pull models.PullRequest) (string, error) {
	return "", nil
}
func (t *testGithubClient) DownloadRepoConfigFile(pull models.PullRequest) (bool, []byte, error) {
	return false, []byte{}, nil
}
func (t *testGithubClient) SupportsSingleFileDownload(repo models.Repo) bool {
	return false
}

func (t *testGithubClient) GetContents(owner, repo, branch, path string) ([]byte, error) {
	return []byte{}, nil

}
func (t *testGithubClient) GetRepoStatuses(repo models.Repo, pull models.PullRequest) ([]*github.RepoStatus, error) {
	return []*github.RepoStatus{}, nil

}
func (t *testGithubClient) GetRepoChecks(repo models.Repo, commitSHA string) ([]*github.CheckRun, error) {
	return []*github.CheckRun{}, nil

}

func (t *testGithubClient) GetPullRequest(repo models.Repo, pullNum int) (*github.PullRequest, error) {
	return t.ExpectedPull, nil

}
func (t *testGithubClient) GetPullRequestFromName(repoName string, repoOwner string, pullNum int) (*github.PullRequest, error) {
	return t.ExpectedPull, nil
}
