package vcs

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	version "github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs/types"
	"github.com/runatlantis/atlantis/server/logging"
	gitlab "github.com/xanzy/go-gitlab"

	. "github.com/runatlantis/atlantis/testing"
)

// Test that the base url gets set properly.
func TestNewGitlabClient_BaseURL(t *testing.T) {
	gitlabClientUnderTest = true
	defer func() { gitlabClientUnderTest = false }()
	cases := []struct {
		Hostname   string
		ExpBaseURL string
	}{
		{
			"gitlab.com",
			"https://gitlab.com/api/v4/",
		},
		{
			"custom.domain",
			"https://custom.domain/api/v4/",
		},
		{
			"http://custom.domain",
			"http://custom.domain/api/v4/",
		},
		{
			"http://custom.domain:8080",
			"http://custom.domain:8080/api/v4/",
		},
		{
			"https://custom.domain",
			"https://custom.domain/api/v4/",
		},
		{
			"https://custom.domain/",
			"https://custom.domain/api/v4/",
		},
		{
			"https://custom.domain/basepath/",
			"https://custom.domain/basepath/api/v4/",
		},
	}

	for _, c := range cases {
		t.Run(c.Hostname, func(t *testing.T) {
			log := logging.NewNoopCtxLogger(t)
			client, err := NewGitlabClient(c.Hostname, "token", log)
			Ok(t, err)
			Equals(t, c.ExpBaseURL, client.Client.BaseURL().String())
		})
	}
}

// This function gets called even if GitlabClient is nil
// so we need to test that.
func TestGitlabClient_SupportsCommonMarkNil(t *testing.T) {
	var gl *GitlabClient
	Equals(t, false, gl.SupportsCommonMark())
}

func TestGitlabClient_SupportsCommonMark(t *testing.T) {
	cases := []struct {
		version string
		exp     bool
	}{
		{
			"11.0",
			false,
		},
		{
			"11.1",
			true,
		},
		{
			"11.2",
			true,
		},
		{
			"12.0",
			true,
		},
	}

	for _, c := range cases {
		t.Run(c.version, func(t *testing.T) {
			vers, err := version.NewVersion(c.version)
			Ok(t, err)
			gl := GitlabClient{
				Version: vers,
			}
			Equals(t, c.exp, gl.SupportsCommonMark())
		})
	}
}

func TestGitlabClient_UpdateStatus(t *testing.T) {
	cases := []struct {
		status   models.VCSStatus
		expState string
	}{
		{
			models.PendingVCSStatus,
			"pending",
		},
		{
			models.SuccessVCSStatus,
			"success",
		},
		{
			models.FailedVCSStatus,
			"failed",
		},
	}
	for _, c := range cases {
		t.Run(c.expState, func(t *testing.T) {
			gotRequest := false
			testServer := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.RequestURI {
					case "/api/v4/projects/runatlantis%2Fatlantis/statuses/sha":
						gotRequest = true

						body, err := io.ReadAll(r.Body)
						Ok(t, err)
						exp := fmt.Sprintf(`{"state":"%s","context":"src","target_url":"https://google.com","description":"description"}`, c.expState)
						Equals(t, exp, string(body))
						defer r.Body.Close()  // nolint: errcheck
						w.Write([]byte("{}")) // nolint: errcheck
					case "/api/v4/":
						// Rate limiter requests.
						w.WriteHeader(http.StatusOK)
					default:
						t.Errorf("got unexpected request at %q", r.RequestURI)
						http.Error(w, "not found", http.StatusNotFound)
					}
				}))

			internalClient, err := gitlab.NewClient("token", gitlab.WithBaseURL(testServer.URL))
			Ok(t, err)
			client := &GitlabClient{
				Client:  internalClient,
				Version: nil,
			}

			repo := models.Repo{
				FullName: "runatlantis/atlantis",
				Owner:    "runatlantis",
				Name:     "atlantis",
			}
			_, err = client.UpdateStatus(context.TODO(), types.UpdateStatusRequest{
				Repo:        repo,
				PullNum:     1,
				Ref:         "sha",
				State:       c.status,
				StatusName:  "src",
				Description: "description",
				DetailsURL:  "https://google.com",
			})

			Ok(t, err)
			Assert(t, gotRequest, "expected to get the request")
		})
	}
}

func TestGitlabClient_MarkdownPullLink(t *testing.T) {
	gitlabClientUnderTest = true
	defer func() { gitlabClientUnderTest = false }()
	client, err := NewGitlabClient("gitlab.com", "token", nil)
	Ok(t, err)
	pull := models.PullRequest{Num: 1}
	s, _ := client.MarkdownPullLink(pull)
	exp := "!1"
	Equals(t, exp, s)
}
