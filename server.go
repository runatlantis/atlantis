package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"encoding/json"

	"github.com/google/go-github/github"
	"github.com/urfave/cli"
	"github.com/hootsuite/atlantis/recovery"
	"github.com/hootsuite/atlantis/locking"
	"github.com/boltdb/bolt"
	"github.com/urfave/negroni"
	"github.com/hootsuite/atlantis/middleware"
	"github.com/hootsuite/atlantis/logging"
	"github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

const (
	lockPath = "/locks"
)

// WebhookServer listens for Github webhooks and runs the necessary Atlantis command
type Server struct {
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
	lockManager      locking.LockManager
}

type ServerConfig struct {
	githubUsername   string
	githubPassword   string
	githubHostname   string
	sshKey           string
	awsAssumeRole    string
	port             int
	scratchDir       string
	awsRegion        string
	s3Bucket         string
	logLevel         string
	atlantisURL      string
	requireApproval  bool
	dataDir          string
	lockingBackend   string
	lockingTable        string
}

type ExecutionContext struct {
	repoOwner         string
	repoName          string
	pullNum           int
	requesterUsername string
	requesterEmail    string
	comment           string
	repoSSHUrl        string
	head              string
	// commit base sha
	base              string
	pullLink          string
	branch            string
	htmlUrl           string
	pullCreator       string
	command           *Command
	log         *logging.SimpleLogger
}

func NewServer(config *ServerConfig, db *bolt.DB) (*Server, error) {
	tp := github.BasicAuthTransport{
		Username: strings.TrimSpace(config.githubUsername),
		Password: strings.TrimSpace(config.githubPassword),
	}
	githubBaseClient := github.NewClient(tp.Client())
	githubClientCtx := context.Background()
	githubBaseClient.BaseURL, _ = url.Parse(fmt.Sprintf("https://%s/api/v3/", config.githubHostname))
	githubClient := &GithubClient{client: githubBaseClient, ctx: githubClientCtx}
	terraformClient := &TerraformClient{
		tfExecutableName: "terraform",
	}
	githubComments := &GithubCommentRenderer{}
	awsConfig := &AWSConfig{
		AWSRegion:  config.awsRegion,
		AWSRoleArn: config.awsAssumeRole,
	}
	var lockManager locking.LockManager
	if config.lockingBackend == dynamoDBLockingBackend {
		session, err := awsConfig.CreateAWSSession()
		if err != nil {
			return nil, errors.Wrap(err, "creating aws session for DynamoDB")
		}
		lockManager = locking.NewDynamoDBLockManager(config.lockingTable, session)
	} else {
		lockManager = locking.NewBoltDBLockManager(db, boltDBRunLocksBucket)
	}
	baseExecutor := BaseExecutor{
		github:                githubClient,
		awsConfig:             awsConfig,
		scratchDir:            config.scratchDir,
		s3Bucket:              config.s3Bucket,
		sshKey:                config.sshKey,
		ghComments:            githubComments,
		terraform:             terraformClient,
		githubCommentRenderer: githubComments,
		lockManager:           lockManager,
	}
	applyExecutor := &ApplyExecutor{BaseExecutor: baseExecutor, requireApproval: config.requireApproval, atlantisGithubUser: config.githubUsername}
	planExecutor := &PlanExecutor{BaseExecutor: baseExecutor, atlantisURL: config.atlantisURL}
	helpExecutor := &HelpExecutor{BaseExecutor: baseExecutor}
	logger := logging.NewSimpleLogger("server", log.New(os.Stderr, "", log.LstdFlags), false, logging.ToLogLevel(config.logLevel))
	return &Server{
		port:             config.port,
		scratchDir:       config.scratchDir,
		awsRegion:        config.awsRegion,
		s3Bucket:         config.s3Bucket,
		applyExecutor:    applyExecutor,
		planExecutor:     planExecutor,
		helpExecutor:     helpExecutor,
		githubBaseClient: githubBaseClient,
		githubClient:     githubClient,
		logger:           logger,
		githubComments:   githubComments,
		requestParser:    &RequestParser{},
		lockManager:      lockManager,
	}, nil
}

func (s *Server) Start() error {
	router := mux.NewRouter()
	router.HandleFunc("/", s.index).Methods("GET").MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
		return r.URL.Path == "/" || r.URL.Path == "/index.html"
	})
	router.PathPrefix("/static/").Handler(http.FileServer(&assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: AssetInfo}))
	router.HandleFunc("/hooks", s.postHooks).Methods("POST")
	router.HandleFunc("/locks/{id}", s.deleteLock).Methods("DELETE")
	// todo: remove this route when there is a detail view
	// right now we need this route because from the pull request comment in GitHub only a GET request can be made
	// in the future, the pull discard link will link to the detail view which will have a Delete button which will
	// make an real DELETE call but we don't have a detail view right now
	router.HandleFunc("/locks/{id}", s.deleteLock).Queries("method", "DELETE").Methods("GET")
	n := negroni.New(&negroni.Recovery{
		Logger:     log.New(os.Stdout, "", log.LstdFlags),
		PrintStack: false,
		StackAll:   false,
		StackSize:  1024 * 8,
	}, middleware.NewNon200Logger(s.logger))
	n.UseHandler(router)
	s.logger.Info("Atlantis started - listening on port %v", s.port)
	return cli.NewExitError(http.ListenAndServe(fmt.Sprintf(":%d", s.port), n), 1)
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	locks, err := s.lockManager.ListLocks()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, "Could not retrieve locks: %s", err)
		return
	}

	type runLock struct {
		locking.Run
		ID string
	}
	var results []runLock
	for id, v := range locks {
		results = append(results, runLock{
			v,
			id,
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
	if err := s.lockManager.Unlock(id); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Failed to unlock: %s", err)
		return
	}
	fmt.Fprint(w, "Unlocked successfully")
}

func (s *Server) postHooks(w http.ResponseWriter, r *http.Request) {
	githubReqID := "X-Github-Delivery=" + r.Header.Get("X-Github-Delivery")

	// try to parse webhook data as a comment event
	decoder := json.NewDecoder(r.Body)
	var comment github.IssueCommentEvent
	if err := decoder.Decode(&comment); err != nil {
		s.logger.Debug("Ignoring non-comment-event request %s", githubReqID)
		fmt.Fprintln(w, "Ignoring")
		return
	}

	if comment.Action == nil || *comment.Action != "created" {
		s.logger.Debug("Ignoring request because the action was not 'created' %s", githubReqID)
		fmt.Fprintln(w, "Ignoring")
		return
	}

	// determine if the comment matches a plan or apply command
	ctx := &ExecutionContext{}
	command, err := s.requestParser.determineCommand(&comment)
	if err != nil {
		s.logger.Debug("Ignoring request: %v %s", err, githubReqID)
		fmt.Fprintln(w, "Ignoring")
		return
	}
	ctx.command = command

	if err = s.requestParser.extractCommentData(&comment, ctx); err != nil {
		s.logger.Err("Failed parsing event: %v %s", err, githubReqID)
		fmt.Fprintln(w, "Ignoring")
		return
	}
	// respond with success and then actually execute the command asynchronously
	fmt.Fprintln(w, "Processing...")
	go s.executeCommand(ctx)
}

func (s *Server) executeCommand(ctx *ExecutionContext) {
	src := fmt.Sprintf("%s/%s/pull/%d", ctx.repoOwner, ctx.repoName, ctx.pullNum)
	// it's safe to reuse the underlying logger s.logger.Log
	ctx.log = logging.NewSimpleLogger(src, s.logger.Log, true, s.logger.Level)
	defer s.recover(ctx)

	// we've got data from the comment, now we need to get data from the actual PR
	pull, _, err := s.githubClient.GetPullRequest(ctx.repoOwner, ctx.repoName, ctx.pullNum)
	if err != nil {
		ctx.log.Err("pull request data api call failed: %v", err)
		return
	}
	if err := s.requestParser.extractPullData(pull, ctx); err != nil {
		ctx.log.Err("failed to extract required fields from comment data: %v", err)
		return
	}

	switch ctx.command.commandType {
	case Plan:
		s.planExecutor.Exec(s.planExecutor.execute, ctx, s.githubClient)
	case Apply:
		s.applyExecutor.Exec(s.applyExecutor.execute, ctx, s.githubClient)
	case Help:
		s.helpExecutor.execute(ctx, s.githubClient)
	default:
		ctx.log.Err("failed to determine desired command, neither plan nor apply")
	}
}

// recover logs and creates a comment on the pull request for panics
func (s *Server) recover(ctx *ExecutionContext) {
	if err := recover(); err != nil {
		ghCtx := s.planExecutor.githubContext(ctx) // this won't have every field but it has the ones needed by CreateComment
		stack := recovery.Stack(3)
		s.githubClient.CreateComment(ghCtx, fmt.Sprintf("**Error: goroutine panic. This is a bug.**\n```\n%s\n%s```", err, stack))
		ctx.log.Err("PANIC: %s\n%s", err, stack)
	}
}
