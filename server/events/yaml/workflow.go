package yaml

import "errors"

type Workflow struct {
	Apply *Stage `yaml:"apply"` // defaults to regular apply steps
	Plan  *Stage `yaml:"plan"`  // defaults to regular plan steps
}

func (p *Workflow) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Check if they forgot to set the "steps" key and just started listing
	// steps, ex.
	//  plan:
	//    - init
	//    - plan
	type MissingSteps struct {
		Apply []interface{}
		Plan  []interface{}
	}
	var missingSteps MissingSteps
	// This will pass if they've just set the key to null, which we don't want
	// since in that case we use the defaults to we also check if the len > 0.
	if err := unmarshal(&missingSteps); err == nil && (len(missingSteps.Apply) > 0 || len(missingSteps.Plan) > 0) {
		return errors.New("missing \"steps\" key")
	}

	// Use a type alias so unmarshal doesn't get into an infinite loop.
	type alias Workflow
	var tmp alias
	if err := unmarshal(&tmp); err != nil {
		return err
	}
	*p = Workflow(tmp)

	// If plan or apply keys aren't specified we use the default workflow.
	if p.Apply == nil {
		p.Apply = &Stage{
			[]StepConfig{
				{
					StepType: "apply",
				},
			},
		}
	}

	if p.Plan == nil {
		p.Plan = &Stage{
			[]StepConfig{
				{
					StepType: "init",
				},
				{
					StepType: "plan",
				},
			},
		}
	}

	return nil
}
