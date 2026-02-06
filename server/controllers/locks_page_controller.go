// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"net/http"
	"sort"
	"time"

	"github.com/runatlantis/atlantis/server/controllers/web_templates"
	"github.com/runatlantis/atlantis/server/events/models"
)

// LocksPageController handles the locks page
type LocksPageController struct {
	template        web_templates.TemplateWriter
	getLocks        func() (map[string]models.ProjectLock, error)
	isApplyLocked   func() bool
	atlantisVersion string
	cleanedBasePath string
}

// NewLocksPageController creates a new LocksPageController
func NewLocksPageController(
	template web_templates.TemplateWriter,
	getLocks func() (map[string]models.ProjectLock, error),
	isApplyLocked func() bool,
	atlantisVersion string,
	cleanedBasePath string,
) *LocksPageController {
	return &LocksPageController{
		template:        template,
		getLocks:        getLocks,
		isApplyLocked:   isApplyLocked,
		atlantisVersion: atlantisVersion,
		cleanedBasePath: cleanedBasePath,
	}
}

// Get renders the locks page
func (c *LocksPageController) Get(w http.ResponseWriter, r *http.Request) {
	locks, err := c.getLocks()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var lockData []web_templates.LockIndexData
	repoSet := make(map[string]bool)

	for id, lock := range locks {
		lockData = append(lockData, web_templates.LockIndexData{
			LockID:        id,
			LockPath:      "/lock?id=" + id,
			RepoFullName:  lock.Project.RepoFullName,
			PullNum:       lock.Pull.Num,
			Path:          lock.Project.Path,
			Workspace:     lock.Workspace,
			LockedBy:      lock.User.Username,
			Time:          lock.Time,
			TimeFormatted: lock.Time.Format(time.RFC1123),
		})
		repoSet[lock.Project.RepoFullName] = true
	}

	// Sort by time descending
	sort.Slice(lockData, func(i, j int) bool {
		return lockData[i].Time.After(lockData[j].Time)
	})

	// Extract unique repositories sorted
	repos := make([]string, 0, len(repoSet))
	for repo := range repoSet {
		repos = append(repos, repo)
	}
	sort.Strings(repos)

	data := web_templates.LocksPageData{
		LayoutData: web_templates.LayoutData{
			AtlantisVersion: c.atlantisVersion,
			CleanedBasePath: c.cleanedBasePath,
			ActiveNav:       "locks",
			ApplyLockActive: c.isApplyLocked(),
		},
		Locks:        lockData,
		TotalCount:   len(lockData),
		Repositories: repos,
	}

	if err := c.template.Execute(w, data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
