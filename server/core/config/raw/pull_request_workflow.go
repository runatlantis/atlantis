package raw

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/runatlantis/atlantis/server/core/config/valid"
)

type PullRequestWorkflows map[string]PullRequestWorkflow

func (w PullRequestWorkflows) ToValid(defaultCfg valid.GlobalCfg) map[string]valid.Workflow {
	validWorkflows := make(map[string]valid.Workflow)
	for k, v := range w {
		validWorkflows[k] = v.ToValid(k)
	}

	// Merge in defaults without overriding.
	for k, v := range defaultCfg.PullRequestWorkflows {
		if _, ok := validWorkflows[k]; !ok {
			validWorkflows[k] = v
		}
	}

	return validWorkflows
}

type PullRequestWorkflow struct {
	Plan        *Stage `yaml:"plan,omitempty" json:"plan,omitempty"`
	PolicyCheck *Stage `yaml:"policy_check,omitempty" json:"policy_check,omitempty"`
}

func (w PullRequestWorkflow) Validate() error {
	return validation.ValidateStruct(&w,
		validation.Field(&w.Plan),
		validation.Field(&w.PolicyCheck),
	)
}

func (w PullRequestWorkflow) toValidStage(stage *Stage, defaultStage valid.Stage) valid.Stage {
	if stage == nil || stage.Steps == nil {
		return defaultStage
	}

	validStage := stage.ToValid()

	// HACK: Extra args should be used for customer args passed in.
	// We should have this logic built into the client and relevant context
	// passed in to allow this.
	// Doing this here minimizes changes right now and ensures that even custom extra args
	// will still have this default applied.
	for _, s := range validStage.Steps {
		if s.StepName == "plan" {
			s.ExtraArgs = append(s.ExtraArgs, "-lock=false")
		}
	}

	return validStage
}

func (w PullRequestWorkflow) ToValid(name string) valid.Workflow {
	v := valid.Workflow{
		Name: name,
	}

	v.Plan = w.toValidStage(w.Plan, valid.DefaultPlanStage)
	v.PolicyCheck = w.toValidStage(w.PolicyCheck, valid.DefaultPolicyCheckStage)

	return v
}
