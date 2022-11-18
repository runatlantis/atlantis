package event

import (
	"context"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/neptune/workflows"
	"github.com/runatlantis/atlantis/server/vcs/provider/github"
)

type CheckSuite struct {
	Action            CheckRunAction
	HeadSha           string
	Repo              models.Repo
	Sender            models.User
	InstallationToken int64
	Branch            string
}

type CheckSuiteHandler struct {
	Logger       logging.Logger
	Scheduler    scheduler
	RootDeployer rootDeployer
}

func (h *CheckSuiteHandler) Handle(ctx context.Context, event CheckSuite) error {
	if event.Action.GetType() != ReRequestedActionType {
		h.Logger.DebugContext(ctx, "ignoring checks event that isn't a rerequested action")
		return nil
	}
	// Block force applies
	if event.Branch != event.Repo.DefaultBranch {
		h.Logger.DebugContext(ctx, "dropping event branch unexpected ref")
		return nil
	}
	return h.Scheduler.Schedule(ctx, func(ctx context.Context) error {
		return h.handle(ctx, event)
	})
}

func (h *CheckSuiteHandler) handle(ctx context.Context, event CheckSuite) error {
	builderOptions := BuilderOptions{
		RepoFetcherOptions: github.RepoFetcherOptions{
			ShallowClone: true,
		},
		FileFetcherOptions: github.FileFetcherOptions{
			Sha: event.HeadSha,
		},
	}
	rootDeployOptions := RootDeployOptions{
		Repo:              event.Repo,
		Branch:            event.Branch,
		Revision:          event.HeadSha,
		Sender:            event.Sender,
		InstallationToken: event.InstallationToken,
		BuilderOptions:    builderOptions,
		Trigger:           workflows.ManualTrigger,
		Rerun:             true,
	}
	return h.RootDeployer.Deploy(ctx, rootDeployOptions)
}
