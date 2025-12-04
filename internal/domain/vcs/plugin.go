package vcs

import (
	"context"
	"fmt"
)

// VCSPlugin defines the interface for VCS providers
type VCSPlugin interface {
	// Basic VCS operations
	GetRepository(ctx context.Context, owner, name string) (*Repository, error)
	GetPullRequest(ctx context.Context, repo Repository, number int) (*PullRequest, error)
	CreateCommitStatus(ctx context.Context, repo Repository, sha string, status CommitStatus) error
	
	// Capability detection
	Capabilities() VCSCapabilities
	
	// Feature implementations (only called if capability is supported)
	CheckMergeableBypass(ctx context.Context, pr *PullRequest) (bool, error)
	ValidateTeamMembership(ctx context.Context, user string, teams []string) (bool, error)
	ValidateGroupMembership(ctx context.Context, user string, groups []string) (bool, error)
}

// VCSCapabilities defines what features each VCS provider supports
type VCSCapabilities struct {
	SupportsMergeableBypass bool
	SupportsTeamAllowlist   bool
	SupportsGroupAllowlist  bool
	SupportsCustomFields    bool
	MaxPageSize            int
}

// FeatureConfig holds feature settings independent of VCS
type FeatureConfig struct {
	// General feature flags
	AllowMergeableBypass bool
	TeamAllowlist        []string
	GroupAllowlist       []string
	
	// VCS-specific configuration (not features)
	GitHubHostname    string
	GitLabURL         string
	AzureDevOpsOrg    string
}

// VCSRegistry manages available VCS plugins
type VCSRegistry interface {
	Register(name string, plugin VCSPlugin) error
	Get(name string) (VCSPlugin, error)
	List() []string
}

// FeatureValidator ensures features are only used where supported
type FeatureValidator struct {
	registry VCSRegistry
}

func (fv *FeatureValidator) ValidateFeatures(vcsType string, config FeatureConfig) error {
	plugin, err := fv.registry.Get(vcsType)
	if err != nil {
		return err
	}
	
	capabilities := plugin.Capabilities()
	
	if config.AllowMergeableBypass && !capabilities.SupportsMergeableBypass {
		return NewUnsupportedFeatureError(vcsType, "mergeable-bypass")
	}
	
	if len(config.TeamAllowlist) > 0 && !capabilities.SupportsTeamAllowlist {
		return NewUnsupportedFeatureError(vcsType, "team-allowlist")
	}
	
	if len(config.GroupAllowlist) > 0 && !capabilities.SupportsGroupAllowlist {
		return NewUnsupportedFeatureError(vcsType, "group-allowlist")
	}
	
	return nil
}

type UnsupportedFeatureError struct {
	VCS     string
	Feature string
}

func (e *UnsupportedFeatureError) Error() string {
	return fmt.Sprintf("feature '%s' is not supported by VCS provider '%s'", e.Feature, e.VCS)
}

func NewUnsupportedFeatureError(vcs, feature string) *UnsupportedFeatureError {
	return &UnsupportedFeatureError{VCS: vcs, Feature: feature}
} 