package events_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	runtime_mocks "github.com/runatlantis/atlantis/server/core/runtime/mocks"
	runtimematchers "github.com/runatlantis/atlantis/server/core/runtime/mocks/matchers"

	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/models/testdata"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

var postWh events.DefaultPostWorkflowHooksCommandRunner
var postWhWorkingDir *mocks.MockWorkingDir
var postWhWorkingDirLocker *mocks.MockWorkingDirLocker
var whPostWorkflowHookRunner *runtime_mocks.MockPostWorkflowHookRunner
var postCommitStatusUpdater *mocks.MockCommitStatusUpdater

func postWorkflowHooksSetup(t *testing.T) {
	RegisterMockTestingT(t)
	vcsClient := vcsmocks.NewMockClient()
	postWhWorkingDir = mocks.NewMockWorkingDir()
	postWhWorkingDirLocker = mocks.NewMockWorkingDirLocker()
	whPostWorkflowHookRunner = runtime_mocks.NewMockPostWorkflowHookRunner()
	postCommitStatusUpdater = mocks.NewMockCommitStatusUpdater()
	postWorkflowHookURLGenerator := mocks.NewMockPostWorkflowHookURLGenerator()

	postWh = events.DefaultPostWorkflowHooksCommandRunner{
		VCSClient:              vcsClient,
		WorkingDirLocker:       postWhWorkingDirLocker,
		WorkingDir:             postWhWorkingDir,
		PostWorkflowHookRunner: whPostWorkflowHookRunner,
		CommitStatusUpdater:    postCommitStatusUpdater,
		Router:                 postWorkflowHookURLGenerator,
	}
}

func TestRunPostHooks_Clone(t *testing.T) {

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
		HookID:   uuid.NewString(),
	}

	t.Run("success hooks in cfg", func(t *testing.T) {
		postWorkflowHooksSetup(t)

		unlockCalled := newBool(false)
		unlockFn := func() {
			unlockCalled = newBool(true)
		}

		globalCfg := valid.GlobalCfg{
			Repos: []valid.Repo{
				{
					ID: testdata.GithubRepo.ID(),
					PostWorkflowHooks: []*valid.WorkflowHook{
						&testHook,
					},
				},
			},
		}

		postWh.GlobalCfg = globalCfg

		When(postWhWorkingDirLocker.TryLock(testdata.GithubRepo.FullName, newPull.Num, events.DefaultWorkspace, events.DefaultRepoRelDir)).ThenReturn(unlockFn, nil)
		When(postWhWorkingDir.Clone(log, testdata.GithubRepo, newPull, events.DefaultWorkspace)).ThenReturn(repoDir, false, nil)
		When(whPostWorkflowHookRunner.Run(pCtx, testHook.RunCommand, repoDir)).ThenReturn(result, runtimeDesc, nil)

		err := postWh.RunPostHooks(ctx, nil)

		Ok(t, err)
		whPostWorkflowHookRunner.VerifyWasCalledOnce().Run(runtimematchers.AnyModelsWorkflowHookCommandContext(), EqString(testHook.RunCommand), EqString(repoDir))
		Assert(t, *unlockCalled == true, "unlock function called")
	})
	t.Run("success hooks not in cfg", func(t *testing.T) {
		postWorkflowHooksSetup(t)
		globalCfg := valid.GlobalCfg{
			Repos: []valid.Repo{
				// one with hooks but mismatched id
				{
					ID: "id1",
					PostWorkflowHooks: []*valid.WorkflowHook{
						&testHook,
					},
				},
				// one with the correct id but no hooks
				{
					ID:                testdata.GithubRepo.ID(),
					PostWorkflowHooks: []*valid.WorkflowHook{},
				},
			},
		}

		postWh.GlobalCfg = globalCfg

		err := postWh.RunPostHooks(ctx, nil)

		Ok(t, err)

		whPostWorkflowHookRunner.VerifyWasCalled(Never()).Run(pCtx, testHook.RunCommand, repoDir)
		postWhWorkingDirLocker.VerifyWasCalled(Never()).TryLock(testdata.GithubRepo.FullName, newPull.Num, events.DefaultWorkspace, "path")
		postWhWorkingDir.VerifyWasCalled(Never()).Clone(log, testdata.GithubRepo, newPull, events.DefaultWorkspace)
	})
	t.Run("error locking work dir", func(t *testing.T) {
		postWorkflowHooksSetup(t)

		globalCfg := valid.GlobalCfg{
			Repos: []valid.Repo{
				{
					ID: testdata.GithubRepo.ID(),
					PostWorkflowHooks: []*valid.WorkflowHook{
						&testHook,
					},
				},
			},
		}

		postWh.GlobalCfg = globalCfg

		When(postWhWorkingDirLocker.TryLock(testdata.GithubRepo.FullName, newPull.Num, events.DefaultWorkspace, events.DefaultRepoRelDir)).ThenReturn(func() {}, errors.New("some error"))

		err := postWh.RunPostHooks(ctx, nil)

		Assert(t, err != nil, "error not nil")
		postWhWorkingDir.VerifyWasCalled(Never()).Clone(log, testdata.GithubRepo, newPull, events.DefaultWorkspace)
		whPostWorkflowHookRunner.VerifyWasCalled(Never()).Run(pCtx, testHook.RunCommand, repoDir)
	})

	t.Run("error cloning", func(t *testing.T) {
		postWorkflowHooksSetup(t)

		unlockCalled := newBool(false)
		unlockFn := func() {
			unlockCalled = newBool(true)
		}

		globalCfg := valid.GlobalCfg{
			Repos: []valid.Repo{
				{
					ID: testdata.GithubRepo.ID(),
					PostWorkflowHooks: []*valid.WorkflowHook{
						&testHook,
					},
				},
			},
		}

		postWh.GlobalCfg = globalCfg

		When(postWhWorkingDirLocker.TryLock(testdata.GithubRepo.FullName, newPull.Num, events.DefaultWorkspace, events.DefaultRepoRelDir)).ThenReturn(unlockFn, nil)
		When(postWhWorkingDir.Clone(log, testdata.GithubRepo, newPull, events.DefaultWorkspace)).ThenReturn(repoDir, false, errors.New("some error"))

		err := postWh.RunPostHooks(ctx, nil)

		Assert(t, err != nil, "error not nil")

		whPostWorkflowHookRunner.VerifyWasCalled(Never()).Run(pCtx, testHook.RunCommand, repoDir)
		Assert(t, *unlockCalled == true, "unlock function called")
	})

	t.Run("error running post hook", func(t *testing.T) {
		postWorkflowHooksSetup(t)

		unlockCalled := newBool(false)
		unlockFn := func() {
			unlockCalled = newBool(true)
		}

		globalCfg := valid.GlobalCfg{
			Repos: []valid.Repo{
				{
					ID: testdata.GithubRepo.ID(),
					PostWorkflowHooks: []*valid.WorkflowHook{
						&testHook,
					},
				},
			},
		}

		postWh.GlobalCfg = globalCfg

		When(postWhWorkingDirLocker.TryLock(testdata.GithubRepo.FullName, newPull.Num, events.DefaultWorkspace, events.DefaultRepoRelDir)).ThenReturn(unlockFn, nil)
		When(postWhWorkingDir.Clone(log, testdata.GithubRepo, newPull, events.DefaultWorkspace)).ThenReturn(repoDir, false, nil)
		When(whPostWorkflowHookRunner.Run(runtimematchers.AnyModelsWorkflowHookCommandContext(), EqString(testHook.RunCommand), EqString(repoDir))).ThenReturn(result, runtimeDesc, errors.New("some error"))

		err := postWh.RunPostHooks(ctx, nil)

		Assert(t, err != nil, "error not nil")
		Assert(t, *unlockCalled == true, "unlock function called")
	})

	t.Run("comment args passed to webhooks", func(t *testing.T) {
		postWorkflowHooksSetup(t)

		unlockCalled := newBool(false)
		unlockFn := func() {
			unlockCalled = newBool(true)
		}

		globalCfg := valid.GlobalCfg{
			Repos: []valid.Repo{
				{
					ID: testdata.GithubRepo.ID(),
					PostWorkflowHooks: []*valid.WorkflowHook{
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

		postWh.GlobalCfg = globalCfg

		When(postWhWorkingDirLocker.TryLock(testdata.GithubRepo.FullName, newPull.Num, events.DefaultWorkspace, events.DefaultRepoRelDir)).ThenReturn(unlockFn, nil)
		When(postWhWorkingDir.Clone(log, testdata.GithubRepo, newPull, events.DefaultWorkspace)).ThenReturn(repoDir, false, nil)
		When(whPostWorkflowHookRunner.Run(runtimematchers.AnyModelsWorkflowHookCommandContext(), EqString(testHook.RunCommand), EqString(repoDir))).ThenReturn(result, runtimeDesc, nil)

		err := postWh.RunPostHooks(ctx, cmd)

		Ok(t, err)
		whPostWorkflowHookRunner.VerifyWasCalledOnce().Run(runtimematchers.AnyModelsWorkflowHookCommandContext(), EqString(testHook.RunCommand), EqString(repoDir))
		Assert(t, *unlockCalled == true, "unlock function called")
	})
}
