package workflows

import (
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/prrevision"
	"go.temporal.io/sdk/workflow"
)

type PRRevisionRevisionRequest = prrevision.Request

var PRRevisionTaskQueue = prrevision.TaskQueue

func PRRevision(ctx workflow.Context, request PRRevisionRevisionRequest) error {
	return prrevision.Workflow(ctx, request)
}
