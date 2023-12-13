package events_test

import (
	"fmt"
	"testing"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/config/raw"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/models"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/stretchr/testify/assert"
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
					ApprovalStatus: models.ApprovalStatus{IsApproved: true},
					Mergeable:      true,
				},
				ProjectPlanStatus: models.PassedPolicyCheckStatus,
			},
			setup: func(workingDir *mocks.MockWorkingDir) {
				When(workingDir.HasDiverged(Any[string]())).ThenReturn(false)
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
				PullReqStatus:    models.PullReqStatus{Mergeable: false},
			},
			wantFailure: "Pull request must be mergeable before running plan.",
			wantErr:     assert.NoError,
		},
		{
			name: "fail by diverged",
			ctx: command.ProjectContext{
				PlanRequirements: []string{raw.UnDivergedRequirement},
			},
			setup: func(workingDir *mocks.MockWorkingDir) {
				When(workingDir.HasDiverged(Any[string]())).ThenReturn(true)
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
			name: "pass full requirements",
			ctx: command.ProjectContext{
				ApplyRequirements: fullRequirements,
				PullReqStatus: models.PullReqStatus{
					ApprovalStatus: models.ApprovalStatus{IsApproved: true},
					Mergeable:      true,
				},
				ProjectPlanStatus: models.PassedPolicyCheckStatus,
			},
			setup: func(workingDir *mocks.MockWorkingDir) {
				When(workingDir.HasDiverged(Any[string]())).ThenReturn(false)
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
						Approvals:     0,
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
				PullReqStatus:     models.PullReqStatus{Mergeable: false},
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
				When(workingDir.HasDiverged(Any[string]())).ThenReturn(true)
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
					ApprovalStatus: models.ApprovalStatus{IsApproved: true},
					Mergeable:      true,
				},
				ProjectPlanStatus: models.PassedPolicyCheckStatus,
			},
			setup: func(workingDir *mocks.MockWorkingDir) {
				When(workingDir.HasDiverged(Any[string]())).ThenReturn(false)
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
				PullReqStatus:      models.PullReqStatus{Mergeable: false},
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
				When(workingDir.HasDiverged(Any[string]())).ThenReturn(true)
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
