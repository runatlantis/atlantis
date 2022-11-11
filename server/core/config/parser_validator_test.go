package config_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/server/core/config/raw"

	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/core/config"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	. "github.com/runatlantis/atlantis/testing"
)

var globalCfg = valid.GlobalCfg{
	Repos: []valid.Repo{
		{
			IDRegex:              regexp.MustCompile(".*"),
			AllowCustomWorkflows: Bool(true),
			AllowedOverrides:     []string{"apply_requirements", "workflow"},
			CheckoutStrategy:     "branch",
		},
	},
}

func TestHasRepoCfg_DirDoesNotExist(t *testing.T) {
	r := config.ParserValidator{}
	exists, err := r.HasRepoCfg("/not/exist")
	Ok(t, err)
	Equals(t, false, exists)
}

func TestHasRepoCfg_FileDoesNotExist(t *testing.T) {
	tmpDir, cleanup := TempDir(t)
	defer cleanup()
	r := config.ParserValidator{}
	exists, err := r.HasRepoCfg(tmpDir)
	Ok(t, err)
	Equals(t, false, exists)
}

func TestHasRepoCfg_InvalidFileExtension(t *testing.T) {
	tmpDir, cleanup := TempDir(t)
	defer cleanup()
	_, err := os.Create(filepath.Join(tmpDir, "atlantis.yml"))
	Ok(t, err)

	r := config.ParserValidator{}
	_, err = r.HasRepoCfg(tmpDir)
	ErrContains(t, "found \"atlantis.yml\" as config file; rename using the .yaml extension - \"atlantis.yaml\"", err)
}

func TestParseRepoCfg_DirDoesNotExist(t *testing.T) {
	r := config.ParserValidator{}
	_, err := r.ParseRepoCfg("/not/exist", globalCfg, "")
	Assert(t, os.IsNotExist(err), "exp not exist err")
}

func TestParseRepoCfg_FileDoesNotExist(t *testing.T) {
	tmpDir, cleanup := TempDir(t)
	defer cleanup()
	r := config.ParserValidator{}
	_, err := r.ParseRepoCfg(tmpDir, globalCfg, "")
	Assert(t, os.IsNotExist(err), "exp not exist err")
}

func TestParseRepoCfg_BadPermissions(t *testing.T) {
	tmpDir, cleanup := TempDir(t)
	defer cleanup()
	err := os.WriteFile(filepath.Join(tmpDir, "atlantis.yaml"), nil, 0000)
	Ok(t, err)

	r := config.ParserValidator{}
	_, err = r.ParseRepoCfg(tmpDir, globalCfg, "")
	ErrContains(t, "unable to read atlantis.yaml file: ", err)
}

// Test both ParseRepoCfg and ParseGlobalCfg when given in valid YAML.
// We only have a few cases here because we assume the YAML library to be
// well tested. See https://github.com/go-yaml/yaml/blob/v2/decode_test.go#L810.
func TestParseCfgs_InvalidYAML(t *testing.T) {
	cases := []struct {
		description string
		input       string
		expErr      string
	}{
		{
			"random characters",
			"slkjds",
			"yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `slkjds` into",
		},
		{
			"just a colon",
			":",
			"yaml: did not find expected key",
		},
	}

	tmpDir, cleanup := TempDir(t)
	defer cleanup()

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			confPath := filepath.Join(tmpDir, "atlantis.yaml")
			err := os.WriteFile(confPath, []byte(c.input), 0600)
			Ok(t, err)
			r := config.ParserValidator{}
			_, err = r.ParseRepoCfg(tmpDir, globalCfg, "")
			ErrContains(t, c.expErr, err)
			_, err = r.ParseGlobalCfg(confPath, valid.NewGlobalCfg("somedir"))
			ErrContains(t, c.expErr, err)
		})
	}
}

func TestParseRepoCfg(t *testing.T) {
	tfVersion, _ := version.NewVersion("v0.11.0")
	cases := []struct {
		description string
		input       string
		expErr      string
		exp         valid.RepoCfg
	}{
		// Version key.
		{
			description: "no version",
			input: `
projects:
- dir: "."
`,
			expErr: "version: is required. If you've just upgraded Atlantis you need to rewrite your atlantis.yaml for version 3. See www.runatlantis.io/docs/upgrading-atlantis-yaml.html.",
		},
		{
			description: "unsupported version",
			input: `
version: 0
projects:
- dir: "."
`,
			expErr: "version: only versions 2 and 3 are supported.",
		},
		{
			description: "empty version",
			input: `
version:
projects:
- dir: "."
`,
			expErr: "version: is required. If you've just upgraded Atlantis you need to rewrite your atlantis.yaml for version 3. See www.runatlantis.io/docs/upgrading-atlantis-yaml.html.",
		},
		{
			description: "version 2",
			input: `
version: 2
workflows:
  custom:
    plan:
      steps:
      - run: old 'shell parsing'
`,
			exp: valid.RepoCfg{
				Version: 2,
				Workflows: map[string]valid.Workflow{
					"custom": {
						Name:        "custom",
						Apply:       valid.DefaultApplyStage,
						PolicyCheck: valid.DefaultPolicyCheckStage,
						Plan: valid.Stage{
							Steps: []valid.Step{
								{
									StepName:   "run",
									RunCommand: "old shell parsing",
								},
							},
						},
					},
				},
			},
		},

		// Projects key.
		{
			description: "empty projects list",
			input: `
version: 3
projects:`,
			exp: valid.RepoCfg{
				Version:   3,
				Projects:  nil,
				Workflows: map[string]valid.Workflow{},
			},
		},
		{
			description: "project dir not set",
			input: `
version: 3
projects:
- `,
			expErr: "projects: (0: (dir: cannot be blank.).).",
		},
		{
			description: "project dir set",
			input: `
version: 3
projects:
- dir: .`,
			exp: valid.RepoCfg{
				Version: 3,
				Projects: []valid.Project{
					{
						Dir:              ".",
						Workspace:        "default",
						WorkflowName:     nil,
						TerraformVersion: nil,
						Autoplan: valid.Autoplan{
							WhenModified: []string{"**/*.tf*", "**/terragrunt.hcl"},
							Enabled:      true,
						},
						ApplyRequirements: nil,
					},
				},
				Workflows: map[string]valid.Workflow{},
			},
		},
		{
			description: "autoplan should be enabled by default",
			input: `
version: 3
projects:
- dir: "."
`,
			exp: valid.RepoCfg{
				Version: 3,
				Projects: []valid.Project{
					{
						Dir:       ".",
						Workspace: "default",
						Autoplan: valid.Autoplan{
							WhenModified: []string{"**/*.tf*", "**/terragrunt.hcl"},
							Enabled:      true,
						},
					},
				},
				Workflows: make(map[string]valid.Workflow),
			},
		},
		{
			description: "autoplan should be enabled if only when_modified set",
			input: `
version: 3
projects:
- dir: "."
  autoplan:
    when_modified: ["**/*.tf*"]
`,
			exp: valid.RepoCfg{
				Version: 3,
				Projects: []valid.Project{
					{
						Dir:       ".",
						Workspace: "default",
						Autoplan: valid.Autoplan{
							WhenModified: []string{"**/*.tf*"},
							Enabled:      true,
						},
					},
				},
				Workflows: make(map[string]valid.Workflow),
			},
		},
		{
			description: "if workflows not defined there are none",
			input: `
version: 3
projects:
- dir: "."
`,
			exp: valid.RepoCfg{
				Version: 3,
				Projects: []valid.Project{
					{
						Dir:       ".",
						Workspace: "default",
						Autoplan: valid.Autoplan{
							WhenModified: []string{"**/*.tf*", "**/terragrunt.hcl"},
							Enabled:      true,
						},
					},
				},
				Workflows: make(map[string]valid.Workflow),
			},
		},
		{
			description: "if workflows key set but with no workflows there are none",
			input: `
version: 3
projects:
- dir: "."
workflows: ~
`,
			exp: valid.RepoCfg{
				Version: 3,
				Projects: []valid.Project{
					{
						Dir:       ".",
						Workspace: "default",
						Autoplan: valid.Autoplan{
							WhenModified: []string{"**/*.tf*", "**/terragrunt.hcl"},
							Enabled:      true,
						},
					},
				},
				Workflows: make(map[string]valid.Workflow),
			},
		},
		{
			description: "if a plan or apply explicitly defines an empty steps key then it gets the defaults",
			input: `
version: 3
projects:
- dir: "."
workflows:
  default:
    plan:
      steps:
    apply:
      steps:
`,
			exp: valid.RepoCfg{
				Version: 3,
				Projects: []valid.Project{
					{
						Dir:       ".",
						Workspace: "default",
						Autoplan: valid.Autoplan{
							WhenModified: []string{"**/*.tf*", "**/terragrunt.hcl"},
							Enabled:      true,
						},
					},
				},
				Workflows: map[string]valid.Workflow{
					"default": {
						Name:        "default",
						Plan:        valid.DefaultPlanStage,
						Apply:       valid.DefaultApplyStage,
						PolicyCheck: valid.DefaultPolicyCheckStage,
					},
				},
			},
		},
		{
			description: "project fields set except autoplan",
			input: `
version: 3
projects:
- dir: .
  workspace: myworkspace
  terraform_version: v0.11.0
  apply_requirements: [approved]
  workflow: myworkflow
workflows:
  myworkflow: ~`,
			exp: valid.RepoCfg{
				Version: 3,
				Projects: []valid.Project{
					{
						Dir:              ".",
						Workspace:        "myworkspace",
						WorkflowName:     String("myworkflow"),
						TerraformVersion: tfVersion,
						Autoplan: valid.Autoplan{
							WhenModified: []string{"**/*.tf*", "**/terragrunt.hcl"},
							Enabled:      true,
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
			},
		},
		{
			description: "project field with autoplan",
			input: `
version: 3
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
			exp: valid.RepoCfg{
				Version: 3,
				Projects: []valid.Project{
					{
						Dir:              ".",
						Workspace:        "myworkspace",
						WorkflowName:     String("myworkflow"),
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
			},
		},
		{
			description: "project field with mergeable apply requirement",
			input: `
version: 3
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
			exp: valid.RepoCfg{
				Version: 3,
				Projects: []valid.Project{
					{
						Dir:              ".",
						Workspace:        "myworkspace",
						WorkflowName:     String("myworkflow"),
						TerraformVersion: tfVersion,
						Autoplan: valid.Autoplan{
							WhenModified: []string{"**/*.tf*", "**/terragrunt.hcl"},
							Enabled:      false,
						},
						ApplyRequirements: []string{"mergeable"},
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
			},
		},
		{
			description: "project field with undiverged apply requirement",
			input: `
version: 3
projects:
- dir: .
  workspace: myworkspace
  terraform_version: v0.11.0
  apply_requirements: [undiverged]
  workflow: myworkflow
  autoplan:
    enabled: false
workflows:
  myworkflow: ~`,
			exp: valid.RepoCfg{
				Version: 3,
				Projects: []valid.Project{
					{
						Dir:              ".",
						Workspace:        "myworkspace",
						WorkflowName:     String("myworkflow"),
						TerraformVersion: tfVersion,
						Autoplan: valid.Autoplan{
							WhenModified: []string{"**/*.tf*", "**/terragrunt.hcl"},
							Enabled:      false,
						},
						ApplyRequirements: []string{"undiverged"},
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
			},
		},
		{
			description: "project field with mergeable and approved apply requirements",
			input: `
version: 3
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
			exp: valid.RepoCfg{
				Version: 3,
				Projects: []valid.Project{
					{
						Dir:              ".",
						Workspace:        "myworkspace",
						WorkflowName:     String("myworkflow"),
						TerraformVersion: tfVersion,
						Autoplan: valid.Autoplan{
							WhenModified: []string{"**/*.tf*", "**/terragrunt.hcl"},
							Enabled:      false,
						},
						ApplyRequirements: []string{"mergeable", "approved"},
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
			},
		},
		{
			description: "project field with undiverged and approved apply requirements",
			input: `
version: 3
projects:
- dir: .
  workspace: myworkspace
  terraform_version: v0.11.0
  apply_requirements: [undiverged, approved]
  workflow: myworkflow
  autoplan:
    enabled: false
workflows:
  myworkflow: ~`,
			exp: valid.RepoCfg{
				Version: 3,
				Projects: []valid.Project{
					{
						Dir:              ".",
						Workspace:        "myworkspace",
						WorkflowName:     String("myworkflow"),
						TerraformVersion: tfVersion,
						Autoplan: valid.Autoplan{
							WhenModified: []string{"**/*.tf*", "**/terragrunt.hcl"},
							Enabled:      false,
						},
						ApplyRequirements: []string{"undiverged", "approved"},
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
			},
		},
		{
			description: "project field with undiverged and mergeable apply requirements",
			input: `
version: 3
projects:
- dir: .
  workspace: myworkspace
  terraform_version: v0.11.0
  apply_requirements: [undiverged, mergeable]
  workflow: myworkflow
  autoplan:
    enabled: false
workflows:
  myworkflow: ~`,
			exp: valid.RepoCfg{
				Version: 3,
				Projects: []valid.Project{
					{
						Dir:              ".",
						Workspace:        "myworkspace",
						WorkflowName:     String("myworkflow"),
						TerraformVersion: tfVersion,
						Autoplan: valid.Autoplan{
							WhenModified: []string{"**/*.tf*", "**/terragrunt.hcl"},
							Enabled:      false,
						},
						ApplyRequirements: []string{"undiverged", "mergeable"},
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
			},
		},
		{
			description: "project field with undiverged, mergeable and approved apply requirements",
			input: `
version: 3
projects:
- dir: .
  workspace: myworkspace
  terraform_version: v0.11.0
  apply_requirements: [undiverged, mergeable, approved]
  workflow: myworkflow
  autoplan:
    enabled: false
workflows:
  myworkflow: ~`,
			exp: valid.RepoCfg{
				Version: 3,
				Projects: []valid.Project{
					{
						Dir:              ".",
						Workspace:        "myworkspace",
						WorkflowName:     String("myworkflow"),
						TerraformVersion: tfVersion,
						Autoplan: valid.Autoplan{
							WhenModified: []string{"**/*.tf*", "**/terragrunt.hcl"},
							Enabled:      false,
						},
						ApplyRequirements: []string{"undiverged", "mergeable", "approved"},
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
			},
		},
		{
			description: "project dir with ..",
			input: `
version: 3
projects:
- dir: ..`,
			expErr: "projects: (0: (dir: cannot contain '..'.).).",
		},

		// Project must have dir set.
		{
			description: "project with no config",
			input: `
version: 3
projects:
-`,
			expErr: "projects: (0: (dir: cannot be blank.).).",
		},
		{
			description: "project with no config at index 1",
			input: `
version: 3
projects:
- dir: "."
-`,
			expErr: "projects: (1: (dir: cannot be blank.).).",
		},
		{
			description: "project with unknown key",
			input: `
version: 3
projects:
- unknown: value`,
			expErr: "yaml: unmarshal errors:\n  line 4: field unknown not found in type raw.Project",
		},
		{
			description: "referencing workflow that doesn't exist",
			input: `
version: 3
projects:
- dir: .
  workflow: undefined`,
			expErr: "workflow \"undefined\" is not defined anywhere",
		},
		{
			description: "two projects with same dir/workspace without names",
			input: `
version: 3
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
version: 3
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
version: 3
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
version: 3
projects:
- name: myname
  dir: .
  workspace: workspace
- name: myname2
  dir: .
  workspace: workspace`,
			exp: valid.RepoCfg{
				Version: 3,
				Projects: []valid.Project{
					{
						Name:      String("myname"),
						Dir:       ".",
						Workspace: "workspace",
						Autoplan: valid.Autoplan{
							WhenModified: []string{"**/*.tf*", "**/terragrunt.hcl"},
							Enabled:      true,
						},
					},
					{
						Name:      String("myname2"),
						Dir:       ".",
						Workspace: "workspace",
						Autoplan: valid.Autoplan{
							WhenModified: []string{"**/*.tf*", "**/terragrunt.hcl"},
							Enabled:      true,
						},
					},
				},
				Workflows: map[string]valid.Workflow{},
			},
		},
		{
			description: "if steps are set then we parse them properly",
			input: `
version: 3
projects:
- dir: "."
workflows:
  default:
    plan:
      steps:
      - init
      - plan
    policy_check:
      steps:
      - init
      - policy_check
    apply:
      steps:
      - plan # NOTE: we don't validate if they make sense
      - apply
`,
			exp: valid.RepoCfg{
				Version: 3,
				Projects: []valid.Project{
					{
						Dir:       ".",
						Workspace: "default",
						Autoplan: valid.Autoplan{
							WhenModified: []string{"**/*.tf*", "**/terragrunt.hcl"},
							Enabled:      true,
						},
					},
				},
				Workflows: map[string]valid.Workflow{
					"default": {
						Name: "default",
						Plan: valid.Stage{
							Steps: []valid.Step{
								{
									StepName: "init",
								},
								{
									StepName: "plan",
								},
							},
						},
						PolicyCheck: valid.Stage{
							Steps: []valid.Step{
								{
									StepName: "init",
								},
								{
									StepName: "policy_check",
								},
							},
						},
						Apply: valid.Stage{
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
version: 3
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
    policy_check:
      steps:
      - policy_check:
          extra_args:
          - arg1
    apply:
      steps:
      - plan:
          extra_args: [a, b]
      - apply:
          extra_args: ["a", "b"]
`,
			exp: valid.RepoCfg{
				Version: 3,
				Projects: []valid.Project{
					{
						Dir:       ".",
						Workspace: "default",
						Autoplan: valid.Autoplan{
							WhenModified: []string{"**/*.tf*", "**/terragrunt.hcl"},
							Enabled:      true,
						},
					},
				},
				Workflows: map[string]valid.Workflow{
					"default": {
						Name: "default",
						Plan: valid.Stage{
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
						PolicyCheck: valid.Stage{
							Steps: []valid.Step{
								{
									StepName:  "policy_check",
									ExtraArgs: []string{"arg1"},
								},
							},
						},
						Apply: valid.Stage{
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
version: 3
projects:
- dir: "."
workflows:
  default:
    plan:
      steps:
      - run: "echo \"plan hi\""
    policy_check:
      steps:
      - run: "echo \"opa hi\""
    apply:
      steps:
      - run: echo apply "arg 2"
`,
			exp: valid.RepoCfg{
				Version: 3,
				Projects: []valid.Project{
					{
						Dir:       ".",
						Workspace: "default",
						Autoplan: valid.Autoplan{
							WhenModified: []string{"**/*.tf*", "**/terragrunt.hcl"},
							Enabled:      true,
						},
					},
				},
				Workflows: map[string]valid.Workflow{
					"default": {
						Name: "default",
						Plan: valid.Stage{
							Steps: []valid.Step{
								{
									StepName:   "run",
									RunCommand: "echo \"plan hi\"",
								},
							},
						},
						PolicyCheck: valid.Stage{
							Steps: []valid.Step{
								{
									StepName:   "run",
									RunCommand: "echo \"opa hi\"",
								},
							},
						},
						Apply: valid.Stage{
							Steps: []valid.Step{
								{
									StepName:   "run",
									RunCommand: "echo apply \"arg 2\"",
								},
							},
						},
					},
				},
			},
		},
		{
			description: "env steps",
			input: `
version: 3
projects:
- dir: "."
workflows:
  default:
    plan:
      steps:
      - env:
          name: env_name
          value: env_value
    policy_check:
      steps:
      - env:
          name: env_name
          value: env_value
    apply:
      steps:
      - env:
          name: env_name
          command: command and args
`,
			exp: valid.RepoCfg{
				Version: 3,
				Projects: []valid.Project{
					{
						Dir:       ".",
						Workspace: "default",
						Autoplan: valid.Autoplan{
							WhenModified: []string{"**/*.tf*", "**/terragrunt.hcl"},
							Enabled:      true,
						},
					},
				},
				Workflows: map[string]valid.Workflow{
					"default": {
						Name: "default",
						Plan: valid.Stage{
							Steps: []valid.Step{
								{
									StepName:    "env",
									EnvVarName:  "env_name",
									EnvVarValue: "env_value",
								},
							},
						},
						PolicyCheck: valid.Stage{
							Steps: []valid.Step{
								{
									StepName:    "env",
									EnvVarName:  "env_name",
									EnvVarValue: "env_value",
								},
							},
						},
						Apply: valid.Stage{
							Steps: []valid.Step{
								{
									StepName:   "env",
									EnvVarName: "env_name",
									RunCommand: "command and args",
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
			err := os.WriteFile(filepath.Join(tmpDir, "atlantis.yaml"), []byte(c.input), 0600)
			Ok(t, err)

			r := config.ParserValidator{}
			act, err := r.ParseRepoCfg(tmpDir, globalCfg, "")
			if c.expErr != "" {
				ErrEquals(t, c.expErr, err)
				return
			}
			Ok(t, err)
			Equals(t, c.exp, act)
		})
	}
}

// Test that we fail if the global validation fails. We test global validation
// more completely in GlobalCfg.ValidateRepoCfg().
func TestParseRepoCfg_GlobalValidation(t *testing.T) {
	tmpDir, cleanup := TempDir(t)
	defer cleanup()

	repoCfg := `
version: 3
projects:
- dir: .
  workflow: custom
workflows:
  custom: ~`
	err := os.WriteFile(filepath.Join(tmpDir, "atlantis.yaml"), []byte(repoCfg), 0600)
	Ok(t, err)

	r := config.ParserValidator{}

	_, err = r.ParseRepoCfg(tmpDir, valid.NewGlobalCfg("somedir"), "repo_id")
	ErrEquals(t, "repo config not allowed to set 'workflow' key: server-side config needs 'allowed_overrides: [workflow]'", err)
}

func TestParseGlobalCfg_NotExist(t *testing.T) {
	r := config.ParserValidator{}
	_, err := r.ParseGlobalCfg("/not/exist", valid.NewGlobalCfg("somedir"))
	ErrEquals(t, "unable to read /not/exist file: open /not/exist: no such file or directory", err)
}

func TestParseGlobalCfg(t *testing.T) {
	defaultCfg := valid.NewGlobalCfg("somedir")
	preWorkflowHook := &valid.PreWorkflowHook{
		StepName:   "run",
		RunCommand: "custom workflow command",
	}
	preWorkflowHooks := []*valid.PreWorkflowHook{preWorkflowHook}
	defaultCfg.Temporal.TerraformTaskQueue = raw.DefaultTaskqueue

	customWorkflow1 := valid.Workflow{
		Name: "custom1",
		Plan: valid.Stage{
			Steps: []valid.Step{
				{
					StepName:   "run",
					RunCommand: "custom command",
				},
				{
					StepName:  "init",
					ExtraArgs: []string{"extra", "args"},
				},
				{
					StepName: "plan",
				},
			},
		},
		PolicyCheck: valid.Stage{
			Steps: []valid.Step{
				{
					StepName:   "run",
					RunCommand: "custom command",
				},
				{
					StepName:  "plan",
					ExtraArgs: []string{"extra", "args"},
				},
				{
					StepName: "policy_check",
				},
			},
		},
		Apply: valid.Stage{
			Steps: []valid.Step{
				{
					StepName:   "run",
					RunCommand: "custom command",
				},
				{
					StepName: "apply",
				},
			},
		},
	}

	conftestVersion, _ := version.NewVersion("v1.0.0")

	cases := map[string]struct {
		input  string
		expErr string
		exp    valid.GlobalCfg
	}{
		"empty file": {
			input:  "",
			expErr: "file <tmp> was empty",
		},
		"invalid fields": {
			input:  "invalid: key",
			expErr: "yaml: unmarshal errors:\n  line 1: field invalid not found in type raw.GlobalCfg",
		},
		"no id specified": {
			input: `repos:
- apply_requirements: []`,
			expErr: "repos: (0: (id: cannot be blank.).).",
		},
		"invalid id regex": {
			input: `repos:
- id: /?/`,
			expErr: "repos: (0: (id: parsing: /?/: error parsing regexp: missing argument to repetition operator: `?`.).).",
		},
		"invalid branch regex": {
			input: `repos:
- id: /.*/
  branch: /?/`,
			expErr: "repos: (0: (branch: parsing: /?/: error parsing regexp: missing argument to repetition operator: `?`.).).",
		},
		"workflow doesn't exist": {
			input: `repos:
- id: /.*/
  workflow: notdefined`,
			expErr: "workflow \"notdefined\" is not defined",
		},
		"invalid allowed_override": {
			input: `repos:
- id: /.*/
  allowed_overrides: [invalid]`,
			expErr: "repos: (0: (allowed_overrides: \"invalid\" is not a valid override, only \"apply_requirements\" and \"workflow\" are supported.).).",
		},
		"invalid apply_requirement": {
			input: `repos:
- id: /.*/
  apply_requirements: [invalid]`,
			expErr: "repos: (0: (apply_requirements: \"invalid\" is not a valid apply_requirement, supported apply requirements are: \"approved\", \"mergeable\", \"undiverged\", \"unlocked\".).).",
		},
		"no workflows key": {
			input: `repos: []`,
			exp:   defaultCfg,
		},
		"workflows empty": {
			input: `workflows:`,
			exp:   defaultCfg,
		},
		"workflow name but the rest is empty": {
			input: `
workflows:
  name:`,
			exp: valid.GlobalCfg{
				Repos: defaultCfg.Repos,
				Workflows: map[string]valid.Workflow{
					"default": defaultCfg.Workflows["default"],
					"name": {
						Name:        "name",
						Apply:       valid.DefaultApplyStage,
						Plan:        valid.DefaultPlanStage,
						PolicyCheck: valid.DefaultPolicyCheckStage,
					},
				},
				PullRequestWorkflows: map[string]valid.Workflow{
					"default": defaultCfg.PullRequestWorkflows["default"],
				},
				DeploymentWorkflows: map[string]valid.Workflow{
					"default": defaultCfg.DeploymentWorkflows["default"],
				},
				Temporal:          valid.Temporal{TerraformTaskQueue: raw.DefaultTaskqueue},
				PersistenceConfig: defaultCfg.PersistenceConfig,
			},
		},
		"workflow stages empty": {
			input: `
workflows:
  name:
    apply:
    plan:
`,
			exp: valid.GlobalCfg{
				Repos: defaultCfg.Repos,
				Workflows: map[string]valid.Workflow{
					"default": defaultCfg.Workflows["default"],
					"name": {
						Name:        "name",
						Apply:       valid.DefaultApplyStage,
						Plan:        valid.DefaultPlanStage,
						PolicyCheck: valid.DefaultPolicyCheckStage,
					},
				},
				PullRequestWorkflows: map[string]valid.Workflow{
					"default": defaultCfg.PullRequestWorkflows["default"],
				},
				DeploymentWorkflows: map[string]valid.Workflow{
					"default": defaultCfg.DeploymentWorkflows["default"],
				},
				Temporal:          valid.Temporal{TerraformTaskQueue: raw.DefaultTaskqueue},
				PersistenceConfig: defaultCfg.PersistenceConfig,
			},
		},
		"workflow steps empty": {
			input: `
workflows:
  name:
    apply:
      steps:
    plan:
      steps:`,
			exp: valid.GlobalCfg{
				Repos: defaultCfg.Repos,
				Workflows: map[string]valid.Workflow{
					"default": defaultCfg.Workflows["default"],
					"name": {
						Name:        "name",
						Plan:        valid.DefaultPlanStage,
						PolicyCheck: valid.DefaultPolicyCheckStage,
						Apply:       valid.DefaultApplyStage,
					},
				},
				PullRequestWorkflows: map[string]valid.Workflow{
					"default": defaultCfg.PullRequestWorkflows["default"],
				},
				DeploymentWorkflows: map[string]valid.Workflow{
					"default": defaultCfg.DeploymentWorkflows["default"],
				},
				Temporal:          valid.Temporal{TerraformTaskQueue: raw.DefaultTaskqueue},
				PersistenceConfig: defaultCfg.PersistenceConfig,
			},
		},
		"all keys specified": {
			input: `
repos:
- id: github.com/owner/repo

  apply_requirements: [approved, mergeable]
  pre_workflow_hooks:
    - run: custom workflow command
  workflow: custom1
  allowed_overrides: [apply_requirements, workflow]
  allow_custom_workflows: true
  checkout_strategy: merge
- id: /.*/
  branch: /(master|main)/
  pre_workflow_hooks:
    - run: custom workflow command
workflows:
  custom1:
    plan:
      steps:
      - run: custom command
      - init:
          extra_args: [extra, args]
      - plan
    policy_check:
      steps:
      - run: custom command
      - plan:
          extra_args: [extra, args]
      - policy_check
    apply:
      steps:
      - run: custom command
      - apply
policies:
  conftest_version: v1.0.0
  policy_sets:
    - name: good-policy
      path: rel/path/to/policy
      source: local
`,
			exp: valid.GlobalCfg{
				Repos: []valid.Repo{
					defaultCfg.Repos[0],
					{
						ID:                   "github.com/owner/repo",
						ApplyRequirements:    []string{"approved", "mergeable", "policies_passed"},
						PreWorkflowHooks:     preWorkflowHooks,
						Workflow:             &customWorkflow1,
						AllowedOverrides:     []string{"apply_requirements", "workflow"},
						AllowCustomWorkflows: Bool(true),
						CheckoutStrategy:     "merge",
					},
					{
						IDRegex:           regexp.MustCompile(".*"),
						BranchRegex:       regexp.MustCompile("(master|main)"),
						ApplyRequirements: []string{"policies_passed"},
						PreWorkflowHooks:  preWorkflowHooks,
						CheckoutStrategy:  "branch",
					},
				},
				Workflows: map[string]valid.Workflow{
					"default": defaultCfg.Workflows["default"],
					"custom1": customWorkflow1,
				},
				PolicySets: valid.PolicySets{
					Version: conftestVersion,
					PolicySets: []valid.PolicySet{
						{
							Name:   "good-policy",
							Path:   "rel/path/to/policy",
							Source: valid.LocalPolicySet,
						},
					},
				},
				PullRequestWorkflows: map[string]valid.Workflow{
					"default": defaultCfg.PullRequestWorkflows["default"],
				},
				DeploymentWorkflows: map[string]valid.Workflow{
					"default": defaultCfg.DeploymentWorkflows["default"],
				},
				Temporal:          valid.Temporal{TerraformTaskQueue: raw.DefaultTaskqueue},
				PersistenceConfig: defaultCfg.PersistenceConfig,
			},
		},
		"id regex with trailing slash": {
			input: `
repos:
- id: /github.com//
`,
			exp: valid.GlobalCfg{
				Repos: []valid.Repo{
					defaultCfg.Repos[0],
					{
						IDRegex:          regexp.MustCompile("github.com/"),
						CheckoutStrategy: "branch",
					},
				},
				Workflows: map[string]valid.Workflow{
					"default": defaultCfg.Workflows["default"],
				},
				PullRequestWorkflows: map[string]valid.Workflow{
					"default": defaultCfg.PullRequestWorkflows["default"],
				},
				DeploymentWorkflows: map[string]valid.Workflow{
					"default": defaultCfg.DeploymentWorkflows["default"],
				},
				Temporal:          valid.Temporal{TerraformTaskQueue: raw.DefaultTaskqueue},
				PersistenceConfig: defaultCfg.PersistenceConfig,
			},
		},
		"referencing default workflow": {
			input: `
repos:
- id: github.com/owner/repo
  workflow: default
`,
			exp: valid.GlobalCfg{
				Repos: []valid.Repo{
					defaultCfg.Repos[0],
					{
						ID:               "github.com/owner/repo",
						Workflow:         defaultCfg.Repos[0].Workflow,
						CheckoutStrategy: "branch",
					},
				},
				Workflows: map[string]valid.Workflow{
					"default": defaultCfg.Workflows["default"],
				},
				PullRequestWorkflows: map[string]valid.Workflow{
					"default": defaultCfg.PullRequestWorkflows["default"],
				},
				DeploymentWorkflows: map[string]valid.Workflow{
					"default": defaultCfg.DeploymentWorkflows["default"],
				},
				Temporal:          valid.Temporal{TerraformTaskQueue: raw.DefaultTaskqueue},
				PersistenceConfig: defaultCfg.PersistenceConfig,
			},
		},
		"redefine default workflow": {
			input: `
workflows:
  default:
    plan:
      steps:
      - run: custom
    policy_check:
      steps: []
    apply:
     steps: []
`,
			exp: valid.GlobalCfg{
				Repos: []valid.Repo{
					{
						IDRegex:           regexp.MustCompile(".*"),
						BranchRegex:       regexp.MustCompile(".*"),
						ApplyRequirements: []string{},
						Workflow: &valid.Workflow{
							Name: "default",
							Apply: valid.Stage{
								Steps: nil,
							},
							PolicyCheck: valid.Stage{
								Steps: nil,
							},
							Plan: valid.Stage{
								Steps: []valid.Step{
									{
										StepName:   "run",
										RunCommand: "custom",
									},
								},
							},
						},
						PullRequestWorkflow: &valid.Workflow{
							Name:        "default",
							Plan:        valid.DefaultLocklessPlanStage,
							PolicyCheck: valid.DefaultPolicyCheckStage,
						},
						DeploymentWorkflow: &valid.Workflow{
							Name:  "default",
							Plan:  valid.DefaultPlanStage,
							Apply: valid.DefaultApplyStage,
						},
						AllowedWorkflows:     []string{},
						AllowedOverrides:     []string{},
						AllowCustomWorkflows: Bool(false),
						CheckoutStrategy:     "branch",
					},
				},
				Workflows: map[string]valid.Workflow{
					"default": {
						Name: "default",
						Apply: valid.Stage{
							Steps: nil,
						},
						Plan: valid.Stage{
							Steps: []valid.Step{
								{
									StepName:   "run",
									RunCommand: "custom",
								},
							},
						},
					},
				},
				PullRequestWorkflows: map[string]valid.Workflow{
					"default": {
						Name:        "default",
						Plan:        valid.DefaultLocklessPlanStage,
						PolicyCheck: valid.DefaultPolicyCheckStage,
					},
				},
				DeploymentWorkflows: map[string]valid.Workflow{
					"default": {
						Name:  "default",
						Plan:  valid.DefaultPlanStage,
						Apply: valid.DefaultApplyStage,
					},
				},
				Temporal:          valid.Temporal{TerraformTaskQueue: raw.DefaultTaskqueue},
				PersistenceConfig: defaultCfg.PersistenceConfig,
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			r := config.ParserValidator{}
			tmp, cleanup := TempDir(t)
			defer cleanup()
			path := filepath.Join(tmp, "conf.yaml")
			Ok(t, os.WriteFile(path, []byte(c.input), 0600))

			act, err := r.ParseGlobalCfg(path, valid.NewGlobalCfg("somedir"))

			if c.expErr != "" {
				expErr := strings.Replace(c.expErr, "<tmp>", path, -1)
				ErrEquals(t, expErr, err)
				return
			}
			Ok(t, err)

			if !act.PolicySets.HasPolicies() {
				c.exp.PolicySets = act.PolicySets
			}

			Equals(t, c.exp, act)
			// Have to hand-compare regexes because Equals doesn't do it.
			for i, actRepo := range act.Repos {
				expRepo := c.exp.Repos[i]
				if expRepo.IDRegex != nil {
					Assert(t, expRepo.IDRegex.String() == actRepo.IDRegex.String(),
						"%q != %q for repos[%d]", expRepo.IDRegex.String(), actRepo.IDRegex.String(), i)
				}
				if expRepo.BranchRegex != nil {
					Assert(t, expRepo.BranchRegex.String() == actRepo.BranchRegex.String(),
						"%q != %q for repos[%d]", expRepo.BranchRegex.String(), actRepo.BranchRegex.String(), i)
				}
			}
		})
	}
}

func TestParseGlobalCfg_PlatformMode(t *testing.T) {
	defaultCfg := valid.NewGlobalCfg("somedir")
	preWorkflowHook := &valid.PreWorkflowHook{
		StepName:   "run",
		RunCommand: "custom workflow command",
	}
	preWorkflowHooks := []*valid.PreWorkflowHook{preWorkflowHook}

	customPlan1 := valid.Stage{
		Steps: []valid.Step{
			{
				StepName:   "run",
				RunCommand: "custom command",
			},
			{
				StepName:  "init",
				ExtraArgs: []string{"extra", "args"},
			},
			{
				StepName: "plan",
			},
		},
	}
	customPolicyCheck1 := valid.Stage{
		Steps: []valid.Step{
			{
				StepName:   "run",
				RunCommand: "custom command",
			},
			{
				StepName:  "plan",
				ExtraArgs: []string{"extra", "args"},
			},
			{
				StepName: "policy_check",
			},
		},
	}
	customApply1 := valid.Stage{
		Steps: []valid.Step{
			{
				StepName:   "run",
				RunCommand: "custom command",
			},
			{
				StepName: "apply",
			},
		},
	}

	customPulRequestWorkflow1 := valid.Workflow{
		Name:        "custom1",
		Plan:        customPlan1,
		PolicyCheck: customPolicyCheck1,
	}

	customDeploymentWorkflow1 := valid.Workflow{
		Name:  "custom1",
		Plan:  customPlan1,
		Apply: customApply1,
	}

	conftestVersion, _ := version.NewVersion("v1.0.0")
	defaultWorkflow := defaultCfg.Workflows["default"]

	cases := map[string]struct {
		input  string
		expErr string
		exp    valid.GlobalCfg
	}{
		"pull_request_workflows don't support apply": {
			input: `
pull_request_workflows:
  default:
    apply:
      steps:
        - run: custom
`,
			expErr: "yaml: unmarshal errors:\n  line 4: field apply not found in type raw.PullRequestWorkflow",
		},
		"deployment_workflows don't support policy_checks": {
			input: `
deployment_workflows:
  default:
    policy_check:
      steps:
        - run: custom
`,
			expErr: "yaml: unmarshal errors:\n  line 4: field policy_check not found in type raw.DeploymentWorkflow",
		},
		"all keys specified": {
			input: `
repos:
- id: github.com/owner/repo
  pre_workflow_hooks:
    - run: custom workflow command
  pull_request_workflow: custom1
  deployment_workflow: custom1
  allowed_overrides: [apply_requirements, pull_request_workflow, deployment_workflow, workflow]
  allow_custom_workflows: true
- id: /.*/
  branch: /(master|main)/
  pre_workflow_hooks:
    - run: custom workflow command
pull_request_workflows:
  custom1:
    plan:
      steps:
      - run: custom command
      - init:
          extra_args: [extra, args]
      - plan
    policy_check:
      steps:
      - run: custom command
      - plan:
          extra_args: [extra, args]
      - policy_check
deployment_workflows:
  custom1:
    plan:
      steps:
      - run: custom command
      - init:
          extra_args: [extra, args]
      - plan
    apply:
      steps:
      - run: custom command
      - apply
policies:
  conftest_version: v1.0.0
  policy_sets:
    - name: good-policy
      path: rel/path/to/policy
      source: local
`,
			exp: valid.GlobalCfg{
				Repos: []valid.Repo{
					defaultCfg.Repos[0],
					{
						ID:                   "github.com/owner/repo",
						PreWorkflowHooks:     preWorkflowHooks,
						PullRequestWorkflow:  &customPulRequestWorkflow1,
						DeploymentWorkflow:   &customDeploymentWorkflow1,
						ApplyRequirements:    []string{"policies_passed"},
						AllowedOverrides:     []string{"apply_requirements", "pull_request_workflow", "deployment_workflow", "workflow"},
						AllowCustomWorkflows: Bool(true),
						CheckoutStrategy:     "branch",
					},
					{
						IDRegex:           regexp.MustCompile(".*"),
						BranchRegex:       regexp.MustCompile("(master|main)"),
						ApplyRequirements: []string{"policies_passed"},
						PreWorkflowHooks:  preWorkflowHooks,
						CheckoutStrategy:  "branch",
					},
				},
				Workflows: defaultCfg.Workflows,
				PullRequestWorkflows: map[string]valid.Workflow{
					"default": defaultCfg.PullRequestWorkflows["default"],
					"custom1": customPulRequestWorkflow1,
				},
				DeploymentWorkflows: map[string]valid.Workflow{
					"default": defaultCfg.DeploymentWorkflows["default"],
					"custom1": customDeploymentWorkflow1,
				},
				Temporal: valid.Temporal{TerraformTaskQueue: raw.DefaultTaskqueue},
				PolicySets: valid.PolicySets{
					Version: conftestVersion,
					PolicySets: []valid.PolicySet{
						{
							Name:   "good-policy",
							Path:   "rel/path/to/policy",
							Source: valid.LocalPolicySet,
						},
					},
				},
				PersistenceConfig: defaultCfg.PersistenceConfig,
			},
		},
		"redefine default platform workflows": {
			input: `
pull_request_workflows:
  default:
    plan:
      steps:
      - run: custom
    policy_check:
      steps: []
deployment_workflows:
  default:
    plan:
      steps: []
    apply:
      steps:
      - run: custom
`,
			exp: valid.GlobalCfg{
				Repos: []valid.Repo{
					{
						IDRegex:           regexp.MustCompile(".*"),
						BranchRegex:       regexp.MustCompile(".*"),
						ApplyRequirements: []string{},
						Workflow:          &defaultWorkflow,
						PullRequestWorkflow: &valid.Workflow{
							Name: "default",
							Apply: valid.Stage{
								Steps: nil,
							},
							PolicyCheck: valid.Stage{
								Steps: nil,
							},
							Plan: valid.Stage{
								Steps: []valid.Step{
									{
										StepName:   "run",
										RunCommand: "custom",
									},
								},
							},
						},
						DeploymentWorkflow: &valid.Workflow{
							Name: "default",
							Apply: valid.Stage{
								Steps: []valid.Step{
									{
										StepName:   "run",
										RunCommand: "custom",
									},
								},
							},
							PolicyCheck: valid.Stage{
								Steps: nil,
							},
							Plan: valid.Stage{
								Steps: nil,
							},
						},
						AllowedWorkflows:     []string{},
						AllowedOverrides:     []string{},
						AllowCustomWorkflows: Bool(false),
						CheckoutStrategy:     "branch",
					},
				},
				Workflows: defaultCfg.Workflows,
				PullRequestWorkflows: map[string]valid.Workflow{
					"default": {
						Name: "default",
						Plan: valid.Stage{
							Steps: []valid.Step{
								{
									StepName:   "run",
									RunCommand: "custom",
								},
							},
						},
					},
				},
				DeploymentWorkflows: map[string]valid.Workflow{
					"default": {
						Name: "default",
						Apply: valid.Stage{
							Steps: []valid.Step{
								{
									StepName:   "run",
									RunCommand: "custom",
								},
							},
						},
					},
				},
				Temporal:          valid.Temporal{TerraformTaskQueue: raw.DefaultTaskqueue},
				PersistenceConfig: defaultCfg.PersistenceConfig,
			},
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			r := &config.ParserValidator{}
			tmp, cleanup := TempDir(t)
			defer cleanup()
			path := filepath.Join(tmp, "conf.yaml")
			Ok(t, os.WriteFile(path, []byte(c.input), 0600))

			act, err := r.ParseGlobalCfg(path, valid.NewGlobalCfg("somedir"))

			if c.expErr != "" {
				expErr := strings.Replace(c.expErr, "<tmp>", path, -1)
				ErrEquals(t, expErr, err)
				return
			}
			Ok(t, err)

			if !act.PolicySets.HasPolicies() {
				c.exp.PolicySets = act.PolicySets
			}

			Equals(t, c.exp, act)
			// Have to hand-compare regexes because Equals doesn't do it.
			for i, actRepo := range act.Repos {
				expRepo := c.exp.Repos[i]
				if expRepo.IDRegex != nil {
					Assert(t, expRepo.IDRegex.String() == actRepo.IDRegex.String(),
						"%q != %q for repos[%d]", expRepo.IDRegex.String(), actRepo.IDRegex.String(), i)
				}
				if expRepo.BranchRegex != nil {
					Assert(t, expRepo.BranchRegex.String() == actRepo.BranchRegex.String(),
						"%q != %q for repos[%d]", expRepo.BranchRegex.String(), actRepo.BranchRegex.String(), i)
				}
			}
		})
	}
}

// Test that if we pass in JSON strings everything should parse fine.
func TestParserValidator_ParseGlobalCfgJSON(t *testing.T) {
	customWorkflow := valid.Workflow{
		Name: "custom",
		Plan: valid.Stage{
			Steps: []valid.Step{
				{
					StepName: "init",
				},
				{
					StepName:  "plan",
					ExtraArgs: []string{"extra", "args"},
				},
				{
					StepName:   "run",
					RunCommand: "custom plan",
				},
			},
		},
		PolicyCheck: valid.Stage{
			Steps: []valid.Step{
				{
					StepName: "plan",
				},
				{
					StepName:   "run",
					RunCommand: "custom policy_check",
				},
			},
		},
		Apply: valid.Stage{
			Steps: []valid.Step{
				{
					StepName:   "run",
					RunCommand: "my custom command",
				},
			},
		},
	}

	conftestVersion, _ := version.NewVersion("v1.0.0")
	gCfg := valid.NewGlobalCfg("somedir")
	gCfg.Temporal.TerraformTaskQueue = raw.DefaultTaskqueue

	cases := map[string]struct {
		json   string
		exp    valid.GlobalCfg
		expErr string
	}{
		"empty string": {
			json:   "",
			expErr: "unexpected end of JSON input",
		},
		"empty object": {
			json: "{}",
			exp:  gCfg,
		},
		"setting all keys": {
			json: `
{
  "repos": [
    {
      "id": "/.*/",
      "workflow": "custom",
      "allowed_workflows": ["custom"],
      "apply_requirements": ["mergeable", "approved"],
      "allowed_overrides": ["workflow", "apply_requirements"],
      "allow_custom_workflows": true
    },
    {
      "id": "github.com/owner/repo"
    }
  ],
  "workflows": {
    "custom": {
      "plan": {
        "steps": [
          "init",
          {"plan": {"extra_args": ["extra", "args"]}},
          {"run": "custom plan"}
        ]
      },
      "policy_check": {
        "steps": [
          "plan",
          {"run": "custom policy_check"}
        ]
      },
      "apply": {
        "steps": [
          {"run": "my custom command"}
        ]
      }
    }
  },
  "policies": {
    "conftest_version": "v1.0.0",
    "policy_sets": [
      {
        "name": "good-policy",
        "source": "local",
        "path": "rel/path/to/policy"
      }
    ]
  }
}
`,
			exp: valid.GlobalCfg{
				Repos: []valid.Repo{
					valid.NewGlobalCfg("somedir").Repos[0],
					{
						IDRegex:              regexp.MustCompile(".*"),
						ApplyRequirements:    []string{"mergeable", "approved", "policies_passed"},
						Workflow:             &customWorkflow,
						AllowedWorkflows:     []string{"custom"},
						AllowedOverrides:     []string{"workflow", "apply_requirements"},
						AllowCustomWorkflows: Bool(true),
						CheckoutStrategy:     "branch",
					},
					{
						ID:                   "github.com/owner/repo",
						IDRegex:              nil,
						ApplyRequirements:    []string{"policies_passed"},
						AllowedOverrides:     nil,
						AllowCustomWorkflows: nil,
						CheckoutStrategy:     "branch",
					},
				},
				Workflows: map[string]valid.Workflow{
					"default": valid.NewGlobalCfg("somedir").Workflows["default"],
					"custom":  customWorkflow,
				},
				PullRequestWorkflows: map[string]valid.Workflow{
					"default": valid.NewGlobalCfg("somedir").PullRequestWorkflows["default"],
				},
				DeploymentWorkflows: map[string]valid.Workflow{
					"default": valid.NewGlobalCfg("somedir").DeploymentWorkflows["default"],
				},
				Temporal: valid.Temporal{TerraformTaskQueue: raw.DefaultTaskqueue},
				PolicySets: valid.PolicySets{
					Version: conftestVersion,
					PolicySets: []valid.PolicySet{
						{
							Name:   "good-policy",
							Path:   "rel/path/to/policy",
							Source: valid.LocalPolicySet,
						},
					},
				},
				PersistenceConfig: gCfg.PersistenceConfig,
			},
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			pv := &config.ParserValidator{}
			cfg, err := pv.ParseGlobalCfgJSON(c.json, valid.NewGlobalCfg("somedir"))
			if c.expErr != "" {
				ErrEquals(t, c.expErr, err)
				return
			}
			Ok(t, err)

			if !cfg.PolicySets.HasPolicies() {
				c.exp.PolicySets = cfg.PolicySets
			}

			Equals(t, c.exp, cfg)
		})
	}
}

// Test that if we pass in JSON strings everything should parse fine.
func TestParserValidator_ParseGlobalCfgV2JSON(t *testing.T) {
	customPullRequestWorkflow := valid.Workflow{
		Name: "custom",
		Plan: valid.Stage{
			Steps: []valid.Step{
				{
					StepName: "init",
				},
				{
					StepName:  "plan",
					ExtraArgs: []string{"extra", "args"},
				},
				{
					StepName:   "run",
					RunCommand: "custom plan",
				},
			},
		},
		PolicyCheck: valid.Stage{
			Steps: []valid.Step{
				{
					StepName: "plan",
				},
				{
					StepName:   "run",
					RunCommand: "custom policy_check",
				},
			},
		},
	}

	customDeploymentWorkflow := valid.Workflow{
		Name: "custom",
		Plan: valid.Stage{
			Steps: []valid.Step{
				{
					StepName: "init",
				},
				{
					StepName:  "plan",
					ExtraArgs: []string{"extra", "args"},
				},
				{
					StepName:   "run",
					RunCommand: "custom plan",
				},
			},
		},
		Apply: valid.Stage{
			Steps: []valid.Step{
				{
					StepName:   "run",
					RunCommand: "my custom command",
				},
			},
		},
	}

	conftestVersion, _ := version.NewVersion("v1.0.0")
	globalCfg := valid.NewGlobalCfg("somedir")
	globalCfg.Temporal.TerraformTaskQueue = raw.DefaultTaskqueue

	cases := map[string]struct {
		json   string
		exp    valid.GlobalCfg
		expErr string
	}{
		"empty string": {
			json:   "",
			expErr: "unexpected end of JSON input",
		},
		"empty object": {
			json: "{}",
			exp:  globalCfg,
		},
		"setting all keys": {
			json: `
{
  "repos": [
    {
      "id": "/.*/",
      "pull_request_workflow": "custom",
      "deployment_workflow": "custom",
      "allowed_pull_request_workflows": ["custom"],
      "allowed_deployment_workflows": ["custom"],
      "allowed_overrides": ["pull_request_workflow", "deployment_workflow"],
      "allow_custom_workflows": true
    },
    {
      "id": "github.com/owner/repo"
    }
  ],
  "pull_request_workflows": {
    "custom": {
      "plan": {
        "steps": [
          "init",
          {"plan": {"extra_args": ["extra", "args"]}},
          {"run": "custom plan"}
        ]
      },
      "policy_check": {
        "steps": [
          "plan",
          {"run": "custom policy_check"}
        ]
      }
    }
  },
  "deployment_workflows": {
    "custom": {
      "plan": {
        "steps": [
          "init",
          {"plan": {"extra_args": ["extra", "args"]}},
          {"run": "custom plan"}
        ]
      },
      "apply": {
        "steps": [
          {"run": "my custom command"}
        ]
      }
    }
  },
  "policies": {
    "conftest_version": "v1.0.0",
    "policy_sets": [
      {
        "name": "good-policy",
        "source": "local",
        "path": "rel/path/to/policy"
      }
    ]
  }
}
`,
			exp: valid.GlobalCfg{
				Repos: []valid.Repo{
					globalCfg.Repos[0],
					{
						IDRegex:                     regexp.MustCompile(".*"),
						ApplyRequirements:           []string{"policies_passed"},
						PullRequestWorkflow:         &customPullRequestWorkflow,
						DeploymentWorkflow:          &customDeploymentWorkflow,
						AllowedPullRequestWorkflows: []string{"custom"},
						AllowedDeploymentWorkflows:  []string{"custom"},
						AllowedOverrides:            []string{"pull_request_workflow", "deployment_workflow"},
						AllowCustomWorkflows:        Bool(true),
						CheckoutStrategy:            "branch",
					},
					{
						ID:                          "github.com/owner/repo",
						IDRegex:                     nil,
						ApplyRequirements:           []string{"policies_passed"},
						AllowedOverrides:            nil,
						AllowedPullRequestWorkflows: nil,
						AllowedDeploymentWorkflows:  nil,
						AllowCustomWorkflows:        nil,
						CheckoutStrategy:            "branch",
					},
				},
				Workflows: map[string]valid.Workflow{
					"default": globalCfg.Workflows["default"],
				},
				PullRequestWorkflows: map[string]valid.Workflow{
					"default": globalCfg.PullRequestWorkflows["default"],
					"custom":  customPullRequestWorkflow,
				},
				DeploymentWorkflows: map[string]valid.Workflow{
					"default": globalCfg.DeploymentWorkflows["default"],
					"custom":  customDeploymentWorkflow,
				},
				Temporal: valid.Temporal{TerraformTaskQueue: raw.DefaultTaskqueue},
				PolicySets: valid.PolicySets{
					Version: conftestVersion,
					PolicySets: []valid.PolicySet{
						{
							Name:   "good-policy",
							Path:   "rel/path/to/policy",
							Source: valid.LocalPolicySet,
						},
					},
				},
				PersistenceConfig: globalCfg.PersistenceConfig,
			},
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			pv := &config.ParserValidator{}

			cfg, err := pv.ParseGlobalCfgJSON(c.json, globalCfg)
			if c.expErr != "" {
				ErrEquals(t, c.expErr, err)
				return
			}
			Ok(t, err)

			if !cfg.PolicySets.HasPolicies() {
				c.exp.PolicySets = cfg.PolicySets
			}

			Equals(t, c.exp, cfg)
		})
	}
}

// Test legacy shell parsing vs v3 parsing.
func TestParseRepoCfg_V2ShellParsing(t *testing.T) {
	cases := []struct {
		in       string
		expV2    string
		expV2Err string
	}{
		{
			in:    "echo a b",
			expV2: "echo a b",
		},
		{
			in:    "echo 'a b'",
			expV2: "echo a b",
		},
		{
			in:       "echo 'a b",
			expV2Err: "unable to parse \"echo 'a b\": EOF found when expecting closing quote.",
		},
		{
			in:    `mkdir a/b/c || printf \'your main.tf file does not provide default region.\\ncheck\'`,
			expV2: `mkdir a/b/c || printf 'your main.tf file does not provide default region.\ncheck'`,
		},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			v2Dir, cleanup2 := TempDir(t)
			defer cleanup2()
			v3Dir, cleanup3 := TempDir(t)
			defer cleanup3()
			v2Path := filepath.Join(v2Dir, "atlantis.yaml")
			v3Path := filepath.Join(v3Dir, "atlantis.yaml")
			cfg := fmt.Sprintf(`workflows:
  custom:
    plan:
      steps:
      - run: %s
    apply:
      steps:
      - run: %s`, c.in, c.in)
			Ok(t, os.WriteFile(v2Path, []byte("version: 2\n"+cfg), 0600))
			Ok(t, os.WriteFile(v3Path, []byte("version: 3\n"+cfg), 0600))

			p := &config.ParserValidator{}

			v2Cfg, err := p.ParseRepoCfg(v2Dir, globalCfg, "")
			if c.expV2Err != "" {
				ErrEquals(t, c.expV2Err, err)
			} else {
				Ok(t, err)
				Equals(t, c.expV2, v2Cfg.Workflows["custom"].Plan.Steps[0].RunCommand)
				Equals(t, c.expV2, v2Cfg.Workflows["custom"].Apply.Steps[0].RunCommand)
			}
			v3Cfg, err := p.ParseRepoCfg(v3Dir, globalCfg, "")
			Ok(t, err)
			Equals(t, c.in, v3Cfg.Workflows["custom"].Plan.Steps[0].RunCommand)
			Equals(t, c.in, v3Cfg.Workflows["custom"].Apply.Steps[0].RunCommand)
		})
	}
}
