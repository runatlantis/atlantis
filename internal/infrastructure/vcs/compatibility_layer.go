package vcs

import (
	"context"
	"fmt"

	"github.com/runatlantis/atlantis/internal/domain/vcs"
	"github.com/runatlantis/atlantis/server"
	"github.com/runatlantis/atlantis/server/events/models"
	vcsClient "github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
)

// CompatibilityLayer provides backward compatibility during VCS plugin migration
type CompatibilityLayer struct {
	// Legacy VCS client proxy (existing system)
	legacyClient vcsClient.Client
	
	// New plugin registry
	pluginRegistry vcs.VCSRegistry
	
	// Configuration
	userConfig server.UserConfig
	logger     logging.SimpleLogging
	
	// Migration settings
	enableNewSystem bool
	fallbackToLegacy bool
}

// CompatibilityConfig holds configuration for the compatibility layer
type CompatibilityConfig struct {
	// EnableNewSystem enables the new plugin system
	EnableNewSystem bool
	
	// FallbackToLegacy falls back to legacy system if plugin fails
	FallbackToLegacy bool
	
	// LogMigrationWarnings logs warnings about deprecated features
	LogMigrationWarnings bool
}

// NewCompatibilityLayer creates a new backward compatibility layer
func NewCompatibilityLayer(
	legacyClient vcsClient.Client,
	pluginRegistry vcs.VCSRegistry,
	userConfig server.UserConfig,
	logger logging.SimpleLogging,
	config CompatibilityConfig,
) *CompatibilityLayer {
	return &CompatibilityLayer{
		legacyClient:     legacyClient,
		pluginRegistry:   pluginRegistry,
		userConfig:       userConfig,
		logger:           logger,
		enableNewSystem:  config.EnableNewSystem,
		fallbackToLegacy: config.FallbackToLegacy,
	}
}

// GetModifiedFiles retrieves modified files with compatibility handling
func (c *CompatibilityLayer) GetModifiedFiles(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest) ([]string, error) {
	if c.enableNewSystem {
		// Try new plugin system first
		plugin, err := c.getVCSPlugin(repo.VCSHost.Type)
		if err == nil {
			// Convert models to new types
			vcsRepo := c.convertToVCSRepository(repo)
			vcsFiles, err := c.getModifiedFilesFromPlugin(plugin, vcsRepo, pull.Num)
			if err == nil {
				c.logger.Debug("Successfully used new VCS plugin for GetModifiedFiles")
				return vcsFiles, nil
			}
			
			if !c.fallbackToLegacy {
				return nil, fmt.Errorf("new VCS plugin failed and fallback disabled: %w", err)
			}
			
			c.logger.Warn("New VCS plugin failed, falling back to legacy client: %v", err)
		}
	}
	
	// Use legacy system
	c.logDeprecationWarning("GetModifiedFiles", "Use VCS plugin system instead")
	return c.legacyClient.GetModifiedFiles(logger, repo, pull)
}

// CreateComment creates a comment with compatibility handling
func (c *CompatibilityLayer) CreateComment(logger logging.SimpleLogging, repo models.Repo, pullNum int, comment string, command string) error {
	if c.enableNewSystem {
		// Try new plugin system first
		_, err := c.getVCSPlugin(repo.VCSHost.Type)
		if err == nil {
			// For now, delegate to legacy since comment creation isn't in the plugin interface yet
			// In a full implementation, this would use the plugin
			c.logger.Debug("Comment creation not yet implemented in plugin system, using legacy")
		}
	}
	
	// Use legacy system
	c.logDeprecationWarning("CreateComment", "VCS plugin comment creation coming in future release")
	return c.legacyClient.CreateComment(logger, repo, pullNum, comment, command)
}

// PullIsMergeable checks if pull request is mergeable with compatibility handling
func (c *CompatibilityLayer) PullIsMergeable(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest, vcsstatusname string, ignoreVCSStatusNames []string) (bool, error) {
	if c.enableNewSystem {
		// Try new plugin system first
		plugin, err := c.getVCSPlugin(repo.VCSHost.Type)
		if err == nil {
			// Check if plugin supports mergeable bypass
			capabilities := plugin.Capabilities()
			if capabilities.SupportsMergeableBypass {
				vcsRepo := c.convertToVCSRepository(repo)
				vcsPR, err := plugin.GetPullRequest(context.Background(), vcsRepo, pull.Num)
				if err == nil {
					// Use plugin's mergeable check
					if adapter, ok := plugin.(*LegacyVCSClientAdapter); ok {
						isMergeable, err := adapter.CheckMergeableBypass(context.Background(), vcsPR)
						if err == nil {
							c.logger.Debug("Successfully used new VCS plugin for PullIsMergeable")
							return isMergeable, nil
						}
					}
				}
			}
			
			if !c.fallbackToLegacy {
				return false, fmt.Errorf("new VCS plugin failed and fallback disabled: %w", err)
			}
			
			c.logger.Warn("New VCS plugin failed, falling back to legacy client: %v", err)
		}
	}
	
	// Use legacy system
	c.logDeprecationWarning("PullIsMergeable", "Use VCS plugin CheckMergeableBypass method")
	return c.legacyClient.PullIsMergeable(logger, repo, pull, vcsstatusname, ignoreVCSStatusNames)
}

// UpdateStatus updates commit status with compatibility handling
func (c *CompatibilityLayer) UpdateStatus(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest, state models.CommitStatus, src string, description string, url string) error {
	if c.enableNewSystem {
		// Try new plugin system first
		plugin, err := c.getVCSPlugin(repo.VCSHost.Type)
		if err == nil {
			// Convert to new types
			vcsRepo := c.convertToVCSRepository(repo)
			vcsStatus := vcs.CommitStatus{
				State:       c.convertCommitState(state),
				Description: description,
				Context:     src,
				TargetURL:   url,
			}
			
			err = plugin.CreateCommitStatus(context.Background(), vcsRepo, pull.HeadCommit, vcsStatus)
			if err == nil {
				c.logger.Debug("Successfully used new VCS plugin for UpdateStatus")
				return nil
			}
			
			if !c.fallbackToLegacy {
				return fmt.Errorf("new VCS plugin failed and fallback disabled: %w", err)
			}
			
			c.logger.Warn("New VCS plugin failed, falling back to legacy client: %v", err)
		}
	}
	
	// Use legacy system
	c.logDeprecationWarning("UpdateStatus", "Use VCS plugin CreateCommitStatus method")
	return c.legacyClient.UpdateStatus(logger, repo, pull, state, src, description, url)
}

// GetTeamNamesForUser gets team names with compatibility handling
func (c *CompatibilityLayer) GetTeamNamesForUser(logger logging.SimpleLogging, repo models.Repo, user models.User) ([]string, error) {
	if c.enableNewSystem {
		// Try new plugin system first
		plugin, err := c.getVCSPlugin(repo.VCSHost.Type)
		if err == nil {
			// Check if plugin supports team validation
			capabilities := plugin.Capabilities()
			if capabilities.SupportsTeamAllowlist {
				if adapter, ok := plugin.(*LegacyVCSClientAdapter); ok {
					// For now, delegate to legacy implementation within adapter
					// In a full implementation, this would be a proper plugin method
					_, err := adapter.ValidateTeamMembership(context.Background(), user.Username, []string{})
					if err == nil {
						c.logger.Debug("Successfully used new VCS plugin for team validation")
						// Return empty list for now since we only validated membership
						return []string{}, nil
					}
				}
			}
			
			if !c.fallbackToLegacy {
				return nil, fmt.Errorf("new VCS plugin failed and fallback disabled: %w", err)
			}
			
			c.logger.Warn("New VCS plugin failed, falling back to legacy client: %v", err)
		}
	}
	
	// Use legacy system
	c.logDeprecationWarning("GetTeamNamesForUser", "Use VCS plugin ValidateTeamMembership method")
	return c.legacyClient.GetTeamNamesForUser(logger, repo, user)
}

// Helper methods

func (c *CompatibilityLayer) getVCSPlugin(vcsHostType models.VCSHostType) (vcs.VCSPlugin, error) {
	vcsType := c.convertVCSHostType(vcsHostType)
	return c.pluginRegistry.Get(vcsType)
}

func (c *CompatibilityLayer) convertVCSHostType(hostType models.VCSHostType) string {
	switch hostType {
	case models.Github:
		return "github"
	case models.Gitlab:
		return "gitlab"
	case models.BitbucketCloud, models.BitbucketServer:
		return "bitbucket"
	case models.AzureDevops:
		return "azuredevops"
	case models.Gitea:
		return "gitea"
	default:
		return "unknown"
	}
}

func (c *CompatibilityLayer) convertToVCSRepository(repo models.Repo) vcs.Repository {
	return vcs.Repository{
		FullName:      repo.FullName,
		Owner:         repo.Owner,
		Name:          repo.Name,
		HTMLURL:       fmt.Sprintf("https://%s/%s", repo.VCSHost.Hostname, repo.FullName),
		CloneURL:      repo.CloneURL,
		DefaultBranch: "main", // Default assumption
	}
}

func (c *CompatibilityLayer) convertCommitState(state models.CommitStatus) vcs.CommitState {
	switch state {
	case models.PendingCommitStatus:
		return vcs.CommitPending
	case models.SuccessCommitStatus:
		return vcs.CommitSuccess
	case models.FailedCommitStatus:
		return vcs.CommitFailure
	default:
		return vcs.CommitPending
	}
}

func (c *CompatibilityLayer) getModifiedFilesFromPlugin(plugin vcs.VCSPlugin, repo vcs.Repository, pullNum int) ([]string, error) {
	// This would be implemented when the plugin interface includes file operations
	// For now, return an error to trigger fallback
	return nil, fmt.Errorf("file operations not yet implemented in plugin interface")
}

func (c *CompatibilityLayer) logDeprecationWarning(method, suggestion string) {
	c.logger.Warn("DEPRECATION WARNING: %s is using legacy VCS client. %s. See GitHub issue #5574 for migration guide.", method, suggestion)
}

// Delegate remaining methods to legacy client for now
// These would be gradually migrated to the plugin system

func (c *CompatibilityLayer) ReactToComment(logger logging.SimpleLogging, repo models.Repo, pullNum int, commentID int64, reaction string) error {
	c.logDeprecationWarning("ReactToComment", "VCS plugin reaction support coming in future release")
	return c.legacyClient.ReactToComment(logger, repo, pullNum, commentID, reaction)
}

func (c *CompatibilityLayer) HidePrevCommandComments(logger logging.SimpleLogging, repo models.Repo, pullNum int, command string, dir string) error {
	c.logDeprecationWarning("HidePrevCommandComments", "VCS plugin comment management coming in future release")
	return c.legacyClient.HidePrevCommandComments(logger, repo, pullNum, command, dir)
}

func (c *CompatibilityLayer) PullIsApproved(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest) (models.ApprovalStatus, error) {
	c.logDeprecationWarning("PullIsApproved", "VCS plugin approval checking coming in future release")
	return c.legacyClient.PullIsApproved(logger, repo, pull)
}

func (c *CompatibilityLayer) DiscardReviews(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest) error {
	c.logDeprecationWarning("DiscardReviews", "VCS plugin review management coming in future release")
	return c.legacyClient.DiscardReviews(logger, repo, pull)
}

func (c *CompatibilityLayer) MergePull(logger logging.SimpleLogging, pull models.PullRequest, pullOptions models.PullRequestOptions) error {
	c.logDeprecationWarning("MergePull", "VCS plugin merge operations coming in future release")
	return c.legacyClient.MergePull(logger, pull, pullOptions)
}

func (c *CompatibilityLayer) MarkdownPullLink(pull models.PullRequest) (string, error) {
	c.logDeprecationWarning("MarkdownPullLink", "VCS plugin link generation coming in future release")
	return c.legacyClient.MarkdownPullLink(pull)
}

func (c *CompatibilityLayer) SupportsSingleFileDownload(repo models.Repo) bool {
	c.logDeprecationWarning("SupportsSingleFileDownload", "VCS plugin file operations coming in future release")
	return c.legacyClient.SupportsSingleFileDownload(repo)
}

func (c *CompatibilityLayer) GetFileContent(logger logging.SimpleLogging, pull models.PullRequest, fileName string) (bool, []byte, error) {
	c.logDeprecationWarning("GetFileContent", "VCS plugin file operations coming in future release")
	return c.legacyClient.GetFileContent(logger, pull, fileName)
}

func (c *CompatibilityLayer) GetCloneURL(logger logging.SimpleLogging, vcsHostType models.VCSHostType, repo string) (string, error) {
	c.logDeprecationWarning("GetCloneURL", "VCS plugin repository operations coming in future release")
	return c.legacyClient.GetCloneURL(logger, vcsHostType, repo)
}

func (c *CompatibilityLayer) GetPullLabels(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest) ([]string, error) {
	c.logDeprecationWarning("GetPullLabels", "VCS plugin label operations coming in future release")
	return c.legacyClient.GetPullLabels(logger, repo, pull)
} 