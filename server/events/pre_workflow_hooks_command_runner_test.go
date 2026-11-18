// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"errors"
	"testing"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	runtime_mocks "github.com/runatlantis/atlantis/server/core/runtime/mocks"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/models/testdata"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

var preWh events.DefaultPreWorkflowHooksCommandRunner
var preWhWorkingDir *mocks.MockWorkingDir
var preWhWorkingDirLocker *mocks.MockWorkingDirLocker
var whPreWorkflowHookRunner *runtime_mocks.MockPreWorkflowHookRunner
var preCommitStatusUpdater *mocks.MockCommitStatusUpdater

func preWorkflowHooksSetup(t *testing.T) {
	RegisterMockTestingT(t)
	vcsClient := vcsmocks.NewMockClient()
	preWhWorkingDir = mocks.NewMockWorkingDir()
	preWhWorkingDirLocker = mocks.NewMockWorkingDirLocker()
	When(preWhWorkingDirLocker.TryLockPull(Any[string](), Any[int](), Any[command.Name](),
		Any[events.WorkingDirLockMetadata]())).ThenReturn(func() {}, nil)
	whPreWorkflowHookRunner = runtime_mocks.NewMockPreWorkflowHookRunner()
	preCommitStatusUpdater = mocks.NewMockCommitStatusUpdater()
	preWorkflowHookURLGenerator := mocks.NewMockPreWorkflowHookURLGenerator()

	preWh = events.DefaultPreWorkflowHooksCommandRunner{
		VCSClient:             vcsClient,
		WorkingDirLocker:      preWhWorkingDirLocker,
		WorkingDir:            preWhWorkingDir,
		PreWorkflowHookRunner: whPreWorkflowHookRunner,
		CommitStatusUpdater:   preCommitStatusUpdater,
		Router:                preWorkflowHookURLGenerator,
	}
}

func newBool(b bool) *bool {
	return &b
}

func TestRunPreHooks_Clone(t *testing.T) {

	log := logging.NewNoopLogger(t)

	var newPull = testdata.Pull
	newPull.BaseRepo = testdata.GithubRepo

	ctx := &command.Context{
		Pull:     newPull,
		HeadRepo: testdata.GithubRepo,
		User:     testdata.User,
		Log:      log,
	}

	defaultShell := "sh"
	defaultShellArgs := "-c"

	testHook := valid.WorkflowHook{
		StepName:   "test",
		RunCommand: "some command",
	}

	testHookWithShell := valid.WorkflowHook{
		StepName:   "test1",
		RunCommand: "echo test1",
		Shell:      "bash",
	}

	testHookWithShellArgs := valid.WorkflowHook{
		StepName:   "test2",
		RunCommand: "echo test2",
		ShellArgs:  "-ce",
	}

	testHookWithShellandShellArgs := valid.WorkflowHook{
		StepName:   "test3",
		RunCommand: "echo test3",
		Shell:      "bash",
		ShellArgs:  "-ce",
	}

	testHookWithPlanCommand := valid.WorkflowHook{
		StepName:   "test4",
		RunCommand: "echo test4",
		Commands:   "plan",
	}

	testHookWithPlanApplyCommands := valid.WorkflowHook{
		StepName:   "test5",
		RunCommand: "echo test5",
		Commands:   "plan, apply",
	}

	repoDir := "path/to/repo"
	result := "some result"
	runtimeDesc := ""

	pCtx := models.WorkflowHookCommandContext{
		BaseRepo:    testdata.GithubRepo,
		HeadRepo:    testdata.GithubRepo,
		Pull:        newPull,
		Log:         log,
		User:        testdata.User,
		Verbose:     false,
		CommandName: "plan",
	}

	planCmd := &events.CommentCommand{
		Name: command.Plan,
	}

	applyCmd := &events.CommentCommand{
		Name: command.Apply,
	}

	t.Run("success hooks in cfg", func(t *testing.T) {
		preWorkflowHooksSetup(t)

		var unlockCalled = newBool(false)
		unlockFn := func() {
			unlockCalled = newBool(true)
		}

		globalCfg := valid.GlobalCfg{
			Repos: []valid.Repo{
				{
					ID: testdata.GithubRepo.ID(),
					PreWorkflowHooks: []*valid.WorkflowHook{
						&testHook,
					},
				},
			},
		}

		preWh.GlobalCfg = globalCfg

		When(preWhWorkingDirLocker.TryLock(testdata.GithubRepo.FullName, newPull.Num, events.DefaultWorkspace,
			events.DefaultRepoRelDir, "", command.Plan, events.WorkingDirLockMetadataForPull(newPull))).ThenReturn(unlockFn, nil)
		When(preWhWorkingDir.Clone(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(newPull),
			Eq(events.DefaultWorkspace))).ThenReturn(repoDir, nil)
		When(whPreWorkflowHookRunner.Run(Any[models.WorkflowHookCommandContext](), Eq(testHook.RunCommand),
			Any[string](), Any[string](), Eq(repoDir))).ThenReturn(result, runtimeDesc, nil)

		err := preWh.RunPreHooks(ctx, planCmd)

		Ok(t, err)
		whPreWorkflowHookRunner.VerifyWasCalledOnce().Run(Any[models.WorkflowHookCommandContext](),
			Eq(testHook.RunCommand), Eq(defaultShell), Eq(defaultShellArgs), Eq(repoDir))
		Assert(t, *unlockCalled == true, "unlock function called")
	})

	t.Run("success hooks not in cfg", func(t *testing.T) {
		preWorkflowHooksSetup(t)
		globalCfg := valid.GlobalCfg{
			Repos: []valid.Repo{
				// one with hooks but mismatched id
				{
					ID: "id1",
					PreWorkflowHooks: []*valid.WorkflowHook{
						&testHook,
					},
				},
				// one with the correct id but no hooks
				{
					ID:               testdata.GithubRepo.ID(),
					PreWorkflowHooks: []*valid.WorkflowHook{},
				},
			},
		}

		preWh.GlobalCfg = globalCfg

		err := preWh.RunPreHooks(ctx, planCmd)

		Ok(t, err)

		whPreWorkflowHookRunner.VerifyWasCalled(Never()).Run(Any[models.WorkflowHookCommandContext](), Eq(testHook.RunCommand),
			Eq(defaultShell), Eq(defaultShellArgs), Eq(repoDir))
		preWhWorkingDirLocker.VerifyWasCalled(Never()).TryLock(testdata.GithubRepo.FullName, newPull.Num, events.DefaultWorkspace, "", "", command.Plan, events.WorkingDirLockMetadataForPull(newPull))
		preWhWorkingDirLocker.VerifyWasCalled(Never()).TryLockPull(Any[string](), Any[int](), Any[command.Name](), Any[events.WorkingDirLockMetadata]())
		preWhWorkingDir.VerifyWasCalled(Never()).Clone(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(newPull),
			Eq(events.DefaultWorkspace))
	})

	t.Run("error locking work dir", func(t *testing.T) {
		preWorkflowHooksSetup(t)

		globalCfg := valid.GlobalCfg{
			Repos: []valid.Repo{
				{
					ID: testdata.GithubRepo.ID(),
					PreWorkflowHooks: []*valid.WorkflowHook{
						&testHook,
					},
				},
			},
		}

		preWh.GlobalCfg = globalCfg

		When(preWhWorkingDirLocker.TryLock(testdata.GithubRepo.FullName, newPull.Num, events.DefaultWorkspace,
			events.DefaultRepoRelDir, "", command.Plan, events.WorkingDirLockMetadataForPull(newPull))).ThenReturn(func() {}, errors.New("some error"))

		err := preWh.RunPreHooks(ctx, planCmd)

		Assert(t, err != nil, "error not nil")
		preWhWorkingDir.VerifyWasCalled(Never()).Clone(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(newPull),
			Eq(events.DefaultWorkspace))
		whPreWorkflowHookRunner.VerifyWasCalled(Never()).Run(Any[models.WorkflowHookCommandContext](), Eq(testHook.RunCommand),
			Eq(defaultShell), Eq(defaultShellArgs), Eq(repoDir))
	})

	t.Run("error cloning", func(t *testing.T) {
		preWorkflowHooksSetup(t)

		var unlockCalled = newBool(false)
		unlockFn := func() {
			unlockCalled = newBool(true)
		}

		globalCfg := valid.GlobalCfg{
			Repos: []valid.Repo{
				{
					ID: testdata.GithubRepo.ID(),
					PreWorkflowHooks: []*valid.WorkflowHook{
						&testHook,
					},
				},
			},
		}

		preWh.GlobalCfg = globalCfg

		When(preWhWorkingDirLocker.TryLock(testdata.GithubRepo.FullName, newPull.Num, events.DefaultWorkspace,
			events.DefaultRepoRelDir, "", command.Plan, events.WorkingDirLockMetadataForPull(newPull))).ThenReturn(unlockFn, nil)
		When(preWhWorkingDir.Clone(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(newPull),
			Eq(events.DefaultWorkspace))).ThenReturn(repoDir, errors.New("some error"))

		err := preWh.RunPreHooks(ctx, planCmd)

		Assert(t, err != nil, "error not nil")

		whPreWorkflowHookRunner.VerifyWasCalled(Never()).Run(Any[models.WorkflowHookCommandContext](), Eq(testHook.RunCommand),
			Eq(defaultShell), Eq(defaultShellArgs), Eq(repoDir))
		Assert(t, *unlockCalled == true, "unlock function called")
	})

	t.Run("error running pre hook", func(t *testing.T) {
		preWorkflowHooksSetup(t)

		var unlockCalled = newBool(false)
		unlockFn := func() {
			unlockCalled = newBool(true)
		}

		globalCfg := valid.GlobalCfg{
			Repos: []valid.Repo{
				{
					ID: testdata.GithubRepo.ID(),
					PreWorkflowHooks: []*valid.WorkflowHook{
						&testHook,
					},
				},
			},
		}

		preWh.GlobalCfg = globalCfg

		When(preWhWorkingDirLocker.TryLock(testdata.GithubRepo.FullName, newPull.Num, events.DefaultWorkspace,
			events.DefaultRepoRelDir, "", command.Plan, events.WorkingDirLockMetadataForPull(newPull))).ThenReturn(unlockFn, nil)
		When(preWhWorkingDir.Clone(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(newPull),
			Eq(events.DefaultWorkspace))).ThenReturn(repoDir, nil)
		When(whPreWorkflowHookRunner.Run(Any[models.WorkflowHookCommandContext](), Eq(testHook.RunCommand),
			Any[string](), Any[string](), Eq(repoDir))).ThenReturn(result, runtimeDesc, errors.New("some error"))

		err := preWh.RunPreHooks(ctx, planCmd)

		Assert(t, err != nil, "error not nil")
		Assert(t, *unlockCalled == true, "unlock function called")
	})

	t.Run("comment args and projectname passed to webhooks", func(t *testing.T) {
		preWorkflowHooksSetup(t)

		var unlockCalled = newBool(false)
		unlockFn := func() {
			unlockCalled = newBool(true)
		}

		globalCfg := valid.GlobalCfg{
			Repos: []valid.Repo{
				{
					ID: testdata.GithubRepo.ID(),
					PreWorkflowHooks: []*valid.WorkflowHook{
						&testHook,
					},
				},
			},
		}

		planCmd := &events.CommentCommand{
			Name:        command.Plan,
			Flags:       []string{"comment", "args"},
			ProjectName: "*",
		}

		expectedCtx := pCtx
		expectedCtx.HookDescription = "Pre workflow hook #0"
		expectedCtx.HookStepName = "pre plan #0"
		expectedCtx.EscapedCommentArgs = []string{"\\c\\o\\m\\m\\e\\n\\t", "\\a\\r\\g\\s"}
		expectedCtx.ProjectName = "*"

		preWh.GlobalCfg = globalCfg

		When(preWhWorkingDirLocker.TryLock(testdata.GithubRepo.FullName, newPull.Num, events.DefaultWorkspace,
			events.DefaultRepoRelDir, "", command.Plan, events.WorkingDirLockMetadataForPull(newPull))).ThenReturn(unlockFn, nil)
		When(preWhWorkingDir.Clone(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(newPull),
			Eq(events.DefaultWorkspace))).ThenReturn(repoDir, nil)
		When(whPreWorkflowHookRunner.Run(
			ArgThat[models.WorkflowHookCommandContext](WorkflowHookCommandContextMatcher{expected: expectedCtx}),
			Eq(testHook.RunCommand),
			Any[string](),
			Any[string](),
			Eq(repoDir),
		)).ThenReturn(result, runtimeDesc, nil)

		err := preWh.RunPreHooks(ctx, planCmd)

		Ok(t, err)
		whPreWorkflowHookRunner.VerifyWasCalledOnce().Run(
			ArgThat[models.WorkflowHookCommandContext](WorkflowHookCommandContextMatcher{expected: expectedCtx}),
			Eq(testHook.RunCommand),
			Eq(defaultShell),
			Eq(defaultShellArgs),
			Eq(repoDir))
		Assert(t, *unlockCalled == true, "unlock function called")
	})

	t.Run("shell passed to webhooks", func(t *testing.T) {
		preWorkflowHooksSetup(t)

		var unlockCalled = newBool(false)
		unlockFn := func() {
			unlockCalled = newBool(true)
		}

		globalCfg := valid.GlobalCfg{
			Repos: []valid.Repo{
				{
					ID: testdata.GithubRepo.ID(),
					PreWorkflowHooks: []*valid.WorkflowHook{
						&testHookWithShell,
					},
				},
			},
		}

		preWh.GlobalCfg = globalCfg

		When(preWhWorkingDirLocker.TryLock(testdata.GithubRepo.FullName, newPull.Num, events.DefaultWorkspace,
			events.DefaultRepoRelDir, "", command.Plan, events.WorkingDirLockMetadataForPull(newPull))).ThenReturn(unlockFn, nil)
		When(preWhWorkingDir.Clone(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(newPull),
			Eq(events.DefaultWorkspace))).ThenReturn(repoDir, nil)
		When(whPreWorkflowHookRunner.Run(Any[models.WorkflowHookCommandContext](), Eq(testHookWithShell.RunCommand),
			Any[string](), Any[string](), Eq(repoDir))).ThenReturn(result, runtimeDesc, nil)

		err := preWh.RunPreHooks(ctx, planCmd)

		Ok(t, err)
		whPreWorkflowHookRunner.VerifyWasCalledOnce().Run(Any[models.WorkflowHookCommandContext](),
			Eq(testHookWithShell.RunCommand), Eq(testHookWithShell.Shell), Eq(defaultShellArgs), Eq(repoDir))
		Assert(t, *unlockCalled == true, "unlock function called")
	})

	t.Run("shellArgs passed to webhooks", func(t *testing.T) {
		preWorkflowHooksSetup(t)

		var unlockCalled = newBool(false)
		unlockFn := func() {
			unlockCalled = newBool(true)
		}

		globalCfg := valid.GlobalCfg{
			Repos: []valid.Repo{
				{
					ID: testdata.GithubRepo.ID(),
					PreWorkflowHooks: []*valid.WorkflowHook{
						&testHookWithShellArgs,
					},
				},
			},
		}

		preWh.GlobalCfg = globalCfg

		When(preWhWorkingDirLocker.TryLock(testdata.GithubRepo.FullName, newPull.Num, events.DefaultWorkspace,
			events.DefaultRepoRelDir, "", command.Plan, events.WorkingDirLockMetadataForPull(newPull))).ThenReturn(unlockFn, nil)
		When(preWhWorkingDir.Clone(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(newPull),
			Eq(events.DefaultWorkspace))).ThenReturn(repoDir, nil)
		When(whPreWorkflowHookRunner.Run(Any[models.WorkflowHookCommandContext](), Eq(testHook.RunCommand),
			Any[string](), Any[string](), Eq(repoDir))).ThenReturn(result, runtimeDesc, nil)

		err := preWh.RunPreHooks(ctx, planCmd)

		Ok(t, err)
		whPreWorkflowHookRunner.VerifyWasCalledOnce().Run(Any[models.WorkflowHookCommandContext](),
			Eq(testHookWithShellArgs.RunCommand), Eq(defaultShell), Eq(testHookWithShellArgs.ShellArgs), Eq(repoDir))
		Assert(t, *unlockCalled == true, "unlock function called")
	})

	t.Run("Shell and ShellArgs passed to webhooks", func(t *testing.T) {
		preWorkflowHooksSetup(t)

		var unlockCalled = newBool(false)
		unlockFn := func() {
			unlockCalled = newBool(true)
		}

		globalCfg := valid.GlobalCfg{
			Repos: []valid.Repo{
				{
					ID: testdata.GithubRepo.ID(),
					PreWorkflowHooks: []*valid.WorkflowHook{
						&testHookWithShellandShellArgs,
					},
				},
			},
		}

		preWh.GlobalCfg = globalCfg

		When(preWhWorkingDirLocker.TryLock(testdata.GithubRepo.FullName, newPull.Num, events.DefaultWorkspace,
			events.DefaultRepoRelDir, "", command.Plan, events.WorkingDirLockMetadataForPull(newPull))).ThenReturn(unlockFn, nil)
		When(preWhWorkingDir.Clone(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(newPull),
			Eq(events.DefaultWorkspace))).ThenReturn(repoDir, nil)
		When(whPreWorkflowHookRunner.Run(Any[models.WorkflowHookCommandContext](), Eq(testHookWithShellandShellArgs.RunCommand),
			Any[string](), Any[string](), Eq(repoDir))).ThenReturn(result, runtimeDesc, nil)

		err := preWh.RunPreHooks(ctx, planCmd)

		Ok(t, err)
		whPreWorkflowHookRunner.VerifyWasCalledOnce().Run(Any[models.WorkflowHookCommandContext](),
			Eq(testHookWithShellandShellArgs.RunCommand), Eq(testHookWithShellandShellArgs.Shell),
			Eq(testHookWithShellandShellArgs.ShellArgs), Eq(repoDir))
		Assert(t, *unlockCalled == true, "unlock function called")
	})

	t.Run("Commands 'plan' set on webhook and plan command", func(t *testing.T) {
		preWorkflowHooksSetup(t)

		var unlockCalled = newBool(false)
		unlockFn := func() {
			unlockCalled = newBool(true)
		}

		globalCfg := valid.GlobalCfg{
			Repos: []valid.Repo{
				{
					ID: testdata.GithubRepo.ID(),
					PreWorkflowHooks: []*valid.WorkflowHook{
						&testHookWithPlanCommand,
					},
				},
			},
		}

		preWh.GlobalCfg = globalCfg

		When(preWhWorkingDirLocker.TryLock(testdata.GithubRepo.FullName, newPull.Num, events.DefaultWorkspace,
			events.DefaultRepoRelDir, "", command.Plan, events.WorkingDirLockMetadataForPull(newPull))).ThenReturn(unlockFn, nil)
		When(preWhWorkingDir.Clone(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(newPull),
			Eq(events.DefaultWorkspace))).ThenReturn(repoDir, nil)
		When(whPreWorkflowHookRunner.Run(Any[models.WorkflowHookCommandContext](), Eq(testHookWithPlanCommand.RunCommand),
			Any[string](), Any[string](), Eq(repoDir))).ThenReturn(result, runtimeDesc, nil)

		err := preWh.RunPreHooks(ctx, planCmd)

		Ok(t, err)
		whPreWorkflowHookRunner.VerifyWasCalledOnce().Run(Any[models.WorkflowHookCommandContext](),
			Eq(testHookWithPlanCommand.RunCommand), Any[string](), Any[string](), Eq(repoDir))
		Assert(t, *unlockCalled == true, "unlock function called")
	})

	t.Run("Commands 'plan' set on webhook and non-plan command", func(t *testing.T) {
		preWorkflowHooksSetup(t)

		var unlockCalled = newBool(false)
		unlockFn := func() {
			unlockCalled = newBool(true)
		}

		globalCfg := valid.GlobalCfg{
			Repos: []valid.Repo{
				{
					ID: testdata.GithubRepo.ID(),
					PreWorkflowHooks: []*valid.WorkflowHook{
						&testHookWithPlanCommand,
					},
				},
			},
		}

		preWh.GlobalCfg = globalCfg

		When(preWhWorkingDirLocker.TryLock(testdata.GithubRepo.FullName, newPull.Num, events.DefaultWorkspace,
			events.DefaultRepoRelDir, "", command.Apply, events.WorkingDirLockMetadataForPull(newPull))).ThenReturn(unlockFn, nil)
		When(preWhWorkingDir.Clone(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(newPull),
			Eq(events.DefaultWorkspace))).ThenReturn(repoDir, nil)
		When(whPreWorkflowHookRunner.Run(Any[models.WorkflowHookCommandContext](), Eq(testHookWithPlanCommand.RunCommand),
			Any[string](), Any[string](), Eq(repoDir))).ThenReturn(result, runtimeDesc, nil)

		err := preWh.RunPreHooks(ctx, applyCmd)

		Ok(t, err)
		whPreWorkflowHookRunner.VerifyWasCalled(Never()).Run(Any[models.WorkflowHookCommandContext](),
			Eq(testHookWithPlanCommand.RunCommand), Any[string](), Any[string](), Eq(repoDir))
		Assert(t, *unlockCalled == true, "unlock function called")
	})

	t.Run("Commands 'plan, apply' set on webhook and plan command", func(t *testing.T) {
		preWorkflowHooksSetup(t)

		var unlockCalled = newBool(false)
		unlockFn := func() {
			unlockCalled = newBool(true)
		}

		globalCfg := valid.GlobalCfg{
			Repos: []valid.Repo{
				{
					ID: testdata.GithubRepo.ID(),
					PreWorkflowHooks: []*valid.WorkflowHook{
						&testHookWithPlanApplyCommands,
					},
				},
			},
		}

		preWh.GlobalCfg = globalCfg

		When(preWhWorkingDirLocker.TryLock(testdata.GithubRepo.FullName, newPull.Num, events.DefaultWorkspace,
			events.DefaultRepoRelDir, "", command.Plan, events.WorkingDirLockMetadataForPull(newPull))).ThenReturn(unlockFn, nil)
		When(preWhWorkingDir.Clone(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(newPull),
			Eq(events.DefaultWorkspace))).ThenReturn(repoDir, nil)
		When(whPreWorkflowHookRunner.Run(Any[models.WorkflowHookCommandContext](), Eq(testHookWithPlanApplyCommands.RunCommand),
			Any[string](), Any[string](), Eq(repoDir))).ThenReturn(result, runtimeDesc, nil)

		err := preWh.RunPreHooks(ctx, planCmd)

		Ok(t, err)
		whPreWorkflowHookRunner.VerifyWasCalledOnce().Run(Any[models.WorkflowHookCommandContext](),
			Eq(testHookWithPlanApplyCommands.RunCommand), Any[string](), Any[string](), Eq(repoDir))
		Assert(t, *unlockCalled == true, "unlock function called")
	})
}

func TestRunPreHooksSuppressesVCSStatus(t *testing.T) {
	for _, c := range []struct {
		name string
		err  error
	}{
		{name: "success"},
		{name: "failure", err: errors.New("hook failed")},
	} {
		t.Run(c.name, func(t *testing.T) {
			preWorkflowHooksSetup(t)
			log := logging.NewNoopLogger(t)
			var newPull = testdata.Pull
			newPull.BaseRepo = testdata.GithubRepo
			ctx := &command.Context{
				Pull:              newPull,
				HeadRepo:          testdata.GithubRepo,
				User:              testdata.User,
				Log:               log,
				SuppressVCSStatus: true,
				SuppressJobOutput: true,
			}
			planCmd := &events.CommentCommand{Name: command.Plan}
			hook := valid.WorkflowHook{RunCommand: "echo hook"}
			repoDir := "path/to/repo"
			preWh.GlobalCfg = valid.GlobalCfg{Repos: []valid.Repo{{
				ID:               testdata.GithubRepo.ID(),
				PreWorkflowHooks: []*valid.WorkflowHook{&hook},
			}}}
			When(preWhWorkingDirLocker.TryLock(testdata.GithubRepo.FullName, newPull.Num, events.DefaultWorkspace,
				events.DefaultRepoRelDir, "", command.Plan, events.WorkingDirLockMetadataForPull(newPull))).ThenReturn(func() {}, nil)
			When(preWhWorkingDir.Clone(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(newPull),
				Eq(events.DefaultWorkspace))).ThenReturn(repoDir, nil)
			var capturedHookCtx models.WorkflowHookCommandContext
			When(whPreWorkflowHookRunner.Run(Any[models.WorkflowHookCommandContext](), Eq(hook.RunCommand),
				Any[string](), Any[string](), Eq(repoDir))).Then(func(args []Param) ReturnValues {
				capturedHookCtx = args[0].(models.WorkflowHookCommandContext)
				return ReturnValues{"", "", c.err}
			})

			err := preWh.RunPreHooks(ctx, planCmd)
			if c.err != nil {
				ErrContains(t, c.err.Error(), err)
			} else {
				Ok(t, err)
			}
			whPreWorkflowHookRunner.VerifyWasCalledOnce().Run(Any[models.WorkflowHookCommandContext](), Eq(hook.RunCommand),
				Any[string](), Any[string](), Eq(repoDir))
			Assert(t, capturedHookCtx.SuppressJobOutput, "expected hook runtime context to suppress job output")
			preCommitStatusUpdater.VerifyWasCalled(Never()).UpdatePreWorkflowHook(Any[logging.SimpleLogging](),
				Any[models.PullRequest](), Any[models.CommitStatus](), Any[string](), Any[string](), Any[string]())
		})
	}
}

func preWorkflowHooksLockSetup(t *testing.T) (*command.Context, *events.DefaultWorkingDirLocker) {
	preWorkflowHooksSetup(t)
	pull := testdata.Pull
	pull.BaseRepo = testdata.GithubRepo
	ctx := &command.Context{
		Pull:     pull,
		HeadRepo: testdata.GithubRepo,
		User:     testdata.User,
		Log:      logging.NewNoopLogger(t),
	}
	locker := events.NewDefaultWorkingDirLocker()
	preWh.WorkingDirLocker = locker
	preWh.GlobalCfg = valid.GlobalCfg{Repos: []valid.Repo{{
		ID:               pull.BaseRepo.ID(),
		PreWorkflowHooks: []*valid.WorkflowHook{{RunCommand: "echo hook"}},
	}}}
	When(preWhWorkingDir.Clone(Any[logging.SimpleLogging](), Eq(ctx.HeadRepo), Eq(pull),
		Eq(events.DefaultWorkspace))).ThenReturn("path/to/repo", nil)
	return ctx, locker
}

func TestRunPreHooks_DoesNotLockActiveCommandWorkspace(t *testing.T) {
	for _, c := range []struct {
		name   string
		active command.Name
		hook   command.Name
	}{
		{name: "apply during plan", active: command.Plan, hook: command.Apply},
		{name: "plan during apply", active: command.Apply, hook: command.Plan},
		{name: "autoplan during apply", active: command.Apply, hook: command.Autoplan},
	} {
		t.Run(c.name, func(t *testing.T) {
			ctx, locker := preWorkflowHooksLockSetup(t)
			metadata := events.WorkingDirLockMetadataForPull(ctx.Pull)
			unlockPull, err := locker.TryLockPull(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num, c.active, metadata)
			Ok(t, err)
			defer unlockPull()

			err = preWh.RunPreHooks(ctx, &events.CommentCommand{Name: c.hook})

			ErrContains(t, "pull request", err)
			preWhWorkingDir.VerifyWasCalled(Never()).Clone(Any[logging.SimpleLogging](), Any[models.Repo](),
				Any[models.PullRequest](), Any[string]())
			whPreWorkflowHookRunner.VerifyWasCalled(Never()).Run(Any[models.WorkflowHookCommandContext](),
				Any[string](), Any[string](), Any[string](), Any[string]())
			unlockWorkspace, err := locker.TryLock(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num, events.DefaultWorkspace,
				events.DefaultRepoRelDir, "", c.active, metadata)
			Ok(t, err)
			unlockWorkspace()
			Assert(t, locker.HasCommandLock(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num, c.active), "active command must retain its lock")
		})
	}
}

func TestRunPreHooks_BlocksCommandDuringHook(t *testing.T) {
	for _, cmdName := range []command.Name{command.Plan, command.Autoplan, command.Apply} {
		name := cmdName.String()
		if cmdName == command.Autoplan {
			name = "autoplan"
		}
		t.Run(name, func(t *testing.T) {
			ctx, locker := preWorkflowHooksLockSetup(t)
			metadata := events.WorkingDirLockMetadataForPull(ctx.Pull)
			When(whPreWorkflowHookRunner.Run(Any[models.WorkflowHookCommandContext](), Any[string](),
				Any[string](), Any[string](), Any[string]())).Then(func([]Param) ReturnValues {
				unlock, err := locker.TryLockPull(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num, command.Plan, metadata)
				defer unlock()
				ErrContains(t, "pull request", err)
				otherUnlock, err := locker.TryLockPull(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num+1, command.Plan, metadata)
				Ok(t, err)
				otherUnlock()
				return ReturnValues{"", "", nil}
			})

			Ok(t, preWh.RunPreHooks(ctx, &events.CommentCommand{Name: cmdName}))
			whPreWorkflowHookRunner.VerifyWasCalledOnce().Run(Any[models.WorkflowHookCommandContext](), Any[string](),
				Any[string](), Any[string](), Any[string]())
		})
	}
}

func TestRunPreHooks_ReleasesLocks(t *testing.T) {
	for _, c := range []struct {
		name          string
		workspaceBusy bool
		cloneErr      error
		hookErr       error
	}{
		{name: "success"},
		{name: "workspace locked", workspaceBusy: true},
		{name: "clone error", cloneErr: errors.New("clone error")},
		{name: "hook error", hookErr: errors.New("hook error")},
	} {
		t.Run(c.name, func(t *testing.T) {
			ctx, locker := preWorkflowHooksLockSetup(t)
			metadata := events.WorkingDirLockMetadataForPull(ctx.Pull)
			var releaseWorkspace func()
			if c.workspaceBusy {
				var err error
				releaseWorkspace, err = locker.TryLock(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num, events.DefaultWorkspace,
					events.DefaultRepoRelDir, "", command.Apply, metadata)
				Ok(t, err)
				defer releaseWorkspace()
			}
			When(preWhWorkingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](),
				Any[string]())).ThenReturn("path/to/repo", c.cloneErr)
			When(whPreWorkflowHookRunner.Run(Any[models.WorkflowHookCommandContext](), Any[string](),
				Any[string](), Any[string](), Any[string]())).ThenReturn("", "", c.hookErr)

			err := preWh.RunPreHooks(ctx, &events.CommentCommand{Name: command.Apply})

			if c.workspaceBusy || c.cloneErr != nil || c.hookErr != nil {
				Assert(t, err != nil, "expected hook to fail")
			} else {
				Ok(t, err)
			}
			unlock, err := locker.TryLockPull(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num, command.Apply, metadata)
			Ok(t, err)
			unlock()
			if releaseWorkspace != nil {
				releaseWorkspace()
			}
			When(preWhWorkingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](),
				Any[string]())).ThenReturn("path/to/repo", nil)
			When(whPreWorkflowHookRunner.Run(Any[models.WorkflowHookCommandContext](), Any[string](),
				Any[string](), Any[string](), Any[string]())).ThenReturn("", "", nil)
			Ok(t, preWh.RunPreHooks(ctx, &events.CommentCommand{Name: command.Apply}))
		})
	}
}

func TestRunPreHooks_ControlCommandsDoNotTakePullLock(t *testing.T) {
	for _, cmdName := range []command.Name{command.Cancel, command.Unlock} {
		t.Run(cmdName.String(), func(t *testing.T) {
			ctx, locker := preWorkflowHooksLockSetup(t)
			unlock, err := locker.TryLockPull(ctx.Pull.BaseRepo.FullName, ctx.Pull.Num, command.Plan,
				events.WorkingDirLockMetadataForPull(ctx.Pull))
			Ok(t, err)
			defer unlock()

			Ok(t, preWh.RunPreHooks(ctx, &events.CommentCommand{Name: cmdName}))
			whPreWorkflowHookRunner.VerifyWasCalledOnce().Run(Any[models.WorkflowHookCommandContext](), Any[string](),
				Any[string](), Any[string](), Any[string]())
		})
	}
}
