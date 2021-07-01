package runtime

import (
	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
)

// PolicyCheckStepRunner runs a policy check command given a ctx
type PolicyCheckStepRunner struct {
	versionEnsurer ExecutorVersionEnsurer
	executor       Executor
}

// NewPolicyCheckStepRunner creates a new step runner from an executor workflow
func NewPolicyCheckStepRunner(defaultTfVersion *version.Version, executorWorkflow VersionedExecutorWorkflow) (Runner, error) {

	runner := &PlanTypeStepRunnerDelegate{
		defaultRunner: &PolicyCheckStepRunner{
			versionEnsurer: executorWorkflow,
			executor:       executorWorkflow,
		},
		remotePlanRunner: RemoteBackendUnsupportedRunner{},
	}

	return NewMinimumVersionStepRunnerDelegate(minimumShowTfVersion, defaultTfVersion, runner)
}

// Run ensures a given version for the executable, builds the args from the project context and then runs executable returning the result
func (p *PolicyCheckStepRunner) Run(ctx models.ProjectCommandContext, extraArgs []string, path string, envs map[string]string) (string, error) {
	executable, err := p.versionEnsurer.EnsureExecutorVersion(ctx.Log, ctx.PolicySets.Version)

	if err != nil {
		return "", errors.Wrapf(err, "ensuring policy executor version")
	}

	return p.executor.Run(ctx, executable, envs, path, extraArgs)
}
