package events_test

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/mocks/matchers"
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
			When(workingDir.Clone(logger, baseRepo, headRepo, pull, false, "default")).ThenReturn(tmpDir, nil)
			if c.AtlantisYAML != "" {
				err := ioutil.WriteFile(filepath.Join(tmpDir, yaml.AtlantisYAMLFilename), []byte(c.AtlantisYAML), 0600)
				Ok(t, err)
			}
			err := ioutil.WriteFile(filepath.Join(tmpDir, "main.tf"), nil, 0600)
			Ok(t, err)

			vcsClient := vcsmocks.NewMockClientProxy()
			When(vcsClient.GetModifiedFiles(baseRepo, pull)).ThenReturn([]string{"main.tf"}, nil)

			builder := &events.DefaultProjectCommandBuilder{
				WorkingDirLocker:    events.NewDefaultWorkingDirLocker(),
				WorkingDir:          workingDir,
				ParserValidator:     &yaml.ParserValidator{},
				VCSClient:           vcsClient,
				ProjectFinder:       &events.DefaultProjectFinder{},
				AllowRepoConfig:     true,
				PendingPlanFinder:   &events.PendingPlanFinder{},
				AllowRepoConfigFlag: "allow-repo-config",
				CommentBuilder:      &events.CommentParser{},
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

// Test building a plan and apply command for one project.
func TestDefaultProjectCommandBuilder_BuildSinglePlanApplyCommand(t *testing.T) {
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
					When(workingDir.Clone(logger, baseRepo, headRepo, pull, false, expWorkspace)).ThenReturn(tmpDir, nil)
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
					WorkingDirLocker:    events.NewDefaultWorkingDirLocker(),
					WorkingDir:          workingDir,
					ParserValidator:     &yaml.ParserValidator{},
					VCSClient:           vcsClient,
					ProjectFinder:       &events.DefaultProjectFinder{},
					AllowRepoConfig:     true,
					AllowRepoConfigFlag: "allow-repo-config",
					CommentBuilder:      &events.CommentParser{},
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

// Test building plan command for multiple projects when the comment
// isn't for a specific project, i.e. atlantis plan and there's no atlantis.yaml.
// In this case we should just use the list of modified projects.
func TestDefaultProjectCommandBuilder_BuildMultiPlanNoAtlantisYAML(t *testing.T) {
	RegisterMockTestingT(t)
	tmpDir, cleanup := DirStructure(t, map[string]interface{}{
		"project1": map[string]interface{}{
			"main.tf": nil,
		},
		"project2": map[string]interface{}{
			"main.tf": nil,
		},
	})
	defer cleanup()
	workingDir := mocks.NewMockWorkingDir()
	When(workingDir.Clone(
		matchers.AnyPtrToLoggingSimpleLogger(),
		matchers.AnyModelsRepo(),
		matchers.AnyModelsRepo(),
		matchers.AnyModelsPullRequest(),
		AnyBool(),
		AnyString())).ThenReturn(tmpDir, nil)
	vcsClient := vcsmocks.NewMockClientProxy()
	When(vcsClient.GetModifiedFiles(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())).ThenReturn([]string{"project1/main.tf", "project2/main.tf"}, nil)

	builder := &events.DefaultProjectCommandBuilder{
		WorkingDirLocker:    events.NewDefaultWorkingDirLocker(),
		WorkingDir:          workingDir,
		ParserValidator:     &yaml.ParserValidator{},
		VCSClient:           vcsClient,
		ProjectFinder:       &events.DefaultProjectFinder{},
		AllowRepoConfig:     true,
		AllowRepoConfigFlag: "allow-repo-config",
		CommentBuilder:      &events.CommentParser{},
	}

	ctxs, err := builder.BuildPlanCommands(&events.CommandContext{
		BaseRepo: models.Repo{},
		HeadRepo: models.Repo{},
		Pull:     models.PullRequest{},
		User:     models.User{},
		Log:      logging.NewNoopLogger(),
	}, &events.CommentCommand{
		RepoRelDir:  "",
		Flags:       nil,
		Name:        events.PlanCommand,
		Verbose:     false,
		Workspace:   "",
		ProjectName: "",
	})
	Ok(t, err)
	Equals(t, 2, len(ctxs))
	Equals(t, "project1", ctxs[0].RepoRelDir)
	Equals(t, "default", ctxs[0].Workspace)
	var nilProjectConfig *valid.Project
	Equals(t, nilProjectConfig, ctxs[0].ProjectConfig)
	Equals(t, "project2", ctxs[1].RepoRelDir)
	Equals(t, "default", ctxs[1].Workspace)
	Equals(t, nilProjectConfig, ctxs[1].ProjectConfig)
}

// Test building plan command for multiple projects when the comment
// isn't for a specific project, i.e. atlantis plan and there's no atlantis.yaml.
// In this case there are no modified files so there should be 0 plans.
func TestDefaultProjectCommandBuilder_BuildMultiPlanNoAtlantisYAMLNoModified(t *testing.T) {
	RegisterMockTestingT(t)
	tmpDir, cleanup := TempDir(t)
	defer cleanup()
	workingDir := mocks.NewMockWorkingDir()
	When(workingDir.Clone(
		matchers.AnyPtrToLoggingSimpleLogger(),
		matchers.AnyModelsRepo(),
		matchers.AnyModelsRepo(),
		matchers.AnyModelsPullRequest(),
		AnyBool(),
		AnyString())).ThenReturn(tmpDir, nil)
	vcsClient := vcsmocks.NewMockClientProxy()
	When(vcsClient.GetModifiedFiles(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())).ThenReturn([]string{}, nil)

	builder := &events.DefaultProjectCommandBuilder{
		WorkingDirLocker:    events.NewDefaultWorkingDirLocker(),
		WorkingDir:          workingDir,
		ParserValidator:     &yaml.ParserValidator{},
		VCSClient:           vcsClient,
		ProjectFinder:       &events.DefaultProjectFinder{},
		AllowRepoConfig:     true,
		AllowRepoConfigFlag: "allow-repo-config",
		CommentBuilder:      &events.CommentParser{},
	}

	ctxs, err := builder.BuildPlanCommands(&events.CommandContext{
		BaseRepo: models.Repo{},
		HeadRepo: models.Repo{},
		Pull:     models.PullRequest{},
		User:     models.User{},
		Log:      logging.NewNoopLogger(),
	}, &events.CommentCommand{
		RepoRelDir:  "",
		Flags:       nil,
		Name:        events.PlanCommand,
		Verbose:     false,
		Workspace:   "",
		ProjectName: "",
	})
	Ok(t, err)
	Equals(t, 0, len(ctxs))
}

// Test building plan command for multiple projects when the comment
// isn't for a specific project, i.e. atlantis plan and there is an atlantis.yaml.
// In this case we should follow the when_modified section of the autoplan config.
func TestDefaultProjectCommandBuilder_BuildMultiPlanWithAtlantisYAML(t *testing.T) {
	RegisterMockTestingT(t)
	tmpDir, cleanup := DirStructure(t, map[string]interface{}{
		"project1": map[string]interface{}{
			"main.tf": nil,
		},
		"project2": map[string]interface{}{
			"main.tf": nil,
		},
		"project3": map[string]interface{}{
			"main.tf": nil,
		},
	})
	defer cleanup()
	yamlCfg := `version: 2
projects:
- dir: project1 # project1 uses the defaults
- dir: project2 # project2 has autoplan disabled but should use default when_modified
  autoplan:
    enabled: false
- dir: project3 # project3 has an empty when_modified
  autoplan:
    enabled: false
    when_modified: []
`
	err := ioutil.WriteFile(filepath.Join(tmpDir, yaml.AtlantisYAMLFilename), []byte(yamlCfg), 0600)
	Ok(t, err)

	workingDir := mocks.NewMockWorkingDir()
	When(workingDir.Clone(
		matchers.AnyPtrToLoggingSimpleLogger(),
		matchers.AnyModelsRepo(),
		matchers.AnyModelsRepo(),
		matchers.AnyModelsPullRequest(),
		AnyBool(),
		AnyString())).ThenReturn(tmpDir, nil)
	vcsClient := vcsmocks.NewMockClientProxy()
	When(vcsClient.GetModifiedFiles(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())).ThenReturn([]string{
		"project1/main.tf", "project2/main.tf", "project3/main.tf",
	}, nil)

	builder := &events.DefaultProjectCommandBuilder{
		WorkingDirLocker:    events.NewDefaultWorkingDirLocker(),
		WorkingDir:          workingDir,
		ParserValidator:     &yaml.ParserValidator{},
		VCSClient:           vcsClient,
		ProjectFinder:       &events.DefaultProjectFinder{},
		AllowRepoConfig:     true,
		AllowRepoConfigFlag: "allow-repo-config",
		CommentBuilder:      &events.CommentParser{},
	}

	ctxs, err := builder.BuildPlanCommands(&events.CommandContext{
		BaseRepo: models.Repo{},
		HeadRepo: models.Repo{},
		Pull:     models.PullRequest{},
		User:     models.User{},
		Log:      logging.NewNoopLogger(),
	}, &events.CommentCommand{
		RepoRelDir:  "",
		Flags:       nil,
		Name:        events.PlanCommand,
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

// Test building plan command for multiple projects when the comment
// isn't for a specific project, i.e. atlantis plan and there is an atlantis.yaml.
// In this case we should follow the when_modified section of the autoplan config.
func TestDefaultProjectCommandBuilder_BuildMultiPlanWithAtlantisYAMLWorkspaces(t *testing.T) {
	RegisterMockTestingT(t)
	tmpDir, cleanup := DirStructure(t, map[string]interface{}{
		"main.tf": nil,
	})
	defer cleanup()
	yamlCfg := `version: 2
projects:
- dir: .
  workspace: staging
- dir: .
  workspace: production
`
	err := ioutil.WriteFile(filepath.Join(tmpDir, yaml.AtlantisYAMLFilename), []byte(yamlCfg), 0600)
	Ok(t, err)

	workingDir := mocks.NewMockWorkingDir()
	When(workingDir.Clone(
		matchers.AnyPtrToLoggingSimpleLogger(),
		matchers.AnyModelsRepo(),
		matchers.AnyModelsRepo(),
		matchers.AnyModelsPullRequest(),
		AnyBool(),
		AnyString())).ThenReturn(tmpDir, nil)
	vcsClient := vcsmocks.NewMockClientProxy()
	When(vcsClient.GetModifiedFiles(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())).ThenReturn([]string{"main.tf"}, nil)

	builder := &events.DefaultProjectCommandBuilder{
		WorkingDirLocker:    events.NewDefaultWorkingDirLocker(),
		WorkingDir:          workingDir,
		ParserValidator:     &yaml.ParserValidator{},
		VCSClient:           vcsClient,
		ProjectFinder:       &events.DefaultProjectFinder{},
		AllowRepoConfig:     true,
		AllowRepoConfigFlag: "allow-repo-config",
		CommentBuilder:      &events.CommentParser{},
	}

	ctxs, err := builder.BuildPlanCommands(&events.CommandContext{
		BaseRepo: models.Repo{},
		HeadRepo: models.Repo{},
		Pull:     models.PullRequest{},
		User:     models.User{},
		Log:      logging.NewNoopLogger(),
	}, &events.CommentCommand{
		RepoRelDir:  "",
		Flags:       nil,
		Name:        events.PlanCommand,
		Verbose:     false,
		Workspace:   "",
		ProjectName: "",
	})
	Ok(t, err)
	Equals(t, 2, len(ctxs))
	Equals(t, ".", ctxs[0].RepoRelDir)
	Equals(t, "staging", ctxs[0].Workspace)
	Equals(t, ".", ctxs[1].RepoRelDir)
	Equals(t, "production", ctxs[1].Workspace)
}

// Test building apply command for multiple projects when the comment
// isn't for a specific project, i.e. atlantis apply.
// In this case we should apply all outstanding plans.
func TestDefaultProjectCommandBuilder_BuildMultiApply(t *testing.T) {
	RegisterMockTestingT(t)
	tmpDir, cleanup := DirStructure(t, map[string]interface{}{
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
	defer cleanup()
	// Initialize git repos in each workspace so that the .tfplan files get
	// picked up.
	runCmd(t, filepath.Join(tmpDir, "workspace1"), "git", "init")
	runCmd(t, filepath.Join(tmpDir, "workspace2"), "git", "init")

	workingDir := mocks.NewMockWorkingDir()
	When(workingDir.GetPullDir(
		matchers.AnyModelsRepo(),
		matchers.AnyModelsPullRequest())).
		ThenReturn(tmpDir, nil)

	builder := &events.DefaultProjectCommandBuilder{
		WorkingDirLocker:    events.NewDefaultWorkingDirLocker(),
		WorkingDir:          workingDir,
		ParserValidator:     &yaml.ParserValidator{},
		VCSClient:           nil,
		ProjectFinder:       &events.DefaultProjectFinder{},
		AllowRepoConfig:     true,
		AllowRepoConfigFlag: "allow-repo-config",
		PendingPlanFinder:   &events.PendingPlanFinder{},
		CommentBuilder:      &events.CommentParser{},
	}

	ctxs, err := builder.BuildApplyCommands(&events.CommandContext{
		BaseRepo: models.Repo{},
		HeadRepo: models.Repo{},
		Pull:     models.PullRequest{},
		User:     models.User{},
		Log:      logging.NewNoopLogger(),
	}, &events.CommentCommand{
		RepoRelDir:  "",
		Flags:       nil,
		Name:        events.ApplyCommand,
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

// Test that if repo config is disabled we error out if there's an atlantis.yaml
// file.
func TestDefaultProjectCommandBuilder_RepoConfigDisabled(t *testing.T) {
	RegisterMockTestingT(t)
	workingDir := mocks.NewMockWorkingDir()

	tmpDir, cleanup := DirStructure(t, map[string]interface{}{
		"pulldir": map[string]interface{}{
			"workspace": map[string]interface{}{},
		},
	})
	defer cleanup()
	repoDir := filepath.Join(tmpDir, "pulldir/workspace")
	err := ioutil.WriteFile(filepath.Join(repoDir, yaml.AtlantisYAMLFilename), nil, 0600)
	Ok(t, err)

	When(workingDir.Clone(
		matchers.AnyPtrToLoggingSimpleLogger(),
		matchers.AnyModelsRepo(),
		matchers.AnyModelsRepo(),
		matchers.AnyModelsPullRequest(),
		AnyBool(),
		AnyString())).ThenReturn(repoDir, nil)
	When(workingDir.GetWorkingDir(
		matchers.AnyModelsRepo(),
		matchers.AnyModelsPullRequest(),
		AnyString())).ThenReturn(repoDir, nil)

	builder := &events.DefaultProjectCommandBuilder{
		WorkingDirLocker:    events.NewDefaultWorkingDirLocker(),
		WorkingDir:          workingDir,
		ParserValidator:     &yaml.ParserValidator{},
		VCSClient:           nil,
		ProjectFinder:       &events.DefaultProjectFinder{},
		AllowRepoConfig:     false,
		AllowRepoConfigFlag: "allow-repo-config",
		CommentBuilder:      &events.CommentParser{},
	}

	ctx := &events.CommandContext{
		BaseRepo: models.Repo{},
		HeadRepo: models.Repo{},
		Pull:     models.PullRequest{},
		User:     models.User{},
		Log:      logging.NewNoopLogger(),
	}
	_, err = builder.BuildAutoplanCommands(ctx)
	ErrEquals(t, "atlantis.yaml files not allowed because Atlantis is not running with --allow-repo-config", err)

	commentCmd := &events.CommentCommand{
		RepoRelDir:  "",
		Flags:       nil,
		Name:        0,
		Verbose:     false,
		Workspace:   "workspace",
		ProjectName: "",
	}
	_, err = builder.BuildPlanCommands(ctx, commentCmd)
	ErrEquals(t, "atlantis.yaml files not allowed because Atlantis is not running with --allow-repo-config", err)

	_, err = builder.BuildApplyCommands(ctx, commentCmd)
	ErrEquals(t, "atlantis.yaml files not allowed because Atlantis is not running with --allow-repo-config", err)
}

func String(v string) *string { return &v }
