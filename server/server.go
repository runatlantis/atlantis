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
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/runatlantis/atlantis/server/events/terraform/filter"
	"github.com/runatlantis/atlantis/server/neptune/storage"

	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/runatlantis/atlantis/server/instrumentation"
	"github.com/runatlantis/atlantis/server/static"

	"github.com/mitchellh/go-homedir"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/core/runtime/policy"
	"github.com/runatlantis/atlantis/server/jobs"
	"github.com/runatlantis/atlantis/server/lyft/aws"
	"github.com/runatlantis/atlantis/server/lyft/aws/sns"
	"github.com/runatlantis/atlantis/server/lyft/aws/sqs"
	lyftCommands "github.com/runatlantis/atlantis/server/lyft/command"
	lyftRuntime "github.com/runatlantis/atlantis/server/lyft/core/runtime"
	"github.com/runatlantis/atlantis/server/lyft/feature"
	"github.com/runatlantis/atlantis/server/lyft/scheduled"
	"github.com/runatlantis/atlantis/server/metrics"
	github_converter "github.com/runatlantis/atlantis/server/vcs/provider/github/converter"
	"github.com/runatlantis/atlantis/server/wrappers"
	"github.com/uber-go/tally/v4"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/controllers"
	events_controllers "github.com/runatlantis/atlantis/server/controllers/events"
	"github.com/runatlantis/atlantis/server/controllers/templates"
	"github.com/runatlantis/atlantis/server/controllers/websocket"
	cfgParser "github.com/runatlantis/atlantis/server/core/config"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/core/runtime"
	"github.com/runatlantis/atlantis/server/core/terraform"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/command/policies"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/events/vcs/bitbucketcloud"
	"github.com/runatlantis/atlantis/server/events/vcs/bitbucketserver"
	lyft_vcs "github.com/runatlantis/atlantis/server/events/vcs/lyft"
	"github.com/runatlantis/atlantis/server/events/webhooks"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/vcs/markdown"
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
	// ProjectJobsViewRouteName is the named route in mux.Router for the log stream view.
	ProjectJobsViewRouteName = "project-jobs-detail"
	// binDirName is the name of the directory inside our data dir where
	// we download binaries.
	BinDirName = "bin"
	// terraformPluginCacheDir is the name of the dir inside our data dir
	// where we tell terraform to cache plugins and modules.
	TerraformPluginCacheDirName = "plugin-cache"
)

// Server runs the Atlantis web server.
type Server struct {
	AtlantisVersion               string
	AtlantisURL                   *url.URL
	Router                        *mux.Router
	Port                          int
	PreWorkflowHooksCommandRunner events.PreWorkflowHooksCommandRunner
	CommandRunner                 *events.DefaultCommandRunner
	CtxLogger                     logging.Logger
	StatsScope                    tally.Scope
	StatsCloser                   io.Closer
	Locker                        locking.Locker
	ApplyLocker                   locking.ApplyLocker
	VCSPostHandler                sqs.VCSPostHandler
	GithubAppController           *controllers.GithubAppController
	LocksController               *controllers.LocksController
	StatusController              *controllers.StatusController
	JobsController                *controllers.JobsController
	IndexTemplate                 templates.TemplateWriter
	LockDetailTemplate            templates.TemplateWriter
	ProjectJobsTemplate           templates.TemplateWriter
	ProjectJobsErrorTemplate      templates.TemplateWriter
	SSLCertFile                   string
	SSLKeyFile                    string
	Drainer                       *events.Drainer
	ScheduledExecutorService      *scheduled.ExecutorService
	ProjectCmdOutputHandler       jobs.ProjectCommandOutputHandler
	LyftMode                      Mode
	CancelWorker                  context.CancelFunc
}

// Config holds config for server that isn't passed in by the user.
type Config struct {
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
	ctxLogger, err := logging.NewLoggerFromLevel(userConfig.ToLogLevel())
	if err != nil {
		return nil, err
	}

	var supportedVCSHosts []models.VCSHostType

	// not to be used directly, currently this is just used
	// for reporting rate limits
	var rawGithubClient *vcs.GithubClient

	var githubClient vcs.IGithubClient
	var githubAppEnabled bool
	var githubCredentials vcs.GithubCredentials
	var gitlabClient *vcs.GitlabClient
	var bitbucketCloudClient *bitbucketcloud.Client
	var bitbucketServerClient *bitbucketserver.Client
	var azuredevopsClient *vcs.AzureDevopsClient
	var featureAllocator feature.Allocator

	mergeabilityChecker := vcs.NewLyftPullMergeabilityChecker(userConfig.VCSStatusName)

	validator := &cfgParser.ParserValidator{}

	globalCfg := valid.NewGlobalCfg(userConfig.DataDir)

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

	statsScope, closer, err := metrics.NewScope(globalCfg.Metrics, ctxLogger, userConfig.StatsNamespace)

	logFilter := filter.LogFilter{
		Regexes: globalCfg.TerraformLogFilter.Regexes,
	}

	if err != nil {
		return nil, errors.Wrapf(err, "instantiating metrics scope")
	}

	if userConfig.GithubUser != "" || userConfig.GithubAppID != 0 {
		supportedVCSHosts = append(supportedVCSHosts, models.Github)
		if userConfig.GithubUser != "" {
			githubCredentials = &vcs.GithubUserCredentials{
				User:  userConfig.GithubUser,
				Token: userConfig.GithubToken,
			}
		} else if userConfig.GithubAppID != 0 && userConfig.GithubAppKeyFile != "" {
			privateKey, err := ioutil.ReadFile(userConfig.GithubAppKeyFile)
			if err != nil {
				return nil, err
			}
			githubCredentials = &vcs.GithubAppCredentials{
				AppID:    userConfig.GithubAppID,
				Key:      privateKey,
				Hostname: userConfig.GithubHostname,
				AppSlug:  userConfig.GithubAppSlug,
			}
			githubAppEnabled = true
		} else if userConfig.GithubAppID != 0 && userConfig.GithubAppKey != "" {
			githubCredentials = &vcs.GithubAppCredentials{
				AppID:    userConfig.GithubAppID,
				Key:      []byte(userConfig.GithubAppKey),
				Hostname: userConfig.GithubHostname,
				AppSlug:  userConfig.GithubAppSlug,
			}
			githubAppEnabled = true
		}

		var err error

		rawGithubClient, err = vcs.NewGithubClient(userConfig.GithubHostname, githubCredentials, ctxLogger, mergeabilityChecker)
		if err != nil {
			return nil, err
		}

		featureAllocator, err = feature.NewGHSourcedAllocator(
			feature.RepoConfig{
				Owner:  userConfig.FFOwner,
				Repo:   userConfig.FFRepo,
				Branch: userConfig.FFBranch,
				Path:   userConfig.FFPath,
			}, rawGithubClient, ctxLogger)

		if err != nil {
			return nil, errors.Wrap(err, "initializing feature allocator")
		}

		githubClient = vcs.NewInstrumentedGithubClient(rawGithubClient, statsScope, ctxLogger)
	}
	if userConfig.GitlabUser != "" {
		supportedVCSHosts = append(supportedVCSHosts, models.Gitlab)
		var err error
		gitlabClient, err = vcs.NewGitlabClient(userConfig.GitlabHostname, userConfig.GitlabToken, ctxLogger)
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
	if userConfig.AzureDevopsUser != "" {
		supportedVCSHosts = append(supportedVCSHosts, models.AzureDevops)
		var err error
		azuredevopsClient, err = vcs.NewAzureDevopsClient("dev.azure.com", userConfig.AzureDevopsUser, userConfig.AzureDevopsToken)
		if err != nil {
			return nil, err
		}
	}

	if userConfig.WriteGitCreds {
		home, err := homedir.Dir()
		if err != nil {
			return nil, errors.Wrap(err, "getting home dir to write ~/.git-credentials file")
		}
		if userConfig.GithubUser != "" {
			if err := events.WriteGitCreds(userConfig.GithubUser, userConfig.GithubToken, userConfig.GithubHostname, home, ctxLogger, false); err != nil {
				return nil, err
			}
		}
		if userConfig.GitlabUser != "" {
			if err := events.WriteGitCreds(userConfig.GitlabUser, userConfig.GitlabToken, userConfig.GitlabHostname, home, ctxLogger, false); err != nil {
				return nil, err
			}
		}
		if userConfig.BitbucketUser != "" {
			// The default BitbucketBaseURL is https://api.bitbucket.org which can't actually be used for git
			// so we override it here only if it's that to be bitbucket.org
			bitbucketBaseURL := userConfig.BitbucketBaseURL
			if bitbucketBaseURL == "https://api.bitbucket.org" {
				bitbucketBaseURL = "bitbucket.org"
			}
			if err := events.WriteGitCreds(userConfig.BitbucketUser, userConfig.BitbucketToken, bitbucketBaseURL, home, ctxLogger, false); err != nil {
				return nil, err
			}
		}
		if userConfig.AzureDevopsUser != "" {
			if err := events.WriteGitCreds(userConfig.AzureDevopsUser, userConfig.AzureDevopsToken, "dev.azure.com", home, ctxLogger, false); err != nil {
				return nil, err
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
	vcsClient := vcs.NewClientProxy(githubClient, gitlabClient, bitbucketCloudClient, bitbucketServerClient, azuredevopsClient)
	vcsStatusUpdater := &command.VCSStatusUpdater{
		Client: vcsClient,
		TitleBuilder: vcs.StatusTitleBuilder{
			TitlePrefix: userConfig.VCSStatusName,
		},
		DefaultDetailsURL: userConfig.DefaultCheckrunDetailsURL,
	}

	binDir, err := mkSubDir(userConfig.DataDir, BinDirName)

	if err != nil {
		return nil, err
	}

	cacheDir, err := mkSubDir(userConfig.DataDir, TerraformPluginCacheDirName)

	if err != nil {
		return nil, err
	}

	parsedURL, err := ParseAtlantisURL(userConfig.AtlantisURL)
	if err != nil {
		return nil, errors.Wrapf(err,
			"parsing --%s flag %q", config.AtlantisURLFlag, userConfig.AtlantisURL)
	}

	underlyingRouter := mux.NewRouter()
	router := &Router{
		AtlantisURL:               parsedURL,
		LockViewRouteIDQueryParam: LockViewRouteIDQueryParam,
		LockViewRouteName:         LockViewRouteName,
		ProjectJobsViewRouteName:  ProjectJobsViewRouteName,
		Underlying:                underlyingRouter,
	}

	projectJobsScope := statsScope.SubScope("getprojectjobs")

	storageClient, err := storage.NewClient(globalCfg.PersistenceConfig.Jobs)
	if err != nil {
		return nil, errors.Wrapf(err, "initializing stow client")
	}

	storageBackend, err := jobs.NewStorageBackend(storageClient, ctxLogger, featureAllocator, projectJobsScope)
	if err != nil {
		return nil, errors.Wrapf(err, "initializing storage backend")
	}

	jobStore := jobs.NewJobStore(storageBackend, statsScope.SubScope("jobstore"))

	var projectCmdOutputHandler jobs.ProjectCommandOutputHandler
	// When TFE is enabled log streaming is not necessary.

	projectCmdOutput := make(chan *jobs.ProjectCmdOutputLine)
	projectCmdOutputHandler = jobs.NewAsyncProjectCommandOutputHandler(
		projectCmdOutput,
		ctxLogger,
		jobStore,
		logFilter,
	)

	terraformClient, err := terraform.NewClient(
		binDir,
		cacheDir,
		userConfig.DefaultTFVersion,
		config.DefaultTFVersionFlag,
		userConfig.TFDownloadURL,
		&terraform.DefaultDownloader{},
		true,
		projectCmdOutputHandler)

	if err != nil {
		return nil, errors.Wrap(err, "initializing terraform")
	}

	templateResolver := markdown.TemplateResolver{
		DisableMarkdownFolding:   userConfig.DisableMarkdownFolding,
		GitlabSupportsCommonMark: gitlabClient.SupportsCommonMark(),
		GlobalCfg:                globalCfg,
		LogFilter:                logFilter,
	}
	markdownRenderer := &markdown.Renderer{
		DisableApplyAll:          userConfig.DisableApplyAll,
		DisableApply:             userConfig.DisableApply,
		EnableDiffMarkdownFormat: userConfig.EnableDiffMarkdownFormat,
		TemplateResolver:         templateResolver,
	}

	boltdb, err := db.New(userConfig.DataDir)
	if err != nil {
		return nil, err
	}
	var lockingClient locking.Locker
	var applyLockingClient locking.ApplyLocker

	lockingClient = locking.NewClient(boltdb)
	applyLockingClient = locking.NewApplyClient(boltdb, userConfig.DisableApply)
	workingDirLocker := events.NewDefaultWorkingDirLocker()

	var workingDir events.WorkingDir = &events.FileWorkspace{
		DataDir:   userConfig.DataDir,
		GlobalCfg: globalCfg,
	}
	// provide fresh tokens before clone from the GitHub Apps integration, proxy workingDir
	if githubAppEnabled {
		if !userConfig.WriteGitCreds {
			return nil, errors.New("Github App requires --write-git-creds to support cloning")
		}
		workingDir = &events.GithubAppWorkingDir{
			WorkingDir:     workingDir,
			Credentials:    githubCredentials,
			GithubHostname: userConfig.GithubHostname,
		}
	}

	projectLocker := &events.DefaultProjectLocker{
		Locker:    lockingClient,
		VCSClient: vcsClient,
	}

	deleteLockCommand := &events.DefaultDeleteLockCommand{
		Locker:           lockingClient,
		Logger:           ctxLogger,
		WorkingDir:       workingDir,
		WorkingDirLocker: workingDirLocker,
		DB:               boltdb,
	}

	pullClosedExecutor := events.NewInstrumentedPullClosedExecutor(
		statsScope,
		ctxLogger,
		&events.PullClosedExecutor{
			Locker:                   lockingClient,
			WorkingDir:               workingDir,
			Logger:                   ctxLogger,
			DB:                       boltdb,
			PullClosedTemplate:       &events.PullClosedEventTemplate{},
			LogStreamResourceCleaner: projectCmdOutputHandler,
			VCSClient:                vcsClient,
		},
	)

	eventParser := &events.EventParser{
		GithubUser:         userConfig.GithubUser,
		GithubToken:        userConfig.GithubToken,
		GitlabUser:         userConfig.GitlabUser,
		GitlabToken:        userConfig.GitlabToken,
		AllowDraftPRs:      userConfig.PlanDrafts,
		BitbucketUser:      userConfig.BitbucketUser,
		BitbucketToken:     userConfig.BitbucketToken,
		BitbucketServerURL: userConfig.BitbucketBaseURL,
		AzureDevopsUser:    userConfig.AzureDevopsUser,
		AzureDevopsToken:   userConfig.AzureDevopsToken,
	}
	commentParser := &events.CommentParser{
		GithubUser:      userConfig.GithubUser,
		GitlabUser:      userConfig.GitlabUser,
		BitbucketUser:   userConfig.BitbucketUser,
		AzureDevopsUser: userConfig.AzureDevopsUser,
		ApplyDisabled:   userConfig.DisableApply,
	}
	defaultTfVersion := terraformClient.DefaultVersion()
	pendingPlanFinder := &events.DefaultPendingPlanFinder{}

	drainer := &events.Drainer{}
	statusController := &controllers.StatusController{
		Logger:  ctxLogger,
		Drainer: drainer,
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

	legacyProjectContextBuilder := wrappers.
		WrapProjectContext(events.NewProjectCommandContextBuilder(commentParser)).
		WithInstrumentation(statsScope)

	projectContextBuilder := wrappers.
		WrapProjectContext(events.NewPlatformModeProjectCommandContextBuilder(commentParser, legacyProjectContextBuilder, ctxLogger, featureAllocator)).
		WithInstrumentation(statsScope)

	if userConfig.EnablePolicyChecks {
		projectContextBuilder = projectContextBuilder.EnablePolicyChecks(commentParser)
	}

	projectCommandBuilder := events.NewProjectCommandBuilder(
		projectContextBuilder,
		validator,
		&events.DefaultProjectFinder{},
		vcsClient,
		workingDir,
		workingDirLocker,
		globalCfg,
		pendingPlanFinder,
		userConfig.EnableRegExpCmd,
		userConfig.AutoplanFileList,
		ctxLogger,
		userConfig.MaxProjectsPerPR,
	)

	initStepRunner := &runtime.InitStepRunner{
		TerraformExecutor: terraformClient,
		DefaultTFVersion:  defaultTfVersion,
	}

	planStepRunner := &runtime.PlanStepRunner{
		TerraformExecutor: terraformClient,
		DefaultTFVersion:  defaultTfVersion,
		VCSStatusUpdater:  vcsStatusUpdater,
		AsyncTFExec:       terraformClient,
	}

	destroyPlanStepRunner := &lyftRuntime.DestroyPlanStepRunner{
		StepRunner: planStepRunner,
	}

	showStepRunner, err := runtime.NewShowStepRunner(terraformClient, defaultTfVersion)
	if err != nil {
		return nil, errors.Wrap(err, "initializing show step runner")
	}

	conftestExecutor := policy.NewConfTestExecutorWorkflow(ctxLogger, binDir, &terraform.DefaultDownloader{})
	policyCheckStepRunner, err := runtime.NewPolicyCheckStepRunner(
		defaultTfVersion,
		conftestExecutor,
	)
	if err != nil {
		return nil, errors.Wrap(err, "initializing policy check runner")
	}

	applyStepRunner := &runtime.ApplyStepRunner{
		TerraformExecutor: terraformClient,
		VCSStatusUpdater:  vcsStatusUpdater,
		AsyncTFExec:       terraformClient,
	}

	versionStepRunner := &runtime.VersionStepRunner{
		TerraformExecutor: terraformClient,
		DefaultTFVersion:  defaultTfVersion,
	}

	runStepRunner := &runtime.RunStepRunner{
		TerraformExecutor: terraformClient,
		DefaultTFVersion:  defaultTfVersion,
		TerraformBinDir:   binDir,
	}

	envStepRunner := &runtime.EnvStepRunner{
		RunStepRunner: runStepRunner,
	}

	stepsRunner := runtime.NewStepsRunner(
		initStepRunner,
		destroyPlanStepRunner,
		showStepRunner,
		policyCheckStepRunner,
		applyStepRunner,
		versionStepRunner,
		runStepRunner,
		envStepRunner,
	)

	dbUpdater := &events.DBUpdater{
		DB: boltdb,
	}

	checksOutputUpdater := &events.ChecksOutputUpdater{
		VCSClient:        vcsClient,
		MarkdownRenderer: markdownRenderer,
		TitleBuilder:     vcs.StatusTitleBuilder{TitlePrefix: userConfig.VCSStatusName},
		JobURLGenerator:  router,
	}

	session, err := aws.NewSession()
	if err != nil {
		return nil, errors.Wrap(err, "initializing new aws session")
	}

	var snsWriter sns.Writer

	if userConfig.LyftAuditJobsSnsTopicArn != "" {
		snsWriter = sns.NewWriterWithStats(
			session,
			userConfig.LyftAuditJobsSnsTopicArn,
			statsScope.SubScope("aws.sns.jobs"),
		)
	} else {
		snsWriter = sns.NewNoopWriter()
	}

	applyRequirementHandler := &events.AggregateApplyRequirements{
		WorkingDir: workingDir,
	}

	unwrappedPrjCmdRunner := events.NewProjectCommandRunner(
		stepsRunner,
		workingDir,
		webhooksManager,
		workingDirLocker,
		applyRequirementHandler,
	)

	statusUpdater := command.ProjectStatusUpdater{
		ProjectJobURLGenerator:  router,
		JobCloser:               projectCmdOutputHandler,
		ProjectVCSStatusUpdater: vcsStatusUpdater,
	}

	legacyPrjCmdRunner := wrappers.
		WrapProjectRunner(unwrappedPrjCmdRunner).
		WithSync(
			projectLocker,
			router,
		).
		WithAuditing(snsWriter).
		WithInstrumentation().
		WithJobs(
			statusUpdater,
		)

	unwrappedPRPrjCmdRunner := events.NewProjectCommandRunner(
		stepsRunner,
		workingDir,
		webhooksManager,
		workingDirLocker,
		applyRequirementHandler,
	)

	platformModePrjCmdRunner := wrappers.
		WrapProjectRunner(unwrappedPRPrjCmdRunner).
		WithAuditing(snsWriter).
		WithInstrumentation().
		WithJobs(
			statusUpdater,
		)

	prjCmdRunner := &lyftCommands.PlatformModeProjectRunner{
		PlatformModeRunner: platformModePrjCmdRunner,
		PrModeRunner:       legacyPrjCmdRunner,
		Allocator:          featureAllocator,
		Logger:             ctxLogger,
	}

	pullReqStatusFetcher := lyft_vcs.NewSQBasedPullStatusFetcher(
		githubClient,
		mergeabilityChecker,
	)

	policyCheckCommandRunner := events.NewPolicyCheckCommandRunner(
		dbUpdater,
		checksOutputUpdater,
		vcsStatusUpdater,
		prjCmdRunner,
		userConfig.ParallelPoolSize,
	)

	planCommandRunner := events.NewPlanCommandRunner(
		vcsClient,
		pendingPlanFinder,
		workingDir,
		vcsStatusUpdater,
		projectCommandBuilder,
		prjCmdRunner,
		dbUpdater,
		checksOutputUpdater,
		policyCheckCommandRunner,
		userConfig.ParallelPoolSize,
	)

	applyCommandRunner := events.NewApplyCommandRunner(
		vcsClient,
		userConfig.DisableApplyAll,
		applyLockingClient,
		vcsStatusUpdater,
		projectCommandBuilder,
		prjCmdRunner,
		checksOutputUpdater,
		dbUpdater,
		userConfig.ParallelPoolSize,
		pullReqStatusFetcher,
	)

	policyCheckOutputGenerator := policies.CommandOutputGenerator{
		PrjCommandRunner:  prjCmdRunner,
		PrjCommandBuilder: projectCommandBuilder,
	}

	approvePoliciesCommandRunner := events.NewApprovePoliciesCommandRunner(
		vcsStatusUpdater,
		projectCommandBuilder,
		prjCmdRunner,
		checksOutputUpdater,
		dbUpdater,
		&policyCheckOutputGenerator,
	)

	unlockCommandRunner := events.NewUnlockCommandRunner(
		deleteLockCommand,
		vcsClient,
	)

	pullOutputUpdater := &events.PullOutputUpdater{
		VCSClient:            vcsClient,
		MarkdownRenderer:     markdownRenderer,
		HidePrevPlanComments: userConfig.HidePrevPlanComments,
	}

	// Using pull updater for version commands until we move off of PR comments entirely
	versionCommandRunner := events.NewVersionCommandRunner(
		pullOutputUpdater,
		projectCommandBuilder,
		prjCmdRunner,
		userConfig.ParallelPoolSize,
	)

	commentCommandRunnerByCmd := map[command.Name]command.Runner{
		command.Plan:            planCommandRunner,
		command.Apply:           applyCommandRunner,
		command.ApprovePolicies: approvePoliciesCommandRunner,
		command.Unlock:          unlockCommandRunner,
		command.Version:         versionCommandRunner,
	}
	cmdStatsScope := statsScope.SubScope("cmd")
	staleCommandChecker := &events.StaleCommandHandler{
		StaleStatsScope: cmdStatsScope.SubScope("stale"),
	}
	commandRunner := &events.DefaultCommandRunner{
		VCSClient:                     vcsClient,
		CommentCommandRunnerByCmd:     commentCommandRunnerByCmd,
		GlobalCfg:                     globalCfg,
		StatsScope:                    cmdStatsScope,
		DisableAutoplan:               userConfig.DisableAutoplan,
		Drainer:                       drainer,
		PreWorkflowHooksCommandRunner: preWorkflowHooksCommandRunner,
		PullStatusFetcher:             boltdb,
		StaleCommandChecker:           staleCommandChecker,
		VCSStatusUpdater:              vcsStatusUpdater,
		Logger:                        ctxLogger,
	}

	forceApplyCommandRunner := &events.ForceApplyCommandRunner{
		CommandRunner: commandRunner,
		VCSClient:     vcsClient,
		Logger:        ctxLogger,
	}

	repoAllowlist, err := events.NewRepoAllowlistChecker(userConfig.RepoAllowlist)
	if err != nil {
		return nil, err
	}
	locksController := &controllers.LocksController{
		AtlantisVersion:    config.AtlantisVersion,
		AtlantisURL:        parsedURL,
		Locker:             lockingClient,
		ApplyLocker:        applyLockingClient,
		Logger:             ctxLogger,
		VCSClient:          vcsClient,
		LockDetailTemplate: templates.LockTemplate,
		WorkingDir:         workingDir,
		WorkingDirLocker:   workingDirLocker,
		DB:                 boltdb,
		DeleteLockCommand:  deleteLockCommand,
	}

	wsMux := websocket.NewInstrumentedMultiplexor(
		websocket.NewMultiplexor(
			ctxLogger,
			controllers.JobIDKeyGenerator{},
			projectCmdOutputHandler,
		),
		projectJobsScope,
	)

	jobsController := &controllers.JobsController{
		AtlantisVersion:          config.AtlantisVersion,
		AtlantisURL:              parsedURL,
		Logger:                   ctxLogger,
		ProjectJobsTemplate:      templates.ProjectJobsTemplate,
		ProjectJobsErrorTemplate: templates.ProjectJobsErrorTemplate,
		Db:                       boltdb,
		WsMux:                    wsMux,
		StatsScope:               projectJobsScope,
		KeyGenerator:             controllers.JobIDKeyGenerator{},
	}
	githubAppController := &controllers.GithubAppController{
		AtlantisURL:         parsedURL,
		Logger:              ctxLogger,
		GithubSetupComplete: githubAppEnabled,
		GithubHostname:      userConfig.GithubHostname,
		GithubOrg:           userConfig.GithubOrg,
		GithubStatusName:    userConfig.VCSStatusName,
	}

	scheduledExecutorService := scheduled.NewExecutorService(
		events.NewFileWorkDirIterator(
			githubClient,
			eventParser,
			userConfig.DataDir,
			ctxLogger,
		),
		statsScope,
		ctxLogger,
		&events.PullClosedExecutor{
			VCSClient:                vcsClient,
			Locker:                   lockingClient,
			WorkingDir:               workingDir,
			Logger:                   ctxLogger,
			DB:                       boltdb,
			LogStreamResourceCleaner: projectCmdOutputHandler,

			// using a specific template to signal that this is from an async process
			PullClosedTemplate: scheduled.NewGCStaleClosedPull(),
		},

		// using a pullclosed executor for stale open PRs. Naming is weird, we need to come up with something better.
		&events.PullClosedExecutor{
			VCSClient:                vcsClient,
			Locker:                   lockingClient,
			WorkingDir:               workingDir,
			Logger:                   ctxLogger,
			DB:                       boltdb,
			LogStreamResourceCleaner: projectCmdOutputHandler,

			// using a specific template to signal that this is from an async process
			PullClosedTemplate: scheduled.NewGCStaleOpenPull(),
		},

		rawGithubClient,
	)

	ctx, cancel := context.WithCancel(context.Background())

	repoConverter := github_converter.RepoConverter{
		GithubUser:  userConfig.GithubUser,
		GithubToken: userConfig.GithubToken,
	}

	pullConverter := github_converter.PullConverter{
		RepoConverter: repoConverter,
	}

	defaultEventsController := events_controllers.NewVCSEventsController(
		statsScope,
		[]byte(userConfig.GithubWebhookSecret),
		userConfig.PlanDrafts,
		forceApplyCommandRunner,
		commentParser,
		eventParser,
		pullClosedExecutor,
		repoAllowlist,
		vcsClient,
		ctxLogger,
		userConfig.DisableApply,
		[]byte(userConfig.GitlabWebhookSecret),
		supportedVCSHosts,
		[]byte(userConfig.BitbucketWebhookSecret),
		[]byte(userConfig.AzureDevopsWebhookUser),
		[]byte(userConfig.AzureDevopsWebhookPassword),
		repoConverter,
		pullConverter,
		githubClient,
		azuredevopsClient,
		gitlabClient,
	)

	var vcsPostHandler sqs.VCSPostHandler
	lyftMode := userConfig.ToLyftMode()
	switch lyftMode {
	case Default: // default eventsController handles POST
		vcsPostHandler = defaultEventsController
		ctxLogger.Info("running Atlantis in default mode")
	case Worker: // an SQS worker is set up to handle messages via default eventsController
		worker, err := sqs.NewGatewaySQSWorker(ctx, statsScope, ctxLogger, userConfig.LyftWorkerQueueURL, defaultEventsController)
		if err != nil {
			ctxLogger.Error("unable to set up worker", map[string]interface{}{
				"err": err,
			})
			cancel()
			return nil, errors.Wrapf(err, "setting up sqs handler for worker mode")
		}
		go worker.Work(ctx)
		ctxLogger.Info("running Atlantis in worker mode", map[string]interface{}{
			"queue": userConfig.LyftWorkerQueueURL,
		})
	}

	return &Server{
		AtlantisVersion:               config.AtlantisVersion,
		AtlantisURL:                   parsedURL,
		Router:                        underlyingRouter,
		Port:                          userConfig.Port,
		PreWorkflowHooksCommandRunner: preWorkflowHooksCommandRunner,
		CommandRunner:                 commandRunner,
		CtxLogger:                     ctxLogger,
		StatsScope:                    statsScope,
		StatsCloser:                   closer,
		Locker:                        lockingClient,
		ApplyLocker:                   applyLockingClient,
		VCSPostHandler:                vcsPostHandler,
		GithubAppController:           githubAppController,
		LocksController:               locksController,
		JobsController:                jobsController,
		StatusController:              statusController,
		IndexTemplate:                 templates.IndexTemplate,
		LockDetailTemplate:            templates.LockTemplate,
		ProjectJobsTemplate:           templates.ProjectJobsTemplate,
		ProjectJobsErrorTemplate:      templates.ProjectJobsErrorTemplate,
		SSLKeyFile:                    userConfig.SSLKeyFile,
		SSLCertFile:                   userConfig.SSLCertFile,
		Drainer:                       drainer,
		ScheduledExecutorService:      scheduledExecutorService,
		ProjectCmdOutputHandler:       projectCmdOutputHandler,
		LyftMode:                      lyftMode,
		CancelWorker:                  cancel,
	}, nil
}

// Start creates the routes and starts serving traffic.
func (s *Server) Start() error {
	s.Router.HandleFunc("/healthz", s.Healthz).Methods("GET")
	s.Router.HandleFunc("/status", s.StatusController.Get).Methods("GET")
	if s.LyftMode != Worker {
		s.Router.HandleFunc("/events", s.VCSPostHandler.Post).Methods("POST")
	}
	s.Router.HandleFunc("/", s.Index).Methods("GET").MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
		return r.URL.Path == "/" || r.URL.Path == "/index.html"
	})
	s.Router.PathPrefix("/static/").Handler(http.FileServer(&assetfs.AssetFS{Asset: static.Asset, AssetDir: static.AssetDir, AssetInfo: static.AssetInfo}))
	s.Router.HandleFunc("/apply/lock", s.LocksController.LockApply).Methods("POST").Queries()
	s.Router.HandleFunc("/apply/unlock", s.LocksController.UnlockApply).Methods("DELETE").Queries()
	s.Router.HandleFunc("/locks", s.LocksController.DeleteLock).Methods("DELETE").Queries("id", "{id:.*}")
	s.Router.HandleFunc("/lock", s.LocksController.GetLock).Methods("GET").
		Queries(LockViewRouteIDQueryParam, fmt.Sprintf("{%s}", LockViewRouteIDQueryParam)).Name(LockViewRouteName)
	s.Router.HandleFunc("/jobs/{job-id}", s.JobsController.GetProjectJobs).Methods("GET").Name(ProjectJobsViewRouteName)
	s.Router.HandleFunc("/jobs/{job-id}/ws", s.JobsController.GetProjectJobsWS).Methods("GET")
	s.Router.HandleFunc("/github-app/exchange-code", s.GithubAppController.ExchangeCode).Methods("GET")
	s.Router.HandleFunc("/github-app/setup", s.GithubAppController.New).Methods("GET")

	n := negroni.New(&negroni.Recovery{
		Logger:     log.New(os.Stdout, "", log.LstdFlags),
		PrintStack: false,
		StackAll:   false,
		StackSize:  1024 * 8,
	}, NewRequestLogger(s.CtxLogger))
	n.UseHandler(s.Router)

	defer s.CtxLogger.Close()

	// Ensure server gracefully drains connections when stopped.
	stop := make(chan os.Signal, 1)
	// Stop on SIGINTs and SIGTERMs.
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go s.ScheduledExecutorService.Run()

	go func() {
		s.ProjectCmdOutputHandler.Handle()
	}()

	server := &http.Server{Addr: fmt.Sprintf(":%d", s.Port), Handler: n}
	go func() {
		s.CtxLogger.Info(fmt.Sprintf("Atlantis started - listening on port %v", s.Port))

		var err error
		if s.SSLCertFile != "" && s.SSLKeyFile != "" {
			err = server.ListenAndServeTLS(s.SSLCertFile, s.SSLKeyFile)
		} else {
			err = server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			s.CtxLogger.Error(err.Error())
		}
	}()
	<-stop

	// Shutdown sqs polling. Any received messages being processed will either succeed/fail depending on if drainer started.
	if s.LyftMode == Worker {
		s.CtxLogger.Warn("Received interrupt. Shutting down the sqs handler")
		s.CancelWorker()
	}

	s.CtxLogger.Warn("Received interrupt. Waiting for in-progress operations to complete")
	s.waitForDrain()

	// flush stats before shutdown
	if err := s.StatsCloser.Close(); err != nil {
		s.CtxLogger.Error(err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // nolint: vet
	defer cancel()
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
			s.CtxLogger.Info("All in-progress operations complete, shutting down")
			return
		case <-ticker.C:
			s.CtxLogger.Info(fmt.Sprintf("Waiting for in-progress operations to complete, current in-progress ops: %d", s.Drainer.GetStatus().InProgressOps))
		}
	}
}

// Index is the / route.
func (s *Server) Index(w http.ResponseWriter, _ *http.Request) {
	locks, err := s.Locker.List()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, "Could not retrieve locks: %s", err)
		return
	}

	var lockResults []templates.LockIndexData
	for id, v := range locks {
		lockURL, _ := s.Router.Get(LockViewRouteName).URL("id", url.QueryEscape(id))
		lockResults = append(lockResults, templates.LockIndexData{
			// NOTE: must use .String() instead of .Path because we need the
			// query params as part of the lock URL.
			LockPath:      lockURL.String(),
			RepoFullName:  v.Project.RepoFullName,
			PullNum:       v.Pull.Num,
			Path:          v.Project.Path,
			Workspace:     v.Workspace,
			Time:          v.Time,
			TimeFormatted: v.Time.Format("02-01-2006 15:04:05"),
		})
	}

	applyCmdLock, err := s.ApplyLocker.CheckApplyLock()
	s.CtxLogger.Info(fmt.Sprintf("Apply Lock: %v", applyCmdLock))
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, "Could not retrieve global apply lock: %s", err)
		return
	}

	applyLockData := templates.ApplyLockData{
		Time:          applyCmdLock.Time,
		Locked:        applyCmdLock.Locked,
		TimeFormatted: applyCmdLock.Time.Format("02-01-2006 15:04:05"),
	}
	//Sort by date - newest to oldest.
	sort.SliceStable(lockResults, func(i, j int) bool { return lockResults[i].Time.After(lockResults[j].Time) })

	err = s.IndexTemplate.Execute(w, templates.IndexData{
		Locks:           lockResults,
		ApplyLock:       applyLockData,
		AtlantisVersion: s.AtlantisVersion,
		CleanedBasePath: s.AtlantisURL.Path,
	})
	if err != nil {
		s.CtxLogger.Error(err.Error())
	}
}

func mkSubDir(parentDir string, subDir string) (string, error) {
	fullDir := filepath.Join(parentDir, subDir)
	if err := os.MkdirAll(fullDir, 0700); err != nil {
		return "", errors.Wrapf(err, "unable to creare dir %q", fullDir)
	}

	return fullDir, nil
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
