// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.
//
// This file wires Atlantis command ingress into the etcd active-active owner
// router (design §"Route coverage", §573). EtcdCommandRouter decorates the real
// CommandRunner: every comment command and autoplan is turned into a
// credential-free command envelope and dispatched through PR-ownership
// resolution, so exactly one replica executes the work for a given pull request
// and no handler runs directly on an arbitrary replica. It also implements the
// etcd Executor: when this replica owns (or is forwarded) a command, it fences
// the execution with a generation-bound barrier and runs the wrapped runner.
package events

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/runatlantis/atlantis/server/core/etcd"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
)

const (
	routedKindComment  = "comment"
	routedKindAutoplan = "autoplan"

	// routeIngressTimeout bounds the ingress dispatch: ownership resolution plus
	// at most one forward hop. It is generous relative to the etcd request timeout
	// so a single reroute still completes.
	routeIngressTimeout = 30 * time.Second

	// routeUnavailableComment is surfaced to the user when a command cannot be
	// routed to an owner or its fenced execution could not start. Atlantis fails
	// closed rather than risk two replicas acting on the same pull request.
	routeUnavailableComment = "Atlantis could not coordinate this command across its high-availability replicas right now. No action was taken; please try again."

	// routeBlockedComment is surfaced when an unresolved execution from an earlier
	// owner generation still fences this pull request (design §558).
	routeBlockedComment = "Atlantis has an unresolved in-flight execution for this pull request from a previous server generation. No action was taken; this pull request is held until the earlier execution is resolved by an operator."

	// routeUncertainComment is surfaced when a command ran but the owning replica
	// lost its lease during execution, so the outcome cannot be confirmed.
	routeUncertainComment = "Atlantis started this command but lost ownership of the pull request during execution across its high-availability replicas, so the outcome is **uncertain**. Do not assume it completed or was skipped; verify the actual state (e.g. Terraform state / provider) before retrying. This pull request is held until an operator resolves it."
)

// EtcdCoordinator is the subset of the etcd runtime the router depends on. It
// keeps every unexported coordination type inside the etcd package, so the
// router is unit-testable with a fake.
type EtcdCoordinator interface {
	// Route resolves PR ownership and admits the command locally or forwards it
	// to the owning replica (ingress side).
	Route(ctx context.Context, cmd etcd.Command) (etcd.Result, error)
	// Execute fences and runs an admitted command on the owning replica, invoking
	// run to perform the actual work (owner side).
	Execute(cmd etcd.Command, run func() bool) etcd.ExecuteOutcome
	// ReopenPull clears a closed pull's lifecycle so a reopened pull request
	// accepts new locks again. It is a no-op when the pull is not closed.
	ReopenPull(ctx context.Context, vcsHostname, repoFullName string, pullNum int) error
}

// routedPayload is the serialized command body forwarded between replicas. It
// carries no VCS credentials — the executing replica uses its own configured VCS
// client. Pointers mirror the nullable arguments of the CommandRunner methods.
type routedPayload struct {
	Kind     string              `json:"kind"`
	BaseRepo models.Repo         `json:"base_repo"`
	HeadRepo *models.Repo        `json:"head_repo,omitempty"`
	Pull     *models.PullRequest `json:"pull,omitempty"`
	User     models.User         `json:"user"`
	PullNum  int                 `json:"pull_num"`
	Comment  *CommentCommand     `json:"comment,omitempty"`
}

// EtcdCommandRouter decorates a CommandRunner with owner-aware routing. It
// satisfies both CommandRunner (ingress) and etcd.Executor (owner-side run).
type EtcdCommandRouter struct {
	delegate    CommandRunner
	coordinator EtcdCoordinator
	vcsClient   vcs.Client
	logger      logging.SimpleLogging
}

// NewEtcdCommandRouter builds the routing decorator over the real command runner.
func NewEtcdCommandRouter(delegate CommandRunner, coordinator EtcdCoordinator, vcsClient vcs.Client, logger logging.SimpleLogging) *EtcdCommandRouter {
	return &EtcdCommandRouter{
		delegate:    delegate,
		coordinator: coordinator,
		vcsClient:   vcsClient,
		logger:      logger,
	}
}

// RunAutoplanCommand dispatches an autoplan through owner routing. The wrapped
// runner executes it on whichever replica owns the pull request.
func (r *EtcdCommandRouter) RunAutoplanCommand(baseRepo models.Repo, headRepo models.Repo, pull models.PullRequest, user models.User) {
	// An autoplan follows an opened/updated/reopened pull event, so it is the
	// signal that a previously-closed pull is live again: clear any closed
	// lifecycle before routing so the pull accepts new locks (design §450).
	reopenCtx, cancel := context.WithTimeout(context.Background(), routeIngressTimeout)
	if err := r.coordinator.ReopenPull(reopenCtx, baseRepo.VCSHost.Hostname, baseRepo.FullName, pull.Num); err != nil {
		r.logger.Warn("etcd routing: reopening pull %s#%d lifecycle: %s", baseRepo.FullName, pull.Num, err)
	}
	cancel()

	hr := headRepo
	p := pull
	payload := routedPayload{
		Kind:     routedKindAutoplan,
		BaseRepo: baseRepo,
		HeadRepo: &hr,
		Pull:     &p,
		User:     user,
		PullNum:  pull.Num,
	}
	// Autoplan has no provider comment id; the head commit is its stable identity,
	// so a redelivered push event for the same commit deduplicates while a new push
	// (new commit) runs (design §509).
	dedupKey := ""
	if pull.HeadCommit != "" {
		dedupKey = fmt.Sprintf("%s#%d/autoplan/%s", baseRepo.FullName, pull.Num, pull.HeadCommit)
	}
	r.dispatch("autoplan", baseRepo, pull.Num, dedupKey, payload)
}

// RunCommentCommand dispatches a parsed comment command through owner routing.
func (r *EtcdCommandRouter) RunCommentCommand(baseRepo models.Repo, maybeHeadRepo *models.Repo, maybePull *models.PullRequest, user models.User, pullNum int, cmd *CommentCommand) {
	payload := routedPayload{
		Kind:     routedKindComment,
		BaseRepo: baseRepo,
		HeadRepo: maybeHeadRepo,
		Pull:     maybePull,
		User:     user,
		PullNum:  pullNum,
		Comment:  cmd,
	}
	// The provider's stable comment id (set on ingress) is the dedup identity, so a
	// redelivered comment webhook executes once while a genuinely new comment — even
	// an identical re-issued command — runs (design §509).
	dedupKey := ""
	if cmd != nil && cmd.DeliveryID != "" {
		dedupKey = fmt.Sprintf("%s#%d/comment/%s", baseRepo.FullName, pullNum, cmd.DeliveryID)
	}
	r.dispatch("webhook", baseRepo, pullNum, dedupKey, payload)
}

// dispatch builds the command envelope and routes it. dedupKey is a stable,
// globally-unique identity for the logical command (derived from the provider
// comment id or the autoplan head commit); when empty, a fresh id is used so the
// single-owner guarantee still holds without deduplication. On any non-accepted
// result it fails closed with a user-visible comment rather than silently
// dropping the command or letting an arbitrary replica run it.
func (r *EtcdCommandRouter) dispatch(sourceKind string, baseRepo models.Repo, pullNum int, dedupKey string, payload routedPayload) {
	body, err := json.Marshal(payload)
	if err != nil {
		r.logger.Err("etcd routing: encoding command for pull %s#%d: %s", baseRepo.FullName, pullNum, err)
		r.failClosed(baseRepo, pullNum, routeUnavailableComment)
		return
	}

	// A stable dedup key makes a redelivered webhook resolve to the same admission
	// record (executing once, design §509). Absent one, a fresh id preserves
	// single-owner execution without deduplicating (non-HA behavior).
	deliveryID := dedupKey
	if deliveryID == "" {
		deliveryID = uuid.NewString()
	}

	cmd := etcd.Command{
		Identity: etcd.CommandIdentity{
			SourceKind:  sourceKind,
			VCSHostname: baseRepo.VCSHost.Hostname,
			DeliveryID:  deliveryID,
		},
		Scope: etcd.PullScope{
			VCSHostname: baseRepo.VCSHost.Hostname,
			Repository:  baseRepo.FullName,
			PullNum:     pullNum,
		},
		Body: body,
	}

	ctx, cancel := context.WithTimeout(context.Background(), routeIngressTimeout)
	defer cancel()

	res, err := r.coordinator.Route(ctx, cmd)
	if err != nil {
		r.logger.Err("etcd routing: dispatching command for pull %s#%d: %s", baseRepo.FullName, pullNum, err)
		r.failClosed(baseRepo, pullNum, routeUnavailableComment)
		return
	}

	if res.Status != http.StatusAccepted {
		r.logger.Warn("etcd routing: command for pull %s#%d not admitted (status %d: %s)", baseRepo.FullName, pullNum, res.Status, res.Message)
		r.failClosed(baseRepo, pullNum, routeUnavailableComment)
	}
}

// Register implements etcd.Executor. It decodes an admitted command and runs it
// on this replica under a generation-bound execution barrier. It returns quickly;
// the barrier, admission lifecycle, and actual work run in a goroutine.
func (r *EtcdCommandRouter) Register(_ context.Context, cmd etcd.Command) error {
	var payload routedPayload
	if err := json.Unmarshal(cmd.Body, &payload); err != nil {
		return err
	}
	go r.execute(cmd, payload)
	return nil
}

// execute fences and runs an admitted command, surfacing a fail-closed comment
// when the fenced execution could not start.
func (r *EtcdCommandRouter) execute(cmd etcd.Command, payload routedPayload) {
	outcome := r.coordinator.Execute(cmd, func() bool { return r.run(payload) })
	switch outcome {
	case etcd.ExecuteRan:
		// The wrapped runner reported its own status/comments.
	case etcd.ExecuteBlocked:
		r.logger.Warn("etcd routing: pull %s#%d blocked by an older-generation execution barrier", payload.BaseRepo.FullName, payload.PullNum)
		r.failClosed(payload.BaseRepo, payload.PullNum, routeBlockedComment)
	case etcd.ExecuteUncertain:
		// The command ran but ownership was lost mid-execution, so its outcome
		// cannot be confirmed. Do not claim "no action taken".
		r.logger.Warn("etcd routing: pull %s#%d execution outcome is uncertain (ownership lost during execution)", payload.BaseRepo.FullName, payload.PullNum)
		r.failClosed(payload.BaseRepo, payload.PullNum, routeUncertainComment)
	default:
		r.logger.Warn("etcd routing: fenced execution for pull %s#%d did not start (outcome %d)", payload.BaseRepo.FullName, payload.PullNum, outcome)
		r.failClosed(payload.BaseRepo, payload.PullNum, routeUnavailableComment)
	}
}

// run invokes the wrapped runner for the decoded command. A panic is contained
// and reported as a failed execution so the barrier is still cleared.
func (r *EtcdCommandRouter) run(payload routedPayload) (success bool) {
	defer func() {
		if p := recover(); p != nil {
			r.logger.Err("etcd routing: recovered panic executing pull %s#%d: %v", payload.BaseRepo.FullName, payload.PullNum, p)
			success = false
		}
	}()
	switch payload.Kind {
	case routedKindAutoplan:
		if payload.HeadRepo == nil || payload.Pull == nil {
			r.logger.Err("etcd routing: autoplan payload for pull %s#%d missing head repo or pull", payload.BaseRepo.FullName, payload.PullNum)
			return false
		}
		r.delegate.RunAutoplanCommand(payload.BaseRepo, *payload.HeadRepo, *payload.Pull, payload.User)
	case routedKindComment:
		r.delegate.RunCommentCommand(payload.BaseRepo, payload.HeadRepo, payload.Pull, payload.User, payload.PullNum, payload.Comment)
	default:
		r.logger.Err("etcd routing: unknown command kind %q for pull %s#%d", payload.Kind, payload.BaseRepo.FullName, payload.PullNum)
		return false
	}
	return true
}

// failClosed comments on the pull request that no action was taken. Best-effort:
// a comment failure is logged, never retried into a duplicate command.
func (r *EtcdCommandRouter) failClosed(baseRepo models.Repo, pullNum int, msg string) {
	if r.vcsClient == nil {
		return
	}
	if err := r.vcsClient.CreateComment(r.logger, baseRepo, pullNum, msg, ""); err != nil {
		r.logger.Err("etcd routing: commenting fail-closed on pull %s#%d: %s", baseRepo.FullName, pullNum, err)
	}
}
