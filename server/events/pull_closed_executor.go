// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package events

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"slices"
	"sort"
	"strings"
	"text/template"

	"github.com/runatlantis/atlantis/server/logging"

	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/core/runtime"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/jobs"
)

//go:generate go tool pegomock generate github.com/runatlantis/atlantis/server/events --package mocks -o mocks/mock_resource_cleaner.go ResourceCleaner

type ResourceCleaner interface {
	CleanUp(pullInfo jobs.PullInfo)
}

//go:generate go tool pegomock generate github.com/runatlantis/atlantis/server/events --package mocks -o mocks/mock_pull_cleaner.go PullCleaner

// PullCleaner cleans up pull requests after they're closed/merged.
type PullCleaner interface {
	// CleanUpPull deletes the workspaces used by the pull request on disk
	// and deletes any locks associated with this pull request for all workspaces.
	CleanUpPull(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest) error
}

// PullClosedExecutor executes the tasks required to clean up a closed pull
// request.
type PullClosedExecutor struct {
	Locker                   locking.Locker
	VCSClient                vcs.Client
	WorkingDir               WorkingDir
	Database                 db.Database
	PullClosedTemplate       PullCleanupTemplate
	LogStreamResourceCleaner ResourceCleaner
	CancellationTracker      CancellationTracker
	PlanStore                runtime.PlanStore
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
	Execute(wr io.Writer, data any) error
}

type PullClosedEventTemplate struct{}

func (t *PullClosedEventTemplate) Execute(wr io.Writer, data any) error {
	return pullClosedTemplate.Execute(wr, data)
}

// CleanUpPull cleans up after a closed pull request.
func (p *PullClosedExecutor) CleanUpPull(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest) error {
	pullStatus, err := p.Database.GetPullStatus(pull)
	if err != nil {
		// Log and continue to clean up other resources.
		logger.Err("retrieving pull status: %s", err)
	}

	if pullStatus != nil {
		for _, project := range pullStatus.Projects {
			jobContext := jobs.PullInfo{
				PullNum:      pull.Num,
				Repo:         pull.BaseRepo.Name,
				RepoFullName: pull.BaseRepo.FullName,
				ProjectName:  project.ProjectName,
				Path:         project.RepoRelDir,
				Workspace:    project.Workspace,
			}
			p.LogStreamResourceCleaner.CleanUp(jobContext)
		}
	}

	// Collect cleanup errors but always attempt lock deletion. Returning early
	// after workspace failure left Redis project locks forever when the local
	// clone was already gone (e.g. multi-pod apply with external plan store).
	var cleanupErrs []error
	if err := p.WorkingDir.Delete(logger, repo, pull); err != nil {
		cleanupErrs = append(cleanupErrs, fmt.Errorf("cleaning workspace: %w", err))
	}

	// Always attempt external plan cleanup even if workspace deletion failed,
	// so S3 objects are not orphaned when local delete errors.
	if p.PlanStore != nil {
		if err := p.PlanStore.DeleteForPull(repo.Owner, repo.Name, pull.Num); err != nil {
			logger.Warn("failed to delete plans from external store: %s", err)
		}
	}

	// Delete locks even when workspace cleanup failed so merge does not leave
	// stale project locks that block other PRs.
	locks, err := p.Locker.UnlockByPull(repo.FullName, pull.Num)
	if err != nil {
		cleanupErrs = append(cleanupErrs, fmt.Errorf("cleaning up locks: %w", err))
	}

	// Delete pull from DB.
	if err := p.Database.DeletePullStatus(pull); err != nil {
		logger.Err("deleting pull from db: %s", err)
	}

	// Clear any operations to avoid unbounded growth.
	if p.CancellationTracker != nil {
		p.CancellationTracker.Clear(pull)
	}

	// Comment when we successfully unlocked at least one project.
	if len(locks) > 0 {
		templateData := p.buildTemplateData(locks)
		var buf bytes.Buffer
		if err = pullClosedTemplate.Execute(&buf, templateData); err != nil {
			cleanupErrs = append(cleanupErrs, fmt.Errorf("rendering template for comment: %w", err))
		} else if err := p.VCSClient.CreateComment(logger, repo, pull.Num, buf.String(), ""); err != nil {
			cleanupErrs = append(cleanupErrs, fmt.Errorf("creating cleanup comment: %w", err))
		}
	}

	return errors.Join(cleanupErrs...)
}

// buildTemplateData formats the lock data into a slice that can easily be
// templated for the VCS comment. We organize all the workspaces by their
// respective project paths so the comment can look like:
// dir: {dir}, workspaces: {all-workspaces}
func (p *PullClosedExecutor) buildTemplateData(locks []models.ProjectLock) []templatedProject {
	workspacesByPath := make(map[string][]string)
	for _, l := range locks {
		path := l.Project.Path
		// Check if workspace already exists to avoid duplicates
		if !slices.Contains(workspacesByPath[path], l.Workspace) {
			workspacesByPath[path] = append(workspacesByPath[path], l.Workspace)
		}
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
		sort.Strings(workspace)
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
