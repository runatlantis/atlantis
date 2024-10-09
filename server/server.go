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
	"crypto/tls"
	"embed"
	"flag"
	"fmt"
	"io"
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

	"github.com/mitchellh/go-homedir"
	tally "github.com/uber-go/tally/v4"
	prometheus "github.com/uber-go/tally/v4/prometheus"
	"github.com/urfave/negroni/v3"

	cfg "github.com/runatlantis/atlantis/server/core/config"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/core/redis"
	"github.com/runatlantis/atlantis/server/jobs"
	"github.com/runatlantis/atlantis/server/metrics"
	"github.com/runatlantis/atlantis/server/scheduled"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/controllers"
	events_controllers "github.com/runatlantis/atlantis/server/controllers/events"
	"github.com/runatlantis/atlantis/server/controllers/web_templates"
	"github.com/runatlantis/atlantis/server/controllers/websocket"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/core/runtime"
	"github.com/runatlantis/atlantis/server/core/runtime/policy"
	"github.com/runatlantis/atlantis/server/core/terraform"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/events/vcs/bitbucketcloud"
	"github.com/runatlantis/atlantis/server/events/vcs/bitbucketserver"
	"github.com/runatlantis/atlantis/server/events/vcs/gitea"
	"github.com/runatlantis/atlantis/server/events/webhooks"
	"github.com/runatlantis/atlantis/server/logging"
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
	AtlantisVersion                string
	AtlantisURL                    *url.URL
	Router                         *mux.Router
	Port                           int
	PostWorkflowHooksCommandRunner *events.DefaultPostWorkflowHooksCommandRunner
	PreWorkflowHooksCommandRunner  *events.DefaultPreWorkflowHooksCommandRunner
	CommandRunner                  *events.DefaultCommandRunner
	Logger                         logging.SimpleLogging
	StatsScope                     tally.Scope
	StatsReporter                  tally.BaseStatsReporter
	StatsCloser                    io.Closer
	Locker                         locking.Locker
	ApplyLocker                    locking.ApplyLocker
	VCSEventsController            *events_controllers.VCSEventsController
	GithubAppController            *controllers.GithubAppController
	LocksController                *controllers.LocksController
	StatusController               *controllers.StatusController
	JobsController                 *controllers.JobsController
	APIController                  *controllers.APIController
	IndexTemplate                  web_templates.TemplateWriter
	LockDetailTemplate             web_templates.TemplateWriter
	ProjectJobsTemplate            web_templates.TemplateWriter
	ProjectJobsErrorTemplate       web_templates.TemplateWriter
	SSLCertFile                    string
	SSLKeyFile                     string
	CertLastRefreshTime            time.Time
	KeyLastRefreshTime             time.Time
	SSLCert                        *tls.Certificate
	Drainer                        *events.Drainer
	WebAuthentication              bool
	WebUsername                    string
	WebPassword                    string
	ProjectCmdOutputHandler        jobs.ProjectCommandOutputHandler
	ScheduledExecutorService       *scheduled.ExecutorService
	DisableGlobalApplyLock         bool
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
	// BranchRegex is a regex that is used to match against the base branch
	// that is being modified for this event. If the regex matches, we'll
	// send the webhook, ex. "main.*".
	BranchRegex string `mapstructure:"branch-regex"`
	// Kind is the type of webhook we should send, ex. slack.
	Kind string `mapstructure:"kind"`
	// Channel is the channel to send this webhook to. It only applies to
	// slack webhooks. Should be without '#'.
	Channel string `mapstructure:"channel"`
}

//go:embed static
var staticAssets embed.FS

// NewServer returns a new server. If there are issues starting the server or
// its dependencies an error will be returned. This is like the main() function
// for the server CLI command because it injects all the dependencies.
func NewServer(userConfig UserConfig, config Config) (*Server, error) {
	logging.SuppressDefaultLogging()
	logger, err := logging.NewStructuredLoggerFromLevel(userConfig.ToLogLevel())

	if err != nil {
		return nil, err
	}

	var supportedVCSHosts []models.VCSHostType
	var githubClient vcs.IGithubClient
	var githubAppEnabled bool
	var githubConfig vcs.GithubConfig
	var githubCredentials vcs.GithubCredentials
	var gitlabClient *vcs.GitlabClient
	var bitbucketCloudClient *bitbucketcloud.Client
	var bitbucketServerClient *bitbucketserver.Client
	var azuredevopsClient *vcs.AzureDevopsClient
	var giteaClient *gitea.GiteaClient

	policyChecksEnabled := false
	if userConfig.EnablePolicyChecksFlag {
		logger.Info("Policy Checks are enabled")
		policyChecksEnabled = true
	}

	allowCommands, err := userConfig.ToAllowCommandNames()
	if err != nil {
		return nil, err
	}
	disableApply := true
	for _, allowCommand := range allowCommands {
		if allowCommand == command.Apply {
			disableApply = false
			break
		}
	}

	validator := &cfg.ParserValidator{}

	globalCfg := valid.NewGlobalCfgFromArgs(
		valid.GlobalCfgArgs{
			PolicyCheckEnabled: userConfig.EnablePolicyChecksFlag,
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

	statsScope, statsReporter, closer, err := metrics.NewScope(globalCfg.Metrics, logger, userConfig.StatsNamespace)

	if err != nil {
		return nil, errors.Wrapf(err, "instantiating metrics scope")
	}

	if userConfig.GithubUser != "" || userConfig.GithubAppID != 0 {
		if userConfig.GithubAllowMergeableBypassApply {
			githubConfig = vcs.GithubConfig{
				AllowMergeableBypassApply: true,
			}
		}
		supportedVCSHosts = append(supportedVCSHosts, models.Github)
		if userConfig.GithubUser != "" {
			githubCredentials = &vcs.GithubUserCredentials{
				User:  userConfig.GithubUser,
				Token: userConfig.GithubToken,
			}
		} else if userConfig.GithubAppID != 0 && userConfig.GithubAppKeyFile != "" {
			privateKey, err := os.ReadFile(userConfig.GithubAppKeyFile)
			if err != nil {
				return nil, err
			}
			githubCredentials = &vcs.GithubAppCredentials{
				AppID:          userConfig.GithubAppID,
				InstallationID: userConfig.GithubAppInstallationID,
				Key:            privateKey,
				Hostname:       userConfig.GithubHostname,
				AppSlug:        userConfig.GithubAppSlug,
			}
			githubAppEnabled = true
		} else if userConfig.GithubAppID != 0 && userConfig.GithubAppKey != "" {
			githubCredentials = &vcs.GithubAppCredentials{
				AppID:          userConfig.GithubAppID,
				InstallationID: userConfig.GithubAppInstallationID,
				Key:            []byte(userConfig.GithubAppKey),
				Hostname:       userConfig.GithubHostname,
				AppSlug:        userConfig.GithubAppSlug,
			}
			githubAppEnabled = true
		}

		var err error
		rawGithubClient, err := vcs.NewGithubClient(userConfig.GithubHostname, githubCredentials, githubConfig, userConfig.MaxCommentsPerCommand, logger)
		if err != nil {
			return nil, err
		}

		githubClient = vcs.NewInstrumentedGithubClient(rawGithubClient, statsScope, logger)
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
		azuredevopsClient, err = vcs.NewAzureDevopsClient(userConfig.AzureDevOpsHostname, userConfig.AzureDevopsUser, userConfig.AzureDevopsToken)
		if err != nil {
			return nil, err
		}
	}
	if userConfig.GiteaToken != "" {
		supportedVCSHosts = append(supportedVCSHosts, models.Gitea)

		giteaClient, err = gitea.NewClient(userConfig.GiteaBaseURL, userConfig.GiteaUser, userConfig.GiteaToken, userConfig.GiteaPageSize, logger)
		if err != nil {
			fmt.Println("error setting up gitea client", "error", err)
			return nil, errors.Wrapf(err, "setting up Gitea client")
		} else {
			logger.Info("gitea client configured successfully")
		}
	}

	logger.Info("Supported VCS Hosts", "hosts", supportedVCSHosts)

	home, err := homedir.Dir()
	if err != nil {
		return nil, errors.Wrap(err, "getting home dir to write ~/.git-credentials file")
	}

	if userConfig.WriteGitCreds {
		if userConfig.GithubUser != "" {
			if err := vcs.WriteGitCreds(userConfig.GithubUser, userConfig.GithubToken, userConfig.GithubHostname, home, logger, false); err != nil {
				return nil, err
			}
		}
		if userConfig.GitlabUser != "" {
			if err := vcs.WriteGitCreds(userConfig.GitlabUser, userConfig.GitlabToken, userConfig.GitlabHostname, home, logger, false); err != nil {
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
			if err := vcs.WriteGitCreds(userConfig.BitbucketUser, userConfig.BitbucketToken, bitbucketBaseURL, home, logger, false); err != nil {
				return nil, err
			}
		}
		if userConfig.AzureDevopsUser != "" {
			if err := vcs.WriteGitCreds(userConfig.AzureDevopsUser, userConfig.AzureDevopsToken, "dev.azure.com", home, logger, false); err != nil {
				return nil, err
			}
		}
		if userConfig.GiteaUser != "" {
			if err := vcs.WriteGitCreds(userConfig.GiteaUser, userConfig.GiteaToken, userConfig.GiteaBaseURL, home, logger, false); err != nil {
				return nil, err
			}
		}
	}

	// default the project files used to generate the module index to the autoplan-file-list if autoplan-modules is true
	// but no files are specified
	if userConfig.AutoplanModules && userConfig.AutoplanModulesFromProjects == "" {
		userConfig.AutoplanModulesFromProjects = userConfig.AutoplanFileList
	}

	var webhooksConfig []webhooks.Config
	for _, c := range userConfig.Webhooks {
		config := webhooks.Config{
			Channel:        c.Channel,
			BranchRegex:    c.BranchRegex,
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
	vcsClient := vcs.NewClientProxy(githubClient, gitlabClient, bitbucketCloudClient, bitbucketServerClient, azuredevopsClient, giteaClient)
	commitStatusUpdater := &events.DefaultCommitStatusUpdater{Client: vcsClient, StatusName: userConfig.VCSStatusName}

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

	var projectCmdOutputHandler jobs.ProjectCommandOutputHandler

	if userConfig.TFEToken != "" && !userConfig.TFELocalExecutionMode {
		// When TFE is enabled and using remote execution mode log streaming is not necessary.
		projectCmdOutputHandler = &jobs.NoopProjectOutputHandler{}
	} else {
		projectCmdOutput := make(chan *jobs.ProjectCmdOutputLine)
		projectCmdOutputHandler = jobs.NewAsyncProjectCommandOutputHandler(
			projectCmdOutput,
			logger,
		)
	}

	distribution := terraform.NewDistributionTerraform()
	if userConfig.TFDistribution == "opentofu" {
		distribution = terraform.NewDistributionOpenTofu()
	}

	terraformClient, err := terraform.NewClient(
		logger,
		distribution,
		binDir,
		cacheDir,
		userConfig.TFEToken,
		userConfig.TFEHostname,
		userConfig.DefaultTFVersion,
		config.DefaultTFVersionFlag,
		userConfig.TFDownloadURL,
		userConfig.TFDownload,
		userConfig.UseTFPluginCache,
		projectCmdOutputHandler)
	// The flag.Lookup call is to detect if we're running in a unit test. If we
	// are, then we don't error out because we don't have/want terraform
	// installed on our CI system where the unit tests run.
	if err != nil && flag.Lookup("test.v") == nil {
		return nil, errors.Wrap(err, fmt.Sprintf("initializing %s", userConfig.TFDistribution))
	}
	markdownRenderer := events.NewMarkdownRenderer(
		gitlabClient.SupportsCommonMark(),
		userConfig.DisableApplyAll,
		disableApply,
		userConfig.DisableMarkdownFolding,
		userConfig.DisableRepoLocking,
		userConfig.EnableDiffMarkdownFormat,
		userConfig.MarkdownTemplateOverridesDir,
		userConfig.ExecutableName,
		userConfig.HideUnchangedPlanComments,
	)

	var lockingClient locking.Locker
	var applyLockingClient locking.ApplyLocker
	var backend locking.Backend

	switch dbtype := userConfig.LockingDBType; dbtype {
	case "redis":
		logger.Info("Utilizing Redis DB")
		backend, err = redis.New(userConfig.RedisHost, userConfig.RedisPort, userConfig.RedisPassword, userConfig.RedisTLSEnabled, userConfig.RedisInsecureSkipVerify, userConfig.RedisDB)
		if err != nil {
			return nil, err
		}
	case "boltdb":
		logger.Info("Utilizing BoltDB")
		backend, err = db.New(userConfig.DataDir)
		if err != nil {
			return nil, err
		}
	}

	noOpLocker := locking.NewNoOpLocker()
	if userConfig.DisableRepoLocking {
		logger.Info("Repo Locking is disabled")
		lockingClient = noOpLocker
	} else {
		lockingClient = locking.NewClient(backend)
	}
	disableGlobalApplyLock := false
	if userConfig.DisableGlobalApplyLock {
		disableGlobalApplyLock = true
	}

	applyLockingClient = locking.NewApplyClient(backend, disableApply, disableGlobalApplyLock)
	workingDirLocker := events.NewDefaultWorkingDirLocker()

	var workingDir events.WorkingDir = &events.FileWorkspace{
		DataDir:          userConfig.DataDir,
		CheckoutMerge:    userConfig.CheckoutStrategy == "merge",
		CheckoutDepth:    userConfig.CheckoutDepth,
		GithubAppEnabled: githubAppEnabled,
	}

	scheduledExecutorService := scheduled.NewExecutorService(
		statsScope,
		logger,
	)

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

		githubAppTokenRotator := vcs.NewGithubAppTokenRotator(logger, githubCredentials, userConfig.GithubHostname, home)
		tokenJd, err := githubAppTokenRotator.GenerateJob()
		if err != nil {
			return nil, errors.Wrap(err, "could not write credentials")
		}
		scheduledExecutorService.AddJob(tokenJd)
	}

	projectLocker := &events.DefaultProjectLocker{
		Locker:     lockingClient,
		NoOpLocker: noOpLocker,
		VCSClient:  vcsClient,
	}
	deleteLockCommand := &events.DefaultDeleteLockCommand{
		Locker:           lockingClient,
		WorkingDir:       workingDir,
		WorkingDirLocker: workingDirLocker,
		Backend:          backend,
	}

	pullClosedExecutor := events.NewInstrumentedPullClosedExecutor(
		statsScope,
		logger,
		&events.PullClosedExecutor{
			Locker:                   lockingClient,
			WorkingDir:               workingDir,
			Backend:                  backend,
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
		GiteaUser:          userConfig.GiteaUser,
		GiteaToken:         userConfig.GiteaToken,
		AllowDraftPRs:      userConfig.PlanDrafts,
		BitbucketUser:      userConfig.BitbucketUser,
		BitbucketToken:     userConfig.BitbucketToken,
		BitbucketServerURL: userConfig.BitbucketBaseURL,
		AzureDevopsUser:    userConfig.AzureDevopsUser,
		AzureDevopsToken:   userConfig.AzureDevopsToken,
	}
	commentParser := events.NewCommentParser(
		userConfig.GithubUser,
		userConfig.GitlabUser,
		userConfig.GiteaUser,
		userConfig.BitbucketUser,
		userConfig.AzureDevopsUser,
		userConfig.ExecutableName,
		allowCommands,
	)
	defaultTfVersion := terraformClient.DefaultVersion()
	pendingPlanFinder := &events.DefaultPendingPlanFinder{}
	runStepRunner := &runtime.RunStepRunner{
		TerraformExecutor:       terraformClient,
		DefaultTFVersion:        defaultTfVersion,
		TerraformBinDir:         terraformClient.TerraformBinDir(),
		ProjectCmdOutputHandler: projectCmdOutputHandler,
	}
	drainer := &events.Drainer{}
	statusController := &controllers.StatusController{
		Logger:          logger,
		Drainer:         drainer,
		AtlantisVersion: config.AtlantisVersion,
	}
	preWorkflowHooksCommandRunner := &events.DefaultPreWorkflowHooksCommandRunner{
		VCSClient:        vcsClient,
		GlobalCfg:        globalCfg,
		WorkingDirLocker: workingDirLocker,
		WorkingDir:       workingDir,
		PreWorkflowHookRunner: runtime.DefaultPreWorkflowHookRunner{
			OutputHandler: projectCmdOutputHandler,
		},
		CommitStatusUpdater: commitStatusUpdater,
		Router:              router,
	}
	postWorkflowHooksCommandRunner := &events.DefaultPostWorkflowHooksCommandRunner{
		VCSClient:        vcsClient,
		GlobalCfg:        globalCfg,
		WorkingDirLocker: workingDirLocker,
		WorkingDir:       workingDir,
		PostWorkflowHookRunner: runtime.DefaultPostWorkflowHookRunner{
			OutputHandler: projectCmdOutputHandler,
		},
		CommitStatusUpdater: commitStatusUpdater,
		Router:              router,
	}
	projectCommandBuilder := events.NewInstrumentedProjectCommandBuilder(
		logger,
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
		userConfig.Automerge,
		userConfig.ParallelPlan,
		userConfig.ParallelApply,
		userConfig.AutoplanModulesFromProjects,
		userConfig.AutoplanFileList,
		userConfig.RestrictFileList,
		userConfig.SilenceNoProjects,
		userConfig.IncludeGitUntrackedFiles,
		userConfig.AutoDiscoverModeFlag,
		statsScope,
		terraformClient,
	)

	showStepRunner, err := runtime.NewShowStepRunner(terraformClient, defaultTfVersion)

	if err != nil {
		return nil, errors.Wrap(err, "initializing show step runner")
	}

	policyCheckStepRunner, err := runtime.NewPolicyCheckStepRunner(
		defaultTfVersion,
		policy.NewConfTestExecutorWorkflow(logger, binDir, &policy.ConfTestGoGetterVersionDownloader{}),
	)

	if err != nil {
		return nil, errors.Wrap(err, "initializing policy check step runner")
	}

	applyRequirementHandler := &events.DefaultCommandRequirementHandler{
		WorkingDir: workingDir,
	}

	projectCommandRunner := &events.DefaultProjectCommandRunner{
		VcsClient:        vcsClient,
		Locker:           projectLocker,
		LockURLGenerator: router,
		InitStepRunner: &runtime.InitStepRunner{
			TerraformExecutor: terraformClient,
			DefaultTFVersion:  defaultTfVersion,
		},
		PlanStepRunner:        runtime.NewPlanStepRunner(terraformClient, defaultTfVersion, commitStatusUpdater, terraformClient),
		ShowStepRunner:        showStepRunner,
		PolicyCheckStepRunner: policyCheckStepRunner,
		ApplyStepRunner: &runtime.ApplyStepRunner{
			TerraformExecutor:   terraformClient,
			DefaultTFVersion:    defaultTfVersion,
			CommitStatusUpdater: commitStatusUpdater,
			AsyncTFExec:         terraformClient,
		},
		RunStepRunner: runStepRunner,
		EnvStepRunner: &runtime.EnvStepRunner{
			RunStepRunner: runStepRunner,
		},
		MultiEnvStepRunner: &runtime.MultiEnvStepRunner{
			RunStepRunner: runStepRunner,
		},
		VersionStepRunner: &runtime.VersionStepRunner{
			TerraformExecutor: terraformClient,
			DefaultTFVersion:  defaultTfVersion,
		},
		ImportStepRunner:          runtime.NewImportStepRunner(terraformClient, defaultTfVersion),
		StateRmStepRunner:         runtime.NewStateRmStepRunner(terraformClient, defaultTfVersion),
		WorkingDir:                workingDir,
		Webhooks:                  webhooksManager,
		WorkingDirLocker:          workingDirLocker,
		CommandRequirementHandler: applyRequirementHandler,
	}

	dbUpdater := &events.DBUpdater{
		Backend: backend,
	}

	pullUpdater := &events.PullUpdater{
		HidePrevPlanComments: userConfig.HidePrevPlanComments,
		VCSClient:            vcsClient,
		MarkdownRenderer:     markdownRenderer,
	}

	autoMerger := &events.AutoMerger{
		VCSClient:       vcsClient,
		GlobalAutomerge: userConfig.Automerge,
	}

	projectOutputWrapper := &events.ProjectOutputWrapper{
		JobMessageSender:     projectCmdOutputHandler,
		ProjectCommandRunner: projectCommandRunner,
		JobURLSetter:         jobs.NewJobURLSetter(router, commitStatusUpdater),
	}
	instrumentedProjectCmdRunner := events.NewInstrumentedProjectCommandRunner(
		statsScope,
		projectOutputWrapper,
	)

	policyCheckCommandRunner := events.NewPolicyCheckCommandRunner(
		dbUpdater,
		pullUpdater,
		commitStatusUpdater,
		instrumentedProjectCmdRunner,
		userConfig.ParallelPoolSize,
		userConfig.SilenceVCSStatusNoProjects,
		userConfig.QuietPolicyChecks,
	)

	pullReqStatusFetcher := vcs.NewPullReqStatusFetcher(vcsClient, userConfig.VCSStatusName)
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
		backend,
		lockingClient,
		userConfig.DiscardApprovalOnPlanFlag,
		pullReqStatusFetcher,
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
		backend,
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
		vcsClient,
	)

	unlockCommandRunner := events.NewUnlockCommandRunner(
		deleteLockCommand,
		vcsClient,
		userConfig.SilenceNoProjects,
		userConfig.DisableUnlockLabel,
	)

	versionCommandRunner := events.NewVersionCommandRunner(
		pullUpdater,
		projectCommandBuilder,
		projectOutputWrapper,
		userConfig.ParallelPoolSize,
		userConfig.SilenceNoProjects,
	)

	importCommandRunner := events.NewImportCommandRunner(
		pullUpdater,
		pullReqStatusFetcher,
		projectCommandBuilder,
		instrumentedProjectCmdRunner,
		userConfig.SilenceNoProjects,
	)

	stateCommandRunner := events.NewStateCommandRunner(
		pullUpdater,
		projectCommandBuilder,
		instrumentedProjectCmdRunner,
	)

	commentCommandRunnerByCmd := map[command.Name]events.CommentCommandRunner{
		command.Plan:            planCommandRunner,
		command.Apply:           applyCommandRunner,
		command.ApprovePolicies: approvePoliciesCommandRunner,
		command.Unlock:          unlockCommandRunner,
		command.Version:         versionCommandRunner,
		command.Import:          importCommandRunner,
		command.State:           stateCommandRunner,
	}

	var teamAllowlistChecker command.TeamAllowlistChecker
	if globalCfg.TeamAuthz.Command != "" {
		teamAllowlistChecker = &events.ExternalTeamAllowlistChecker{
			Command:                     globalCfg.TeamAuthz.Command,
			ExtraArgs:                   globalCfg.TeamAuthz.Args,
			ExternalTeamAllowlistRunner: &runtime.DefaultExternalTeamAllowlistRunner{},
		}
	} else {
		teamAllowlistChecker, err = command.NewTeamAllowlistChecker(userConfig.GithubTeamAllowlist)
		if err != nil {
			return nil, err
		}
	}

	varFileAllowlistChecker, err := events.NewVarFileAllowlistChecker(userConfig.VarFileAllowlist)
	if err != nil {
		return nil, err
	}

	commandRunner := &events.DefaultCommandRunner{
		VCSClient:                      vcsClient,
		GithubPullGetter:               githubClient,
		GitlabMergeRequestGetter:       gitlabClient,
		AzureDevopsPullGetter:          azuredevopsClient,
		GiteaPullGetter:                giteaClient,
		CommentCommandRunnerByCmd:      commentCommandRunnerByCmd,
		EventParser:                    eventParser,
		FailOnPreWorkflowHookError:     userConfig.FailOnPreWorkflowHookError,
		Logger:                         logger,
		GlobalCfg:                      globalCfg,
		StatsScope:                     statsScope.SubScope("cmd"),
		AllowForkPRs:                   userConfig.AllowForkPRs,
		AllowForkPRsFlag:               config.AllowForkPRsFlag,
		SilenceForkPRErrors:            userConfig.SilenceForkPRErrors,
		SilenceForkPRErrorsFlag:        config.SilenceForkPRErrorsFlag,
		DisableAutoplan:                userConfig.DisableAutoplan,
		DisableAutoplanLabel:           userConfig.DisableAutoplanLabel,
		Drainer:                        drainer,
		PreWorkflowHooksCommandRunner:  preWorkflowHooksCommandRunner,
		PostWorkflowHooksCommandRunner: postWorkflowHooksCommandRunner,
		PullStatusFetcher:              backend,
		TeamAllowlistChecker:           teamAllowlistChecker,
		VarFileAllowlistChecker:        varFileAllowlistChecker,
		CommitStatusUpdater:            commitStatusUpdater,
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
		LockDetailTemplate: web_templates.LockTemplate,
		WorkingDir:         workingDir,
		WorkingDirLocker:   workingDirLocker,
		Backend:            backend,
		DeleteLockCommand:  deleteLockCommand,
	}

	wsMux := websocket.NewMultiplexor(
		logger,
		controllers.JobIDKeyGenerator{},
		projectCmdOutputHandler,
		userConfig.WebsocketCheckOrigin,
	)

	jobsController := &controllers.JobsController{
		AtlantisVersion:          config.AtlantisVersion,
		AtlantisURL:              parsedURL,
		Logger:                   logger,
		ProjectJobsTemplate:      web_templates.ProjectJobsTemplate,
		ProjectJobsErrorTemplate: web_templates.ProjectJobsErrorTemplate,
		Backend:                  backend,
		WsMux:                    wsMux,
		KeyGenerator:             controllers.JobIDKeyGenerator{},
		StatsScope:               statsScope.SubScope("api"),
	}
	apiController := &controllers.APIController{
		APISecret:                      []byte(userConfig.APISecret),
		Locker:                         lockingClient,
		Logger:                         logger,
		Parser:                         eventParser,
		ProjectCommandBuilder:          projectCommandBuilder,
		ProjectPlanCommandRunner:       instrumentedProjectCmdRunner,
		ProjectApplyCommandRunner:      instrumentedProjectCmdRunner,
		FailOnPreWorkflowHookError:     userConfig.FailOnPreWorkflowHookError,
		PreWorkflowHooksCommandRunner:  preWorkflowHooksCommandRunner,
		PostWorkflowHooksCommandRunner: postWorkflowHooksCommandRunner,
		RepoAllowlistChecker:           repoAllowlist,
		Scope:                          statsScope.SubScope("api"),
		VCSClient:                      vcsClient,
	}

	eventsController := &events_controllers.VCSEventsController{
		CommandRunner:                   commandRunner,
		PullCleaner:                     pullClosedExecutor,
		Parser:                          eventParser,
		CommentParser:                   commentParser,
		Logger:                          logger,
		Scope:                           statsScope,
		ApplyDisabled:                   disableApply,
		GithubWebhookSecret:             []byte(userConfig.GithubWebhookSecret),
		GithubRequestValidator:          &events_controllers.DefaultGithubRequestValidator{},
		GitlabRequestParserValidator:    &events_controllers.DefaultGitlabRequestParserValidator{},
		GitlabWebhookSecret:             []byte(userConfig.GitlabWebhookSecret),
		RepoAllowlistChecker:            repoAllowlist,
		SilenceAllowlistErrors:          userConfig.SilenceAllowlistErrors,
		EmojiReaction:                   userConfig.EmojiReaction,
		ExecutableName:                  userConfig.ExecutableName,
		SupportedVCSHosts:               supportedVCSHosts,
		VCSClient:                       vcsClient,
		BitbucketWebhookSecret:          []byte(userConfig.BitbucketWebhookSecret),
		AzureDevopsWebhookBasicUser:     []byte(userConfig.AzureDevopsWebhookUser),
		AzureDevopsWebhookBasicPassword: []byte(userConfig.AzureDevopsWebhookPassword),
		AzureDevopsRequestValidator:     &events_controllers.DefaultAzureDevopsRequestValidator{},
		GiteaWebhookSecret:              []byte(userConfig.GiteaWebhookSecret),
	}
	githubAppController := &controllers.GithubAppController{
		AtlantisURL:         parsedURL,
		Logger:              logger,
		GithubSetupComplete: githubAppEnabled,
		GithubHostname:      userConfig.GithubHostname,
		GithubOrg:           userConfig.GithubOrg,
	}

	return &Server{
		AtlantisVersion:                config.AtlantisVersion,
		AtlantisURL:                    parsedURL,
		Router:                         underlyingRouter,
		Port:                           userConfig.Port,
		PostWorkflowHooksCommandRunner: postWorkflowHooksCommandRunner,
		PreWorkflowHooksCommandRunner:  preWorkflowHooksCommandRunner,
		CommandRunner:                  commandRunner,
		Logger:                         logger,
		StatsScope:                     statsScope,
		StatsReporter:                  statsReporter,
		StatsCloser:                    closer,
		Locker:                         lockingClient,
		ApplyLocker:                    applyLockingClient,
		VCSEventsController:            eventsController,
		GithubAppController:            githubAppController,
		LocksController:                locksController,
		JobsController:                 jobsController,
		StatusController:               statusController,
		APIController:                  apiController,
		IndexTemplate:                  web_templates.IndexTemplate,
		LockDetailTemplate:             web_templates.LockTemplate,
		ProjectJobsTemplate:            web_templates.ProjectJobsTemplate,
		ProjectJobsErrorTemplate:       web_templates.ProjectJobsErrorTemplate,
		SSLKeyFile:                     userConfig.SSLKeyFile,
		SSLCertFile:                    userConfig.SSLCertFile,
		DisableGlobalApplyLock:         userConfig.DisableGlobalApplyLock,
		Drainer:                        drainer,
		ProjectCmdOutputHandler:        projectCmdOutputHandler,
		WebAuthentication:              userConfig.WebBasicAuth,
		WebUsername:                    userConfig.WebUsername,
		WebPassword:                    userConfig.WebPassword,
		ScheduledExecutorService:       scheduledExecutorService,
	}, nil
}

// Start creates the routes and starts serving traffic.
func (s *Server) Start() error {
	s.Router.HandleFunc("/", s.Index).Methods("GET").MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
		return r.URL.Path == "/" || r.URL.Path == "/index.html"
	})
	s.Router.HandleFunc("/healthz", s.Healthz).Methods("GET")
	s.Router.HandleFunc("/status", s.StatusController.Get).Methods("GET")
	s.Router.PathPrefix("/static/").Handler(http.FileServer(http.FS(staticAssets)))
	s.Router.HandleFunc("/events", s.VCSEventsController.Post).Methods("POST")
	s.Router.HandleFunc("/api/plan", s.APIController.Plan).Methods("POST")
	s.Router.HandleFunc("/api/apply", s.APIController.Apply).Methods("POST")
	s.Router.HandleFunc("/github-app/exchange-code", s.GithubAppController.ExchangeCode).Methods("GET")
	s.Router.HandleFunc("/github-app/setup", s.GithubAppController.New).Methods("GET")
	s.Router.HandleFunc("/locks", s.LocksController.DeleteLock).Methods("DELETE").Queries("id", "{id:.*}")
	s.Router.HandleFunc("/lock", s.LocksController.GetLock).Methods("GET").
		Queries(LockViewRouteIDQueryParam, fmt.Sprintf("{%s}", LockViewRouteIDQueryParam)).Name(LockViewRouteName)
	s.Router.HandleFunc("/jobs/{job-id}", s.JobsController.GetProjectJobs).Methods("GET").Name(ProjectJobsViewRouteName)
	s.Router.HandleFunc("/jobs/{job-id}/ws", s.JobsController.GetProjectJobsWS).Methods("GET")

	r, ok := s.StatsReporter.(prometheus.Reporter)
	if ok {
		s.Router.Handle(s.CommandRunner.GlobalCfg.Metrics.Prometheus.Endpoint, r.HTTPHandler())
	}
	if !s.DisableGlobalApplyLock {
		s.Router.HandleFunc("/apply/lock", s.LocksController.LockApply).Methods("POST").Queries()
		s.Router.HandleFunc("/apply/unlock", s.LocksController.UnlockApply).Methods("DELETE").Queries()
	}

	n := negroni.New(&negroni.Recovery{
		Logger:     log.New(os.Stdout, "", log.LstdFlags),
		PrintStack: false,
		StackAll:   false,
		StackSize:  1024 * 8,
	}, NewRequestLogger(s))
	n.UseHandler(s.Router)

	defer s.Logger.Flush()

	// Ensure server gracefully drains connections when stopped.
	stop := make(chan os.Signal, 1)
	// Stop on SIGINTs and SIGTERMs.
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go s.ScheduledExecutorService.Run()

	go func() {
		s.ProjectCmdOutputHandler.Handle()
	}()

	tlsConfig := &tls.Config{GetCertificate: s.GetSSLCertificate, MinVersion: tls.VersionTLS12}

	server := &http.Server{Addr: fmt.Sprintf(":%d", s.Port), Handler: n, TLSConfig: tlsConfig, ReadHeaderTimeout: 10 * time.Second}
	go func() {
		s.Logger.Info("Atlantis started - listening on port %v", s.Port)

		var err error
		if s.SSLCertFile != "" && s.SSLKeyFile != "" {
			err = server.ListenAndServeTLS("", "")
		} else {
			err = server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			s.Logger.Err(err.Error())
		}
	}()
	<-stop

	s.Logger.Warn("Received interrupt. Waiting for in-progress operations to complete")
	s.waitForDrain()

	// flush stats before shutdown
	if err := s.StatsCloser.Close(); err != nil {
		s.Logger.Err(err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("while shutting down: %s", err)
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
			s.Logger.Info("Waiting for in-progress operations to complete, current in-progress ops: %d", s.Drainer.GetStatus().InProgressOps)
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

	var lockResults []web_templates.LockIndexData
	for id, v := range locks {
		lockURL, _ := s.Router.Get(LockViewRouteName).URL("id", url.QueryEscape(id))
		lockResults = append(lockResults, web_templates.LockIndexData{
			// NOTE: must use .String() instead of .Path because we need the
			// query params as part of the lock URL.
			LockPath:      lockURL.String(),
			RepoFullName:  v.Project.RepoFullName,
			LockedBy:      v.Pull.Author,
			PullNum:       v.Pull.Num,
			Path:          v.Project.Path,
			Workspace:     v.Workspace,
			Time:          v.Time,
			TimeFormatted: v.Time.Format("2006-01-02 15:04:05"),
		})
	}

	applyCmdLock, err := s.ApplyLocker.CheckApplyLock()
	s.Logger.Info("Apply Lock: %v", applyCmdLock)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, "Could not retrieve global apply lock: %s", err)
		return
	}

	applyLockData := web_templates.ApplyLockData{
		Time:                   applyCmdLock.Time,
		Locked:                 applyCmdLock.Locked,
		GlobalApplyLockEnabled: applyCmdLock.GlobalApplyLockEnabled,
		TimeFormatted:          applyCmdLock.Time.Format("2006-01-02 15:04:05"),
	}
	//Sort by date - newest to oldest.
	sort.SliceStable(lockResults, func(i, j int) bool { return lockResults[i].Time.After(lockResults[j].Time) })

	err = s.IndexTemplate.Execute(w, web_templates.IndexData{
		Locks:            lockResults,
		PullToJobMapping: preparePullToJobMappings(s),
		ApplyLock:        applyLockData,
		AtlantisVersion:  s.AtlantisVersion,
		CleanedBasePath:  s.AtlantisURL.Path,
	})
	if err != nil {
		s.Logger.Err(err.Error())
	}
}

func preparePullToJobMappings(s *Server) []jobs.PullInfoWithJobIDs {

	pullToJobMappings := s.ProjectCmdOutputHandler.GetPullToJobMapping()

	for i := range pullToJobMappings {
		for j := range pullToJobMappings[i].JobIDInfos {
			jobUrl, _ := s.Router.Get(ProjectJobsViewRouteName).URL("job-id", pullToJobMappings[i].JobIDInfos[j].JobID)
			pullToJobMappings[i].JobIDInfos[j].JobIDUrl = jobUrl.String()
			pullToJobMappings[i].JobIDInfos[j].TimeFormatted = pullToJobMappings[i].JobIDInfos[j].Time.Format("2006-01-02 15:04:05")
		}

		//Sort by date - newest to oldest.
		sort.SliceStable(pullToJobMappings[i].JobIDInfos, func(x, y int) bool {
			return pullToJobMappings[i].JobIDInfos[x].Time.After(pullToJobMappings[i].JobIDInfos[y].Time)
		})
	}

	//Sort by repository, project, path, workspace then date.
	sort.SliceStable(pullToJobMappings, func(x, y int) bool {
		if pullToJobMappings[x].Pull.RepoFullName != pullToJobMappings[y].Pull.RepoFullName {
			return pullToJobMappings[x].Pull.RepoFullName < pullToJobMappings[y].Pull.RepoFullName
		}
		if pullToJobMappings[x].Pull.ProjectName != pullToJobMappings[y].Pull.ProjectName {
			return pullToJobMappings[x].Pull.ProjectName < pullToJobMappings[y].Pull.ProjectName
		}
		if pullToJobMappings[x].Pull.Path != pullToJobMappings[y].Pull.Path {
			return pullToJobMappings[x].Pull.Path < pullToJobMappings[y].Pull.Path
		}
		return pullToJobMappings[x].Pull.Workspace < pullToJobMappings[y].Pull.Workspace
	})

	return pullToJobMappings
}

func mkSubDir(parentDir string, subDir string) (string, error) {
	fullDir := filepath.Join(parentDir, subDir)
	if err := os.MkdirAll(fullDir, 0700); err != nil {
		return "", errors.Wrapf(err, "unable to create dir %q", fullDir)
	}

	return fullDir, nil
}

// Healthz returns the health check response. It always returns a 200 currently.
func (s *Server) Healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(healthzData) // nolint: errcheck
}

var healthzData = []byte(`{
  "status": "ok"
}`)

func (s *Server) GetSSLCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	certStat, err := os.Stat(s.SSLCertFile)
	if err != nil {
		return nil, fmt.Errorf("while getting cert file modification time: %w", err)
	}

	keyStat, err := os.Stat(s.SSLKeyFile)
	if err != nil {
		return nil, fmt.Errorf("while getting key file modification time: %w", err)
	}

	if s.SSLCert == nil || certStat.ModTime() != s.CertLastRefreshTime || keyStat.ModTime() != s.KeyLastRefreshTime {
		cert, err := tls.LoadX509KeyPair(s.SSLCertFile, s.SSLKeyFile)
		if err != nil {
			return nil, fmt.Errorf("while loading tls cert: %w", err)
		}

		s.SSLCert = &cert
		s.CertLastRefreshTime = certStat.ModTime()
		s.KeyLastRefreshTime = keyStat.ModTime()
	}
	return s.SSLCert, nil
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
