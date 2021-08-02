// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.

package events

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/template"

	"github.com/runatlantis/atlantis/server/core/db"

	"github.com/runatlantis/atlantis/server/logging"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_pull_cleaner.go PullCleaner

// PullCleaner cleans up pull requests after they're closed/merged.
type PullCleaner interface {
	// CleanUpPull deletes the workspaces used by the pull request on disk
	// and deletes any locks associated with this pull request for all workspaces.
	CleanUpPull(repo models.Repo, pull models.PullRequest) error
}

// PullClosedExecutor executes the tasks required to clean up a closed pull
// request.
type PullClosedExecutor struct {
	Locker     locking.Locker
	VCSClient  vcs.Client
	WorkingDir WorkingDir
	Logger     logging.SimpleLogging
	DB         *db.BoltDB
}

type templatedProject struct {
	RepoRelDir    string
	Workspaces    string
	DequeueStatus string
}

var pullClosedTemplate = template.Must(template.New("").Parse(
	"Locks and plans deleted for the projects and workspaces modified in this pull request:\n" +
		"{{ range . }}\n" +
		"- dir: `{{ .RepoRelDir }}` {{ .Workspaces }}\n" +
		"{{ .DequeueStatus }}{{ end }}"))

// CleanUpPull cleans up after a closed pull request.
func (p *PullClosedExecutor) CleanUpPull(repo models.Repo, pull models.PullRequest) error {
	// TODO monikma extend the tests
	if err := p.WorkingDir.Delete(repo, pull); err != nil {
		return errors.Wrap(err, "cleaning workspace")
	}

	// Finally, delete locks. We do this last because when someone
	// unlocks a project, right now we don't actually delete the plan
	// so we might have plans laying around but no locks.
	locks, dequeueStatus, err := p.Locker.UnlockByPull(repo.FullName, pull.Num)
	if err != nil {
		return errors.Wrap(err, "cleaning up locks")
	}

	// Delete pull from DB.
	if err := p.DB.DeletePullStatus(pull); err != nil {
		p.Logger.Err("deleting pull from db: %s", err)
	}

	// If there are no locks then there's no need to comment.
	if len(locks) == 0 {
		return nil
	}

	templateData := p.buildTemplateData(locks, dequeueStatus)
	var buf bytes.Buffer
	if err = pullClosedTemplate.Execute(&buf, templateData); err != nil {
		return errors.Wrap(err, "rendering template for comment")
	}

	var commentErr = p.VCSClient.CreateComment(repo, pull.Num, buf.String(), "")

	// TODO monikma do you know a nicer method for potentially appending new error to the existing one
	commentErr = p.triggerPlansForDequeuedPRs(repo, dequeueStatus, commentErr)

	// TODO monikma if I am not mistaken, this method executes asynchronusly and the information of PRs being dequeued will not
	// bubble up to the Atlantis "Pull request closed" comment. That is not nice.
	return commentErr
}

func (p *PullClosedExecutor) triggerPlansForDequeuedPRs(repo models.Repo, dequeueStatus models.DequeueStatus, commentErr error) error {
	// TODO monikma #4 use exact dequeued comment instead of hardcoding it
	for _, lock := range dequeueStatus.ProjectLocks {
		planVcsMessage := "atlantis plan -d " + lock.Project.Path
		if err := p.VCSClient.CreateComment(repo, lock.Pull.Num, planVcsMessage, ""); err != nil {
			// TODO monikma at this point planning queue will be interrupted, how to resolve from this?
			commentErr = fmt.Errorf("%s\nunable to comment on PR %s: %s", commentErr, lock.Pull.Num, commentErr)
		}
	}
	return commentErr
}

// buildTemplateData formats the lock data into a slice that can easily be
// templated for the VCS comment. We organize all the workspaces by their
// respective project paths so the comment can look like:
// dir: {dir}, workspaces: {all-workspaces}
func (p *PullClosedExecutor) buildTemplateData(locks []models.ProjectLock, dequeueStatus models.DequeueStatus) []templatedProject {
	workspacesByPath := make(map[string][]string)
	for _, l := range locks {
		path := l.Project.Path
		workspacesByPath[path] = append(workspacesByPath[path], l.Workspace)
	}

	// sort keys so we can write deterministic tests
	var sortedPaths []string
	for p := range workspacesByPath {
		sortedPaths = append(sortedPaths, p)
	}
	sort.Strings(sortedPaths)

	var projects []templatedProject
	for _, p := range sortedPaths {
		workspace := workspacesByPath[p]
		workspacesStr := fmt.Sprintf("`%s`", strings.Join(workspace, "`, `"))
		if len(workspace) == 1 {
			projects = append(projects, templatedProject{
				RepoRelDir:    p,
				Workspaces:    "workspace: " + workspacesStr,
				DequeueStatus: dequeueStatus.StringFilterProject(p),
			})
		} else {
			projects = append(projects, templatedProject{
				RepoRelDir:    p,
				Workspaces:    "workspaces: " + workspacesStr,
				DequeueStatus: dequeueStatus.StringFilterProject(p),
			})

		}
	}
	return projects
}
