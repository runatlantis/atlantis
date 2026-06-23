package vcs

import (
	"context"
	"fmt"
)

// MockVCSPlugin is a test implementation of VCSPlugin
type MockVCSPlugin struct {
	capabilities        VCSCapabilities
	repositories        map[string]*Repository
	pullRequests        map[string]*PullRequest
	commitStatuses      []CommitStatus
	teamMemberships     map[string][]string
	groupMemberships    map[string][]string
	mergeableBypassFunc func(*PullRequest) bool
}

func NewMockVCSPlugin(caps VCSCapabilities) *MockVCSPlugin {
	return &MockVCSPlugin{
		capabilities:     caps,
		repositories:     make(map[string]*Repository),
		pullRequests:     make(map[string]*PullRequest),
		teamMemberships:  make(map[string][]string),
		groupMemberships: make(map[string][]string),
	}
}

func (m *MockVCSPlugin) GetRepository(ctx context.Context, owner, name string) (*Repository, error) {
	key := owner + "/" + name
	if repo, exists := m.repositories[key]; exists {
		return repo, nil
	}
	return nil, fmt.Errorf("repository %s not found", key)
}

func (m *MockVCSPlugin) GetPullRequest(ctx context.Context, repo Repository, number int) (*PullRequest, error) {
	key := fmt.Sprintf("%s#%d", repo.FullName, number)
	if pr, exists := m.pullRequests[key]; exists {
		return pr, nil
	}
	return nil, fmt.Errorf("pull request %s not found", key)
}

func (m *MockVCSPlugin) CreateCommitStatus(ctx context.Context, repo Repository, sha string, status CommitStatus) error {
	m.commitStatuses = append(m.commitStatuses, status)
	return nil
}

func (m *MockVCSPlugin) Capabilities() VCSCapabilities {
	return m.capabilities
}

func (m *MockVCSPlugin) CheckMergeableBypass(ctx context.Context, pr *PullRequest) (bool, error) {
	if !m.capabilities.SupportsMergeableBypass {
		return false, NewUnsupportedFeatureError("mock", "mergeable-bypass")
	}
	if m.mergeableBypassFunc != nil {
		return m.mergeableBypassFunc(pr), nil
	}
	return pr.Mergeable != nil && *pr.Mergeable, nil
}

func (m *MockVCSPlugin) ValidateTeamMembership(ctx context.Context, user string, teams []string) (bool, error) {
	if !m.capabilities.SupportsTeamAllowlist {
		return false, NewUnsupportedFeatureError("mock", "team-allowlist")
	}
	
	userTeams, exists := m.teamMemberships[user]
	if !exists {
		return false, nil
	}
	
	for _, team := range teams {
		for _, userTeam := range userTeams {
			if team == userTeam {
				return true, nil
			}
		}
	}
	return false, nil
}

func (m *MockVCSPlugin) ValidateGroupMembership(ctx context.Context, user string, groups []string) (bool, error) {
	if !m.capabilities.SupportsGroupAllowlist {
		return false, NewUnsupportedFeatureError("mock", "group-allowlist")
	}
	
	userGroups, exists := m.groupMemberships[user]
	if !exists {
		return false, nil
	}
	
	for _, group := range groups {
		for _, userGroup := range userGroups {
			if group == userGroup {
				return true, nil
			}
		}
	}
	return false, nil
}

// Test helpers
func (m *MockVCSPlugin) AddRepository(repo *Repository) {
	m.repositories[repo.FullName] = repo
}

func (m *MockVCSPlugin) AddPullRequest(repo Repository, pr *PullRequest) {
	key := fmt.Sprintf("%s#%d", repo.FullName, pr.Number)
	m.pullRequests[key] = pr
}

func (m *MockVCSPlugin) AddTeamMembership(user string, teams []string) {
	m.teamMemberships[user] = teams
}

func (m *MockVCSPlugin) AddGroupMembership(user string, groups []string) {
	m.groupMemberships[user] = groups
}

func (m *MockVCSPlugin) SetMergeableBypassFunc(fn func(*PullRequest) bool) {
	m.mergeableBypassFunc = fn
}

func (m *MockVCSPlugin) GetCommitStatuses() []CommitStatus {
	return m.commitStatuses
} 