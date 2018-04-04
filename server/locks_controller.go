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

// EventsController handles all webhook requests which signify 'events' in the
// VCS host, ex. GitHub. It's split out from Server to make testing easier.
type LockController struct {
	AtlantisVersion    string
	Locker             locking.Locker
	Logger             *logging.SimpleLogger
	VCSClient          vcs.ClientProxy
	LockDetailTemplate TemplateWriter
}

var lockDeletedTemplate = template.Must(template.New("").Parse(
	"The following Locks and plans deleted were deleted via the Atlantis UI:\n" +
		"- path: `{{ .Path	}}` {{ .Workspace }}"))

func (l *LockController) GetLockRoute(w http.ResponseWriter, r *http.Request) {
	id, ok := mux.Vars(r)["id"]
	if !ok {
		l.respond(w, "No lock id in request")
		return
	}
	idUnencoded, err := url.QueryUnescape(id)
	if err != nil {
		l.respond(w, "Invalid lock id", err)
		return
	}
	lock, err := l.GetLock(idUnencoded)
	t := l.GetLockTemplate(lock, id, idUnencoded)
	if err != nil {
		l.respond(w, "Invalid lock id", err)
		return
	}
	err = l.LockDetailTemplate.Execute(w, t) // nolint: errcheck
	if err != nil {
		fmt.Fprint(w, err.Error())
	}

}

func (l *LockController) GetLock(id string) (*models.ProjectLock, error) {

	lock, err := l.Locker.GetLock(id)
	if err != nil {
		return nil, errors.Wrap(err, "Error getting lock")
	}
	if lock == nil {
		return nil, errors.New("No corresponding lock for given id")
	}

	return lock, nil

}

func (l *LockController) GetLockTemplate(lock *models.ProjectLock, id string, idUnencoded string) LockDetailData {
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

func (l *LockController) DeleteLockRoute(w http.ResponseWriter, r *http.Request) {
	id, ok := mux.Vars(r)["id"]
	if !ok || id == "" {
		l.respond(w, "No lock id in request")
		return
	}
	idUnencoded, err := url.PathUnescape(id)
	if err != nil {
		l.respond(w, "Invalid lock id: %s. Failed with %s", id, err)
		return
	}
	err = l.DeleteLock(idUnencoded)
	if err != nil {
		l.respond(w, "Failed to delete lock id: %s. Failed with %s", idUnencoded, err)
		return
	}
}

func (l *LockController) DeleteLock(id string) error {
	lock, err := l.Locker.Unlock(id)
	if err != nil {
		return errors.Wrap(err, "Failed to delete lock")
	}
	if lock == nil {
		return errors.New("No corresponding lock for given id")
	}
	err = l.CommentOnPullRequest(lock)
	if err != nil {
		return errors.Wrap(err, "error commenting on pull request")
	}
	return nil
}

// writes a commment on pull request
// exported for testing
func (l *LockController) CommentOnPullRequest(lock *models.ProjectLock) error {
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
	l.Logger.Debug("%v", lock.Pull.Repo)
	return l.VCSClient.CreateComment(lock.Pull.Repo, lock.Pull.Num, buf.String())
}

func (l *LockController) respond(w http.ResponseWriter, format string, args ...interface{}) {
	response := fmt.Sprintf(format, args...)
	l.Logger.Log(logging.Warn, response)
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprintln(w, response)
}
