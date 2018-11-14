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

func (w Workflow) ToValid() valid.Workflow {
	var v valid.Workflow
	if w.Apply != nil {
		apply := w.Apply.ToValid()
		v.Apply = &apply
	}
	if w.Plan != nil {
		plan := w.Plan.ToValid()
		v.Plan = &plan
	}
	return v
}
