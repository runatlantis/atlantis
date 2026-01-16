// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package raw

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	validation "github.com/go-ozzo/ozzo-validation"
	version "github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/core/config/valid"
)

const (
	DefaultWorkspace      = "default"
	ApprovedRequirement   = "approved"
	MergeableRequirement  = "mergeable"
	UnDivergedRequirement = "undiverged"
)

type Project struct {
	Name                      *string    `yaml:"name,omitempty"`
	Branch                    *string    `yaml:"branch,omitempty"`
	Dir                       *string    `yaml:"dir,omitempty"`
	Workspace                 *string    `yaml:"workspace,omitempty"`
	Workflow                  *string    `yaml:"workflow,omitempty"`
	TerraformDistribution     *string    `yaml:"terraform_distribution,omitempty"`
	TerraformVersion          *string    `yaml:"terraform_version,omitempty"`
	Autoplan                  *Autoplan  `yaml:"autoplan,omitempty"`
	PlanRequirements          []string   `yaml:"plan_requirements,omitempty"`
	ApplyRequirements         []string   `yaml:"apply_requirements,omitempty"`
	ImportRequirements        []string   `yaml:"import_requirements,omitempty"`
	DependsOn                 []string   `yaml:"depends_on,omitempty"`
	DeleteSourceBranchOnMerge *bool      `yaml:"delete_source_branch_on_merge,omitempty"`
	RepoLocking               *bool      `yaml:"repo_locking,omitempty"`
	RepoLocks                 *RepoLocks `yaml:"repo_locks,omitempty"`
	ExecutionOrderGroup       *int       `yaml:"execution_order_group,omitempty"`
	PolicyCheck               *bool      `yaml:"policy_check,omitempty"`
	CustomPolicyCheck         *bool      `yaml:"custom_policy_check,omitempty"`
	SilencePRComments         []string   `yaml:"silence_pr_comments,omitempty"`
}

func (p Project) Validate() error {
	validDir := func(value any) error {
		dir := *value.(*string)
		if strings.Contains(dir, "..") {
			return errors.New("cannot contain '..'")
		}
		// If the dir contains glob pattern characters, validate the pattern
		if ContainsGlobPattern(dir) {
			if err := ValidateGlobPattern(dir); err != nil {
				return err
			}
		}
		return nil
	}

	validName := func(value any) error {
		strPtr := value.(*string)
		if strPtr == nil {
			return nil
		}
		if *strPtr == "" {
			return errors.New("if set cannot be empty")
		}
		if !validProjectName(*strPtr) {
			return fmt.Errorf("%q is not allowed: must contain only URL safe characters", *strPtr)
		}
		return nil
	}

	branchValid := func(value any) error {
		strPtr := value.(*string)
		if strPtr == nil {
			return nil
		}
		branch := *strPtr
		if !strings.HasPrefix(branch, "/") || !strings.HasSuffix(branch, "/") {
			return errors.New("regex must begin and end with a slash '/'")
		}
		withoutSlashes := branch[1 : len(branch)-1]
		_, err := regexp.Compile(withoutSlashes)
		if err != nil {
			return fmt.Errorf("parsing: %s: %w", branch, err)
		}
		return nil
	}

	DependsOn := func(value any) error {
		return nil
	}

	// Validate that name doesn't contain glob patterns - glob expansion only works for 'dir'
	if p.Name != nil && ContainsGlobPattern(*p.Name) {
		return errors.New("name: cannot contain glob pattern characters ('*', '?', '['); glob expansion is only supported in the 'dir' field")
	}

	// Cross-field validation: name cannot be used with glob patterns in dir
	// because glob patterns expand to multiple projects which can't share the same name
	if p.Name != nil && p.Dir != nil && ContainsGlobPattern(*p.Dir) {
		return errors.New("name: cannot be used with glob patterns in 'dir'; glob patterns expand to multiple projects which cannot share the same name")
	}

	return validation.ValidateStruct(&p,
		validation.Field(&p.Dir, validation.Required, validation.By(validDir)),
		validation.Field(&p.PlanRequirements, validation.By(validPlanReq)),
		validation.Field(&p.ApplyRequirements, validation.By(validApplyReq)),
		validation.Field(&p.ImportRequirements, validation.By(validImportReq)),
		validation.Field(&p.TerraformDistribution, validation.By(validDistribution)),
		validation.Field(&p.TerraformVersion, validation.By(VersionValidator)),
		validation.Field(&p.DependsOn, validation.By(DependsOn)),
		validation.Field(&p.Name, validation.By(validName)),
		validation.Field(&p.Branch, validation.By(branchValid)),
	)
}

func (p Project) ToValid() valid.Project {
	var v valid.Project
	// Prepend ./ and then run .Clean() so we're guaranteed to have a relative
	// directory. This is necessary because we use this dir without sanitation
	// in DefaultProjectFinder.
	cleanedDir := filepath.Clean("./" + *p.Dir)
	v.Dir = cleanedDir

	if p.Branch != nil {
		branch := *p.Branch
		withoutSlashes := branch[1 : len(branch)-1]
		// Safe to use MustCompile because we test it in Validate().
		v.BranchRegex = regexp.MustCompile(withoutSlashes)
	}

	if p.Workspace == nil || *p.Workspace == "" {
		v.Workspace = DefaultWorkspace
	} else {
		v.Workspace = *p.Workspace
	}

	v.WorkflowName = p.Workflow
	if p.TerraformVersion != nil {
		v.TerraformVersion, _ = version.NewVersion(*p.TerraformVersion)
	}
	if p.TerraformDistribution != nil {
		v.TerraformDistribution = p.TerraformDistribution
	}
	if p.Autoplan == nil {
		v.Autoplan = DefaultAutoPlan()
	} else {
		v.Autoplan = p.Autoplan.ToValid()
	}

	// There are no default apply/import requirements.
	v.PlanRequirements = p.PlanRequirements
	v.ApplyRequirements = p.ApplyRequirements
	v.ImportRequirements = p.ImportRequirements

	v.Name = p.Name

	v.DependsOn = p.DependsOn

	if p.DeleteSourceBranchOnMerge != nil {
		v.DeleteSourceBranchOnMerge = p.DeleteSourceBranchOnMerge
	}

	if p.RepoLocking != nil {
		v.RepoLocking = p.RepoLocking
	}

	if p.RepoLocks != nil {
		v.RepoLocks = p.RepoLocks.ToValid()
	}

	if p.ExecutionOrderGroup != nil {
		v.ExecutionOrderGroup = *p.ExecutionOrderGroup
	}

	if p.PolicyCheck != nil {
		v.PolicyCheck = p.PolicyCheck
	}

	if p.CustomPolicyCheck != nil {
		v.CustomPolicyCheck = p.CustomPolicyCheck
	}

	if p.SilencePRComments != nil {
		v.SilencePRComments = p.SilencePRComments
	}

	return v
}

// validProjectName returns true if the project name is valid.
// Since the name might be used in URLs and definitely in files we don't
// support any characters that must be url escaped *except* for '/' because
// users like to name their projects to match the directory it's in.
func validProjectName(name string) bool {
	nameWithoutSlashes := strings.ReplaceAll(name, "/", "-")
	return nameWithoutSlashes == url.QueryEscape(nameWithoutSlashes)
}

func validPlanReq(value any) error {
	reqs := value.([]string)
	for _, r := range reqs {
		if r != ApprovedRequirement && r != MergeableRequirement && r != UnDivergedRequirement {
			return fmt.Errorf("%q is not a valid plan_requirement, only %q, %q and %q are supported", r, ApprovedRequirement, MergeableRequirement, UnDivergedRequirement)
		}
	}
	return nil
}

func validApplyReq(value any) error {
	reqs := value.([]string)
	for _, r := range reqs {
		if r != ApprovedRequirement && r != MergeableRequirement && r != UnDivergedRequirement {
			return fmt.Errorf("%q is not a valid apply_requirement, only %q, %q and %q are supported", r, ApprovedRequirement, MergeableRequirement, UnDivergedRequirement)
		}
	}
	return nil
}

func validImportReq(value any) error {
	reqs := value.([]string)
	for _, r := range reqs {
		if r != ApprovedRequirement && r != MergeableRequirement && r != UnDivergedRequirement {
			return fmt.Errorf("%q is not a valid import_requirement, only %q, %q and %q are supported", r, ApprovedRequirement, MergeableRequirement, UnDivergedRequirement)
		}
	}
	return nil
}

func validDistribution(value any) error {
	distribution := value.(*string)
	if distribution != nil && *distribution != "terraform" && *distribution != "opentofu" {
		return fmt.Errorf("'%s' is not a valid terraform_distribution, only '%s' and '%s' are supported", *distribution, "terraform", "opentofu")
	}
	return nil
}

// ContainsGlobPattern returns true if the string contains glob pattern characters.
// This is used to detect if a dir field should be treated as a glob pattern
// for expansion into multiple projects.
func ContainsGlobPattern(s string) bool {
	return strings.ContainsAny(s, "*?[")
}

// ValidateGlobPattern validates that a glob pattern is syntactically correct
// using the doublestar library.
func ValidateGlobPattern(pattern string) error {
	if !doublestar.ValidatePattern(pattern) {
		return fmt.Errorf("invalid glob pattern %q", pattern)
	}
	return nil
}
