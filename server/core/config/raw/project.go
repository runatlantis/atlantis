package raw

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation"
	version "github.com/hashicorp/go-version"
	"github.com/pkg/errors"
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
	hasDotDot := func(value interface{}) error {
		if strings.Contains(*value.(*string), "..") {
			return errors.New("cannot contain '..'")
		}
		return nil
	}

	validName := func(value interface{}) error {
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

	branchValid := func(value interface{}) error {
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
		return errors.Wrapf(err, "parsing: %s", branch)
	}

	DependsOn := func(value interface{}) error {
		return nil
	}

	return validation.ValidateStruct(&p,
		validation.Field(&p.Dir, validation.Required, validation.By(hasDotDot)),
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
	nameWithoutSlashes := strings.Replace(name, "/", "-", -1)
	return nameWithoutSlashes == url.QueryEscape(nameWithoutSlashes)
}

func validPlanReq(value interface{}) error {
	reqs := value.([]string)
	for _, r := range reqs {
		if r != ApprovedRequirement && r != MergeableRequirement && r != UnDivergedRequirement {
			return fmt.Errorf("%q is not a valid plan_requirement, only %q, %q and %q are supported", r, ApprovedRequirement, MergeableRequirement, UnDivergedRequirement)
		}
	}
	return nil
}

func validApplyReq(value interface{}) error {
	reqs := value.([]string)
	for _, r := range reqs {
		if r != ApprovedRequirement && r != MergeableRequirement && r != UnDivergedRequirement {
			return fmt.Errorf("%q is not a valid apply_requirement, only %q, %q and %q are supported", r, ApprovedRequirement, MergeableRequirement, UnDivergedRequirement)
		}
	}
	return nil
}

func validImportReq(value interface{}) error {
	reqs := value.([]string)
	for _, r := range reqs {
		if r != ApprovedRequirement && r != MergeableRequirement && r != UnDivergedRequirement {
			return fmt.Errorf("%q is not a valid import_requirement, only %q, %q and %q are supported", r, ApprovedRequirement, MergeableRequirement, UnDivergedRequirement)
		}
	}
	return nil
}

func validDistribution(value interface{}) error {
	distribution := value.(*string)
	if distribution != nil && *distribution != "terraform" && *distribution != "opentofu" {
		return fmt.Errorf("%q is not a valid terraform_distribution, only %q and %q are supported", *distribution, "terraform", "opentofu")
	}
	return nil
}
