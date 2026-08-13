// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"context"
	"errors"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/core/ownership"
	redisdb "github.com/runatlantis/atlantis/server/core/redis"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/stretchr/testify/require"
)

func TestReplicaRouting_DistributesPRsAndPreservesWholePRAffinity(t *testing.T) {
	h := newHAHarness(t)

	require.NoError(t, h.dispatcherA.DispatchComment(haCommentDispatch(t, 101, "dev", command.Plan)))
	require.NoError(t, h.dispatcherB.DispatchComment(haCommentDispatch(t, 202, "dev", command.Plan)))
	require.NoError(t, h.dispatcherB.DispatchComment(haCommentDispatch(t, 101, "prod", command.Apply)))
	require.NoError(t, h.dispatcherA.DispatchComment(haCommentDispatch(t, 202, "prod", command.Apply)))

	require.Equal(t, []haRunnerCall{
		{PullNum: 101, Workspace: "dev", Command: command.Plan},
		{PullNum: 101, Workspace: "prod", Command: command.Apply, PlanPresent: true},
	}, h.runnerA.snapshot())
	require.Equal(t, []haRunnerCall{
		{PullNum: 202, Workspace: "dev", Command: command.Plan},
		{PullNum: 202, Workspace: "prod", Command: command.Apply, PlanPresent: true},
	}, h.runnerB.snapshot())
	require.Equal(t, 1, h.diskA.deleteCount(101))
	require.Equal(t, 1, h.diskB.deleteCount(202))
}

func TestReplicaRouting_LeaseReplacementDuringResetReroutesBeforeExecution(t *testing.T) {
	h := newHAHarness(t)
	const pullNum = 250
	resetStarted, continueReset := h.diskA.blockNextDelete()
	dispatch := haCommentDispatch(t, pullNum, "default", command.Plan)
	dispatchErr := make(chan error, 1)
	go func() {
		dispatchErr <- h.dispatcherA.DispatchComment(dispatch)
	}()

	select {
	case <-resetStarted:
	case <-time.After(time.Second):
		t.Fatal("local reset did not start")
	}
	h.redis.FastForward(31 * time.Second)
	replacement, err := h.storeB.Claim(context.Background(), haOwnershipKey(pullNum))
	require.NoError(t, err)
	require.Equal(t, "replica-b", replacement.ReplicaID)
	close(continueReset)
	select {
	case err := <-dispatchErr:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("stale dispatch did not reroute")
	}

	require.Empty(t, h.runnerA.snapshot(), "expired owner must not start the command after reset")
	require.Equal(t, []haRunnerCall{{PullNum: pullNum, Workspace: "default", Command: command.Plan}}, h.runnerB.snapshot())
}

func TestReplicaRouting_OwnerLossRequiresReplanAndKeepsSamePRLockUsable(t *testing.T) {
	h := newHAHarness(t)
	const pullNum = 303

	require.NoError(t, h.dispatcherA.DispatchComment(haCommentDispatch(t, pullNum, "default", command.Plan)))
	ownerA, found, err := h.storeA.Current(context.Background(), haOwnershipKey(pullNum))
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, "replica-a", ownerA.ReplicaID)

	project := models.Project{RepoFullName: "owner/repo", Path: "infra", ProjectName: "default"}
	pull := models.PullRequest{Num: pullNum, BaseRepo: models.Repo{FullName: "owner/repo"}}
	projectLocker := &events.DefaultProjectLocker{
		Locker:         locking.NewClient(h.database),
		NoOpLocker:     locking.NewNoOpLocker(),
		ExecutableName: "atlantis",
	}
	locked, err := projectLocker.TryLock(logging.NewNoopLogger(t), pull, models.User{Username: "alice"}, "default", project, true)
	require.NoError(t, err)
	require.True(t, locked.LockAcquired)

	// Simulate stale files on a reused local volume before replica B takes over.
	h.diskB.setPlan(pullNum, true)
	h.redis.FastForward(31 * time.Second)
	require.NoError(t, h.dispatcherB.DispatchComment(haCommentDispatch(t, pullNum, "default", command.Apply)))

	ownerB, found, err := h.storeB.Current(context.Background(), haOwnershipKey(pullNum))
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, "replica-b", ownerB.ReplicaID)
	require.NotEqual(t, ownerA.ClaimID, ownerB.ClaimID)
	require.Equal(t, 1, h.diskB.deleteCount(pullNum))
	require.False(t, h.runnerB.lastApplyHadPlan(), "apply after takeover must not use a stale local plan")

	// The shared lock belongs to the same PR, so the new owner can safely re-plan.
	locked, err = projectLocker.TryLock(logging.NewNoopLogger(t), pull, models.User{Username: "alice"}, "default", project, true)
	require.NoError(t, err)
	require.True(t, locked.LockAcquired)

	require.NoError(t, h.dispatcherB.DispatchComment(haCommentDispatch(t, pullNum, "default", command.Plan)))
	require.NoError(t, h.dispatcherA.DispatchComment(haCommentDispatch(t, pullNum, "default", command.Apply)))
	require.True(t, h.runnerB.lastApplyHadPlan())
	require.Equal(t, 1, h.diskB.deleteCount(pullNum), "one ownership claim must reset local state only once")
}

type haHarness struct {
	redis       *miniredis.Miniredis
	database    *redisdb.RedisDB
	storeA      *redisdb.OwnerStore
	storeB      *redisdb.OwnerStore
	dispatcherA *events.RoutedCommandDispatcher
	dispatcherB *events.RoutedCommandDispatcher
	diskA       *haPlanDisk
	diskB       *haPlanDisk
	runnerA     *haCommandRunner
	runnerB     *haCommandRunner
}

func newHAHarness(t *testing.T) *haHarness {
	t.Helper()
	redisServer := miniredis.RunT(t)
	host, portString, err := net.SplitHostPort(redisServer.Addr())
	require.NoError(t, err)
	port, err := strconv.Atoi(portString)
	require.NoError(t, err)
	database, err := redisdb.NewWithConfig(redisdb.Config{Hostname: host, Port: port})
	require.NoError(t, err)
	logger := logging.NewNoopLogger(t)
	storeA, err := redisdb.NewOwnerStore(database, redisdb.OwnerStoreConfig{
		ReplicaID: "replica-a", AdvertiseURL: "http://replica-a", TTL: 30 * time.Second,
	}, logger)
	require.NoError(t, err)
	storeB, err := redisdb.NewOwnerStore(database, redisdb.OwnerStoreConfig{
		ReplicaID: "replica-b", AdvertiseURL: "http://replica-b", TTL: 30 * time.Second,
	}, logger)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, storeA.Close())
		require.NoError(t, storeB.Close())
		require.NoError(t, database.Close())
	})

	diskA, diskB := newHAPlanDisk(), newHAPlanDisk()
	runnerA, runnerB := &haCommandRunner{disk: diskA}, &haCommandRunner{disk: diskB}
	executorA := &events.LocalCommandExecutor{
		Hydrator: &haRepoHydrator{}, Runner: runnerA, PullCleaner: haPullCleaner{}, WorkingDir: diskA,
		ClaimGuard: events.NewLocalClaimGuard(), Owners: storeA, Logger: logger, TestingMode: true,
	}
	executorB := &events.LocalCommandExecutor{
		Hydrator: &haRepoHydrator{}, Runner: runnerB, PullCleaner: haPullCleaner{}, WorkingDir: diskB,
		ClaimGuard: events.NewLocalClaimGuard(), Owners: storeB, Logger: logger, TestingMode: true,
	}
	forwarder := &haDirectForwarder{targets: map[string]haForwardTarget{
		"http://replica-a": {replicaID: "replica-a", store: storeA, executor: executorA},
		"http://replica-b": {replicaID: "replica-b", store: storeB, executor: executorB},
	}}

	return &haHarness{
		redis: redisServer, database: database, storeA: storeA, storeB: storeB,
		dispatcherA: events.NewRoutedCommandDispatcher("replica-a", storeA, executorA, forwarder),
		dispatcherB: events.NewRoutedCommandDispatcher("replica-b", storeB, executorB, forwarder),
		diskA:       diskA, diskB: diskB, runnerA: runnerA, runnerB: runnerB,
	}
}

func haCommentDispatch(t *testing.T, pullNum int, workspace string, name command.Name) events.CommentDispatch {
	t.Helper()
	repo, err := models.NewRepo(models.Github, "owner/repo", "https://github.com/owner/repo", "receiver", "receiver-token", "")
	require.NoError(t, err)
	pull := models.PullRequest{Num: pullNum, BaseRepo: repo}
	dispatch, err := events.NewCommentDispatch(
		repo, nil, &pull, models.User{Username: "alice"}, pullNum,
		&events.CommentCommand{Name: name, Workspace: workspace},
	)
	require.NoError(t, err)
	return dispatch
}

func haOwnershipKey(pullNum int) ownership.Key {
	return ownership.Key{VCSHostname: "github.com", RepoFullName: "owner/repo", PullNum: pullNum}
}

type haRepoHydrator struct{}

func (*haRepoHydrator) HydrateRepo(ref events.RepoRef) (models.Repo, error) {
	return models.NewRepo(ref.VCSHost.Type, ref.FullName, ref.CloneURL, "owner-local", "owner-token", "")
}

type haPlanDisk struct {
	mu             sync.Mutex
	plans          map[int]bool
	deletes        map[int]int
	deleteStarted  chan struct{}
	continueDelete chan struct{}
}

func newHAPlanDisk() *haPlanDisk {
	return &haPlanDisk{plans: make(map[int]bool), deletes: make(map[int]int)}
}

func (d *haPlanDisk) Delete(_ logging.SimpleLogging, _ models.Repo, pull models.PullRequest) error {
	d.mu.Lock()
	d.plans[pull.Num] = false
	d.deletes[pull.Num]++
	deleteStarted := d.deleteStarted
	continueDelete := d.continueDelete
	d.deleteStarted = nil
	d.continueDelete = nil
	d.mu.Unlock()
	if deleteStarted != nil {
		close(deleteStarted)
		<-continueDelete
	}
	return nil
}

func (d *haPlanDisk) blockNextDelete() (<-chan struct{}, chan<- struct{}) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.deleteStarted = make(chan struct{})
	d.continueDelete = make(chan struct{})
	return d.deleteStarted, d.continueDelete
}

func (d *haPlanDisk) setPlan(pullNum int, present bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.plans[pullNum] = present
}

func (d *haPlanDisk) hasPlan(pullNum int) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.plans[pullNum]
}

func (d *haPlanDisk) deleteCount(pullNum int) int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.deletes[pullNum]
}

type haRunnerCall struct {
	PullNum     int
	Workspace   string
	Command     command.Name
	PlanPresent bool
}

type haCommandRunner struct {
	mu    sync.Mutex
	disk  *haPlanDisk
	calls []haRunnerCall
}

func (r *haCommandRunner) RunCommentCommand(_ models.Repo, _ *models.Repo, pull *models.PullRequest, _ models.User, pullNum int, cmd *events.CommentCommand) {
	call := haRunnerCall{PullNum: pullNum, Workspace: cmd.Workspace, Command: cmd.Name}
	if cmd.Name == command.Apply {
		call.PlanPresent = r.disk.hasPlan(pullNum)
	}
	if cmd.Name == command.Plan {
		r.disk.setPlan(pullNum, true)
	}
	r.mu.Lock()
	r.calls = append(r.calls, call)
	r.mu.Unlock()
}

func (r *haCommandRunner) RunAutoplanCommand(models.Repo, models.Repo, models.PullRequest, models.User) {
}

func (r *haCommandRunner) RunRoutedCommentCommand(base models.Repo, head *models.Repo, pull *models.PullRequest, user models.User, pullNum int, cmd *events.CommentCommand, _ command.RoutingContext) {
	r.RunCommentCommand(base, head, pull, user, pullNum, cmd)
}

func (r *haCommandRunner) RunRoutedAutoplanCommand(base models.Repo, head models.Repo, pull models.PullRequest, user models.User, _ command.RoutingContext) {
	r.RunAutoplanCommand(base, head, pull, user)
}

func (r *haCommandRunner) snapshot() []haRunnerCall {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]haRunnerCall(nil), r.calls...)
}

func (r *haCommandRunner) lastApplyHadPlan() bool {
	calls := r.snapshot()
	for i := len(calls) - 1; i >= 0; i-- {
		if calls[i].Command == command.Apply {
			return calls[i].PlanPresent
		}
	}
	return false
}

type haPullCleaner struct{}

func (haPullCleaner) CleanUpPull(logging.SimpleLogging, models.Repo, models.PullRequest) error {
	return nil
}

type haForwardTarget struct {
	replicaID string
	store     ownership.Store
	executor  events.CommandExecutor
}

type haDirectForwarder struct {
	targets map[string]haForwardTarget
}

func (f *haDirectForwarder) ForwardComment(owner ownership.Record, dispatch events.CommentDispatch) error {
	return f.forward(owner, dispatch.OwnershipKey(), func(target haForwardTarget) error {
		return target.executor.ExecuteComment(dispatch, owner.ClaimID)
	})
}

func (f *haDirectForwarder) ForwardAutoplan(owner ownership.Record, dispatch events.AutoplanDispatch) error {
	return f.forward(owner, dispatch.OwnershipKey(), func(target haForwardTarget) error {
		return target.executor.ExecuteAutoplan(dispatch, owner.ClaimID)
	})
}

func (f *haDirectForwarder) ForwardPullClosed(owner ownership.Record, dispatch events.PullClosedDispatch) error {
	return f.forward(owner, dispatch.OwnershipKey(), func(target haForwardTarget) error {
		return target.executor.ExecutePullClosed(dispatch, owner.ClaimID)
	})
}

func (f *haDirectForwarder) forward(owner ownership.Record, key ownership.Key, execute func(haForwardTarget) error) error {
	target, ok := f.targets[owner.AdvertiseURL]
	if !ok {
		return errors.New("unknown owner URL")
	}
	current, found, err := target.store.Current(context.Background(), key)
	if err != nil {
		return err
	}
	if !found || current.ReplicaID != target.replicaID || current.ClaimID != owner.ClaimID || !target.store.Owns(key, owner.ClaimID) {
		return events.ErrOwnershipChanged
	}
	return execute(target)
}
