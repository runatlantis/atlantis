// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/runatlantis/atlantis/server/core/ownership"
	"github.com/runatlantis/atlantis/server/events/models"
)

// RepoRef is a credential-free repository reference used for internal routing.
type RepoRef struct {
	FullName string         `json:"full_name"`
	CloneURL string         `json:"clone_url"`
	VCSHost  models.VCSHost `json:"vcs_host"`
}

// NewRepoRef removes credentials and nonessential URL components from a repository.
func NewRepoRef(repo models.Repo) (RepoRef, error) {
	parsed, err := url.Parse(repo.CloneURL)
	if err != nil {
		return RepoRef{}, fmt.Errorf("parsing clone URL: %w", err)
	}
	if !isHTTPCloneURL(parsed) {
		return RepoRef{}, errorsForCloneURL(parsed)
	}
	parsed.User = nil
	parsed.RawQuery = ""
	parsed.ForceQuery = false
	parsed.Fragment = ""
	parsed.RawFragment = ""
	return RepoRef{FullName: repo.FullName, CloneURL: parsed.String(), VCSHost: repo.VCSHost}, nil
}

// PullRef contains pull request data without its credential-bearing base repository.
type PullRef struct {
	Num                      int                     `json:"num"`
	HeadCommit               string                  `json:"head_commit"`
	URL                      string                  `json:"url"`
	HeadBranch               string                  `json:"head_branch"`
	BaseBranch               string                  `json:"base_branch"`
	HardenedNonPRRefCheckout bool                    `json:"hardened_non_pr_ref_checkout"`
	Author                   string                  `json:"author"`
	Body                     string                  `json:"body"`
	State                    models.PullRequestState `json:"state"`
}

// NewPullRef removes the base repository from a pull request.
func NewPullRef(pull models.PullRequest) PullRef {
	return PullRef{
		Num:                      pull.Num,
		HeadCommit:               pull.HeadCommit,
		URL:                      pull.URL,
		HeadBranch:               pull.HeadBranch,
		BaseBranch:               pull.BaseBranch,
		HardenedNonPRRefCheckout: pull.HardenedNonPRRefCheckout,
		Author:                   pull.Author,
		Body:                     pull.Body,
		State:                    pull.State,
	}
}

// ToModel rebuilds a pull request with an owner-local base repository.
func (p PullRef) ToModel(baseRepo models.Repo) models.PullRequest {
	return models.PullRequest{
		Num:                      p.Num,
		HeadCommit:               p.HeadCommit,
		URL:                      p.URL,
		HeadBranch:               p.HeadBranch,
		BaseBranch:               p.BaseBranch,
		HardenedNonPRRefCheckout: p.HardenedNonPRRefCheckout,
		Author:                   p.Author,
		Body:                     p.Body,
		State:                    p.State,
		BaseRepo:                 baseRepo,
	}
}

// CommentDispatch is a routed pull request comment command.
type CommentDispatch struct {
	BaseRepo RepoRef         `json:"base_repo"`
	HeadRepo *RepoRef        `json:"head_repo,omitempty"`
	Pull     *PullRef        `json:"pull,omitempty"`
	User     models.User     `json:"user"`
	PullNum  int             `json:"pull_num"`
	Command  *CommentCommand `json:"command"`
}

// AutoplanDispatch is a routed pull request autoplan command.
type AutoplanDispatch struct {
	BaseRepo RepoRef     `json:"base_repo"`
	HeadRepo RepoRef     `json:"head_repo"`
	Pull     PullRef     `json:"pull"`
	User     models.User `json:"user"`
}

// PullClosedDispatch is routed owner-local pull request cleanup.
type PullClosedDispatch struct {
	BaseRepo RepoRef `json:"base_repo"`
	Pull     PullRef `json:"pull"`
}

// NewCommentDispatch creates a credential-free comment command envelope.
func NewCommentDispatch(
	baseRepo models.Repo,
	headRepo *models.Repo,
	pull *models.PullRequest,
	user models.User,
	pullNum int,
	cmd *CommentCommand,
) (CommentDispatch, error) {
	baseRef, err := NewRepoRef(baseRepo)
	if err != nil {
		return CommentDispatch{}, err
	}
	var headRef *RepoRef
	if headRepo != nil {
		value, err := NewRepoRef(*headRepo)
		if err != nil {
			return CommentDispatch{}, err
		}
		headRef = &value
	}
	var pullRef *PullRef
	if pull != nil {
		value := NewPullRef(*pull)
		pullRef = &value
	}
	return CommentDispatch{
		BaseRepo: baseRef,
		HeadRepo: headRef,
		Pull:     pullRef,
		User:     user,
		PullNum:  pullNum,
		Command:  cmd,
	}, nil
}

// NewAutoplanDispatch creates a credential-free autoplan envelope.
func NewAutoplanDispatch(baseRepo, headRepo models.Repo, pull models.PullRequest, user models.User) (AutoplanDispatch, error) {
	baseRef, err := NewRepoRef(baseRepo)
	if err != nil {
		return AutoplanDispatch{}, err
	}
	headRef, err := NewRepoRef(headRepo)
	if err != nil {
		return AutoplanDispatch{}, err
	}
	return AutoplanDispatch{BaseRepo: baseRef, HeadRepo: headRef, Pull: NewPullRef(pull), User: user}, nil
}

// NewPullClosedDispatch creates a credential-free pull cleanup envelope.
func NewPullClosedDispatch(baseRepo models.Repo, pull models.PullRequest) (PullClosedDispatch, error) {
	baseRef, err := NewRepoRef(baseRepo)
	if err != nil {
		return PullClosedDispatch{}, err
	}
	return PullClosedDispatch{BaseRepo: baseRef, Pull: NewPullRef(pull)}, nil
}

// OwnershipKey returns the pull request affinity key for a comment command.
func (d CommentDispatch) OwnershipKey() ownership.Key {
	return dispatchOwnershipKey(d.BaseRepo, d.PullNum)
}

// OwnershipKey returns the pull request affinity key for an autoplan command.
func (d AutoplanDispatch) OwnershipKey() ownership.Key {
	return dispatchOwnershipKey(d.BaseRepo, d.Pull.Num)
}

// OwnershipKey returns the pull request affinity key for pull cleanup.
func (d PullClosedDispatch) OwnershipKey() ownership.Key {
	return dispatchOwnershipKey(d.BaseRepo, d.Pull.Num)
}

func dispatchOwnershipKey(repo RepoRef, pullNum int) ownership.Key {
	return ownership.Key{
		VCSHostname:  strings.ToLower(strings.TrimSpace(repo.VCSHost.Hostname)),
		RepoFullName: strings.TrimSpace(repo.FullName),
		PullNum:      pullNum,
	}
}

// HydrateRepo rebuilds a credential-free repository with this replica's VCS credentials.
func (e *EventParser) HydrateRepo(ref RepoRef) (models.Repo, error) {
	if err := validateRepoRef(ref); err != nil {
		return models.Repo{}, err
	}

	user, token, vcsHostname := "", "", ""
	switch ref.VCSHost.Type {
	case models.Github:
		user, token = e.GithubUser, e.GithubToken
		if e.GithubTokenFile != "" {
			contents, err := os.ReadFile(e.GithubTokenFile)
			if err != nil {
				return models.Repo{}, fmt.Errorf("reading github token file: %w", err)
			}
			token = strings.TrimSpace(string(contents))
		}
	case models.Gitlab:
		user, token, vcsHostname = e.GitlabUser, e.GitlabToken, e.GitlabHostname
	case models.Gitea:
		user, token = e.GiteaUser, e.GiteaToken
	case models.BitbucketCloud, models.BitbucketServer:
		user, token = e.BitbucketUser, e.BitbucketToken
	case models.AzureDevops:
		user, token = e.AzureDevopsUser, e.AzureDevopsToken
	default:
		return models.Repo{}, fmt.Errorf("unsupported forwarded VCS type %v", ref.VCSHost.Type)
	}

	repo, err := models.NewRepo(ref.VCSHost.Type, ref.FullName, ref.CloneURL, user, token, vcsHostname)
	if err != nil {
		return models.Repo{}, fmt.Errorf("rebuilding forwarded repository: %w", err)
	}
	if !strings.EqualFold(repo.VCSHost.Hostname, ref.VCSHost.Hostname) || repo.VCSHost.Type != ref.VCSHost.Type || repo.FullName != ref.FullName {
		return models.Repo{}, errorsForRepositoryIdentity(ref, repo)
	}
	return repo, nil
}

func validateRepoRef(ref RepoRef) error {
	parsed, err := url.Parse(ref.CloneURL)
	if err != nil {
		return fmt.Errorf("parsing forwarded clone URL: %w", err)
	}
	if !isHTTPCloneURL(parsed) || parsed.User != nil || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" {
		return errorsForCloneURL(parsed)
	}
	if ref.FullName == "" || ref.VCSHost.Hostname == "" || !strings.EqualFold(parsed.Hostname(), ref.VCSHost.Hostname) {
		return fmt.Errorf("forwarded repository identity does not match clone URL")
	}
	return nil
}

func isHTTPCloneURL(parsed *url.URL) bool {
	return (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
}

func errorsForCloneURL(parsed *url.URL) error {
	return fmt.Errorf("clone URL %q must be credential-free absolute HTTP(S)", parsed.Redacted())
}

func errorsForRepositoryIdentity(ref RepoRef, repo models.Repo) error {
	return fmt.Errorf(
		"forwarded repository %q on %q does not match rebuilt repository %q on %q",
		ref.FullName,
		ref.VCSHost.Hostname,
		repo.FullName,
		repo.VCSHost.Hostname,
	)
}
