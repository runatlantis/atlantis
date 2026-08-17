// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/runatlantis/atlantis/server/core/ownership"
	commandpkg "github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

// ErrOwnershipChanged asks an ingress replica to resolve ownership one more time.
var ErrOwnershipChanged = errors.New("pull request ownership changed")

const ownershipOperationTimeout = 5 * time.Second

//go:generate go tool pegomock generate github.com/runatlantis/atlantis/server/events --package mocks -o mocks/mock_command_dispatcher.go CommandDispatcher

// CommandDispatcher routes parsed VCS work to the pull request owner.
type CommandDispatcher interface {
	DispatchComment(CommentDispatch) error
	DispatchAutoplan(AutoplanDispatch) error
	DispatchPullClosed(PullClosedDispatch) error
}

// CommandExecutor runs routed work on the owning replica.
type CommandExecutor interface {
	ExecuteComment(CommentDispatch, string) error
	ExecuteAutoplan(AutoplanDispatch, string) error
	ExecutePullClosed(PullClosedDispatch, string) error
}

// CommandForwarder sends routed work to another replica.
type CommandForwarder interface {
	ForwardComment(ownership.Record, CommentDispatch) error
	ForwardAutoplan(ownership.Record, AutoplanDispatch) error
	ForwardPullClosed(ownership.Record, PullClosedDispatch) error
}

// RepoHydrator adds owner-local VCS credentials to a routed repository reference.
type RepoHydrator interface {
	HydrateRepo(RepoRef) (models.Repo, error)
}

// PullStateCleaner removes replica-local state for a pull request.
type PullStateCleaner interface {
	Delete(logging.SimpleLogging, models.Repo, models.PullRequest) error
}

// RoutedCommandDispatcher resolves a renewable owner before dispatching work.
type RoutedCommandDispatcher struct {
	replicaID string
	owners    ownership.Store
	local     CommandExecutor
	forwarder CommandForwarder
}

// NewRoutedCommandDispatcher creates an ownership-aware command dispatcher.
func NewRoutedCommandDispatcher(replicaID string, owners ownership.Store, local CommandExecutor, forwarder CommandForwarder) *RoutedCommandDispatcher {
	return &RoutedCommandDispatcher{
		replicaID: replicaID,
		owners:    owners,
		local:     local,
		forwarder: forwarder,
	}
}

func (d *RoutedCommandDispatcher) DispatchComment(command CommentDispatch) error {
	return d.route(
		command.OwnershipKey(),
		func(claimID string) error { return d.local.ExecuteComment(command, claimID) },
		func(owner ownership.Record) error { return d.forwarder.ForwardComment(owner, command) },
	)
}

func (d *RoutedCommandDispatcher) DispatchAutoplan(command AutoplanDispatch) error {
	return d.route(
		command.OwnershipKey(),
		func(claimID string) error { return d.local.ExecuteAutoplan(command, claimID) },
		func(owner ownership.Record) error { return d.forwarder.ForwardAutoplan(owner, command) },
	)
}

func (d *RoutedCommandDispatcher) DispatchPullClosed(command PullClosedDispatch) error {
	return d.route(
		command.OwnershipKey(),
		func(claimID string) error { return d.local.ExecutePullClosed(command, claimID) },
		func(owner ownership.Record) error { return d.forwarder.ForwardPullClosed(owner, command) },
	)
}

func (d *RoutedCommandDispatcher) route(key ownership.Key, execute func(string) error, forward func(ownership.Record) error) error {
	ctx, cancel := boundedOwnershipContext(context.Background())
	defer cancel()

	for attempt := 0; attempt < 2; attempt++ {
		if err := d.owners.Ready(ctx); err != nil {
			return fmt.Errorf("checking ownership readiness: %w", err)
		}
		owner, err := d.owners.Claim(ctx, key)
		if err != nil {
			return fmt.Errorf("claiming command owner: %w", err)
		}
		if owner.ReplicaID == d.replicaID && d.owners.Owns(key, owner.ClaimID) {
			err = execute(owner.ClaimID)
			if err == nil {
				return nil
			}
			if errors.Is(err, ErrOwnershipChanged) && attempt == 0 {
				continue
			}
			return err
		}

		err = forward(owner)
		if err == nil {
			return nil
		}
		if errors.Is(err, ErrOwnershipChanged) && attempt == 0 {
			continue
		}
		return fmt.Errorf("forwarding command to %s: %w", owner.ReplicaID, err)
	}
	return ErrOwnershipChanged
}

// LocalCommandExecutor hydrates credentials and runs commands on the owner.
type LocalCommandExecutor struct {
	Hydrator    RepoHydrator
	Runner      CommandRunner
	PullCleaner PullCleaner
	WorkingDir  PullStateCleaner
	ClaimGuard  *LocalClaimGuard
	Owners      ownership.Store
	Logger      logging.SimpleLogging
	TestingMode bool
	asyncWG     sync.WaitGroup
}

func (e *LocalCommandExecutor) ExecuteComment(command CommentDispatch, claimID string) error {
	baseRepo, err := e.hydrate(command.BaseRepo)
	if err != nil {
		return fmt.Errorf("hydrating base repository: %w", err)
	}
	var headRepo *models.Repo
	if command.HeadRepo != nil {
		hydrated, err := e.hydrate(*command.HeadRepo)
		if err != nil {
			return fmt.Errorf("hydrating head repository: %w", err)
		}
		headRepo = &hydrated
	}
	var pull *models.PullRequest
	if command.Pull != nil {
		hydrated := command.Pull.ToModel(baseRepo)
		pull = &hydrated
	}
	statePull := models.PullRequest{Num: command.PullNum, BaseRepo: baseRepo}
	if pull != nil {
		statePull = *pull
	}
	logger := e.commandLogger(baseRepo, command.PullNum)
	if e.Runner == nil {
		return errors.New("command runner is required")
	}
	lease, release, localStateReset, err := e.acquire(command.OwnershipKey(), claimID, logger, baseRepo, statePull, true)
	if err != nil {
		return err
	}
	if lease == nil {
		e.run(func() {
			e.Runner.RunCommentCommand(baseRepo, headRepo, pull, command.User, command.PullNum, command.Command)
		})
		return nil
	}
	routedRunner, ok := e.Runner.(RoutedCommandRunner)
	if !ok {
		release()
		return errors.New("routed command runner is required")
	}
	if err := lease.Admit(context.Background()); err != nil {
		release()
		return err
	}

	e.run(func() {
		defer release()
		routedRunner.RunRoutedCommentCommand(
			baseRepo, headRepo, pull, command.User, command.PullNum, command.Command,
			commandpkg.RoutingContext{Lease: lease, RecoverExternalPlans: localStateReset},
		)
	})
	return nil
}

func (e *LocalCommandExecutor) ExecuteAutoplan(command AutoplanDispatch, claimID string) error {
	baseRepo, err := e.hydrate(command.BaseRepo)
	if err != nil {
		return fmt.Errorf("hydrating base repository: %w", err)
	}
	headRepo, err := e.hydrate(command.HeadRepo)
	if err != nil {
		return fmt.Errorf("hydrating head repository: %w", err)
	}
	pull := command.Pull.ToModel(baseRepo)
	logger := e.commandLogger(baseRepo, pull.Num)
	if e.Runner == nil {
		return errors.New("command runner is required")
	}
	lease, release, localStateReset, err := e.acquire(command.OwnershipKey(), claimID, logger, baseRepo, pull, true)
	if err != nil {
		return err
	}
	if lease == nil {
		e.run(func() {
			e.Runner.RunAutoplanCommand(baseRepo, headRepo, pull, command.User)
		})
		return nil
	}
	routedRunner, ok := e.Runner.(RoutedCommandRunner)
	if !ok {
		release()
		return errors.New("routed command runner is required")
	}
	if err := lease.Admit(context.Background()); err != nil {
		release()
		return err
	}

	e.run(func() {
		defer release()
		routedRunner.RunRoutedAutoplanCommand(
			baseRepo, headRepo, pull, command.User,
			commandpkg.RoutingContext{Lease: lease, RecoverExternalPlans: localStateReset},
		)
	})
	return nil
}

func (e *LocalCommandExecutor) ExecutePullClosed(command PullClosedDispatch, claimID string) error {
	baseRepo, err := e.hydrate(command.BaseRepo)
	if err != nil {
		return fmt.Errorf("hydrating base repository: %w", err)
	}
	pull := command.Pull.ToModel(baseRepo)
	logger := e.commandLogger(baseRepo, pull.Num)
	if e.PullCleaner == nil {
		return errors.New("pull cleaner is required")
	}
	key := command.OwnershipKey()
	lease, release, _, err := e.acquire(key, claimID, logger, baseRepo, pull, false)
	if err != nil {
		return err
	}
	if lease != nil {
		if err := lease.Admit(context.Background()); err != nil {
			release()
			return err
		}
	}

	e.run(func() {
		defer release()
		if err := e.finishPullClosed(key, claimID, logger, baseRepo, pull); err != nil && logger != nil {
			logger.Err("unable to finish pull request cleanup: %s", err)
		}
	})
	return nil
}

func (e *LocalCommandExecutor) finishPullClosed(
	key ownership.Key,
	claimID string,
	logger logging.SimpleLogging,
	baseRepo models.Repo,
	pull models.PullRequest,
) error {
	if err := e.PullCleaner.CleanUpPull(logger, baseRepo, pull); err != nil {
		return fmt.Errorf("cleaning up closed pull request: %w", err)
	}
	if claimID == "" {
		return nil
	}
	if e.Owners == nil {
		return errors.New("ownership store is required")
	}
	ctx, cancel := boundedOwnershipContext(context.Background())
	defer cancel()
	if err := e.Owners.Release(ctx, key, claimID); err != nil {
		return fmt.Errorf("releasing pull request ownership: %w", err)
	}
	if err := e.ClaimGuard.Forget(key, claimID); err != nil {
		return fmt.Errorf("forgetting local ownership claim: %w", err)
	}
	return nil
}

func (e *LocalCommandExecutor) hydrate(ref RepoRef) (models.Repo, error) {
	if e.Hydrator == nil {
		return models.Repo{}, errors.New("repository hydrator is required")
	}
	return e.Hydrator.HydrateRepo(ref)
}

func (e *LocalCommandExecutor) acquire(
	key ownership.Key,
	claimID string,
	logger logging.SimpleLogging,
	repo models.Repo,
	pull models.PullRequest,
	resetLocalState bool,
) (commandpkg.ExecutionLease, func(), bool, error) {
	noopRelease := func() {}
	if claimID == "" {
		return nil, noopRelease, false, nil
	}
	if e.ClaimGuard == nil {
		return nil, noopRelease, false, errors.New("local claim guard is required")
	}
	if resetLocalState && e.WorkingDir == nil {
		return nil, noopRelease, false, errors.New("working directory is required")
	}
	if e.Owners == nil {
		return nil, noopRelease, false, errors.New("ownership store is required")
	}
	lease := &ownershipExecutionLease{owners: e.Owners, key: key, claimID: claimID}
	var reset func() error
	if resetLocalState {
		reset = func() error {
			if err := e.WorkingDir.Delete(logger, repo, pull); err != nil {
				return fmt.Errorf("resetting local pull state for new owner: %w", err)
			}
			return nil
		}
	}
	release, localStateReset, err := e.ClaimGuard.Acquire(key, claimID, func() error {
		return lease.Admit(context.Background())
	}, reset)
	if err != nil {
		return nil, noopRelease, false, err
	}
	return lease, release, localStateReset, nil
}

type ownershipExecutionLease struct {
	owners  ownership.Store
	key     ownership.Key
	claimID string
}

func (l *ownershipExecutionLease) Admit(ctx context.Context) error {
	ctx, cancel := boundedOwnershipContext(ctx)
	defer cancel()

	admitted, err := l.owners.Admit(ctx, l.key, l.claimID)
	if err != nil {
		return fmt.Errorf("admitting ownership claim: %w", err)
	}
	if !admitted {
		return ErrOwnershipChanged
	}
	return nil
}

func boundedOwnershipContext(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, ownershipOperationTimeout)
}

func (e *LocalCommandExecutor) commandLogger(repo models.Repo, pullNum int) logging.SimpleLogging {
	if e.Logger == nil {
		return nil
	}
	return e.Logger.With("repo", repo.FullName, "pull", strconv.Itoa(pullNum))
}

func (e *LocalCommandExecutor) run(fn func()) {
	if e.TestingMode {
		fn()
		return
	}
	e.asyncWG.Add(1)
	go func() {
		defer e.asyncWG.Done()
		fn()
	}()
}

// Wait blocks until all asynchronously accepted commands have returned.
// Call it only after HTTP shutdown prevents new dispatches.
func (e *LocalCommandExecutor) Wait() {
	e.asyncWG.Wait()
}
