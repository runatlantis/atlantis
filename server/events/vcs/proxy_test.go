// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package vcs_test

import (
	"errors"
	"testing"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

// makeRepo creates a Repo with the given VCS host type.
func makeRepo(hostType models.VCSHostType) models.Repo {
	return models.Repo{
		VCSHost: models.VCSHost{Type: hostType},
	}
}

// makePull creates a PullRequest whose BaseRepo uses the given VCS host type.
func makePull(hostType models.VCSHostType) models.PullRequest {
	return models.PullRequest{BaseRepo: makeRepo(hostType)}
}

// TestNewClientProxy_NilClientsUseNotConfigured verifies that passing nil for any
// VCS client causes NewClientProxy to install a NotConfiguredVCSClient placeholder.
func TestNewClientProxy_NilClientsUseNotConfigured(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)

	// Build a proxy where every slot is nil – all should become NotConfiguredVCSClient.
	proxy := vcs.NewClientProxy(nil, nil, nil, nil, nil, nil)

	repo := makeRepo(models.Github)
	pull := makePull(models.Github)
	// NotConfiguredVCSClient returns an error for most methods.
	_, err := proxy.GetModifiedFiles(logger, repo, pull)
	Assert(t, err != nil, "expected error from NotConfiguredVCSClient")
}

// TestClientProxy_RoutesToCorrectClient verifies that each proxy method dispatches
// to the mock client that matches the VCS host type embedded in the repo/pull.
func TestClientProxy_RoutesToCorrectClient(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)

	hostTypes := []models.VCSHostType{
		models.Github,
		models.Gitlab,
		models.BitbucketCloud,
		models.BitbucketServer,
		models.AzureDevops,
		models.Gitea,
	}

	for _, hostType := range hostTypes {
		t.Run(hostType.String(), func(t *testing.T) {
			RegisterMockTestingT(t)

			gh := vcsmocks.NewMockClient()
			gl := vcsmocks.NewMockClient()
			bbc := vcsmocks.NewMockClient()
			bbs := vcsmocks.NewMockClient()
			ado := vcsmocks.NewMockClient()
			gitea := vcsmocks.NewMockClient()

			proxy := vcs.NewClientProxy(gh, gl, bbc, bbs, ado, gitea)

			repo := makeRepo(hostType)
			pull := makePull(hostType)

			// Identify which mock corresponds to this host type.
			var expected *vcsmocks.MockClient
			switch hostType {
			case models.Github:
				expected = gh
			case models.Gitlab:
				expected = gl
			case models.BitbucketCloud:
				expected = bbc
			case models.BitbucketServer:
				expected = bbs
			case models.AzureDevops:
				expected = ado
			case models.Gitea:
				expected = gitea
			}

			// --- GetModifiedFiles ---
			When(expected.GetModifiedFiles(logger, repo, pull)).ThenReturn([]string{"a.tf"}, nil)
			files, err := proxy.GetModifiedFiles(logger, repo, pull)
			Ok(t, err)
			Equals(t, []string{"a.tf"}, files)
			expected.VerifyWasCalledOnce().GetModifiedFiles(logger, repo, pull)

			// --- CreateComment ---
			When(expected.CreateComment(logger, repo, pull.Num, "msg", "plan")).ThenReturn(nil)
			err = proxy.CreateComment(logger, repo, pull.Num, "msg", "plan")
			Ok(t, err)
			expected.VerifyWasCalledOnce().CreateComment(logger, repo, pull.Num, "msg", "plan")

			// --- HidePrevCommandComments ---
			When(expected.HidePrevCommandComments(logger, repo, pull.Num, "plan", "")).ThenReturn(nil)
			err = proxy.HidePrevCommandComments(logger, repo, pull.Num, "plan", "")
			Ok(t, err)
			expected.VerifyWasCalledOnce().HidePrevCommandComments(logger, repo, pull.Num, "plan", "")

			// --- PullIsApproved ---
			approvalStatus := models.ApprovalStatus{IsApproved: true}
			When(expected.PullIsApproved(logger, repo, pull)).ThenReturn(approvalStatus, nil)
			gotApproval, err := proxy.PullIsApproved(logger, repo, pull)
			Ok(t, err)
			Equals(t, approvalStatus, gotApproval)
			expected.VerifyWasCalledOnce().PullIsApproved(logger, repo, pull)

			// --- DiscardReviews ---
			When(expected.DiscardReviews(logger, repo, pull)).ThenReturn(nil)
			err = proxy.DiscardReviews(logger, repo, pull)
			Ok(t, err)
			expected.VerifyWasCalledOnce().DiscardReviews(logger, repo, pull)

			// --- PullIsMergeable ---
			mergeableStatus := models.MergeableStatus{IsMergeable: true}
			When(expected.PullIsMergeable(logger, repo, pull, "atlantis", []string{})).ThenReturn(mergeableStatus, nil)
			gotMergeable, err := proxy.PullIsMergeable(logger, repo, pull, "atlantis", []string{})
			Ok(t, err)
			Equals(t, mergeableStatus, gotMergeable)
			expected.VerifyWasCalledOnce().PullIsMergeable(logger, repo, pull, "atlantis", []string{})

			// --- UpdateStatus ---
			When(expected.UpdateStatus(logger, repo, pull, models.SuccessCommitStatus, "src", "desc", "url")).ThenReturn(nil)
			err = proxy.UpdateStatus(logger, repo, pull, models.SuccessCommitStatus, "src", "desc", "url")
			Ok(t, err)
			expected.VerifyWasCalledOnce().UpdateStatus(logger, repo, pull, models.SuccessCommitStatus, "src", "desc", "url")

			// --- MergePull ---
			opts := models.PullRequestOptions{DeleteSourceBranchOnMerge: true}
			When(expected.MergePull(logger, pull, opts)).ThenReturn(nil)
			err = proxy.MergePull(logger, pull, opts)
			Ok(t, err)
			expected.VerifyWasCalledOnce().MergePull(logger, pull, opts)

			// --- MarkdownPullLink ---
			When(expected.MarkdownPullLink(pull)).ThenReturn("[#1](url)", nil)
			link, err := proxy.MarkdownPullLink(pull)
			Ok(t, err)
			Equals(t, "[#1](url)", link)
			expected.VerifyWasCalledOnce().MarkdownPullLink(pull)

			// --- GetTeamNamesForUser ---
			user := models.User{Username: "bob"}
			When(expected.GetTeamNamesForUser(logger, repo, user)).ThenReturn([]string{"team-a"}, nil)
			teams, err := proxy.GetTeamNamesForUser(logger, repo, user)
			Ok(t, err)
			Equals(t, []string{"team-a"}, teams)
			expected.VerifyWasCalledOnce().GetTeamNamesForUser(logger, repo, user)

			// --- GetFileContent ---
			When(expected.GetFileContent(logger, repo, "main", "foo.tf")).ThenReturn(true, []byte("content"), nil)
			found, content, err := proxy.GetFileContent(logger, repo, "main", "foo.tf")
			Ok(t, err)
			Equals(t, true, found)
			Equals(t, []byte("content"), content)
			expected.VerifyWasCalledOnce().GetFileContent(logger, repo, "main", "foo.tf")

			// --- SupportsSingleFileDownload ---
			When(expected.SupportsSingleFileDownload(repo)).ThenReturn(true)
			Equals(t, true, proxy.SupportsSingleFileDownload(repo))
			expected.VerifyWasCalledOnce().SupportsSingleFileDownload(repo)

			// --- GetCloneURL ---
			When(expected.GetCloneURL(logger, hostType, "owner/repo")).ThenReturn("https://clone.url", nil)
			cloneURL, err := proxy.GetCloneURL(logger, hostType, "owner/repo")
			Ok(t, err)
			Equals(t, "https://clone.url", cloneURL)
			expected.VerifyWasCalledOnce().GetCloneURL(logger, hostType, "owner/repo")

			// --- GetPullLabels ---
			When(expected.GetPullLabels(logger, repo, pull)).ThenReturn([]string{"label-a"}, nil)
			labels, err := proxy.GetPullLabels(logger, repo, pull)
			Ok(t, err)
			Equals(t, []string{"label-a"}, labels)
			expected.VerifyWasCalledOnce().GetPullLabels(logger, repo, pull)

			// --- ReactToComment ---
			When(expected.ReactToComment(logger, repo, pull.Num, int64(42), "eyes")).ThenReturn(nil)
			err = proxy.ReactToComment(logger, repo, pull.Num, int64(42), "eyes")
			Ok(t, err)
			expected.VerifyWasCalledOnce().ReactToComment(logger, repo, pull.Num, int64(42), "eyes")
		})
	}
}

// TestClientProxy_PropagatesErrors verifies that errors returned by the underlying
// client are propagated back to the caller unchanged.
func TestClientProxy_PropagatesErrors(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)

	gh := vcsmocks.NewMockClient()
	proxy := vcs.NewClientProxy(gh, nil, nil, nil, nil, nil)

	repo := makeRepo(models.Github)
	pull := makePull(models.Github)
	wantErr := errors.New("api error")

	When(gh.GetModifiedFiles(logger, repo, pull)).ThenReturn(nil, wantErr)
	_, err := proxy.GetModifiedFiles(logger, repo, pull)
	Assert(t, err != nil, "expected error")
	Equals(t, wantErr, err)

	When(gh.CreateComment(logger, repo, pull.Num, "msg", "plan")).ThenReturn(wantErr)
	err = proxy.CreateComment(logger, repo, pull.Num, "msg", "plan")
	Assert(t, err != nil, "expected error")
	Equals(t, wantErr, err)

	When(gh.PullIsApproved(logger, repo, pull)).ThenReturn(models.ApprovalStatus{}, wantErr)
	_, err = proxy.PullIsApproved(logger, repo, pull)
	Assert(t, err != nil, "expected error")
	Equals(t, wantErr, err)

	When(gh.PullIsMergeable(logger, repo, pull, "atlantis", []string{})).ThenReturn(models.MergeableStatus{}, wantErr)
	_, err = proxy.PullIsMergeable(logger, repo, pull, "atlantis", []string{})
	Assert(t, err != nil, "expected error")
	Equals(t, wantErr, err)

	When(gh.UpdateStatus(logger, repo, pull, models.SuccessCommitStatus, "src", "desc", "url")).ThenReturn(wantErr)
	err = proxy.UpdateStatus(logger, repo, pull, models.SuccessCommitStatus, "src", "desc", "url")
	Assert(t, err != nil, "expected error")
	Equals(t, wantErr, err)
}
