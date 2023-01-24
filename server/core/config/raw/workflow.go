package raw

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/runatlantis/atlantis/server/core/config/valid"
)

type Workflow struct {
	Apply       *Stage `yaml:"apply,omitempty" json:"apply,omitempty"`
	Plan        *Stage `yaml:"plan,omitempty" json:"plan,omitempty"`
	PolicyCheck *Stage `yaml:"policy_check,omitempty" json:"policy_check,omitempty"`
	Import      *Stage `yaml:"import,omitempty" json:"import,omitempty"`
	StateRm     *Stage `yaml:"state_rm,omitempty" json:"state_rm,omitempty"`
}

func (w Workflow) Validate() error {
	return validation.ValidateStruct(&w,
		validation.Field(&w.Apply),
		validation.Field(&w.Plan),
		validation.Field(&w.PolicyCheck),
		validation.Field(&w.Import),
		validation.Field(&w.StateRm),
	)
}

func (w Workflow) toValidStage(stage *Stage, defaultStage valid.Stage) valid.Stage {
	if stage == nil || stage.Steps == nil {
		return defaultStage
	}

	return stage.ToValid()
}

func (w Workflow) ToValid(name string) valid.Workflow {
	v := valid.Workflow{
		Name: name,
	}

	v.Apply = w.toValidStage(w.Apply, valid.DefaultApplyStage)
	v.Plan = w.toValidStage(w.Plan, valid.DefaultPlanStage)
	v.PolicyCheck = w.toValidStage(w.PolicyCheck, valid.DefaultPolicyCheckStage)
	v.Import = w.toValidStage(w.Import, valid.DefaultImportStage)
	v.StateRm = w.toValidStage(w.StateRm, valid.DefaultStateRmStage)

	return v
}
