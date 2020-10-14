package events

import (
	"github.com/runatlantis/atlantis/server/events/db"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
	"github.com/runatlantis/atlantis/server/logging"
)

type WorkflowHookRunner interface {
	Run(ctx models.WorkflowHookCommandContext)
}

type WorkflowHooksCommandRunner interface {
	RunPreHooks(baseRepo models.Repo, headRepo models.Repo, pull models.PullRequest, user models.User)
}

// DefaultWorkflowHooksCommandRunner is the first step when processing a workflow hook commands.
type DefaultWorkflowHooksCommandRunner struct {
	VCSClient                vcs.Client
	GithubPullGetter         GithubPullGetter
	AzureDevopsPullGetter    AzureDevopsPullGetter
	GitlabMergeRequestGetter GitlabMergeRequestGetter
	Logger                   logging.SimpleLogging
	WorkingDir               WorkingDir
	GlobalCfg                valid.GlobalCfg
	DB                       *db.BoltDB
	Drainer                  *Drainer
	DeleteLockCommand        DeleteLockCommand
}

func (w *DefaultWorkflowHooksCommandRunner) RunPreHooks(
	baseRepo models.Repo,
	headRepo models.Repo,
	pull models.PullRequest,
	user models.User,
) {
	w.Logger.Info("Running Pre Hooks for repo: %s and pr: %d", baseRepo.ID(), pull.Num)
	return
}
