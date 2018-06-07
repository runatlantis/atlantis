package yaml

import (
	"fmt"

	"github.com/flynn-archive/go-shlex"
	"github.com/pkg/errors"
)

type RepoConfig struct {
	Version   int                 `yaml:"version"`
	Projects  []Project           `yaml:"projects"`
	Workflows map[string]Workflow `yaml:"workflows"`
}

type Project struct {
	Dir               string    `yaml:"dir"`
	Workspace         string    `yaml:"workspace"`
	Workflow          string    `yaml:"workflow"`
	TerraformVersion  string    `yaml:"terraform_version"`
	AutoPlan          *AutoPlan `yaml:"auto_plan,omitempty"`
	ApplyRequirements []string  `yaml:"apply_requirements"`
}

func (p *Project) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Use a type alias so unmarshal doesn't get into an infinite loop.
	type alias Project
	// Set up defaults.
	defaults := alias{
		Workspace: defaultWorkspace,
		AutoPlan: &AutoPlan{
			Enabled:      true,
			WhenModified: []string{"**/*.tf"},
		},
	}
	if err := unmarshal(&defaults); err != nil {
		return err
	}
	*p = Project(defaults)
	return nil
}

type AutoPlan struct {
	WhenModified []string `yaml:"when_modified"`
	Enabled      bool     `yaml:"enabled"` // defaults to true
}

func (a *AutoPlan) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Use a type alias so unmarshal doesn't get into an infinite loop.
	type alias AutoPlan
	// Set up defaults.
	defaults := alias{
		// If not specified, we assume it's enabled.
		Enabled: true,
	}
	if err := unmarshal(&defaults); err != nil {
		return err
	}
	*a = AutoPlan(defaults)
	return nil
}

type Workflow struct {
	Apply *Stage `yaml:"apply"` // defaults to regular apply steps
	Plan  *Stage `yaml:"plan"`  // defaults to regular plan steps
}

func (p *Workflow) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Check if they forgot to set the "steps" key.
	type MissingSteps struct {
		Apply []interface{}
		Plan  []interface{}
	}
	var missingSteps MissingSteps
	if err := unmarshal(&missingSteps); err == nil {
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

type Stage struct {
	Steps []StepConfig `yaml:"steps"` // can either be a built in step like 'plan' or a custom step like 'run: echo hi'
}

type StepConfig struct {
	StepType  string
	ExtraArgs []string
	// Run will be set if the StepType is "run". This is for custom commands.
	// Ex. if the key is `run: echo hi` then Run will be "echo hi".
	Run []string
}

func (s *StepConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// First try to unmarshal as a single string, ex.
	// steps:
	// - init
	// - plan
	var singleString string
	err := unmarshal(&singleString)
	if err == nil {
		if singleString != "init" && singleString != "plan" && singleString != "apply" {
			return fmt.Errorf("unsupported step type: %q", singleString)
		}
		s.StepType = singleString
		return nil
	}

	// Next, try to unmarshal as a built-in command with extra_args set, ex.
	// steps:
	// - init:
	///    extra_args: ["arg1"]
	//
	// We need to create a struct for each step so go-yaml knows to call into
	// our routine based on the key (ex. init, plan, etc).
	// We use a map[string]interface{} as the value so we can manually
	// validate key names and return better errors. This is instead of:
	//    Init struct{
	//    	ExtraArgs []string `yaml:"extra_args"`
	//    } `yaml:"init"`

	validateBuiltIn := func(stepType string, args map[string]interface{}) error {
		s.StepType = stepType
		for k, v := range args {
			if k != "extra_args" {
				return fmt.Errorf("unsupported key %q for step %s – the only supported key is extra_args", k, stepType)
			}

			// parse as []string
			val, ok := v.([]interface{})
			if !ok {
				return fmt.Errorf("expected array of strings as value of extra_args, not %q", v)
			}
			var finalVals []string
			for _, i := range val {
				finalVals = append(finalVals, fmt.Sprintf("%s", i))
			}
			s.ExtraArgs = finalVals
		}
		return nil
	}
	var initStep struct {
		Init map[string]interface{} `yaml:"init"`
	}
	if err = unmarshal(&initStep); err == nil {
		return validateBuiltIn("init", initStep.Init)
	}

	var planStep struct {
		Plan map[string]interface{} `yaml:"plan"`
	}
	if err = unmarshal(&planStep); err == nil {
		return validateBuiltIn("plan", planStep.Plan)
	}

	var applyStep struct {
		Apply map[string]interface{} `yaml:"apply"`
	}
	if err = unmarshal(&applyStep); err == nil {
		return validateBuiltIn("apply", applyStep.Apply)
	}

	// Try to unmarshal as a custom run step, ex.
	// steps:
	// - run: my command
	var runStep struct {
		Run string `yaml:"run"`
	}
	if err = unmarshal(&runStep); err == nil {
		s.StepType = "run"
		parts, err := shlex.Split(runStep.Run)
		if err != nil {
			return err
		}
		s.Run = parts
	}

	return err
}
