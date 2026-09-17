// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/boltdb"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestUnlockCommandRunner_OptionalPublication(t *testing.T) {
	for _, mode := range []string{"nil", "coordinated", "nil error", "coordinated error", "nil busy", "nil ambiguous"} {
		t.Run(mode, func(t *testing.T) {
			RegisterMockTestingT(t)
			storage, err := boltdb.New(t.TempDir())
			Ok(t, err)
			t.Cleanup(func() { Ok(t, storage.Close()) })
			pull := models.PullRequest{Num: 42, BaseRepo: models.Repo{FullName: "owner/repo"}}
			if mode == "nil busy" || mode == "nil ambiguous" {
				_, err = storage.AcquirePublicationLease(context.Background(), pull, "other-owner", time.Minute)
				Ok(t, err)
				if mode == "nil ambiguous" {
					Ok(t, storage.BeginPublication(context.Background(), pull, command.PublicationFence{Owner: "other-owner"}))
				}
			}
			deleter := mocks.NewMockDeleteLockCommand()
			client := vcsmocks.NewMockClient()
			runner := events.NewUnlockCommandRunner(deleter, client, false, "")
			Assert(t, runner.Publication == nil, "constructor must remain usable before server injection")
			coordinated := strings.HasPrefix(mode, "coordinated")
			if coordinated {
				runner.Publication = events.NewPublicationCoordinator(storage, context.Background())
			}
			called := false
			When(deleter.DeleteLocksByPull(Any[logging.SimpleLogging](), Eq(pull), Any[command.PublicationWriteMode]())).Then(func(args []Param) ReturnValues {
				called = true
				writeMode := args[2].(command.PublicationWriteMode)
				if coordinated {
					fence, ok := writeMode.(command.PublicationFence)
					Assert(t, ok, "injected coordinator must pass a fence")
					lease, err := storage.GetPublicationLease(context.Background(), pull)
					Ok(t, err)
					Assert(t, lease != nil, "fence must be live during deletion")
					Equals(t, lease.Owner, fence.Owner)
				} else {
					Equals(t, command.PublicationWriteMode(command.NoClaim{}), writeMode)
				}
				if strings.HasSuffix(mode, "error") {
					return ReturnValues{0, errors.New("discard unavailable")}
				}
				// Use the real backend to prove an optional coordinator cannot bypass
				// another owner's live or ambiguous publication record.
				if _, err := storage.UpdatePullWithResults(pull, []command.ProjectResult{{Command: command.Plan, Workspace: "default", RepoRelDir: ".", ProjectCommandOutput: command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}}}}, writeMode); err != nil {
					return ReturnValues{0, err}
				}
				return ReturnValues{1, nil}
			})
			ctx := &command.Context{Log: logging.NewNoopLogger(t), Pull: pull}
			runner.Run(ctx, &events.CommentCommand{Name: command.Unlock})
			Assert(t, called, "unlock must attempt the original operation")
			success := "All Atlantis locks for this PR have been unlocked and plans discarded"
			if mode == "nil" || mode == "coordinated" {
				Assert(t, !ctx.CommandHasErrors, "successful unlock must remain successful")
				client.VerifyWasCalledOnce().CreateComment(Any[logging.SimpleLogging](), Eq(pull.BaseRepo), Eq(pull.Num), Eq(success), Eq("unlock"))
			} else {
				Assert(t, ctx.CommandHasErrors, "discard/fencing errors must be visible")
				client.VerifyWasCalled(Never()).CreateComment(Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Eq(success), Any[string]())
				client.VerifyWasCalledOnce().CreateComment(Any[logging.SimpleLogging](), Eq(pull.BaseRepo), Eq(pull.Num), Any[string](), Eq("unlock"))
			}
		})
	}
}

// Exercise the real bulk cleanup helper: a mock which writes a project result
// cannot expose the missing-status path that used to skip publication fencing.
func TestUnlockCommandRunner_RealCleanupWithoutStatus(t *testing.T) {
	for _, mode := range []string{"nil", "coordinated", "nil busy", "nil ambiguous"} {
		t.Run(mode, func(t *testing.T) {
			RegisterMockTestingT(t)
			storage, err := boltdb.New(t.TempDir())
			Ok(t, err)
			t.Cleanup(func() { Ok(t, storage.Close()) })
			pull := models.PullRequest{Num: 42, BaseRepo: models.Repo{FullName: "owner/repo"}}
			locker := locking.NewClient(storage)
			lock, err := locker.TryLock(models.NewProject(pull.BaseRepo.FullName, "path", ""), "default", pull, models.User{})
			Ok(t, err)
			if mode == "nil busy" || mode == "nil ambiguous" {
				_, err = storage.AcquirePublicationLease(context.Background(), pull, "other-owner", time.Minute)
				Ok(t, err)
				if mode == "nil ambiguous" {
					Ok(t, storage.BeginPublication(context.Background(), pull, command.PublicationFence{Owner: "other-owner"}))
				}
			}
			workingDir := &publicationCleanupWorkingDir{}
			deleter := &events.DefaultDeleteLockCommand{Database: storage, Locker: locker, WorkingDir: workingDir}
			client := vcsmocks.NewMockClient()
			runner := events.NewUnlockCommandRunner(deleter, client, false, "")
			if mode == "coordinated" {
				runner.Publication = events.NewPublicationCoordinator(storage, context.Background())
			}
			ctx := &command.Context{Log: logging.NewNoopLogger(t), Pull: pull}
			runner.Run(ctx, &events.CommentCommand{Name: command.Unlock})
			remaining, err := locker.GetLock(lock.LockKey)
			Ok(t, err)
			success := "All Atlantis locks for this PR have been unlocked and plans discarded"
			if mode == "nil" || mode == "coordinated" {
				Assert(t, !ctx.CommandHasErrors, "usable constructor/injected coordinator must permit cleanup")
				Assert(t, remaining == nil, "successful cleanup must remove the lock")
				Equals(t, 1, workingDir.deleted)
				client.VerifyWasCalledOnce().CreateComment(Any[logging.SimpleLogging](), Eq(pull.BaseRepo), Eq(pull.Num), Eq(success), Eq("unlock"))
			} else {
				Assert(t, ctx.CommandHasErrors, "publication conflict must remain visible")
				Assert(t, remaining != nil, "conflicting command must retain the lock")
				Equals(t, 0, workingDir.deleted)
				client.VerifyWasCalled(Never()).CreateComment(Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Eq(success), Any[string]())
				client.VerifyWasCalledOnce().CreateComment(Any[logging.SimpleLogging](), Eq(pull.BaseRepo), Eq(pull.Num), Eq("Failed to delete PR locks"), Eq("unlock"))
			}
		})
	}
}
