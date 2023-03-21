package valid_test

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/graymeta/stow"
	"github.com/graymeta/stow/local"
	"github.com/hashicorp/go-version"
	"github.com/mohae/deepcopy"
	"github.com/runatlantis/atlantis/server/core/config"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	. "github.com/runatlantis/atlantis/testing"
)

func TestNewGlobalCfg(t *testing.T) {
	expDefaultWorkflow := valid.Workflow{
		Name: "default",
		Apply: valid.Stage{
			Steps: []valid.Step{
				{
					StepName: "apply",
				},
			},
		},
		PolicyCheck: valid.Stage{
			Steps: []valid.Step{
				{
					StepName: "show",
				},
				{
					StepName: "policy_check",
				},
			},
		},
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
	}
	expDefaultPRWorkflow := valid.Workflow{
		Name: "default",
		PolicyCheck: valid.Stage{
			Steps: []valid.Step{
				{
					StepName: "show",
				},
				{
					StepName: "policy_check",
				},
			},
		},
		Plan: valid.Stage{
			Steps: []valid.Step{
				{
					StepName: "init",
				},
				{
					StepName:  "plan",
					ExtraArgs: []string{"-lock=false"},
				},
			},
		},
	}
	expDefaultDeploymentWorkflow := valid.Workflow{
		Name: "default",
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
				{
					StepName: "plan",
				},
			},
		},
	}

	rebaseEnabled := true
	baseCfg := valid.GlobalCfg{
		Repos: []valid.Repo{
			{
				IDRegex:              regexp.MustCompile(".*"),
				BranchRegex:          regexp.MustCompile(".*"),
				ApplyRequirements:    []string{},
				Workflow:             &expDefaultWorkflow,
				PullRequestWorkflow:  &expDefaultPRWorkflow,
				DeploymentWorkflow:   &expDefaultDeploymentWorkflow,
				AllowedWorkflows:     []string{},
				AllowedOverrides:     []string{},
				AllowCustomWorkflows: Bool(false),
				CheckoutStrategy:     "branch",
				RebaseEnabled:        &rebaseEnabled,
			},
		},
		Workflows: map[string]valid.Workflow{
			"default": expDefaultWorkflow,
		},
		PullRequestWorkflows: map[string]valid.Workflow{
			"default": expDefaultPRWorkflow,
		},
		DeploymentWorkflows: map[string]valid.Workflow{
			"default": expDefaultDeploymentWorkflow,
		},
		PersistenceConfig: valid.PersistenceConfig{
			Deployments: valid.StoreConfig{
				BackendType: valid.LocalBackend,
				Prefix:      valid.DefaultDeploymentsPrefix,
				Config: stow.ConfigMap{
					local.ConfigKeyPath: "somedir",
				},
				ContainerName: valid.LocalStore,
			},
			Jobs: valid.StoreConfig{
				BackendType: valid.LocalBackend,
				Prefix:      valid.DefaultJobsPrefix,
				Config: stow.ConfigMap{
					local.ConfigKeyPath: "somedir",
				},
				ContainerName: valid.LocalStore,
			},
		},
	}

	t.Run("new global config", func(t *testing.T) {
		act := valid.NewGlobalCfg("somedir")
		// For each test, we change our expected cfg based on the parameters.
		exp := deepcopy.Copy(baseCfg).(valid.GlobalCfg)
		exp.Repos[0].IDRegex = regexp.MustCompile(".*") // deepcopy doesn't copy the regex.
		exp.Repos[0].BranchRegex = regexp.MustCompile(".*")

		Equals(t, exp, act)

		// Have to hand-compare regexes because Equals doesn't do it.
		for i, actRepo := range act.Repos {
			expRepo := exp.Repos[i]
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

func TestGlobalCfg_ValidateRepoCfg(t *testing.T) {
	cases := map[string]struct {
		gCfg   valid.GlobalCfg
		rCfg   valid.RepoCfg
		repoID string
		expErr string
	}{
		"repo uses workflow that is defined server side but not allowed (with custom workflows)": {
			gCfg: valid.GlobalCfg{
				Repos: []valid.Repo{
					valid.NewGlobalCfg("somedir").Repos[0],
					{
						ID:                   "github.com/owner/repo",
						AllowCustomWorkflows: Bool(true),
						AllowedOverrides:     []string{"workflow"},
						AllowedWorkflows:     []string{"allowed"},
					},
				},
				Workflows: map[string]valid.Workflow{
					"allowed":   {},
					"forbidden": {},
				},
			},
			rCfg: valid.RepoCfg{
				Projects: []valid.Project{
					{
						Dir:          ".",
						Workspace:    "default",
						WorkflowName: String("forbidden"),
					},
				},
			},
			repoID: "github.com/owner/repo",
			expErr: "workflow \"forbidden\" is not allowed for this repo",
		},
		"repo uses workflow that is defined server side but not allowed (without custom workflows)": {
			gCfg: valid.GlobalCfg{
				Repos: []valid.Repo{
					valid.NewGlobalCfg("somedir").Repos[0],
					{
						ID:                   "github.com/owner/repo",
						AllowCustomWorkflows: Bool(false),
						AllowedOverrides:     []string{"workflow"},
						AllowedWorkflows:     []string{"allowed"},
					},
				},
				Workflows: map[string]valid.Workflow{
					"allowed":   {},
					"forbidden": {},
				},
			},
			rCfg: valid.RepoCfg{
				Projects: []valid.Project{
					{
						Dir:          ".",
						Workspace:    "default",
						WorkflowName: String("forbidden"),
					},
				},
			},
			repoID: "github.com/owner/repo",
			expErr: "workflow \"forbidden\" is not allowed for this repo",
		},
		"repo uses workflow that is defined in both places with same name (without custom workflows)": {
			gCfg: valid.GlobalCfg{
				Repos: []valid.Repo{
					valid.NewGlobalCfg("somedir").Repos[0],
					{
						ID:                   "github.com/owner/repo",
						AllowCustomWorkflows: Bool(false),
						AllowedOverrides:     []string{"workflow"},
						AllowedWorkflows:     []string{"duplicated"},
					},
				},
				Workflows: map[string]valid.Workflow{
					"duplicated": {},
				},
			},
			rCfg: valid.RepoCfg{
				Projects: []valid.Project{
					{
						Dir:          ".",
						Workspace:    "default",
						WorkflowName: String("duplicated"),
					},
				},
				Workflows: map[string]valid.Workflow{
					"duplicated": {},
				},
			},
			repoID: "github.com/owner/repo",
			expErr: "repo config not allowed to define custom workflows: server-side config needs 'allow_custom_workflows: true'",
		},
		"repo uses workflow that is defined repo side, but not allowed (with custom workflows)": {
			gCfg: valid.GlobalCfg{
				Repos: []valid.Repo{
					valid.NewGlobalCfg("somedir").Repos[0],
					{
						ID:                   "github.com/owner/repo",
						AllowCustomWorkflows: Bool(true),
						AllowedOverrides:     []string{"workflow"},
						AllowedWorkflows:     []string{"none"},
					},
				},
				Workflows: map[string]valid.Workflow{
					"forbidden": {},
				},
			},
			rCfg: valid.RepoCfg{
				Projects: []valid.Project{
					{
						Dir:          ".",
						Workspace:    "default",
						WorkflowName: String("repodefined"),
					},
				},
				Workflows: map[string]valid.Workflow{
					"repodefined": {},
				},
			},
			repoID: "github.com/owner/repo",
			expErr: "",
		},
		"repo uses workflow that is defined server side and allowed (without custom workflows)": {
			gCfg: valid.GlobalCfg{
				Repos: []valid.Repo{
					valid.NewGlobalCfg("somedir").Repos[0],
					{
						ID:                   "github.com/owner/repo",
						AllowCustomWorkflows: Bool(false),
						AllowedOverrides:     []string{"workflow"},
						AllowedWorkflows:     []string{"allowed"},
					},
				},
				Workflows: map[string]valid.Workflow{
					"allowed":   {},
					"forbidden": {},
				},
			},
			rCfg: valid.RepoCfg{
				Projects: []valid.Project{
					{
						Dir:          ".",
						Workspace:    "default",
						WorkflowName: String("allowed"),
					},
				},
			},
			repoID: "github.com/owner/repo",
			expErr: "",
		},
		"repo uses workflow that is defined server side and allowed (with custom workflows)": {
			gCfg: valid.GlobalCfg{
				Repos: []valid.Repo{
					valid.NewGlobalCfg("somedir").Repos[0],
					{
						ID:                   "github.com/owner/repo",
						AllowCustomWorkflows: Bool(true),
						AllowedOverrides:     []string{"workflow"},
						AllowedWorkflows:     []string{"allowed"},
					},
				},
				Workflows: map[string]valid.Workflow{
					"allowed":   {},
					"forbidden": {},
				},
			},
			rCfg: valid.RepoCfg{
				Projects: []valid.Project{
					{
						Dir:          ".",
						Workspace:    "default",
						WorkflowName: String("allowed"),
					},
				},
			},
			repoID: "github.com/owner/repo",
			expErr: "",
		},
		"workflow not allowed": {
			gCfg: valid.NewGlobalCfg("somedir"),
			rCfg: valid.RepoCfg{
				Projects: []valid.Project{
					{
						WorkflowName: String("invalid"),
					},
				},
			},
			repoID: "github.com/owner/repo",
			expErr: "repo config not allowed to set 'workflow' key: server-side config needs 'allowed_overrides: [workflow]'",
		},
		"custom workflows not allowed": {
			gCfg: valid.NewGlobalCfg("somedir"),
			rCfg: valid.RepoCfg{
				Workflows: map[string]valid.Workflow{
					"custom": {},
				},
			},
			repoID: "github.com/owner/repo",
			expErr: "repo config not allowed to define custom workflows: server-side config needs 'allow_custom_workflows: true'",
		},
		"custom workflows allowed": {
			gCfg: valid.GlobalCfg{
				Repos: []valid.Repo{
					{
						IDRegex:              regexp.MustCompile(".*"),
						BranchRegex:          regexp.MustCompile(".*"),
						AllowCustomWorkflows: Bool(true),
					},
				},
			},
			rCfg: valid.RepoCfg{
				Workflows: map[string]valid.Workflow{
					"custom": {},
				},
			},
			repoID: "github.com/owner/repo",
			expErr: "",
		},
		"repo uses custom workflow defined on repo": {
			gCfg: valid.GlobalCfg{
				Repos: []valid.Repo{
					{
						IDRegex:              regexp.MustCompile(".*"),
						AllowCustomWorkflows: Bool(true),
						AllowedOverrides:     []string{"workflow", "pull_request_workflow", "deployment_workflow"},
					},
				},
			},

			rCfg: valid.RepoCfg{
				Projects: []valid.Project{
					{
						Dir:          ".",
						Workspace:    "default",
						WorkflowName: String("repodefined"),
					},
				},
				Workflows: map[string]valid.Workflow{
					"repodefined": {},
				},
				PullRequestWorkflows: map[string]valid.Workflow{
					"repodefined": {},
				},
				DeploymentWorkflows: map[string]valid.Workflow{
					"repodefined": {},
				},
			},
			repoID: "github.com/owner/repo",
			expErr: "",
		},
		"custom workflows allowed for this repo only": {
			gCfg: valid.GlobalCfg{
				Repos: []valid.Repo{
					valid.NewGlobalCfg("somedir").Repos[0],
					{
						ID:                   "github.com/owner/repo",
						AllowCustomWorkflows: Bool(true),
					},
				},
			},
			rCfg: valid.RepoCfg{
				Workflows: map[string]valid.Workflow{
					"custom": {},
				},
			},
			repoID: "github.com/owner/repo",
			expErr: "",
		},
		"repo uses global workflow": {
			gCfg: valid.GlobalCfg{
				Repos: []valid.Repo{
					{
						IDRegex:              regexp.MustCompile(".*"),
						BranchRegex:          regexp.MustCompile(".*"),
						AllowedOverrides:     []string{"workflow", "pull_request_workflow", "deployment_workflow"},
						AllowCustomWorkflows: Bool(false),
					},
				},
				Workflows: map[string]valid.Workflow{
					"default": {
						Name:        valid.DefaultWorkflowName,
						Apply:       valid.DefaultApplyStage,
						Plan:        valid.DefaultPlanStage,
						PolicyCheck: valid.DefaultPolicyCheckStage,
					},
				},
				PullRequestWorkflows: map[string]valid.Workflow{
					"default": {
						Name:        "default",
						PolicyCheck: valid.DefaultPolicyCheckStage,
						Plan:        valid.DefaultLocklessPlanStage,
					},
				},
				DeploymentWorkflows: map[string]valid.Workflow{
					"default": {

						Name:  "default",
						Apply: valid.DefaultApplyStage,
						Plan:  valid.DefaultPlanStage,
					},
				},
			},
			rCfg: valid.RepoCfg{
				Projects: []valid.Project{
					{
						Dir:                     ".",
						Workspace:               "default",
						WorkflowName:            String("default"),
						PullRequestWorkflowName: String("default"),
						DeploymentWorkflowName:  String("default"),
					},
				},
			},
			repoID: "github.com/owner/repo",
			expErr: "",
		},
		"apply_reqs not allowed": {
			gCfg: valid.GlobalCfg{
				Repos: []valid.Repo{
					{
						IDRegex:     regexp.MustCompile(".*"),
						BranchRegex: regexp.MustCompile(".*"),
					},
				},
			},
			rCfg: valid.RepoCfg{
				Projects: []valid.Project{
					{
						Dir:               ".",
						Workspace:         "default",
						ApplyRequirements: []string{""},
					},
				},
			},
			repoID: "github.com/owner/repo",
			expErr: "repo config not allowed to set 'apply_requirements' key: server-side config needs 'allowed_overrides: [apply_requirements]'",
		},
		"repo workflow doesn't exist": {
			gCfg: valid.GlobalCfg{
				Repos: []valid.Repo{
					{
						IDRegex:              regexp.MustCompile(".*"),
						AllowCustomWorkflows: Bool(true),
						AllowedOverrides:     []string{"workflow"},
					},
				},
			},
			rCfg: valid.RepoCfg{
				Projects: []valid.Project{
					{
						Dir:          ".",
						Workspace:    "default",
						WorkflowName: String("doesntexist"),
					},
				},
			},
			repoID: "github.com/owner/repo",
			expErr: "workflow \"doesntexist\" is not defined anywhere",
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			actErr := c.gCfg.ValidateRepoCfg(c.rCfg, c.repoID)
			if c.expErr == "" {
				Ok(t, actErr)
			} else {
				ErrEquals(t, c.expErr, actErr)
			}
		})
	}
}

func TestGlobalCfg_WithPolicySets(t *testing.T) {
	version, _ := version.NewVersion("v1.0.0")
	cases := map[string]struct {
		gCfg   string
		proj   valid.Project
		repoID string
		exp    valid.MergedProjectCfg
	}{
		"policies are added to MergedProjectCfg when present": {
			gCfg: `
repos:
- id: /.*/
policies:
  policy_sets:
    - name: good-policy
      source: local
      path: rel/path/to/source
`,
			repoID: "github.com/owner/repo",
			proj: valid.Project{
				Dir:          ".",
				Workspace:    "default",
				WorkflowName: String("custom"),
			},
			exp: valid.MergedProjectCfg{
				ApplyRequirements: []string{"policies_passed"},
				Workflow: valid.Workflow{
					Name:        "default",
					Apply:       valid.DefaultApplyStage,
					Plan:        valid.DefaultPlanStage,
					PolicyCheck: valid.DefaultPolicyCheckStage,
				},
				PullRequestWorkflow: valid.Workflow{
					Name:        "default",
					PolicyCheck: valid.DefaultPolicyCheckStage,
					Plan:        valid.DefaultLocklessPlanStage,
				},
				DeploymentWorkflow: valid.Workflow{
					Name:  "default",
					Apply: valid.DefaultApplyStage,
					Plan:  valid.DefaultPlanStage,
				},
				PolicySets: valid.PolicySets{
					Version: nil,
					PolicySets: []valid.PolicySet{
						{
							Name:   "good-policy",
							Path:   "rel/path/to/source",
							Source: "local",
						},
					},
				},
				RepoRelDir:      ".",
				Workspace:       "default",
				Name:            "",
				AutoplanEnabled: false,
				RebaseEnabled:   true,
			},
		},
		"policies set correct version if specified": {
			gCfg: `
repos:
- id: /.*/
policies:
  conftest_version: v1.0.0
  policy_sets:
    - name: good-policy
      source: local
      path: rel/path/to/source
`,
			repoID: "github.com/owner/repo",
			proj: valid.Project{
				Dir:          ".",
				Workspace:    "default",
				WorkflowName: String("custom"),
			},
			exp: valid.MergedProjectCfg{
				ApplyRequirements: []string{"policies_passed"},
				Workflow: valid.Workflow{
					Name:        "default",
					Apply:       valid.DefaultApplyStage,
					Plan:        valid.DefaultPlanStage,
					PolicyCheck: valid.DefaultPolicyCheckStage,
				},
				PullRequestWorkflow: valid.Workflow{
					Name:        "default",
					PolicyCheck: valid.DefaultPolicyCheckStage,
					Plan:        valid.DefaultLocklessPlanStage,
				},
				DeploymentWorkflow: valid.Workflow{
					Name:  "default",
					Apply: valid.DefaultApplyStage,
					Plan:  valid.DefaultPlanStage,
				},
				PolicySets: valid.PolicySets{
					Version: version,
					PolicySets: []valid.PolicySet{
						{
							Name:   "good-policy",
							Path:   "rel/path/to/source",
							Source: "local",
						},
					},
				},
				RepoRelDir:      ".",
				Workspace:       "default",
				Name:            "",
				AutoplanEnabled: false,
				RebaseEnabled:   true,
			},
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			tmp, cleanup := TempDir(t)
			defer cleanup()
			var global valid.GlobalCfg
			if c.gCfg != "" {
				path := filepath.Join(tmp, "config.yaml")
				Ok(t, os.WriteFile(path, []byte(c.gCfg), 0600))
				var err error
				global, err = (&config.ParserValidator{}).ParseGlobalCfg(path, valid.NewGlobalCfg("somedir"))
				Ok(t, err)
			} else {
				global = valid.NewGlobalCfg("somedir")
			}

			Equals(t,
				c.exp,
				global.MergeProjectCfg(c.repoID, c.proj, valid.RepoCfg{}))
		})
	}
}

func TestGlobalCfg_MergeProjectCfg(t *testing.T) {
	var emptyPolicySets valid.PolicySets

	cases := map[string]struct {
		gCfg          string
		repoID        string
		proj          valid.Project
		repoWorkflows map[string]valid.Workflow
		exp           valid.MergedProjectCfg
	}{
		"repos can use server-side defined workflow if allowed": {
			gCfg: `
repos:
- id: /.*/
  allowed_overrides: [workflow]
workflows:
  custom:
    plan:
      steps: [plan]`,
			repoID: "github.com/owner/repo",
			proj: valid.Project{
				Dir:          ".",
				Workspace:    "default",
				WorkflowName: String("custom"),
			},
			repoWorkflows: nil,
			exp: valid.MergedProjectCfg{
				ApplyRequirements: []string{},
				Workflow: valid.Workflow{
					Name:        "custom",
					Apply:       valid.DefaultApplyStage,
					PolicyCheck: valid.DefaultPolicyCheckStage,
					Plan: valid.Stage{
						Steps: []valid.Step{
							{
								StepName: "plan",
							},
						},
					},
				},
				PullRequestWorkflow: valid.Workflow{
					Name:        "default",
					PolicyCheck: valid.DefaultPolicyCheckStage,
					Plan:        valid.DefaultLocklessPlanStage,
				},
				DeploymentWorkflow: valid.Workflow{
					Name:  "default",
					Apply: valid.DefaultApplyStage,
					Plan:  valid.DefaultPlanStage,
				},
				RepoRelDir:      ".",
				Workspace:       "default",
				Name:            "",
				AutoplanEnabled: false,
				PolicySets:      emptyPolicySets,
				RebaseEnabled:   true,
			},
		},
		"repos can use server-side defined pr and deployment workflow if allowed": {
			gCfg: `
repos:
- id: /.*/
  allowed_overrides: [workflow, pull_request_workflow, deployment_workflow]
pull_request_workflows:
  custom:
    plan:
      steps: [plan]
deployment_workflows:
  custom:
    plan:
      steps: [plan]
    apply:
      steps: []`,
			repoID: "github.com/owner/repo",
			proj: valid.Project{
				Dir:                     ".",
				Workspace:               "default",
				PullRequestWorkflowName: String("custom"),
				DeploymentWorkflowName:  String("custom"),
			},
			repoWorkflows: nil,
			exp: valid.MergedProjectCfg{
				ApplyRequirements: []string{},
				Workflow: valid.Workflow{
					Name:        "default",
					Apply:       valid.DefaultApplyStage,
					PolicyCheck: valid.DefaultPolicyCheckStage,
					Plan:        valid.DefaultPlanStage,
				},
				PullRequestWorkflow: valid.Workflow{
					Name:        "custom",
					PolicyCheck: valid.DefaultPolicyCheckStage,
					Plan: valid.Stage{
						Steps: []valid.Step{
							{
								StepName: "plan",
							},
						},
					},
				},
				DeploymentWorkflow: valid.Workflow{
					Name: "custom",
					Plan: valid.Stage{
						Steps: []valid.Step{
							{
								StepName: "plan",
							},
						},
					},
					Apply: valid.Stage{},
				},
				RepoRelDir:      ".",
				Workspace:       "default",
				Name:            "",
				AutoplanEnabled: false,
				PolicySets:      emptyPolicySets,
				RebaseEnabled:   true,
			},
		},
		"repo-side apply reqs win out if allowed": {
			gCfg: `
repos:
- id: /.*/
  allowed_overrides: [apply_requirements]
  apply_requirements: [approved]
`,
			repoID: "github.com/owner/repo",
			proj: valid.Project{
				Dir:               ".",
				Workspace:         "default",
				ApplyRequirements: []string{"mergeable"},
			},
			repoWorkflows: nil,
			exp: valid.MergedProjectCfg{
				ApplyRequirements: []string{"mergeable"},
				Workflow: valid.Workflow{
					Name:        "default",
					Apply:       valid.DefaultApplyStage,
					PolicyCheck: valid.DefaultPolicyCheckStage,
					Plan:        valid.DefaultPlanStage,
				},
				PullRequestWorkflow: valid.Workflow{
					Name:        "default",
					PolicyCheck: valid.DefaultPolicyCheckStage,
					Plan:        valid.DefaultLocklessPlanStage,
				},
				DeploymentWorkflow: valid.Workflow{
					Name:  "default",
					Apply: valid.DefaultApplyStage,
					Plan:  valid.DefaultPlanStage,
				},
				RepoRelDir:      ".",
				Workspace:       "default",
				Name:            "",
				AutoplanEnabled: false,
				PolicySets:      emptyPolicySets,
				RebaseEnabled:   true,
			},
		},
		"last server-side match wins": {
			gCfg: `
repos:
- id: /.*/
  apply_requirements: [approved]
- id: /github.com/.*/
  apply_requirements: [mergeable]
- id: github.com/owner/repo
  apply_requirements: [approved, mergeable]
`,
			repoID: "github.com/owner/repo",
			proj: valid.Project{
				Dir:       "mydir",
				Workspace: "myworkspace",
				Name:      String("myname"),
			},
			repoWorkflows: nil,
			exp: valid.MergedProjectCfg{
				ApplyRequirements: []string{"approved", "mergeable"},
				Workflow: valid.Workflow{
					Name:        "default",
					Apply:       valid.DefaultApplyStage,
					PolicyCheck: valid.DefaultPolicyCheckStage,
					Plan:        valid.DefaultPlanStage,
				},
				PullRequestWorkflow: valid.Workflow{
					Name:        "default",
					PolicyCheck: valid.DefaultPolicyCheckStage,
					Plan:        valid.DefaultLocklessPlanStage,
				},
				DeploymentWorkflow: valid.Workflow{
					Name:  "default",
					Apply: valid.DefaultApplyStage,
					Plan:  valid.DefaultPlanStage,
				},
				RepoRelDir:      "mydir",
				Workspace:       "myworkspace",
				Name:            "myname",
				AutoplanEnabled: false,
				PolicySets:      emptyPolicySets,
				RebaseEnabled:   true,
			},
		},
		"autoplan is set properly": {
			gCfg:   "",
			repoID: "github.com/owner/repo",
			proj: valid.Project{
				Dir:       "mydir",
				Workspace: "myworkspace",
				Name:      String("myname"),
				Autoplan: valid.Autoplan{
					WhenModified: []string{".tf"},
					Enabled:      true,
				},
			},
			repoWorkflows: nil,
			exp: valid.MergedProjectCfg{
				ApplyRequirements: []string{},
				Workflow: valid.Workflow{
					Name:        "default",
					Apply:       valid.DefaultApplyStage,
					PolicyCheck: valid.DefaultPolicyCheckStage,
					Plan:        valid.DefaultPlanStage,
				},
				PullRequestWorkflow: valid.Workflow{
					Name:        "default",
					PolicyCheck: valid.DefaultPolicyCheckStage,
					Plan:        valid.DefaultLocklessPlanStage,
				},
				DeploymentWorkflow: valid.Workflow{
					Name:  "default",
					Apply: valid.DefaultApplyStage,
					Plan:  valid.DefaultPlanStage,
				},
				RepoRelDir:      "mydir",
				Workspace:       "myworkspace",
				Name:            "myname",
				AutoplanEnabled: true,
				WhenModified:    []string{".tf"},
				PolicySets:      emptyPolicySets,
				RebaseEnabled:   true,
			},
		},
		"merge platform mode default config": {
			gCfg:   "",
			repoID: "github.com/owner/repo",
			proj: valid.Project{
				Dir:       "mydir",
				Workspace: "myworkspace",
				Name:      String("myname"),
				Autoplan: valid.Autoplan{
					WhenModified: []string{".tf"},
					Enabled:      true,
				},
			},
			exp: valid.MergedProjectCfg{
				ApplyRequirements: []string{},
				Workflow: valid.Workflow{
					Name:        "default",
					Apply:       valid.DefaultApplyStage,
					PolicyCheck: valid.DefaultPolicyCheckStage,
					Plan:        valid.DefaultPlanStage,
				},
				PullRequestWorkflow: valid.Workflow{
					Name:        "default",
					PolicyCheck: valid.DefaultPolicyCheckStage,
					Plan:        valid.DefaultLocklessPlanStage,
				},
				DeploymentWorkflow: valid.Workflow{
					Name:  "default",
					Apply: valid.DefaultApplyStage,
					Plan:  valid.DefaultPlanStage,
				},
				RepoRelDir:      "mydir",
				Workspace:       "myworkspace",
				Name:            "myname",
				AutoplanEnabled: true,
				WhenModified:    []string{".tf"},
				PolicySets:      emptyPolicySets,
				RebaseEnabled:   true,
			},
		},
		"project level override of workflow mode": {
			gCfg:   "",
			repoID: "github.com/owner/repo",
			proj: valid.Project{
				Dir:       "mydir",
				Workspace: "myworkspace",
				Name:      String("myname"),
				Autoplan: valid.Autoplan{
					WhenModified: []string{".tf"},
					Enabled:      true,
				},
				WorkflowModeType: String("platform"),
			},
			exp: valid.MergedProjectCfg{
				ApplyRequirements: []string{},
				Workflow: valid.Workflow{
					Name:        "default",
					Apply:       valid.DefaultApplyStage,
					PolicyCheck: valid.DefaultPolicyCheckStage,
					Plan:        valid.DefaultPlanStage,
				},
				PullRequestWorkflow: valid.Workflow{
					Name:        "default",
					PolicyCheck: valid.DefaultPolicyCheckStage,
					Plan:        valid.DefaultLocklessPlanStage,
				},
				DeploymentWorkflow: valid.Workflow{
					Name:  "default",
					Apply: valid.DefaultApplyStage,
					Plan:  valid.DefaultPlanStage,
				},
				RepoRelDir:      "mydir",
				Workspace:       "myworkspace",
				Name:            "myname",
				AutoplanEnabled: true,
				WhenModified:    []string{".tf"},
				PolicySets:      emptyPolicySets,
				RebaseEnabled:   true,
				WorkflowMode:    valid.PlatformWorkflowMode,
			},
		},
		"disable rebase when configured": {
			gCfg: `
repos:
- id: /.*/
  apply_requirements: [approved]
- id: /github.com/.*/
  apply_requirements: [mergeable]
- id: github.com/owner/repo
  rebase_enabled: false
  apply_requirements: [approved, mergeable]
`,
			repoID: "github.com/owner/repo",
			proj: valid.Project{
				Dir:       "mydir",
				Workspace: "myworkspace",
				Name:      String("myname"),
			},
			repoWorkflows: nil,
			exp: valid.MergedProjectCfg{
				ApplyRequirements: []string{"approved", "mergeable"},
				Workflow: valid.Workflow{
					Name:        "default",
					Apply:       valid.DefaultApplyStage,
					PolicyCheck: valid.DefaultPolicyCheckStage,
					Plan:        valid.DefaultPlanStage,
				},
				PullRequestWorkflow: valid.Workflow{
					Name:        "default",
					PolicyCheck: valid.DefaultPolicyCheckStage,
					Plan:        valid.DefaultLocklessPlanStage,
				},
				DeploymentWorkflow: valid.Workflow{
					Name:  "default",
					Apply: valid.DefaultApplyStage,
					Plan:  valid.DefaultPlanStage,
				},
				RepoRelDir:      "mydir",
				Workspace:       "myworkspace",
				Name:            "myname",
				AutoplanEnabled: false,
				PolicySets:      emptyPolicySets,
				RebaseEnabled:   false,
			},
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			tmp, cleanup := TempDir(t)
			defer cleanup()
			global := valid.NewGlobalCfg("somedir")

			if c.gCfg != "" {
				path := filepath.Join(tmp, "config.yaml")
				Ok(t, os.WriteFile(path, []byte(c.gCfg), 0600))
				var err error
				global, err = (&config.ParserValidator{}).ParseGlobalCfg(path, global)
				Ok(t, err)
			}

			global.PolicySets = emptyPolicySets
			Equals(t, c.exp, global.MergeProjectCfg(c.repoID, c.proj, valid.RepoCfg{Workflows: c.repoWorkflows}))
		})
	}
}

func TestRepo_IDMatches(t *testing.T) {
	// Test exact matches.
	Equals(t, false, (valid.Repo{ID: "github.com/owner/repo"}).IDMatches("github.com/runatlantis/atlantis"))
	Equals(t, true, (valid.Repo{ID: "github.com/owner/repo"}).IDMatches("github.com/owner/repo"))

	// Test regexes.
	Equals(t, true, (valid.Repo{IDRegex: regexp.MustCompile(".*")}).IDMatches("github.com/owner/repo"))
	Equals(t, true, (valid.Repo{IDRegex: regexp.MustCompile("github.com")}).IDMatches("github.com/owner/repo"))
	Equals(t, false, (valid.Repo{IDRegex: regexp.MustCompile("github.com/anotherowner")}).IDMatches("github.com/owner/repo"))
	Equals(t, true, (valid.Repo{IDRegex: regexp.MustCompile("github.com/(owner|runatlantis)")}).IDMatches("github.com/owner/repo"))
	Equals(t, true, (valid.Repo{IDRegex: regexp.MustCompile("github.com/owner.*")}).IDMatches("github.com/owner/repo"))
}

func TestRepo_IDString(t *testing.T) {
	Equals(t, "github.com/owner/repo", (valid.Repo{ID: "github.com/owner/repo"}).IDString())
	Equals(t, "/regex.*/", (valid.Repo{IDRegex: regexp.MustCompile("regex.*")}).IDString())
}

func TestRepo_BranchMatches(t *testing.T) {
	// Test matches when no branch regex is set.
	Equals(t, true, (valid.Repo{}).BranchMatches("main"))

	// Test regexes.
	Equals(t, true, (valid.Repo{BranchRegex: regexp.MustCompile(".*")}).BranchMatches("main"))
	Equals(t, true, (valid.Repo{BranchRegex: regexp.MustCompile("main")}).BranchMatches("main"))
	Equals(t, false, (valid.Repo{BranchRegex: regexp.MustCompile("^main$")}).BranchMatches("foo-main"))
	Equals(t, false, (valid.Repo{BranchRegex: regexp.MustCompile("^main$")}).BranchMatches("main-foo"))
	Equals(t, true, (valid.Repo{BranchRegex: regexp.MustCompile("(main|master)")}).BranchMatches("main"))
	Equals(t, true, (valid.Repo{BranchRegex: regexp.MustCompile("(main|master)")}).BranchMatches("master"))
	Equals(t, true, (valid.Repo{BranchRegex: regexp.MustCompile("release")}).BranchMatches("release-stage"))
	Equals(t, false, (valid.Repo{BranchRegex: regexp.MustCompile("release")}).BranchMatches("main"))
}

func TestGlobalCfg_MatchingRepo(t *testing.T) {
	defaultRepo := valid.Repo{
		IDRegex:           regexp.MustCompile(".*"),
		BranchRegex:       regexp.MustCompile(".*"),
		ApplyRequirements: []string{},
	}
	repo1 := valid.Repo{
		IDRegex:           regexp.MustCompile(".*"),
		BranchRegex:       regexp.MustCompile("^main$"),
		ApplyRequirements: []string{"approved"},
	}
	repo2 := valid.Repo{
		ID:                "github.com/owner/repo",
		BranchRegex:       regexp.MustCompile("^master$"),
		ApplyRequirements: []string{"approved", "mergeable"},
	}

	cases := map[string]struct {
		gCfg   valid.GlobalCfg
		repoID string
		exp    *valid.Repo
	}{
		"matches to default": {
			gCfg: valid.GlobalCfg{
				Repos: []valid.Repo{
					defaultRepo,
					repo2,
				},
			},
			repoID: "foo",
			exp:    &defaultRepo,
		},
		"matches to IDRegex": {
			gCfg: valid.GlobalCfg{
				Repos: []valid.Repo{
					defaultRepo,
					repo1,
					repo2,
				},
			},
			repoID: "foo",
			exp:    &repo1,
		},
		"matches to ID": {
			gCfg: valid.GlobalCfg{
				Repos: []valid.Repo{
					defaultRepo,
					repo1,
					repo2,
				},
			},
			repoID: "github.com/owner/repo",
			exp:    &repo2,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			Equals(t, c.exp, c.gCfg.MatchingRepo(c.repoID))
		})
	}
}
