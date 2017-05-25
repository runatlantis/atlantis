package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli"
)

const (
	ghUsernameFlag       = "gh-username"
	ghPasswordFlag       = "gh-password"
	ghHostnameFlag       = "gh-hostname"
	sshKeyFlag           = "ssh-key"
	portFlag             = "port"
	awsAssumeRoleFlag    = "aws-assume-role-arn"
	scratchDirFlag       = "scratch-dir"
	awsRegionFlag        = "aws-region"
	s3BucketFlag         = "s3-bucket"
	logLevelFlag         = "log-level"
	version              = "2.0.0"
	configFileFlag       = "config-file"
	requireApprovalFlag  = "require-approval"

	defaultGHHostname = "github.com"
	defaultPort       = 4141
	defaultScratchDir = "/tmp/atlantis"
	defaultRegion     = "us-east-1"
	defaultS3Bucket   = "atlantis"
	defaultLogLevel   = "info"
)

type AtlantisConfig struct {
	GithubHostname   *string                 `json:"gh-hostname"`
	GithubUsername   *string                 `json:"gh-username"`
	GithubPassword   *string                 `json:"gh-password"`
	SSHKey           *string                 `json:"ssh-key"`
	AssumeRole       *string                 `json:"aws-assume-role-arn"`
	Port             int                     `json:"port"`
	ScratchDir       *string                 `json:"scratch-dir"`
	AWSRegion        *string                 `json:"aws-region"`
	S3Bucket         *string                 `json:"s3-bucket"`
	LogLevel         *string                 `json:"log-level"`
	RequireApproval  bool                    `json:"require-approval"`
}

func main() {
	app := cli.NewApp()
	configureCli(app)
	app.Action = mainAction
	app.Run(os.Args)
}

// mainAction is the only "action" for this cli program. It starts the web server
func mainAction(c *cli.Context) error {
	conf, err := validateConfig(c)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return NewServer(conf).Start()
}

// configureCli sets up the flags, name of the cli, etc.
func configureCli(app *cli.App) {
	app.Name = "atlantis"
	app.Usage = "Terraform collaboration tool"
	app.Version = version
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   ghUsernameFlag,
			Usage:  "GitHub `username` of API user",
			EnvVar: "GITHUB_USERNAME",
		},
		cli.StringFlag{
			Name:   ghPasswordFlag,
			Usage:  "GitHub `password` of API user",
			EnvVar: "GITHUB_PASSWORD",
		},
		cli.StringFlag{
			Name:  ghHostnameFlag,
			Value: "github.com",
			Usage: "GitHub `hostname`",
		},
		cli.StringFlag{
			Name:  sshKeyFlag,
			Usage: "Path to SSH key used for Github.",
		},
		cli.StringFlag{
			Name:  awsAssumeRoleFlag,
			Usage: "The Amazon Resource Name (`arn`) to assume when running Terraform commands. If not specified, will use AWS credentials via environment variables, or credentials files",
		},
		cli.IntFlag{
			Name:  fmt.Sprintf("%s, p", portFlag),
			Value: defaultPort,
			Usage: "The `port` to listen for GitHub hooks on",
		},
		cli.StringFlag{
			Name:  scratchDirFlag,
			Value: defaultScratchDir,
			Usage: "The `directory` to use as a temporary workspace",
		},
		cli.StringFlag{
			Name:  awsRegionFlag,
			Value: defaultRegion,
			Usage: "The Amazon `region` to connect to for API actions.",
		},
		cli.StringFlag{
			Name:  s3BucketFlag,
			Value: defaultS3Bucket,
			Usage: "The S3 bucket name to store atlantis data (terraform plans, terraform state, etc.)",
		},
		cli.BoolFlag{
			Name:  requireApprovalFlag,
			Usage: "Require pull requests to be \"Approved\" before allowing the apply command to be run",
		},
		cli.StringFlag{
			Name:  logLevelFlag,
			Value: defaultLogLevel,
			Usage: "debug, info, warn, or error",
		},
		cli.StringFlag{
			Name:  configFileFlag,
			Usage: "Load configuration from `file`",
		},
		// todo: allow option of where to log
		// todo: network interface to bind to? by default localhost
	}
	cli.AppHelpTemplate = `{{.Name}} - {{.Usage}}

usage: {{.HelpName}} [options]
{{if .VisibleFlags}}
options:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}{{if .Version}}
version: {{.Version}}{{end}}
`
}

func validateConfig(c *cli.Context) (*ServerConfig, error) {
	// set defaults for non-required flags if not specified
	port := defaultPort
	ghHostname := defaultGHHostname
	ghUsername := ""
	ghPassword := ""
	scratchDir := defaultScratchDir
	awsRegion := defaultRegion
	s3Bucket := defaultS3Bucket
	logLevel := defaultLogLevel
	sshKey := ""
	assumeRole := ""
	requireApproval := false

	// check if config file flag is set
	if c.IsSet(configFileFlag) {
		configFile, err := os.Open(c.String(configFileFlag))
		if err != nil {
			return nil, fmt.Errorf("couldn't read config file: %v", err)
		}

		config := &AtlantisConfig{}
		jsonParser := json.NewDecoder(configFile)
		if err = jsonParser.Decode(config); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %v", err)
		}

		// set config from file
		if config.Port != 0 {
			port = config.Port
		}
		if config.GithubHostname != nil {
			ghHostname = *config.GithubHostname
		}
		if config.GithubUsername != nil {
			ghUsername = *config.GithubUsername
		}
		if config.GithubPassword != nil {
			ghPassword = *config.GithubPassword
		}
		if config.ScratchDir != nil {
			scratchDir = *config.ScratchDir
		}
		if config.AWSRegion != nil {
			awsRegion = *config.AWSRegion
		}
		if config.S3Bucket != nil {
			s3Bucket = *config.S3Bucket
		}
		if config.LogLevel != nil {
			logLevel = *config.LogLevel
		}
		if config.SSHKey != nil {
			sshKey = *config.SSHKey
		}
		if config.AssumeRole != nil {
			assumeRole = *config.AssumeRole
		}
		requireApproval = config.RequireApproval
	}

	// override config file if command line flags are set
	if c.IsSet(ghUsernameFlag) {
		ghUsername = c.String(ghUsernameFlag)
	}
	// check required flags
	if ghUsername == "" {
		return nil, fmt.Errorf("Error: must specify the --%s flag", ghUsernameFlag)
	}
	if c.IsSet(ghPasswordFlag) {
		ghPassword = c.String(ghPasswordFlag)
	}
	if ghPassword == "" {
		return nil, fmt.Errorf("Error: must specify the --%s flag", ghPasswordFlag)
	}
	if c.IsSet(portFlag) {
		port = c.Int(portFlag)
	}
	if c.IsSet(ghHostnameFlag) {
		ghHostname = c.String(ghHostnameFlag)
	}
	if c.IsSet(scratchDirFlag) {
		scratchDir = c.String(scratchDirFlag)
	}
	if c.IsSet(awsRegionFlag) {
		awsRegion = c.String(awsRegionFlag)
	}
	if c.IsSet(sshKeyFlag) {
		sshKey = c.String(sshKeyFlag)
	}
	if c.IsSet(awsAssumeRoleFlag) {
		assumeRole = c.String(awsAssumeRoleFlag)
	}
	if c.IsSet(s3BucketFlag) {
		s3Bucket = c.String(s3BucketFlag)
	}
	if c.IsSet(logLevelFlag) {
		logLevel = strings.ToLower(c.String(logLevelFlag))
	}
	if c.IsSet(requireApprovalFlag) {
		requireApproval = c.Bool(requireApprovalFlag)
	}

	if logLevel != "debug" && logLevel != "info" && logLevel != "warn" && logLevel != "error" {
		return nil, errors.New("Invalid log level")
	}

	return &ServerConfig{
		githubUsername:   ghUsername,
		githubPassword:   ghPassword,
		githubHostname:   ghHostname,
		sshKey:           sshKey,
		awsAssumeRole:    assumeRole,
		port:             port,
		scratchDir:       scratchDir,
		awsRegion:        awsRegion,
		s3Bucket:         s3Bucket,
		logLevel:         logLevel,
		requireApproval:  requireApproval,
	}, nil
}
