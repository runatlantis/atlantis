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
	AtlantisURLFlag     = "atlantis-url"
	ConfigFlag          = "config"
	DataDirFlag         = "data-dir"
	GHHostnameFlag      = "gh-hostname"
	GHTokenFlag         = "gh-token"
	GHUserFlag          = "gh-user"
	GHWebHookSecret     = "gh-webhook-secret"
	LogLevelFlag        = "log-level"
	PortFlag            = "port"
	RequireApprovalFlag = "require-approval"
)

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
		name:        GHTokenFlag,
		description: "[REQUIRED] GitHub token of API user. Can also be specified via the ATLANTIS_GH_TOKEN environment variable.",
		env:         "ATLANTIS_GH_TOKEN",
	},
	{
		name:        GHUserFlag,
		description: "[REQUIRED] GitHub username of API user.",
	},
	{
		name:        GHWebHookSecret,
		description: "Optional secret used for GitHub webhooks (see https://developer.github.com/webhooks/securing/). If not specified, Atlantis won't validate the incoming webhook call.",
		env:         "ATLANTIS_GH_WEBHOOK_SECRET",
	},
	{
		name:        LogLevelFlag,
		description: "Log level. Either debug, info, warn, or error.",
		value:       "info",
	},
}
var boolFlags = []boolFlag{
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
	Viper *viper.Viper
	// SilenceOutput set to true means nothing gets printed.
	// Useful for testing to keep the logs clean.
	SilenceOutput bool
}

// ServerCreator creates servers.
// It's an abstraction to help us test.
type ServerCreator interface {
	NewServer(config server.ServerConfig) (ServerStarter, error)
}

// DefaultServerCreator is the concrete implementation of ServerCreator.
type DefaultServerCreator struct{}

// ServerStarter is for starting up a server.
// It's an abstraction to help us test.
type ServerStarter interface {
	Start() error
}

// NewServer returns the real Atlantis server object.
func (d *DefaultServerCreator) NewServer(config server.ServerConfig) (ServerStarter, error) {
	return server.NewServer(config)
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
		PreRunE: s.withErrPrint(func(cmd *cobra.Command, args []string) error {
			return s.preRun()
		}),
		RunE: s.withErrPrint(func(cmd *cobra.Command, args []string) error {
			return s.run()
		}),
	}

	// if a user passes in an invalid flag, tell them what the flag was
	c.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
		fmt.Fprintf(os.Stderr, "\033[31mError: %s\033[39m\n\n", err.Error())
		return err
	})

	// set string flags
	for _, f := range stringFlags {
		c.Flags().String(f.name, f.value, f.description)
		if f.env != "" {
			s.Viper.BindEnv(f.name, f.env)
		}
		s.Viper.BindPFlag(f.name, c.Flags().Lookup(f.name))
	}

	// set int flags
	for _, f := range intFlags {
		c.Flags().Int(f.name, f.value, f.description)
		s.Viper.BindPFlag(f.name, c.Flags().Lookup(f.name))
	}

	// set bool flags
	for _, f := range boolFlags {
		c.Flags().Bool(f.name, f.value, f.description)
		s.Viper.BindPFlag(f.name, c.Flags().Lookup(f.name))
	}

	return c
}

func (s *ServerCmd) preRun() error {
	// if passed a config file then try and load it
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
	var config server.ServerConfig
	if err := s.Viper.Unmarshal(&config); err != nil {
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
	server, err := s.ServerCreator.NewServer(config)
	if err != nil {
		return errors.Wrap(err, "initializing server")
	}
	return server.Start()
}

func validate(config server.ServerConfig) error {
	logLevel := config.LogLevel
	if logLevel != "debug" && logLevel != "info" && logLevel != "warn" && logLevel != "error" {
		return errors.New("invalid log level: not one of debug, info, warn, error")
	}
	if config.GithubUser == "" {
		return fmt.Errorf("--%s must be set", GHUserFlag)
	}
	if config.GithubToken == "" {
		return fmt.Errorf("--%s must be set", GHTokenFlag)
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
func (s *ServerCmd) withErrPrint(f func(*cobra.Command, []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := f(cmd, args)
		if err != nil && !s.SilenceOutput {
			fmt.Fprintf(os.Stderr, "\033[31mError: %s\033[39m\n\n", err.Error())
		}
		return err
	}
}
