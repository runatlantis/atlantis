// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

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
		defaultAutoDiscover valid.AutoDiscoverMode
		projects            []valid.Project
		expEnabled          bool
	}{
		{
			description:         "default enabled with no projects",
			defaultAutoDiscover: valid.AutoDiscoverEnabledMode,
			expEnabled:          true,
		},
		{
			description:         "default disabled with no projects",
			defaultAutoDiscover: valid.AutoDiscoverDisabledMode,
			expEnabled:          false,
		},
		{
			description:         "auto mode with no projects",
			defaultAutoDiscover: valid.AutoDiscoverAutoMode,
			expEnabled:          true,
		},
		{
			description:         "default enabled with projects",
			defaultAutoDiscover: valid.AutoDiscoverEnabledMode,
			projects:            []valid.Project{{}},
			expEnabled:          true,
		},
		{
			description:         "default disabled with projects",
			defaultAutoDiscover: valid.AutoDiscoverDisabledMode,
			projects:            []valid.Project{{}},
			expEnabled:          false,
		},
		{
			description:         "auto mode with projects",
			defaultAutoDiscover: valid.AutoDiscoverAutoMode,
			projects:            []valid.Project{{}},
			expEnabled:          false,
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			r := valid.RepoCfg{
				Projects:     c.projects,
				AutoDiscover: nil,
			}
			enabled := r.AutoDiscoverEnabled(c.defaultAutoDiscover)
			Equals(t, c.expEnabled, enabled)
		})
	}
}

func TestConfig_FindProjectsByDirPattern(t *testing.T) {
	cases := []struct {
		description string
		pattern     string
		projects    []valid.Project
		expProjects []valid.Project
	}{
		{
			description: "simple wildcard matches multiple projects",
			pattern:     "modules/*",
			projects: []valid.Project{
				{Dir: "modules/vpc", Workspace: "default"},
				{Dir: "modules/rds", Workspace: "default"},
				{Dir: "apps/api", Workspace: "default"},
			},
			expProjects: []valid.Project{
				{Dir: "modules/vpc", Workspace: "default"},
				{Dir: "modules/rds", Workspace: "default"},
			},
		},
		{
			description: "double star matches nested directories",
			pattern:     "environments/**",
			projects: []valid.Project{
				{Dir: "environments/prod/app", Workspace: "default"},
				{Dir: "environments/staging/app", Workspace: "default"},
				{Dir: "environments/dev", Workspace: "default"},
				{Dir: "modules/vpc", Workspace: "default"},
			},
			expProjects: []valid.Project{
				{Dir: "environments/prod/app", Workspace: "default"},
				{Dir: "environments/staging/app", Workspace: "default"},
				{Dir: "environments/dev", Workspace: "default"},
			},
		},
		{
			description: "question mark matches single character",
			pattern:     "env?/*",
			projects: []valid.Project{
				{Dir: "env1/app", Workspace: "default"},
				{Dir: "env2/app", Workspace: "default"},
				{Dir: "envX/app", Workspace: "default"},
				{Dir: "environment/app", Workspace: "default"},
			},
			expProjects: []valid.Project{
				{Dir: "env1/app", Workspace: "default"},
				{Dir: "env2/app", Workspace: "default"},
				{Dir: "envX/app", Workspace: "default"},
			},
		},
		{
			description: "character class matches specific characters",
			pattern:     "env[0-9]/*",
			projects: []valid.Project{
				{Dir: "env1/app", Workspace: "default"},
				{Dir: "env2/app", Workspace: "default"},
				{Dir: "envX/app", Workspace: "default"},
			},
			expProjects: []valid.Project{
				{Dir: "env1/app", Workspace: "default"},
				{Dir: "env2/app", Workspace: "default"},
			},
		},
		{
			description: "no matches returns empty slice",
			pattern:     "nonexistent/*",
			projects: []valid.Project{
				{Dir: "modules/vpc", Workspace: "default"},
			},
			expProjects: nil,
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			r := valid.RepoCfg{
				Projects: c.projects,
			}
			projects := r.FindProjectsByDirPattern(c.pattern)
			Equals(t, c.expProjects, projects)
		})
	}
}

func TestConfig_FindProjectsByDirPatternWorkspace(t *testing.T) {
	cases := []struct {
		description string
		pattern     string
		workspace   string
		projects    []valid.Project
		expProjects []valid.Project
	}{
		{
			description: "matches pattern and workspace",
			pattern:     "modules/*",
			workspace:   "default",
			projects: []valid.Project{
				{Dir: "modules/vpc", Workspace: "default"},
				{Dir: "modules/vpc", Workspace: "staging"},
				{Dir: "modules/rds", Workspace: "default"},
			},
			expProjects: []valid.Project{
				{Dir: "modules/vpc", Workspace: "default"},
				{Dir: "modules/rds", Workspace: "default"},
			},
		},
		{
			description: "workspace filter excludes non-matching",
			pattern:     "modules/*",
			workspace:   "production",
			projects: []valid.Project{
				{Dir: "modules/vpc", Workspace: "default"},
				{Dir: "modules/vpc", Workspace: "staging"},
			},
			expProjects: nil,
		},
		{
			description: "double star with workspace filter",
			pattern:     "environments/**",
			workspace:   "staging",
			projects: []valid.Project{
				{Dir: "environments/us-east/app", Workspace: "staging"},
				{Dir: "environments/us-west/app", Workspace: "production"},
				{Dir: "environments/eu/app", Workspace: "staging"},
			},
			expProjects: []valid.Project{
				{Dir: "environments/us-east/app", Workspace: "staging"},
				{Dir: "environments/eu/app", Workspace: "staging"},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			r := valid.RepoCfg{
				Projects: c.projects,
			}
			projects := r.FindProjectsByDirPatternWorkspace(c.pattern, c.workspace)
			Equals(t, c.expProjects, projects)
		})
	}
}

func TestContainsDirGlobPattern(t *testing.T) {
	cases := []struct {
		input    string
		expected bool
	}{
		{"modules/*", true},
		{"modules/**", true},
		{"env?/app", true},
		{"env[0-9]/app", true},
		{"modules/vpc", false},
		{".", false},
		{"path/to/dir", false},
	}
	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			result := valid.ContainsDirGlobPattern(c.input)
			Equals(t, c.expected, result)
		})
	}
}
