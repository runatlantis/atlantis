package server

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"bytes"
	"io/ioutil"

	"github.com/elazarl/go-bindata-assetfs"
	gh "github.com/google/go-github/github"
	"github.com/gorilla/mux"
	"github.com/hootsuite/atlantis/github"
	"github.com/hootsuite/atlantis/locking"
	"github.com/hootsuite/atlantis/locking/boltdb"
	"github.com/hootsuite/atlantis/logging"
	"github.com/hootsuite/atlantis/models"
	"github.com/hootsuite/atlantis/run"
	"github.com/hootsuite/atlantis/terraform"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"github.com/urfave/negroni"
)

const (
	lockRoute = "lock-detail"
	// atlantisUserTFVar is the name of the variable we execute terraform
	// with containing the github username of who is running the command
	atlantisUserTFVar = "atlantis_user"
)

// Server listens for GitHub events and runs the necessary Atlantis command
type Server struct {
	router              *mux.Router
	port                int
	commandHandler      *CommandHandler
	pullClosedExecutor  *PullClosedExecutor
	logger              *logging.SimpleLogger
	eventParser         *EventParser
	lockingClient       *locking.Client
	atlantisURL         string
	githubWebHookSecret []byte
}

// the mapstructure tags correspond to flags in cmd/server.go
type ServerConfig struct {
	AtlantisURL         string `mapstructure:"atlantis-url"`
	DataDir             string `mapstructure:"data-dir"`
	GithubHostname      string `mapstructure:"gh-hostname"`
	GithubToken         string `mapstructure:"gh-token"`
	GithubUser          string `mapstructure:"gh-user"`
	GithubWebHookSecret string `mapstructure:"gh-webhook-secret"`
	LogLevel            string `mapstructure:"log-level"`
	Port                int    `mapstructure:"port"`
	RequireApproval     bool   `mapstructure:"require-approval"`
}

type CommandContext struct {
	BaseRepo models.Repo
	HeadRepo models.Repo
	Pull     models.PullRequest
	User     models.User
	Command  *Command
	Log      *logging.SimpleLogger
}

func NewServer(config ServerConfig) (*Server, error) {
	// if ~ was used in data-dir convert that to actual home directory otherwise we'll
	// create a directory call "~" instead of actually using home
	if strings.HasPrefix(config.DataDir, "~/") {
		expanded, err := homedir.Expand(config.DataDir)
		if err != nil {
			return nil, errors.Wrap(err, "determining home directory")
		}
		config.DataDir = expanded
	}

	githubClient, err := github.NewClient(config.GithubHostname, config.GithubUser, config.GithubToken)
	if err != nil {
		return nil, err
	}
	githubStatus := &GithubStatus{client: githubClient}
	terraformClient, err := terraform.NewClient()
	if err != nil {
		return nil, errors.Wrap(err, "initializing terraform")
	}
	githubComments := &GithubCommentRenderer{}

	boltdb, err := boltdb.New(config.DataDir)
	if err != nil {
		return nil, err
	}
	lockingClient := locking.NewClient(boltdb)
	run := &run.Run{}
	configReader := &ConfigReader{}
	concurrentRunLocker := NewConcurrentRunLocker()
	workspace := &Workspace{
		dataDir: config.DataDir,
	}
	applyExecutor := &ApplyExecutor{
		github:                githubClient,
		githubStatus:          githubStatus,
		terraform:             terraformClient,
		githubCommentRenderer: githubComments,
		lockingClient:         lockingClient,
		requireApproval:       config.RequireApproval,
		run:                   run,
		configReader:          configReader,
		concurrentRunLocker:   concurrentRunLocker,
		workspace:             workspace,
	}
	planExecutor := &PlanExecutor{
		github:                githubClient,
		githubStatus:          githubStatus,
		terraform:             terraformClient,
		githubCommentRenderer: githubComments,
		lockingClient:         lockingClient,
		run:                   run,
		configReader:          configReader,
		concurrentRunLocker:   concurrentRunLocker,
		workspace:             workspace,
	}
	helpExecutor := &HelpExecutor{
		github: githubClient,
	}
	pullClosedExecutor := &PullClosedExecutor{
		github:    githubClient,
		locking:   lockingClient,
		workspace: workspace,
	}
	logger := logging.NewSimpleLogger("server", log.New(os.Stderr, "", log.LstdFlags), false, logging.ToLogLevel(config.LogLevel))
	eventParser := &EventParser{
		GithubUser:  config.GithubUser,
		GithubToken: config.GithubToken,
	}
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
		router:              router,
		port:                config.Port,
		commandHandler:      commandHandler,
		pullClosedExecutor:  pullClosedExecutor,
		eventParser:         eventParser,
		logger:              logger,
		lockingClient:       lockingClient,
		atlantisURL:         config.AtlantisURL,
		githubWebHookSecret: []byte(config.GithubWebHookSecret),
	}, nil
}

func (s *Server) Start() error {
	s.router.HandleFunc("/", s.index).Methods("GET").MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
		return r.URL.Path == "/" || r.URL.Path == "/index.html"
	})
	s.router.PathPrefix("/static/").Handler(http.FileServer(&assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: AssetInfo}))
	s.router.HandleFunc("/events", s.postEvents).Methods("POST")
	s.router.HandleFunc("/locks", s.deleteLock).Methods("DELETE").Queries("id", "{id:.*}")
	lockRoute := s.router.HandleFunc("/lock", s.getLock).Methods("GET").Queries("id", "{id}").Name(lockRoute)
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
	}, NewRequestLogger(s.logger))
	n.UseHandler(s.router)
	s.logger.Warn("Atlantis started - listening on port %v", s.port)
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
		url, _ := s.router.Get(lockRoute).URL("id", url.QueryEscape(id))
		results = append(results, lock{
			LockURL:      url.String(),
			RepoFullName: v.Project.RepoFullName,
			PullNum:      v.Pull.Num,
			Time:         v.Time,
		})
	}
	indexTemplate.Execute(w, results)
}

func (s *Server) getLock(w http.ResponseWriter, r *http.Request) {
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
	if lock == nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "no lock found at that id")
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
	if !ok || id == "" {
		s.respond(w, logging.Warn, http.StatusBadRequest, "No id in request")
		return
	}
	idUnencoded, err := url.PathUnescape(id)
	if err != nil {
		s.respond(w, logging.Warn, http.StatusBadRequest, "Invalid lock id: %s", err)
		return
	}
	lock, err := s.lockingClient.Unlock(idUnencoded)
	if err != nil {
		s.respond(w, logging.Error, http.StatusInternalServerError, "Failed to delete lock %s: %s", idUnencoded, err)
		return
	}
	if lock == nil {
		s.respond(w, logging.Warn, http.StatusBadRequest, "No lock found at id %s", idUnencoded)
		return
	}
	s.respond(w, logging.Info, http.StatusOK, "Deleted lock id %s", idUnencoded)
}

// postEvents handles comment and pull request events from GitHub
func (s *Server) postEvents(w http.ResponseWriter, r *http.Request) {
	githubReqID := "X-Github-Delivery=" + r.Header.Get("X-Github-Delivery")
	var payload []byte

	// webhook requests can either be application/json or application/x-www-form-urlencoded.
	// We accept both to make it easier on users that may choose x-www-form-urlencoded by mistake
	// todo: use go-github's ValidatePayload method if https://github.com/google/go-github/pull/693 is merged
	if r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
		// GitHub stores the json payload as a form value
		payloadForm := r.FormValue("payload")
		if payloadForm == "" {
			s.respond(w, logging.Warn, http.StatusBadRequest, "request did not contain expected 'payload' form value")
			return
		}
		if len(s.githubWebHookSecret) != 0 {
			// github calculates the signature based on the query escaped
			// post body. In order to use go-github's ValidatePayload method
			// that only accepts an http request we need to override r.Body
			// with a value that was the original raw body before it was
			// parsed.
			rawPayload := fmt.Sprintf("payload=%s", url.QueryEscape(payloadForm))
			r.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(rawPayload)))
			_, err := gh.ValidatePayload(r, s.githubWebHookSecret)
			if err != nil {
				s.respond(w, logging.Warn, http.StatusBadRequest, "webhook failed secret key validation")
				return
			}
		}
		payload = []byte(payloadForm)
	} else {
		// else read it as json
		if len(s.githubWebHookSecret) != 0 {
			var err error
			payload, err = gh.ValidatePayload(r, s.githubWebHookSecret)
			if err != nil {
				s.respond(w, logging.Warn, http.StatusBadRequest, "webhook failed secret key validation")
				return
			}
		} else {
			// if we're not validating against the webhook secret then
			// we can't use the ValidatePayload method and need to read
			// the request body ourselves.
			defer r.Body.Close()
			var err error
			payload, err = ioutil.ReadAll(r.Body)
			if err != nil {
				s.respond(w, logging.Warn, http.StatusBadRequest, "could not read body: %s", err)
				return
			}
		}
	}

	event, _ := gh.ParseWebHook(gh.WebHookType(r), payload)
	switch event := event.(type) {
	case *gh.IssueCommentEvent:
		s.handleCommentEvent(w, event, githubReqID)
	case *gh.PullRequestEvent:
		s.handlePullRequestEvent(w, event, githubReqID)
	default:
		s.respond(w, logging.Debug, http.StatusOK, "Ignoring unsupported event %s", githubReqID)
	}
}

// handlePullRequestEvent will delete any locks associated with the pull request
func (s *Server) handlePullRequestEvent(w http.ResponseWriter, pullEvent *gh.PullRequestEvent, githubReqID string) {
	if pullEvent.GetAction() != "closed" {
		s.respond(w, logging.Debug, http.StatusOK, "Ignoring pull request event since action was not closed %s", githubReqID)
		return
	}
	pull, _, err := s.eventParser.ExtractPullData(pullEvent.PullRequest)
	if err != nil {
		s.respond(w, logging.Error, http.StatusBadRequest, "Error parsing pull data: %s", err)
		return
	}
	repo, err := s.eventParser.ExtractRepoData(pullEvent.Repo)
	if err != nil {
		s.respond(w, logging.Error, http.StatusBadRequest, "Error parsing repo data: %s", err)
		return
	}

	if err := s.pullClosedExecutor.CleanUpPull(repo, pull); err != nil {
		s.respond(w, logging.Error, http.StatusInternalServerError, "Error cleaning pull request: %s", err)
		return
	}
	s.logger.Info("deleted locks and workspace for repo %s, pull %d", repo.FullName, pull.Num)
	fmt.Fprint(w, "Pull request cleaned successfully")
}

func (s *Server) handleCommentEvent(w http.ResponseWriter, event *gh.IssueCommentEvent, githubReqID string) {
	if event.GetAction() != "created" {
		s.respond(w, logging.Debug, http.StatusOK, "Ignoring comment event since action was not created %s", githubReqID)
		return
	}

	ctx := &CommandContext{}
	command, err := s.eventParser.DetermineCommand(event)
	if err != nil {
		s.respond(w, logging.Debug, http.StatusOK, "Ignoring: %s %s", err, githubReqID)
		return
	}
	ctx.Command = command

	if err = s.eventParser.ExtractCommentData(event, ctx); err != nil {
		s.respond(w, logging.Error, http.StatusInternalServerError, "Failed parsing event: %v %s", err, githubReqID)
		return
	}
	// respond with success and then actually execute the command asynchronously
	fmt.Fprintln(w, "Processing...")
	go s.commandHandler.ExecuteCommand(ctx)
}

func (s *Server) respond(w http.ResponseWriter, lvl logging.LogLevel, code int, format string, args ...interface{}) {
	response := fmt.Sprintf(format, args...)
	s.logger.Log(lvl, response)
	w.WriteHeader(code)
	fmt.Fprintln(w, response)
}
