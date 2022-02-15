package events

import (
	"os"
	"path/filepath"
	"testing"

	version "github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/core/config"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/matchers"
	"github.com/runatlantis/atlantis/server/events/models"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	logging_matchers "github.com/runatlantis/atlantis/server/logging/mocks/matchers"
	. "github.com/runatlantis/atlantis/testing"
)

// Test different permutations of global and repo config.
func TestBuildProjectCmdCtx(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	emptyPolicySets := valid.PolicySets{
		Version:    nil,
		PolicySets: []valid.PolicySet{},
	}
	baseRepo := models.Repo{
		FullName: "owner/repo",
		VCSHost: models.VCSHost{
			Hostname: "github.com",
		},
	}
	pull := models.PullRequest{
		BaseRepo: baseRepo,
	}
	cases := map[string]struct {
		globalCfg     string
		repoCfg       string
		expErr        string
		expCtx        models.ProjectCommandContext
		expPlanSteps  []string
		expApplySteps []string
	}{
		// Test that if we've set global defaults and no project config
		// that the global defaults are used.
		"global defaults": {
			globalCfg: `
repos:
- id: /.*/
  workflow: default
workflows:
  default:
    plan:
      steps:
      - init
      - plan
    apply:
      steps:
      - apply`,
			repoCfg: "",
			expCtx: models.ProjectCommandContext{
				ApplyCmd:           "atlantis apply -d project1 -w myworkspace",
				BaseRepo:           baseRepo,
				EscapedCommentArgs: []string{`\f\l\a\g`},
				AutomergeEnabled:   false,
				AutoplanEnabled:    true,
				HeadRepo:           models.Repo{},
				Log:                logger,
				PullReqStatus: models.PullReqStatus{
					Mergeable: true,
				},
				Pull:              pull,
				ProjectName:       "",
				ApplyRequirements: []string{},
				RePlanCmd:         "atlantis plan -d project1 -w myworkspace -- flag",
				RepoRelDir:        "project1",
				User:              models.User{},
				Verbose:           true,
				Workspace:         "myworkspace",
				PolicySets:        emptyPolicySets,
			},
			expPlanSteps:  []string{"init", "plan"},
			expApplySteps: []string{"apply"},
		},

		// Test that if we've set global defaults, that they are used but the
		// allowed project config values also come through.
		"global defaults with repo cfg": {
			globalCfg: `
repos:
- id: /.*/
  workflow: default
workflows:
  default:
    plan:
      steps:
      - init
      - plan
    apply:
      steps:
      - apply`,
			repoCfg: `
version: 3
automerge: true
projects:
- dir: project1
  workspace: myworkspace
  autoplan:
    enabled: true
    when_modified: [../modules/**/*.tf]
  terraform_version: v10.0
  `,
			expCtx: models.ProjectCommandContext{
				ApplyCmd:           "atlantis apply -d project1 -w myworkspace",
				BaseRepo:           baseRepo,
				EscapedCommentArgs: []string{`\f\l\a\g`},
				AutomergeEnabled:   true,
				AutoplanEnabled:    true,
				HeadRepo:           models.Repo{},
				Log:                logger,
				PullReqStatus: models.PullReqStatus{
					Mergeable: true,
				},
				Pull:              pull,
				ProjectName:       "",
				ApplyRequirements: []string{},
				RepoConfigVersion: 3,
				RePlanCmd:         "atlantis plan -d project1 -w myworkspace -- flag",
				RepoRelDir:        "project1",
				TerraformVersion:  mustVersion("10.0"),
				User:              models.User{},
				Verbose:           true,
				Workspace:         "myworkspace",
				PolicySets:        emptyPolicySets,
			},
			expPlanSteps:  []string{"init", "plan"},
			expApplySteps: []string{"apply"},
		},

		// Set a global apply req that should be used.
		"global apply_requirements": {
			globalCfg: `
repos:
- id: /.*/
  workflow: default
  apply_requirements: [approved, mergeable]
workflows:
  default:
    plan:
      steps:
      - init
      - plan
    apply:
      steps:
      - apply`,
			repoCfg: `
version: 3
automerge: true
projects:
- dir: project1
  workspace: myworkspace
  autoplan:
    enabled: true
    when_modified: [../modules/**/*.tf]
  terraform_version: v10.0
`,
			expCtx: models.ProjectCommandContext{
				ApplyCmd:           "atlantis apply -d project1 -w myworkspace",
				BaseRepo:           baseRepo,
				EscapedCommentArgs: []string{`\f\l\a\g`},
				AutomergeEnabled:   true,
				AutoplanEnabled:    true,
				HeadRepo:           models.Repo{},
				Log:                logger,
				PullReqStatus: models.PullReqStatus{
					Mergeable: true,
				},
				Pull:              pull,
				ProjectName:       "",
				ApplyRequirements: []string{"approved", "mergeable"},
				RepoConfigVersion: 3,
				RePlanCmd:         "atlantis plan -d project1 -w myworkspace -- flag",
				RepoRelDir:        "project1",
				TerraformVersion:  mustVersion("10.0"),
				User:              models.User{},
				Verbose:           true,
				Workspace:         "myworkspace",
				PolicySets:        emptyPolicySets,
			},
			expPlanSteps:  []string{"init", "plan"},
			expApplySteps: []string{"apply"},
		},

		// If we have global config that matches a specific repo, it should be used.
		"specific repo": {
			globalCfg: `
repos:
- id: /.*/
  workflow: default
- id: github.com/owner/repo
  workflow: specific
  apply_requirements: [approved]
workflows:
  default:
    plan:
      steps:
      - init
      - plan
    apply:
      steps:
      - apply
  specific:
    plan:
      steps:
      - plan
    apply:
      steps: []`,
			repoCfg: `
version: 3
automerge: true
projects:
- dir: project1
  workspace: myworkspace
  autoplan:
    enabled: true
    when_modified: [../modules/**/*.tf]
  terraform_version: v10.0
`,
			expCtx: models.ProjectCommandContext{
				ApplyCmd:           "atlantis apply -d project1 -w myworkspace",
				BaseRepo:           baseRepo,
				EscapedCommentArgs: []string{`\f\l\a\g`},
				AutomergeEnabled:   true,
				AutoplanEnabled:    true,
				HeadRepo:           models.Repo{},
				Log:                logger,
				PullReqStatus: models.PullReqStatus{
					Mergeable: true,
				},
				Pull:              pull,
				ProjectName:       "",
				ApplyRequirements: []string{"approved"},
				RepoConfigVersion: 3,
				RePlanCmd:         "atlantis plan -d project1 -w myworkspace -- flag",
				RepoRelDir:        "project1",
				TerraformVersion:  mustVersion("10.0"),
				User:              models.User{},
				Verbose:           true,
				Workspace:         "myworkspace",
				PolicySets:        emptyPolicySets,
			},
			expPlanSteps:  []string{"plan"},
			expApplySteps: []string{},
		},

		// We should get an error if the repo sets an apply req when its
		// not allowed.
		"repo defines apply_requirements": {
			globalCfg: `
repos:
- id: /.*/
  workflow: default
  apply_requirements: [approved, mergeable]
workflows:
  default:
    plan:
      steps:
      - init
      - plan
    apply:
      steps:
      - apply`,
			repoCfg: `
version: 3
automerge: true
projects:
- dir: project1
  workspace: myworkspace
  apply_requirements: []
`,
			expErr: "repo config not allowed to set 'apply_requirements' key: server-side config needs 'allowed_overrides: [apply_requirements]'",
		},

		// We should get an error if a repo sets a workflow when it's not allowed.
		"repo sets its own workflow": {
			globalCfg: `
repos:
- id: /.*/
  workflow: default
  apply_requirements: [approved, mergeable]
workflows:
  default:
    plan:
      steps:
      - init
      - plan
    apply:
      steps:
      - apply`,
			repoCfg: `
version: 3
automerge: true
projects:
- dir: project1
  workspace: myworkspace
  workflow: default
`,
			expErr: "repo config not allowed to set 'workflow' key: server-side config needs 'allowed_overrides: [workflow]'",
		},

		// We should get an error if a repo defines a workflow when it's not
		// allowed.
		"repo defines new workflow": {
			globalCfg: `
repos:
- id: /.*/
  workflow: default
  apply_requirements: [approved, mergeable]
workflows:
  default:
    plan:
      steps:
      - init
      - plan
    apply:
      steps:
      - apply`,
			repoCfg: `
version: 3
automerge: true
projects:
- dir: project1
  workspace: myworkspace
workflows:
  new: ~
`,
			expErr: "repo config not allowed to define custom workflows: server-side config needs 'allow_custom_workflows: true'",
		},

		// If the repos are allowed to set everything then their config should
		// come through.
		"full repo permissions": {
			globalCfg: `
repos:
- id: /.*/
  workflow: default
  apply_requirements: [approved]
  allowed_overrides: [apply_requirements, workflow]
  allow_custom_workflows: true
workflows:
  default:
    plan:
      steps: []
    apply:
      steps: []
`,
			repoCfg: `
version: 3
automerge: true
projects:
- dir: project1
  workspace: myworkspace
  autoplan:
    enabled: true
    when_modified: [../modules/**/*.tf]
  terraform_version: v10.0
  apply_requirements: []
  workflow: custom
workflows:
  custom:
    plan:
      steps:
      - plan
    apply:
      steps:
      - apply
`,
			expCtx: models.ProjectCommandContext{
				ApplyCmd:           "atlantis apply -d project1 -w myworkspace",
				BaseRepo:           baseRepo,
				EscapedCommentArgs: []string{`\f\l\a\g`},
				AutomergeEnabled:   true,
				AutoplanEnabled:    true,
				HeadRepo:           models.Repo{},
				Log:                logger,
				PullReqStatus: models.PullReqStatus{
					Mergeable: true,
				},
				Pull:              pull,
				ProjectName:       "",
				ApplyRequirements: []string{},
				RepoConfigVersion: 3,
				RePlanCmd:         "atlantis plan -d project1 -w myworkspace -- flag",
				RepoRelDir:        "project1",
				TerraformVersion:  mustVersion("10.0"),
				User:              models.User{},
				Verbose:           true,
				Workspace:         "myworkspace",
				PolicySets:        emptyPolicySets,
			},
			expPlanSteps:  []string{"plan"},
			expApplySteps: []string{"apply"},
		},

		// Repos can choose server-side workflows.
		"repos choose server-side workflow": {
			globalCfg: `
repos:
- id: /.*/
  workflow: default
  allowed_overrides: [workflow]
workflows:
  default:
    plan:
      steps: []
    apply:
      steps: []
  custom:
    plan:
      steps: [plan]
    apply:
      steps: [apply]
`,
			repoCfg: `
version: 3
automerge: true
projects:
- dir: project1
  workspace: myworkspace
  autoplan:
    enabled: true
    when_modified: [../modules/**/*.tf]
  terraform_version: v10.0
  workflow: custom
`,
			expCtx: models.ProjectCommandContext{
				ApplyCmd:           "atlantis apply -d project1 -w myworkspace",
				BaseRepo:           baseRepo,
				EscapedCommentArgs: []string{`\f\l\a\g`},
				AutomergeEnabled:   true,
				AutoplanEnabled:    true,
				HeadRepo:           models.Repo{},
				Log:                logger,
				PullReqStatus: models.PullReqStatus{
					Mergeable: true,
				},
				Pull:              pull,
				ProjectName:       "",
				ApplyRequirements: []string{},
				RepoConfigVersion: 3,
				RePlanCmd:         "atlantis plan -d project1 -w myworkspace -- flag",
				RepoRelDir:        "project1",
				TerraformVersion:  mustVersion("10.0"),
				User:              models.User{},
				Verbose:           true,
				Workspace:         "myworkspace",
				PolicySets:        emptyPolicySets,
			},
			expPlanSteps:  []string{"plan"},
			expApplySteps: []string{"apply"},
		},

		// Repo-side workflows with the same name override server-side if
		// allowed.
		"repo-side workflow override": {
			globalCfg: `
repos:
- id: /.*/
  workflow: custom
  allowed_overrides: [workflow]
  allow_custom_workflows: true
workflows:
  custom:
    plan:
      steps: [plan]
    apply:
      steps: [apply]
`,
			repoCfg: `
version: 3
automerge: true
projects:
- dir: project1
  workspace: myworkspace
  autoplan:
    enabled: true
    when_modified: [../modules/**/*.tf]
  terraform_version: v10.0
  workflow: custom
workflows:
  custom:
    plan:
      steps: []
    apply:
      steps: []
`,
			expCtx: models.ProjectCommandContext{
				ApplyCmd:           "atlantis apply -d project1 -w myworkspace",
				BaseRepo:           baseRepo,
				EscapedCommentArgs: []string{`\f\l\a\g`},
				AutomergeEnabled:   true,
				AutoplanEnabled:    true,
				HeadRepo:           models.Repo{},
				Log:                logger,
				PullReqStatus: models.PullReqStatus{
					Mergeable: true,
				},
				Pull:              pull,
				ProjectName:       "",
				ApplyRequirements: []string{},
				RepoConfigVersion: 3,
				RePlanCmd:         "atlantis plan -d project1 -w myworkspace -- flag",
				RepoRelDir:        "project1",
				TerraformVersion:  mustVersion("10.0"),
				User:              models.User{},
				Verbose:           true,
				Workspace:         "myworkspace",
				PolicySets:        emptyPolicySets,
			},
			expPlanSteps:  []string{},
			expApplySteps: []string{},
		},
		// Test that if we leave keys undefined, that they don't override.
		"cascading matches": {
			globalCfg: `
repos:
- id: /.*/
  apply_requirements: [approved]
- id: github.com/owner/repo
  workflow: custom
workflows:
  custom:
    plan:
      steps: [plan]
`,
			repoCfg: `
version: 3
projects:
- dir: project1
  workspace: myworkspace
`,
			expCtx: models.ProjectCommandContext{
				ApplyCmd:           "atlantis apply -d project1 -w myworkspace",
				BaseRepo:           baseRepo,
				EscapedCommentArgs: []string{`\f\l\a\g`},
				AutomergeEnabled:   false,
				AutoplanEnabled:    true,
				HeadRepo:           models.Repo{},
				Log:                logger,
				PullReqStatus: models.PullReqStatus{
					Mergeable: true,
				},
				Pull:              pull,
				ProjectName:       "",
				ApplyRequirements: []string{"approved"},
				RepoConfigVersion: 3,
				RePlanCmd:         "atlantis plan -d project1 -w myworkspace -- flag",
				RepoRelDir:        "project1",
				User:              models.User{},
				Verbose:           true,
				Workspace:         "myworkspace",
				PolicySets:        emptyPolicySets,
			},
			expPlanSteps:  []string{"plan"},
			expApplySteps: []string{"apply"},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			tmp, cleanup := DirStructure(t, map[string]interface{}{
				"project1": map[string]interface{}{
					"main.tf": nil,
				},
				"modules": map[string]interface{}{
					"module": map[string]interface{}{
						"main.tf": nil,
					},
				},
			})
			defer cleanup()

			workingDir := NewMockWorkingDir()
			When(workingDir.Clone(matchers.AnyPtrToLoggingSimpleLogger(), matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest(), AnyString())).ThenReturn(tmp, false, nil)
			vcsClient := vcsmocks.NewMockClient()
			When(vcsClient.GetModifiedFiles(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())).ThenReturn([]string{"modules/module/main.tf"}, nil)

			// Write and parse the global config file.
			globalCfgPath := filepath.Join(tmp, "global.yaml")
			Ok(t, os.WriteFile(globalCfgPath, []byte(c.globalCfg), 0600))
			parser := &config.ParserValidator{}
			globalCfgArgs := valid.GlobalCfgArgs{
				AllowRepoCfg:  false,
				MergeableReq:  false,
				ApprovedReq:   false,
				UnDivergedReq: false,
			}
			globalCfg, err := parser.ParseGlobalCfg(globalCfgPath, valid.NewGlobalCfgFromArgs(globalCfgArgs))
			Ok(t, err)

			if c.repoCfg != "" {
				Ok(t, os.WriteFile(filepath.Join(tmp, "atlantis.yaml"), []byte(c.repoCfg), 0600))
			}

			builder := NewProjectCommandBuilder(
				false,
				parser,
				&DefaultProjectFinder{},
				vcsClient,
				workingDir,
				NewDefaultWorkingDirLocker(),
				globalCfg,
				&DefaultPendingPlanFinder{},
				&CommentParser{},
				false,
				false,
				"**/*.tf,**/*.tfvars,**/*.tfvars.json,**/terragrunt.hcl,**/.terraform.lock.hcl",
			)

			// We run a test for each type of command.
			for _, cmd := range []models.CommandName{models.PlanCommand, models.ApplyCommand} {
				t.Run(cmd.String(), func(t *testing.T) {
					ctxs, err := builder.buildProjectCommandCtx(&CommandContext{
						Log: logger,
						Pull: models.PullRequest{
							BaseRepo: baseRepo,
						},
						PullRequestStatus: models.PullReqStatus{
							Mergeable: true,
						},
					}, cmd, "", []string{"flag"}, tmp, "project1", "myworkspace", true)

					if c.expErr != "" {
						ErrEquals(t, c.expErr, err)
						return
					}
					ctx := ctxs[0]

					Ok(t, err)

					// Construct expected steps.
					var stepNames []string
					switch cmd {
					case models.PlanCommand:
						stepNames = c.expPlanSteps
					case models.ApplyCommand:
						stepNames = c.expApplySteps
					}
					var expSteps []valid.Step
					for _, stepName := range stepNames {
						expSteps = append(expSteps, valid.Step{
							StepName: stepName,
						})
					}

					c.expCtx.CommandName = cmd
					// Init fields we couldn't in our cases map.
					c.expCtx.Steps = expSteps
					ctx.PolicySets = emptyPolicySets

					// Job ID cannot be compared since its generated at random
					ctx.JobID = ""

					Equals(t, c.expCtx, ctx)
					// Equals() doesn't compare TF version properly so have to
					// use .String().
					if c.expCtx.TerraformVersion != nil {
						Equals(t, c.expCtx.TerraformVersion.String(), ctx.TerraformVersion.String())
					}
				})
			}
		})
	}
}

func TestBuildProjectCmdCtx_WithRegExpCmdEnabled(t *testing.T) {
	emptyPolicySets := valid.PolicySets{
		Version:    nil,
		PolicySets: []valid.PolicySet{},
	}
	baseRepo := models.Repo{
		FullName: "owner/repo",
		VCSHost: models.VCSHost{
			Hostname: "github.com",
		},
	}
	pull := models.PullRequest{
		BaseRepo: baseRepo,
	}
	cases := map[string]struct {
		globalCfg     string
		repoCfg       string
		expErr        string
		expCtx        models.ProjectCommandContext
		expPlanSteps  []string
		expApplySteps []string
	}{

		// Test that if we've set global defaults, that they are used but the
		// allowed project config values also come through.
		"global defaults with repo cfg": {
			globalCfg: `
repos:
- id: /.*/
  workflow: default
workflows:
  default:
    plan:
      steps:
      - init
      - plan
    apply:
      steps:
      - apply`,
			repoCfg: `
version: 3
automerge: true
projects:
- name: myproject_1
  dir: project1
  workspace: myworkspace
  autoplan:
    enabled: true
    when_modified: [../modules/**/*.tf]
  terraform_version: v10.0
- name: myproject_2
  dir: project2
  workspace: myworkspace
  autoplan:
    enabled: true
    when_modified: [../modules/**/*.tf]
  terraform_version: v10.0
- name: myproject_3
  dir: project3
  workspace: myworkspace
  autoplan:
    enabled: true
    when_modified: [../modules/**/*.tf]
  terraform_version: v10.0
  `,
			expCtx: models.ProjectCommandContext{
				ApplyCmd:           "atlantis apply -p myproject_1",
				BaseRepo:           baseRepo,
				EscapedCommentArgs: []string{`\f\l\a\g`},
				AutomergeEnabled:   true,
				AutoplanEnabled:    true,
				HeadRepo:           models.Repo{},
				Log:                logging.NewNoopLogger(t),
				PullReqStatus: models.PullReqStatus{
					Mergeable: true,
				},
				Pull:              pull,
				ProjectName:       "myproject_1",
				ApplyRequirements: []string{},
				RepoConfigVersion: 3,
				RePlanCmd:         "atlantis plan -p myproject_1 -- flag",
				RepoRelDir:        "project1",
				TerraformVersion:  mustVersion("10.0"),
				User:              models.User{},
				Verbose:           true,
				Workspace:         "myworkspace",
				PolicySets:        emptyPolicySets,
			},
			expPlanSteps:  []string{"init", "plan"},
			expApplySteps: []string{"apply"},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			tmp, cleanup := DirStructure(t, map[string]interface{}{
				"project1": map[string]interface{}{
					"main.tf": nil,
				},
				"modules": map[string]interface{}{
					"module": map[string]interface{}{
						"main.tf": nil,
					},
				},
			})
			defer cleanup()

			workingDir := NewMockWorkingDir()
			When(workingDir.Clone(logging_matchers.AnyPtrToLoggingSimpleLogger(), matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest(), AnyString())).ThenReturn(tmp, false, nil)
			vcsClient := vcsmocks.NewMockClient()
			When(vcsClient.GetModifiedFiles(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())).ThenReturn([]string{"modules/module/main.tf"}, nil)

			// Write and parse the global config file.
			globalCfgPath := filepath.Join(tmp, "global.yaml")
			Ok(t, os.WriteFile(globalCfgPath, []byte(c.globalCfg), 0600))
			parser := &config.ParserValidator{}
			globalCfg, err := parser.ParseGlobalCfg(globalCfgPath, valid.NewGlobalCfg(false, false, false))
			Ok(t, err)

			if c.repoCfg != "" {
				Ok(t, os.WriteFile(filepath.Join(tmp, "atlantis.yaml"), []byte(c.repoCfg), 0600))
			}

			builder := NewProjectCommandBuilder(
				false,
				parser,
				&DefaultProjectFinder{},
				vcsClient,
				workingDir,
				NewDefaultWorkingDirLocker(),
				globalCfg,
				&DefaultPendingPlanFinder{},
				&CommentParser{},
				false,
				true,
				"**/*.tf,**/*.tfvars,**/*.tfvars.json,**/terragrunt.hcl,**/.terraform.lock.hcl",
			)

			// We run a test for each type of command, again specific projects
			for _, cmd := range []models.CommandName{models.PlanCommand, models.ApplyCommand} {
				t.Run(cmd.String(), func(t *testing.T) {
					ctxs, err := builder.buildProjectCommandCtx(&CommandContext{
						Pull: models.PullRequest{
							BaseRepo: baseRepo,
						},
						Log: logging.NewNoopLogger(t),
						PullRequestStatus: models.PullReqStatus{
							Mergeable: true,
						},
					}, cmd, "myproject_[1-2]", []string{"flag"}, tmp, "project1", "myworkspace", true)

					if c.expErr != "" {
						ErrEquals(t, c.expErr, err)
						return
					}
					ctx := ctxs[0]

					Ok(t, err)

					Equals(t, 2, len(ctxs))
					// Construct expected steps.
					var stepNames []string
					switch cmd {
					case models.PlanCommand:
						stepNames = c.expPlanSteps
					case models.ApplyCommand:
						stepNames = c.expApplySteps
					}
					var expSteps []valid.Step
					for _, stepName := range stepNames {
						expSteps = append(expSteps, valid.Step{
							StepName: stepName,
						})
					}

					c.expCtx.CommandName = cmd
					// Init fields we couldn't in our cases map.
					c.expCtx.Steps = expSteps
					ctx.PolicySets = emptyPolicySets

					// Job ID cannot be compared since its generated at random
					ctx.JobID = ""

					Equals(t, c.expCtx, ctx)
					// Equals() doesn't compare TF version properly so have to
					// use .String().
					if c.expCtx.TerraformVersion != nil {
						Equals(t, c.expCtx.TerraformVersion.String(), ctx.TerraformVersion.String())
					}
				})
			}
		})
	}
}

func TestBuildProjectCmdCtx_WithPolicCheckEnabled(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	emptyPolicySets := valid.PolicySets{
		Version:    nil,
		PolicySets: []valid.PolicySet{},
	}
	baseRepo := models.Repo{
		FullName: "owner/repo",
		VCSHost: models.VCSHost{
			Hostname: "github.com",
		},
	}
	pull := models.PullRequest{
		BaseRepo: baseRepo,
	}
	cases := map[string]struct {
		globalCfg           string
		repoCfg             string
		expErr              string
		expCtx              models.ProjectCommandContext
		expPolicyCheckSteps []string
	}{
		// Test that if we've set global defaults and no project config
		// that the global defaults are used.
		"global defaults": {
			globalCfg: `
repos:
- id: /.*/
`,
			repoCfg: "",
			expCtx: models.ProjectCommandContext{
				ApplyCmd:           "atlantis apply -d project1 -w myworkspace",
				BaseRepo:           baseRepo,
				EscapedCommentArgs: []string{`\f\l\a\g`},
				AutomergeEnabled:   false,
				AutoplanEnabled:    true,
				HeadRepo:           models.Repo{},
				Log:                logger,
				PullReqStatus: models.PullReqStatus{
					Mergeable: true,
				},
				Pull:              pull,
				ProjectName:       "",
				ApplyRequirements: []string{},
				RePlanCmd:         "atlantis plan -d project1 -w myworkspace -- flag",
				RepoRelDir:        "project1",
				User:              models.User{},
				Verbose:           true,
				Workspace:         "myworkspace",
				PolicySets:        emptyPolicySets,
			},
			expPolicyCheckSteps: []string{"show", "policy_check"},
		},

		// If the repos are allowed to set everything then their config should
		// come through.
		"full repo permissions": {
			globalCfg: `
repos:
- id: /.*/
  workflow: default
  apply_requirements: [approved]
  allowed_overrides: [apply_requirements, workflow]
  allow_custom_workflows: true
workflows:
  default:
    policy_check:
      steps: []
`,
			repoCfg: `
version: 3
automerge: true
projects:
- dir: project1
  workspace: myworkspace
  autoplan:
    enabled: true
    when_modified: [../modules/**/*.tf]
  terraform_version: v10.0
  apply_requirements: []
  workflow: custom
workflows:
  custom:
    policy_check:
      steps:
      - policy_check
`,
			expCtx: models.ProjectCommandContext{
				ApplyCmd:           "atlantis apply -d project1 -w myworkspace",
				BaseRepo:           baseRepo,
				EscapedCommentArgs: []string{`\f\l\a\g`},
				AutomergeEnabled:   true,
				AutoplanEnabled:    true,
				HeadRepo:           models.Repo{},
				Log:                logger,
				PullReqStatus: models.PullReqStatus{
					Mergeable: true,
				},
				Pull:              pull,
				ProjectName:       "",
				ApplyRequirements: []string{},
				RepoConfigVersion: 3,
				RePlanCmd:         "atlantis plan -d project1 -w myworkspace -- flag",
				RepoRelDir:        "project1",
				TerraformVersion:  mustVersion("10.0"),
				User:              models.User{},
				Verbose:           true,
				Workspace:         "myworkspace",
				PolicySets:        emptyPolicySets,
			},
			expPolicyCheckSteps: []string{"policy_check"},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			tmp, cleanup := DirStructure(t, map[string]interface{}{
				"project1": map[string]interface{}{
					"main.tf": nil,
				},
				"modules": map[string]interface{}{
					"module": map[string]interface{}{
						"main.tf": nil,
					},
				},
			})
			defer cleanup()

			workingDir := NewMockWorkingDir()
			When(workingDir.Clone(matchers.AnyPtrToLoggingSimpleLogger(), matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest(), AnyString())).ThenReturn(tmp, false, nil)
			vcsClient := vcsmocks.NewMockClient()
			When(vcsClient.GetModifiedFiles(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())).ThenReturn([]string{"modules/module/main.tf"}, nil)

			// Write and parse the global config file.
			globalCfgPath := filepath.Join(tmp, "global.yaml")
			Ok(t, os.WriteFile(globalCfgPath, []byte(c.globalCfg), 0600))
			parser := &config.ParserValidator{}
			globalCfgArgs := valid.GlobalCfgArgs{
				AllowRepoCfg:  false,
				MergeableReq:  false,
				ApprovedReq:   false,
				UnDivergedReq: false,
			}

			globalCfg, err := parser.ParseGlobalCfg(globalCfgPath, valid.NewGlobalCfgFromArgs(globalCfgArgs))
			Ok(t, err)

			if c.repoCfg != "" {
				Ok(t, os.WriteFile(filepath.Join(tmp, "atlantis.yaml"), []byte(c.repoCfg), 0600))
			}

			builder := NewProjectCommandBuilder(
				true,
				parser,
				&DefaultProjectFinder{},
				vcsClient,
				workingDir,
				NewDefaultWorkingDirLocker(),
				globalCfg,
				&DefaultPendingPlanFinder{},
				&CommentParser{},
				false,
				false,
				"**/*.tf,**/*.tfvars,**/*.tfvars.json,**/terragrunt.hcl,**/.terraform.lock.hcl",
			)

			cmd := models.PolicyCheckCommand
			t.Run(cmd.String(), func(t *testing.T) {
				ctxs, err := builder.buildProjectCommandCtx(&CommandContext{
					Log: logger,
					Pull: models.PullRequest{
						BaseRepo: baseRepo,
					},
					PullRequestStatus: models.PullReqStatus{
						Mergeable: true,
					},
				}, models.PlanCommand, "", []string{"flag"}, tmp, "project1", "myworkspace", true)

				if c.expErr != "" {
					ErrEquals(t, c.expErr, err)
					return
				}

				ctx := ctxs[1]

				Ok(t, err)

				// Construct expected steps.
				var stepNames []string
				var expSteps []valid.Step

				stepNames = c.expPolicyCheckSteps
				for _, stepName := range stepNames {
					expSteps = append(expSteps, valid.Step{
						StepName: stepName,
					})
				}

				c.expCtx.CommandName = cmd
				// Init fields we couldn't in our cases map.
				c.expCtx.Steps = expSteps
				ctx.PolicySets = emptyPolicySets

				// Job ID cannot be compared since its generated at random
				ctx.JobID = ""

				Equals(t, c.expCtx, ctx)
				// Equals() doesn't compare TF version properly so have to
				// use .String().
				if c.expCtx.TerraformVersion != nil {
					Equals(t, c.expCtx.TerraformVersion.String(), ctx.TerraformVersion.String())
				}
			})
		})
	}
}

func mustVersion(v string) *version.Version {
	vers, err := version.NewVersion(v)
	if err != nil {
		panic(err)
	}
	return vers
}
