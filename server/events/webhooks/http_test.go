// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package webhooks_test

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/webhooks"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

var httpApplyResult = webhooks.ApplyResult{
	Workspace: "production",
	Repo: models.Repo{
		FullName: "runatlantis/atlantis",
	},
	Pull: models.PullRequest{
		Num:        1,
		URL:        "url",
		BaseBranch: "main",
	},
	User: models.User{
		Username: "lkysow",
	},
	ProjectName: "test-project",
	Directory:   "testing/prod/directory",
	Success:     true,
}

func TestHttpWebhookWithHeaders(t *testing.T) {
	expectedHeaders := map[string][]string{
		"Authorization":   {"Bearer token"},
		"X-Custom-Header": {"value1", "value2"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Equals(t, r.Header.Get("Content-Type"), "application/json")
		for k, v := range expectedHeaders {
			Equals(t, r.Header.Values(k), v)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	webhook := webhooks.HttpWebhook{
		Client:         &webhooks.HttpClient{Client: http.DefaultClient, Headers: expectedHeaders},
		URL:            server.URL,
		WorkspaceRegex: regexp.MustCompile(".*"),
		BranchRegex:    regexp.MustCompile(".*"),
		ProjectRegex:   regexp.MustCompile(".*"),
		DirectoryRegex: regexp.MustCompile(".*"),
	}

	err := webhook.Send(logging.NewNoopLogger(t), httpApplyResult)
	Ok(t, err)
}

func TestHttpWebhookNoHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Equals(t, r.Header.Get("Content-Type"), "application/json")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	webhook := webhooks.HttpWebhook{
		Client:         &webhooks.HttpClient{Client: http.DefaultClient},
		URL:            server.URL,
		WorkspaceRegex: regexp.MustCompile(".*"),
		BranchRegex:    regexp.MustCompile(".*"),
		ProjectRegex:   regexp.MustCompile(".*"),
		DirectoryRegex: regexp.MustCompile(".*"),
	}

	err := webhook.Send(logging.NewNoopLogger(t), httpApplyResult)
	Ok(t, err)
}

func TestHttpWebhook500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	webhook := webhooks.HttpWebhook{
		Client:         &webhooks.HttpClient{Client: http.DefaultClient},
		URL:            server.URL,
		WorkspaceRegex: regexp.MustCompile(".*"),
		BranchRegex:    regexp.MustCompile(".*"),
		ProjectRegex:   regexp.MustCompile(".*"),
		DirectoryRegex: regexp.MustCompile(".*"),
	}

	err := webhook.Send(logging.NewNoopLogger(t), httpApplyResult)
	ErrContains(t, "sending webhook", err)
}

func TestHttpNoRegexMatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Assert(t, false, "webhook should not be sent")
	}))
	defer server.Close()

	tt := []struct {
		name string
		wr   *regexp.Regexp
		br   *regexp.Regexp
		pr   *regexp.Regexp
		dr   *regexp.Regexp
	}{
		{
			name: "no workspace match",
			wr:   regexp.MustCompile("other"),
			br:   regexp.MustCompile(".*"),
			pr:   regexp.MustCompile(".*"),
			dr:   regexp.MustCompile(".*"),
		},
		{
			name: "no branch match",
			wr:   regexp.MustCompile(".*"),
			br:   regexp.MustCompile("other"),
			pr:   regexp.MustCompile(".*"),
			dr:   regexp.MustCompile(".*"),
		},
		{
			name: "no project match",
			wr:   regexp.MustCompile(".*"),
			br:   regexp.MustCompile(".*"),
			pr:   regexp.MustCompile("other"),
			dr:   regexp.MustCompile(".*"),
		},
		{
			name: "no directory match",
			wr:   regexp.MustCompile(".*"),
			br:   regexp.MustCompile(".*"),
			pr:   regexp.MustCompile(".*"),
			dr:   regexp.MustCompile("other"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			webhook := webhooks.HttpWebhook{
				Client:         &webhooks.HttpClient{Client: http.DefaultClient},
				URL:            server.URL,
				WorkspaceRegex: tc.wr,
				BranchRegex:    tc.br,
				ProjectRegex:   tc.pr,
				DirectoryRegex: tc.dr,
			}
			err := webhook.Send(logging.NewNoopLogger(t), httpApplyResult)
			Ok(t, err)
		})
	}
}

func TestHttpMultipleRegexMatch(t *testing.T) {
	tt := []struct {
		name         string
		wr           *regexp.Regexp
		br           *regexp.Regexp
		pr           *regexp.Regexp
		dr           *regexp.Regexp
		shouldBeSent bool
	}{
		{
			name:         "all regexes match",
			wr:           regexp.MustCompile("production"),
			br:           regexp.MustCompile("main"),
			pr:           regexp.MustCompile("test-project"),
			dr:           regexp.MustCompile("testing/prod/directory"),
			shouldBeSent: true,
		},
		{
			name:         "no match - workspace regex wrong, others correct",
			wr:           regexp.MustCompile("notproduction"),
			br:           regexp.MustCompile("main"),
			pr:           regexp.MustCompile("test-project"),
			dr:           regexp.MustCompile("testing/directory"),
			shouldBeSent: false,
		},
		{
			name:         "no match - directory regex wrong, others correct or allowed",
			wr:           regexp.MustCompile(".*"),
			br:           regexp.MustCompile(".*"),
			pr:           regexp.MustCompile("test-project"),
			dr:           regexp.MustCompile("wrong/directory"),
			shouldBeSent: false,
		},
		{
			name:         "match - directory regex allowing, others correct or allowed",
			wr:           regexp.MustCompile(".*"),
			br:           regexp.MustCompile(".*"),
			pr:           regexp.MustCompile(".*"),
			dr:           regexp.MustCompile("testing/.*/directory"),
			shouldBeSent: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new server for each test case
			webhookCalled := false
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				webhookCalled = true
			}))
			defer server.Close()

			webhook := webhooks.HttpWebhook{
				Client:         &webhooks.HttpClient{Client: http.DefaultClient},
				URL:            server.URL,
				WorkspaceRegex: tc.wr,
				BranchRegex:    tc.br,
				ProjectRegex:   tc.pr,
				DirectoryRegex: tc.dr,
			}
			err := webhook.Send(logging.NewNoopLogger(t), httpApplyResult)
			Ok(t, err)

			Assert(t, tc.shouldBeSent == webhookCalled,
				"webhook expected to be called: %v, was: %v",
				tc.shouldBeSent,
				webhookCalled)
		})
	}
}
