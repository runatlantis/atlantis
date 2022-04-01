package instrumentation

import (
	"context"

	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/metrics"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/logging/fields"
)

type PreWorkflowHookRunner struct {
	events.PreWorkflowHooksCommandRunner
	Logger logging.Logger
}

func (r *PreWorkflowHookRunner) RunPreHooks(ctx context.Context, cmdCtx *command.Context) error {
	scope := cmdCtx.Scope.SubScope("pre_workflow_hook")

	executionSuccess := scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := scope.Counter(metrics.ExecutionErrorMetric)

	err := r.PreWorkflowHooksCommandRunner.RunPreHooks(ctx, cmdCtx)
	if err != nil {
		executionError.Inc(1)
		return err
	}

	//TODO: thread context and use related logging methods.
	r.Logger.InfoContext(ctx, "pre-workflow-hook success", fields.PullRequest(cmdCtx.Pull))
	executionSuccess.Inc(1)
	return nil
}
