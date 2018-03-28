package server

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/locking"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
)

// LocksController handles all webhook requests which signify 'events' in the
// VCS host, ex. GitHub. It's split out from Server to make testing easier.
type LocksController struct {
	AtlantisVersion    string
	Locker             locking.Locker
	Logger             *logging.SimpleLogger
	VCSClient          vcs.ClientProxy
	LockDetailTemplate TemplateWriter
}

var lockDeletedTemplate = template.Must(template.New("").Parse(
	"**Warning**: The plan for path: `{{ .Path }}` workspace: `{{ .Workspace }}` were deleted via the Atlantis UI.\n\n" +
		"To `apply` you must run `plan` again."))

// GetLockRoute is the GET /locks/{id} route. It renders the lock detail view.
func (l *LocksController) GetLockRoute(w http.ResponseWriter, r *http.Request) {
	id, ok := mux.Vars(r)["id"]
	if !ok {
		l.respond(w, http.StatusBadRequest, "No lock id in request")
		return
	}

	l.GetLock(w, r, id)
}

// GetLock handles a lock detail page view. getLockRoute is expected to
// be called before. This function was extracted to make it testable.
func (l *LocksController) GetLock(w http.ResponseWriter, _ *http.Request, id string) {
	idUnencoded, err := url.QueryUnescape(id)
	if err != nil {
		l.respond(w, http.StatusBadRequest, "Invalid lock id", err)
		return
	}
	lock, err := l.Locker.GetLock(idUnencoded)
	if err != nil {
		l.respond(w, http.StatusInternalServerError, "failed getting lock", err)
		return
	}
	if lock == nil {
		l.respond(w, http.StatusNotFound, "failed getting lock:", errors.New("no corresponding lock for given id"))
		return
	}
	t := l.GetLockTemplate(lock, id, idUnencoded) // nolint: errcheck
	l.LockDetailTemplate.Execute(w, t)
}

func (l *LocksController) GetLockTemplate(lock *models.ProjectLock, id string, idUnencoded string) LockDetailData {
	// Extract the repo owner and repo name.
	repo := strings.Split(lock.Project.RepoFullName, "/")
	return LockDetailData{
		LockKeyEncoded:  id,
		LockKey:         idUnencoded,
		RepoOwner:       repo[0],
		RepoName:        repo[1],
		PullRequestLink: lock.Pull.URL,
		LockedBy:        lock.Pull.Author,
		Workspace:       lock.Workspace,
		AtlantisVersion: l.AtlantisVersion,
	}
}

// DeleteLockRoute handles deleting the lock at id.
func (l *LocksController) DeleteLockRoute(w http.ResponseWriter, r *http.Request) {
	id, ok := mux.Vars(r)["id"]
	if !ok || id == "" {
		l.respond(w, http.StatusBadRequest, "No lock id in request")
		return
	}

	lock := l.DeleteLock(w, r, id)

	err := l.CommentOnPullRequest(lock)
	if err != nil {
		l.respond(w, http.StatusInternalServerError, "Failed commenting on pull request: %s", err)
		return
	}
	l.respond(w, http.StatusOK, "Deleted lock id %s", id)
}

// DeleteLock deletes the lock
// DeleteLockRoute should be called first. This method is split out to make this route testable.
func (l *LocksController) DeleteLock(w http.ResponseWriter, _ *http.Request, id string) *models.ProjectLock {
	idUnencoded, err := url.PathUnescape(id)
	if err != nil {
		l.respond(w, http.StatusBadRequest, "Invalid lock id: %s. Failed with error: %s", err)
		return nil
	}
	lock, err := l.Locker.Unlock(idUnencoded)
	if err != nil {
		l.respond(w, http.StatusInternalServerError, "deleting lock failed with: %s", err)
		return nil
	}
	if lock == nil {
		l.respond(w, http.StatusNotFound, "Error deleting lock: %s", errors.New("no corresponding lock for given id"))
		return nil
	}
	return lock
}

// Writes a commment on pull request
// Exported for testing
func (l *LocksController) CommentOnPullRequest(lock *models.ProjectLock) error {
	// templateData := buildTemplateData(locks)
	templateData := struct {
		Path      string
		Workspace string
	}{
		lock.Project.Path,
		lock.Workspace,
	}
	var buf bytes.Buffer
	if err := lockDeletedTemplate.Execute(&buf, templateData); err != nil {
		return errors.Wrap(err, "rendering template for comment")
	}
	l.Logger.Debug("%v", lock.Pull.HeadRepo)
	return l.VCSClient.CreateComment(lock.Pull.HeadRepo, lock.Pull.Num, buf.String())
}

func (l *LocksController) respond(w http.ResponseWriter, responseCode int, format string, args ...interface{}) {
	response := fmt.Sprintf(format, args...)
	l.Logger.Log(logging.Warn, response)
	w.WriteHeader(responseCode)
	fmt.Fprintln(w, response)
}
