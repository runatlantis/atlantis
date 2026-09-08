// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/google/go-github/v88/github"
	"github.com/google/uuid"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

var (
	ErrPublicationWaitLimit = errors.New("publication is busy; maximum wait reached, retry the command")
	ErrPublicationShutdown  = errors.New("server is shutting down; retry the command")
)

type PublicationClock interface {
	After(time.Duration) <-chan time.Time
}

type realPublicationClock struct{}

func (realPublicationClock) After(delay time.Duration) <-chan time.Time { return time.After(delay) }

// PublicationCoordinator scopes leases to durable transitions and VCS calls.
// Callers must finish Run before executing Terraform, and use a new Run to
// persist and publish its result after refreshing generation/head identity.
type PublicationCoordinator struct {
	Store               db.PublicationLeaseStore
	CancellationTracker CancellationTracker
	Clock               PublicationClock
	Shutdown            context.Context
	MaxWait             time.Duration
	LeaseTTL            time.Duration
	RetryInterval       time.Duration
}

func NewPublicationCoordinator(store db.PublicationLeaseStore, shutdown context.Context) *PublicationCoordinator {
	return &PublicationCoordinator{Store: store, Clock: realPublicationClock{}, Shutdown: shutdown, MaxWait: 5 * time.Second, LeaseTTL: 30 * time.Second, RetryInterval: 25 * time.Millisecond}
}

type HeldPublicationLease struct {
	mu            sync.Mutex
	deletionMu    sync.Mutex
	unresolved    bool
	statusDeleted bool
	coordinator   *PublicationCoordinator
	pull          models.PullRequest
	fence         db.PublicationFence
	ctx           context.Context
}

func (l *HeldPublicationLease) Fence() db.PublicationFence { return l.fence }

// Publish guards one terminal remote mutation. An unknown failure leaves the
// marker intact because the VCS may still finish that request after timeout.
// Pending/cosmetic calls use the ordinary lease and do not retain this marker.
func (l *HeldPublicationLease) Publish(remote func() error) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.unresolved {
		return db.ErrPublicationAmbiguous
	}
	if l.statusDeleted {
		return db.ErrPublicationOwnerLost
	}

	if err := l.ctx.Err(); err != nil {
		return context.Cause(l.ctx)
	}
	if err := l.coordinator.Store.BeginPublication(l.ctx, l.pull, l.fence); err != nil {
		// No remote call was issued. Resolve only this owner if the DB write may
		// have committed despite a cancelled response.
		cleanup, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = l.coordinator.Store.CompletePublication(cleanup, l.pull, l.fence)
		return err
	}
	remoteErr := remote()
	if remoteErr != nil && !publicationDefinitelyRejected(remoteErr) {
		l.unresolved = true
		return errors.Join(db.ErrPublicationAmbiguous, fmt.Errorf("publication owner %q: %w", l.fence.Owner, remoteErr))
	}
	cleanup, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := l.coordinator.Store.CompletePublication(cleanup, l.pull, l.fence); err != nil {
		l.unresolved = true
		return errors.Join(db.ErrPublicationAmbiguous, fmt.Errorf("recording known publication completion: %w", err), remoteErr)
	}
	return remoteErr
}

// DeleteStatus ends the critical section with the backend's atomic status and
// lease deletion. Renewal and release must not mistake that intentional deletion
// for lost ownership, and no later publication may use this lease.
func (l *HeldPublicationLease) DeleteStatus(remove func() error) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.deletionMu.Lock()
	defer l.deletionMu.Unlock()
	if l.unresolved {
		return db.ErrPublicationAmbiguous
	}
	if l.statusDeleted {
		return db.ErrPublicationOwnerLost
	}
	if err := l.ctx.Err(); err != nil {
		return context.Cause(l.ctx)
	}
	if err := remove(); err != nil {
		return err
	}
	l.statusDeleted = true
	return nil
}

// publicationDefinitelyRejected recognizes provider responses that confirm the
// mutation was rejected. Timeouts, server errors, and unrecognized errors remain
// ambiguous; an error string alone is not evidence of a completed request.
func publicationDefinitelyRejected(err error) bool {
	var response *http.Response
	switch e := err.(type) {
	case *github.ErrorResponse:
		response = e.Response
	case *github.RateLimitError:
		response = e.Response
	case *github.AbuseRateLimitError:
		response = e.Response
	case *gitlab.ErrorResponse:
		response = e.Response
	case interface{ Unwrap() []error }:
		children := e.Unwrap()
		if len(children) == 0 {
			return false
		}
		for _, child := range children {
			if !publicationDefinitelyRejected(child) {
				return false
			}
		}
		return true
	case interface{ Unwrap() error }:
		return publicationDefinitelyRejected(e.Unwrap())
	}
	return response != nil && response.StatusCode >= 400 && response.StatusCode < 500 && response.StatusCode != http.StatusRequestTimeout
}

// Run does not silently drop a busy command: it returns a retryable error for
// callers to render through their existing command/HTTP error path.
func (p *PublicationCoordinator) Run(ctx context.Context, pull models.PullRequest, operation func(*HeldPublicationLease) error) (resultErr error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if p.Store == nil || p.Clock == nil || p.MaxWait <= 0 || p.LeaseTTL <= 0 || p.RetryInterval <= 0 {
		return fmt.Errorf("publication coordinator is not configured")
	}
	commandCtx, cancelCommand := context.WithCancelCause(ctx)
	defer cancelCommand(context.Canceled)
	if p.Shutdown != nil {
		stop := context.AfterFunc(p.Shutdown, func() { cancelCommand(ErrPublicationShutdown) })
		defer stop()
		if p.Shutdown.Err() != nil {
			return ErrPublicationShutdown
		}
	}
	owner := uuid.NewString()
	acquired := false
	defer func() {
		if !acquired {
			cleanup, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_ = p.Store.ReleasePublicationLease(cleanup, pull, db.PublicationFence{Owner: owner})
		}
	}()
	waitCtx, cancelWait := context.WithCancelCause(commandCtx)
	defer cancelWait(context.Canceled)
	waitLimit := p.Clock.After(p.MaxWait)
	waitFinished := make(chan struct{})
	go func() {
		defer close(waitFinished)
		select {
		case <-waitLimit:
			cancelWait(ErrPublicationWaitLimit)
		case <-waitCtx.Done():
		}
	}()
	for {
		if waitCtx.Err() != nil {
			return context.Cause(waitCtx)
		}
		_, err := p.Store.AcquirePublicationLease(waitCtx, pull, owner, p.LeaseTTL)
		if err == nil {
			if waitCtx.Err() != nil {
				return context.Cause(waitCtx)
			}
			acquired = true
			break
		}
		if waitCtx.Err() != nil {
			return context.Cause(waitCtx)
		}
		if !errors.Is(err, db.ErrPublicationBusy) {
			return err
		}
		select {
		case <-waitCtx.Done():
			return context.Cause(waitCtx)
		case <-p.Clock.After(p.RetryInterval):
		}
	}
	cancelWait(context.Canceled)
	<-waitFinished
	held := &HeldPublicationLease{coordinator: p, pull: pull, fence: db.PublicationFence{Owner: owner}, ctx: commandCtx}
	stopRenewal := make(chan struct{})
	renewalDone := make(chan struct{})
	go func() {
		defer close(renewalDone)
		for {
			select {
			case <-stopRenewal:
				return
			case <-commandCtx.Done():
				return
			case <-p.Clock.After(p.LeaseTTL / 3):
				held.deletionMu.Lock()
				if held.statusDeleted {
					held.deletionMu.Unlock()
					return
				}
				_, err := p.Store.RenewPublicationLease(commandCtx, pull, held.fence, p.LeaseTTL)
				held.deletionMu.Unlock()
				if err != nil {
					cancelCommand(fmt.Errorf("renewing publication lease: %w", err))
					return
				}
			}
		}
	}()
	defer func() {
		close(stopRenewal)
		<-renewalDone
		if commandCtx.Err() != nil {
			resultErr = errors.Join(resultErr, context.Cause(commandCtx))
		}
		if held.statusDeleted {
			return
		}
		cleanup, cancelCleanup := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancelCleanup()
		releaseErr := p.Store.ReleasePublicationLease(cleanup, pull, held.fence)
		resultErr = errors.Join(resultErr, releaseErr)
	}()
	return operation(held)
}

// RunCommand binds required write ownership only for this short section. The
// synthetic API remains explicitly unfenced; it does not publish PR results.
func (p *PublicationCoordinator) RunCommand(ctx *command.Context, operation func() error) error {
	if p == nil || ctx.Pull.Num <= 0 {
		return operation()
	}
	ctx.PublicationRequired = true
	if ctx.TerminalPublisher != nil {
		return operation()
	}
	commandCtx := ctx.CommandContext
	if commandCtx == nil {
		commandCtx = context.Background()
	}
	if tracker, ok := p.CancellationTracker.(interface {
		CommandContext(context.Context, models.PullRequest) (context.Context, context.CancelFunc)
	}); ok {
		var cancel context.CancelFunc
		commandCtx, cancel = tracker.CommandContext(commandCtx, ctx.Pull)
		defer cancel()
	}
	return p.Run(commandCtx, ctx.Pull, func(held *HeldPublicationLease) error {
		if ctx.Log != nil {
			ctx.Log.Info("acquired publication lease %q", held.Fence().Owner)
		}
		oldFence, oldPublisher := ctx.PublicationFence, ctx.TerminalPublisher
		fence := held.Fence()
		ctx.PublicationFence, ctx.TerminalPublisher = &fence, held
		defer func() { ctx.PublicationFence, ctx.TerminalPublisher = oldFence, oldPublisher }()
		return operation()
	})
}

func publishTerminal(ctx *command.Context, remote func() error) error {
	if ctx.TerminalPublisher != nil {
		return ctx.TerminalPublisher.Publish(remote)
	}
	if ctx.PublicationRequired {
		return db.ErrPublicationOwnerLost
	}
	return remote()
}

// RunWithObservedStatus fences commands which do not install their own plan
// generation. Compare the durable status observed at command start before
// publishing an early failure, policy approval, or remediation result.
func (p *PublicationCoordinator) RunWithObservedStatus(ctx *command.Context, database db.Database, fetcher LivePullHeadFetcher, operation func() error) error {
	expected := ctx.PullStatus
	return p.RunCommand(ctx, func() error {
		if ctx.TerminalPublisher != nil {
			if fetcher != nil {
				live, err := fetcher.GetLivePullIdentity(command.ProjectContext{Log: ctx.Log, Pull: ctx.Pull, PullStatus: ctx.PullStatus, API: ctx.API})
				if err != nil {
					return err
				}
				if err := validateCommandStartIdentity(command.ProjectContext{Pull: ctx.Pull}, live); err != nil {
					return fmt.Errorf("%w: %w", db.ErrPlanGenerationSuperseded, err)
				}
			}
			if database == nil {
				return fmt.Errorf("reading publication status: database is unavailable")
			}
			current, err := database.GetPullStatus(ctx.Pull)
			if err != nil {
				return err
			}
			if !reflect.DeepEqual(expected, current) {
				return db.ErrPlanGenerationSuperseded
			}
		}
		return operation()
	})
}
