package workflows

import (
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/revision"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/request"
	"github.com/uber-go/tally/v4"
	"go.temporal.io/sdk/workflow"
)

// Export anything that callers need such as requests, signals, etc.
type DeployRequest = deploy.Request
type Repo = request.Repo
type Root = request.Root
type Job = request.Job
type Step = request.Step
type AppCredentials = request.AppCredentials
type HeadCommit = request.Commit
type Ref = request.Ref

type PlanMode = request.PlanMode

const DestroyPlanMode = request.DestroyPlanMode
const NormalPlanMode = request.NormalPlanMode

type DeployNewRevisionSignalRequest = revision.NewRevisionRequest

var DeployTaskQueue = deploy.TaskQueue

var DeployNewRevisionSignalID = deploy.NewRevisionSignalID

type DeployActivities struct {
	activities.Deploy
}

func NewDeployActivities(appConfig githubapp.Config, scope tally.Scope) (*DeployActivities, error) {
	deployActivities, err := activities.NewDeploy(appConfig, scope)

	if err != nil {
		return nil, errors.Wrap(err, "initializing deploy activities")
	}

	return &DeployActivities{
		Deploy: *deployActivities,
	}, nil
}

func Deploy(ctx workflow.Context, request DeployRequest) error {
	return deploy.Workflow(ctx, request, Terraform)
}
