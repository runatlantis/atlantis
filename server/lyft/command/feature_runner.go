package command

import (
	"fmt"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/lyft/feature"
)

// DefaultProjectCommandRunner implements ProjectCommandRunner.
type PlatformModeProjectRunner struct { //create object and test
	PlatformModeRunner events.ProjectCommandRunner
	PrModeRunner       events.ProjectCommandRunner
	Allocator          feature.Allocator
	Logger             logging.Logger
}

// Plan runs terraform plan for the project described by ctx.
func (p *PlatformModeProjectRunner) Plan(ctx command.ProjectContext) command.ProjectResult {
	shouldAllocate, err := p.Allocator.ShouldAllocate(feature.PlatformMode, feature.FeatureContext{RepoName: ctx.HeadRepo.FullName})
	if err != nil {
		p.Logger.ErrorContext(ctx.RequestCtx, fmt.Sprintf("unable to allocate for feature: %s, error: %s", feature.PlatformMode, err))
	}

	if shouldAllocate && (ctx.WorkflowModeType == valid.PlatformWorkflowMode) {
		return p.PlatformModeRunner.Plan(ctx)
	}

	return p.PrModeRunner.Plan(ctx)
}

// PolicyCheck evaluates policies defined with Rego for the project described by ctx.
func (p *PlatformModeProjectRunner) PolicyCheck(ctx command.ProjectContext) command.ProjectResult {
	shouldAllocate, err := p.Allocator.ShouldAllocate(feature.PlatformMode, feature.FeatureContext{RepoName: ctx.HeadRepo.FullName})
	if err != nil {
		p.Logger.ErrorContext(ctx.RequestCtx, fmt.Sprintf("unable to allocate for feature: %s, error: %s", feature.PlatformMode, err))
	}

	if shouldAllocate && (ctx.WorkflowModeType == valid.PlatformWorkflowMode) {
		return p.PlatformModeRunner.PolicyCheck(ctx)
	}

	return p.PrModeRunner.PolicyCheck(ctx)
}

// Apply runs terraform apply for the project described by ctx.
func (p *PlatformModeProjectRunner) Apply(ctx command.ProjectContext) command.ProjectResult {
	shouldAllocate, err := p.Allocator.ShouldAllocate(feature.PlatformMode, feature.FeatureContext{RepoName: ctx.HeadRepo.FullName})
	if err != nil {
		p.Logger.ErrorContext(ctx.RequestCtx, fmt.Sprintf("unable to allocate for feature: %s, error: %s", feature.PlatformMode, err))
	}

	if shouldAllocate && (ctx.WorkflowModeType == valid.PlatformWorkflowMode) {
		return command.ProjectResult{
			Command:      command.Apply,
			RepoRelDir:   ctx.RepoRelDir,
			Workspace:    ctx.Workspace,
			ProjectName:  ctx.ProjectName,
			StatusID:     ctx.StatusID,
			ApplySuccess: "atlantis apply is disabled for this project. Please track the deployment when the PR is merged. ",
		}
	}

	return p.PrModeRunner.Apply(ctx)
}

func (p *PlatformModeProjectRunner) ApprovePolicies(ctx command.ProjectContext) command.ProjectResult {
	shouldAllocate, err := p.Allocator.ShouldAllocate(feature.PlatformMode, feature.FeatureContext{RepoName: ctx.HeadRepo.FullName})
	if err != nil {
		p.Logger.ErrorContext(ctx.RequestCtx, fmt.Sprintf("unable to allocate for feature: %s, error: %s", feature.PlatformMode, err))
	}

	if shouldAllocate && (ctx.WorkflowModeType == valid.PlatformWorkflowMode) {
		return p.PlatformModeRunner.ApprovePolicies(ctx)
	}

	return p.PrModeRunner.ApprovePolicies(ctx)
}

func (p *PlatformModeProjectRunner) Version(ctx command.ProjectContext) command.ProjectResult {
	shouldAllocate, err := p.Allocator.ShouldAllocate(feature.PlatformMode, feature.FeatureContext{RepoName: ctx.HeadRepo.FullName})
	if err != nil {
		p.Logger.ErrorContext(ctx.RequestCtx, fmt.Sprintf("unable to allocate for feature: %s, error: %s", feature.PlatformMode, err))
	}

	if shouldAllocate && (ctx.WorkflowModeType == valid.PlatformWorkflowMode) {
		return p.PlatformModeRunner.Version(ctx)
	}

	return p.PrModeRunner.Version(ctx)
}
