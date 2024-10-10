package server

import (
	"strings"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/logging"
)

// UserConfig holds config values passed in by the user.
// The mapstructure tags correspond to flags in cmd/server.go and are used when
// the config is parsed from a YAML file.
type UserConfig struct {
	AllowForkPRs                bool   `mapstructure:"allow-fork-prs"`
	AllowCommands               string `mapstructure:"allow-commands"`
	AtlantisURL                 string `mapstructure:"atlantis-url"`
	AutoDiscoverModeFlag        string `mapstructure:"autodiscover-mode"`
	Automerge                   bool   `mapstructure:"automerge"`
	AutoplanFileList            string `mapstructure:"autoplan-file-list"`
	AutoplanModules             bool   `mapstructure:"autoplan-modules"`
	AutoplanModulesFromProjects string `mapstructure:"autoplan-modules-from-projects"`
	AzureDevopsToken            string `mapstructure:"azuredevops-token"`
	AzureDevopsUser             string `mapstructure:"azuredevops-user"`
	AzureDevopsWebhookPassword  string `mapstructure:"azuredevops-webhook-password"`
	AzureDevopsWebhookUser      string `mapstructure:"azuredevops-webhook-user"`
	AzureDevOpsHostname         string `mapstructure:"azuredevops-hostname"`
	BitbucketBaseURL            string `mapstructure:"bitbucket-base-url"`
	BitbucketToken              string `mapstructure:"bitbucket-token"`
	BitbucketUser               string `mapstructure:"bitbucket-user"`
	BitbucketWebhookSecret      string `mapstructure:"bitbucket-webhook-secret"`
	CheckoutDepth               int    `mapstructure:"checkout-depth"`
	CheckoutStrategy            string `mapstructure:"checkout-strategy"`
	DataDir                     string `mapstructure:"data-dir"`
	DisableApplyAll             bool   `mapstructure:"disable-apply-all"`
	DisableAutoplan             bool   `mapstructure:"disable-autoplan"`
	DisableAutoplanLabel        string `mapstructure:"disable-autoplan-label"`
	DisableMarkdownFolding      bool   `mapstructure:"disable-markdown-folding"`
	DisableRepoLocking          bool   `mapstructure:"disable-repo-locking"`
	DisableGlobalApplyLock      bool   `mapstructure:"disable-global-apply-lock"`
	DisableUnlockLabel          string `mapstructure:"disable-unlock-label"`
	DiscardApprovalOnPlanFlag   bool   `mapstructure:"discard-approval-on-plan"`
	EmojiReaction               string `mapstructure:"emoji-reaction"`
	EnablePolicyChecksFlag      bool   `mapstructure:"enable-policy-checks"`
	EnableRegExpCmd             bool   `mapstructure:"enable-regexp-cmd"`
	EnableDiffMarkdownFormat    bool   `mapstructure:"enable-diff-markdown-format"`
	ExecutableName              string `mapstructure:"executable-name"`
	// Fail and do not run the Atlantis command request if any of the pre workflow hooks error.
	FailOnPreWorkflowHookError      bool   `mapstructure:"fail-on-pre-workflow-hook-error"`
	HideUnchangedPlanComments       bool   `mapstructure:"hide-unchanged-plan-comments"`
	GithubAllowMergeableBypassApply bool   `mapstructure:"gh-allow-mergeable-bypass-apply"`
	GithubHostname                  string `mapstructure:"gh-hostname"`
	GithubToken                     string `mapstructure:"gh-token"`
	GithubUser                      string `mapstructure:"gh-user"`
	GithubWebhookSecret             string `mapstructure:"gh-webhook-secret"`
	GithubOrg                       string `mapstructure:"gh-org"`
	GithubAppID                     int64  `mapstructure:"gh-app-id"`
	GithubAppKey                    string `mapstructure:"gh-app-key"`
	GithubAppKeyFile                string `mapstructure:"gh-app-key-file"`
	GithubAppSlug                   string `mapstructure:"gh-app-slug"`
	GithubAppInstallationID         int64  `mapstructure:"gh-app-installation-id"`
	GithubTeamAllowlist             string `mapstructure:"gh-team-allowlist"`
	GiteaBaseURL                    string `mapstructure:"gitea-base-url"`
	GiteaToken                      string `mapstructure:"gitea-token"`
	GiteaUser                       string `mapstructure:"gitea-user"`
	GiteaWebhookSecret              string `mapstructure:"gitea-webhook-secret"`
	GiteaPageSize                   int    `mapstructure:"gitea-page-size"`
	GitlabHostname                  string `mapstructure:"gitlab-hostname"`
	GitlabToken                     string `mapstructure:"gitlab-token"`
	GitlabUser                      string `mapstructure:"gitlab-user"`
	GitlabWebhookSecret             string `mapstructure:"gitlab-webhook-secret"`
	IncludeGitUntrackedFiles        bool   `mapstructure:"include-git-untracked-files"`
	APISecret                       string `mapstructure:"api-secret"`
	HidePrevPlanComments            bool   `mapstructure:"hide-prev-plan-comments"`
	LockingDBType                   string `mapstructure:"locking-db-type"`
	LogLevel                        string `mapstructure:"log-level"`
	MarkdownTemplateOverridesDir    string `mapstructure:"markdown-template-overrides-dir"`
	MaxCommentsPerCommand           int    `mapstructure:"max-comments-per-command"`
	ParallelPoolSize                int    `mapstructure:"parallel-pool-size"`
	ParallelPlan                    bool   `mapstructure:"parallel-plan"`
	ParallelApply                   bool   `mapstructure:"parallel-apply"`
	StatsNamespace                  string `mapstructure:"stats-namespace"`
	PlanDrafts                      bool   `mapstructure:"allow-draft-prs"`
	Port                            int    `mapstructure:"port"`
	QuietPolicyChecks               bool   `mapstructure:"quiet-policy-checks"`
	RedisDB                         int    `mapstructure:"redis-db"`
	RedisHost                       string `mapstructure:"redis-host"`
	RedisPassword                   string `mapstructure:"redis-password"`
	RedisPort                       int    `mapstructure:"redis-port"`
	RedisTLSEnabled                 bool   `mapstructure:"redis-tls-enabled"`
	RedisInsecureSkipVerify         bool   `mapstructure:"redis-insecure-skip-verify"`
	RepoConfig                      string `mapstructure:"repo-config"`
	RepoConfigJSON                  string `mapstructure:"repo-config-json"`
	RepoAllowlist                   string `mapstructure:"repo-allowlist"`

	// SilenceNoProjects is whether Atlantis should respond to a PR if no projects are found.
	SilenceNoProjects   bool `mapstructure:"silence-no-projects"`
	SilenceForkPRErrors bool `mapstructure:"silence-fork-pr-errors"`
	// SilenceVCSStatusNoPlans is whether autoplan should set commit status if no plans
	// are found.
	SilenceVCSStatusNoPlans bool `mapstructure:"silence-vcs-status-no-plans"`
	// SilenceVCSStatusNoProjects is whether autoplan should set commit status if no projects
	// are found.
	SilenceVCSStatusNoProjects bool            `mapstructure:"silence-vcs-status-no-projects"`
	SilenceAllowlistErrors     bool            `mapstructure:"silence-allowlist-errors"`
	SkipCloneNoChanges         bool            `mapstructure:"skip-clone-no-changes"`
	SlackToken                 string          `mapstructure:"slack-token"`
	SSLCertFile                string          `mapstructure:"ssl-cert-file"`
	SSLKeyFile                 string          `mapstructure:"ssl-key-file"`
	RestrictFileList           bool            `mapstructure:"restrict-file-list"`
	TFDistribution             string          `mapstructure:"tf-distribution"`
	TFDownload                 bool            `mapstructure:"tf-download"`
	TFDownloadURL              string          `mapstructure:"tf-download-url"`
	TFEHostname                string          `mapstructure:"tfe-hostname"`
	TFELocalExecutionMode      bool            `mapstructure:"tfe-local-execution-mode"`
	TFEToken                   string          `mapstructure:"tfe-token"`
	VarFileAllowlist           string          `mapstructure:"var-file-allowlist"`
	VCSStatusName              string          `mapstructure:"vcs-status-name"`
	DefaultTFVersion           string          `mapstructure:"default-tf-version"`
	Webhooks                   []WebhookConfig `mapstructure:"webhooks" flag:"false"`
	WebBasicAuth               bool            `mapstructure:"web-basic-auth"`
	WebUsername                string          `mapstructure:"web-username"`
	WebPassword                string          `mapstructure:"web-password"`
	WriteGitCreds              bool            `mapstructure:"write-git-creds"`
	WebsocketCheckOrigin       bool            `mapstructure:"websocket-check-origin"`
	UseTFPluginCache           bool            `mapstructure:"use-tf-plugin-cache"`
}

// ToAllowCommandNames parse AllowCommands into a slice of CommandName
func (u UserConfig) ToAllowCommandNames() ([]command.Name, error) {
	var allowCommands []command.Name
	var hasAll bool
	for _, input := range strings.Split(u.AllowCommands, ",") {
		if input == "" {
			continue
		}
		if input == "all" {
			hasAll = true
			continue
		}
		cmd, err := command.ParseCommandName(input)
		if err != nil {
			return nil, err
		}
		allowCommands = append(allowCommands, cmd)
	}
	if hasAll {
		return command.AllCommentCommands, nil
	}
	return allowCommands, nil
}

// ToLogLevel returns the LogLevel object corresponding to the user-passed
// log level.
func (u UserConfig) ToLogLevel() logging.LogLevel {
	switch u.LogLevel {
	case "debug":
		return logging.Debug
	case "info":
		return logging.Info
	case "warn":
		return logging.Warn
	case "error":
		return logging.Error
	}
	return logging.Info
}
