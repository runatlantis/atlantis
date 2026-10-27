// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package github_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs/github"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

var stackedPull = models.PullRequest{
	Num:        2,
	HeadCommit: "abc123",
	BaseRepo: models.Repo{
		FullName: "owner/repo",
		Owner:    "owner",
		Name:     "repo",
		VCSHost: models.VCSHost{
			Type:     models.Github,
			Hostname: "github.com",
		},
	},
	Stack: &models.PullRequestStack{Number: 1, Position: 2, Size: 3, BaseBranch: "main"},
}

// Test that stacked pull requests are merged through the asynchronous merge
// endpoint, and that its result is polled until the merge completes.
func TestClient_MergePullStacked(t *testing.T) {
	cases := map[string]struct {
		// statuses are the responses to the merge request followed by each poll.
		statuses []string
		timeout  time.Duration
		expErr   string
	}{
		"merged immediately": {
			statuses: []string{"merged"},
		},
		"merged after polling": {
			statuses: []string{"pending", "pending", "merged"},
		},
		"added to merge queue": {
			statuses: []string{"pending", "enqueued"},
		},
		"failed": {
			statuses: []string{"pending", "failed"},
			expErr:   "could not merge stacked pull request: failed: status failed",
		},
		"timed out": {
			statuses: []string{"pending"},
			timeout:  20 * time.Millisecond,
			expErr:   "timed out after 20ms waiting for GitHub to merge stacked pull request (merge request uuid-1 is still pending)",
		},
	}

	jsBytes, err := os.ReadFile("testdata/repo.json")
	Ok(t, err)

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			requests := 0
			respond := func(w http.ResponseWriter, code int) {
				status := c.statuses[min(requests, len(c.statuses)-1)]
				requests++
				w.WriteHeader(code)
				fmt.Fprintf(w, `{"status":%q,"details":{"uuid":"uuid-1","message":"status %s"}}`, status, status) // nolint: errcheck
			}
			testServer := httptest.NewTLSServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.RequestURI {
					case "/api/v3/repos/owner/repo":
						w.Write(jsBytes) // nolint: errcheck
					case "/api/v3/repos/owner/repo/pulls/2/merge-async":
						Equals(t, http.MethodPut, r.Method)
						body, err := io.ReadAll(r.Body)
						Ok(t, err)
						Equals(t, `{"merge_method":"squash","sha":"abc123"}`+"\n", string(body))
						respond(w, http.StatusAccepted)
					case "/api/v3/repos/owner/repo/pulls/2/merge-async/uuid-1":
						Equals(t, http.MethodGet, r.Method)
						code := http.StatusOK
						if c.statuses[min(requests, len(c.statuses)-1)] == "pending" {
							code = http.StatusAccepted
						}
						respond(w, code)
					default:
						t.Errorf("got unexpected request at %q", r.RequestURI)
						http.Error(w, "not found", http.StatusNotFound)
					}
				}))

			testServerURL, err := url.Parse(testServer.URL)
			Ok(t, err)
			client, err := github.New(testServerURL.Host, &github.UserCredentials{"user", "pass", ""}, github.Config{
				AsyncMergePollInterval: time.Millisecond,
				AsyncMergeTimeout:      c.timeout,
			}, 0, logging.NewNoopLogger(t))
			Ok(t, err)
			defer disableSSLVerification()()

			err = client.MergePull(logging.NewNoopLogger(t), stackedPull, models.PullRequestOptions{MergeMethod: "squash"})
			if c.expErr == "" {
				Ok(t, err)
				Equals(t, len(c.statuses), requests)
			} else {
				ErrEquals(t, c.expErr, err)
			}
		})
	}
}

func TestClient_GetPullRequestStack(t *testing.T) {
	stacks := []map[string]any{{
		"id":     1,
		"number": 7,
		"base":   map[string]any{"ref": "main"},
		"open":   true,
		"pull_requests": []map[string]any{
			{"number": 1, "state": "closed", "draft": false, "merged_at": "2026-10-01T00:00:00Z", "head": map[string]any{"ref": "a", "sha": "sha1"}},
			{"number": 2, "state": "open", "draft": false, "merged_at": nil, "head": map[string]any{"ref": "b", "sha": "sha2"}},
			{"number": 3, "state": "open", "draft": true, "merged_at": nil, "head": map[string]any{"ref": "c", "sha": "sha3"}},
		},
	}}

	cases := map[string]struct {
		response any
		exp      []models.StackedPull
	}{
		"stacked": {
			response: stacks,
			exp: []models.StackedPull{
				{Num: 1, HeadCommit: "sha1", Merged: true},
				{Num: 2, HeadCommit: "sha2", Open: true},
				{Num: 3, HeadCommit: "sha3", Open: true, Draft: true},
			},
		},
		"not stacked": {
			response: []any{},
			exp:      nil,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			testServer := httptest.NewTLSServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.RequestURI {
					case "/api/v3/repos/owner/repo/stacks?pull_request=2":
						Ok(t, json.NewEncoder(w).Encode(c.response))
					default:
						t.Errorf("got unexpected request at %q", r.RequestURI)
						http.Error(w, "not found", http.StatusNotFound)
					}
				}))

			testServerURL, err := url.Parse(testServer.URL)
			Ok(t, err)
			client, err := github.New(testServerURL.Host, &github.UserCredentials{"user", "pass", ""}, github.Config{}, 0, logging.NewNoopLogger(t))
			Ok(t, err)
			defer disableSSLVerification()()

			pulls, err := client.GetPullRequestStack(logging.NewNoopLogger(t), stackedPull.BaseRepo, 2)
			Ok(t, err)
			Equals(t, c.exp, pulls)
		})
	}
}
