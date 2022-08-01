package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/controllers"
	cfgParser "github.com/runatlantis/atlantis/server/core/config"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/core/runtime"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/instrumentation"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/lyft/aws"
	"github.com/runatlantis/atlantis/server/lyft/aws/sns"
	lyft_checks "github.com/runatlantis/atlantis/server/lyft/checks"
	"github.com/runatlantis/atlantis/server/lyft/feature"
	lyft_gateway "github.com/runatlantis/atlantis/server/lyft/gateway"
	"github.com/runatlantis/atlantis/server/metrics"
	"github.com/runatlantis/atlantis/server/vcs/markdown"
	github_converter "github.com/runatlantis/atlantis/server/vcs/provider/github/converter"
	"github.com/runatlantis/atlantis/server/wrappers"
	"github.com/urfave/cli"
	"github.com/urfave/negroni"
)

// TODO: let's make this struct nicer using actual OOP instead of just a god type struct
type Config struct {
	DataDir             string
	AutoplanFileList    string
	RepoAllowList       string
	MaxProjectsPerPR    int
	FFOwner             string
	FFRepo              string
	FFBranch            string
	FFPath              string
	GithubHostname      string
	GithubWebhookSecret string
	GithubAppID         int64
	GithubAppKeyFile    string
	GithubAppSlug       string
	GithubStatusName    string
	LogLevel            logging.LogLevel
	StatsNamespace      string
	Port                int
	RepoConfig          string
	TFDownloadURL       string
	SNSTopicArn         string
	SSLKeyFile          string
	SSLCertFile         string
}

type Server struct {
	StatsCloser      io.Closer
	Router           *mux.Router
	EventController  *lyft_gateway.VCSEventsController
	StatusController *controllers.StatusController
	Logger           logging.Logger
	Port             int
	Drainer          *events.Drainer
	SSLKeyFile       string
	SSLCertFile      string
}

// NewServer injects all dependencies nothing should "start" here
func NewServer(config Config) (*Server, error) {
	ctxLogger, err := logging.NewLoggerFromLevel(config.LogLevel)
	if err != nil {
		return nil, err
	}

	repoAllowlist, err := events.NewRepoAllowlistChecker(config.RepoAllowList)
	if err != nil {
		return nil, err
	}

	globalCfg := valid.NewGlobalCfg()
	validator := &cfgParser.ParserValidator{}
	if config.RepoConfig != "" {
		globalCfg, err = validator.ParseGlobalCfg(config.RepoConfig, globalCfg)
		if err != nil {
			return nil, errors.Wrapf(err, "parsing %s file", config.RepoConfig)
		}
	}

	statsScope, closer, err := metrics.NewScope(globalCfg.Metrics, ctxLogger, config.StatsNamespace)
	if err != nil {
		return nil, err
	}

	privateKey, err := ioutil.ReadFile(config.GithubAppKeyFile)
	if err != nil {
		return nil, err
	}
	githubCredentials := &vcs.GithubAppCredentials{
		AppID:    config.GithubAppID,
		Key:      privateKey,
		Hostname: config.GithubHostname,
		AppSlug:  config.GithubAppSlug,
	}

	mergeabilityChecker := vcs.NewLyftPullMergeabilityChecker(config.GithubStatusName)

	rawGithubClient, err := vcs.NewGithubClient(config.GithubHostname, githubCredentials, ctxLogger, mergeabilityChecker)
	if err != nil {
		return nil, err
	}

	featureAllocator, err := feature.NewGHSourcedAllocator(
		feature.RepoConfig{
			Owner:  config.FFOwner,
			Repo:   config.FFRepo,
			Branch: config.FFBranch,
			Path:   config.FFPath,
		}, rawGithubClient, ctxLogger)

	if err != nil {
		return nil, errors.Wrap(err, "initializing feature allocator")
	}

	// [WENGINES-4643] TODO: Remove this wrapped client once github checks is stable
	checksWrapperGhClient := &lyft_checks.ChecksClientWrapper{
		FeatureAllocator: featureAllocator,
		Logger:           ctxLogger,
		GithubClient:     rawGithubClient,
	}

	vcsClient := vcs.NewInstrumentedGithubClient(rawGithubClient, checksWrapperGhClient, statsScope, ctxLogger)

	workingDirLocker := events.NewDefaultWorkingDirLocker()

	var workingDir events.WorkingDir = &events.GithubAppWorkingDir{
		WorkingDir: &events.FileWorkspace{
			DataDir:   config.DataDir,
			GlobalCfg: globalCfg,
		},
		Credentials: githubCredentials,
	}

	var preWorkflowHooksCommandRunner events.PreWorkflowHooksCommandRunner
	preWorkflowHooksCommandRunner = &events.DefaultPreWorkflowHooksCommandRunner{
		VCSClient:             vcsClient,
		GlobalCfg:             globalCfg,
		WorkingDirLocker:      workingDirLocker,
		WorkingDir:            workingDir,
		PreWorkflowHookRunner: runtime.DefaultPreWorkflowHookRunner{},
	}
	preWorkflowHooksCommandRunner = &instrumentation.PreWorkflowHookRunner{
		PreWorkflowHooksCommandRunner: preWorkflowHooksCommandRunner,
		Logger:                        ctxLogger,
	}

	templateResolver := markdown.TemplateResolver{
		DisableMarkdownFolding:   false,
		GitlabSupportsCommonMark: false,
		GlobalCfg:                globalCfg,
	}
	markdownRenderer := &markdown.Renderer{
		DisableApplyAll:          false,
		DisableApply:             false,
		EnableDiffMarkdownFormat: true,
		TemplateResolver:         templateResolver,
	}

	pullOutputUpdater := events.PullOutputUpdater{
		VCSClient:            vcsClient,
		MarkdownRenderer:     markdownRenderer,
		HidePrevPlanComments: true,
	}

	checksOutputUpdater := events.ChecksOutputUpdater{
		VCSClient:        vcsClient,
		MarkdownRenderer: markdownRenderer,
		TitleBuilder:     vcs.StatusTitleBuilder{TitlePrefix: config.GithubStatusName},
	}

	// [WENGINES-4643] TODO: Remove pullOutputUpdater once github checks is stable
	outputUpdater := &events.FeatureAwareChecksOutputUpdater{
		ChecksOutputUpdater: checksOutputUpdater,
		PullOutputUpdater:   pullOutputUpdater,
		FeatureAllocator:    featureAllocator,
		Logger:              ctxLogger,
	}

	session, err := aws.NewSession()
	if err != nil {
		return nil, errors.Wrap(err, "initializing new aws session")
	}

	drainer := &events.Drainer{}
	statusController := &controllers.StatusController{
		Logger:  ctxLogger,
		Drainer: drainer,
	}

	commentParser := &events.CommentParser{
		GithubUser: config.GithubAppSlug,
	}

	projectContextBuilder := wrappers.
		WrapProjectContext(events.NewProjectCommandContextBuilder(commentParser)).
		WithInstrumentation(statsScope).
		EnablePolicyChecks(commentParser)

	projectCommandBuilder := events.NewProjectCommandBuilder(
		projectContextBuilder,
		validator,
		&events.DefaultProjectFinder{},
		vcsClient,
		workingDir,
		workingDirLocker,
		globalCfg,
		&events.DefaultPendingPlanFinder{},
		true,
		config.AutoplanFileList,
		ctxLogger,
		config.MaxProjectsPerPR,
	)

	gatewaySnsWriter := sns.NewWriterWithStats(session, config.SNSTopicArn, statsScope.SubScope("aws.sns.gateway"))
	autoplanValidator := &lyft_gateway.AutoplanValidator{
		Scope:                         statsScope.SubScope("validator"),
		VCSClient:                     vcsClient,
		PreWorkflowHooksCommandRunner: preWorkflowHooksCommandRunner,
		Drainer:                       drainer,
		GlobalCfg:                     globalCfg,
		CommitStatusUpdater:           &command.VCSStatusUpdater{Client: vcsClient, TitleBuilder: vcs.StatusTitleBuilder{TitlePrefix: config.GithubStatusName}},
		PrjCmdBuilder:                 projectCommandBuilder,
		OutputUpdater:                 outputUpdater,
		WorkingDir:                    workingDir,
		WorkingDirLocker:              workingDirLocker,
	}

	repoConverter := github_converter.RepoConverter{}

	pullConverter := github_converter.PullConverter{
		RepoConverter: repoConverter,
	}

	gatewayEventsController := lyft_gateway.NewVCSEventsController(
		statsScope,
		[]byte(config.GithubWebhookSecret),
		false,
		autoplanValidator,
		gatewaySnsWriter,
		commentParser,
		repoAllowlist,
		vcsClient,
		ctxLogger,
		[]models.VCSHostType{models.Github},
		repoConverter,
		pullConverter,
		vcsClient,
	)

	router := mux.NewRouter()

	return &Server{
		StatsCloser:      closer,
		Router:           router,
		EventController:  gatewayEventsController,
		StatusController: statusController,
		Logger:           ctxLogger,
		Port:             config.Port,
		Drainer:          drainer,
	}, nil

}

// Start is blocking and listens for incoming requests until a configured shutdown
// signal is received.
func (s *Server) Start() error {

	// TODO: remove router initialization from here
	// I assume this is currently happening to ensure that healthz is returning ready
	// only when we are actually ready to receive requests, however, a better to do this is
	// by using some atomic value to determine when our server is ready to start taking traffic.
	// This way we'd have clean separation between setup and execution
	s.Router.HandleFunc("/healthz", s.Healthz).Methods("GET")
	s.Router.HandleFunc("/status", s.StatusController.Get).Methods("GET")
	s.Router.HandleFunc("/events", s.EventController.Post).Methods("POST")

	n := negroni.New(&negroni.Recovery{
		Logger:     log.New(os.Stdout, "", log.LstdFlags),
		PrintStack: false,
		StackAll:   false,
		StackSize:  1024 * 8,
	})
	n.UseHandler(s.Router)

	defer s.Logger.Close()

	// Ensure server gracefully drains connections when stopped.
	stop := make(chan os.Signal, 1)
	// Stop on SIGINTs and SIGTERMs.
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	server := &http.Server{Addr: fmt.Sprintf(":%d", s.Port), Handler: n}
	go func() {
		s.Logger.Info(fmt.Sprintf("Atlantis started - listening on port %v", s.Port))

		var err error
		if s.SSLCertFile != "" && s.SSLKeyFile != "" {
			err = server.ListenAndServeTLS(s.SSLCertFile, s.SSLKeyFile)
		} else {
			err = server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			s.Logger.Error(err.Error())
		}
	}()
	<-stop

	s.Logger.Warn("Received interrupt. Waiting for in-progress operations to complete")
	s.waitForDrain()

	// flush stats before shutdown
	if err := s.StatsCloser.Close(); err != nil {
		s.Logger.Error(err.Error())
	}

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second) // nolint: vet
	if err := server.Shutdown(ctx); err != nil {
		return cli.NewExitError(fmt.Sprintf("while shutting down: %s", err), 1)
	}

	return nil

}

// waitForDrain blocks until draining is complete.
func (s *Server) waitForDrain() {
	drainComplete := make(chan bool, 1)
	go func() {
		s.Drainer.ShutdownBlocking()
		drainComplete <- true
	}()
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-drainComplete:
			s.Logger.Info("All in-progress operations complete, shutting down")
			return
		case <-ticker.C:
			s.Logger.Info(fmt.Sprintf("Waiting for in-progress operations to complete, current in-progress ops: %d", s.Drainer.GetStatus().InProgressOps))
		}
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
