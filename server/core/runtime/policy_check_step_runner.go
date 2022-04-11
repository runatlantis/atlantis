package runtime

import (
	"context"

	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/command"
)

// PolicyCheckStepRunner runs a policy check command given a ctx
type PolicyCheckStepRunner struct {
	VersionEnsurer ExecutorVersionEnsurer
	Executor       Executor
}

// NewPolicyCheckStepRunner creates a new step runner from an Executor workflow
func NewPolicyCheckStepRunner(defaultTfVersion *version.Version, executorWorkflow VersionedExecutorWorkflow) (Runner, error) {

	runner := &PlanTypeStepRunnerDelegate{
		defaultRunner: &PolicyCheckStepRunner{
			VersionEnsurer: executorWorkflow,
			Executor:       executorWorkflow,
		},
		remotePlanRunner: RemoteBackendUnsupportedRunner{},
	}

	return NewMinimumVersionStepRunnerDelegate(minimumShowTfVersion, defaultTfVersion, runner)
}

// Run ensures a given version for the executable, builds the args from the project context and then runs executable returning the result
func (p *PolicyCheckStepRunner) Run(ctx context.Context, cmdCtx command.ProjectContext, extraArgs []string, path string, envs map[string]string) (string, error) {
	executable, err := p.VersionEnsurer.EnsureExecutorVersion(cmdCtx.Log, cmdCtx.PolicySets.Version)

	if err != nil {
		return "", errors.Wrapf(err, "ensuring policy Executor version")
	}

	return p.Executor.Run(ctx, cmdCtx, executable, envs, path, extraArgs)
}
