// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"reflect"

	"github.com/google/uuid"
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

	lock, err := l.Locker.GetLock(idUnencoded)
	if err != nil {
		l.respond(w, logging.Error, http.StatusInternalServerError, "getting lock before deletion failed with: '%s'", err)
		return
	}
	if lock == nil {
		l.respond(w, logging.Info, http.StatusNotFound, "No lock found at id '%s'", idUnencoded)
		return
	}
	if lock.Pull.BaseRepo == (models.Repo{}) {
		lock, err = l.DeleteLockCommand.DeleteLock(l.Logger, idUnencoded)
		if err != nil {
			l.respond(w, logging.Error, http.StatusInternalServerError, "deleting legacy lock failed with: '%s'", err)
			return
		}
		if lock == nil {
			l.respond(w, logging.Info, http.StatusNotFound, "No lock found at id '%s'", idUnencoded)
			return
		}
		l.Logger.Debug("skipping durable plan update and pull request comment because legacy lock BaseRepo is empty")
		l.respond(w, logging.Info, http.StatusOK, "Deleted lock id '%s'", id)
		return
	}

	publicationToken := uuid.NewString()
	if err := l.Database.AcquirePlanPublicationClaim(lock.Pull, publicationToken); err != nil {
		l.respond(w, logging.Warn, http.StatusConflict, "Lock cannot be deleted while plan state is being published: '%s'", err)
		return
	}
	claimActive := true
	retainClaim := false
	claimPull := lock.Pull
	defer func() {
		if !claimActive || retainClaim {
			return
		}
		if err := l.Database.ReleasePlanPublicationClaim(claimPull, publicationToken); err != nil {
			l.Logger.Err("releasing lock deletion publication claim: %s", err)
		}
	}()

	currentLock, err := l.Locker.GetLock(idUnencoded)
	if err != nil {
		l.respond(w, logging.Error, http.StatusInternalServerError, "rechecking lock before deletion failed with: '%s'", err)
		return
	}
	if currentLock == nil || !reflect.DeepEqual(lock, currentLock) {
		l.respond(w, logging.Warn, http.StatusConflict, "Lock changed before deletion; refresh and retry")
		return
	}

	pullStatus, err := l.Database.GetPullStatus(lock.Pull)
	if err != nil {
		l.respond(w, logging.Error, http.StatusInternalServerError, "reading durable plan status before lock deletion failed with: '%s'", err)
		return
	}
	if pullStatus != nil {
		var matchingProjects []models.ProjectStatus
		for _, project := range pullStatus.Projects {
			projectWorkspace := project.Workspace
			if projectWorkspace == "" {
				projectWorkspace = events.DefaultWorkspace
			}
			lockWorkspace := lock.Workspace
			if lockWorkspace == "" {
				lockWorkspace = events.DefaultWorkspace
			}
			if projectWorkspace != lockWorkspace || filepath.Clean(project.RepoRelDir) != filepath.Clean(lock.Project.Path) {
				continue
			}
			if lock.Project.ProjectName != "" && project.ProjectName != lock.Project.ProjectName {
				continue
			}
			matchingProjects = append(matchingProjects, project)
		}
		if lock.Project.ProjectName == "" && len(matchingProjects) > 1 {
			l.respond(w, logging.Warn, http.StatusConflict, "Lock matches multiple named projects; replan or unlock the pull request instead")
			return
		}
		if len(matchingProjects) == 1 {
			project := matchingProjects[0]
			if project.PlanGeneration != "" {
				l.respond(w, logging.Warn, http.StatusConflict, "Lock cannot be deleted while plan generation '%s' is active", project.PlanGeneration)
				return
			}
			_, err = l.Database.UpdateDiscardResultsForPlanGeneration(lock.Pull, []command.ProjectResult{{
				Command:                command.Unlock,
				Workspace:              project.Workspace,
				RepoRelDir:             project.RepoRelDir,
				ProjectName:            project.ProjectName,
				AcceptedPlanGeneration: project.AcceptedPlanGeneration,
			}}, publicationToken)
			if err != nil {
				statusCode := http.StatusInternalServerError
				if db.IsPlanGenerationObsolete(err) || errors.Is(err, db.ErrPlanGenerationStateInvalid) {
					statusCode = http.StatusConflict
				}
				l.respond(w, logging.Warn, statusCode, "discarding durable plan before lock deletion failed with: '%s'", err)
				return
			}
		}
	}

	lock, err = l.DeleteLockCommand.DeleteLock(l.Logger, idUnencoded)
	if err != nil {
		l.respond(w, logging.Error, http.StatusInternalServerError, "deleting lock failed with: '%s'", err)
		return
	}
	if lock == nil {
		l.respond(w, logging.Info, http.StatusNotFound, "No lock found at id '%s'", idUnencoded)
		return
	}

	// NOTE: Because BaseRepo was added to the PullRequest model later, previous
	// installations of Atlantis will have locks in their DB that do not have
	// this field on PullRequest. We skip commenting in this case.
	if lock.Pull.BaseRepo != (models.Repo{}) {
		// Once the lock has been deleted, comment back on the pull request.
		comment := fmt.Sprintf("**Warning**: The plan for dir: `%s` workspace: `%s` was **discarded** via the Atlantis UI.\n\n"+
			"To `apply` this plan you must run `plan` again.", lock.Project.Path, lock.Workspace)
		if err = l.VCSClient.CreateComment(l.Logger, lock.Pull.BaseRepo, lock.Pull.Num, comment, ""); err != nil {
			l.Logger.Warn("failed commenting on pull request: %s", err)
			retainClaim = true
		}
	} else {
		l.Logger.Debug("skipping commenting on pull request and deleting workspace because BaseRepo field is empty")
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
