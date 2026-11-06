package vcs

import (
	"context"
	"fmt"

	"github.com/google/go-github/v71/github"
	"github.com/runatlantis/atlantis/internal/domain/vcs"
	"golang.org/x/oauth2"
)

// GitHubPlugin implements VCSPlugin for GitHub
type GitHubPlugin struct {
	client        *github.Client
	hostname      string
	teamAllowlist []string
}

// GitHubConfig holds GitHub-specific configuration
type GitHubConfig struct {
	Token         string
	Hostname      string
	TeamAllowlist []string
}

// NewGitHubPlugin creates a new GitHub VCS plugin
func NewGitHubPlugin(config GitHubConfig) *GitHubPlugin {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.Token},
	)
	tc := oauth2.NewClient(ctx, ts)

	var client *github.Client
	if config.Hostname != "" && config.Hostname != "github.com" {
		// GitHub Enterprise
		baseURL := fmt.Sprintf("https://%s/api/v3/", config.Hostname)
		client, _ = github.NewEnterpriseClient(baseURL, baseURL, tc)
	} else {
		// GitHub.com
		client = github.NewClient(tc)
	}

	return &GitHubPlugin{
		client:        client,
		hostname:      config.Hostname,
		teamAllowlist: config.TeamAllowlist,
	}
}

// GetRepository retrieves repository information from GitHub
func (g *GitHubPlugin) GetRepository(ctx context.Context, owner, name string) (*vcs.Repository, error) {
	repo, _, err := g.client.Repositories.Get(ctx, owner, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository %s/%s: %w", owner, name, err)
	}

	return &vcs.Repository{
		FullName:      repo.GetFullName(),
		Owner:         repo.GetOwner().GetLogin(),
		Name:          repo.GetName(),
		HTMLURL:       repo.GetHTMLURL(),
		CloneURL:      repo.GetCloneURL(),
		DefaultBranch: repo.GetDefaultBranch(),
	}, nil
}

// GetPullRequest retrieves pull request information from GitHub
func (g *GitHubPlugin) GetPullRequest(ctx context.Context, repo vcs.Repository, number int) (*vcs.PullRequest, error) {
	pr, _, err := g.client.PullRequests.Get(ctx, repo.Owner, repo.Name, number)
	if err != nil {
		return nil, fmt.Errorf("failed to get pull request %s#%d: %w", repo.FullName, number, err)
	}

	var state vcs.PullRequestState
	switch pr.GetState() {
	case "open":
		state = vcs.PullRequestOpen
	case "closed":
		if pr.GetMerged() {
			state = vcs.PullRequestMerged
		} else {
			state = vcs.PullRequestClosed
		}
	default:
		state = vcs.PullRequestClosed
	}

	return &vcs.PullRequest{
		Number:     pr.GetNumber(),
		Title:      pr.GetTitle(),
		Author:     pr.GetUser().GetLogin(),
		HeadSHA:    pr.GetHead().GetSHA(),
		HeadBranch: pr.GetHead().GetRef(),
		BaseBranch: pr.GetBase().GetRef(),
		State:      state,
		URL:        pr.GetHTMLURL(),
		UpdatedAt:  pr.GetUpdatedAt().Time,
		Mergeable:  pr.Mergeable,
	}, nil
}

// CreateCommitStatus creates a commit status on GitHub
func (g *GitHubPlugin) CreateCommitStatus(ctx context.Context, repo vcs.Repository, sha string, status vcs.CommitStatus) error {
	state := string(status.State)
	
	repoStatus := &github.RepoStatus{
		State:       &state,
		Description: &status.Description,
		Context:     &status.Context,
	}
	
	if status.TargetURL != "" {
		repoStatus.TargetURL = &status.TargetURL
	}

	_, _, err := g.client.Repositories.CreateStatus(ctx, repo.Owner, repo.Name, sha, repoStatus)
	if err != nil {
		return fmt.Errorf("failed to create commit status: %w", err)
	}

	return nil
}

// Capabilities returns GitHub's VCS capabilities
func (g *GitHubPlugin) Capabilities() vcs.VCSCapabilities {
	return vcs.VCSCapabilities{
		SupportsMergeableBypass: true,
		SupportsTeamAllowlist:   true,
		SupportsGroupAllowlist:  false, // GitHub uses teams, not groups
		SupportsCustomFields:    true,
		MaxPageSize:            100,
	}
}

// CheckMergeableBypass checks if mergeable bypass is allowed for a PR
// This addresses gh-allow-mergeable-bypass-apply from issue #5574
func (g *GitHubPlugin) CheckMergeableBypass(ctx context.Context, pr *vcs.PullRequest) (bool, error) {
	if !g.Capabilities().SupportsMergeableBypass {
		return false, vcs.NewUnsupportedFeatureError("github", "mergeable-bypass")
	}

	// If mergeable is nil, we can't determine mergeability, so allow bypass
	if pr.Mergeable == nil {
		return true, nil
	}

	// If PR is mergeable, no bypass needed
	if *pr.Mergeable {
		return false, nil
	}

	// PR is not mergeable, check if we have additional logic here
	// (e.g., checking for specific labels, author permissions, etc.)
	return true, nil
}

// ValidateTeamMembership validates if a user is a member of allowed teams
// This addresses gh-team-allowlist from issue #5574
func (g *GitHubPlugin) ValidateTeamMembership(ctx context.Context, user string, teams []string) (bool, error) {
	if !g.Capabilities().SupportsTeamAllowlist {
		return false, vcs.NewUnsupportedFeatureError("github", "team-allowlist")
	}

	if len(teams) == 0 {
		return true, nil // No team restrictions
	}

	// Check each team for membership
	for _, teamSlug := range teams {
		// Parse team slug to get org and team name
		// Format: "org/team" or just "team" (if single org context)
		org, team := parseTeamSlug(teamSlug)
		if org == "" {
			// If no org specified, we need context - this could be improved
			// with additional configuration
			continue
		}

		membership, _, err := g.client.Teams.GetTeamMembershipBySlug(ctx, org, team, user)
		if err != nil {
			// User not found in team or team doesn't exist
			continue
		}

		if membership.GetState() == "active" {
			return true, nil
		}
	}

	return false, nil
}

// ValidateGroupMembership is not supported by GitHub (uses teams instead)
func (g *GitHubPlugin) ValidateGroupMembership(ctx context.Context, user string, groups []string) (bool, error) {
	return false, vcs.NewUnsupportedFeatureError("github", "group-allowlist")
}

// parseTeamSlug parses a team slug in format "org/team" or returns empty org for "team"
func parseTeamSlug(teamSlug string) (org, team string) {
	// Simple implementation - could be enhanced with better parsing
	for i, c := range teamSlug {
		if c == '/' {
			return teamSlug[:i], teamSlug[i+1:]
		}
	}
	return "", teamSlug
} 