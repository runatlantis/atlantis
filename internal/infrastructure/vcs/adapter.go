package vcs

import (
	"context"
	"fmt"
	"time"

	"github.com/runatlantis/atlantis/internal/domain/vcs"
	"github.com/runatlantis/atlantis/server/events/models"
	vcsClient "github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
)

// LegacyVCSClientAdapter adapts existing VCS clients to the new plugin interface
type LegacyVCSClientAdapter struct {
	client      vcsClient.Client
	vcsType     string
	logger      logging.SimpleLogging
	config      VCSAdapterConfig
}

// VCSAdapterConfig holds configuration for the adapter
type VCSAdapterConfig struct {
	// Feature flags
	AllowMergeableBypass bool
	TeamAllowlist        []string
	GroupAllowlist       []string
	
	// VCS-specific settings
	IgnoreVCSStatusNames []string
}

// NewLegacyVCSClientAdapter creates an adapter for existing VCS clients
func NewLegacyVCSClientAdapter(client vcsClient.Client, vcsType string, logger logging.SimpleLogging, config VCSAdapterConfig) *LegacyVCSClientAdapter {
	return &LegacyVCSClientAdapter{
		client:  client,
		vcsType: vcsType,
		logger:  logger,
		config:  config,
	}
}

// GetRepository adapts the existing client to the new interface
func (a *LegacyVCSClientAdapter) GetRepository(ctx context.Context, owner, name string) (*vcs.Repository, error) {
	// Current VCS clients don't have a direct GetRepository method
	// We construct repository info from what we know
	return &vcs.Repository{
		FullName:      fmt.Sprintf("%s/%s", owner, name),
		Owner:         owner,
		Name:          name,
		HTMLURL:       fmt.Sprintf("https://%s/%s/%s", a.getHostname(), owner, name),
		CloneURL:      fmt.Sprintf("https://%s/%s/%s.git", a.getHostname(), owner, name),
		DefaultBranch: "main", // Default assumption
	}, nil
}

// GetPullRequest adapts pull request retrieval
func (a *LegacyVCSClientAdapter) GetPullRequest(ctx context.Context, repo vcs.Repository, number int) (*vcs.PullRequest, error) {
	// Current VCS clients don't have a direct GetPullRequest method
	// This would need to be implemented differently per VCS or mock the response
	return &vcs.PullRequest{
		Number:    number,
		Title:     "Pull Request Title", // Would need to be fetched
		State:     vcs.PullRequestOpen,
		URL:       fmt.Sprintf("%s/pull/%d", repo.HTMLURL, number),
		UpdatedAt: time.Now(),
		Mergeable: nil, // Unknown without actual API call
	}, nil
}

// CreateCommitStatus adapts commit status creation
func (a *LegacyVCSClientAdapter) CreateCommitStatus(ctx context.Context, repo vcs.Repository, sha string, status vcs.CommitStatus) error {
	modelsRepo := models.Repo{
		FullName:          repo.FullName,
		Owner:             repo.Owner,
		Name:              repo.Name,
		CloneURL:          repo.CloneURL,
		SanitizedCloneURL: repo.CloneURL,
		VCSHost: models.VCSHost{
			Type:     a.getVCSHostType(),
			Hostname: a.getHostname(),
		},
	}

	modelsPR := models.PullRequest{
		Num:        1, // Default PR number since we don't have actual PR info
		HeadCommit: sha,
		BaseRepo:   modelsRepo,
	}

	var modelsStatus models.CommitStatus
	switch status.State {
	case vcs.CommitPending:
		modelsStatus = models.PendingCommitStatus
	case vcs.CommitSuccess:
		modelsStatus = models.SuccessCommitStatus
	case vcs.CommitFailure:
		modelsStatus = models.FailedCommitStatus
	case vcs.CommitError:
		modelsStatus = models.FailedCommitStatus
	default:
		modelsStatus = models.PendingCommitStatus
	}

	return a.client.UpdateStatus(
		a.logger,
		modelsRepo,
		modelsPR,
		modelsStatus,
		status.Context,
		status.Description,
		status.TargetURL,
	)
}

// Capabilities returns the capabilities of this adapter
func (a *LegacyVCSClientAdapter) Capabilities() vcs.VCSCapabilities {
	switch a.vcsType {
	case "github":
		return vcs.VCSCapabilities{
			SupportsMergeableBypass: true,
			SupportsTeamAllowlist:   true,
			SupportsGroupAllowlist:  false,
			SupportsCustomFields:    false,
			MaxPageSize:            100,
		}
	case "gitlab":
		return vcs.VCSCapabilities{
			SupportsMergeableBypass: true,
			SupportsTeamAllowlist:   false,
			SupportsGroupAllowlist:  true,
			SupportsCustomFields:    false,
			MaxPageSize:            100,
		}
	case "azuredevops":
		return vcs.VCSCapabilities{
			SupportsMergeableBypass: false,
			SupportsTeamAllowlist:   false,
			SupportsGroupAllowlist:  false,
			SupportsCustomFields:    true,
			MaxPageSize:            100,
		}
	default:
		return vcs.VCSCapabilities{
			SupportsMergeableBypass: false,
			SupportsTeamAllowlist:   false,
			SupportsGroupAllowlist:  false,
			SupportsCustomFields:    false,
			MaxPageSize:            50,
		}
	}
}

// CheckMergeableBypass implements mergeable bypass check for GitHub
func (a *LegacyVCSClientAdapter) CheckMergeableBypass(ctx context.Context, pr *vcs.PullRequest) (bool, error) {
	if !a.config.AllowMergeableBypass {
		return false, nil
	}

	if a.vcsType != "github" {
		return false, fmt.Errorf("mergeable bypass only supported for GitHub")
	}

	// Convert to models for existing client
	modelsRepo := models.Repo{
		FullName: "owner/repo", // Would need actual repo info
		VCSHost: models.VCSHost{
			Type: models.Github,
		},
	}

	modelsPR := models.PullRequest{
		Num: pr.Number,
	}

	isMergeable, err := a.client.PullIsMergeable(
		a.logger,
		modelsRepo,
		modelsPR,
		"atlantis", // VCS status name
		a.config.IgnoreVCSStatusNames,
	)

	if err != nil {
		return false, fmt.Errorf("failed to check if PR is mergeable: %w", err)
	}

	return isMergeable, nil
}

// ValidateTeamMembership implements team membership validation
func (a *LegacyVCSClientAdapter) ValidateTeamMembership(ctx context.Context, user string, teams []string) (bool, error) {
	if len(a.config.TeamAllowlist) == 0 {
		return true, nil // No team restriction
	}

	// Convert to models for existing client
	modelsRepo := models.Repo{
		VCSHost: models.VCSHost{
			Type: a.getVCSHostType(),
		},
	}

	modelsUser := models.User{
		Username: user,
	}

	userTeams, err := a.client.GetTeamNamesForUser(a.logger, modelsRepo, modelsUser)
	if err != nil {
		return false, fmt.Errorf("failed to get user teams: %w", err)
	}

	// Check if user is in any of the allowed teams
	for _, userTeam := range userTeams {
		for _, allowedTeam := range a.config.TeamAllowlist {
			if userTeam == allowedTeam {
				return true, nil
			}
		}
	}

	return false, nil
}

// ValidateGroupMembership implements group membership validation (primarily for GitLab)
func (a *LegacyVCSClientAdapter) ValidateGroupMembership(ctx context.Context, user string, groups []string) (bool, error) {
	if len(a.config.GroupAllowlist) == 0 {
		return true, nil // No group restriction
	}

	if a.vcsType != "gitlab" {
		return false, fmt.Errorf("group membership validation only supported for GitLab")
	}

	// For GitLab, groups are similar to teams in the existing implementation
	return a.ValidateTeamMembership(ctx, user, groups)
}

// Helper methods
func (a *LegacyVCSClientAdapter) getVCSHostType() models.VCSHostType {
	switch a.vcsType {
	case "github":
		return models.Github
	case "gitlab":
		return models.Gitlab
	case "bitbucket":
		return models.BitbucketCloud
	case "azuredevops":
		return models.AzureDevops
	case "gitea":
		return models.Gitea
	default:
		return models.Github
	}
}

func (a *LegacyVCSClientAdapter) getHostname() string {
	switch a.vcsType {
	case "github":
		return "github.com"
	case "gitlab":
		return "gitlab.com"
	case "bitbucket":
		return "bitbucket.org"
	case "azuredevops":
		return "dev.azure.com"
	case "gitea":
		return "gitea.com"
	default:
		return "github.com"
	}
} 