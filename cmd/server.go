package cmd

import (
	"fmt"
	"github.com/hootsuite/atlantis/server"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
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
	ghPasswordFlag      = "gh-password"
	ghUserFlag          = "gh-user"
	lockingBackendFlag  = "locking-backend"
	lockingTableFlag    = "locking-dynamodb-table"
	logLevelFlag        = "log-level"
	portFlag            = "port"
	requireApprovalFlag = "require-approval"
	s3BucketFlag        = "s3-bucket"
	scratchDirFlag      = "scratch-dir"
	sshKeyFlag          = "ssh-key"
)

var stringFlags = []stringFlag{
	{
		name:        atlantisURLFlag,
		description: "Url that Atlantis can be reached at. Defaults to http://$(hostname):$port where $port is the port flag.",
	},
	{
		name:        awsAssumeRoleFlag,
		description: "The Amazon Resource Name (`arn`) to assume when running Terraform commands. If not specified, will use AWS credentials via environment variables, or credentials files.",
	},
	{
		name:        awsRegionFlag,
		description: "The Amazon region to connect to for API actions.",
		value:       "us-east-1",
	},
	{
		name:        configFlag,
		description: "Config file.",
	},
	{
		name:        dataDirFlag,
		description: "Path to directory to store Atlantis data.",
		value:       "/var/lib/atlantis",
	},
	{
		name:        ghHostnameFlag,
		description: "Hostname of Github installation.",
		value:       "api.github.com",
	},
	{
		name:        ghPasswordFlag,
		description: "GitHub password of API user. Can also be specified via the ATLANTIS_GH_PASSWORD environment variable.",
		env:         "ATLANTIS_GH_PASSWORD",
	},
	{
		name:        ghUserFlag,
		description: "GitHub username of API user. Can also be specified via the ATLANTIS_GH_USER environment variable.",
		env:         "ATLANTIS_GH_USER",
	},
	{
		name:        lockingBackendFlag,
		description: "Which backend to use for locking. file or dynamodb.",
		value:       "file",
	},
	{
		name:        lockingTableFlag,
		description: "Name of table in DynamoDB to use for locking. Only read if locking-backend is set to dynamodb.",
		value:       "atlantis-locks",
	},
	{
		name:        logLevelFlag,
		description: "Log level. Either debug, info, warn, or error.",
		value:       "warn",
	},
	{
		name:        s3BucketFlag,
		description: "The S3 bucket name to store atlantis data (terraform plans, terraform state, etc).",
		value:       "atlantis",
	},
	{
		name:        scratchDirFlag,
		description: "Path to directory to use as a temporary workspace for checking out repos.",
		value:       "/tmp/atlantis",
	},
	{
		name:        sshKeyFlag,
		description: "Path to SSH key used for GitHub.",
	},
}
var boolFlags = []boolFlag{
	{
		name:        requireApprovalFlag,
		description: "Require pull requests to be \"Approved\" before allowing the apply command to be run.",
		value:       true,
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
Config values are overridden by environment variables which in turn are overridden by flags.`,
	SilenceUsage: true,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		// if passed a config file then try and load it
		configFile := viper.GetString(configFlag)
		if configFile != "" {
			viper.SetConfigFile(configFile)
			if err := viper.ReadInConfig(); err != nil {
				return fmt.Errorf("invalid config: reading %s: %s", configFile, err)
			}
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
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

		// config looks good, start the server
		server, err := server.NewServer(config)
		if err != nil {
			return errors.Wrap(err, "Failed to initialize server")
		}
		return server.Start()
	},
}

func init() {
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
	if config.GitHubUser == "" {
		return fmt.Errorf("%s must be set", ghUserFlag)
	}
	if config.GitHubPassword == "" {
		return fmt.Errorf("%s must be set", ghPasswordFlag)
	}
	if config.LockingBackend != server.LockingFileBackend && config.LockingBackend != server.LockingDynamoDBBackend {
		return fmt.Errorf("unsupported locking backend %q: not one of %q or %q", config.LockingBackend, server.LockingFileBackend, server.LockingDynamoDBBackend)
	}
	return nil
}

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
