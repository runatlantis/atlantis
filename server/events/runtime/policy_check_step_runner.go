package runtime

import (
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
)

// PolicyCheckStepRunner runs a policy check command given a ctx
type PolicyCheckStepRunner struct {
	versionEnsurer ExecutorVersionEnsurer
	argsResolver   ExecutorArgsResolver
	executor       Executor
}

// NewPolicyCheckStepRunner creates a new step runner from an executor workflow
func NewPolicyCheckStepRunner(executorWorkflow VersionedExecutorWorkflow) *PolicyCheckStepRunner {
	return &PolicyCheckStepRunner{
		versionEnsurer: executorWorkflow,
		argsResolver:   executorWorkflow,
		executor:       executorWorkflow,
	}
}

// Run ensures a given version for the executable, builds the args from the project context and then runs executable returning the result
func (p *PolicyCheckStepRunner) Run(ctx models.ProjectCommandContext, extraArgs []string, path string, envs map[string]string) (string, error) {
	executable, err := p.versionEnsurer.EnsureExecutorVersion(ctx.Log, ctx.PolicySets.Version)

	if err != nil {
		return "", errors.Wrapf(err, "ensuring policy executor version")
	}

	args, err := p.argsResolver.ResolveArgs(ctx)

	if err != nil {
		return "", errors.Wrapf(err, "resolving policy executor args")
	}

	stdOut, err := p.executor.Run(ctx.Log, executable, envs, args)

	if err != nil {
		return "", errors.Wrapf(err, "running policy executor")

	}

	return stdOut, nil
}
