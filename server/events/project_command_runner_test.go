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
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/core/locking"
	smocks "github.com/runatlantis/atlantis/server/core/runtime/mocks"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	commandMocks "github.com/runatlantis/atlantis/server/events/command/mocks"
	"github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/mocks/matchers"
	"github.com/runatlantis/atlantis/server/events/models"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/wrappers"
	. "github.com/runatlantis/atlantis/testing"
	"github.com/stretchr/testify/assert"
)

// Test that it runs the expected plan steps.
func TestDefaultProjectCommandRunner_Plan(t *testing.T) {
	RegisterMockTestingT(t)
	mockWorkingDir := mocks.NewMockWorkingDir()
	mockLocker := mocks.NewMockProjectLocker()
	mockStepsRunner := smocks.NewMockStepsRunner()
	applyRequirementHandler := &events.AggregateApplyRequirements{
		WorkingDir: workingDir,
	}

	runner := events.NewProjectCommandRunner(
		mockStepsRunner,
		mockWorkingDir,
		nil,
		events.NewDefaultWorkingDirLocker(),
		applyRequirementHandler,
	)

	wrappedRunner := wrappers.
		WrapProjectRunner(runner).
		WithSync(mockLocker, mockURLGenerator{})

	When(mockLocker.TryLock(
		matchers.AnyContextContext(),
		matchers.AnyLoggingLogger(),
		matchers.AnyModelsPullRequest(),
		matchers.AnyModelsUser(),
		AnyString(),
		matchers.AnyModelsProject(),
	)).ThenReturn(&events.TryLockResponse{
		LockAcquired: true,
		LockKey:      "lock-key",
	}, nil)
	repoDir, cleanup := TempDir(t)
	defer cleanup()
	When(mockWorkingDir.Clone(
		matchers.AnyLoggingLogger(),
		matchers.AnyModelsRepo(),
		matchers.AnyModelsPullRequest(),
		AnyString(),
	)).ThenReturn(repoDir, false, nil)

	ctx := context.Background()
	prjCtx := command.ProjectContext{
		Log: logging.NewNoopCtxLogger(t),
		Steps: []valid.Step{
			{
				StepName:    "env",
				EnvVarName:  "name",
				EnvVarValue: "value",
			},
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
		Workspace:  "default",
		RepoRelDir: ".",
		RequestCtx: ctx,
	}

	When(mockStepsRunner.Run(ctx, prjCtx, repoDir)).ThenReturn("run\napply\nplan\ninit", nil)
	firstRes := wrappedRunner.Plan(prjCtx)

	Assert(t, firstRes.PlanSuccess != nil, "exp plan success")
	Equals(t, "https://lock-key", firstRes.PlanSuccess.LockURL)
	Equals(t, "run\napply\nplan\ninit", firstRes.PlanSuccess.TerraformOutput)
	mockStepsRunner.VerifyWasCalledOnce().Run(ctx, prjCtx, repoDir)
}

// Test that it runs the expected plan steps.
func TestDefaultProjectCommandRunner_PlanWithSync(t *testing.T) {
	RegisterMockTestingT(t)
	prjCtx := command.ProjectContext{
		RequestCtx: context.TODO(),
		Log:        logging.NewNoopCtxLogger(t),
		Pull: models.PullRequest{
			BaseRepo: models.Repo{
				FullName: "test",
			},
			Num: 1,
		},
		Workspace:  "default",
		RepoRelDir: ".",
	}

	cases := []struct {
		description   string
		usePrjLock    bool
		expFailure    string
		expPlanStatus models.ProjectPlanStatus
	}{
		{
			description:   "plan with locking",
			usePrjLock:    true,
			expFailure:    "This project is currently locked by an unapplied plan from pull . To continue, delete the lock from  or apply that plan and merge the pull request.\n\nOnce the lock is released, comment `atlantis plan` here to re-plan.",
			expPlanStatus: models.ErroredPlanStatus,
		},
		{
			description:   "plan without locking",
			usePrjLock:    false,
			expFailure:    "",
			expPlanStatus: models.PlannedPlanStatus,
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			mockWorkingDir := mocks.NewMockWorkingDir()
			mockVcsClient := vcsmocks.NewMockClient()

			dataDir, cleanup := TempDir(t)
			defer cleanup()

			boltdb, err := db.New(dataDir)
			Ok(t, err)

			lockingClient := locking.NewClient(boltdb)
			projectLocker := &events.DefaultProjectLocker{
				Locker:    lockingClient,
				VCSClient: mockVcsClient,
			}

			applyRequirementHandler := &events.AggregateApplyRequirements{
				WorkingDir: mockWorkingDir,
			}

			runner := events.NewProjectCommandRunner(
				smocks.NewMockStepsRunner(),
				mockWorkingDir,
				nil,
				events.NewDefaultWorkingDirLocker(),
				applyRequirementHandler,
			)

			targetCtx := command.ProjectContext{
				Log: logging.NewNoopCtxLogger(t),
				Pull: models.PullRequest{
					BaseRepo: models.Repo{
						FullName: "test",
					},
					Num: 2,
				},
				Workspace:  "default",
				RepoRelDir: ".",
			}

			wrappedRunner := wrappers.WrapProjectRunner(runner)
			if c.usePrjLock {
				wrappedRunner = wrappedRunner.WithSync(projectLocker, &mockURLGenerator{})
			}

			firstRes := wrappedRunner.Plan(prjCtx)
			targetRes := wrappedRunner.Plan(targetCtx)

			Assert(t, firstRes.IsSuccessful(), "exp first prjCtx to succeed")
			Equals(t, targetRes.PlanStatus(), c.expPlanStatus)
			Equals(t, targetRes.Failure, c.expFailure)
		})
	}
}

type strictTestCommitStatusUpdater struct {
	statusUpdaters []*testCommitStatusUpdater
	count          int
}

// UpdateProject(ctx context.Context, projectCtx ProjectContext, cmdName fmt.Stringer, status models.CommitStatus, url string, statusID string) (string, error)
func (t *strictTestCommitStatusUpdater) UpdateProject(ctx context.Context, projectCtx command.ProjectContext, cmdName fmt.Stringer, status models.VCSStatus, url string, statusID string) (string, error) {
	if t.count > (len(t.statusUpdaters) - 1) {
		return "", errors.New("more calls than expected")
	}

	statusID, err := t.statusUpdaters[t.count].UpdateProject(ctx, projectCtx, cmdName, status, url, statusID)
	t.count++
	return statusID, err
}

type testCommitStatusUpdater struct {
	t           *testing.T
	expCtx      context.Context
	expPrjCtx   command.ProjectContext
	expCmdName  fmt.Stringer
	expStatus   models.VCSStatus
	expURL      string
	expStatusID string

	statusID string
	err      error
}

func (t *testCommitStatusUpdater) UpdateProject(ctx context.Context, projectCtx command.ProjectContext, cmdName fmt.Stringer, status models.VCSStatus, url string, statusID string) (string, error) {
	assert.Equal(t.t, t.expCtx, ctx)
	assert.Equal(t.t, t.expPrjCtx, projectCtx)
	assert.Equal(t.t, t.expCmdName, cmdName)
	assert.Equal(t.t, t.expStatus, status)
	assert.Equal(t.t, t.expURL, url)
	assert.Equal(t.t, t.expStatusID, statusID)

	return t.statusID, t.err
}

type testProjectCommandRunner struct {
	t         *testing.T
	expPrjCtx command.ProjectContext
	result    command.ProjectResult
}

func (t *testProjectCommandRunner) Apply(ctx command.ProjectContext) command.ProjectResult {
	assert.Equal(t.t, t.expPrjCtx, ctx)

	return t.result
}

func (t *testProjectCommandRunner) Plan(ctx command.ProjectContext) command.ProjectResult {
	assert.Equal(t.t, t.expPrjCtx, ctx)

	return t.result
}

func (t *testProjectCommandRunner) PolicyCheck(ctx command.ProjectContext) command.ProjectResult {
	assert.Equal(t.t, t.expPrjCtx, ctx)

	return t.result
}

func (t *testProjectCommandRunner) ApprovePolicies(ctx command.ProjectContext) command.ProjectResult {
	assert.Equal(t.t, t.expPrjCtx, ctx)

	return t.result
}

func (t *testProjectCommandRunner) Version(ctx command.ProjectContext) command.ProjectResult {
	assert.Equal(t.t, t.expPrjCtx, ctx)

	return t.result
}

func TestProjectOutputWrapper(t *testing.T) {
	RegisterMockTestingT(t)
	prjCtx := command.ProjectContext{
		Log: logging.NewNoopCtxLogger(t),
		Steps: []valid.Step{
			{
				StepName: "plan",
			},
		},
		Workspace:  "default",
		RepoRelDir: ".",
		RequestCtx: context.TODO(),
	}

	cases := []struct {
		Description string
		Failure     bool
		Error       bool
		Success     bool
		CommandName command.Name
	}{
		{
			Description: "plan success",
			Success:     true,
			CommandName: command.Plan,
		},
		{
			Description: "plan failure",
			Failure:     true,
			CommandName: command.Plan,
		},
		{
			Description: "plan error",
			Error:       true,
			CommandName: command.Plan,
		},
		{
			Description: "apply success",
			Success:     true,
			CommandName: command.Apply,
		},
		{
			Description: "apply failure",
			Failure:     true,
			CommandName: command.Apply,
		},
		{
			Description: "apply error",
			Error:       true,
			CommandName: command.Apply,
		},
	}

	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			var prjResult command.ProjectResult
			var expCommitStatus models.VCSStatus

			mockJobURLGenerator := commandMocks.NewMockJobURLGenerator()
			mockJobCloser := commandMocks.NewMockJobCloser()

			if c.Success {
				prjResult = command.ProjectResult{
					PlanSuccess:  &models.PlanSuccess{},
					ApplySuccess: "exists",
				}
				expCommitStatus = models.SuccessVCSStatus
			} else if c.Failure {
				prjResult = command.ProjectResult{
					Failure: "failure",
				}
				expCommitStatus = models.FailedVCSStatus
			} else if c.Error {
				prjResult = command.ProjectResult{
					Error: errors.New("error"),
				}
				expCommitStatus = models.FailedVCSStatus
			}

			prjCtx.CommandName = c.CommandName

			mockProjectCommandRunner := testProjectCommandRunner{
				t:         t,
				expPrjCtx: prjCtx,
				result:    prjResult,
			}

			mockCommitStatusUpdater := strictTestCommitStatusUpdater{
				statusUpdaters: []*testCommitStatusUpdater{
					{
						t:           t,
						expCtx:      context.TODO(),
						expPrjCtx:   prjCtx,
						expCmdName:  c.CommandName,
						expStatus:   models.PendingVCSStatus,
						expStatusID: "",
						expURL:      "",
						statusID:    "",
						err:         nil,
					},
					{
						t:           t,
						expCtx:      context.TODO(),
						expPrjCtx:   prjCtx,
						expCmdName:  c.CommandName,
						expStatus:   expCommitStatus,
						expStatusID: "",
						expURL:      "",
						statusID:    "",
						err:         nil,
					},
				},
			}

			projectUpdater := command.ProjectStatusUpdater{
				JobCloser:               mockJobCloser,
				ProjectJobURLGenerator:  mockJobURLGenerator,
				ProjectVCSStatusUpdater: &mockCommitStatusUpdater,
			}

			runner := &events.ProjectOutputWrapper{
				ProjectStatusUpdater: projectUpdater,
				ProjectCommandRunner: &mockProjectCommandRunner,
			}

			switch c.CommandName {
			case command.Plan:
				runner.Plan(prjCtx)
			case command.Apply:
				runner.Apply(prjCtx)
			}
		})
	}
}

// Test what happens if there's no working dir. This signals that the project
// was never planned.
func TestDefaultProjectCommandRunner_ApplyNotCloned(t *testing.T) {
	mockWorkingDir := mocks.NewMockWorkingDir()
	runner := &events.DefaultProjectCommandRunner{
		WorkingDir: mockWorkingDir,
	}
	prjCtx := command.ProjectContext{}
	When(mockWorkingDir.GetWorkingDir(prjCtx.BaseRepo, prjCtx.Pull, prjCtx.Workspace)).ThenReturn("", os.ErrNotExist)

	firstRes := runner.Apply(prjCtx)
	ErrEquals(t, "project has not been cloned–did you run plan?", firstRes.Error)
}

// Test that if approval is required and the PR isn't approved we give an error.
func TestDefaultProjectCommandRunner_ApplyNotApproved(t *testing.T) {
	RegisterMockTestingT(t)
	mockWorkingDir := mocks.NewMockWorkingDir()
	mockSender := mocks.NewMockWebhooksSender()
	runner := &events.DefaultProjectCommandRunner{
		WorkingDir:       mockWorkingDir,
		WorkingDirLocker: events.NewDefaultWorkingDirLocker(),
		AggregateApplyRequirements: &events.AggregateApplyRequirements{
			WorkingDir: mockWorkingDir,
		},
		Webhooks: mockSender,
	}
	prjCtx := command.ProjectContext{
		ApplyRequirements: []string{"approved"},
		PullReqStatus: models.PullReqStatus{
			ApprovalStatus: models.ApprovalStatus{
				IsApproved: false,
			},
		},
	}
	tmp, cleanup := TempDir(t)
	defer cleanup()
	When(mockWorkingDir.GetWorkingDir(prjCtx.BaseRepo, prjCtx.Pull, prjCtx.Workspace)).ThenReturn(tmp, nil)

	firstRes := runner.Apply(prjCtx)
	Equals(t, "Pull request must be approved by at least one person other than the author before running apply.", firstRes.Failure)
}

func TestDefaultProjectCommandRunner_ForceOverridesApplyReqs_IfPlatformMode(t *testing.T) {
	RegisterMockTestingT(t)
	mockWorkingDir := mocks.NewMockWorkingDir()
	mockSender := mocks.NewMockWebhooksSender()
	runner := &events.DefaultProjectCommandRunner{
		WorkingDir:       mockWorkingDir,
		StepsRunner:      smocks.NewMockStepsRunner(),
		WorkingDirLocker: events.NewDefaultWorkingDirLocker(),
		AggregateApplyRequirements: &events.AggregateApplyRequirements{
			WorkingDir: mockWorkingDir,
		},
		Webhooks: mockSender,
	}
	prjCtx := command.ProjectContext{
		PullReqStatus: models.PullReqStatus{
			ApprovalStatus: models.ApprovalStatus{
				IsApproved: false,
			},
		},
		ApplyRequirements: []string{"approved"},
		WorkflowModeType:  valid.PlatformWorkflowMode,
	}
	tmp, cleanup := TempDir(t)
	defer cleanup()
	When(mockWorkingDir.GetWorkingDir(prjCtx.BaseRepo, prjCtx.Pull, prjCtx.Workspace)).ThenReturn(tmp, nil)

	firstRes := runner.Apply(prjCtx)
	Equals(t, "", firstRes.Failure)
}

func TestDefaultProjectCommandRunner_ForceOverridesApplyReqs(t *testing.T) {
	RegisterMockTestingT(t)
	mockWorkingDir := mocks.NewMockWorkingDir()
	mockSender := mocks.NewMockWebhooksSender()
	runner := &events.DefaultProjectCommandRunner{
		WorkingDir:       mockWorkingDir,
		StepsRunner:      smocks.NewMockStepsRunner(),
		WorkingDirLocker: events.NewDefaultWorkingDirLocker(),
		AggregateApplyRequirements: &events.AggregateApplyRequirements{
			WorkingDir: mockWorkingDir,
		},
		Webhooks: mockSender,
	}
	prjCtx := command.ProjectContext{
		PullReqStatus: models.PullReqStatus{
			ApprovalStatus: models.ApprovalStatus{
				IsApproved: false,
			},
		},
		ApplyRequirements: []string{"approved"},
		ForceApply:        true,
	}
	tmp, cleanup := TempDir(t)
	defer cleanup()
	When(mockWorkingDir.GetWorkingDir(prjCtx.BaseRepo, prjCtx.Pull, prjCtx.Workspace)).ThenReturn(tmp, nil)

	firstRes := runner.Apply(prjCtx)
	Equals(t, "", firstRes.Failure)
}

// Test that if mergeable is required and the PR isn't mergeable we give an error.
func TestDefaultProjectCommandRunner_ApplyNotMergeable(t *testing.T) {
	RegisterMockTestingT(t)
	mockWorkingDir := mocks.NewMockWorkingDir()
	runner := &events.DefaultProjectCommandRunner{
		WorkingDir:       mockWorkingDir,
		WorkingDirLocker: events.NewDefaultWorkingDirLocker(),
		StepsRunner:      smocks.NewMockStepsRunner(),
		AggregateApplyRequirements: &events.AggregateApplyRequirements{
			WorkingDir: mockWorkingDir,
		},
	}
	prjCtx := command.ProjectContext{
		PullReqStatus: models.PullReqStatus{
			Mergeable: false,
		},
		ApplyRequirements: []string{"mergeable"},
	}
	tmp, cleanup := TempDir(t)
	defer cleanup()
	When(mockWorkingDir.GetWorkingDir(prjCtx.BaseRepo, prjCtx.Pull, prjCtx.Workspace)).ThenReturn(tmp, nil)

	firstRes := runner.Apply(prjCtx)
	Equals(t, "Pull request must be mergeable before running apply.", firstRes.Failure)
}

// Test that if undiverged is required and the PR is diverged we give an error.
func TestDefaultProjectCommandRunner_ApplyDiverged(t *testing.T) {
	RegisterMockTestingT(t)
	mockWorkingDir := mocks.NewMockWorkingDir()
	runner := &events.DefaultProjectCommandRunner{
		WorkingDir:       mockWorkingDir,
		WorkingDirLocker: events.NewDefaultWorkingDirLocker(),
		StepsRunner:      smocks.NewMockStepsRunner(),
		AggregateApplyRequirements: &events.AggregateApplyRequirements{
			WorkingDir: mockWorkingDir,
		},
	}
	prjCtx := command.ProjectContext{
		ApplyRequirements: []string{"undiverged"},
	}
	tmp, cleanup := TempDir(t)
	defer cleanup()
	When(mockWorkingDir.GetWorkingDir(prjCtx.BaseRepo, prjCtx.Pull, prjCtx.Workspace)).ThenReturn(tmp, nil)
	When(mockWorkingDir.HasDiverged(matchers.AnyLoggingLogger(), AnyString(), matchers.AnyModelsRepo())).ThenReturn(true)

	firstRes := runner.Apply(prjCtx)
	Equals(t, "Default branch must be rebased onto pull request before running apply.", firstRes.Failure)
}

// Test that it runs the expected apply steps.
func TestDefaultProjectCommandRunner_Apply(t *testing.T) {
	cases := []struct {
		description string
		steps       []valid.Step
		applyReqs   []string

		expSteps      []string
		expOut        string
		expFailure    string
		pullMergeable bool
	}{
		{
			description: "normal workflow",
			steps:       valid.DefaultApplyStage.Steps,
			expSteps:    []string{"apply"},
			expOut:      "apply",
		},
		{
			description: "approval required",
			steps:       valid.DefaultApplyStage.Steps,
			applyReqs:   []string{"approved"},
			expSteps:    []string{"approve", "apply"},
			expOut:      "apply",
		},
		{
			description:   "mergeable required",
			steps:         valid.DefaultApplyStage.Steps,
			pullMergeable: true,
			applyReqs:     []string{"mergeable"},
			expSteps:      []string{"apply"},
			expOut:        "apply",
		},
		{
			description:   "mergeable required, pull not mergeable",
			steps:         valid.DefaultApplyStage.Steps,
			pullMergeable: false,
			applyReqs:     []string{"mergeable"},
			expSteps:      []string{""},
			expOut:        "",
			expFailure:    "Pull request must be mergeable before running apply.",
		},
		{
			description:   "mergeable and approved required",
			steps:         valid.DefaultApplyStage.Steps,
			pullMergeable: true,
			applyReqs:     []string{"mergeable", "approved"},
			expSteps:      []string{"approved", "apply"},
			expOut:        "apply",
		},
		{
			description: "workflow with custom apply stage",
			steps: []valid.Step{
				{
					StepName:    "env",
					EnvVarName:  "key",
					EnvVarValue: "value",
				},
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
			expSteps: []string{"env", "run", "apply", "plan", "init"},
			expOut:   "run\napply\nplan\ninit",
		},
	}

	for _, c := range cases {
		if c.description != "workflow with custom apply stage" {
			continue
		}
		t.Run(c.description, func(t *testing.T) {
			RegisterMockTestingT(t)
			mockStepsRunner := smocks.NewMockStepsRunner()
			mockWorkingDir := mocks.NewMockWorkingDir()
			mockSender := mocks.NewMockWebhooksSender()
			applyReqHandler := &events.AggregateApplyRequirements{
				WorkingDir: mockWorkingDir,
			}

			runner := events.DefaultProjectCommandRunner{
				StepsRunner:                mockStepsRunner,
				WorkingDir:                 mockWorkingDir,
				Webhooks:                   mockSender,
				WorkingDirLocker:           events.NewDefaultWorkingDirLocker(),
				AggregateApplyRequirements: applyReqHandler,
			}
			repoDir, cleanup := TempDir(t)
			defer cleanup()
			When(mockWorkingDir.GetWorkingDir(
				matchers.AnyModelsRepo(),
				matchers.AnyModelsPullRequest(),
				AnyString(),
			)).ThenReturn(repoDir, nil)

			ctx := context.Background()
			prjCtx := command.ProjectContext{
				Log:               logging.NewNoopCtxLogger(t),
				Steps:             c.steps,
				Workspace:         "default",
				ApplyRequirements: c.applyReqs,
				RepoRelDir:        ".",
				PullReqStatus: models.PullReqStatus{
					ApprovalStatus: models.ApprovalStatus{
						IsApproved: true,
					},
					Mergeable: true,
				},
				RequestCtx: ctx,
			}

			When(mockStepsRunner.Run(ctx, prjCtx, repoDir)).ThenReturn("run\napply\nplan\ninit", nil)

			firstRes := runner.Apply(prjCtx)
			Equals(t, c.expOut, firstRes.ApplySuccess)
			Equals(t, c.expFailure, firstRes.Failure)

			mockStepsRunner.VerifyWasCalledOnce().Run(ctx, prjCtx, repoDir)
		})
	}
}

// Test that it runs the expected apply steps.
func TestDefaultProjectCommandRunner_ApplyRunStepFailure(t *testing.T) {
	RegisterMockTestingT(t)
	mockStepsRunner := smocks.NewMockStepsRunner()
	mockWorkingDir := mocks.NewMockWorkingDir()
	mockSender := mocks.NewMockWebhooksSender()
	applyReqHandler := &events.AggregateApplyRequirements{
		WorkingDir: mockWorkingDir,
	}

	runner := events.DefaultProjectCommandRunner{
		StepsRunner:                mockStepsRunner,
		WorkingDir:                 mockWorkingDir,
		WorkingDirLocker:           events.NewDefaultWorkingDirLocker(),
		AggregateApplyRequirements: applyReqHandler,
		Webhooks:                   mockSender,
	}
	repoDir, cleanup := TempDir(t)
	defer cleanup()
	When(mockWorkingDir.GetWorkingDir(
		matchers.AnyModelsRepo(),
		matchers.AnyModelsPullRequest(),
		AnyString(),
	)).ThenReturn(repoDir, nil)

	ctx := context.Background()
	prjCtx := command.ProjectContext{
		Log: logging.NewNoopCtxLogger(t),
		Steps: []valid.Step{
			{
				StepName: "apply",
			},
		},
		Workspace:         "default",
		ApplyRequirements: []string{},
		RepoRelDir:        ".",
		PullReqStatus: models.PullReqStatus{
			Mergeable: true,
		},
		RequestCtx: ctx,
	}
	When(mockStepsRunner.Run(ctx, prjCtx, ".")).ThenReturn("apply", fmt.Errorf("something went wrong"))

	firstRes := runner.Apply(prjCtx)
	Assert(t, firstRes.ApplySuccess == "", "exp apply failure")

	mockStepsRunner.VerifyWasCalledOnce().Run(ctx, prjCtx, repoDir)
}

type mockURLGenerator struct{}

func (m mockURLGenerator) GenerateLockURL(lockID string) string {
	return "https://" + lockID
}
