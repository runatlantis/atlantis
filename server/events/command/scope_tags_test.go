package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProjectScopeTags_Loadtags_ExcludesBranchNames(t *testing.T) {
	tests := []struct {
		name     string
		tags     ProjectScopeTags
		expected map[string]string
	}{
		{
			name: "normal project tags",
			tags: ProjectScopeTags{
				BaseRepo:              "runatlantis/atlantis",
				PrNumber:              "123",
				Project:               "my-project",
				ProjectPath:           "terraform/prod",
				TerraformDistribution: "terraform",
				TerraformVersion:      "1.0.0",
				Workspace:             "default",
			},
			expected: map[string]string{
				"base_repo":              "runatlantis/atlantis",
				"pr_number":              "123",
				"project":                "my-project",
				"project_path":           "terraform/prod",
				"terraform_distribution": "terraform",
				"terraform_version":      "1.0.0",
				"workspace":              "default",
			},
		},
		{
			name: "project name with branch-like suffix is normalized",
			tags: ProjectScopeTags{
				BaseRepo:              "runatlantis/atlantis",
				PrNumber:              "123",
				Project:               "my-project-feature-branch",
				ProjectPath:           "terraform/prod",
				TerraformDistribution: "terraform",
				TerraformVersion:      "1.0.0",
				Workspace:             "default",
			},
			expected: map[string]string{
				"base_repo":              "runatlantis/atlantis",
				"pr_number":              "123",
				"project":                "my-project", // branch suffix should be removed
				"project_path":           "terraform/prod",
				"terraform_distribution": "terraform",
				"terraform_version":      "1.0.0",
				"workspace":              "default",
			},
		},
		{
			name: "project name with main branch suffix is normalized",
			tags: ProjectScopeTags{
				BaseRepo:              "runatlantis/atlantis",
				PrNumber:              "123",
				Project:               "my-project-main",
				ProjectPath:           "terraform/prod",
				TerraformDistribution: "terraform",
				TerraformVersion:      "1.0.0",
				Workspace:             "default",
			},
			expected: map[string]string{
				"base_repo":              "runatlantis/atlantis",
				"pr_number":              "123",
				"project":                "my-project", // branch suffix should be removed
				"project_path":           "terraform/prod",
				"terraform_distribution": "terraform",
				"terraform_version":      "1.0.0",
				"workspace":              "default",
			},
		},
		{
			name: "empty values are normalized to unknown",
			tags: ProjectScopeTags{
				BaseRepo:              "",
				PrNumber:              "",
				Project:               "",
				ProjectPath:           "",
				TerraformDistribution: "",
				TerraformVersion:      "",
				Workspace:             "",
			},
			expected: map[string]string{
				"base_repo":              "unknown",
				"pr_number":              "unknown",
				"project":                "unknown",
				"project_path":           "unknown",
				"terraform_distribution": "unknown",
				"terraform_version":      "unknown",
				"workspace":              "unknown",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.tags.Loadtags()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateTagValue_RemovesBranchSuffixes(t *testing.T) {
	tests := []struct {
		key      string
		value    string
		expected string
	}{
		{"project", "my-project-feature-branch", "my-project"},
		{"project", "my-project-main", "my-project"},
		{"project", "my-project-master", "my-project"},
		{"project", "my-project-develop", "my-project"},
		{"project", "my-project-staging", "my-project"},
		{"project", "my-project-prod", "my-project"},
		{"project", "my-project-hotfix-123", "my-project"},
		{"project", "my-project-release-v1.0", "my-project-release-v1.0"}, // not a branch pattern, should remain unchanged
		{"project", "my-project", "my-project"},                           // no suffix, should remain unchanged
		{"project", "my-project-123", "my-project-123"},                   // not a branch pattern
		{"base_repo", "runatlantis/atlantis", "runatlantis/atlantis"},     // not project key
		{"", "empty-key", "empty-key"},
		{"project", "", "unknown"},
		{"project", "   ", "unknown"},
		// Test edge cases
		{"project", "feature-branch", "default-project"}, // only branch name becomes default
		{"project", "main", "default-project"},           // only branch name becomes default
	}

	for _, tt := range tests {
		t.Run(tt.key+"_"+tt.value, func(t *testing.T) {
			result := validateTagValue(tt.key, tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}
