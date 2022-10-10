package command_test

import (
	"context"
	"testing"

	"gopkg.in/go-playground/assert.v1"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	lyftCommand "github.com/runatlantis/atlantis/server/lyft/command"
	"github.com/runatlantis/atlantis/server/lyft/feature"
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
