package server_test

import (
	"fmt"
	"net/http/httptest"
	"regexp"
	"testing"

	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/events/mocks/matchers"
	. "github.com/runatlantis/atlantis/testing"
)

func TestGitHubWorkflowWithPolicyCheck(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	// Ensure we have >= TF 0.12 locally.
	ensureRunning012(t)

	cases := []struct {
		Description string
		// RepoDir is relative to testfixtures/test-repos.
		RepoDir string
		// ModifiedFiles are the list of files that have been modified in this
		// pull request.
		ModifiedFiles []string
		// Comments are what our mock user writes to the pull request.
		Comments []string
		// ExpAutomerge is true if we expect Atlantis to automerge.
		ExpAutomerge bool
		// ExpAutoplan is true if we expect Atlantis to autoplan.
		ExpAutoplan bool
		// ExpParallel is true if we expect Atlantis to run parallel plans or applies.
		ExpParallel bool
		// ExpReplies is a list of files containing the expected replies that
		// Atlantis writes to the pull request in order. A reply from a parallel operation
		// will be matched using a substring check.
		ExpReplies [][]string
		// PolicyCheckEnabled runs integration tests through PolicyCheckProjectCommandBuilder.
		PolicyCheckEnabled bool
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
				{"exp-output-auto-policy-check.txt"},
				{"exp-output-apply.txt"},
				{"exp-output-merge.txt"},
			},
			ExpAutoplan:        true,
			PolicyCheckEnabled: true,
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
				{"exp-output-auto-policy-check.txt"},
				{"exp-output-autoplan.txt"},
				{"exp-output-auto-policy-check.txt"},
				{"exp-output-apply.txt"},
				{"exp-output-merge.txt"},
			},
			PolicyCheckEnabled: true,
		},
		{
			Description:   "policy check enabled: simple with plan comment",
			RepoDir:       "simple",
			ModifiedFiles: []string{"main.tf"},
			ExpAutoplan:   true,
			Comments: []string{
				"atlantis plan",
				"atlantis apply",
			},
			ExpReplies: [][]string{
				{"exp-output-autoplan.txt"},
				{"exp-output-auto-policy-check.txt"},
				{"exp-output-autoplan.txt"},
				{"exp-output-auto-policy-check.txt"},
				{"exp-output-apply.txt"},
				{"exp-output-merge.txt"},
			},
			PolicyCheckEnabled: true,
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
				{"exp-output-auto-policy-check.txt"},
				{"exp-output-atlantis-plan-var-overridden.txt"},
				{"exp-output-atlantis-policy-check.txt"},
				{"exp-output-apply-var.txt"},
				{"exp-output-merge.txt"},
			},
			PolicyCheckEnabled: true,
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
				{"exp-output-auto-policy-check.txt"},
				{"exp-output-atlantis-plan.txt"},
				{"exp-output-atlantis-policy-check.txt"},
				{"exp-output-atlantis-plan-new-workspace.txt"},
				{"exp-output-atlantis-policy-check.txt"},
				{"exp-output-apply-var-default-workspace.txt"},
				{"exp-output-apply-var-new-workspace.txt"},
				{"exp-output-merge-workspaces.txt"},
			},
			PolicyCheckEnabled: true,
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
				{"exp-output-auto-policy-check.txt"},
				{"exp-output-atlantis-plan.txt"},
				{"exp-output-atlantis-policy-check.txt"},
				{"exp-output-atlantis-plan-new-workspace.txt"},
				{"exp-output-atlantis-policy-check.txt"},
				{"exp-output-apply-var-all.txt"},
				{"exp-output-merge-workspaces.txt"},
			},
			PolicyCheckEnabled: true,
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
				{"exp-output-auto-policy-check.txt"},
				{"exp-output-apply-staging.txt"},
				{"exp-output-apply-default.txt"},
				{"exp-output-merge.txt"},
			},
			PolicyCheckEnabled: true,
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
				{"exp-output-auto-policy-check.txt"},
				{"exp-output-apply-all.txt"},
				{"exp-output-merge.txt"},
			},
			PolicyCheckEnabled: true,
		},
		{
			Description:   "simple with atlantis.yaml and plan/apply all",
			RepoDir:       "simple-yaml",
			ModifiedFiles: []string{"main.tf"},
			ExpAutoplan:   true,
			Comments: []string{
				"atlantis plan",
				"atlantis apply",
			},
			ExpReplies: [][]string{
				{"exp-output-autoplan.txt"},
				{"exp-output-auto-policy-check.txt"},
				{"exp-output-autoplan.txt"},
				{"exp-output-auto-policy-check.txt"},
				{"exp-output-apply-all.txt"},
				{"exp-output-merge.txt"},
			},
			PolicyCheckEnabled: true,
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
				{"exp-output-auto-policy-check.txt"},
				{"exp-output-apply-staging.txt"},
				{"exp-output-merge-only-staging.txt"},
			},
			PolicyCheckEnabled: true,
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
				{"exp-output-policy-check-staging.txt"},
				{"exp-output-plan-production.txt"},
				{"exp-output-policy-check-production.txt"},
				{"exp-output-apply-staging.txt"},
				{"exp-output-apply-production.txt"},
				{"exp-output-merge-all-dirs.txt"},
			},
			PolicyCheckEnabled: true,
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
				{"exp-output-auto-policy-check.txt"},
				{"exp-output-apply-staging.txt"},
				{"exp-output-apply-production.txt"},
				{"exp-output-merge.txt"},
			},
			PolicyCheckEnabled: true,
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
				{"exp-output-auto-policy-check.txt"},
				{"exp-output-apply-staging.txt"},
				{"exp-output-apply-default.txt"},
				{"exp-output-merge.txt"},
			},
			PolicyCheckEnabled: true,
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
				{"exp-output-policy-check-staging.txt"},
				{"exp-output-plan-default.txt"},
				{"exp-output-policy-check-default.txt"},
				{"exp-output-apply-staging.txt"},
				{"exp-output-apply-default.txt"},
				{"exp-output-merge.txt"},
			},
			PolicyCheckEnabled: true,
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
				{"exp-output-auto-policy-check.txt"},
				{"exp-output-apply-dir1.txt"},
				{"exp-output-apply-dir2.txt"},
				{"exp-output-automerge.txt"},
				{"exp-output-merge.txt"},
			},
			PolicyCheckEnabled: true,
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
				{"exp-output-auto-policy-check.txt"},
				{"exp-output-apply-staging-workspace.txt"},
				{"exp-output-apply-default-workspace.txt"},
				{"exp-output-merge.txt"},
			},
			PolicyCheckEnabled: true,
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
				{"exp-output-auto-policy-check.txt", "exp-output-auto-policy-check.txt"},
				{"exp-output-apply-all-staging.txt", "exp-output-apply-all-production.txt"},
				{"exp-output-merge.txt"},
			},
			PolicyCheckEnabled: true,
		},
	}
	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			RegisterMockTestingT(t)

			ctrl, vcsClient, githubGetter, atlantisWorkspace := setupE2E(t, c.RepoDir, c.PolicyCheckEnabled)
			// Set the repo to be cloned through the testing backdoor.
			repoDir, headSHA, cleanup := initializeRepo(t, c.RepoDir)
			defer cleanup()
			atlantisWorkspace.TestingOverrideHeadCloneURL = fmt.Sprintf("file://%s", repoDir)

			// Setup test dependencies.
			w := httptest.NewRecorder()
			When(githubGetter.GetPullRequest(AnyRepo(), AnyInt())).ThenReturn(GitHubPullRequestParsed(headSHA), nil)
			When(vcsClient.GetModifiedFiles(AnyRepo(), matchers.AnyModelsPullRequest())).ThenReturn(c.ModifiedFiles, nil)

			// First, send the open pull request event which triggers autoplan.
			pullOpenedReq := GitHubPullRequestOpenedEvent(t, headSHA)
			ctrl.Post(w, pullOpenedReq)
			responseContains(t, w, 200, "Processing...")

			// Now send any other comments.
			for _, comment := range c.Comments {
				commentReq := GitHubCommentEvent(t, comment)
				w = httptest.NewRecorder()
				ctrl.Post(w, commentReq)
				responseContains(t, w, 200, "Processing...")
			}

			// Send the "pull closed" event which would be triggered by the
			// automerge or a manual merge.
			pullClosedReq := GitHubPullRequestClosedEvent(t)
			w = httptest.NewRecorder()
			ctrl.Post(w, pullClosedReq)
			responseContains(t, w, 200, "Pull request cleaned successfully")

			// Now we're ready to verify Atlantis made all the comments back (or
			// replies) that we expect.  We expect each plan to have 2 comments,
			// one for plan one for policy check and apply have 1 for each
			// comment plus one for the locks deleted at the end.
			expNumReplies := len(c.Comments) + 1

			if c.ExpAutoplan {
				expNumReplies++
			}

			// When enabled policy_check runs right after plan. So whenever
			// comment matches plan we add additional call to expected
			// number.
			if c.PolicyCheckEnabled {
				var planRegex = regexp.MustCompile("plan")
				for _, comment := range c.Comments {
					if planRegex.MatchString(comment) {
						expNumReplies++
					}
				}

				// Adding 1 for policy_check autorun
				if c.ExpAutoplan {
					expNumReplies++
				}
			}

			if c.ExpAutomerge {
				expNumReplies++
			}

			_, _, actReplies, _ := vcsClient.VerifyWasCalled(Times(expNumReplies)).CreateComment(AnyRepo(), AnyInt(), AnyString(), AnyString()).GetAllCapturedArguments()
			Assert(t, len(c.ExpReplies) == len(actReplies), "missing expected replies, got %d but expected %d", len(actReplies), len(c.ExpReplies))
			for i, expReply := range c.ExpReplies {
				assertCommentEquals(t, expReply, actReplies[i], c.RepoDir, c.ExpParallel)
			}

			if c.ExpAutomerge {
				// Verify that the merge API call was made.
				vcsClient.VerifyWasCalledOnce().MergePull(matchers.AnyModelsPullRequest())
			} else {
				vcsClient.VerifyWasCalled(Never()).MergePull(matchers.AnyModelsPullRequest())
			}
		})
	}
}
