package events_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock/v4"
	terraform_mocks "github.com/runatlantis/atlantis/server/core/terraform/mocks"

	"github.com/runatlantis/atlantis/server/core/config"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics"
	. "github.com/runatlantis/atlantis/testing"
)

func TestDefaultProjectCommandBuilder_BuildAutoplanCommands(t *testing.T) {
	// expCtxFields define the ctx fields we're going to assert on.
	// Since we're focused on autoplanning here, we don't validate all the
	// fields so the tests are more obvious and targeted.
	type expCtxFields struct {
		ProjectName string
		RepoRelDir  string
		Workspace   string
	}
	cases := []struct {
		Description    string
		AtlantisYAML   string
		ServerSideYAML string
		exp            []expCtxFields
	}{
		{
			Description: "simple atlantis.yaml",
			AtlantisYAML: `
version: 3
projects:
- dir: .
`,
			exp: []expCtxFields{
				{
					ProjectName: "",
					RepoRelDir:  ".",
					Workspace:   "default",
				},
			},
		},
		{
			Description: "some projects disabled",
			AtlantisYAML: `
version: 3
projects:
- dir: .
  autoplan:
    enabled: false
- dir: .
  workspace: myworkspace
  autoplan:
    when_modified: ["main.tf"]
- dir: .
  name: myname
  workspace: myworkspace2
`,
			exp: []expCtxFields{
				{
					ProjectName: "",
					RepoRelDir:  ".",
					Workspace:   "myworkspace",
				},
				{
					ProjectName: "myname",
					RepoRelDir:  ".",
					Workspace:   "myworkspace2",
				},
			},
		},
		{
			Description: "some projects disabled",
			AtlantisYAML: `
version: 3
projects:
- dir: .
  autoplan:
    enabled: false
- dir: .
  workspace: myworkspace
  autoplan:
    when_modified: ["main.tf"]
- dir: .
  workspace: myworkspace2
`,
			exp: []expCtxFields{
				{
					ProjectName: "",
					RepoRelDir:  ".",
					Workspace:   "myworkspace",
				},
				{
					ProjectName: "",
					RepoRelDir:  ".",
					Workspace:   "myworkspace2",
				},
			},
		},
		{
			Description: "no projects modified",
			AtlantisYAML: `
version: 3
projects:
- dir: mydir
`,
			exp: nil,
		},
	}

	logger := logging.NewNoopLogger(t)
	scope, _, _ := metrics.NewLoggingScope(logger, "atlantis")

	terraformClient := terraform_mocks.NewMockClient()
	When(terraformClient.ListAvailableVersions(Any[logging.SimpleLogging]())).ThenReturn([]string{}, nil)

	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			RegisterMockTestingT(t)
			tmpDir := DirStructure(t, map[string]interface{}{
				"main.tf": nil,
			})

			workingDir := mocks.NewMockWorkingDir()
			When(workingDir.Clone(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(tmpDir, false, nil)
			vcsClient := vcsmocks.NewMockClient()
			When(vcsClient.GetModifiedFiles(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn([]string{"main.tf"}, nil)
			if c.AtlantisYAML != "" {
				err := os.WriteFile(filepath.Join(tmpDir, valid.DefaultAtlantisFile), []byte(c.AtlantisYAML), 0600)
				Ok(t, err)
			}

			globalCfgArgs := valid.GlobalCfgArgs{
				AllowRepoCfg:  false,
				MergeableReq:  false,
				ApprovedReq:   false,
				UnDivergedReq: false,
			}

			builder := events.NewProjectCommandBuilder(
				false,
				&config.ParserValidator{},
				&events.DefaultProjectFinder{},
				vcsClient,
				workingDir,
				events.NewDefaultWorkingDirLocker(),
				valid.NewGlobalCfgFromArgs(globalCfgArgs),
				&events.DefaultPendingPlanFinder{},
				&events.CommentParser{ExecutableName: "atlantis"},
				false,
				false,
				false,
				false,
				false,
				"",
				"**/*.tf,**/*.tfvars,**/*.tfvars.json,**/terragrunt.hcl,**/.terraform.lock.hcl",
				false,
				false,
				scope,
				logger,
				terraformClient,
			)

			ctxs, err := builder.BuildAutoplanCommands(&command.Context{
				PullRequestStatus: models.PullReqStatus{
					Mergeable: true,
				},
				Log:   logger,
				Scope: scope,
			})
			Ok(t, err)
			Equals(t, len(c.exp), len(ctxs))
			for i, actCtx := range ctxs {
				expCtx := c.exp[i]
				Equals(t, expCtx.ProjectName, actCtx.ProjectName)
				Equals(t, expCtx.RepoRelDir, actCtx.RepoRelDir)
				Equals(t, expCtx.Workspace, actCtx.Workspace)
			}
		})
	}
}

// Test building a plan and apply command for one project.
func TestDefaultProjectCommandBuilder_BuildSinglePlanApplyCommand(t *testing.T) {
	cases := []struct {
		Description                string
		AtlantisYAML               string
		Cmd                        events.CommentCommand
		Silenced                   bool
		ExpCommentArgs             []string
		ExpWorkspace               string
		ExpDir                     string
		ExpProjectName             string
		ExpErr                     string
		ExpApplyReqs               []string
		EnableAutoMergeUserCfg     bool
		EnableParallelPlanUserCfg  bool
		EnableParallelApplyUserCfg bool
		ExpAutoMerge               bool
		ExpParallelPlan            bool
		ExpParallelApply           bool
		ExpNoProjects              bool
	}{
		{
			Description: "no atlantis.yaml",
			Cmd: events.CommentCommand{
				RepoRelDir: ".",
				Flags:      []string{"commentarg"},
				Name:       command.Plan,
				Workspace:  "myworkspace",
			},
			AtlantisYAML:   "",
			ExpCommentArgs: []string{`\c\o\m\m\e\n\t\a\r\g`},
			ExpWorkspace:   "myworkspace",
			ExpDir:         ".",
			ExpApplyReqs:   []string{},
		},
		{
			Description: "no atlantis.yaml with project flag",
			Cmd: events.CommentCommand{
				RepoRelDir:  ".",
				Name:        command.Plan,
				ProjectName: "myproject",
			},
			AtlantisYAML: "",
			ExpErr:       "cannot specify a project name unless an atlantis.yaml file exists to configure projects",
		},
		{
			Description: "simple atlantis.yaml",
			Cmd: events.CommentCommand{
				RepoRelDir: ".",
				Name:       command.Plan,
				Workspace:  "myworkspace",
			},
			AtlantisYAML: `
version: 3
projects:
- dir: .
  workspace: myworkspace
  apply_requirements: [approved]`,
			ExpApplyReqs: []string{"approved"},
			ExpWorkspace: "myworkspace",
			ExpDir:       ".",
		},
		{
			Description: "atlantis.yaml wrong dir",
			Cmd: events.CommentCommand{
				RepoRelDir: ".",
				Name:       command.Plan,
				Workspace:  "myworkspace",
			},
			AtlantisYAML: `
version: 3
projects:
- dir: notroot
  workspace: myworkspace
  apply_requirements: [approved]`,
			ExpWorkspace: "myworkspace",
			ExpDir:       ".",
			ExpApplyReqs: []string{},
		},
		{
			Description: "atlantis.yaml wrong workspace",
			Cmd: events.CommentCommand{
				RepoRelDir: ".",
				Name:       command.Plan,
				Workspace:  "myworkspace",
			},
			AtlantisYAML: `
version: 3
projects:
- dir: .
  workspace: notmyworkspace
  apply_requirements: [approved]`,
			ExpErr: "running commands in workspace \"myworkspace\" is not allowed because this directory is only configured for the following workspaces: notmyworkspace",
		},
		{
			Description: "atlantis.yaml with projectname",
			Cmd: events.CommentCommand{
				Name:        command.Plan,
				ProjectName: "myproject",
			},
			AtlantisYAML: `
version: 3
projects:
- name: myproject
  dir: .
  workspace: myworkspace
  apply_requirements: [approved]`,
			ExpApplyReqs:   []string{"approved"},
			ExpProjectName: "myproject",
			ExpWorkspace:   "myworkspace",
			ExpDir:         ".",
		},
		{
			Description: "atlantis.yaml with mergeable apply requirement",
			Cmd: events.CommentCommand{
				Name:        command.Plan,
				ProjectName: "myproject",
			},
			AtlantisYAML: `
version: 3
projects:
- name: myproject
  dir: .
  workspace: myworkspace
  apply_requirements: [mergeable]`,
			ExpApplyReqs:   []string{"mergeable"},
			ExpProjectName: "myproject",
			ExpWorkspace:   "myworkspace",
			ExpDir:         ".",
		},
		{
			Description: "atlantis.yaml with mergeable and approved apply requirements",
			Cmd: events.CommentCommand{
				Name:        command.Plan,
				ProjectName: "myproject",
			},
			AtlantisYAML: `
version: 3
projects:
- name: myproject
  dir: .
  workspace: myworkspace
  apply_requirements: [mergeable, approved]`,
			ExpApplyReqs:   []string{"mergeable", "approved"},
			ExpProjectName: "myproject",
			ExpWorkspace:   "myworkspace",
			ExpDir:         ".",
		},
		{
			Description: "atlantis.yaml with multiple dir/workspaces matching",
			Cmd: events.CommentCommand{
				Name:       command.Plan,
				RepoRelDir: ".",
				Workspace:  "myworkspace",
			},
			AtlantisYAML: `
version: 3
projects:
- name: myproject
  dir: .
  workspace: myworkspace
  apply_requirements: [approved]
- name: myproject2
  dir: .
  workspace: myworkspace
`,
			ExpErr: "must specify project name: more than one project defined in atlantis.yaml matched dir: \".\" workspace: \"myworkspace\"",
		},
		{
			Description: "atlantis.yaml with project flag not matching",
			Cmd: events.CommentCommand{
				Name:        command.Plan,
				RepoRelDir:  ".",
				Workspace:   "default",
				ProjectName: "notconfigured",
			},
			AtlantisYAML: `
version: 3
projects:
- dir: .
`,
			ExpErr: "no project with name \"notconfigured\" is defined in atlantis.yaml",
		},
		{
			Description: "atlantis.yaml with project flag not matching but silenced",
			Cmd: events.CommentCommand{
				Name:        command.Plan,
				RepoRelDir:  ".",
				Workspace:   "default",
				ProjectName: "notconfigured",
			},
			AtlantisYAML: `
version: 3
projects:
- dir: .
`,
			Silenced:      true,
			ExpNoProjects: true,
		},
		{
			Description: "atlantis.yaml with ParallelPlan Set to true",
			Cmd: events.CommentCommand{
				Name:        command.Plan,
				RepoRelDir:  ".",
				Workspace:   "default",
				ProjectName: "myproject",
			},
			AtlantisYAML: `
version: 3
parallel_plan: true
projects:
- name: myproject
  dir: .
  workspace: myworkspace
`,
			ExpParallelPlan:  true,
			ExpParallelApply: false,
			ExpDir:           ".",
			ExpWorkspace:     "myworkspace",
			ExpProjectName:   "myproject",
			ExpApplyReqs:     []string{},
		},
		{
			Description: "atlantis.yaml with ParallelPlan/apply and Automerge not set, but set in user conf",
			Cmd: events.CommentCommand{
				Name:        command.Plan,
				RepoRelDir:  ".",
				Workspace:   "default",
				ProjectName: "myproject",
			},
			AtlantisYAML: `
version: 3
projects:
- name: myproject
  dir: .
  workspace: myworkspace
`,
			EnableAutoMergeUserCfg:     true,
			EnableParallelPlanUserCfg:  true,
			EnableParallelApplyUserCfg: true,
			ExpAutoMerge:               true,
			ExpParallelPlan:            true,
			ExpParallelApply:           true,
			ExpDir:                     ".",
			ExpWorkspace:               "myworkspace",
			ExpProjectName:             "myproject",
			ExpApplyReqs:               []string{},
		},
		{
			Description: "atlantis.yaml with ParallelPlan/apply and Automerge set to false, but set to true in user conf",
			Cmd: events.CommentCommand{
				Name:        command.Plan,
				RepoRelDir:  ".",
				Workspace:   "default",
				ProjectName: "myproject",
			},
			AtlantisYAML: `
version: 3
automerge: false
parallel_plan: false
parallel_apply: false
projects:
- name: myproject
  dir: .
  workspace: myworkspace
`,
			EnableAutoMergeUserCfg:     true,
			EnableParallelPlanUserCfg:  true,
			EnableParallelApplyUserCfg: true,
			ExpAutoMerge:               false,
			ExpParallelPlan:            false,
			ExpParallelApply:           false,
			ExpDir:                     ".",
			ExpWorkspace:               "myworkspace",
			ExpProjectName:             "myproject",
			ExpApplyReqs:               []string{},
		},
	}

	logger := logging.NewNoopLogger(t)
	scope, _, _ := metrics.NewLoggingScope(logger, "atlantis")

	for _, c := range cases {
		// NOTE: we're testing both plan and apply here.
		for _, cmdName := range []command.Name{command.Plan, command.Apply} {
			t.Run(c.Description+"_"+cmdName.String(), func(t *testing.T) {
				RegisterMockTestingT(t)
				tmpDir := DirStructure(t, map[string]interface{}{
					"main.tf": nil,
				})

				workingDir := mocks.NewMockWorkingDir()
				When(workingDir.Clone(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(tmpDir, false, nil)
				When(workingDir.GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(tmpDir, nil)
				vcsClient := vcsmocks.NewMockClient()
				When(vcsClient.GetModifiedFiles(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn([]string{"main.tf"}, nil)
				if c.AtlantisYAML != "" {
					err := os.WriteFile(filepath.Join(tmpDir, valid.DefaultAtlantisFile), []byte(c.AtlantisYAML), 0600)
					Ok(t, err)
				}

				globalCfgArgs := valid.GlobalCfgArgs{
					AllowRepoCfg:  true,
					MergeableReq:  false,
					ApprovedReq:   false,
					UnDivergedReq: false,
				}

				terraformClient := terraform_mocks.NewMockClient()
				When(terraformClient.ListAvailableVersions(Any[logging.SimpleLogging]())).ThenReturn([]string{}, nil)

				builder := events.NewProjectCommandBuilder(
					false,
					&config.ParserValidator{},
					&events.DefaultProjectFinder{},
					vcsClient,
					workingDir,
					events.NewDefaultWorkingDirLocker(),
					valid.NewGlobalCfgFromArgs(globalCfgArgs),
					&events.DefaultPendingPlanFinder{},
					&events.CommentParser{ExecutableName: "atlantis"},
					false,
					true,
					c.EnableAutoMergeUserCfg,
					c.EnableParallelPlanUserCfg,
					c.EnableParallelApplyUserCfg,
					"",
					"**/*.tf,**/*.tfvars,**/*.tfvars.json,**/terragrunt.hcl,**/.terraform.lock.hcl",
					false,
					c.Silenced,
					scope,
					logger,
					terraformClient,
				)

				var actCtxs []command.ProjectContext
				var err error
				if cmdName == command.Plan {
					actCtxs, err = builder.BuildPlanCommands(&command.Context{
						Log:   logger,
						Scope: scope,
					}, &c.Cmd)
				} else {
					actCtxs, err = builder.BuildApplyCommands(&command.Context{Log: logger, Scope: scope}, &c.Cmd)
				}

				if c.ExpErr != "" {
					ErrEquals(t, c.ExpErr, err)
					return
				}
				Ok(t, err)
				if c.ExpNoProjects {
					Equals(t, 0, len(actCtxs))
					return
				}
				Equals(t, 1, len(actCtxs))
				actCtx := actCtxs[0]
				Equals(t, c.ExpDir, actCtx.RepoRelDir)
				Equals(t, c.ExpWorkspace, actCtx.Workspace)
				Equals(t, c.ExpCommentArgs, actCtx.EscapedCommentArgs)
				Equals(t, c.ExpProjectName, actCtx.ProjectName)
				Equals(t, c.ExpApplyReqs, actCtx.ApplyRequirements)
				Equals(t, c.ExpAutoMerge, actCtx.AutomergeEnabled)
				Equals(t, c.ExpParallelPlan, actCtx.ParallelPlanEnabled)
				Equals(t, c.ExpParallelApply, actCtx.ParallelApplyEnabled)
			})
		}
	}
}

// Test building a plan and apply command for one project
// with the RestrictFileList
func TestDefaultProjectCommandBuilder_BuildSinglePlanApplyCommand_WithRestrictFileList(t *testing.T) {
	cases := []struct {
		Description        string
		AtlantisYAML       string
		DirectoryStructure map[string]interface{}
		ModifiedFiles      []string
		Cmd                events.CommentCommand
		ExpErr             string
	}{
		{
			Description: "planning a file outside of the changed files",
			Cmd: events.CommentCommand{
				Name:       command.Plan,
				RepoRelDir: "directory-1",
				Workspace:  "default",
			},
			DirectoryStructure: map[string]interface{}{
				"directory-1": map[string]interface{}{
					"main.tf": nil,
				},
				"directory-2": map[string]interface{}{
					"main.tf": nil,
				},
			},
			ModifiedFiles: []string{"directory-2/main.tf"},
			ExpErr:        "the dir \"directory-1\" is not in the plan list of this pull request",
		},
		{
			Description: "planning a file of the changed files",
			Cmd: events.CommentCommand{
				Name:       command.Plan,
				RepoRelDir: "directory-1",
				Workspace:  "default",
			},
			DirectoryStructure: map[string]interface{}{
				"directory-1": map[string]interface{}{
					"main.tf": nil,
				},
				"directory-2": map[string]interface{}{
					"main.tf": nil,
				},
			},
			ModifiedFiles: []string{"directory-1/main.tf"},
		},
		{
			Description: "planning a project outside of the requested changed files",
			Cmd: events.CommentCommand{
				Name:        command.Plan,
				Workspace:   "default",
				ProjectName: "project-1",
			},
			AtlantisYAML: `
version: 3
projects:
- name: project-1
  dir: directory-1
- name: project-2
  dir: directory-2
`,
			DirectoryStructure: map[string]interface{}{
				"directory-1": map[string]interface{}{
					"main.tf": nil,
				},
				"directory-2": map[string]interface{}{
					"main.tf": nil,
				},
			},
			ModifiedFiles: []string{"directory-2/main.tf"},
			ExpErr:        "the following directories are present in the pull request but not in the requested project:\ndirectory-2",
		},
		{
			Description: "planning a project defined in the requested changed files",
			Cmd: events.CommentCommand{
				Name:        command.Plan,
				Workspace:   "default",
				ProjectName: "project-1",
			},
			AtlantisYAML: `
version: 3
projects:
- name: project-1
  dir: directory-1
- name: project-2
  dir: directory-2
`,
			DirectoryStructure: map[string]interface{}{
				"directory-1": map[string]interface{}{
					"main.tf": nil,
				},
				"directory-2": map[string]interface{}{
					"main.tf": nil,
				},
			},
			ModifiedFiles: []string{"directory-1/main.tf"},
		},
	}

	logger := logging.NewNoopLogger(t)
	scope, _, _ := metrics.NewLoggingScope(logger, "atlantis")

	for _, c := range cases {
		t.Run(c.Description+"_"+command.Plan.String(), func(t *testing.T) {
			RegisterMockTestingT(t)
			tmpDir := DirStructure(t, c.DirectoryStructure)

			workingDir := mocks.NewMockWorkingDir()
			When(workingDir.Clone(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(tmpDir, false, nil)
			When(workingDir.GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(tmpDir, nil)
			vcsClient := vcsmocks.NewMockClient()
			When(vcsClient.GetModifiedFiles(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(c.ModifiedFiles, nil)
			if c.AtlantisYAML != "" {
				err := os.WriteFile(filepath.Join(tmpDir, valid.DefaultAtlantisFile), []byte(c.AtlantisYAML), 0600)
				Ok(t, err)
			}

			globalCfgArgs := valid.GlobalCfgArgs{
				AllowRepoCfg:  true,
				MergeableReq:  false,
				ApprovedReq:   false,
				UnDivergedReq: false,
			}

			terraformClient := terraform_mocks.NewMockClient()
			When(terraformClient.ListAvailableVersions(Any[logging.SimpleLogging]())).ThenReturn([]string{}, nil)

			builder := events.NewProjectCommandBuilder(
				false,
				&config.ParserValidator{},
				&events.DefaultProjectFinder{},
				vcsClient,
				workingDir,
				events.NewDefaultWorkingDirLocker(),
				valid.NewGlobalCfgFromArgs(globalCfgArgs),
				&events.DefaultPendingPlanFinder{},
				&events.CommentParser{ExecutableName: "atlantis"},
				false,
				true,
				false,
				false,
				false,
				"",
				"**/*.tf,**/*.tfvars,**/*.tfvars.json,**/terragrunt.hcl,**/.terraform.lock.hcl",
				true,
				false,
				scope,
				logger,
				terraformClient,
			)

			var actCtxs []command.ProjectContext
			var err error
			actCtxs, err = builder.BuildPlanCommands(&command.Context{
				Log:   logger,
				Scope: scope,
			}, &c.Cmd)

			if c.ExpErr != "" {
				ErrEquals(t, c.ExpErr, err)
				return
			}
			Ok(t, err)
			Equals(t, 1, len(actCtxs))
		})
	}
}

func TestDefaultProjectCommandBuilder_BuildPlanCommands(t *testing.T) {
	// expCtxFields define the ctx fields we're going to assert on.
	// Since we're focused on autoplanning here, we don't validate all the
	// fields so the tests are more obvious and targeted.
	type expCtxFields struct {
		ProjectName      string
		RepoRelDir       string
		Workspace        string
		Automerge        bool
		ExpParallelPlan  bool
		ExpParallelApply bool
	}
	cases := map[string]struct {
		AutoMergeUserCfg            bool
		ParallelPlanEnabledUserCfg  bool
		ParallelApplyEnabledUserCfg bool
		DirStructure                map[string]interface{}
		AtlantisYAML                string
		ModifiedFiles               []string
		Exp                         []expCtxFields
	}{
		"no atlantis.yaml": {
			DirStructure: map[string]interface{}{
				"project1": map[string]interface{}{
					"main.tf": nil,
				},
				"project2": map[string]interface{}{
					"main.tf": nil,
				},
			},
			ModifiedFiles: []string{"project1/main.tf", "project2/main.tf"},
			Exp: []expCtxFields{
				{
					ProjectName: "",
					RepoRelDir:  "project1",
					Workspace:   "default",
				},
				{
					ProjectName: "",
					RepoRelDir:  "project2",
					Workspace:   "default",
				},
			},
		},
		"no projects in atlantis.yaml with parallel operations in atlantis.yaml": {
			DirStructure: map[string]interface{}{
				"project1": map[string]interface{}{
					"main.tf": nil,
				},
				"project2": map[string]interface{}{
					"main.tf": nil,
				},
			},
			AtlantisYAML: `
version: 3
automerge: true
parallel_plan: true
parallel_apply: true
`,
			ModifiedFiles: []string{"project1/main.tf", "project2/main.tf"},
			Exp: []expCtxFields{
				{
					ProjectName:      "",
					RepoRelDir:       "project1",
					Workspace:        "default",
					Automerge:        true,
					ExpParallelApply: true,
					ExpParallelPlan:  true,
				},
				{
					ProjectName:      "",
					RepoRelDir:       "project2",
					Workspace:        "default",
					Automerge:        true,
					ExpParallelApply: true,
					ExpParallelPlan:  true,
				},
			},
		},
		"no projects in atlantis.yaml with parallel operations and automerge not in atlantis.yaml, but in user conf": {
			DirStructure: map[string]interface{}{
				"project1": map[string]interface{}{
					"main.tf": nil,
				},
				"project2": map[string]interface{}{
					"main.tf": nil,
				},
			},
			AtlantisYAML: `
version: 3
`,
			AutoMergeUserCfg:            true,
			ParallelPlanEnabledUserCfg:  true,
			ParallelApplyEnabledUserCfg: true,
			ModifiedFiles:               []string{"project1/main.tf", "project2/main.tf"},
			Exp: []expCtxFields{
				{
					ProjectName:      "",
					RepoRelDir:       "project1",
					Workspace:        "default",
					Automerge:        true,
					ExpParallelApply: true,
					ExpParallelPlan:  true,
				},
				{
					ProjectName:      "",
					RepoRelDir:       "project2",
					Workspace:        "default",
					Automerge:        true,
					ExpParallelApply: true,
					ExpParallelPlan:  true,
				},
			},
		},
		"no projects in atlantis.yaml with parallel operations and automerge set to false in atlantis.yaml and true in user conf": {
			DirStructure: map[string]interface{}{
				"project1": map[string]interface{}{
					"main.tf": nil,
				},
				"project2": map[string]interface{}{
					"main.tf": nil,
				},
			},
			AtlantisYAML: `
version: 3
automerge: false
parallel_plan: false
parallel_apply: false
`,
			AutoMergeUserCfg:            true,
			ParallelPlanEnabledUserCfg:  true,
			ParallelApplyEnabledUserCfg: true,
			ModifiedFiles:               []string{"project1/main.tf", "project2/main.tf"},
			Exp: []expCtxFields{
				{
					ProjectName:      "",
					RepoRelDir:       "project1",
					Workspace:        "default",
					Automerge:        false,
					ExpParallelApply: false,
					ExpParallelPlan:  false,
				},
				{
					ProjectName:      "",
					RepoRelDir:       "project2",
					Workspace:        "default",
					Automerge:        false,
					ExpParallelApply: false,
					ExpParallelPlan:  false,
				},
			},
		},
		"no modified files": {
			DirStructure: map[string]interface{}{
				"main.tf": nil,
			},
			ModifiedFiles: []string{},
			Exp:           []expCtxFields{},
		},
		"follow when_modified config": {
			DirStructure: map[string]interface{}{
				"project1": map[string]interface{}{
					"main.tf": nil,
				},
				"project2": map[string]interface{}{
					"main.tf": nil,
				},
				"project3": map[string]interface{}{
					"main.tf": nil,
				},
			},
			AtlantisYAML: `version: 3
projects:
- dir: project1 # project1 uses the defaults
- dir: project2 # project2 has autoplan disabled but should use default when_modified
  autoplan:
    enabled: false
- dir: project3 # project3 has an empty when_modified
  autoplan:
    enabled: false
    when_modified: []`,
			ModifiedFiles: []string{"project1/main.tf", "project2/main.tf", "project3/main.tf"},
			Exp: []expCtxFields{
				{
					ProjectName: "",
					RepoRelDir:  "project1",
					Workspace:   "default",
				},
				{
					ProjectName: "",
					RepoRelDir:  "project2",
					Workspace:   "default",
				},
			},
		},
	}

	logger := logging.NewNoopLogger(t)
	scope, _, _ := metrics.NewLoggingScope(logger, "atlantis")
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			RegisterMockTestingT(t)
			tmpDir := DirStructure(t, c.DirStructure)

			workingDir := mocks.NewMockWorkingDir()
			When(workingDir.Clone(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(tmpDir, false, nil)
			When(workingDir.GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(tmpDir, nil)
			vcsClient := vcsmocks.NewMockClient()
			When(vcsClient.GetModifiedFiles(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(c.ModifiedFiles, nil)
			if c.AtlantisYAML != "" {
				err := os.WriteFile(filepath.Join(tmpDir, valid.DefaultAtlantisFile), []byte(c.AtlantisYAML), 0600)
				Ok(t, err)
			}

			globalCfgArgs := valid.GlobalCfgArgs{
				AllowRepoCfg:  true,
				MergeableReq:  false,
				ApprovedReq:   false,
				UnDivergedReq: false,
			}

			terraformClient := terraform_mocks.NewMockClient()
			When(terraformClient.ListAvailableVersions(Any[logging.SimpleLogging]())).ThenReturn([]string{}, nil)

			builder := events.NewProjectCommandBuilder(
				false,
				&config.ParserValidator{},
				&events.DefaultProjectFinder{},
				vcsClient,
				workingDir,
				events.NewDefaultWorkingDirLocker(),
				valid.NewGlobalCfgFromArgs(globalCfgArgs),
				&events.DefaultPendingPlanFinder{},
				&events.CommentParser{ExecutableName: "atlantis"},
				false,
				false,
				false,
				c.ParallelPlanEnabledUserCfg,
				c.ParallelApplyEnabledUserCfg,
				"",
				"**/*.tf,**/*.tfvars,**/*.tfvars.json,**/terragrunt.hcl,**/.terraform.lock.hcl",
				false,
				false,
				scope,
				logger,
				terraformClient,
			)

			ctxs, err := builder.BuildPlanCommands(
				&command.Context{
					Log:   logger,
					Scope: scope,
				},
				&events.CommentCommand{
					RepoRelDir:  "",
					Flags:       nil,
					Name:        command.Plan,
					Verbose:     false,
					Workspace:   "",
					ProjectName: "",
				})
			Ok(t, err)
			Equals(t, len(c.Exp), len(ctxs))
			for i, actCtx := range ctxs {
				expCtx := c.Exp[i]
				Equals(t, expCtx.ProjectName, actCtx.ProjectName)
				Equals(t, expCtx.RepoRelDir, actCtx.RepoRelDir)
				Equals(t, expCtx.Workspace, actCtx.Workspace)
				Equals(t, expCtx.ExpParallelPlan, actCtx.ParallelPlanEnabled)
				Equals(t, expCtx.ExpParallelApply, actCtx.ParallelApplyEnabled)
			}
		})
	}
}

// Test building apply command for multiple projects when the comment
// isn't for a specific project, i.e. atlantis apply.
// In this case we should apply all outstanding plans.
func TestDefaultProjectCommandBuilder_BuildMultiApply(t *testing.T) {
	RegisterMockTestingT(t)
	tmpDir := DirStructure(t, map[string]interface{}{
		"workspace1": map[string]interface{}{
			"project1": map[string]interface{}{
				"main.tf":          nil,
				"workspace.tfplan": nil,
			},
			"project2": map[string]interface{}{
				"main.tf":          nil,
				"workspace.tfplan": nil,
			},
		},
		"workspace2": map[string]interface{}{
			"project1": map[string]interface{}{
				"main.tf":          nil,
				"workspace.tfplan": nil,
			},
			"project2": map[string]interface{}{
				"main.tf":          nil,
				"workspace.tfplan": nil,
			},
		},
	})
	// Initialize git repos in each workspace so that the .tfplan files get
	// picked up.
	runCmd(t, filepath.Join(tmpDir, "workspace1"), "git", "init")
	runCmd(t, filepath.Join(tmpDir, "workspace2"), "git", "init")

	workingDir := mocks.NewMockWorkingDir()
	When(workingDir.GetPullDir(
		Any[models.Repo](),
		Any[models.PullRequest]())).
		ThenReturn(tmpDir, nil)

	logger := logging.NewNoopLogger(t)

	globalCfgArgs := valid.GlobalCfgArgs{
		AllowRepoCfg:  false,
		MergeableReq:  false,
		ApprovedReq:   false,
		UnDivergedReq: false,
	}
	scope, _, _ := metrics.NewLoggingScope(logger, "atlantis")

	terraformClient := terraform_mocks.NewMockClient()
	When(terraformClient.ListAvailableVersions(Any[logging.SimpleLogging]())).ThenReturn([]string{}, nil)

	builder := events.NewProjectCommandBuilder(
		false,
		&config.ParserValidator{},
		&events.DefaultProjectFinder{},
		nil,
		workingDir,
		events.NewDefaultWorkingDirLocker(),
		valid.NewGlobalCfgFromArgs(globalCfgArgs),
		&events.DefaultPendingPlanFinder{},
		&events.CommentParser{ExecutableName: "atlantis"},
		false,
		false,
		false,
		false,
		false,
		"",
		"**/*.tf,**/*.tfvars,**/*.tfvars.json,**/terragrunt.hcl,**/.terraform.lock.hcl",
		false,
		false,
		scope,
		logger,
		terraformClient,
	)

	ctxs, err := builder.BuildApplyCommands(
		&command.Context{
			Log:   logger,
			Scope: scope,
		},
		&events.CommentCommand{
			RepoRelDir:  "",
			Flags:       nil,
			Name:        command.Apply,
			Verbose:     false,
			Workspace:   "",
			ProjectName: "",
		})
	Ok(t, err)
	Equals(t, 4, len(ctxs))
	Equals(t, "project1", ctxs[0].RepoRelDir)
	Equals(t, "workspace1", ctxs[0].Workspace)
	Equals(t, "project2", ctxs[1].RepoRelDir)
	Equals(t, "workspace1", ctxs[1].Workspace)
	Equals(t, "project1", ctxs[2].RepoRelDir)
	Equals(t, "workspace2", ctxs[2].Workspace)
	Equals(t, "project2", ctxs[3].RepoRelDir)
	Equals(t, "workspace2", ctxs[3].Workspace)
}

// Test that if a directory has a list of workspaces configured then we don't
// allow plans for other workspace names.
func TestDefaultProjectCommandBuilder_WrongWorkspaceName(t *testing.T) {
	RegisterMockTestingT(t)
	workingDir := mocks.NewMockWorkingDir()

	tmpDir := DirStructure(t, map[string]interface{}{
		"pulldir": map[string]interface{}{
			"notconfigured": map[string]interface{}{},
		},
	})
	repoDir := filepath.Join(tmpDir, "pulldir/notconfigured")

	yamlCfg := `version: 3
projects:
- dir: .
  workspace: default
- dir: .
  workspace: staging
`
	err := os.WriteFile(filepath.Join(repoDir, valid.DefaultAtlantisFile), []byte(yamlCfg), 0600)
	Ok(t, err)

	When(workingDir.Clone(
		Any[models.Repo](),
		Any[models.PullRequest](),
		Any[string]())).ThenReturn(repoDir, false, nil)
	When(workingDir.GetWorkingDir(
		Any[models.Repo](),
		Any[models.PullRequest](),
		Any[string]())).ThenReturn(repoDir, nil)

	globalCfgArgs := valid.GlobalCfgArgs{
		AllowRepoCfg:  true,
		MergeableReq:  false,
		ApprovedReq:   false,
		UnDivergedReq: false,
	}
	logger := logging.NewNoopLogger(t)
	scope, _, _ := metrics.NewLoggingScope(logger, "atlantis")
	terraformClient := terraform_mocks.NewMockClient()
	When(terraformClient.ListAvailableVersions(Any[logging.SimpleLogging]())).ThenReturn([]string{}, nil)

	builder := events.NewProjectCommandBuilder(
		false,
		&config.ParserValidator{},
		&events.DefaultProjectFinder{},
		nil,
		workingDir,
		events.NewDefaultWorkingDirLocker(),
		valid.NewGlobalCfgFromArgs(globalCfgArgs),
		&events.DefaultPendingPlanFinder{},
		&events.CommentParser{ExecutableName: "atlantis"},
		false,
		false,
		false,
		false,
		false,
		"",
		"**/*.tf,**/*.tfvars,**/*.tfvars.json,**/terragrunt.hcl,**/.terraform.lock.hcl",
		false,
		false,
		scope,
		logger,
		terraformClient,
	)

	ctx := &command.Context{
		HeadRepo: models.Repo{},
		Pull:     models.PullRequest{},
		User:     models.User{},
		Log:      logger,
		Scope:    scope,
	}
	_, err = builder.BuildPlanCommands(ctx, &events.CommentCommand{
		RepoRelDir:  ".",
		Flags:       nil,
		Name:        command.Plan,
		Verbose:     false,
		Workspace:   "notconfigured",
		ProjectName: "",
	})
	ErrEquals(t, "running commands in workspace \"notconfigured\" is not allowed because this directory is only configured for the following workspaces: default, staging", err)
}

// Test that extra comment args are escaped.
func TestDefaultProjectCommandBuilder_EscapeArgs(t *testing.T) {
	cases := []struct {
		ExtraArgs      []string
		ExpEscapedArgs []string
	}{
		{
			ExtraArgs:      []string{"arg1", "arg2"},
			ExpEscapedArgs: []string{`\a\r\g\1`, `\a\r\g\2`},
		},
		{
			ExtraArgs:      []string{"-var=$(touch bad)"},
			ExpEscapedArgs: []string{`\-\v\a\r\=\$\(\t\o\u\c\h\ \b\a\d\)`},
		},
		{
			ExtraArgs:      []string{"-- ;echo bad"},
			ExpEscapedArgs: []string{`\-\-\ \;\e\c\h\o\ \b\a\d`},
		},
	}

	logger := logging.NewNoopLogger(t)
	scope, _, _ := metrics.NewLoggingScope(logger, "atlantis")

	for _, c := range cases {
		t.Run(strings.Join(c.ExtraArgs, " "), func(t *testing.T) {
			RegisterMockTestingT(t)
			tmpDir := DirStructure(t, map[string]interface{}{
				"main.tf": nil,
			})

			workingDir := mocks.NewMockWorkingDir()
			When(workingDir.Clone(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(tmpDir, false, nil)
			When(workingDir.GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(tmpDir, nil)
			vcsClient := vcsmocks.NewMockClient()
			When(vcsClient.GetModifiedFiles(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn([]string{"main.tf"}, nil)

			globalCfgArgs := valid.GlobalCfgArgs{
				AllowRepoCfg:  true,
				MergeableReq:  false,
				ApprovedReq:   false,
				UnDivergedReq: false,
			}

			terraformClient := terraform_mocks.NewMockClient()
			When(terraformClient.ListAvailableVersions(Any[logging.SimpleLogging]())).ThenReturn([]string{}, nil)

			builder := events.NewProjectCommandBuilder(
				false,
				&config.ParserValidator{},
				&events.DefaultProjectFinder{},
				vcsClient,
				workingDir,
				events.NewDefaultWorkingDirLocker(),
				valid.NewGlobalCfgFromArgs(globalCfgArgs),
				&events.DefaultPendingPlanFinder{},
				&events.CommentParser{ExecutableName: "atlantis"},
				false,
				false,
				false,
				false,
				false,
				"",
				"**/*.tf,**/*.tfvars,**/*.tfvars.json,**/terragrunt.hcl,**/.terraform.lock.hcl",
				false,
				false,
				scope,
				logger,
				terraformClient,
			)

			var actCtxs []command.ProjectContext
			var err error
			actCtxs, err = builder.BuildPlanCommands(&command.Context{
				Log:   logger,
				Scope: scope,
			}, &events.CommentCommand{
				RepoRelDir: ".",
				Flags:      c.ExtraArgs,
				Name:       command.Plan,
				Verbose:    false,
				Workspace:  "default",
			})
			Ok(t, err)
			Equals(t, 1, len(actCtxs))
			actCtx := actCtxs[0]
			Equals(t, c.ExpEscapedArgs, actCtx.EscapedCommentArgs)
		})
	}
}

// Test that terraform version is used when specified in terraform configuration
func TestDefaultProjectCommandBuilder_TerraformVersion(t *testing.T) {
	// For the following tests:
	// If terraform configuration is used, result should be `0.12.8`.
	// If project configuration is used, result should be `0.12.6`.
	// If default is to be used, result should be `nil`.

	baseVersionConfig := `
terraform {
  required_version = "0.12.8"
}
`

	atlantisYamlContent := `
version: 3
projects:
- dir: project1 # project1 uses the defaults
  terraform_version: v0.12.6
`

	type testCase struct {
		DirStructure  map[string]interface{}
		AtlantisYAML  string
		ModifiedFiles []string
		Exp           map[string]string
	}

	testCases := make(map[string]testCase)

	// atlantis.yaml should take precedence over terraform config
	testCases["with project config and terraform config"] = testCase{
		DirStructure: map[string]interface{}{
			"project1": map[string]interface{}{
				"main.tf": baseVersionConfig,
			},
			valid.DefaultAtlantisFile: atlantisYamlContent,
		},
		ModifiedFiles: []string{"project1/main.tf", "project2/main.tf"},
		Exp: map[string]string{
			"project1": "0.12.6",
		},
	}

	testCases["with project config only"] = testCase{
		DirStructure: map[string]interface{}{
			"project1": map[string]interface{}{
				"main.tf": nil,
			},
			valid.DefaultAtlantisFile: atlantisYamlContent,
		},
		ModifiedFiles: []string{"project1/main.tf"},
		Exp: map[string]string{
			"project1": "0.12.6",
		},
	}

	testCases["neither project config or terraform config"] = testCase{
		DirStructure: map[string]interface{}{
			"project1": map[string]interface{}{
				"main.tf": nil,
			},
		},
		ModifiedFiles: []string{"project1/main.tf", "project2/main.tf"},
		Exp: map[string]string{
			"project1": "",
		},
	}

	testCases["project with different terraform config"] = testCase{
		DirStructure: map[string]interface{}{
			"project1": map[string]interface{}{
				"main.tf": baseVersionConfig,
			},
			"project2": map[string]interface{}{
				"main.tf": strings.Replace(baseVersionConfig, "0.12.8", "0.12.9", -1),
			},
		},
		ModifiedFiles: []string{"project1/main.tf", "project2/main.tf"},
		Exp: map[string]string{
			"project1": "0.12.8",
			"project2": "0.12.9",
		},
	}

	logger := logging.NewNoopLogger(t)
	scope, _, _ := metrics.NewLoggingScope(logger, "atlantis")

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			RegisterMockTestingT(t)

			tmpDir := DirStructure(t, testCase.DirStructure)

			vcsClient := vcsmocks.NewMockClient()
			When(vcsClient.GetModifiedFiles(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(testCase.ModifiedFiles, nil)

			workingDir := mocks.NewMockWorkingDir()
			When(workingDir.Clone(
				Any[models.Repo](),
				Any[models.PullRequest](),
				Any[string]())).ThenReturn(tmpDir, false, nil)

			When(workingDir.GetWorkingDir(
				Any[models.Repo](),
				Any[models.PullRequest](),
				Any[string]())).ThenReturn(tmpDir, nil)

			globalCfgArgs := valid.GlobalCfgArgs{
				AllowRepoCfg:  true,
				MergeableReq:  false,
				ApprovedReq:   false,
				UnDivergedReq: false,
			}

			terraformClient := terraform_mocks.NewMockClient()
			When(terraformClient.DetectVersion(Any[logging.SimpleLogging](), Any[string]())).Then(func(params []Param) ReturnValues {
				projectName := filepath.Base(params[1].(string))
				testVersion := testCase.Exp[projectName]
				if testVersion != "" {
					v, _ := version.NewVersion(testVersion)
					return []ReturnValue{v}
				}
				return nil
			})

			builder := events.NewProjectCommandBuilder(
				false,
				&config.ParserValidator{},
				&events.DefaultProjectFinder{},
				vcsClient,
				workingDir,
				events.NewDefaultWorkingDirLocker(),
				valid.NewGlobalCfgFromArgs(globalCfgArgs),
				&events.DefaultPendingPlanFinder{},
				&events.CommentParser{ExecutableName: "atlantis"},
				false,
				false,
				false,
				false,
				false,
				"",
				"**/*.tf,**/*.tfvars,**/*.tfvars.json,**/terragrunt.hcl,**/.terraform.lock.hcl",
				false,
				false,
				scope,
				logger,
				terraformClient,
			)

			actCtxs, err := builder.BuildPlanCommands(
				&command.Context{
					Log:   logger,
					Scope: scope,
				},
				&events.CommentCommand{
					RepoRelDir: "",
					Flags:      nil,
					Name:       command.Plan,
					Verbose:    false,
				})

			Ok(t, err)
			Equals(t, len(testCase.Exp), len(actCtxs))
			for _, actCtx := range actCtxs {
				if testCase.Exp[actCtx.RepoRelDir] != "" {
					Assert(t, actCtx.TerraformVersion != nil, "TerraformVersion is nil, not %s for %s", testCase.Exp[actCtx.RepoRelDir], actCtx.RepoRelDir)
					Equals(t, testCase.Exp[actCtx.RepoRelDir], actCtx.TerraformVersion.String())
				} else {
					Assert(t, actCtx.TerraformVersion == nil, "TerraformVersion is supposed to be nil.")
				}
			}
		})
	}
}

// Test that we don't clone the repo if there were no changes based on the atlantis.yaml file.
func TestDefaultProjectCommandBuilder_SkipCloneNoChanges(t *testing.T) {
	cases := []struct {
		AtlantisYAML   string
		ExpectedCtxs   int
		ExpectedClones InvocationCountMatcher
		ModifiedFiles  []string
	}{
		{
			AtlantisYAML: `
version: 3
projects:
- dir: dir1`,
			ExpectedCtxs:   0,
			ExpectedClones: Never(),
			ModifiedFiles:  []string{"dir2/main.tf"},
		},
		{
			AtlantisYAML: `
version: 3
parallel_plan: true`,
			ExpectedCtxs:   0,
			ExpectedClones: Once(),
			ModifiedFiles:  []string{"README.md"},
		},
	}

	for _, c := range cases {
		RegisterMockTestingT(t)
		vcsClient := vcsmocks.NewMockClient()
		When(vcsClient.GetModifiedFiles(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(c.ModifiedFiles, nil)
		When(vcsClient.SupportsSingleFileDownload(Any[models.Repo]())).ThenReturn(true)
		When(vcsClient.GetFileContent(Any[models.PullRequest](), Any[string]())).ThenReturn(true, []byte(c.AtlantisYAML), nil)
		workingDir := mocks.NewMockWorkingDir()

		logger := logging.NewNoopLogger(t)

		globalCfgArgs := valid.GlobalCfgArgs{
			AllowRepoCfg:  true,
			MergeableReq:  false,
			ApprovedReq:   false,
			UnDivergedReq: false,
		}
		scope, _, _ := metrics.NewLoggingScope(logger, "atlantis")
		terraformClient := terraform_mocks.NewMockClient()
		When(terraformClient.ListAvailableVersions(Any[logging.SimpleLogging]())).ThenReturn([]string{}, nil)

		builder := events.NewProjectCommandBuilder(
			false,
			&config.ParserValidator{},
			&events.DefaultProjectFinder{},
			vcsClient,
			workingDir,
			events.NewDefaultWorkingDirLocker(),
			valid.NewGlobalCfgFromArgs(globalCfgArgs),
			&events.DefaultPendingPlanFinder{},
			&events.CommentParser{ExecutableName: "atlantis"},
			true,
			false,
			false,
			false,
			false,
			"",
			"**/*.tf,**/*.tfvars,**/*.tfvars.json,**/terragrunt.hcl,**/.terraform.lock.hcl",
			false,
			false,
			scope,
			logger,
			terraformClient,
		)

		var actCtxs []command.ProjectContext
		var err error
		actCtxs, err = builder.BuildAutoplanCommands(&command.Context{
			HeadRepo: models.Repo{},
			Pull:     models.PullRequest{},
			User:     models.User{},
			Log:      logger,
			Scope:    scope,
			PullRequestStatus: models.PullReqStatus{
				Mergeable: true,
			},
		})
		Ok(t, err)
		Equals(t, c.ExpectedCtxs, len(actCtxs))
		workingDir.VerifyWasCalled(c.ExpectedClones).Clone(Any[models.Repo](), Any[models.PullRequest](), Any[string]())
	}
}

func TestDefaultProjectCommandBuilder_WithPolicyCheckEnabled_BuildAutoplanCommand(t *testing.T) {
	RegisterMockTestingT(t)
	tmpDir := DirStructure(t, map[string]interface{}{
		"main.tf": nil,
	})

	logger := logging.NewNoopLogger(t)
	scope, _, _ := metrics.NewLoggingScope(logger, "atlantis")

	workingDir := mocks.NewMockWorkingDir()
	When(workingDir.Clone(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(tmpDir, false, nil)
	vcsClient := vcsmocks.NewMockClient()
	When(vcsClient.GetModifiedFiles(Any[models.Repo](), Any[models.PullRequest]())).ThenReturn([]string{"main.tf"}, nil)

	globalCfgArgs := valid.GlobalCfgArgs{
		AllowRepoCfg:       false,
		MergeableReq:       false,
		ApprovedReq:        false,
		UnDivergedReq:      false,
		PolicyCheckEnabled: true,
	}

	globalCfg := valid.NewGlobalCfgFromArgs(globalCfgArgs)
	terraformClient := terraform_mocks.NewMockClient()
	When(terraformClient.ListAvailableVersions(Any[logging.SimpleLogging]())).ThenReturn([]string{}, nil)

	builder := events.NewProjectCommandBuilder(
		true,
		&config.ParserValidator{},
		&events.DefaultProjectFinder{},
		vcsClient,
		workingDir,
		events.NewDefaultWorkingDirLocker(),
		globalCfg,
		&events.DefaultPendingPlanFinder{},
		&events.CommentParser{ExecutableName: "atlantis"},
		false,
		false,
		false,
		false,
		false,
		"",
		"**/*.tf,**/*.tfvars,**/*.tfvars.json,**/terragrunt.hcl,**/.terraform.lock.hcl",
		false,
		false,
		scope,
		logger,
		terraformClient,
	)

	ctxs, err := builder.BuildAutoplanCommands(&command.Context{
		PullRequestStatus: models.PullReqStatus{
			Mergeable: true,
		},
		Log:   logger,
		Scope: scope,
	})

	Ok(t, err)
	Equals(t, 2, len(ctxs))
	planCtx := ctxs[0]
	policyCheckCtx := ctxs[1]
	Equals(t, command.Plan, planCtx.CommandName)
	Equals(t, globalCfg.Workflows["default"].Plan.Steps, planCtx.Steps)
	Equals(t, command.PolicyCheck, policyCheckCtx.CommandName)
	Equals(t, globalCfg.Workflows["default"].PolicyCheck.Steps, policyCheckCtx.Steps)
}

// Test building version command for multiple projects
func TestDefaultProjectCommandBuilder_BuildVersionCommand(t *testing.T) {
	RegisterMockTestingT(t)
	tmpDir := DirStructure(t, map[string]interface{}{
		"workspace1": map[string]interface{}{
			"project1": map[string]interface{}{
				"main.tf":          nil,
				"workspace.tfplan": nil,
			},
			"project2": map[string]interface{}{
				"main.tf":          nil,
				"workspace.tfplan": nil,
			},
		},
		"workspace2": map[string]interface{}{
			"project1": map[string]interface{}{
				"main.tf":          nil,
				"workspace.tfplan": nil,
			},
			"project2": map[string]interface{}{
				"main.tf":          nil,
				"workspace.tfplan": nil,
			},
		},
	})
	// Initialize git repos in each workspace so that the .tfplan files get
	// picked up.
	runCmd(t, filepath.Join(tmpDir, "workspace1"), "git", "init")
	runCmd(t, filepath.Join(tmpDir, "workspace2"), "git", "init")

	workingDir := mocks.NewMockWorkingDir()
	When(workingDir.GetPullDir(
		Any[models.Repo](),
		Any[models.PullRequest]())).
		ThenReturn(tmpDir, nil)

	logger := logging.NewNoopLogger(t)
	scope, _, _ := metrics.NewLoggingScope(logger, "atlantis")

	globalCfgArgs := valid.GlobalCfgArgs{
		AllowRepoCfg:  false,
		MergeableReq:  false,
		ApprovedReq:   false,
		UnDivergedReq: false,
	}
	terraformClient := terraform_mocks.NewMockClient()
	When(terraformClient.ListAvailableVersions(Any[logging.SimpleLogging]())).ThenReturn([]string{}, nil)

	builder := events.NewProjectCommandBuilder(
		false,
		&config.ParserValidator{},
		&events.DefaultProjectFinder{},
		nil,
		workingDir,
		events.NewDefaultWorkingDirLocker(),
		valid.NewGlobalCfgFromArgs(globalCfgArgs),
		&events.DefaultPendingPlanFinder{},
		&events.CommentParser{ExecutableName: "atlantis"},
		false,
		false,
		false,
		false,
		false,
		"",
		"**/*.tf,**/*.tfvars,**/*.tfvars.json,**/terragrunt.hcl,**/.terraform.lock.hcl",
		false,
		false,
		scope,
		logger,
		terraformClient,
	)

	ctxs, err := builder.BuildVersionCommands(
		&command.Context{
			Log:   logger,
			Scope: scope,
		},
		&events.CommentCommand{
			RepoRelDir:  "",
			Flags:       nil,
			Name:        command.Version,
			Verbose:     false,
			Workspace:   "",
			ProjectName: "",
		})
	Ok(t, err)
	Equals(t, 4, len(ctxs))
	Equals(t, "project1", ctxs[0].RepoRelDir)
	Equals(t, "workspace1", ctxs[0].Workspace)
	Equals(t, "project2", ctxs[1].RepoRelDir)
	Equals(t, "workspace1", ctxs[1].Workspace)
	Equals(t, "project1", ctxs[2].RepoRelDir)
	Equals(t, "workspace2", ctxs[2].Workspace)
	Equals(t, "project2", ctxs[3].RepoRelDir)
	Equals(t, "workspace2", ctxs[3].Workspace)
}
