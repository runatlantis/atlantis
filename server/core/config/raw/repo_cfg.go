package raw

import (
	"errors"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/runatlantis/atlantis/server/core/config/valid"
)

// DefaultAutomerge is the default setting for automerge.
const DefaultAutomerge = false

// DefaultParallelApply is the default setting for parallel apply
const DefaultParallelApply = false

// DefaultParallelPlan is the default setting for parallel plan
const DefaultParallelPlan = false

// DefaultParallelPolicyCheck is the default setting for parallel plan
const DefaultParallelPolicyCheck = false

// DefaultDeleteSourceBranchOnMerge being false is the default setting whether or not to remove a source branch on merge
const DefaultDeleteSourceBranchOnMerge = false

// RepoCfg is the raw schema for repo-level atlantis.yaml config.
type RepoCfg struct {
	Version                   *int                `yaml:"version,omitempty"`
	Projects                  []Project           `yaml:"projects,omitempty"`
	Workflows                 map[string]Workflow `yaml:"workflows,omitempty"`
	PolicySets                PolicySets          `yaml:"policies,omitempty"`
	Automerge                 *bool               `yaml:"automerge,omitempty"`
	ParallelApply             *bool               `yaml:"parallel_apply,omitempty"`
	ParallelPlan              *bool               `yaml:"parallel_plan,omitempty"`
	DeleteSourceBranchOnMerge *bool               `yaml:"delete_source_branch_on_merge,omitempty"`
	AllowedRegexpPrefixes     []string            `yaml:"allowed_regexp_prefixes,omitempty"`
}

func (r RepoCfg) Validate() error {
	equals2 := func(value interface{}) error {
		asIntPtr := value.(*int)
		if asIntPtr == nil {
			return errors.New("is required. If you've just upgraded Atlantis you need to rewrite your atlantis.yaml for version 3. See www.runatlantis.io/docs/upgrading-atlantis-yaml.html")
		}
		if *asIntPtr != 2 && *asIntPtr != 3 {
			return errors.New("only versions 2 and 3 are supported")
		}
		return nil
	}
	return validation.ValidateStruct(&r,
		validation.Field(&r.Version, validation.By(equals2)),
		validation.Field(&r.Projects),
		validation.Field(&r.Workflows),
	)
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

	automerge := DefaultAutomerge
	if r.Automerge != nil {
		automerge = *r.Automerge
	}

	parallelApply := DefaultParallelApply
	if r.ParallelApply != nil {
		parallelApply = *r.ParallelApply
	}

	parallelPlan := DefaultParallelPlan
	if r.ParallelPlan != nil {
		parallelPlan = *r.ParallelPlan
	}

	return valid.RepoCfg{
		Version:                   *r.Version,
		Projects:                  validProjects,
		Workflows:                 validWorkflows,
		Automerge:                 automerge,
		ParallelApply:             parallelApply,
		ParallelPlan:              parallelPlan,
		ParallelPolicyCheck:       parallelPlan,
		DeleteSourceBranchOnMerge: r.DeleteSourceBranchOnMerge,
		AllowedRegexpPrefixes:     r.AllowedRegexpPrefixes,
	}
}
