package vcs

import (
	"fmt"
	"strings"

	"github.com/runatlantis/atlantis/internal/domain/vcs"
	"github.com/runatlantis/atlantis/server"
	"github.com/runatlantis/atlantis/server/events/models"
	vcsClient "github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
)

// VCSConfigManager manages VCS configuration and creates appropriate plugins
type VCSConfigManager struct {
	registry    vcs.VCSRegistry
	logger      logging.SimpleLogging
	userConfig  server.UserConfig
}

// NewVCSConfigManager creates a new VCS configuration manager
func NewVCSConfigManager(registry vcs.VCSRegistry, logger logging.SimpleLogging, userConfig server.UserConfig) *VCSConfigManager {
	return &VCSConfigManager{
		registry:   registry,
		logger:     logger,
		userConfig: userConfig,
	}
}

// RegisterLegacyVCSClients registers existing VCS clients as plugins
func (m *VCSConfigManager) RegisterLegacyVCSClients(clientProxy *vcsClient.ClientProxy) error {
	// Register GitHub if configured
	if m.userConfig.GithubUser != "" {
		githubClient, err := m.extractGitHubClient(clientProxy)
		if err != nil {
			return fmt.Errorf("failed to extract GitHub client: %w", err)
		}

		config := VCSAdapterConfig{
			AllowMergeableBypass: m.userConfig.GithubAllowMergeableBypassApply,
			TeamAllowlist:        m.parseAllowlist(m.userConfig.GithubTeamAllowlist),
			IgnoreVCSStatusNames: m.parseIgnoreStatusNames(m.userConfig.IgnoreVCSStatusNames),
		}

		adapter := NewLegacyVCSClientAdapter(githubClient, "github", m.logger, config)
		err = m.registry.Register("github", adapter)
		if err != nil {
			return fmt.Errorf("failed to register GitHub adapter: %w", err)
		}
		m.logger.Info("Registered GitHub VCS plugin via legacy adapter")
	}

	// Register GitLab if configured
	if m.userConfig.GitlabUser != "" {
		gitlabClient, err := m.extractGitLabClient(clientProxy)
		if err != nil {
			return fmt.Errorf("failed to extract GitLab client: %w", err)
		}

		config := VCSAdapterConfig{
			AllowMergeableBypass: false, // GitLab doesn't support this feature
			GroupAllowlist:       m.parseAllowlist(m.userConfig.GitlabGroupAllowlist),
			IgnoreVCSStatusNames: m.parseIgnoreStatusNames(m.userConfig.IgnoreVCSStatusNames),
		}

		adapter := NewLegacyVCSClientAdapter(gitlabClient, "gitlab", m.logger, config)
		err = m.registry.Register("gitlab", adapter)
		if err != nil {
			return fmt.Errorf("failed to register GitLab adapter: %w", err)
		}
		m.logger.Info("Registered GitLab VCS plugin via legacy adapter")
	}

	// Register Azure DevOps if configured
	if m.userConfig.AzureDevopsUser != "" {
		azureClient, err := m.extractAzureDevOpsClient(clientProxy)
		if err != nil {
			return fmt.Errorf("failed to extract Azure DevOps client: %w", err)
		}

		config := VCSAdapterConfig{
			AllowMergeableBypass: false, // Azure DevOps doesn't support this feature
			IgnoreVCSStatusNames: m.parseIgnoreStatusNames(m.userConfig.IgnoreVCSStatusNames),
		}

		adapter := NewLegacyVCSClientAdapter(azureClient, "azuredevops", m.logger, config)
		err = m.registry.Register("azuredevops", adapter)
		if err != nil {
			return fmt.Errorf("failed to register Azure DevOps adapter: %w", err)
		}
		m.logger.Info("Registered Azure DevOps VCS plugin via legacy adapter")
	}

	// Register Bitbucket if configured
	if m.userConfig.BitbucketUser != "" {
		bitbucketClient, err := m.extractBitbucketClient(clientProxy)
		if err != nil {
			return fmt.Errorf("failed to extract Bitbucket client: %w", err)
		}

		config := VCSAdapterConfig{
			AllowMergeableBypass: false, // Bitbucket doesn't support this feature
			IgnoreVCSStatusNames: m.parseIgnoreStatusNames(m.userConfig.IgnoreVCSStatusNames),
		}

		adapter := NewLegacyVCSClientAdapter(bitbucketClient, "bitbucket", m.logger, config)
		err = m.registry.Register("bitbucket", adapter)
		if err != nil {
			return fmt.Errorf("failed to register Bitbucket adapter: %w", err)
		}
		m.logger.Info("Registered Bitbucket VCS plugin via legacy adapter")
	}

	// Register Gitea if configured
	if m.userConfig.GiteaUser != "" {
		giteaClient, err := m.extractGiteaClient(clientProxy)
		if err != nil {
			return fmt.Errorf("failed to extract Gitea client: %w", err)
		}

		config := VCSAdapterConfig{
			AllowMergeableBypass: false, // Gitea doesn't support this feature
			IgnoreVCSStatusNames: m.parseIgnoreStatusNames(m.userConfig.IgnoreVCSStatusNames),
		}

		adapter := NewLegacyVCSClientAdapter(giteaClient, "gitea", m.logger, config)
		err = m.registry.Register("gitea", adapter)
		if err != nil {
			return fmt.Errorf("failed to register Gitea adapter: %w", err)
		}
		m.logger.Info("Registered Gitea VCS plugin via legacy adapter")
	}

	return nil
}

// GetVCSPlugin returns the appropriate VCS plugin for the given VCS type
func (m *VCSConfigManager) GetVCSPlugin(vcsType string) (vcs.VCSPlugin, error) {
	plugin, err := m.registry.Get(vcsType)
	if err != nil {
		return nil, fmt.Errorf("VCS type %s not configured: %w", vcsType, err)
	}
	return plugin, nil
}

// ValidateVCSFeatures validates the given VCS configuration against capabilities
func (m *VCSConfigManager) ValidateVCSFeatures(vcsType string) error {
	validator := vcs.NewFeatureValidator(m.registry)
	
	config := vcs.FeatureConfig{
		AllowMergeableBypass: m.getVCSMergeableBypass(vcsType),
		TeamAllowlist:        m.getVCSTeamAllowlist(vcsType),
		GroupAllowlist:       m.getVCSGroupAllowlist(vcsType),
	}

	return validator.ValidateAndWarn(vcsType, config)
}

// Helper methods to extract configuration
func (m *VCSConfigManager) getVCSMergeableBypass(vcsType string) bool {
	switch vcsType {
	case "github":
		return m.userConfig.GithubAllowMergeableBypassApply
	default:
		return false
	}
}

func (m *VCSConfigManager) getVCSTeamAllowlist(vcsType string) []string {
	switch vcsType {
	case "github":
		return m.parseAllowlist(m.userConfig.GithubTeamAllowlist)
	default:
		return nil
	}
}

func (m *VCSConfigManager) getVCSGroupAllowlist(vcsType string) []string {
	switch vcsType {
	case "gitlab":
		return m.parseAllowlist(m.userConfig.GitlabGroupAllowlist)
	default:
		return nil
	}
}

func (m *VCSConfigManager) parseAllowlist(allowlist string) []string {
	if allowlist == "" {
		return nil
	}
	
	items := strings.Split(allowlist, ",")
	var result []string
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func (m *VCSConfigManager) parseIgnoreStatusNames(ignoreNames string) []string {
	if ignoreNames == "" {
		return nil
	}
	
	items := strings.Split(ignoreNames, ",")
	var result []string
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// Temporary helper methods to extract clients from proxy
// These use reflection or type assertions to get the underlying clients
// In a real implementation, these would be more robust

func (m *VCSConfigManager) extractGitHubClient(proxy *vcsClient.ClientProxy) (vcsClient.Client, error) {
	// For now, return a placeholder. In real implementation, would extract from proxy
	return &vcsClient.NotConfiguredVCSClient{Host: models.Github}, nil
}

func (m *VCSConfigManager) extractGitLabClient(proxy *vcsClient.ClientProxy) (vcsClient.Client, error) {
	return &vcsClient.NotConfiguredVCSClient{Host: models.Gitlab}, nil
}

func (m *VCSConfigManager) extractAzureDevOpsClient(proxy *vcsClient.ClientProxy) (vcsClient.Client, error) {
	return &vcsClient.NotConfiguredVCSClient{Host: models.AzureDevops}, nil
}

func (m *VCSConfigManager) extractBitbucketClient(proxy *vcsClient.ClientProxy) (vcsClient.Client, error) {
	return &vcsClient.NotConfiguredVCSClient{Host: models.BitbucketCloud}, nil
}

func (m *VCSConfigManager) extractGiteaClient(proxy *vcsClient.ClientProxy) (vcsClient.Client, error) {
	return &vcsClient.NotConfiguredVCSClient{Host: models.Gitea}, nil
} 