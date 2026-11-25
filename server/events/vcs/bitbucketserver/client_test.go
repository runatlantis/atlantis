// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package bitbucketserver_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs/bitbucketserver"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

// Test that we include the base path in our base url.
func TestClient_BasePath(t *testing.T) {
	cases := []struct {
		inputURL string
		expURL   string
		expErr   string
	}{
		{
			inputURL: "mycompany.com",
			expErr:   `must have 'http://' or 'https://' in base url "mycompany.com"`,
		},
		{
			inputURL: "https://mycompany.com",
			expURL:   "https://mycompany.com",
		},
		{
			inputURL: "http://mycompany.com",
			expURL:   "http://mycompany.com",
		},
		{
			inputURL: "http://mycompany.com:7990",
			expURL:   "http://mycompany.com:7990",
		},
		{
			inputURL: "http://mycompany.com/",
			expURL:   "http://mycompany.com",
		},
		{
			inputURL: "http://mycompany.com:7990/",
			expURL:   "http://mycompany.com:7990",
		},
		{
			inputURL: "http://mycompany.com/basepath/",
			expURL:   "http://mycompany.com/basepath",
		},
		{
			inputURL: "http://mycompany.com:7990/basepath/",
			expURL:   "http://mycompany.com:7990/basepath",
		},
	}

	for _, c := range cases {
		t.Run(c.inputURL, func(t *testing.T) {
			client, err := bitbucketserver.NewClient(nil, "u", "p", c.inputURL, "atlantis-url")
			if c.expErr != "" {
				ErrEquals(t, c.expErr, err)
			} else {
				Ok(t, err)
				Equals(t, c.expURL, client.BaseURL)
			}
		})
	}
}

// Should follow pagination properly.
func TestClient_GetModifiedFilesPagination(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	respTemplate := `
{
  "values": [
    {
      "path": {
        "toString": "%s"
      }
    },
    {
      "path": {
        "toString": "%s"
      }
    }
  ],
  "size": 2,
  "isLastPage": true,
  "start": 0,
  "limit": 2,
  "nextPageStart": null
}
`
	firstResp := fmt.Sprintf(respTemplate, "file1.txt", "file2.txt")
	secondResp := fmt.Sprintf(respTemplate, "file2.txt", "file3.txt")
	var serverURL string

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		// The first request should hit this URL.
		case "/rest/api/1.0/projects/ow/repos/repo/pull-requests/1/changes?start=0":
			resp := strings.ReplaceAll(firstResp, `"isLastPage": true`, `"isLastPage": false`)
			resp = strings.ReplaceAll(resp, `"nextPageStart": null`, `"nextPageStart": 3`)
			w.Write([]byte(resp)) // nolint: errcheck
			return
			// The second should hit this URL.
		case "/rest/api/1.0/projects/ow/repos/repo/pull-requests/1/changes?start=3":
			w.Write([]byte(secondResp)) // nolint: errcheck
		default:
			t.Errorf("got unexpected request at %q", r.RequestURI)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
	}))
	defer testServer.Close()

	serverURL = testServer.URL
	client, err := bitbucketserver.NewClient(http.DefaultClient, "user", "pass", serverURL, "runatlantis.io")
	Ok(t, err)

	files, err := client.GetModifiedFiles(
		logger,
		models.Repo{
			FullName:          "owner/repo",
			Owner:             "owner",
			Name:              "repo",
			SanitizedCloneURL: fmt.Sprintf("%s/scm/ow/repo.git", serverURL),
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

func TestClient_PullIsMergeable(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	cases := map[string]struct {
		MergeStatus  string
		ExpMergeable models.MergeableStatus
	}{
		"mergeable": {
			MergeStatus: `{"canMerge": true, "conflicted": false, "outcome": "CLEAN", "vetoes": []}`,
			ExpMergeable: models.MergeableStatus{
				IsMergeable: true,
			},
		},
		"not mergeable without vetoes": {
			MergeStatus: `{"canMerge": false, "conflicted": false}`,
			ExpMergeable: models.MergeableStatus{
				IsMergeable: false,
			},
		},
		"merge conflict": {
			MergeStatus: `{"canMerge": false, "conflicted": true, "outcome": "CONFLICTED", "vetoes": []}`,
			ExpMergeable: models.MergeableStatus{
				IsMergeable: false,
				Reason:      "Has conflicts",
			},
		},
		"merge conflict and vetoes": {
			MergeStatus: `{"canMerge": false, "conflicted": true, "outcome": "CONFLICTED", "vetoes": [
				{"summaryMessage": "First veto", "detailedMessage": "First details"},
				{"summaryMessage": "Second veto", "detailedMessage": "Second details"}
			]}`,
			ExpMergeable: models.MergeableStatus{
				IsMergeable: false,
				Reason:      "Has conflicts; First veto; Second veto",
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.RequestURI {
				case "/rest/api/1.0/projects/ow/repos/repo/pull-requests/1/merge":
					w.Write([]byte(c.MergeStatus)) // nolint: errcheck
					return
				default:
					t.Errorf("got unexpected request at %q", r.RequestURI)
					http.Error(w, "not found", http.StatusNotFound)
					return
				}
			}))
			defer testServer.Close()

			client, err := bitbucketserver.NewClient(http.DefaultClient, "user", "pass", testServer.URL, "runatlantis.io")
			Ok(t, err)

			actMergeable, err := client.PullIsMergeable(
				logger,
				models.Repo{
					FullName:          "owner/repo",
					Owner:             "owner",
					Name:              "repo",
					SanitizedCloneURL: fmt.Sprintf("%s/scm/ow/repo.git", testServer.URL),
				}, models.PullRequest{
					Num: 1,
				}, "atlantis-test", []string{})
			Ok(t, err)
			Equals(t, c.ExpMergeable, actMergeable)
		})
	}
}

// Test that the builds vetoes are ignored when they are only caused by our own
// apply statuses, or by the ignored Atlantis services.
func TestClient_PullIsMergeableBuilds(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	const failedVeto = `{"summaryMessage": "Builds", "detailedMessage": "You cannot merge this pull request while it has failed builds."}`
	const inProgressVeto = `{"summaryMessage": "Builds", "detailedMessage": "You cannot merge this pull request while it has in-progress builds."}`
	const missingVeto = `{"summaryMessage": "Builds", "detailedMessage": "You cannot merge this pull request while it has failed builds. You still need a minimum of 1 successful builds before this pull request can be merged."}`
	const otherVeto = `{"summaryMessage": "Other", "detailedMessage": "Other details"}`
	blocked := models.MergeableStatus{Reason: "Builds"}
	mergeable := models.MergeableStatus{IsMergeable: true}

	cases := map[string]struct {
		vetoes     string
		conflicted bool
		required   int
		pages      []string
		exp        models.MergeableStatus
	}{
		"own apply failed": {
			vetoes: failedVeto, required: 1,
			pages: []string{`{"key": "atlantis/plan", "state": "SUCCESSFUL"}, {"key": "atlantis/apply", "state": "FAILED"}`},
			exp:   mergeable,
		},
		"own project apply failed": {
			vetoes: failedVeto,
			pages:  []string{`{"key": "atlantis/apply: p1", "state": "FAILED"}`},
			exp:    mergeable,
		},
		"own apply in progress": {
			vetoes: inProgressVeto,
			pages:  []string{`{"key": "atlantis/apply", "state": "INPROGRESS"}`},
			exp:    mergeable,
		},
		"ignored services": {
			vetoes: failedVeto,
			pages:  []string{`{"key": "other/plan", "state": "FAILED"}, {"key": "another/apply", "state": "INPROGRESS"}`},
			exp:    mergeable,
		},
		"not an ignored service": {
			vetoes: failedVeto,
			pages:  []string{`{"key": "others/apply", "state": "FAILED"}`},
			exp:    blocked,
		},
		"own plan failed": {
			vetoes: failedVeto,
			pages:  []string{`{"key": "atlantis/plan", "state": "FAILED"}, {"key": "atlantis/apply", "state": "FAILED"}`},
			exp:    blocked,
		},
		"other build failed": {
			vetoes: failedVeto,
			pages:  []string{`{"key": "atlantis/apply", "state": "FAILED"}, {"key": "ci", "state": "FAILED"}`},
			exp:    blocked,
		},
		"other build cancelled": {
			vetoes: failedVeto,
			pages:  []string{`{"key": "atlantis/apply", "state": "FAILED"}, {"key": "ci", "state": "CANCELLED"}`},
			exp:    blocked,
		},
		"other build failed in the second page": {
			vetoes: failedVeto,
			pages:  []string{`{"key": "atlantis/apply", "state": "FAILED"}`, `{"key": "ci", "state": "FAILED"}`},
			exp:    blocked,
		},
		"successful build in the second page": {
			vetoes: failedVeto, required: 1,
			pages: []string{`{"key": "atlantis/apply", "state": "FAILED"}`, `{"key": "ci", "state": "SUCCESSFUL"}`},
			exp:   mergeable,
		},
		"successful build missing": {
			vetoes: failedVeto, required: 2,
			pages: []string{`{"key": "atlantis/plan", "state": "SUCCESSFUL"}, {"key": "atlantis/apply", "state": "FAILED"}`},
			exp:   blocked,
		},
		"own applies don't count as successful builds": {
			vetoes: failedVeto, required: 2,
			pages: []string{`{"key": "atlantis/plan", "state": "SUCCESSFUL"}, {"key": "atlantis/apply: p1", "state": "SUCCESSFUL"}, {"key": "atlantis/apply", "state": "FAILED"}`},
			exp:   blocked,
		},
		"no unsuccessful builds": {
			vetoes: failedVeto,
			pages:  []string{`{"key": "ci", "state": "SUCCESSFUL"}`},
			exp:    blocked,
		},
		"veto saying that builds are missing": {
			vetoes: missingVeto,
			exp:    blocked,
		},
		"other veto too": {
			vetoes: failedVeto + "," + otherVeto,
			pages:  []string{`{"key": "atlantis/apply", "state": "FAILED"}`},
			exp:    models.MergeableStatus{Reason: "Other"},
		},
		"conflicts too": {
			vetoes: failedVeto, conflicted: true,
			pages: []string{`{"key": "atlantis/apply", "state": "FAILED"}`},
			exp:   models.MergeableStatus{Reason: "Has conflicts"},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Equals(t, "en", r.Header.Get("Accept-Language"))
				for i, page := range c.pages {
					if r.RequestURI == fmt.Sprintf("/rest/build-status/1.0/commits/abc?start=%d", i) {
						fmt.Fprintf(w, `{"values": [%s], "isLastPage": %t, "nextPageStart": %d}`, page, i == len(c.pages)-1, i+1) // nolint: errcheck
						return
					}
				}
				switch r.RequestURI {
				case "/rest/api/1.0/projects/ow/repos/repo/pull-requests/1/merge":
					fmt.Fprintf(w, `{"canMerge": false, "conflicted": %t, "vetoes": [%s]}`, c.conflicted, c.vetoes) // nolint: errcheck
				case "/rest/api/1.0/projects/ow/repos/repo/settings/pull-requests":
					fmt.Fprintf(w, `{"requiredSuccessfulBuilds": %d}`, c.required) // nolint: errcheck
				default:
					t.Errorf("got unexpected request at %q", r.RequestURI)
					http.Error(w, "not found", http.StatusNotFound)
				}
			}))
			defer testServer.Close()

			client, err := bitbucketserver.NewClient(http.DefaultClient, "user", "pass", testServer.URL, "runatlantis.io")
			Ok(t, err)

			act, err := client.PullIsMergeable(
				logger,
				models.Repo{
					FullName:          "owner/repo",
					Owner:             "owner",
					Name:              "repo",
					SanitizedCloneURL: fmt.Sprintf("%s/scm/ow/repo.git", testServer.URL),
				}, models.PullRequest{
					Num:        1,
					HeadCommit: "abc",
				}, "atlantis", []string{"other", "another"})
			Ok(t, err)
			Equals(t, c.exp, act)
		})
	}
}

// Test that we don't consider it mergeable if we can't tell which builds failed.
func TestClient_PullIsMergeableBuildsErrors(t *testing.T) {
	const settings = `{"requiredSuccessfulBuilds": 1}`
	cases := map[string]struct {
		settings string
		statuses string
		expErr   string
	}{
		"settings not found":   {expErr: "unexpected status code: 404"},
		"missing minimum":      {settings: `{}`, expErr: "was missing fields"},
		"invalid settings":     {settings: `[`, expErr: "parsing response"},
		"statuses not found":   {settings: settings, expErr: "unexpected status code: 404"},
		"missing status state": {settings: settings, statuses: `{"values": [{"key": "a"}], "isLastPage": true}`, expErr: "was missing fields"},
		"missing next page":    {settings: settings, statuses: `{"values": [], "isLastPage": false}`, expErr: "was missing nextPageStart"},
		"invalid statuses":     {settings: settings, statuses: `[`, expErr: "parsing response"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body := ""
				switch r.RequestURI {
				case "/rest/api/1.0/projects/ow/repos/repo/pull-requests/1/merge":
					body = `{"canMerge": false, "conflicted": false, "vetoes": [{"detailedMessage": "You cannot merge this pull request while it has failed builds."}]}`
				case "/rest/api/1.0/projects/ow/repos/repo/settings/pull-requests":
					body = c.settings
				case "/rest/build-status/1.0/commits/abc?start=0":
					body = c.statuses
				}
				if body == "" {
					http.Error(w, "not found", http.StatusNotFound)
					return
				}
				w.Write([]byte(body)) // nolint: errcheck
			}))
			defer testServer.Close()

			client, err := bitbucketserver.NewClient(http.DefaultClient, "user", "pass", testServer.URL, "runatlantis.io")
			Ok(t, err)
			act, err := client.PullIsMergeable(logging.NewNoopLogger(t), models.Repo{
				Name:              "repo",
				SanitizedCloneURL: fmt.Sprintf("%s/scm/ow/repo.git", testServer.URL),
			}, models.PullRequest{Num: 1, HeadCommit: "abc"}, "atlantis", nil)
			ErrContains(t, c.expErr, err)
			Equals(t, false, act.IsMergeable)
		})
	}
}

// Test that we use the correct version parameter in our call to merge the pull
// request.
func TestClient_MergePull(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	pullRequest, err := os.ReadFile(filepath.Join("testdata", "pull-request.json"))
	Ok(t, err)
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		// The first request should hit this URL.
		case "/rest/api/1.0/projects/ow/repos/repo/pull-requests/1":
			w.Write(pullRequest) // nolint: errcheck
			return
		case "/rest/api/1.0/projects/ow/repos/repo/pull-requests/1/merge?version=3":
			Equals(t, "POST", r.Method)
			w.Write(pullRequest) // nolint: errcheck
		default:
			t.Errorf("got unexpected request at %q", r.RequestURI)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
	}))
	defer testServer.Close()

	client, err := bitbucketserver.NewClient(http.DefaultClient, "user", "pass", testServer.URL, "runatlantis.io")
	Ok(t, err)

	err = client.MergePull(
		logger,
		models.PullRequest{
			Num:        1,
			HeadCommit: "",
			URL:        "",
			HeadBranch: "",
			BaseBranch: "",
			Author:     "",
			State:      0,
			BaseRepo: models.Repo{
				FullName:          "owner/repo",
				Owner:             "owner",
				Name:              "repo",
				SanitizedCloneURL: fmt.Sprintf("%s/scm/ow/repo.git", testServer.URL),
				VCSHost: models.VCSHost{
					Type:     models.BitbucketCloud,
					Hostname: "bitbucket.org",
				},
			},
		}, models.PullRequestOptions{
			DeleteSourceBranchOnMerge: false,
		})
	Ok(t, err)
}

// Test that we delete the source branch in our call to merge the pull
// request.
func TestClient_MergePullDeleteSourceBranch(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	pullRequest, err := os.ReadFile(filepath.Join("testdata", "pull-request.json"))
	Ok(t, err)
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		// The first request should hit this URL.
		case "/rest/api/1.0/projects/ow/repos/repo/pull-requests/1":
			w.Write(pullRequest) // nolint: errcheck
			return
		case "/rest/api/1.0/projects/ow/repos/repo/pull-requests/1/merge?version=3":
			Equals(t, "POST", r.Method)
			w.Write(pullRequest) // nolint: errcheck
		case "/rest/branch-utils/1.0/projects/ow/repos/repo/branches":
			Equals(t, "DELETE", r.Method)
			defer r.Body.Close()
			b, err := io.ReadAll(r.Body)
			Ok(t, err)
			var payload bitbucketserver.DeleteSourceBranch
			err = json.Unmarshal(b, &payload)
			Ok(t, err)
			Equals(t, "refs/heads/foo", payload.Name)
			w.WriteHeader(http.StatusNoContent) // nolint: errcheck
		default:
			t.Errorf("got unexpected request at %q", r.RequestURI)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
	}))
	defer testServer.Close()

	client, err := bitbucketserver.NewClient(http.DefaultClient, "user", "pass", testServer.URL, "runatlantis.io")
	Ok(t, err)

	err = client.MergePull(
		logger,
		models.PullRequest{
			Num:        1,
			HeadCommit: "",
			URL:        "",
			HeadBranch: "foo",
			BaseBranch: "",
			Author:     "",
			State:      0,
			BaseRepo: models.Repo{
				FullName:          "owner/repo",
				Owner:             "owner",
				Name:              "repo",
				SanitizedCloneURL: fmt.Sprintf("%s/scm/ow/repo.git", testServer.URL),
				VCSHost: models.VCSHost{
					Type:     models.BitbucketServer,
					Hostname: "bitbucket.org",
				},
			},
		},
		models.PullRequestOptions{
			DeleteSourceBranchOnMerge: true,
		},
	)
	Ok(t, err)
}

func TestClient_MarkdownPullLink(t *testing.T) {
	client, err := bitbucketserver.NewClient(nil, "u", "p", "https://base-url", "atlantis-url")
	Ok(t, err)
	pull := models.PullRequest{Num: 1}
	s, _ := client.MarkdownPullLink(pull)
	exp := "#1"
	Equals(t, exp, s)
}

func TestBitbucketServer_GetPullRequestIdentityMapsDestinationBranch(t *testing.T) {
	assertBitbucketServerPullRequestIdentityIncludesBaseBranch(t)
}

func TestLivePullIdentityFetcher_BitbucketServerIncludesBaseBranch(t *testing.T) {
	assertBitbucketServerPullRequestIdentityIncludesBaseBranch(t)
}

func assertBitbucketServerPullRequestIdentityIncludesBaseBranch(t *testing.T) {
	t.Helper()
	logger := logging.NewNoopLogger(t)
	pullRequest, err := os.ReadFile(filepath.Join("testdata", "pull-request.json"))
	Ok(t, err)
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case "/rest/api/1.0/projects/ow/repos/repo/pull-requests/1":
			w.Write(pullRequest) // nolint: errcheck
			return
		default:
			t.Errorf("got unexpected request at %q", r.RequestURI)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
	}))
	defer testServer.Close()

	client, err := bitbucketserver.NewClient(http.DefaultClient, "user", "pass", testServer.URL, "runatlantis.io")
	Ok(t, err)
	repo := models.Repo{
		FullName:          "owner/repo",
		Owner:             "owner",
		Name:              "repo",
		SanitizedCloneURL: fmt.Sprintf("%s/scm/ow/repo.git", testServer.URL),
		VCSHost: models.VCSHost{
			Type:     models.BitbucketServer,
			Hostname: "bitbucket.org",
		},
	}
	pull := models.PullRequest{Num: 1, BaseRepo: repo}

	identity, err := client.GetPullRequestIdentity(logger, repo, pull)

	Ok(t, err)
	Equals(t, "bdcaa224f4b65edb853a689404ef79cf47d8cdda", identity.HeadCommit)
	Equals(t, "main", identity.BaseBranch)
}
