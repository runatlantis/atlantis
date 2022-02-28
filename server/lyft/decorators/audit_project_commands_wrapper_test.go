package decorators_test

import (
	"encoding/json"
	"errors"
	"testing"

	. "github.com/runatlantis/atlantis/testing"

	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/mocks/matchers"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	snsMocks "github.com/runatlantis/atlantis/server/lyft/aws/sns/mocks"
	"github.com/runatlantis/atlantis/server/lyft/decorators"
	"github.com/runatlantis/atlantis/server/metrics"
)

func TestAuditProjectCommandsWrapper(t *testing.T) {
	RegisterMockTestingT(t)

	cases := []struct {
		Description string
		Success     bool
		Failure     bool
		Error       bool
	}{
		{
			Description: "apply success",
			Success:     true,
		},
		{
			Description: "apply error",
			Error:       true,
		},
		{
			Description: "apply failure",
			Failure:     true,
		},
	}

	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			snsMock := snsMocks.NewMockWriter()
			projectCmdRunnerMock := mocks.NewMockProjectCommandRunner()
			auditPrjCmds := &decorators.AuditProjectCommandWrapper{
				SnsWriter:            snsMock,
				ProjectCommandRunner: projectCmdRunnerMock,
			}

			prjRslt := command.ProjectResult{}

			if c.Error {
				prjRslt.Error = errors.New("oh-no")
			}

			if c.Failure {
				prjRslt.Failure = "oh-no"
			}

			logger := logging.NewNoopLogger(t)

			scope, _, _ := metrics.NewLoggingScope(logger, "atlantis")

			ctx := command.ProjectContext{
				Scope:       scope,
				Log:         logger,
				Steps:       []valid.Step{},
				ProjectName: "test-project",
				User: models.User{
					Username: "test-user",
				},
				Workspace: "default",
				PullReqStatus: models.PullReqStatus{
					ApprovalStatus: models.ApprovalStatus{
						IsApproved: true,
					},
				},
				RepoRelDir: ".",
				Tags: map[string]string{
					"environment":  "production",
					"service_name": "test-service",
				},
			}

			When(snsMock.Write(matchers.AnySliceOfByte())).ThenReturn(nil)
			When(projectCmdRunnerMock.Apply(matchers.AnyModelsProjectCommandContext())).ThenReturn(prjRslt)

			auditPrjCmds.Apply(ctx)

			eventBefore := &decorators.AtlantisJobEvent{}
			eventAfter := &decorators.AtlantisJobEvent{}
			eventPayload := snsMock.VerifyWasCalled(Twice()).Write(matchers.AnySliceOfByte()).GetAllCapturedArguments()

			err := json.Unmarshal(eventPayload[0], eventBefore)
			Ok(t, err)
			err = json.Unmarshal(eventPayload[1], eventAfter)
			Ok(t, err)

			Equals(t, eventBefore.State, decorators.AtlantisJobStateRunning)
			Equals(t, eventBefore.RootName, "test-project")
			Equals(t, eventBefore.Environment, "production")
			Equals(t, eventBefore.InitiatingUser, "test-user")
			Equals(t, eventBefore.Project, "test-service")
			Assert(t, eventBefore.EndTime == "", "end time must be empty")
			Assert(t, eventBefore.StartTime != "", "start time must be set")

			if c.Success {
				Equals(t, eventAfter.State, decorators.AtlantisJobStateSuccess)
			} else {
				Equals(t, eventAfter.State, decorators.AtlantisJobStateFailure)
			}

			Assert(t, eventBefore.StartTime == eventAfter.StartTime, "start time should not change")
			Assert(t, eventAfter.EndTime != "", "end time must be set")
			Assert(t, eventBefore.ID == eventAfter.ID, "id should not change")
		})
	}
}
