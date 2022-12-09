package runtime

import (
	"context"
	"github.com/runatlantis/atlantis/server/lyft/feature"

	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/command"
)

// PolicyCheckStepRunner runs a policy check command given a ctx
type PolicyCheckStepRunner struct {
	VersionEnsurer ExecutorVersionEnsurer
	Executor       Executor
	// TODO: remove these
	LegacyExecutor Executor
	Allocator      feature.Allocator
}

// NewPolicyCheckStepRunner creates a new step runner from an Executor workflow
func NewPolicyCheckStepRunner(
	defaultTfVersion *version.Version,
	versionEnsurer VersionedExecutorWorkflow,
	executor Executor,
	allocator feature.Allocator) (Runner, error) {
	runner := &PlanTypeStepRunnerDelegate{
		defaultRunner: &PolicyCheckStepRunner{
			VersionEnsurer: versionEnsurer,
			LegacyExecutor: versionEnsurer,
			Executor:       executor,
			Allocator:      allocator,
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

	// TODO: remove feature allocator when the legacy executor is deprecated
	shouldAllocate, err := p.Allocator.ShouldAllocate(feature.PolicyV2, feature.FeatureContext{
		RepoName: cmdCtx.BaseRepo.FullName,
	})
	if err != nil {
		cmdCtx.Log.ErrorContext(ctx, "unable to allocate policy v2, continuing with legacy mode")
		return p.LegacyExecutor.Run(ctx, cmdCtx, executable, envs, path, extraArgs)
	}
	if shouldAllocate {
		return p.Executor.Run(ctx, cmdCtx, executable, envs, path, extraArgs)
	}
	return p.LegacyExecutor.Run(ctx, cmdCtx, executable, envs, path, extraArgs)
}
