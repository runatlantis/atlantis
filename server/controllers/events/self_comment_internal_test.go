// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"encoding/json"
	"errors"
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
	err  error
	repo models.Repo
	user models.User
}

func (c *identityClient) IsAtlantisUser(repo models.Repo, user models.User) (bool, error) {
	c.repo, c.user = repo, user
	return c.self, c.err
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
		name, author, comment              string
		status                             int
		body                               string
		wantParses, wantRuns, wantComments int
	}{
		{"no user-read permission", "human", "atlantis plan", 403, `{}`, 1, 1, 0},
		{"human", "human", "atlantis plan", 200, `{"uuid":"self"}`, 1, 1, 0},
		{"other bot", "other-bot", "atlantis plan", 200, `{"uuid":"self"}`, 1, 1, 0},
		{"self command", "self", "atlantis plan", 200, `{"uuid":"self"}`, 0, 0, 0},
		{"self closed PR", "self", "Atlantis commands can't be run on closed pull requests", 200, `{"uuid":"self"}`, 0, 0, 0},
		{"self fork PR", "self", "Atlantis commands can't be run on fork pull requests", 200, `{"uuid":"self"}`, 0, 0, 0},
		{"invalid credentials", "human", "atlantis plan", 401, `{}`, 1, 1, 0},
		{"rate limited", "human", "atlantis plan", 429, `{}`, 1, 1, 0},
		{"API unavailable", "human", "atlantis plan", 500, `{}`, 1, 1, 0},
		{"network error", "human", "atlantis plan", 0, `{}`, 1, 1, 0},
		{"malformed response", "human", "atlantis plan", 200, `{`, 1, 1, 0},
		{"missing UUID", "human", "atlantis plan", 200, `{}`, 1, 1, 0},
		{"self parsing error after lookup failure", "self", "Atlantis commands can't be run on closed pull requests", 500, `{}`, 1, 0, 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			comments := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/2.0/user":
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
			parser := &countingParser{parser: &events.CommentParser{ExecutableName: "atlantis", AllowCommands: []command.Name{command.Plan}}}
			runner := &countingRunner{}
			controller := VCSEventsController{
				VCSClient: client, CommentParser: parser, CommandRunner: runner,
				RepoAllowlistChecker: allowlist, TestingMode: true, EmojiReaction: "eyes",
			}
			repo := models.Repo{FullName: "owner/repo", VCSHost: models.VCSHost{Type: models.BitbucketCloud}}
			user := models.User{ID: test.author, Username: test.author}
			commentAuthor := user
			response := controller.handleCommentEvent(logging.NewNoopLogger(t), repo, nil, nil, user, commentAuthor, 1, test.comment, 2, models.BitbucketCloud)
			Equals(t, 0, response.err.code)
			Equals(t, test.wantParses, parser.calls)
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
			parser := &countingParser{}
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
			Equals(t, 0, parser.calls)
		})
	}
}

func TestAzureCommentPreservesCommandUser(t *testing.T) {
	for _, lookupErr := range []error{nil, errors.New("identity lookup unavailable")} {
		t.Run(fmt.Sprint(lookupErr), func(t *testing.T) {
			client := &identityClient{err: lookupErr}
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
			Equals(t, models.User{ID: "comment-author-id", Username: "atlantis@example.com"}, client.user)
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
