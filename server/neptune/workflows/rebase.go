package workflows

import (
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/rebase"
	"go.temporal.io/sdk/workflow"
)

type RebaseRequest = rebase.Request

func Rebase(ctx workflow.Context, request RebaseRequest) error {
	return rebase.Workflow(ctx, request)
}
