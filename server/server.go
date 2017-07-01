package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"time"

	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/elazarl/go-bindata-assetfs"
	"github.com/google/go-github/github"
	"github.com/gorilla/mux"
	"github.com/hootsuite/atlantis/locking"
	"github.com/hootsuite/atlantis/locking/boltdb"
	"github.com/hootsuite/atlantis/locking/dynamodb"
	"github.com/hootsuite/atlantis/logging"
	"github.com/hootsuite/atlantis/middleware"
	"github.com/hootsuite/atlantis/models"
	"github.com/hootsuite/atlantis/plan"
	"github.com/hootsuite/atlantis/plan/file"
	"github.com/hootsuite/atlantis/plan/s3"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"github.com/urfave/negroni"
	"github.com/hootsuite/atlantis/prerun"
)

const (
	deleteLockRoute        = "delete-lock"
	LockingFileBackend     = "file"
	LockingDynamoDBBackend = "dynamodb"
	PlanFileBackend        = "file"
	PlanS3Backend          = "s3"
)

// Server listens for GitHub events and runs the necessary Atlantis command
type Server struct {
	router             *mux.Router
	port               int
	commandHandler     *CommandHandler
	pullClosedExecutor *PullClosedExecutor
	logger             *logging.SimpleLogger
	eventParser        *EventParser
	lockingClient      *locking.Client
	atlantisURL        string
}

// the mapstructure tags correspond to flags in cmd/server.go
type ServerConfig struct {
	AWSRegion            string `mapstructure:"aws-region"`
	AssumeRole           string `mapstructure:"aws-assume-role-arn"`
	AtlantisURL          string `mapstructure:"atlantis-url"`
	DataDir              string `mapstructure:"data-dir"`
	GithubHostname       string `mapstructure:"gh-hostname"`
	GithubPassword       string `mapstructure:"gh-password"`
	GithubUser           string `mapstructure:"gh-user"`
	LockingBackend       string `mapstructure:"locking-backend"`
	LockingDynamoDBTable string `mapstructure:"locking-dynamodb-table"`
	LogLevel             string `mapstructure:"log-level"`
	Port                 int    `mapstructure:"port"`
	PlanS3Bucket         string `mapstructure:"plan-s3-bucket"`
	PlanS3Prefix         string `mapstructure:"plan-s3-prefix"`
	PlanBackend          string `mapstructure:"plan-backend"`
	RequireApproval      bool   `mapstructure:"require-approval"`
	SSHKey               string `mapstructure:"ssh-key"`
	ScratchDir           string `mapstructure:"scratch-dir"`
}

type CommandContext struct {
	Repo    models.Repo
	Pull    models.PullRequest
	User    models.User
	Command *Command
	Log     *logging.SimpleLogger
}

// todo: These structs have nothing to do with the server. Move to a different file/package #refactor
type ExecutionResult struct {
	SetupError   Templater
	SetupFailure Templater
	PathResults  []PathResult
	Command      CommandType
}

type PathResult struct {
	Path   string
	Status Status
	Result Templater
}

type Templater interface {
	Template() *CompiledTemplate
}

type GeneralError struct {
	Error error
}

func (g GeneralError) Template() *CompiledTemplate {
	return GeneralErrorTmpl
}

// todo: /end


func NewServer(config ServerConfig) (*Server, error) {
	tp := github.BasicAuthTransport{
		Username: strings.TrimSpace(config.GithubUser),
		Password: strings.TrimSpace(config.GithubPassword),
	}
	githubBaseClient := github.NewClient(tp.Client())
	githubClientCtx := context.Background()
	ghHostname := fmt.Sprintf("https://%s/api/v3/", config.GithubHostname)
	if config.GithubHostname == "api.github.com" {
		ghHostname = fmt.Sprintf("https://%s/", config.GithubHostname)
	}
	githubBaseClient.BaseURL, _ = url.Parse(ghHostname)
	githubClient := &GithubClient{client: githubBaseClient, ctx: githubClientCtx}
	githubStatus := &GithubStatus{client: githubClient}
	terraformClient, err := NewTerraformClient()
	if err != nil {
		return nil, errors.Wrap(err, "initializing terraform")
	}
	githubComments := &GithubCommentRenderer{}
	awsConfig := &AWSConfig{
		AWSRegion:  config.AWSRegion,
		AWSRoleArn: config.AssumeRole,
	}

	var awsSession *session.Session
	var lockingClient *locking.Client
	if config.LockingBackend == LockingDynamoDBBackend {
		awsSession, err = awsConfig.CreateAWSSession()
		if err != nil {
			return nil, errors.Wrap(err, "creating aws session for DynamoDB")
		}
		lockingClient = locking.NewClient(dynamodb.New(config.LockingDynamoDBTable, awsSession))
	} else {
		backend, err := boltdb.New(config.DataDir)
		if err != nil {
			return nil, err
		}
		lockingClient = locking.NewClient(backend)
	}
	var planBackend plan.Backend
	if config.PlanBackend == PlanS3Backend {
		if awsSession == nil {
			awsSession, err = awsConfig.CreateAWSSession()
			if err != nil {
				return nil, errors.Wrap(err, "creating aws session for S3")
			}
		}
		planBackend = s3.New(awsSession, config.PlanS3Bucket, config.PlanS3Prefix)
	} else {
		planBackend, err = file.New(config.DataDir)
		if err != nil {
			return nil, errors.Wrap(err, "creating file backend for plans")
		}
	}
	preRun := &prerun.PreRun{}
	configReader := &ConfigReader{}
	applyExecutor := &ApplyExecutor{
		github:                githubClient,
		githubStatus:          githubStatus,
		awsConfig:             awsConfig,
		scratchDir:            config.ScratchDir,
		sshKey:                config.SSHKey,
		terraform:             terraformClient,
		githubCommentRenderer: githubComments,
		lockingClient:         lockingClient,
		requireApproval:       config.RequireApproval,
		planBackend:           planBackend,
		preRun:                preRun,
		configReader: configReader,
	}
	planExecutor := &PlanExecutor{
		github:                githubClient,
		githubStatus:          githubStatus,
		awsConfig:             awsConfig,
		scratchDir:            config.ScratchDir,
		sshKey:                config.SSHKey,
		terraform:             terraformClient,
		githubCommentRenderer: githubComments,
		lockingClient:         lockingClient,
		planBackend:           planBackend,
		preRun:                preRun,
		configReader: configReader,
	}
	helpExecutor := &HelpExecutor{}
	pullClosedExecutor := &PullClosedExecutor{
		planBackend: planBackend,
		github:      githubClient,
		locking:     lockingClient,
	}
	logger := logging.NewSimpleLogger("server", log.New(os.Stderr, "", log.LstdFlags), false, logging.ToLogLevel(config.LogLevel))
	eventParser := &EventParser{}
	commandHandler := &CommandHandler{
		applyExecutor: applyExecutor,
		planExecutor:  planExecutor,
		helpExecutor:  helpExecutor,
		eventParser:   eventParser,
		githubClient:  githubClient,
		logger:        logger,
	}
	router := mux.NewRouter()
	return &Server{
		router:             router,
		port:               config.Port,
		commandHandler:     commandHandler,
		pullClosedExecutor: pullClosedExecutor,
		eventParser:        eventParser,
		logger:             logger,
		lockingClient:      lockingClient,
		atlantisURL:        config.AtlantisURL,
	}, nil
}

func (s *Server) Start() error {
	s.router.HandleFunc("/", s.index).Methods("GET").MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
		return r.URL.Path == "/" || r.URL.Path == "/index.html"
	})
	s.router.PathPrefix("/static/").Handler(http.FileServer(&assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: AssetInfo}))
	s.router.HandleFunc("/hooks", s.postHooks).Methods("POST")
	s.router.HandleFunc("/locks", s.deleteLock).Methods("DELETE").Queries("id", "{id:.*}")
	lockRoute := s.router.HandleFunc("/lock", s.lock).Methods("GET").Queries("id", "{id}")
	// function that planExecutor can use to construct detail view url
	// injecting this here because this is the earliest routes are created
	s.commandHandler.SetLockURL(func(lockID string) string {
		// ignoring error since guaranteed to succeed if "id" is specified
		u, _ := lockRoute.URL("id", url.QueryEscape(lockID))
		return s.atlantisURL + u.RequestURI()
	})
	n := negroni.New(&negroni.Recovery{
		Logger:     log.New(os.Stdout, "", log.LstdFlags),
		PrintStack: false,
		StackAll:   false,
		StackSize:  1024 * 8,
	}, middleware.NewNon200Logger(s.logger))
	n.UseHandler(s.router)
	s.logger.Info("Atlantis started - listening on port %v", s.port)
	return cli.NewExitError(http.ListenAndServe(fmt.Sprintf(":%d", s.port), n), 1)
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	locks, err := s.lockingClient.List()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, "Could not retrieve locks: %s", err)
		return
	}

	type lock struct {
		LockURL      string
		RepoFullName string
		PullNum      int
		Time         time.Time
	}
	var results []lock
	for id, v := range locks {
		results = append(results, lock{
			// todo: make LockURL use the router to get /lock endpoint
			LockURL:      fmt.Sprintf("/lock?id=%s", url.QueryEscape(id)),
			RepoFullName: v.Project.RepoFullName,
			PullNum:      v.Pull.Num,
			Time:         v.Time,
		})
	}
	indexTemplate.Execute(w, results)
}

func (s *Server) lock(w http.ResponseWriter, r *http.Request) {
	id, ok := mux.Vars(r)["id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "no lock id in request")
	}
	// get details for lock id
	idUnencoded, err := url.QueryUnescape(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "invalid lock id")
	}

	// for the given lock key get lock data
	lock, err := s.lockingClient.GetLock(idUnencoded)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
	}

	type lockData struct {
		UnlockURL       string
		LockKeyEncoded  string
		LockKey         string
		RepoOwner       string
		RepoName        string
		PullRequestLink string
		LockedBy        string
		Environment     string
		Time            time.Time
	}

	// extract the repo owner and repo name
	repo := strings.Split(lock.Project.RepoFullName, "/")

	l := lockData{
		LockKeyEncoded:  id,
		LockKey:         idUnencoded,
		RepoOwner:       repo[0],
		RepoName:        repo[1],
		PullRequestLink: lock.Pull.URL,
		LockedBy:        lock.Pull.Author,
		Environment:     lock.Env,
	}

	lockTemplate.Execute(w, l)
}

func (s *Server) deleteLock(w http.ResponseWriter, r *http.Request) {
	id, ok := mux.Vars(r)["id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "no lock id in request")
	}
	idUnencoded, err := url.PathUnescape(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "invalid lock id")
	}
	lock, err := s.lockingClient.Unlock(idUnencoded)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Failed to unlock: %s", err)
		return
	}
	if lock == nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "No lock at that key")
		return
	}
	fmt.Fprint(w, "Unlocked successfully")
}

// postHooks handles comment and pull request events from GitHub
func (s *Server) postHooks(w http.ResponseWriter, r *http.Request) {
	githubReqID := "X-Github-Delivery=" + r.Header.Get("X-Github-Delivery")

	var payload []byte

	// webhook requests can either be application/json or application/x-www-form-urlencoded
	if r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
		// GitHub stores the json payload as a form value
		payloadForm := r.PostFormValue("payload")
		if payloadForm == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "request did not contain expected 'payload' form value")
			return

		}
		payload = []byte(payloadForm)
	} else {
		// else read it as json
		defer r.Body.Close()
		var err error
		payload, err = ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "could not read body: %s", err)
			return
		}
	}

	event, _ := github.ParseWebHook(github.WebHookType(r), payload)
	switch event := event.(type) {
	case *github.IssueCommentEvent:
		s.handleCommentEvent(w, event, githubReqID)
	case *github.PullRequestEvent:
		s.handlePullRequestEvent(w, event, githubReqID)
	default:
		s.logger.Debug("Ignoring unsupported event %s", githubReqID)
		fmt.Fprintln(w, "Ignoring")
	}
}

// handlePullRequestEvent will delete any locks associated with the pull request
func (s *Server) handlePullRequestEvent(w http.ResponseWriter, pullEvent *github.PullRequestEvent, githubReqID string) {
	if pullEvent.GetAction() != "closed" {
		s.logger.Debug("Ignoring pull request event since action was not closed %s", githubReqID)
		fmt.Fprintln(w, "Ignoring")
		return
	}
	pull, err := s.eventParser.ExtractPullData(pullEvent.PullRequest)
	if err != nil {
		s.logger.Err("parsing pull data: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "error parsing request: %s", err)
		return
	}
	repo, err := s.eventParser.ExtractRepoData(pullEvent.Repo)
	if err != nil {
		s.logger.Err("parsing repo data: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "error parsing request: %s", err)
		return
	}

	s.logger.Info("cleaning up locks and plans for repo %s and pull %d", repo.FullName, pull.Num)
	if err := s.pullClosedExecutor.CleanUpPull(repo, pull); err != nil {
		s.logger.Err("cleaning pull request: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error cleaning pull request: %s", err)
		return
	}
	fmt.Fprint(w, "Pull request cleaned successfully")
}

func (s *Server) handleCommentEvent(w http.ResponseWriter, event *github.IssueCommentEvent, githubReqID string) {
	if event.GetAction() != "created" {
		s.logger.Debug("Ignoring comment event since action was not created %s", githubReqID)
		fmt.Fprintln(w, "Ignoring")
		return
	}

	// determine if the comment matches a plan or apply command
	ctx := &CommandContext{}
	command, err := s.eventParser.DetermineCommand(event)
	if err != nil {
		s.logger.Debug("Ignoring request: %v %s", err, githubReqID)
		fmt.Fprintln(w, "Ignoring")
		return
	}
	ctx.Command = command

	if err = s.eventParser.ExtractCommentData(event, ctx); err != nil {
		s.logger.Err("Failed parsing event: %v %s", err, githubReqID)
		fmt.Fprintln(w, "Ignoring")
		return
	}
	// respond with success and then actually execute the command asynchronously
	fmt.Fprintln(w, "Processing...")
	go s.commandHandler.ExecuteCommand(ctx)
}
