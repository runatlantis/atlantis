package yaml

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
