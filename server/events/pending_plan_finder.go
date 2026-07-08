// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/runatlantis/atlantis/server/core/runtime"
	"github.com/runatlantis/atlantis/server/utils"
)

//go:generate go tool pegomock generate --package mocks -o mocks/mock_pending_plan_finder.go PendingPlanFinder

type PendingPlanFinder interface {
	Find(pullDir string) ([]PendingPlan, error)
	DeletePlans(pullDir string) error
}

// DefaultPendingPlanFinder finds unapplied plans.
type DefaultPendingPlanFinder struct {
	DataDir           string
	LocalPlanStoreDir string
}

// PendingPlan is a plan that has not been applied.
type PendingPlan struct {
	// RepoDir is the absolute path to the root of the repo that holds this
	// plan's workspace clone.
	RepoDir string
	// PlanDir is the absolute path to the root of the workspace plan store.
	// If empty, RepoDir is used for backward compatibility.
	PlanDir string
	// RepoRelDir is the relative path from the repo to the project that
	// the plan is for.
	RepoRelDir string
	// Workspace is the workspace this plan should execute in.
	Workspace   string
	ProjectName string
}

func (p PendingPlan) planRepoDir() string {
	if p.PlanDir != "" {
		return p.PlanDir
	}
	return p.RepoDir
}

// Find finds all pending plans in pullDir. pullDir should be the working
// directory where Atlantis will operate on this pull request. It's one level
// up from where Atlantis clones the repo for each workspace.
func (p *DefaultPendingPlanFinder) Find(pullDir string) ([]PendingPlan, error) {
	plans, _, err := p.findWithAbsPaths(pullDir)
	return plans, err
}

func (p *DefaultPendingPlanFinder) findWithAbsPaths(pullDir string) ([]PendingPlan, []string, error) {
	planPullDir, separatePlanDir, err := p.planPullDir(pullDir)
	if err != nil {
		return nil, nil, err
	}
	if separatePlanDir {
		return p.findInPlanStore(pullDir, planPullDir)
	}
	return p.findInGitWorkspaces(pullDir)
}

func (p *DefaultPendingPlanFinder) planPullDir(pullDir string) (string, bool, error) {
	if p.DataDir == "" || p.LocalPlanStoreDir == "" {
		return pullDir, false, nil
	}

	cloneRoot := filepath.Clean(filepath.Join(p.DataDir, workingDirPrefix))
	cleanPullDir := filepath.Clean(pullDir)
	relPullDir, err := filepath.Rel(cloneRoot, cleanPullDir)
	if err != nil {
		return "", false, fmt.Errorf("checking pull directory %q: %w", pullDir, err)
	}
	if relPullDir == ".." || filepath.IsAbs(relPullDir) || strings.HasPrefix(relPullDir, ".."+string(filepath.Separator)) {
		return "", false, fmt.Errorf("pull directory %q escapes clone root %q", pullDir, cloneRoot)
	}

	planPullDir := filepath.Join(p.LocalPlanStoreDir, workingDirPrefix, relPullDir)
	return planPullDir, filepath.Clean(planPullDir) != cleanPullDir, nil
}

func (p *DefaultPendingPlanFinder) findInGitWorkspaces(pullDir string) ([]PendingPlan, []string, error) {
	workspaceDirs, err := os.ReadDir(pullDir)
	if err != nil {
		return nil, nil, err
	}
	absPullDir, err := filepath.Abs(pullDir)
	if err != nil {
		return nil, nil, fmt.Errorf("getting absolute pull directory: %w", err)
	}

	var plans []PendingPlan
	var absPaths []string
	for _, workspaceDir := range workspaceDirs {
		// Skip non-directory entries (files, symlinks); workspace clones are always directories.
		if !workspaceDir.IsDir() {
			continue
		}

		workspace := workspaceDir.Name()
		repoDir, err := workspaceRepoDir(absPullDir, workspace)
		if err != nil {
			return nil, nil, err
		}

		// Skip directories that are not workspace clone roots (e.g. stray
		// directories left by external processes).
		workspaceIsGitRoot, err := isGitWorkTreeRoot(repoDir)
		if err != nil {
			return nil, nil, err
		}
		if !workspaceIsGitRoot {
			continue
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

func (p *DefaultPendingPlanFinder) findInPlanStore(clonePullDir string, planPullDir string) ([]PendingPlan, []string, error) {
	workspaceDirs, err := os.ReadDir(planPullDir)
	if os.IsNotExist(err) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}

	absClonePullDir, err := filepath.Abs(clonePullDir)
	if err != nil {
		return nil, nil, fmt.Errorf("getting absolute clone pull directory: %w", err)
	}
	absPlanPullDir, err := filepath.Abs(planPullDir)
	if err != nil {
		return nil, nil, fmt.Errorf("getting absolute plan pull directory: %w", err)
	}

	var plans []PendingPlan
	var absPaths []string
	for _, workspaceDir := range workspaceDirs {
		if !workspaceDir.IsDir() {
			continue
		}

		workspace := workspaceDir.Name()
		cloneRepoDir, err := workspaceRepoDir(absClonePullDir, workspace)
		if err != nil {
			return nil, nil, err
		}
		if _, err := os.Stat(cloneRepoDir); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, nil, err
		}
		workspaceIsGitRoot, err := isGitWorkTreeRoot(cloneRepoDir)
		if err != nil {
			return nil, nil, err
		}
		if !workspaceIsGitRoot {
			continue
		}
		planRepoDir, err := workspaceRepoDir(absPlanPullDir, workspace)
		if err != nil {
			return nil, nil, err
		}

		if err := filepath.WalkDir(planRepoDir, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				if entry.Name() == ".terragrunt-cache" {
					return filepath.SkipDir
				}
				return nil
			}
			if filepath.Ext(path) != ".tfplan" {
				return nil
			}

			relFile, err := filepath.Rel(planRepoDir, path)
			if err != nil {
				return fmt.Errorf("checking plan path %q: %w", path, err)
			}
			projectName, err := runtime.ProjectNameFromPlanfile(workspace, filepath.Base(relFile))
			if err != nil {
				return err
			}
			plans = append(plans, PendingPlan{
				RepoDir:     cloneRepoDir,
				PlanDir:     planRepoDir,
				RepoRelDir:  filepath.Dir(relFile),
				Workspace:   workspace,
				ProjectName: projectName,
			})
			absPaths = append(absPaths, path)
			return nil
		}); err != nil {
			return nil, nil, err
		}
	}
	return plans, absPaths, nil
}

func workspaceRepoDir(absPullDir, workspace string) (string, error) {
	repoDir := filepath.Clean(filepath.Join(absPullDir, workspace))
	rel, err := filepath.Rel(absPullDir, repoDir)
	if err != nil {
		return "", fmt.Errorf("checking workspace directory %q: %w", workspace, err)
	}
	if rel == "." || rel == ".." || filepath.IsAbs(rel) || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("workspace directory %q escapes pull directory %q", workspace, absPullDir)
	}
	return repoDir, nil
}

func isGitWorkTreeRoot(repoDir string) (bool, error) {
	showPrefixCmd := exec.Command("git", "rev-parse", "--is-inside-work-tree", "--show-prefix") // nolint: gosec
	showPrefixCmd.Dir = repoDir
	showPrefixOut, err := showPrefixCmd.CombinedOutput()
	if err != nil {
		output := string(showPrefixOut)
		if strings.Contains(output, "not a git repository") || strings.Contains(output, "must be run in a work tree") {
			return false, nil
		}
		return false, fmt.Errorf("checking git repository in '%s' directory: %s: %w", repoDir, output, err)
	}
	lines := strings.Split(string(showPrefixOut), "\n")
	return len(lines) >= 2 && lines[0] == "true" && lines[1] == "", nil
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
