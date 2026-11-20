// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"strings"
	"sync"
	"testing"
	"time"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/models/testdata"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics/metricstest"
	. "github.com/runatlantis/atlantis/testing"
)

// Reproduces #7050: pushing a commit while a plan is running for an older
// commit of the same pull request must not fail the new head's autoplan.
func TestIssue7050_PushDuringPlanDoesNotFailNewHead(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()
	setup(t, func(tc *TestConfig) { tc.workingDirLocker = locker })
	setAutoplanLiveHead("new")
	newHead := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num, HeadCommit: "new"}

	// The plan for the previous commit is still running.
	unlockOldPlan, err := locker.TryLockPull(newHead.BaseRepo.FullName, newHead.Num, command.Plan, events.WorkingDirLockMetadata{HeadCommit: "old"})
	Ok(t, err)

	When(projectCommandBuilder.BuildAutoplanCommands(Any[*command.Context]())).ThenReturn([]command.ProjectContext{}, nil)
	When(workingDir.GetPullDir(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(t.TempDir(), nil)

	done := make(chan struct{})
	go func() {
		ch.RunAutoplanCommand(testdata.GithubRepo, testdata.GithubRepo, newHead, testdata.User)
		close(done)
	}()
	// Let the old plan keep running for a moment after the push.
	select {
	case <-done:
	case <-time.After(300 * time.Millisecond):
	}
	unlockOldPlan()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for autoplan")
	}

	commitUpdater.VerifyWasCalled(Never()).UpdateCombined(
		Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Eq(models.FailedCommitStatus), Eq(command.Plan))
	projectCommandBuilder.VerifyWasCalledOnce().BuildAutoplanCommands(Any[*command.Context]())
	commitUpdater.VerifyWasCalledOnce().UpdateCombinedCount(
		Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(newHead), Eq(models.SuccessCommitStatus), Eq(command.Plan), Eq(models.ProjectCounts{}))
}

// Reproduces the report on #7050: `atlantis apply -d dir1` followed by
// `atlantis apply -d dir2` must not fail the second apply with
// "pull request N is currently locked by apply".
func TestIssue7050_TargetedAppliesToDifferentDirsRunConcurrently(t *testing.T) {
	locker := events.NewDefaultWorkingDirLocker()
	vcsClient := setup(t, func(tc *TestConfig) { tc.workingDirLocker = locker })
	pull := models.PullRequest{BaseRepo: testdata.GithubRepo, State: models.OpenPullState, Num: testdata.Pull.Num, HeadCommit: "abc123"}
	newCtx := func() *command.Context {
		return &command.Context{
			User:     testdata.User,
			Log:      logging.NewNoopLogger(t),
			Scope:    metricstest.NewLoggingScope(t, logging.NewNoopLogger(t), "atlantis"),
			Pull:     pull,
			HeadRepo: testdata.GithubRepo,
			Trigger:  command.CommentTrigger,
		}
	}
	applyDir1 := &events.CommentCommand{Name: command.Apply, RepoRelDir: "dir1"}
	applyDir2 := &events.CommentCommand{Name: command.Apply, RepoRelDir: "dir2"}

	firstStarted := make(chan struct{})
	releaseFirst := make(chan struct{})
	var blockFirst sync.Once
	When(projectCommandBuilder.BuildApplyCommands(Any[*command.Context](), Eq(applyDir1))).Then(func([]Param) ReturnValues {
		// Block the first build that runs under the pull-level apply lock,
		// so the second apply arrives while the first one is in progress.
		if locker.HasCommandLock(pull.BaseRepo.FullName, pull.Num, command.Apply) {
			blockFirst.Do(func() {
				close(firstStarted)
				<-releaseFirst
			})
		}
		return ReturnValues{[]command.ProjectContext{}, nil}
	})
	When(projectCommandBuilder.BuildApplyCommands(Any[*command.Context](), Eq(applyDir2))).ThenReturn([]command.ProjectContext{}, nil)

	firstDone := make(chan struct{})
	go func() {
		applyCommandRunner.Run(newCtx(), applyDir1)
		close(firstDone)
	}()
	select {
	case <-firstStarted:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for the first apply to start")
	}

	// The first apply is still running when the second one is requested.
	secondCtx := newCtx()
	applyCommandRunner.Run(secondCtx, applyDir2)
	close(releaseFirst)
	<-firstDone

	projectCommandBuilder.VerifyWasCalled(AtLeast(1)).BuildApplyCommands(Any[*command.Context](), Eq(applyDir2))
	_, _, _, comments, _ := vcsClient.VerifyWasCalled(AtLeast(0)).CreateComment(
		Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]()).GetAllCapturedArguments()
	for _, comment := range comments {
		Assert(t, !strings.Contains(comment, "currently locked by"), "second apply hit the pull lock: %s", comment)
	}
}

type staticLivePullHead string

func (h staticLivePullHead) GetLivePullIdentity(ctx command.ProjectContext) (models.PullRequest, error) {
	pull := ctx.Pull
	pull.HeadCommit = string(h)
	return pull, nil
}

// setAutoplanLiveHead makes the plan runner from setup() report head as the
// pull request's live head commit.
func setAutoplanLiveHead(head string) {
	planCommandRunner.LivePullHeadFetcher = staticLivePullHead(head)
}
