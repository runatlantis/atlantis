// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/terraform"
	tfclientmocks "github.com/runatlantis/atlantis/server/core/terraform/tfclient/mocks"
	"github.com/runatlantis/atlantis/server/metrics/metricstest"

	"github.com/runatlantis/atlantis/server/core/config"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

var defaultUserConfig = struct {
	SkipCloneNoChanges       bool
	EnableRegExpCmd          bool
	EnableAutoMerge          bool
	EnableParallelPlan       bool
	EnableParallelApply      bool
	AutoDetectModuleFiles    string
	AutoplanFileList         string
	RestrictFileList         bool
	SilenceNoProjects        bool
	IncludeGitUntrackedFiles bool
	AutoDiscoverMode         string
}{
	SkipCloneNoChanges:       false,
	EnableRegExpCmd:          false,
	EnableAutoMerge:          false,
	EnableParallelPlan:       false,
	EnableParallelApply:      false,
	AutoDetectModuleFiles:    "",
	AutoplanFileList:         "**/*.tf,**/*.tfvars,**/*.tfvars.json,**/terragrunt.hcl,**/.terraform.lock.hcl",
	RestrictFileList:         false,
	SilenceNoProjects:        false,
	IncludeGitUntrackedFiles: false,
	AutoDiscoverMode:         "auto",
}

func ChangedFiles(dirStructure map[string]any, parent string) []string {
	var files []string
	for k, v := range dirStructure {
		switch v := v.(type) {
		case map[string]any:
			files = append(files, ChangedFiles(v, k)...)
		default:
			files = append(files, filepath.Join(parent, k))
		}
	}
	return files
}

func TestDefaultProjectCommandBuilder_BuildAutoplanCommands(t *testing.T) {
	// expCtxFields define the ctx fields we're going to assert on.
	// Since we're focused on autoplanning here, we don't validate all the
	// fields so the tests are more obvious and targeted.
	type expCtxFields struct {
		ProjectName string
		RepoRelDir  string
		Workspace   string
	}
	defaultTestDirStructure := map[string]any{
		"main.tf": nil,
	}

	cases := []struct {
		Description      string
		AtlantisYAML     string
		ServerSideYAML   string
		TestDirStructure map[string]any
		exp              []expCtxFields
	}{
		{
			Description: "simple atlantis.yaml",
			AtlantisYAML: `
version: 3
projects:
- dir: .
`,
			TestDirStructure: defaultTestDirStructure,
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
			TestDirStructure: defaultTestDirStructure,
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
			TestDirStructure: defaultTestDirStructure,
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
			TestDirStructure: defaultTestDirStructure,
			exp:              nil,
		},
		{
			Description: "workspaces from subdirectories detected",
			TestDirStructure: map[string]any{
				"work": map[string]any{
					"main.tf": `
terraform {
  cloud {
    organization = "atlantis-test"
    workspaces {
      name = "test-workspace1"
    }
  }
}`,
				},
				"test": map[string]any{
					"main.tf": `
terraform {
  cloud {
    organization = "atlantis-test"
    workspaces {
      name = "test-workspace12"
    }
  }
}`,
				},
			},
			exp: []expCtxFields{
				{
					ProjectName: "",
					RepoRelDir:  "test",
					Workspace:   "test-workspace12",
				},
				{
					ProjectName: "",
					RepoRelDir:  "work",
					Workspace:   "test-workspace1",
				},
			},
		},
		{
			Description: "workspaces in parent directory are detected",
			TestDirStructure: map[string]any{
				"main.tf": `
terraform {
  cloud {
    organization = "atlantis-test"
    workspaces {
      name = "test-workspace"
    }
  }
}`,
			},
			exp: []expCtxFields{
				{
					ProjectName: "",
					RepoRelDir:  ".",
					Workspace:   "test-workspace",
				},
			},
		},
	}

	logger := logging.NewNoopLogger(t)
	scope := metricstest.NewLoggingScope(t, logger, "atlantis")
	userConfig := defaultUserConfig

	terraformClient := tfclientmocks.NewMockClient()

	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			RegisterMockTestingT(t)
			tmpDir := DirStructure(t, c.TestDirStructure)
			workingDir := mocks.NewMockWorkingDir()
			When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](),
				Any[string]())).ThenReturn(tmpDir, nil)
			vcsClient := vcsmocks.NewMockClient()
			When(vcsClient.GetModifiedFiles(Any[logging.SimpleLogging](), Any[models.Repo](),
				Any[models.PullRequest]())).ThenReturn(ChangedFiles(c.TestDirStructure, ""), nil)
			if c.AtlantisYAML != "" {
				err := os.WriteFile(filepath.Join(tmpDir, valid.DefaultAtlantisFile), []byte(c.AtlantisYAML), 0600)
				Ok(t, err)
			}

			globalCfgArgs := valid.GlobalCfgArgs{}

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
				userConfig.SkipCloneNoChanges,
				userConfig.EnableRegExpCmd,
				userConfig.EnableAutoMerge,
				userConfig.EnableParallelPlan,
				userConfig.EnableParallelApply,
				userConfig.AutoDetectModuleFiles,
				userConfig.AutoplanFileList,
				userConfig.RestrictFileList,
				userConfig.SilenceNoProjects,
				userConfig.IncludeGitUntrackedFiles,
				userConfig.AutoDiscoverMode,
				scope,
				terraformClient,
			)

			ctxs, err := builder.BuildAutoplanCommands(&command.Context{
				PullRequestStatus: models.PullReqStatus{
					MergeableStatus: models.MergeableStatus{IsMergeable: true},
				},
				Log:   logger,
				Scope: scope,
			})
			Ok(t, err)
			Equals(t, len(c.exp), len(ctxs))

			// Sort so comparisons are deterministic
			sort.Slice(ctxs, func(i, j int) bool {
				if ctxs[i].ProjectName != ctxs[j].ProjectName {
					return ctxs[i].ProjectName < ctxs[j].ProjectName
				}
				if ctxs[i].RepoRelDir != ctxs[j].RepoRelDir {
					return ctxs[i].RepoRelDir < ctxs[j].RepoRelDir
				}
				return ctxs[i].Workspace < ctxs[j].Workspace
			})
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
		AutoDiscoverModeUserCfg    string
		EnableParallelPlanUserCfg  bool
		EnableParallelApplyUserCfg bool
		ExpAutoMerge               bool
		ExpParallelPlan            bool
		ExpParallelApply           bool
		ExpNoProjects              bool
		API                        bool
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
			Description: "atlantis.yaml with projectname and workspace",
			Cmd: events.CommentCommand{
				Name:        command.Plan,
				ProjectName: "myproject",
				Workspace:   "staging",
			},
			AtlantisYAML: `
version: 3
projects:
- name: otherproject
  dir: production
  workspace: production
- name: myproject
  dir: staging
  workspace: staging
  apply_requirements: [approved]`,
			ExpApplyReqs:   []string{"approved"},
			ExpProjectName: "myproject",
			ExpWorkspace:   "staging",
			ExpDir:         "staging",
			API:            true,
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
			ExpErr: "must specify project name: more than one project defined in 'atlantis.yaml' matched dir: '.' workspace: 'myworkspace'",
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
			ExpErr: "no project with name 'notconfigured' is defined in 'atlantis.yaml'",
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
	scope := metricstest.NewLoggingScope(t, logger, "atlantis")
	userConfig := defaultUserConfig

	for _, c := range cases {
		// NOTE: we're testing both plan and apply here.
		for _, cmdName := range []command.Name{command.Plan, command.Apply} {
			t.Run(c.Description+"_"+cmdName.String(), func(t *testing.T) {
				RegisterMockTestingT(t)
				tmpDir := DirStructure(t, map[string]any{
					"main.tf": nil,
				})

				workingDir := mocks.NewMockWorkingDir()
				When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](),
					Any[string]())).ThenReturn(tmpDir, nil)
				When(workingDir.GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(tmpDir, nil)
				vcsClient := vcsmocks.NewMockClient()
				When(vcsClient.GetModifiedFiles(Any[logging.SimpleLogging](), Any[models.Repo](),
					Any[models.PullRequest]())).ThenReturn([]string{"main.tf"}, nil)
				if c.AtlantisYAML != "" {
					err := os.WriteFile(filepath.Join(tmpDir, valid.DefaultAtlantisFile), []byte(c.AtlantisYAML), 0600)
					Ok(t, err)
				}

				globalCfgArgs := valid.GlobalCfgArgs{
					AllowAllRepoSettings: true,
				}

				terraformClient := tfclientmocks.NewMockClient()

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
					userConfig.SkipCloneNoChanges,
					userConfig.EnableRegExpCmd,
					c.EnableAutoMergeUserCfg,
					c.EnableParallelPlanUserCfg,
					c.EnableParallelApplyUserCfg,
					userConfig.AutoDetectModuleFiles,
					userConfig.AutoplanFileList,
					userConfig.RestrictFileList,
					c.Silenced,
					userConfig.IncludeGitUntrackedFiles,
					c.AutoDiscoverModeUserCfg,
					scope,
					terraformClient,
				)

				var actCtxs []command.ProjectContext
				var err error
				cmd := c.Cmd
				if cmdName == command.Plan {
					actCtxs, err = builder.BuildPlanCommands(&command.Context{
						Log:   logger,
						Scope: scope,
						API:   c.API,
					}, &cmd)
				} else {
					actCtxs, err = builder.BuildApplyCommands(&command.Context{Log: logger, Scope: scope, API: c.API}, &cmd)
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

func TestDefaultProjectCommandBuilder_BuildPlanCommandsDiscoverAllProjectsSkipsModifiedFiles(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	scope := metricstest.NewLoggingScope(t, logger, "atlantis")
	tmpDir := DirStructure(t, map[string]any{
		"project1": map[string]any{
			"main.tf": nil,
		},
		"project2": map[string]any{
			"main.tf": nil,
		},
		"modules": map[string]any{
			"network": map[string]any{
				"main.tf": nil,
			},
		},
	})

	workingDir := mocks.NewMockWorkingDir()
	When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[string]())).
		ThenReturn(tmpDir, nil)
	vcsClient := vcsmocks.NewMockClient()
	terraformClient := tfclientmocks.NewMockClient()

	builder := events.NewProjectCommandBuilder(
		false,
		&config.ParserValidator{},
		&events.DefaultProjectFinder{},
		vcsClient,
		workingDir,
		events.NewDefaultWorkingDirLocker(),
		valid.NewGlobalCfgFromArgs(valid.GlobalCfgArgs{AllowAllRepoSettings: true}),
		&events.DefaultPendingPlanFinder{},
		&events.CommentParser{ExecutableName: "atlantis"},
		defaultUserConfig.SkipCloneNoChanges,
		defaultUserConfig.EnableRegExpCmd,
		defaultUserConfig.EnableAutoMerge,
		defaultUserConfig.EnableParallelPlan,
		defaultUserConfig.EnableParallelApply,
		defaultUserConfig.AutoDetectModuleFiles,
		defaultUserConfig.AutoplanFileList,
		defaultUserConfig.RestrictFileList,
		defaultUserConfig.SilenceNoProjects,
		defaultUserConfig.IncludeGitUntrackedFiles,
		defaultUserConfig.AutoDiscoverMode,
		scope,
		terraformClient,
	)

	ctxs, err := builder.BuildPlanCommands(&command.Context{
		Log:      logger,
		Scope:    scope,
		HeadRepo: models.Repo{FullName: "owner/repo"},
		Pull: models.PullRequest{
			Num:      -1,
			BaseRepo: models.Repo{FullName: "owner/repo"},
		},
		API: true,
	}, &events.CommentCommand{Name: command.Plan, DiscoverAllProjects: true})
	Ok(t, err)

	dirs := make([]string, 0, len(ctxs))
	for _, ctx := range ctxs {
		dirs = append(dirs, ctx.RepoRelDir)
	}
	sort.Strings(dirs)
	Equals(t, []string{"project1", "project2"}, dirs)
	vcsClient.VerifyWasCalled(Never()).GetModifiedFiles(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest]())
}

// Test that autodiscover.ignore_paths blocks targeted plan/apply -d commands
// when the directory has no explicit project config (global config ignore_paths).
func TestDefaultProjectCommandBuilder_BuildTargetedCommand_IgnorePaths(t *testing.T) {
	RegisterMockTestingT(t)

	tmpDir := DirStructure(t, map[string]any{
		"environments": map[string]any{
			"prod": map[string]any{
				"main.tf": nil,
			},
			"nonprod": map[string]any{
				"main.tf": nil,
			},
		},
	})

	workingDir := mocks.NewMockWorkingDir()
	When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](),
		Any[string]())).ThenReturn(tmpDir, nil)
	When(workingDir.GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(tmpDir, nil)
	vcsClient := vcsmocks.NewMockClient()
	When(vcsClient.GetModifiedFiles(Any[logging.SimpleLogging](), Any[models.Repo](),
		Any[models.PullRequest]())).ThenReturn([]string{"environments/prod/main.tf", "environments/nonprod/main.tf"}, nil)

	logger := logging.NewNoopLogger(t)
	scope := metricstest.NewLoggingScope(t, logger, "atlantis")

	globalCfgArgs := valid.GlobalCfgArgs{AllowAllRepoSettings: true}
	globalCfg := valid.NewGlobalCfgFromArgs(globalCfgArgs)
	globalCfg.Repos[0].AutoDiscover = &valid.AutoDiscover{
		Mode:        valid.AutoDiscoverEnabledMode,
		IgnorePaths: []string{"environments/prod/**"},
	}

	terraformClient := tfclientmocks.NewMockClient()
	userConfig := defaultUserConfig

	builder := events.NewProjectCommandBuilder(
		false,
		&config.ParserValidator{},
		&events.DefaultProjectFinder{},
		vcsClient,
		workingDir,
		events.NewDefaultWorkingDirLocker(),
		globalCfg,
		&events.DefaultPendingPlanFinder{},
		&events.CommentParser{ExecutableName: "atlantis"},
		userConfig.SkipCloneNoChanges,
		userConfig.EnableRegExpCmd,
		userConfig.EnableAutoMerge,
		userConfig.EnableParallelPlan,
		userConfig.EnableParallelApply,
		userConfig.AutoDetectModuleFiles,
		userConfig.AutoplanFileList,
		userConfig.RestrictFileList,
		userConfig.SilenceNoProjects,
		userConfig.IncludeGitUntrackedFiles,
		userConfig.AutoDiscoverMode,
		scope,
		terraformClient,
	)

	cmdCtx := &command.Context{Log: logger, Scope: scope}

	// Targeted plan -d to ignored path should return no projects
	planCtxs, err := builder.BuildPlanCommands(cmdCtx, &events.CommentCommand{
		Name:       command.Plan,
		RepoRelDir: "environments/prod",
		Workspace:  "default",
	})
	Assert(t, errors.Is(err, events.ErrIgnoredTargetedDir), "expected ignored targeted dir error, got %v", err)
	Equals(t, 0, len(planCtxs))

	// Targeted plan -d to non-ignored path should succeed
	planCtxs, err = builder.BuildPlanCommands(cmdCtx, &events.CommentCommand{
		Name:       command.Plan,
		RepoRelDir: "environments/nonprod",
		Workspace:  "default",
	})
	Ok(t, err)
	Equals(t, 1, len(planCtxs))
	Equals(t, "environments/nonprod", planCtxs[0].RepoRelDir)

	// Targeted apply -d to ignored path should return no projects
	applyCtxs, err := builder.BuildApplyCommands(cmdCtx, &events.CommentCommand{
		Name:       command.Apply,
		RepoRelDir: "environments/prod",
		Workspace:  "default",
	})
	Assert(t, errors.Is(err, events.ErrIgnoredTargetedDir), "expected ignored targeted dir error, got %v", err)
	Equals(t, 0, len(applyCtxs))

	// Targeted apply -d to non-ignored path should succeed
	applyCtxs, err = builder.BuildApplyCommands(cmdCtx, &events.CommentCommand{
		Name:       command.Apply,
		RepoRelDir: "environments/nonprod",
		Workspace:  "default",
	})
	Ok(t, err)
	Equals(t, 1, len(applyCtxs))
	Equals(t, "environments/nonprod", applyCtxs[0].RepoRelDir)
}

func TestDefaultProjectCommandBuilder_BuildWorkspaceOnlyCommand_IgnorePathsNotSkipped(t *testing.T) {
	RegisterMockTestingT(t)

	tmpDir := DirStructure(t, map[string]any{
		"main.tf": nil,
	})
	workingDir := mocks.NewMockWorkingDir()
	When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(tmpDir, nil)
	When(workingDir.GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(tmpDir, nil)
	vcsClient := vcsmocks.NewMockClient()

	logger := logging.NewNoopLogger(t)
	scope := metricstest.NewLoggingScope(t, logger, "atlantis")
	globalCfg := valid.NewGlobalCfgFromArgs(valid.GlobalCfgArgs{AllowAllRepoSettings: true})
	globalCfg.Repos[0].AutoDiscover = &valid.AutoDiscover{
		Mode:        valid.AutoDiscoverEnabledMode,
		IgnorePaths: []string{"**"},
	}
	terraformClient := tfclientmocks.NewMockClient()
	userConfig := defaultUserConfig
	builder := events.NewProjectCommandBuilder(
		false,
		&config.ParserValidator{},
		&events.DefaultProjectFinder{},
		vcsClient,
		workingDir,
		events.NewDefaultWorkingDirLocker(),
		globalCfg,
		&events.DefaultPendingPlanFinder{},
		&events.CommentParser{ExecutableName: "atlantis"},
		userConfig.SkipCloneNoChanges,
		userConfig.EnableRegExpCmd,
		userConfig.EnableAutoMerge,
		userConfig.EnableParallelPlan,
		userConfig.EnableParallelApply,
		userConfig.AutoDetectModuleFiles,
		userConfig.AutoplanFileList,
		userConfig.RestrictFileList,
		userConfig.SilenceNoProjects,
		userConfig.IncludeGitUntrackedFiles,
		userConfig.AutoDiscoverMode,
		scope,
		terraformClient,
	)
	repo := models.Repo{FullName: "runatlantis/atlantis", Owner: "runatlantis", Name: "atlantis"}
	cmdCtx := &command.Context{
		Log:      logger,
		Scope:    scope,
		Pull:     models.PullRequest{BaseRepo: repo, Num: 1, HeadBranch: "feature", BaseBranch: "main"},
		HeadRepo: repo,
	}

	planCtxs, err := builder.BuildPlanCommands(cmdCtx, &events.CommentCommand{
		Name:      command.Plan,
		Workspace: "staging",
	})
	Ok(t, err)
	Equals(t, 1, len(planCtxs))
	Equals(t, events.DefaultRepoRelDir, planCtxs[0].RepoRelDir)
	Equals(t, "staging", planCtxs[0].Workspace)

	applyCtxs, err := builder.BuildApplyCommands(cmdCtx, &events.CommentCommand{
		Name:      command.Apply,
		Workspace: "staging",
	})
	Ok(t, err)
	Equals(t, 1, len(applyCtxs))
	Equals(t, events.DefaultRepoRelDir, applyCtxs[0].RepoRelDir)
	Equals(t, "staging", applyCtxs[0].Workspace)
}

func TestDefaultProjectCommandBuilder_BuildTargetedNonPlanCommand_IgnorePathsWithoutWorkingDir(t *testing.T) {
	RegisterMockTestingT(t)

	workingDir := mocks.NewMockWorkingDir()
	When(workingDir.GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn("", os.ErrNotExist)
	vcsClient := vcsmocks.NewMockClient()

	logger := logging.NewNoopLogger(t)
	scope := metricstest.NewLoggingScope(t, logger, "atlantis")

	globalCfgArgs := valid.GlobalCfgArgs{AllowAllRepoSettings: true}
	globalCfg := valid.NewGlobalCfgFromArgs(globalCfgArgs)
	globalCfg.Repos[0].AutoDiscover = &valid.AutoDiscover{
		Mode:        valid.AutoDiscoverEnabledMode,
		IgnorePaths: []string{"environments/prod/**"},
	}

	terraformClient := tfclientmocks.NewMockClient()
	userConfig := defaultUserConfig

	builder := events.NewProjectCommandBuilder(
		false,
		&config.ParserValidator{},
		&events.DefaultProjectFinder{},
		vcsClient,
		workingDir,
		events.NewDefaultWorkingDirLocker(),
		globalCfg,
		&events.DefaultPendingPlanFinder{},
		&events.CommentParser{ExecutableName: "atlantis"},
		userConfig.SkipCloneNoChanges,
		userConfig.EnableRegExpCmd,
		userConfig.EnableAutoMerge,
		userConfig.EnableParallelPlan,
		userConfig.EnableParallelApply,
		userConfig.AutoDetectModuleFiles,
		userConfig.AutoplanFileList,
		userConfig.RestrictFileList,
		userConfig.SilenceNoProjects,
		userConfig.IncludeGitUntrackedFiles,
		userConfig.AutoDiscoverMode,
		scope,
		terraformClient,
	)

	repo := models.Repo{FullName: "runatlantis/atlantis", Owner: "runatlantis", Name: "atlantis"}
	cmdCtx := &command.Context{
		Log:      logger,
		Scope:    scope,
		Pull:     models.PullRequest{BaseRepo: repo, Num: 1, HeadBranch: "feature", BaseBranch: "main"},
		HeadRepo: repo,
	}

	cases := []struct {
		description string
		cmd         events.CommentCommand
		build       func(*command.Context, *events.CommentCommand) ([]command.ProjectContext, error)
	}{
		{
			description: "apply",
			cmd:         events.CommentCommand{Name: command.Apply, RepoRelDir: "environments/prod", Workspace: "default"},
			build:       builder.BuildApplyCommands,
		},
		{
			description: "approve policies",
			cmd:         events.CommentCommand{Name: command.ApprovePolicies, RepoRelDir: "environments/prod", Workspace: "default"},
			build:       builder.BuildApprovePoliciesCommands,
		},
		{
			description: "import",
			cmd:         events.CommentCommand{Name: command.Import, RepoRelDir: "environments/prod", Workspace: "default"},
			build:       builder.BuildImportCommands,
		},
		{
			description: "version",
			cmd:         events.CommentCommand{Name: command.Version, RepoRelDir: "environments/prod", Workspace: "default"},
			build:       builder.BuildVersionCommands,
		},
		{
			description: "state rm",
			cmd:         events.CommentCommand{Name: command.State, SubName: "rm", RepoRelDir: "environments/prod", Workspace: "default"},
			build:       builder.BuildStateRmCommands,
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			ctxs, err := c.build(cmdCtx, &c.cmd)
			Assert(t, errors.Is(err, events.ErrIgnoredTargetedDir), "expected ignored targeted dir error, got %v", err)
			Equals(t, 0, len(ctxs))
		})
	}
}

func TestDefaultProjectCommandBuilder_ShouldIgnoreTargetedDirPrefersGeneratedLocalConfig(t *testing.T) {
	RegisterMockTestingT(t)

	remoteAtlantisYAML := `
version: 3
autodiscover:
  mode: enabled
  ignore_paths:
  - "environments/prod/**"
`
	localAtlantisYAML := `
version: 3
autodiscover:
  mode: enabled
  ignore_paths:
  - "environments/prod/**"
projects:
- dir: environments/prod
`
	tmpDir := DirStructure(t, map[string]any{
		"environments": map[string]any{
			"prod": map[string]any{
				"main.tf": nil,
			},
		},
	})
	err := os.WriteFile(filepath.Join(tmpDir, valid.DefaultAtlantisFile), []byte(localAtlantisYAML), 0600)
	Ok(t, err)

	workingDir := mocks.NewMockWorkingDir()
	When(workingDir.GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Eq(events.DefaultWorkspace))).ThenReturn(tmpDir, nil)
	vcsClient := vcsmocks.NewMockClient()
	When(vcsClient.GetFileContent(Any[logging.SimpleLogging](), Any[models.Repo](), Eq("feature"), Eq(valid.DefaultAtlantisFile))).ThenReturn(true, []byte(remoteAtlantisYAML), nil)

	logger := logging.NewNoopLogger(t)
	scope := metricstest.NewLoggingScope(t, logger, "atlantis")
	globalCfg := valid.NewGlobalCfgFromArgs(valid.GlobalCfgArgs{AllowAllRepoSettings: true})
	terraformClient := tfclientmocks.NewMockClient()
	userConfig := defaultUserConfig
	builder := events.NewProjectCommandBuilder(
		false,
		&config.ParserValidator{},
		&events.DefaultProjectFinder{},
		vcsClient,
		workingDir,
		events.NewDefaultWorkingDirLocker(),
		globalCfg,
		&events.DefaultPendingPlanFinder{},
		&events.CommentParser{ExecutableName: "atlantis"},
		userConfig.SkipCloneNoChanges,
		userConfig.EnableRegExpCmd,
		userConfig.EnableAutoMerge,
		userConfig.EnableParallelPlan,
		userConfig.EnableParallelApply,
		userConfig.AutoDetectModuleFiles,
		userConfig.AutoplanFileList,
		userConfig.RestrictFileList,
		userConfig.SilenceNoProjects,
		userConfig.IncludeGitUntrackedFiles,
		userConfig.AutoDiscoverMode,
		scope,
		terraformClient,
	)
	baseRepo := models.Repo{Owner: "owner", Name: "repo", FullName: "owner/repo", VCSHost: models.VCSHost{Type: models.Github}}
	ctx := &command.Context{
		Log:      logger,
		Scope:    scope,
		HeadRepo: baseRepo,
		Pull: models.PullRequest{
			Num:        1,
			BaseBranch: "main",
			HeadBranch: "feature",
			BaseRepo:   baseRepo,
		},
	}
	cmd := &events.CommentCommand{Name: command.Plan, RepoRelDir: "environments/prod", Workspace: events.DefaultWorkspace}

	Assert(t, builder.ShouldIgnoreTargetedDir(ctx, cmd), "expected remote config without explicit project to ignore target")

	ctx.PreferLocalRepoCfgForTargetedIgnore = true
	Assert(t, !builder.ShouldIgnoreTargetedDir(ctx, cmd), "expected generated local explicit project to prevent ignored-target skip")
}

func TestDefaultProjectCommandBuilder_ShouldIgnoreTargetedDirUsesHeadCommitForRemoteConfig(t *testing.T) {
	RegisterMockTestingT(t)

	atlantisYAML := "version: 3\n" +
		"autodiscover:\n" +
		"  mode: enabled\n" +
		"  ignore_paths:\n" +
		"  - \"environments/prod/**\"\n"

	workingDir := mocks.NewMockWorkingDir()
	When(workingDir.GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Eq(events.DefaultWorkspace))).ThenReturn("", os.ErrNotExist)
	vcsClient := vcsmocks.NewMockClient()
	When(vcsClient.GetFileContent(Any[logging.SimpleLogging](), Any[models.Repo](), Eq("abc123"), Eq(valid.DefaultAtlantisFile))).ThenReturn(true, []byte(atlantisYAML), nil)

	logger := logging.NewNoopLogger(t)
	scope := metricstest.NewLoggingScope(t, logger, "atlantis")
	globalCfg := valid.NewGlobalCfgFromArgs(valid.GlobalCfgArgs{AllowAllRepoSettings: true})
	terraformClient := tfclientmocks.NewMockClient()
	userConfig := defaultUserConfig
	builder := events.NewProjectCommandBuilder(
		false,
		&config.ParserValidator{},
		&events.DefaultProjectFinder{},
		vcsClient,
		workingDir,
		events.NewDefaultWorkingDirLocker(),
		globalCfg,
		&events.DefaultPendingPlanFinder{},
		&events.CommentParser{ExecutableName: "atlantis"},
		userConfig.SkipCloneNoChanges,
		userConfig.EnableRegExpCmd,
		userConfig.EnableAutoMerge,
		userConfig.EnableParallelPlan,
		userConfig.EnableParallelApply,
		userConfig.AutoDetectModuleFiles,
		userConfig.AutoplanFileList,
		userConfig.RestrictFileList,
		userConfig.SilenceNoProjects,
		userConfig.IncludeGitUntrackedFiles,
		userConfig.AutoDiscoverMode,
		scope,
		terraformClient,
	)
	baseRepo := models.Repo{Owner: "owner", Name: "repo", FullName: "owner/repo", VCSHost: models.VCSHost{Type: models.Github}}
	ctx := &command.Context{
		Log:      logger,
		Scope:    scope,
		HeadRepo: baseRepo,
		Pull: models.PullRequest{
			Num:        1,
			BaseBranch: "main",
			HeadBranch: "feature",
			HeadCommit: "abc123",
			BaseRepo:   baseRepo,
		},
	}
	cmd := &events.CommentCommand{Name: command.Plan, RepoRelDir: "environments/prod", Workspace: events.DefaultWorkspace}

	Assert(t, builder.ShouldIgnoreTargetedDir(ctx, cmd), "expected remote head commit config to ignore target")
	vcsClient.VerifyWasCalledOnce().GetFileContent(Any[logging.SimpleLogging](), Eq(baseRepo), Eq("abc123"), Eq(valid.DefaultAtlantisFile))
	vcsClient.VerifyWasCalled(Never()).GetFileContent(Any[logging.SimpleLogging](), Eq(baseRepo), Eq("feature"), Eq(valid.DefaultAtlantisFile))
}

func TestDefaultProjectCommandBuilder_ShouldIgnoreTargetedDirRespectsGlobProjectDirs(t *testing.T) {
	RegisterMockTestingT(t)

	atlantisYAML := "version: 3\n" +
		"autodiscover:\n" +
		"  mode: enabled\n" +
		"  ignore_paths:\n" +
		"  - \"modules/**\"\n" +
		"projects:\n" +
		"- dir: \"modules/*\"\n"

	workingDir := mocks.NewMockWorkingDir()
	vcsClient := vcsmocks.NewMockClient()
	When(vcsClient.GetFileContent(Any[logging.SimpleLogging](), Any[models.Repo](), Eq("abc123"), Eq(valid.DefaultAtlantisFile))).ThenReturn(true, []byte(atlantisYAML), nil)

	logger := logging.NewNoopLogger(t)
	scope := metricstest.NewLoggingScope(t, logger, "atlantis")
	globalCfg := valid.NewGlobalCfgFromArgs(valid.GlobalCfgArgs{AllowAllRepoSettings: true})
	terraformClient := tfclientmocks.NewMockClient()
	userConfig := defaultUserConfig
	builder := events.NewProjectCommandBuilder(
		false,
		&config.ParserValidator{},
		&events.DefaultProjectFinder{},
		vcsClient,
		workingDir,
		events.NewDefaultWorkingDirLocker(),
		globalCfg,
		&events.DefaultPendingPlanFinder{},
		&events.CommentParser{ExecutableName: "atlantis"},
		userConfig.SkipCloneNoChanges,
		userConfig.EnableRegExpCmd,
		userConfig.EnableAutoMerge,
		userConfig.EnableParallelPlan,
		userConfig.EnableParallelApply,
		userConfig.AutoDetectModuleFiles,
		userConfig.AutoplanFileList,
		userConfig.RestrictFileList,
		userConfig.SilenceNoProjects,
		userConfig.IncludeGitUntrackedFiles,
		userConfig.AutoDiscoverMode,
		scope,
		terraformClient,
	)
	baseRepo := models.Repo{Owner: "owner", Name: "repo", FullName: "owner/repo", VCSHost: models.VCSHost{Type: models.Github}}
	ctx := &command.Context{
		Log:      logger,
		Scope:    scope,
		HeadRepo: baseRepo,
		Pull: models.PullRequest{
			Num:        1,
			BaseBranch: "main",
			HeadBranch: "feature",
			HeadCommit: "abc123",
			BaseRepo:   baseRepo,
		},
	}
	cmd := &events.CommentCommand{Name: command.Apply, RepoRelDir: "modules/foo", Workspace: events.DefaultWorkspace}

	Assert(t, !builder.ShouldIgnoreTargetedDir(ctx, cmd), "expected glob-configured project dir to prevent ignored-target skip")
	vcsClient.VerifyWasCalledOnce().GetFileContent(Any[logging.SimpleLogging](), Eq(baseRepo), Eq("abc123"), Eq(valid.DefaultAtlantisFile))
}

func TestDefaultProjectCommandBuilder_ShouldIgnoreTargetedDirFailsOpenWhenRemoteConfigUnknown(t *testing.T) {
	RegisterMockTestingT(t)

	staleLocalAtlantisYAML := "version: 3\n" +
		"autodiscover:\n" +
		"  mode: enabled\n" +
		"  ignore_paths:\n" +
		"  - \"environments/prod/**\"\n"
	tmpDir := DirStructure(t, map[string]any{
		"environments": map[string]any{
			"prod": map[string]any{
				"main.tf": nil,
			},
		},
	})
	Ok(t, os.WriteFile(filepath.Join(tmpDir, valid.DefaultAtlantisFile), []byte(staleLocalAtlantisYAML), 0600))

	workingDir := mocks.NewMockWorkingDir()
	When(workingDir.GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Eq(events.DefaultWorkspace))).ThenReturn(tmpDir, nil)
	vcsClient := vcsmocks.NewMockClient()
	When(vcsClient.GetFileContent(Any[logging.SimpleLogging](), Any[models.Repo](), Eq("abc123"), Eq(valid.DefaultAtlantisFile))).ThenReturn(false, []byte{}, errors.New("not implemented"))

	logger := logging.NewNoopLogger(t)
	scope := metricstest.NewLoggingScope(t, logger, "atlantis")
	globalCfg := valid.NewGlobalCfgFromArgs(valid.GlobalCfgArgs{AllowAllRepoSettings: true})
	terraformClient := tfclientmocks.NewMockClient()
	userConfig := defaultUserConfig
	builder := events.NewProjectCommandBuilder(
		false,
		&config.ParserValidator{},
		&events.DefaultProjectFinder{},
		vcsClient,
		workingDir,
		events.NewDefaultWorkingDirLocker(),
		globalCfg,
		&events.DefaultPendingPlanFinder{},
		&events.CommentParser{ExecutableName: "atlantis"},
		userConfig.SkipCloneNoChanges,
		userConfig.EnableRegExpCmd,
		userConfig.EnableAutoMerge,
		userConfig.EnableParallelPlan,
		userConfig.EnableParallelApply,
		userConfig.AutoDetectModuleFiles,
		userConfig.AutoplanFileList,
		userConfig.RestrictFileList,
		userConfig.SilenceNoProjects,
		userConfig.IncludeGitUntrackedFiles,
		userConfig.AutoDiscoverMode,
		scope,
		terraformClient,
	)
	baseRepo := models.Repo{Owner: "owner", Name: "repo", FullName: "owner/repo", VCSHost: models.VCSHost{Type: models.Github}}
	ctx := &command.Context{
		Log:      logger,
		Scope:    scope,
		HeadRepo: baseRepo,
		Pull: models.PullRequest{
			Num:        1,
			BaseBranch: "main",
			HeadBranch: "feature",
			HeadCommit: "abc123",
			BaseRepo:   baseRepo,
		},
	}
	cmd := &events.CommentCommand{Name: command.Plan, RepoRelDir: "environments/prod", Workspace: events.DefaultWorkspace}

	Assert(t, !builder.ShouldIgnoreTargetedDir(ctx, cmd), "expected unknown current repo config to avoid pre-clone skip")
	workingDir.VerifyWasCalled(Never()).GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Any[string]())
}

func TestDefaultProjectCommandBuilder_ShouldIgnoreTargetedDirAllowsAuthoritativeMissingRemoteConfig(t *testing.T) {
	RegisterMockTestingT(t)

	workingDir := mocks.NewMockWorkingDir()
	When(workingDir.GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Eq(events.DefaultWorkspace))).ThenReturn("", os.ErrNotExist)
	vcsClient := vcsmocks.NewMockClient()
	When(vcsClient.GetFileContent(Any[logging.SimpleLogging](), Any[models.Repo](), Eq("abc123"), Eq(valid.DefaultAtlantisFile))).ThenReturn(false, []byte{}, nil)

	logger := logging.NewNoopLogger(t)
	scope := metricstest.NewLoggingScope(t, logger, "atlantis")
	globalCfg := valid.NewGlobalCfgFromArgs(valid.GlobalCfgArgs{AllowAllRepoSettings: true})
	globalCfg.Repos[0].AutoDiscover = &valid.AutoDiscover{
		Mode:        valid.AutoDiscoverEnabledMode,
		IgnorePaths: []string{"environments/prod/**"},
	}
	terraformClient := tfclientmocks.NewMockClient()
	userConfig := defaultUserConfig
	builder := events.NewProjectCommandBuilder(
		false,
		&config.ParserValidator{},
		&events.DefaultProjectFinder{},
		vcsClient,
		workingDir,
		events.NewDefaultWorkingDirLocker(),
		globalCfg,
		&events.DefaultPendingPlanFinder{},
		&events.CommentParser{ExecutableName: "atlantis"},
		userConfig.SkipCloneNoChanges,
		userConfig.EnableRegExpCmd,
		userConfig.EnableAutoMerge,
		userConfig.EnableParallelPlan,
		userConfig.EnableParallelApply,
		userConfig.AutoDetectModuleFiles,
		userConfig.AutoplanFileList,
		userConfig.RestrictFileList,
		userConfig.SilenceNoProjects,
		userConfig.IncludeGitUntrackedFiles,
		userConfig.AutoDiscoverMode,
		scope,
		terraformClient,
	)
	baseRepo := models.Repo{Owner: "owner", Name: "repo", FullName: "owner/repo", VCSHost: models.VCSHost{Type: models.Github}}
	ctx := &command.Context{
		Log:      logger,
		Scope:    scope,
		HeadRepo: baseRepo,
		Pull: models.PullRequest{
			Num:        1,
			BaseBranch: "main",
			HeadBranch: "feature",
			HeadCommit: "abc123",
			BaseRepo:   baseRepo,
		},
	}
	cmd := &events.CommentCommand{Name: command.Plan, RepoRelDir: "environments/prod", Workspace: events.DefaultWorkspace}

	Assert(t, builder.ShouldIgnoreTargetedDir(ctx, cmd), "expected authoritative missing repo config to allow global ignored-target skip")
	workingDir.VerifyWasCalled(Never()).GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Any[string]())
}

func TestDefaultProjectCommandBuilder_ShouldIgnoreTargetedDirUsesGlobalIgnoreWhenFileDownloadUnsupported(t *testing.T) {
	RegisterMockTestingT(t)

	workingDir := mocks.NewMockWorkingDir()
	When(workingDir.GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Eq(events.DefaultWorkspace))).ThenReturn("", os.ErrNotExist)
	vcsClient := vcsmocks.NewMockClient()
	When(vcsClient.GetFileContent(Any[logging.SimpleLogging](), Any[models.Repo](), Eq("abc123"), Eq(valid.DefaultAtlantisFile))).ThenReturn(false, []byte{}, errors.New("not implemented"))
	When(vcsClient.SupportsSingleFileDownload(Any[models.Repo]())).ThenReturn(false)

	logger := logging.NewNoopLogger(t)
	scope := metricstest.NewLoggingScope(t, logger, "atlantis")
	globalCfg := valid.NewGlobalCfgFromArgs(valid.GlobalCfgArgs{AllowAllRepoSettings: true})
	globalCfg.Repos[0].AutoDiscover = &valid.AutoDiscover{
		Mode:        valid.AutoDiscoverEnabledMode,
		IgnorePaths: []string{"environments/prod/**"},
	}
	terraformClient := tfclientmocks.NewMockClient()
	userConfig := defaultUserConfig
	builder := events.NewProjectCommandBuilder(
		false,
		&config.ParserValidator{},
		&events.DefaultProjectFinder{},
		vcsClient,
		workingDir,
		events.NewDefaultWorkingDirLocker(),
		globalCfg,
		&events.DefaultPendingPlanFinder{},
		&events.CommentParser{ExecutableName: "atlantis"},
		userConfig.SkipCloneNoChanges,
		userConfig.EnableRegExpCmd,
		userConfig.EnableAutoMerge,
		userConfig.EnableParallelPlan,
		userConfig.EnableParallelApply,
		userConfig.AutoDetectModuleFiles,
		userConfig.AutoplanFileList,
		userConfig.RestrictFileList,
		userConfig.SilenceNoProjects,
		userConfig.IncludeGitUntrackedFiles,
		userConfig.AutoDiscoverMode,
		scope,
		terraformClient,
	)
	baseRepo := models.Repo{Owner: "owner", Name: "repo", FullName: "owner/repo", VCSHost: models.VCSHost{Type: models.AzureDevops}}
	ctx := &command.Context{
		Log:      logger,
		Scope:    scope,
		HeadRepo: baseRepo,
		Pull: models.PullRequest{
			Num:        1,
			BaseBranch: "main",
			HeadBranch: "feature",
			HeadCommit: "abc123",
			BaseRepo:   baseRepo,
		},
	}
	cmd := &events.CommentCommand{Name: command.Apply, RepoRelDir: "environments/prod", Workspace: events.DefaultWorkspace}

	Assert(t, builder.ShouldIgnoreTargetedDir(ctx, cmd), "expected global ignored-target skip when VCS file download is unsupported")
	workingDir.VerifyWasCalledOnce().GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Eq(events.DefaultWorkspace))
}

func TestDefaultProjectCommandBuilder_ShouldIgnoreTargetedDirFileDownloadUnsupportedPreservesExplicitLocalProject(t *testing.T) {
	RegisterMockTestingT(t)

	atlantisYAML := "version: 3\n" +
		"projects:\n" +
		"- name: prod-project\n" +
		"  dir: environments/prod\n"
	tmpDir := DirStructure(t, map[string]any{
		"environments": map[string]any{
			"prod": map[string]any{
				"main.tf": nil,
			},
		},
	})
	Ok(t, os.WriteFile(filepath.Join(tmpDir, valid.DefaultAtlantisFile), []byte(atlantisYAML), 0600))

	workingDir := mocks.NewMockWorkingDir()
	When(workingDir.GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Eq(events.DefaultWorkspace))).ThenReturn(tmpDir, nil)
	vcsClient := vcsmocks.NewMockClient()
	When(vcsClient.GetFileContent(Any[logging.SimpleLogging](), Any[models.Repo](), Eq("abc123"), Eq(valid.DefaultAtlantisFile))).ThenReturn(false, []byte{}, errors.New("not implemented"))
	When(vcsClient.SupportsSingleFileDownload(Any[models.Repo]())).ThenReturn(false)

	logger := logging.NewNoopLogger(t)
	scope := metricstest.NewLoggingScope(t, logger, "atlantis")
	globalCfg := valid.NewGlobalCfgFromArgs(valid.GlobalCfgArgs{AllowAllRepoSettings: true})
	globalCfg.Repos[0].AutoDiscover = &valid.AutoDiscover{
		Mode:        valid.AutoDiscoverEnabledMode,
		IgnorePaths: []string{"environments/prod/**"},
	}
	terraformClient := tfclientmocks.NewMockClient()
	userConfig := defaultUserConfig
	builder := events.NewProjectCommandBuilder(
		false,
		&config.ParserValidator{},
		&events.DefaultProjectFinder{},
		vcsClient,
		workingDir,
		events.NewDefaultWorkingDirLocker(),
		globalCfg,
		&events.DefaultPendingPlanFinder{},
		&events.CommentParser{ExecutableName: "atlantis"},
		userConfig.SkipCloneNoChanges,
		userConfig.EnableRegExpCmd,
		userConfig.EnableAutoMerge,
		userConfig.EnableParallelPlan,
		userConfig.EnableParallelApply,
		userConfig.AutoDetectModuleFiles,
		userConfig.AutoplanFileList,
		userConfig.RestrictFileList,
		userConfig.SilenceNoProjects,
		userConfig.IncludeGitUntrackedFiles,
		userConfig.AutoDiscoverMode,
		scope,
		terraformClient,
	)
	baseRepo := models.Repo{Owner: "owner", Name: "repo", FullName: "owner/repo", VCSHost: models.VCSHost{Type: models.AzureDevops}}
	ctx := &command.Context{
		Log:      logger,
		Scope:    scope,
		HeadRepo: baseRepo,
		Pull: models.PullRequest{
			Num:        1,
			BaseBranch: "main",
			HeadBranch: "feature",
			HeadCommit: "abc123",
			BaseRepo:   baseRepo,
		},
	}
	cmd := &events.CommentCommand{Name: command.Apply, RepoRelDir: "environments/prod", Workspace: events.DefaultWorkspace}

	Assert(t, !builder.ShouldIgnoreTargetedDir(ctx, cmd), "expected local explicit project to prevent ignored-target skip when file download is unsupported")
	workingDir.VerifyWasCalledOnce().GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Eq(events.DefaultWorkspace))
}

func TestDefaultProjectCommandBuilder_ShouldIgnoreTargetedDirFileDownloadUnsupportedFailsOpenForPlan(t *testing.T) {
	RegisterMockTestingT(t)

	workingDir := mocks.NewMockWorkingDir()
	vcsClient := vcsmocks.NewMockClient()
	When(vcsClient.GetFileContent(Any[logging.SimpleLogging](), Any[models.Repo](), Eq("abc123"), Eq(valid.DefaultAtlantisFile))).ThenReturn(false, []byte{}, errors.New("not implemented"))
	When(vcsClient.SupportsSingleFileDownload(Any[models.Repo]())).ThenReturn(false)

	logger := logging.NewNoopLogger(t)
	scope := metricstest.NewLoggingScope(t, logger, "atlantis")
	globalCfg := valid.NewGlobalCfgFromArgs(valid.GlobalCfgArgs{AllowAllRepoSettings: true})
	globalCfg.Repos[0].AutoDiscover = &valid.AutoDiscover{
		Mode:        valid.AutoDiscoverEnabledMode,
		IgnorePaths: []string{"environments/prod/**"},
	}
	terraformClient := tfclientmocks.NewMockClient()
	userConfig := defaultUserConfig
	builder := events.NewProjectCommandBuilder(
		false,
		&config.ParserValidator{},
		&events.DefaultProjectFinder{},
		vcsClient,
		workingDir,
		events.NewDefaultWorkingDirLocker(),
		globalCfg,
		&events.DefaultPendingPlanFinder{},
		&events.CommentParser{ExecutableName: "atlantis"},
		userConfig.SkipCloneNoChanges,
		userConfig.EnableRegExpCmd,
		userConfig.EnableAutoMerge,
		userConfig.EnableParallelPlan,
		userConfig.EnableParallelApply,
		userConfig.AutoDetectModuleFiles,
		userConfig.AutoplanFileList,
		userConfig.RestrictFileList,
		userConfig.SilenceNoProjects,
		userConfig.IncludeGitUntrackedFiles,
		userConfig.AutoDiscoverMode,
		scope,
		terraformClient,
	)
	baseRepo := models.Repo{Owner: "owner", Name: "repo", FullName: "owner/repo", VCSHost: models.VCSHost{Type: models.AzureDevops}}
	ctx := &command.Context{
		Log:      logger,
		Scope:    scope,
		HeadRepo: baseRepo,
		Pull: models.PullRequest{
			Num:        1,
			BaseBranch: "main",
			HeadBranch: "feature",
			HeadCommit: "abc123",
			BaseRepo:   baseRepo,
		},
	}
	cmd := &events.CommentCommand{Name: command.Plan, RepoRelDir: "environments/prod", Workspace: events.DefaultWorkspace}

	Assert(t, !builder.ShouldIgnoreTargetedDir(ctx, cmd), "expected plan to avoid pre-clone skip when file download is unsupported")
	workingDir.VerifyWasCalled(Never()).GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Any[string]())
}

func TestDefaultProjectCommandBuilder_ShouldIgnoreTargetedDirMergeCheckoutWithLocalConfigFailsOpen(t *testing.T) {
	RegisterMockTestingT(t)

	staleLocalAtlantisYAML := "version: 3\n" +
		"autodiscover:\n" +
		"  mode: enabled\n" +
		"  ignore_paths:\n" +
		"  - \"environments/prod/**\"\n"

	logger := logging.NewNoopLogger(t)
	scope := metricstest.NewLoggingScope(t, logger, "atlantis")
	globalCfg := valid.NewGlobalCfgFromArgs(valid.GlobalCfgArgs{AllowAllRepoSettings: true})
	terraformClient := tfclientmocks.NewMockClient()
	vcsClient := vcsmocks.NewMockClient()

	dataDir := t.TempDir()
	workingDir := &events.FileWorkspace{
		DataDir:       dataDir,
		CheckoutMerge: true,
	}
	baseRepo := models.Repo{Owner: "owner", Name: "repo", FullName: "owner/repo", VCSHost: models.VCSHost{Type: models.Github}}
	pull := models.PullRequest{
		Num:        1,
		BaseBranch: "main",
		HeadBranch: "feature",
		BaseRepo:   baseRepo,
	}
	repoDir := filepath.Join(dataDir, "repos", baseRepo.FullName, "1", events.DefaultWorkspace)
	Ok(t, os.MkdirAll(repoDir, 0700))
	Ok(t, os.WriteFile(filepath.Join(repoDir, valid.DefaultAtlantisFile), []byte(staleLocalAtlantisYAML), 0600))

	userConfig := defaultUserConfig
	builder := events.NewProjectCommandBuilder(
		false,
		&config.ParserValidator{},
		&events.DefaultProjectFinder{},
		vcsClient,
		workingDir,
		events.NewDefaultWorkingDirLocker(),
		globalCfg,
		&events.DefaultPendingPlanFinder{},
		&events.CommentParser{ExecutableName: "atlantis"},
		userConfig.SkipCloneNoChanges,
		userConfig.EnableRegExpCmd,
		userConfig.EnableAutoMerge,
		userConfig.EnableParallelPlan,
		userConfig.EnableParallelApply,
		userConfig.AutoDetectModuleFiles,
		userConfig.AutoplanFileList,
		userConfig.RestrictFileList,
		userConfig.SilenceNoProjects,
		userConfig.IncludeGitUntrackedFiles,
		userConfig.AutoDiscoverMode,
		scope,
		terraformClient,
	)
	ctx := &command.Context{
		Log:      logger,
		Scope:    scope,
		HeadRepo: baseRepo,
		Pull:     pull,
	}
	cmd := &events.CommentCommand{Name: command.Plan, RepoRelDir: "environments/prod", Workspace: events.DefaultWorkspace}

	Assert(t, !builder.ShouldIgnoreTargetedDir(ctx, cmd), "expected merge checkout to avoid pre-clone skip from stale local config")
	vcsClient.VerifyWasCalled(Never()).GetFileContent(Any[logging.SimpleLogging](), Any[models.Repo](), Any[string](), Any[string]())
}

func TestDefaultProjectCommandBuilder_ShouldIgnoreTargetedDirMergeCheckoutWithoutLocalConfigFailsOpen(t *testing.T) {
	RegisterMockTestingT(t)

	remoteAtlantisYAML := "version: 3\n" +
		"autodiscover:\n" +
		"  mode: enabled\n" +
		"  ignore_paths:\n" +
		"  - \"environments/prod/**\"\n"

	logger := logging.NewNoopLogger(t)
	scope := metricstest.NewLoggingScope(t, logger, "atlantis")
	globalCfg := valid.NewGlobalCfgFromArgs(valid.GlobalCfgArgs{AllowAllRepoSettings: true})
	terraformClient := tfclientmocks.NewMockClient()
	vcsClient := vcsmocks.NewMockClient()
	When(vcsClient.GetFileContent(Any[logging.SimpleLogging](), Any[models.Repo](), Eq("feature"), Eq(valid.DefaultAtlantisFile))).ThenReturn(true, []byte(remoteAtlantisYAML), nil)

	workingDir := &events.FileWorkspace{
		DataDir:       t.TempDir(),
		CheckoutMerge: true,
	}
	userConfig := defaultUserConfig
	builder := events.NewProjectCommandBuilder(
		false,
		&config.ParserValidator{},
		&events.DefaultProjectFinder{},
		vcsClient,
		workingDir,
		events.NewDefaultWorkingDirLocker(),
		globalCfg,
		&events.DefaultPendingPlanFinder{},
		&events.CommentParser{ExecutableName: "atlantis"},
		userConfig.SkipCloneNoChanges,
		userConfig.EnableRegExpCmd,
		userConfig.EnableAutoMerge,
		userConfig.EnableParallelPlan,
		userConfig.EnableParallelApply,
		userConfig.AutoDetectModuleFiles,
		userConfig.AutoplanFileList,
		userConfig.RestrictFileList,
		userConfig.SilenceNoProjects,
		userConfig.IncludeGitUntrackedFiles,
		userConfig.AutoDiscoverMode,
		scope,
		terraformClient,
	)
	baseRepo := models.Repo{Owner: "owner", Name: "repo", FullName: "owner/repo", VCSHost: models.VCSHost{Type: models.Github}}
	ctx := &command.Context{
		Log:      logger,
		Scope:    scope,
		HeadRepo: baseRepo,
		Pull: models.PullRequest{
			Num:        1,
			BaseBranch: "main",
			HeadBranch: "feature",
			BaseRepo:   baseRepo,
		},
	}
	cmd := &events.CommentCommand{Name: command.Plan, RepoRelDir: "environments/prod", Workspace: events.DefaultWorkspace}

	Assert(t, !builder.ShouldIgnoreTargetedDir(ctx, cmd), "expected merge checkout without local config to avoid pre-clone skip")
	vcsClient.VerifyWasCalled(Never()).GetFileContent(Any[logging.SimpleLogging](), Any[models.Repo](), Any[string](), Any[string]())
}

// Test that autodiscover.ignore_paths set in repo-level atlantis.yaml blocks
// targeted plan/apply -d commands for non-configured projects.
func TestDefaultProjectCommandBuilder_BuildTargetedCommand_IgnorePathsRepoCfg(t *testing.T) {
	RegisterMockTestingT(t)

	atlantisYAML := `
version: 3
autodiscover:
  mode: enabled
  ignore_paths:
  - "environments/prod/**"
`
	tmpDir := DirStructure(t, map[string]any{
		"environments": map[string]any{
			"prod": map[string]any{
				"main.tf": nil,
			},
			"nonprod": map[string]any{
				"main.tf": nil,
			},
		},
	})
	err := os.WriteFile(filepath.Join(tmpDir, valid.DefaultAtlantisFile), []byte(atlantisYAML), 0600)
	Ok(t, err)

	workingDir := mocks.NewMockWorkingDir()
	When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](),
		Any[string]())).ThenReturn(tmpDir, nil)
	When(workingDir.GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(tmpDir, nil)
	vcsClient := vcsmocks.NewMockClient()
	When(vcsClient.GetModifiedFiles(Any[logging.SimpleLogging](), Any[models.Repo](),
		Any[models.PullRequest]())).ThenReturn([]string{"environments/prod/main.tf", "environments/nonprod/main.tf"}, nil)

	logger := logging.NewNoopLogger(t)
	scope := metricstest.NewLoggingScope(t, logger, "atlantis")

	globalCfgArgs := valid.GlobalCfgArgs{AllowAllRepoSettings: true}
	globalCfg := valid.NewGlobalCfgFromArgs(globalCfgArgs)

	terraformClient := tfclientmocks.NewMockClient()
	userConfig := defaultUserConfig

	builder := events.NewProjectCommandBuilder(
		false,
		&config.ParserValidator{},
		&events.DefaultProjectFinder{},
		vcsClient,
		workingDir,
		events.NewDefaultWorkingDirLocker(),
		globalCfg,
		&events.DefaultPendingPlanFinder{},
		&events.CommentParser{ExecutableName: "atlantis"},
		userConfig.SkipCloneNoChanges,
		userConfig.EnableRegExpCmd,
		userConfig.EnableAutoMerge,
		userConfig.EnableParallelPlan,
		userConfig.EnableParallelApply,
		userConfig.AutoDetectModuleFiles,
		userConfig.AutoplanFileList,
		userConfig.RestrictFileList,
		userConfig.SilenceNoProjects,
		userConfig.IncludeGitUntrackedFiles,
		userConfig.AutoDiscoverMode,
		scope,
		terraformClient,
	)

	cmdCtx := &command.Context{Log: logger, Scope: scope}

	// Targeted plan -d to ignored path should return no projects
	planCtxs, err := builder.BuildPlanCommands(cmdCtx, &events.CommentCommand{
		Name:       command.Plan,
		RepoRelDir: "environments/prod",
		Workspace:  "default",
	})
	Assert(t, errors.Is(err, events.ErrIgnoredTargetedDir), "expected ignored targeted dir error, got %v", err)
	Equals(t, 0, len(planCtxs))

	// Non-ignored path should work
	planCtxs, err = builder.BuildPlanCommands(cmdCtx, &events.CommentCommand{
		Name:       command.Plan,
		RepoRelDir: "environments/nonprod",
		Workspace:  "default",
	})
	Ok(t, err)
	Equals(t, 1, len(planCtxs))
	Equals(t, "environments/nonprod", planCtxs[0].RepoRelDir)

	// Targeted apply -d to ignored path should return no projects
	applyCtxs, err := builder.BuildApplyCommands(cmdCtx, &events.CommentCommand{
		Name:       command.Apply,
		RepoRelDir: "environments/prod",
		Workspace:  "default",
	})
	Assert(t, errors.Is(err, events.ErrIgnoredTargetedDir), "expected ignored targeted dir error, got %v", err)
	Equals(t, 0, len(applyCtxs))

	// Targeted apply -d to non-ignored path should work
	applyCtxs, err = builder.BuildApplyCommands(cmdCtx, &events.CommentCommand{
		Name:       command.Apply,
		RepoRelDir: "environments/nonprod",
		Workspace:  "default",
	})
	Ok(t, err)
	Equals(t, 1, len(applyCtxs))
	Equals(t, "environments/nonprod", applyCtxs[0].RepoRelDir)
}

// Test that targeted -d commands to a path with an explicit project config
// are NOT blocked by ignore_paths.
func TestDefaultProjectCommandBuilder_BuildTargetedCommand_IgnorePathsExplicitProjectAllowed(t *testing.T) {
	RegisterMockTestingT(t)

	atlantisYAML := `
version: 3
autodiscover:
  mode: enabled
  ignore_paths:
  - "environments/prod/**"
projects:
- name: prod-project
  dir: environments/prod
`
	tmpDir := DirStructure(t, map[string]any{
		"environments": map[string]any{
			"prod": map[string]any{
				"main.tf": nil,
			},
		},
	})
	err := os.WriteFile(filepath.Join(tmpDir, valid.DefaultAtlantisFile), []byte(atlantisYAML), 0600)
	Ok(t, err)

	workingDir := mocks.NewMockWorkingDir()
	When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](),
		Any[string]())).ThenReturn(tmpDir, nil)
	When(workingDir.GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(tmpDir, nil)
	vcsClient := vcsmocks.NewMockClient()
	When(vcsClient.GetModifiedFiles(Any[logging.SimpleLogging](), Any[models.Repo](),
		Any[models.PullRequest]())).ThenReturn([]string{"environments/prod/main.tf"}, nil)

	logger := logging.NewNoopLogger(t)
	scope := metricstest.NewLoggingScope(t, logger, "atlantis")

	globalCfgArgs := valid.GlobalCfgArgs{AllowAllRepoSettings: true}
	globalCfg := valid.NewGlobalCfgFromArgs(globalCfgArgs)

	terraformClient := tfclientmocks.NewMockClient()
	userConfig := defaultUserConfig

	builder := events.NewProjectCommandBuilder(
		false,
		&config.ParserValidator{},
		&events.DefaultProjectFinder{},
		vcsClient,
		workingDir,
		events.NewDefaultWorkingDirLocker(),
		globalCfg,
		&events.DefaultPendingPlanFinder{},
		&events.CommentParser{ExecutableName: "atlantis"},
		userConfig.SkipCloneNoChanges,
		userConfig.EnableRegExpCmd,
		userConfig.EnableAutoMerge,
		userConfig.EnableParallelPlan,
		userConfig.EnableParallelApply,
		userConfig.AutoDetectModuleFiles,
		userConfig.AutoplanFileList,
		userConfig.RestrictFileList,
		userConfig.SilenceNoProjects,
		userConfig.IncludeGitUntrackedFiles,
		userConfig.AutoDiscoverMode,
		scope,
		terraformClient,
	)

	cmdCtx := &command.Context{Log: logger, Scope: scope}

	// Targeted plan -d to ignored path with explicit config should succeed
	planCtxs, err := builder.BuildPlanCommands(cmdCtx, &events.CommentCommand{
		Name:       command.Plan,
		RepoRelDir: "environments/prod",
		Workspace:  "default",
	})
	Ok(t, err)
	Equals(t, 1, len(planCtxs))
	Equals(t, "prod-project", planCtxs[0].ProjectName)

	// Targeted apply -d to ignored path with explicit config should succeed
	applyCtxs, err := builder.BuildApplyCommands(cmdCtx, &events.CommentCommand{
		Name:       command.Apply,
		RepoRelDir: "environments/prod",
		Workspace:  "default",
	})
	Ok(t, err)
	Equals(t, 1, len(applyCtxs))
	Equals(t, "prod-project", applyCtxs[0].ProjectName)
}

// Test building a plan and apply command for one project
// with the RestrictFileList
func TestDefaultProjectCommandBuilder_BuildSinglePlanApplyCommand_WithRestrictFileList(t *testing.T) {
	cases := []struct {
		Description         string
		AtlantisYAML        string
		DirectoryStructure  map[string]any
		ModifiedFiles       []string
		Cmd                 events.CommentCommand
		EnableRegExpCmd     bool
		ExpErr              string
		ExpNoProjects       bool
		ExpSkipFileList     bool
		SkipPRModifiedFiles bool
		ExpProjectNames     []string
		ExpCloneWorkspaces  []string
	}{
		{
			Description: "planning a file outside of the changed files",
			Cmd: events.CommentCommand{
				Name:       command.Plan,
				RepoRelDir: "directory-1",
				Workspace:  "default",
			},
			DirectoryStructure: map[string]any{
				"directory-1": map[string]any{
					"main.tf": nil,
				},
				"directory-2": map[string]any{
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
			DirectoryStructure: map[string]any{
				"directory-1": map[string]any{
					"main.tf": nil,
				},
				"directory-2": map[string]any{
					"main.tf": nil,
				},
			},
			ModifiedFiles: []string{"directory-1/main.tf"},
		},
		{
			Description: "planning an ignored targeted dir outside of the changed files",
			Cmd: events.CommentCommand{
				Name:       command.Plan,
				RepoRelDir: "ignored",
				Workspace:  "default",
			},
			AtlantisYAML: `
version: 3
autodiscover:
  mode: enabled
  ignore_paths:
  - ignored/**
`,
			DirectoryStructure: map[string]any{
				"ignored": map[string]any{
					"main.tf": nil,
				},
				"directory-2": map[string]any{
					"main.tf": nil,
				},
			},
			ModifiedFiles:   []string{"directory-2/main.tf"},
			ExpNoProjects:   true,
			ExpSkipFileList: true,
		},
		{
			Description: "API drift targeted path skips pull request modified file filtering",
			Cmd: events.CommentCommand{
				Name:       command.Plan,
				RepoRelDir: "directory-1",
				Workspace:  "default",
			},
			DirectoryStructure: map[string]any{
				"directory-1": map[string]any{
					"main.tf": nil,
				},
				"directory-2": map[string]any{
					"main.tf": nil,
				},
			},
			ModifiedFiles:       []string{"directory-2/main.tf"},
			ExpSkipFileList:     true,
			SkipPRModifiedFiles: true,
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
			DirectoryStructure: map[string]any{
				"directory-1": map[string]any{
					"main.tf": nil,
				},
				"directory-2": map[string]any{
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
			DirectoryStructure: map[string]any{
				"directory-1": map[string]any{
					"main.tf": nil,
				},
				"directory-2": map[string]any{
					"main.tf": nil,
				},
			},
			ModifiedFiles: []string{"directory-1/main.tf"},
		},
		{
			Description: "planning a regexp project only includes changed matching projects",
			Cmd: events.CommentCommand{
				Name:        command.Plan,
				Workspace:   "default",
				ProjectName: ".*/shared",
			},
			EnableRegExpCmd: true,
			AtlantisYAML: `
version: 3
projects:
- name: athens/shared
  dir: layers/athens/shared
- name: cicd/shared
  dir: layers/cicd/shared
`,
			DirectoryStructure: map[string]any{
				"layers": map[string]any{
					"athens": map[string]any{
						"shared": map[string]any{
							"main.tf": nil,
						},
					},
					"cicd": map[string]any{
						"shared": map[string]any{
							"main.tf": nil,
						},
					},
				},
			},
			ModifiedFiles:   []string{"layers/cicd/shared/main.tf"},
			ExpProjectNames: []string{"cicd/shared"},
		},
		{
			Description: "planning a regexp project with non-default workspace clones the workspace before returning",
			Cmd: events.CommentCommand{
				Name:        command.Plan,
				Workspace:   "staging",
				ProjectName: ".*/shared",
			},
			EnableRegExpCmd: true,
			AtlantisYAML: `
version: 3
projects:
- name: athens/shared
  dir: layers/athens/shared
  workspace: staging
- name: cicd/shared
  dir: layers/cicd/shared
  workspace: staging
`,
			DirectoryStructure: map[string]any{
				"layers": map[string]any{
					"athens": map[string]any{
						"shared": map[string]any{
							"main.tf": nil,
						},
					},
					"cicd": map[string]any{
						"shared": map[string]any{
							"main.tf": nil,
						},
					},
				},
			},
			ModifiedFiles:      []string{"layers/cicd/shared/main.tf"},
			ExpProjectNames:    []string{"cicd/shared"},
			ExpCloneWorkspaces: []string{"default", "staging"},
		},
		{
			Description: "planning a regexp project outside of the changed files",
			Cmd: events.CommentCommand{
				Name:        command.Plan,
				Workspace:   "default",
				ProjectName: "athens/.*",
			},
			EnableRegExpCmd: true,
			AtlantisYAML: `
version: 3
projects:
- name: athens/shared
  dir: layers/athens/shared
- name: cicd/shared
  dir: layers/cicd/shared
`,
			DirectoryStructure: map[string]any{
				"layers": map[string]any{
					"athens": map[string]any{
						"shared": map[string]any{
							"main.tf": nil,
						},
					},
					"cicd": map[string]any{
						"shared": map[string]any{
							"main.tf": nil,
						},
					},
				},
			},
			ModifiedFiles: []string{"layers/cicd/shared/main.tf"},
			ExpErr:        "the following directories are present in the pull request but not in the requested project:\nlayers/cicd/shared",
		},
	}

	logger := logging.NewNoopLogger(t)
	scope := metricstest.NewLoggingScope(t, logger, "atlantis")
	userConfig := defaultUserConfig
	userConfig.RestrictFileList = true

	for _, c := range cases {
		t.Run(c.Description+"_"+command.Plan.String(), func(t *testing.T) {
			RegisterMockTestingT(t)
			tmpDir := DirStructure(t, c.DirectoryStructure)

			workingDir := mocks.NewMockWorkingDir()
			When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](),
				Any[string]())).ThenReturn(tmpDir, nil)
			When(workingDir.GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(tmpDir, nil)
			vcsClient := vcsmocks.NewMockClient()
			When(vcsClient.GetModifiedFiles(Any[logging.SimpleLogging](), Any[models.Repo](),
				Any[models.PullRequest]())).ThenReturn(c.ModifiedFiles, nil)
			if c.AtlantisYAML != "" {
				err := os.WriteFile(filepath.Join(tmpDir, valid.DefaultAtlantisFile), []byte(c.AtlantisYAML), 0600)
				Ok(t, err)
			}

			globalCfgArgs := valid.GlobalCfgArgs{
				AllowAllRepoSettings: true,
			}

			terraformClient := tfclientmocks.NewMockClient()

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
				userConfig.SkipCloneNoChanges,
				c.EnableRegExpCmd,
				userConfig.EnableAutoMerge,
				userConfig.EnableParallelPlan,
				userConfig.EnableParallelApply,
				userConfig.AutoDetectModuleFiles,
				userConfig.AutoplanFileList,
				userConfig.RestrictFileList,
				userConfig.SilenceNoProjects,
				userConfig.IncludeGitUntrackedFiles,
				userConfig.AutoDiscoverMode,
				scope,
				terraformClient,
			)

			var actCtxs []command.ProjectContext
			var err error
			cmd := c.Cmd
			actCtxs, err = builder.BuildPlanCommands(&command.Context{
				Log:                 logger,
				Scope:               scope,
				SkipPRModifiedFiles: c.SkipPRModifiedFiles,
			}, &cmd)

			if c.ExpNoProjects {
				Assert(t, errors.Is(err, events.ErrIgnoredTargetedDir), "expected ignored targeted dir error, got %v", err)
				Equals(t, 0, len(actCtxs))
				if c.ExpSkipFileList {
					vcsClient.VerifyWasCalled(Never()).GetModifiedFiles(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest]())
				}
				return
			}
			if c.ExpErr != "" {
				ErrEquals(t, c.ExpErr, err)
				return
			}
			Ok(t, err)
			if len(c.ExpCloneWorkspaces) > 0 {
				_, _, _, cloneWorkspaces := workingDir.VerifyWasCalled(Times(len(c.ExpCloneWorkspaces))).Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[string]()).GetAllCapturedArguments()
				Equals(t, c.ExpCloneWorkspaces, cloneWorkspaces)
			}
			if c.ExpSkipFileList {
				vcsClient.VerifyWasCalled(Never()).GetModifiedFiles(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest]())
			}
			if len(c.ExpProjectNames) == 0 {
				Equals(t, 1, len(actCtxs))
				return
			}
			var actProjectNames []string
			for _, actCtx := range actCtxs {
				actProjectNames = append(actProjectNames, actCtx.ProjectName)
			}
			sort.Strings(actProjectNames)
			Equals(t, c.ExpProjectNames, actProjectNames)
		})
	}
}

func TestDefaultProjectCommandBuilder_BuildPlanCommand_RegExpProjectWithoutRestrictFileList(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	scope := metricstest.NewLoggingScope(t, logger, "atlantis")
	tmpDir := DirStructure(t, map[string]any{
		"layers": map[string]any{
			"athens": map[string]any{
				"shared": map[string]any{
					"main.tf": nil,
				},
			},
			"cicd": map[string]any{
				"shared": map[string]any{
					"main.tf": nil,
				},
			},
		},
	})
	err := os.WriteFile(filepath.Join(tmpDir, valid.DefaultAtlantisFile), []byte(`
version: 3
projects:
- name: athens/shared
  dir: layers/athens/shared
- name: cicd/shared
  dir: layers/cicd/shared
`), 0600)
	Ok(t, err)

	workingDir := mocks.NewMockWorkingDir()
	When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](),
		Any[string]())).ThenReturn(tmpDir, nil)
	When(workingDir.GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(tmpDir, nil)

	builder := events.NewProjectCommandBuilder(
		false,
		&config.ParserValidator{},
		&events.DefaultProjectFinder{},
		vcsmocks.NewMockClient(),
		workingDir,
		events.NewDefaultWorkingDirLocker(),
		valid.NewGlobalCfgFromArgs(valid.GlobalCfgArgs{AllowAllRepoSettings: true}),
		&events.DefaultPendingPlanFinder{},
		&events.CommentParser{ExecutableName: "atlantis"},
		defaultUserConfig.SkipCloneNoChanges,
		true,
		defaultUserConfig.EnableAutoMerge,
		defaultUserConfig.EnableParallelPlan,
		defaultUserConfig.EnableParallelApply,
		defaultUserConfig.AutoDetectModuleFiles,
		defaultUserConfig.AutoplanFileList,
		false,
		defaultUserConfig.SilenceNoProjects,
		defaultUserConfig.IncludeGitUntrackedFiles,
		defaultUserConfig.AutoDiscoverMode,
		scope,
		tfclientmocks.NewMockClient(),
	)

	actCtxs, err := builder.BuildPlanCommands(&command.Context{
		Log:   logger,
		Scope: scope,
	}, &events.CommentCommand{
		Name:        command.Plan,
		Workspace:   "default",
		ProjectName: ".*/shared",
	})

	Ok(t, err)
	var actProjectNames []string
	for _, actCtx := range actCtxs {
		actProjectNames = append(actProjectNames, actCtx.ProjectName)
	}
	sort.Strings(actProjectNames)
	Equals(t, []string{"athens/shared", "cicd/shared"}, actProjectNames)
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
		AutoDiscover     valid.AutoDiscover
		ExpParallelPlan  bool
		ExpParallelApply bool
	}
	cases := map[string]struct {
		AutoMergeUserCfg            bool
		ParallelPlanEnabledUserCfg  bool
		ParallelApplyEnabledUserCfg bool
		DirStructure                map[string]any
		AtlantisYAML                string
		ModifiedFiles               []string
		Exp                         []expCtxFields
	}{
		"no atlantis.yaml": {
			DirStructure: map[string]any{
				"project1": map[string]any{
					"main.tf": nil,
				},
				"project2": map[string]any{
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
			DirStructure: map[string]any{
				"project1": map[string]any{
					"main.tf": nil,
				},
				"project2": map[string]any{
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
			DirStructure: map[string]any{
				"project1": map[string]any{
					"main.tf": nil,
				},
				"project2": map[string]any{
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
			DirStructure: map[string]any{
				"project1": map[string]any{
					"main.tf": nil,
				},
				"project2": map[string]any{
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
			DirStructure: map[string]any{
				"main.tf": nil,
			},
			ModifiedFiles: []string{},
			Exp:           []expCtxFields{},
		},
		"follow when_modified config": {
			DirStructure: map[string]any{
				"project1": map[string]any{
					"main.tf": nil,
				},
				"project2": map[string]any{
					"main.tf": nil,
				},
				"project3": map[string]any{
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
		"follow autodiscover enabled config": {
			DirStructure: map[string]any{
				"project1": map[string]any{
					"main.tf": nil,
				},
				"project2": map[string]any{
					"main.tf": nil,
				},
				"project3": map[string]any{
					"main.tf": nil,
				},
			},
			AtlantisYAML: `version: 3
autodiscover:
  mode: enabled
projects:
- name: project1-custom-name
  dir: project1`,
			ModifiedFiles: []string{"project1/main.tf", "project2/main.tf"},
			// project2 is autodiscovered, whereas project1 is not
			Exp: []expCtxFields{
				{
					ProjectName: "project1-custom-name",
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
		"autodiscover enabled but ignoring explicit project": {
			DirStructure: map[string]any{
				"project1": map[string]any{
					"main.tf": nil,
				},
				"project2": map[string]any{
					"main.tf": nil,
				},
				"project3": map[string]any{
					"main.tf": nil,
				},
			},
			AtlantisYAML: `version: 3
autodiscover:
  mode: enabled
  ignore_paths:
  - project1
projects:
- name: project1-custom-name
  dir: project1`,
			ModifiedFiles: []string{"project1/main.tf", "project2/main.tf"},
			// project2 is autodiscover-ignored, but configured explicitly so added
			// project1 is autodiscoverd as normal
			Exp: []expCtxFields{
				{
					ProjectName: "project1-custom-name",
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
		"autodiscover enabled but project excluded by empty when_modified": {
			DirStructure: map[string]any{
				"project1": map[string]any{
					"main.tf": nil,
				},
				"project2": map[string]any{
					"main.tf": nil,
				},
				"project3": map[string]any{
					"main.tf": nil,
				},
			},
			AtlantisYAML: `version: 3
autodiscover:
  mode: enabled
projects:
- dir: project1
  autoplan:
    when_modified: []`,
			ModifiedFiles: []string{"project1/main.tf", "project2/main.tf"},
			Exp: []expCtxFields{
				{
					ProjectName: "",
					RepoRelDir:  "project2",
					Workspace:   "default",
				},
			},
		},
	}

	logger := logging.NewNoopLogger(t)
	scope := metricstest.NewLoggingScope(t, logger, "atlantis")
	userConfig := defaultUserConfig

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			RegisterMockTestingT(t)
			tmpDir := DirStructure(t, c.DirStructure)

			workingDir := mocks.NewMockWorkingDir()
			When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](),
				Any[string]())).ThenReturn(tmpDir, nil)
			When(workingDir.GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(tmpDir, nil)
			vcsClient := vcsmocks.NewMockClient()
			When(vcsClient.GetModifiedFiles(Any[logging.SimpleLogging](), Any[models.Repo](),
				Any[models.PullRequest]())).ThenReturn(c.ModifiedFiles, nil)
			if c.AtlantisYAML != "" {
				err := os.WriteFile(filepath.Join(tmpDir, valid.DefaultAtlantisFile), []byte(c.AtlantisYAML), 0600)
				Ok(t, err)
			}

			globalCfgArgs := valid.GlobalCfgArgs{
				AllowAllRepoSettings: true,
			}

			terraformClient := tfclientmocks.NewMockClient()

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
				userConfig.SkipCloneNoChanges,
				userConfig.EnableRegExpCmd,
				userConfig.EnableAutoMerge,
				c.ParallelPlanEnabledUserCfg,
				c.ParallelApplyEnabledUserCfg,
				userConfig.AutoDetectModuleFiles,
				userConfig.AutoplanFileList,
				userConfig.RestrictFileList,
				userConfig.SilenceNoProjects,
				userConfig.IncludeGitUntrackedFiles,
				userConfig.AutoDiscoverMode,
				scope,
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
					Verbose:     true,
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
	tmpDir := DirStructure(t, map[string]any{
		"workspace1": map[string]any{
			"project1": map[string]any{
				"main.tf":          nil,
				"workspace.tfplan": nil,
			},
			"project2": map[string]any{
				"main.tf":          nil,
				"workspace.tfplan": nil,
			},
		},
		"workspace2": map[string]any{
			"project1": map[string]any{
				"main.tf":          nil,
				"workspace.tfplan": nil,
			},
			"project2": map[string]any{
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
	userConfig := defaultUserConfig

	globalCfgArgs := valid.GlobalCfgArgs{}
	scope := metricstest.NewLoggingScope(t, logger, "atlantis")

	terraformClient := tfclientmocks.NewMockClient()

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
		userConfig.SkipCloneNoChanges,
		userConfig.EnableRegExpCmd,
		userConfig.EnableAutoMerge,
		userConfig.EnableParallelPlan,
		userConfig.EnableParallelApply,
		userConfig.AutoDetectModuleFiles,
		userConfig.AutoplanFileList,
		userConfig.RestrictFileList,
		userConfig.SilenceNoProjects,
		userConfig.IncludeGitUntrackedFiles,
		userConfig.AutoDiscoverMode,
		scope,
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

// Test that autodiscover.ignore_paths is respected during multi-apply.
// Plans in ignored paths (e.g. .terraform/modules/) should not be applied.
func TestDefaultProjectCommandBuilder_BuildMultiApply_IgnorePaths(t *testing.T) {
	RegisterMockTestingT(t)
	tmpDir := DirStructure(t, map[string]any{
		"default": map[string]any{
			"project1": map[string]any{
				"main.tf":        nil,
				"default.tfplan": nil,
			},
			"project2": map[string]any{
				"main.tf":        nil,
				"default.tfplan": nil,
			},
			".terraform": map[string]any{
				"modules": map[string]any{
					"some_module": map[string]any{
						"test-deployment": map[string]any{
							"main.tf":        nil,
							"default.tfplan": nil,
						},
					},
				},
			},
		},
	})
	// Initialize git repo so .tfplan files get picked up as untracked.
	runCmd(t, filepath.Join(tmpDir, "default"), "git", "init")

	workingDir := mocks.NewMockWorkingDir()
	When(workingDir.GetPullDir(
		Any[models.Repo](),
		Any[models.PullRequest]())).
		ThenReturn(tmpDir, nil)
	When(workingDir.GetWorkingDir(
		Any[models.Repo](),
		Any[models.PullRequest](),
		Any[string]())).
		ThenReturn(filepath.Join(tmpDir, "default"), nil)

	logger := logging.NewNoopLogger(t)
	userConfig := defaultUserConfig

	globalCfgArgs := valid.GlobalCfgArgs{
		AllowAllRepoSettings: true,
	}
	globalCfg := valid.NewGlobalCfgFromArgs(globalCfgArgs)
	// Set ignore_paths on the global config's AutoDiscover so that
	// isAutoDiscoverPathIgnored filters .terraform paths.
	globalCfg.Repos[0].AutoDiscover = &valid.AutoDiscover{
		Mode:        valid.AutoDiscoverEnabledMode,
		IgnorePaths: []string{".terraform/**"},
	}
	scope := metricstest.NewLoggingScope(t, logger, "atlantis")

	terraformClient := tfclientmocks.NewMockClient()

	builder := events.NewProjectCommandBuilder(
		false,
		&config.ParserValidator{},
		&events.DefaultProjectFinder{},
		nil,
		workingDir,
		events.NewDefaultWorkingDirLocker(),
		globalCfg,
		&events.DefaultPendingPlanFinder{},
		&events.CommentParser{ExecutableName: "atlantis"},
		userConfig.SkipCloneNoChanges,
		userConfig.EnableRegExpCmd,
		userConfig.EnableAutoMerge,
		userConfig.EnableParallelPlan,
		userConfig.EnableParallelApply,
		userConfig.AutoDetectModuleFiles,
		userConfig.AutoplanFileList,
		userConfig.RestrictFileList,
		userConfig.SilenceNoProjects,
		userConfig.IncludeGitUntrackedFiles,
		userConfig.AutoDiscoverMode,
		scope,
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
	Equals(t, 2, len(ctxs))
	Equals(t, "project1", ctxs[0].RepoRelDir)
	Equals(t, "default", ctxs[0].Workspace)
	Equals(t, "project2", ctxs[1].RepoRelDir)
	Equals(t, "default", ctxs[1].Workspace)
}

// Test that autodiscover.ignore_paths set in repo-level atlantis.yaml (not
// global config) is respected during multi-apply.
func TestDefaultProjectCommandBuilder_BuildMultiApply_IgnorePathsRepoCfg(t *testing.T) {
	RegisterMockTestingT(t)

	atlantisYAML := `
version: 3
autodiscover:
  mode: enabled
  ignore_paths:
  - ".terraform/**"
`
	tmpDir := DirStructure(t, map[string]any{
		"default": map[string]any{
			"atlantis.yaml": atlantisYAML,
			"project1": map[string]any{
				"main.tf":        nil,
				"default.tfplan": nil,
			},
			"project2": map[string]any{
				"main.tf":        nil,
				"default.tfplan": nil,
			},
			".terraform": map[string]any{
				"modules": map[string]any{
					"some_module": map[string]any{
						"test-deployment": map[string]any{
							"main.tf":        nil,
							"default.tfplan": nil,
						},
					},
				},
			},
		},
	})
	// Initialize git repo so .tfplan files get picked up as untracked.
	runCmd(t, filepath.Join(tmpDir, "default"), "git", "init")

	workingDir := mocks.NewMockWorkingDir()
	When(workingDir.GetPullDir(
		Any[models.Repo](),
		Any[models.PullRequest]())).
		ThenReturn(tmpDir, nil)
	When(workingDir.GetWorkingDir(
		Any[models.Repo](),
		Any[models.PullRequest](),
		Any[string]())).
		ThenReturn(filepath.Join(tmpDir, "default"), nil)

	logger := logging.NewNoopLogger(t)
	userConfig := defaultUserConfig

	// No global-level ignore_paths — only the repo atlantis.yaml has it.
	globalCfgArgs := valid.GlobalCfgArgs{
		AllowAllRepoSettings: true,
	}
	globalCfg := valid.NewGlobalCfgFromArgs(globalCfgArgs)
	scope := metricstest.NewLoggingScope(t, logger, "atlantis")

	terraformClient := tfclientmocks.NewMockClient()

	builder := events.NewProjectCommandBuilder(
		false,
		&config.ParserValidator{},
		&events.DefaultProjectFinder{},
		nil,
		workingDir,
		events.NewDefaultWorkingDirLocker(),
		globalCfg,
		&events.DefaultPendingPlanFinder{},
		&events.CommentParser{ExecutableName: "atlantis"},
		userConfig.SkipCloneNoChanges,
		userConfig.EnableRegExpCmd,
		userConfig.EnableAutoMerge,
		userConfig.EnableParallelPlan,
		userConfig.EnableParallelApply,
		userConfig.AutoDetectModuleFiles,
		userConfig.AutoplanFileList,
		userConfig.RestrictFileList,
		userConfig.SilenceNoProjects,
		userConfig.IncludeGitUntrackedFiles,
		userConfig.AutoDiscoverMode,
		scope,
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
	Equals(t, 2, len(ctxs))
	Equals(t, "project1", ctxs[0].RepoRelDir)
	Equals(t, "default", ctxs[0].Workspace)
	Equals(t, "project2", ctxs[1].RepoRelDir)
	Equals(t, "default", ctxs[1].Workspace)
}

// Test that when global AutoDiscover is set (mode enabled) but has no
// IgnorePaths, isAutoDiscoverPathIgnored falls through to the repo config's
// ignore_paths. This validates the guard change from "!= nil" to
// "!= nil && IgnorePaths != nil".
func TestDefaultProjectCommandBuilder_BuildMultiApply_GlobalAutoDiscoverFallthrough(t *testing.T) {
	RegisterMockTestingT(t)

	atlantisYAML := `
version: 3
autodiscover:
  mode: enabled
  ignore_paths:
  - ".terraform/**"
`
	tmpDir := DirStructure(t, map[string]any{
		"default": map[string]any{
			"atlantis.yaml": atlantisYAML,
			"project1": map[string]any{
				"main.tf":        nil,
				"default.tfplan": nil,
			},
			".terraform": map[string]any{
				"modules": map[string]any{
					"some_module": map[string]any{
						"test-deployment": map[string]any{
							"main.tf":        nil,
							"default.tfplan": nil,
						},
					},
				},
			},
		},
	})
	runCmd(t, filepath.Join(tmpDir, "default"), "git", "init")

	workingDir := mocks.NewMockWorkingDir()
	When(workingDir.GetPullDir(
		Any[models.Repo](),
		Any[models.PullRequest]())).
		ThenReturn(tmpDir, nil)
	When(workingDir.GetWorkingDir(
		Any[models.Repo](),
		Any[models.PullRequest](),
		Any[string]())).
		ThenReturn(filepath.Join(tmpDir, "default"), nil)

	logger := logging.NewNoopLogger(t)
	userConfig := defaultUserConfig

	globalCfgArgs := valid.GlobalCfgArgs{
		AllowAllRepoSettings: true,
	}
	globalCfg := valid.NewGlobalCfgFromArgs(globalCfgArgs)
	// Global AutoDiscover with mode set but NO IgnorePaths.
	// The guard change means this falls through to repo config's ignore_paths.
	globalCfg.Repos[0].AutoDiscover = &valid.AutoDiscover{
		Mode: valid.AutoDiscoverEnabledMode,
	}
	scope := metricstest.NewLoggingScope(t, logger, "atlantis")

	terraformClient := tfclientmocks.NewMockClient()

	builder := events.NewProjectCommandBuilder(
		false,
		&config.ParserValidator{},
		&events.DefaultProjectFinder{},
		nil,
		workingDir,
		events.NewDefaultWorkingDirLocker(),
		globalCfg,
		&events.DefaultPendingPlanFinder{},
		&events.CommentParser{ExecutableName: "atlantis"},
		userConfig.SkipCloneNoChanges,
		userConfig.EnableRegExpCmd,
		userConfig.EnableAutoMerge,
		userConfig.EnableParallelPlan,
		userConfig.EnableParallelApply,
		userConfig.AutoDetectModuleFiles,
		userConfig.AutoplanFileList,
		userConfig.RestrictFileList,
		userConfig.SilenceNoProjects,
		userConfig.IncludeGitUntrackedFiles,
		userConfig.AutoDiscoverMode,
		scope,
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
	// .terraform/modules/some_module/test-deployment should be filtered by
	// the repo config's ignore_paths, even though global AutoDiscover is set.
	Equals(t, 1, len(ctxs))
	Equals(t, "project1", ctxs[0].RepoRelDir)
	Equals(t, "default", ctxs[0].Workspace)
}

func TestDefaultProjectCommandBuilder_BuildMultiApply_ExplicitPlanInIgnoredPath(t *testing.T) {
	RegisterMockTestingT(t)

	atlantisYAML := "version: 3\nautodiscover:\n  mode: auto\n  ignore_paths:\n  - \"ignored/**\"\n"
	tmpDir := DirStructure(t, map[string]any{
		"default": map[string]any{
			"atlantis.yaml": atlantisYAML,
			"ignored": map[string]any{
				"dir": map[string]any{
					"main.tf":        nil,
					"default.tfplan": nil,
				},
			},
		},
	})
	runCmd(t, filepath.Join(tmpDir, "default"), "git", "init")

	workingDir := mocks.NewMockWorkingDir()
	When(workingDir.GetPullDir(
		Any[models.Repo](),
		Any[models.PullRequest]())).
		ThenReturn(tmpDir, nil)
	When(workingDir.GetWorkingDir(
		Any[models.Repo](),
		Any[models.PullRequest](),
		Any[string]())).
		ThenReturn(filepath.Join(tmpDir, "default"), nil)

	logger := logging.NewNoopLogger(t)
	userConfig := defaultUserConfig

	globalCfgArgs := valid.GlobalCfgArgs{
		AllowAllRepoSettings: true,
	}
	globalCfg := valid.NewGlobalCfgFromArgs(globalCfgArgs)
	scope := metricstest.NewLoggingScope(t, logger, "atlantis")

	terraformClient := tfclientmocks.NewMockClient()

	builder := events.NewProjectCommandBuilder(
		false,
		&config.ParserValidator{},
		&events.DefaultProjectFinder{},
		nil,
		workingDir,
		events.NewDefaultWorkingDirLocker(),
		globalCfg,
		&events.DefaultPendingPlanFinder{},
		&events.CommentParser{ExecutableName: "atlantis"},
		userConfig.SkipCloneNoChanges,
		userConfig.EnableRegExpCmd,
		userConfig.EnableAutoMerge,
		userConfig.EnableParallelPlan,
		userConfig.EnableParallelApply,
		userConfig.AutoDetectModuleFiles,
		userConfig.AutoplanFileList,
		userConfig.RestrictFileList,
		userConfig.SilenceNoProjects,
		userConfig.IncludeGitUntrackedFiles,
		userConfig.AutoDiscoverMode,
		scope,
		terraformClient,
	)

	ctxs, err := builder.BuildApplyCommands(
		&command.Context{
			Log:   logger,
			Scope: scope,
			PullStatus: &models.PullStatus{
				Projects: []models.ProjectStatus{
					{
						Workspace:  "default",
						RepoRelDir: "ignored/dir",
						Status:     models.PlannedPlanStatus,
					},
				},
			},
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
	Equals(t, 1, len(ctxs))
	Equals(t, "ignored/dir", ctxs[0].RepoRelDir)
	Equals(t, "default", ctxs[0].Workspace)
}

func TestDefaultProjectCommandBuilder_BuildMultiApply_IgnoreStaleNamedPlanInIgnoredPath(t *testing.T) {
	RegisterMockTestingT(t)

	atlantisYAML := "version: 3\nautodiscover:\n  mode: enabled\n  ignore_paths:\n  - \"ignored/**\"\nprojects:\n- name: app\n  dir: app\n"
	tmpDir := DirStructure(t, map[string]any{
		"default": map[string]any{
			"atlantis.yaml": atlantisYAML,
			"app": map[string]any{
				"main.tf":            nil,
				"app-default.tfplan": nil,
			},
			"ignored": map[string]any{
				"stale": map[string]any{
					"main.tf":            nil,
					"app-default.tfplan": nil,
				},
			},
		},
	})
	runCmd(t, filepath.Join(tmpDir, "default"), "git", "init")

	workingDir := mocks.NewMockWorkingDir()
	When(workingDir.GetPullDir(
		Any[models.Repo](),
		Any[models.PullRequest]())).
		ThenReturn(tmpDir, nil)
	When(workingDir.GetWorkingDir(
		Any[models.Repo](),
		Any[models.PullRequest](),
		Any[string]())).
		ThenReturn(filepath.Join(tmpDir, "default"), nil)

	logger := logging.NewNoopLogger(t)
	userConfig := defaultUserConfig

	globalCfgArgs := valid.GlobalCfgArgs{
		AllowAllRepoSettings: true,
	}
	globalCfg := valid.NewGlobalCfgFromArgs(globalCfgArgs)
	scope := metricstest.NewLoggingScope(t, logger, "atlantis")

	terraformClient := tfclientmocks.NewMockClient()

	builder := events.NewProjectCommandBuilder(
		false,
		&config.ParserValidator{},
		&events.DefaultProjectFinder{},
		nil,
		workingDir,
		events.NewDefaultWorkingDirLocker(),
		globalCfg,
		&events.DefaultPendingPlanFinder{},
		&events.CommentParser{ExecutableName: "atlantis"},
		userConfig.SkipCloneNoChanges,
		userConfig.EnableRegExpCmd,
		userConfig.EnableAutoMerge,
		userConfig.EnableParallelPlan,
		userConfig.EnableParallelApply,
		userConfig.AutoDetectModuleFiles,
		userConfig.AutoplanFileList,
		userConfig.RestrictFileList,
		userConfig.SilenceNoProjects,
		userConfig.IncludeGitUntrackedFiles,
		userConfig.AutoDiscoverMode,
		scope,
		terraformClient,
	)

	ctxs, err := builder.BuildApplyCommands(
		&command.Context{
			Log:   logger,
			Scope: scope,
			PullStatus: &models.PullStatus{
				Projects: []models.ProjectStatus{
					{
						Workspace:   "default",
						RepoRelDir:  "app",
						ProjectName: "app",
						Status:      models.PlannedPlanStatus,
					},
				},
			},
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
	Equals(t, 1, len(ctxs))
	Equals(t, "app", ctxs[0].RepoRelDir)
	Equals(t, "app", ctxs[0].ProjectName)
	Equals(t, "default", ctxs[0].Workspace)
}

// Test that if a directory has a list of workspaces configured then we don't
// allow plans for other workspace names.
func TestDefaultProjectCommandBuilder_WrongWorkspaceName(t *testing.T) {
	RegisterMockTestingT(t)
	workingDir := mocks.NewMockWorkingDir()

	tmpDir := DirStructure(t, map[string]any{
		"pulldir": map[string]any{
			"notconfigured": map[string]any{},
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

	When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](),
		Any[string]())).ThenReturn(repoDir, nil)
	When(workingDir.GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(repoDir, nil)

	globalCfgArgs := valid.GlobalCfgArgs{
		AllowAllRepoSettings: true,
	}
	logger := logging.NewNoopLogger(t)
	scope := metricstest.NewLoggingScope(t, logger, "atlantis")
	userConfig := defaultUserConfig

	terraformClient := tfclientmocks.NewMockClient()

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
		userConfig.SkipCloneNoChanges,
		userConfig.EnableRegExpCmd,
		userConfig.EnableAutoMerge,
		userConfig.EnableParallelPlan,
		userConfig.EnableParallelApply,
		userConfig.AutoDetectModuleFiles,
		userConfig.AutoplanFileList,
		userConfig.RestrictFileList,
		userConfig.SilenceNoProjects,
		userConfig.IncludeGitUntrackedFiles,
		userConfig.AutoDiscoverMode,
		scope,
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
	scope := metricstest.NewLoggingScope(t, logger, "atlantis")
	userConfig := defaultUserConfig

	for _, c := range cases {
		t.Run(strings.Join(c.ExtraArgs, " "), func(t *testing.T) {
			RegisterMockTestingT(t)
			tmpDir := DirStructure(t, map[string]any{
				"main.tf": nil,
			})

			workingDir := mocks.NewMockWorkingDir()
			When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](),
				Any[string]())).ThenReturn(tmpDir, nil)
			When(workingDir.GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(tmpDir, nil)
			vcsClient := vcsmocks.NewMockClient()
			When(vcsClient.GetModifiedFiles(Any[logging.SimpleLogging](), Any[models.Repo](),
				Any[models.PullRequest]())).ThenReturn([]string{"main.tf"}, nil)

			globalCfgArgs := valid.GlobalCfgArgs{
				AllowAllRepoSettings: true,
			}

			terraformClient := tfclientmocks.NewMockClient()

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
				userConfig.SkipCloneNoChanges,
				userConfig.EnableRegExpCmd,
				userConfig.EnableAutoMerge,
				userConfig.EnableParallelPlan,
				userConfig.EnableParallelApply,
				userConfig.AutoDetectModuleFiles,
				userConfig.AutoplanFileList,
				userConfig.RestrictFileList,
				userConfig.SilenceNoProjects,
				userConfig.IncludeGitUntrackedFiles,
				userConfig.AutoDiscoverMode,
				scope,
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

	opentofuAtlantisYamlContent := `
version: 3
projects:
- dir: project1
  terraform_distribution: opentofu
`

	type testCase struct {
		DirStructure    map[string]any
		AtlantisYAML    string
		ModifiedFiles   []string
		Exp             map[string]string
		ExpDistribution map[string]string
	}

	testCases := make(map[string]testCase)

	// atlantis.yaml should take precedence over terraform config
	testCases["with project config and terraform config"] = testCase{
		DirStructure: map[string]any{
			"project1": map[string]any{
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
		DirStructure: map[string]any{
			"project1": map[string]any{
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
		DirStructure: map[string]any{
			"project1": map[string]any{
				"main.tf": nil,
			},
		},
		ModifiedFiles: []string{"project1/main.tf", "project2/main.tf"},
		Exp: map[string]string{
			"project1": "",
		},
	}

	testCases["project with different terraform config"] = testCase{
		DirStructure: map[string]any{
			"project1": map[string]any{
				"main.tf": baseVersionConfig,
			},
			"project2": map[string]any{
				"main.tf": strings.ReplaceAll(baseVersionConfig, "0.12.8", "0.12.9"),
			},
		},
		ModifiedFiles: []string{"project1/main.tf", "project2/main.tf"},
		Exp: map[string]string{
			"project1": "0.12.8",
			"project2": "0.12.9",
		},
	}

	testCases["opentofu project detects version with opentofu distribution"] = testCase{
		DirStructure: map[string]any{
			"project1": map[string]any{
				"main.tf": baseVersionConfig,
			},
			valid.DefaultAtlantisFile: opentofuAtlantisYamlContent,
		},
		ModifiedFiles: []string{"project1/main.tf"},
		Exp: map[string]string{
			"project1": "0.12.8",
		},
		ExpDistribution: map[string]string{
			"project1": "tofu",
		},
	}

	logger := logging.NewNoopLogger(t)
	scope := metricstest.NewLoggingScope(t, logger, "atlantis")
	userConfig := defaultUserConfig

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			RegisterMockTestingT(t)

			tmpDir := DirStructure(t, testCase.DirStructure)

			vcsClient := vcsmocks.NewMockClient()
			When(vcsClient.GetModifiedFiles(Any[logging.SimpleLogging](), Any[models.Repo](),
				Any[models.PullRequest]())).ThenReturn(testCase.ModifiedFiles, nil)
			workingDir := mocks.NewMockWorkingDir()
			When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](),
				Any[string]())).ThenReturn(tmpDir, nil)
			When(workingDir.GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(tmpDir, nil)

			globalCfgArgs := valid.GlobalCfgArgs{
				AllowAllRepoSettings: true,
			}

			terraformClient := tfclientmocks.NewMockClient()
			detectedDistributions := map[string]string{}
			When(terraformClient.DetectVersion(Any[logging.SimpleLogging](), Any[terraform.Distribution](), Any[string]())).Then(func(params []Param) ReturnValues {
				projectName := filepath.Base(params[2].(string))
				if params[1] != nil {
					detectedDistributions[projectName] = params[1].(terraform.Distribution).BinName()
				}
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
				userConfig.SkipCloneNoChanges,
				userConfig.EnableRegExpCmd,
				userConfig.EnableAutoMerge,
				userConfig.EnableParallelPlan,
				userConfig.EnableParallelApply,
				userConfig.AutoDetectModuleFiles,
				userConfig.AutoplanFileList,
				userConfig.RestrictFileList,
				userConfig.SilenceNoProjects,
				userConfig.IncludeGitUntrackedFiles,
				userConfig.AutoDiscoverMode,
				scope,
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
			for project, expDistribution := range testCase.ExpDistribution {
				Equals(t, expDistribution, detectedDistributions[project])
			}
		})
	}
}

// Test that we don't clone the repo if there were no changes based on the atlantis.yaml file.
func TestDefaultProjectCommandBuilder_SkipCloneNoChanges(t *testing.T) {
	cases := []struct {
		AtlantisYAML             string
		IsFork                   bool
		ExpectedCtxs             int
		ExpectedClones           int
		ExpectedGetFileContents  int
		ModifiedFiles            []string
		IncludeGitUntrackedFiles bool
	}{
		{
			AtlantisYAML: `
version: 3
projects:
- dir: dir1`,
			ExpectedCtxs:             0,
			ExpectedClones:           0,
			ExpectedGetFileContents:  1,
			ModifiedFiles:            []string{"dir2/main.tf"},
			IncludeGitUntrackedFiles: false,
		},
		{
			AtlantisYAML: `
version: 3
projects:
- dir: dir1`,
			ExpectedCtxs:             0,
			ExpectedClones:           1,
			ExpectedGetFileContents:  0,
			ModifiedFiles:            []string{"dir2/main.tf"},
			IncludeGitUntrackedFiles: true,
		},
		{
			AtlantisYAML: `
version: 3
projects:
- dir: dir1`,
			IsFork:                  true,
			ExpectedCtxs:            0,
			ExpectedClones:          0,
			ExpectedGetFileContents: 1,
			ModifiedFiles:           []string{"dir2/main.tf"},
		},
		{
			AtlantisYAML: `
version: 3
parallel_plan: true`,
			ExpectedCtxs:             0,
			ExpectedClones:           1,
			ExpectedGetFileContents:  1,
			ModifiedFiles:            []string{"README.md"},
			IncludeGitUntrackedFiles: false,
		},
		{
			AtlantisYAML: `
version: 3
autodiscover:
  mode: enabled
projects:
- dir: dir1`,
			ExpectedCtxs:             0,
			ExpectedClones:           1,
			ExpectedGetFileContents:  1,
			ModifiedFiles:            []string{"dir2/main.tf"},
			IncludeGitUntrackedFiles: false,
		},
	}

	userConfig := defaultUserConfig
	userConfig.SkipCloneNoChanges = true

	for _, c := range cases {
		RegisterMockTestingT(t)
		vcsClient := vcsmocks.NewMockClient()
		When(vcsClient.GetModifiedFiles(
			Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest]())).ThenReturn(c.ModifiedFiles, nil)
		When(vcsClient.SupportsSingleFileDownload(Any[models.Repo]())).ThenReturn(true)
		When(vcsClient.GetFileContent(
			Any[logging.SimpleLogging](), Any[models.Repo](), Any[string](), Any[string]())).ThenReturn(true, []byte(c.AtlantisYAML), nil)
		workingDir := mocks.NewMockWorkingDir()

		logger := logging.NewNoopLogger(t)

		globalCfgArgs := valid.GlobalCfgArgs{
			AllowAllRepoSettings: true,
		}
		scope := metricstest.NewLoggingScope(t, logger, "atlantis")
		terraformClient := tfclientmocks.NewMockClient()

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
			userConfig.SkipCloneNoChanges,
			userConfig.EnableRegExpCmd,
			userConfig.EnableAutoMerge,
			userConfig.EnableParallelPlan,
			userConfig.EnableParallelApply,
			userConfig.AutoDetectModuleFiles,
			userConfig.AutoplanFileList,
			userConfig.RestrictFileList,
			userConfig.SilenceNoProjects,
			c.IncludeGitUntrackedFiles,
			userConfig.AutoDiscoverMode,
			scope,
			terraformClient,
		)

		var actCtxs []command.ProjectContext
		var err error

		baseRepo := models.Repo{Owner: "owner"}
		headRepo := baseRepo
		if c.IsFork {
			headRepo.Owner = "repoForker"
		}

		actCtxs, err = builder.BuildAutoplanCommands(&command.Context{
			HeadRepo: headRepo,
			Pull: models.PullRequest{
				BaseRepo: baseRepo,
			},
			User:  models.User{},
			Log:   logger,
			Scope: scope,
			PullRequestStatus: models.PullReqStatus{
				MergeableStatus: models.MergeableStatus{IsMergeable: true},
			},
		})

		Ok(t, err)
		Equals(t, c.ExpectedCtxs, len(actCtxs))
		workingDir.VerifyWasCalled(Times(c.ExpectedClones)).Clone(Any[logging.SimpleLogging](), Any[models.Repo](),
			Any[models.PullRequest](), Any[string]())
		res := vcsClient.VerifyWasCalled(Times(c.ExpectedGetFileContents)).GetFileContent(Any[logging.SimpleLogging](), Any[models.Repo](), Any[string](), Any[string]())
		if c.ExpectedGetFileContents > 0 {
			_, actRepo, _, _ := res.GetCapturedArguments()
			Equals(t, headRepo, actRepo)
		}
	}
}

func TestDefaultProjectCommandBuilder_WithPolicyCheckEnabled_BuildAutoplanCommand(t *testing.T) {
	RegisterMockTestingT(t)
	tmpDir := DirStructure(t, map[string]any{
		"main.tf": nil,
	})

	logger := logging.NewNoopLogger(t)
	scope := metricstest.NewLoggingScope(t, logger, "atlantis")
	userConfig := defaultUserConfig

	workingDir := mocks.NewMockWorkingDir()
	When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](),
		Any[string]())).ThenReturn(tmpDir, nil)
	vcsClient := vcsmocks.NewMockClient()
	When(vcsClient.GetModifiedFiles(Any[logging.SimpleLogging](), Any[models.Repo](),
		Any[models.PullRequest]())).ThenReturn([]string{"main.tf"}, nil)

	globalCfgArgs := valid.GlobalCfgArgs{
		AllowAllRepoSettings: false,
		PolicyCheckEnabled:   true,
	}

	globalCfg := valid.NewGlobalCfgFromArgs(globalCfgArgs)
	terraformClient := tfclientmocks.NewMockClient()

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
		userConfig.SkipCloneNoChanges,
		userConfig.EnableRegExpCmd,
		userConfig.EnableAutoMerge,
		userConfig.EnableParallelPlan,
		userConfig.EnableParallelApply,
		userConfig.AutoDetectModuleFiles,
		userConfig.AutoplanFileList,
		userConfig.RestrictFileList,
		userConfig.SilenceNoProjects,
		userConfig.IncludeGitUntrackedFiles,
		userConfig.AutoDiscoverMode,
		scope,
		terraformClient,
	)

	ctxs, err := builder.BuildAutoplanCommands(&command.Context{
		PullRequestStatus: models.PullReqStatus{
			MergeableStatus: models.MergeableStatus{IsMergeable: true},
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
	tmpDir := DirStructure(t, map[string]any{
		"workspace1": map[string]any{
			"project1": map[string]any{
				"main.tf":          nil,
				"workspace.tfplan": nil,
			},
			"project2": map[string]any{
				"main.tf":          nil,
				"workspace.tfplan": nil,
			},
		},
		"workspace2": map[string]any{
			"project1": map[string]any{
				"main.tf":          nil,
				"workspace.tfplan": nil,
			},
			"project2": map[string]any{
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
	scope := metricstest.NewLoggingScope(t, logger, "atlantis")
	userConfig := defaultUserConfig

	globalCfgArgs := valid.GlobalCfgArgs{
		AllowAllRepoSettings: false,
	}
	terraformClient := tfclientmocks.NewMockClient()

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
		userConfig.SkipCloneNoChanges,
		userConfig.EnableRegExpCmd,
		userConfig.EnableAutoMerge,
		userConfig.EnableParallelPlan,
		userConfig.EnableParallelApply,
		userConfig.AutoDetectModuleFiles,
		userConfig.AutoplanFileList,
		userConfig.RestrictFileList,
		userConfig.SilenceNoProjects,
		userConfig.IncludeGitUntrackedFiles,
		userConfig.AutoDiscoverMode,
		scope,
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

// Test
func TestDefaultProjectCommandBuilder_BuildPlanCommands_Single_With_RestrictFileList_And_IncludeGitUntrackedFiles(t *testing.T) {
	testDir1 := "directory-1"
	testDir2 := "directory-2"

	cases := []struct {
		Description        string
		AtlantisYAML       string
		DirectoryStructure map[string]any
		ModifiedFiles      []string
		UntrackedFiles     []string
		Cmd                events.CommentCommand
		ExpRepoRelDir      string
		ExpErr             string
	}{
		{
			Description: "planning a git untracked file project in a modified directory",
			Cmd: events.CommentCommand{
				Name:       command.Plan,
				RepoRelDir: testDir1 + "/ci-cdktf.out/stacks/test",
				Workspace:  "default",
			},
			DirectoryStructure: map[string]any{
				testDir1: map[string]any{
					"main.ts": nil,
				},
			},
			ModifiedFiles:  []string{testDir1 + "/main.ts"},
			UntrackedFiles: []string{testDir1 + "/ci-cdktf.out/stacks/test/cdk.tf.json"},
			ExpRepoRelDir:  testDir1 + "/ci-cdktf.out/stacks/test",
		},
		{
			Description: "planning a git untracked file project outside a modified directory",
			Cmd: events.CommentCommand{
				Name:       command.Plan,
				RepoRelDir: testDir2 + "/ci-cdktf.out/stacks/test",
				Workspace:  "default",
			},
			DirectoryStructure: map[string]any{
				testDir1: map[string]any{
					"main.ts": nil,
				},
			},
			ModifiedFiles:  []string{testDir1 + "/main.ts"},
			UntrackedFiles: []string{testDir1 + "/ci-cdktf.out/stacks/test/cdk.tf.json"},
			ExpErr:         "the dir \"" + testDir2 + "/ci-cdktf.out/stacks/test\" is not in the plan list of this pull request",
		},
	}

	globalCfgArgs := valid.GlobalCfgArgs{
		AllowAllRepoSettings: true,
	}

	logger := logging.NewNoopLogger(t)
	scope := metricstest.NewLoggingScope(t, logger, "atlantis")
	userConfig := defaultUserConfig
	userConfig.RestrictFileList = true
	userConfig.IncludeGitUntrackedFiles = true

	for _, c := range cases {
		t.Run(c.Description+"_"+command.Plan.String(), func(t *testing.T) {
			RegisterMockTestingT(t)
			tmpDir := DirStructure(t, c.DirectoryStructure)

			workingDir := mocks.NewMockWorkingDir()
			When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](),
				Any[string]())).ThenReturn(tmpDir, nil)
			When(workingDir.GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(tmpDir, nil)
			When(workingDir.GetGitUntrackedFiles(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](),
				Any[string]())).ThenReturn(c.UntrackedFiles, nil)
			vcsClient := vcsmocks.NewMockClient()
			When(vcsClient.GetModifiedFiles(Any[logging.SimpleLogging](), Any[models.Repo](),
				Any[models.PullRequest]())).ThenReturn(c.ModifiedFiles, nil)
			if c.AtlantisYAML != "" {
				err := os.WriteFile(filepath.Join(tmpDir, valid.DefaultAtlantisFile), []byte(c.AtlantisYAML), 0600)
				Ok(t, err)
			}

			terraformClient := tfclientmocks.NewMockClient()

			builder := events.NewProjectCommandBuilder(
				false, // policyChecksSupported
				&config.ParserValidator{},
				&events.DefaultProjectFinder{},
				vcsClient,
				workingDir,
				events.NewDefaultWorkingDirLocker(),
				valid.NewGlobalCfgFromArgs(globalCfgArgs),
				&events.DefaultPendingPlanFinder{},
				&events.CommentParser{ExecutableName: "atlantis"},
				userConfig.SkipCloneNoChanges,
				userConfig.EnableRegExpCmd,
				userConfig.EnableAutoMerge,
				userConfig.EnableParallelPlan,
				userConfig.EnableParallelApply,
				userConfig.AutoDetectModuleFiles,
				userConfig.AutoplanFileList,
				userConfig.RestrictFileList,
				userConfig.SilenceNoProjects,
				userConfig.IncludeGitUntrackedFiles,
				userConfig.AutoDiscoverMode,
				scope,
				terraformClient,
			)

			var actCtxs []command.ProjectContext
			var err error
			cmd := c.Cmd
			actCtxs, err = builder.BuildPlanCommands(&command.Context{
				Log:   logger,
				Scope: scope,
			}, &cmd)
			if c.ExpErr != "" {
				ErrEquals(t, c.ExpErr, err)
				return
			}
			Ok(t, err)
			Equals(t, 1, len(actCtxs))
			actCtx := actCtxs[0]
			Equals(t, c.ExpRepoRelDir, actCtx.RepoRelDir)
		})
	}
}

func TestDefaultProjectCommandBuilder_BuildPlanCommands_with_IncludeGitUntrackedFiles(t *testing.T) {
	testDir1 := "directory-1"

	cases := []struct {
		Description        string
		AtlantisYAML       string
		DirectoryStructure map[string]any
		ModifiedFiles      []string
		UntrackedFiles     []string
		Cmd                events.CommentCommand
		ExpRepoRelDir      string
		ExpErr             string
	}{
		{
			Description: "planning with a git untracked file",
			Cmd: events.CommentCommand{
				Name: command.Plan,
			},
			DirectoryStructure: map[string]any{
				testDir1: map[string]any{
					"main.ts": nil,
					"ci-cdktf.out": map[string]any{
						"stacks": map[string]any{
							"test": map[string]any{
								"cdk.tf.json": nil,
							},
						},
					},
				},
			},
			ModifiedFiles:  []string{testDir1 + "/main.ts"},
			UntrackedFiles: []string{testDir1 + "/ci-cdktf.out/stacks/test/cdk.tf.json"},
			ExpRepoRelDir:  testDir1 + "/ci-cdktf.out/stacks/test",
		},
	}

	globalCfgArgs := valid.GlobalCfgArgs{
		AllowAllRepoSettings: true,
	}

	logger := logging.NewNoopLogger(t)
	scope := metricstest.NewLoggingScope(t, logger, "atlantis")
	userConfig := defaultUserConfig
	userConfig.IncludeGitUntrackedFiles = true
	userConfig.AutoplanFileList = "**/cdk.tf.json"

	for _, c := range cases {
		t.Run(c.Description+"_"+command.Plan.String(), func(t *testing.T) {
			RegisterMockTestingT(t)
			tmpDir := DirStructure(t, c.DirectoryStructure)

			workingDir := mocks.NewMockWorkingDir()
			When(workingDir.Clone(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](),
				Any[string]())).ThenReturn(tmpDir, nil)
			When(workingDir.GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(tmpDir, nil)
			When(workingDir.GetGitUntrackedFiles(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](),
				Any[string]())).ThenReturn(c.UntrackedFiles, nil)
			vcsClient := vcsmocks.NewMockClient()
			When(vcsClient.GetModifiedFiles(Any[logging.SimpleLogging](), Any[models.Repo](),
				Any[models.PullRequest]())).ThenReturn(c.ModifiedFiles, nil)
			if c.AtlantisYAML != "" {
				err := os.WriteFile(filepath.Join(tmpDir, valid.DefaultAtlantisFile), []byte(c.AtlantisYAML), 0600)
				Ok(t, err)
			}

			terraformClient := tfclientmocks.NewMockClient()

			builder := events.NewProjectCommandBuilder(
				false, // policyChecksSupported
				&config.ParserValidator{},
				&events.DefaultProjectFinder{},
				vcsClient,
				workingDir,
				events.NewDefaultWorkingDirLocker(),
				valid.NewGlobalCfgFromArgs(globalCfgArgs),
				&events.DefaultPendingPlanFinder{},
				&events.CommentParser{ExecutableName: "atlantis"},
				userConfig.SkipCloneNoChanges,
				userConfig.EnableRegExpCmd,
				userConfig.EnableAutoMerge,
				userConfig.EnableParallelPlan,
				userConfig.EnableParallelApply,
				userConfig.AutoDetectModuleFiles,
				userConfig.AutoplanFileList,
				userConfig.RestrictFileList,
				userConfig.SilenceNoProjects,
				userConfig.IncludeGitUntrackedFiles,
				userConfig.AutoDiscoverMode,
				scope,
				terraformClient,
			)

			var actCtxs []command.ProjectContext
			var err error
			cmd := c.Cmd
			actCtxs, err = builder.BuildPlanCommands(&command.Context{
				Log:   logger,
				Scope: scope,
			}, &cmd)
			if c.ExpErr != "" {
				ErrEquals(t, c.ExpErr, err)
				return
			}
			Ok(t, err)
			Equals(t, 1, len(actCtxs))
			actCtx := actCtxs[0]
			Equals(t, c.ExpRepoRelDir, actCtx.RepoRelDir)
		})
	}
}
