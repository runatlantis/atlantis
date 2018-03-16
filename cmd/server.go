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
//
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// To add a new flag you must:
// 1. Add a const with the flag name (in alphabetic order).
// 2. Add a new field to server.UserConfig and set the mapstructure tag equal to the flag name.
// 3. Add your flag's description etc. to the stringFlags, intFlags, or boolFlags slices.
const (
	AtlantisURLFlag     = "atlantis-url"
	AllowForkPRsFlag    = "allow-fork-prs"
	ConfigFlag          = "config"
	DataDirFlag         = "data-dir"
	GHHostnameFlag      = "gh-hostname"
	GHTokenFlag         = "gh-token"
	GHUserFlag          = "gh-user"
	GHWebHookSecret     = "gh-webhook-secret" // nolint: gas
	GitlabHostnameFlag  = "gitlab-hostname"
	GitlabTokenFlag     = "gitlab-token"
	GitlabUserFlag      = "gitlab-user"
	GitlabWebHookSecret = "gitlab-webhook-secret"
	LogLevelFlag        = "log-level"
	PortFlag            = "port"
	RepoWhitelistFlag   = "repo-whitelist"
	RequireApprovalFlag = "require-approval"
	SSLCertFileFlag     = "ssl-cert-file"
	SSLKeyFileFlag      = "ssl-key-file"
)

const RedTermStart = "\033[31m"
const RedTermEnd = "\033[39m"

var stringFlags = []stringFlag{
	{
		name:        AtlantisURLFlag,
		description: "URL that Atlantis can be reached at. Defaults to http://$(hostname):$port where $port is from --" + PortFlag + ".",
	},
	{
		name:        ConfigFlag,
		description: "Path to config file.",
	},
	{
		name:        DataDirFlag,
		description: "Path to directory to store Atlantis data.",
		value:       "~/.atlantis",
	},
	{
		name:        GHHostnameFlag,
		description: "Hostname of your Github Enterprise installation. If using github.com, no need to set.",
		value:       "github.com",
	},
	{
		name:        GHUserFlag,
		description: "GitHub username of API user.",
	},
	{
		name:        GHTokenFlag,
		description: "GitHub token of API user. Can also be specified via the ATLANTIS_GH_TOKEN environment variable.",
		env:         "ATLANTIS_GH_TOKEN",
	},
	{
		name: GHWebHookSecret,
		description: "Secret used to validate GitHub webhooks (see https://developer.github.com/webhooks/securing/)." +
			" SECURITY WARNING: If not specified, Atlantis won't be able to validate that the incoming webhook call came from GitHub. " +
			"This means that an attacker could spoof calls to Atlantis and cause it to perform malicious actions. " +
			"Should be specified via the ATLANTIS_GH_WEBHOOK_SECRET environment variable.",
		env: "ATLANTIS_GH_WEBHOOK_SECRET",
	},
	{
		name:        GitlabHostnameFlag,
		description: "Hostname of your GitLab Enterprise installation. If using gitlab.com, no need to set.",
		value:       "gitlab.com",
	},
	{
		name:        GitlabUserFlag,
		description: "GitLab username of API user.",
	},
	{
		name:        GitlabTokenFlag,
		description: "GitLab token of API user. Can also be specified via the ATLANTIS_GITLAB_TOKEN environment variable.",
		env:         "ATLANTIS_GITLAB_TOKEN",
	},
	{
		name: GitlabWebHookSecret,
		description: "Optional secret used to validate GitLab webhooks." +
			" SECURITY WARNING: If not specified, Atlantis won't be able to validate that the incoming webhook call came from GitLab. " +
			"This means that an attacker could spoof calls to Atlantis and cause it to perform malicious actions. " +
			"Should be specified via the ATLANTIS_GITLAB_WEBHOOK_SECRET environment variable.",
		env: "ATLANTIS_GITLAB_WEBHOOK_SECRET",
	},
	{
		name:        LogLevelFlag,
		description: "Log level. Either debug, info, warn, or error.",
		value:       "info",
	},
	{
		name: RepoWhitelistFlag,
		description: "Comma separated list of repositories that Atlantis will operate on. " +
			"The format is {hostname}/{owner}/{repo}, ex. github.com/runatlantis/atlantis. '*' matches any characters until the next comma and can be used for example to whitelist " +
			"all repos: '*' (not recommended), an entire hostname: 'internalgithub.com/*' or an organization: 'github.com/runatlantis/*'.",
	},
	{
		name:        SSLCertFileFlag,
		description: "File containing x509 Certificate used for serving HTTPS. If the cert is signed by a CA, the file should be the concatenation of the server's certificate, any intermediates, and the CA's certificate.",
	},
	{
		name:        SSLKeyFileFlag,
		description: fmt.Sprintf("File containing x509 private key matching --%s.", SSLCertFileFlag),
	},
}
var boolFlags = []boolFlag{
	{
		name:        AllowForkPRsFlag,
		description: "Allow Atlantis to run on pull requests from forks. A security issue for public repos.",
		value:       false,
	},
	{
		name:        RequireApprovalFlag,
		description: "Require pull requests to be \"Approved\" before allowing the apply command to be run.",
		value:       false,
	},
}
var intFlags = []intFlag{
	{
		name:        PortFlag,
		description: "Port to bind to.",
		value:       4141,
	},
}

type stringFlag struct {
	name        string
	description string
	value       string
	env         string
}
type intFlag struct {
	name        string
	description string
	value       int
}
type boolFlag struct {
	name        string
	description string
	value       bool
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
		Use:   "server",
		Short: "Start the atlantis server",
		Long: `Start the atlantis server

Flags can also be set in a yaml config file (see --` + ConfigFlag + `).
Config file values are overridden by environment variables which in turn are overridden by flags.`,
		SilenceErrors: true,
		SilenceUsage:  s.SilenceOutput,
		PreRunE: s.withErrPrint(func(cmd *cobra.Command, args []string) error {
			return s.preRun()
		}),
		RunE: s.withErrPrint(func(cmd *cobra.Command, args []string) error {
			return s.run()
		}),
	}
	// Replace the call in their template to use the usage function that wraps
	// columns to make for a nicer output.
	usageWithWrappedCols := strings.Replace(c.UsageTemplate(), ".FlagUsages", ".FlagUsagesWrapped 120", -1)
	c.SetUsageTemplate(usageWithWrappedCols)

	// If a user passes in an invalid flag, tell them what the flag was.
	c.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
		fmt.Fprintf(os.Stderr, "\033[31mError: %s\033[39m\n\n", err.Error())
		return err
	})

	// Set string flags.
	for _, f := range stringFlags {
		c.Flags().String(f.name, f.value, "> "+f.description)
		if f.env != "" {
			s.Viper.BindEnv(f.name, f.env) // nolint: errcheck
		}
		s.Viper.BindPFlag(f.name, c.Flags().Lookup(f.name)) // nolint: errcheck
	}

	// Set int flags.
	for _, f := range intFlags {
		c.Flags().Int(f.name, f.value, "> "+f.description)
		s.Viper.BindPFlag(f.name, c.Flags().Lookup(f.name)) // nolint: errcheck
	}

	// Set bool flags.
	for _, f := range boolFlags {
		c.Flags().Bool(f.name, f.value, "> "+f.description)
		s.Viper.BindPFlag(f.name, c.Flags().Lookup(f.name)) // nolint: errcheck
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
	if err := s.validate(userConfig); err != nil {
		return err
	}
	if err := s.setAtlantisURL(&userConfig); err != nil {
		return err
	}
	if err := s.setDataDir(&userConfig); err != nil {
		return err
	}
	s.securityWarnings(&userConfig)
	s.trimAtSymbolFromUsers(&userConfig)

	// Config looks good. Start the server.
	server, err := s.ServerCreator.NewServer(userConfig, server.Config{
		AllowForkPRsFlag: AllowForkPRsFlag,
		AtlantisVersion:  s.AtlantisVersion,
	})
	if err != nil {
		return errors.Wrap(err, "initializing server")
	}
	return server.Start()
}

// nolint: gocyclo
func (s *ServerCmd) validate(userConfig server.UserConfig) error {
	logLevel := userConfig.LogLevel
	if logLevel != "debug" && logLevel != "info" && logLevel != "warn" && logLevel != "error" {
		return errors.New("invalid log level: not one of debug, info, warn, error")
	}

	if (userConfig.SSLKeyFile == "") != (userConfig.SSLCertFile == "") {
		return fmt.Errorf("--%s and --%s are both required for ssl", SSLKeyFileFlag, SSLCertFileFlag)
	}

	// The following combinations are valid.
	// 1. github user and token set
	// 2. gitlab user and token set
	// 3. all 4 set
	vcsErr := fmt.Errorf("--%s and --%s or --%s and --%s must be set", GHUserFlag, GHTokenFlag, GitlabUserFlag, GitlabTokenFlag)
	if ((userConfig.GithubUser == "") != (userConfig.GithubToken == "")) || ((userConfig.GitlabUser == "") != (userConfig.GitlabToken == "")) {
		return vcsErr
	}
	// At this point, we know that there can't be a single user/token without
	// its partner, but we haven't checked if any user/token is set at all.
	if userConfig.GithubUser == "" && userConfig.GitlabUser == "" {
		return vcsErr
	}

	if userConfig.RepoWhitelist == "" {
		return fmt.Errorf("--%s must be set for security purposes", RepoWhitelistFlag)
	}

	return nil
}

// setAtlantisURL sets the externally accessible URL for atlantis.
func (s *ServerCmd) setAtlantisURL(userConfig *server.UserConfig) error {
	if userConfig.AtlantisURL == "" {
		hostname, err := os.Hostname()
		if err != nil {
			return fmt.Errorf("Failed to determine hostname: %v", err)
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
}

func (s *ServerCmd) securityWarnings(userConfig *server.UserConfig) {
	if userConfig.GithubUser != "" && userConfig.GithubWebHookSecret == "" && !s.SilenceOutput {
		fmt.Fprintf(os.Stderr, "%s[WARN] No GitHub webhook secret set. This could allow attackers to spoof requests from GitHub. See https://git.io/vAF3t%s\n", RedTermStart, RedTermEnd)
	}
	if userConfig.GitlabUser != "" && userConfig.GitlabWebHookSecret == "" && !s.SilenceOutput {
		fmt.Fprintf(os.Stderr, "%s[WARN] No GitLab webhook secret set. This could allow attackers to spoof requests from GitLab. See https://git.io/vAF3t%s\n", RedTermStart, RedTermEnd)
	}
}

// withErrPrint prints out any errors to a terminal in red.
func (s *ServerCmd) withErrPrint(f func(*cobra.Command, []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := f(cmd, args)
		if err != nil && !s.SilenceOutput {
			fmt.Fprintf(os.Stderr, "%s[ERROR] %s%s\n\n", RedTermStart, err.Error(), RedTermEnd)
		}
		return err
	}
}
