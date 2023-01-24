package events_test

import (
	"testing"

	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/models/testdata"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics"
	. "github.com/runatlantis/atlantis/testing"
)

func TestImportCommandRunner_Run(t *testing.T) {
	RegisterMockTestingT(t)

	tests := []struct {
		name          string
		pullReqStatus models.PullReqStatus
		projectCmds   []command.ProjectContext
		expComment    string
	}{
		{
			name: "success with zero projects",
			pullReqStatus: models.PullReqStatus{
				ApprovalStatus: models.ApprovalStatus{IsApproved: true},
				Mergeable:      true,
			},
			projectCmds: []command.ProjectContext{},
			expComment:  "Ran Import for 0 projects:",
		},
		{
			name: "failure with multiple projects",
			pullReqStatus: models.PullReqStatus{
				ApprovalStatus: models.ApprovalStatus{IsApproved: true},
				Mergeable:      true,
			},
			projectCmds: []command.ProjectContext{{}, {}},
			expComment:  "**Import Failed**: import cannot run on multiple projects. please specify one project.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vcsClient := setup(t)

			scopeNull, _, _ := metrics.NewLoggingScope(logger, "atlantis")
			modelPull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num}
			ctx := &command.Context{
				User:     testdata.User,
				Log:      logging.NewNoopLogger(t),
				Scope:    scopeNull,
				Pull:     modelPull,
				HeadRepo: testdata.GithubRepo,
				Trigger:  command.CommentTrigger,
			}
			cmd := &events.CommentCommand{Name: command.Import}

			When(pullReqStatusFetcher.FetchPullStatus(modelPull)).ThenReturn(tt.pullReqStatus, nil)
			When(projectCommandBuilder.BuildImportCommands(ctx, cmd)).ThenReturn(tt.projectCmds, nil)

			importCommandRunner.Run(ctx, cmd)

			Assert(t, ctx.PullRequestStatus.Mergeable == true, "PullRequestStatus must be set for import_requirements")
			vcsClient.VerifyWasCalledOnce().CreateComment(testdata.GithubRepo, modelPull.Num, tt.expComment, "import")
		})
	}
}
