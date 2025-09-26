package vcs

// GithubConfig allows for custom github-specific functionality and behavior
type GithubConfig struct {
	AllowMergeableBypassApply bool
	// DisableRepoLevelSecurityRules disables querying repository rules which may not be supported on older GitHub Enterprise instances
	DisableRepoLevelSecurityRules bool
}
