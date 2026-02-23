# Pegomock to uber-go/mock Migration Plan

## Overview

Migrate all mock generation and test assertions from `github.com/petergtz/pegomock/v4` to `go.uber.org/mock` (uber-go/mock v0.6.0). Pegomock is unmaintained and causing issues; uber-go/mock is the actively maintained community fork of golang/mock.

**Prior attempt**: PR #5758 (abandoned) attempted a big-bang migration (+5,878/-13,831 lines). This plan takes an incremental approach — one package group at a time, ensuring tests pass at each stage.

**Migration scope**: 66 mock files, 67 go:generate directives, 53 test files, ~1,982 pegomock references.

## Key Pattern Differences

| Aspect | Pegomock | Gomock (uber-go/mock) |
|--------|----------|----------------------|
| Constructor | `NewMockX()` | `NewMockX(ctrl)` with `gomock.NewController(t)` |
| Stubbing | `When(mock.Method(Any[T]())).ThenReturn(vals)` | `mock.EXPECT().Method(gomock.Any()).Return(vals)` |
| Verify once | Post-hoc `mock.VerifyWasCalledOnce().Method()` | Pre-execution `mock.EXPECT().Method().Times(1)` |
| Verify never | `mock.VerifyWasCalled(Never()).Method()` | Don't set EXPECT (gomock auto-fails on unexpected) |
| Unstubbed calls | Returns zero values silently | **PANICS** — must add `.AnyTimes()` for incidental calls |

## Stages

### Stage 1: Leaf Packages — server/core/db and server/core/locking
**Goal**: Migrate the foundational mock interfaces (Database, Locker, ApplyLocker, ApplyLockChecker) and all their consumers.
**Success Criteria**: All `./server/...` tests pass with gomock-generated mocks for these interfaces.
**Tests**: `go test ./server/... -count=1`
**Status**: Complete

**Files changed**:
- `server/core/db/db.go` — go:generate directive updated to mockgen
- `server/core/locking/locking.go` — go:generate directive updated to mockgen
- `server/core/locking/apply_locking.go` — 2 go:generate directives updated to mockgen
- `server/core/db/mocks/mock_database.go` — regenerated with mockgen
- `server/core/locking/mocks/mock_locker.go` — regenerated with mockgen
- `server/core/locking/mocks/mock_apply_locker.go` — regenerated with mockgen
- `server/core/locking/mocks/mock_apply_lock_checker.go` — regenerated with mockgen
- `server/core/locking/locking_test.go` — fully rewritten with gomock
- `server/server_internal_test.go` — fully rewritten with gomock (DoAndReturn pattern)
- `server/server_test.go` — coexistence: gomock for locking mocks, pegomock for TemplateWriter
- `server/controllers/locks_controller_test.go` — gomock for locking mocks, pegomock for events mocks
- `server/controllers/api_controller_test.go` — gomock for Locker, pegomock for events mocks
- `server/events/delete_lock_command_test.go` — gomock for locking Locker, pegomock for WorkingDir
- `server/events/pull_closed_executor_test.go` — gomock for locking Locker, pegomock for events mocks
- `server/events/project_locker_test.go` — fully rewritten with gomock (no pegomock)
- `server/events/command_runner_test.go` — gomock for ApplyLockChecker and Locker, pegomock for events mocks
- `server/events/apply_command_runner_test.go` — gomock for ApplyLockChecker, pegomock for events mocks
- `go.mod` — added `go.uber.org/mock v0.6.0`

### Stage 2: Events Package — server/events/mocks
**Goal**: Migrate the large events mock package (ProjectCommandBuilder, ProjectCommandRunner, EventParsing, WorkingDir, etc.)
**Success Criteria**: All `./server/...` tests pass.
**Tests**: `go test ./server/... -count=1`
**Status**: Not Started

**Key interfaces**: ProjectCommandBuilder, ProjectCommandRunner, EventParsing, GithubPullGetter, GitlabMergeRequestGetter, AzureDevopsPullGetter, WorkingDir, WorkingDirLocker, PendingPlanFinder, DeleteLockCommand, CommitStatusUpdater, PreWorkflowHooksCommandRunner, PostWorkflowHooksCommandRunner, CancellationTracker, ResourceCleaner

### Stage 3: VCS Package — server/events/vcs/mocks
**Goal**: Migrate VCS client mocks (Client, PullReqStatusFetcher)
**Success Criteria**: All `./server/...` tests pass.
**Tests**: `go test ./server/... -count=1`
**Status**: Not Started

### Stage 4: Remaining Packages
**Goal**: Migrate remaining mock packages (jobs/mocks, logging/mocks, scheduled/mocks, webhooks/mocks)
**Success Criteria**: All `./server/...` tests pass.
**Tests**: `go test ./server/... -count=1`
**Status**: Not Started

### Stage 5: Remove Pegomock Dependency
**Goal**: Remove all pegomock imports, delete pegomock-generated code, remove dependency from go.mod.
**Success Criteria**: `grep -r pegomock` returns zero results. `go mod tidy` succeeds. All tests pass.
**Tests**: `go test ./... -count=1`
**Status**: Not Started

### Stage 6: Cleanup and Documentation
**Goal**: Remove this plan file. Update CONTRIBUTING.md if it references mock generation. Verify CI passes.
**Success Criteria**: PR merged, CI green, no references to pegomock in codebase.
**Status**: Not Started
