package raw_test

import (
	"testing"

	"github.com/go-ozzo/ozzo-validation"
	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/events/yaml/raw"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
	. "github.com/runatlantis/atlantis/testing"
	"gopkg.in/yaml.v2"
)

func TestProject_UnmarshalYAML(t *testing.T) {
	cases := []struct {
		description string
		input       string
		exp         raw.Project
	}{
		{
			description: "omit unset fields",
			input:       "",
			exp: raw.Project{
				Dir:               nil,
				Workspace:         nil,
				Workflow:          nil,
				TerraformVersion:  nil,
				Autoplan:          nil,
				ApplyRequirements: nil,
				Name:              nil,
			},
		},
		{
			description: "all fields set",
			input: `
name: myname
dir: mydir
workspace: workspace
workflow: workflow
terraform_version: v0.11.0
autoplan:
  when_modified: []
  enabled: false
apply_requirements:
- mergeable`,
			exp: raw.Project{
				Name:             String("myname"),
				Dir:              String("mydir"),
				Workspace:        String("workspace"),
				Workflow:         String("workflow"),
				TerraformVersion: String("v0.11.0"),
				Autoplan: &raw.Autoplan{
					WhenModified: []string{},
					Enabled:      Bool(false),
				},
				ApplyRequirements: []string{"mergeable"},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			var p raw.Project
			err := yaml.UnmarshalStrict([]byte(c.input), &p)
			Ok(t, err)
			Equals(t, c.exp, p)
		})
	}
}

func TestProject_Validate(t *testing.T) {
	cases := []struct {
		description string
		input       raw.Project
		expErr      string
	}{
		{
			description: "minimal fields",
			input: raw.Project{
				Dir: String("."),
			},
			expErr: "",
		},
		{
			description: "dir empty",
			input: raw.Project{
				Dir: nil,
			},
			expErr: "dir: cannot be blank.",
		},
		{
			description: "dir with ..",
			input: raw.Project{
				Dir: String("../mydir"),
			},
			expErr: "dir: cannot contain '..'.",
		},
		{
			description: "apply reqs with unsupported",
			input: raw.Project{
				Dir:               String("."),
				ApplyRequirements: []string{"unsupported"},
			},
			expErr: "apply_requirements: \"unsupported\" not supported, only approved is supported.",
		},
		{
			description: "apply reqs with valid",
			input: raw.Project{
				Dir:               String("."),
				ApplyRequirements: []string{"approved"},
			},
			expErr: "",
		},
		{
			description: "empty tf version string",
			input: raw.Project{
				Dir:              String("."),
				TerraformVersion: String(""),
			},
			expErr: "terraform_version: version \"\" could not be parsed: Malformed version: .",
		},
		{
			description: "tf version with v prepended",
			input: raw.Project{
				Dir:              String("."),
				TerraformVersion: String("v1"),
			},
			expErr: "",
		},
		{
			description: "tf version without prepended v",
			input: raw.Project{
				Dir:              String("."),
				TerraformVersion: String("1"),
			},
			expErr: "",
		},
		{
			description: "empty string for project name",
			input: raw.Project{
				Dir:  String("."),
				Name: String(""),
			},
			expErr: "name: if set cannot be empty.",
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

func TestProject_ToValid(t *testing.T) {
	tfVersionPointEleven, _ := version.NewVersion("v0.11.0")
	cases := []struct {
		description string
		input       raw.Project
		exp         valid.Project
	}{
		{
			description: "minimal values",
			input: raw.Project{
				Dir: String("."),
			},
			exp: valid.Project{
				Dir:              ".",
				Workspace:        "default",
				Workflow:         nil,
				TerraformVersion: nil,
				Autoplan: valid.Autoplan{
					WhenModified: []string{"**/*.tf*"},
					Enabled:      true,
				},
				ApplyRequirements: nil,
				Name:              nil,
			},
		},
		{
			description: "all set",
			input: raw.Project{
				Dir:              String("."),
				Workspace:        String("myworkspace"),
				Workflow:         String("myworkflow"),
				TerraformVersion: String("v0.11.0"),
				Autoplan: &raw.Autoplan{
					WhenModified: []string{"hi"},
					Enabled:      Bool(false),
				},
				ApplyRequirements: []string{"approved"},
				Name:              String("myname"),
			},
			exp: valid.Project{
				Dir:              ".",
				Workspace:        "myworkspace",
				Workflow:         String("myworkflow"),
				TerraformVersion: tfVersionPointEleven,
				Autoplan: valid.Autoplan{
					WhenModified: []string{"hi"},
					Enabled:      false,
				},
				ApplyRequirements: []string{"approved"},
				Name:              String("myname"),
			},
		},
		{
			description: "tf version without 'v'",
			input: raw.Project{
				Dir:              String("."),
				TerraformVersion: String("0.11.0"),
			},
			exp: valid.Project{
				Dir:              ".",
				Workspace:        "default",
				TerraformVersion: tfVersionPointEleven,
				Autoplan: valid.Autoplan{
					WhenModified: []string{"**/*.tf*"},
					Enabled:      true,
				},
			},
		},
		{
			description: "dir with /",
			input: raw.Project{
				Dir: String("/"),
			},
			exp: valid.Project{
				Dir:       ".",
				Workspace: "default",
				Autoplan: valid.Autoplan{
					WhenModified: []string{"**/*.tf*"},
					Enabled:      true,
				},
			},
		},
		{
			description: "dir with trailing slash",
			input: raw.Project{
				Dir: String("mydir/"),
			},
			exp: valid.Project{
				Dir:       "mydir",
				Workspace: "default",
				Autoplan: valid.Autoplan{
					WhenModified: []string{"**/*.tf*"},
					Enabled:      true,
				},
			},
		},
		{
			description: "unclean dir",
			input: raw.Project{
				// This won't actually be allowed since it doesn't validate.
				Dir: String("./mydir/anotherdir/../"),
			},
			exp: valid.Project{
				Dir:       "mydir",
				Workspace: "default",
				Autoplan: valid.Autoplan{
					WhenModified: []string{"**/*.tf*"},
					Enabled:      true,
				},
			},
		},
		{
			description: "workspace set to empty string",
			input: raw.Project{
				Dir:       String("."),
				Workspace: String(""),
			},
			exp: valid.Project{
				Dir:       ".",
				Workspace: "default",
				Autoplan: valid.Autoplan{
					WhenModified: []string{"**/*.tf*"},
					Enabled:      true,
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
