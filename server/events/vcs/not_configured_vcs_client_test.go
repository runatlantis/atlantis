// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package vcs_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

// TestNotConfiguredVCSClient_ErrorMethods verifies that methods that should
// return an error all do so with a message that mentions the host type.
func TestNotConfiguredVCSClient_ErrorMethods(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	client := &vcs.NotConfiguredVCSClient{Host: models.Github}

	repo := models.Repo{}
	pull := models.PullRequest{}

	_, err := client.GetModifiedFiles(logger, repo, pull)
	Assert(t, err != nil, "GetModifiedFiles should return an error")
	ErrContains(t, "Github", err)

	err = client.CreateComment(logger, repo, 1, "msg", "plan")
	Assert(t, err != nil, "CreateComment should return an error")

	_, err = client.PullIsApproved(logger, repo, pull)
	Assert(t, err != nil, "PullIsApproved should return an error")

	_, err = client.PullIsMergeable(logger, repo, pull, "atlantis", nil)
	Assert(t, err != nil, "PullIsMergeable should return an error")

	err = client.UpdateStatus(logger, repo, pull, models.SuccessCommitStatus, "src", "desc", "url")
	Assert(t, err != nil, "UpdateStatus should return an error")

	err = client.MergePull(logger, pull, models.PullRequestOptions{})
	Assert(t, err != nil, "MergePull should return an error")

	_, err = client.MarkdownPullLink(pull)
	Assert(t, err != nil, "MarkdownPullLink should return an error")

	_, err = client.GetTeamNamesForUser(logger, repo, models.User{})
	Assert(t, err != nil, "GetTeamNamesForUser should return an error")

	_, _, err = client.GetFileContent(logger, repo, "main", "foo.tf")
	Assert(t, err != nil, "GetFileContent should return an error")

	_, err = client.GetCloneURL(logger, models.Github, "owner/repo")
	Assert(t, err != nil, "GetCloneURL should return an error")

	_, err = client.GetPullLabels(logger, repo, pull)
	Assert(t, err != nil, "GetPullLabels should return an error")
}

// TestNotConfiguredVCSClient_NoopMethods verifies that no-op methods return
// no error (they silently ignore calls).
func TestNotConfiguredVCSClient_NoopMethods(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	client := &vcs.NotConfiguredVCSClient{Host: models.Gitlab}

	repo := models.Repo{}
	pull := models.PullRequest{}

	// HidePrevCommandComments is a no-op – returns nil.
	err := client.HidePrevCommandComments(logger, repo, 1, "plan", "")
	Ok(t, err)

	// ReactToComment is a no-op – returns nil.
	err = client.ReactToComment(logger, repo, 1, int64(99), "eyes")
	Ok(t, err)

	// DiscardReviews is a no-op – returns nil.
	err = client.DiscardReviews(logger, repo, pull)
	Ok(t, err)

	// SupportsSingleFileDownload always returns false.
	Equals(t, false, client.SupportsSingleFileDownload(repo))
}

// TestNotConfiguredVCSClient_ErrorMessageContainsHost verifies that the error
// message returned by error-producing methods always identifies the host.
func TestNotConfiguredVCSClient_ErrorMessageContainsHost(t *testing.T) {
	cases := []models.VCSHostType{
		models.Github,
		models.Gitlab,
		models.BitbucketCloud,
		models.BitbucketServer,
		models.AzureDevops,
		models.Gitea,
	}

	logger := logging.NewNoopLogger(t)
	for _, hostType := range cases {
		t.Run(hostType.String(), func(t *testing.T) {
			client := &vcs.NotConfiguredVCSClient{Host: hostType}
			repo := models.Repo{}
			pull := models.PullRequest{}

			_, err := client.GetModifiedFiles(logger, repo, pull)
			Assert(t, err != nil, "expected error")
			ErrContains(t, hostType.String(), err)
		})
	}
}
