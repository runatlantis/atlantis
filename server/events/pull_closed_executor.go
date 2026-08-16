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
	"time"

	"github.com/google/uuid"
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
func (p *PullClosedExecutor) CleanUpPull(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest) (cleanupErr error) {
	publicationToken := uuid.NewString()
	for {
		err := p.Database.AcquirePlanPublicationClaim(pull, publicationToken)
		if err == nil {
			break
		}
		if !errors.Is(err, db.ErrPlanPublicationBusy) {
			return fmt.Errorf("claiming pull cleanup publication boundary: %w", err)
		}
		time.Sleep(25 * time.Millisecond)
	}
	claimActive := true
	retainClaim := false
	defer func() {
		if !claimActive {
			return
		}
		if retainClaim {
			logger.Warn("retaining pull cleanup publication claim after ambiguous final comment publication")
			return
		}
		if err := p.Database.ReleasePlanPublicationClaim(pull, publicationToken); err != nil {
			cleanupErr = errors.Join(cleanupErr, fmt.Errorf("releasing pull cleanup publication boundary: %w", err))
		}
	}()

	closeGeneration := uuid.NewString()
	closedStatus, err := p.Database.BeginPlanGenerationReplacing(pull, nil, closeGeneration, publicationToken)
	if err != nil {
		return fmt.Errorf("invalidating durable plans before pull cleanup: %w", err)
	}

	for _, project := range closedStatus.Projects {
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

	var workspaceErr error
	if err := p.WorkingDir.Delete(logger, repo, pull); err != nil {
		workspaceErr = fmt.Errorf("cleaning workspace: %w", err)
	}

	// Always attempt external plan cleanup even if workspace deletion failed,
	// so S3 objects are not orphaned when local delete errors.
	if p.PlanStore != nil {
		if err := p.PlanStore.DeleteForPull(repo.Owner, repo.Name, pull.Num); err != nil {
			logger.Warn("failed to delete plans from external store: %s", err)
		}
	}

	if workspaceErr != nil {
		return workspaceErr
	}

	// Finally, delete locks. We do this last because when someone
	// unlocks a project, right now we don't actually delete the plan
	// so we might have plans laying around but no locks.
	locks, err := p.Locker.UnlockByPull(repo.FullName, pull.Num)
	if err != nil {
		return fmt.Errorf("cleaning up locks: %w", err)
	}

	// Clear any operations to avoid unbounded growth.
	if p.CancellationTracker != nil {
		p.CancellationTracker.Clear(pull)
	}

	// If there are no locks then there's no need to comment.
	if len(locks) == 0 {
		return nil
	}

	templateData := p.buildTemplateData(locks)
	var buf bytes.Buffer
	if err = pullClosedTemplate.Execute(&buf, templateData); err != nil {
		return fmt.Errorf("rendering template for comment: %w", err)
	}
	if err := p.VCSClient.CreateComment(logger, repo, pull.Num, buf.String(), ""); err != nil {
		retainClaim = true
		return err
	}
	return nil
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
