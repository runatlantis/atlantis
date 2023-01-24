package events_test

import (
	"errors"
	"testing"

	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	runtime_mocks "github.com/runatlantis/atlantis/server/core/runtime/mocks"
	runtimematchers "github.com/runatlantis/atlantis/server/core/runtime/mocks/matchers"
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

	testHook := valid.WorkflowHook{
		StepName:   "test",
		RunCommand: "some command",
	}

	repoDir := "path/to/repo"
	result := "some result"
	runtimeDesc := ""

	pCtx := models.WorkflowHookCommandContext{
		BaseRepo: testdata.GithubRepo,
		HeadRepo: testdata.GithubRepo,
		Pull:     newPull,
		Log:      log,
		User:     testdata.User,
		Verbose:  false,
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
		When(preWhWorkingDir.Clone(log, testdata.GithubRepo, newPull, events.DefaultWorkspace)).ThenReturn(repoDir, false, nil)
		When(whPreWorkflowHookRunner.Run(runtimematchers.AnyModelsWorkflowHookCommandContext(), EqString(testHook.RunCommand), EqString(repoDir))).ThenReturn(result, runtimeDesc, nil)

		err := preWh.RunPreHooks(ctx, nil)

		Ok(t, err)
		whPreWorkflowHookRunner.VerifyWasCalledOnce().Run(runtimematchers.AnyModelsWorkflowHookCommandContext(), EqString(testHook.RunCommand), EqString(repoDir))
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

		err := preWh.RunPreHooks(ctx, nil)

		Ok(t, err)

		whPreWorkflowHookRunner.VerifyWasCalled(Never()).Run(runtimematchers.AnyModelsWorkflowHookCommandContext(), EqString(testHook.RunCommand), EqString(repoDir))
		preWhWorkingDirLocker.VerifyWasCalled(Never()).TryLock(testdata.GithubRepo.FullName, newPull.Num, events.DefaultWorkspace, "")
		preWhWorkingDir.VerifyWasCalled(Never()).Clone(log, testdata.GithubRepo, newPull, events.DefaultWorkspace)
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

		err := preWh.RunPreHooks(ctx, nil)

		Assert(t, err != nil, "error not nil")
		preWhWorkingDir.VerifyWasCalled(Never()).Clone(log, testdata.GithubRepo, newPull, events.DefaultWorkspace)
		whPreWorkflowHookRunner.VerifyWasCalled(Never()).Run(runtimematchers.AnyModelsWorkflowHookCommandContext(), EqString(testHook.RunCommand), EqString(repoDir))
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
		When(preWhWorkingDir.Clone(log, testdata.GithubRepo, newPull, events.DefaultWorkspace)).ThenReturn(repoDir, false, errors.New("some error"))

		err := preWh.RunPreHooks(ctx, nil)

		Assert(t, err != nil, "error not nil")

		whPreWorkflowHookRunner.VerifyWasCalled(Never()).Run(runtimematchers.AnyModelsWorkflowHookCommandContext(), EqString(testHook.RunCommand), EqString(repoDir))
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
		When(preWhWorkingDir.Clone(log, testdata.GithubRepo, newPull, events.DefaultWorkspace)).ThenReturn(repoDir, false, nil)
		When(whPreWorkflowHookRunner.Run(runtimematchers.AnyModelsWorkflowHookCommandContext(), EqString(testHook.RunCommand), EqString(repoDir))).ThenReturn(result, runtimeDesc, errors.New("some error"))

		err := preWh.RunPreHooks(ctx, nil)

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
			Flags: []string{"comment", "args"},
		}

		expectedCtx := pCtx
		expectedCtx.EscapedCommentArgs = []string{"\\c\\o\\m\\m\\e\\n\\t", "\\a\\r\\g\\s"}

		preWh.GlobalCfg = globalCfg

		When(preWhWorkingDirLocker.TryLock(testdata.GithubRepo.FullName, newPull.Num, events.DefaultWorkspace, events.DefaultRepoRelDir)).ThenReturn(unlockFn, nil)
		When(preWhWorkingDir.Clone(log, testdata.GithubRepo, newPull, events.DefaultWorkspace)).ThenReturn(repoDir, false, nil)
		When(whPreWorkflowHookRunner.Run(runtimematchers.AnyModelsWorkflowHookCommandContext(), EqString(testHook.RunCommand), EqString(repoDir))).ThenReturn(result, runtimeDesc, nil)

		err := preWh.RunPreHooks(ctx, cmd)

		Ok(t, err)
		whPreWorkflowHookRunner.VerifyWasCalledOnce().Run(runtimematchers.AnyModelsWorkflowHookCommandContext(), EqString(testHook.RunCommand), EqString(repoDir))
		Assert(t, *unlockCalled == true, "unlock function called")
	})
}
