// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package vcs_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/events/vcs/azuredevops"
	"github.com/runatlantis/atlantis/server/events/vcs/bitbucketcloud"
	"github.com/runatlantis/atlantis/server/events/vcs/bitbucketserver"
	"github.com/runatlantis/atlantis/server/events/vcs/gitea"
	"github.com/runatlantis/atlantis/server/events/vcs/github"
	atlantisgitlab "github.com/runatlantis/atlantis/server/events/vcs/gitlab"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func newIdentityClient(t *testing.T, host models.VCSHostType, configuredUser, identity string, status int) vcs.Client {
	t.Helper()
	logger := logging.NewNoopLogger(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/v1/version" {
			_, _ = fmt.Fprint(w, `{"version":"1.25.4"}`)
			return
		}
		w.WriteHeader(status)
		switch host {
		case models.Gitlab:
			Equals(t, "/api/v4/user", r.URL.Path)
			_, _ = fmt.Fprintf(w, `{"username":%q}`, identity)
		case models.Gitea:
			Equals(t, "/api/v1/user", r.URL.Path)
			_, _ = fmt.Fprintf(w, `{"login":%q}`, identity)
		case models.BitbucketCloud:
			Equals(t, "/2.0/user", r.URL.Path)
			apiUser, token, _ := r.BasicAuth()
			Equals(t, "api@example.com", apiUser)
			Equals(t, "token", token)
			_, _ = fmt.Fprintf(w, `{"account_id":"different-from-uuid","uuid":%q}`, identity)
		case models.AzureDevops:
			Equals(t, "/owner/_apis/connectionData", r.URL.Path)
			_, token, _ := r.BasicAuth()
			Equals(t, "token", token)
			_, _ = fmt.Fprintf(w, `{"authenticatedUser":{"id":%q},"authorizedUser":{"id":"different-user"}}`, identity)
		default:
			t.Errorf("unexpected identity request for %s", host)
		}
	}))
	t.Cleanup(server.Close)
	var client vcs.Client
	var err error
	switch host {
	case models.Github:
		client, err = github.New("github.com", &github.UserCredentials{User: configuredUser, Token: "token"}, github.Config{}, 0, logger)
	case models.BitbucketServer:
		client, err = bitbucketserver.NewClient(server.Client(), configuredUser, "token", server.URL, "")
	case models.BitbucketCloud:
		bb := bitbucketcloud.New(server.Client(), configuredUser, "token", "api@example.com", "")
		bb.BaseURL = server.URL
		client = bb
	case models.Gitea:
		client, err = gitea.New(server.URL, configuredUser, "token", 100, logger)
	case models.Gitlab:
		var sdk *gitlab.Client
		sdk, err = gitlab.NewClient("token", gitlab.WithBaseURL(server.URL))
		client = &atlantisgitlab.Client{Client: sdk}
	case models.AzureDevops:
		var ado *azuredevops.Client
		ado, err = azuredevops.New("dev.azure.com", configuredUser, "token")
		baseURL, parseErr := url.Parse(server.URL + "/")
		Ok(t, parseErr)
		ado.Client.BaseURL = *baseURL
		client = ado
	}
	Ok(t, err)
	return client
}

func TestIsAtlantisUser(t *testing.T) {
	for _, host := range []models.VCSHostType{models.Github, models.Gitlab, models.BitbucketCloud, models.BitbucketServer, models.AzureDevops, models.Gitea} {
		t.Run(host.String(), func(t *testing.T) {
			client := newIdentityClient(t, host, "atlantis", "atlantis", http.StatusOK)
			repo := models.Repo{FullName: "owner/project/repo", VCSHost: models.VCSHost{Type: host}}
			for _, c := range []struct {
				author string
				self   bool
			}{
				{"atlantis", true}, {"human", false}, {"other-atlantis[bot]", false}, {"", false},
			} {
				t.Run(c.author, func(t *testing.T) {
					self, err := client.IsAtlantisUser(repo, models.User{Username: c.author, ID: c.author})
					Ok(t, err)
					Equals(t, c.self, self)
				})
			}
			if host == models.BitbucketCloud {
				// Regression for the previous package-wide UUID cache.
				other := newIdentityClient(t, host, "other", "other", http.StatusOK)
				self, err := other.IsAtlantisUser(repo, models.User{Username: "atlantis", ID: "atlantis"})
				Ok(t, err)
				Equals(t, false, self)
			}
		})
	}
}

func TestIsAtlantisUser_MissingIdentity(t *testing.T) {
	for _, host := range []models.VCSHostType{models.Github, models.Gitlab, models.BitbucketCloud, models.BitbucketServer, models.AzureDevops, models.Gitea} {
		t.Run(host.String(), func(t *testing.T) {
			client := newIdentityClient(t, host, "", "", http.StatusOK)
			for _, author := range []string{"", "human"} {
				self, _ := client.IsAtlantisUser(models.Repo{FullName: "owner/project/repo"}, models.User{Username: author, ID: author})
				Equals(t, false, self)
			}
		})
	}
}

func TestIsAtlantisUser_AuthenticationError(t *testing.T) {
	for _, host := range []models.VCSHostType{models.Gitlab, models.BitbucketCloud, models.AzureDevops, models.Gitea} {
		t.Run(host.String(), func(t *testing.T) {
			client := newIdentityClient(t, host, "atlantis", "atlantis", http.StatusUnauthorized)
			self, err := client.IsAtlantisUser(models.Repo{FullName: "owner/project/repo"}, models.User{Username: "atlantis", ID: "atlantis"})
			Assert(t, err != nil, "lookup error must be returned")
			Equals(t, false, self)
		})
	}
}

func TestIsAtlantisUser_StableID(t *testing.T) {
	for _, host := range []models.VCSHostType{models.BitbucketCloud, models.AzureDevops} {
		t.Run(host.String(), func(t *testing.T) {
			client := newIdentityClient(t, host, "configured-name", "account-id", http.StatusOK)
			for _, c := range []struct {
				user models.User
				self bool
			}{
				{models.User{Username: "renamed-account", ID: "account-id"}, true},
				{models.User{Username: "configured-name", ID: "other-id"}, false},
				{models.User{Username: "configured-name"}, false},
			} {
				self, err := client.IsAtlantisUser(models.Repo{FullName: "owner/project/repo"}, c.user)
				Ok(t, err)
				Equals(t, c.self, self)
			}
		})
	}
}
