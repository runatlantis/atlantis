package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

<<<<<<< d425de95fa1f3ee312a8b569290932163bc93bf1
=======
	"io/ioutil"

>>>>>>> Adding ci for atlantis
	"github.com/elazarl/go-bindata-assetfs"
	"github.com/google/go-github/github"
	"github.com/gorilla/mux"
	"github.com/hootsuite/atlantis/locking"
	"github.com/hootsuite/atlantis/locking/boltdb"
	"github.com/hootsuite/atlantis/locking/dynamodb"
	"github.com/hootsuite/atlantis/logging"
	"github.com/hootsuite/atlantis/middleware"
<<<<<<< d425de95fa1f3ee312a8b569290932163bc93bf1
	"github.com/hootsuite/atlantis/models"
=======
>>>>>>> Adding ci for atlantis
	"github.com/hootsuite/atlantis/recovery"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"github.com/urfave/negroni"
<<<<<<< d425de95fa1f3ee312a8b569290932163bc93bf1
	"io/ioutil"
	"path/filepath"
	"time"
=======
>>>>>>> Adding ci for atlantis
)

const (
	deleteLockRoute        = "delete-lock"
	LockingFileBackend     = "file"
	LockingDynamoDBBackend = "dynamodb"
)

// Server listens for Github webhooks and runs the necessary Atlantis command
type Server struct {
	router           *mux.Router
	port             int
	scratchDir       string
	awsRegion        string
	s3Bucket         string
	githubBaseClient *github.Client
	githubClient     *GithubClient
	applyExecutor    *ApplyExecutor
	planExecutor     *PlanExecutor
	helpExecutor     *HelpExecutor
	logger           *logging.SimpleLogger
	githubComments   *GithubCommentRenderer
	requestParser    *RequestParser
	lockingClient    *locking.Client
	atlantisURL      string
}

// the mapstructure tags correspond to flags in cmd/server.go
type ServerConfig struct {
	GitHubHostname       string `mapstructure:"gh-hostname"`
	GitHubUser           string `mapstructure:"gh-user"`
	GitHubPassword       string `mapstructure:"gh-password"`
	SSHKey               string `mapstructure:"ssh-key"`
	AssumeRole           string `mapstructure:"aws-assume-role-arn"`
	Port                 int    `mapstructure:"port"`
	ScratchDir           string `mapstructure:"scratch-dir"`
	AWSRegion            string `mapstructure:"aws-region"`
	S3Bucket             string `mapstructure:"s3-bucket"`
	LogLevel             string `mapstructure:"log-level"`
	AtlantisURL          string `mapstructure:"atlantis-url"`
	RequireApproval      bool   `mapstructure:"require-approval"`
	DataDir              string `mapstructure:"data-dir"`
	LockingBackend       string `mapstructure:"locking-backend"`
	LockingDynamoDBTable string `mapstructure:"locking-dynamodb-table"`
}

<<<<<<< d425de95fa1f3ee312a8b569290932163bc93bf1
// todo: rename to Command
type CommandContext struct {
	Repo    models.Repo
	Pull    models.PullRequest
	User    models.User
	Command *Command
	Log     *logging.SimpleLogger
}

type ExecutionResult struct {
	SetupError   Templater
	SetupFailure Templater
	PathResults  []PathResult
	Command      CommandType
}

type PathResult struct {
	Path   string
	Status string // todo: this should be an enum for success/error/failure
	Result Templater
}

type ExecutionPath struct {
	// Absolute is the full path on the OS where we will execute.
	// Will never end with a '/'.
	Absolute string
	// Relative is the path relative to the repo root.
	// Will never end with a '/'.
	Relative string
}

func NewExecutionPath(absolutePath string, relativePath string) ExecutionPath {
	return ExecutionPath{filepath.Clean(absolutePath), filepath.Clean(relativePath)}
}

type Templater interface {
	Template() *CompiledTemplate
}

type GeneralError struct {
	Error error
}

func (g GeneralError) Template() *CompiledTemplate {
	return GeneralErrorTmpl
=======
type ExecutionContext struct {
	repoFullName      string
	pullNum           int
	requesterUsername string
	requesterEmail    string
	comment           string
	repoSSHUrl        string
	head              string
	// commit base sha
	base        string
	pullLink    string
	branch      string
	htmlUrl     string
	pullCreator string
	command     *Command
	log         *logging.SimpleLogger
>>>>>>> Adding ci for atlantis
}

func NewServer(config ServerConfig) (*Server, error) {
	tp := github.BasicAuthTransport{
		Username: strings.TrimSpace(config.GitHubUser),
		Password: strings.TrimSpace(config.GitHubPassword),
	}
	githubBaseClient := github.NewClient(tp.Client())
	githubClientCtx := context.Background()
	ghHostname := fmt.Sprintf("https://%s/api/v3/", config.GitHubHostname)
	if config.GitHubHostname == "api.github.com" {
		ghHostname = fmt.Sprintf("https://%s/", config.GitHubHostname)
	}
	githubBaseClient.BaseURL, _ = url.Parse(ghHostname)
	githubClient := &GithubClient{client: githubBaseClient, ctx: githubClientCtx}
	terraformClient := &TerraformClient{
		tfExecutableName: "terraform",
	}
	githubComments := &GithubCommentRenderer{}
	awsConfig := &AWSConfig{
		AWSRegion:  config.AWSRegion,
		AWSRoleArn: config.AssumeRole,
	}
	var lockingClient *locking.Client
	if config.LockingBackend == LockingDynamoDBBackend {
		session, err := awsConfig.CreateAWSSession()
		if err != nil {
			return nil, errors.Wrap(err, "creating aws session for DynamoDB")
		}
		lockingClient = locking.NewClient(dynamodb.New(config.LockingDynamoDBTable, session))
	} else {
		backend, err := boltdb.New(config.DataDir)
		if err != nil {
			return nil, err
		}
		lockingClient = locking.NewClient(backend)
	}
	applyExecutor := &ApplyExecutor{
		github:                githubClient,
		awsConfig:             awsConfig,
		scratchDir:            config.ScratchDir,
		s3Bucket:              config.S3Bucket,
		sshKey:                config.SSHKey,
		terraform:             terraformClient,
		githubCommentRenderer: githubComments,
		lockingClient:         lockingClient,
		requireApproval:       config.RequireApproval,
	}
	planExecutor := &PlanExecutor{
		github:                githubClient,
		awsConfig:             awsConfig,
		scratchDir:            config.ScratchDir,
		s3Bucket:              config.S3Bucket,
		sshKey:                config.SSHKey,
		terraform:             terraformClient,
		githubCommentRenderer: githubComments,
		lockingClient:         lockingClient,
	}
	helpExecutor := &HelpExecutor{}
	logger := logging.NewSimpleLogger("server", log.New(os.Stderr, "", log.LstdFlags), false, logging.ToLogLevel(config.LogLevel))
	router := mux.NewRouter()
	return &Server{
		router:           router,
		port:             config.Port,
		scratchDir:       config.ScratchDir,
		awsRegion:        config.AWSRegion,
		s3Bucket:         config.S3Bucket,
		applyExecutor:    applyExecutor,
		planExecutor:     planExecutor,
		helpExecutor:     helpExecutor,
		githubBaseClient: githubBaseClient,
		githubClient:     githubClient,
		logger:           logger,
		githubComments:   githubComments,
		requestParser:    &RequestParser{},
		lockingClient:    lockingClient,
		atlantisURL:      config.AtlantisURL,
	}, nil
}

func (s *Server) Start() error {
	s.router.HandleFunc("/", s.index).Methods("GET").MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
		return r.URL.Path == "/" || r.URL.Path == "/index.html"
	})
	s.router.PathPrefix("/static/").Handler(http.FileServer(&assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: AssetInfo}))
	s.router.HandleFunc("/hooks", s.postHooks).Methods("POST")
	s.router.HandleFunc("/locks", s.deleteLock).Methods("DELETE").Queries("id", "{id:.*}")
	// todo: remove this route when there is a detail view
	// right now we need this route because from the pull request comment in GitHub only a GET request can be made
	// in the future, the pull discard link will link to the detail view which will have a Delete button which will
	// make an real DELETE call but we don't have a detail view right now
	deleteLockRoute := s.router.HandleFunc("/locks", s.deleteLock).Queries("id", "{id}", "method", "DELETE").Methods("GET").Name(deleteLockRoute)

	// function that planExecutor can use to construct delete lock urls
	// injecting this here because this is the earliest routes are created
	s.planExecutor.DeleteLockURL = func(lockID string) string {
		// ignoring error since guaranteed to succeed if "id" is specified
		u, _ := deleteLockRoute.URL("id", url.QueryEscape(lockID))
		return s.atlantisURL + u.RequestURI()
	}
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
		UnlockURL    string
		RepoFullName string
		PullNum      int
		Time         time.Time
	}
	var results []lock
	for id, v := range locks {
		u, _ := s.router.Get(deleteLockRoute).URL("id", url.QueryEscape(id))
		results = append(results, lock{
			UnlockURL:    u.String(),
			RepoFullName: v.Project.RepoFullName,
			PullNum:      v.PullNum,
			Time:         v.Time,
		})
	}
	indexTemplate.Execute(w, results)
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
	if err := s.lockingClient.Unlock(idUnencoded); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Failed to unlock: %s", err)
		return
	}
	fmt.Fprint(w, "Unlocked successfully")
}

// postHooks handles comment and pull request events from GitHub
func (s *Server) postHooks(w http.ResponseWriter, r *http.Request) {
	githubReqID := "X-Github-Delivery=" + r.Header.Get("X-Github-Delivery")

	defer r.Body.Close()
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "could not read body: %s\n", err)
		return
	}

	// Try to unmarshal the request into the supported event types
	var commentEvent github.IssueCommentEvent
	var pullEvent github.PullRequestEvent
	if json.Unmarshal(bytes, &commentEvent) == nil && s.isCommentCreatedEvent(commentEvent) {
		s.logger.Debug("Handling comment event %s", githubReqID)
		s.handleCommentCreatedEvent(w, commentEvent, githubReqID)
	} else if json.Unmarshal(bytes, &pullEvent) == nil && s.isPullClosedEvent(pullEvent) {
		s.logger.Debug("Handling pull request event %s", githubReqID)
		s.handlePullClosedEvent(w, pullEvent, githubReqID)
	} else {
		s.logger.Debug("Ignoring unsupported event %s", githubReqID)
		fmt.Fprintln(w, "Ignoring")
	}
}

// handlePullClosedEvent will delete any locks associated with the pull request
func (s *Server) handlePullClosedEvent(w http.ResponseWriter, pullEvent github.PullRequestEvent, githubReqID string) {
	repo := *pullEvent.Repo.FullName
	pullNum := *pullEvent.PullRequest.Number
	s.logger.Debug("Unlocking locks for repo %s and pull %d %s", repo, pullNum, githubReqID)
	err := s.lockingClient.UnlockByPull(repo, pullNum)
	if err != nil {
		s.logger.Err("unlocking locks for repo %s pull %d: %v", repo, pullNum, err)
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, "Error unlocking locks: %v\n", err)
		return
	}
<<<<<<< d425de95fa1f3ee312a8b569290932163bc93bf1
=======

>>>>>>> Adding ci for atlantis
	fmt.Fprintln(w, "Locks unlocked")
}

func (s *Server) handleCommentCreatedEvent(w http.ResponseWriter, comment github.IssueCommentEvent, githubReqID string) {
	// determine if the comment matches a plan or apply command
	ctx := &CommandContext{}
	command, err := s.requestParser.determineCommand(&comment)
	if err != nil {
		s.logger.Debug("Ignoring request: %v %s", err, githubReqID)
		fmt.Fprintln(w, "Ignoring")
		return
	}
	ctx.Command = command

	if err = s.requestParser.extractCommentData(&comment, ctx); err != nil {
		s.logger.Err("Failed parsing event: %v %s", err, githubReqID)
		fmt.Fprintln(w, "Ignoring")
		return
	}
	// respond with success and then actually execute the command asynchronously
	fmt.Fprintln(w, "Processing...")
	go s.executeCommand(ctx)
}

func (s *Server) executeCommand(ctx *CommandContext) {
	src := fmt.Sprintf("%s/pull/%d", ctx.Repo.FullName, ctx.Pull.Num)
	// it's safe to reuse the underlying logger s.logger.Log
	ctx.Log = logging.NewSimpleLogger(src, s.logger.Log, true, s.logger.Level)
	defer s.recover(ctx)

	// we've got data from the comment, now we need to get data from the actual PR
	pull, _, err := s.githubClient.GetPullRequest(ctx.Repo, ctx.Pull.Num)
	if err != nil {
		ctx.Log.Err("pull request data api call failed: %v", err)
		return
	}
	if err := s.requestParser.extractPullData(pull, ctx); err != nil {
		ctx.Log.Err("failed to extract required fields from comment data: %v", err)
		return
	}

	switch ctx.Command.commandType {
	case Plan:
		s.planExecutor.execute(ctx, s.githubClient)
	case Apply:
		s.applyExecutor.execute(ctx, s.githubClient)
	case Help:
		s.helpExecutor.execute(ctx, s.githubClient)
	default:
		ctx.Log.Err("failed to determine desired command, neither plan nor apply")
	}
}

func (s *Server) isCommentCreatedEvent(event github.IssueCommentEvent) bool {
	return event.Action != nil && *event.Action == "created" && event.Comment != nil
}

func (s *Server) isPullClosedEvent(event github.PullRequestEvent) bool {
	return event.Action != nil && *event.Action == "closed" && event.PullRequest != nil
}

// recover logs and creates a comment on the pull request for panics
func (s *Server) recover(ctx *CommandContext) {
	if err := recover(); err != nil {
		stack := recovery.Stack(3)
		s.githubClient.CreateComment(ctx, fmt.Sprintf("**Error: goroutine panic. This is a bug.**\n```\n%s\n%s```", err, stack))
		ctx.Log.Err("PANIC: %s\n%s", err, stack)
	}
}
