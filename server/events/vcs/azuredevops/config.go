// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package azuredevops

// Config allows for custom Azure DevOps-specific functionality and behavior.
type Config struct {
	// AllowMergeableBypassApply enables the functionality to allow the mergeable
	// check to ignore the apply required status check. When enabled, a pull request
	// can be considered mergeable even if the Atlantis apply status check is failing,
	// as long as all other branch policy requirements are satisfied.
	AllowMergeableBypassApply bool

	// BypassMergeRequirementTeams is a list of Azure DevOps security group names
	// (including Azure AD groups synced to Azure DevOps) that are allowed to merge PRs
	// when the apply status check is bypassed. If empty and AllowMergeableBypassApply
	// is enabled, any user can merge with bypass.
	// When set, only members of these groups can merge with bypass, and an audit
	// comment will be added to the PR documenting who performed the bypass merge.
	//
	// Example group names:
	//   - "GL-SGA-DevOps-EcomDevopsTeamMembers" (AAD group synced to Azure DevOps)
	//   - "[Project Name]\\Contributors" (built-in project group)
	//
	// Note: This uses the Azure DevOps Graph API to check group membership,
	// which requires the vso.graph scope for the configured token.
	BypassMergeRequirementTeams []string
}
