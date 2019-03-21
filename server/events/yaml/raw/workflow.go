package raw

import (
	"github.com/go-ozzo/ozzo-validation"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
)

type Workflow struct {
	Apply *Stage `yaml:"apply,omitempty"`
	Plan  *Stage `yaml:"plan,omitempty"`
}

func (w Workflow) Validate() error {
	return validation.ValidateStruct(&w,
		validation.Field(&w.Apply),
		validation.Field(&w.Plan),
	)
}

func (w Workflow) ToValid(name string) valid.Workflow {
	v := valid.Workflow{
		Name: name,
	}
	if w.Apply == nil {
		v.Apply = valid.DefaultApplyStage
	} else {
		v.Apply = w.Apply.ToValid()
	}
	if w.Plan == nil {
		v.Plan = valid.DefaultPlanStage
	} else {
		v.Plan = w.Plan.ToValid()
	}
	return v
}
