// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/drmaxgit/go-azuredevops/azuredevops"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/events/vcs/bitbucketcloud"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

// Embedding unimplemented interfaces makes unexpected calls panic, including
// reactions, extra comments, and commands after a self-comment.
type identityClient struct {
	vcs.Client
	self bool
	repo models.Repo
	user models.User
}

func (c *identityClient) IsAtlantisUser(repo models.Repo, user models.User) (bool, error) {
	c.repo, c.user = repo, user
	return c.self, nil
}

type countingParser struct {
	calls  int
	parser events.CommentParsing
}

func (p *countingParser) Parse(comment string, host models.VCSHostType) events.CommentParseResult {
	p.calls++
	if p.parser != nil {
		return p.parser.Parse(comment, host)
	}
	return events.CommentParseResult{Command: &events.CommentCommand{Name: command.Plan}}
}

type countingRunner struct {
	events.CommandRunner
	calls int
	user  models.User
}

func (r *countingRunner) RunCommentCommand(_ models.Repo, _ *models.Repo, _ *models.PullRequest, user models.User, _ int, _ *events.CommentCommand) {
	r.calls++
	r.user = user
}

func TestHandleCommentEvent_IdentityLookup(t *testing.T) {
	for _, test := range []struct {
		name, author, comment               string
		status                              int
		body                                string
		wantLookups, wantRuns, wantComments int
	}{
		{"no user-read permission", "human", "atlantis plan", 403, `{}`, 0, 1, 0},
		{"human", "human", "atlantis plan", 200, `{"uuid":"self"}`, 0, 1, 0},
		{"other bot", "other-bot", "atlantis plan", 200, `{"uuid":"self"}`, 0, 1, 0},
		{"self plan", "self", "atlantis plan", 200, `{"uuid":"self"}`, 0, 1, 0},
		{"self apply", "self", "atlantis apply", 200, `{"uuid":"self"}`, 0, 1, 0},
		{"self closed PR", "self", "Atlantis commands can't be run on closed pull requests", 200, `{"uuid":"self"}`, 1, 0, 0},
		{"self fork PR", "self", "Atlantis commands can't be run on fork pull requests", 200, `{"uuid":"self"}`, 1, 0, 0},
		{"self help", "self", "atlantis help", 200, `{"uuid":"self"}`, 1, 0, 0},
		{"self flag help", "self", "atlantis --help", 200, `{"uuid":"self"}`, 1, 0, 0},
		{"self dash", "self", "atlantis -", 200, `{"uuid":"self"}`, 1, 0, 0},
		{"self suggestion", "self", "terraform plan", 200, `{"uuid":"self"}`, 1, 0, 0},
		{"self invalid flags", "self", "atlantis plan --unknown", 200, `{"uuid":"self"}`, 1, 0, 0},
		{"human help", "human", "atlantis help", 200, `{"uuid":"self"}`, 1, 0, 1},
		{"human parsing error", "human", "atlantis plan '", 200, `{"uuid":"self"}`, 1, 0, 1},
		{"ignored comment", "human", "LGTM", 500, `{}`, 0, 0, 0},
		{"ignored self output", "self", "Atlantis plan results\n1 to add", 500, `{}`, 0, 0, 0},
		{"invalid credentials", "human", "atlantis help", 401, `{}`, 1, 0, 1},
		{"no identity permission", "human", "atlantis help", 403, `{}`, 1, 0, 1},
		{"rate limited", "human", "atlantis help", 429, `{}`, 1, 0, 1},
		{"API unavailable", "human", "atlantis help", 500, `{}`, 1, 0, 1},
		{"network error", "human", "atlantis help", 0, `{}`, 1, 0, 1},
		{"malformed response", "human", "atlantis help", 200, `{`, 1, 0, 1},
		{"missing UUID", "human", "atlantis help", 200, `{}`, 1, 0, 1},
		{"self parsing error after lookup failure", "self", "Atlantis commands can't be run on closed pull requests", 500, `{}`, 1, 0, 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			comments, lookups := 0, 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/2.0/user":
					lookups++
					if test.status == 0 {
						conn, _, err := w.(http.Hijacker).Hijack()
						Ok(t, err)
						_ = conn.Close()
						return
					}
					w.WriteHeader(test.status)
					_, _ = w.Write([]byte(test.body))
				case "/2.0/repositories/owner/repo/pullrequests/1/comments":
					comments++
					w.WriteHeader(http.StatusCreated)
				default:
					t.Errorf("unexpected request %s", r.URL)
				}
			}))
			defer server.Close()
			client := bitbucketcloud.New(server.Client(), "git-user", "token", "api@example.com", "")
			client.BaseURL = server.URL
			allowlist, err := events.NewRepoAllowlistChecker("*")
			Ok(t, err)
			parser := &countingParser{parser: &events.CommentParser{ExecutableName: "atlantis", AllowCommands: command.AllCommentCommands}}
			runner := &countingRunner{}
			controller := VCSEventsController{
				VCSClient: client, CommentParser: parser, CommandRunner: runner,
				RepoAllowlistChecker: allowlist, TestingMode: true, EmojiReaction: "eyes",
			}
			repo := models.Repo{FullName: "owner/repo", VCSHost: models.VCSHost{Type: models.BitbucketCloud}}
			user := models.User{ID: test.author, Username: test.author}
			response := controller.handleCommentEvent(logging.NewNoopLogger(t), repo, nil, nil, user, user, 1, test.comment, 2, models.BitbucketCloud)
			Equals(t, 0, response.err.code)
			Equals(t, 1, parser.calls)
			Equals(t, test.wantLookups, lookups)
			Equals(t, test.wantRuns, runner.calls)
			Equals(t, test.wantComments, comments)
		})
	}
}

func TestCommentWebhooks_SelfAuthor(t *testing.T) {
	readFixture := func(t *testing.T, name string) map[string]any {
		t.Helper()
		data, err := os.ReadFile(name)
		Ok(t, err)
		var fixture map[string]any
		Ok(t, json.Unmarshal(data, &fixture))
		return fixture
	}
	for _, host := range []models.VCSHostType{models.BitbucketCloud, models.AzureDevops} {
		t.Run(host.String(), func(t *testing.T) {
			client := &identityClient{self: true}
			body := "Atlantis commands can't be run on closed pull requests"
			parser := &countingParser{parser: &events.CommentParser{ExecutableName: "atlantis"}}
			logger := logging.NewNoopLogger(t)
			controller := VCSEventsController{
				VCSClient: client, CommentParser: parser, Logger: logger,
				Parser:        &events.EventParser{BitbucketServerURL: "https://bitbucket.example.com"},
				EmojiReaction: "eyes", TestingMode: true,
			}
			w := httptest.NewRecorder()
			expectedUser := ""
			switch host {
			case models.BitbucketCloud:
				fixture := readFixture(t, "../../events/testdata/bitbucket-cloud-comment-event.json")
				fixture["comment"].(map[string]any)["content"].(map[string]any)["raw"] = body
				data, err := json.Marshal(fixture)
				Ok(t, err)
				controller.HandleBitbucketCloudCommentEvent(w, data, "test")
				expectedUser = "557058:dc3817de-68b5-45cd-b81c-5c39d2560090"
				Equals(t, "{bf34a99b-8a11-452c-8fbc-bdffc340e584}", client.user.ID)
			case models.AzureDevops:
				controller.HandleAzureDevopsPullRequestCommentedEvent(w, azureCommentEvent(t, body), "test")
				expectedUser = "atlantis@example.com"
				Equals(t, "comment-author-id", client.user.ID)
			}
			ResponseContains(t, w, http.StatusOK, "Ignoring comment authored by Atlantis")
			Equals(t, expectedUser, client.user.Username)
			Equals(t, host, client.repo.VCSHost.Type)
			Equals(t, 1, parser.calls)
		})
	}
}

func TestAzureCommentPreservesCommandUser(t *testing.T) {
	for _, self := range []bool{false, true} {
		t.Run(fmt.Sprint(self), func(t *testing.T) {
			client := &identityClient{self: self}
			runner := &countingRunner{}
			allowlist, err := events.NewRepoAllowlistChecker("*")
			Ok(t, err)
			controller := VCSEventsController{
				VCSClient: client, CommandRunner: runner, CommentParser: &countingParser{},
				Parser: &events.EventParser{}, Logger: logging.NewNoopLogger(t),
				RepoAllowlistChecker: allowlist, TestingMode: true,
			}
			w := httptest.NewRecorder()
			controller.HandleAzureDevopsPullRequestCommentedEvent(w, azureCommentEvent(t, "atlantis plan"), "test")
			ResponseContains(t, w, http.StatusOK, "Processing...")
			Equals(t, models.User{}, client.user)
			Equals(t, 1, runner.calls)
			Equals(t, models.User{Username: "human@example.com"}, runner.user)
		})
	}
}

func azureCommentEvent(t *testing.T, body string) *azuredevops.Event {
	t.Helper()
	data := fmt.Sprintf(`{"eventType":"ms.vss-code.git-pullrequest-comment-event","resource":{"comment":{"content":%q,"author":{"id":"comment-author-id","uniqueName":"atlantis@example.com"}},"pullRequest":{"pullRequestId":1,"createdBy":{"id":"pr-author-id","uniqueName":"human@example.com"},"repository":{"name":"repo","project":{"name":"project"},"webUrl":"https://dev.azure.com/owner/project/_git/repo"}}}}`, body)
	event, err := azuredevops.ParseWebHook([]byte(data))
	Ok(t, err)
	return event
}
