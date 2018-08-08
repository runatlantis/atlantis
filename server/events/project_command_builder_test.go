package events_test

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/events/yaml"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestDefaultProjectCommandBuilder_BuildAutoplanCommands(t *testing.T) {
	// exp defines what we will assert on. We don't check all fields in the
	// actual contexts.
	type exp struct {
		projectConfig *valid.Project
		dir           string
		workspace     string
	}
	cases := []struct {
		Description  string
		AtlantisYAML string
		exp          []exp
	}{
		{
			Description:  "no atlantis.yaml",
			AtlantisYAML: "",
			exp: []exp{
				{
					projectConfig: nil,
					dir:           ".",
					workspace:     "default",
				},
			},
		},
		{
			Description: "autoplan disabled",
			AtlantisYAML: `
version: 2
projects:
- dir: .
  autoplan:
    enabled: false`,
			exp: nil,
		},
		{
			Description: "simple atlantis.yaml",
			AtlantisYAML: `
version: 2
projects:
- dir: .
`,
			exp: []exp{
				{
					projectConfig: &valid.Project{
						Dir:       ".",
						Workspace: "default",
						Autoplan: valid.Autoplan{
							Enabled:      true,
							WhenModified: []string{"**/*.tf*"},
						},
					},
					dir:       ".",
					workspace: "default",
				},
			},
		},
		{
			Description: "some projects disabled",
			AtlantisYAML: `
version: 2
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
			exp: []exp{
				{
					projectConfig: &valid.Project{
						Dir:       ".",
						Workspace: "myworkspace",
						Autoplan: valid.Autoplan{
							Enabled:      true,
							WhenModified: []string{"main.tf"},
						},
					},
					dir:       ".",
					workspace: "myworkspace",
				},
				{
					projectConfig: &valid.Project{
						Dir:       ".",
						Workspace: "myworkspace2",
						Autoplan: valid.Autoplan{
							Enabled:      true,
							WhenModified: []string{"**/*.tf*"},
						},
					},
					dir:       ".",
					workspace: "myworkspace2",
				},
			},
		},
		{
			Description: "some projects disabled",
			AtlantisYAML: `
version: 2
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
			exp: []exp{
				{
					projectConfig: &valid.Project{
						Dir:       ".",
						Workspace: "myworkspace",
						Autoplan: valid.Autoplan{
							Enabled:      true,
							WhenModified: []string{"main.tf"},
						},
					},
					dir:       ".",
					workspace: "myworkspace",
				},
				{
					projectConfig: &valid.Project{
						Dir:       ".",
						Workspace: "myworkspace2",
						Autoplan: valid.Autoplan{
							Enabled:      true,
							WhenModified: []string{"**/*.tf*"},
						},
					},
					dir:       ".",
					workspace: "myworkspace2",
				},
			},
		},
		{
			Description: "no projects modified",
			AtlantisYAML: `
version: 2
projects:
- dir: mydir
`,
			exp: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			RegisterMockTestingT(t)
			tmpDir, cleanup := TempDir(t)
			defer cleanup()

			baseRepo := models.Repo{}
			headRepo := models.Repo{}
			pull := models.PullRequest{}
			logger := logging.NewNoopLogger()
			workingDir := mocks.NewMockWorkingDir()
			When(workingDir.Clone(logger, baseRepo, headRepo, pull, "default")).ThenReturn(tmpDir, nil)
			if c.AtlantisYAML != "" {
				err := ioutil.WriteFile(filepath.Join(tmpDir, yaml.AtlantisYAMLFilename), []byte(c.AtlantisYAML), 0600)
				Ok(t, err)
			}
			err := ioutil.WriteFile(filepath.Join(tmpDir, "main.tf"), nil, 0600)
			Ok(t, err)

			vcsClient := vcsmocks.NewMockClientProxy()
			When(vcsClient.GetModifiedFiles(baseRepo, pull)).ThenReturn([]string{"main.tf"}, nil)

			builder := &events.DefaultProjectCommandBuilder{
				WorkingDirLocker: events.NewDefaultWorkingDirLocker(),
				WorkingDir:       workingDir,
				ParserValidator:  &yaml.ParserValidator{},
				VCSClient:        vcsClient,
				ProjectFinder:    &events.DefaultProjectFinder{},
				AllowRepoConfig:  true,
			}

			ctxs, err := builder.BuildAutoplanCommands(&events.CommandContext{
				BaseRepo: baseRepo,
				HeadRepo: headRepo,
				Pull:     pull,
				User:     models.User{},
				Log:      logger,
			})
			Ok(t, err)
			Equals(t, len(c.exp), len(ctxs))

			for i, actCtx := range ctxs {
				expCtx := c.exp[i]
				Equals(t, baseRepo, actCtx.BaseRepo)
				Equals(t, baseRepo, actCtx.HeadRepo)
				Equals(t, pull, actCtx.Pull)
				Equals(t, models.User{}, actCtx.User)
				Equals(t, logger, actCtx.Log)
				Equals(t, 0, len(actCtx.CommentArgs))

				Equals(t, expCtx.projectConfig, actCtx.ProjectConfig)
				Equals(t, expCtx.dir, actCtx.RepoRelDir)
				Equals(t, expCtx.workspace, actCtx.Workspace)
			}
		})
	}
}

func TestDefaultProjectCommandBuilder_BuildPlanApplyCommand(t *testing.T) {
	cases := []struct {
		Description      string
		AtlantisYAML     string
		Cmd              events.CommentCommand
		ExpProjectConfig *valid.Project
		ExpCommentArgs   []string
		ExpWorkspace     string
		ExpDir           string
		ExpErr           string
	}{
		{
			Description: "no atlantis.yaml",
			Cmd: events.CommentCommand{
				RepoRelDir: ".",
				Flags:      []string{"commentarg"},
				Name:       events.PlanCommand,
				Workspace:  "myworkspace",
			},
			AtlantisYAML:     "",
			ExpProjectConfig: nil,
			ExpCommentArgs:   []string{"commentarg"},
			ExpWorkspace:     "myworkspace",
			ExpDir:           ".",
		},
		{
			Description: "no atlantis.yaml with project flag",
			Cmd: events.CommentCommand{
				RepoRelDir:  ".",
				Name:        events.PlanCommand,
				ProjectName: "myproject",
			},
			AtlantisYAML: "",
			ExpErr:       "cannot specify a project name unless an atlantis.yaml file exists to configure projects",
		},
		{
			Description: "simple atlantis.yaml",
			Cmd: events.CommentCommand{
				RepoRelDir: ".",
				Name:       events.PlanCommand,
				Workspace:  "myworkspace",
			},
			AtlantisYAML: `
version: 2
projects:
- dir: .
  workspace: myworkspace
  apply_requirements: [approved]`,
			ExpProjectConfig: &valid.Project{
				Dir:       ".",
				Workspace: "myworkspace",
				Autoplan: valid.Autoplan{
					WhenModified: []string{"**/*.tf*"},
					Enabled:      true,
				},
				ApplyRequirements: []string{"approved"},
			},
			ExpWorkspace: "myworkspace",
			ExpDir:       ".",
		},
		{
			Description: "atlantis.yaml wrong dir",
			Cmd: events.CommentCommand{
				RepoRelDir: ".",
				Name:       events.PlanCommand,
				Workspace:  "myworkspace",
			},
			AtlantisYAML: `
version: 2
projects:
- dir: notroot
  workspace: myworkspace
  apply_requirements: [approved]`,
			ExpProjectConfig: nil,
			ExpWorkspace:     "myworkspace",
			ExpDir:           ".",
		},
		{
			Description: "atlantis.yaml wrong workspace",
			Cmd: events.CommentCommand{
				RepoRelDir: ".",
				Name:       events.PlanCommand,
				Workspace:  "myworkspace",
			},
			AtlantisYAML: `
version: 2
projects:
- dir: .
  workspace: notmyworkspace
  apply_requirements: [approved]`,
			ExpProjectConfig: nil,
			ExpWorkspace:     "myworkspace",
			ExpDir:           ".",
		},
		{
			Description: "atlantis.yaml with projectname",
			Cmd: events.CommentCommand{
				Name:        events.PlanCommand,
				ProjectName: "myproject",
			},
			AtlantisYAML: `
version: 2
projects:
- name: myproject
  dir: .
  workspace: myworkspace
  apply_requirements: [approved]`,
			ExpProjectConfig: &valid.Project{
				Dir:       ".",
				Workspace: "myworkspace",
				Autoplan: valid.Autoplan{
					WhenModified: []string{"**/*.tf*"},
					Enabled:      true,
				},
				ApplyRequirements: []string{"approved"},
				Name:              String("myproject"),
			},
			ExpWorkspace: "myworkspace",
			ExpDir:       ".",
		},
		{
			Description: "atlantis.yaml with multiple dir/workspaces matching",
			Cmd: events.CommentCommand{
				Name:       events.PlanCommand,
				RepoRelDir: ".",
				Workspace:  "myworkspace",
			},
			AtlantisYAML: `
version: 2
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
				Name:        events.PlanCommand,
				RepoRelDir:  ".",
				Workspace:   "default",
				ProjectName: "notconfigured",
			},
			AtlantisYAML: `
version: 2
projects:
- dir: .
`,
			ExpErr: "no project with name \"notconfigured\" is defined in atlantis.yaml",
		},
	}

	for _, c := range cases {
		// NOTE: we're testing both plan and apply here.
		for _, cmdName := range []events.CommandName{events.PlanCommand, events.ApplyCommand} {
			t.Run(c.Description, func(t *testing.T) {
				RegisterMockTestingT(t)
				tmpDir, cleanup := TempDir(t)
				defer cleanup()

				baseRepo := models.Repo{}
				headRepo := models.Repo{}
				pull := models.PullRequest{}
				logger := logging.NewNoopLogger()
				workingDir := mocks.NewMockWorkingDir()
				expWorkspace := c.Cmd.Workspace
				if expWorkspace == "" {
					expWorkspace = "default"
				}
				if cmdName == events.PlanCommand {
					When(workingDir.Clone(logger, baseRepo, headRepo, pull, expWorkspace)).ThenReturn(tmpDir, nil)
				} else {
					When(workingDir.GetWorkingDir(baseRepo, pull, expWorkspace)).ThenReturn(tmpDir, nil)
				}
				if c.AtlantisYAML != "" {
					err := ioutil.WriteFile(filepath.Join(tmpDir, yaml.AtlantisYAMLFilename), []byte(c.AtlantisYAML), 0600)
					Ok(t, err)
				}
				err := ioutil.WriteFile(filepath.Join(tmpDir, "main.tf"), nil, 0600)
				Ok(t, err)

				vcsClient := vcsmocks.NewMockClientProxy()
				When(vcsClient.GetModifiedFiles(baseRepo, pull)).ThenReturn([]string{"main.tf"}, nil)

				builder := &events.DefaultProjectCommandBuilder{
					WorkingDirLocker: events.NewDefaultWorkingDirLocker(),
					WorkingDir:       workingDir,
					ParserValidator:  &yaml.ParserValidator{},
					VCSClient:        vcsClient,
					ProjectFinder:    &events.DefaultProjectFinder{},
					AllowRepoConfig:  true,
				}

				cmdCtx := &events.CommandContext{
					BaseRepo: baseRepo,
					HeadRepo: headRepo,
					Pull:     pull,
					User:     models.User{},
					Log:      logger,
				}
				var actCtxs []models.ProjectCommandContext
				if cmdName == events.PlanCommand {
					actCtxs, err = builder.BuildPlanCommands(cmdCtx, &c.Cmd)
				} else {
					actCtxs, err = builder.BuildApplyCommands(cmdCtx, &c.Cmd)
				}

				if c.ExpErr != "" {
					ErrEquals(t, c.ExpErr, err)
					return
				}
				Ok(t, err)

				Equals(t, 1, len(actCtxs))
				actCtx := actCtxs[0]
				Equals(t, baseRepo, actCtx.BaseRepo)
				Equals(t, baseRepo, actCtx.HeadRepo)
				Equals(t, pull, actCtx.Pull)
				Equals(t, models.User{}, actCtx.User)
				Equals(t, logger, actCtx.Log)

				Equals(t, c.ExpProjectConfig, actCtx.ProjectConfig)
				Equals(t, c.ExpDir, actCtx.RepoRelDir)
				Equals(t, c.ExpWorkspace, actCtx.Workspace)
				Equals(t, c.ExpCommentArgs, actCtx.CommentArgs)
			})
		}
	}
}

func String(v string) *string { return &v }
