package server

import "github.com/runatlantis/atlantis/server/logging"

// UserConfig holds config values passed in by the user.
// The mapstructure tags correspond to flags in cmd/server.go and are used when
// the config is parsed from a YAML file.
type UserConfig struct {
	AllowForkPRs           bool   `mapstructure:"allow-fork-prs"`
	AllowRepoConfig        bool   `mapstructure:"allow-repo-config"`
	AtlantisURL            string `mapstructure:"atlantis-url"`
	Automerge              bool   `mapstructure:"automerge"`
	BitbucketBaseURL       string `mapstructure:"bitbucket-base-url"`
	BitbucketToken         string `mapstructure:"bitbucket-token"`
	BitbucketUser          string `mapstructure:"bitbucket-user"`
	BitbucketWebhookSecret string `mapstructure:"bitbucket-webhook-secret"`
	CheckoutStrategy       string `mapstructure:"checkout-strategy"`
	DataDir                string `mapstructure:"data-dir"`
	DisableApplyAll        bool   `mapstructure:"disable-apply-all"`
	GithubHostname         string `mapstructure:"gh-hostname"`
	GithubToken            string `mapstructure:"gh-token"`
	GithubUser             string `mapstructure:"gh-user"`
	GithubWebhookSecret    string `mapstructure:"gh-webhook-secret"`
	GitlabHostname         string `mapstructure:"gitlab-hostname"`
	GitlabToken            string `mapstructure:"gitlab-token"`
	GitlabUser             string `mapstructure:"gitlab-user"`
	GitlabWebhookSecret    string `mapstructure:"gitlab-webhook-secret"`
	LogLevel               string `mapstructure:"log-level"`
	Port                   int    `mapstructure:"port"`
	RepoConfig             string `mapstructure:"repo-config"`
	RepoConfigJSON         string `mapstructure:"repo-config-json"`
	RepoWhitelist          string `mapstructure:"repo-whitelist"`
	// RequireApproval is whether to require pull request approval before
	// allowing terraform apply's to be run.
	RequireApproval bool `mapstructure:"require-approval"`
	// RequireMergeable is whether to require pull requests to be mergeable before
	// allowing terraform apply's to run.
	RequireMergeable       bool            `mapstructure:"require-mergeable"`
	SilenceWhitelistErrors bool            `mapstructure:"silence-whitelist-errors"`
	SlackToken             string          `mapstructure:"slack-token"`
	SSLCertFile            string          `mapstructure:"ssl-cert-file"`
	SSLKeyFile             string          `mapstructure:"ssl-key-file"`
	TFEToken               string          `mapstructure:"tfe-token"`
	DefaultTFVersion       string          `mapstructure:"default-tf-version"`
	Webhooks               []WebhookConfig `mapstructure:"webhooks"`
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
