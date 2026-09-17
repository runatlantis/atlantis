// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

var (
	ErrPublicationRecoveryNotNeeded = errors.New("publication outcome is not ambiguous; expired leases recover automatically")
	ErrPublicationBusy              = errors.New("another command is publishing this pull request; retry shortly")
	ErrPublicationOwnerLost         = errors.New("publication lease ownership expired or changed; retry the command")
	ErrPublicationAmbiguous         = errors.New("an earlier terminal VCS publication has an unknown outcome; reconcile that publication before retrying")
)

// PublicationFence is a required argument of fenced status mutations. An empty
// fence is invalid; legacy callers use a separate explicit unfenced operation.
type PublicationFence = command.PublicationFence

// PublicationLease protects short durable-transition/publication sections, never
// Terraform execution. Publishing marks an external request which may complete
// after its caller crashes. An expired publishing record cannot safely be
// replaced: VCS providers do not enforce the database's fencing token.
type PublicationLease struct {
	Owner             string
	DeadlineUnixMilli int64
	Publishing        bool
}

// AcquirePublicationLease is a pure transition committed atomically by each
// backend. Supplying time explicitly keeps expiry tests deterministic.
func AcquirePublicationLease(current *PublicationLease, owner string, now time.Time, ttl time.Duration) (PublicationLease, error) {
	if owner == "" || ttl <= 0 {
		return PublicationLease{}, fmt.Errorf("invalid publication lease owner or duration")
	}
	if current != nil {
		if current.DeadlineUnixMilli > now.UnixMilli() {
			return PublicationLease{}, ErrPublicationBusy
		}
		if current.Publishing {
			return PublicationLease{}, ErrPublicationAmbiguous
		}
		// Tokens identify one acquisition, not a process or a reusable lock name.
		if current.Owner == owner {
			return PublicationLease{}, ErrPublicationOwnerLost
		}
	}
	return PublicationLease{Owner: owner, DeadlineUnixMilli: now.Add(ttl).UnixMilli()}, nil
}

func CheckPublicationFence(current *PublicationLease, fence PublicationFence, now time.Time) error {
	if current == nil || fence.Owner == "" || current.Owner != fence.Owner || current.DeadlineUnixMilli <= now.UnixMilli() {
		return ErrPublicationOwnerLost
	}
	return nil
}

func RenewPublicationLease(current *PublicationLease, fence PublicationFence, now time.Time, ttl time.Duration) (PublicationLease, error) {
	if err := CheckPublicationFence(current, fence, now); err != nil {
		return PublicationLease{}, err
	}
	if ttl <= 0 {
		return PublicationLease{}, fmt.Errorf("invalid publication lease duration")
	}
	next := *current
	next.DeadlineUnixMilli = now.Add(ttl).UnixMilli()
	return next, nil
}

func BeginPublication(current *PublicationLease, fence PublicationFence, now time.Time) (PublicationLease, error) {
	if err := CheckPublicationFence(current, fence, now); err != nil {
		return PublicationLease{}, err
	}
	if current.Publishing {
		return PublicationLease{}, ErrPublicationAmbiguous
	}
	next := *current
	next.Publishing = true
	return next, nil
}

// CompletePublication records a known remote outcome, even if ownership expired
// while awaiting the response. It cannot permit more remote calls: subsequent
// calls must pass CheckPublicationFence again. No replacement can acquire an
// expired Publishing lease before this exact owner resolves it.
func CompletePublication(current *PublicationLease, fence PublicationFence) (PublicationLease, error) {
	if current == nil || fence.Owner == "" || current.Owner != fence.Owner {
		return PublicationLease{}, ErrPublicationOwnerLost
	}
	next := *current
	next.Publishing = false
	return next, nil
}

// CanReleasePublicationLease never deletes a replacement owner or an unresolved
// remote publication. Normal pre-execution failures leave Publishing false and
// can always release their exact ownership, including after expiry.
func CanReleasePublicationLease(current *PublicationLease, fence PublicationFence) error {
	if current == nil || fence.Owner == "" || current.Owner != fence.Owner {
		return ErrPublicationOwnerLost
	}
	if current.Publishing {
		return ErrPublicationAmbiguous
	}
	return nil
}

// PublicationLeaseStore provides atomic owner/deadline transitions. Callers must
// use the returned owner token for every fenced durable write and publication.
type PublicationLeaseStore interface {
	AcquirePublicationLease(context.Context, models.PullRequest, string, time.Duration) (PublicationLease, error)
	RenewPublicationLease(context.Context, models.PullRequest, PublicationFence, time.Duration) (PublicationLease, error)
	BeginPublication(context.Context, models.PullRequest, PublicationFence) error
	CompletePublication(context.Context, models.PullRequest, PublicationFence) error
	ReleasePublicationLease(context.Context, models.PullRequest, PublicationFence) error
}

// CheckPublicationWrite is evaluated inside the same backend transaction as the
// durable mutation, including deletion. Unfenced writes cannot bypass a live or
// ambiguous publication lease.
func CheckPublicationWrite(current *PublicationLease, mode command.PublicationWriteMode, now time.Time) error {
	switch fence := mode.(type) {
	case command.PublicationFence:
		if err := CheckPublicationFence(current, fence, now); err != nil {
			return err
		}
		if current.Publishing {
			return ErrPublicationAmbiguous
		}
		return nil
	case command.NoClaim:
		if current == nil {
			return nil
		}
		if current.Publishing {
			return ErrPublicationAmbiguous
		}
		if current.DeadlineUnixMilli > now.UnixMilli() {
			return ErrPublicationBusy
		}
		return nil
	default:
		return ErrPublicationOwnerLost
	}
}

// PublicationRecoveryStore supports exceptional operator recovery for both
// backends. The caller must first establish that the original remote operation
// cannot still complete. Normal expired, non-publishing leases need no recovery.
type PublicationRecoveryStore interface {
	GetPublicationLease(context.Context, models.PullRequest) (*PublicationLease, error)
	RecoverPublicationLease(context.Context, models.PullRequest, PublicationFence) error
}

func CanRecoverPublicationLease(current *PublicationLease, fence PublicationFence, now time.Time) error {
	if current == nil || fence.Owner == "" || current.Owner != fence.Owner {
		return ErrPublicationOwnerLost
	}
	if current.DeadlineUnixMilli > now.UnixMilli() {
		return ErrPublicationBusy
	}
	if !current.Publishing {
		return ErrPublicationRecoveryNotNeeded
	}
	return nil
}
