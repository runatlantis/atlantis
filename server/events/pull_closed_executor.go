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
	"io"
	"sort"
	"strings"
	"text/template"

	"github.com/runatlantis/atlantis/server/core/db"

	"github.com/runatlantis/atlantis/server/logging"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/jobs"
)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_resource_cleaner.go ResourceCleaner

type ResourceCleaner interface {
	CleanUp(pullInfo jobs.PullInfo)
}

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
	Locker                   locking.Locker
	VCSClient                vcs.Client
	WorkingDir               WorkingDir
	Logger                   logging.SimpleLogging
	DB                       *db.BoltDB
	PullClosedTemplate       PullCleanupTemplate
	LogStreamResourceCleaner ResourceCleaner
}

type templatedProject struct {
	RepoRelDir string
	Workspaces string
}

var pullClosedTemplate = template.Must(template.New("").Parse(
	"Locks and plans deleted for the projects and workspaces modified in this pull request:\n" +
		"{{ range . }}\n" +
		"- dir: `{{ .RepoRelDir }}` {{ .Workspaces }}{{ end }}"))

type PullCleanupTemplate interface {
	Execute(wr io.Writer, data interface{}) error
}

type PullClosedEventTemplate struct{}

func (t *PullClosedEventTemplate) Execute(wr io.Writer, data interface{}) error {
	return pullClosedTemplate.Execute(wr, data)
}

// CleanUpPull cleans up after a closed pull request.
func (p *PullClosedExecutor) CleanUpPull(repo models.Repo, pull models.PullRequest) error {
	pullStatus, err := p.DB.GetPullStatus(pull)
	if err != nil {
		// Log and continue to clean up other resources.
		p.Logger.Err("retrieving pull status: %s", err)
	}

	if pullStatus != nil {
		for _, project := range pullStatus.Projects {
			jobContext := jobs.PullInfo{
				PullNum:     pull.Num,
				Repo:        pull.BaseRepo.Name,
				Workspace:   project.Workspace,
				ProjectName: project.ProjectName,
			}
			p.LogStreamResourceCleaner.CleanUp(jobContext)
		}
	}

	if err := p.WorkingDir.Delete(repo, pull); err != nil {
		return errors.Wrap(err, "cleaning workspace")
	}

	// Finally, delete locks. We do this last because when someone
	// unlocks a project, right now we don't actually delete the plan
	// so we might have plans laying around but no locks.
	locks, err := p.Locker.UnlockByPull(repo.FullName, pull.Num)
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

	templateData := p.buildTemplateData(locks)
	var buf bytes.Buffer
	if err = pullClosedTemplate.Execute(&buf, templateData); err != nil {
		return errors.Wrap(err, "rendering template for comment")
	}
	return p.VCSClient.CreateComment(repo, pull.Num, buf.String(), "")
}

// buildTemplateData formats the lock data into a slice that can easily be
// templated for the VCS comment. We organize all the workspaces by their
// respective project paths so the comment can look like:
// dir: {dir}, workspaces: {all-workspaces}
func (p *PullClosedExecutor) buildTemplateData(locks []models.ProjectLock) []templatedProject {
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
				RepoRelDir: p,
				Workspaces: "workspace: " + workspacesStr,
			})
		} else {
			projects = append(projects, templatedProject{
				RepoRelDir: p,
				Workspaces: "workspaces: " + workspacesStr,
			})

		}
	}
	return projects
}
