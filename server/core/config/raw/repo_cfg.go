package raw

import (
	"errors"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/runatlantis/atlantis/server/core/config/valid"
)

// DefaultEmojiReaction is the default emoji reaction for repos
const DefaultEmojiReaction = ""

// DefaultAbortOnExcecutionOrderFail being false is the default setting for abort on execution group failiures
const DefaultAbortOnExcecutionOrderFail = false

// RepoCfg is the raw schema for repo-level atlantis.yaml config.
type RepoCfg struct {
	Version                    *int                `yaml:"version,omitempty"`
	Projects                   []Project           `yaml:"projects,omitempty"`
	Workflows                  map[string]Workflow `yaml:"workflows,omitempty"`
	PolicySets                 PolicySets          `yaml:"policies,omitempty"`
	Automerge                  *bool               `yaml:"automerge,omitempty"`
	ParallelApply              *bool               `yaml:"parallel_apply,omitempty"`
	ParallelPlan               *bool               `yaml:"parallel_plan,omitempty"`
	DeleteSourceBranchOnMerge  *bool               `yaml:"delete_source_branch_on_merge,omitempty"`
	EmojiReaction              *string             `yaml:"emoji_reaction,omitempty"`
	AllowedRegexpPrefixes      []string            `yaml:"allowed_regexp_prefixes,omitempty"`
	AbortOnExcecutionOrderFail *bool               `yaml:"abort_on_execution_order_fail,omitempty"`
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

	automerge := r.Automerge
	parallelApply := r.ParallelApply
	parallelPlan := r.ParallelPlan

	emojiReaction := DefaultEmojiReaction
	if r.EmojiReaction != nil {
		emojiReaction = *r.EmojiReaction
	}

	abortOnExcecutionOrderFail := DefaultAbortOnExcecutionOrderFail
	if r.AbortOnExcecutionOrderFail != nil {
		abortOnExcecutionOrderFail = *r.AbortOnExcecutionOrderFail
	}

	return valid.RepoCfg{
		Version:                    *r.Version,
		Projects:                   validProjects,
		Workflows:                  validWorkflows,
		Automerge:                  automerge,
		ParallelApply:              parallelApply,
		ParallelPlan:               parallelPlan,
		ParallelPolicyCheck:        parallelPlan,
		DeleteSourceBranchOnMerge:  r.DeleteSourceBranchOnMerge,
		AllowedRegexpPrefixes:      r.AllowedRegexpPrefixes,
		EmojiReaction:              emojiReaction,
		AbortOnExcecutionOrderFail: abortOnExcecutionOrderFail,
	}
}
