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

	cmd := &events.CommentCommand{
		Name: command.Plan,
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

		When(preWhWorkingDirLocker.TryLock(testdata.GithubRepo.FullName, newPull.Num, events.DefaultWorkspace, events.DefaultRepoRelDir)).ThenReturn(unlockFn, nil)
		When(preWhWorkingDir.Clone(testdata.GithubRepo, newPull, events.DefaultWorkspace)).ThenReturn(repoDir, false, nil)
		When(whPreWorkflowHookRunner.Run(Any[models.WorkflowHookCommandContext](), Eq(testHook.RunCommand),
			Any[string](), Any[string](), Eq(repoDir))).ThenReturn(result, runtimeDesc, nil)

		err := preWh.RunPreHooks(ctx, cmd)

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

		err := preWh.RunPreHooks(ctx, cmd)

		Ok(t, err)

		whPreWorkflowHookRunner.VerifyWasCalled(Never()).Run(Any[models.WorkflowHookCommandContext](), Eq(testHook.RunCommand), Eq(defaultShell), Eq(defaultShellArgs), Eq(repoDir))
		preWhWorkingDirLocker.VerifyWasCalled(Never()).TryLock(testdata.GithubRepo.FullName, newPull.Num, events.DefaultWorkspace, "")
		preWhWorkingDir.VerifyWasCalled(Never()).Clone(testdata.GithubRepo, newPull, events.DefaultWorkspace)
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

		When(preWhWorkingDirLocker.TryLock(testdata.GithubRepo.FullName, newPull.Num, events.DefaultWorkspace, events.DefaultRepoRelDir)).ThenReturn(func() {}, errors.New("some error"))

		err := preWh.RunPreHooks(ctx, cmd)

		Assert(t, err != nil, "error not nil")
		preWhWorkingDir.VerifyWasCalled(Never()).Clone(testdata.GithubRepo, newPull, events.DefaultWorkspace)
		whPreWorkflowHookRunner.VerifyWasCalled(Never()).Run(Any[models.WorkflowHookCommandContext](), Eq(testHook.RunCommand), Eq(defaultShell), Eq(defaultShellArgs), Eq(repoDir))
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

		When(preWhWorkingDirLocker.TryLock(testdata.GithubRepo.FullName, newPull.Num, events.DefaultWorkspace, events.DefaultRepoRelDir)).ThenReturn(unlockFn, nil)
		When(preWhWorkingDir.Clone(testdata.GithubRepo, newPull, events.DefaultWorkspace)).ThenReturn(repoDir, false, errors.New("some error"))

		err := preWh.RunPreHooks(ctx, cmd)

		Assert(t, err != nil, "error not nil")

		whPreWorkflowHookRunner.VerifyWasCalled(Never()).Run(Any[models.WorkflowHookCommandContext](), Eq(testHook.RunCommand), Eq(defaultShell), Eq(defaultShellArgs), Eq(repoDir))
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

		When(preWhWorkingDirLocker.TryLock(testdata.GithubRepo.FullName, newPull.Num, events.DefaultWorkspace, events.DefaultRepoRelDir)).ThenReturn(unlockFn, nil)
		When(preWhWorkingDir.Clone(testdata.GithubRepo, newPull, events.DefaultWorkspace)).ThenReturn(repoDir, false, nil)
		When(whPreWorkflowHookRunner.Run(Any[models.WorkflowHookCommandContext](), Eq(testHook.RunCommand),
			Any[string](), Any[string](), Eq(repoDir))).ThenReturn(result, runtimeDesc, errors.New("some error"))

		err := preWh.RunPreHooks(ctx, cmd)

		Assert(t, err != nil, "error not nil")
		Assert(t, *unlockCalled == true, "unlock function called")
	})

	t.Run("comment args passed to webhooks", func(t *testing.T) {
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

		cmd := &events.CommentCommand{
			Name:  command.Plan,
			Flags: []string{"comment", "args"},
		}

		expectedCtx := pCtx
		expectedCtx.EscapedCommentArgs = []string{"\\c\\o\\m\\m\\e\\n\\t", "\\a\\r\\g\\s"}

		preWh.GlobalCfg = globalCfg

		When(preWhWorkingDirLocker.TryLock(testdata.GithubRepo.FullName, newPull.Num, events.DefaultWorkspace, events.DefaultRepoRelDir)).ThenReturn(unlockFn, nil)
		When(preWhWorkingDir.Clone(testdata.GithubRepo, newPull, events.DefaultWorkspace)).ThenReturn(repoDir, false, nil)
		When(whPreWorkflowHookRunner.Run(Any[models.WorkflowHookCommandContext](), Eq(testHook.RunCommand), Any[string](), Any[string](), Eq(repoDir))).ThenReturn(result, runtimeDesc, nil)

		err := preWh.RunPreHooks(ctx, cmd)

		Ok(t, err)
		whPreWorkflowHookRunner.VerifyWasCalledOnce().Run(Any[models.WorkflowHookCommandContext](), Eq(testHook.RunCommand), Eq(defaultShell), Eq(defaultShellArgs), Eq(repoDir))
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

		When(preWhWorkingDirLocker.TryLock(testdata.GithubRepo.FullName, newPull.Num, events.DefaultWorkspace, events.DefaultRepoRelDir)).ThenReturn(unlockFn, nil)
		When(preWhWorkingDir.Clone(testdata.GithubRepo, newPull, events.DefaultWorkspace)).ThenReturn(repoDir, false, nil)
		When(whPreWorkflowHookRunner.Run(Any[models.WorkflowHookCommandContext](), Eq(testHookWithShell.RunCommand),
			Any[string](), Any[string](), Eq(repoDir))).ThenReturn(result, runtimeDesc, nil)

		err := preWh.RunPreHooks(ctx, cmd)

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

		When(preWhWorkingDirLocker.TryLock(testdata.GithubRepo.FullName, newPull.Num, events.DefaultWorkspace, events.DefaultRepoRelDir)).ThenReturn(unlockFn, nil)
		When(preWhWorkingDir.Clone(testdata.GithubRepo, newPull, events.DefaultWorkspace)).ThenReturn(repoDir, false, nil)
		When(whPreWorkflowHookRunner.Run(Any[models.WorkflowHookCommandContext](), Eq(testHook.RunCommand),
			Any[string](), Any[string](), Eq(repoDir))).ThenReturn(result, runtimeDesc, nil)

		err := preWh.RunPreHooks(ctx, cmd)

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

		When(preWhWorkingDirLocker.TryLock(testdata.GithubRepo.FullName, newPull.Num, events.DefaultWorkspace, events.DefaultRepoRelDir)).ThenReturn(unlockFn, nil)
		When(preWhWorkingDir.Clone(testdata.GithubRepo, newPull, events.DefaultWorkspace)).ThenReturn(repoDir, false, nil)
		When(whPreWorkflowHookRunner.Run(Any[models.WorkflowHookCommandContext](),
			Eq(testHookWithShellandShellArgs.RunCommand), Any[string](), Any[string](), Eq(repoDir))).ThenReturn(result, runtimeDesc, nil)

		err := preWh.RunPreHooks(ctx, cmd)

		Ok(t, err)
		whPreWorkflowHookRunner.VerifyWasCalledOnce().Run(Any[models.WorkflowHookCommandContext](),
			Eq(testHookWithShellandShellArgs.RunCommand), Eq(testHookWithShellandShellArgs.Shell),
			Eq(testHookWithShellandShellArgs.ShellArgs), Eq(repoDir))
		Assert(t, *unlockCalled == true, "unlock function called")
	})
}
