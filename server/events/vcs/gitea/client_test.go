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
