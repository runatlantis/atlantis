package valid_test

import (
	"testing"

	validation "github.com/go-ozzo/ozzo-validation"
	version "github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/core/config/raw"
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
							WhenModified: raw.DefaultAutoPlanWhenModified,
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
						WhenModified: raw.DefaultAutoPlanWhenModified,
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
							WhenModified: raw.DefaultAutoPlanWhenModified,
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
							WhenModified: raw.DefaultAutoPlanWhenModified,
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
							WhenModified: raw.DefaultAutoPlanWhenModified,
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
							WhenModified: raw.DefaultAutoPlanWhenModified,
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
						WhenModified: raw.DefaultAutoPlanWhenModified,
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
						WhenModified: raw.DefaultAutoPlanWhenModified,
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
							WhenModified: raw.DefaultAutoPlanWhenModified,
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
						WhenModified: raw.DefaultAutoPlanWhenModified,
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

func TestConfig_AutoDiscoverEnabled(t *testing.T) {
	cases := []struct {
		description         string
		repoAutoDiscover    valid.AutoDiscoverMode
		defaultAutoDiscover valid.AutoDiscoverMode
		projects            []valid.Project
		expEnabled          bool
	}{
		{
			description:         "repo disabled autodiscover default enabled",
			repoAutoDiscover:    valid.AutoDiscoverDisabledMode,
			defaultAutoDiscover: valid.AutoDiscoverEnabledMode,
			expEnabled:          false,
		},
		{
			description:         "repo disabled autodiscover default disabled",
			repoAutoDiscover:    valid.AutoDiscoverDisabledMode,
			defaultAutoDiscover: valid.AutoDiscoverDisabledMode,
			expEnabled:          false,
		},
		{
			description:         "repo enabled autodiscover default enabled",
			repoAutoDiscover:    valid.AutoDiscoverEnabledMode,
			defaultAutoDiscover: valid.AutoDiscoverEnabledMode,
			expEnabled:          true,
		},
		{
			description:         "repo enabled autodiscover default disabled",
			repoAutoDiscover:    valid.AutoDiscoverEnabledMode,
			defaultAutoDiscover: valid.AutoDiscoverDisabledMode,
			expEnabled:          true,
		},
		{
			description:         "repo set auto autodiscover with no projects default enabled",
			repoAutoDiscover:    valid.AutoDiscoverAutoMode,
			defaultAutoDiscover: valid.AutoDiscoverEnabledMode,
			expEnabled:          true,
		},
		{
			description:         "repo set auto autodiscover with no projects default disabled",
			repoAutoDiscover:    valid.AutoDiscoverAutoMode,
			defaultAutoDiscover: valid.AutoDiscoverDisabledMode,
			expEnabled:          true,
		},
		{
			description:         "repo set auto autodiscover with a project default enabled",
			repoAutoDiscover:    valid.AutoDiscoverAutoMode,
			defaultAutoDiscover: valid.AutoDiscoverEnabledMode,
			projects:            []valid.Project{{}},
			expEnabled:          false,
		},
		{
			description:         "repo set auto autodiscover with a project default disabled",
			repoAutoDiscover:    valid.AutoDiscoverAutoMode,
			defaultAutoDiscover: valid.AutoDiscoverDisabledMode,
			projects:            []valid.Project{{}},
			expEnabled:          false,
		},
		{
			description:         "repo unset autodiscover with no projects default enabled",
			defaultAutoDiscover: valid.AutoDiscoverEnabledMode,
			expEnabled:          true,
		},
		{
			description:         "repo unset autodiscover with no projects default disabled",
			defaultAutoDiscover: valid.AutoDiscoverDisabledMode,
			expEnabled:          false,
		},
		{
			description:         "repo unset autodiscover with no projects default auto",
			defaultAutoDiscover: valid.AutoDiscoverAutoMode,
			expEnabled:          true,
		},
		{
			description:         "repo unset autodiscover with a project default enabled",
			projects:            []valid.Project{{}},
			defaultAutoDiscover: valid.AutoDiscoverEnabledMode,
			expEnabled:          true,
		},
		{
			description:         "repo unset autodiscover with a project default disabled",
			projects:            []valid.Project{{}},
			defaultAutoDiscover: valid.AutoDiscoverDisabledMode,
			expEnabled:          false,
		},
		{
			description:         "repo unset autodiscover with a project default auto",
			projects:            []valid.Project{{}},
			defaultAutoDiscover: valid.AutoDiscoverAutoMode,
			expEnabled:          false,
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			r := valid.RepoCfg{
				Projects:     c.projects,
				AutoDiscover: nil,
			}
			if c.repoAutoDiscover != "" {
				r.AutoDiscover = &valid.AutoDiscover{c.repoAutoDiscover}
			}
			enabled := r.AutoDiscoverEnabled(c.defaultAutoDiscover)
			Equals(t, c.expEnabled, enabled)
		})
	}
}
