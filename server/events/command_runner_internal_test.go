// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestApplyUpdateCommitStatus(t *testing.T) {
	cases := map[string]struct {
		cmd             command.Name
		pullStatus      models.PullStatus
		expStatus       models.CommitStatus
		expNumSuccess   int
		expNumTotal     int
		expNumErrored   int
		expNumNoChanges int
	}{
		"apply, one pending": {
			cmd: command.Apply,
			pullStatus: models.PullStatus{
				Projects: []models.ProjectStatus{
					{
						Status: models.PlannedPlanStatus,
					},
					{
						Status: models.AppliedPlanStatus,
					},
				},
			},
			expStatus:     models.PendingCommitStatus,
			expNumSuccess: 1,
			expNumTotal:   2,
		},
		"apply, all successful": {
			cmd: command.Apply,
			pullStatus: models.PullStatus{
				Projects: []models.ProjectStatus{
					{
						Status: models.AppliedPlanStatus,
					},
					{
						Status: models.AppliedPlanStatus,
					},
				},
			},
			expStatus:     models.SuccessCommitStatus,
			expNumSuccess: 2,
			expNumTotal:   2,
		},
		"apply, one errored, one pending": {
			cmd: command.Apply,
			pullStatus: models.PullStatus{
				Projects: []models.ProjectStatus{
					{
						Status: models.AppliedPlanStatus,
					},
					{
						Status: models.ErroredApplyStatus,
					},
					{
						Status: models.PlannedPlanStatus,
					},
				},
			},
			expStatus:     models.FailedCommitStatus,
			expNumSuccess: 1,
			expNumTotal:   3,
			expNumErrored: 1,
		},
		"apply, one planned no changes": {
			cmd: command.Apply,
			pullStatus: models.PullStatus{
				Projects: []models.ProjectStatus{
					{
						Status: models.AppliedPlanStatus,
					},
					{
						Status: models.PlannedNoChangesPlanStatus,
					},
				},
			},
			expStatus:       models.SuccessCommitStatus,
			expNumSuccess:   2,
			expNumTotal:     2,
			expNumNoChanges: 1,
		},
		"apply, one planned no changes and one pending": {
			cmd: command.Apply,
			pullStatus: models.PullStatus{
				Projects: []models.ProjectStatus{
					{
						Status: models.PlannedNoChangesPlanStatus,
					},
					{
						Status: models.PlannedPlanStatus,
					},
				},
			},
			expStatus:       models.PendingCommitStatus,
			expNumSuccess:   1,
			expNumTotal:     2,
			expNumNoChanges: 1,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			csu := &MockCSU{}
			cr := &ApplyCommandRunner{
				commitStatusUpdater: csu,
			}
			cr.updateCommitStatus(&command.Context{}, c.pullStatus)
			Equals(t, models.Repo{}, csu.CalledRepo)
			Equals(t, models.PullRequest{}, csu.CalledPull)
			Equals(t, c.expStatus, csu.CalledStatus)
			Equals(t, c.cmd, csu.CalledCommand)
			Equals(t, models.ProjectCounts{Success: c.expNumSuccess, Total: c.expNumTotal, Errored: c.expNumErrored, NoChanges: c.expNumNoChanges}, csu.CalledCounts)
		})
	}
}

func TestPlanUpdatePlanCommitStatus(t *testing.T) {
	cases := map[string]struct {
		cmd           command.Name
		pullStatus    models.PullStatus
		expStatus     models.CommitStatus
		expNumSuccess int
		expNumTotal   int
		expNumErrored int
	}{
		"single plan success": {
			cmd: command.Plan,
			pullStatus: models.PullStatus{
				Projects: []models.ProjectStatus{
					{
						Status: models.PlannedPlanStatus,
					},
				},
			},
			expStatus:     models.SuccessCommitStatus,
			expNumSuccess: 1,
			expNumTotal:   1,
		},
		"one plan error, other errors": {
			cmd: command.Plan,
			pullStatus: models.PullStatus{
				Projects: []models.ProjectStatus{
					{
						Status: models.ErroredPlanStatus,
					},
					{
						Status: models.PlannedPlanStatus,
					},
					{
						Status: models.AppliedPlanStatus,
					},
					{
						Status: models.ErroredApplyStatus,
					},
				},
			},
			expStatus:     models.FailedCommitStatus,
			expNumSuccess: 3,
			expNumTotal:   4,
			expNumErrored: 1,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			csu := &MockCSU{}
			cr := &PlanCommandRunner{
				commitStatusUpdater: csu,
			}
			cr.updateCommitStatus(&command.Context{}, c.pullStatus, command.Plan)
			Equals(t, models.Repo{}, csu.CalledRepo)
			Equals(t, models.PullRequest{}, csu.CalledPull)
			Equals(t, c.expStatus, csu.CalledStatus)
			Equals(t, c.cmd, csu.CalledCommand)
			Equals(t, models.ProjectCounts{Success: c.expNumSuccess, Total: c.expNumTotal, Errored: c.expNumErrored}, csu.CalledCounts)
		})
	}
}

func TestPlanUpdateApplyCommitStatus(t *testing.T) {
	cases := map[string]struct {
		cmd                  command.Name
		pullStatus           models.PullStatus
		expStatus            models.CommitStatus
		doNotCallUpdateApply bool // In certain situations, we don't expect updateCommitStatus to call the underlying commitStatusUpdater code at all
		expNumSuccess        int
		expNumTotal          int
		expNumErrored        int
		expNumNoChanges      int
	}{
		"all plans success with no changes": {
			cmd: command.Apply,
			pullStatus: models.PullStatus{
				Projects: []models.ProjectStatus{
					{
						Status: models.PlannedNoChangesPlanStatus,
					},
					{
						Status: models.PlannedNoChangesPlanStatus,
					},
				},
			},
			expStatus:       models.SuccessCommitStatus,
			expNumSuccess:   2,
			expNumTotal:     2,
			expNumNoChanges: 2,
		},
		"one plan, one plan success with no changes": {
			cmd: command.Apply,
			pullStatus: models.PullStatus{
				Projects: []models.ProjectStatus{
					{
						Status: models.PlannedNoChangesPlanStatus,
					},
					{
						Status: models.PlannedPlanStatus,
					},
				},
			},
			doNotCallUpdateApply: true,
		},
		"one plan, one apply, one plan success with no changes": {
			cmd: command.Apply,
			pullStatus: models.PullStatus{
				Projects: []models.ProjectStatus{
					{
						Status: models.PlannedNoChangesPlanStatus,
					},
					{
						Status: models.AppliedPlanStatus,
					},
					{
						Status: models.PlannedPlanStatus,
					},
				},
			},
			doNotCallUpdateApply: true,
		},
		"one apply error, one apply, one plan success with no changes": {
			cmd: command.Apply,
			pullStatus: models.PullStatus{
				Projects: []models.ProjectStatus{
					{
						Status: models.PlannedNoChangesPlanStatus,
					},
					{
						Status: models.AppliedPlanStatus,
					},
					{
						Status: models.ErroredApplyStatus,
					},
				},
			},
			expStatus:       models.FailedCommitStatus,
			expNumSuccess:   2,
			expNumTotal:     3,
			expNumErrored:   1,
			expNumNoChanges: 1,
		},
		"one apply error, one pending plan": {
			cmd: command.Apply,
			pullStatus: models.PullStatus{
				Projects: []models.ProjectStatus{
					{
						Status: models.ErroredApplyStatus,
					},
					{
						Status: models.AppliedPlanStatus,
					},
					{
						Status: models.PlannedPlanStatus,
					},
				},
			},
			expStatus:     models.FailedCommitStatus,
			expNumSuccess: 1,
			expNumTotal:   3,
			expNumErrored: 1,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			csu := &MockCSU{}
			cr := &PlanCommandRunner{
				commitStatusUpdater: csu,
			}
			cr.updateCommitStatus(&command.Context{}, c.pullStatus, command.Apply)
			if c.doNotCallUpdateApply {
				Equals(t, csu.Called, false)
			} else {
				Equals(t, csu.Called, true)
				Equals(t, models.Repo{}, csu.CalledRepo)
				Equals(t, models.PullRequest{}, csu.CalledPull)
				Equals(t, c.expStatus, csu.CalledStatus)
				Equals(t, c.cmd, csu.CalledCommand)
				Equals(t, models.ProjectCounts{Success: c.expNumSuccess, Total: c.expNumTotal, Errored: c.expNumErrored, NoChanges: c.expNumNoChanges}, csu.CalledCounts)
			}
		})
	}
}

func TestPlanUpdateApplyCommitStatus_GitLabPendingWithNoChanges(t *testing.T) {
	csu := &MockCSU{}
	runner := &PlanCommandRunner{
		commitStatusUpdater: csu,
		PendingApplyStatus:  true,
	}
	pullStatus := models.PullStatus{
		Projects: []models.ProjectStatus{
			{
				Status: models.PlannedNoChangesPlanStatus,
			},
			{
				Status: models.PlannedPlanStatus,
			},
		},
	}
	ctx := &command.Context{
		Log: logging.NewNoopLogger(t),
		Pull: models.PullRequest{
			BaseRepo: models.Repo{
				VCSHost: models.VCSHost{
					Type: models.Gitlab,
				},
			},
		},
	}

	runner.updateCommitStatus(ctx, pullStatus, command.Apply)

	Equals(t, true, csu.Called)
	Equals(t, models.PendingCommitStatus, csu.CalledStatus)
	Equals(t, command.Apply, csu.CalledCommand)
	Equals(t, models.ProjectCounts{Success: 1, Total: 2, NoChanges: 1}, csu.CalledCounts)
}

func TestPolicyCheckUpdateCommitStatus(t *testing.T) {
	cases := map[string]struct {
		pullStatus    models.PullStatus
		expStatus     models.CommitStatus
		expNumSuccess int
		expNumTotal   int
		expNumErrored int
	}{
		"all policy checks passed": {
			pullStatus: models.PullStatus{
				Projects: []models.ProjectStatus{
					{
						Status: models.PassedPolicyCheckStatus,
					},
					{
						Status: models.PassedPolicyCheckStatus,
					},
				},
			},
			expStatus:     models.SuccessCommitStatus,
			expNumSuccess: 2,
			expNumTotal:   2,
		},
		"one policy check failed, one untouched project": {
			pullStatus: models.PullStatus{
				Projects: []models.ProjectStatus{
					{
						Status: models.ErroredPolicyCheckStatus,
					},
					{
						Status: models.PlannedPlanStatus,
					},
				},
			},
			expStatus:     models.FailedCommitStatus,
			expNumTotal:   2,
			expNumErrored: 1,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			csu := &MockCSU{}
			runner := &PolicyCheckCommandRunner{
				commitStatusUpdater: csu,
			}
			runner.updateCommitStatus(&command.Context{}, c.pullStatus)
			Equals(t, models.Repo{}, csu.CalledRepo)
			Equals(t, models.PullRequest{}, csu.CalledPull)
			Equals(t, c.expStatus, csu.CalledStatus)
			Equals(t, command.PolicyCheck, csu.CalledCommand)
			Equals(t, models.ProjectCounts{Success: c.expNumSuccess, Total: c.expNumTotal, Errored: c.expNumErrored}, csu.CalledCounts)
		})
	}
}

func TestApprovePoliciesUpdateCommitStatus(t *testing.T) {
	pullStatus := models.PullStatus{
		Projects: []models.ProjectStatus{
			{
				Status: models.ErroredPolicyCheckStatus,
			},
			{
				Status: models.PlannedPlanStatus,
			},
		},
	}
	csu := &MockCSU{}
	runner := &ApprovePoliciesCommandRunner{
		commitStatusUpdater: csu,
	}

	runner.updateCommitStatus(&command.Context{}, pullStatus)

	Equals(t, models.FailedCommitStatus, csu.CalledStatus)
	Equals(t, command.PolicyCheck, csu.CalledCommand)
	Equals(t, models.ProjectCounts{Total: 2, Errored: 1}, csu.CalledCounts)
}

type MockCSU struct {
	CalledRepo    models.Repo
	CalledPull    models.PullRequest
	CalledStatus  models.CommitStatus
	CalledCommand command.Name
	CalledCounts  models.ProjectCounts
	Called        bool
}

func (m *MockCSU) UpdateCombinedCount(_ logging.SimpleLogging, repo models.Repo, pull models.PullRequest, status models.CommitStatus, command command.Name, counts models.ProjectCounts) error {
	m.Called = true
	m.CalledRepo = repo
	m.CalledPull = pull
	m.CalledStatus = status
	m.CalledCommand = command
	m.CalledCounts = counts
	return nil
}

func (m *MockCSU) UpdateCombined(_ logging.SimpleLogging, _ models.Repo, _ models.PullRequest, _ models.CommitStatus, _ command.Name) error {
	return nil
}

func (m *MockCSU) UpdateProject(_ command.ProjectContext, _ command.Name, _ models.CommitStatus, _ string, _ *command.ProjectResult) error {
	return nil
}

func (m *MockCSU) UpdatePreWorkflowHook(_ logging.SimpleLogging, _ models.PullRequest, _ models.CommitStatus, _ string, _ string, _ string) error {
	return nil
}

func (m *MockCSU) UpdatePostWorkflowHook(_ logging.SimpleLogging, _ models.PullRequest, _ models.CommitStatus, _ string, _ string, _ string) error {
	return nil
}
