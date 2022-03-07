package valid_test

import (
	"testing"

	validation "github.com/go-ozzo/ozzo-validation"
	version "github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	. "github.com/runatlantis/atlantis/testing"
)

func TestConfig_FindProjectsByDir(t *testing.T) {
	tfVersion, _ := version.NewVersion("v0.11.0")
	cases := []struct {
		description string
		nameRegex   string
		input       valid.RepoCfg
		expProjects []valid.Project
	}{
		{
			description: "Find projects with 'dev' prefix as allowed prefix",
			nameRegex:   "dev.*",
			input: valid.RepoCfg{
				Version: 3,
				Projects: []valid.Project{
					{
						Dir:              ".",
						Name:             String("dev_terragrunt_myproject"),
						Workspace:        "myworkspace",
						TerraformVersion: tfVersion,
						Autoplan: valid.Autoplan{
							WhenModified: []string{"**/*.tf*", "**/terragrunt.hcl"},
							Enabled:      false,
						},
						ApplyRequirements: []string{"approved"},
					},
				},
				Workflows: map[string]valid.Workflow{
					"myworkflow": {
						Name:        "myworkflow",
						Apply:       valid.DefaultApplyStage,
						Plan:        valid.DefaultPlanStage,
						PolicyCheck: valid.DefaultPolicyCheckStage,
					},
				},
				AllowedRegexpPrefixes: []string{"dev", "staging"},
			},
			expProjects: []valid.Project{
				{
					Dir:              ".",
					Name:             String("dev_terragrunt_myproject"),
					Workspace:        "myworkspace",
					TerraformVersion: tfVersion,
					Autoplan: valid.Autoplan{
						WhenModified: []string{"**/*.tf*", "**/terragrunt.hcl"},
						Enabled:      false,
					},
					ApplyRequirements: []string{"approved"},
				},
			},
		},
		{
			description: "Only find projects with allowed prefix",
			nameRegex:   ".*",
			input: valid.RepoCfg{
				Version: 3,
				Projects: []valid.Project{
					{
						Dir:              ".",
						Name:             String("dev_terragrunt_myproject"),
						Workspace:        "myworkspace",
						TerraformVersion: tfVersion,
						Autoplan: valid.Autoplan{
							WhenModified: []string{"**/*.tf*", "**/terragrunt.hcl"},
							Enabled:      false,
						},
						ApplyRequirements: []string{"approved"},
					},
					{
						Dir:              ".",
						Name:             String("staging_terragrunt_myproject"),
						Workspace:        "myworkspace",
						TerraformVersion: tfVersion,
						Autoplan: valid.Autoplan{
							WhenModified: []string{"**/*.tf*", "**/terragrunt.hcl"},
							Enabled:      false,
						},
						ApplyRequirements: []string{"approved"},
					},
				},
				Workflows: map[string]valid.Workflow{
					"myworkflow": {
						Name:        "myworkflow",
						Apply:       valid.DefaultApplyStage,
						Plan:        valid.DefaultPlanStage,
						PolicyCheck: valid.DefaultPolicyCheckStage,
					},
				},
				AllowedRegexpPrefixes: []string{"dev", "staging"},
			},
			expProjects: nil,
		},
		{
			description: "Find all projects without restrictions of allowed prefix",
			nameRegex:   ".*",
			input: valid.RepoCfg{
				Version: 3,
				Projects: []valid.Project{
					{
						Dir:              ".",
						Name:             String("dev_terragrunt_myproject"),
						Workspace:        "myworkspace",
						TerraformVersion: tfVersion,
						Autoplan: valid.Autoplan{
							WhenModified: []string{"**/*.tf*", "**/terragrunt.hcl"},
							Enabled:      false,
						},
						ApplyRequirements: []string{"approved"},
					},
					{
						Dir:              ".",
						Name:             String("staging_terragrunt_myproject"),
						Workspace:        "myworkspace",
						TerraformVersion: tfVersion,
						Autoplan: valid.Autoplan{
							WhenModified: []string{"**/*.tf*", "**/terragrunt.hcl"},
							Enabled:      false,
						},
						ApplyRequirements: []string{"approved"},
					},
				},
				Workflows: map[string]valid.Workflow{
					"myworkflow": {
						Name:        "myworkflow",
						Apply:       valid.DefaultApplyStage,
						Plan:        valid.DefaultPlanStage,
						PolicyCheck: valid.DefaultPolicyCheckStage,
					},
				},
				AllowedRegexpPrefixes: nil,
			},
			expProjects: []valid.Project{
				{
					Dir:              ".",
					Name:             String("dev_terragrunt_myproject"),
					Workspace:        "myworkspace",
					TerraformVersion: tfVersion,
					Autoplan: valid.Autoplan{
						WhenModified: []string{"**/*.tf*", "**/terragrunt.hcl"},
						Enabled:      false,
					},
					ApplyRequirements: []string{"approved"},
				},
				{
					Dir:              ".",
					Name:             String("staging_terragrunt_myproject"),
					Workspace:        "myworkspace",
					TerraformVersion: tfVersion,
					Autoplan: valid.Autoplan{
						WhenModified: []string{"**/*.tf*", "**/terragrunt.hcl"},
						Enabled:      false,
					},
					ApplyRequirements: []string{"approved"},
				},
			},
		},
		{
			description: "Always find exact matches even if the prefix is not allowed",
			nameRegex:   ".*",
			input: valid.RepoCfg{
				Version: 3,
				Projects: []valid.Project{
					{
						Dir:              ".",
						Name:             String("prod_terragrunt_myproject"),
						Workspace:        "myworkspace",
						TerraformVersion: tfVersion,
						Autoplan: valid.Autoplan{
							WhenModified: []string{"**/*.tf*", "**/terragrunt.hcl"},
							Enabled:      false,
						},
						ApplyRequirements: []string{"approved"},
					},
				},
				Workflows: map[string]valid.Workflow{
					"myworkflow": {
						Name:        "myworkflow",
						Apply:       valid.DefaultApplyStage,
						Plan:        valid.DefaultPlanStage,
						PolicyCheck: valid.DefaultPolicyCheckStage,
					},
				},
				AllowedRegexpPrefixes: []string{"dev", "staging"},
			},
			expProjects: []valid.Project{
				{
					Dir:              ".",
					Name:             String("prod_terragrunt_myproject"),
					Workspace:        "myworkspace",
					TerraformVersion: tfVersion,
					Autoplan: valid.Autoplan{
						WhenModified: []string{"**/*.tf*", "**/terragrunt.hcl"},
						Enabled:      false,
					},
					ApplyRequirements: []string{"approved"},
				},
			},
		},
	}
	validation.ErrorTag = "yaml"
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			projects := c.input.FindProjectsByName(c.nameRegex)
			Equals(t, c.expProjects, projects)
		})
	}
}
