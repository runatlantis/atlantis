package command_test

import (
	"context"
	"testing"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	lyftCommand "github.com/runatlantis/atlantis/server/lyft/command"
	"github.com/runatlantis/atlantis/server/lyft/feature"
	"github.com/runatlantis/atlantis/server/neptune/template"
	"github.com/stretchr/testify/assert"
)

type testAllocator struct {
	expectedFeatureName feature.Name
	expectedCtx         feature.FeatureContext
	expectedT           *testing.T

	expectedResult bool
	expectedErr    error
}

func (t *testAllocator) ShouldAllocate(name feature.Name, ctx feature.FeatureContext) (bool, error) {
	assert.Equal(t.expectedT, t.expectedFeatureName, name)
	assert.Equal(t.expectedT, t.expectedCtx, ctx)

	return t.expectedResult, t.expectedErr
}

type testRunner struct {
	expectedPlanResult            command.ProjectResult
	expectedPolicyCheckResult     command.ProjectResult
	expectedApplyResult           command.ProjectResult
	expectedApprovePoliciesResult command.ProjectResult
	expectedVersionResult         command.ProjectResult
}

// Plan runs terraform plan for the project described by ctx.
func (r *testRunner) Plan(ctx command.ProjectContext) command.ProjectResult {
	return r.expectedPlanResult
}

// PolicyCheck evaluates policies defined with Rego for the project described by ctx.
func (r *testRunner) PolicyCheck(ctx command.ProjectContext) command.ProjectResult {
	return r.expectedPolicyCheckResult
}

// Apply runs terraform apply for the project described by ctx.
func (r *testRunner) Apply(ctx command.ProjectContext) command.ProjectResult {
	return r.expectedApplyResult
}

func (r *testRunner) ApprovePolicies(ctx command.ProjectContext) command.ProjectResult {
	return r.expectedApprovePoliciesResult
}

func (r *testRunner) Version(ctx command.ProjectContext) command.ProjectResult {
	return r.expectedVersionResult
}

type testCMDRunner struct {
	expectedCmd *command.Comment
	t           *testing.T
	called      bool
}

func (r *testCMDRunner) Run(ctx *command.Context, cmd *command.Comment) {
	r.called = true
	assert.Equal(r.t, r.expectedCmd, cmd)
}

type TestBuilder struct {
	Type   valid.WorkflowModeType
	called bool
}

func (b *TestBuilder) BuildApplyCommands(ctx *command.Context, comment *command.Comment) ([]command.ProjectContext, error) {
	b.called = true
	return []command.ProjectContext{
		{
			WorkflowModeType: b.Type,
		},
	}, nil
}

type TestMultiBuilder struct {
	called bool
}

func (b *TestMultiBuilder) BuildApplyCommands(ctx *command.Context, comment *command.Comment) ([]command.ProjectContext, error) {
	b.called = true
	return []command.ProjectContext{
		{
			WorkflowModeType: valid.PlatformWorkflowMode,
		},
		{
			WorkflowModeType: valid.DefaultWorkflowMode,
		},
	}, nil
}

type TestCommenter struct {
	expectedComment string
	expectedPullNum int
	expectedRepo    models.Repo
	expectedCommand string
	expectedT       *testing.T

	called bool
}

func (c *TestCommenter) CreateComment(repo models.Repo, pullNum int, comment string, command string) error {

	c.called = true
	assert.Equal(c.expectedT, c.expectedComment, comment)
	assert.Equal(c.expectedT, c.expectedPullNum, pullNum)
	assert.Equal(c.expectedT, c.expectedRepo, repo)
	assert.Equal(c.expectedT, c.expectedCommand, command)

	return nil
}

func TestPlatformModeRunner_allocatesButNotPlatformMode(t *testing.T) {

	ctx := &command.Context{
		RequestCtx: context.Background(),
		HeadRepo: models.Repo{
			FullName: "owner/repo",
		},
		Pull: models.PullRequest{
			Num: 1,
			BaseRepo: models.Repo{
				FullName: "owner/base",
			},
		},
	}
	cmd := &command.Comment{
		Workspace: "hi",
	}

	commenter := &TestCommenter{
		expectedT:       t,
		expectedComment: "Platform mode does not support legacy apply commands. Please merge your PR to apply the changes. ",
		expectedPullNum: 1,
		expectedRepo:    ctx.Pull.BaseRepo,
	}

	builder := &TestBuilder{
		Type: valid.DefaultWorkflowMode,
	}
	runner := &testCMDRunner{
		t:           t,
		expectedCmd: cmd,
	}

	subject := &lyftCommand.PlatformModeRunner{
		Allocator: &testAllocator{
			expectedFeatureName: feature.PlatformMode,
			expectedT:           t,
			expectedCtx:         feature.FeatureContext{RepoName: "owner/repo"},
			expectedResult:      true,
		},
		Logger:  logging.NewNoopCtxLogger(t),
		Builder: builder,
		TemplateLoader: template.Loader[lyftCommand.LegacyApplyCommentInput]{
			GlobalCfg: valid.GlobalCfg{},
		},
		VCSClient: commenter,
		Runner:    runner,
	}

	subject.Run(ctx, cmd)

	assert.True(t, runner.called)
	assert.True(t, builder.called)
	assert.False(t, commenter.called)
}

func TestPlatformModeRunner_allocatesButPartialPlatformMode(t *testing.T) {

	ctx := &command.Context{
		RequestCtx: context.Background(),
		HeadRepo: models.Repo{
			FullName: "owner/repo",
		},
		Pull: models.PullRequest{
			Num: 1,
			BaseRepo: models.Repo{
				FullName: "owner/base",
			},
		},
	}
	cmd := &command.Comment{
		Workspace: "hi",
	}

	commenter := &TestCommenter{
		expectedT:       t,
		expectedComment: "Platform mode does not support legacy apply commands. Please merge your PR to apply the changes. ",
		expectedPullNum: 1,
		expectedRepo:    ctx.Pull.BaseRepo,
	}

	builder := &TestMultiBuilder{}
	runner := &testCMDRunner{
		t:           t,
		expectedCmd: cmd,
	}

	subject := &lyftCommand.PlatformModeRunner{
		Allocator: &testAllocator{
			expectedFeatureName: feature.PlatformMode,
			expectedT:           t,
			expectedCtx:         feature.FeatureContext{RepoName: "owner/repo"},
			expectedResult:      true,
		},
		Logger:  logging.NewNoopCtxLogger(t),
		Builder: builder,
		TemplateLoader: template.Loader[lyftCommand.LegacyApplyCommentInput]{
			GlobalCfg: valid.GlobalCfg{},
		},
		VCSClient: commenter,
		Runner:    runner,
	}

	subject.Run(ctx, cmd)

	assert.True(t, runner.called)
	assert.True(t, builder.called)
	assert.False(t, commenter.called)
}

func TestPlatformModeRunner_doesntAllocate(t *testing.T) {

	ctx := &command.Context{
		RequestCtx: context.Background(),
		HeadRepo: models.Repo{
			FullName: "owner/repo",
		},
		Pull: models.PullRequest{
			Num: 1,
			BaseRepo: models.Repo{
				FullName: "owner/base",
			},
		},
	}
	cmd := &command.Comment{
		Workspace: "hi",
	}

	commenter := &TestCommenter{
		expectedT:       t,
		expectedComment: "Platform mode does not support legacy apply commands. Please merge your PR to apply the changes. ",
		expectedPullNum: 1,
		expectedRepo:    ctx.Pull.BaseRepo,
	}

	builder := &TestBuilder{
		Type: valid.PlatformWorkflowMode,
	}
	runner := &testCMDRunner{
		t:           t,
		expectedCmd: cmd,
	}

	subject := &lyftCommand.PlatformModeRunner{
		Allocator: &testAllocator{
			expectedFeatureName: feature.PlatformMode,
			expectedT:           t,
			expectedCtx:         feature.FeatureContext{RepoName: "owner/repo"},
			expectedResult:      false,
		},
		Logger:  logging.NewNoopCtxLogger(t),
		Builder: builder,
		TemplateLoader: template.Loader[lyftCommand.LegacyApplyCommentInput]{
			GlobalCfg: valid.GlobalCfg{},
		},
		VCSClient: commenter,
		Runner:    runner,
	}

	subject.Run(ctx, cmd)

	assert.True(t, runner.called)
	assert.False(t, builder.called)
	assert.False(t, commenter.called)
}

func TestPlatformModeRunner_success(t *testing.T) {

	ctx := &command.Context{
		RequestCtx: context.Background(),
		HeadRepo: models.Repo{
			FullName: "owner/repo",
		},
		Pull: models.PullRequest{
			Num: 1,
			BaseRepo: models.Repo{
				FullName: "owner/base",
			},
		},
	}
	cmd := &command.Comment{
		Workspace: "hi",
	}

	builder := &TestBuilder{
		Type: valid.PlatformWorkflowMode,
	}
	runner := &testCMDRunner{
		t:           t,
		expectedCmd: cmd,
	}

	commenter := &TestCommenter{
		expectedT:       t,
		expectedComment: "Platform mode does not support legacy apply commands. Please merge your PR to apply the changes. ",
		expectedPullNum: 1,
		expectedRepo:    ctx.Pull.BaseRepo,
	}

	subject := &lyftCommand.PlatformModeRunner{
		Allocator: &testAllocator{
			expectedFeatureName: feature.PlatformMode,
			expectedT:           t,
			expectedCtx:         feature.FeatureContext{RepoName: "owner/repo"},
			expectedResult:      true,
		},
		Logger:  logging.NewNoopCtxLogger(t),
		Builder: builder,
		TemplateLoader: template.Loader[lyftCommand.LegacyApplyCommentInput]{
			GlobalCfg: valid.GlobalCfg{},
		},
		VCSClient: commenter,
		Runner:    runner,
	}

	subject.Run(ctx, cmd)

	assert.True(t, runner.called)
	assert.True(t, builder.called)
	assert.True(t, commenter.called)

}

func TestPlatformModeProjectRunner_plan(t *testing.T) {
	expectedResult := command.ProjectResult{
		JobID: "1234y",
	}

	cases := []struct {
		description      string
		shouldAllocate   bool
		workflowModeType valid.WorkflowModeType
		platformRunner   events.ProjectCommandRunner
		prModeRunner     events.ProjectCommandRunner
		subject          lyftCommand.PlatformModeProjectRunner
	}{
		{
			description:      "allocated and platform mode enabled",
			shouldAllocate:   true,
			workflowModeType: valid.PlatformWorkflowMode,
			platformRunner: &testRunner{
				expectedPlanResult: expectedResult,
			},
			prModeRunner: &testRunner{},
		},
		{
			description:      "allocated and platform mode not enabled",
			shouldAllocate:   true,
			workflowModeType: valid.DefaultWorkflowMode,
			platformRunner:   &testRunner{},
			prModeRunner: &testRunner{
				expectedPlanResult: expectedResult,
			},
		},
		{
			description:      "not allocated and platform mode enabled",
			shouldAllocate:   false,
			workflowModeType: valid.PlatformWorkflowMode,
			platformRunner:   &testRunner{},
			prModeRunner: &testRunner{
				expectedPlanResult: expectedResult,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			subject := lyftCommand.PlatformModeProjectRunner{
				PlatformModeRunner: c.platformRunner,
				PrModeRunner:       c.prModeRunner,
				Allocator: &testAllocator{
					expectedResult:      c.shouldAllocate,
					expectedFeatureName: feature.PlatformMode,
					expectedCtx: feature.FeatureContext{
						RepoName: "nish/repo",
					},
					expectedT: t,
				},
				Logger: logging.NewNoopCtxLogger(t),
			}

			result := subject.Plan(command.ProjectContext{
				RequestCtx: context.Background(),
				HeadRepo: models.Repo{
					FullName: "nish/repo",
				},
				WorkflowModeType: c.workflowModeType,
			})

			assert.Equal(t, expectedResult, result)
		})
	}
}

func TestPlatformModeProjectRunner_policyCheck(t *testing.T) {
	expectedResult := command.ProjectResult{
		JobID: "1234y",
	}

	cases := []struct {
		description      string
		shouldAllocate   bool
		workflowModeType valid.WorkflowModeType
		platformRunner   events.ProjectCommandRunner
		prModeRunner     events.ProjectCommandRunner
		subject          lyftCommand.PlatformModeProjectRunner
	}{
		{
			description:      "allocated and platform mode enabled",
			shouldAllocate:   true,
			workflowModeType: valid.PlatformWorkflowMode,
			platformRunner: &testRunner{
				expectedPolicyCheckResult: expectedResult,
			},
			prModeRunner: &testRunner{},
		},
		{
			description:      "allocated and platform mode not enabled",
			shouldAllocate:   true,
			workflowModeType: valid.DefaultWorkflowMode,
			platformRunner:   &testRunner{},
			prModeRunner: &testRunner{
				expectedPolicyCheckResult: expectedResult,
			},
		},
		{
			description:      "not allocated and platform mode enabled",
			shouldAllocate:   false,
			workflowModeType: valid.PlatformWorkflowMode,
			platformRunner:   &testRunner{},
			prModeRunner: &testRunner{
				expectedPolicyCheckResult: expectedResult,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			subject := lyftCommand.PlatformModeProjectRunner{
				PlatformModeRunner: c.platformRunner,
				PrModeRunner:       c.prModeRunner,
				Allocator: &testAllocator{
					expectedResult:      c.shouldAllocate,
					expectedFeatureName: feature.PlatformMode,
					expectedCtx: feature.FeatureContext{
						RepoName: "nish/repo",
					},
					expectedT: t,
				},
				Logger: logging.NewNoopCtxLogger(t),
			}

			result := subject.PolicyCheck(command.ProjectContext{
				RequestCtx: context.Background(),
				HeadRepo: models.Repo{
					FullName: "nish/repo",
				},
				WorkflowModeType: c.workflowModeType,
			})

			assert.Equal(t, expectedResult, result)
		})
	}
}

func TestPlatformModeProjectRunner_apply(t *testing.T) {
	cases := []struct {
		description      string
		shouldAllocate   bool
		workflowModeType valid.WorkflowModeType
		platformRunner   events.ProjectCommandRunner
		prModeRunner     events.ProjectCommandRunner
		subject          lyftCommand.PlatformModeProjectRunner
		expectedResult   command.ProjectResult
	}{
		{
			description:      "allocated and platform mode enabled",
			shouldAllocate:   true,
			workflowModeType: valid.PlatformWorkflowMode,
			platformRunner: &testRunner{
				expectedApplyResult: command.ProjectResult{
					RepoRelDir:   "reldir",
					Workspace:    "default",
					ProjectName:  "project",
					StatusID:     "id",
					Command:      command.Apply,
					ApplySuccess: "atlantis apply is disabled for this project. Please track the deployment when the PR is merged. ",
				},
			},
			expectedResult: command.ProjectResult{
				RepoRelDir:   "reldir",
				Workspace:    "default",
				ProjectName:  "project",
				StatusID:     "id",
				Command:      command.Apply,
				ApplySuccess: "atlantis apply is disabled for this project. Please track the deployment when the PR is merged. ",
			},
			prModeRunner: &testRunner{},
		},
		{
			description:      "allocated and platform mode not enabled",
			shouldAllocate:   true,
			workflowModeType: valid.DefaultWorkflowMode,
			platformRunner:   &testRunner{},
			prModeRunner: &testRunner{
				expectedApplyResult: command.ProjectResult{
					JobID: "1234y",
				},
			},
			expectedResult: command.ProjectResult{
				JobID: "1234y",
			},
		},
		{
			description:      "not allocated and platform mode enabled",
			shouldAllocate:   false,
			workflowModeType: valid.PlatformWorkflowMode,
			platformRunner:   &testRunner{},
			prModeRunner: &testRunner{
				expectedApplyResult: command.ProjectResult{
					JobID: "1234y",
				},
			},
			expectedResult: command.ProjectResult{
				JobID: "1234y",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			subject := lyftCommand.PlatformModeProjectRunner{
				PlatformModeRunner: c.platformRunner,
				PrModeRunner:       c.prModeRunner,
				Allocator: &testAllocator{
					expectedResult:      c.shouldAllocate,
					expectedFeatureName: feature.PlatformMode,
					expectedCtx: feature.FeatureContext{
						RepoName: "nish/repo",
					},
					expectedT: t,
				},
				Logger: logging.NewNoopCtxLogger(t),
			}

			result := subject.Apply(command.ProjectContext{
				RequestCtx: context.Background(),
				HeadRepo: models.Repo{
					FullName: "nish/repo",
				},
				RepoRelDir:       "reldir",
				Workspace:        "default",
				ProjectName:      "project",
				StatusID:         "id",
				WorkflowModeType: c.workflowModeType,
			})

			assert.Equal(t, c.expectedResult, result)
		})
	}
}

func TestPlatformModeProjectRunner_approvePolicies(t *testing.T) {
	expectedResult := command.ProjectResult{
		JobID: "1234y",
	}

	cases := []struct {
		description      string
		shouldAllocate   bool
		workflowModeType valid.WorkflowModeType
		platformRunner   events.ProjectCommandRunner
		prModeRunner     events.ProjectCommandRunner
		subject          lyftCommand.PlatformModeProjectRunner
	}{
		{
			description:      "allocated and platform mode enabled",
			shouldAllocate:   true,
			workflowModeType: valid.PlatformWorkflowMode,
			platformRunner: &testRunner{
				expectedApprovePoliciesResult: expectedResult,
			},
			prModeRunner: &testRunner{},
		},
		{
			description:      "allocated and platform mode not enabled",
			shouldAllocate:   true,
			workflowModeType: valid.DefaultWorkflowMode,
			platformRunner:   &testRunner{},
			prModeRunner: &testRunner{
				expectedApprovePoliciesResult: expectedResult,
			},
		},
		{
			description:      "not allocated and platform mode enabled",
			shouldAllocate:   false,
			workflowModeType: valid.PlatformWorkflowMode,
			platformRunner:   &testRunner{},
			prModeRunner: &testRunner{
				expectedApprovePoliciesResult: expectedResult,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			subject := lyftCommand.PlatformModeProjectRunner{
				PlatformModeRunner: c.platformRunner,
				PrModeRunner:       c.prModeRunner,
				Allocator: &testAllocator{
					expectedResult:      c.shouldAllocate,
					expectedFeatureName: feature.PlatformMode,
					expectedCtx: feature.FeatureContext{
						RepoName: "nish/repo",
					},
					expectedT: t,
				},
				Logger: logging.NewNoopCtxLogger(t),
			}

			result := subject.ApprovePolicies(command.ProjectContext{
				RequestCtx: context.Background(),
				HeadRepo: models.Repo{
					FullName: "nish/repo",
				},
				WorkflowModeType: c.workflowModeType,
			})

			assert.Equal(t, expectedResult, result)
		})
	}
}

func TestPlatformModeProjectRunner_version(t *testing.T) {
	expectedResult := command.ProjectResult{
		JobID: "1234y",
	}

	cases := []struct {
		description      string
		shouldAllocate   bool
		workflowModeType valid.WorkflowModeType
		platformRunner   events.ProjectCommandRunner
		prModeRunner     events.ProjectCommandRunner
		subject          lyftCommand.PlatformModeProjectRunner
	}{
		{
			description:      "allocated and platform mode enabled",
			shouldAllocate:   true,
			workflowModeType: valid.PlatformWorkflowMode,
			platformRunner: &testRunner{
				expectedVersionResult: expectedResult,
			},
			prModeRunner: &testRunner{},
		},
		{
			description:      "allocated and platform mode not enabled",
			shouldAllocate:   true,
			workflowModeType: valid.DefaultWorkflowMode,
			platformRunner:   &testRunner{},
			prModeRunner: &testRunner{
				expectedVersionResult: expectedResult,
			},
		},
		{
			description:      "not allocated and platform mode enabled",
			shouldAllocate:   false,
			workflowModeType: valid.PlatformWorkflowMode,
			platformRunner:   &testRunner{},
			prModeRunner: &testRunner{
				expectedVersionResult: expectedResult,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			subject := lyftCommand.PlatformModeProjectRunner{
				PlatformModeRunner: c.platformRunner,
				PrModeRunner:       c.prModeRunner,
				Allocator: &testAllocator{
					expectedResult:      c.shouldAllocate,
					expectedFeatureName: feature.PlatformMode,
					expectedCtx: feature.FeatureContext{
						RepoName: "nish/repo",
					},
					expectedT: t,
				},
				Logger: logging.NewNoopCtxLogger(t),
			}

			result := subject.Version(command.ProjectContext{
				RequestCtx: context.Background(),
				HeadRepo: models.Repo{
					FullName: "nish/repo",
				},
				WorkflowModeType: c.workflowModeType,
			})

			assert.Equal(t, expectedResult, result)
		})
	}
}
