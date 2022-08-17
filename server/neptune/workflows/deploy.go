package workflows

import (
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/revision"
	"github.com/uber-go/tally/v4"
	"go.temporal.io/sdk/workflow"
)

// Export anything that callers need such as requests, signals, etc.
type DeployRequest = deploy.Request
type Repo = deploy.Repo
type Root = deploy.Root
type Job = deploy.Job
type Step = deploy.Step
type AppCredentials = deploy.AppCredentials

type DeployNewRevisionSignalRequest = revision.NewRevisionRequest

var DeployTaskQueue = deploy.TaskQueue

var DeployNewRevisionSignalID = deploy.NewRevisionSignalID

type Activities struct {
	activities.Deploy
}

func NewActivities(appConfig githubapp.Config, scope tally.Scope) (*Activities, error) {
	deployActivities, err := activities.NewDeploy(appConfig, scope)

	if err != nil {
		return nil, errors.Wrap(err, "initializing deploy activities")
	}

	return &Activities{
		Deploy: *deployActivities,
	}, nil
}

func Deploy(ctx workflow.Context, request DeployRequest) error {
	return deploy.Workflow(ctx, request, Terraform)
}
