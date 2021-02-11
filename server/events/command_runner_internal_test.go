package events

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

func TestApplyUpdateCommitStatus(t *testing.T) {
	cases := map[string]struct {
		cmd           models.CommandName
		pullStatus    models.PullStatus
		expStatus     models.CommitStatus
		expNumSuccess int
		expNumTotal   int
	}{
		"apply, one pending": {
			cmd: models.ApplyCommand,
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
			cmd: models.ApplyCommand,
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
			cmd: models.ApplyCommand,
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
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			csu := &MockCSU{}
			cr := &ApplyCommandRunner{
				commitStatusUpdater: csu,
			}
			cr.updateCommitStatus(&CommandContext{}, c.pullStatus)
			Equals(t, models.Repo{}, csu.CalledRepo)
			Equals(t, models.PullRequest{}, csu.CalledPull)
			Equals(t, c.expStatus, csu.CalledStatus)
			Equals(t, c.cmd, csu.CalledCommand)
			Equals(t, c.expNumSuccess, csu.CalledNumSuccess)
			Equals(t, c.expNumTotal, csu.CalledNumTotal)
		})
	}
}

func TestPlanUpdateCommitStatus(t *testing.T) {
	cases := map[string]struct {
		cmd           models.CommandName
		pullStatus    models.PullStatus
		expStatus     models.CommitStatus
		expNumSuccess int
		expNumTotal   int
	}{
		"single plan success": {
			cmd: models.PlanCommand,
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
			cmd: models.PlanCommand,
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
			cr.updateCommitStatus(&CommandContext{}, c.pullStatus)
			Equals(t, models.Repo{}, csu.CalledRepo)
			Equals(t, models.PullRequest{}, csu.CalledPull)
			Equals(t, c.expStatus, csu.CalledStatus)
			Equals(t, c.cmd, csu.CalledCommand)
			Equals(t, c.expNumSuccess, csu.CalledNumSuccess)
			Equals(t, c.expNumTotal, csu.CalledNumTotal)
		})
	}
}

type MockCSU struct {
	CalledRepo       models.Repo
	CalledPull       models.PullRequest
	CalledStatus     models.CommitStatus
	CalledCommand    models.CommandName
	CalledNumSuccess int
	CalledNumTotal   int
}

func (m *MockCSU) UpdateCombinedCount(repo models.Repo, pull models.PullRequest, status models.CommitStatus, command models.CommandName, numSuccess int, numTotal int) error {
	m.CalledRepo = repo
	m.CalledPull = pull
	m.CalledStatus = status
	m.CalledCommand = command
	m.CalledNumSuccess = numSuccess
	m.CalledNumTotal = numTotal
	return nil
}
func (m *MockCSU) UpdateCombined(repo models.Repo, pull models.PullRequest, status models.CommitStatus, command models.CommandName) error {
	return nil
}
func (m *MockCSU) UpdateProject(ctx models.ProjectCommandContext, cmdName models.CommandName, status models.CommitStatus, url string) error {
	return nil
}
