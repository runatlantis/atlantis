package yaml_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/events/yaml"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
	. "github.com/runatlantis/atlantis/testing"
)

func TestReadConfig_DirDoesNotExist(t *testing.T) {
	r := yaml.ParserValidator{}
	_, err := r.ReadConfig("/not/exist")
	Assert(t, os.IsNotExist(err), "exp nil ptr")

	exists, err := r.HasConfigFile("/not/exist")
	Ok(t, err)
	Equals(t, false, exists)
}

func TestReadConfig_FileDoesNotExist(t *testing.T) {
	tmpDir, cleanup := TempDir(t)
	defer cleanup()

	r := yaml.ParserValidator{}
	_, err := r.ReadConfig(tmpDir)
	Assert(t, os.IsNotExist(err), "exp nil ptr")

	exists, err := r.HasConfigFile(tmpDir)
	Ok(t, err)
	Equals(t, false, exists)
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
			"parsing atlantis.yaml: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `slkjds` into raw.Config",
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
	tfVersion, _ := version.NewVersion("v0.11.0")
	cases := []struct {
		description string
		input       string
		expErr      string
		exp         valid.Config
	}{
		// Version key.
		{
			description: "no version",
			input: `
projects:
- dir: "."
`,
			expErr: "version: is required. If you've just upgraded Atlantis you need to rewrite your atlantis.yaml for version 2. See www.runatlantis.io/docs/upgrading-atlantis-yaml-to-version-2.html.",
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
			exp: valid.Config{
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
			exp: valid.Config{
				Version: 2,
				Projects: []valid.Project{
					{
						Dir:              ".",
						Workspace:        "default",
						Workflow:         nil,
						TerraformVersion: nil,
						Autoplan: valid.Autoplan{
							WhenModified: []string{"**/*.tf*"},
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
			exp: valid.Config{
				Version: 2,
				Projects: []valid.Project{
					{
						Dir:              ".",
						Workspace:        "myworkspace",
						Workflow:         String("myworkflow"),
						TerraformVersion: tfVersion,
						Autoplan: valid.Autoplan{
							WhenModified: []string{"**/*.tf*"},
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
			exp: valid.Config{
				Version: 2,
				Projects: []valid.Project{
					{
						Dir:              ".",
						Workspace:        "myworkspace",
						Workflow:         String("myworkflow"),
						TerraformVersion: tfVersion,
						Autoplan: valid.Autoplan{
							WhenModified: []string{"**/*.tf*"},
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
			description: "project field with mergeable apply requirement",
			input: `
version: 2
projects:
- dir: .
  workspace: myworkspace
  terraform_version: v0.11.0
  apply_requirements: [mergeable]
  workflow: myworkflow
  autoplan:
    enabled: false
workflows:
  myworkflow: ~`,
			exp: valid.Config{
				Version: 2,
				Projects: []valid.Project{
					{
						Dir:              ".",
						Workspace:        "myworkspace",
						Workflow:         String("myworkflow"),
						TerraformVersion: tfVersion,
						Autoplan: valid.Autoplan{
							WhenModified: []string{"**/*.tf*"},
							Enabled:      false,
						},
						ApplyRequirements: []string{"mergeable"},
					},
				},
				Workflows: map[string]valid.Workflow{
					"myworkflow": {},
				},
			},
		},
		{
			description: "project field with mergeable and approved apply requirements",
			input: `
version: 2
projects:
- dir: .
  workspace: myworkspace
  terraform_version: v0.11.0
  apply_requirements: [mergeable, approved]
  workflow: myworkflow
  autoplan:
    enabled: false
workflows:
  myworkflow: ~`,
			exp: valid.Config{
				Version: 2,
				Projects: []valid.Project{
					{
						Dir:              ".",
						Workspace:        "myworkspace",
						Workflow:         String("myworkflow"),
						TerraformVersion: tfVersion,
						Autoplan: valid.Autoplan{
							WhenModified: []string{"**/*.tf*"},
							Enabled:      false,
						},
						ApplyRequirements: []string{"mergeable", "approved"},
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
		{
			description: "two projects with same dir/workspace without names",
			input: `
version: 2
projects:
- dir: .
  workspace: workspace
- dir: .
  workspace: workspace`,
			expErr: "there are two or more projects with dir: \".\" workspace: \"workspace\" that are not all named; they must have a 'name' key so they can be targeted for apply's separately",
		},
		{
			description: "two projects with same dir/workspace only one with name",
			input: `
version: 2
projects:
- name: myname
  dir: .
  workspace: workspace
- dir: .
  workspace: workspace`,
			expErr: "there are two or more projects with dir: \".\" workspace: \"workspace\" that are not all named; they must have a 'name' key so they can be targeted for apply's separately",
		},
		{
			description: "two projects with same dir/workspace both with same name",
			input: `
version: 2
projects:
- name: myname
  dir: .
  workspace: workspace
- name: myname
  dir: .
  workspace: workspace`,
			expErr: "found two or more projects with name \"myname\"; project names must be unique",
		},
		{
			description: "two projects with same dir/workspace with different names",
			input: `
version: 2
projects:
- name: myname
  dir: .
  workspace: workspace
- name: myname2
  dir: .
  workspace: workspace`,
			exp: valid.Config{
				Version: 2,
				Projects: []valid.Project{
					{
						Name:      String("myname"),
						Dir:       ".",
						Workspace: "workspace",
						Autoplan: valid.Autoplan{
							WhenModified: []string{"**/*.tf*"},
							Enabled:      true,
						},
					},
					{
						Name:      String("myname2"),
						Dir:       ".",
						Workspace: "workspace",
						Autoplan: valid.Autoplan{
							WhenModified: []string{"**/*.tf*"},
							Enabled:      true,
						},
					},
				},
				Workflows: map[string]valid.Workflow{},
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
				WhenModified: []string{"**/*.tf*"},
			},
			Workspace:         "default",
			ApplyRequirements: nil,
			Dir:               ".",
		},
	}

	cases := []struct {
		description string
		input       string
		expOutput   valid.Config
	}{
		{
			description: "uses project defaults",
			input: `
version: 2
projects:
- dir: "."`,
			expOutput: valid.Config{
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
    when_modified: ["**/*.tf*"]
`,
			expOutput: valid.Config{
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
			expOutput: valid.Config{
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
			expOutput: valid.Config{
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
			expOutput: valid.Config{
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
			expOutput: valid.Config{
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
			expOutput: valid.Config{
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
			expOutput: valid.Config{
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

// String is a helper routine that allocates a new string value
// to store v and returns a pointer to it.
func String(v string) *string { return &v }
