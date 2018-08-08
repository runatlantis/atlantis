package server

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/locking"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
)

// LocksController handles all requests relating to Atlantis locks.
type LocksController struct {
	AtlantisVersion    string
	Locker             locking.Locker
	Logger             *logging.SimpleLogger
	VCSClient          vcs.ClientProxy
	LockDetailTemplate TemplateWriter
	WorkingDir         events.WorkingDir
	WorkingDirLocker   events.WorkingDirLocker
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
		l.respond(w, logging.Info, http.StatusNotFound, "No lock found at id %q", idUnencoded)
		return
	}

	// Extract the repo owner and repo name.
	repo := strings.Split(lock.Project.RepoFullName, "/")
	viewData := LockDetailData{
		LockKeyEncoded:  id,
		LockKey:         idUnencoded,
		RepoOwner:       repo[0],
		RepoName:        repo[1],
		PullRequestLink: lock.Pull.URL,
		LockedBy:        lock.Pull.Author,
		Workspace:       lock.Workspace,
		AtlantisVersion: l.AtlantisVersion,
	}
	l.LockDetailTemplate.Execute(w, viewData) // nolint: errcheck
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
		l.respond(w, logging.Warn, http.StatusBadRequest, "Invalid lock id %q. Failed with error: %s", id, err)
		return
	}
	lock, err := l.Locker.Unlock(idUnencoded)
	if err != nil {
		l.respond(w, logging.Error, http.StatusInternalServerError, "deleting lock failed with: %s", err)
		return
	}
	if lock == nil {
		l.respond(w, logging.Info, http.StatusNotFound, "No lock found at id %q", idUnencoded)
		return
	}

	// NOTE: Because BaseRepo was added to the PullRequest model later, previous
	// installations of Atlantis will have locks in their DB that do not have
	// this field on PullRequest. We skip commenting and deleting the working dir in this case.
	if lock.Pull.BaseRepo != (models.Repo{}) {
		unlock, err := l.WorkingDirLocker.TryLock(lock.Pull.BaseRepo.FullName, lock.Workspace, lock.Pull.Num)
		if err != nil {
			l.Logger.Err("unable to obtain working dir lock when trying to delete old plans: %s", err)
		} else {
			defer unlock()
			// nolint: vetshadow
			if err := l.WorkingDir.DeleteForWorkspace(lock.Pull.BaseRepo, lock.Pull, lock.Workspace); err != nil {
				l.Logger.Err("unable to delete workspace: %s", err)
			}
		}

		// Once the lock has been deleted, comment back on the pull request.
		comment := fmt.Sprintf("**Warning**: The plan for dir: `%s` workspace: `%s` was **discarded** via the Atlantis UI.\n\n"+
			"To `apply` this plan you must run `plan` again.", lock.Project.Path, lock.Workspace)
		err = l.VCSClient.CreateComment(lock.Pull.BaseRepo, lock.Pull.Num, comment)
		if err != nil {
			l.respond(w, logging.Error, http.StatusInternalServerError, "Failed commenting on pull request: %s", err)
			return
		}
	} else {
		l.Logger.Debug("skipping commenting on pull request and deleting workspace because BaseRepo field is empty")
	}
	l.respond(w, logging.Info, http.StatusOK, "Deleted lock id %q", id)
}

// respond is a helper function to respond and log the response. lvl is the log
// level to log at, code is the HTTP response code.
func (l *LocksController) respond(w http.ResponseWriter, lvl logging.LogLevel, responseCode int, format string, args ...interface{}) {
	response := fmt.Sprintf(format, args...)
	l.Logger.Log(lvl, response)
	w.WriteHeader(responseCode)
	fmt.Fprintln(w, response)
}
