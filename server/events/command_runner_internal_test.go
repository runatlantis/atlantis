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
		cmd                                        command.Name
		SetAtlantisApplyCheckSuccessfulIfNoChanges bool
		pullStatus                                 models.PullStatus
		expStatus                                  models.CommitStatus
		expNumSuccess                              int
		expNumTotal                                int
	}{
		"apply, one pending": {
			cmd: command.Apply,
			SetAtlantisApplyCheckSuccessfulIfNoChanges: true,
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
			SetAtlantisApplyCheckSuccessfulIfNoChanges: true,
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
			SetAtlantisApplyCheckSuccessfulIfNoChanges: true,
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
		},
		"apply, one planned no changes": {
			cmd: command.Apply,
			SetAtlantisApplyCheckSuccessfulIfNoChanges: false,
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
			expStatus:     models.PendingCommitStatus,
			expNumSuccess: 1,
			expNumTotal:   2,
		},
		"apply, one planned no changes, skip apply when no changes": {
			cmd: command.Apply,
			SetAtlantisApplyCheckSuccessfulIfNoChanges: true,
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
			expStatus:     models.SuccessCommitStatus,
			expNumSuccess: 2,
			expNumTotal:   2,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			csu := &MockCSU{}
			cr := &ApplyCommandRunner{
				commitStatusUpdater:                        csu,
				SetAtlantisApplyCheckSuccessfulIfNoChanges: c.SetAtlantisApplyCheckSuccessfulIfNoChanges,
			}
			cr.updateCommitStatus(&command.Context{}, c.pullStatus)
			Equals(t, models.Repo{}, csu.CalledRepo)
			Equals(t, models.PullRequest{}, csu.CalledPull)
			Equals(t, c.expStatus, csu.CalledStatus)
			Equals(t, c.cmd, csu.CalledCommand)
			Equals(t, c.expNumSuccess, csu.CalledNumSuccess)
			Equals(t, c.expNumTotal, csu.CalledNumTotal)
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
			Equals(t, c.expNumSuccess, csu.CalledNumSuccess)
			Equals(t, c.expNumTotal, csu.CalledNumTotal)
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
			expStatus:     models.SuccessCommitStatus,
			expNumSuccess: 2,
			expNumTotal:   2,
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
			expStatus:     models.FailedCommitStatus,
			expNumSuccess: 2,
			expNumTotal:   3,
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
				Equals(t, c.expNumSuccess, csu.CalledNumSuccess)
				Equals(t, c.expNumTotal, csu.CalledNumTotal)
			}
		})
	}
}

type MockCSU struct {
	CalledRepo       models.Repo
	CalledPull       models.PullRequest
	CalledStatus     models.CommitStatus
	CalledCommand    command.Name
	CalledNumSuccess int
	CalledNumTotal   int
	Called           bool
}

func (m *MockCSU) UpdateCombinedCount(_ logging.SimpleLogging, repo models.Repo, pull models.PullRequest, status models.CommitStatus, command command.Name, numSuccess int, numTotal int) error {
	m.Called = true
	m.CalledRepo = repo
	m.CalledPull = pull
	m.CalledStatus = status
	m.CalledCommand = command
	m.CalledNumSuccess = numSuccess
	m.CalledNumTotal = numTotal
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
