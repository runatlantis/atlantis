package yaml_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/runatlantis/atlantis/server/events/yaml"
	. "github.com/runatlantis/atlantis/testing"
)

func TestReadConfig_DirDoesNotExist(t *testing.T) {
	r := yaml.ParserValidator{}
	_, err := r.ReadConfig("/not/exist")
	Assert(t, os.IsNotExist(err), "exp nil ptr")
}

func TestReadConfig_FileDoesNotExist(t *testing.T) {
	tmpDir, cleanup := TempDir(t)
	defer cleanup()

	r := yaml.ParserValidator{}
	_, err := r.ReadConfig(tmpDir)
	Assert(t, os.IsNotExist(err), "exp nil ptr")
}

func TestReadConfig_BadPermissions(t *testing.T) {
	tmpDir, cleanup := TempDir(t)
	defer cleanup()
	err := ioutil.WriteFile(filepath.Join(tmpDir, "atlantis.yaml"), nil, 0000)
	Ok(t, err)

	r := yaml.ParserValidator{}
	_, err = r.ReadConfig(tmpDir)
	ErrContains(t, "unable to read atlantis.yaml file: ", err)
}

func TestReadConfig_UnmarshalErrors(t *testing.T) {
	// We only have a few cases here because we assume the YAML library to be
	// well tested. See https://github.com/go-yaml/yaml/blob/v2/decode_test.go#L810.
	cases := []struct {
		description string
		input       string
		expErr      string
	}{
		{
			"random characters",
			"slkjds",
			"parsing atlantis.yaml: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `slkjds` into yaml.Config",
		},
		{
			"just a colon",
			":",
			"parsing atlantis.yaml: yaml: did not find expected key",
		},
	}

	tmpDir, cleanup := TempDir(t)
	defer cleanup()

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			err := ioutil.WriteFile(filepath.Join(tmpDir, "atlantis.yaml"), []byte(c.input), 0600)
			Ok(t, err)
			r := yaml.ParserValidator{}
			_, err = r.ReadConfig(tmpDir)
			ErrEquals(t, c.expErr, err)
		})
	}
}

func TestReadConfig_Invalid(t *testing.T) {
	cases := []struct {
		description string
		input       string
		expErr      string
	}{
		// Invalid version.
		{
			description: "no version",
			input: `
projects:
- dir: "."
`,
			expErr: "unknown version: must have \"version: 2\" set",
		},
		{
			description: "unsupported version",
			input: `
version: 0
projects:
- dir: "."
`,
			expErr: "unknown version: must have \"version: 2\" set",
		},
		{
			description: "empty version",
			input: `
version: ~
projects:
- dir: "."
`,
			expErr: "unknown version: must have \"version: 2\" set",
		},

		// No projects specified.
		{
			description: "no projects key",
			input:       `version: 2`,
			expErr:      "'projects' key must exist and contain at least one element",
		},
		{
			description: "empty projects list",
			input: `
version: 2
projects:`,
			expErr: "'projects' key must exist and contain at least one element",
		},

		// Project must have dir set.
		{
			description: "project with no config",
			input: `
version: 2
projects:
-`,
			expErr: "project at index 0 invalid: dir key must be set and non-empty",
		},
		{
			description: "project without dir set",
			input: `
version: 2
projects:
- workspace: "staging"`,
			expErr: "project at index 0 invalid: dir key must be set and non-empty",
		},
		{
			description: "project with dir set to empty string",
			input: `
version: 2
projects:
- dir: ""`,
			expErr: "project at index 0 invalid: dir key must be set and non-empty",
		},
		{
			description: "project with no config at index 1",
			input: `
version: 2
projects:
- dir: "."
-`,
			expErr: "project at index 1 invalid: dir key must be set and non-empty",
		},
		{
			description: "project with unknown key",
			input: `
version: 2
projects:
- unknown: value`,
			expErr: "yaml: unmarshal errors:\n  line 4: field unknown not found in struct yaml.alias",
		},
	}

	tmpDir, cleanup := TempDir(t)
	defer cleanup()

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			err := ioutil.WriteFile(filepath.Join(tmpDir, "atlantis.yaml"), []byte(c.input), 0600)
			Ok(t, err)

			r := yaml.ParserValidator{}
			_, err = r.ReadConfig(tmpDir)
			ErrEquals(t, "parsing atlantis.yaml: "+c.expErr, err)
		})
	}
}

func TestReadConfig_Successes(t *testing.T) {
	basicProjects := []yaml.Project{
		{
			AutoPlan: &yaml.AutoPlan{
				Enabled:      true,
				WhenModified: []string{"**/*.tf"},
			},
			Workspace:         "default",
			TerraformVersion:  "",
			ApplyRequirements: nil,
			Workflow:          "",
			Dir:               ".",
		},
	}

	cases := []struct {
		description string
		input       string
		expOutput   yaml.Config
	}{
		{
			description: "uses project defaults",
			input: `
version: 2
projects:
- dir: "."`,
			expOutput: yaml.Config{
				Version:  2,
				Projects: basicProjects,
			},
		},
		{
			description: "autoplan is enabled by default",
			input: `
version: 2
projects:
- dir: "."
  auto_plan:
    when_modified: ["**/*.tf"]
`,
			expOutput: yaml.Config{
				Version:  2,
				Projects: basicProjects,
			},
		},
		{
			description: "if workflows not defined, there are none",
			input: `
version: 2
projects:
- dir: "."
`,
			expOutput: yaml.Config{
				Version:  2,
				Projects: basicProjects,
			},
		},
		{
			description: "if workflows key set but with no workflows there are none",
			input: `
version: 2
projects:
- dir: "."
workflows: ~
`,
			expOutput: yaml.Config{
				Version:  2,
				Projects: basicProjects,
			},
		},
		{
			description: "if a workflow is defined but set to null we use the defaults",
			input: `
version: 2
projects:
- dir: "."
workflows:
  default: ~
`,
			expOutput: yaml.Config{
				Version:  2,
				Projects: basicProjects,
				Workflows: map[string]yaml.Workflow{
					"default": {
						Plan: &yaml.Stage{
							Steps: []yaml.StepConfig{
								{
									StepType: "init",
								},
								{
									StepType: "plan",
								},
							},
						},
						Apply: &yaml.Stage{
							Steps: []yaml.StepConfig{
								{
									StepType: "apply",
								},
							},
						},
					},
				},
			},
		},
		{
			description: "if a plan or apply has no steps defined then we use the defaults",
			input: `
version: 2
projects:
- dir: "."
workflows:
  default:
    plan:
    apply:
`,
			expOutput: yaml.Config{
				Version:  2,
				Projects: basicProjects,
				Workflows: map[string]yaml.Workflow{
					"default": {
						Plan: &yaml.Stage{
							Steps: []yaml.StepConfig{
								{
									StepType: "init",
								},
								{
									StepType: "plan",
								},
							},
						},
						Apply: &yaml.Stage{
							Steps: []yaml.StepConfig{
								{
									StepType: "apply",
								},
							},
						},
					},
				},
			},
		},
		{
			description: "if a plan or apply explicitly defines an empty steps key then there are no steps",
			input: `
version: 2
projects:
- dir: "."
workflows:
  default:
    plan:
      steps:
    apply:
      steps:
`,
			expOutput: yaml.Config{
				Version:  2,
				Projects: basicProjects,
				Workflows: map[string]yaml.Workflow{
					"default": {
						Plan: &yaml.Stage{
							Steps: nil,
						},
						Apply: &yaml.Stage{
							Steps: nil,
						},
					},
				},
			},
		},
		{
			description: "if steps are set then we parse them properly",
			input: `
version: 2
projects:
- dir: "."
workflows:
  default:
    plan:
      steps:
      - init
      - plan
    apply:
      steps:
      - plan # we don't validate if they make sense
      - apply
`,
			expOutput: yaml.Config{
				Version:  2,
				Projects: basicProjects,
				Workflows: map[string]yaml.Workflow{
					"default": {
						Plan: &yaml.Stage{
							Steps: []yaml.StepConfig{
								{
									StepType: "init",
								},
								{
									StepType: "plan",
								},
							},
						},
						Apply: &yaml.Stage{
							Steps: []yaml.StepConfig{
								{
									StepType: "plan",
								},
								{
									StepType: "apply",
								},
							},
						},
					},
				},
			},
		},
		{
			description: "we parse extra_args for the steps",
			input: `
version: 2
projects:
- dir: "."
workflows:
  default:
    plan:
      steps:
      - init:
          extra_args: []
      - plan:
          extra_args:
          - arg1
          - arg2
    apply:
      steps:
      - plan:
          extra_args: [a, b]
      - apply:
          extra_args: ["a", "b"]
`,
			expOutput: yaml.Config{
				Version:  2,
				Projects: basicProjects,
				Workflows: map[string]yaml.Workflow{
					"default": {
						Plan: &yaml.Stage{
							Steps: []yaml.StepConfig{
								{
									StepType:  "init",
									ExtraArgs: []string{},
								},
								{
									StepType:  "plan",
									ExtraArgs: []string{"arg1", "arg2"},
								},
							},
						},
						Apply: &yaml.Stage{
							Steps: []yaml.StepConfig{
								{
									StepType:  "plan",
									ExtraArgs: []string{"a", "b"},
								},
								{
									StepType:  "apply",
									ExtraArgs: []string{"a", "b"},
								},
							},
						},
					},
				},
			},
		},
		{
			description: "custom steps are parsed",
			input: `
version: 2
projects:
- dir: "."
workflows:
  default:
    plan:
      steps:
      - run: "echo \"plan hi\""
    apply:
      steps:
      - run: echo apply "arg 2"
`,
			expOutput: yaml.Config{
				Version:  2,
				Projects: basicProjects,
				Workflows: map[string]yaml.Workflow{
					"default": {
						Plan: &yaml.Stage{
							Steps: []yaml.StepConfig{
								{
									StepType: "run",
									Run:      []string{"echo", "plan hi"},
								},
							},
						},
						Apply: &yaml.Stage{
							Steps: []yaml.StepConfig{
								{
									StepType: "run",
									Run:      []string{"echo", "apply", "arg 2"},
								},
							},
						},
					},
				},
			},
		},
	}

	tmpDir, cleanup := TempDir(t)
	defer cleanup()

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			err := ioutil.WriteFile(filepath.Join(tmpDir, "atlantis.yaml"), []byte(c.input), 0600)
			Ok(t, err)

			r := yaml.ParserValidator{}
			act, err := r.ReadConfig(tmpDir)
			Ok(t, err)
			Equals(t, c.expOutput, act)
		})
	}
}
