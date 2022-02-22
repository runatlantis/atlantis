package raw

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/runatlantis/atlantis/server/core/config/valid"
)

type DeploymentWorkflows map[string]DeploymentWorkflow

func (w DeploymentWorkflows) ToValid(defaultCfg valid.GlobalCfg) map[string]valid.Workflow {
	validWorkflows := make(map[string]valid.Workflow)
	for k, v := range w {
		validWorkflows[k] = v.ToValid(k)
	}

	// Merge in defaults without overriding.
	for k, v := range defaultCfg.DeploymentWorkflows {
		if _, ok := validWorkflows[k]; !ok {
			validWorkflows[k] = v
		}
	}

	return validWorkflows
}

type DeploymentWorkflow struct {
	Apply *Stage `yaml:"apply,omitempty" json:"apply,omitempty"`
	Plan  *Stage `yaml:"plan,omitempty" json:"plan,omitempty"`
}

func (w DeploymentWorkflow) Validate() error {
	return validation.ValidateStruct(&w,
		validation.Field(&w.Apply),
		validation.Field(&w.Plan),
	)
}

func (w DeploymentWorkflow) toValidStage(stage *Stage, defaultStage valid.Stage) valid.Stage {
	if stage == nil || stage.Steps == nil {
		return defaultStage
	}

	return stage.ToValid()
}

func (w DeploymentWorkflow) ToValid(name string) valid.Workflow {
	v := valid.Workflow{
		Name: name,
	}

	v.Apply = w.toValidStage(w.Apply, valid.DefaultApplyStage)
	v.Plan = w.toValidStage(w.Plan, valid.DefaultPlanStage)

	return v
}
