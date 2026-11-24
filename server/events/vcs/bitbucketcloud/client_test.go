// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package bitbucketcloud_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs/bitbucketcloud"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

const (
	diffstatURL     = "/2.0/repositories/owner/repo/pullrequests/1/diffstat"
	mergeabilityURL = "/2.0/repositories/owner/repo/pullrequests/1/mergeability/checks"
	statusesURL     = "/2.0/repositories/owner/repo/pullrequests/1/statuses"
)

// Should follow pagination properly.
func TestClient_GetModifiedFilesPagination(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	respTemplate := `
{
    "pagelen": 1,
    "values": [
        {
            "type": "diffstat",
            "status": "modified",
            "lines_removed": 1,
            "lines_added": 2,
            "old": {
                "path": "%s",
                "type": "commit_file",
                "links": {
                    "self": {
                        "href": "https://api.bitbucket.org/2.0/repositories/bitbucket/geordi/src/e1749643d655d7c7014001a6c0f58abaf42ad850/setup.py"
                    }
                }
            },
            "new": {
                "path": "%s",
                "type": "commit_file",
                "links": {
                    "self": {
                        "href": "https://api.bitbucket.org/2.0/repositories/bitbucket/geordi/src/d222fa235229c55dad20b190b0b571adf737d5a6/setup.py"
                    }
                }
            }
        }
    ],
    "page": 1,
    "size": 1
`
	firstResp := fmt.Sprintf(respTemplate, "file1.txt", "file2.txt")
	secondResp := fmt.Sprintf(respTemplate, "file2.txt", "file3.txt")
	var serverURL string

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		// The first request should hit this URL.
		case diffstatURL:
			resp := firstResp + fmt.Sprintf(`,"next": "%s%s?page=2"}`, serverURL, diffstatURL)
			w.Write([]byte(resp)) // nolint: errcheck
			return
			// The second should hit this URL.
		case fmt.Sprintf("%s?page=2", diffstatURL):
			w.Write([]byte(secondResp + "}")) // nolint: errcheck
		default:
			t.Errorf("got unexpected request at %q", r.RequestURI)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
	}))
	defer testServer.Close()

	serverURL = testServer.URL
	client := bitbucketcloud.New(http.DefaultClient, "user", "pass", "", "runatlantis.io")
	client.BaseURL = testServer.URL

	files, err := client.GetModifiedFiles(
		logger,
		models.Repo{
			FullName:          "owner/repo",
			Owner:             "owner",
			Name:              "repo",
			CloneURL:          "",
			SanitizedCloneURL: "",
			VCSHost: models.VCSHost{
				Type:     models.BitbucketCloud,
				Hostname: "bitbucket.org",
			},
		}, models.PullRequest{
			Num: 1,
		})
	Ok(t, err)
	Equals(t, []string{"file1.txt", "file2.txt", "file3.txt"}, files)
}

// If the "old" key in the list of files is nil we shouldn't error.
func TestClient_GetModifiedFilesOldNil(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	resp := `
{
  "pagelen": 500,
  "values": [
    {
      "status": "added",
      "old": null,
      "lines_removed": 0,
      "lines_added": 2,
      "new": {
        "path": "parent/child/file1.txt",
        "type": "commit_file",
        "links": {
          "self": {
            "href": "https://api.bitbucket.org/2.0/repositories/lkysow/atlantis-example/src/1ed8205eec00dab4f1c0a8c486a4492c98c51f8e/main.tf"
          }
        }
      },
      "type": "diffstat"
    }
  ],
  "page": 1,
  "size": 1
}`

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		// The first request should hit this URL.
		case diffstatURL:
			w.Write([]byte(resp)) // nolint: errcheck
			return
		default:
			t.Errorf("got unexpected request at %q", r.RequestURI)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
	}))
	defer testServer.Close()

	client := bitbucketcloud.New(http.DefaultClient, "user", "pass", "", "runatlantis.io")
	client.BaseURL = testServer.URL

	files, err := client.GetModifiedFiles(
		logger,
		models.Repo{
			FullName:          "owner/repo",
			Owner:             "owner",
			Name:              "repo",
			CloneURL:          "",
			SanitizedCloneURL: "",
			VCSHost: models.VCSHost{
				Type:     models.BitbucketCloud,
				Hostname: "bitbucket.org",
			},
		}, models.PullRequest{
			Num: 1,
		})
	Ok(t, err)
	Equals(t, []string{"parent/child/file1.txt"}, files)
}

func TestClient_PullIsApproved(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	cases := []struct {
		description string
		testdata    string
		exp         bool
	}{
		{
			"no approvers",
			"pull-unapproved.json",
			false,
		},
		{
			"approver is the author",
			"pull-approved-by-author.json",
			false,
		},
		{
			"single approver",
			"pull-approved.json",
			true,
		},
		{
			"two approvers one author",
			"pull-approved-multiple.json",
			true,
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			json, err := os.ReadFile(filepath.Join("testdata", c.testdata))
			Ok(t, err)
			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.RequestURI {
				// The first request should hit this URL.
				case "/2.0/repositories/owner/repo/pullrequests/1":
					w.Write(json) // nolint: errcheck
					return
				default:
					t.Errorf("got unexpected request at %q", r.RequestURI)
					http.Error(w, "not found", http.StatusNotFound)
					return
				}
			}))
			defer testServer.Close()

			client := bitbucketcloud.New(http.DefaultClient, "user", "pass", "", "runatlantis.io")
			client.BaseURL = testServer.URL

			repo, err := models.NewRepo(models.BitbucketServer, "owner/repo", "https://bitbucket.org/owner/repo.git", "user", "token", "")
			Ok(t, err)
			approvalStatus, err := client.PullIsApproved(
				logger,
				repo, models.PullRequest{
					Num:        1,
					HeadBranch: "branch",
					Author:     "author",
					BaseRepo:   repo,
				})
			Ok(t, err)
			Equals(t, c.exp, approvalStatus.IsApproved)
		})
	}
}

func TestBitbucketCloud_GetPullRequestIdentityMapsDestinationBranch(t *testing.T) {
	assertBitbucketCloudPullRequestIdentityIncludesBaseBranch(t)
}

func TestLivePullIdentityFetcher_BitbucketCloudIncludesBaseBranch(t *testing.T) {
	assertBitbucketCloudPullRequestIdentityIncludesBaseBranch(t)
}

func assertBitbucketCloudPullRequestIdentityIncludesBaseBranch(t *testing.T) {
	t.Helper()
	logger := logging.NewNoopLogger(t)
	pullRequest, err := os.ReadFile(filepath.Join("testdata", "pull-approved.json"))
	Ok(t, err)
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case "/2.0/repositories/owner/repo/pullrequests/1":
			w.Write(pullRequest) // nolint: errcheck
			return
		default:
			t.Errorf("got unexpected request at %q", r.RequestURI)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
	}))
	defer testServer.Close()

	client := bitbucketcloud.New(http.DefaultClient, "user", "pass", "", "runatlantis.io")
	client.BaseURL = testServer.URL
	repo := models.Repo{
		FullName: "owner/repo",
		Owner:    "owner",
		Name:     "repo",
		VCSHost: models.VCSHost{
			Type:     models.BitbucketCloud,
			Hostname: "bitbucket.org",
		},
	}
	pull := models.PullRequest{Num: 1, BaseRepo: repo}

	identity, err := client.GetPullRequestIdentity(logger, repo, pull)

	Ok(t, err)
	Equals(t, "75d1e7c57cd9", identity.HeadCommit)
	Equals(t, "main", identity.BaseBranch)
}

func TestClient_PullIsMergeable(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	cases := map[string]struct {
		Checks       string
		Statuses     string
		Status       int
		ExpMergeable models.MergeableStatus
		ExpErr       string
	}{
		"mergeable": {
			Checks: `{"values": [
				{"type": "pullrequest_state_check", "status": "PASSED", "required": true, "blocking": false, "state": "OPEN"},
				{"type": "current_user_permission_check", "status": "PASSED", "required": true, "blocking": false},
				{"type": "git_mergeability_check", "status": "PASSED", "required": true, "blocking": false, "reason": "clean"}
			]}`,
			ExpMergeable: models.MergeableStatus{
				IsMergeable: true,
			},
		},
		"failed but not blocking": {
			Checks: `{"values": [
				{"type": "standard_merge_check", "status": "FAILED", "required": false, "blocking": false, "check": {"kind": "minimum_approvals"}, "requirement": {"minimum_approvals": 2}, "observed": {"approval_count": 1}}
			]}`,
			ExpMergeable: models.MergeableStatus{
				IsMergeable: true,
			},
		},
		"atlantis user not allowed to merge": {
			Checks: `{"values": [
				{"type": "current_user_permission_check", "status": "FAILED", "required": true, "blocking": true}
			]}`,
			ExpMergeable: models.MergeableStatus{
				IsMergeable: true,
			},
		},
		"merge conflict": {
			Checks: `{"values": [
				{"type": "pullrequest_state_check", "status": "PASSED", "required": true, "blocking": false, "state": "OPEN"},
				{"type": "git_mergeability_check", "status": "FAILED", "required": true, "blocking": true, "reason": "conflicts"}
			]}`,
			ExpMergeable: models.MergeableStatus{
				IsMergeable: false,
				Reason:      "Blocked by conflicts",
			},
		},
		"multiple blocking checks": {
			Checks: `{"values": [
				{"type": "current_user_permission_check", "status": "FAILED", "required": true, "blocking": true},
				{"type": "standard_merge_check", "status": "FAILED", "required": true, "blocking": true, "check": {"kind": "minimum_approvals"}, "requirement": {"minimum_approvals": 2}, "observed": {"approval_count": 1}},
				{"type": "custom_pre_merge_check", "status": "FAILED", "required": true, "blocking": true, "check": {"id": "ari", "name": "My check"}}
			]}`,
			ExpMergeable: models.MergeableStatus{
				IsMergeable: false,
				Reason:      "Blocked by minimum_approvals, My check",
			},
		},
		"draft": {
			Checks: `{"values": [
				{"type": "pullrequest_state_check", "status": "FAILED", "required": true, "blocking": true, "state": "DRAFT"}
			]}`,
			ExpMergeable: models.MergeableStatus{
				IsMergeable: false,
				Reason:      "Blocked by DRAFT",
			},
		},
		"builds blocked by apply statuses": {
			Checks: `{"values": [
				{"type": "standard_merge_check", "status": "FAILED", "required": true, "blocking": true, "check": {"kind": "failed_builds"}},
				{"type": "standard_merge_check", "status": "FAILED", "required": true, "blocking": true, "check": {"kind": "in_progress_builds"}}
			]}`,
			Statuses: `{"values": [{"key": "atlantis/plan", "state": "SUCCESSFUL"}, {"key": "atlantis/apply", "state": "INPROGRESS"}, {"key": "atlantis/apply: p1", "state": "FAILED"}]}`,
			ExpMergeable: models.MergeableStatus{
				IsMergeable: true,
			},
		},
		"builds blocked by apply statuses and another check": {
			Checks: `{"values": [
				{"type": "standard_merge_check", "status": "FAILED", "required": true, "blocking": true, "check": {"kind": "failed_builds"}},
				{"type": "standard_merge_check", "status": "FAILED", "required": true, "blocking": true, "check": {"kind": "resolved_tasks"}}
			]}`,
			Statuses: `{"values": [{"key": "atlantis/apply", "state": "FAILED"}]}`,
			ExpMergeable: models.MergeableStatus{
				IsMergeable: false,
				Reason:      "Blocked by resolved_tasks",
			},
		},
		"builds blocked by plan status": {
			Checks: `{"values": [
				{"type": "standard_merge_check", "status": "FAILED", "required": true, "blocking": true, "check": {"kind": "failed_builds"}},
				{"type": "standard_merge_check", "status": "FAILED", "required": true, "blocking": true, "check": {"kind": "in_progress_builds"}}
			]}`,
			Statuses: `{"values": [{"key": "atlantis/plan", "state": "FAILED"}, {"key": "atlantis/apply", "state": "FAILED"}]}`,
			ExpMergeable: models.MergeableStatus{
				IsMergeable: false,
				Reason:      "Blocked by failed_builds, in_progress_builds",
			},
		},
		"builds blocked by another build": {
			Checks: `{"values": [
				{"type": "standard_merge_check", "status": "FAILED", "required": true, "blocking": true, "check": {"kind": "failed_builds"}},
				{"type": "standard_merge_check", "status": "FAILED", "required": true, "blocking": true, "check": {"kind": "in_progress_builds"}}
			]}`,
			Statuses: `{"values": [{"key": "ci", "state": "STOPPED"}, {"key": "atlantis/apply", "state": "FAILED"}]}`,
			ExpMergeable: models.MergeableStatus{
				IsMergeable: false,
				Reason:      "Blocked by failed_builds, in_progress_builds",
			},
		},
		"builds blocked by ignored atlantis": {
			Checks: `{"values": [
				{"type": "standard_merge_check", "status": "FAILED", "required": true, "blocking": true, "check": {"kind": "failed_builds"}},
				{"type": "standard_merge_check", "status": "FAILED", "required": true, "blocking": true, "check": {"kind": "in_progress_builds"}}
			]}`,
			Statuses: `{"values": [{"key": "ignored/plan", "state": "INPROGRESS"}, {"key": "ignored/apply: p1", "state": "FAILED"}, {"key": "atlantis/apply", "state": "FAILED"}]}`,
			ExpMergeable: models.MergeableStatus{
				IsMergeable: true,
			},
		},
		"builds blocked by ignored atlantis and plan status": {
			Checks: `{"values": [
				{"type": "standard_merge_check", "status": "FAILED", "required": true, "blocking": true, "check": {"kind": "failed_builds"}},
				{"type": "standard_merge_check", "status": "FAILED", "required": true, "blocking": true, "check": {"kind": "in_progress_builds"}}
			]}`,
			Statuses: `{"values": [{"key": "ignored/apply", "state": "FAILED"}, {"key": "atlantis/plan", "state": "FAILED"}]}`,
			ExpMergeable: models.MergeableStatus{
				IsMergeable: false,
				Reason:      "Blocked by failed_builds, in_progress_builds",
			},
		},
		"builds blocked by atlantis with similar name": {
			Checks: `{"values": [
				{"type": "standard_merge_check", "status": "FAILED", "required": true, "blocking": true, "check": {"kind": "failed_builds"}},
				{"type": "standard_merge_check", "status": "FAILED", "required": true, "blocking": true, "check": {"kind": "in_progress_builds"}}
			]}`,
			Statuses: `{"values": [{"key": "ignored-not/apply", "state": "FAILED"}]}`,
			ExpMergeable: models.MergeableStatus{
				IsMergeable: false,
				Reason:      "Blocked by failed_builds, in_progress_builds",
			},
		},
		"builds blocked by another atlantis": {
			Checks: `{"values": [
				{"type": "standard_merge_check", "status": "FAILED", "required": true, "blocking": true, "check": {"kind": "failed_builds"}},
				{"type": "standard_merge_check", "status": "FAILED", "required": true, "blocking": true, "check": {"kind": "in_progress_builds"}}
			]}`,
			Statuses: `{"values": [{"key": "other/apply", "state": "FAILED"}]}`,
			ExpMergeable: models.MergeableStatus{
				IsMergeable: false,
				Reason:      "Blocked by failed_builds, in_progress_builds",
			},
		},
		"builds blocked but all statuses successful": {
			Checks: `{"values": [
				{"type": "standard_merge_check", "status": "FAILED", "required": true, "blocking": true, "check": {"kind": "failed_builds"}},
				{"type": "standard_merge_check", "status": "FAILED", "required": true, "blocking": true, "check": {"kind": "in_progress_builds"}}
			]}`,
			Statuses: `{"values": [{"key": "atlantis/apply", "state": "SUCCESSFUL"}]}`,
			ExpMergeable: models.MergeableStatus{
				IsMergeable: false,
				Reason:      "Blocked by failed_builds, in_progress_builds",
			},
		},
		"missing status state": {
			Checks: `{"values": [
				{"type": "standard_merge_check", "status": "FAILED", "required": true, "blocking": true, "check": {"kind": "failed_builds"}},
				{"type": "standard_merge_check", "status": "FAILED", "required": true, "blocking": true, "check": {"kind": "in_progress_builds"}}
			]}`,
			Statuses: `{"values": [{"key": "atlantis/apply"}]}`,
			ExpErr:   "was missing fields",
		},
		"missing values": {
			Checks: `{}`,
			ExpErr: "was missing fields",
		},
		"missing blocking": {
			Checks: `{"values": [{"type": "pullrequest_state_check", "status": "PASSED", "required": true}]}`,
			ExpErr: "was missing fields",
		},
		"http error": {
			Status: http.StatusInternalServerError,
			ExpErr: "unexpected status code: 500",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.RequestURI {
				case mergeabilityURL:
					if c.Status != 0 {
						w.WriteHeader(c.Status)
					}
					w.Write([]byte(c.Checks)) // nolint: errcheck
					return
				case statusesURL:
					w.Write([]byte(c.Statuses)) // nolint: errcheck
					return
				default:
					t.Errorf("got unexpected request at %q", r.RequestURI)
					http.Error(w, "not found", http.StatusNotFound)
					return
				}
			}))
			defer testServer.Close()

			client := bitbucketcloud.New(http.DefaultClient, "user", "pass", "", "runatlantis.io")
			client.BaseURL = testServer.URL

			actMergeable, err := client.PullIsMergeable(
				logger,
				models.Repo{
					FullName:          "owner/repo",
					Owner:             "owner",
					Name:              "repo",
					CloneURL:          "",
					SanitizedCloneURL: "",
					VCSHost: models.VCSHost{
						Type:     models.BitbucketCloud,
						Hostname: "bitbucket.org",
					},
				}, models.PullRequest{
					Num: 1,
				}, "atlantis", []string{"ignored"})
			if c.ExpErr != "" {
				ErrContains(t, c.ExpErr, err)
				return
			}
			Ok(t, err)
			Equals(t, c.ExpMergeable, actMergeable)
		})
	}

}

// Should recognize the apply statuses even if UpdateStatus truncated their keys.
func TestClient_PullIsMergeableTruncatedStatuses(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	cases := []struct {
		nameLength   int
		failed       string
		expMergeable bool
	}{
		{nameLength: 8, failed: "apply", expMergeable: true},
		{nameLength: 8, failed: "plan", expMergeable: false},
		{nameLength: bitbucketcloud.MaxVCSStatusNameLength, failed: "apply", expMergeable: true},
		{nameLength: bitbucketcloud.MaxVCSStatusNameLength, failed: "plan", expMergeable: false},
		{nameLength: bitbucketcloud.MaxVCSStatusNameLength, failed: "policy_check", expMergeable: false},
		{nameLength: bitbucketcloud.MaxVCSStatusNameLength, failed: "pre_workflow_hook", expMergeable: false},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%d characters and failed %s", c.nameLength, c.failed), func(t *testing.T) {
			var statuses []string
			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.RequestURI {
				case "/2.0/repositories/owner/repo/commit/sha/statuses/build":
					body, err := io.ReadAll(r.Body)
					Ok(t, err)
					statuses = append(statuses, string(body))
				case mergeabilityURL:
					w.Write([]byte(`{"values": [{"type": "standard_merge_check", "status": "FAILED", "required": true, "blocking": true, "check": {"kind": "failed_builds"}}]}`)) // nolint: errcheck
				case statusesURL:
					fmt.Fprintf(w, `{"values": [%s]}`, strings.Join(statuses, ",")) // nolint: errcheck
				default:
					t.Errorf("got unexpected request at %q", r.RequestURI)
					http.Error(w, "not found", http.StatusNotFound)
				}
			}))
			defer testServer.Close()

			client := bitbucketcloud.New(http.DefaultClient, "user", "pass", "", "runatlantis.io")
			client.BaseURL = testServer.URL
			repo := models.Repo{FullName: "owner/repo"}
			pull := models.PullRequest{Num: 1, HeadCommit: "sha"}
			name := strings.Repeat("a", c.nameLength)

			for _, suffix := range []string{"", ": a/long/project/directory"} {
				Ok(t, client.UpdateStatus(logger, repo, pull, models.FailedCommitStatus, fmt.Sprintf("%s/%s%s", name, c.failed, suffix), "", ""))
			}

			actMergeable, err := client.PullIsMergeable(logger, repo, pull, name, []string{})
			Ok(t, err)
			Equals(t, c.expMergeable, actMergeable.IsMergeable)
		})
	}
}

// Should follow the pagination of the commit statuses properly, and reject next
// page URLs from a different origin to prevent SSRF.
func TestClient_PullIsMergeableStatusesPagination(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	cases := map[string]struct {
		NextOrigin   string
		ExpMergeable models.MergeableStatus
		ExpErr       string
	}{
		"same origin": {
			ExpMergeable: models.MergeableStatus{
				IsMergeable: false,
				Reason:      "Blocked by failed_builds",
			},
		},
		"different origin": {
			NextOrigin: "http://internal.evil.host",
			ExpErr:     "does not match base URL origin",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			nextOrigin := c.NextOrigin
			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.RequestURI {
				case mergeabilityURL:
					w.Write([]byte(`{"values": [{"type": "standard_merge_check", "status": "FAILED", "required": true, "blocking": true, "check": {"kind": "failed_builds"}}]}`)) // nolint: errcheck
				case statusesURL:
					fmt.Fprintf(w, `{"values": [{"key": "atlantis/apply", "state": "FAILED"}], "next": "%s%s?page=2"}`, nextOrigin, statusesURL) // nolint: errcheck
				case statusesURL + "?page=2":
					w.Write([]byte(`{"values": [{"key": "ci", "state": "FAILED"}]}`)) // nolint: errcheck
				default:
					t.Errorf("got unexpected request at %q", r.RequestURI)
					http.Error(w, "not found", http.StatusNotFound)
				}
			}))
			defer testServer.Close()
			if nextOrigin == "" {
				nextOrigin = testServer.URL
			}

			client := bitbucketcloud.New(http.DefaultClient, "user", "pass", "", "runatlantis.io")
			client.BaseURL = testServer.URL

			actMergeable, err := client.PullIsMergeable(logger, models.Repo{FullName: "owner/repo"}, models.PullRequest{Num: 1}, "atlantis", []string{})
			if c.ExpErr != "" {
				ErrContains(t, c.ExpErr, err)
				return
			}
			Ok(t, err)
			Equals(t, c.ExpMergeable, actMergeable)
		})
	}
}

// TestClient_GetModifiedFiles_SSRFPrevented verifies that a malicious "next"
// pagination URL pointing to a different origin is rejected to prevent SSRF.
func TestClient_GetModifiedFiles_SSRFPrevented(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	resp := `{
		"pagelen": 1,
		"values": [
			{
				"status": "modified",
				"old": {"path": "file1.txt"},
				"new": {"path": "file1.txt"}
			}
		],
		"next": "http://internal.evil.host/steal-secrets"
	}`

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case diffstatURL:
			w.Write([]byte(resp)) // nolint: errcheck
		default:
			t.Errorf("got unexpected request at %q", r.RequestURI)
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer testServer.Close()

	client := bitbucketcloud.New(http.DefaultClient, "user", "pass", "", "runatlantis.io")
	client.BaseURL = testServer.URL

	_, err := client.GetModifiedFiles(logger, models.Repo{
		FullName: "owner/repo",
		VCSHost:  models.VCSHost{Type: models.BitbucketCloud, Hostname: "bitbucket.org"},
	}, models.PullRequest{Num: 1})
	Assert(t, err != nil, "expected error for malicious next page URL")
	Assert(t, strings.Contains(err.Error(), "does not match base URL origin"), "error should mention origin mismatch, got: %s", err.Error())
}

func TestClient_MarkdownPullLink(t *testing.T) {
	client := bitbucketcloud.New(http.DefaultClient, "user", "pass", "", "runatlantis.io")
	pull := models.PullRequest{Num: 1}
	s, _ := client.MarkdownPullLink(pull)
	exp := "#1"
	Equals(t, exp, s)
}

func TestClient_GetMyUUID(t *testing.T) {
	json, err := os.ReadFile(filepath.Join("testdata", "user.json"))
	Ok(t, err)

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case "/2.0/user":
			w.Write(json) // nolint: errcheck
			return
		default:
			t.Errorf("got unexpected request at %q", r.RequestURI)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
	}))
	defer testServer.Close()

	client := bitbucketcloud.New(http.DefaultClient, "user", "pass", "", "runatlantis.io")
	client.BaseURL = testServer.URL
	v, _ := client.GetMyUUID()
	Equals(t, v, "{00000000-0000-0000-0000-000000000001}")
}

func TestClient_HidePRCommentsRequiresUserRead(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Equals(t, "/2.0/user", r.URL.Path)
		http.Error(w, "missing account scope", http.StatusForbidden)
	}))
	defer server.Close()
	client := bitbucketcloud.New(server.Client(), "user", "pass", "", "")
	client.BaseURL = server.URL
	err := client.HidePrevCommandComments(logging.NewNoopLogger(t), models.Repo{}, 1, "plan", "")
	Assert(t, err != nil, "comment hiding must still require user-read permission")
	Assert(t, strings.Contains(err.Error(), "check required scope"), "expected scope error, got %v", err)
}

func TestClient_GetComment(t *testing.T) {
	json, err := os.ReadFile(filepath.Join("testdata", "comments.json"))
	Ok(t, err)

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case "/2.0/repositories/myorg/myrepo/pullrequests/5/comments":
			w.Write(json) // nolint: errcheck
			return
		default:
			t.Errorf("got unexpected request at %q", r.RequestURI)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
	}))
	defer testServer.Close()

	client := bitbucketcloud.New(http.DefaultClient, "user", "pass", "", "runatlantis.io")
	client.BaseURL = testServer.URL
	v, _ := client.GetPullRequestComments(
		models.Repo{
			FullName:          "myorg/myrepo",
			Owner:             "owner",
			Name:              "myrepo",
			CloneURL:          "",
			SanitizedCloneURL: "",
			VCSHost: models.VCSHost{
				Type:     models.BitbucketCloud,
				Hostname: "bitbucket.org",
			},
		}, 5)

	Equals(t, len(v), 5)
	exp := "Plan"
	Assert(t, strings.Contains(v[1].Content.Raw, exp), "Comment should contain word \"%s\", has \"%s\"", exp, v[1].Content.Raw)
}

func TestClient_DeleteComment(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case "/2.0/repositories/myorg/myrepo/pullrequests/5/comments/1":
			if r.Method == "DELETE" {
				w.WriteHeader(http.StatusNoContent)
			}
			return
		default:
			t.Errorf("got unexpected request at %q", r.RequestURI)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
	}))
	defer testServer.Close()

	client := bitbucketcloud.New(http.DefaultClient, "user", "pass", "", "runatlantis.io")
	client.BaseURL = testServer.URL
	err := client.DeletePullRequestComment(
		models.Repo{
			FullName:          "myorg/myrepo",
			Owner:             "owner",
			Name:              "myrepo",
			CloneURL:          "",
			SanitizedCloneURL: "",
			VCSHost: models.VCSHost{
				Type:     models.BitbucketCloud,
				Hostname: "bitbucket.org",
			},
		}, 5, 1)
	Ok(t, err)
}

func TestClient_HidePRComments(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	comments, err := os.ReadFile(filepath.Join("testdata", "comments.json"))
	Ok(t, err)
	json, err := os.ReadFile(filepath.Join("testdata", "user.json"))
	Ok(t, err)

	called := 0

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		// we have two comments in the test file
		// The code is going to delete them all and then create a new one
		case "/2.0/repositories/myorg/myrepo/pullrequests/5/comments/498931882":
			if r.Method == "DELETE" {
				w.WriteHeader(http.StatusNoContent)
			}
			w.Write([]byte("")) // nolint: errcheck
			called += 1
			return
			// This is the second one
		case "/2.0/repositories/myorg/myrepo/pullrequests/5/comments/498931784":
			if r.Method == "DELETE" {
				http.Error(w, "", http.StatusNoContent)
			}
			w.Write([]byte("")) // nolint: errcheck
			called += 1
			return
		case "/2.0/repositories/myorg/myrepo/pullrequests/5/comments/49893111":
			Assert(t, r.Method != "DELETE", "Shouldn't delete this one")
			return
		case "/2.0/repositories/myorg/myrepo/pullrequests/5/comments":
			w.Write(comments) // nolint: errcheck
			return
		case "/2.0/user":
			w.Write(json) // nolint: errcheck
			return
		default:
			t.Errorf("got unexpected request at %q", r.RequestURI)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
	}))
	defer testServer.Close()

	client := bitbucketcloud.New(http.DefaultClient, "user", "pass", "", "runatlantis.io")
	client.BaseURL = testServer.URL
	err = client.HidePrevCommandComments(logger,
		models.Repo{
			FullName:          "myorg/myrepo",
			Owner:             "owner",
			Name:              "myrepo",
			CloneURL:          "",
			SanitizedCloneURL: "",
			VCSHost: models.VCSHost{
				Type:     models.BitbucketCloud,
				Hostname: "bitbucket.org",
			},
		}, 5, "plan", "")
	Ok(t, err)
	Equals(t, 2, called)
}
