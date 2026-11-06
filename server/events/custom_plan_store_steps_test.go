// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/core/runtime"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestProjectCommandRunner_PlanProducingRunStep(t *testing.T) {
	tests := []struct {
		name             string
		planContents     []byte
		producePlan      bool
		produceDirectory bool
		preexistingLocal bool
		skipIfEmpty      bool
		customPlanErr    error
		saveErr          error
		removeErr        error
		wantSaveCalls    int
		wantRemoveCalls  int
		wantStaleRemote  bool
		wantLocalAbsent  bool
		wantErr          string
	}{
		{
			name:          "stores custom plan",
			planContents:  []byte("custom plan"),
			producePlan:   true,
			wantSaveCalls: 1,
		},
		{
			name:            "replaces and removes stale stored plan for empty custom plan",
			producePlan:     true,
			skipIfEmpty:     true,
			wantSaveCalls:   1,
			wantRemoveCalls: 1,
		},
		{
			name:            "leaves only non-executable object when empty plan removal fails",
			producePlan:     true,
			skipIfEmpty:     true,
			removeErr:       errors.New("delete unavailable"),
			wantSaveCalls:   1,
			wantRemoveCalls: 1,
		},
		{
			name:          "stores empty custom plan by default",
			producePlan:   true,
			wantSaveCalls: 1,
		},
		{
			name:        "requires plan file even when skipping empty plans",
			skipIfEmpty: true,
			wantErr:     "finding plan file",
		},
		{
			name:          "returns storage error",
			planContents:  []byte("custom plan"),
			producePlan:   true,
			saveErr:       errors.New("store unavailable"),
			wantSaveCalls: 1,
			wantErr:       "saving plan: store unavailable",
		},
		{
			name:             "does not store after custom plan failure",
			customPlanErr:    errors.New("plan failed"),
			preexistingLocal: true,
			wantStaleRemote:  true,
			wantLocalAbsent:  true,
			wantErr:          "plan failed",
		},
		{
			name:             "rejects non-regular plan file",
			produceDirectory: true,
			wantErr:          "is not a regular file",
		},
		{
			name:             "successful no-op producer cannot save stale local plan",
			preexistingLocal: true,
			wantStaleRemote:  true,
			wantLocalAbsent:  true,
			wantErr:          "finding plan file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RegisterMockTestingT(t)
			mockWorkingDir := mocks.NewMockWorkingDir()
			mockLocker := mocks.NewMockProjectLocker()
			mockRequirements := mocks.NewMockCommandRequirementHandler()
			store := &recordingWorkflowPlanStore{
				saveErr:   tt.saveErr,
				removeErr: tt.removeErr,
				object:    []byte("stale executable plan"),
			}
			repoDir := t.TempDir()
			ctx := command.ProjectContext{
				Log:         logging.NewNoopLogger(t),
				CommandName: command.Plan,
				Steps: []valid.Step{
					{
						StepName:   "run",
						RunCommand: "custom plan",
						PlanStore: &valid.RunPlanStore{
							Mode:        valid.RunPlanStoreSaveMode,
							SkipIfEmpty: tt.skipIfEmpty,
						},
					},
				},
				Workspace:  "default",
				RepoRelDir: ".",
				Pull: models.PullRequest{
					Num:      1,
					BaseRepo: models.Repo{FullName: "owner/repo", Owner: "owner", Name: "repo"},
				},
			}
			planPath := runtime.GetPlanFilePath(ctx, repoDir)
			producer := &planProducingCustomStepRunner{
				planPath:         planPath,
				contents:         tt.planContents,
				producePlan:      tt.producePlan,
				produceDirectory: tt.produceDirectory,
				err:              tt.customPlanErr,
			}
			runner := events.DefaultProjectCommandRunner{
				Locker:                    mockLocker,
				LockURLGenerator:          mockURLGenerator{},
				RunStepRunner:             producer,
				WorkingDir:                mockWorkingDir,
				WorkingDirLocker:          events.NewDefaultWorkingDirLocker(),
				CommandRequirementHandler: mockRequirements,
				PlanStore:                 store,
			}
			if tt.preexistingLocal {
				Ok(t, os.WriteFile(planPath, []byte("stale local plan"), 0o600))
			}

			When(mockWorkingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Eq(ctx.Pull), Eq(ctx.Workspace))).ThenReturn(repoDir, nil)
			When(mockWorkingDir.GitReadLock(ctx.Pull.BaseRepo, ctx.Pull, ctx.Workspace)).ThenReturn(func() {})
			When(mockLocker.TryLock(Any[logging.SimpleLogging](), Eq(ctx.Pull), Any[models.User](), Eq(ctx.Workspace), Any[models.Project](), AnyBool())).
				ThenReturn(&events.TryLockResponse{LockAcquired: true, LockKey: "lock-key", UnlockFn: func() error { return nil }}, nil)

			result := runner.Plan(ctx)

			Equals(t, tt.wantSaveCalls, len(store.savePaths))
			if tt.wantSaveCalls == 1 {
				Equals(t, planPath, store.savePaths[0])
			}
			Equals(t, tt.wantRemoveCalls, len(store.removePaths))
			if tt.skipIfEmpty && tt.producePlan {
				Equals(t, 0, len(store.object))
			}
			if tt.wantStaleRemote {
				Equals(t, "stale executable plan", string(store.object))
			}
			if tt.wantLocalAbsent {
				_, statErr := os.Stat(planPath)
				Assert(t, os.IsNotExist(statErr), "expected stale local plan to be removed, got %v", statErr)
			}
			if tt.wantErr != "" {
				Assert(t, result.Error != nil, "expected plan error")
				Assert(t, strings.Contains(result.Error.Error(), tt.wantErr), "got %q", result.Error)
				return
			}
			Ok(t, result.Error)
			Assert(t, result.PlanSuccess != nil, "expected plan success")
			Equals(t, "planned", result.PlanSuccess.TerraformOutput)
		})
	}
}

func TestProjectCommandRunner_PlanConsumingRunStep(t *testing.T) {
	tests := []struct {
		name            string
		customApplyErr  error
		removeErr       error
		skipLoadWrite   bool
		loadAsDirectory bool
		wantRun         bool
		wantRemoveCalls int
		wantErr         string
	}{
		{
			name:            "removes plan after custom apply",
			wantRun:         true,
			wantRemoveCalls: 1,
		},
		{
			name:           "retains plan after custom apply failure",
			customApplyErr: errors.New("apply failed"),
			wantRun:        true,
			wantErr:        "apply failed",
		},
		{
			name:            "does not fail completed apply when removal fails",
			removeErr:       errors.New("store unavailable"),
			wantRun:         true,
			wantRemoveCalls: 1,
		},
		{
			name:          "rejects missing plan before custom apply",
			skipLoadWrite: true,
			wantErr:       "finding plan file",
		},
		{
			name:            "rejects non-regular plan before custom apply",
			loadAsDirectory: true,
			wantErr:         "is not a regular file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RegisterMockTestingT(t)
			mockRun := mocks.NewMockCustomStepRunner()
			mockWorkingDir := mocks.NewMockWorkingDir()
			mockLocker := mocks.NewMockProjectLocker()
			mockRequirements := mocks.NewMockCommandRequirementHandler()
			store := &recordingWorkflowPlanStore{
				loadContents:    []byte("custom plan"),
				removeErr:       tt.removeErr,
				skipLoadWrite:   tt.skipLoadWrite,
				loadAsDirectory: tt.loadAsDirectory,
			}
			runner := events.DefaultProjectCommandRunner{
				Locker:                    mockLocker,
				LockURLGenerator:          mockURLGenerator{},
				RunStepRunner:             mockRun,
				WorkingDir:                mockWorkingDir,
				WorkingDirLocker:          events.NewDefaultWorkingDirLocker(),
				CommandRequirementHandler: mockRequirements,
				PlanStore:                 store,
			}
			repoDir := t.TempDir()
			ctx := command.ProjectContext{
				Log:         logging.NewNoopLogger(t),
				CommandName: command.Apply,
				Steps: []valid.Step{
					{
						StepName:   "run",
						RunCommand: "custom apply",
						PlanStore:  &valid.RunPlanStore{Mode: valid.RunPlanStoreConsumeMode},
					},
				},
				Workspace:  "default",
				RepoRelDir: ".",
				Pull: models.PullRequest{
					Num:      1,
					BaseRepo: models.Repo{FullName: "owner/repo", Owner: "owner", Name: "repo"},
				},
			}
			planPath := runtime.GetPlanFilePath(ctx, repoDir)

			When(mockWorkingDir.GetWorkingDir(ctx.Pull.BaseRepo, ctx.Pull, ctx.Workspace)).ThenReturn(repoDir, nil)
			When(mockWorkingDir.GitReadLock(ctx.Pull.BaseRepo, ctx.Pull, ctx.Workspace)).ThenReturn(func() {})
			When(mockLocker.TryLock(Any[logging.SimpleLogging](), Eq(ctx.Pull), Any[models.User](), Eq(ctx.Workspace), Any[models.Project](), AnyBool())).
				ThenReturn(&events.TryLockResponse{LockAcquired: true, LockKey: "lock-key"}, nil)
			When(mockRun.Run(ctx, nil, "custom apply", repoDir, map[string]string{}, true, nil, nil)).ThenReturn("applied", tt.customApplyErr)

			result := runner.Apply(ctx)

			Equals(t, 1, len(store.loadPaths))
			Equals(t, planPath, store.loadPaths[0])
			Equals(t, tt.wantRemoveCalls, len(store.removePaths))
			if tt.wantRemoveCalls == 1 {
				Equals(t, planPath, store.removePaths[0])
			}
			if tt.wantRun {
				mockRun.VerifyWasCalledOnce().Run(ctx, nil, "custom apply", repoDir, map[string]string{}, true, nil, nil)
			} else {
				mockRun.VerifyWasCalled(Never()).Run(Any[command.ProjectContext](), Any[*valid.CommandShell](), Any[string](), Any[string](), Any[map[string]string](), AnyBool(), Any[[]valid.PostProcessRunOutputOption](), Any[[]*regexp.Regexp]())
			}
			if tt.wantErr != "" {
				Assert(t, result.Error != nil, "expected apply error")
				Assert(t, strings.Contains(result.Error.Error(), tt.wantErr), "got %q", result.Error)
				return
			}
			Ok(t, result.Error)
			Equals(t, "applied", result.ApplySuccess)
		})
	}
}

func TestProjectCommandRunner_RevalidatesPlanImmediatelyBeforeConsumingRun(t *testing.T) {
	RegisterMockTestingT(t)
	mockWorkingDir := mocks.NewMockWorkingDir()
	mockLocker := mocks.NewMockProjectLocker()
	mockRequirements := mocks.NewMockCommandRequirementHandler()
	db := newTestBoltDB(t)
	store := &recordingWorkflowPlanStore{loadContents: []byte("original plan")}
	repoDir := t.TempDir()
	ctx := command.ProjectContext{
		Log:         logging.NewNoopLogger(t),
		CommandName: command.Apply,
		Steps: []valid.Step{
			{StepName: "run", RunCommand: "prepare"},
			{
				StepName:   "run",
				RunCommand: "custom apply",
				PlanStore:  &valid.RunPlanStore{Mode: valid.RunPlanStoreConsumeMode},
			},
		},
		Workspace:   "default",
		RepoRelDir:  ".",
		ProjectName: "project",
		Pull: models.PullRequest{
			Num:        1,
			HeadCommit: "abc123",
			BaseRepo:   models.Repo{FullName: "owner/repo", Owner: "owner", Name: "repo"},
		},
	}
	_, err := db.UpdatePullWithResults(ctx.Pull, []command.ProjectResult{plannedProjectResult(ctx.RepoRelDir, ctx.Workspace, ctx.ProjectName)})
	Ok(t, err)
	planPath := runtime.GetPlanFilePath(ctx, repoDir)
	run := &commandMutatingCustomStepRunner{
		mutateCommand: "prepare",
		planPath:      planPath,
	}
	runner := events.DefaultProjectCommandRunner{
		Locker:                    mockLocker,
		RunStepRunner:             run,
		WorkingDir:                mockWorkingDir,
		WorkingDirLocker:          events.NewDefaultWorkingDirLocker(),
		CommandRequirementHandler: mockRequirements,
		ApplyPlanValidator:        &events.DefaultApplyPlanValidator{PullStatusFetcher: db},
		PlanStore:                 store,
	}

	When(mockWorkingDir.GetWorkingDir(ctx.Pull.BaseRepo, ctx.Pull, ctx.Workspace)).ThenReturn(repoDir, nil)
	When(mockWorkingDir.GitReadLock(ctx.Pull.BaseRepo, ctx.Pull, ctx.Workspace)).ThenReturn(func() {})
	When(mockLocker.TryLock(Any[logging.SimpleLogging](), Eq(ctx.Pull), Any[models.User](), Eq(ctx.Workspace), Any[models.Project](), AnyBool())).
		ThenReturn(&events.TryLockResponse{LockAcquired: true, LockKey: "lock-key"}, nil)

	result := runner.Apply(ctx)

	Assert(t, result.Error != nil, "expected plan mutation error")
	Assert(t, strings.Contains(result.Error.Error(), "plan file changed"), "got %q", result.Error)
	Equals(t, []string{"prepare"}, run.commands)
	Equals(t, 0, len(store.removePaths))
}

type recordingWorkflowPlanStore struct {
	loadContents    []byte
	saveErr         error
	removeErr       error
	loadPaths       []string
	savePaths       []string
	removePaths     []string
	object          []byte
	skipLoadWrite   bool
	loadAsDirectory bool
}

func (s *recordingWorkflowPlanStore) Save(_ command.ProjectContext, planPath string) error {
	s.savePaths = append(s.savePaths, planPath)
	contents, err := os.ReadFile(planPath)
	if err != nil {
		return err
	}
	s.object = contents
	return s.saveErr
}

func (s *recordingWorkflowPlanStore) Load(_ command.ProjectContext, planPath string) error {
	s.loadPaths = append(s.loadPaths, planPath)
	if s.skipLoadWrite {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(planPath), 0o700); err != nil {
		return err
	}
	if s.loadAsDirectory {
		return os.Mkdir(planPath, 0o700)
	}
	return os.WriteFile(planPath, s.loadContents, 0o600)
}

func (s *recordingWorkflowPlanStore) Remove(_ command.ProjectContext, planPath string) error {
	s.removePaths = append(s.removePaths, planPath)
	if s.removeErr == nil {
		s.object = nil
	}
	return s.removeErr
}

func (s *recordingWorkflowPlanStore) ListWorkspaces(string, string, int) ([]string, error) {
	return nil, nil
}

func (s *recordingWorkflowPlanStore) RestorePlans(string, string, string, int) error {
	return nil
}

func (s *recordingWorkflowPlanStore) DeleteForPull(string, string, int) error {
	return nil
}

func (s *recordingWorkflowPlanStore) DeletePlanForProject(string, string, int, string, string, string) error {
	return nil
}

type planProducingCustomStepRunner struct {
	planPath         string
	contents         []byte
	producePlan      bool
	produceDirectory bool
	err              error
}

func (r *planProducingCustomStepRunner) Run(
	_ command.ProjectContext,
	_ *valid.CommandShell,
	_ string,
	_ string,
	_ map[string]string,
	_ bool,
	_ []valid.PostProcessRunOutputOption,
	_ []*regexp.Regexp,
) (string, error) {
	if r.err != nil {
		return "planned", r.err
	}
	if r.produceDirectory {
		if err := os.Mkdir(r.planPath, 0o700); err != nil {
			return "", err
		}
	} else if r.producePlan {
		if err := os.WriteFile(r.planPath, r.contents, 0o600); err != nil {
			return "", err
		}
	}
	return "planned", nil
}

type commandMutatingCustomStepRunner struct {
	mutateCommand string
	planPath      string
	commands      []string
}

func (r *commandMutatingCustomStepRunner) Run(
	_ command.ProjectContext,
	_ *valid.CommandShell,
	cmd string,
	_ string,
	_ map[string]string,
	_ bool,
	_ []valid.PostProcessRunOutputOption,
	_ []*regexp.Regexp,
) (string, error) {
	r.commands = append(r.commands, cmd)
	if cmd == r.mutateCommand {
		if err := os.WriteFile(r.planPath, []byte("mutated plan"), 0o600); err != nil {
			return "", err
		}
	}
	return cmd, nil
}
