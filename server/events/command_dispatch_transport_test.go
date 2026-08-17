// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/stretchr/testify/require"
)

func TestNewCommentDispatch_DoesNotSerializeCloneCredentials(t *testing.T) {
	repo, err := models.NewRepo(
		models.Github,
		"owner/repo",
		"https://github.com/owner/repo",
		"atlantis-user",
		"receiver-vcs-token",
		"",
	)
	require.NoError(t, err)

	dispatch, err := events.NewCommentDispatch(
		repo,
		nil,
		nil,
		models.User{Username: "alice", Teams: []string{"platform"}},
		12,
		&events.CommentCommand{Name: command.Plan},
	)
	require.NoError(t, err)
	payload, err := json.Marshal(dispatch)
	require.NoError(t, err)
	require.NotContains(t, string(payload), "receiver-vcs-token")
	require.NotContains(t, string(payload), "atlantis-user")
	require.NotContains(t, string(payload), "<redacted>")
	require.Contains(t, string(payload), "https://github.com/owner/repo.git")
}

func TestPullRef_RoundTripsAllNonRepositoryFields(t *testing.T) {
	baseRepo := models.Repo{FullName: "owner/repo"}
	pull := models.PullRequest{
		Num:                      12,
		HeadCommit:               "abc123",
		URL:                      "https://github.com/owner/repo/pull/12",
		HeadBranch:               "feature",
		BaseBranch:               "main",
		HardenedNonPRRefCheckout: true,
		Author:                   "alice",
		Body:                     "body",
		State:                    models.OpenPullState,
		BaseRepo:                 models.Repo{FullName: "credential-bearing/source"},
	}

	roundTrip := events.NewPullRef(pull).ToModel(baseRepo)
	pull.BaseRepo = baseRepo
	require.Equal(t, pull, roundTrip)
}

func TestEventParser_HydrateRepoUsesLocalToken(t *testing.T) {
	parser := events.EventParser{GithubUser: "owner-user", GithubToken: "owner-vcs-token"}
	repo, err := parser.HydrateRepo(events.RepoRef{
		FullName: "owner/repo",
		CloneURL: "https://github.com/owner/repo.git",
		VCSHost:  models.VCSHost{Type: models.Github, Hostname: "github.com"},
	})
	require.NoError(t, err)
	require.Contains(t, repo.CloneURL, "owner-user:owner-vcs-token@github.com")
}

func TestEventParser_HydrateRepoSupportsWebhookProviders(t *testing.T) {
	parser := events.EventParser{
		GithubUser:       "github-user",
		GithubToken:      "github-token",
		GitlabUser:       "gitlab-user",
		GitlabToken:      "gitlab-token",
		GitlabHostname:   "gitlab.example.com/gitlab",
		GiteaUser:        "gitea-user",
		GiteaToken:       "gitea-token",
		BitbucketUser:    "bitbucket-user",
		BitbucketToken:   "bitbucket-token",
		AzureDevopsUser:  "azure-user",
		AzureDevopsToken: "azure-token",
	}
	tests := []struct {
		name          string
		vcsType       models.VCSHostType
		fullName      string
		cloneURL      string
		vcsHostname   string
		expectedUser  string
		expectedToken string
	}{
		{"github", models.Github, "owner/repo", "https://github.example.com/owner/repo", "", "github-user", "github-token"},
		{"gitlab subpath", models.Gitlab, "group/subgroup/repo", "https://gitlab.example.com/gitlab/group/subgroup/repo", "gitlab.example.com/gitlab", "gitlab-user", "gitlab-token"},
		{"gitea", models.Gitea, "owner/repo", "https://gitea.example.com/owner/repo", "", "gitea-user", "gitea-token"},
		{"bitbucket cloud", models.BitbucketCloud, "owner/repo", "https://bitbucket.org/owner/repo", "", "bitbucket-user", "bitbucket-token"},
		{"bitbucket server", models.BitbucketServer, "PROJECT/repo", "https://bitbucket.example.com/scm/project/repo.git", "", "bitbucket-user", "bitbucket-token"},
		{"azure devops", models.AzureDevops, "org/project/repo", "https://dev.azure.com/org/project/_git/repo", "", "azure-user", "azure-token"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source, err := models.NewRepo(test.vcsType, test.fullName, test.cloneURL, "receiver", "receiver-token", test.vcsHostname)
			require.NoError(t, err)
			ref, err := events.NewRepoRef(source)
			require.NoError(t, err)

			hydrated, err := parser.HydrateRepo(ref)
			require.NoError(t, err)
			require.Equal(t, test.vcsType, hydrated.VCSHost.Type)
			require.Equal(t, test.fullName, hydrated.FullName)
			parsed, err := url.Parse(hydrated.CloneURL)
			require.NoError(t, err)
			require.Equal(t, test.expectedUser, parsed.User.Username())
			password, present := parsed.User.Password()
			require.True(t, present)
			require.Equal(t, test.expectedToken, password)
		})
	}
}

func TestEventParser_HydrateRepoReadsRotatedTokenFile(t *testing.T) {
	tokenPath := filepath.Join(t.TempDir(), "token")
	require.NoError(t, os.WriteFile(tokenPath, []byte("rotated-token\n"), 0o600))
	parser := events.EventParser{
		GithubUser: "owner-user", GithubToken: "stale-token", GithubTokenFile: tokenPath,
	}

	repo, err := parser.HydrateRepo(events.RepoRef{
		FullName: "owner/repo",
		CloneURL: "https://github.com/owner/repo.git",
		VCSHost:  models.VCSHost{Type: models.Github, Hostname: "github.com"},
	})
	require.NoError(t, err)
	require.Contains(t, repo.CloneURL, "owner-user:rotated-token@github.com")
	require.NotContains(t, repo.CloneURL, "stale-token")
}

func TestEventParser_HydrateRepoRejectsUnsafeCloneURL(t *testing.T) {
	parser := events.EventParser{GithubUser: "owner-user", GithubToken: "owner-vcs-token"}
	tests := []events.RepoRef{
		{
			FullName: "owner/repo",
			CloneURL: "https://attacker:token@github.com/owner/repo.git",
			VCSHost:  models.VCSHost{Type: models.Github, Hostname: "github.com"},
		},
		{
			FullName: "owner/repo",
			CloneURL: "file:///tmp/repo",
			VCSHost:  models.VCSHost{Type: models.Github, Hostname: "github.com"},
		},
		{
			FullName: "owner/repo",
			CloneURL: "https://git.example.com/owner/repo.git",
			VCSHost:  models.VCSHost{Type: models.Github, Hostname: "github.com"},
		},
	}

	for _, ref := range tests {
		_, err := parser.HydrateRepo(ref)
		require.Error(t, err)
	}
}
