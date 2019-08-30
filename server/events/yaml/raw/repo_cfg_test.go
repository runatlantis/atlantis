package raw_test

import (
	"testing"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/runatlantis/atlantis/server/events/yaml/raw"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
	. "github.com/runatlantis/atlantis/testing"
	yaml "gopkg.in/yaml.v2"
)

func TestConfig_UnmarshalYAML(t *testing.T) {
	cases := []struct {
		description string
		input       string
		exp         raw.RepoCfg
		expErr      string
	}{
		{
			description: "no data",
			input:       "",
			exp: raw.RepoCfg{
				Version:   nil,
				Projects:  nil,
				Workflows: nil,
			},
		},
		{
			description: "yaml nil",
			input:       "~",
			exp: raw.RepoCfg{
				Version:   nil,
				Projects:  nil,
				Workflows: nil,
			},
		},
		{
			description: "invalid key",
			input:       "invalid: key",
			exp: raw.RepoCfg{
				Version:   nil,
				Projects:  nil,
				Workflows: nil,
			},
			expErr: "yaml: unmarshal errors:\n  line 1: field invalid not found in type raw.RepoCfg",
		},
		{
			description: "version set to 2",
			input:       "version: 2",
			exp: raw.RepoCfg{
				Version:   Int(2),
				Projects:  nil,
				Workflows: nil,
			},
		},
		{
			description: "version set to 3",
			input:       "version: 3",
			exp: raw.RepoCfg{
				Version:   Int(3),
				Projects:  nil,
				Workflows: nil,
			},
		},
		{
			description: "projects key without value",
			input:       "projects:",
			exp: raw.RepoCfg{
				Version:   nil,
				Projects:  nil,
				Workflows: nil,
			},
		},
		{
			description: "workflows key without value",
			input:       "workflows:",
			exp: raw.RepoCfg{
				Version:   nil,
				Projects:  nil,
				Workflows: nil,
			},
		},
		{
			description: "projects with a map",
			input:       "projects:\n  key: value",
			exp: raw.RepoCfg{
				Version:   nil,
				Projects:  nil,
				Workflows: nil,
			},
			expErr: "yaml: unmarshal errors:\n  line 2: cannot unmarshal !!map into []raw.Project",
		},
		{
			description: "projects with a scalar",
			input:       "projects: value",
			exp: raw.RepoCfg{
				Version:   nil,
				Projects:  nil,
				Workflows: nil,
			},
			expErr: "yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `value` into []raw.Project",
		},
		{
			description: "automerge not a boolean",
			input:       "version: 3\nautomerge: notabool",
			exp: raw.RepoCfg{
				Version:   nil,
				Projects:  nil,
				Workflows: nil,
			},
			expErr: "yaml: unmarshal errors:\n  line 2: cannot unmarshal !!str `notabool` into bool",
		},
		{
			description: "should use values if set",
			input: `
version: 3
automerge: true
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
			exp: raw.RepoCfg{
				Version:   Int(3),
				Automerge: Bool(true),
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
			},
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			var conf raw.RepoCfg
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
		input       raw.RepoCfg
		expErr      string
	}{
		{
			description: "version not nil",
			input: raw.RepoCfg{
				Version: nil,
			},
			expErr: "version: is required. If you've just upgraded Atlantis you need to rewrite your atlantis.yaml for version 3. See www.runatlantis.io/docs/upgrading-atlantis-yaml.html.",
		},
		{
			description: "version not 2 or 3",
			input: raw.RepoCfg{
				Version: Int(1),
			},
			expErr: "version: only versions 2 and 3 are supported.",
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
		input       raw.RepoCfg
		exp         valid.RepoCfg
	}{
		{
			description: "nothing set",
			input:       raw.RepoCfg{Version: Int(2)},
			exp: valid.RepoCfg{
				Version:   2,
				Workflows: make(map[string]valid.Workflow),
			},
		},
		{
			description: "set to empty",
			input: raw.RepoCfg{
				Version:   Int(2),
				Workflows: map[string]raw.Workflow{},
				Projects:  []raw.Project{},
			},
			exp: valid.RepoCfg{
				Version:   2,
				Workflows: map[string]valid.Workflow{},
				Projects:  nil,
			},
		},
		{
			description: "automerge ommitted",
			input: raw.RepoCfg{
				Version: Int(2),
			},
			exp: valid.RepoCfg{
				Version:   2,
				Automerge: false,
				Workflows: map[string]valid.Workflow{},
			},
		},
		{
			description: "automerge true",
			input: raw.RepoCfg{
				Version:   Int(2),
				Automerge: Bool(true),
			},
			exp: valid.RepoCfg{
				Version:   2,
				Automerge: true,
				Workflows: map[string]valid.Workflow{},
			},
		},
		{
			description: "automerge false",
			input: raw.RepoCfg{
				Version:   Int(2),
				Automerge: Bool(false),
			},
			exp: valid.RepoCfg{
				Version:   2,
				Automerge: false,
				Workflows: map[string]valid.Workflow{},
			},
		},
		{
			description: "only plan stage set",
			input: raw.RepoCfg{
				Version: Int(2),
				Workflows: map[string]raw.Workflow{
					"myworkflow": {
						Plan:  &raw.Stage{},
						Apply: nil,
					},
				},
			},
			exp: valid.RepoCfg{
				Version:   2,
				Automerge: false,
				Workflows: map[string]valid.Workflow{
					"myworkflow": {
						Name: "myworkflow",
						Plan: valid.DefaultPlanStage,
						Apply: valid.Stage{
							Steps: []valid.Step{
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
			description: "everything set",
			input: raw.RepoCfg{
				Version:   Int(2),
				Automerge: Bool(true),
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
			exp: valid.RepoCfg{
				Version:   2,
				Automerge: true,
				Workflows: map[string]valid.Workflow{
					"myworkflow": {
						Name: "myworkflow",
						Apply: valid.Stage{
							Steps: []valid.Step{
								{
									StepName: "apply",
								},
							},
						},
						Plan: valid.Stage{
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
