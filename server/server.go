package server

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/mux"
	"github.com/hootsuite/atlantis/server/github"
	"github.com/hootsuite/atlantis/server/locking"
	"github.com/hootsuite/atlantis/server/locking/boltdb"
	"github.com/hootsuite/atlantis/server/logging"
	"github.com/hootsuite/atlantis/server/run"
	"github.com/hootsuite/atlantis/server/static"
	"github.com/hootsuite/atlantis/server/terraform"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"github.com/urfave/negroni"
	"github.com/hootsuite/atlantis/server/events"
)

const lockRoute = "lock-detail"

// Server listens for GitHub events and runs the necessary Atlantis command
type Server struct {
	router           *mux.Router
	port             int
	commandHandler   *events.CommandHandler
	logger           *logging.SimpleLogger
	eventParser      *events.EventParser
	locker           locking.Locker
	atlantisURL      string
	eventsController *EventsController
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
	githubStatus := &events.GithubStatus{Client: githubClient}
	terraformClient, err := terraform.NewClient()
	if err != nil {
		return nil, errors.Wrap(err, "initializing terraform")
	}
	githubComments := &events.GithubCommentRenderer{}

	boltdb, err := boltdb.New(config.DataDir)
	if err != nil {
		return nil, err
	}
	lockingClient := locking.NewClient(boltdb)
	run := &run.Run{}
	configReader := &events.ConfigReader{}
	concurrentRunLocker := events.NewEnvLock()
	workspace := &events.FileWorkspace{
		DataDir: config.DataDir,
	}
	applyExecutor := &events.ApplyExecutor{
		Github:          githubClient,
		Terraform:       terraformClient,
		Locker:          lockingClient,
		RequireApproval: config.RequireApproval,
		Run:             run,
		ConfigReader:    configReader,
		Workspace:       workspace,
	}
	planExecutor := &events.PlanExecutor{
		Github:       githubClient,
		Terraform:    terraformClient,
		Locker:       lockingClient,
		Run:          run,
		ConfigReader: configReader,
		Workspace:    workspace,
	}
	helpExecutor := &events.HelpExecutor{}
	pullClosedExecutor := &events.PullClosedExecutor{
		Github:    githubClient,
		Locker:    lockingClient,
		Workspace: workspace,
	}
	logger := logging.NewSimpleLogger("server", log.New(os.Stderr, "", log.LstdFlags), false, logging.ToLogLevel(config.LogLevel))
	eventParser := &events.EventParser{
		GithubUser:  config.GithubUser,
		GithubToken: config.GithubToken,
	}
	commandHandler := &events.CommandHandler{
		ApplyExecutor:     applyExecutor,
		PlanExecutor:      planExecutor,
		HelpExecutor:      helpExecutor,
		EventParser:       eventParser,
		GHClient:          githubClient,
		GHStatus:          githubStatus,
		EnvLocker:         concurrentRunLocker,
		GHCommentRenderer: githubComments,
		Logger:            logger,
	}
	eventsController := &EventsController{
		commandHandler:      commandHandler,
		pullClosedExecutor:  pullClosedExecutor,
		parser:              eventParser,
		logger:              logger,
		githubWebHookSecret: []byte(config.GithubWebHookSecret),
	}
	router := mux.NewRouter()
	return &Server{
		router:           router,
		port:             config.Port,
		commandHandler:   commandHandler,
		eventParser:      eventParser,
		logger:           logger,
		locker:           lockingClient,
		atlantisURL:      config.AtlantisURL,
		eventsController: eventsController,
	}, nil
}

func (s *Server) Start() error {
	s.router.HandleFunc("/", s.index).Methods("GET").MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
		return r.URL.Path == "/" || r.URL.Path == "/index.html"
	})
	s.router.PathPrefix("/static/").Handler(http.FileServer(&assetfs.AssetFS{Asset: static.Asset, AssetDir: static.AssetDir, AssetInfo: static.AssetInfo}))
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
	locks, err := s.locker.List()
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
	lock, err := s.locker.GetLock(idUnencoded)
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
	lock, err := s.locker.Unlock(idUnencoded)
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

// postEvents handles POST requests to our /events endpoint. These should be
// GitHub webhook requests.
func (s *Server) postEvents(w http.ResponseWriter, r *http.Request) {
	s.eventsController.Post(w, r)
}
func (s *Server) respond(w http.ResponseWriter, lvl logging.LogLevel, code int, format string, args ...interface{}) {
	response := fmt.Sprintf(format, args...)
	s.logger.Log(lvl, response)
	w.WriteHeader(code)
	fmt.Fprintln(w, response)
}
