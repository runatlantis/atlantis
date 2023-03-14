package workflows

import (
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/prrevision"
	"go.temporal.io/sdk/workflow"
)

type PRRevisionRevisionRequest = prrevision.Request

func PRRevision(ctx workflow.Context, request PRRevisionRevisionRequest) error {
	return prrevision.Workflow(ctx, request)
}
