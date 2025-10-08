package status

import (
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

//go:generate pegomock generate github.com/runatlantis/atlantis/server/events/status --package mocks -o mocks/mock_status_policy.go StatusPolicy

// StatusPolicy defines the business logic for when to set, clear, or silence status checks
type StatusPolicy interface {
	// DecideOnStart determines what to do when a command starts
	DecideOnStart(ctx *command.Context, cmdName command.Name) StatusDecision

	// DecideOnEnd determines what to do when a command completes
	DecideOnEnd(ctx *command.Context, cmdName command.Name, result *command.Result) StatusDecision

	// DecideOnNoProjects determines what to do when no projects are found
	DecideOnNoProjects(ctx *command.Context, cmdName command.Name) StatusDecision
}

// SilencePolicy implements StatusPolicy with silence flag awareness
type SilencePolicy struct {
	// Silence flags from user configuration
	SilenceNoProjects          bool
	SilenceVCSStatusNoPlans    bool
	SilenceVCSStatusNoProjects bool
	SilenceForkPRErrors        bool
}

// NewSilencePolicy creates a new SilencePolicy with the given silence flags
func NewSilencePolicy(silenceNoProjects, silenceVCSStatusNoPlans, silenceVCSStatusNoProjects, silenceForkPRErrors bool) StatusPolicy {
	return &SilencePolicy{
		SilenceNoProjects:          silenceNoProjects,
		SilenceVCSStatusNoPlans:    silenceVCSStatusNoPlans,
		SilenceVCSStatusNoProjects: silenceVCSStatusNoProjects,
		SilenceForkPRErrors:        silenceForkPRErrors,
	}
}

// DecideOnStart determines what to do when starting a command
func (p *SilencePolicy) DecideOnStart(ctx *command.Context, cmdName command.Name) StatusDecision {
	// Check fork PR silence
	if p.shouldSilenceForkPR(ctx) {
		return StatusDecision{
			Operation:   OperationSilence,
			Reason:      DefaultSilenceReasons.ForkPRError,
			SilenceType: "SilenceForkPRErrors",
		}
	}

	// Check if we should silence due to VCS status flags
	if p.SilenceVCSStatusNoProjects {
		return StatusDecision{
			Operation:   OperationSilence,
			Reason:      "silence VCS status for projects enabled",
			SilenceType: "SilenceVCSStatusNoProjects",
		}
	}

	// Default: set pending status
	return StatusDecision{
		Operation: OperationSet,
		Status:    models.PendingCommitStatus,
		Reason:    "command starting - setting pending status",
	}
}

// DecideOnEnd determines what to do when a command completes
func (p *SilencePolicy) DecideOnEnd(ctx *command.Context, cmdName command.Name, result *command.Result) StatusDecision {
	// Check fork PR silence first
	if p.shouldSilenceForkPR(ctx) {
		return StatusDecision{
			Operation:   OperationSilence,
			Reason:      DefaultSilenceReasons.ForkPRError,
			SilenceType: "SilenceForkPRErrors",
		}
	}

	// If command failed, always set failure (unless fork PR silenced)
	if result.HasErrors() {
		return StatusDecision{
			Operation: OperationSet,
			Status:    models.FailedCommitStatus,
			Reason:    "command failed",
		}
	}

	// Count successful projects
	numSuccess := 0
	numTotal := len(result.ProjectResults)
	for _, projectResult := range result.ProjectResults {
		if !projectResult.IsSuccessful() {
			continue
		}
		numSuccess++
	}

	// If silence is enabled, don't set status
	if p.SilenceVCSStatusNoProjects {
		return StatusDecision{
			Operation:   OperationSilence,
			Reason:      "silence VCS status enabled",
			SilenceType: "SilenceVCSStatusNoProjects",
		}
	}

	// Default: set success status with counts
	return StatusDecision{
		Operation:  OperationSet,
		Status:     models.SuccessCommitStatus,
		NumSuccess: numSuccess,
		NumTotal:   numTotal,
		Reason:     "command completed successfully",
	}
}

// DecideOnNoProjects determines what to do when no projects are found
func (p *SilencePolicy) DecideOnNoProjects(ctx *command.Context, cmdName command.Name) StatusDecision {
	// Check fork PR silence first
	if p.shouldSilenceForkPR(ctx) {
		return StatusDecision{
			Operation:   OperationSilence,
			Reason:      DefaultSilenceReasons.ForkPRError,
			SilenceType: "SilenceForkPRErrors",
		}
	}

	// Check silence flags specific to no projects/plans
	if p.shouldSilenceNoProjects(cmdName) {
		silenceType := p.getSilenceTypeForNoProjects(cmdName)
		return StatusDecision{
			Operation:   OperationSilence,
			Reason:      DefaultSilenceReasons.NoProjects,
			SilenceType: silenceType,
		}
	}

	// Default: set success status with 0/0
	return StatusDecision{
		Operation:  OperationSet,
		Status:     models.SuccessCommitStatus,
		NumSuccess: 0,
		NumTotal:   0,
		Reason:     "no projects found - setting success 0/0",
	}
}

// shouldSilenceForkPR checks if this is a fork PR that should be silenced
func (p *SilencePolicy) shouldSilenceForkPR(ctx *command.Context) bool {
	// A fork PR is when head repo owner != base repo owner
	isForkPR := ctx.HeadRepo.Owner != ctx.Pull.BaseRepo.Owner

	// We silence if it's a fork PR and SilenceForkPRErrors is enabled
	return isForkPR && p.SilenceForkPRErrors
}

// shouldSilenceNoProjects checks if we should silence when no projects are found
func (p *SilencePolicy) shouldSilenceNoProjects(cmdName command.Name) bool {
	// Check different silence flags based on command
	switch cmdName {
	case command.Plan:
		return p.SilenceVCSStatusNoProjects || p.SilenceVCSStatusNoPlans
	case command.Apply, command.PolicyCheck:
		return p.SilenceVCSStatusNoProjects
	default:
		return p.SilenceVCSStatusNoProjects
	}
}

// getSilenceTypeForNoProjects returns which silence flag is causing the silence
func (p *SilencePolicy) getSilenceTypeForNoProjects(cmdName command.Name) string {
	if p.SilenceVCSStatusNoProjects {
		return "SilenceVCSStatusNoProjects"
	}
	if cmdName == command.Plan && p.SilenceVCSStatusNoPlans {
		return "SilenceVCSStatusNoPlans"
	}
	return "unknown"
}
