package raw

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/utils"
)

const (
	ExtraArgsKey        = "extra_args"
	NameArgKey          = "name"
	CommandArgKey       = "command"
	ValueArgKey         = "value"
	OutputArgKey        = "output"
	RunStepName         = "run"
	PlanStepName        = "plan"
	ShowStepName        = "show"
	PolicyCheckStepName = "policy_check"
	ApplyStepName       = "apply"
	InitStepName        = "init"
	EnvStepName         = "env"
	MultiEnvStepName    = "multienv"
	ImportStepName      = "import"
	StateRmStepName     = "state_rm"
	ShellArgKey         = "shell"
	ShellArgsArgKey     = "shellArgs"
)

/*
Step represents a single action/command to perform. In YAML, it can be set as
1. A single string for a built-in command:
  - init
  - plan
  - policy_check

2. A map for an env step with name and command or value, or a run step with a command and output config
  - env:
    name: test_command
    command: echo 312
  - env:
    name: test_value
    value: value
  - env:
    name: test_bash_command
    command: echo ${test_value::7}
    shell: bash
    shellArgs: ["--verbose", "-c"]
  - multienv:
    command: envs.sh
    output: hide
    shell: sh
    shellArgs: -c
  - run:
    command: my custom command
    output: hide

3. A map for a built-in command and extra_args:
  - plan:
    extra_args: [-var-file=staging.tfvars]

4. A map for a custom run command:
  - run: my custom command

Here we parse step in the most generic fashion possible. See fields for more
details.
*/
type Step struct {
	// Key will be set in case #1 and #3 above to the key. In case #2, there
	// could be multiple keys (since the element is a map) so we don't set Key.
	Key *string
	// StringVal will be set in case #4 above.
	StringVal map[string]string
	// Map will be set in case #3 above.
	Map map[string]map[string][]string
	// CommandMap will be set in case #2 above.
	CommandMap map[string]map[string]interface{}
}

func (s *Step) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return s.unmarshalGeneric(unmarshal)
}

func (s Step) MarshalYAML() (interface{}, error) {
	return s.marshalGeneric()
}

func (s *Step) UnmarshalJSON(data []byte) error {
	return s.unmarshalGeneric(func(i interface{}) error {
		return json.Unmarshal(data, i)
	})
}

func (s *Step) MarshalJSON() ([]byte, error) {
	out, err := s.marshalGeneric()
	if err != nil {
		return nil, err
	}
	return json.Marshal(out)
}

func (s Step) validStepName(stepName string) bool {
	return stepName == InitStepName ||
		stepName == PlanStepName ||
		stepName == ApplyStepName ||
		stepName == EnvStepName ||
		stepName == MultiEnvStepName ||
		stepName == ShowStepName ||
		stepName == PolicyCheckStepName ||
		stepName == ImportStepName ||
		stepName == StateRmStepName
}

func (s Step) Validate() error {
	validStep := func(value interface{}) error {
		str := *value.(*string)
		if !s.validStepName(str) {
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
			if !s.validStepName(stepName) {
				return fmt.Errorf("%q is not a valid step type", stepName)
			}
			var argKeys []string
			for k := range args {
				argKeys = append(argKeys, k)
			}
			// Sort so tests can be deterministic.
			sort.Strings(argKeys)

			// args should contain a single 'extra_args' key.
			if len(argKeys) > 1 {
				return fmt.Errorf("built-in steps only support a single %s key, found %d: %s",
					ExtraArgsKey, len(argKeys), strings.Join(argKeys, ","))
			}
			for k := range args {
				if k != ExtraArgsKey {
					return fmt.Errorf("built-in steps only support a single %s key, found %q in step %s",
						ExtraArgsKey, k, stepName)
				}
			}
		}
		return nil
	}

	envOrRunOrMultiEnvStep := func(value interface{}) error {
		elem := value.(map[string]map[string]interface{})
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
		if len(keys) == 0 {
			return fmt.Errorf("step element must contain at least 1 key")
		}

		stepName := keys[0]
		args := elem[keys[0]]

		var argKeys []string
		for k := range args {
			argKeys = append(argKeys, k)
		}
		argMap := make(map[string]interface{})
		for k, v := range args {
			argMap[k] = v
		}
		// Sort so tests can be deterministic.
		sort.Strings(argKeys)

		// Validate keys common for all the steps.
		if utils.SlicesContains(argKeys, ShellArgKey) && !utils.SlicesContains(argKeys, CommandArgKey) {
			return fmt.Errorf("workflow steps only support %q key in combination with %q key",
				ShellArgKey, CommandArgKey)
		}
		if utils.SlicesContains(argKeys, ShellArgsArgKey) && !utils.SlicesContains(argKeys, ShellArgKey) {
			return fmt.Errorf("workflow steps only support %q key in combination with %q key",
				ShellArgsArgKey, ShellArgKey)
		}

		switch t := argMap[ShellArgsArgKey].(type) {
		case nil:
		case string:
		case []interface{}:
			for _, e := range t {
				if _, ok := e.(string); !ok {
					return fmt.Errorf("%q step %q option must contain only strings, found %v",
						stepName, ShellArgsArgKey, e)
				}
			}
		default:
			return fmt.Errorf("%q step %q option must be a string or a list of strings, found %v",
				stepName, ShellArgsArgKey, t)
		}
		delete(argMap, ShellArgsArgKey)
		delete(argMap, ShellArgKey)

		// Validate keys per step type.
		switch stepName {
		case EnvStepName:
			foundNameKey := false
			for _, k := range argKeys {
				if k != NameArgKey && k != CommandArgKey && k != ValueArgKey && k != ShellArgKey && k != ShellArgsArgKey {
					return fmt.Errorf("env steps only support keys %q, %q, %q, %q and %q, found key %q",
						NameArgKey, ValueArgKey, CommandArgKey, ShellArgKey, ShellArgsArgKey, k)
				}
				if k == NameArgKey {
					foundNameKey = true
				}
			}
			delete(argMap, CommandArgKey)
			if !foundNameKey {
				return fmt.Errorf("env steps must have a %q key set", NameArgKey)
			}
			delete(argMap, NameArgKey)
			if utils.SlicesContains(argKeys, ValueArgKey) && utils.SlicesContains(argKeys, CommandArgKey) {
				return fmt.Errorf("env steps only support one of the %q or %q keys, found both",
					ValueArgKey, CommandArgKey)
			}
			delete(argMap, ValueArgKey)
		case RunStepName, MultiEnvStepName:
			if _, ok := argMap[CommandArgKey].(string); !ok {
				return fmt.Errorf("%q step must have a %q key set", stepName, CommandArgKey)
			}
			delete(argMap, CommandArgKey)
			if v, ok := argMap[OutputArgKey].(string); ok {
				if stepName == RunStepName && (v != valid.PostProcessRunOutputShow &&
					v != valid.PostProcessRunOutputHide && v != valid.PostProcessRunOutputStripRefreshing) {
					return fmt.Errorf("run step %q option must be one of %q, %q, or %q",
						OutputArgKey, valid.PostProcessRunOutputShow, valid.PostProcessRunOutputHide,
						valid.PostProcessRunOutputStripRefreshing)
				} else if stepName == MultiEnvStepName && (v != valid.PostProcessRunOutputShow &&
					v != valid.PostProcessRunOutputHide) {
					return fmt.Errorf("multienv step %q option must be %q or %q",
						OutputArgKey, valid.PostProcessRunOutputShow, valid.PostProcessRunOutputHide)
				}
			}
			delete(argMap, OutputArgKey)
		default:
			return fmt.Errorf("%q is not a valid step type", stepName)
		}

		if len(argMap) > 0 {
			var argKeys []string
			for k := range argMap {
				argKeys = append(argKeys, k)
			}
			// Sort so tests can be deterministic.
			sort.Strings(argKeys)
			return fmt.Errorf("%q steps only support keys %q, %q, %q and %q, found extra keys %q",
				stepName, CommandArgKey, OutputArgKey, ShellArgKey, ShellArgsArgKey, strings.Join(argKeys, ","))
		}

		return nil
	}

	runOrMultiEnvStep := func(value interface{}) error {
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
			if stepName != RunStepName && stepName != MultiEnvStepName {
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
	if len(s.CommandMap) > 0 {
		return validation.Validate(s.CommandMap, validation.By(envOrRunOrMultiEnvStep))
	}
	if len(s.StringVal) > 0 {
		return validation.Validate(s.StringVal, validation.By(runOrMultiEnvStep))
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
	if len(s.CommandMap) > 0 {
		// After validation we assume there's only one key and it's a valid
		// step name so we just use the first one.
		for stepName, stepArgs := range s.CommandMap {
			step := valid.Step{StepName: stepName}
			if name, ok := stepArgs[NameArgKey].(string); ok {
				step.EnvVarName = name
			}
			if command, ok := stepArgs[CommandArgKey].(string); ok {
				step.RunCommand = command
			}
			if value, ok := stepArgs[ValueArgKey].(string); ok {
				step.EnvVarValue = value
			}
			if output, ok := stepArgs[OutputArgKey].(string); ok {
				step.Output = valid.PostProcessRunOutputOption(output)
			}
			if shell, ok := stepArgs[ShellArgKey].(string); ok {
				step.RunShell = &valid.CommandShell{
					Shell:     shell,
					ShellArgs: []string{"-c"},
				}
			}
			if step.StepName == RunStepName && step.Output == "" {
				step.Output = valid.PostProcessRunOutputShow
			}

			switch t := stepArgs[ShellArgsArgKey].(type) {
			case nil:
			case string:
				step.RunShell.ShellArgs = strings.Split(t, " ")
			case []interface{}:
				step.RunShell.ShellArgs = []string{}
				for _, e := range t {
					step.RunShell.ShellArgs = append(step.RunShell.ShellArgs, e.(string))
				}
			}

			return step
		}
	}

	// This will trigger in case #3 (see Step docs).
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

	// This will trigger in case #4 (see Step docs).
	if len(s.StringVal) > 0 {
		// After validation we assume there's only one key and it's a valid
		// step name so we just use the first one.
		for stepName, v := range s.StringVal {
			return valid.Step{
				StepName:   stepName,
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

	// Try to unmarshal as a custom run step, ex.
	// steps:
	//   - run: my command
	// We validate if the key is run later.
	var runStep map[string]string
	err = unmarshal(&runStep)
	if err == nil {
		s.StringVal = runStep
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

	// This represents a command steps env, run, and multienv, ex:
	// steps:
	//   - env:
	//       name: k
	//       command: exec
	//   - run:
	//       name: test_bash_command
	//       command: echo ${test_value::7}
	//       shell: bash
	//       shellArgs: ["--verbose", "-c"]
	var commandStep map[string]map[string]interface{}
	err = unmarshal(&commandStep)
	if err == nil {
		s.CommandMap = commandStep
		return nil
	}

	return err
}

func (s Step) marshalGeneric() (interface{}, error) {
	if len(s.StringVal) != 0 {
		return s.StringVal, nil
	} else if len(s.Map) != 0 {
		return s.Map, nil
	} else if len(s.CommandMap) != 0 {
		return s.CommandMap, nil
	} else if s.Key != nil {
		return s.Key, nil
	}

	// empty step should be marshalled to null, although this is generally
	// unexpected behavior.
	return nil, nil
}
