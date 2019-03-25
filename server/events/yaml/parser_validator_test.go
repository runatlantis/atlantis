package yaml_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/events/yaml"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
	. "github.com/runatlantis/atlantis/testing"
)

var globalCfg = valid.NewGlobalCfg(true, false, false)

func TestHasRepoCfg_DirDoesNotExist(t *testing.T) {
	r := yaml.ParserValidator{}
	exists, err := r.HasRepoCfg("/not/exist")
	Ok(t, err)
	Equals(t, false, exists)
}

func TestHasRepoCfg_FileDoesNotExist(t *testing.T) {
	tmpDir, cleanup := TempDir(t)
	defer cleanup()
	r := yaml.ParserValidator{}
	exists, err := r.HasRepoCfg(tmpDir)
	Ok(t, err)
	Equals(t, false, exists)
}

func TestParseRepoCfg_DirDoesNotExist(t *testing.T) {
	r := yaml.ParserValidator{}
	_, err := r.ParseRepoCfg("/not/exist", globalCfg, "")
	Assert(t, os.IsNotExist(err), "exp not exist err")
}

func TestParseRepoCfg_FileDoesNotExist(t *testing.T) {
	tmpDir, cleanup := TempDir(t)
	defer cleanup()
	r := yaml.ParserValidator{}
	_, err := r.ParseRepoCfg(tmpDir, globalCfg, "")
	Assert(t, os.IsNotExist(err), "exp not exist err")
}

func TestParseRepoCfg_BadPermissions(t *testing.T) {
	tmpDir, cleanup := TempDir(t)
	defer cleanup()
	err := ioutil.WriteFile(filepath.Join(tmpDir, "atlantis.yaml"), nil, 0000)
	Ok(t, err)

	r := yaml.ParserValidator{}
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
			err := ioutil.WriteFile(confPath, []byte(c.input), 0600)
			Ok(t, err)
			r := yaml.ParserValidator{}
			_, err = r.ParseRepoCfg(tmpDir, globalCfg, "")
			ErrContains(t, c.expErr, err)
			_, err = r.ParseGlobalCfg(confPath, valid.NewGlobalCfg(false, false, false))
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
			description: "autoplan should be enabled by default",
			input: `
version: 2
projects:
- dir: "."
  autoplan:
    when_modified: ["**/*.tf*"]
`,
			exp: valid.Config{
				Version: 2,
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
version: 2
projects:
- dir: "."
`,
			exp: valid.Config{
				Version: 2,
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
			description: "if workflows key set but with no workflows there are none",
			input: `
version: 2
projects:
- dir: "."
workflows: ~
`,
			exp: valid.Config{
				Version: 2,
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
			description: "if a plan or apply explicitly defines an empty steps key then it gets the defaults",
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
			exp: valid.Config{
				Version: 2,
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
				Workflows: map[string]valid.Workflow{
					"default": {
						Name:  "default",
						Plan:  valid.DefaultPlanStage,
						Apply: valid.DefaultApplyStage,
					},
				},
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
					"myworkflow": {
						Name:  "myworkflow",
						Apply: valid.DefaultApplyStage,
						Plan:  valid.DefaultPlanStage,
					},
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
					"myworkflow": {
						Name:  "myworkflow",
						Apply: valid.DefaultApplyStage,
						Plan:  valid.DefaultPlanStage,
					},
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
					"myworkflow": {
						Name:  "myworkflow",
						Apply: valid.DefaultApplyStage,
						Plan:  valid.DefaultPlanStage,
					},
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
					"myworkflow": {
						Name:  "myworkflow",
						Apply: valid.DefaultApplyStage,
						Plan:  valid.DefaultPlanStage,
					},
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
			expErr: "workflow \"undefined\" is not defined anywhere",
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
      - plan # NOTE: we don't validate if they make sense
      - apply
`,
			exp: valid.Config{
				Version: 2,
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
			exp: valid.Config{
				Version: 2,
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
			exp: valid.Config{
				Version: 2,
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
				Workflows: map[string]valid.Workflow{
					"default": {
						Name: "default",
						Plan: valid.Stage{
							Steps: []valid.Step{
								{
									StepName:   "run",
									RunCommand: []string{"echo", "plan hi"},
								},
							},
						},
						Apply: valid.Stage{
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
version: 2
projects:
- dir: .
  workflow: custom
workflows:
  custom: ~`
	err := ioutil.WriteFile(filepath.Join(tmpDir, "atlantis.yaml"), []byte(repoCfg), 0600)
	Ok(t, err)

	r := yaml.ParserValidator{}
	_, err = r.ParseRepoCfg(tmpDir, valid.NewGlobalCfg(false, false, false), "repo_id")
	ErrEquals(t, "repo config not allowed to set 'workflow' key: server-side config needs 'allowed_overrides: [workflow]'", err)
}

func TestParseGlobalCfg_NotExist(t *testing.T) {
	r := yaml.ParserValidator{}
	_, err := r.ParseGlobalCfg("/not/exist", valid.NewGlobalCfg(false, false, false))
	ErrEquals(t, "unable to read /not/exist file: open /not/exist: no such file or directory", err)
}

func TestParseGlobalCfg(t *testing.T) {
	defaultCfg := valid.NewGlobalCfg(false, false, false)
	customWorkflow1 := valid.Workflow{
		Name: "custom1",
		Plan: valid.Stage{
			Steps: []valid.Step{
				{
					StepName:   "run",
					RunCommand: []string{"custom", "command"},
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
		Apply: valid.Stage{
			Steps: []valid.Step{
				{
					StepName:   "run",
					RunCommand: []string{"custom", "command"},
				},
				{
					StepName: "apply",
				},
			},
		},
	}

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
			expErr: "yaml: unmarshal errors:\n  line 1: field invalid not found in struct raw.GlobalCfg",
		},
		"no id specified": {
			input: `repos:
- apply_requirements: []`,
			expErr: "repos: (0: (id: cannot be blank.).).",
		},
		"invalid regex": {
			input: `repos:
- id: /?/`,
			expErr: "repos: (0: (id: parsing: /?/: error parsing regexp: missing argument to repetition operator: `?`.).).",
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
			expErr: "repos: (0: (apply_requirements: \"invalid\" is not a valid apply_requirement, only \"approved\" and \"mergeable\" are supported.).).",
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
						Name:  "name",
						Apply: valid.DefaultApplyStage,
						Plan:  valid.DefaultPlanStage,
					},
				},
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
						Name:  "name",
						Apply: valid.DefaultApplyStage,
						Plan:  valid.DefaultPlanStage,
					},
				},
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
						Name:  "name",
						Plan:  valid.DefaultPlanStage,
						Apply: valid.DefaultApplyStage,
					},
				},
			},
		},
		"all keys specified": {
			input: `
repos:
- id: github.com/owner/repo
  apply_requirements: [approved, mergeable]
  workflow: custom1
  allowed_overrides: [apply_requirements, workflow]
  allow_custom_workflows: true
- id: /.*/

workflows:
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
`,
			exp: valid.GlobalCfg{
				Repos: []valid.Repo{
					defaultCfg.Repos[0],
					{
						ID:                   "github.com/owner/repo",
						ApplyRequirements:    []string{"approved", "mergeable"},
						Workflow:             &customWorkflow1,
						AllowedOverrides:     []string{"apply_requirements", "workflow"},
						AllowCustomWorkflows: Bool(true),
					},
					{
						IDRegex: regexp.MustCompile(".*"),
					},
				},
				Workflows: map[string]valid.Workflow{
					"default": defaultCfg.Workflows["default"],
					"custom1": customWorkflow1,
				},
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
						IDRegex: regexp.MustCompile("github.com/"),
					},
				},
				Workflows: map[string]valid.Workflow{
					"default": defaultCfg.Workflows["default"],
				},
			},
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			r := yaml.ParserValidator{}
			tmp, cleanup := TempDir(t)
			defer cleanup()
			path := filepath.Join(tmp, "conf.yaml")
			Ok(t, ioutil.WriteFile(path, []byte(c.input), 0600))

			act, err := r.ParseGlobalCfg(path, valid.NewGlobalCfg(false, false, false))
			if c.expErr != "" {
				expErr := strings.Replace(c.expErr, "<tmp>", path, -1)
				ErrEquals(t, expErr, err)
				return
			}
			Ok(t, err)
			Equals(t, c.exp, act)
			// Have to hand-compare regexes because Equals doesn't do it.
			for i, actRepo := range act.Repos {
				expRepo := c.exp.Repos[i]
				if expRepo.IDRegex != nil {
					Assert(t, expRepo.IDRegex.String() == actRepo.IDRegex.String(),
						"%q != %q for repos[%d]", expRepo.IDRegex.String(), actRepo.IDRegex.String(), i)
				}
			}
		})
	}
}

// String is a helper routine that allocates a new string value
// to store v and returns a pointer to it.
func String(v string) *string { return &v }

// Bool is a helper routine that allocates a new bool value
// to store v and returns a pointer to it.
func Bool(v bool) *bool { return &v }
