package yaml

import (
	"fmt"

	"github.com/flynn-archive/go-shlex"
	"github.com/pkg/errors"
)

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

	// This represents a step with extra_args, ex:
	//   init:
	//     extra_args: [a, b]
	var step map[string]map[string][]string
	if err = unmarshal(&step); err == nil {
		if len(step) != 1 {
			return errors.New("each step can have only one map key, you probably have something like:\nsteps:\n  - key1: val\n    key2: val")
		}

		for k, v := range step {
			if k != "init" && k != "plan" && k != "apply" {
				return fmt.Errorf("unsupported step %q", k)
			}

			extraArgs, ok := v["extra_args"]
			if !ok {
				return errors.New("the only supported key for a step is 'extra_args'")
			}

			s.StepType = k
			s.ExtraArgs = extraArgs
			return nil
		}
	}

	// Try to unmarshal as a custom run step, ex.
	// steps:
	// - run: my command
	var runStep map[string]string
	if err = unmarshal(&runStep); err == nil {
		if len(runStep) != 1 {
			return errors.New("each step can have only one map key, you probably have something like:\nsteps:\n  - key1: val\n    key2: val")
		}

		for k, v := range runStep {
			if k != "run" {
				return fmt.Errorf("unsupported step %q", k)
			}

			s.StepType = "run"
			parts, err := shlex.Split(v)
			if err != nil {
				return err
			}
			s.Run = parts
		}
	}

	return err
}
