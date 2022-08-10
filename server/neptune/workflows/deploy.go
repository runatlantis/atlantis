package workflows

import (
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/revision"
	"go.temporal.io/sdk/workflow"
)

// Export anything that callers need such as requests, signals, etc.
type DeployRequest = deploy.Request
type Repo = deploy.Repo

type DeployNewRevisionSignalRequest = revision.NewRevisionRequest

var DeployTaskQueue = deploy.TaskQueue

var DeployNewRevisionSignalID = deploy.NewRevisionSignalID

func Deploy(ctx workflow.Context, request DeployRequest) error {
	return deploy.Workflow(ctx, request)
}
