package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/palantir/go-githubapp/githubapp"
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
	"github.com/runatlantis/atlantis/server/lyft/feature"
	lyft_gateway "github.com/runatlantis/atlantis/server/lyft/gateway"
	"github.com/runatlantis/atlantis/server/metrics"
	"github.com/runatlantis/atlantis/server/neptune/gateway/event"
	"github.com/runatlantis/atlantis/server/neptune/gateway/event/preworkflow"
	httpInternal "github.com/runatlantis/atlantis/server/neptune/http"
	"github.com/runatlantis/atlantis/server/neptune/sync"
	internalSync "github.com/runatlantis/atlantis/server/neptune/sync"
	"github.com/runatlantis/atlantis/server/neptune/sync/crons"
	"github.com/runatlantis/atlantis/server/neptune/temporal"
	middleware "github.com/runatlantis/atlantis/server/neptune/workflows/activities/github"
	"github.com/runatlantis/atlantis/server/vcs/markdown"
	"github.com/runatlantis/atlantis/server/vcs/provider/github"
	github_converter "github.com/runatlantis/atlantis/server/vcs/provider/github/converter"
	"github.com/runatlantis/atlantis/server/wrappers"
	"github.com/urfave/cli"
	"github.com/urfave/negroni"
	"go.temporal.io/sdk/client"
	"golang.org/x/sync/errgroup"
)

// TODO: let's make this struct nicer using actual OOP instead of just a god type struct
type Config struct {
	DataDir                   string
	AutoplanFileList          string
	AppCfg                    githubapp.Config
	RepoAllowList             string
	MaxProjectsPerPR          int
	FFOwner                   string
	FFRepo                    string
	FFBranch                  string
	FFPath                    string
	GithubHostname            string
	GithubWebhookSecret       string
	GithubAppID               int64
	GithubAppKeyFile          string
	GithubAppSlug             string
	GithubStatusName          string
	LogLevel                  logging.LogLevel
	StatsNamespace            string
	Port                      int
	RepoConfig                string
	TFDownloadURL             string
	SNSTopicArn               string
	SSLKeyFile                string
	SSLCertFile               string
	DefaultCheckrunDetailsURL string
}

type Server struct {
	Crons          []*internalSync.Cron
	StatsCloser    io.Closer
	Handler        http.Handler
	Logger         logging.Logger
	Port           int
	Drainer        *events.Drainer
	Scheduler      *sync.AsyncScheduler
	Server         httpInternal.ServerProxy
	TemporalClient client.Client
	CronScheduler  *internalSync.CronScheduler
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

	globalCfg := valid.NewGlobalCfg(config.DataDir)
	validator := &cfgParser.ParserValidator{}
	if config.RepoConfig != "" {
		globalCfg, err = validator.ParseGlobalCfg(config.RepoConfig, globalCfg)
		if err != nil {
			return nil, errors.Wrapf(err, "parsing %s file", config.RepoConfig)
		}
	}

	statsReporter, err := metrics.NewReporter(globalCfg.Metrics, ctxLogger)

	if err != nil {
		return nil, err
	}

	statsScope, closer := metrics.NewScopeWithReporter(globalCfg.Metrics, ctxLogger, config.StatsNamespace, statsReporter)
	if err != nil {
		return nil, err
	}
	statsScope = statsScope.Tagged(map[string]string{
		"mode": "gateway",
	})

	privateKey, err := os.ReadFile(config.GithubAppKeyFile)
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

	vcsClient := vcs.NewInstrumentedGithubClient(rawGithubClient, statsScope, ctxLogger)

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

	outputUpdater := &events.ChecksOutputUpdater{
		VCSClient:        vcsClient,
		MarkdownRenderer: markdownRenderer,
		TitleBuilder:     vcs.StatusTitleBuilder{TitlePrefix: config.GithubStatusName},
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

	syncScheduler := &sync.SynchronousScheduler{
		Logger: ctxLogger,
	}
	asyncScheduler := sync.NewAsyncScheduler(ctxLogger, syncScheduler)

	gatewaySnsWriter := sns.NewWriterWithStats(session, config.SNSTopicArn, statsScope.SubScope("aws.sns.gateway"))
	autoplanValidator := &lyft_gateway.AutoplanValidator{
		Scope:                         statsScope.SubScope("validator"),
		VCSClient:                     vcsClient,
		PreWorkflowHooksCommandRunner: preWorkflowHooksCommandRunner,
		Drainer:                       drainer,
		GlobalCfg:                     globalCfg,
		VCSStatusUpdater:              &command.VCSStatusUpdater{Client: vcsClient, TitleBuilder: vcs.StatusTitleBuilder{TitlePrefix: config.GithubStatusName}},
		PrjCmdBuilder:                 projectCommandBuilder,
		OutputUpdater:                 outputUpdater,
		WorkingDir:                    workingDir,
		WorkingDirLocker:              workingDirLocker,
		Allocator:                     featureAllocator,
	}

	repoConverter := github_converter.RepoConverter{}

	pullConverter := github_converter.PullConverter{
		RepoConverter: repoConverter,
	}

	opts := &temporal.Options{
		StatsReporter: statsReporter,
	}
	opts = opts.WithClientInterceptors(temporal.NewMetricsInterceptor(statsScope))
	temporalClient, err := temporal.NewClient(ctxLogger, globalCfg.Temporal, opts)

	if err != nil {
		return nil, errors.Wrap(err, "initializing temporal client")
	}

	repoFetcher := &github.RepoFetcher{
		DataDir:           config.DataDir,
		GithubCredentials: githubCredentials,
		GithubHostname:    config.GithubHostname,
		Logger:            ctxLogger,
	}
	hooksRunner := &preworkflow.HooksRunner{
		GlobalCfg:    globalCfg,
		HookExecutor: &preworkflow.HookExecutor{},
	}
	clientCreator, err := githubapp.NewDefaultCachingClientCreator(
		config.AppCfg,
		githubapp.WithClientMiddleware(
			middleware.ClientMetrics(statsScope.SubScope("github")),
		))
	if err != nil {
		return nil, errors.Wrap(err, "creating github client creator")
	}

	rootConfigBuilder := &event.RootConfigBuilder{
		RepoFetcher:     repoFetcher,
		HooksRunner:     hooksRunner,
		ParserValidator: &event.ParserValidator{GlobalCfg: globalCfg},
		RootFinder:      &event.RepoRootFinder{Logger: ctxLogger},
		FileFetcher:     &github.RemoteFileFetcher{ClientCreator: clientCreator},
		GlobalCfg:       globalCfg,
		Logger:          ctxLogger,
		Scope:           statsScope.SubScope("event.filters.root"),
	}

	checkRunFetcher := &github.CheckRunsFetcher{
		AppID:         config.GithubAppID,
		ClientCreator: clientCreator,
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
		featureAllocator,
		syncScheduler,
		asyncScheduler,
		temporalClient,
		rootConfigBuilder,
		checkRunFetcher,
	)

	router := mux.NewRouter()
	router.HandleFunc("/healthz", Healthz).Methods("GET")
	router.HandleFunc("/status", statusController.Get).Methods("GET")
	router.HandleFunc("/events", gatewayEventsController.Post).Methods("POST")
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)

	n := negroni.New(&negroni.Recovery{
		Logger:     log.New(os.Stdout, "", log.LstdFlags),
		PrintStack: false,
		StackAll:   false,
		StackSize:  1024 * 8,
	})
	n.UseHandler(router)

	s := httpInternal.ServerProxy{
		Server: &http.Server{
			Addr:              fmt.Sprintf(":%d", config.Port),
			Handler:           n,
			ReadHeaderTimeout: time.Second * 10,
		},
		SSLCertFile: config.SSLCertFile,
		SSLKeyFile:  config.SSLKeyFile,
	}

	cronScheduler := internalSync.NewCronScheduler(ctxLogger)

	return &Server{
		Crons: []*internalSync.Cron{
			{
				Executor:  crons.NewRuntimeStats(statsScope).Run,
				Frequency: 1 * time.Minute,
			},
		},
		StatsCloser:    closer,
		Handler:        n,
		Scheduler:      asyncScheduler,
		Logger:         ctxLogger,
		Port:           config.Port,
		Drainer:        drainer,
		TemporalClient: temporalClient,
		Server:         s,
		CronScheduler:  cronScheduler,
	}, nil

}

// Start is blocking and listens for incoming requests until a configured shutdown
// signal is received.
func (s *Server) Start() error {
	// we create a base context that is marked done when we get a sigterm.
	// we should use this context for other async work to ensure we
	// are gracefully handling shutdown and not dropping data.
	mainCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// error group here makes it easier to add other processes and share a ctx between them
	group, gCtx := errgroup.WithContext(mainCtx)
	group.Go(func() error {
		s.Logger.Info(fmt.Sprintf("Atlantis started - listening on port %v", s.Port))
		err := s.Server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			s.Logger.Error(err.Error())
		}

		return err
	})

	for _, c := range s.Crons {
		s.CronScheduler.Schedule(c)
	}

	<-gCtx.Done()
	s.Logger.Warn("Received interrupt. Waiting for in-progress operations to complete")

	if err := s.Shutdown(); err != nil {
		return err
	}

	if err := group.Wait(); err != nil {
		return err
	}

	return nil
}

func (s *Server) Shutdown() error {
	defer s.Logger.Close()
	defer s.TemporalClient.Close()

	// legacy way of draining ops, we should remove
	// in favor of context based approach.
	s.waitForDrain()

	// block on async work for 30 seconds max
	s.Scheduler.Shutdown(30 * time.Second)

	// flush stats before shutdown
	if err := s.StatsCloser.Close(); err != nil {
		s.Logger.Error(err.Error())
	}

	// wait for 5 seconds to shutdown http server and drain existing requests if any.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Server.Shutdown(ctx); err != nil {
		return cli.NewExitError(fmt.Sprintf("while shutting down: %s", err), 1)
	}

	s.CronScheduler.Shutdown(5 * time.Second)

	s.TemporalClient.Close()

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
func Healthz(w http.ResponseWriter, _ *http.Request) {
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
