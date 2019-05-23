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
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server"
	"github.com/runatlantis/atlantis/server/events/vcs/bitbucketcloud"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// To add a new flag you must:
// 1. Add a const with the flag name (in alphabetic order).
// 2. Add a new field to server.UserConfig and set the mapstructure tag equal to the flag name.
// 3. Add your flag's description etc. to the stringFlags, intFlags, or boolFlags slices.
const (
	// Flag names.
	AllowForkPRsFlag           = "allow-fork-prs"
	AllowRepoConfigFlag        = "allow-repo-config"
	AtlantisURLFlag            = "atlantis-url"
	AutomergeFlag              = "automerge"
	BitbucketBaseURLFlag       = "bitbucket-base-url"
	BitbucketTokenFlag         = "bitbucket-token"
	BitbucketUserFlag          = "bitbucket-user"
	BitbucketWebhookSecretFlag = "bitbucket-webhook-secret"
	ConfigFlag                 = "config"
	CheckoutStrategyFlag       = "checkout-strategy"
	DataDirFlag                = "data-dir"
	DefaultTFVersionFlag       = "default-tf-version"
	DisableApplyAllFlag        = "disable-apply-all"
	GHHostnameFlag             = "gh-hostname"
	GHTokenFlag                = "gh-token"
	GHUserFlag                 = "gh-user"
	GHWebhookSecretFlag        = "gh-webhook-secret" // nolint: gosec
	GitlabHostnameFlag         = "gitlab-hostname"
	GitlabTokenFlag            = "gitlab-token"
	GitlabUserFlag             = "gitlab-user"
	GitlabWebhookSecretFlag    = "gitlab-webhook-secret" // nolint: gosec
	LogLevelFlag               = "log-level"
	PortFlag                   = "port"
	RepoConfigFlag             = "repo-config"
	RepoConfigJSONFlag         = "repo-config-json"
	RepoWhitelistFlag          = "repo-whitelist"
	RequireApprovalFlag        = "require-approval"
	RequireMergeableFlag       = "require-mergeable"
	SilenceWhitelistErrorsFlag = "silence-whitelist-errors"
	SlackTokenFlag             = "slack-token"
	SSLCertFileFlag            = "ssl-cert-file"
	SSLKeyFileFlag             = "ssl-key-file"
	TFETokenFlag               = "tfe-token"

	// Flag defaults.
	DefaultCheckoutStrategy = "branch"
	DefaultBitbucketBaseURL = bitbucketcloud.BaseURL
	DefaultDataDir          = "~/.atlantis"
	DefaultGHHostname       = "github.com"
	DefaultGitlabHostname   = "gitlab.com"
	DefaultLogLevel         = "info"
	DefaultPort             = 4141
)

var stringFlags = map[string]stringFlag{
	AtlantisURLFlag: {
		description: "URL that Atlantis can be reached at. Defaults to http://$(hostname):$port where $port is from --" + PortFlag + ". Supports a base path ex. https://example.com/basepath.",
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
			" If set to merge, Atlantis will check out the destination branch of the pull request (ex. master)" +
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
	GHHostnameFlag: {
		description:  "Hostname of your Github Enterprise installation. If using github.com, no need to set.",
		defaultValue: DefaultGHHostname,
	},
	GHUserFlag: {
		description: "GitHub username of API user.",
	},
	GHTokenFlag: {
		description: "GitHub token of API user. Can also be specified via the ATLANTIS_GH_TOKEN environment variable.",
	},
	GHWebhookSecretFlag: {
		description: "Secret used to validate GitHub webhooks (see https://developer.github.com/webhooks/securing/)." +
			" SECURITY WARNING: If not specified, Atlantis won't be able to validate that the incoming webhook call came from GitHub. " +
			"This means that an attacker could spoof calls to Atlantis and cause it to perform malicious actions. " +
			"Should be specified via the ATLANTIS_GH_WEBHOOK_SECRET environment variable.",
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
	LogLevelFlag: {
		description:  "Log level. Either debug, info, warn, or error.",
		defaultValue: DefaultLogLevel,
	},
	RepoConfigFlag: {
		description: "Path to a repo config file, used to customize how Atlantis runs on each repo. See runatlantis.io/docs for more details.",
	},
	RepoConfigJSONFlag: {
		description: "Specify repo config as a JSON string. Useful if you don't want to write a config file to disk.",
	},
	RepoWhitelistFlag: {
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
	TFETokenFlag: {
		description: "API token for Terraform Enterprise. This will be used to generate a ~/.terraformrc file." +
			" Only set if using TFE as a backend." +
			" Should be specified via the ATLANTIS_TFE_TOKEN environment variable for security.",
	},
	DefaultTFVersionFlag: {
		description: "Terraform version to default to (ex. v0.12.0). Will download if not yet on disk." +
			" If not set, Atlantis uses the terraform binary in its PATH.",
	},
}

var boolFlags = map[string]boolFlag{
	AllowForkPRsFlag: {
		description:  "Allow Atlantis to run on pull requests from forks. A security issue for public repos.",
		defaultValue: false,
	},
	AllowRepoConfigFlag: {
		description: "Allow repositories to use atlantis.yaml files to customize the commands Atlantis runs." +
			" Should only be enabled in a trusted environment since it enables a pull request to run arbitrary commands" +
			" on the Atlantis server.",
		defaultValue: false,
		hidden:       true,
	},
	AutomergeFlag: {
		description:  "Automatically merge pull requests when all plans are successfully applied.",
		defaultValue: false,
	},
	DisableApplyAllFlag: {
		description:  "Disable \"atlantis apply\" command so a specific project/workspace/directory has to be specified for applies.",
		defaultValue: false,
	},
	RequireApprovalFlag: {
		description:  "Require pull requests to be \"Approved\" before allowing the apply command to be run.",
		defaultValue: false,
		hidden:       true,
	},
	RequireMergeableFlag: {
		description:  "Require pull requests to be mergeable before allowing the apply command to be run.",
		defaultValue: false,
		hidden:       true,
	},
	SilenceWhitelistErrorsFlag: {
		description:  "Silences the posting of whitelist error comments.",
		defaultValue: false,
	},
}
var intFlags = map[string]intFlag{
	PortFlag: {
		description:  "Port to bind to.",
		defaultValue: DefaultPort,
	},
}

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
	Logger          *logging.SimpleLogger
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
	c.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
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
	s.setDefaults(&userConfig)

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
	if err := s.deprecationWarnings(&userConfig); err != nil {
		return err
	}
	s.securityWarnings(&userConfig)
	s.trimAtSymbolFromUsers(&userConfig)

	// Config looks good. Start the server.
	server, err := s.ServerCreator.NewServer(userConfig, server.Config{
		AllowForkPRsFlag:     AllowForkPRsFlag,
		AtlantisURLFlag:      AtlantisURLFlag,
		AtlantisVersion:      s.AtlantisVersion,
		DefaultTFVersionFlag: DefaultTFVersionFlag,
		RepoConfigJSONFlag:   RepoConfigJSONFlag,
	})
	if err != nil {
		return errors.Wrap(err, "initializing server")
	}
	return server.Start()
}

func (s *ServerCmd) setDefaults(c *server.UserConfig) {
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
	if c.BitbucketBaseURL == "" {
		c.BitbucketBaseURL = DefaultBitbucketBaseURL
	}
	if c.LogLevel == "" {
		c.LogLevel = DefaultLogLevel
	}
	if c.Port == 0 {
		c.Port = DefaultPort
	}
}

func (s *ServerCmd) validate(userConfig server.UserConfig) error {
	logLevel := userConfig.LogLevel
	if logLevel != "debug" && logLevel != "info" && logLevel != "warn" && logLevel != "error" {
		return errors.New("invalid log level: not one of debug, info, warn, error")
	}
	checkoutStrat := userConfig.CheckoutStrategy
	if checkoutStrat != "branch" && checkoutStrat != "merge" {
		return errors.New("invalid checkout strategy: not one of branch or merge")
	}

	if (userConfig.SSLKeyFile == "") != (userConfig.SSLCertFile == "") {
		return fmt.Errorf("--%s and --%s are both required for ssl", SSLKeyFileFlag, SSLCertFileFlag)
	}

	// The following combinations are valid.
	// 1. github user and token set
	// 2. gitlab user and token set
	// 3. bitbucket user and token set
	// 4. any combination of the above
	vcsErr := fmt.Errorf("--%s/--%s or --%s/--%s or --%s/--%s must be set", GHUserFlag, GHTokenFlag, GitlabUserFlag, GitlabTokenFlag, BitbucketUserFlag, BitbucketTokenFlag)
	if ((userConfig.GithubUser == "") != (userConfig.GithubToken == "")) || ((userConfig.GitlabUser == "") != (userConfig.GitlabToken == "")) || ((userConfig.BitbucketUser == "") != (userConfig.BitbucketToken == "")) {
		return vcsErr
	}
	// At this point, we know that there can't be a single user/token without
	// its partner, but we haven't checked if any user/token is set at all.
	if userConfig.GithubUser == "" && userConfig.GitlabUser == "" && userConfig.BitbucketUser == "" {
		return vcsErr
	}

	if userConfig.RepoWhitelist == "" {
		return fmt.Errorf("--%s must be set for security purposes", RepoWhitelistFlag)
	}
	if strings.Contains(userConfig.RepoWhitelist, "://") {
		return fmt.Errorf("--%s cannot contain ://, should be hostnames only", RepoWhitelistFlag)
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
	} {
		if strings.Contains(token, "\n") {
			s.Logger.Warn("--%s contains a newline which is usually unintentional", name)
		}
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

// trimAtSymbolFromUsers trims @ from the front of the github and gitlab usernames
func (s *ServerCmd) trimAtSymbolFromUsers(userConfig *server.UserConfig) {
	userConfig.GithubUser = strings.TrimPrefix(userConfig.GithubUser, "@")
	userConfig.GitlabUser = strings.TrimPrefix(userConfig.GitlabUser, "@")
	userConfig.BitbucketUser = strings.TrimPrefix(userConfig.BitbucketUser, "@")
}

func (s *ServerCmd) securityWarnings(userConfig *server.UserConfig) {
	if userConfig.GithubUser != "" && userConfig.GithubWebhookSecret == "" && !s.SilenceOutput {
		s.Logger.Warn("no GitHub webhook secret set. This could allow attackers to spoof requests from GitHub")
	}
	if userConfig.GitlabUser != "" && userConfig.GitlabWebhookSecret == "" && !s.SilenceOutput {
		s.Logger.Warn("no GitLab webhook secret set. This could allow attackers to spoof requests from GitLab")
	}
	if userConfig.BitbucketUser != "" && userConfig.BitbucketBaseURL != DefaultBitbucketBaseURL && userConfig.BitbucketWebhookSecret == "" && !s.SilenceOutput {
		s.Logger.Warn("no Bitbucket webhook secret set. This could allow attackers to spoof requests from Bitbucket")
	}
	if userConfig.BitbucketUser != "" && userConfig.BitbucketBaseURL == DefaultBitbucketBaseURL && !s.SilenceOutput {
		s.Logger.Warn("Bitbucket Cloud does not support webhook secrets. This could allow attackers to spoof requests from Bitbucket. Ensure you are whitelisting Bitbucket IPs")
	}
}

// deprecationWarnings prints a warning if flags that are deprecated are
// being used. Right now this only applies to flags that have been made obsolete
// due to server-side config.
func (s *ServerCmd) deprecationWarnings(userConfig *server.UserConfig) error {
	var applyReqs []string
	var deprecatedFlags []string
	if userConfig.RequireApproval {
		deprecatedFlags = append(deprecatedFlags, RequireApprovalFlag)
		applyReqs = append(applyReqs, valid.ApprovedApplyReq)
	}
	if userConfig.RequireMergeable {
		deprecatedFlags = append(deprecatedFlags, RequireMergeableFlag)
		applyReqs = append(applyReqs, valid.MergeableApplyReq)
	}

	// Build up strings with what the recommended yaml and json config should
	// be instead of using the deprecated flags.
	yamlCfg := "---\nrepos:\n- id: /.*/"
	jsonCfg := `{"repos":[{"id":"/.*/"`
	if len(applyReqs) > 0 {
		yamlCfg += fmt.Sprintf("\n  apply_requirements: [%s]", strings.Join(applyReqs, ", "))
		jsonCfg += fmt.Sprintf(`, "apply_requirements":["%s"]`, strings.Join(applyReqs, "\", \""))

	}
	if userConfig.AllowRepoConfig {
		deprecatedFlags = append(deprecatedFlags, AllowRepoConfigFlag)
		yamlCfg += "\n  allowed_overrides: [apply_requirements, workflow]\n  allow_custom_workflows: true"
		jsonCfg += `, "allowed_overrides":["apply_requirements","workflow"], "allow_custom_workflows":true`
	}
	jsonCfg += "}]}"

	if len(deprecatedFlags) > 0 {
		warning := "WARNING: "
		if len(deprecatedFlags) == 1 {
			warning += fmt.Sprintf("Flag --%s has been deprecated.", deprecatedFlags[0])
		} else {
			warning += fmt.Sprintf("Flags --%s and --%s have been deprecated.", strings.Join(deprecatedFlags[0:len(deprecatedFlags)-1], ", --"), deprecatedFlags[len(deprecatedFlags)-1:][0])
		}
		warning += fmt.Sprintf("\nCreate a --%s file with the following config instead:\n\n%s\n\nor use --%s='%s'\n",
			RepoConfigFlag,
			yamlCfg,
			RepoConfigJSONFlag,
			jsonCfg,
		)
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
