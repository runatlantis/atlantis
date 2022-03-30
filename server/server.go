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

	assetfs "github.com/elazarl/go-bindata-assetfs"
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
	lyftDecorators "github.com/runatlantis/atlantis/server/lyft/decorators"
	"github.com/runatlantis/atlantis/server/lyft/feature"
	"github.com/runatlantis/atlantis/server/lyft/gateway"
	"github.com/runatlantis/atlantis/server/lyft/scheduled"
	"github.com/runatlantis/atlantis/server/metrics"
	"github.com/uber-go/tally"

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
	"github.com/runatlantis/atlantis/server/events/command/apply"
	"github.com/runatlantis/atlantis/server/events/command/plan"
	"github.com/runatlantis/atlantis/server/events/command/policies"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/events/vcs/bitbucketcloud"
	"github.com/runatlantis/atlantis/server/events/vcs/bitbucketserver"
	lyft_vcs "github.com/runatlantis/atlantis/server/events/vcs/lyft"
	"github.com/runatlantis/atlantis/server/events/webhooks"
	"github.com/runatlantis/atlantis/server/logging"
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
	PreWorkflowHooksCommandRunner *events.DefaultPreWorkflowHooksCommandRunner
	CommandRunner                 *events.DefaultCommandRunner
	Logger                        logging.SimpleLogging
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
	AllowForkPRsFlag        string
	AtlantisURLFlag         string
	AtlantisVersion         string
	DefaultTFVersionFlag    string
	RepoConfigJSONFlag      string
	SilenceForkPRErrorsFlag string
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
	logger, err := logging.NewStructuredLoggerFromLevel(userConfig.ToLogLevel())

	if err != nil {
		return nil, err
	}

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

	mergeabilityChecker := vcs.NewLyftPullMergeabilityChecker(userConfig.VCSStatusName)

	policyChecksEnabled := false
	if userConfig.EnablePolicyChecksFlag {
		logger.Infof("Policy Checks are enabled")
		policyChecksEnabled = true
	}

	validator := &cfgParser.ParserValidator{}

	globalCfg := valid.NewGlobalCfgFromArgs(
		valid.GlobalCfgArgs{
			AllowRepoCfg:        userConfig.AllowRepoConfig,
			MergeableReq:        userConfig.RequireMergeable,
			ApprovedReq:         userConfig.RequireApproval,
			UnDivergedReq:       userConfig.RequireUnDiverged,
			SQUnLockedReq:       userConfig.RequireSQUnlocked,
			PolicyCheckEnabled:  userConfig.EnablePolicyChecksFlag,
			PlatformModeEnabled: userConfig.EnablePlatformMode,
		})
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

	statsScope, closer, err := metrics.NewScope(globalCfg.Metrics, logger, userConfig.StatsNamespace)

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
		rawGithubClient, err = vcs.NewGithubClient(userConfig.GithubHostname, githubCredentials, logger, mergeabilityChecker)
		if err != nil {
			return nil, err
		}

		githubClient = vcs.NewInstrumentedGithubClient(rawGithubClient, statsScope, ctxLogger)
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
			if err := events.WriteGitCreds(userConfig.GithubUser, userConfig.GithubToken, userConfig.GithubHostname, home, logger, false); err != nil {
				return nil, err
			}
		}
		if userConfig.GitlabUser != "" {
			if err := events.WriteGitCreds(userConfig.GitlabUser, userConfig.GitlabToken, userConfig.GitlabHostname, home, logger, false); err != nil {
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
			if err := events.WriteGitCreds(userConfig.BitbucketUser, userConfig.BitbucketToken, bitbucketBaseURL, home, logger, false); err != nil {
				return nil, err
			}
		}
		if userConfig.AzureDevopsUser != "" {
			if err := events.WriteGitCreds(userConfig.AzureDevopsUser, userConfig.AzureDevopsToken, "dev.azure.com", home, logger, false); err != nil {
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
	commitStatusUpdater := &command.VCSStatusUpdater{Client: vcsClient, TitleBuilder: vcs.StatusTitleBuilder{TitlePrefix: userConfig.VCSStatusName}}

	binDir, err := mkSubDir(userConfig.DataDir, BinDirName)

	if err != nil {
		return nil, err
	}

	cacheDir, err := mkSubDir(userConfig.DataDir, TerraformPluginCacheDirName)

	if err != nil {
		return nil, err
	}

	featureAllocator, err := feature.NewGHSourcedAllocator(
		feature.RepoConfig{
			Owner:  userConfig.FFOwner,
			Repo:   userConfig.FFRepo,
			Branch: userConfig.FFBranch,
			Path:   userConfig.FFPath,
		}, githubClient, logger)

	if err != nil {
		return nil, errors.Wrap(err, "initializing feature allocator")
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

	storageBackend, err := jobs.NewStorageBackend(globalCfg.Jobs, logger, featureAllocator, projectJobsScope)
	if err != nil {
		return nil, errors.Wrapf(err, "initializing storage backend")
	}

	jobStore := jobs.NewJobStore(storageBackend)

	var projectCmdOutputHandler jobs.ProjectCommandOutputHandler
	// When TFE is enabled log streaming is not necessary.

	if userConfig.TFEToken != "" {
		projectCmdOutputHandler = &jobs.NoopProjectOutputHandler{}
	} else {
		projectCmdOutput := make(chan *jobs.ProjectCmdOutputLine)
		projectCmdOutputHandler = jobs.NewAsyncProjectCommandOutputHandler(
			projectCmdOutput,
			logger,
			jobStore,
		)
	}

	terraformClient, err := terraform.NewClient(
		logger,
		binDir,
		cacheDir,
		userConfig.TFEToken,
		userConfig.TFEHostname,
		userConfig.DefaultTFVersion,
		config.DefaultTFVersionFlag,
		userConfig.TFDownloadURL,
		&terraform.DefaultDownloader{},
		true,
		projectCmdOutputHandler,
		featureAllocator)

	if err != nil {
		return nil, errors.Wrap(err, "initializing terraform")
	}
	markdownRenderer := &events.MarkdownRenderer{
		GitlabSupportsCommonMark: gitlabClient.SupportsCommonMark(),
		DisableApplyAll:          userConfig.DisableApplyAll,
		DisableMarkdownFolding:   userConfig.DisableMarkdownFolding,
		DisableApply:             userConfig.DisableApply,
		DisableRepoLocking:       userConfig.DisableRepoLocking,
		EnableDiffMarkdownFormat: userConfig.EnableDiffMarkdownFormat,
	}

	boltdb, err := db.New(userConfig.DataDir)
	if err != nil {
		return nil, err
	}
	var lockingClient locking.Locker
	var applyLockingClient locking.ApplyLocker
	if userConfig.DisableRepoLocking {
		lockingClient = locking.NewNoOpLocker()
	} else {
		lockingClient = locking.NewClient(boltdb)
	}
	applyLockingClient = locking.NewApplyClient(boltdb, userConfig.DisableApply)
	workingDirLocker := events.NewDefaultWorkingDirLocker()

	var workingDir events.WorkingDir = &events.FileWorkspace{
		DataDir:       userConfig.DataDir,
		CheckoutMerge: userConfig.CheckoutStrategy == "merge",
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
		Logger:           logger,
		WorkingDir:       workingDir,
		WorkingDirLocker: workingDirLocker,
		DB:               boltdb,
	}

	pullClosedExecutor := events.NewInstrumentedPullClosedExecutor(
		statsScope,
		logger,
		&events.PullClosedExecutor{
			Locker:                   lockingClient,
			WorkingDir:               workingDir,
			Logger:                   logger,
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
		Logger:  logger,
		Drainer: drainer,
	}
	preWorkflowHooksCommandRunner := &events.DefaultPreWorkflowHooksCommandRunner{
		VCSClient:             vcsClient,
		GlobalCfg:             globalCfg,
		WorkingDirLocker:      workingDirLocker,
		WorkingDir:            workingDir,
		PreWorkflowHookRunner: runtime.DefaultPreWorkflowHookRunner{},
	}
	projectCommandBuilder := events.NewProjectCommandBuilderWithLimit(
		policyChecksEnabled,
		validator,
		&events.DefaultProjectFinder{},
		vcsClient,
		workingDir,
		workingDirLocker,
		globalCfg,
		pendingPlanFinder,
		commentParser,
		userConfig.SkipCloneNoChanges,
		userConfig.EnableRegExpCmd,
		userConfig.AutoplanFileList,
		statsScope,
		logger,
		userConfig.MaxProjectsPerPR,
	)
	applyRequirementHandler := &events.AggregateApplyRequirements{
		WorkingDir: workingDir,
	}

	initStepRunner := &runtime.InitStepRunner{
		TerraformExecutor: terraformClient,
		DefaultTFVersion:  defaultTfVersion,
	}

	planStepRunner := &runtime.PlanStepRunner{
		TerraformExecutor:   terraformClient,
		DefaultTFVersion:    defaultTfVersion,
		CommitStatusUpdater: commitStatusUpdater,
		AsyncTFExec:         terraformClient,
	}

	destroyPlanStepRunner := &lyftRuntime.DestroyPlanStepRunner{
		StepRunner: planStepRunner,
	}

	showStepRunner, err := runtime.NewShowStepRunner(terraformClient, defaultTfVersion)
	if err != nil {
		return nil, errors.Wrap(err, "initializing show step runner")
	}

	conftestExecutor := policy.NewConfTestExecutorWorkflow(logger, binDir, &terraform.DefaultDownloader{})
	policyCheckStepRunner, err := runtime.NewPolicyCheckStepRunner(
		defaultTfVersion,
		conftestExecutor,
	)
	if err != nil {
		return nil, errors.Wrap(err, "initializing policy check runner")
	}

	applyStepRunner := &runtime.ApplyStepRunner{
		TerraformExecutor:   terraformClient,
		CommitStatusUpdater: commitStatusUpdater,
		AsyncTFExec:         terraformClient,
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

	projectCommandRunner := events.NewProjectCommandRunner(
		stepsRunner,
		projectLocker,
		router,
		workingDir,
		webhooksManager,
		workingDirLocker,
		applyRequirementHandler,
	)

	dbUpdater := &events.DBUpdater{
		DB: boltdb,
	}

	pullUpdater := &events.PullUpdater{
		HidePrevPlanComments: userConfig.HidePrevPlanComments,
		VCSClient:            vcsClient,
		MarkdownRenderer:     markdownRenderer,
		GlobalCfg:            globalCfg,
	}

	autoMerger := &events.AutoMerger{
		VCSClient:       vcsClient,
		GlobalAutomerge: userConfig.Automerge,
	}

	projectOutputWrapper := &events.ProjectOutputWrapper{
		ProjectCommandRunner: projectCommandRunner,
		JobURLSetter:         jobs.NewJobURLSetter(router, commitStatusUpdater),
		JobCloser:            projectCmdOutputHandler,
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
	auditProjectCmdRunner := &lyftDecorators.AuditProjectCommandWrapper{
		SnsWriter:            snsWriter,
		ProjectCommandRunner: projectOutputWrapper,
	}

	instrumentedProjectCmdRunner := &events.InstrumentedProjectCommandRunner{
		ProjectCommandRunner: auditProjectCmdRunner,
	}

	policyCheckCommandRunner := events.NewPolicyCheckCommandRunner(
		dbUpdater,
		pullUpdater,
		commitStatusUpdater,
		instrumentedProjectCmdRunner,
		userConfig.ParallelPoolSize,
		userConfig.SilenceVCSStatusNoProjects,
	)

	planCommandRunner := events.NewPlanCommandRunner(
		userConfig.SilenceVCSStatusNoPlans,
		userConfig.SilenceVCSStatusNoProjects,
		vcsClient,
		pendingPlanFinder,
		workingDir,
		commitStatusUpdater,
		projectCommandBuilder,
		instrumentedProjectCmdRunner,
		dbUpdater,
		pullUpdater,
		policyCheckCommandRunner,
		autoMerger,
		userConfig.ParallelPoolSize,
		userConfig.SilenceNoProjects,
		boltdb,
	)

	pullReqStatusFetcher := lyft_vcs.NewSQBasedPullStatusFetcher(
		githubClient,
		mergeabilityChecker,
	)
	applyCommandRunner := events.NewApplyCommandRunner(
		vcsClient,
		userConfig.DisableApplyAll,
		applyLockingClient,
		commitStatusUpdater,
		projectCommandBuilder,
		instrumentedProjectCmdRunner,
		autoMerger,
		pullUpdater,
		dbUpdater,
		userConfig.ParallelPoolSize,
		userConfig.SilenceNoProjects,
		userConfig.SilenceVCSStatusNoProjects,
		pullReqStatusFetcher,
	)

	approvePoliciesCommandRunner := events.NewApprovePoliciesCommandRunner(
		commitStatusUpdater,
		projectCommandBuilder,
		instrumentedProjectCmdRunner,
		pullUpdater,
		dbUpdater,
		userConfig.SilenceNoProjects,
		userConfig.SilenceVCSStatusNoPlans,
	)

	unlockCommandRunner := events.NewUnlockCommandRunner(
		deleteLockCommand,
		vcsClient,
		userConfig.SilenceNoProjects,
	)

	versionCommandRunner := events.NewVersionCommandRunner(
		pullUpdater,
		projectCommandBuilder,
		projectCommandRunner,
		userConfig.ParallelPoolSize,
		userConfig.SilenceNoProjects,
	)

	featuredPlanRunner := lyftCommands.NewPlatformModeFeatureRunner(
		featureAllocator,
		userConfig.EnablePlatformMode,
		logger,
		plan.NewRunner(vcsClient),
		planCommandRunner,
	)

	featuredApplyRunner := lyftCommands.NewPlatformModeFeatureRunner(
		featureAllocator,
		userConfig.EnablePlatformMode,
		logger,
		apply.NewRunner(vcsClient),
		applyCommandRunner,
	)

	featuredPolicyCheckRunner := lyftCommands.NewPlatformModeFeatureRunner(
		featureAllocator,
		userConfig.EnablePlatformMode,
		logger,
		policies.NewRunner(vcsClient),
		approvePoliciesCommandRunner,
	)

	commentCommandRunnerByCmd := map[command.Name]command.Runner{
		command.Plan:            featuredPlanRunner,
		command.Apply:           featuredApplyRunner,
		command.ApprovePolicies: featuredPolicyCheckRunner,
		command.Unlock:          unlockCommandRunner,
		command.Version:         versionCommandRunner,
	}
	cmdStatsScope := statsScope.SubScope("cmd")
	staleCommandChecker := &events.StaleCommandHandler{
		StaleStatsScope: cmdStatsScope.SubScope("stale"),
	}
	commandRunner := &events.DefaultCommandRunner{
		VCSClient:                     vcsClient,
		GithubPullGetter:              githubClient,
		GitlabMergeRequestGetter:      gitlabClient,
		AzureDevopsPullGetter:         azuredevopsClient,
		CommentCommandRunnerByCmd:     commentCommandRunnerByCmd,
		EventParser:                   eventParser,
		GlobalCfg:                     globalCfg,
		StatsScope:                    cmdStatsScope,
		AllowForkPRs:                  userConfig.AllowForkPRs,
		AllowForkPRsFlag:              config.AllowForkPRsFlag,
		SilenceForkPRErrors:           userConfig.SilenceForkPRErrors,
		SilenceForkPRErrorsFlag:       config.SilenceForkPRErrorsFlag,
		DisableAutoplan:               userConfig.DisableAutoplan,
		Drainer:                       drainer,
		PreWorkflowHooksCommandRunner: preWorkflowHooksCommandRunner,
		PullStatusFetcher:             boltdb,
		StaleCommandChecker:           staleCommandChecker,
	}

	forceApplyCommandRunner := &events.ForceApplyCommandRunner{
		CommandRunner: commandRunner,
		VCSClient:     vcsClient,
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
		Logger:             logger,
		VCSClient:          vcsClient,
		LockDetailTemplate: templates.LockTemplate,
		WorkingDir:         workingDir,
		WorkingDirLocker:   workingDirLocker,
		DB:                 boltdb,
		DeleteLockCommand:  deleteLockCommand,
	}

	wsMux := websocket.NewInstrumentedMultiplexor(
		websocket.NewMultiplexor(
			logger,
			controllers.JobIDKeyGenerator{},
			projectCmdOutputHandler,
		),
		projectJobsScope,
	)

	jobsController := &controllers.JobsController{
		AtlantisVersion:          config.AtlantisVersion,
		AtlantisURL:              parsedURL,
		Logger:                   logger,
		ProjectJobsTemplate:      templates.ProjectJobsTemplate,
		ProjectJobsErrorTemplate: templates.ProjectJobsErrorTemplate,
		Db:                       boltdb,
		WsMux:                    wsMux,
		StatsScope:               projectJobsScope,
		KeyGenerator:             controllers.JobIDKeyGenerator{},
	}
	githubAppController := &controllers.GithubAppController{
		AtlantisURL:         parsedURL,
		Logger:              logger,
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
			logger,
		),
		statsScope,
		logger,
		&events.PullClosedExecutor{
			VCSClient:                vcsClient,
			Locker:                   lockingClient,
			WorkingDir:               workingDir,
			Logger:                   logger,
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
			Logger:                   logger,
			DB:                       boltdb,
			LogStreamResourceCleaner: projectCmdOutputHandler,

			// using a specific template to signal that this is from an async process
			PullClosedTemplate: scheduled.NewGCStaleOpenPull(),
		},

		rawGithubClient,
	)

	ctx, cancel := context.WithCancel(context.Background())
	gatewaySnsWriter := sns.NewWriterWithStats(session, userConfig.LyftGatewaySnsTopicArn, statsScope.SubScope("aws.sns.gateway"))
	autoplanValidator := &gateway.AutoplanValidator{
		Scope:                         statsScope.SubScope("validator"),
		VCSClient:                     vcsClient,
		PreWorkflowHooksCommandRunner: preWorkflowHooksCommandRunner,
		Drainer:                       drainer,
		GlobalCfg:                     globalCfg,
		AllowForkPRs:                  userConfig.AllowForkPRs,
		AllowForkPRsFlag:              config.AllowForkPRsFlag,
		SilenceForkPRErrors:           userConfig.SilenceForkPRErrors,
		SilenceForkPRErrorsFlag:       config.SilenceForkPRErrorsFlag,
		SilenceVCSStatusNoPlans:       userConfig.SilenceVCSStatusNoPlans,
		SilenceVCSStatusNoProjects:    userConfig.SilenceVCSStatusNoProjects,
		CommitStatusUpdater:           commitStatusUpdater,
		PrjCmdBuilder:                 projectCommandBuilder,
		PullUpdater:                   pullUpdater,
		WorkingDir:                    workingDir,
		WorkingDirLocker:              workingDirLocker,
	}
	gatewayEventsController := &gateway.VCSEventsController{
		Logger:                 logger,
		Scope:                  statsScope,
		Parser:                 eventParser,
		CommentParser:          commentParser,
		GithubWebhookSecret:    []byte(userConfig.GithubWebhookSecret),
		GithubRequestValidator: &events_controllers.DefaultGithubRequestValidator{},
		RepoAllowlistChecker:   repoAllowlist,
		SilenceAllowlistErrors: userConfig.SilenceAllowlistErrors,
		VCSClient:              vcsClient,
		SNSWriter:              gatewaySnsWriter,
		AutoplanValidator:      autoplanValidator,
	}
	defaultEventsController := &events_controllers.VCSEventsController{
		CommandRunner:                   forceApplyCommandRunner,
		PullCleaner:                     pullClosedExecutor,
		Parser:                          eventParser,
		CommentParser:                   commentParser,
		Logger:                          logger,
		Scope:                           statsScope,
		ApplyDisabled:                   userConfig.DisableApply,
		GithubWebhookSecret:             []byte(userConfig.GithubWebhookSecret),
		GithubRequestValidator:          &events_controllers.DefaultGithubRequestValidator{},
		GitlabRequestParserValidator:    &events_controllers.DefaultGitlabRequestParserValidator{},
		GitlabWebhookSecret:             []byte(userConfig.GitlabWebhookSecret),
		RepoAllowlistChecker:            repoAllowlist,
		SilenceAllowlistErrors:          userConfig.SilenceAllowlistErrors,
		SupportedVCSHosts:               supportedVCSHosts,
		VCSClient:                       vcsClient,
		BitbucketWebhookSecret:          []byte(userConfig.BitbucketWebhookSecret),
		AzureDevopsWebhookBasicUser:     []byte(userConfig.AzureDevopsWebhookUser),
		AzureDevopsWebhookBasicPassword: []byte(userConfig.AzureDevopsWebhookPassword),
		AzureDevopsRequestValidator:     &events_controllers.DefaultAzureDevopsRequestValidator{},
	}
	var vcsPostHandler sqs.VCSPostHandler
	lyftMode := userConfig.ToLyftMode()
	//TODO: remove logs after testing is complete
	switch lyftMode {
	case Default: // default eventsController handles POST
		vcsPostHandler = defaultEventsController
		logger.Infof("running Atlantis in default mode")
	case Gateway: // gateway eventsController handles POST
		vcsPostHandler = gatewayEventsController
		logger.With("sns", userConfig.LyftGatewaySnsTopicArn).Infof("running Atlantis in gateway mode")
	case Hybrid: // gateway eventsController handles POST, and SQS worker is set up to handle messages via default eventsController
		vcsPostHandler = gatewayEventsController
		worker, err := sqs.NewGatewaySQSWorker(ctx, statsScope, logger, userConfig.LyftWorkerQueueURL, defaultEventsController)
		if err != nil {
			logger.With("err", err).Errorf("unable to set up worker")
			cancel()
			return nil, errors.Wrapf(err, "setting up sqs handler for hybrid mode")
		}
		go worker.Work(ctx)
		logger.With("queue", userConfig.LyftWorkerQueueURL, "sns", userConfig.LyftGatewaySnsTopicArn).Infof("running Atlantis in hybrid mode")
	case Worker: // an SQS worker is set up to handle messages via default eventsController
		worker, err := sqs.NewGatewaySQSWorker(ctx, statsScope, logger, userConfig.LyftWorkerQueueURL, defaultEventsController)
		if err != nil {
			logger.With("err", err).Errorf("unable to set up worker")
			cancel()
			return nil, errors.Wrapf(err, "setting up sqs handler for worker mode")
		}
		go worker.Work(ctx)
		logger.With("queue", userConfig.LyftWorkerQueueURL).Infof("running Atlantis in worker mode")
	}

	return &Server{
		AtlantisVersion:               config.AtlantisVersion,
		AtlantisURL:                   parsedURL,
		Router:                        underlyingRouter,
		Port:                          userConfig.Port,
		PreWorkflowHooksCommandRunner: preWorkflowHooksCommandRunner,
		CommandRunner:                 commandRunner,
		Logger:                        logger,
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
	if s.LyftMode != Gateway {
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
	}
	s.Router.HandleFunc("/github-app/exchange-code", s.GithubAppController.ExchangeCode).Methods("GET")
	s.Router.HandleFunc("/github-app/setup", s.GithubAppController.New).Methods("GET")

	n := negroni.New(&negroni.Recovery{
		Logger:     log.New(os.Stdout, "", log.LstdFlags),
		PrintStack: false,
		StackAll:   false,
		StackSize:  1024 * 8,
	}, NewRequestLogger(s.Logger))
	n.UseHandler(s.Router)

	defer s.Logger.Close()
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
		s.Logger.Infof("Atlantis started - listening on port %v", s.Port)

		var err error
		if s.SSLCertFile != "" && s.SSLKeyFile != "" {
			err = server.ListenAndServeTLS(s.SSLCertFile, s.SSLKeyFile)
		} else {
			err = server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			s.Logger.Errorf(err.Error())
		}
	}()
	<-stop

	// Shutdown sqs polling. Any received messages being processed will either succeed/fail depending on if drainer started.
	if s.LyftMode == Hybrid || s.LyftMode == Worker {
		s.Logger.Warnf("Received interrupt. Shutting down the sqs handler")
		s.CancelWorker()
	}

	s.Logger.Warnf("Received interrupt. Waiting for in-progress operations to complete")
	s.waitForDrain()

	// flush stats before shutdown
	if err := s.StatsCloser.Close(); err != nil {
		s.Logger.Errorf(err.Error())
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
			s.Logger.Infof("All in-progress operations complete, shutting down")
			return
		case <-ticker.C:
			s.Logger.Infof("Waiting for in-progress operations to complete, current in-progress ops: %d", s.Drainer.GetStatus().InProgressOps)
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
	s.Logger.Infof("Apply Lock: %v", applyCmdLock)
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
		s.Logger.Errorf(err.Error())
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
