package raw

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
)

const (
	ExtraArgsKey  = "extra_args"
	RunStepName   = "run"
	PlanStepName  = "plan"
	ApplyStepName = "apply"
	InitStepName  = "init"
)

// Step represents a single action/command to perform. In YAML, it can be set as
// 1. A single string for a built-in command:
//    - init
//    - plan
// 2. A map for a built-in command and extra_args:
//    - plan:
//        extra_args: [-var-file=staging.tfvars]
// 3. A map for a custom run command:
//    - run: my custom command
// Here we parse step in the most generic fashion possible. See fields for more
// details.
type Step struct {
	// Key will be set in case #1 and #3 above to the key. In case #2, there
	// could be multiple keys (since the element is a map) so we don't set Key.
	Key *string
	// Map will be set in case #2 above.
	Map map[string]map[string][]string
	// StringVal will be set in case #3 above.
	StringVal map[string]string
}

func (s *Step) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return s.unmarshalGeneric(unmarshal)
}

func (s *Step) UnmarshalJSON(data []byte) error {
	return s.unmarshalGeneric(func(i interface{}) error {
		return json.Unmarshal(data, i)
	})
}
func (s Step) Validate() error {
	validStep := func(value interface{}) error {
		str := *value.(*string)
		if str != InitStepName && str != PlanStepName && str != ApplyStepName {
			return fmt.Errorf("%q is not a valid step type, maybe you omitted the 'run' key", str)
		}
		return nil
	}

	extraArgs := func(value interface{}) error {
		elem := value.(map[string]map[string][]string)
		var keys []string
		for k := range elem {
			keys = append(keys, k)
		}
		// Sort so tests can be deterministic.
		sort.Strings(keys)

		if len(keys) > 1 {
			return fmt.Errorf("step element can only contain a single key, found %d: %s",
				len(keys), strings.Join(keys, ","))
		}
		for stepName, args := range elem {
			if stepName != InitStepName && stepName != PlanStepName && stepName != ApplyStepName {
				return fmt.Errorf("%q is not a valid step type", stepName)
			}
			var argKeys []string
			for k := range args {
				argKeys = append(argKeys, k)
			}

			// args should contain a single 'extra_args' key.
			if len(argKeys) > 1 {
				return fmt.Errorf("built-in steps only support a single %s key, found %d: %s",
					ExtraArgsKey, len(argKeys), strings.Join(argKeys, ","))
			}
			for k := range args {
				if k != ExtraArgsKey {
					return fmt.Errorf("built-in steps only support a single %s key, found %q in step %s", ExtraArgsKey, k, stepName)
				}
			}
		}
		return nil
	}

	runStep := func(value interface{}) error {
		elem := value.(map[string]string)
		var keys []string
		for k := range elem {
			keys = append(keys, k)
		}
		// Sort so tests can be deterministic.
		sort.Strings(keys)

		if len(keys) > 1 {
			return fmt.Errorf("step element can only contain a single key, found %d: %s",
				len(keys), strings.Join(keys, ","))
		}
		for stepName := range elem {
			if stepName != RunStepName {
				return fmt.Errorf("%q is not a valid step type", stepName)
			}
		}
		return nil
	}

	if s.Key != nil {
		return validation.Validate(s.Key, validation.By(validStep))
	}
	if len(s.Map) > 0 {
		return validation.Validate(s.Map, validation.By(extraArgs))
	}
	if len(s.StringVal) > 0 {
		return validation.Validate(s.StringVal, validation.By(runStep))
	}
	return errors.New("step element is empty")
}

func (s Step) ToValid() valid.Step {
	// This will trigger in case #1 (see Step docs).
	if s.Key != nil {
		return valid.Step{
			StepName: *s.Key,
		}
	}

	// This will trigger in case #2 (see Step docs).
	if len(s.Map) > 0 {
		// After validation we assume there's only one key and it's a valid
		// step name so we just use the first one.
		for stepName, stepArgs := range s.Map {
			return valid.Step{
				StepName:  stepName,
				ExtraArgs: stepArgs[ExtraArgsKey],
			}
		}
	}

	// This will trigger in case #3 (see Step docs).
	if len(s.StringVal) > 0 {
		// After validation we assume there's only one key and it's a valid
		// step name so we just use the first one.
		for _, v := range s.StringVal {
			return valid.Step{
				StepName:   RunStepName,
				RunCommand: v,
			}
		}
	}

	panic("step was not valid. This is a bug!")
}

// unmarshalGeneric is used by UnmarshalJSON and UnmarshalYAML to unmarshal
// a step into one of its three forms. We need to implement a custom unmarshal
// function because steps can either be:
// 1. a built-in step: " - init"
// 2. a built-in step with extra_args: " - init: {extra_args: [arg1] }"
// 3. a custom run step: " - run: my custom command"
// It takes a parameter unmarshal that is a function that tries to unmarshal
// the current element into a given object.
func (s *Step) unmarshalGeneric(unmarshal func(interface{}) error) error {

	// First try to unmarshal as a single string, ex.
	// steps:
	// - init
	// - plan
	// We validate if it's a legal string later.
	var singleString string
	err := unmarshal(&singleString)
	if err == nil {
		s.Key = &singleString
		return nil
	}

	// This represents a step with extra_args, ex:
	//   init:
	//     extra_args: [a, b]
	// We validate if there's a single key in the map and if the value is a
	// legal value later.
	var step map[string]map[string][]string
	err = unmarshal(&step)
	if err == nil {
		s.Map = step
		return nil
	}

	// Try to unmarshal as a custom run step, ex.
	// steps:
	// - run: my command
	// We validate if the key is run later.
	var runStep map[string]string
	err = unmarshal(&runStep)
	if err == nil {
		s.StringVal = runStep
		return nil
	}

	return err
}
