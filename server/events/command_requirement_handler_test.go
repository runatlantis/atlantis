// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"testing"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/config/raw"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAggregateApplyRequirements_ValidatePlanProject(t *testing.T) {
	repoDir := "repoDir"
	fullRequirements := []string{
		raw.ApprovedRequirement,
		valid.PoliciesPassedCommandReq,
		raw.MergeableRequirement,
		raw.UnDivergedRequirement,
	}
	tests := []struct {
		name        string
		ctx         command.ProjectContext
		setup       func(workingDir *mocks.MockWorkingDir)
		wantFailure string
		wantErr     assert.ErrorAssertionFunc
	}{
		{
			name:    "pass no requirements",
			ctx:     command.ProjectContext{},
			wantErr: assert.NoError,
		},
		{
			name: "pass full requirements",
			ctx: command.ProjectContext{
				PlanRequirements: fullRequirements,
				PullReqStatus: models.PullReqStatus{
					ApprovalStatus:  models.ApprovalStatus{IsApproved: true},
					MergeableStatus: models.MergeableStatus{IsMergeable: true},
				},
				ProjectPlanStatus: models.PassedPolicyCheckStatus,
			},
			setup: func(workingDir *mocks.MockWorkingDir) {
				When(workingDir.HasDivergedFromPullHead(Any[logging.SimpleLogging](), Any[string](), Any[string](), Any[[]string](), Any[models.PullRequest]())).ThenReturn(false)
			},
			wantErr: assert.NoError,
		},
		{
			name: "fail by no approved",
			ctx: command.ProjectContext{
				PlanRequirements: []string{raw.ApprovedRequirement},
				PullReqStatus: models.PullReqStatus{
					ApprovalStatus: models.ApprovalStatus{IsApproved: false},
				},
			},
			wantFailure: "Pull request must be approved according to the project's approval rules before running plan.",
			wantErr:     assert.NoError,
		},
		{
			name: "fail by no mergeable",
			ctx: command.ProjectContext{
				PlanRequirements: []string{raw.MergeableRequirement},
				PullReqStatus: models.PullReqStatus{
					MergeableStatus: models.MergeableStatus{IsMergeable: false},
				},
			},
			wantFailure: "Pull request must be mergeable before running plan.",
			wantErr:     assert.NoError,
		},
		{
			name: "fail by no mergeable with reason",
			ctx: command.ProjectContext{
				PlanRequirements: []string{raw.MergeableRequirement},
				PullReqStatus: models.PullReqStatus{
					MergeableStatus: models.MergeableStatus{
						IsMergeable: false,
						Reason:      "some reason",
					},
				},
			},
			wantFailure: "Pull request must be mergeable before running plan (some reason).",
			wantErr:     assert.NoError,
		},
		{
			name: "fail by diverged",
			ctx: command.ProjectContext{
				PlanRequirements: []string{raw.UnDivergedRequirement},
			},
			setup: func(workingDir *mocks.MockWorkingDir) {
				When(workingDir.HasDivergedFromPullHead(Any[logging.SimpleLogging](), Any[string](), Any[string](), Any[[]string](), Any[models.PullRequest]())).ThenReturn(true)
			},
			wantFailure: "Default branch must be rebased onto pull request before running plan.",
			wantErr:     assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RegisterMockTestingT(t)
			workingDir := mocks.NewMockWorkingDir()
			a := &events.DefaultCommandRequirementHandler{WorkingDir: workingDir}
			if tt.setup != nil {
				tt.setup(workingDir)
			}
			gotFailure, err := a.ValidatePlanProject(repoDir, tt.ctx)
			if !tt.wantErr(t, err, fmt.Sprintf("ValidatePlanProject(%v, %v)", repoDir, tt.ctx)) {
				return
			}
			assert.Equalf(t, tt.wantFailure, gotFailure, "ValidatePlanProject(%v, %v)", repoDir, tt.ctx)
		})
	}
}

func TestAggregateApplyRequirements_ValidateApplyProject(t *testing.T) {
	repoDir := "repoDir"
	fullRequirements := []string{
		raw.ApprovedRequirement,
		valid.PoliciesPassedCommandReq,
		raw.MergeableRequirement,
		raw.UnDivergedRequirement,
	}
	tests := []struct {
		name        string
		ctx         command.ProjectContext
		setup       func(workingDir *mocks.MockWorkingDir)
		wantFailure string
		wantErr     assert.ErrorAssertionFunc
	}{
		{
			name:    "pass no requirements",
			ctx:     command.ProjectContext{},
			wantErr: assert.NoError,
		},
		{
			name: "API call without PR enforces approved requirement by default",
			ctx: command.ProjectContext{
				Log:               logging.NewNoopLogger(t),
				API:               true,
				Pull:              models.PullRequest{Num: -1},
				ApplyRequirements: []string{raw.ApprovedRequirement},
				PullReqStatus: models.PullReqStatus{
					ApprovalStatus: models.ApprovalStatus{IsApproved: false},
				},
			},
			wantFailure: "Pull request must be approved according to the project's approval rules before running apply.",
			wantErr:     assert.NoError,
		},
		{
			name: "explicit non-PR API opt-in skips approved requirement",
			ctx: command.ProjectContext{
				Log:                logging.NewNoopLogger(t),
				API:                true,
				SkipPRRequirements: true,
				Pull:               models.PullRequest{Num: -1},
				ApplyRequirements:  []string{raw.ApprovedRequirement},
				PullReqStatus: models.PullReqStatus{
					ApprovalStatus: models.ApprovalStatus{IsApproved: false},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "API call without PR enforces mergeable requirement by default",
			ctx: command.ProjectContext{
				Log:               logging.NewNoopLogger(t),
				API:               true,
				Pull:              models.PullRequest{Num: -2},
				ApplyRequirements: []string{raw.MergeableRequirement},
				PullReqStatus: models.PullReqStatus{
					MergeableStatus: models.MergeableStatus{IsMergeable: false},
				},
			},
			wantFailure: "Pull request must be mergeable before running apply.",
			wantErr:     assert.NoError,
		},
		{
			name: "explicit non-PR API opt-in skips mergeable requirement",
			ctx: command.ProjectContext{
				Log:                logging.NewNoopLogger(t),
				API:                true,
				SkipPRRequirements: true,
				Pull:               models.PullRequest{Num: -2},
				ApplyRequirements:  []string{raw.MergeableRequirement},
				PullReqStatus: models.PullReqStatus{
					MergeableStatus: models.MergeableStatus{IsMergeable: false},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "API call with PR still enforces approved requirement",
			ctx: command.ProjectContext{
				Log:               logging.NewNoopLogger(t),
				API:               true,
				Pull:              models.PullRequest{Num: 123},
				ApplyRequirements: []string{raw.ApprovedRequirement},
				PullReqStatus: models.PullReqStatus{
					ApprovalStatus: models.ApprovalStatus{IsApproved: false},
				},
			},
			wantFailure: "Pull request must be approved according to the project's approval rules before running apply.",
			wantErr:     assert.NoError,
		},
		{
			name: "API call without PR still enforces undiverged requirement",
			ctx: command.ProjectContext{
				Log:               logging.NewNoopLogger(t),
				API:               true,
				Pull:              models.PullRequest{Num: 0},
				ApplyRequirements: []string{raw.UnDivergedRequirement},
			},
			setup: func(workingDir *mocks.MockWorkingDir) {
				When(workingDir.HasDiverged(
					Any[logging.SimpleLogging](),
					Any[string](),
					Any[string](),
					Any[[]string](),
					Any[models.PullRequest](),
				)).ThenReturn(true)
			},
			wantFailure: "Default branch must be rebased onto pull request before running apply.",
			wantErr:     assert.NoError,
		},
		{
			name: "opted-in API call without PR fails policies_passed when no policy status exists",
			ctx: command.ProjectContext{
				Log:               logging.NewNoopLogger(t),
				API:               true,
				Pull:              models.PullRequest{Num: -1},
				RunPolicyChecks:   true,
				ApplyRequirements: []string{valid.PoliciesPassedCommandReq},
				PolicySets: valid.PolicySets{
					PolicySets: []valid.PolicySet{{
						Name:         "policy1",
						ApproveCount: 1,
					}},
				},
			},
			wantFailure: "All policies must pass for project before running apply.",
			wantErr:     assert.NoError,
		},
		{
			name: "API call without policy execution uses existing policy status semantics",
			ctx: command.ProjectContext{
				Log:               logging.NewNoopLogger(t),
				API:               true,
				Pull:              models.PullRequest{Num: 0},
				ApplyRequirements: []string{valid.PoliciesPassedCommandReq},
				PolicySets: valid.PolicySets{
					PolicySets: []valid.PolicySet{{
						Name:         "policy1",
						ApproveCount: 1,
					}},
				},
			},
			wantFailure: "",
			wantErr:     assert.NoError,
		},
		{
			name: "API call without PR still enforces policies_passed requirement",
			ctx: command.ProjectContext{
				Log:               logging.NewNoopLogger(t),
				API:               true,
				Pull:              models.PullRequest{Num: 0},
				ApplyRequirements: []string{valid.PoliciesPassedCommandReq},
				ProjectPlanStatus: models.ErroredPolicyCheckStatus,
				ProjectPolicyStatus: []models.PolicySetStatus{
					{
						PolicySetName: "policy1",
						Passed:        false,
						Approvals:     []models.PolicySetApproval{},
					},
				},
				PolicySets: valid.PolicySets{
					PolicySets: []valid.PolicySet{
						{
							Name:         "policy1",
							ApproveCount: 1,
						},
					},
				},
			},
			wantFailure: "All policies must pass for project before running apply.",
			wantErr:     assert.NoError,
		},
		{
			name: "pass full requirements",
			ctx: command.ProjectContext{
				ApplyRequirements: fullRequirements,
				PullReqStatus: models.PullReqStatus{
					ApprovalStatus:  models.ApprovalStatus{IsApproved: true},
					MergeableStatus: models.MergeableStatus{IsMergeable: true},
				},
				ProjectPlanStatus: models.PassedPolicyCheckStatus,
			},
			setup: func(workingDir *mocks.MockWorkingDir) {
				When(workingDir.HasDiverged(Any[logging.SimpleLogging](), Any[string](), Any[string](), Any[[]string](), Any[models.PullRequest]())).ThenReturn(false)
			},
			wantErr: assert.NoError,
		},
		{
			name: "fail by no approved",
			ctx: command.ProjectContext{
				ApplyRequirements: []string{raw.ApprovedRequirement},
				PullReqStatus: models.PullReqStatus{
					ApprovalStatus: models.ApprovalStatus{IsApproved: false},
				},
			},
			wantFailure: "Pull request must be approved according to the project's approval rules before running apply.",
			wantErr:     assert.NoError,
		},
		{
			name: "fail by no policy passed",
			ctx: command.ProjectContext{
				ApplyRequirements: []string{valid.PoliciesPassedCommandReq},
				ProjectPlanStatus: models.ErroredPolicyCheckStatus,
				ProjectPolicyStatus: []models.PolicySetStatus{
					{
						PolicySetName: "policy1",
						Passed:        false,
						Approvals:     nil,
					},
				},
				PolicySets: valid.PolicySets{
					PolicySets: []valid.PolicySet{
						{
							Name:         "policy1",
							ApproveCount: 1,
						},
					},
				},
			},
			wantFailure: "All policies must pass for project before running apply.",
			wantErr:     assert.NoError,
		},
		{
			name: "fail by no mergeable",
			ctx: command.ProjectContext{
				ApplyRequirements: []string{raw.MergeableRequirement},
				PullReqStatus: models.PullReqStatus{
					MergeableStatus: models.MergeableStatus{IsMergeable: false},
				},
			},
			wantFailure: "Pull request must be mergeable before running apply.",
			wantErr:     assert.NoError,
		},
		{
			name: "fail by diverged",
			ctx: command.ProjectContext{
				ApplyRequirements: []string{raw.UnDivergedRequirement},
			},
			setup: func(workingDir *mocks.MockWorkingDir) {
				When(workingDir.HasDiverged(Any[logging.SimpleLogging](), Any[string](), Any[string](), Any[[]string](), Any[models.PullRequest]())).ThenReturn(true)
			},
			wantFailure: "Default branch must be rebased onto pull request before running apply.",
			wantErr:     assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RegisterMockTestingT(t)
			workingDir := mocks.NewMockWorkingDir()
			a := &events.DefaultCommandRequirementHandler{WorkingDir: workingDir}
			if tt.setup != nil {
				tt.setup(workingDir)
			}
			gotFailure, err := a.ValidateApplyProject(repoDir, tt.ctx)
			if !tt.wantErr(t, err, fmt.Sprintf("ValidateApplyProject(%v, %v)", repoDir, tt.ctx)) {
				return
			}
			assert.Equalf(t, tt.wantFailure, gotFailure, "ValidateApplyProject(%v, %v)", repoDir, tt.ctx)
		})
	}
}

// TestAggregateApplyRequirements_MergeableScopedToProject verifies that the
// mergeable apply requirement is evaluated per project: a failing Atlantis plan
// status for a different project in the same pull request must not block
// applying a project whose own plan succeeded, while the project's own plan,
// external CI, MR-level signals, and lone aggregate plan failures still block.
func TestAggregateApplyRequirements_MergeableScopedToProject(t *testing.T) {
	repoDir := "repoDir"
	const vcsStatusName = "atlantis"
	mergeableReq := []string{raw.MergeableRequirement}
	longProjectName := strings.Repeat("very-long-project-name-", 20)
	truncatedOwnPlanStatus := truncateStatusContextForTest(fmt.Sprintf("%s/plan: %s", vcsStatusName, longProjectName))

	tests := []struct {
		name          string
		vcsStatusName string
		ctx           command.ProjectContext
		wantFailure   string
	}{
		{
			name:          "other project's plan failure does not block this project's apply",
			vcsStatusName: vcsStatusName,
			ctx: command.ProjectContext{
				ProjectName:       "app",
				ApplyRequirements: mergeableReq,
				PullReqStatus: models.PullReqStatus{
					MergeableStatus: models.MergeableStatus{
						IsMergeable:      false,
						Reason:           "Pipeline atlantis/plan: network has status failed",
						BlockingStatuses: []string{"atlantis/plan: network"},
					},
				},
			},
			wantFailure: "",
		},
		{
			name:          "combined plan status with another project's plan failure does not block this project's apply",
			vcsStatusName: vcsStatusName,
			ctx: command.ProjectContext{
				ProjectName:       "app",
				ApplyRequirements: mergeableReq,
				PullReqStatus: models.PullReqStatus{
					MergeableStatus: models.MergeableStatus{
						IsMergeable:      false,
						Reason:           "Pipeline atlantis/plan has status failed",
						BlockingStatuses: []string{"atlantis/plan", "atlantis/plan: network"},
					},
				},
			},
			wantFailure: "",
		},
		{
			name:          "this project's own plan failure still blocks",
			vcsStatusName: vcsStatusName,
			ctx: command.ProjectContext{
				ProjectName:       "app",
				ApplyRequirements: mergeableReq,
				PullReqStatus: models.PullReqStatus{
					MergeableStatus: models.MergeableStatus{
						IsMergeable:      false,
						Reason:           "Pipeline atlantis/plan: app has status failed",
						BlockingStatuses: []string{"atlantis/plan: app"},
					},
				},
			},
			wantFailure: "Pull request must be mergeable before running apply (Pipeline atlantis/plan: app has status failed).",
		},
		{
			name:          "combined plan status alone still blocks a per-project apply",
			vcsStatusName: vcsStatusName,
			ctx: command.ProjectContext{
				ProjectName:       "app",
				ApplyRequirements: mergeableReq,
				PullReqStatus: models.PullReqStatus{
					MergeableStatus: models.MergeableStatus{
						IsMergeable:      false,
						Reason:           "Pipeline atlantis/plan has status failed",
						BlockingStatuses: []string{"atlantis/plan"},
					},
				},
			},
			wantFailure: "Pull request must be mergeable before running apply (Pipeline atlantis/plan has status failed).",
		},
		{
			name:          "external CI failure still blocks",
			vcsStatusName: vcsStatusName,
			ctx: command.ProjectContext{
				ProjectName:       "app",
				ApplyRequirements: mergeableReq,
				PullReqStatus: models.PullReqStatus{
					MergeableStatus: models.MergeableStatus{
						IsMergeable:      false,
						Reason:           "Pipeline ci/build has status failed",
						BlockingStatuses: []string{"ci/build"},
					},
				},
			},
			wantFailure: "Pull request must be mergeable before running apply (Pipeline ci/build has status failed).",
		},
		{
			name:          "other project's plan plus this project's own plan still blocks",
			vcsStatusName: vcsStatusName,
			ctx: command.ProjectContext{
				ProjectName:       "app",
				ApplyRequirements: mergeableReq,
				PullReqStatus: models.PullReqStatus{
					MergeableStatus: models.MergeableStatus{
						IsMergeable:      false,
						Reason:           "Pipeline atlantis/plan: network has status failed",
						BlockingStatuses: []string{"atlantis/plan: network", "atlantis/plan: app"},
					},
				},
			},
			wantFailure: "Pull request must be mergeable before running apply (Pipeline atlantis/plan: network has status failed).",
		},
		{
			name:          "other project's plan plus external CI still blocks",
			vcsStatusName: vcsStatusName,
			ctx: command.ProjectContext{
				ProjectName:       "app",
				ApplyRequirements: mergeableReq,
				PullReqStatus: models.PullReqStatus{
					MergeableStatus: models.MergeableStatus{
						IsMergeable:      false,
						Reason:           "Pipeline atlantis/plan: network has status failed",
						BlockingStatuses: []string{"atlantis/plan: network", "ci/build"},
					},
				},
			},
			wantFailure: "Pull request must be mergeable before running apply (Pipeline atlantis/plan: network has status failed).",
		},
		{
			name:          "not mergeable with no blocking statuses (e.g. approvals) still blocks",
			vcsStatusName: vcsStatusName,
			ctx: command.ProjectContext{
				ProjectName:       "app",
				ApplyRequirements: mergeableReq,
				PullReqStatus: models.PullReqStatus{
					MergeableStatus: models.MergeableStatus{
						IsMergeable: false,
						Reason:      "Still require 2 approvals",
					},
				},
			},
			wantFailure: "Pull request must be mergeable before running apply (Still require 2 approvals).",
		},
		{
			name:          "project without a name ignores another project's plan failure",
			vcsStatusName: vcsStatusName,
			ctx: command.ProjectContext{
				RepoRelDir:        "infra/staging",
				Workspace:         "default",
				ApplyRequirements: mergeableReq,
				PullReqStatus: models.PullReqStatus{
					MergeableStatus: models.MergeableStatus{
						IsMergeable:      false,
						Reason:           "Pipeline atlantis/plan: infra/production/default has status failed",
						BlockingStatuses: []string{"atlantis/plan: infra/production/default"},
					},
				},
			},
			wantFailure: "",
		},
		{
			name:          "project without a name is still blocked by its own plan",
			vcsStatusName: vcsStatusName,
			ctx: command.ProjectContext{
				RepoRelDir:        "infra/staging",
				Workspace:         "default",
				ApplyRequirements: mergeableReq,
				PullReqStatus: models.PullReqStatus{
					MergeableStatus: models.MergeableStatus{
						IsMergeable:      false,
						Reason:           "Pipeline atlantis/plan: infra/staging/default has status failed",
						BlockingStatuses: []string{"atlantis/plan: infra/staging/default"},
					},
				},
			},
			wantFailure: "Pull request must be mergeable before running apply (Pipeline atlantis/plan: infra/staging/default has status failed).",
		},
		{
			name:          "truncated own project plan status still blocks",
			vcsStatusName: vcsStatusName,
			ctx: command.ProjectContext{
				ProjectName:       longProjectName,
				ApplyRequirements: mergeableReq,
				PullReqStatus: models.PullReqStatus{
					MergeableStatus: models.MergeableStatus{
						IsMergeable:      false,
						Reason:           fmt.Sprintf("Pipeline %s has status failed", truncatedOwnPlanStatus),
						BlockingStatuses: []string{truncatedOwnPlanStatus},
					},
				},
			},
			wantFailure: fmt.Sprintf("Pull request must be mergeable before running apply (Pipeline %s has status failed).", truncatedOwnPlanStatus),
		},
		{
			name:          "empty vcs status name disables per-project scoping and stays blocked",
			vcsStatusName: "",
			ctx: command.ProjectContext{
				ProjectName:       "app",
				ApplyRequirements: mergeableReq,
				PullReqStatus: models.PullReqStatus{
					MergeableStatus: models.MergeableStatus{
						IsMergeable:      false,
						Reason:           "Pipeline atlantis/plan: network has status failed",
						BlockingStatuses: []string{"atlantis/plan: network"},
					},
				},
			},
			wantFailure: "Pull request must be mergeable before running apply (Pipeline atlantis/plan: network has status failed).",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RegisterMockTestingT(t)
			workingDir := mocks.NewMockWorkingDir()
			a := &events.DefaultCommandRequirementHandler{
				WorkingDir:    workingDir,
				VCSStatusName: tt.vcsStatusName,
			}
			gotFailure, err := a.ValidateApplyProject(repoDir, tt.ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.wantFailure, gotFailure)
		})
	}
}

func TestAggregateCommandRequirements_MergeableScopingOnlyAppliesToApply(t *testing.T) {
	repoDir := "repoDir"
	const vcsStatusName = "atlantis"
	mergeableReq := []string{raw.MergeableRequirement}
	blockedByOtherProjectPlan := models.PullReqStatus{
		MergeableStatus: models.MergeableStatus{
			IsMergeable:      false,
			Reason:           "Pipeline atlantis/plan: network has status failed",
			BlockingStatuses: []string{"atlantis/plan: network"},
		},
	}

	tests := []struct {
		name        string
		ctx         command.ProjectContext
		validate    func(*events.DefaultCommandRequirementHandler, string, command.ProjectContext) (string, error)
		wantFailure string
	}{
		{
			name: "plan command still requires MR-wide mergeability",
			ctx: command.ProjectContext{
				ProjectName:      "app",
				PlanRequirements: mergeableReq,
				PullReqStatus:    blockedByOtherProjectPlan,
			},
			validate:    (*events.DefaultCommandRequirementHandler).ValidatePlanProject,
			wantFailure: "Pull request must be mergeable before running plan (Pipeline atlantis/plan: network has status failed).",
		},
		{
			name: "import command still requires MR-wide mergeability",
			ctx: command.ProjectContext{
				ProjectName:        "app",
				ImportRequirements: mergeableReq,
				PullReqStatus:      blockedByOtherProjectPlan,
			},
			validate:    (*events.DefaultCommandRequirementHandler).ValidateImportProject,
			wantFailure: "Pull request must be mergeable before running import (Pipeline atlantis/plan: network has status failed).",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RegisterMockTestingT(t)
			workingDir := mocks.NewMockWorkingDir()
			a := &events.DefaultCommandRequirementHandler{
				WorkingDir:    workingDir,
				VCSStatusName: vcsStatusName,
			}
			gotFailure, err := tt.validate(a, repoDir, tt.ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.wantFailure, gotFailure)
		})
	}
}

func truncateStatusContextForTest(s string) string {
	const maxStatusContext = 255
	if len(s) <= maxStatusContext {
		return s
	}
	hash := sha256.Sum256([]byte(s))
	suffix := fmt.Sprintf("-%x", hash[:6])
	return s[:maxStatusContext-len(suffix)] + suffix
}

func TestRequirements_ValidateProjectDependencies(t *testing.T) {
	tests := []struct {
		name        string
		ctx         command.ProjectContext
		setup       func(workingDir *mocks.MockWorkingDir)
		wantFailure string
		wantErr     assert.ErrorAssertionFunc
	}{
		{
			name:    "pass no dependencies",
			ctx:     command.ProjectContext{},
			wantErr: assert.NoError,
		},
		{
			name: "pass all dependencies applied",
			ctx: command.ProjectContext{
				DependsOn: []string{"project1"},
				PullStatus: &models.PullStatus{
					Projects: []models.ProjectStatus{
						{
							ProjectName: "project1",
							Status:      models.AppliedPlanStatus,
						},
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "Fail all dependencies are not applied",
			ctx: command.ProjectContext{
				DependsOn: []string{"project1", "project2"},
				PullStatus: &models.PullStatus{
					Projects: []models.ProjectStatus{
						{
							ProjectName: "project1",
							Status:      models.PlannedPlanStatus,
						},
						{
							ProjectName: "project2",
							Status:      models.ErroredApplyStatus,
						},
					},
				},
			},
			wantFailure: "Can't apply your project unless you apply its dependencies: [project1]",
			wantErr:     assert.NoError,
		},
		{
			name: "Fail one of dependencies is not applied",
			ctx: command.ProjectContext{
				DependsOn: []string{"project1", "project2"},
				PullStatus: &models.PullStatus{
					Projects: []models.ProjectStatus{
						{
							ProjectName: "project1",
							Status:      models.AppliedPlanStatus,
						},
						{
							ProjectName: "project2",
							Status:      models.ErroredApplyStatus,
						},
					},
				},
			},
			wantFailure: "Can't apply your project unless you apply its dependencies: [project2]",
			wantErr:     assert.NoError,
		},
		{
			name: "Should not fail if one of dependencies is not applied but it has no changes to apply",
			ctx: command.ProjectContext{
				DependsOn: []string{"project1", "project2"},
				PullStatus: &models.PullStatus{
					Projects: []models.ProjectStatus{
						{
							ProjectName: "project1",
							Status:      models.AppliedPlanStatus,
						},
						{
							ProjectName: "project2",
							Status:      models.PlannedNoChangesPlanStatus,
						},
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "In the case of more than one dependency, should not continue to check dependencies if one of them is not in applied status",
			ctx: command.ProjectContext{
				DependsOn: []string{"project1", "project2"},
				PullStatus: &models.PullStatus{
					Projects: []models.ProjectStatus{
						{
							ProjectName: "project1",
							Status:      models.AppliedPlanStatus,
						},
						{
							ProjectName: "project2",
							Status:      models.ErroredApplyStatus,
						},
						{
							ProjectName: "project3",
							Status:      models.PlannedPlanStatus,
						},
					},
				},
			},
			wantFailure: "Can't apply your project unless you apply its dependencies: [project2]",
			wantErr:     assert.NoError,
		},
		{
			name: "Fail missing dependency when strict dependency status is required",
			ctx: command.ProjectContext{
				DependsOn:                 []string{"project1"},
				FailOnMissingDependencies: true,
				PullStatus: &models.PullStatus{
					Projects: []models.ProjectStatus{},
				},
			},
			wantFailure: "Can't apply your project unless you apply its dependencies: [project1]",
			wantErr:     assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RegisterMockTestingT(t)
			workingDir := mocks.NewMockWorkingDir()
			a := &events.DefaultCommandRequirementHandler{WorkingDir: workingDir}
			gotFailure, err := a.ValidateProjectDependencies(tt.ctx)
			if !tt.wantErr(t, err, fmt.Sprintf("ValidateProjectDependencies(%v)", tt.ctx)) {
				return
			}
			assert.Equalf(t, tt.wantFailure, gotFailure, "ValidateProjectDependencies(%v)", tt.ctx)
		})
	}
}

func TestAggregateApplyRequirements_ValidateImportProject(t *testing.T) {
	repoDir := "repoDir"
	fullRequirements := []string{
		raw.ApprovedRequirement,
		raw.MergeableRequirement,
		raw.UnDivergedRequirement,
	}
	tests := []struct {
		name        string
		ctx         command.ProjectContext
		setup       func(workingDir *mocks.MockWorkingDir)
		wantFailure string
		wantErr     assert.ErrorAssertionFunc
	}{
		{
			name:    "pass no requirements",
			ctx:     command.ProjectContext{},
			wantErr: assert.NoError,
		},
		{
			name: "pass full requirements",
			ctx: command.ProjectContext{
				ImportRequirements: fullRequirements,
				PullReqStatus: models.PullReqStatus{
					ApprovalStatus:  models.ApprovalStatus{IsApproved: true},
					MergeableStatus: models.MergeableStatus{IsMergeable: true},
				},
				ProjectPlanStatus: models.PassedPolicyCheckStatus,
			},
			setup: func(workingDir *mocks.MockWorkingDir) {
				When(workingDir.HasDiverged(Any[logging.SimpleLogging](), Any[string](), Any[string](), Any[[]string](), Any[models.PullRequest]())).ThenReturn(false)
			},
			wantErr: assert.NoError,
		},
		{
			name: "fail by no approved",
			ctx: command.ProjectContext{
				ImportRequirements: []string{raw.ApprovedRequirement},
				PullReqStatus: models.PullReqStatus{
					ApprovalStatus: models.ApprovalStatus{IsApproved: false},
				},
			},
			wantFailure: "Pull request must be approved according to the project's approval rules before running import.",
			wantErr:     assert.NoError,
		},
		{
			name: "fail by no mergeable",
			ctx: command.ProjectContext{
				ImportRequirements: []string{raw.MergeableRequirement},
				PullReqStatus: models.PullReqStatus{
					MergeableStatus: models.MergeableStatus{IsMergeable: false},
				},
			},
			wantFailure: "Pull request must be mergeable before running import.",
			wantErr:     assert.NoError,
		},
		{
			name: "fail by diverged",
			ctx: command.ProjectContext{
				ImportRequirements: []string{raw.UnDivergedRequirement},
			},
			setup: func(workingDir *mocks.MockWorkingDir) {
				When(workingDir.HasDiverged(Any[logging.SimpleLogging](), Any[string](), Any[string](), Any[[]string](), Any[models.PullRequest]())).ThenReturn(true)
			},
			wantFailure: "Default branch must be rebased onto pull request before running import.",
			wantErr:     assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RegisterMockTestingT(t)
			workingDir := mocks.NewMockWorkingDir()
			a := &events.DefaultCommandRequirementHandler{WorkingDir: workingDir}
			if tt.setup != nil {
				tt.setup(workingDir)
			}
			gotFailure, err := a.ValidateImportProject(repoDir, tt.ctx)
			if !tt.wantErr(t, err, fmt.Sprintf("ValidateImportProject(%v, %v)", repoDir, tt.ctx)) {
				return
			}
			assert.Equalf(t, tt.wantFailure, gotFailure, "ValidateImportProject(%v, %v)", repoDir, tt.ctx)
		})
	}
}
