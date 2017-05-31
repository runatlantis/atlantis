package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli"
	"github.com/boltdb/bolt"
	"time"
	"github.com/pkg/errors"
	"path"
)

const (
	// flags
	ghUsernameFlag       = "gh-username"
	ghPasswordFlag      = "gh-password"
	ghHostnameFlag      = "gh-hostname"
	sshKeyFlag          = "ssh-key"
	portFlag            = "port"
	awsAssumeRoleFlag   = "aws-assume-role-arn"
	scratchDirFlag      = "scratch-dir"
	awsRegionFlag       = "aws-region"
	s3BucketFlag        = "s3-bucket"
	logLevelFlag        = "log-level"
	configFileFlag      = "config-file"
	requireApprovalFlag = "require-approval"
	atlantisURLFlag     = "atlantis-url"
	dataDirFlag         = "data-dir"
	lockingBackendFlag  = "locking-backend"
	lockingTableFlag    = "locking-table"

	// defaults
	boltDBRunLocksBucket   = "runLocks"
	version                = "2.0.0"
	defaultGHHostname      = "github.com"
	defaultPort            = 4141
	defaultScratchDir      = "/tmp/atlantis"
	defaultRegion          = "us-east-1"
	defaultS3Bucket        = "atlantis"
	defaultLogLevel        = "info"
	defaultDataDir         = "/var/lib/atlantis"
	fileLockingBackend     = "file"
	dynamoDBLockingBackend = "dynamodb"
	defaultLockingBackend  = fileLockingBackend
	defaultLockingTable    = "atlantis-locks"
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
	AtlantisURL      *string                 `json:"atlantis-url"`
	RequireApproval  bool                    `json:"require-approval"`
	DataDir          *string                 `json:"data-dir"`
	LockingBackend   *string                 `json:"locking-backend"`
	LockingTable     *string                 `json:"locking-table"`
}

func main() {
	app := cli.NewApp()
	configureCli(app)
	app.Action = mainAction
	app.Run(os.Args)
}

// mainAction is the only "action" for this cli program. It starts the web server and initializes the bolt db
func mainAction(c *cli.Context) error {
	hostname, err := os.Hostname()
	if err != nil {
		return cli.NewExitError(fmt.Errorf("Failed to determine hostname: %v", err), 1)
	}
	conf, err := validateConfig(c, hostname)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	var db *bolt.DB
	// if file is the locking backend then initialize the bolt DB
	if conf.lockingBackend == fileLockingBackend {
		if err := os.MkdirAll(conf.dataDir, 0755); err != nil {
			return cli.NewExitError(fmt.Errorf("Failed to create data dir: %v", err), 1)
		}
		db, err = bolt.Open(path.Join(conf.dataDir, "atlantis.db"), 0600, &bolt.Options{Timeout: 1 * time.Second})
		if err != nil {
			if err.Error() == "timeout" {
				return cli.NewExitError(errors.New("Failed to start Bolt DB due to a timeout. A likely cause is Atlantis already running"), 1)
			}
			return cli.NewExitError(fmt.Errorf("Failed to start Bolt DB: %v", err), 1)
		}
		if err = initDB(db); err != nil {
			return cli.NewExitError(fmt.Errorf("Failed to initialize Bolt DB: %v", err), 1)
		}
		defer db.Close()
	}
	server, err := NewServer(conf, db)
	if err != nil {
		return cli.NewExitError(fmt.Errorf("Failed to start server: %s", err), 1)
	}
	return server.Start()
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
		cli.StringFlag{
			Name:  atlantisURLFlag,
			Usage: "The url that Atlantis can be reached at. Defaults to http, hostname, and configured port",
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
		cli.StringFlag{
			Name: dataDirFlag,
			Value: defaultDataDir,
			Usage: "Directory to store Atlantis data",
		},
		cli.StringFlag{
			Name: lockingBackendFlag,
			Value: defaultLockingBackend,
			Usage: "Which backend to use for locking: file (default) or dynamodb",
		},
		cli.StringFlag{
			Name:  lockingTableFlag,
			Value: defaultLockingTable,
			Usage: "Name of the DynamoDB table used for locking. Only required if locking-backend is set to dynamodb. Defaults to 'atlantis-locks'",
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

func validateConfig(c *cli.Context, hostname string) (*ServerConfig, error) {
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
	atlantisURL := ""
	dataDir := defaultDataDir
	lockingBackend := defaultLockingBackend
	lockingTable := defaultLockingTable

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
		if config.AtlantisURL != nil {
			atlantisURL = *config.AtlantisURL
		}
		if config.DataDir != nil {
			dataDir = *config.DataDir
		}
		if config.LockingBackend != nil {
			lockingBackend = *config.LockingBackend
		}
		if config.LockingTable != nil {
			lockingTable = *config.LockingTable
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
	if c.IsSet(atlantisURLFlag) {
		atlantisURL = c.String(atlantisURLFlag)
	} else if atlantisURL == "" {
		atlantisURL = fmt.Sprintf("http://%s:%d", hostname, port)
	}
	if logLevel != "debug" && logLevel != "info" && logLevel != "warn" && logLevel != "error" {
		return nil, errors.New("Invalid log level")
	}
	if c.IsSet(dataDirFlag) {
		dataDir = c.String(dataDirFlag)
	}
	if c.IsSet(lockingBackendFlag) {
		lockingBackend = c.String(lockingBackendFlag)
	}
	if c.IsSet(lockingTableFlag) {
		lockingTable = c.String(lockingTableFlag)
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
		atlantisURL:      atlantisURL,
		requireApproval:  requireApproval,
		dataDir:          dataDir,
		lockingBackend:   lockingBackend,
		lockingTable:     lockingTable,
	}, nil
}

func initDB(db *bolt.DB) error {
	return db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(boltDBRunLocksBucket)); err != nil {
			return errors.Wrap(err, "failed to create bucket")
		}
		return nil
	})
}
