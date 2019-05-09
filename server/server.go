// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.

// Package server handles the web server and executing commands that come in
// via webhooks.
package server

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/runatlantis/atlantis/server/events/db"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"

	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/locking"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/runtime"
	"github.com/runatlantis/atlantis/server/events/terraform"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/events/vcs/bitbucketcloud"
	"github.com/runatlantis/atlantis/server/events/vcs/bitbucketserver"
	"github.com/runatlantis/atlantis/server/events/webhooks"
	"github.com/runatlantis/atlantis/server/events/yaml"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/static"
	"github.com/urfave/cli"
	"github.com/urfave/negroni"
)

const (
	// LockViewRouteName is the named route in mux.Router for the lock view.
	// The route can be retrieved by this name, ex:
	//   mux.Router.Get(LockViewRouteName)
	LockViewRouteName = "lock-detail"
	// LockViewRouteIDQueryParam is the query parameter needed to construct the lock view
	// route. ex:
	//   mux.Router.Get(LockViewRouteName).URL(LockViewRouteIDQueryParam, "my id")
	LockViewRouteIDQueryParam = "id"
)

// Server runs the Atlantis web server.
type Server struct {
	AtlantisVersion    string
	AtlantisURL        *url.URL
	Router             *mux.Router
	Port               int
	CommandRunner      *events.DefaultCommandRunner
	Logger             *logging.SimpleLogger
	Locker             locking.Locker
	EventsController   *EventsController
	LocksController    *LocksController
	IndexTemplate      TemplateWriter
	LockDetailTemplate TemplateWriter
	SSLCertFile        string
	SSLKeyFile         string
}

// Config holds config for server that isn't passed in by the user.
type Config struct {
	AllowForkPRsFlag     string
	AtlantisURLFlag      string
	AtlantisVersion      string
	DefaultTFVersionFlag string
	RepoConfigJSONFlag   string
}

// WebhookConfig is nested within UserConfig. It's used to configure webhooks.
type WebhookConfig struct {
	// Event is the type of event we should send this webhook for, ex. apply.
	Event string `mapstructure:"event"`
	// WorkspaceRegex is a regex that is used to match against the workspace
	// that is being modified for this event. If the regex matches, we'll
	// send the webhook, ex. "production.*".
	WorkspaceRegex string `mapstructure:"workspace-regex"`
	// Kind is the type of webhook we should send, ex. slack.
	Kind string `mapstructure:"kind"`
	// Channel is the channel to send this webhook to. It only applies to
	// slack webhooks. Should be without '#'.
	Channel string `mapstructure:"channel"`
}

// NewServer returns a new server. If there are issues starting the server or
// its dependencies an error will be returned. This is like the main() function
// for the server CLI command because it injects all the dependencies.
func NewServer(userConfig UserConfig, config Config) (*Server, error) {
	logger := logging.NewSimpleLogger("server", false, userConfig.ToLogLevel())
	var supportedVCSHosts []models.VCSHostType
	var githubClient *vcs.GithubClient
	var gitlabClient *vcs.GitlabClient
	var bitbucketCloudClient *bitbucketcloud.Client
	var bitbucketServerClient *bitbucketserver.Client
	if userConfig.GithubUser != "" {
		supportedVCSHosts = append(supportedVCSHosts, models.Github)
		var err error
		githubClient, err = vcs.NewGithubClient(userConfig.GithubHostname, userConfig.GithubUser, userConfig.GithubToken)
		if err != nil {
			return nil, err
		}
	}
	if userConfig.GitlabUser != "" {
		supportedVCSHosts = append(supportedVCSHosts, models.Gitlab)
		var err error
		gitlabClient, err = vcs.NewGitlabClient(userConfig.GitlabHostname, userConfig.GitlabToken, logger)
		if err != nil {
			return nil, err
		}
	}
	if userConfig.BitbucketUser != "" {
		if userConfig.BitbucketBaseURL == bitbucketcloud.BaseURL {
			supportedVCSHosts = append(supportedVCSHosts, models.BitbucketCloud)
			bitbucketCloudClient = bitbucketcloud.NewClient(
				http.DefaultClient,
				userConfig.BitbucketUser,
				userConfig.BitbucketToken,
				userConfig.AtlantisURL)
		} else {
			supportedVCSHosts = append(supportedVCSHosts, models.BitbucketServer)
			var err error
			bitbucketServerClient, err = bitbucketserver.NewClient(
				http.DefaultClient,
				userConfig.BitbucketUser,
				userConfig.BitbucketToken,
				userConfig.BitbucketBaseURL,
				userConfig.AtlantisURL)
			if err != nil {
				return nil, errors.Wrapf(err, "setting up Bitbucket Server client")
			}
		}
	}

	var webhooksConfig []webhooks.Config
	for _, c := range userConfig.Webhooks {
		config := webhooks.Config{
			Channel:        c.Channel,
			Event:          c.Event,
			Kind:           c.Kind,
			WorkspaceRegex: c.WorkspaceRegex,
		}
		webhooksConfig = append(webhooksConfig, config)
	}
	webhooksManager, err := webhooks.NewMultiWebhookSender(webhooksConfig, webhooks.NewSlackClient(userConfig.SlackToken))
	if err != nil {
		return nil, errors.Wrap(err, "initializing webhooks")
	}
	vcsClient := vcs.NewClientProxy(githubClient, gitlabClient, bitbucketCloudClient, bitbucketServerClient)
	commitStatusUpdater := &events.DefaultCommitStatusUpdater{Client: vcsClient}
	terraformClient, err := terraform.NewClient(logger, userConfig.DataDir, userConfig.TFEToken, userConfig.DefaultTFVersion, config.DefaultTFVersionFlag, &terraform.DefaultDownloader{})
	// The flag.Lookup call is to detect if we're running in a unit test. If we
	// are, then we don't error out because we don't have/want terraform
	// installed on our CI system where the unit tests run.
	if err != nil && flag.Lookup("test.v") == nil {
		return nil, errors.Wrap(err, "initializing terraform")
	}
	markdownRenderer := &events.MarkdownRenderer{
		GitlabSupportsCommonMark: gitlabClient.SupportsCommonMark(),
		DisableApplyAll:          userConfig.DisableApplyAll,
	}
	boltdb, err := db.New(userConfig.DataDir)
	if err != nil {
		return nil, err
	}
	lockingClient := locking.NewClient(boltdb)
	workingDirLocker := events.NewDefaultWorkingDirLocker()
	workingDir := &events.FileWorkspace{
		DataDir:       userConfig.DataDir,
		CheckoutMerge: userConfig.CheckoutStrategy == "merge",
	}
	projectLocker := &events.DefaultProjectLocker{
		Locker: lockingClient,
	}
	parsedURL, err := ParseAtlantisURL(userConfig.AtlantisURL)
	if err != nil {
		return nil, errors.Wrapf(err,
			"parsing --%s flag %q", config.AtlantisURLFlag, userConfig.AtlantisURL)
	}
	validator := &yaml.ParserValidator{}

	globalCfg := valid.NewGlobalCfg(userConfig.AllowRepoConfig, userConfig.RequireMergeable, userConfig.RequireApproval)
	if userConfig.RepoConfig != "" {
		globalCfg, err = validator.ParseGlobalCfg(userConfig.RepoConfig, globalCfg)
		if err != nil {
			return nil, errors.Wrapf(err, "parsing %s file", userConfig.RepoConfig)
		}
	} else if userConfig.RepoConfigJSON != "" {
		globalCfg, err = validator.ParseGlobalCfgJSON(userConfig.RepoConfigJSON, globalCfg)
		if err != nil {
			return nil, errors.Wrapf(err, "parsing --%s", config.RepoConfigJSONFlag)
		}
	}

	underlyingRouter := mux.NewRouter()
	router := &Router{
		AtlantisURL:               parsedURL,
		LockViewRouteIDQueryParam: LockViewRouteIDQueryParam,
		LockViewRouteName:         LockViewRouteName,
		Underlying:                underlyingRouter,
	}
	pullClosedExecutor := &events.PullClosedExecutor{
		VCSClient:  vcsClient,
		Locker:     lockingClient,
		WorkingDir: workingDir,
		Logger:     logger,
		DB:         boltdb,
	}
	eventParser := &events.EventParser{
		GithubUser:         userConfig.GithubUser,
		GithubToken:        userConfig.GithubToken,
		GitlabUser:         userConfig.GitlabUser,
		GitlabToken:        userConfig.GitlabToken,
		BitbucketUser:      userConfig.BitbucketUser,
		BitbucketToken:     userConfig.BitbucketToken,
		BitbucketServerURL: userConfig.BitbucketBaseURL,
	}
	commentParser := &events.CommentParser{
		GithubUser:    userConfig.GithubUser,
		GitlabUser:    userConfig.GitlabUser,
		BitbucketUser: userConfig.BitbucketUser,
	}
	defaultTfVersion := terraformClient.DefaultVersion()
	pendingPlanFinder := &events.DefaultPendingPlanFinder{}
	commandRunner := &events.DefaultCommandRunner{
		VCSClient:                vcsClient,
		GithubPullGetter:         githubClient,
		GitlabMergeRequestGetter: gitlabClient,
		CommitStatusUpdater:      commitStatusUpdater,
		EventParser:              eventParser,
		MarkdownRenderer:         markdownRenderer,
		Logger:                   logger,
		AllowForkPRs:             userConfig.AllowForkPRs,
		AllowForkPRsFlag:         config.AllowForkPRsFlag,
		DisableApplyAll:          userConfig.DisableApplyAll,
		ProjectCommandBuilder: &events.DefaultProjectCommandBuilder{
			ParserValidator:   validator,
			ProjectFinder:     &events.DefaultProjectFinder{},
			VCSClient:         vcsClient,
			WorkingDir:        workingDir,
			WorkingDirLocker:  workingDirLocker,
			GlobalCfg:         globalCfg,
			PendingPlanFinder: pendingPlanFinder,
			CommentBuilder:    commentParser,
		},
		ProjectCommandRunner: &events.DefaultProjectCommandRunner{
			Locker:           projectLocker,
			LockURLGenerator: router,
			InitStepRunner: &runtime.InitStepRunner{
				TerraformExecutor: terraformClient,
				DefaultTFVersion:  defaultTfVersion,
			},
			PlanStepRunner: &runtime.PlanStepRunner{
				TerraformExecutor:   terraformClient,
				DefaultTFVersion:    defaultTfVersion,
				CommitStatusUpdater: commitStatusUpdater,
				AsyncTFExec:         terraformClient,
			},
			ApplyStepRunner: &runtime.ApplyStepRunner{
				TerraformExecutor:   terraformClient,
				CommitStatusUpdater: commitStatusUpdater,
				AsyncTFExec:         terraformClient,
			},
			RunStepRunner: &runtime.RunStepRunner{
				DefaultTFVersion: defaultTfVersion,
			},
			PullApprovedChecker: vcsClient,
			WorkingDir:          workingDir,
			Webhooks:            webhooksManager,
			WorkingDirLocker:    workingDirLocker,
		},
		WorkingDir:        workingDir,
		PendingPlanFinder: pendingPlanFinder,
		DB:                boltdb,
		GlobalAutomerge:   userConfig.Automerge,
	}
	repoWhitelist, err := events.NewRepoWhitelistChecker(userConfig.RepoWhitelist)
	if err != nil {
		return nil, err
	}
	locksController := &LocksController{
		AtlantisVersion:    config.AtlantisVersion,
		AtlantisURL:        parsedURL,
		Locker:             lockingClient,
		Logger:             logger,
		VCSClient:          vcsClient,
		LockDetailTemplate: lockTemplate,
		WorkingDir:         workingDir,
		WorkingDirLocker:   workingDirLocker,
		DB:                 boltdb,
	}
	eventsController := &EventsController{
		CommandRunner:                commandRunner,
		PullCleaner:                  pullClosedExecutor,
		Parser:                       eventParser,
		CommentParser:                commentParser,
		Logger:                       logger,
		GithubWebhookSecret:          []byte(userConfig.GithubWebhookSecret),
		GithubRequestValidator:       &DefaultGithubRequestValidator{},
		GitlabRequestParserValidator: &DefaultGitlabRequestParserValidator{},
		GitlabWebhookSecret:          []byte(userConfig.GitlabWebhookSecret),
		RepoWhitelistChecker:         repoWhitelist,
		SilenceWhitelistErrors:       userConfig.SilenceWhitelistErrors,
		SupportedVCSHosts:            supportedVCSHosts,
		VCSClient:                    vcsClient,
		BitbucketWebhookSecret:       []byte(userConfig.BitbucketWebhookSecret),
	}
	return &Server{
		AtlantisVersion:    config.AtlantisVersion,
		AtlantisURL:        parsedURL,
		Router:             underlyingRouter,
		Port:               userConfig.Port,
		CommandRunner:      commandRunner,
		Logger:             logger,
		Locker:             lockingClient,
		EventsController:   eventsController,
		LocksController:    locksController,
		IndexTemplate:      indexTemplate,
		LockDetailTemplate: lockTemplate,
		SSLKeyFile:         userConfig.SSLKeyFile,
		SSLCertFile:        userConfig.SSLCertFile,
	}, nil
}

// Start creates the routes and starts serving traffic.
func (s *Server) Start() error {
	s.Router.HandleFunc("/", s.Index).Methods("GET").MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
		return r.URL.Path == "/" || r.URL.Path == "/index.html"
	})
	s.Router.HandleFunc("/healthz", s.Healthz).Methods("GET")
	s.Router.PathPrefix("/static/").Handler(http.FileServer(&assetfs.AssetFS{Asset: static.Asset, AssetDir: static.AssetDir, AssetInfo: static.AssetInfo}))
	s.Router.HandleFunc("/events", s.EventsController.Post).Methods("POST")
	s.Router.HandleFunc("/locks", s.LocksController.DeleteLock).Methods("DELETE").Queries("id", "{id:.*}")
	s.Router.HandleFunc("/lock", s.LocksController.GetLock).Methods("GET").
		Queries(LockViewRouteIDQueryParam, fmt.Sprintf("{%s}", LockViewRouteIDQueryParam)).Name(LockViewRouteName)
	n := negroni.New(&negroni.Recovery{
		Logger:     log.New(os.Stdout, "", log.LstdFlags),
		PrintStack: false,
		StackAll:   false,
		StackSize:  1024 * 8,
	}, NewRequestLogger(s.Logger))
	n.UseHandler(s.Router)

	// Ensure server gracefully drains connections when stopped.
	stop := make(chan os.Signal, 1)
	// Stop on SIGINTs and SIGTERMs.
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	server := &http.Server{Addr: fmt.Sprintf(":%d", s.Port), Handler: n}
	go func() {
		s.Logger.Info("Atlantis started - listening on port %v", s.Port)

		var err error
		if s.SSLCertFile != "" && s.SSLKeyFile != "" {
			err = server.ListenAndServeTLS(s.SSLCertFile, s.SSLKeyFile)
		} else {
			err = server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			s.Logger.Err(err.Error())
		}
	}()
	<-stop

	s.Logger.Warn("Received interrupt. Safely shutting down")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second) // nolint: vet
	if err := server.Shutdown(ctx); err != nil {
		return cli.NewExitError(fmt.Sprintf("while shutting down: %s", err), 1)
	}
	return nil
}

// Index is the / route.
func (s *Server) Index(w http.ResponseWriter, _ *http.Request) {
	locks, err := s.Locker.List()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, "Could not retrieve locks: %s", err)
		return
	}

	var lockResults []LockIndexData
	for id, v := range locks {
		lockURL, _ := s.Router.Get(LockViewRouteName).URL("id", url.QueryEscape(id))
		lockResults = append(lockResults, LockIndexData{
			// NOTE: must use .String() instead of .Path because we need the
			// query params as part of the lock URL.
			LockPath:     lockURL.String(),
			RepoFullName: v.Project.RepoFullName,
			PullNum:      v.Pull.Num,
			Time:         v.Time,
		})
	}
	err = s.IndexTemplate.Execute(w, IndexData{
		Locks:           lockResults,
		AtlantisVersion: s.AtlantisVersion,
		CleanedBasePath: s.AtlantisURL.Path,
	})
	if err != nil {
		s.Logger.Err(err.Error())
	}
}

// Healthz returns the health check response. It always returns a 200 currently.
func (s *Server) Healthz(w http.ResponseWriter, _ *http.Request) {
	data, err := json.MarshalIndent(&struct {
		Status string `json:"status"`
	}{
		Status: "ok",
	}, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error creating status json response: %s", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data) // nolint: errcheck
}

// ParseAtlantisURL parses the user-passed atlantis URL to ensure it is valid
// and we can use it in our templates.
// It removes any trailing slashes from the path so we can concatenate it
// with other paths without checking.
func ParseAtlantisURL(u string) (*url.URL, error) {
	parsed, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	if !(parsed.Scheme == "http" || parsed.Scheme == "https") {
		return nil, errors.New("http or https must be specified")
	}
	// We want the path to end without a trailing slash so we know how to
	// use it in the rest of the program.
	parsed.Path = strings.TrimSuffix(parsed.Path, "/")
	return parsed, nil
}
