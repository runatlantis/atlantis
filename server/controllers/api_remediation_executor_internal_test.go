// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/locking"
	. "github.com/runatlantis/atlantis/server/core/locking/mocks"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	. "github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics/metricstest"
	. "github.com/runatlantis/atlantis/testing"
	"go.uber.org/mock/gomock"
)

func TestAPIRemediationExecutor_ExecuteApplyProjectsHonorsGlobalApplyLock(t *testing.T) {
	RegisterMockTestingT(t)
	gmockCtrl := gomock.NewController(t)
	logger := logging.NewNoopLogger(t)
	baseRepo := models.Repo{
		FullName: "owner/repo",
		VCSHost: models.VCSHost{
			Hostname: "github.com",
			Type:     models.Github,
		},
	}

	applyLockChecker := NewMockApplyLocker(gmockCtrl)
	applyLockChecker.EXPECT().CheckApplyLock().Return(locking.ApplyCommandLock{Locked: true}, nil)
	projectCommandBuilder := NewMockProjectCommandBuilder()
	projectCommandRunner := NewMockProjectCommandRunner()

	executor := &apiRemediationExecutor{
		controller: &APIController{
			ApplyLockChecker:          applyLockChecker,
			ProjectCommandBuilder:     projectCommandBuilder,
			ProjectPlanCommandRunner:  projectCommandRunner,
			ProjectApplyCommandRunner: projectCommandRunner,
		},
		baseRepo: baseRepo,
		logger:   logger,
	}

	results, err := executor.ExecuteApplyProjects("owner/repo", "main", "Github", []models.ProjectDrift{{
		ProjectName: "app",
		Path:        "app",
		Workspace:   events.DefaultWorkspace,
	}})

	Assert(t, err != nil, "expected apply lock error")
	Assert(t, strings.Contains(err.Error(), "apply is disabled globally"), "expected apply lock error, got %v", err)
	Equals(t, 0, len(results))
	projectCommandBuilder.VerifyWasCalled(Never()).BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())
	projectCommandBuilder.VerifyWasCalled(Never()).BuildApplyCommands(Any[*command.Context](), Any[*events.CommentCommand]())
	projectCommandRunner.VerifyWasCalled(Never()).Apply(Any[command.ProjectContext]())
}

func TestAPIRemediationExecutor_ExecuteApplyAbortsWhenPreApplyPlanHasErrors(t *testing.T) {
	RegisterMockTestingT(t)
	gmockCtrl := gomock.NewController(t)
	logger := logging.NewNoopLogger(t)
	baseRepo := models.Repo{
		FullName: "owner/repo",
		VCSHost: models.VCSHost{
			Hostname: "github.com",
			Type:     models.Github,
		},
	}

	locker := NewMockLocker(gmockCtrl)
	locker.EXPECT().UnlockByPull(baseRepo.FullName, gomock.Any()).Return(nil, nil).AnyTimes()
	applyLockChecker := NewMockApplyLocker(gmockCtrl)
	applyLockChecker.EXPECT().CheckApplyLock().Return(locking.ApplyCommandLock{}, nil)

	workingDirLocker := NewMockWorkingDirLocker()
	When(workingDirLocker.TryLock(Any[string](), Any[int](), Any[string](), Any[string](), Any[string](), Any[command.Name]())).
		ThenReturn(func() {}, nil)
	workingDir := NewMockWorkingDir()
	When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[string]())).
		ThenReturn(t.TempDir(), nil)

	projectCommandBuilder := NewMockProjectCommandBuilder()
	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		ThenReturn([]command.ProjectContext{{
			CommandName: command.Plan,
			ProjectName: "app",
			RepoRelDir:  ".",
			Workspace:   events.DefaultWorkspace,
		}}, nil)

	projectCommandRunner := NewMockProjectCommandRunner()
	When(projectCommandRunner.Plan(Any[command.ProjectContext]())).ThenReturn(command.ProjectCommandOutput{
		Error: errors.New("latest plan failed"),
	})

	preWorkflowHooksCommandRunner := NewMockPreWorkflowHooksCommandRunner()
	When(preWorkflowHooksCommandRunner.RunPreHooks(Any[*command.Context](), Any[*events.CommentCommand]())).ThenReturn(nil)
	postWorkflowHooksCommandRunner := NewMockPostWorkflowHooksCommandRunner()
	When(postWorkflowHooksCommandRunner.RunPostHooks(Any[*command.Context](), Any[*events.CommentCommand]())).ThenReturn(nil)
	commitStatusUpdater := NewMockCommitStatusUpdater()
	When(commitStatusUpdater.UpdateCombined(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[models.CommitStatus](), Any[command.Name]())).
		ThenReturn(nil)

	controller := &APIController{
		Locker:                         locker,
		ApplyLockChecker:               applyLockChecker,
		Logger:                         logger,
		Scope:                          metricstest.NewLoggingScope(t, logger, "null"),
		ProjectCommandBuilder:          projectCommandBuilder,
		ProjectPlanCommandRunner:       projectCommandRunner,
		ProjectApplyCommandRunner:      projectCommandRunner,
		PreWorkflowHooksCommandRunner:  preWorkflowHooksCommandRunner,
		PostWorkflowHooksCommandRunner: postWorkflowHooksCommandRunner,
		WorkingDir:                     workingDir,
		WorkingDirLocker:               workingDirLocker,
		CommitStatusUpdater:            commitStatusUpdater,
	}
	executor := &apiRemediationExecutor{
		controller: controller,
		baseRepo:   baseRepo,
		logger:     logger,
	}

	output, err := executor.ExecuteApply("owner/repo", "main", "Github", "app", ".", events.DefaultWorkspace)

	Assert(t, err != nil, "expected pre-apply plan error")
	Assert(t, strings.Contains(err.Error(), "plan had errors"), "expected plan error, got %v", err)
	Assert(t, strings.Contains(output, "latest plan failed"), "expected plan failure output, got %q", output)
	projectCommandBuilder.VerifyWasCalled(Never()).BuildApplyCommands(Any[*command.Context](), Any[*events.CommentCommand]())
	projectCommandRunner.VerifyWasCalled(Never()).Apply(Any[command.ProjectContext]())
}

func TestAPIRemediationExecutor_ExecuteApplyProjectsSeedsPullStatusForDependencies(t *testing.T) {
	RegisterMockTestingT(t)
	gmockCtrl := gomock.NewController(t)
	logger := logging.NewNoopLogger(t)
	baseRepo := models.Repo{
		FullName: "owner/repo",
		VCSHost: models.VCSHost{
			Hostname: "github.com",
			Type:     models.Github,
		},
	}

	locker := NewMockLocker(gmockCtrl)
	locker.EXPECT().UnlockByPull(baseRepo.FullName, gomock.Any()).Return(nil, nil).AnyTimes()
	applyLockChecker := NewMockApplyLocker(gmockCtrl)
	applyLockChecker.EXPECT().CheckApplyLock().Return(locking.ApplyCommandLock{}, nil)

	workingDirLocker := NewMockWorkingDirLocker()
	When(workingDirLocker.TryLock(Any[string](), Any[int](), Any[string](), Any[string](), Any[string](), Any[command.Name]())).
		ThenReturn(func() {}, nil)
	workingDir := NewMockWorkingDir()
	When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[string]())).
		ThenReturn(t.TempDir(), nil)
	When(workingDir.Delete(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest]())).
		ThenReturn(nil)

	projectCommandBuilder := NewMockProjectCommandBuilder()
	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		Then(func(args []Param) ReturnValues {
			cmd := args[1].(*events.CommentCommand)
			executionOrderGroup := 0
			if cmd.ProjectName == "app" {
				executionOrderGroup = 1
			}
			return ReturnValues{[]command.ProjectContext{{
				CommandName:         command.Plan,
				ProjectName:         cmd.ProjectName,
				RepoRelDir:          cmd.RepoRelDir,
				Workspace:           cmd.Workspace,
				ExecutionOrderGroup: executionOrderGroup,
			}}, nil}
		})

	var capturedPullStatus *models.PullStatus
	When(projectCommandBuilder.BuildApplyCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		Then(func(args []Param) ReturnValues {
			ctx := args[0].(*command.Context)
			cmd := args[1].(*events.CommentCommand)
			capturedPullStatus = ctx.PullStatus
			Assert(t, capturedPullStatus != nil, "expected pull status before building apply commands")
			Equals(t, 2, len(capturedPullStatus.Projects))
			Equals(t, models.PlannedPlanStatus, capturedPullStatus.Projects[0].Status)
			Equals(t, models.PlannedPlanStatus, capturedPullStatus.Projects[1].Status)

			executionOrderGroup := 0
			dependsOn := []string(nil)
			if cmd.ProjectName == "app" {
				executionOrderGroup = 1
				dependsOn = []string{"network"}
			}
			return ReturnValues{[]command.ProjectContext{{
				CommandName:         command.Apply,
				ProjectName:         cmd.ProjectName,
				RepoRelDir:          cmd.RepoRelDir,
				Workspace:           cmd.Workspace,
				DependsOn:           dependsOn,
				PullStatus:          capturedPullStatus,
				ExecutionOrderGroup: executionOrderGroup,
			}}, nil}
		})

	projectCommandRunner := NewMockProjectCommandRunner()
	When(projectCommandRunner.Plan(Any[command.ProjectContext]())).ThenReturn(command.ProjectCommandOutput{
		PlanSuccess: &models.PlanSuccess{TerraformOutput: "No changes."},
	})
	applyCalls := 0
	applyOrder := []string{}
	When(projectCommandRunner.Apply(Any[command.ProjectContext]())).Then(func(args []Param) ReturnValues {
		applyCalls++
		projectCtx := args[0].(command.ProjectContext)
		applyOrder = append(applyOrder, projectCtx.ProjectName)
		if projectCtx.ProjectName == "app" {
			Equals(t, models.AppliedPlanStatus, capturedPullStatus.Projects[0].Status)
		}
		return ReturnValues{command.ProjectCommandOutput{ApplySuccess: "success"}}
	})

	preWorkflowHooksCommandRunner := NewMockPreWorkflowHooksCommandRunner()
	When(preWorkflowHooksCommandRunner.RunPreHooks(Any[*command.Context](), Any[*events.CommentCommand]())).ThenReturn(nil)
	postWorkflowHooksCommandRunner := NewMockPostWorkflowHooksCommandRunner()
	When(postWorkflowHooksCommandRunner.RunPostHooks(Any[*command.Context](), Any[*events.CommentCommand]())).ThenReturn(nil)
	commitStatusUpdater := NewMockCommitStatusUpdater()
	When(commitStatusUpdater.UpdateCombined(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[models.CommitStatus](), Any[command.Name]())).
		ThenReturn(nil)

	controller := &APIController{
		Locker:                         locker,
		ApplyLockChecker:               applyLockChecker,
		Logger:                         logger,
		Scope:                          metricstest.NewLoggingScope(t, logger, "null"),
		ProjectCommandBuilder:          projectCommandBuilder,
		ProjectPlanCommandRunner:       projectCommandRunner,
		ProjectApplyCommandRunner:      projectCommandRunner,
		PreWorkflowHooksCommandRunner:  preWorkflowHooksCommandRunner,
		PostWorkflowHooksCommandRunner: postWorkflowHooksCommandRunner,
		WorkingDir:                     workingDir,
		WorkingDirLocker:               workingDirLocker,
		CommitStatusUpdater:            commitStatusUpdater,
	}
	executor := &apiRemediationExecutor{
		controller: controller,
		baseRepo:   baseRepo,
		logger:     logger,
	}

	results, err := executor.ExecuteApplyProjects("owner/repo", "main", "Github", []models.ProjectDrift{
		{
			ProjectName: "app",
			Path:        "app",
			Workspace:   events.DefaultWorkspace,
		},
		{
			ProjectName: "network",
			Path:        "network",
			Workspace:   events.DefaultWorkspace,
		},
	})

	Ok(t, err)
	Equals(t, 2, len(results))
	Equals(t, 2, applyCalls)
	Equals(t, []string{"network", "app"}, applyOrder)
	Equals(t, models.AppliedPlanStatus, capturedPullStatus.Projects[0].Status)
	Equals(t, models.AppliedPlanStatus, capturedPullStatus.Projects[1].Status)

	_, capturedHookCmds := preWorkflowHooksCommandRunner.VerifyWasCalled(Times(1)).
		RunPreHooks(Any[*command.Context](), Any[*events.CommentCommand]()).
		GetAllCapturedArguments()
	Equals(t, command.Plan, capturedHookCmds[0].Name)
}

func TestAPIRemediationExecutor_ExecuteApplyProjectsAbortsLaterExecutionGroups(t *testing.T) {
	RegisterMockTestingT(t)
	gmockCtrl := gomock.NewController(t)
	logger := logging.NewNoopLogger(t)
	baseRepo := models.Repo{
		FullName: "owner/repo",
		VCSHost: models.VCSHost{
			Hostname: "github.com",
			Type:     models.Github,
		},
	}

	locker := NewMockLocker(gmockCtrl)
	locker.EXPECT().UnlockByPull(baseRepo.FullName, gomock.Any()).Return(nil, nil).AnyTimes()
	applyLockChecker := NewMockApplyLocker(gmockCtrl)
	applyLockChecker.EXPECT().CheckApplyLock().Return(locking.ApplyCommandLock{}, nil)

	workingDirLocker := NewMockWorkingDirLocker()
	When(workingDirLocker.TryLock(Any[string](), Any[int](), Any[string](), Any[string](), Any[string](), Any[command.Name]())).
		ThenReturn(func() {}, nil)
	workingDir := NewMockWorkingDir()
	When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[string]())).
		ThenReturn(t.TempDir(), nil)
	When(workingDir.Delete(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest]())).
		ThenReturn(nil)

	projectCommandBuilder := NewMockProjectCommandBuilder()
	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		Then(func(args []Param) ReturnValues {
			cmd := args[1].(*events.CommentCommand)
			executionOrderGroup := 0
			if cmd.ProjectName == "app" {
				executionOrderGroup = 1
			}
			return ReturnValues{[]command.ProjectContext{{
				CommandName:         command.Plan,
				ProjectName:         cmd.ProjectName,
				RepoRelDir:          cmd.RepoRelDir,
				Workspace:           cmd.Workspace,
				ExecutionOrderGroup: executionOrderGroup,
			}}, nil}
		})
	When(projectCommandBuilder.BuildApplyCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		Then(func(args []Param) ReturnValues {
			cmd := args[1].(*events.CommentCommand)
			executionOrderGroup := 0
			if cmd.ProjectName == "app" {
				executionOrderGroup = 1
			}
			return ReturnValues{[]command.ProjectContext{{
				CommandName:               command.Apply,
				ProjectName:               cmd.ProjectName,
				RepoRelDir:                cmd.RepoRelDir,
				Workspace:                 cmd.Workspace,
				ExecutionOrderGroup:       executionOrderGroup,
				AbortOnExecutionOrderFail: true,
			}}, nil}
		})

	projectCommandRunner := NewMockProjectCommandRunner()
	When(projectCommandRunner.Plan(Any[command.ProjectContext]())).ThenReturn(command.ProjectCommandOutput{
		PlanSuccess: &models.PlanSuccess{TerraformOutput: "No changes."},
	})
	applyOrder := []string{}
	When(projectCommandRunner.Apply(Any[command.ProjectContext]())).Then(func(args []Param) ReturnValues {
		projectCtx := args[0].(command.ProjectContext)
		applyOrder = append(applyOrder, projectCtx.ProjectName)
		if projectCtx.ProjectName == "network" {
			return ReturnValues{command.ProjectCommandOutput{Error: errors.New("network apply failed")}}
		}
		return ReturnValues{command.ProjectCommandOutput{ApplySuccess: "success"}}
	})

	preWorkflowHooksCommandRunner := NewMockPreWorkflowHooksCommandRunner()
	When(preWorkflowHooksCommandRunner.RunPreHooks(Any[*command.Context](), Any[*events.CommentCommand]())).ThenReturn(nil)
	postWorkflowHooksCommandRunner := NewMockPostWorkflowHooksCommandRunner()
	When(postWorkflowHooksCommandRunner.RunPostHooks(Any[*command.Context](), Any[*events.CommentCommand]())).ThenReturn(nil)
	commitStatusUpdater := NewMockCommitStatusUpdater()
	When(commitStatusUpdater.UpdateCombined(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[models.CommitStatus](), Any[command.Name]())).
		ThenReturn(nil)

	executor := &apiRemediationExecutor{
		controller: &APIController{
			Locker:                         locker,
			ApplyLockChecker:               applyLockChecker,
			Logger:                         logger,
			Scope:                          metricstest.NewLoggingScope(t, logger, "null"),
			ProjectCommandBuilder:          projectCommandBuilder,
			ProjectPlanCommandRunner:       projectCommandRunner,
			ProjectApplyCommandRunner:      projectCommandRunner,
			PreWorkflowHooksCommandRunner:  preWorkflowHooksCommandRunner,
			PostWorkflowHooksCommandRunner: postWorkflowHooksCommandRunner,
			WorkingDir:                     workingDir,
			WorkingDirLocker:               workingDirLocker,
			CommitStatusUpdater:            commitStatusUpdater,
		},
		baseRepo: baseRepo,
		logger:   logger,
	}

	results, err := executor.ExecuteApplyProjects("owner/repo", "main", "Github", []models.ProjectDrift{
		{
			ProjectName: "network",
			Path:        "network",
			Workspace:   events.DefaultWorkspace,
		},
		{
			ProjectName: "app",
			Path:        "app",
			Workspace:   events.DefaultWorkspace,
		},
	})

	Assert(t, err != nil, "expected apply failure")
	Assert(t, strings.Contains(err.Error(), "apply had errors"), "expected apply error, got %v", err)
	Equals(t, []string{"network"}, applyOrder)
	Equals(t, 2, len(results))
	Equals(t, models.RemediationStatusFailed, results[0].Status)
	Equals(t, "network apply failed", results[0].Error)
	Equals(t, models.RemediationStatusFailed, results[1].Status)
	Equals(t, "apply skipped because an earlier execution group failed", results[1].Error)
}

func TestAPIRemediationExecutor_ExecuteApplyProjectsDoesNotSkipPRRequirements(t *testing.T) {
	RegisterMockTestingT(t)
	gmockCtrl := gomock.NewController(t)
	logger := logging.NewNoopLogger(t)
	baseRepo := models.Repo{
		FullName: "owner/repo",
		VCSHost: models.VCSHost{
			Hostname: "github.com",
			Type:     models.Github,
		},
	}

	locker := NewMockLocker(gmockCtrl)
	locker.EXPECT().UnlockByPull(baseRepo.FullName, gomock.Any()).Return(nil, nil).AnyTimes()
	applyLockChecker := NewMockApplyLocker(gmockCtrl)
	applyLockChecker.EXPECT().CheckApplyLock().Return(locking.ApplyCommandLock{}, nil)

	workingDirLocker := NewMockWorkingDirLocker()
	When(workingDirLocker.TryLock(Any[string](), Any[int](), Any[string](), Any[string](), Any[string](), Any[command.Name]())).
		ThenReturn(func() {}, nil)
	workingDir := NewMockWorkingDir()
	When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[string]())).
		ThenReturn(t.TempDir(), nil)
	When(workingDir.Delete(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest]())).
		ThenReturn(nil)

	projectCommandBuilder := NewMockProjectCommandBuilder()
	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		Then(func(args []Param) ReturnValues {
			ctx := args[0].(*command.Context)
			Assert(t, !ctx.SkipPRRequirements, "remediation apply must not bypass PR-state requirements during pre-apply plan")
			Assert(t, ctx.SuppressVCSStatus, "remediation apply should suppress normal VCS status writes")
			cmd := args[1].(*events.CommentCommand)
			return ReturnValues{[]command.ProjectContext{{
				CommandName: command.Plan,
				ProjectName: cmd.ProjectName,
				RepoRelDir:  cmd.RepoRelDir,
				Workspace:   cmd.Workspace,
			}}, nil}
		})
	When(projectCommandBuilder.BuildApplyCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		Then(func(args []Param) ReturnValues {
			ctx := args[0].(*command.Context)
			Assert(t, !ctx.SkipPRRequirements, "remediation apply must not bypass PR-state requirements during apply")
			Assert(t, ctx.SuppressVCSStatus, "remediation apply should suppress normal VCS status writes")
			cmd := args[1].(*events.CommentCommand)
			return ReturnValues{[]command.ProjectContext{{
				CommandName:        command.Apply,
				ProjectName:        cmd.ProjectName,
				RepoRelDir:         cmd.RepoRelDir,
				Workspace:          cmd.Workspace,
				SkipPRRequirements: ctx.SkipPRRequirements,
				SuppressVCSStatus:  ctx.SuppressVCSStatus,
			}}, nil}
		})

	projectCommandRunner := NewMockProjectCommandRunner()
	When(projectCommandRunner.Plan(Any[command.ProjectContext]())).ThenReturn(command.ProjectCommandOutput{
		PlanSuccess: &models.PlanSuccess{TerraformOutput: "No changes."},
	})
	When(projectCommandRunner.Apply(Any[command.ProjectContext]())).Then(func(args []Param) ReturnValues {
		projectCtx := args[0].(command.ProjectContext)
		Assert(t, !projectCtx.SkipPRRequirements, "remediation apply project context must not bypass PR-state requirements")
		Assert(t, projectCtx.SuppressVCSStatus, "remediation apply project context should suppress normal VCS status writes")
		return ReturnValues{command.ProjectCommandOutput{ApplySuccess: "success"}}
	})

	preWorkflowHooksCommandRunner := NewMockPreWorkflowHooksCommandRunner()
	When(preWorkflowHooksCommandRunner.RunPreHooks(Any[*command.Context](), Any[*events.CommentCommand]())).ThenReturn(nil)
	postWorkflowHooksCommandRunner := NewMockPostWorkflowHooksCommandRunner()
	When(postWorkflowHooksCommandRunner.RunPostHooks(Any[*command.Context](), Any[*events.CommentCommand]())).ThenReturn(nil)

	executor := &apiRemediationExecutor{
		controller: &APIController{
			Locker:                         locker,
			ApplyLockChecker:               applyLockChecker,
			Logger:                         logger,
			Scope:                          metricstest.NewLoggingScope(t, logger, "null"),
			ProjectCommandBuilder:          projectCommandBuilder,
			ProjectPlanCommandRunner:       projectCommandRunner,
			ProjectApplyCommandRunner:      projectCommandRunner,
			PreWorkflowHooksCommandRunner:  preWorkflowHooksCommandRunner,
			PostWorkflowHooksCommandRunner: postWorkflowHooksCommandRunner,
			WorkingDir:                     workingDir,
			WorkingDirLocker:               workingDirLocker,
		},
		baseRepo: baseRepo,
		logger:   logger,
	}

	results, err := executor.ExecuteApplyProjects("owner/repo", "main", "Github", []models.ProjectDrift{{
		ProjectName: "app",
		Path:        "app",
		Workspace:   events.DefaultWorkspace,
	}})

	Ok(t, err)
	Equals(t, 1, len(results))
	Equals(t, models.RemediationStatusSuccess, results[0].Status)
}

func TestVerifyNonPRBaseBranchReachability(t *testing.T) {
	repoDir, mainCommit, tagCommit, unrelatedCommit := initReachabilityGitRepo(t)
	logger := logging.NewNoopLogger(t)

	cases := []struct {
		name       string
		headRef    string
		headCommit string
		baseBranch string
		wantErr    string
	}{
		{
			name:       "raw SHA reachable from base branch",
			headRef:    mainCommit,
			headCommit: mainCommit,
		},
		{
			name:       "bare tag reachable from base branch",
			headRef:    "v1.0.0",
			headCommit: tagCommit,
		},
		{
			name:       "raw SHA not reachable from base branch",
			headRef:    unrelatedCommit,
			headCommit: unrelatedCommit,
			wantErr:    "is not reachable from base_branch",
		},
		{
			name:       "ambiguous ref matching base branch still verifies reachability",
			headRef:    "prod",
			headCommit: mainCommit,
			baseBranch: "prod",
			wantErr:    "verifying API base branch",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := &command.Context{
				Log: logger,
				Pull: models.PullRequest{
					Num:        -1,
					HeadBranch: tc.headRef,
					HeadCommit: tc.headCommit,
					BaseBranch: "main",
				},
			}
			if tc.baseBranch != "" {
				ctx.Pull.BaseBranch = tc.baseBranch
			}

			err := verifyNonPRBaseBranchReachability(ctx, repoDir)
			if tc.wantErr == "" {
				Ok(t, err)
				return
			}
			Assert(t, err != nil, "expected reachability error")
			Assert(t, strings.Contains(err.Error(), tc.wantErr), "expected %q in %v", tc.wantErr, err)
		})
	}
}

func TestAPIRemediationExecutor_ExecuteApplyProjectsRejectsStaleCachedDrift(t *testing.T) {
	RegisterMockTestingT(t)
	gmockCtrl := gomock.NewController(t)
	logger := logging.NewNoopLogger(t)
	baseRepo := models.Repo{
		FullName: "owner/repo",
		VCSHost: models.VCSHost{
			Hostname: "github.com",
			Type:     models.Github,
		},
	}

	locker := NewMockLocker(gmockCtrl)
	locker.EXPECT().UnlockByPull(baseRepo.FullName, gomock.Any()).Return(nil, nil).AnyTimes()
	applyLockChecker := NewMockApplyLocker(gmockCtrl)
	applyLockChecker.EXPECT().CheckApplyLock().Return(locking.ApplyCommandLock{}, nil)

	workingDirLocker := NewMockWorkingDirLocker()
	When(workingDirLocker.TryLock(Any[string](), Any[int](), Any[string](), Any[string](), Any[string](), Any[command.Name]())).
		ThenReturn(func() {}, nil)
	workingDir := NewMockWorkingDir()
	repoDir, checkedOutCommit := initRemediationGitRepo(t)
	When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[string]())).
		ThenReturn(repoDir, nil)
	When(workingDir.Delete(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest]())).
		ThenReturn(nil)

	projectCommandBuilder := NewMockProjectCommandBuilder()
	projectCommandRunner := NewMockProjectCommandRunner()

	executor := &apiRemediationExecutor{
		controller: &APIController{
			Locker:                    locker,
			ApplyLockChecker:          applyLockChecker,
			Logger:                    logger,
			Scope:                     metricstest.NewLoggingScope(t, logger, "null"),
			ProjectCommandBuilder:     projectCommandBuilder,
			ProjectPlanCommandRunner:  projectCommandRunner,
			ProjectApplyCommandRunner: projectCommandRunner,
			WorkingDir:                workingDir,
			WorkingDirLocker:          workingDirLocker,
		},
		baseRepo: baseRepo,
		logger:   logger,
	}

	results, err := executor.ExecuteApplyProjects("owner/repo", "main", "Github", []models.ProjectDrift{{
		ProjectName:    "app",
		Path:           "app",
		Workspace:      events.DefaultWorkspace,
		Ref:            "main",
		DetectionID:    "detect-1",
		ResolvedCommit: "0000000000000000000000000000000000000000",
		Drift:          models.DriftSummary{HasDrift: true},
	}})

	Assert(t, err != nil, "expected stale cached drift error")
	Assert(t, strings.Contains(err.Error(), "rerun drift detection before remediation apply"), "expected stale drift message, got %v", err)
	Assert(t, strings.Contains(err.Error(), checkedOutCommit), "expected checked out commit in error, got %v", err)
	Equals(t, 0, len(results))
	projectCommandBuilder.VerifyWasCalled(Never()).BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())
	projectCommandBuilder.VerifyWasCalled(Never()).BuildApplyCommands(Any[*command.Context](), Any[*events.CommentCommand]())
	projectCommandRunner.VerifyWasCalled(Never()).Plan(Any[command.ProjectContext]())
	projectCommandRunner.VerifyWasCalled(Never()).Apply(Any[command.ProjectContext]())
}

func TestAPIRemediationExecutor_ExecuteApplyProjectsPolicyFailureSkipsApply(t *testing.T) {
	RegisterMockTestingT(t)
	gmockCtrl := gomock.NewController(t)
	logger := logging.NewNoopLogger(t)
	baseRepo := models.Repo{
		FullName: "owner/repo",
		VCSHost: models.VCSHost{
			Hostname: "github.com",
			Type:     models.Github,
		},
	}

	locker := NewMockLocker(gmockCtrl)
	locker.EXPECT().UnlockByPull(baseRepo.FullName, gomock.Any()).Return(nil, nil).AnyTimes()
	applyLockChecker := NewMockApplyLocker(gmockCtrl)
	applyLockChecker.EXPECT().CheckApplyLock().Return(locking.ApplyCommandLock{}, nil)
	workingDirLocker := NewMockWorkingDirLocker()
	When(workingDirLocker.TryLock(Any[string](), Any[int](), Any[string](), Any[string](), Any[string](), Any[command.Name]())).
		ThenReturn(func() {}, nil)
	workingDir := NewMockWorkingDir()
	When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[string]())).
		ThenReturn(t.TempDir(), nil)
	When(workingDir.Delete(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest]())).
		ThenReturn(nil)

	projectCommandBuilder := NewMockProjectCommandBuilder()
	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		ThenReturn([]command.ProjectContext{
			{
				CommandName: command.Plan,
				ProjectName: "app",
				RepoRelDir:  "app",
				Workspace:   events.DefaultWorkspace,
			},
			{
				CommandName: command.PolicyCheck,
				ProjectName: "app",
				RepoRelDir:  "app",
				Workspace:   events.DefaultWorkspace,
			},
		}, nil)

	projectCommandRunner := NewMockProjectCommandRunner()
	When(projectCommandRunner.Plan(Any[command.ProjectContext]())).ThenReturn(command.ProjectCommandOutput{
		PlanSuccess: &models.PlanSuccess{TerraformOutput: "No changes."},
	})
	When(projectCommandRunner.PolicyCheck(Any[command.ProjectContext]())).ThenReturn(command.ProjectCommandOutput{
		Failure: "policy failed",
	})
	preWorkflowHooksCommandRunner := NewMockPreWorkflowHooksCommandRunner()
	When(preWorkflowHooksCommandRunner.RunPreHooks(Any[*command.Context](), Any[*events.CommentCommand]())).ThenReturn(nil)
	postWorkflowHooksCommandRunner := NewMockPostWorkflowHooksCommandRunner()
	When(postWorkflowHooksCommandRunner.RunPostHooks(Any[*command.Context](), Any[*events.CommentCommand]())).ThenReturn(nil)
	commitStatusUpdater := NewMockCommitStatusUpdater()
	When(commitStatusUpdater.UpdateCombined(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[models.CommitStatus](), Any[command.Name]())).
		ThenReturn(nil)

	executor := &apiRemediationExecutor{
		controller: &APIController{
			Locker:                          locker,
			ApplyLockChecker:                applyLockChecker,
			Logger:                          logger,
			Scope:                           metricstest.NewLoggingScope(t, logger, "null"),
			ProjectCommandBuilder:           projectCommandBuilder,
			ProjectPlanCommandRunner:        projectCommandRunner,
			ProjectPolicyCheckCommandRunner: projectCommandRunner,
			ProjectApplyCommandRunner:       projectCommandRunner,
			PreWorkflowHooksCommandRunner:   preWorkflowHooksCommandRunner,
			PostWorkflowHooksCommandRunner:  postWorkflowHooksCommandRunner,
			WorkingDir:                      workingDir,
			WorkingDirLocker:                workingDirLocker,
			CommitStatusUpdater:             commitStatusUpdater,
		},
		baseRepo: baseRepo,
		logger:   logger,
	}

	results, err := executor.ExecuteApplyProjects("owner/repo", "main", "Github", []models.ProjectDrift{{
		ProjectName: "app",
		Path:        "app",
		Workspace:   events.DefaultWorkspace,
	}})

	Assert(t, err != nil, "expected policy failure")
	Equals(t, 1, len(results))
	Equals(t, models.RemediationStatusFailed, results[0].Status)
	Equals(t, "policy failed", results[0].Error)
	projectCommandRunner.VerifyWasCalled(Times(1)).PolicyCheck(Any[command.ProjectContext]())
	projectCommandRunner.VerifyWasCalled(Never()).Apply(Any[command.ProjectContext]())
}

func TestProjectRemediationResultsFromPlanReconcilesProjectOnlyTargets(t *testing.T) {
	results := projectRemediationResultsFromPlan([]models.ProjectDrift{{
		ProjectName: "app",
	}}, &command.Result{ProjectResults: []command.ProjectResult{{
		ProjectName: "app",
		RepoRelDir:  "apps/app",
		Workspace:   events.DefaultWorkspace,
		ProjectCommandOutput: command.ProjectCommandOutput{
			PlanSuccess: &models.PlanSuccess{TerraformOutput: "No changes."},
		},
	}}})

	Equals(t, 1, len(results))
	Equals(t, "apps/app", results[0].Path)
	Equals(t, events.DefaultWorkspace, results[0].Workspace)
	Equals(t, models.RemediationStatusRunning, results[0].Status)

	results = mergeApplyRemediationResults(results, &command.Result{ProjectResults: []command.ProjectResult{{
		ProjectName: "app",
		RepoRelDir:  "apps/app",
		Workspace:   events.DefaultWorkspace,
		ProjectCommandOutput: command.ProjectCommandOutput{
			ApplySuccess: "success",
		},
	}}})
	Equals(t, 1, len(results))
	Equals(t, models.RemediationStatusSuccess, results[0].Status)
}

func TestProjectRemediationResultsFromPlanReconcilesPathOnlyTargets(t *testing.T) {
	results := projectRemediationResultsFromPlan([]models.ProjectDrift{{
		Path:      "apps/app",
		Workspace: events.DefaultWorkspace,
	}}, &command.Result{ProjectResults: []command.ProjectResult{{
		ProjectName: "app",
		RepoRelDir:  "apps/app",
		Workspace:   events.DefaultWorkspace,
		ProjectCommandOutput: command.ProjectCommandOutput{
			PlanSuccess: &models.PlanSuccess{TerraformOutput: "No changes."},
		},
	}}})

	Equals(t, 1, len(results))
	Equals(t, "app", results[0].ProjectName)
	Equals(t, "apps/app", results[0].Path)
	Equals(t, events.DefaultWorkspace, results[0].Workspace)
	Equals(t, models.RemediationStatusRunning, results[0].Status)

	results = mergeApplyRemediationResults(results, &command.Result{ProjectResults: []command.ProjectResult{{
		ProjectName: "app",
		RepoRelDir:  "apps/app",
		Workspace:   events.DefaultWorkspace,
		ProjectCommandOutput: command.ProjectCommandOutput{
			ApplySuccess: "success",
		},
	}}})
	Equals(t, 1, len(results))
	Equals(t, models.RemediationStatusSuccess, results[0].Status)
}

func TestProjectRemediationResultsFromPlanPreservesPlanOutputWhenPolicyCheckFollows(t *testing.T) {
	results := projectRemediationResultsFromPlan([]models.ProjectDrift{{
		ProjectName: "app",
		Path:        "app",
		Workspace:   events.DefaultWorkspace,
	}}, &command.Result{ProjectResults: []command.ProjectResult{
		{
			Command:     command.Plan,
			ProjectName: "app",
			RepoRelDir:  "app",
			Workspace:   events.DefaultWorkspace,
			ProjectCommandOutput: command.ProjectCommandOutput{
				PlanSuccess: &models.PlanSuccess{TerraformOutput: "Plan: 1 to add, 0 to change, 0 to destroy."},
			},
		},
		{
			Command:     command.PolicyCheck,
			ProjectName: "app",
			RepoRelDir:  "app",
			Workspace:   events.DefaultWorkspace,
			ProjectCommandOutput: command.ProjectCommandOutput{
				PolicyCheckResults: &models.PolicyCheckResults{},
			},
		},
	}})

	Equals(t, 1, len(results))
	Equals(t, "Plan: 1 to add, 0 to change, 0 to destroy.", results[0].PlanOutput)
	Equals(t, models.RemediationStatusRunning, results[0].Status)
}

func TestAPIRemediationExecutor_ExecutePlanPreservesNamedTargetWorkspace(t *testing.T) {
	RegisterMockTestingT(t)
	gmockCtrl := gomock.NewController(t)
	logger := logging.NewNoopLogger(t)
	baseRepo := models.Repo{
		FullName: "owner/repo",
		VCSHost: models.VCSHost{
			Hostname: "github.com",
			Type:     models.Github,
		},
	}

	locker := NewMockLocker(gmockCtrl)
	locker.EXPECT().UnlockByPull(baseRepo.FullName, gomock.Any()).Return(nil, nil).AnyTimes()

	workingDirLocker := NewMockWorkingDirLocker()
	When(workingDirLocker.TryLock(Any[string](), Any[int](), Any[string](), Any[string](), Any[string](), Any[command.Name]())).
		ThenReturn(func() {}, nil)
	workingDir := NewMockWorkingDir()
	When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[string]())).
		ThenReturn(t.TempDir(), nil)

	var capturedCmd *events.CommentCommand
	projectCommandBuilder := NewMockProjectCommandBuilder()
	When(projectCommandBuilder.BuildPlanCommands(Any[*command.Context](), Any[*events.CommentCommand]())).
		Then(func(args []Param) ReturnValues {
			capturedCmd = args[1].(*events.CommentCommand)
			return ReturnValues{[]command.ProjectContext{{
				CommandName: command.Plan,
				ProjectName: capturedCmd.ProjectName,
				RepoRelDir:  capturedCmd.RepoRelDir,
				Workspace:   capturedCmd.Workspace,
			}}, nil}
		})

	projectCommandRunner := NewMockProjectCommandRunner()
	When(projectCommandRunner.Plan(Any[command.ProjectContext]())).ThenReturn(command.ProjectCommandOutput{
		PlanSuccess: &models.PlanSuccess{TerraformOutput: "No changes."},
	})

	preWorkflowHooksCommandRunner := NewMockPreWorkflowHooksCommandRunner()
	When(preWorkflowHooksCommandRunner.RunPreHooks(Any[*command.Context](), Any[*events.CommentCommand]())).ThenReturn(nil)
	postWorkflowHooksCommandRunner := NewMockPostWorkflowHooksCommandRunner()
	When(postWorkflowHooksCommandRunner.RunPostHooks(Any[*command.Context](), Any[*events.CommentCommand]())).ThenReturn(nil)
	commitStatusUpdater := NewMockCommitStatusUpdater()
	When(commitStatusUpdater.UpdateCombined(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[models.CommitStatus](), Any[command.Name]())).
		ThenReturn(nil)

	controller := &APIController{
		Locker:                         locker,
		Logger:                         logger,
		Scope:                          metricstest.NewLoggingScope(t, logger, "null"),
		ProjectCommandBuilder:          projectCommandBuilder,
		ProjectPlanCommandRunner:       projectCommandRunner,
		PreWorkflowHooksCommandRunner:  preWorkflowHooksCommandRunner,
		PostWorkflowHooksCommandRunner: postWorkflowHooksCommandRunner,
		WorkingDir:                     workingDir,
		WorkingDirLocker:               workingDirLocker,
		CommitStatusUpdater:            commitStatusUpdater,
	}
	executor := &apiRemediationExecutor{
		controller: controller,
		baseRepo:   baseRepo,
		logger:     logger,
	}

	_, _, err := executor.ExecutePlan("owner/repo", "main", "Github", "app", "modules/app", "staging")

	Ok(t, err)
	Assert(t, capturedCmd != nil, "expected BuildPlanCommands to be called")
	Equals(t, "app", capturedCmd.ProjectName)
	Equals(t, "modules/app", capturedCmd.RepoRelDir)
	Equals(t, "staging", capturedCmd.Workspace)
}

func initRemediationGitRepo(t *testing.T) (string, string) {
	t.Helper()
	repoDir := remediationGitTempDir(t)
	runRemediationGit(t, repoDir, "init", "-q")
	runRemediationGit(t, repoDir, "config", "user.email", "test@example.com")
	runRemediationGit(t, repoDir, "config", "user.name", "Atlantis Test")
	Ok(t, os.WriteFile(filepath.Join(repoDir, "main.tf"), []byte("resource \"null_resource\" \"app\" {}\n"), 0600))
	runRemediationGit(t, repoDir, "add", "main.tf")
	runRemediationGit(t, repoDir, "commit", "-q", "-m", "initial")

	out := runRemediationGit(t, repoDir, "rev-parse", "HEAD")
	return repoDir, strings.TrimSpace(out)
}

func initReachabilityGitRepo(t *testing.T) (string, string, string, string) {
	t.Helper()
	rootDir := remediationGitTempDir(t)
	originDir := filepath.Join(rootDir, "origin.git")
	repoDir := filepath.Join(rootDir, "work")
	Ok(t, os.MkdirAll(repoDir, 0700))
	runRemediationGit(t, "", "init", "--bare", originDir)
	runRemediationGit(t, repoDir, "init", "-q")
	runRemediationGit(t, repoDir, "config", "user.email", "test@example.com")
	runRemediationGit(t, repoDir, "config", "user.name", "Atlantis Test")
	Ok(t, os.WriteFile(filepath.Join(repoDir, "main.tf"), []byte("resource \"null_resource\" \"app\" {}\n"), 0600))
	runRemediationGit(t, repoDir, "add", "main.tf")
	runRemediationGit(t, repoDir, "commit", "-q", "-m", "initial")
	tagCommit := strings.TrimSpace(runRemediationGit(t, repoDir, "rev-parse", "HEAD"))
	runRemediationGit(t, repoDir, "tag", "v1.0.0")
	Ok(t, os.WriteFile(filepath.Join(repoDir, "main.tf"), []byte("resource \"null_resource\" \"app\" { triggers = { value = \"main\" } }\n"), 0600))
	runRemediationGit(t, repoDir, "add", "main.tf")
	runRemediationGit(t, repoDir, "commit", "-q", "-m", "main update")
	mainCommit := strings.TrimSpace(runRemediationGit(t, repoDir, "rev-parse", "HEAD"))
	runRemediationGit(t, repoDir, "branch", "-M", "main")
	runRemediationGit(t, repoDir, "remote", "add", "origin", "file://"+originDir)
	runRemediationGit(t, repoDir, "push", "-q", "origin", "main", "v1.0.0")

	runRemediationGit(t, repoDir, "checkout", "--orphan", "untrusted")
	runRemediationGit(t, repoDir, "rm", "-q", "-f", "main.tf")
	Ok(t, os.WriteFile(filepath.Join(repoDir, "main.tf"), []byte("resource \"null_resource\" \"untrusted\" {}\n"), 0600))
	runRemediationGit(t, repoDir, "add", "main.tf")
	runRemediationGit(t, repoDir, "commit", "-q", "-m", "untrusted")
	unrelatedCommit := strings.TrimSpace(runRemediationGit(t, repoDir, "rev-parse", "HEAD"))

	return repoDir, mainCommit, tagCommit, unrelatedCommit
}

func remediationGitTempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", t.Name())
	Ok(t, err)
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})
	return dir
}

func runRemediationGit(t *testing.T, repoDir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	if repoDir != "" {
		cmd.Dir = repoDir
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %s: %v", strings.Join(args, " "), strings.TrimSpace(string(out)), err)
	}
	return string(out)
}
