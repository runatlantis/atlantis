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
	"github.hootops.com/production-delivery/atlantis/recovery"
	"github.com/urfave/negroni"
	"github.hootops.com/production-delivery/atlantis/middleware"
	"github.hootops.com/production-delivery/atlantis/logging"
	"github.com/elazarl/go-bindata-assetfs"
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
	requireApproval  bool
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
	base        string
	pullLink    string
	branch      string
	htmlUrl     string
	pullCreator string
	command     *Command
	log         *logging.SimpleLogger
}

func NewServer(config *ServerConfig) *Server {
	tp := github.BasicAuthTransport{
		Username: strings.TrimSpace(config.githubUsername),
		Password: strings.TrimSpace(config.githubPassword),
	}
	githubBaseClient := github.NewClient(tp.Client())
	githubClientCtx := context.Background()
	githubBaseClient.BaseURL, _ = url.Parse(fmt.Sprintf("https://%s/api/v3/", config.githubHostname))
	githubClient := &GithubClient{client: githubBaseClient, ctx: githubClientCtx}
	stashClient := &StashClient{}
	terraformClient := &TerraformClient{
		tfExecutableName: "terraform",
	}
	githubComments := &GithubCommentRenderer{}
	awsConfig := &AWSConfig{
		AWSRegion:  config.awsRegion,
		AWSRoleArn: config.awsAssumeRole,
	}
	baseExecutor := BaseExecutor{
		github:                githubClient,
		awsConfig:             awsConfig,
		scratchDir:            config.scratchDir,
		s3Bucket:              config.s3Bucket,
		sshKey:                config.sshKey,
		stash:                 &StashPRClient{client: stashClient},
		ghComments:            githubComments,
		terraform:             terraformClient,
		githubCommentRenderer: githubComments,
	}
	applyExecutor := &ApplyExecutor{BaseExecutor: baseExecutor, requireApproval: config.requireApproval, atlantisGithubUser: config.githubUsername}
	planExecutor := &PlanExecutor{BaseExecutor: baseExecutor}
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
	}
}

func (s *Server) Start() error {
	router := http.NewServeMux()
	router.HandleFunc("/", s.index)
	router.Handle("/static/", http.FileServer(&assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: AssetInfo}))
	router.HandleFunc("/hooks", s.postHooks)
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
	if r.URL.Path != "/" && r.URL.Path != "/index.html" {
		http.NotFound(w, r)
		return
	}
	// this is just a placeholder to test the ui
	type result struct {
		RepoOwner     string
		RepoName      string
		PullRequestId int
		Timestamp     string
	}
	res := result{}
	res.RepoOwner = "anubhavmishra"
	res.RepoName = "atlantis-test"
	res.PullRequestId = 10
	res.Timestamp = "2017-05-08T11:18:43.91646206-07:00"

	var results []result
	results = append(results, res)
	indexTemplate.Execute(w, results)
}

func (s *Server) postHooks(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.NotFound(w, r)
		return
	}
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
