// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/runatlantis/atlantis/server/core/ownership"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

// ErrOwnershipChanged asks an ingress replica to resolve ownership one more time.
var ErrOwnershipChanged = errors.New("pull request ownership changed")

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
	for attempt := 0; attempt < 2; attempt++ {
		if err := d.owners.Ready(context.Background()); err != nil {
			return fmt.Errorf("checking ownership readiness: %w", err)
		}
		owner, err := d.owners.Claim(context.Background(), key)
		if err != nil {
			return fmt.Errorf("claiming command owner: %w", err)
		}
		if owner.ReplicaID == d.replicaID && d.owners.Owns(key, owner.ClaimID) {
			return execute(owner.ClaimID)
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
	if err := e.prepare(command.OwnershipKey(), claimID, logger, baseRepo, statePull); err != nil {
		return err
	}
	if e.Runner == nil {
		return errors.New("command runner is required")
	}

	e.run(func() {
		e.Runner.RunCommentCommand(baseRepo, headRepo, pull, command.User, command.PullNum, command.Command)
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
	if err := e.prepare(command.OwnershipKey(), claimID, logger, baseRepo, pull); err != nil {
		return err
	}
	if e.Runner == nil {
		return errors.New("command runner is required")
	}

	e.run(func() {
		e.Runner.RunAutoplanCommand(baseRepo, headRepo, pull, command.User)
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
	if err := e.prepare(command.OwnershipKey(), claimID, logger, baseRepo, pull); err != nil {
		return err
	}
	if e.PullCleaner == nil {
		return errors.New("pull cleaner is required")
	}
	if err := e.PullCleaner.CleanUpPull(logger, baseRepo, pull); err != nil {
		return err
	}
	if claimID == "" {
		return nil
	}
	if e.Owners == nil {
		return errors.New("ownership store is required")
	}
	if err := e.Owners.Release(context.Background(), command.OwnershipKey(), claimID); err != nil {
		return fmt.Errorf("releasing pull request ownership: %w", err)
	}
	if err := e.ClaimGuard.Forget(command.OwnershipKey(), claimID); err != nil {
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

func (e *LocalCommandExecutor) prepare(key ownership.Key, claimID string, logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest) error {
	if claimID == "" {
		return nil
	}
	if e.ClaimGuard == nil {
		return errors.New("local claim guard is required")
	}
	if e.WorkingDir == nil {
		return errors.New("working directory is required")
	}
	if err := e.ClaimGuard.Prepare(key, claimID, func() error {
		return e.WorkingDir.Delete(logger, repo, pull)
	}); err != nil {
		return fmt.Errorf("resetting local pull state for new owner: %w", err)
	}
	return nil
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
