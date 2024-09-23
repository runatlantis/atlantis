package events_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/google/go-github/v63/github"
	"github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock/v4"

	"github.com/runatlantis/atlantis/server"
	events_controllers "github.com/runatlantis/atlantis/server/controllers/events"
	"github.com/runatlantis/atlantis/server/core/config"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/core/runtime"
	runtimemocks "github.com/runatlantis/atlantis/server/core/runtime/mocks"
	"github.com/runatlantis/atlantis/server/core/runtime/policy"
	"github.com/runatlantis/atlantis/server/core/terraform"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/events/webhooks"
	jobmocks "github.com/runatlantis/atlantis/server/jobs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics"
	. "github.com/runatlantis/atlantis/testing"
)

// In the e2e test, we use `conftest` not `conftest$version`.
// Because if depends on the version, we need to upgrade test base image before e2e fix it.
const conftestCommand = "conftest"

var applyLocker locking.ApplyLocker
var userConfig server.UserConfig

type NoopTFDownloader struct{}

var mockPreWorkflowHookRunner *runtimemocks.MockPreWorkflowHookRunner

var mockPostWorkflowHookRunner *runtimemocks.MockPostWorkflowHookRunner

func (m *NoopTFDownloader) GetAny(_, _ string) error {
	return nil
}

func (m *NoopTFDownloader) Install(_ string, _ string, _ *version.Version) (string, error) {
	return "", nil
}

type LocalConftestCache struct {
}

func (m *LocalConftestCache) Get(_ *version.Version) (string, error) {
	return exec.LookPath(conftestCommand)
}

func TestGitHubWorkflow(t *testing.T) {

	if testing.Short() {
		t.SkipNow()
	}
	// Ensure we have >= TF 0.14 locally.
	ensureRunning014(t)

	cases := []struct {
		Description string
		// RepoDir is relative to testdata/test-repos.
		RepoDir string
		// RepoConfigFile is path for atlantis.yaml
		RepoConfigFile string
		// ModifiedFiles are the list of files that have been modified in this
		// pull request.
		ModifiedFiles []string
		// Comments are what our mock user writes to the pull request.
		Comments []string
		// ApplyLock creates an apply lock that temporarily disables apply command
		ApplyLock bool
		// AllowCommands flag what kind of atlantis commands are available.
		AllowCommands []command.Name
		// DisableAutoplan flag disable auto plans when any pull request is opened.
		DisableAutoplan bool
		// DisablePreWorkflowHooks if set to true, pre-workflow hooks will be disabled
		DisablePreWorkflowHooks bool
		// ExpAutomerge is true if we expect Atlantis to automerge.
		ExpAutomerge bool
		// ExpAutoplan is true if we expect Atlantis to autoplan.
		ExpAutoplan bool
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
		// ExpAllowResponseCommentBack allow http response content with "Commenting back on pull request"
		ExpAllowResponseCommentBack bool
		// ExpParseFailedCount represents how many times test sends invalid commands
		ExpParseFailedCount int
		// ExpNoLocksToDelete whether we expect that there are no locks at the end to delete
		ExpNoLocksToDelete bool
	}{
		{
			Description:        "no comment or change",
			RepoDir:            "simple",
			ModifiedFiles:      []string{},
			Comments:           []string{},
			ExpReplies:         [][]string{},
			ExpNoLocksToDelete: true,
		},
		{
			Description:   "no comment",
			RepoDir:       "simple",
			ModifiedFiles: []string{"main.tf"},
			Comments:      []string{},
			ExpReplies: [][]string{
				{"exp-output-autoplan.txt"},
				{"exp-output-merge.txt"},
			},
			ExpAutoplan: true,
		},
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
			ExpAutoplan: true,
		},
		{
			Description:   "simple with plan comment",
			RepoDir:       "simple",
			ModifiedFiles: []string{"main.tf"},
			ExpAutoplan:   true,
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
			ExpAutoplan:   true,
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
			ExpAutoplan:   true,
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
			ExpAutoplan:   true,
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
			Description:   "simple with allow commands",
			RepoDir:       "simple",
			AllowCommands: []command.Name{command.Plan, command.Apply},
			Comments: []string{
				"atlantis import ADDRESS ID",
			},
			ExpReplies: [][]string{
				{"exp-output-allow-command-unknown-import.txt"},
			},
			ExpAllowResponseCommentBack: true,
			ExpParseFailedCount:         1,
			ExpNoLocksToDelete:          true,
		},
		{
			Description:   "simple with atlantis.yaml",
			RepoDir:       "simple-yaml",
			ModifiedFiles: []string{"main.tf"},
			ExpAutoplan:   true,
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
			Description:             "simple with atlantis.yaml - autoplan disabled",
			RepoDir:                 "simple-yaml",
			ModifiedFiles:           []string{"main.tf"},
			DisableAutoplan:         true,
			DisablePreWorkflowHooks: true,
			ExpAutoplan:             false,
			Comments: []string{
				"atlantis plan -w staging",
				"atlantis plan -w default",
				"atlantis apply -w staging",
			},
			ExpReplies: [][]string{
				{"exp-output-plan-staging.txt"},
				{"exp-output-plan-default.txt"},
				{"exp-output-apply-staging.txt"},
				{"exp-output-merge.txt"},
			},
		},
		{
			Description:   "simple with atlantis.yaml and apply all",
			RepoDir:       "simple-yaml",
			ModifiedFiles: []string{"main.tf"},
			ExpAutoplan:   true,
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
			Description:    "custom repo config file",
			RepoDir:        "repo-config-file",
			RepoConfigFile: "infrastructure/custom-name-atlantis.yaml",
			ModifiedFiles: []string{
				"infrastructure/staging/main.tf",
				"infrastructure/production/main.tf",
			},
			ExpAutoplan: true,
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
			Description:   "modules staging only",
			RepoDir:       "modules",
			ModifiedFiles: []string{"staging/main.tf"},
			ExpAutoplan:   true,
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
			Description:             "modules staging only - autoplan disabled",
			RepoDir:                 "modules",
			ModifiedFiles:           []string{"staging/main.tf"},
			DisableAutoplan:         true,
			DisablePreWorkflowHooks: true,
			ExpAutoplan:             false,
			Comments: []string{
				"atlantis plan -d staging",
				"atlantis apply -d staging",
			},
			ExpReplies: [][]string{
				{"exp-output-plan-staging.txt"},
				{"exp-output-apply-staging.txt"},
				{"exp-output-merge-only-staging.txt"},
			},
		},
		{
			Description:   "modules modules only",
			RepoDir:       "modules",
			ModifiedFiles: []string{"modules/null/main.tf"},
			ExpAutoplan:   false,
			Comments: []string{
				"atlantis plan -d staging",
				"atlantis plan -d production",
				"atlantis apply -d staging",
				"atlantis apply -d production",
			},
			ExpReplies: [][]string{
				{"exp-output-plan-staging.txt"},
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
			ExpAutoplan:   true,
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
			ExpAutoplan:   true,
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
			Description:   "tfvars no autoplan",
			RepoDir:       "tfvars-yaml-no-autoplan",
			ModifiedFiles: []string{"main.tf"},
			ExpAutoplan:   false,
			Comments: []string{
				"atlantis plan -p staging",
				"atlantis plan -p default",
				"atlantis apply -p staging",
				"atlantis apply -p default",
			},
			ExpReplies: [][]string{
				{"exp-output-plan-staging.txt"},
				{"exp-output-plan-default.txt"},
				{"exp-output-apply-staging.txt"},
				{"exp-output-apply-default.txt"},
				{"exp-output-merge.txt"},
			},
		},
		{
			Description:   "automerge",
			RepoDir:       "automerge",
			ExpAutomerge:  true,
			ExpAutoplan:   true,
			ModifiedFiles: []string{"dir1/main.tf", "dir2/main.tf"},
			Comments: []string{
				"atlantis apply -d dir1",
				"atlantis apply -d dir2",
			},
			ExpReplies: [][]string{
				{"exp-output-autoplan.txt"},
				{"exp-output-apply-dir1.txt"},
				{"exp-output-apply-dir2.txt"},
				{"exp-output-automerge.txt"},
				{"exp-output-merge.txt"},
			},
		},
		{
			Description:   "server-side cfg",
			RepoDir:       "server-side-cfg",
			ExpAutomerge:  false,
			ExpAutoplan:   true,
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
			ExpAutoplan:   true,
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
			ApplyLock:     true,
			ExpAutoplan:   true,
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
			Description:   "omitting apply from allow commands always takes presedence",
			RepoDir:       "simple-yaml",
			ModifiedFiles: []string{"main.tf"},
			AllowCommands: []command.Name{command.Plan},
			ApplyLock:     false,
			ExpAutoplan:   true,
			Comments: []string{
				"atlantis apply",
			},
			ExpParseFailedCount:         1,
			ExpAllowResponseCommentBack: true,
			ExpReplies: [][]string{
				{"exp-output-autoplan.txt"},
				// Disabling apply is implementing by omitting it from the apply list
				// See: https://github.com/runatlantis/atlantis/pull/2877
				{"exp-output-allow-command-unknown-apply.txt"},
				{"exp-output-merge.txt"},
			},
		},
		{
			Description:   "import single project",
			RepoDir:       "import-single-project",
			ModifiedFiles: []string{"main.tf"},
			ExpAutoplan:   true,
			Comments: []string{
				"atlantis import random_id.dummy1 AA",
				"atlantis apply",
				"atlantis import random_id.dummy2 BB",
				"atlantis plan",
			},
			ExpReplies: [][]string{
				{"exp-output-autoplan.txt"},
				{"exp-output-import-dummy1.txt"},
				{"exp-output-apply-no-projects.txt"},
				{"exp-output-import-dummy2.txt"},
				{"exp-output-plan-again.txt"},
				{"exp-output-merge.txt"},
			},
		},
		{
			Description: "import workspace",
			RepoDir:     "import-workspace",
			Comments: []string{
				"atlantis import -d dir1 -w ops 'random_id.dummy1[0]' AA",
				"atlantis import -p dir1-ops 'random_id.dummy2[0]' BB",
				"atlantis plan -p dir1-ops",
			},
			ExpReplies: [][]string{
				{"exp-output-import-dir1-ops-dummy1.txt"},
				{"exp-output-import-dir1-ops-dummy2.txt"},
				{"exp-output-plan.txt"},
				{"exp-output-merge.txt"},
			},
		},
		{
			Description:   "import single project with -var",
			RepoDir:       "import-single-project-var",
			ModifiedFiles: []string{"main.tf"},
			ExpAutoplan:   true,
			Comments: []string{
				"atlantis import 'random_id.for_each[\"overridden\"]' AA -- -var var=overridden",
				"atlantis import random_id.count[0] BB",
				"atlantis plan -- -var var=overridden",
			},
			ExpReplies: [][]string{
				{"exp-output-autoplan.txt"},
				{"exp-output-import-foreach.txt"},
				{"exp-output-import-count.txt"},
				{"exp-output-plan-again.txt"},
				{"exp-output-merge.txt"},
			},
		},
		{
			Description:   "import multiple project",
			RepoDir:       "import-multiple-project",
			ModifiedFiles: []string{"dir1/main.tf", "dir2/main.tf"},
			ExpAutoplan:   true,
			Comments: []string{
				"atlantis import random_id.dummy1 AA",
				"atlantis import -d dir1 random_id.dummy1 AA",
				"atlantis plan",
			},
			ExpReplies: [][]string{
				{"exp-output-autoplan.txt"},
				{"exp-output-import-multiple-projects.txt"},
				{"exp-output-import-dummy1.txt"},
				{"exp-output-plan-again.txt"},
				{"exp-output-merge.txt"},
			},
		},
		{
			Description:   "state rm single project",
			RepoDir:       "state-rm-single-project",
			ModifiedFiles: []string{"main.tf"},
			ExpAutoplan:   true,
			Comments: []string{
				"atlantis import random_id.simple AA",
				"atlantis import 'random_id.for_each[\"overridden\"]' BB -- -var var=overridden",
				"atlantis import random_id.count[0] BB",
				"atlantis plan -- -var var=overridden",
				"atlantis state rm 'random_id.for_each[\"overridden\"]' -- -lock=false",
				"atlantis state rm random_id.count[0] random_id.simple",
				"atlantis plan -- -var var=overridden",
			},
			ExpReplies: [][]string{
				{"exp-output-autoplan.txt"},
				{"exp-output-import-simple.txt"},
				{"exp-output-import-foreach.txt"},
				{"exp-output-import-count.txt"},
				{"exp-output-plan.txt"},
				{"exp-output-state-rm-foreach.txt"},
				{"exp-output-state-rm-multiple.txt"},
				{"exp-output-plan-again.txt"},
				{"exp-output-merged.txt"},
			},
		},
		{
			Description: "state rm workspace",
			RepoDir:     "state-rm-workspace",
			Comments: []string{
				"atlantis import -p dir1-ops 'random_id.dummy1[0]' AA",
				"atlantis plan -p dir1-ops",
				"atlantis state rm -p dir1-ops 'random_id.dummy1[0]'",
				"atlantis plan -p dir1-ops",
			},
			ExpReplies: [][]string{
				{"exp-output-import-dummy1.txt"},
				{"exp-output-plan.txt"},
				{"exp-output-state-rm-dummy1.txt"},
				{"exp-output-plan-again.txt"},
				{"exp-output-merge.txt"},
			},
		},
		{
			Description:   "state rm multiple project",
			RepoDir:       "state-rm-multiple-project",
			ModifiedFiles: []string{"dir1/main.tf", "dir2/main.tf"},
			ExpAutoplan:   true,
			Comments: []string{
				"atlantis import -d dir1 random_id.dummy AA",
				"atlantis import -d dir2 random_id.dummy BB",
				"atlantis plan",
				"atlantis state rm random_id.dummy",
				"atlantis plan",
			},
			ExpReplies: [][]string{
				{"exp-output-autoplan.txt"},
				{"exp-output-import-dummy1.txt"},
				{"exp-output-import-dummy2.txt"},
				{"exp-output-plan.txt"},
				{"exp-output-state-rm-multiple-projects.txt"},
				{"exp-output-plan-again.txt"},
				{"exp-output-merged.txt"},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			RegisterMockTestingT(t)

			// reset userConfig
			userConfig = server.UserConfig{}

			opt := setupOption{
				repoConfigFile:          c.RepoConfigFile,
				allowCommands:           c.AllowCommands,
				disableAutoplan:         c.DisableAutoplan,
				disablePreWorkflowHooks: c.DisablePreWorkflowHooks,
			}
			ctrl, vcsClient, githubGetter, atlantisWorkspace := setupE2E(t, c.RepoDir, opt)
			// Set the repo to be cloned through the testing backdoor.
			repoDir, headSHA := initializeRepo(t, c.RepoDir)
			atlantisWorkspace.TestingOverrideHeadCloneURL = fmt.Sprintf("file://%s", repoDir)

			// Setup test dependencies.
			w := httptest.NewRecorder()
			When(githubGetter.GetPullRequest(
				Any[logging.SimpleLogging](), Any[models.Repo](), Any[int]())).ThenReturn(GitHubPullRequestParsed(headSHA), nil)
			When(vcsClient.GetModifiedFiles(
				Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(c.ModifiedFiles, nil)

			// First, send the open pull request event which triggers autoplan.
			pullOpenedReq := GitHubPullRequestOpenedEvent(t, headSHA)
			ctrl.Post(w, pullOpenedReq)
			ResponseContains(t, w, 200, "Processing...")

			// Create global apply lock if required
			if c.ApplyLock {
				_, _ = applyLocker.LockApply()
			}

			// Now send any other comments.
			for _, comment := range c.Comments {
				commentReq := GitHubCommentEvent(t, comment)
				w = httptest.NewRecorder()
				ctrl.Post(w, commentReq)
				if c.ExpAllowResponseCommentBack {
					ResponseContains(t, w, 200, "Commenting back on pull request")
				} else {
					ResponseContains(t, w, 200, "Processing...")
				}
			}

			// Send the "pull closed" event which would be triggered by the
			// automerge or a manual merge.
			pullClosedReq := GitHubPullRequestClosedEvent(t)
			w = httptest.NewRecorder()
			ctrl.Post(w, pullClosedReq)
			ResponseContains(t, w, 200, "Pull request cleaned successfully")

			expNumHooks := len(c.Comments) - c.ExpParseFailedCount
			// if auto plan is disabled, hooks will not be called on pull request opened event
			if !c.DisableAutoplan {
				expNumHooks++
			}
			// Let's verify the pre-workflow hook was called for each comment including the pull request opened event
			if !c.DisablePreWorkflowHooks {
				mockPreWorkflowHookRunner.VerifyWasCalled(Times(expNumHooks)).Run(Any[models.WorkflowHookCommandContext](),
					Eq("some dummy command"), Any[string](), Any[string](), Any[string]())
			}
			// Let's verify the post-workflow hook was called for each comment including the pull request opened event
			mockPostWorkflowHookRunner.VerifyWasCalled(Times(expNumHooks)).Run(Any[models.WorkflowHookCommandContext](),
				Eq("some post dummy command"), Any[string](), Any[string](), Any[string]())

			// Now we're ready to verify Atlantis made all the comments back (or
			// replies) that we expect.  We expect each plan to have 1 comment,
			// and apply have 1 for each comment
			expNumReplies := len(c.Comments)

			// If there are locks to delete at the end, that will take a comment
			if !c.ExpNoLocksToDelete {
				expNumReplies++
			}

			if c.ExpAutoplan {
				expNumReplies++
			}

			if c.ExpAutomerge {
				expNumReplies++
			}

			_, _, _, actReplies, _ := vcsClient.VerifyWasCalled(Times(expNumReplies)).CreateComment(
				Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]()).GetAllCapturedArguments()
			Assert(t, len(c.ExpReplies) == len(actReplies), "missing expected replies, got %d but expected %d", len(actReplies), len(c.ExpReplies))
			for i, expReply := range c.ExpReplies {
				assertCommentEquals(t, expReply, actReplies[i], c.RepoDir, c.ExpParallel)
			}

			if c.ExpAutomerge {
				// Verify that the merge API call was made.
				vcsClient.VerifyWasCalledOnce().MergePull(Any[logging.SimpleLogging](), Any[models.PullRequest](), Any[models.PullRequestOptions]())
			} else {
				vcsClient.VerifyWasCalled(Never()).MergePull(Any[logging.SimpleLogging](), Any[models.PullRequest](), Any[models.PullRequestOptions]())
			}
		})
	}
}

func TestSimpleWorkflow_terraformLockFile(t *testing.T) {

	if testing.Short() {
		t.SkipNow()
	}
	// Ensure we have >= TF 0.14 locally.
	ensureRunning014(t)

	cases := []struct {
		Description string
		// RepoDir is relative to testdata/test-repos.
		RepoDir string
		// ModifiedFiles are the list of files that have been modified in this
		// pull request.
		ModifiedFiles []string
		// ExpAutoplan is true if we expect Atlantis to autoplan.
		ExpAutoplan bool
		// Comments are what our mock user writes to the pull request.
		Comments []string
		// ExpReplies is a list of files containing the expected replies that
		// Atlantis writes to the pull request in order. A reply from a parallel operation
		// will be matched using a substring check.
		ExpReplies [][]string
		// LockFileTracked deterims if the `.terraform.lock.hcl` file is tracked in git
		// if this is true we dont expect the lockfile to be modified by terraform init
		// if false we expect the lock file to be updated
		LockFileTracked bool
	}{
		{
			Description:   "simple with plan comment lockfile staged",
			RepoDir:       "simple-with-lockfile",
			ModifiedFiles: []string{"main.tf"},
			ExpAutoplan:   true,
			Comments: []string{
				"atlantis plan",
			},
			ExpReplies: [][]string{
				{"exp-output-autoplan.txt"},
				{"exp-output-plan.txt"},
			},
			LockFileTracked: true,
		},
		{
			Description:   "simple with plan comment lockfile not staged",
			RepoDir:       "simple-with-lockfile",
			ModifiedFiles: []string{"main.tf"},
			Comments: []string{
				"atlantis plan",
			},
			ExpReplies: [][]string{
				{"exp-output-autoplan.txt"},
				{"exp-output-plan.txt"},
			},
			LockFileTracked: false,
		},
		{
			Description:   "Modified .terraform.lock.hcl triggers autoplan ",
			RepoDir:       "simple-with-lockfile",
			ModifiedFiles: []string{".terraform.lock.hcl"},
			ExpAutoplan:   true,
			Comments: []string{
				"atlantis plan",
			},
			ExpReplies: [][]string{
				{"exp-output-autoplan.txt"},
				{"exp-output-plan.txt"},
			},
			LockFileTracked: true,
		},
	}
	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			RegisterMockTestingT(t)

			// reset userConfig
			userConfig = server.UserConfig{}

			ctrl, vcsClient, githubGetter, atlantisWorkspace := setupE2E(t, c.RepoDir, setupOption{})
			// Set the repo to be cloned through the testing backdoor.
			repoDir, headSHA := initializeRepo(t, c.RepoDir)

			oldLockFilePath, err := filepath.Abs(filepath.Join("testdata", "null_provider_lockfile_old_version"))
			Ok(t, err)
			oldLockFileContent, err := os.ReadFile(oldLockFilePath)
			Ok(t, err)

			if c.LockFileTracked {
				runCmd(t, "", "cp", oldLockFilePath, fmt.Sprintf("%s/.terraform.lock.hcl", repoDir))
				runCmd(t, repoDir, "git", "add", ".terraform.lock.hcl")
				runCmd(t, repoDir, "git", "commit", "-am", "stage .terraform.lock.hcl")
			}

			atlantisWorkspace.TestingOverrideHeadCloneURL = fmt.Sprintf("file://%s", repoDir)

			// Setup test dependencies.
			w := httptest.NewRecorder()
			When(githubGetter.GetPullRequest(Any[logging.SimpleLogging](), Any[models.Repo](), Any[int]())).ThenReturn(GitHubPullRequestParsed(headSHA), nil)
			When(vcsClient.GetModifiedFiles(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(c.ModifiedFiles, nil)

			// First, send the open pull request event which triggers autoplan.
			pullOpenedReq := GitHubPullRequestOpenedEvent(t, headSHA)
			ctrl.Post(w, pullOpenedReq)
			ResponseContains(t, w, 200, "Processing...")

			// check lock file content
			actualLockFileContent, err := os.ReadFile(fmt.Sprintf("%s/repos/runatlantis/atlantis-tests/2/default/.terraform.lock.hcl", atlantisWorkspace.DataDir))
			Ok(t, err)
			if c.LockFileTracked {
				if string(oldLockFileContent) != string(actualLockFileContent) {
					t.Error("Expected terraform.lock.hcl file not to be different as it has been staged")
					t.FailNow()
				}
			} else {
				if string(oldLockFileContent) == string(actualLockFileContent) {
					t.Error("Expected terraform.lock.hcl file to be different as it should have been updated")
					t.FailNow()
				}
			}

			if !c.LockFileTracked {
				// replace the lock file generated by the previous init to simulate
				// dependcies needing updating in a latter plan
				runCmd(t, "", "cp", oldLockFilePath, fmt.Sprintf("%s/repos/runatlantis/atlantis-tests/2/default/.terraform.lock.hcl", atlantisWorkspace.DataDir))
			}

			// Now send any other comments.
			for _, comment := range c.Comments {
				commentReq := GitHubCommentEvent(t, comment)
				w = httptest.NewRecorder()
				ctrl.Post(w, commentReq)
				ResponseContains(t, w, 200, "Processing...")
			}

			// check lock file content
			actualLockFileContent, err = os.ReadFile(fmt.Sprintf("%s/repos/runatlantis/atlantis-tests/2/default/.terraform.lock.hcl", atlantisWorkspace.DataDir))
			Ok(t, err)
			if c.LockFileTracked {
				if string(oldLockFileContent) != string(actualLockFileContent) {
					t.Error("Expected terraform.lock.hcl file not to be different as it has been staged")
					t.FailNow()
				}
			} else {
				if string(oldLockFileContent) == string(actualLockFileContent) {
					t.Error("Expected terraform.lock.hcl file to be different as it should have been updated")
					t.FailNow()
				}
			}

			// Let's verify the pre-workflow hook was called for each comment including the pull request opened event
			mockPreWorkflowHookRunner.VerifyWasCalled(Times(2)).Run(Any[models.WorkflowHookCommandContext](),
				Eq("some dummy command"), Any[string](), Any[string](), Any[string]())

			// Now we're ready to verify Atlantis made all the comments back (or
			// replies) that we expect.  We expect each plan to have 1 comment,
			// and apply have 1 for each comment plus one for the locks deleted at the
			// end.

			_, _, _, actReplies, _ := vcsClient.VerifyWasCalled(Times(2)).CreateComment(
				Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]()).GetAllCapturedArguments()
			Assert(t, len(c.ExpReplies) == len(actReplies), "missing expected replies, got %d but expected %d", len(actReplies), len(c.ExpReplies))
			for i, expReply := range c.ExpReplies {
				assertCommentEquals(t, expReply, actReplies[i], c.RepoDir, false)
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
	// Ensure we have conftest locally.
	ensureRunningConftest(t)

	cases := []struct {
		Description string
		// RepoDir is relative to testdata/test-repos.
		RepoDir string
		// ModifiedFiles are the list of files that have been modified in this
		// pull request.
		ModifiedFiles []string
		// Comments are what our mock user writes to the pull request.
		Comments []string
		// PolicyCheck is true if we expect Atlantis to run policy checking
		PolicyCheck bool
		// ExpAutomerge is true if we expect Atlantis to automerge.
		ExpAutomerge bool
		// ExpAutoplan is true if we expect Atlantis to autoplan.
		ExpAutoplan bool
		// ExpPolicyChecks is true if we expect Atlantis to execute policy checks
		ExpPolicyChecks bool
		// ExpQuietPolicyChecks is true if we expect Atlantis to exclude policy check output
		// when there's no error
		ExpQuietPolicyChecks bool
		// ExpQuietPolicyCheckFailure is true when we expect Atlantis to post back policy check failures
		// even when QuietPolicyChecks is enabled
		ExpQuietPolicyCheckFailure bool
		// ExpParallel is true if we expect Atlantis to run parallel plans or applies.
		ExpParallel bool
		// ExpReplies is a list of files containing the expected replies that
		// Atlantis writes to the pull request in order. A reply from a parallel operation
		// will be matched using a substring check.
		ExpReplies [][]string
	}{
		{
			Description:     "1 failing policy and 1 passing policy ",
			RepoDir:         "policy-checks-multi-projects",
			ModifiedFiles:   []string{"dir1/main.tf,", "dir2/main.tf"},
			PolicyCheck:     true,
			ExpAutoplan:     true,
			ExpPolicyChecks: true,
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
			Description:     "failing policy without policies passing using extra args",
			RepoDir:         "policy-checks-extra-args",
			ModifiedFiles:   []string{"main.tf"},
			PolicyCheck:     true,
			ExpAutoplan:     true,
			ExpPolicyChecks: true,
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
			Description:     "failing policy without policies passing",
			RepoDir:         "policy-checks",
			ModifiedFiles:   []string{"main.tf"},
			PolicyCheck:     true,
			ExpAutoplan:     true,
			ExpPolicyChecks: true,
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
			Description:     "failing policy without policies passing and custom run steps",
			RepoDir:         "policy-checks-custom-run-steps",
			ModifiedFiles:   []string{"main.tf"},
			PolicyCheck:     true,
			ExpAutoplan:     true,
			ExpPolicyChecks: true,
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
			Description:     "failing policy additional apply requirements specified",
			RepoDir:         "policy-checks-apply-reqs",
			ModifiedFiles:   []string{"main.tf"},
			PolicyCheck:     true,
			ExpAutoplan:     true,
			ExpPolicyChecks: true,
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
			Description:     "failing policy approved by non owner",
			RepoDir:         "policy-checks-diff-owner",
			ModifiedFiles:   []string{"main.tf"},
			PolicyCheck:     true,
			ExpAutoplan:     true,
			ExpPolicyChecks: true,
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
		{
			Description:          "successful policy checks with quiet flag enabled",
			RepoDir:              "policy-checks-success-silent",
			ModifiedFiles:        []string{"main.tf"},
			PolicyCheck:          true,
			ExpAutoplan:          true,
			ExpPolicyChecks:      true,
			ExpQuietPolicyChecks: true,
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
			Description:                "failing policy checks with quiet flag enabled",
			RepoDir:                    "policy-checks",
			ModifiedFiles:              []string{"main.tf"},
			PolicyCheck:                true,
			ExpAutoplan:                true,
			ExpPolicyChecks:            true,
			ExpQuietPolicyChecks:       true,
			ExpQuietPolicyCheckFailure: true,
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
			Description:     "failing policy with approval and policy approval clear",
			RepoDir:         "policy-checks-clear-approval",
			ModifiedFiles:   []string{"main.tf"},
			PolicyCheck:     true,
			ExpAutoplan:     true,
			ExpPolicyChecks: true,
			Comments: []string{
				"atlantis approve_policies",
				"atlantis approve_policies --clear-policy-approval",
				"atlantis apply",
			},
			ExpReplies: [][]string{
				{"exp-output-autoplan.txt"},
				{"exp-output-auto-policy-check.txt"},
				{"exp-output-approve-policies-success.txt"},
				{"exp-output-approve-policies-clear.txt"},
				{"exp-output-apply-failed.txt"},
				{"exp-output-merge.txt"},
			},
		},
		{
			Description:     "policy checking disabled on specific repo",
			RepoDir:         "policy-checks-disabled-repo",
			ModifiedFiles:   []string{"main.tf"},
			PolicyCheck:     true,
			ExpAutoplan:     true,
			ExpPolicyChecks: false,
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
			Description:     "policy checking disabled on specific repo server side",
			RepoDir:         "policy-checks-disabled-repo-server-side",
			ModifiedFiles:   []string{"main.tf"},
			PolicyCheck:     true,
			ExpAutoplan:     true,
			ExpPolicyChecks: false,
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
			Description:     "policy checking enabled on specific repo but disabled globally",
			RepoDir:         "policy-checks-enabled-repo",
			ModifiedFiles:   []string{"main.tf"},
			PolicyCheck:     false,
			ExpAutoplan:     true,
			ExpPolicyChecks: false,
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
			Description:     "policy checking enabled on specific repo server side but disabled globally",
			RepoDir:         "policy-checks-enabled-repo-server-side",
			ModifiedFiles:   []string{"main.tf"},
			PolicyCheck:     false,
			ExpAutoplan:     true,
			ExpPolicyChecks: false,
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
			Description:     "policy checking disabled on previous regex match but not on repo",
			RepoDir:         "policy-checks-disabled-previous-match",
			ModifiedFiles:   []string{"main.tf"},
			PolicyCheck:     true,
			ExpAutoplan:     true,
			ExpPolicyChecks: false,
			Comments: []string{
				"atlantis apply",
			},
			ExpReplies: [][]string{
				{"exp-output-autoplan.txt"},
				{"exp-output-apply.txt"},
				{"exp-output-merge.txt"},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			RegisterMockTestingT(t)

			// reset userConfig
			userConfig = server.UserConfig{}
			userConfig.EnablePolicyChecksFlag = c.PolicyCheck
			userConfig.QuietPolicyChecks = c.ExpQuietPolicyChecks

			ctrl, vcsClient, githubGetter, atlantisWorkspace := setupE2E(t, c.RepoDir, setupOption{})

			// Set the repo to be cloned through the testing backdoor.
			repoDir, headSHA := initializeRepo(t, c.RepoDir)
			atlantisWorkspace.TestingOverrideHeadCloneURL = fmt.Sprintf("file://%s", repoDir)

			// Setup test dependencies.
			w := httptest.NewRecorder()
			When(vcsClient.PullIsMergeable(
				Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Eq("atlantis-test"))).ThenReturn(true, nil)
			When(vcsClient.PullIsApproved(
				Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(models.ApprovalStatus{
				IsApproved: true,
			}, nil)
			When(githubGetter.GetPullRequest(
				Any[logging.SimpleLogging](), Any[models.Repo](), Any[int]())).ThenReturn(GitHubPullRequestParsed(headSHA), nil)
			When(vcsClient.GetModifiedFiles(
				Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(c.ModifiedFiles, nil)

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
			// automerge or a manual merge.
			pullClosedReq := GitHubPullRequestClosedEvent(t)
			w = httptest.NewRecorder()
			ctrl.Post(w, pullClosedReq)
			ResponseContains(t, w, 200, "Pull request cleaned successfully")

			// Now we're ready to verify Atlantis made all the comments back (or
			// replies) that we expect.  We expect each plan to have 2 comments,
			// one for plan one for policy check and apply have 1 for each
			// comment plus one for the locks deleted at the end.
			expNumReplies := len(c.Comments) + 1

			if c.ExpAutoplan {
				expNumReplies++
				expNumReplies++
			}

			var planRegex = regexp.MustCompile("plan")
			for _, comment := range c.Comments {
				if planRegex.MatchString(comment) {
					expNumReplies++
				}
			}

			if c.ExpAutomerge {
				expNumReplies++
			}

			if c.ExpQuietPolicyChecks && !c.ExpQuietPolicyCheckFailure {
				expNumReplies--
			}

			if !c.ExpPolicyChecks {
				expNumReplies--
			}
			_, _, _, actReplies, _ := vcsClient.VerifyWasCalled(Times(expNumReplies)).CreateComment(
				Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]()).GetAllCapturedArguments()

			Assert(t, len(c.ExpReplies) == len(actReplies), "missing expected replies, got %d but expected %d", len(actReplies), len(c.ExpReplies))
			for i, expReply := range c.ExpReplies {
				assertCommentEquals(t, expReply, actReplies[i], c.RepoDir, c.ExpParallel)
			}

			if c.ExpAutomerge {
				// Verify that the merge API call was made.
				vcsClient.VerifyWasCalledOnce().MergePull(Any[logging.SimpleLogging](), Any[models.PullRequest](), Any[models.PullRequestOptions]())
			} else {
				vcsClient.VerifyWasCalled(Never()).MergePull(Any[logging.SimpleLogging](), Any[models.PullRequest](), Any[models.PullRequestOptions]())
			}
		})
	}
}

type setupOption struct {
	repoConfigFile          string
	allowCommands           []command.Name
	disableAutoplan         bool
	disablePreWorkflowHooks bool
}

func setupE2E(t *testing.T, repoDir string, opt setupOption) (events_controllers.VCSEventsController, *vcsmocks.MockClient, *mocks.MockGithubPullGetter, *events.FileWorkspace) {
	allowForkPRs := false
	discardApprovalOnPlan := true
	dataDir, binDir, cacheDir := mkSubDirs(t)

	// Mocks.
	e2eVCSClient := vcsmocks.NewMockClient()
	e2eStatusUpdater := &events.DefaultCommitStatusUpdater{Client: e2eVCSClient}
	e2eGithubGetter := mocks.NewMockGithubPullGetter()
	e2eGitlabGetter := mocks.NewMockGitlabMergeRequestGetter()
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()

	// Real dependencies.
	logging.SuppressDefaultLogging()
	logger := logging.NewNoopLogger(t)

	eventParser := &events.EventParser{
		GithubUser:  "github-user",
		GithubToken: "github-token",
		GitlabUser:  "gitlab-user",
		GitlabToken: "gitlab-token",
	}
	allowCommands := command.AllCommentCommands
	if opt.allowCommands != nil {
		allowCommands = opt.allowCommands
	}
	disableApply := true
	disableGlobalApplyLock := false
	for _, allowCommand := range allowCommands {
		if allowCommand == command.Apply {
			disableApply = false
			break
		}
	}
	commentParser := &events.CommentParser{
		GithubUser:     "github-user",
		GitlabUser:     "gitlab-user",
		ExecutableName: "atlantis",
		AllowCommands:  allowCommands,
	}
	terraformClient, err := terraform.NewClient(logger, binDir, cacheDir, "", "", "", "default-tf-version", "https://releases.hashicorp.com", &NoopTFDownloader{}, true, false, projectCmdOutputHandler)
	Ok(t, err)
	boltdb, err := db.New(dataDir)
	Ok(t, err)
	backend := boltdb
	lockingClient := locking.NewClient(boltdb)
	noOpLocker := locking.NewNoOpLocker()
	applyLocker = locking.NewApplyClient(boltdb, disableApply, disableGlobalApplyLock)
	projectLocker := &events.DefaultProjectLocker{
		Locker:     lockingClient,
		NoOpLocker: noOpLocker,
		VCSClient:  e2eVCSClient,
	}
	workingDir := &events.FileWorkspace{
		DataDir:                     dataDir,
		TestingOverrideHeadCloneURL: "override-me",
	}
	var preWorkflowHooks []*valid.WorkflowHook
	if !opt.disablePreWorkflowHooks {
		preWorkflowHooks = []*valid.WorkflowHook{
			{
				StepName:   "global_hook",
				RunCommand: "some dummy command",
			},
		}
	}

	defaultTFVersion := terraformClient.DefaultVersion()
	locker := events.NewDefaultWorkingDirLocker()
	parser := &config.ParserValidator{}

	globalCfgArgs := valid.GlobalCfgArgs{
		RepoConfigFile:       opt.repoConfigFile,
		AllowAllRepoSettings: true,
		PreWorkflowHooks:     preWorkflowHooks,
		PostWorkflowHooks: []*valid.WorkflowHook{
			{
				StepName:   "global_hook",
				RunCommand: "some post dummy command",
			},
		},
		PolicyCheckEnabled: userConfig.EnablePolicyChecksFlag,
	}
	globalCfg := valid.NewGlobalCfgFromArgs(globalCfgArgs)
	expCfgPath := filepath.Join(absRepoPath(t, repoDir), "repos.yaml")
	if _, err := os.Stat(expCfgPath); err == nil {
		globalCfg, err = parser.ParseGlobalCfg(expCfgPath, globalCfg)
		Ok(t, err)
	}
	drainer := &events.Drainer{}

	parallelPoolSize := 1
	silenceNoProjects := false

	disableUnlockLabel := "do-not-unlock"

	statusUpdater := runtimemocks.NewMockStatusUpdater()
	commitStatusUpdater := mocks.NewMockCommitStatusUpdater()
	asyncTfExec := runtimemocks.NewMockAsyncTFExec()

	mockPreWorkflowHookRunner = runtimemocks.NewMockPreWorkflowHookRunner()
	preWorkflowHookURLGenerator := mocks.NewMockPreWorkflowHookURLGenerator()
	preWorkflowHooksCommandRunner := &events.DefaultPreWorkflowHooksCommandRunner{
		VCSClient:             e2eVCSClient,
		GlobalCfg:             globalCfg,
		WorkingDirLocker:      locker,
		WorkingDir:            workingDir,
		PreWorkflowHookRunner: mockPreWorkflowHookRunner,
		CommitStatusUpdater:   commitStatusUpdater,
		Router:                preWorkflowHookURLGenerator,
	}

	mockPostWorkflowHookRunner = runtimemocks.NewMockPostWorkflowHookRunner()
	postWorkflowHookURLGenerator := mocks.NewMockPostWorkflowHookURLGenerator()
	postWorkflowHooksCommandRunner := &events.DefaultPostWorkflowHooksCommandRunner{
		VCSClient:              e2eVCSClient,
		GlobalCfg:              globalCfg,
		WorkingDirLocker:       locker,
		WorkingDir:             workingDir,
		PostWorkflowHookRunner: mockPostWorkflowHookRunner,
		CommitStatusUpdater:    commitStatusUpdater,
		Router:                 postWorkflowHookURLGenerator,
	}
	statsScope, _, _ := metrics.NewLoggingScope(logger, "atlantis")

	projectCommandBuilder := events.NewProjectCommandBuilder(
		userConfig.EnablePolicyChecksFlag,
		parser,
		&events.DefaultProjectFinder{},
		e2eVCSClient,
		workingDir,
		locker,
		globalCfg,
		&events.DefaultPendingPlanFinder{},
		commentParser,
		false,
		false,
		false,
		false,
		false,
		"",
		"**/*.tf,**/*.tfvars,**/*.tfvars.json,**/terragrunt.hcl,**/.terraform.lock.hcl",
		false,
		false,
		false,
		"auto",
		statsScope,
		terraformClient,
	)

	showStepRunner, err := runtime.NewShowStepRunner(terraformClient, defaultTFVersion)

	Ok(t, err)

	conftextExec := policy.NewConfTestExecutorWorkflow(logger, binDir, &NoopTFDownloader{})

	// swapping out version cache to something that always returns local conftest
	// binary
	conftextExec.VersionCache = &LocalConftestCache{}

	policyCheckRunner, err := runtime.NewPolicyCheckStepRunner(
		defaultTFVersion,
		conftextExec,
	)

	Ok(t, err)

	projectCommandRunner := &events.DefaultProjectCommandRunner{
		VcsClient:        e2eVCSClient,
		Locker:           projectLocker,
		LockURLGenerator: &mockLockURLGenerator{},
		InitStepRunner: &runtime.InitStepRunner{
			TerraformExecutor: terraformClient,
			DefaultTFVersion:  defaultTFVersion,
		},
		PlanStepRunner: runtime.NewPlanStepRunner(
			terraformClient,
			defaultTFVersion,
			statusUpdater,
			asyncTfExec,
		),
		ShowStepRunner:        showStepRunner,
		PolicyCheckStepRunner: policyCheckRunner,
		ApplyStepRunner: &runtime.ApplyStepRunner{
			TerraformExecutor: terraformClient,
		},
		ImportStepRunner:  runtime.NewImportStepRunner(terraformClient, defaultTFVersion),
		StateRmStepRunner: runtime.NewStateRmStepRunner(terraformClient, defaultTFVersion),
		RunStepRunner: &runtime.RunStepRunner{
			TerraformExecutor:       terraformClient,
			DefaultTFVersion:        defaultTFVersion,
			ProjectCmdOutputHandler: projectCmdOutputHandler,
		},
		WorkingDir:       workingDir,
		Webhooks:         &mockWebhookSender{},
		WorkingDirLocker: locker,
		CommandRequirementHandler: &events.DefaultCommandRequirementHandler{
			WorkingDir: workingDir,
		},
	}

	dbUpdater := &events.DBUpdater{
		Backend: backend,
	}

	pullUpdater := &events.PullUpdater{
		HidePrevPlanComments: false,
		VCSClient:            e2eVCSClient,
		MarkdownRenderer:     events.NewMarkdownRenderer(false, false, false, false, false, false, "", "atlantis", false),
	}

	autoMerger := &events.AutoMerger{
		VCSClient:       e2eVCSClient,
		GlobalAutomerge: false,
	}

	policyCheckCommandRunner := events.NewPolicyCheckCommandRunner(
		dbUpdater,
		pullUpdater,
		e2eStatusUpdater,
		projectCommandRunner,
		parallelPoolSize,
		false,
		userConfig.QuietPolicyChecks,
	)

	e2ePullReqStatusFetcher := vcs.NewPullReqStatusFetcher(e2eVCSClient, "atlantis-test")

	planCommandRunner := events.NewPlanCommandRunner(
		false,
		false,
		e2eVCSClient,
		&events.DefaultPendingPlanFinder{},
		workingDir,
		e2eStatusUpdater,
		projectCommandBuilder,
		projectCommandRunner,
		dbUpdater,
		pullUpdater,
		policyCheckCommandRunner,
		autoMerger,
		parallelPoolSize,
		silenceNoProjects,
		boltdb,
		lockingClient,
		discardApprovalOnPlan,
		e2ePullReqStatusFetcher,
	)

	applyCommandRunner := events.NewApplyCommandRunner(
		e2eVCSClient,
		false,
		applyLocker,
		e2eStatusUpdater,
		projectCommandBuilder,
		projectCommandRunner,
		autoMerger,
		pullUpdater,
		dbUpdater,
		boltdb,
		parallelPoolSize,
		silenceNoProjects,
		false,
		e2ePullReqStatusFetcher,
	)

	approvePoliciesCommandRunner := events.NewApprovePoliciesCommandRunner(
		e2eStatusUpdater,
		projectCommandBuilder,
		projectCommandRunner,
		pullUpdater,
		dbUpdater,
		silenceNoProjects,
		false,
		e2eVCSClient,
	)

	unlockCommandRunner := events.NewUnlockCommandRunner(
		mocks.NewMockDeleteLockCommand(),
		e2eVCSClient,
		silenceNoProjects,
		disableUnlockLabel,
	)

	versionCommandRunner := events.NewVersionCommandRunner(
		pullUpdater,
		projectCommandBuilder,
		projectCommandRunner,
		parallelPoolSize,
		silenceNoProjects,
	)

	importCommandRunner := events.NewImportCommandRunner(
		pullUpdater,
		e2ePullReqStatusFetcher,
		projectCommandBuilder,
		projectCommandRunner,
		silenceNoProjects,
	)

	stateCommandRunner := events.NewStateCommandRunner(
		pullUpdater,
		projectCommandBuilder,
		projectCommandRunner,
	)

	commentCommandRunnerByCmd := map[command.Name]events.CommentCommandRunner{
		command.Plan:            planCommandRunner,
		command.Apply:           applyCommandRunner,
		command.ApprovePolicies: approvePoliciesCommandRunner,
		command.Unlock:          unlockCommandRunner,
		command.Version:         versionCommandRunner,
		command.Import:          importCommandRunner,
		command.State:           stateCommandRunner,
	}

	commandRunner := &events.DefaultCommandRunner{
		EventParser:                    eventParser,
		VCSClient:                      e2eVCSClient,
		GithubPullGetter:               e2eGithubGetter,
		GitlabMergeRequestGetter:       e2eGitlabGetter,
		Logger:                         logger,
		GlobalCfg:                      globalCfg,
		StatsScope:                     statsScope,
		AllowForkPRs:                   allowForkPRs,
		AllowForkPRsFlag:               "allow-fork-prs",
		CommentCommandRunnerByCmd:      commentCommandRunnerByCmd,
		Drainer:                        drainer,
		PreWorkflowHooksCommandRunner:  preWorkflowHooksCommandRunner,
		PostWorkflowHooksCommandRunner: postWorkflowHooksCommandRunner,
		PullStatusFetcher:              backend,
		DisableAutoplan:                opt.disableAutoplan,
	}

	repoAllowlistChecker, err := events.NewRepoAllowlistChecker("*")
	Ok(t, err)

	ctrl := events_controllers.VCSEventsController{
		TestingMode:   true,
		CommandRunner: commandRunner,
		PullCleaner: &events.PullClosedExecutor{
			Locker:                   lockingClient,
			VCSClient:                e2eVCSClient,
			WorkingDir:               workingDir,
			Backend:                  backend,
			PullClosedTemplate:       &events.PullClosedEventTemplate{},
			LogStreamResourceCleaner: projectCmdOutputHandler,
		},
		Logger:                       logger,
		Scope:                        statsScope,
		Parser:                       eventParser,
		CommentParser:                commentParser,
		GithubWebhookSecret:          nil,
		GithubRequestValidator:       &events_controllers.DefaultGithubRequestValidator{},
		GitlabRequestParserValidator: &events_controllers.DefaultGitlabRequestParserValidator{},
		GitlabWebhookSecret:          nil,
		RepoAllowlistChecker:         repoAllowlistChecker,
		SupportedVCSHosts:            []models.VCSHostType{models.Gitlab, models.Github, models.BitbucketCloud},
		VCSClient:                    e2eVCSClient,
	}
	return ctrl, e2eVCSClient, e2eGithubGetter, workingDir
}

type mockLockURLGenerator struct{}

func (m *mockLockURLGenerator) GenerateLockURL(_ string) string {
	return "lock-url"
}

type mockWebhookSender struct{}

func (w *mockWebhookSender) Send(_ logging.SimpleLogging, _ webhooks.ApplyResult) error {
	return nil
}

func GitHubCommentEvent(t *testing.T, comment string) *http.Request {
	requestJSON, err := os.ReadFile(filepath.Join("testdata", "githubIssueCommentEvent.json"))
	Ok(t, err)
	escapedComment, err := json.Marshal(comment)
	Ok(t, err)
	requestJSON = []byte(strings.Replace(string(requestJSON), "\"###comment body###\"", string(escapedComment), 1))
	req, err := http.NewRequest("POST", "/events", bytes.NewBuffer(requestJSON))
	Ok(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(githubHeader, "issue_comment")
	return req
}

func GitHubPullRequestOpenedEvent(t *testing.T, headSHA string) *http.Request {
	requestJSON, err := os.ReadFile(filepath.Join("testdata", "githubPullRequestOpenedEvent.json"))
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
	requestJSON, err := os.ReadFile(filepath.Join("testdata", "githubPullRequestClosedEvent.json"))
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
			Ref: github.String("main"),
		},
		User: &github.User{
			Login: github.String("atlantisbot"),
		},
	}
}

// absRepoPath returns the absolute path to the test repo under dir repoDir.
func absRepoPath(t *testing.T, repoDir string) string {
	path, err := filepath.Abs(filepath.Join("testdata", "test-repos", repoDir))
	Ok(t, err)
	return path
}

// initializeRepo copies the repo data from testdata and initializes a new
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
			// If in CI, we write the diff to the console. Otherwise we write the diff
			// to file so we can use our local diff viewer.
			if os.Getenv("CI") == "true" {
				t.Logf("exp: %s, got: %s", expStr, act)
				t.FailNow()
			} else {
				actFile := filepath.Join(absRepoPath(t, repoDir), expFile+".act")
				err := os.WriteFile(actFile, []byte(act), 0600)
				Ok(t, err)
				cwd, err := os.Getwd()
				Ok(t, err)
				rel, err := filepath.Rel(cwd, actFile)
				Ok(t, err)
				t.Errorf("%q was different, wrote actual comment to %q", expFile, rel)
			}
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

// Will fail test if conftest isn't in path
func ensureRunningConftest(t *testing.T) {
	// use `conftest` command instead `conftest$version`, so tests may fail on the environment cause the output logs may become change by version.
	t.Logf("conftest check may fail depends on conftest version. please use latest stable conftest.")
	_, err := exec.LookPath(conftestCommand)
	if err != nil {
		t.Logf(`%s must be installed to run this test
- on local, please install conftest command or run 'make docker/test-all'
- on CI, please check testing-env docker image contains conftest command. see testing/Dockerfile
`, conftestCommand)
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
