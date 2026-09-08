// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/runatlantis/atlantis/server/controllers/web_templates"

	"github.com/gorilla/mux"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
)

// LocksController handles all requests relating to Atlantis locks.
type LocksController struct {
	Publication        *events.PublicationCoordinator
	AtlantisVersion    string                       `validate:"required"`
	AtlantisURL        *url.URL                     `validate:"required"`
	Locker             locking.Locker               `validate:"required"`
	Logger             logging.SimpleLogging        `validate:"required"`
	ApplyLocker        locking.ApplyLocker          `validate:"required"`
	VCSClient          vcs.Client                   `validate:"required"`
	LockDetailTemplate web_templates.TemplateWriter `validate:"required"`
	WorkingDir         events.WorkingDir            `validate:"required"`
	WorkingDirLocker   events.WorkingDirLocker      `validate:"required"`
	Database           db.Database                  `validate:"required"`
	DeleteLockCommand  events.DeleteLockCommand     `validate:"required"`
}

// LockApply handles creating a global apply lock.
// If Lock already exists it will be a no-op
func (l *LocksController) LockApply(w http.ResponseWriter, _ *http.Request) {
	lock, err := l.ApplyLocker.LockApply()
	if err != nil {
		l.respond(w, logging.Error, http.StatusInternalServerError, "creating apply lock failed with: %s", err)
		return
	}

	l.respond(w, logging.Info, http.StatusOK, "Apply Lock is acquired on %s", lock.Time.Format("2006-01-02 15:04:05"))
}

// UnlockApply handles releasing a global apply lock.
// If Lock doesn't exists it will be a no-op
func (l *LocksController) UnlockApply(w http.ResponseWriter, _ *http.Request) {
	err := l.ApplyLocker.UnlockApply()
	if err != nil {
		l.respond(w, logging.Error, http.StatusInternalServerError, "deleting apply lock failed with: %s", err)
		return
	}

	l.respond(w, logging.Info, http.StatusOK, "Deleted apply lock")
}

// GetLock is the GET /locks/{id} route. It renders the lock detail view.
func (l *LocksController) GetLock(w http.ResponseWriter, r *http.Request) {
	id, ok := mux.Vars(r)["id"]
	if !ok {
		l.respond(w, logging.Warn, http.StatusBadRequest, "No lock id in request")
		return
	}

	idUnencoded, err := url.QueryUnescape(id)
	if err != nil {
		l.respond(w, logging.Warn, http.StatusBadRequest, "Invalid lock id: %s", err)
		return
	}
	lock, err := l.Locker.GetLock(idUnencoded)
	if err != nil {
		l.respond(w, logging.Error, http.StatusInternalServerError, "Failed getting lock: %s", err)
		return
	}
	if lock == nil {
		l.respond(w, logging.Info, http.StatusNotFound, "No lock found at id '%s'", idUnencoded)
		return
	}

	owner, repo := models.SplitRepoFullName(lock.Project.RepoFullName)
	viewData := web_templates.LockDetailData{
		LockKeyEncoded:  id,
		LockKey:         idUnencoded,
		PullRequestLink: lock.Pull.URL,
		LockedBy:        lock.Pull.Author,
		Workspace:       lock.Workspace,
		AtlantisVersion: l.AtlantisVersion,
		CleanedBasePath: l.AtlantisURL.Path,
		RepoOwner:       owner,
		RepoName:        repo,
	}

	err = l.LockDetailTemplate.Execute(w, viewData)
	if err != nil {
		l.Logger.Err("%s", err.Error())
	}
}

// DeleteLock handles deleting the lock at id and commenting back on the
// pull request that the lock has been deleted.
func (l *LocksController) DeleteLock(w http.ResponseWriter, r *http.Request) {
	id, ok := mux.Vars(r)["id"]
	if !ok || id == "" {
		l.respond(w, logging.Warn, http.StatusBadRequest, "No lock id in request")
		return
	}

	idUnencoded, err := url.PathUnescape(id)
	if err != nil {
		l.respond(w, logging.Warn, http.StatusBadRequest, "Invalid lock id '%s'. Failed with error: '%s'", id, err)
		return
	}

	ctx := &command.Context{CommandContext: r.Context(), Log: l.Logger}
	if l.Publication != nil {
		// Resolve the pull before waiting. DeleteLock reads the lock again and
		// the required fence prevents changing a replacement pull's status.
		lock, err := l.Locker.GetLock(idUnencoded)
		if err != nil {
			l.respond(w, logging.Error, http.StatusInternalServerError, "reading lock failed: %s", err)
			return
		}
		if lock == nil {
			l.respond(w, logging.Info, http.StatusNotFound, "No lock found at id '%s'", idUnencoded)
			return
		}
		ctx.Pull = lock.Pull
	}
	var lock *models.ProjectLock
	var discarded bool
	err = l.Publication.RunCommand(ctx, func() error {
		var err error
		lock, discarded, err = l.DeleteLockCommand.DeleteLock(l.Logger, idUnencoded, ctx.PublicationMode())
		return err
	})
	if err != nil {
		code := http.StatusInternalServerError
		if errors.Is(err, db.ErrPlanStatusNotFound) || errors.Is(err, db.ErrPlanGenerationSuperseded) || errors.Is(err, db.ErrPublicationBusy) || errors.Is(err, db.ErrPublicationOwnerLost) || errors.Is(err, db.ErrPublicationAmbiguous) || errors.Is(err, events.ErrPublicationWaitLimit) {
			code = http.StatusConflict
			w.Header().Set("Retry-After", "1")
		} else if errors.Is(err, events.ErrPublicationShutdown) {
			code = http.StatusServiceUnavailable
			w.Header().Set("Retry-After", "1")
		} else if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			code = http.StatusRequestTimeout
		}
		l.respond(w, logging.Error, code, "deleting lock failed with: '%s'", err)
		return
	}

	if lock == nil {
		l.respond(w, logging.Info, http.StatusNotFound, "No lock found at id '%s'", idUnencoded)
		return
	}

	// NOTE: Because BaseRepo was added to the PullRequest model later, previous
	// installations of Atlantis will have locks in their DB that do not have
	// this field on PullRequest. We skip commenting in this case.
	if discarded && lock.Pull.BaseRepo != (models.Repo{}) {

		// This courtesy comment reports a completed discard. It is outside the
		// lease and does not authorize a plan or publish a reusable success check.
		comment := fmt.Sprintf("**Warning**: The plan for dir: `%s` workspace: `%s` was **discarded** via the Atlantis UI.\n\n"+
			"To `apply` this plan you must run `plan` again.", lock.Project.Path, lock.Workspace)
		if err = l.VCSClient.CreateComment(l.Logger, lock.Pull.BaseRepo, lock.Pull.Num, comment, ""); err != nil {
			l.Logger.Warn("failed commenting on pull request: %s", err)
		}
	} else {
		l.Logger.Debug("skipping discard comment because the plan was already applied or repository metadata is unavailable")
	}
	l.respond(w, logging.Info, http.StatusOK, "Deleted lock id '%s'", id)
}

// respond is a helper function to respond and log the response. lvl is the log
// level to log at, code is the HTTP response code.
func (l *LocksController) respond(w http.ResponseWriter, lvl logging.LogLevel, responseCode int, format string, args ...any) {
	response := fmt.Sprintf(format, args...)
	l.Logger.Log(lvl, response)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(responseCode)
	fmt.Fprintln(w, response) // #nosec G705 -- response body is served as text/plain, not interpreted as HTML
}
