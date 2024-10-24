// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.

package cmd

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/moby/patternmatcher"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/runatlantis/atlantis/server"
	"github.com/runatlantis/atlantis/server/events/vcs/bitbucketcloud"
	"github.com/runatlantis/atlantis/server/logging"
)

// checkout strategies
const (
	CheckoutStrategyBranch = "branch"
	CheckoutStrategyMerge  = "merge"
)

// TF distributions
const (
	TFDistributionTerraform = "terraform"
	TFDistributionOpenTofu  = "opentofu"
)

// To add a new flag you must:
// 1. Add a const with the flag name (in alphabetic order).
// 2. Add a new field to server.UserConfig and set the mapstructure tag equal to the flag name.
// 3. Add your flag's description etc. to the stringFlags, intFlags, or boolFlags slices.
const (
	// Flag names.
	ADWebhookPasswordFlag            = "azuredevops-webhook-password" // nolint: gosec
	ADWebhookUserFlag                = "azuredevops-webhook-user"
	ADTokenFlag                      = "azuredevops-token" // nolint: gosec
	ADUserFlag                       = "azuredevops-user"
	ADHostnameFlag                   = "azuredevops-hostname"
	AllowCommandsFlag                = "allow-commands"
	AllowForkPRsFlag                 = "allow-fork-prs"
	AtlantisURLFlag                  = "atlantis-url"
	AutoDiscoverModeFlag             = "autodiscover-mode"
	AutomergeFlag                    = "automerge"
	ParallelPlanFlag                 = "parallel-plan"
	ParallelApplyFlag                = "parallel-apply"
	AutoplanModules                  = "autoplan-modules"
	AutoplanModulesFromProjects      = "autoplan-modules-from-projects"
	AutoplanFileListFlag             = "autoplan-file-list"
	BitbucketBaseURLFlag             = "bitbucket-base-url"
	BitbucketTokenFlag               = "bitbucket-token"
	BitbucketUserFlag                = "bitbucket-user"
	BitbucketWebhookSecretFlag       = "bitbucket-webhook-secret"
	CheckoutDepthFlag                = "checkout-depth"
	CheckoutStrategyFlag             = "checkout-strategy"
	ConfigFlag                       = "config"
	DataDirFlag                      = "data-dir"
	DefaultTFVersionFlag             = "default-tf-version"
	DisableApplyAllFlag              = "disable-apply-all"
	DisableAutoplanFlag              = "disable-autoplan"
	DisableAutoplanLabelFlag         = "disable-autoplan-label"
	DisableMarkdownFoldingFlag       = "disable-markdown-folding"
	DisableRepoLockingFlag           = "disable-repo-locking"
	DisableGlobalApplyLockFlag       = "disable-global-apply-lock"
	DisableUnlockLabelFlag           = "disable-unlock-label"
	DiscardApprovalOnPlanFlag        = "discard-approval-on-plan"
	EmojiReaction                    = "emoji-reaction"
	EnableDiffMarkdownFormat         = "enable-diff-markdown-format"
	EnablePolicyChecksFlag           = "enable-policy-checks"
	EnableRegExpCmdFlag              = "enable-regexp-cmd"
	ExecutableName                   = "executable-name"
	FailOnPreWorkflowHookError       = "fail-on-pre-workflow-hook-error"
	HideUnchangedPlanComments        = "hide-unchanged-plan-comments"
	GHHostnameFlag                   = "gh-hostname"
	GHTeamAllowlistFlag              = "gh-team-allowlist"
	GHTokenFlag                      = "gh-token"
	GHUserFlag                       = "gh-user"
	GHAppIDFlag                      = "gh-app-id"
	GHAppKeyFlag                     = "gh-app-key"
	GHAppKeyFileFlag                 = "gh-app-key-file"
	GHAppSlugFlag                    = "gh-app-slug"
	GHAppInstallationIDFlag          = "gh-app-installation-id"
	GHOrganizationFlag               = "gh-org"
	GHWebhookSecretFlag              = "gh-webhook-secret"               // nolint: gosec
	GHAllowMergeableBypassApply      = "gh-allow-mergeable-bypass-apply" // nolint: gosec
	GiteaBaseURLFlag                 = "gitea-base-url"
	GiteaTokenFlag                   = "gitea-token"
	GiteaUserFlag                    = "gitea-user"
	GiteaWebhookSecretFlag           = "gitea-webhook-secret" // nolint: gosec
	GiteaPageSizeFlag                = "gitea-page-size"
	GitlabHostnameFlag               = "gitlab-hostname"
	GitlabTokenFlag                  = "gitlab-token"
	GitlabUserFlag                   = "gitlab-user"
	GitlabWebhookSecretFlag          = "gitlab-webhook-secret" // nolint: gosec
	IncludeGitUntrackedFiles         = "include-git-untracked-files"
	APISecretFlag                    = "api-secret"
	HidePrevPlanComments             = "hide-prev-plan-comments"
	QuietPolicyChecks                = "quiet-policy-checks"
	LockingDBType                    = "locking-db-type"
	LogLevelFlag                     = "log-level"
	MarkdownTemplateOverridesDirFlag = "markdown-template-overrides-dir"
	MaxCommentsPerCommand            = "max-comments-per-command"
	ParallelPoolSize                 = "parallel-pool-size"
	StatsNamespace                   = "stats-namespace"
	AllowDraftPRs                    = "allow-draft-prs"
	PortFlag                         = "port"
	RedisDB                          = "redis-db"
	RedisHost                        = "redis-host"
	RedisPassword                    = "redis-password"
	RedisPort                        = "redis-port"
	RedisTLSEnabled                  = "redis-tls-enabled"
	RedisInsecureSkipVerify          = "redis-insecure-skip-verify"
	RepoConfigFlag                   = "repo-config"
	RepoConfigJSONFlag               = "repo-config-json"
	RepoAllowlistFlag                = "repo-allowlist"
	SilenceNoProjectsFlag            = "silence-no-projects"
	SilenceForkPRErrorsFlag          = "silence-fork-pr-errors"
	SilenceVCSStatusNoPlans          = "silence-vcs-status-no-plans"
	SilenceVCSStatusNoProjectsFlag   = "silence-vcs-status-no-projects"
	SilenceAllowlistErrorsFlag       = "silence-allowlist-errors"
	SkipCloneNoChanges               = "skip-clone-no-changes"
	SlackTokenFlag                   = "slack-token"
	SSLCertFileFlag                  = "ssl-cert-file"
	SSLKeyFileFlag                   = "ssl-key-file"
	RestrictFileList                 = "restrict-file-list"
	TFDistributionFlag               = "tf-distribution"
	TFDownloadFlag                   = "tf-download"
	TFDownloadURLFlag                = "tf-download-url"
	UseTFPluginCache                 = "use-tf-plugin-cache"
	VarFileAllowlistFlag             = "var-file-allowlist"
	VCSStatusName                    = "vcs-status-name"
	TFEHostnameFlag                  = "tfe-hostname"
	TFELocalExecutionModeFlag        = "tfe-local-execution-mode"
	TFETokenFlag                     = "tfe-token"
	WriteGitCredsFlag                = "write-git-creds" // nolint: gosec
	WebBasicAuthFlag                 = "web-basic-auth"
	WebUsernameFlag                  = "web-username"
	WebPasswordFlag                  = "web-password"
	WebsocketCheckOrigin             = "websocket-check-origin"

	// NOTE: Must manually set these as defaults in the setDefaults function.
	DefaultADBasicUser                  = ""
	DefaultADBasicPassword              = ""
	DefaultADHostname                   = "dev.azure.com"
	DefaultAutoDiscoverMode             = "auto"
	DefaultAutoplanFileList             = "**/*.tf,**/*.tfvars,**/*.tfvars.json,**/terragrunt.hcl,**/.terraform.lock.hcl"
	DefaultAllowCommands                = "version,plan,apply,unlock,approve_policies"
	DefaultCheckoutStrategy             = CheckoutStrategyBranch
	DefaultCheckoutDepth                = 0
	DefaultBitbucketBaseURL             = bitbucketcloud.BaseURL
	DefaultDataDir                      = "~/.atlantis"
	DefaultEmojiReaction                = ""
	DefaultExecutableName               = "atlantis"
	DefaultMarkdownTemplateOverridesDir = "~/.markdown_templates"
	DefaultGHHostname                   = "github.com"
	DefaultGiteaBaseURL                 = "https://gitea.com"
	DefaultGiteaPageSize                = 30
	DefaultGitlabHostname               = "gitlab.com"
	DefaultLockingDBType                = "boltdb"
	DefaultLogLevel                     = "info"
	DefaultMaxCommentsPerCommand        = 100
	DefaultParallelPoolSize             = 15
	DefaultStatsNamespace               = "atlantis"
	DefaultPort                         = 4141
	DefaultRedisDB                      = 0
	DefaultRedisPort                    = 6379
	DefaultRedisTLSEnabled              = false
	DefaultRedisInsecureSkipVerify      = false
	DefaultTFDistribution               = TFDistributionTerraform
	DefaultTFDownloadURL                = "https://releases.hashicorp.com"
	DefaultTFDownload                   = true
	DefaultTFEHostname                  = "app.terraform.io"
	DefaultVCSStatusName                = "atlantis"
	DefaultWebBasicAuth                 = false
	DefaultWebUsername                  = "atlantis"
	DefaultWebPassword                  = "atlantis"
)

var stringFlags = map[string]stringFlag{
	ADTokenFlag: {
		description: "Azure DevOps token of API user. Can also be specified via the ATLANTIS_AZUREDEVOPS_TOKEN environment variable.",
	},
	ADUserFlag: {
		description: "Azure DevOps username of API user.",
	},
	ADWebhookPasswordFlag: {
		description: "Azure DevOps basic HTTP authentication password for inbound webhooks " +
			"(see https://docs.microsoft.com/en-us/azure/devops/service-hooks/authorize?view=azure-devops)." +
			" SECURITY WARNING: If not specified, Atlantis won't be able to validate that the incoming webhook call came from your Azure DevOps org. " +
			"This means that an attacker could spoof calls to Atlantis and cause it to perform malicious actions. " +
			"Should be specified via the ATLANTIS_AZUREDEVOPS_WEBHOOK_PASSWORD environment variable.",
		defaultValue: "",
	},
	ADWebhookUserFlag: {
		description:  "Azure DevOps basic HTTP authentication username for inbound webhooks.",
		defaultValue: "",
	},
	ADHostnameFlag: {
		description:  "Azure DevOps hostname to support cloud and self hosted instances.",
		defaultValue: "dev.azure.com",
	},
	AllowCommandsFlag: {
		description:  "Comma separated list of acceptable atlantis commands.",
		defaultValue: DefaultAllowCommands,
	},
	AtlantisURLFlag: {
		description: "URL that Atlantis can be reached at. Defaults to http://$(hostname):$port where $port is from --" + PortFlag + ". Supports a base path ex. https://example.com/basepath.",
	},
	AutoDiscoverModeFlag: {
		description: "Auto discover mode controls whether projects in a repo are discovered by Atlantis. Defaults to 'auto' which " +
			"means projects will be discovered when no explicit projects are defined in repo config. Also supports 'enabled' (always " +
			"discover projects) and 'disabled' (never discover projects).",
		defaultValue: DefaultAutoDiscoverMode,
	},
	AutoplanModulesFromProjects: {
		description: "Comma separated list of file patterns to select projects Atlantis will index for module dependencies." +
			" Indexed projects will automatically be planned if a module they depend on is modified." +
			" Patterns use the dockerignore (https://docs.docker.com/engine/reference/builder/#dockerignore-file) syntax." +
			" A custom Workflow that uses autoplan 'when_modified' will ignore this value.",
		defaultValue: "",
	},
	AutoplanFileListFlag: {
		description: "Comma separated list of file patterns that Atlantis will use to check if a directory contains modified files that should trigger project planning." +
			" Patterns use the dockerignore (https://docs.docker.com/engine/reference/builder/#dockerignore-file) syntax." +
			" Use single quotes to avoid shell expansion of '*'. Defaults to '" + DefaultAutoplanFileList + "'." +
			" A custom Workflow that uses autoplan 'when_modified' will ignore this value.",
		defaultValue: DefaultAutoplanFileList,
	},
	BitbucketUserFlag: {
		description: "Bitbucket username of API user.",
	},
	BitbucketTokenFlag: {
		description: "Bitbucket app password of API user. Can also be specified via the ATLANTIS_BITBUCKET_TOKEN environment variable.",
	},
	BitbucketBaseURLFlag: {
		description: "Base URL of Bitbucket Server (aka Stash) installation." +
			" Must include 'http://' or 'https://'." +
			" If using Bitbucket Cloud (bitbucket.org), do not set.",
		defaultValue: DefaultBitbucketBaseURL,
	},
	BitbucketWebhookSecretFlag: {
		description: "Secret used to validate Bitbucket webhooks. Only Bitbucket Server supports webhook secrets." +
			" SECURITY WARNING: If not specified, Atlantis won't be able to validate that the incoming webhook call came from Bitbucket. " +
			"This means that an attacker could spoof calls to Atlantis and cause it to perform malicious actions. " +
			"Should be specified via the ATLANTIS_BITBUCKET_WEBHOOK_SECRET environment variable.",
	},
	CheckoutStrategyFlag: {
		description: "How to check out pull requests. Accepts either 'branch' (default) or 'merge'." +
			" If set to branch, Atlantis will check out the source branch of the pull request." +
			" If set to merge, Atlantis will check out the destination branch of the pull request (ex. main, master)" +
			" and then locally perform a git merge of the source branch." +
			" This effectively means Atlantis operates on the repo as it will look" +
			" after the pull request is merged.",
		defaultValue: "branch",
	},
	ConfigFlag: {
		description: "Path to yaml config file where flag values can also be set.",
	},
	DataDirFlag: {
		description:  "Path to directory to store Atlantis data.",
		defaultValue: DefaultDataDir,
	},
	DisableAutoplanLabelFlag: {
		description:  "Pull request label to disable atlantis auto planning feature only if present.",
		defaultValue: "",
	},
	DisableUnlockLabelFlag: {
		description:  "Pull request label to disable atlantis unlock feature only if present.",
		defaultValue: "",
	},
	EmojiReaction: {
		description:  "Emoji Reaction to use to react to comments.",
		defaultValue: DefaultEmojiReaction,
	},
	ExecutableName: {
		description:  "Comment command executable name.",
		defaultValue: DefaultExecutableName,
	},
	GHHostnameFlag: {
		description:  "Hostname of your Github Enterprise installation. If using github.com, no need to set.",
		defaultValue: DefaultGHHostname,
	},
	GHTeamAllowlistFlag: {
		description: "Comma separated list of key-value pairs representing the GitHub teams and the operations that " +
			"the members of a particular team are allowed to perform. " +
			"The format is {team}:{command},{team}:{command}. " +
			"Valid values for 'command' are 'plan', 'apply' and '*', e.g. 'dev:plan,ops:apply,devops:*'" +
			"This example gives the users from the 'dev' GitHub team the permissions to execute the 'plan' command, " +
			"the 'ops' team the permissions to execute the 'apply' command, " +
			"and allows the 'devops' team to perform any operation. If this argument is not provided, the default value (*:*) " +
			"will be used and the default behavior will be to not check permissions " +
			"and to allow users from any team to perform any operation.",
	},
	GHUserFlag: {
		description:  "GitHub username of API user.",
		defaultValue: "",
	},
	GHTokenFlag: {
		description: "GitHub token of API user. Can also be specified via the ATLANTIS_GH_TOKEN environment variable.",
	},
	GHAppKeyFlag: {
		description:  "The GitHub App's private key",
		defaultValue: "",
	},
	GHAppKeyFileFlag: {
		description:  "A path to a file containing the GitHub App's private key",
		defaultValue: "",
	},
	GHAppSlugFlag: {
		description: "The Github app slug (ie. the URL-friendly name of your GitHub App)",
	},
	GHOrganizationFlag: {
		description:  "The name of the GitHub organization to use during the creation of a Github App for Atlantis",
		defaultValue: "",
	},
	GHWebhookSecretFlag: {
		description: "Secret used to validate GitHub webhooks (see https://developer.github.com/webhooks/securing/)." +
			" SECURITY WARNING: If not specified, Atlantis won't be able to validate that the incoming webhook call came from GitHub. " +
			"This means that an attacker could spoof calls to Atlantis and cause it to perform malicious actions. " +
			"Should be specified via the ATLANTIS_GH_WEBHOOK_SECRET environment variable.",
	},
	GiteaBaseURLFlag: {
		description: "Base URL of Gitea server installation. Must include 'http://' or 'https://'.",
	},
	GiteaUserFlag: {
		description:  "Gitea username of API user.",
		defaultValue: "",
	},
	GiteaTokenFlag: {
		description: "Gitea token of API user. Can also be specified via the ATLANTIS_GITEA_TOKEN environment variable.",
	},
	GiteaWebhookSecretFlag: {
		description: "Optional secret used to validate Gitea webhooks." +
			" SECURITY WARNING: If not specified, Atlantis won't be able to validate that the incoming webhook call came from Gitea. " +
			"This means that an attacker could spoof calls to Atlantis and cause it to perform malicious actions. " +
			"Should be specified via the ATLANTIS_GITEA_WEBHOOK_SECRET environment variable.",
	},
	GitlabHostnameFlag: {
		description:  "Hostname of your GitLab Enterprise installation. If using gitlab.com, no need to set.",
		defaultValue: DefaultGitlabHostname,
	},
	GitlabUserFlag: {
		description: "GitLab username of API user.",
	},
	GitlabTokenFlag: {
		description: "GitLab token of API user. Can also be specified via the ATLANTIS_GITLAB_TOKEN environment variable.",
	},
	GitlabWebhookSecretFlag: {
		description: "Optional secret used to validate GitLab webhooks." +
			" SECURITY WARNING: If not specified, Atlantis won't be able to validate that the incoming webhook call came from GitLab. " +
			"This means that an attacker could spoof calls to Atlantis and cause it to perform malicious actions. " +
			"Should be specified via the ATLANTIS_GITLAB_WEBHOOK_SECRET environment variable.",
	},
	APISecretFlag: {
		description: "Secret used to validate requests made to the /api/* endpoints",
	},
	LockingDBType: {
		description:  "The locking database type to use for storing plan and apply locks.",
		defaultValue: DefaultLockingDBType,
	},
	LogLevelFlag: {
		description:  "Log level. Either debug, info, warn, or error.",
		defaultValue: DefaultLogLevel,
	},
	MarkdownTemplateOverridesDirFlag: {
		description:  "Directory for custom overrides to the markdown templates used for comments.",
		defaultValue: DefaultMarkdownTemplateOverridesDir,
	},
	StatsNamespace: {
		description:  "Namespace for aggregating stats.",
		defaultValue: DefaultStatsNamespace,
	},
	RedisHost: {
		description: "The Redis Hostname for when using a Locking DB type of 'redis'.",
	},
	RedisPassword: {
		description: "The Redis Password for when using a Locking DB type of 'redis'.",
	},
	RepoConfigFlag: {
		description: "Path to a repo config file, used to customize how Atlantis runs on each repo. See runatlantis.io/docs for more details.",
	},
	RepoConfigJSONFlag: {
		description: "Specify repo config as a JSON string. Useful if you don't want to write a config file to disk.",
	},
	RepoAllowlistFlag: {
		description: "Comma separated list of repositories that Atlantis will operate on. " +
			"The format is {hostname}/{owner}/{repo}, ex. github.com/runatlantis/atlantis. '*' matches any characters until the next comma. Examples: " +
			"all repos: '*' (not secure), an entire hostname: 'internalgithub.com/*' or an organization: 'github.com/runatlantis/*'." +
			" For Bitbucket Server, {owner} is the name of the project (not the key).",
	},
	SlackTokenFlag: {
		description: "API token for Slack notifications.",
	},
	SSLCertFileFlag: {
		description: "File containing x509 Certificate used for serving HTTPS. If the cert is signed by a CA, the file should be the concatenation of the server's certificate, any intermediates, and the CA's certificate.",
	},
	SSLKeyFileFlag: {
		description: fmt.Sprintf("File containing x509 private key matching --%s.", SSLCertFileFlag),
	},
	TFDistributionFlag: {
		description:  fmt.Sprintf("Which TF distribution to use. Can be set to %s or %s.", TFDistributionTerraform, TFDistributionOpenTofu),
		defaultValue: DefaultTFDistribution,
	},
	TFDownloadURLFlag: {
		description:  "Base URL to download Terraform versions from.",
		defaultValue: DefaultTFDownloadURL,
	},
	TFEHostnameFlag: {
		description:  "Hostname of your Terraform Enterprise installation. If using Terraform Cloud no need to set.",
		defaultValue: DefaultTFEHostname,
	},
	TFETokenFlag: {
		description: "API token for Terraform Cloud/Enterprise. This will be used to generate a ~/.terraformrc file." +
			" Only set if using TFC/E as a remote backend." +
			" Should be specified via the ATLANTIS_TFE_TOKEN environment variable for security.",
	},
	DefaultTFVersionFlag: {
		description: "Terraform version to default to (ex. v0.12.0). Will download if not yet on disk." +
			" If not set, Atlantis uses the terraform binary in its PATH.",
	},
	VarFileAllowlistFlag: {
		description: "Comma-separated list of additional paths where variable definition files can be read from." +
			" If this argument is not provided, it defaults to Atlantis' data directory, determined by the --data-dir argument.",
	},
	VCSStatusName: {
		description:  "Name used to identify Atlantis for pull request statuses.",
		defaultValue: DefaultVCSStatusName,
	},
	WebUsernameFlag: {
		description:  "Username used for Web Basic Authentication on Atlantis HTTP Middleware",
		defaultValue: DefaultWebUsername,
	},
	WebPasswordFlag: {
		description:  "Password used for Web Basic Authentication on Atlantis HTTP Middleware",
		defaultValue: DefaultWebPassword,
	},
}

var boolFlags = map[string]boolFlag{
	AllowForkPRsFlag: {
		description:  "Allow Atlantis to run on pull requests from forks. A security issue for public repos.",
		defaultValue: false,
	},
	AutoplanModules: {
		description:  "Automatically plan projects that have a changed module from the local repository.",
		defaultValue: false,
	},
	AutomergeFlag: {
		description:  "Automatically merge pull requests when all plans are successfully applied.",
		defaultValue: false,
	},
	DisableApplyAllFlag: {
		description:  "Disable \"atlantis apply\" command without any flags (i.e. apply all). A specific project/workspace/directory has to be specified for applies.",
		defaultValue: false,
	},
	DisableAutoplanFlag: {
		description:  "Disable atlantis auto planning feature",
		defaultValue: false,
	},

	DisableRepoLockingFlag: {
		description: "Disable atlantis locking repos",
	},
	DisableGlobalApplyLockFlag: {
		description: "Disable atlantis global apply lock in UI",
	},
	DiscardApprovalOnPlanFlag: {
		description:  "Enables the discarding of approval if a new plan has been executed. Currently only Github is supported",
		defaultValue: false,
	},
	EnablePolicyChecksFlag: {
		description:  "Enable atlantis to run user defined policy checks.  This is explicitly disabled for TFE/TFC backends since plan files are inaccessible.",
		defaultValue: false,
	},
	EnableRegExpCmdFlag: {
		description:  "Enable Atlantis to use regular expressions on plan/apply commands when \"-p\" flag is passed with it.",
		defaultValue: false,
	},
	EnableDiffMarkdownFormat: {
		description:  "Enable Atlantis to format Terraform plan output into a markdown-diff friendly format for color-coding purposes.",
		defaultValue: false,
	},
	FailOnPreWorkflowHookError: {
		description:  "Fail and do not run the requested Atlantis command if any of the pre workflow hooks error.",
		defaultValue: false,
	},
	GHAllowMergeableBypassApply: {
		description:  "Feature flag to enable functionality to allow mergeable check to ignore apply required check",
		defaultValue: false,
	},
	AllowDraftPRs: {
		description:  "Enable autoplan for Github Draft Pull Requests",
		defaultValue: false,
	},
	HidePrevPlanComments: {
		description: "Hide previous plan comments to reduce clutter in the PR. " +
			"VCS support is limited to: GitHub.",
		defaultValue: false,
	},
	IncludeGitUntrackedFiles: {
		description:  "Include git untracked files in the Atlantis modified file scope.",
		defaultValue: false,
	},
	ParallelPlanFlag: {
		description:  "Run plan operations in parallel.",
		defaultValue: false,
	},
	ParallelApplyFlag: {
		description:  "Run apply operations in parallel.",
		defaultValue: false,
	},
	QuietPolicyChecks: {
		description:  "Exclude policy check comments from pull requests unless there's an actual error from conftest. This also excludes warnings.",
		defaultValue: false,
	},
	RedisTLSEnabled: {
		description:  "Enable TLS on the connection to Redis with a min TLS version of 1.2",
		defaultValue: DefaultRedisTLSEnabled,
	},
	RedisInsecureSkipVerify: {
		description:  "Controls whether the Redis client verifies the Redis server's certificate chain and host name. If true, accepts any certificate presented by the server and any host name in that certificate.",
		defaultValue: DefaultRedisInsecureSkipVerify,
	},
	SilenceNoProjectsFlag: {
		description:  "Silences Atlants from responding to PRs when it finds no projects.",
		defaultValue: false,
	},
	SilenceForkPRErrorsFlag: {
		description:  "Silences the posting of fork pull requests not allowed error comments.",
		defaultValue: false,
	},
	SilenceVCSStatusNoPlans: {
		description:  "Silences VCS commit status when autoplan finds no projects to plan.",
		defaultValue: false,
	},
	SilenceVCSStatusNoProjectsFlag: {
		description:  "Silences VCS commit status when for all commands when a project is not defined.",
		defaultValue: false,
	},
	SilenceAllowlistErrorsFlag: {
		description:  "Silences the posting of allowlist error comments.",
		defaultValue: false,
	},
	DisableMarkdownFoldingFlag: {
		description:  "Toggle off folding in markdown output.",
		defaultValue: false,
	},
	WriteGitCredsFlag: {
		description: "Write out a .git-credentials file with the provider user and token to allow cloning private modules over HTTPS or SSH." +
			" This writes secrets to disk and should only be enabled in a secure environment.",
		defaultValue: false,
	},
	SkipCloneNoChanges: {
		description:  "Skips cloning the PR repo if there are no projects were changed in the PR.",
		defaultValue: false,
	},
	TFDownloadFlag: {
		description:  "Allow Atlantis to list & download Terraform versions. Setting this to false can be helpful in air-gapped environments.",
		defaultValue: DefaultTFDownload,
	},
	TFELocalExecutionModeFlag: {
		description:  "Enable if you're using local execution mode (instead of TFE/C's remote execution mode).",
		defaultValue: false,
	},
	WebBasicAuthFlag: {
		description:  "Switches on or off the Basic Authentication on the HTTP Middleware interface",
		defaultValue: DefaultWebBasicAuth,
	},
	RestrictFileList: {
		description:  "Block plan requests from projects outside the files modified in the pull request.",
		defaultValue: false,
	},
	WebsocketCheckOrigin: {
		description:  "Enable websocket origin check",
		defaultValue: false,
	},
	HideUnchangedPlanComments: {
		description:  "Remove no-changes plan comments from the pull request.",
		defaultValue: false,
	},
	UseTFPluginCache: {
		description:  "Enable the use of the Terraform plugin cache",
		defaultValue: true,
	},
}
var intFlags = map[string]intFlag{
	CheckoutDepthFlag: {
		description: fmt.Sprintf("Used only if --%s=%s.", CheckoutStrategyFlag, CheckoutStrategyMerge) +
			" How many commits to include in each of base and feature branches when cloning repository." +
			" If merge base is further behind than this number of commits from any of branches heads, full fetch will be performed.",
		defaultValue: DefaultCheckoutDepth,
	},
	MaxCommentsPerCommand: {
		description:  "If non-zero, the maximum number of comments to split command output into before truncating.",
		defaultValue: DefaultMaxCommentsPerCommand,
	},
	GiteaPageSizeFlag: {
		description:  "Optional value that specifies the number of results per page to expect from Gitea.",
		defaultValue: DefaultGiteaPageSize,
	},
	ParallelPoolSize: {
		description:  "Max size of the wait group that runs parallel plans and applies (if enabled).",
		defaultValue: DefaultParallelPoolSize,
	},
	PortFlag: {
		description:  "Port to bind to.",
		defaultValue: DefaultPort,
	},
	RedisDB: {
		description:  "The Redis Database to use when using a Locking DB type of 'redis'.",
		defaultValue: DefaultRedisDB,
	},
	RedisPort: {
		description:  "The Redis Port for when using a Locking DB type of 'redis'.",
		defaultValue: DefaultRedisPort,
	},
}

var int64Flags = map[string]int64Flag{
	GHAppIDFlag: {
		description:  "GitHub App Id. If defined, initializes the GitHub client with app-based credentials",
		defaultValue: 0,
	},
	GHAppInstallationIDFlag: {
		description: "GitHub App Installation Id. If defined, initializes the GitHub client with app-based credentials " +
			"using this specific GitHub Application Installation ID, otherwise it attempts to auto-detect it. " +
			"Note that this value must be set if you want to have one App and multiple installations of that same " +
			"application.",
		defaultValue: 0,
	},
}

// ValidLogLevels are the valid log levels that can be set
var ValidLogLevels = []string{"debug", "info", "warn", "error"}

type stringFlag struct {
	description  string
	defaultValue string
	hidden       bool
}
type intFlag struct {
	description  string
	defaultValue int
	hidden       bool
}
type int64Flag struct {
	description  string
	defaultValue int64
	hidden       bool
}
type boolFlag struct {
	description  string
	defaultValue bool
	hidden       bool
}

// ServerCmd is an abstraction that helps us test. It allows
// us to mock out starting the actual server.
type ServerCmd struct {
	ServerCreator ServerCreator
	Viper         *viper.Viper
	// SilenceOutput set to true means nothing gets printed.
	// Useful for testing to keep the logs clean.
	SilenceOutput   bool
	AtlantisVersion string
	Logger          logging.SimpleLogging
}

// ServerCreator creates servers.
// It's an abstraction to help us test.
type ServerCreator interface {
	NewServer(userConfig server.UserConfig, config server.Config) (ServerStarter, error)
}

// DefaultServerCreator is the concrete implementation of ServerCreator.
type DefaultServerCreator struct{}

// ServerStarter is for starting up a server.
// It's an abstraction to help us test.
type ServerStarter interface {
	Start() error
}

// NewServer returns the real Atlantis server object.
func (d *DefaultServerCreator) NewServer(userConfig server.UserConfig, config server.Config) (ServerStarter, error) {
	return server.NewServer(userConfig, config)
}

// Init returns the runnable cobra command.
func (s *ServerCmd) Init() *cobra.Command {
	c := &cobra.Command{
		Use:           "server",
		Short:         "Start the atlantis server",
		Long:          `Start the atlantis server and listen for webhook calls.`,
		SilenceErrors: true,
		SilenceUsage:  true,
		PreRunE: s.withErrPrint(func(cmd *cobra.Command, args []string) error {
			return s.preRun()
		}),
		RunE: s.withErrPrint(func(cmd *cobra.Command, args []string) error {
			return s.run()
		}),
	}

	// Configure viper to accept env vars prefixed with ATLANTIS_ that can be
	// used instead of flags.
	s.Viper.SetEnvPrefix("ATLANTIS")
	s.Viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	s.Viper.AutomaticEnv()
	s.Viper.SetTypeByDefaultValue(true)

	c.SetUsageTemplate(usageTmpl(stringFlags, intFlags, boolFlags))
	// If a user passes in an invalid flag, tell them what the flag was.
	c.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		s.printErr(err)
		return err
	})

	// Set string flags.
	for name, f := range stringFlags {
		usage := f.description
		if f.defaultValue != "" {
			usage = fmt.Sprintf("%s (default %q)", usage, f.defaultValue)
		}
		c.Flags().String(name, "", usage+"\n")
		s.Viper.BindPFlag(name, c.Flags().Lookup(name)) // nolint: errcheck
		if f.hidden {
			c.Flags().MarkHidden(name) // nolint: errcheck
		}
	}

	// Set int flags.
	for name, f := range intFlags {
		usage := f.description
		if f.defaultValue != 0 {
			usage = fmt.Sprintf("%s (default %d)", usage, f.defaultValue)
		}
		c.Flags().Int(name, 0, usage+"\n")
		if f.hidden {
			c.Flags().MarkHidden(name) // nolint: errcheck
		}
		s.Viper.BindPFlag(name, c.Flags().Lookup(name)) // nolint: errcheck
	}

	// Set int64 flags.
	for name, f := range int64Flags {
		usage := f.description
		if f.defaultValue != 0 {
			usage = fmt.Sprintf("%s (default %d)", usage, f.defaultValue)
		}
		c.Flags().Int(name, 0, usage+"\n")
		if f.hidden {
			c.Flags().MarkHidden(name) // nolint: errcheck
		}
		s.Viper.BindPFlag(name, c.Flags().Lookup(name)) // nolint: errcheck
	}

	// Set bool flags.
	for name, f := range boolFlags {
		c.Flags().Bool(name, f.defaultValue, f.description+"\n")
		if f.hidden {
			c.Flags().MarkHidden(name) // nolint: errcheck
		}
		s.Viper.BindPFlag(name, c.Flags().Lookup(name)) // nolint: errcheck
	}

	return c
}

func (s *ServerCmd) preRun() error {
	// If passed a config file then try and load it.
	configFile := s.Viper.GetString(ConfigFlag)
	if configFile != "" {
		s.Viper.SetConfigFile(configFile)
		if err := s.Viper.ReadInConfig(); err != nil {
			return errors.Wrapf(err, "invalid config: reading %s", configFile)
		}
	}
	return nil
}

func (s *ServerCmd) run() error {
	var userConfig server.UserConfig
	if err := s.Viper.Unmarshal(&userConfig); err != nil {
		return err
	}
	s.setDefaults(&userConfig, s.Viper)

	// Now that we've parsed the config we can set our local logger to the
	// right level.
	s.Logger.SetLevel(userConfig.ToLogLevel())

	if err := s.validate(userConfig); err != nil {
		return err
	}
	if err := s.setAtlantisURL(&userConfig); err != nil {
		return err
	}
	if err := s.setDataDir(&userConfig); err != nil {
		return err
	}
	if err := s.setMarkdownTemplateOverridesDir(&userConfig); err != nil {
		return err
	}
	s.setVarFileAllowlist(&userConfig)
	if err := s.deprecationWarnings(&userConfig); err != nil {
		return err
	}
	s.securityWarnings(&userConfig)
	s.trimAtSymbolFromUsers(&userConfig)

	// Config looks good. Start the server.
	server, err := s.ServerCreator.NewServer(userConfig, server.Config{
		AllowForkPRsFlag:        AllowForkPRsFlag,
		AtlantisURLFlag:         AtlantisURLFlag,
		AtlantisVersion:         s.AtlantisVersion,
		DefaultTFVersionFlag:    DefaultTFVersionFlag,
		RepoConfigJSONFlag:      RepoConfigJSONFlag,
		SilenceForkPRErrorsFlag: SilenceForkPRErrorsFlag,
	})

	if err != nil {
		return errors.Wrap(err, "initializing server")
	}
	return server.Start()
}

func (s *ServerCmd) setDefaults(c *server.UserConfig, v *viper.Viper) {
	if c.AzureDevOpsHostname == "" {
		c.AzureDevOpsHostname = DefaultADHostname
	}
	if c.AutoplanFileList == "" {
		c.AutoplanFileList = DefaultAutoplanFileList
	}
	if c.CheckoutDepth <= 0 {
		c.CheckoutDepth = DefaultCheckoutDepth
	}
	if c.AllowCommands == "" {
		c.AllowCommands = DefaultAllowCommands
	}
	if c.CheckoutStrategy == "" {
		c.CheckoutStrategy = DefaultCheckoutStrategy
	}
	if c.DataDir == "" {
		c.DataDir = DefaultDataDir
	}
	if c.GithubHostname == "" {
		c.GithubHostname = DefaultGHHostname
	}
	if c.GitlabHostname == "" {
		c.GitlabHostname = DefaultGitlabHostname
	}
	if c.GiteaBaseURL == "" {
		c.GiteaBaseURL = DefaultGiteaBaseURL
	}
	if c.GiteaPageSize == 0 {
		c.GiteaPageSize = DefaultGiteaPageSize
	}
	if c.BitbucketBaseURL == "" {
		c.BitbucketBaseURL = DefaultBitbucketBaseURL
	}
	if c.EmojiReaction == "" {
		c.EmojiReaction = DefaultEmojiReaction
	}
	if c.ExecutableName == "" {
		c.ExecutableName = DefaultExecutableName
	}
	if c.LockingDBType == "" {
		c.LockingDBType = DefaultLockingDBType
	}
	if c.LogLevel == "" {
		c.LogLevel = DefaultLogLevel
	}
	if c.MarkdownTemplateOverridesDir == "" {
		c.MarkdownTemplateOverridesDir = DefaultMarkdownTemplateOverridesDir
	}
	if !v.IsSet("max-comments-per-command") {
		c.MaxCommentsPerCommand = DefaultMaxCommentsPerCommand
	}
	if c.ParallelPoolSize == 0 {
		c.ParallelPoolSize = DefaultParallelPoolSize
	}
	if c.StatsNamespace == "" {
		c.StatsNamespace = DefaultStatsNamespace
	}
	if c.Port == 0 {
		c.Port = DefaultPort
	}
	if c.RedisDB == 0 {
		c.RedisDB = DefaultRedisDB
	}
	if c.RedisPort == 0 {
		c.RedisPort = DefaultRedisPort
	}
	if c.TFDistribution == "" {
		c.TFDistribution = DefaultTFDistribution
	}
	if c.TFDownloadURL == "" {
		c.TFDownloadURL = DefaultTFDownloadURL
	}
	if c.VCSStatusName == "" {
		c.VCSStatusName = DefaultVCSStatusName
	}
	if c.TFEHostname == "" {
		c.TFEHostname = DefaultTFEHostname
	}
	if c.WebUsername == "" {
		c.WebUsername = DefaultWebUsername
	}
	if c.WebPassword == "" {
		c.WebPassword = DefaultWebPassword
	}
	if c.AutoDiscoverModeFlag == "" {
		c.AutoDiscoverModeFlag = DefaultAutoDiscoverMode
	}
}

func (s *ServerCmd) validate(userConfig server.UserConfig) error {
	userConfig.LogLevel = strings.ToLower(userConfig.LogLevel)
	if !isValidLogLevel(userConfig.LogLevel) {
		return fmt.Errorf("invalid log level: must be one of %v", ValidLogLevels)
	}

	if userConfig.TFDistribution != TFDistributionTerraform && userConfig.TFDistribution != TFDistributionOpenTofu {
		return fmt.Errorf("invalid tf distribution: expected one of %s or %s",
			TFDistributionTerraform, TFDistributionOpenTofu)
	}

	checkoutStrategy := userConfig.CheckoutStrategy
	if checkoutStrategy != CheckoutStrategyBranch && checkoutStrategy != CheckoutStrategyMerge {
		return fmt.Errorf("invalid checkout strategy: not one of %s or %s",
			CheckoutStrategyBranch, CheckoutStrategyMerge)
	}

	if (userConfig.SSLKeyFile == "") != (userConfig.SSLCertFile == "") {
		return fmt.Errorf("--%s and --%s are both required for ssl", SSLKeyFileFlag, SSLCertFileFlag)
	}

	// The following combinations are valid.
	// 1. github user and token set
	// 2. github app ID and (key file set or key set)
	// 3. gitea user and token set
	// 4. gitlab user and token set
	// 5. bitbucket user and token set
	// 6. azuredevops user and token set
	// 7. any combination of the above
	vcsErr := fmt.Errorf("--%s/--%s or --%s/--%s or --%s/--%s or --%s/--%s or --%s/--%s or --%s/--%s or --%s/--%s must be set", GHUserFlag, GHTokenFlag, GHAppIDFlag, GHAppKeyFileFlag, GHAppIDFlag, GHAppKeyFlag, GiteaUserFlag, GiteaTokenFlag, GitlabUserFlag, GitlabTokenFlag, BitbucketUserFlag, BitbucketTokenFlag, ADUserFlag, ADTokenFlag)
	if ((userConfig.GithubUser == "") != (userConfig.GithubToken == "")) ||
		((userConfig.GiteaUser == "") != (userConfig.GiteaToken == "")) ||
		((userConfig.GitlabUser == "") != (userConfig.GitlabToken == "")) ||
		((userConfig.BitbucketUser == "") != (userConfig.BitbucketToken == "")) ||
		((userConfig.AzureDevopsUser == "") != (userConfig.AzureDevopsToken == "")) {
		return vcsErr
	}
	if (userConfig.GithubAppID != 0) && ((userConfig.GithubAppKey == "") && (userConfig.GithubAppKeyFile == "")) {
		return vcsErr
	}
	if (userConfig.GithubAppID == 0) && ((userConfig.GithubAppKey != "") || (userConfig.GithubAppKeyFile != "")) {
		return vcsErr
	}
	// At this point, we know that there can't be a single user/token without
	// its partner, but we haven't checked if any user/token is set at all.
	if userConfig.GithubAppID == 0 && userConfig.GithubUser == "" && userConfig.GiteaUser == "" && userConfig.GitlabUser == "" && userConfig.BitbucketUser == "" && userConfig.AzureDevopsUser == "" {
		return vcsErr
	}

	if userConfig.RepoAllowlist == "" {
		return fmt.Errorf("--%s must be set for security purposes", RepoAllowlistFlag)
	}
	if strings.Contains(userConfig.RepoAllowlist, "://") {
		return fmt.Errorf("--%s cannot contain ://, should be hostnames only", RepoAllowlistFlag)
	}

	if userConfig.BitbucketBaseURL == DefaultBitbucketBaseURL && userConfig.BitbucketWebhookSecret != "" {
		return fmt.Errorf("--%s cannot be specified for Bitbucket Cloud because it is not supported by Bitbucket", BitbucketWebhookSecretFlag)
	}

	parsed, err := url.Parse(userConfig.BitbucketBaseURL)
	if err != nil {
		return fmt.Errorf("error parsing --%s flag value %q: %s", BitbucketWebhookSecretFlag, userConfig.BitbucketBaseURL, err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("--%s must have http:// or https://, got %q", BitbucketBaseURLFlag, userConfig.BitbucketBaseURL)
	}

	parsed, err = url.Parse(userConfig.GiteaBaseURL)
	if err != nil {
		return fmt.Errorf("error parsing --%s flag value %q: %s", GiteaWebhookSecretFlag, userConfig.GiteaBaseURL, err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("--%s must have http:// or https://, got %q", GiteaBaseURLFlag, userConfig.GiteaBaseURL)
	}

	if userConfig.RepoConfig != "" && userConfig.RepoConfigJSON != "" {
		return fmt.Errorf("cannot use --%s and --%s at the same time", RepoConfigFlag, RepoConfigJSONFlag)
	}

	// Warn if any tokens have newlines.
	for name, token := range map[string]string{
		GHTokenFlag:                userConfig.GithubToken,
		GHWebhookSecretFlag:        userConfig.GithubWebhookSecret,
		GitlabTokenFlag:            userConfig.GitlabToken,
		GitlabWebhookSecretFlag:    userConfig.GitlabWebhookSecret,
		BitbucketTokenFlag:         userConfig.BitbucketToken,
		BitbucketWebhookSecretFlag: userConfig.BitbucketWebhookSecret,
		GiteaTokenFlag:             userConfig.GiteaToken,
		GiteaWebhookSecretFlag:     userConfig.GiteaWebhookSecret,
	} {
		if strings.Contains(token, "\n") {
			s.Logger.Warn("--%s contains a newline which is usually unintentional", name)
		}
	}

	if userConfig.TFEHostname != DefaultTFEHostname && userConfig.TFEToken == "" {
		return fmt.Errorf("if setting --%s, must set --%s", TFEHostnameFlag, TFETokenFlag)
	}

	_, patternErr := patternmatcher.New(strings.Split(userConfig.AutoplanFileList, ","))
	if patternErr != nil {
		return errors.Wrapf(patternErr, "invalid pattern in --%s, %s", AutoplanFileListFlag, userConfig.AutoplanFileList)
	}

	if _, err := userConfig.ToAllowCommandNames(); err != nil {
		return errors.Wrapf(err, "invalid --%s", AllowCommandsFlag)
	}

	return nil
}

// setAtlantisURL sets the externally accessible URL for atlantis.
func (s *ServerCmd) setAtlantisURL(userConfig *server.UserConfig) error {
	if userConfig.AtlantisURL == "" {
		hostname, err := os.Hostname()
		if err != nil {
			return errors.Wrap(err, "failed to determine hostname")
		}
		userConfig.AtlantisURL = fmt.Sprintf("http://%s:%d", hostname, userConfig.Port)
	}
	return nil
}

// setDataDir checks if ~ was used in data-dir and converts it to the actual
// home directory. If we don't do this, we'll create a directory called "~"
// instead of actually using home. It also converts relative paths to absolute.
func (s *ServerCmd) setDataDir(userConfig *server.UserConfig) error {
	finalPath := userConfig.DataDir

	// Convert ~ to the actual home dir.
	if strings.HasPrefix(finalPath, "~/") {
		var err error
		finalPath, err = homedir.Expand(finalPath)
		if err != nil {
			return errors.Wrap(err, "determining home directory")
		}
	}

	// Convert relative paths to absolute.
	finalPath, err := filepath.Abs(finalPath)
	if err != nil {
		return errors.Wrap(err, "making data-dir absolute")
	}
	userConfig.DataDir = finalPath
	return nil
}

// setMarkdownTemplateOverridesDir checks if ~ was used in markdown-template-overrides-dir and converts it to the actual
// home directory. If we don't do this, we'll create a directory called "~"
// instead of actually using home. It also converts relative paths to absolute.
func (s *ServerCmd) setMarkdownTemplateOverridesDir(userConfig *server.UserConfig) error {
	finalPath := userConfig.MarkdownTemplateOverridesDir

	// Convert ~ to the actual home dir.
	if strings.HasPrefix(finalPath, "~/") {
		var err error
		finalPath, err = homedir.Expand(finalPath)
		if err != nil {
			return errors.Wrap(err, "determining home directory")
		}
	}

	// Convert relative paths to absolute.
	finalPath, err := filepath.Abs(finalPath)
	if err != nil {
		return errors.Wrap(err, "making markdown-template-overrides-dir absolute")
	}
	userConfig.MarkdownTemplateOverridesDir = finalPath
	return nil
}

// setVarFileAllowlist checks if var-file-allowlist is unassigned and makes it default to data-dir for better backward
// compatibility.
func (s *ServerCmd) setVarFileAllowlist(userConfig *server.UserConfig) {
	if userConfig.VarFileAllowlist == "" {
		userConfig.VarFileAllowlist = userConfig.DataDir
	}
}

// trimAtSymbolFromUsers trims @ from the front of the github and gitlab usernames
func (s *ServerCmd) trimAtSymbolFromUsers(userConfig *server.UserConfig) {
	userConfig.GithubUser = strings.TrimPrefix(userConfig.GithubUser, "@")
	userConfig.GiteaUser = strings.TrimPrefix(userConfig.GiteaUser, "@")
	userConfig.GitlabUser = strings.TrimPrefix(userConfig.GitlabUser, "@")
	userConfig.BitbucketUser = strings.TrimPrefix(userConfig.BitbucketUser, "@")
	userConfig.AzureDevopsUser = strings.TrimPrefix(userConfig.AzureDevopsUser, "@")
}

func (s *ServerCmd) securityWarnings(userConfig *server.UserConfig) {
	if userConfig.GithubUser != "" && userConfig.GithubWebhookSecret == "" && !s.SilenceOutput {
		s.Logger.Warn("no GitHub webhook secret set. This could allow attackers to spoof requests from GitHub")
	}
	if userConfig.GiteaUser != "" && userConfig.GiteaWebhookSecret == "" && !s.SilenceOutput {
		s.Logger.Warn("no Gitea webhook secret set. This could allow attackers to spoof requests from Gitea")
	}
	if userConfig.GitlabUser != "" && userConfig.GitlabWebhookSecret == "" && !s.SilenceOutput {
		s.Logger.Warn("no GitLab webhook secret set. This could allow attackers to spoof requests from GitLab")
	}
	if userConfig.BitbucketUser != "" && userConfig.BitbucketBaseURL != DefaultBitbucketBaseURL && userConfig.BitbucketWebhookSecret == "" && !s.SilenceOutput {
		s.Logger.Warn("no Bitbucket webhook secret set. This could allow attackers to spoof requests from Bitbucket")
	}
	if userConfig.BitbucketUser != "" && userConfig.BitbucketBaseURL == DefaultBitbucketBaseURL && !s.SilenceOutput {
		s.Logger.Warn("Bitbucket Cloud does not support webhook secrets. This could allow attackers to spoof requests from Bitbucket. Ensure you are allowing only Bitbucket IPs")
	}
	if userConfig.AzureDevopsWebhookUser != "" && userConfig.AzureDevopsWebhookPassword == "" && !s.SilenceOutput {
		s.Logger.Warn("no Azure DevOps webhook user and password set. This could allow attackers to spoof requests from Azure DevOps.")
	}
}

// deprecationWarnings prints a warning if flags that are deprecated are
// being used.
func (s *ServerCmd) deprecationWarnings(userConfig *server.UserConfig) error {
	var deprecatedFlags []string

	// Currently there are no deprecated flags; if flags become deprecated, add them here like so
	//       if userConfig.SomeDeprecatedFlag {
	//               deprecatedFlags = append(deprecatedFlags, SomeDeprecatedFlag)
	//       }
	//

	if len(deprecatedFlags) > 0 {
		warning := "WARNING: "
		if len(deprecatedFlags) == 1 {
			warning += fmt.Sprintf("Flag --%s has been deprecated.", deprecatedFlags[0])
		} else {
			warning += fmt.Sprintf("Flags --%s and --%s have been deprecated.", strings.Join(deprecatedFlags[0:len(deprecatedFlags)-1], ", --"), deprecatedFlags[len(deprecatedFlags)-1:][0])
		}
		fmt.Println(warning)
	}

	return nil
}

// withErrPrint prints out any cmd errors to stderr.
func (s *ServerCmd) withErrPrint(f func(*cobra.Command, []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := f(cmd, args)
		if err != nil && !s.SilenceOutput {
			s.printErr(err)
		}
		return err
	}
}

// printErr prints err to stderr using a red terminal colour.
func (s *ServerCmd) printErr(err error) {
	fmt.Fprintf(os.Stderr, "%sError: %s%s\n", "\033[31m", err.Error(), "\033[39m")
}

func isValidLogLevel(level string) bool {
	for _, logLevel := range ValidLogLevels {
		if logLevel == level {
			return true
		}
	}

	return false
}
