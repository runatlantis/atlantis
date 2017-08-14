package cmd

import (
	"fmt"
	"os"

	"strings"

	"github.com/hootsuite/atlantis/server"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// To add a new flag you must
// 1. Add a const with the flag name (in alphabetic order)
// 2. Add a new field to server.ServerConfig and set the mapstructure tag equal to the flag name
// 3. Add your flag's description etc. to the stringFlags, intFlags, or boolFlags slices
const (
	atlantisURLFlag     = "atlantis-url"
	awsAssumeRoleFlag   = "aws-assume-role-arn"
	awsRegionFlag       = "aws-region"
	configFlag          = "config"
	dataDirFlag         = "data-dir"
	ghHostnameFlag      = "gh-hostname"
	ghTokenFlag         = "gh-token"
	ghUserFlag          = "gh-user"
	ghWebHookSecret     = "gh-webhook-secret"
	logLevelFlag        = "log-level"
	portFlag            = "port"
	requireApprovalFlag = "require-approval"
)

var stringFlags = []stringFlag{
	{
		name:        atlantisURLFlag,
		description: "URL that Atlantis can be reached at. Defaults to http://$(hostname):$port where $port is from --" + portFlag + ".",
	},
	{
		name:        awsAssumeRoleFlag,
		description: "ARN of the role to assume when running Terraform against AWS. If not using assume role, no need to set.",
	},
	{
		name:        awsRegionFlag,
		description: "Amazon region to use for assume role. If not setting --" + awsAssumeRoleFlag + " then ignore.",
		value:       "us-east-1",
	},
	{
		name:        configFlag,
		description: "Path to config file.",
	},
	{
		name:        dataDirFlag,
		description: "Path to directory to store Atlantis data.",
		value:       "~/.atlantis",
	},
	{
		name:        ghHostnameFlag,
		description: "Hostname of your Github Enterprise installation. If using github.com, no need to set.",
		value:       "github.com",
	},
	{
		name:        ghTokenFlag,
		description: "[REQUIRED] GitHub token of API user. Can also be specified via the ATLANTIS_GH_TOKEN environment variable.",
		env:         "ATLANTIS_GH_TOKEN",
	},
	{
		name:        ghUserFlag,
		description: "[REQUIRED] GitHub username of API user.",
	},
	{
		name:        ghWebHookSecret,
		description: "Optional secret used for GitHub webhooks (see https://developer.github.com/webhooks/securing/). If not specified, Atlantis won't validate the incoming webhook call.",
		env:         "ATLANTIS_GH_WEBHOOK_SECRET",
	},
	{
		name:        logLevelFlag,
		description: "Log level. Either debug, info, warn, or error.",
		value:       "info",
	},
}
var boolFlags = []boolFlag{
	{
		name:        requireApprovalFlag,
		description: "Require pull requests to be \"Approved\" before allowing the apply command to be run.",
		value:       false,
	},
}
var intFlags = []intFlag{
	{
		name:        portFlag,
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

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the atlantis server",
	Long: `Start the atlantis server

Flags can also be set in a yaml config file (see --` + configFlag + `).
Config file values are overridden by environment variables which in turn are overridden by flags.`,
	SilenceErrors: true,

	PreRunE: withErrPrint(func(cmd *cobra.Command, args []string) error {
		// if passed a config file then try and load it
		configFile := viper.GetString(configFlag)
		if configFile != "" {
			viper.SetConfigFile(configFile)
			if err := viper.ReadInConfig(); err != nil {
				return errors.Wrapf(err, "invalid config: reading %s", configFile)
			}
		}
		return nil
	}),
	RunE: withErrPrint(func(cmd *cobra.Command, args []string) error {
		var config server.ServerConfig
		if err := viper.Unmarshal(&config); err != nil {
			return err
		}
		if err := validate(config); err != nil {
			return err
		}
		if err := setAtlantisURL(&config); err != nil {
			return err
		}
		sanitizeGithubUser(&config)

		// config looks good, start the server
		server, err := server.NewServer(config)
		if err != nil {
			return errors.Wrap(err, "initializing server")
		}
		return server.Start()
	}),
}

func init() {
	// if a user passes in an invalid flag, tell them what the flag was
	serverCmd.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
		fmt.Fprintf(os.Stderr, "\033[31mError: %s\033[39m\n\n", err.Error())
		return err
	})

	// set string flags
	for _, f := range stringFlags {
		serverCmd.Flags().String(f.name, f.value, f.description)
		if f.env != "" {
			viper.BindEnv(f.name, f.env)
		}
		viper.BindPFlag(f.name, serverCmd.Flags().Lookup(f.name))
	}

	// set int flags
	for _, f := range intFlags {
		serverCmd.Flags().Int(f.name, f.value, f.description)
		viper.BindPFlag(f.name, serverCmd.Flags().Lookup(f.name))
	}

	// set bool flags
	for _, f := range boolFlags {
		serverCmd.Flags().Bool(f.name, f.value, f.description)
		viper.BindPFlag(f.name, serverCmd.Flags().Lookup(f.name))
	}

	RootCmd.AddCommand(serverCmd)
}

func validate(config server.ServerConfig) error {
	logLevel := config.LogLevel
	if logLevel != "debug" && logLevel != "info" && logLevel != "warn" && logLevel != "error" {
		return errors.New("invalid log level: not one of debug, info, warn, error")
	}
	if config.GithubUser == "" {
		return fmt.Errorf("--%s must be set", ghUserFlag)
	}
	if config.GithubToken == "" {
		return fmt.Errorf("--%s must be set", ghTokenFlag)
	}
	return nil
}

// setAtlantisURL sets the externally accessible URL for atlantis.
func setAtlantisURL(config *server.ServerConfig) error {
	if config.AtlantisURL == "" {
		hostname, err := os.Hostname()
		if err != nil {
			return fmt.Errorf("Failed to determine hostname: %v", err)
		}
		config.AtlantisURL = fmt.Sprintf("http://%s:%d", hostname, config.Port)
	}
	return nil
}

// sanitizeGithubUser trims @ from the front of the github username if it exists.
func sanitizeGithubUser(config *server.ServerConfig) {
	config.GithubUser = strings.TrimPrefix(config.GithubUser, "@")
}

// withErrPrint prints out any errors to a terminal in red.
func withErrPrint(f func(*cobra.Command, []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := f(cmd, args); err != nil {
			fmt.Fprintf(os.Stderr, "\033[31mError: %s\033[39m\n\n", err.Error())
			return err
		}
		return nil
	}
}
