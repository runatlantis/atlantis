// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/runatlantis/atlantis/server/core/runtime"
	"github.com/runatlantis/atlantis/server/utils"
)

//go:generate go tool pegomock generate --package mocks -o mocks/mock_pending_plan_finder.go PendingPlanFinder

type PendingPlanFinder interface {
	Find(pullDir string) ([]PendingPlan, error)
	DeletePlans(pullDir string) error
}

// DefaultPendingPlanFinder finds unapplied plans.
type DefaultPendingPlanFinder struct{}

// PendingPlan is a plan that has not been applied.
type PendingPlan struct {
	// RepoDir is the absolute path to the root of the repo that holds this
	// plan.
	RepoDir string
	// RepoRelDir is the relative path from the repo to the project that
	// the plan is for.
	RepoRelDir string
	// Workspace is the workspace this plan should execute in.
	Workspace   string
	ProjectName string
}

// Find finds all pending plans in pullDir. pullDir should be the working
// directory where Atlantis will operate on this pull request. It's one level
// up from where Atlantis clones the repo for each workspace.
func (p *DefaultPendingPlanFinder) Find(pullDir string) ([]PendingPlan, error) {
	plans, _, err := p.findWithAbsPaths(pullDir)
	return plans, err
}

func (p *DefaultPendingPlanFinder) findWithAbsPaths(pullDir string) ([]PendingPlan, []string, error) {
	workspaceDirs, err := os.ReadDir(pullDir)
	if err != nil {
		return nil, nil, err
	}
	var plans []PendingPlan
	var absPaths []string
	for _, workspaceDir := range workspaceDirs {
		// Skip non-directory entries (files, symlinks) — workspace clones are always directories.
		if !workspaceDir.IsDir() {
			continue
		}

		workspace := workspaceDir.Name()
		repoDir := filepath.Join(pullDir, workspace)

		// Skip directories that are not git repositories (e.g. stray directories
		// left by external processes). Workspace clones always have a .git entry.
		if _, statErr := os.Stat(filepath.Join(repoDir, ".git")); statErr != nil {
			if os.IsNotExist(statErr) || errors.Is(statErr, syscall.ENOTDIR) {
				continue
			}
			return nil, nil, fmt.Errorf("checking git repository in '%s': %w", repoDir, statErr)
		}

		// Any generated plans should be untracked by git since Atlantis created
		// them.
		lsCmd := exec.Command("git", "ls-files", ".", "--others") // nolint: gosec
		lsCmd.Dir = repoDir
		lsOut, err := lsCmd.CombinedOutput()
		if err != nil {
			return nil, nil, fmt.Errorf("running 'git ls-files . --others' in '%s' directory: %s: %w", repoDir, string(lsOut), err)
		}
		for file := range strings.SplitSeq(string(lsOut), "\n") {
			if filepath.Ext(file) == ".tfplan" {
				// Ignore .terragrunt-cache dirs (#487)
				if strings.Contains(file, ".terragrunt-cache/") {
					continue
				}

				projectName, err := runtime.ProjectNameFromPlanfile(workspace, filepath.Base(file))
				if err != nil {
					return nil, nil, err
				}
				plans = append(plans, PendingPlan{
					RepoDir:     repoDir,
					RepoRelDir:  filepath.Dir(file),
					Workspace:   workspace,
					ProjectName: projectName,
				})
				absPaths = append(absPaths, filepath.Join(repoDir, file))
			}
		}
	}
	return plans, absPaths, nil
}

// deletePlans deletes all plans in pullDir.
func (p *DefaultPendingPlanFinder) DeletePlans(pullDir string) error {
	_, absPaths, err := p.findWithAbsPaths(pullDir)
	if err != nil {
		return err
	}
	for _, path := range absPaths {
		if err := utils.RemoveIgnoreNonExistent(path); err != nil {
			return fmt.Errorf("delete plan at %s: %w", path, err)
		}
	}
	return nil
}
