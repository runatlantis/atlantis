// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"errors"
	"testing"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

// childTeamVCSClient combines vcs.NotConfiguredVCSClient (satisfying vcs.Client)
// with a configurable team hierarchy map, so tests can exercise GetChildTeams
// without a real VCS connection.
type childTeamVCSClient struct {
	vcs.NotConfiguredVCSClient
	children map[string][]string
	errOn    map[string]bool
}

func (c *childTeamVCSClient) GetChildTeams(_ logging.SimpleLogging, _ models.Repo, teamSlug string) ([]string, error) {
	if c.errOn[teamSlug] {
		return nil, errors.New("API error for " + teamSlug)
	}
	return c.children[teamSlug], nil
}

func TestFetchDescendantTeams(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	repo := models.Repo{Owner: "test-org"}

	t.Run("leaf team returns empty", func(t *testing.T) {
		fetcher := &childTeamVCSClient{children: map[string][]string{}}
		result, err := fetchDescendantTeams(fetcher, logger, repo, "leaf-team", 20)
		Ok(t, err)
		Equals(t, 0, len(result))
	})

	t.Run("single level of children", func(t *testing.T) {
		fetcher := &childTeamVCSClient{children: map[string][]string{
			"parent": {"child-a", "child-b"},
		}}
		result, err := fetchDescendantTeams(fetcher, logger, repo, "parent", 20)
		Ok(t, err)
		Equals(t, []string{"child-a", "child-b"}, result)
	})

	t.Run("multiple levels of nesting", func(t *testing.T) {
		fetcher := &childTeamVCSClient{children: map[string][]string{
			"grandparent": {"parent"},
			"parent":      {"child"},
		}}
		result, err := fetchDescendantTeams(fetcher, logger, repo, "grandparent", 20)
		Ok(t, err)
		Equals(t, []string{"parent", "child"}, result)
	})

	t.Run("maxDepth=0 returns nothing", func(t *testing.T) {
		fetcher := &childTeamVCSClient{children: map[string][]string{
			"parent": {"child"},
		}}
		result, err := fetchDescendantTeams(fetcher, logger, repo, "parent", 0)
		Ok(t, err)
		Equals(t, []string(nil), result)
	})

	t.Run("maxDepth=1 returns only direct children", func(t *testing.T) {
		fetcher := &childTeamVCSClient{children: map[string][]string{
			"grandparent": {"parent"},
			"parent":      {"child"},
		}}
		result, err := fetchDescendantTeams(fetcher, logger, repo, "grandparent", 1)
		Ok(t, err)
		Equals(t, []string{"parent"}, result)
	})

	t.Run("error at root propagates", func(t *testing.T) {
		fetcher := &childTeamVCSClient{
			children: map[string][]string{},
			errOn:    map[string]bool{"parent": true},
		}
		_, err := fetchDescendantTeams(fetcher, logger, repo, "parent", 20)
		Assert(t, err != nil, "expected error to propagate from root team")
	})

	t.Run("error in recursive call is logged and skipped", func(t *testing.T) {
		// parent fetches OK; fetching child-a's children errors.
		// child-b and its subtree should still be traversed.
		fetcher := &childTeamVCSClient{
			children: map[string][]string{
				"parent":  {"child-a", "child-b"},
				"child-b": {"grandchild-b"},
			},
			errOn: map[string]bool{"child-a": true},
		}
		result, err := fetchDescendantTeams(fetcher, logger, repo, "parent", 20)
		Ok(t, err)
		// child-a is included (direct child of parent); its subtree is skipped on error.
		// child-b and grandchild-b are also traversed successfully.
		Equals(t, []string{"child-a", "child-b", "grandchild-b"}, result)
	})

	t.Run("cycle is handled without infinite loop", func(t *testing.T) {
		// a -> b -> a forms a cycle; visited set should break it.
		fetcher := &childTeamVCSClient{children: map[string][]string{
			"team-a": {"team-b"},
			"team-b": {"team-a"},
		}}
		result, err := fetchDescendantTeams(fetcher, logger, repo, "team-a", 20)
		Ok(t, err)
		Equals(t, []string{"team-b"}, result)
	})
}

func TestCheckUserPermissions(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	repo := models.Repo{Owner: "test-org"}

	t.Run("no rules allows everyone", func(t *testing.T) {
		cr := &DefaultCommandRunner{
			Logger:               logger,
			TeamAllowlistChecker: &command.DefaultTeamAllowlistChecker{},
		}
		user := models.User{Username: "alice", Teams: []string{"some-team"}}
		ok, err := cr.checkUserPermissions(repo, &user, "plan")
		Ok(t, err)
		Assert(t, ok, "expected allowed when no rules")
	})

	t.Run("fast path: direct team member is allowed", func(t *testing.T) {
		checker, _ := command.NewTeamAllowlistChecker("dev-team:plan")
		cr := &DefaultCommandRunner{
			Logger:               logger,
			TeamAllowlistChecker: checker,
			VCSClient:            &vcs.NotConfiguredVCSClient{}, // GetChildTeams returns nil (no hierarchy support)
		}
		user := models.User{Username: "alice", Teams: []string{"dev-team"}}
		ok, err := cr.checkUserPermissions(repo, &user, "plan")
		Ok(t, err)
		Assert(t, ok, "expected direct member to be allowed")
	})

	t.Run("fast path: non-member without hierarchy support is rejected", func(t *testing.T) {
		checker, _ := command.NewTeamAllowlistChecker("dev-team:plan")
		cr := &DefaultCommandRunner{
			Logger:               logger,
			TeamAllowlistChecker: checker,
			VCSClient:            &vcs.NotConfiguredVCSClient{}, // GetChildTeams returns nil (no hierarchy support)
		}
		user := models.User{Username: "alice", Teams: []string{"other-team"}}
		ok, err := cr.checkUserPermissions(repo, &user, "plan")
		Ok(t, err)
		Assert(t, !ok, "expected non-member to be rejected when no hierarchy support")
	})

	hierarchyCases := map[string]struct {
		allowlist   string
		userTeams   []string
		hierarchy   map[string][]string
		cmdName     string
		expectAllow bool
	}{
		"slow path: user in direct child team is allowed": {
			allowlist:   "parent-team:plan",
			userTeams:   []string{"child-team"},
			hierarchy:   map[string][]string{"parent-team": {"child-team"}},
			cmdName:     "plan",
			expectAllow: true,
		},
		"slow path: user in grandchild team is allowed": {
			allowlist: "grandparent-team:plan",
			userTeams: []string{"grandchild-team"},
			hierarchy: map[string][]string{
				"grandparent-team": {"parent-team"},
				"parent-team":      {"grandchild-team"},
			},
			cmdName:     "plan",
			expectAllow: true,
		},
		"slow path: user not in any descendant is rejected": {
			allowlist:   "parent-team:plan",
			userTeams:   []string{"unrelated-team"},
			hierarchy:   map[string][]string{"parent-team": {"child-team"}},
			cmdName:     "plan",
			expectAllow: false,
		},
		"slow path: user in child team but wrong command is rejected": {
			allowlist:   "parent-team:apply",
			userTeams:   []string{"child-team"},
			hierarchy:   map[string][]string{"parent-team": {"child-team"}},
			cmdName:     "plan",
			expectAllow: false,
		},
		// *:plan allows everyone including users with no team memberships.
		// The fast path handles this via IsCommandAllowedForAnyTeam's zero-team wildcard check.
		"fast path: wildcard team rule allows user with no teams": {
			allowlist:   "*:plan",
			userTeams:   []string{},
			hierarchy:   map[string][]string{},
			cmdName:     "plan",
			expectAllow: true,
		},
	}

	for name, tc := range hierarchyCases {
		t.Run(name, func(t *testing.T) {
			checker, err := command.NewTeamAllowlistChecker(tc.allowlist)
			Ok(t, err)
			cr := &DefaultCommandRunner{
				Logger:               logger,
				TeamAllowlistChecker: checker,
				VCSClient: &childTeamVCSClient{children: tc.hierarchy},
			}
			user := models.User{Username: "testuser", Teams: tc.userTeams}
			ok, checkErr := cr.checkUserPermissions(repo, &user, tc.cmdName)
			Ok(t, checkErr)
			Equals(t, tc.expectAllow, ok)
		})
	}

	t.Run("slow path: matched parent team is appended to user.Teams", func(t *testing.T) {
		checker, err := command.NewTeamAllowlistChecker("parent-team:plan")
		Ok(t, err)
		cr := &DefaultCommandRunner{
			Logger:               logger,
			TeamAllowlistChecker: checker,
			VCSClient: &childTeamVCSClient{
				children: map[string][]string{"parent-team": {"child-team"}},
			},
		}
		user := models.User{Username: "alice", Teams: []string{"child-team"}}
		ok, checkErr := cr.checkUserPermissions(repo, &user, "plan")
		Ok(t, checkErr)
		Assert(t, ok, "expected child team member to be allowed")
		// The matched parent team should be appended so per-project allowlist checks pass.
		Assert(t, len(user.Teams) == 2, "expected user.Teams to contain both child-team and parent-team")
		Equals(t, "parent-team", user.Teams[1])
	})
}

func TestApplyUpdateCommitStatus(t *testing.T) {
	cases := map[string]struct {
		cmd           command.Name
		pullStatus    models.PullStatus
		expStatus     models.CommitStatus
		expNumSuccess int
		expNumTotal   int
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
			expStatus:     models.SuccessCommitStatus,
			expNumSuccess: 2,
			expNumTotal:   2,
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
