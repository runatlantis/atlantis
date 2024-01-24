package raw_test

import (
	"testing"

	validation "github.com/go-ozzo/ozzo-validation"
	version "github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/core/config/raw"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	. "github.com/runatlantis/atlantis/testing"
	yaml "gopkg.in/yaml.v3"
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
				Dir:                nil,
				Workspace:          nil,
				Workflow:           nil,
				TerraformVersion:   nil,
				Autoplan:           nil,
				PlanRequirements:   nil,
				ApplyRequirements:  nil,
				ImportRequirements: nil,
				Name:               nil,
				Branch:             nil,
			},
		},
		{
			description: "all fields set including mergeable apply requirement",
			input: `
name: myname
branch: mybranch
dir: mydir
workspace: workspace
workflow: workflow
terraform_version: v0.11.0
autoplan:
  when_modified: []
  enabled: false
plan_requirements:
- mergeable
apply_requirements:
- mergeable
import_requirements:
- mergeable
execution_order_group: 10`,
			exp: raw.Project{
				Name:             String("myname"),
				Branch:           String("mybranch"),
				Dir:              String("mydir"),
				Workspace:        String("workspace"),
				Workflow:         String("workflow"),
				TerraformVersion: String("v0.11.0"),
				Autoplan: &raw.Autoplan{
					WhenModified: []string{},
					Enabled:      Bool(false),
				},
				PlanRequirements:    []string{"mergeable"},
				ApplyRequirements:   []string{"mergeable"},
				ImportRequirements:  []string{"mergeable"},
				ExecutionOrderGroup: Int(10),
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
			description: "not a regexp for branch",
			input: raw.Project{
				Branch: String("text"),
				Dir:    String("."),
			},
			expErr: "branch: regex must begin and end with a slash '/'.",
		},
		{
			description: "invalid regexp for branch",
			input: raw.Project{
				Branch: String("/(text/"),
				Dir:    String("."),
			},
			expErr: "branch: parsing: /(text/: error parsing regexp: missing closing ): `(text`.",
		},
		{
			description: "plan reqs with unsupported",
			input: raw.Project{
				Dir:              String("."),
				PlanRequirements: []string{"unsupported"},
			},
			expErr: "plan_requirements: \"unsupported\" is not a valid plan_requirement, only \"approved\", \"mergeable\" and \"undiverged\" are supported.",
		},
		{
			description: "plan reqs with undiverged, mergeable and approved requirements",
			input: raw.Project{
				Dir:              String("."),
				PlanRequirements: []string{"undiverged", "mergeable", "approved"},
			},
			expErr: "",
		},
		{
			description: "plan reqs with approved requirement",
			input: raw.Project{
				Dir:              String("."),
				PlanRequirements: []string{"approved"},
			},
			expErr: "",
		},
		{
			description: "plan reqs with mergeable requirement",
			input: raw.Project{
				Dir:              String("."),
				PlanRequirements: []string{"mergeable"},
			},
			expErr: "",
		},
		{
			description: "plan reqs with mergeable and approved requirements",
			input: raw.Project{
				Dir:              String("."),
				PlanRequirements: []string{"mergeable", "approved"},
			},
			expErr: "",
		},
		{
			description: "apply reqs with unsupported",
			input: raw.Project{
				Dir:               String("."),
				ApplyRequirements: []string{"unsupported"},
			},
			expErr: "apply_requirements: \"unsupported\" is not a valid apply_requirement, only \"approved\", \"mergeable\" and \"undiverged\" are supported.",
		},
		{
			description: "apply reqs with approved requirement",
			input: raw.Project{
				Dir:               String("."),
				ApplyRequirements: []string{"approved"},
			},
			expErr: "",
		},
		{
			description: "apply reqs with mergeable requirement",
			input: raw.Project{
				Dir:               String("."),
				ApplyRequirements: []string{"mergeable"},
			},
			expErr: "",
		},
		{
			description: "apply reqs with undiverged requirement",
			input: raw.Project{
				Dir:               String("."),
				ApplyRequirements: []string{"undiverged"},
			},
			expErr: "",
		},
		{
			description: "apply reqs with mergeable and approved requirements",
			input: raw.Project{
				Dir:               String("."),
				ApplyRequirements: []string{"mergeable", "approved"},
			},
			expErr: "",
		},
		{
			description: "apply reqs with undiverged and approved requirements",
			input: raw.Project{
				Dir:               String("."),
				ApplyRequirements: []string{"undiverged", "approved"},
			},
			expErr: "",
		},
		{
			description: "apply reqs with undiverged and mergeable requirements",
			input: raw.Project{
				Dir:               String("."),
				ApplyRequirements: []string{"undiverged", "mergeable"},
			},
			expErr: "",
		},
		{
			description: "apply reqs with undiverged, mergeable and approved requirements",
			input: raw.Project{
				Dir:               String("."),
				ApplyRequirements: []string{"undiverged", "mergeable", "approved"},
			},
			expErr: "",
		},
		{
			description: "import reqs with unsupported",
			input: raw.Project{
				Dir:                String("."),
				ImportRequirements: []string{"unsupported"},
			},
			expErr: "import_requirements: \"unsupported\" is not a valid import_requirement, only \"approved\", \"mergeable\" and \"undiverged\" are supported.",
		},
		{
			description: "import reqs with undiverged, mergeable and approved requirements",
			input: raw.Project{
				Dir:                String("."),
				ImportRequirements: []string{"undiverged", "mergeable", "approved"},
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
		{
			description: "project name with slashes",
			input: raw.Project{
				Dir:  String("."),
				Name: String("my/name"),
			},
			expErr: "",
		},
		{
			description: "project name with emoji",
			input: raw.Project{
				Dir:  String("."),
				Name: String("😀"),
			},
			expErr: "name: \"😀\" is not allowed: must contain only URL safe characters.",
		},
		{
			description: "project name with spaces",
			input: raw.Project{
				Dir:  String("."),
				Name: String("name with spaces"),
			},
			expErr: "name: \"name with spaces\" is not allowed: must contain only URL safe characters.",
		},
		{
			description: "project name with +",
			input: raw.Project{
				Dir:  String("."),
				Name: String("namewith+"),
			},
			expErr: "name: \"namewith+\" is not allowed: must contain only URL safe characters.",
		},
		{
			description: `project name with \`,
			input: raw.Project{
				Dir:  String("."),
				Name: String(`namewith\`),
			},
			expErr: `name: "namewith\\" is not allowed: must contain only URL safe characters.`,
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
				BranchRegex:      nil,
				Workspace:        "default",
				WorkflowName:     nil,
				TerraformVersion: nil,
				Autoplan: valid.Autoplan{
					WhenModified: raw.DefaultAutoPlanWhenModified,
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
				ApplyRequirements:   []string{"approved"},
				Name:                String("myname"),
				ExecutionOrderGroup: Int(10),
			},
			exp: valid.Project{
				Dir:              ".",
				Workspace:        "myworkspace",
				WorkflowName:     String("myworkflow"),
				TerraformVersion: tfVersionPointEleven,
				Autoplan: valid.Autoplan{
					WhenModified: []string{"hi"},
					Enabled:      false,
				},
				ApplyRequirements:   []string{"approved"},
				Name:                String("myname"),
				ExecutionOrderGroup: 10,
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
					WhenModified: raw.DefaultAutoPlanWhenModified,
					Enabled:      true,
				},
			},
		},
		// Directories.
		{
			description: "dir set to /",
			input: raw.Project{
				Dir: String("/"),
			},
			exp: valid.Project{
				Dir:       ".",
				Workspace: "default",
				Autoplan: valid.Autoplan{
					WhenModified: raw.DefaultAutoPlanWhenModified,
					Enabled:      true,
				},
			},
		},
		{
			description: "dir starting with /",
			input: raw.Project{
				Dir: String("/a/b/c"),
			},
			exp: valid.Project{
				Dir:       "a/b/c",
				Workspace: "default",
				Autoplan: valid.Autoplan{
					WhenModified: raw.DefaultAutoPlanWhenModified,
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
					WhenModified: raw.DefaultAutoPlanWhenModified,
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
					WhenModified: raw.DefaultAutoPlanWhenModified,
					Enabled:      true,
				},
			},
		},
		{
			description: "dir set to ./",
			input: raw.Project{
				Dir: String("./"),
			},
			exp: valid.Project{
				Dir:       ".",
				Workspace: "default",
				Autoplan: valid.Autoplan{
					WhenModified: raw.DefaultAutoPlanWhenModified,
					Enabled:      true,
				},
			},
		},
		{
			description: "dir set to ././",
			input: raw.Project{
				Dir: String("././"),
			},
			exp: valid.Project{
				Dir:       ".",
				Workspace: "default",
				Autoplan: valid.Autoplan{
					WhenModified: raw.DefaultAutoPlanWhenModified,
					Enabled:      true,
				},
			},
		},
		{
			description: "dir set to .",
			input: raw.Project{
				Dir: String("."),
			},
			exp: valid.Project{
				Dir:       ".",
				Workspace: "default",
				Autoplan: valid.Autoplan{
					WhenModified: raw.DefaultAutoPlanWhenModified,
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
					WhenModified: raw.DefaultAutoPlanWhenModified,
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
