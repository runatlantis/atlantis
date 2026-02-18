// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/runatlantis/atlantis/server/controllers/web_templates"

	"github.com/gorilla/mux"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
)

// LocksController handles all requests relating to Atlantis locks.
type LocksController struct {
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
	lockedBy := lock.Pull.Author
	if lock.IsManualLock {
		lockedBy = lock.User.Username
	}
	viewData := web_templates.LockDetailData{
		LockKeyEncoded:  id,
		LockKey:         idUnencoded,
		PullRequestLink: lock.Pull.URL,
		LockedBy:        lockedBy,
		Workspace:       lock.Workspace,
		AtlantisVersion: l.AtlantisVersion,
		CleanedBasePath: l.AtlantisURL.Path,
		RepoOwner:       owner,
		RepoName:        repo,
		Note:            lock.Note,
		IsManualLock:    lock.IsManualLock,
	}

	err = l.LockDetailTemplate.Execute(w, viewData)
	if err != nil {
		l.Logger.Err(err.Error())
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

	lock, err := l.DeleteLockCommand.DeleteLock(l.Logger, idUnencoded)
	if err != nil {
		l.respond(w, logging.Error, http.StatusInternalServerError, "deleting lock failed with: '%s'", err)
		return
	}

	if lock == nil {
		l.respond(w, logging.Info, http.StatusNotFound, "No lock found at id '%s'", idUnencoded)
		return
	}

	if lock.IsManualLock {
		l.Logger.Info("manual lock '%s' released", id)
	} else if lock.Pull.BaseRepo != (models.Repo{}) {
		// NOTE: Because BaseRepo was added to the PullRequest model later, previous
		// installations of Atlantis will have locks in their DB that do not have
		// this field on PullRequest. We skip commenting in this case.
		if err := l.Database.UpdateProjectStatus(lock.Pull, lock.Workspace, lock.Project.Path, models.DiscardedPlanStatus); err != nil {
			l.Logger.Err("unable to update project status: %s", err)
		}

		// Once the lock has been deleted, comment back on the pull request.
		comment := fmt.Sprintf("**Warning**: The plan for dir: `%s` workspace: `%s` was **discarded** via the Atlantis UI.\n\n"+
			"To `apply` this plan you must run `plan` again.", lock.Project.Path, lock.Workspace)
		if err = l.VCSClient.CreateComment(l.Logger, lock.Pull.BaseRepo, lock.Pull.Num, comment, ""); err != nil {
			l.Logger.Warn("failed commenting on pull request: %s", err)
		}
	} else {
		l.Logger.Debug("skipping commenting on pull request and deleting workspace because BaseRepo field is empty")
	}
	l.respond(w, logging.Info, http.StatusOK, "Deleted lock id '%s'", id)
}

// CreateManualLock handles creating a manual project lock via the web UI.
func (l *LocksController) CreateManualLock(w http.ResponseWriter, r *http.Request) {
	repoFullName := r.FormValue("repo_full_name")
	path := r.FormValue("path")
	projectName := r.FormValue("project_name")
	workspace := r.FormValue("workspace")
	note := r.FormValue("note")

	if repoFullName == "" || workspace == "" || note == "" {
		l.respond(w, logging.Warn, http.StatusBadRequest, "repo_full_name, workspace, and note are required")
		return
	}
	if path == "" {
		path = "."
	}

	// Derive username from basic auth if available, otherwise default.
	username := "atlantis-ui"
	if user, _, ok := r.BasicAuth(); ok && user != "" {
		username = user
	}

	project := models.NewProject(repoFullName, path, projectName)
	user := models.User{Username: username}

	lockAttempt, err := l.Locker.ManualLock(project, workspace, note, user)
	if err != nil {
		l.respond(w, logging.Error, http.StatusInternalServerError, "creating manual lock failed: %s", err)
		return
	}

	if !lockAttempt.LockAcquired {
		lockURL := fmt.Sprintf("%s/lock?id=%s", l.AtlantisURL.String(), url.QueryEscape(lockAttempt.LockKey))
		w.Header().Set("X-Lock-URL", lockURL)
		l.respond(w, logging.Warn, http.StatusConflict, "This project is already locked.")
		return
	}

	l.respond(w, logging.Info, http.StatusOK, "Manual lock created for %s", lockAttempt.LockKey)
}

// respond is a helper function to respond and log the response. lvl is the log
// level to log at, code is the HTTP response code.
func (l *LocksController) respond(w http.ResponseWriter, lvl logging.LogLevel, responseCode int, format string, args ...any) {
	response := fmt.Sprintf(format, args...)
	l.Logger.Log(lvl, response)
	w.WriteHeader(responseCode)
	fmt.Fprintln(w, response)
}
