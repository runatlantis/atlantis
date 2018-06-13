package yaml_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/runatlantis/atlantis/server/events/yaml"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
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
			"parsing atlantis.yaml: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `slkjds` into raw.Spec",
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

func TestReadConfig(t *testing.T) {
	cases := []struct {
		description string
		input       string
		expErr      string
		exp         valid.Spec
	}{
		// Version key.
		{
			description: "no version",
			input: `
projects:
- dir: "."
`,
			expErr: "version: is required.",
		},
		{
			description: "unsupported version",
			input: `
version: 0
projects:
- dir: "."
`,
			expErr: "version: must equal 2.",
		},
		{
			description: "empty version",
			input: `
version: ~
projects:
- dir: "."
`,
			expErr: "version: must equal 2.",
		},

		// Projects key.
		{
			description: "empty projects list",
			input: `
version: 2
projects:`,
			exp: valid.Spec{
				Version:   2,
				Projects:  nil,
				Workflows: map[string]valid.Workflow{},
			},
		},
		{
			description: "project dir not set",
			input: `
version: 2
projects:
- `,
			expErr: "projects: (0: (dir: cannot be blank.).).",
		},
		{
			description: "project dir set",
			input: `
version: 2
projects:
- dir: .`,
			exp: valid.Spec{
				Version: 2,
				Projects: []valid.Project{
					{
						Dir:              ".",
						Workspace:        "default",
						Workflow:         nil,
						TerraformVersion: nil,
						Autoplan: valid.Autoplan{
							WhenModified: []string{"**/*.tf"},
							Enabled:      true,
						},
						ApplyRequirements: nil,
					},
				},
				Workflows: map[string]valid.Workflow{},
			},
		},
		{
			description: "project fields set except autoplan",
			input: `
version: 2
projects:
- dir: .
  workspace: myworkspace
  terraform_version: v0.11.0
  apply_requirements: [approved]
  workflow: myworkflow
workflows:
  myworkflow: ~`,
			exp: valid.Spec{
				Version: 2,
				Projects: []valid.Project{
					{
						Dir:              ".",
						Workspace:        "myworkspace",
						Workflow:         String("myworkflow"),
						TerraformVersion: String("v0.11.0"),
						Autoplan: valid.Autoplan{
							WhenModified: []string{"**/*.tf"},
							Enabled:      true,
						},
						ApplyRequirements: []string{"approved"},
					},
				},
				Workflows: map[string]valid.Workflow{
					"myworkflow": {},
				},
			},
		},
		{
			description: "project field with autoplan",
			input: `
version: 2
projects:
- dir: .
  workspace: myworkspace
  terraform_version: v0.11.0
  apply_requirements: [approved]
  workflow: myworkflow
  autoplan:
    enabled: false
workflows:
  myworkflow: ~`,
			exp: valid.Spec{
				Version: 2,
				Projects: []valid.Project{
					{
						Dir:              ".",
						Workspace:        "myworkspace",
						Workflow:         String("myworkflow"),
						TerraformVersion: String("v0.11.0"),
						Autoplan: valid.Autoplan{
							WhenModified: []string{"**/*.tf"},
							Enabled:      false,
						},
						ApplyRequirements: []string{"approved"},
					},
				},
				Workflows: map[string]valid.Workflow{
					"myworkflow": {},
				},
			},
		},
		{
			description: "project dir with ..",
			input: `
version: 2
projects:
- dir: ..`,
			expErr: "projects: (0: (dir: cannot contain '..'.).).",
		},

		// Project must have dir set.
		{
			description: "project with no config",
			input: `
version: 2
projects:
-`,
			expErr: "projects: (0: (dir: cannot be blank.).).",
		},
		{
			description: "project with no config at index 1",
			input: `
version: 2
projects:
- dir: "."
-`,
			expErr: "projects: (1: (dir: cannot be blank.).).",
		},
		{
			description: "project with unknown key",
			input: `
version: 2
projects:
- unknown: value`,
			expErr: "yaml: unmarshal errors:\n  line 4: field unknown not found in struct raw.Project",
		},
		{
			description: "referencing workflow that doesn't exist",
			input: `
version: 2
projects:
- dir: .
  workflow: undefined`,
			expErr: "workflow \"undefined\" is not defined",
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
			if c.expErr != "" {
				ErrEquals(t, "parsing atlantis.yaml: "+c.expErr, err)
				return
			}
			Ok(t, err)
			Equals(t, c.exp, act)
		})
	}
}

func TestReadConfig_Successes(t *testing.T) {
	basicProjects := []valid.Project{
		{
			Autoplan: valid.Autoplan{
				Enabled:      true,
				WhenModified: []string{"**/*.tf"},
			},
			Workspace:         "default",
			ApplyRequirements: nil,
			Dir:               ".",
		},
	}

	cases := []struct {
		description string
		input       string
		expOutput   valid.Spec
	}{
		{
			description: "uses project defaults",
			input: `
version: 2
projects:
- dir: "."`,
			expOutput: valid.Spec{
				Version:   2,
				Projects:  basicProjects,
				Workflows: make(map[string]valid.Workflow),
			},
		},
		{
			description: "autoplan is enabled by default",
			input: `
version: 2
projects:
- dir: "."
  autoplan:
    when_modified: ["**/*.tf"]
`,
			expOutput: valid.Spec{
				Version:   2,
				Projects:  basicProjects,
				Workflows: make(map[string]valid.Workflow),
			},
		},
		{
			description: "if workflows not defined there are none",
			input: `
version: 2
projects:
- dir: "."
`,
			expOutput: valid.Spec{
				Version:   2,
				Projects:  basicProjects,
				Workflows: make(map[string]valid.Workflow),
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
			expOutput: valid.Spec{
				Version:   2,
				Projects:  basicProjects,
				Workflows: make(map[string]valid.Workflow),
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
			expOutput: valid.Spec{
				Version:  2,
				Projects: basicProjects,
				Workflows: map[string]valid.Workflow{
					"default": {
						Plan: &valid.Stage{
							Steps: nil,
						},
						Apply: &valid.Stage{
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
			expOutput: valid.Spec{
				Version:  2,
				Projects: basicProjects,
				Workflows: map[string]valid.Workflow{
					"default": {
						Plan: &valid.Stage{
							Steps: []valid.Step{
								{
									StepName: "init",
								},
								{
									StepName: "plan",
								},
							},
						},
						Apply: &valid.Stage{
							Steps: []valid.Step{
								{
									StepName: "plan",
								},
								{
									StepName: "apply",
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
			expOutput: valid.Spec{
				Version:  2,
				Projects: basicProjects,
				Workflows: map[string]valid.Workflow{
					"default": {
						Plan: &valid.Stage{
							Steps: []valid.Step{
								{
									StepName:  "init",
									ExtraArgs: []string{},
								},
								{
									StepName:  "plan",
									ExtraArgs: []string{"arg1", "arg2"},
								},
							},
						},
						Apply: &valid.Stage{
							Steps: []valid.Step{
								{
									StepName:  "plan",
									ExtraArgs: []string{"a", "b"},
								},
								{
									StepName:  "apply",
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
			expOutput: valid.Spec{
				Version:  2,
				Projects: basicProjects,
				Workflows: map[string]valid.Workflow{
					"default": {
						Plan: &valid.Stage{
							Steps: []valid.Step{
								{
									StepName:   "run",
									RunCommand: []string{"echo", "plan hi"},
								},
							},
						},
						Apply: &valid.Stage{
							Steps: []valid.Step{
								{
									StepName:   "run",
									RunCommand: []string{"echo", "apply", "arg 2"},
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

// Bool is a helper routine that allocates a new bool value
// to store v and returns a pointer to it.
func Bool(v bool) *bool { return &v }

// Int is a helper routine that allocates a new int value
// to store v and returns a pointer to it.
func Int(v int) *int { return &v }

// String is a helper routine that allocates a new string value
// to store v and returns a pointer to it.
func String(v string) *string { return &v }
