// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.

package events_test

import (
	"os"
	"strings"
	"testing"

	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/mocks/matchers"
	"github.com/runatlantis/atlantis/server/events/models"
	mocks2 "github.com/runatlantis/atlantis/server/events/runtime/mocks"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestDefaultProjectCommandRunner_Plan(t *testing.T) {
	cases := []struct {
		description string
		projCfg     *valid.Project
		globalCfg   *valid.Config
		expSteps    []string
		expOut      string
	}{
		{
			description: "use defaults",
			projCfg:     nil,
			globalCfg:   nil,
			expSteps:    []string{"init", "plan"},
			expOut:      "init\nplan",
		},
		{
			description: "no workflow, use defaults",
			projCfg: &valid.Project{
				Dir: ".",
			},
			globalCfg: &valid.Config{
				Version: 2,
				Projects: []valid.Project{
					{
						Dir: ".",
					},
				},
			},
			expSteps: []string{"init", "plan"},
			expOut:   "init\nplan",
		},
		{
			description: "workflow without plan stage set",
			projCfg: &valid.Project{
				Dir:      ".",
				Workflow: String("myworkflow"),
			},
			globalCfg: &valid.Config{
				Version: 2,
				Projects: []valid.Project{
					{
						Dir: ".",
					},
				},
				Workflows: map[string]valid.Workflow{
					"myworkflow": {
						Apply: &valid.Stage{
							Steps: nil,
						},
					},
				},
			},
			expSteps: []string{"init", "plan"},
			expOut:   "init\nplan",
		},
		{
			description: "workflow with custom plan stage",
			projCfg: &valid.Project{
				Dir:      ".",
				Workflow: String("myworkflow"),
			},
			globalCfg: &valid.Config{
				Version: 2,
				Projects: []valid.Project{
					{
						Dir: ".",
					},
				},
				Workflows: map[string]valid.Workflow{
					"myworkflow": {
						Plan: &valid.Stage{
							Steps: []valid.Step{
								{
									StepName: "run",
								},
								{
									StepName: "apply",
								},
								{
									StepName: "plan",
								},
								{
									StepName: "init",
								},
							},
						},
					},
				},
			},
			expSteps: []string{"run", "apply", "plan", "init"},
			expOut:   "run\napply\nplan\ninit",
		},
	}

	for _, c := range cases {
		t.Run(strings.Join(c.expSteps, ","), func(t *testing.T) {
			RegisterMockTestingT(t)
			mockInit := mocks.NewMockStepRunner()
			mockPlan := mocks.NewMockStepRunner()
			mockApply := mocks.NewMockStepRunner()
			mockRun := mocks.NewMockStepRunner()
			mockWorkingDir := mocks.NewMockWorkingDir()
			mockLocker := mocks.NewMockProjectLocker()

			runner := events.DefaultProjectCommandRunner{
				Locker:              mockLocker,
				LockURLGenerator:    mockURLGenerator{},
				InitStepRunner:      mockInit,
				PlanStepRunner:      mockPlan,
				ApplyStepRunner:     mockApply,
				RunStepRunner:       mockRun,
				PullApprovedChecker: nil,
				WorkingDir:          mockWorkingDir,
				Webhooks:            nil,
				WorkingDirLocker:    events.NewDefaultWorkingDirLocker(),
			}

			repoDir := "/tmp/mydir"
			When(mockWorkingDir.Clone(
				matchers.AnyPtrToLoggingSimpleLogger(),
				matchers.AnyModelsRepo(),
				matchers.AnyModelsRepo(),
				matchers.AnyModelsPullRequest(),
				AnyBool(),
				AnyString(),
			)).ThenReturn(repoDir, nil)
			When(mockLocker.TryLock(
				matchers.AnyPtrToLoggingSimpleLogger(),
				matchers.AnyModelsPullRequest(),
				matchers.AnyModelsUser(),
				AnyString(),
				matchers.AnyModelsProject(),
			)).ThenReturn(&events.TryLockResponse{
				LockAcquired: true,
				LockKey:      "lock-key",
			}, nil)

			ctx := models.ProjectCommandContext{
				Log:           logging.NewNoopLogger(),
				ProjectConfig: c.projCfg,
				Workspace:     "default",
				GlobalConfig:  c.globalCfg,
				RepoRelDir:    ".",
			}
			When(mockInit.Run(ctx, nil, repoDir)).ThenReturn("init", nil)
			When(mockPlan.Run(ctx, nil, repoDir)).ThenReturn("plan", nil)
			When(mockApply.Run(ctx, nil, repoDir)).ThenReturn("apply", nil)
			When(mockRun.Run(ctx, nil, repoDir)).ThenReturn("run", nil)

			res := runner.Plan(ctx)

			Assert(t, res.PlanSuccess != nil, "exp plan success")
			Equals(t, "https://lock-key", res.PlanSuccess.LockURL)
			Equals(t, c.expOut, res.PlanSuccess.TerraformOutput)

			for _, step := range c.expSteps {
				switch step {
				case "init":
					mockInit.VerifyWasCalledOnce().Run(ctx, nil, repoDir)
				case "plan":
					mockPlan.VerifyWasCalledOnce().Run(ctx, nil, repoDir)
				case "apply":
					mockApply.VerifyWasCalledOnce().Run(ctx, nil, repoDir)
				case "run":
					mockRun.VerifyWasCalledOnce().Run(ctx, nil, repoDir)
				}
			}
		})
	}
}

func TestDefaultProjectCommandRunner_ApplyNotCloned(t *testing.T) {
	mockWorkingDir := mocks.NewMockWorkingDir()
	runner := &events.DefaultProjectCommandRunner{
		WorkingDir: mockWorkingDir,
	}
	ctx := models.ProjectCommandContext{}
	When(mockWorkingDir.GetWorkingDir(ctx.BaseRepo, ctx.Pull, ctx.Workspace)).ThenReturn("", os.ErrNotExist)

	res := runner.Apply(ctx)
	ErrEquals(t, "project has not been cloned–did you run plan?", res.Error)
}

func TestDefaultProjectCommandRunner_ApplyNotApproved(t *testing.T) {
	RegisterMockTestingT(t)
	mockWorkingDir := mocks.NewMockWorkingDir()
	mockApproved := mocks2.NewMockPullApprovedChecker()
	runner := &events.DefaultProjectCommandRunner{
		WorkingDir:              mockWorkingDir,
		PullApprovedChecker:     mockApproved,
		WorkingDirLocker:        events.NewDefaultWorkingDirLocker(),
		RequireApprovalOverride: true,
	}
	ctx := models.ProjectCommandContext{}
	When(mockWorkingDir.GetWorkingDir(ctx.BaseRepo, ctx.Pull, ctx.Workspace)).ThenReturn("/tmp/mydir", nil)
	When(mockApproved.PullIsApproved(ctx.BaseRepo, ctx.Pull)).ThenReturn(false, nil)

	res := runner.Apply(ctx)
	Equals(t, "Pull request must be approved before running apply.", res.Failure)
}

func TestDefaultProjectCommandRunner_Apply(t *testing.T) {
	cases := []struct {
		description string
		projCfg     *valid.Project
		globalCfg   *valid.Config
		expSteps    []string
		expOut      string
	}{
		{
			description: "use defaults",
			projCfg:     nil,
			globalCfg:   nil,
			expSteps:    []string{"apply"},
			expOut:      "apply",
		},
		{
			description: "no workflow, use defaults",
			projCfg: &valid.Project{
				Dir: ".",
			},
			globalCfg: &valid.Config{
				Version: 2,
				Projects: []valid.Project{
					{
						Dir: ".",
					},
				},
			},
			expSteps: []string{"apply"},
			expOut:   "apply",
		},
		{
			description: "no workflow, approval required, use defaults",
			projCfg: &valid.Project{
				Dir:               ".",
				ApplyRequirements: []string{"approved"},
			},
			globalCfg: &valid.Config{
				Version: 2,
				Projects: []valid.Project{
					{
						Dir:               ".",
						ApplyRequirements: []string{"approved"},
					},
				},
			},
			expSteps: []string{"approved", "apply"},
			expOut:   "apply",
		},
		{
			description: "workflow without apply stage set",
			projCfg: &valid.Project{
				Dir:      ".",
				Workflow: String("myworkflow"),
			},
			globalCfg: &valid.Config{
				Version: 2,
				Projects: []valid.Project{
					{
						Dir: ".",
					},
				},
				Workflows: map[string]valid.Workflow{
					"myworkflow": {
						Plan: &valid.Stage{
							Steps: nil,
						},
					},
				},
			},
			expSteps: []string{"apply"},
			expOut:   "apply",
		},
		{
			description: "workflow with custom apply stage",
			projCfg: &valid.Project{
				Dir:      ".",
				Workflow: String("myworkflow"),
			},
			globalCfg: &valid.Config{
				Version: 2,
				Projects: []valid.Project{
					{
						Dir: ".",
					},
				},
				Workflows: map[string]valid.Workflow{
					"myworkflow": {
						Apply: &valid.Stage{
							Steps: []valid.Step{
								{
									StepName: "run",
								},
								{
									StepName: "apply",
								},
								{
									StepName: "plan",
								},
								{
									StepName: "init",
								},
							},
						},
					},
				},
			},
			expSteps: []string{"run", "apply", "plan", "init"},
			expOut:   "run\napply\nplan\ninit",
		},
	}

	for _, c := range cases {
		t.Run(strings.Join(c.expSteps, ","), func(t *testing.T) {
			RegisterMockTestingT(t)
			mockInit := mocks.NewMockStepRunner()
			mockPlan := mocks.NewMockStepRunner()
			mockApply := mocks.NewMockStepRunner()
			mockRun := mocks.NewMockStepRunner()
			mockApproved := mocks2.NewMockPullApprovedChecker()
			mockWorkingDir := mocks.NewMockWorkingDir()
			mockLocker := mocks.NewMockProjectLocker()
			mockSender := mocks.NewMockWebhooksSender()

			runner := events.DefaultProjectCommandRunner{
				Locker:              mockLocker,
				LockURLGenerator:    mockURLGenerator{},
				InitStepRunner:      mockInit,
				PlanStepRunner:      mockPlan,
				ApplyStepRunner:     mockApply,
				RunStepRunner:       mockRun,
				PullApprovedChecker: mockApproved,
				WorkingDir:          mockWorkingDir,
				Webhooks:            mockSender,
				WorkingDirLocker:    events.NewDefaultWorkingDirLocker(),
			}

			repoDir := "/tmp/mydir"
			When(mockWorkingDir.GetWorkingDir(
				matchers.AnyModelsRepo(),
				matchers.AnyModelsPullRequest(),
				AnyString(),
			)).ThenReturn(repoDir, nil)

			ctx := models.ProjectCommandContext{
				Log:           logging.NewNoopLogger(),
				ProjectConfig: c.projCfg,
				Workspace:     "default",
				GlobalConfig:  c.globalCfg,
				RepoRelDir:    ".",
			}
			When(mockInit.Run(ctx, nil, repoDir)).ThenReturn("init", nil)
			When(mockPlan.Run(ctx, nil, repoDir)).ThenReturn("plan", nil)
			When(mockApply.Run(ctx, nil, repoDir)).ThenReturn("apply", nil)
			When(mockRun.Run(ctx, nil, repoDir)).ThenReturn("run", nil)
			When(mockApproved.PullIsApproved(ctx.BaseRepo, ctx.Pull)).ThenReturn(true, nil)

			res := runner.Apply(ctx)
			Equals(t, c.expOut, res.ApplySuccess)

			for _, step := range c.expSteps {
				switch step {
				case "approved":
					mockApproved.VerifyWasCalledOnce().PullIsApproved(ctx.BaseRepo, ctx.Pull)
				case "init":
					mockInit.VerifyWasCalledOnce().Run(ctx, nil, repoDir)
				case "plan":
					mockPlan.VerifyWasCalledOnce().Run(ctx, nil, repoDir)
				case "apply":
					mockApply.VerifyWasCalledOnce().Run(ctx, nil, repoDir)
				case "run":
					mockRun.VerifyWasCalledOnce().Run(ctx, nil, repoDir)
				}
			}
		})
	}
}

type mockURLGenerator struct{}

func (m mockURLGenerator) GenerateLockURL(lockID string) string {
	return "https://" + lockID
}
