// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"errors"
	"strings"
	"testing"

	. "github.com/petergtz/pegomock/v4"
	dbmocks "github.com/runatlantis/atlantis/server/core/db/mocks"
	lockingmocks "github.com/runatlantis/atlantis/server/core/locking/mocks"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/models/testdata"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/jobs"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
	"go.uber.org/mock/gomock"
)

func TestReconfigureCommandRunner_RunCleansPullState(t *testing.T) {
	RegisterMockTestingT(t)
	ctrl := gomock.NewController(t)
	database := dbmocks.NewMockDatabase(ctrl)
	locker := lockingmocks.NewMockLocker(ctrl)
	workingDir := mocks.NewMockWorkingDir()
	resourceCleaner := mocks.NewMockResourceCleaner()
	cancellationTracker := mocks.NewMockCancellationTracker()
	pullUpdater := reconfigurePullUpdater()
	runner := events.NewReconfigureCommandRunner(
		workingDir,
		events.NewDefaultWorkingDirLocker(),
		locker,
		database,
		pullUpdater,
		resourceCleaner,
		cancellationTracker,
	)
	pull := reconfigurePull()
	pullStatus := &models.PullStatus{
		Projects: []models.ProjectStatus{
			{ProjectName: "app", RepoRelDir: "app", Workspace: "default"},
			{ProjectName: "db", RepoRelDir: "db", Workspace: "prod"},
		},
	}
	ctx := &command.Context{
		Log:        logging.NewNoopLogger(t),
		Pull:       pull,
		PullStatus: pullStatus,
	}

	database.EXPECT().GetPullStatus(pull).Return(pullStatus, nil)
	When(workingDir.Delete(Any[logging.SimpleLogging](), Eq(pull.BaseRepo), Eq(pull))).ThenReturn(nil)
	locker.EXPECT().UnlockByPull(pull.BaseRepo.FullName, pull.Num).Return(nil, nil)
	database.EXPECT().DeletePullStatus(pull).Return(nil)

	runner.Run(ctx, &events.CommentCommand{Name: command.Reconfigure})

	workingDir.VerifyWasCalledOnce().Delete(Any[logging.SimpleLogging](), Eq(pull.BaseRepo), Eq(pull))
	resourceCleaner.VerifyWasCalled(Times(2)).CleanUp(Any[jobs.PullInfo]())
	cancellationTracker.VerifyWasCalledOnce().Clear(pull)
	Assert(t, ctx.PullStatus == nil, "expected reconfigure to clear stale pull status from command context")
	Assert(t, !ctx.CommandHasErrors, "expected reconfigure to succeed")
}

func TestReconfigureCommandRunner_RunStopsWhenPullIsLocked(t *testing.T) {
	RegisterMockTestingT(t)
	ctrl := gomock.NewController(t)
	database := dbmocks.NewMockDatabase(ctrl)
	locker := lockingmocks.NewMockLocker(ctrl)
	workingDir := mocks.NewMockWorkingDir()
	vcsClient := vcsmocks.NewMockClient()
	pullUpdater := reconfigurePullUpdaterWithClient(vcsClient)
	workingDirLocker := events.NewDefaultWorkingDirLocker()
	pull := reconfigurePull()
	_, err := workingDirLocker.TryLock(pull.BaseRepo.FullName, pull.Num, "default", ".", "app", command.Plan)
	Ok(t, err)
	runner := events.NewReconfigureCommandRunner(
		workingDir,
		workingDirLocker,
		locker,
		database,
		pullUpdater,
		nil,
		nil,
	)
	ctx := &command.Context{
		Log:  logging.NewNoopLogger(t),
		Pull: pull,
	}

	runner.Run(ctx, &events.CommentCommand{Name: command.Reconfigure})

	workingDir.VerifyWasCalled(Never()).Delete(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest]())
	_, _, _, comment, _ := vcsClient.VerifyWasCalledOnce().CreateComment(
		Any[logging.SimpleLogging](), Eq(pull.BaseRepo), Eq(pull.Num), Any[string](), Eq(command.Reconfigure.String())).GetCapturedArguments()
	Assert(t, strings.Contains(comment, "currently locked by"), "expected lock error comment, got %q", comment)
	Assert(t, ctx.CommandHasErrors, "expected reconfigure to stop on active pull lock")
}

func TestReconfigureCommandRunner_RunStopsOnWorkspaceDeleteFailure(t *testing.T) {
	RegisterMockTestingT(t)
	ctrl := gomock.NewController(t)
	database := dbmocks.NewMockDatabase(ctrl)
	locker := lockingmocks.NewMockLocker(ctrl)
	workingDir := mocks.NewMockWorkingDir()
	vcsClient := vcsmocks.NewMockClient()
	pullUpdater := reconfigurePullUpdaterWithClient(vcsClient)
	pull := reconfigurePull()
	runner := events.NewReconfigureCommandRunner(
		workingDir,
		events.NewDefaultWorkingDirLocker(),
		locker,
		database,
		pullUpdater,
		nil,
		nil,
	)
	ctx := &command.Context{
		Log:  logging.NewNoopLogger(t),
		Pull: pull,
	}

	database.EXPECT().GetPullStatus(pull).Return(nil, nil)
	When(workingDir.Delete(Any[logging.SimpleLogging](), Eq(pull.BaseRepo), Eq(pull))).ThenReturn(errors.New("boom"))

	runner.Run(ctx, &events.CommentCommand{Name: command.Reconfigure})

	_, _, _, comment, _ := vcsClient.VerifyWasCalledOnce().CreateComment(
		Any[logging.SimpleLogging](), Eq(pull.BaseRepo), Eq(pull.Num), Any[string](), Eq(command.Reconfigure.String())).GetCapturedArguments()
	Assert(t, strings.Contains(comment, "deleting pull request workspace: boom"), "expected delete error comment, got %q", comment)
	Assert(t, ctx.CommandHasErrors, "expected reconfigure to stop on delete failure")
}

func reconfigurePullUpdater() *events.PullUpdater {
	return reconfigurePullUpdaterWithClient(vcsmocks.NewMockClient())
}

func reconfigurePullUpdaterWithClient(vcsClient *vcsmocks.MockClient) *events.PullUpdater {
	return &events.PullUpdater{
		VCSClient:        vcsClient,
		MarkdownRenderer: events.NewMarkdownRenderer(false, false, false, false, false, false, "", "atlantis", false, false),
	}
}

func reconfigurePull() models.PullRequest {
	pull := testdata.Pull
	pull.BaseRepo = testdata.GithubRepo
	return pull
}
