package api

import (
	"context"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/neptune/gateway/api/request"
	"github.com/runatlantis/atlantis/server/neptune/gateway/event"
	"github.com/runatlantis/atlantis/server/neptune/sync"
	"github.com/runatlantis/atlantis/server/neptune/workflows"
	internalGH "github.com/runatlantis/atlantis/server/vcs/provider/github"
)

const (
	RepoVarKey     = "repo"
	OwnerVarKey    = "owner"
	UsernameVarKey = "username"
	RootVarKey     = "root"
)

type rootDeployer interface {
	Deploy(ctx context.Context, deployOptions event.RootDeployOptions) error
}

type scheduler interface {
	Schedule(ctx context.Context, f sync.Executor) error
}

type DeployHandler struct {
	Deployer  rootDeployer
	Scheduler scheduler
	Logger    logging.Logger
}

func (c *DeployHandler) Handle(ctx context.Context, r request.Deploy) error {
	c.Logger.Info("scheduling deploy API request")

	return c.Scheduler.Schedule(ctx,
		func(ctx context.Context) error {
			return c.Deployer.Deploy(ctx, event.RootDeployOptions{
				Repo:      r.Repo,
				Branch:    r.Branch,
				Revision:  r.Revision,
				RootNames: r.RootNames,

				// this we won't have, maybe we can add some gh auth to add this
				Sender: r.User,

				InstallationToken: r.InstallationToken,

				RepoFetcherOptions: &internalGH.RepoFetcherOptions{
					CloneDepth: 1,
				},

				Trigger: workflows.ManualTrigger,
			})
		},
	)
}
