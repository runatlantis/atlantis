// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package gitea_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs/gitea"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestClient_GetModifiedFilesPagination(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	pages := make([]int, 0, 2)
	filesPage1 := mustReadTestData(t, "list-pull-request-files-page-1.json")
	filesPage2 := mustReadTestData(t, "list-pull-request-files-page-2.json")

	client := newTestClient(t, 2, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/repos/owner/repo/pulls/1/files" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.String())
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		page, err := strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil {
			t.Errorf("invalid page query: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		pages = append(pages, page)

		if gotLimit := r.URL.Query().Get("limit"); gotLimit != "2" {
			t.Errorf("expected limit=2, got %q", gotLimit)
		}

		switch page {
		case 1:
			w.Header().Set("Link", `</repos/owner/repo/pulls/1/files?page=2&limit=2>; rel="next"`)
			_, _ = w.Write(filesPage1)
		case 2:
			_, _ = w.Write(filesPage2)
		default:
			t.Errorf("unexpected page %d", page)
			http.Error(w, "not found", http.StatusNotFound)
		}
	})

	files, err := client.GetModifiedFiles(logger, models.Repo{Owner: "owner", Name: "repo"}, models.PullRequest{Num: 1})
	Ok(t, err)
	Equals(t, []string{"modules/network/main.tf", "modules/app/variables.tf", "README.md"}, files)
	Equals(t, []int{1, 2}, pages)
}

func TestClient_PullIsApprovedPagination(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	pages := make([]int, 0, 2)
	reviewsPage1Comment := mustReadTestData(t, "list-pull-reviews-page-1-comment.json")
	reviewsPage2Approved := mustReadTestData(t, "list-pull-reviews-page-2-approved.json")

	client := newTestClient(t, 2, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/repos/owner/repo/pulls/1/reviews" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.String())
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		page, err := strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil {
			t.Errorf("invalid page query: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		pages = append(pages, page)

		if gotLimit := r.URL.Query().Get("limit"); gotLimit != "2" {
			t.Errorf("expected limit=2, got %q", gotLimit)
		}

		switch page {
		case 1:
			w.Header().Set("Link", `</repos/owner/repo/pulls/1/reviews?page=2&limit=2>; rel="next"`)
			_, _ = w.Write(reviewsPage1Comment)
		case 2:
			_, _ = w.Write(reviewsPage2Approved)
		default:
			t.Errorf("unexpected page %d", page)
			http.Error(w, "not found", http.StatusNotFound)
		}
	})

	approvalStatus, err := client.PullIsApproved(logger, models.Repo{Owner: "owner", Name: "repo"}, models.PullRequest{Num: 1})
	Ok(t, err)
	Assert(t, approvalStatus.IsApproved, "expected pull request to be approved")
	Equals(t, "reviewer-2", approvalStatus.ApprovedBy)

	submittedAt, err := time.Parse(time.RFC3339, "2026-01-11T12:34:56Z")
	Ok(t, err)
	Equals(t, submittedAt, approvalStatus.Date)
	Equals(t, []int{1, 2}, pages)
}

// giteaReview is a compact description of a Gitea pull review used to build
// API responses for approval tests.
type giteaReview struct {
	id        int
	login     string
	state     string
	dismissed bool
	stale     bool
	submitted string
}

func giteaReviewsJSON(reviews []giteaReview) string {
	entries := make([]string, 0, len(reviews))
	for _, r := range reviews {
		entries = append(entries, fmt.Sprintf(`{
  "id": %d,
  "user": {"id": %d, "login": %q, "login_name": %q},
  "state": %q,
  "commit_id": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
  "stale": %t,
  "official": true,
  "dismissed": %t,
  "comments_count": 0,
  "submitted_at": %q
}`, r.id, 100+r.id, r.login, r.login, r.state, r.stale, r.dismissed, r.submitted))
	}
	return "[" + strings.Join(entries, ",") + "]"
}

func TestClient_PullIsApproved_EffectiveReviewState(t *testing.T) {
	cases := []struct {
		description   string
		reviews       []giteaReview
		expApproved   bool
		expApprovedBy string
	}{
		{
			description: "dismissed approval only is not approved",
			reviews: []giteaReview{
				{id: 1, login: "alice", state: "APPROVED", dismissed: true, submitted: "2026-01-10T01:00:00Z"},
			},
			expApproved: false,
		},
		{
			description: "approval superseded by request changes from the same user is not approved",
			reviews: []giteaReview{
				{id: 1, login: "alice", state: "APPROVED", submitted: "2026-01-10T01:00:00Z"},
				{id: 2, login: "alice", state: "REQUEST_CHANGES", submitted: "2026-01-10T02:00:00Z"},
			},
			expApproved: false,
		},
		{
			description: "stale approval is not approved",
			reviews: []giteaReview{
				{id: 1, login: "alice", state: "APPROVED", stale: true, submitted: "2026-01-10T01:00:00Z"},
			},
			expApproved: false,
		},
		{
			description: "approval followed by a dismissed approval from the same user is not approved",
			reviews: []giteaReview{
				{id: 1, login: "alice", state: "APPROVED", submitted: "2026-01-10T01:00:00Z"},
				{id: 2, login: "alice", state: "APPROVED", dismissed: true, submitted: "2026-01-10T02:00:00Z"},
			},
			expApproved: false,
		},
		{
			description: "approval followed by a stale approval from the same user is not approved",
			reviews: []giteaReview{
				{id: 1, login: "alice", state: "APPROVED", submitted: "2026-01-10T01:00:00Z"},
				{id: 2, login: "alice", state: "APPROVED", stale: true, submitted: "2026-01-10T02:00:00Z"},
			},
			expApproved: false,
		},
		{
			description: "approval followed by a dismissed request changes from the same user is not approved",
			reviews: []giteaReview{
				{id: 1, login: "alice", state: "APPROVED", submitted: "2026-01-10T01:00:00Z"},
				{id: 2, login: "alice", state: "REQUEST_CHANGES", dismissed: true, submitted: "2026-01-10T02:00:00Z"},
			},
			expApproved: false,
		},
		{
			description: "dismissed approval followed by a fresh approval from the same user is approved",
			reviews: []giteaReview{
				{id: 1, login: "alice", state: "APPROVED", dismissed: true, submitted: "2026-01-10T01:00:00Z"},
				{id: 2, login: "alice", state: "APPROVED", submitted: "2026-01-10T02:00:00Z"},
			},
			expApproved:   true,
			expApprovedBy: "alice",
		},
		{
			description: "approval by one user plus dismissed approval by another is approved",
			reviews: []giteaReview{
				{id: 1, login: "alice", state: "APPROVED", submitted: "2026-01-10T01:00:00Z"},
				{id: 2, login: "bob", state: "APPROVED", dismissed: true, submitted: "2026-01-10T02:00:00Z"},
			},
			expApproved:   true,
			expApprovedBy: "alice",
		},
		{
			description: "dismissed comment after approval does not revoke the approval",
			reviews: []giteaReview{
				{id: 1, login: "alice", state: "APPROVED", submitted: "2026-01-10T01:00:00Z"},
				{id: 2, login: "alice", state: "COMMENT", dismissed: true, submitted: "2026-01-10T02:00:00Z"},
			},
			expApproved:   true,
			expApprovedBy: "alice",
		},
		{
			description: "approval by one user plus request changes by another is approved",
			reviews: []giteaReview{
				{id: 1, login: "alice", state: "APPROVED", submitted: "2026-01-10T01:00:00Z"},
				{id: 2, login: "bob", state: "REQUEST_CHANGES", submitted: "2026-01-10T02:00:00Z"},
			},
			expApproved:   true,
			expApprovedBy: "alice",
		},
		{
			description: "request changes followed by approval from the same user is approved",
			reviews: []giteaReview{
				{id: 1, login: "alice", state: "REQUEST_CHANGES", submitted: "2026-01-10T01:00:00Z"},
				{id: 2, login: "alice", state: "APPROVED", submitted: "2026-01-10T02:00:00Z"},
			},
			expApproved:   true,
			expApprovedBy: "alice",
		},
		{
			description: "comment after approval does not revoke the approval",
			reviews: []giteaReview{
				{id: 1, login: "alice", state: "APPROVED", submitted: "2026-01-10T01:00:00Z"},
				{id: 2, login: "alice", state: "COMMENT", submitted: "2026-01-10T02:00:00Z"},
			},
			expApproved:   true,
			expApprovedBy: "alice",
		},
		{
			description: "normal approval is approved",
			reviews: []giteaReview{
				{id: 1, login: "alice", state: "APPROVED", submitted: "2026-01-10T01:00:00Z"},
			},
			expApproved:   true,
			expApprovedBy: "alice",
		},
		{
			description: "no reviews is not approved",
			reviews:     nil,
			expApproved: false,
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			logger := logging.NewNoopLogger(t)
			body := giteaReviewsJSON(c.reviews)
			client := newTestClient(t, 100, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != "/repos/owner/repo/pulls/1/reviews" {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.String())
					http.Error(w, "not found", http.StatusNotFound)
					return
				}
				_, _ = w.Write([]byte(body))
			})

			approvalStatus, err := client.PullIsApproved(logger, models.Repo{Owner: "owner", Name: "repo"}, models.PullRequest{Num: 1})
			Ok(t, err)
			Equals(t, c.expApproved, approvalStatus.IsApproved)
			Equals(t, c.expApprovedBy, approvalStatus.ApprovedBy)
		})
	}
}

func TestClient_PullIsApproved_SupersededAcrossPages(t *testing.T) {
	// The approval is on page 1 and the same reviewer's later REQUEST_CHANGES
	// is on page 2. The client must read every page before deciding.
	logger := logging.NewNoopLogger(t)
	page1 := giteaReviewsJSON([]giteaReview{
		{id: 1, login: "alice", state: "APPROVED", submitted: "2026-01-10T01:00:00Z"},
	})
	page2 := giteaReviewsJSON([]giteaReview{
		{id: 2, login: "alice", state: "REQUEST_CHANGES", submitted: "2026-01-10T02:00:00Z"},
	})

	client := newTestClient(t, 1, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/repos/owner/repo/pulls/1/reviews" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.String())
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		switch r.URL.Query().Get("page") {
		case "1":
			w.Header().Set("Link", `</repos/owner/repo/pulls/1/reviews?page=2&limit=1>; rel="next"`)
			_, _ = w.Write([]byte(page1))
		case "2":
			_, _ = w.Write([]byte(page2))
		default:
			t.Errorf("unexpected page %q", r.URL.Query().Get("page"))
			http.Error(w, "not found", http.StatusNotFound)
		}
	})

	approvalStatus, err := client.PullIsApproved(logger, models.Repo{Owner: "owner", Name: "repo"}, models.PullRequest{Num: 1})
	Ok(t, err)
	Equals(t, false, approvalStatus.IsApproved)
}

func TestClient_DiscardReviewsPagination(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	listPages := make([]int, 0, 2)
	dismissedReviewIDs := make([]int, 0, 4)
	reviewsPage1ID11 := mustReadTestData(t, "list-pull-reviews-page-1.json")
	reviewsPage2ID22 := mustReadTestData(t, "list-pull-reviews-page-2.json")

	client := newTestClient(t, 2, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/repos/owner/repo/pulls/1/reviews":
			page, err := strconv.Atoi(r.URL.Query().Get("page"))
			if err != nil {
				t.Errorf("invalid page query: %v", err)
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			listPages = append(listPages, page)

			switch page {
			case 1:
				w.Header().Set("Link", `</repos/owner/repo/pulls/1/reviews?page=2&limit=2>; rel="next"`)
				_, _ = w.Write(reviewsPage1ID11)
			case 2:
				_, _ = w.Write(reviewsPage2ID22)
			default:
				t.Errorf("unexpected page %d", page)
				http.Error(w, "not found", http.StatusNotFound)
			}
		case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/repos/owner/repo/pulls/1/reviews/") && strings.HasSuffix(r.URL.Path, "/dismissals"):
			reviewID, err := parseDismissReviewID(r.URL.Path)
			if err != nil {
				t.Errorf("failed to parse dismiss review id from path %q: %v", r.URL.Path, err)
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			dismissedReviewIDs = append(dismissedReviewIDs, reviewID)
			w.WriteHeader(http.StatusOK)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.String())
			http.Error(w, "not found", http.StatusNotFound)
		}
	})

	err := client.DiscardReviews(logger, models.Repo{Owner: "owner", Name: "repo"}, models.PullRequest{Num: 1})
	Ok(t, err)
	Equals(t, []int{1, 2}, listPages)
	Equals(t, []int{11, 12, 22, 23}, dismissedReviewIDs)
}

func TestClient_GetPullLabelsPagination(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	pages := make([]int, 0, 2)
	labelsPage1 := mustReadTestData(t, "list-issue-labels-page-1.json")
	labelsPage2 := mustReadTestData(t, "list-issue-labels-page-2.json")

	client := newTestClient(t, 2, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/repos/owner/repo/issues/1/labels" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.String())
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		page, err := strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil {
			t.Errorf("invalid page query: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		pages = append(pages, page)

		if gotLimit := r.URL.Query().Get("limit"); gotLimit != "2" {
			t.Errorf("expected limit=2, got %q", gotLimit)
		}

		switch page {
		case 1:
			w.Header().Set("Link", `</repos/owner/repo/issues/1/labels?page=2&limit=2>; rel="next"`)
			_, _ = w.Write(labelsPage1)
		case 2:
			_, _ = w.Write(labelsPage2)
		default:
			t.Errorf("unexpected page %d", page)
			http.Error(w, "not found", http.StatusNotFound)
		}
	})

	labels, err := client.GetPullLabels(logger, models.Repo{Owner: "owner", Name: "repo"}, models.PullRequest{Num: 1})
	Ok(t, err)
	Equals(t, []string{"label-one", "label-alpha", "label-two"}, labels)
	Equals(t, []int{1, 2}, pages)
}

func TestClient_NilResponseErrorsDoNotPanic(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	repo := models.Repo{Owner: "owner", Name: "repo"}
	pull := models.PullRequest{
		Num:        1,
		HeadCommit: "abc123",
		BaseRepo:   repo,
	}

	cases := []struct {
		description string
		run         func(client *gitea.Client) error
	}{
		{
			description: "GetPullRequest",
			run: func(client *gitea.Client) error {
				_, err := client.GetPullRequest(logger, repo, pull.Num)
				return err
			},
		},
		{
			description: "GetModifiedFiles",
			run: func(client *gitea.Client) error {
				_, err := client.GetModifiedFiles(logger, repo, pull)
				return err
			},
		},
		{
			description: "CreateComment",
			run: func(client *gitea.Client) error {
				return client.CreateComment(logger, repo, pull.Num, "comment", "plan")
			},
		},
		{
			description: "ReactToComment",
			run: func(client *gitea.Client) error {
				return client.ReactToComment(logger, repo, pull.Num, 1, "+1")
			},
		},
		{
			description: "HidePrevCommandComments",
			run: func(client *gitea.Client) error {
				return client.HidePrevCommandComments(logger, repo, pull.Num, "plan", "")
			},
		},
		{
			description: "PullIsApproved",
			run: func(client *gitea.Client) error {
				_, err := client.PullIsApproved(logger, repo, pull)
				return err
			},
		},
		{
			description: "UpdateStatus",
			run: func(client *gitea.Client) error {
				return client.UpdateStatus(logger, repo, pull, models.PendingCommitStatus, "atlantis/plan", "pending", "")
			},
		},
		{
			description: "MergePull",
			run: func(client *gitea.Client) error {
				return client.MergePull(logger, pull, models.PullRequestOptions{})
			},
		},
		{
			description: "GetFileContent",
			run: func(client *gitea.Client) error {
				_, _, err := client.GetFileContent(logger, repo, "main", "atlantis.yaml")
				return err
			},
		},
		{
			description: "GetPullLabels",
			run: func(client *gitea.Client) error {
				_, err := client.GetPullLabels(logger, repo, pull)
				return err
			},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			client := newClosedTestClient(t, 2)

			var err error
			func() {
				defer func() {
					if recovered := recover(); recovered != nil {
						t.Fatalf("unexpected panic: %v", recovered)
					}
				}()
				err = c.run(client)
			}()

			Assert(t, err != nil, "expected network error")
		})
	}
}

func TestClient_GetModifiedFilesPaginationEmergencyBreak(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	pages := make([]int, 0, 500)

	client := newTestClient(t, 1, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/repos/owner/repo/pulls/1/files" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.String())
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		page, err := strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil {
			t.Errorf("invalid page query: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		pages = append(pages, page)

		if page > 500 {
			t.Errorf("expected pagination emergency break at page 500, got page %d", page)
			http.Error(w, "too many pages", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Link", fmt.Sprintf(`</repos/owner/repo/pulls/1/files?page=%d&limit=1>; rel="next"`, page+1))
		_, _ = w.Write([]byte(`[]`))
	})

	files, err := client.GetModifiedFiles(logger, models.Repo{Owner: "owner", Name: "repo"}, models.PullRequest{Num: 1})
	Ok(t, err)
	Equals(t, []string{}, files)
	Equals(t, 500, len(pages))
	Equals(t, 500, pages[len(pages)-1])
}

func newTestClient(t *testing.T, pageSize int, apiHandler http.HandlerFunc) *gitea.Client {
	t.Helper()
	versionResponse := mustReadTestData(t, "version-response.json")

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if r.Method == http.MethodGet && path == "/api/v1/version" {
			_, _ = w.Write(versionResponse)
			return
		}

		if strings.HasPrefix(path, "/api/v1") {
			rewritten := r.Clone(r.Context())
			urlCopy := *r.URL
			urlCopy.Path = strings.TrimPrefix(path, "/api/v1")
			rewritten.URL = &urlCopy
			apiHandler(w, rewritten)
			return
		}

		apiHandler(w, r)
	}))
	t.Cleanup(testServer.Close)

	client, err := gitea.New(testServer.URL, "user", "token", pageSize, logging.NewNoopLogger(t))
	Ok(t, err)
	return client
}

func newClosedTestClient(t *testing.T, pageSize int) *gitea.Client {
	t.Helper()
	versionResponse := mustReadTestData(t, "version-response.json")

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/v1/version" {
			_, _ = w.Write(versionResponse)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))

	client, err := gitea.New(testServer.URL, "user", "token", pageSize, logging.NewNoopLogger(t))
	Ok(t, err)
	testServer.Close()
	return client
}

func mustReadTestData(t *testing.T, filename string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", filename))
	Ok(t, err)
	return data
}

func parseDismissReviewID(path string) (int, error) {
	prefix := "/repos/owner/repo/pulls/1/reviews/"
	suffix := "/dismissals"
	reviewIDStr := strings.TrimSuffix(strings.TrimPrefix(path, prefix), suffix)
	if reviewIDStr == path || reviewIDStr == "" {
		return 0, fmt.Errorf("unexpected dismiss path: %s", path)
	}

	reviewID, err := strconv.Atoi(reviewIDStr)
	if err != nil {
		return 0, err
	}
	return reviewID, nil
}

func TestClient_PullIsApproved_TombstoneAcrossPages(t *testing.T) {
	// The approval is on page 1 and the same reviewer's later dismissed
	// review is on page 2. The dismissed review revokes the reviewer's stance,
	// so the older approval must not be revived.
	logger := logging.NewNoopLogger(t)
	page1 := giteaReviewsJSON([]giteaReview{
		{id: 1, login: "alice", state: "APPROVED", submitted: "2026-01-10T01:00:00Z"},
	})
	page2 := giteaReviewsJSON([]giteaReview{
		{id: 2, login: "alice", state: "APPROVED", dismissed: true, submitted: "2026-01-10T02:00:00Z"},
	})

	client := newTestClient(t, 1, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/repos/owner/repo/pulls/1/reviews" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.String())
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		switch r.URL.Query().Get("page") {
		case "1":
			w.Header().Set("Link", `</repos/owner/repo/pulls/1/reviews?page=2&limit=1>; rel="next"`)
			_, _ = w.Write([]byte(page1))
		case "2":
			_, _ = w.Write([]byte(page2))
		default:
			t.Errorf("unexpected page %q", r.URL.Query().Get("page"))
			http.Error(w, "not found", http.StatusNotFound)
		}
	})

	approvalStatus, err := client.PullIsApproved(logger, models.Repo{Owner: "owner", Name: "repo"}, models.PullRequest{Num: 1})
	Ok(t, err)
	Equals(t, false, approvalStatus.IsApproved)
}
