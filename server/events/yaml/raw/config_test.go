package raw_test

import (
	"testing"

	"github.com/go-ozzo/ozzo-validation"
	"github.com/runatlantis/atlantis/server/events/yaml/raw"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
	. "github.com/runatlantis/atlantis/testing"
	"gopkg.in/yaml.v2"
)

func TestConfig_UnmarshalYAML(t *testing.T) {
	cases := []struct {
		description string
		input       string
		exp         raw.Config
		expErr      string
	}{
		{
			description: "no data",
			input:       "",
			exp: raw.Config{
				Version:   nil,
				Projects:  nil,
				Workflows: nil,
				Automerge: nil,
			},
		},
		{
			description: "yaml nil",
			input:       "~",
			exp: raw.Config{
				Version:   nil,
				Projects:  nil,
				Workflows: nil,
				Automerge: nil,
			},
		},
		{
			description: "invalid key",
			input:       "invalid: key",
			exp: raw.Config{
				Version:   nil,
				Projects:  nil,
				Workflows: nil,
				Automerge: nil,
			},
			expErr: "yaml: unmarshal errors:\n  line 1: field invalid not found in struct raw.Config",
		},
		{
			description: "version set",
			input:       "version: 2",
			exp: raw.Config{
				Version:   Int(2),
				Projects:  nil,
				Workflows: nil,
				Automerge: nil,
			},
		},
		{
			description: "projects key without value",
			input:       "projects:",
			exp: raw.Config{
				Version:   nil,
				Projects:  nil,
				Workflows: nil,
				Automerge: nil,
			},
		},
		{
			description: "workflows key without value",
			input:       "workflows:",
			exp: raw.Config{
				Version:   nil,
				Projects:  nil,
				Workflows: nil,
			},
		},
		{
			description: "projects with a map",
			input:       "projects:\n  key: value",
			exp: raw.Config{
				Version:   nil,
				Projects:  nil,
				Workflows: nil,
				Automerge: nil,
			},
			expErr: "yaml: unmarshal errors:\n  line 2: cannot unmarshal !!map into []raw.Project",
		},
		{
			description: "projects with a scalar",
			input:       "projects: value",
			exp: raw.Config{
				Version:   nil,
				Projects:  nil,
				Workflows: nil,
				Automerge: nil,
			},
			expErr: "yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `value` into []raw.Project",
		},
		{
			description: "should use values if set",
			input: `
version: 2
automerge: false
projects:
- dir: mydir
  workspace: myworkspace
  workflow: default
  terraform_version: v0.11.0
  autoplan:
    enabled: false
    when_modified: []
  apply_requirements: [mergeable]
workflows:
  default:
    plan:
      steps: []
    apply:
     steps: []`,
			exp: raw.Config{
				Version: Int(2),
				Projects: []raw.Project{
					{
						Dir:              String("mydir"),
						Workspace:        String("myworkspace"),
						Workflow:         String("default"),
						TerraformVersion: String("v0.11.0"),
						Autoplan: &raw.Autoplan{
							WhenModified: []string{},
							Enabled:      Bool(false),
						},
						ApplyRequirements: []string{"mergeable"},
					},
				},
				Workflows: map[string]raw.Workflow{
					"default": {
						Apply: &raw.Stage{
							Steps: []raw.Step{},
						},
						Plan: &raw.Stage{
							Steps: []raw.Step{},
						},
					},
				},
				Automerge: Bool(false),
			},
		},
		{
			description: "automerge set to true",
			input: `
version: 2
automerge: true
`,
			exp: raw.Config{
				Version:   Int(2),
				Projects:  nil,
				Workflows: nil,
				Automerge: Bool(true),
			},
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			var conf raw.Config
			err := yaml.UnmarshalStrict([]byte(c.input), &conf)
			if c.expErr != "" {
				ErrEquals(t, c.expErr, err)
				return
			}
			Ok(t, err)
			Equals(t, c.exp, conf)
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	cases := []struct {
		description string
		input       raw.Config
		expErr      string
	}{
		{
			description: "version not nil",
			input: raw.Config{
				Version: nil,
			},
			expErr: "version: is required. If you've just upgraded Atlantis you need to rewrite your atlantis.yaml for version 2. See www.runatlantis.io/docs/upgrading-atlantis-yaml-to-version-2.html.",
		},
		{
			description: "version not 1",
			input: raw.Config{
				Version: Int(1),
			},
			expErr: "version: must equal 2.",
		},
	}
	validation.ErrorTag = "yaml"
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			err := c.input.Validate()
			if c.expErr == "" {
				Ok(t, err)
			} else {
				ErrEquals(t, c.expErr, err)
			}
		})
	}
}

func TestConfig_ToValid(t *testing.T) {
	cases := []struct {
		description string
		input       raw.Config
		exp         valid.Config
	}{
		{
			description: "nothing set",
			input:       raw.Config{Version: Int(2)},
			exp: valid.Config{
				Version:   2,
				Workflows: make(map[string]valid.Workflow),
			},
		},
		{
			description: "set to empty",
			input: raw.Config{
				Version:   Int(2),
				Workflows: map[string]raw.Workflow{},
				Projects:  []raw.Project{},
			},
			exp: valid.Config{
				Version:   2,
				Workflows: map[string]valid.Workflow{},
				Projects:  nil,
			},
		},
		{
			description: "everything set",
			input: raw.Config{
				Version: Int(2),
				Workflows: map[string]raw.Workflow{
					"myworkflow": {
						Apply: &raw.Stage{
							Steps: []raw.Step{
								{
									Key: String("apply"),
								},
							},
						},
						Plan: &raw.Stage{
							Steps: []raw.Step{
								{
									Key: String("init"),
								},
							},
						},
					},
				},
				Projects: []raw.Project{
					{
						Dir: String("mydir"),
					},
				},
			},
			exp: valid.Config{
				Version: 2,
				Workflows: map[string]valid.Workflow{
					"myworkflow": {
						Apply: &valid.Stage{
							Steps: []valid.Step{
								{
									StepName: "apply",
								},
							},
						},
						Plan: &valid.Stage{
							Steps: []valid.Step{
								{
									StepName: "init",
								},
							},
						},
					},
				},
				Projects: []valid.Project{
					{
						Dir:       "mydir",
						Workspace: "default",
						Autoplan: valid.Autoplan{
							WhenModified: []string{"**/*.tf*"},
							Enabled:      true,
						},
					},
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			Equals(t, c.exp, c.input.ToValid())
		})
	}
}
