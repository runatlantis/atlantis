// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package raw

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/runatlantis/atlantis/server/core/config/valid"
)

// DefaultEmojiReaction is the default emoji reaction for repos
const DefaultEmojiReaction = ""

// DefaultAbortOnExecutionOrderFail being false is the default setting for abort on execution group failures
const DefaultAbortOnExecutionOrderFail = false

// RepoCfg is the raw schema for repo-level atlantis.yaml config.
type RepoCfg struct {
	Version                   *int                `yaml:"version,omitempty"`
	Projects                  []Project           `yaml:"projects,omitempty"`
	Workflows                 map[string]Workflow `yaml:"workflows,omitempty"`
	PolicySets                PolicySets          `yaml:"policies,omitempty"`
	AutoDiscover              *AutoDiscover       `yaml:"autodiscover,omitempty"`
	Automerge                 *bool               `yaml:"automerge,omitempty"`
	ParallelApply             *bool               `yaml:"parallel_apply,omitempty"`
	ParallelPlan              *bool               `yaml:"parallel_plan,omitempty"`
	DeleteSourceBranchOnMerge *bool               `yaml:"delete_source_branch_on_merge,omitempty"`
	EmojiReaction             *string             `yaml:"emoji_reaction,omitempty"`
	AllowedRegexpPrefixes     []string            `yaml:"allowed_regexp_prefixes,omitempty"`
	AbortOnExecutionOrderFail *bool               `yaml:"abort_on_execution_order_fail,omitempty"`
	RepoLocks                 *RepoLocks          `yaml:"repo_locks,omitempty"`
	SilencePRComments         []string            `yaml:"silence_pr_comments,omitempty"`
}

func (r RepoCfg) Validate() error {
	equals2 := func(value any) error {
		asIntPtr := value.(*int)
		if asIntPtr == nil {
			return errors.New("is required. If you've just upgraded Atlantis you need to rewrite your atlantis.yaml for version 3. See www.runatlantis.io/docs/upgrading-atlantis-yaml.html")
		}
		if *asIntPtr != 2 && *asIntPtr != 3 {
			return errors.New("only versions 2 and 3 are supported")
		}
		return nil
	}
	if err := validation.ValidateStruct(&r,
		validation.Field(&r.Version, validation.By(equals2)),
		validation.Field(&r.Projects),
		validation.Field(&r.Workflows),
	); err != nil {
		return err
	}
	// Validated separately from the struct fields because it needs to see all
	// the projects at once, not one at a time.
	return r.validateDependsOn()
}

// validateDependsOn checks the depends_on graph. Dependencies are matched
// against project names when an apply runs, and an unmatched name is not an
// error at apply time unless the server sets fail-on-missing-dependencies, so a
// typo would otherwise silently drop the dependency instead of enforcing it. A
// project depending on itself, or a cycle, can never satisfy its dependencies
// and would block those applies forever, so both are rejected too.
//
// This runs before projects are filtered by the pull request's base branch, so
// depending on a project that only exists on another branch is still valid.
func (r RepoCfg) validateDependsOn() error {
	names := make(map[string]bool, len(r.Projects))
	for _, p := range r.Projects {
		if p.Name != nil && *p.Name != "" {
			names[*p.Name] = true
		}
	}

	dependsOn := make(map[string][]string, len(r.Projects))
	for _, p := range r.Projects {
		for _, dep := range p.DependsOn {
			if !names[dep] {
				return fmt.Errorf("depends_on: %s depends on %q which is not a project name defined in this repo config", describeProject(p), dep)
			}
			if p.Name != nil && *p.Name == dep {
				return fmt.Errorf("depends_on: project %q cannot depend on itself", dep)
			}
		}
		if p.Name != nil && *p.Name != "" {
			dependsOn[*p.Name] = append(dependsOn[*p.Name], p.DependsOn...)
		}
	}

	if cycle := findDependsOnCycle(dependsOn); len(cycle) > 0 {
		return fmt.Errorf("depends_on: projects cannot depend on each other in a cycle: %s", strings.Join(cycle, " -> "))
	}
	return nil
}

// findDependsOnCycle returns the projects forming a dependency cycle, starting
// and ending at the same project, or nil if the graph is acyclic.
func findDependsOnCycle(dependsOn map[string][]string) []string {
	const (
		visiting = 1
		done     = 2
	)
	state := make(map[string]int, len(dependsOn))
	var path []string

	var walk func(name string) []string
	walk = func(name string) []string {
		switch state[name] {
		case done:
			return nil
		case visiting:
			// Trim the path to where the cycle starts so the message only
			// contains the projects in it.
			for i, p := range path {
				if p == name {
					return append(append([]string{}, path[i:]...), name)
				}
			}
			return []string{name, name}
		}
		state[name] = visiting
		path = append(path, name)
		for _, dep := range dependsOn[name] {
			if cycle := walk(dep); cycle != nil {
				return cycle
			}
		}
		path = path[:len(path)-1]
		state[name] = done
		return nil
	}

	// Sorted so the reported cycle is deterministic across runs.
	names := make([]string, 0, len(dependsOn))
	for name := range dependsOn {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if cycle := walk(name); cycle != nil {
			return cycle
		}
	}
	return nil
}

// describeProject identifies a project in an error message, falling back to its
// dir when it has no name.
func describeProject(p Project) string {
	if p.Name != nil && *p.Name != "" {
		return fmt.Sprintf("project %q", *p.Name)
	}
	if p.Dir != nil {
		return fmt.Sprintf("the project in dir %q", *p.Dir)
	}
	return "a project"
}

func (r RepoCfg) ToValid() valid.RepoCfg {
	validWorkflows := make(map[string]valid.Workflow)
	for k, v := range r.Workflows {
		validWorkflows[k] = v.ToValid(k)
	}

	var validProjects []valid.Project
	for _, p := range r.Projects {
		validProjects = append(validProjects, p.ToValid())
	}

	automerge := r.Automerge
	parallelApply := r.ParallelApply
	parallelPlan := r.ParallelPlan

	emojiReaction := DefaultEmojiReaction
	if r.EmojiReaction != nil {
		emojiReaction = *r.EmojiReaction
	}

	abortOnExecutionOrderFail := DefaultAbortOnExecutionOrderFail
	if r.AbortOnExecutionOrderFail != nil {
		abortOnExecutionOrderFail = *r.AbortOnExecutionOrderFail
	}

	var autoDiscover *valid.AutoDiscover
	if r.AutoDiscover != nil {
		autoDiscover = r.AutoDiscover.ToValid()
	}

	var repoLocks *valid.RepoLocks
	if r.RepoLocks != nil {
		repoLocks = r.RepoLocks.ToValid()
	}
	return valid.RepoCfg{
		Version:                   *r.Version,
		Projects:                  validProjects,
		Workflows:                 validWorkflows,
		AutoDiscover:              autoDiscover,
		Automerge:                 automerge,
		ParallelApply:             parallelApply,
		ParallelPlan:              parallelPlan,
		ParallelPolicyCheck:       parallelPlan,
		DeleteSourceBranchOnMerge: r.DeleteSourceBranchOnMerge,
		AllowedRegexpPrefixes:     r.AllowedRegexpPrefixes,
		EmojiReaction:             emojiReaction,
		AbortOnExecutionOrderFail: abortOnExecutionOrderFail,
		RepoLocks:                 repoLocks,
		SilencePRComments:         r.SilencePRComments,
	}
}
