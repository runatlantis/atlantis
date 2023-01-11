package events_test

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/core/config"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/matchers"
	"github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics"
	"github.com/runatlantis/atlantis/server/wrappers"
	. "github.com/runatlantis/atlantis/testing"
	"github.com/uber-go/tally/v4"
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

	logger := logging.NewNoopCtxLogger(t)
	scope, _, _ := metrics.NewLoggingScope(logger, "atlantis")

	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			RegisterMockTestingT(t)
			tmpDir, cleanup := DirStructure(t, map[string]interface{}{
				"main.tf": nil,
			})
			defer cleanup()

			workingDir := mocks.NewMockWorkingDir()
			When(workingDir.Clone(matchers.AnyLoggingLogger(), matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest(), AnyString())).ThenReturn(tmpDir, false, nil)
			vcsClient := vcsmocks.NewMockClient()
			When(vcsClient.GetModifiedFiles(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())).ThenReturn([]string{"main.tf"}, nil)
			if c.AtlantisYAML != "" {
				err := os.WriteFile(filepath.Join(tmpDir, config.AtlantisYAMLFilename), []byte(c.AtlantisYAML), 0600)
				Ok(t, err)
			}

			builder := events.NewProjectCommandBuilder(
				events.NewProjectCommandContextBuilder(&events.CommentParser{}),
				&config.ParserValidator{},
				&events.DefaultProjectFinder{},
				vcsClient,
				workingDir,
				events.NewDefaultWorkingDirLocker(),
				valid.NewGlobalCfg("somedir"),
				&events.DefaultPendingPlanFinder{},
				false,
				"**/*.tf,**/*.tfvars,**/*.tfvars.json,**/terragrunt.hcl",
				logger,
				events.InfiniteProjectsPerPR,
			)

			ctxs, err := builder.BuildAutoplanCommands(&command.Context{
				RequestCtx: context.TODO(),
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
		Description      string
		AtlantisYAML     string
		Cmd              command.Comment
		ExpCommentArgs   []string
		ExpWorkspace     string
		ExpDir           string
		ExpProjectName   string
		ExpErr           string
		ExpApplyReqs     []string
		ExpParallelApply bool
		ExpParallelPlan  bool
	}{
		{
			Description: "no atlantis.yaml",
			Cmd: command.Comment{
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
			Cmd: command.Comment{
				RepoRelDir:  ".",
				Name:        command.Plan,
				ProjectName: "myproject",
			},
			AtlantisYAML: "",
			ExpErr:       "cannot specify a project name unless an atlantis.yaml file exists to configure projects",
		},
		{
			Description: "simple atlantis.yaml",
			Cmd: command.Comment{
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
			Cmd: command.Comment{
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
			Cmd: command.Comment{
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
			Cmd: command.Comment{
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
			Cmd: command.Comment{
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
			Cmd: command.Comment{
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
			Cmd: command.Comment{
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
			Cmd: command.Comment{
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
			Description: "atlantis.yaml with ParallelPlan Set to true",
			Cmd: command.Comment{
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
	}

	logger := logging.NewNoopCtxLogger(t)
	scope, _, _ := metrics.NewLoggingScope(logger, "atlantis")

	for _, c := range cases {
		// NOTE: we're testing both plan and apply here.
		for _, cmdName := range []command.Name{command.Plan, command.Apply} {
			t.Run(c.Description+"_"+cmdName.String(), func(t *testing.T) {
				RegisterMockTestingT(t)
				tmpDir, cleanup := DirStructure(t, map[string]interface{}{
					"main.tf": nil,
				})
				defer cleanup()

				workingDir := mocks.NewMockWorkingDir()
				When(workingDir.Clone(matchers.AnyLoggingLogger(), matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest(), AnyString())).ThenReturn(tmpDir, false, nil)
				When(workingDir.GetWorkingDir(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest(), AnyString())).ThenReturn(tmpDir, nil)
				vcsClient := vcsmocks.NewMockClient()
				When(vcsClient.GetModifiedFiles(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())).ThenReturn([]string{"main.tf"}, nil)
				if c.AtlantisYAML != "" {
					err := os.WriteFile(filepath.Join(tmpDir, config.AtlantisYAMLFilename), []byte(c.AtlantisYAML), 0600)
					Ok(t, err)
				}

				globalCfg := valid.NewGlobalCfg("somedir")
				globalCfg.Repos[0].AllowedOverrides = []string{"apply_requirements"}

				builder := events.NewProjectCommandBuilder(
					events.NewProjectCommandContextBuilder(&events.CommentParser{}),
					&config.ParserValidator{},
					&events.DefaultProjectFinder{},
					vcsClient,
					workingDir,
					events.NewDefaultWorkingDirLocker(),
					globalCfg,
					&events.DefaultPendingPlanFinder{},
					true,
					"**/*.tf,**/*.tfvars,**/*.tfvars.json,**/terragrunt.hcl",
					logger,
					events.InfiniteProjectsPerPR,
				)

				var actCtxs []command.ProjectContext
				var err error
				if cmdName == command.Plan {
					actCtxs, err = builder.BuildPlanCommands(&command.Context{
						RequestCtx: context.TODO(),
						Log:        logger,
						Scope:      scope,
					}, &c.Cmd)
				} else {
					actCtxs, err = builder.BuildApplyCommands(&command.Context{Log: logger, Scope: scope, RequestCtx: context.TODO()}, &c.Cmd)
				}

				if c.ExpErr != "" {
					ErrEquals(t, c.ExpErr, err)
					return
				}
				Ok(t, err)
				Equals(t, 1, len(actCtxs))
				actCtx := actCtxs[0]
				Equals(t, c.ExpDir, actCtx.RepoRelDir)
				Equals(t, c.ExpWorkspace, actCtx.Workspace)
				Equals(t, c.ExpCommentArgs, actCtx.EscapedCommentArgs)
				Equals(t, c.ExpProjectName, actCtx.ProjectName)
				Equals(t, c.ExpApplyReqs, actCtx.ApplyRequirements)
				Equals(t, c.ExpParallelApply, actCtx.ParallelApplyEnabled)
				Equals(t, c.ExpParallelPlan, actCtx.ParallelPlanEnabled)
			})
		}
	}
}

func TestDefaultProjectCommandBuilder_BuildPlanCommands(t *testing.T) {
	// expCtxFields define the ctx fields we're going to assert on.
	// Since we're focused on autoplanning here, we don't validate all the
	// fields so the tests are more obvious and targeted.
	type expCtxFields struct {
		ProjectName string
		RepoRelDir  string
		Workspace   string
	}
	cases := map[string]struct {
		DirStructure  map[string]interface{}
		AtlantisYAML  string
		ModifiedFiles []string
		Exp           []expCtxFields
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

	logger := logging.NewNoopCtxLogger(t)
	scope, _, _ := metrics.NewLoggingScope(logger, "atlantis")
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			RegisterMockTestingT(t)
			tmpDir, cleanup := DirStructure(t, c.DirStructure)
			defer cleanup()

			workingDir := mocks.NewMockWorkingDir()
			When(workingDir.Clone(matchers.AnyLoggingLogger(), matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest(), AnyString())).ThenReturn(tmpDir, false, nil)
			When(workingDir.GetWorkingDir(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest(), AnyString())).ThenReturn(tmpDir, nil)
			vcsClient := vcsmocks.NewMockClient()
			When(vcsClient.GetModifiedFiles(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())).ThenReturn(c.ModifiedFiles, nil)
			if c.AtlantisYAML != "" {
				err := os.WriteFile(filepath.Join(tmpDir, config.AtlantisYAMLFilename), []byte(c.AtlantisYAML), 0600)
				Ok(t, err)
			}

			builder := events.NewProjectCommandBuilder(
				events.NewProjectCommandContextBuilder(&events.CommentParser{}),
				&config.ParserValidator{},
				&events.DefaultProjectFinder{},
				vcsClient,
				workingDir,
				events.NewDefaultWorkingDirLocker(),
				valid.NewGlobalCfg("somedir"),
				&events.DefaultPendingPlanFinder{},
				false,
				"**/*.tf,**/*.tfvars,**/*.tfvars.json,**/terragrunt.hcl",
				logger,
				events.InfiniteProjectsPerPR,
			)

			ctxs, err := builder.BuildPlanCommands(
				&command.Context{
					Log:        logger,
					Scope:      scope,
					RequestCtx: context.TODO(),
				},
				&command.Comment{
					RepoRelDir:  "",
					Flags:       nil,
					Name:        command.Plan,
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
			}
		})
	}
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

	logger := logging.NewNoopCtxLogger(t)

	scope, _, _ := metrics.NewLoggingScope(logger, "atlantis")

	builder := events.NewProjectCommandBuilder(
		events.NewProjectCommandContextBuilder(&events.CommentParser{}),
		&config.ParserValidator{},
		&events.DefaultProjectFinder{},
		nil,
		workingDir,
		events.NewDefaultWorkingDirLocker(),
		valid.NewGlobalCfg("somedir"),
		&events.DefaultPendingPlanFinder{},
		false,
		"**/*.tf,**/*.tfvars,**/*.tfvars.json,**/terragrunt.hcl",
		logger,
		events.InfiniteProjectsPerPR,
	)

	ctxs, err := builder.BuildApplyCommands(
		&command.Context{
			Log:        logger,
			Scope:      scope,
			RequestCtx: context.TODO(),
		},
		&command.Comment{
			RepoRelDir:  "",
			Flags:       nil,
			Name:        command.Apply,
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

	tmpDir, cleanup := DirStructure(t, map[string]interface{}{
		"pulldir": map[string]interface{}{
			"notconfigured": map[string]interface{}{},
		},
	})
	defer cleanup()
	repoDir := filepath.Join(tmpDir, "pulldir/notconfigured")

	yamlCfg := `version: 3
projects:
- dir: .
  workspace: default
- dir: .
  workspace: staging
`
	err := os.WriteFile(filepath.Join(repoDir, config.AtlantisYAMLFilename), []byte(yamlCfg), 0600)
	Ok(t, err)

	When(workingDir.Clone(
		matchers.AnyLoggingLogger(),
		matchers.AnyModelsRepo(),
		matchers.AnyModelsPullRequest(),
		AnyString())).ThenReturn(repoDir, false, nil)
	When(workingDir.GetWorkingDir(
		matchers.AnyModelsRepo(),
		matchers.AnyModelsPullRequest(),
		AnyString())).ThenReturn(repoDir, nil)

	logger := logging.NewNoopCtxLogger(t)
	scope, _, _ := metrics.NewLoggingScope(logger, "atlantis")

	builder := events.NewProjectCommandBuilder(
		events.NewProjectCommandContextBuilder(&events.CommentParser{}),
		&config.ParserValidator{},
		&events.DefaultProjectFinder{},
		nil,
		workingDir,
		events.NewDefaultWorkingDirLocker(),
		valid.NewGlobalCfg("somedir"),
		&events.DefaultPendingPlanFinder{},
		false,
		"**/*.tf,**/*.tfvars,**/*.tfvars.json,**/terragrunt.hcl",
		logger,
		events.InfiniteProjectsPerPR,
	)

	ctx := &command.Context{
		RequestCtx: context.TODO(),
		HeadRepo:   models.Repo{},
		Pull:       models.PullRequest{},
		User:       models.User{},
		Log:        logging.NewNoopCtxLogger(t),
		Scope:      scope,
	}
	_, err = builder.BuildPlanCommands(ctx, &command.Comment{
		RepoRelDir:  ".",
		Flags:       nil,
		Name:        command.Plan,
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

	logger := logging.NewNoopCtxLogger(t)
	scope, _, _ := metrics.NewLoggingScope(logger, "atlantis")

	for _, c := range cases {
		t.Run(strings.Join(c.ExtraArgs, " "), func(t *testing.T) {
			RegisterMockTestingT(t)
			tmpDir, cleanup := DirStructure(t, map[string]interface{}{
				"main.tf": nil,
			})
			defer cleanup()

			workingDir := mocks.NewMockWorkingDir()
			When(workingDir.Clone(matchers.AnyLoggingLogger(), matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest(), AnyString())).ThenReturn(tmpDir, false, nil)
			When(workingDir.GetWorkingDir(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest(), AnyString())).ThenReturn(tmpDir, nil)
			vcsClient := vcsmocks.NewMockClient()
			When(vcsClient.GetModifiedFiles(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())).ThenReturn([]string{"main.tf"}, nil)

			builder := events.NewProjectCommandBuilder(
				events.NewProjectCommandContextBuilder(&events.CommentParser{}),
				&config.ParserValidator{},
				&events.DefaultProjectFinder{},
				vcsClient,
				workingDir,
				events.NewDefaultWorkingDirLocker(),
				valid.NewGlobalCfg("somedir"),
				&events.DefaultPendingPlanFinder{},
				false,
				"**/*.tf,**/*.tfvars,**/*.tfvars.json,**/terragrunt.hcl",
				logger,
				events.InfiniteProjectsPerPR,
			)

			var actCtxs []command.ProjectContext
			var err error
			actCtxs, err = builder.BuildPlanCommands(&command.Context{
				RequestCtx: context.TODO(),
				Log:        logger,
				Scope:      scope,
			}, &command.Comment{
				RepoRelDir: ".",
				Flags:      c.ExtraArgs,
				Name:       command.Plan,
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
  required_version = "%s0.12.8"
}
`

	atlantisYamlContent := `
version: 3
projects:
- dir: project1 # project1 uses the defaults
  terraform_version: v0.12.6
`

	exactSymbols := []string{"", "="}
	nonExactSymbols := []string{">", ">=", "<", "<=", "~="}

	type testCase struct {
		DirStructure  map[string]interface{}
		AtlantisYAML  string
		ModifiedFiles []string
		Exp           map[string][]int
	}

	testCases := make(map[string]testCase)

	for _, exactSymbol := range exactSymbols {
		testCases[fmt.Sprintf("exact version in terraform config using \"%s\"", exactSymbol)] = testCase{
			DirStructure: map[string]interface{}{
				"project1": map[string]interface{}{
					"main.tf": fmt.Sprintf(baseVersionConfig, exactSymbol),
				},
			},
			ModifiedFiles: []string{"project1/main.tf"},
			Exp: map[string][]int{
				"project1": {0, 12, 8},
			},
		}
	}

	for _, nonExactSymbol := range nonExactSymbols {
		testCases[fmt.Sprintf("non-exact version in terraform config using \"%s\"", nonExactSymbol)] = testCase{
			DirStructure: map[string]interface{}{
				"project1": map[string]interface{}{
					"main.tf": fmt.Sprintf(baseVersionConfig, nonExactSymbol),
				},
			},
			ModifiedFiles: []string{"project1/main.tf"},
			Exp: map[string][]int{
				"project1": nil,
			},
		}
	}

	// atlantis.yaml should take precedence over terraform config
	testCases["with project config and terraform config"] = testCase{
		DirStructure: map[string]interface{}{
			"project1": map[string]interface{}{
				"main.tf": fmt.Sprintf(baseVersionConfig, exactSymbols[0]),
			},
			config.AtlantisYAMLFilename: atlantisYamlContent,
		},
		ModifiedFiles: []string{"project1/main.tf", "project2/main.tf"},
		Exp: map[string][]int{
			"project1": {0, 12, 6},
		},
	}

	testCases["with project config only"] = testCase{
		DirStructure: map[string]interface{}{
			"project1": map[string]interface{}{
				"main.tf": nil,
			},
			config.AtlantisYAMLFilename: atlantisYamlContent,
		},
		ModifiedFiles: []string{"project1/main.tf"},
		Exp: map[string][]int{
			"project1": {0, 12, 6},
		},
	}

	testCases["neither project config or terraform config"] = testCase{
		DirStructure: map[string]interface{}{
			"project1": map[string]interface{}{
				"main.tf": nil,
			},
		},
		ModifiedFiles: []string{"project1/main.tf", "project2/main.tf"},
		Exp: map[string][]int{
			"project1": nil,
		},
	}

	testCases["project with different terraform config"] = testCase{
		DirStructure: map[string]interface{}{
			"project1": map[string]interface{}{
				"main.tf": fmt.Sprintf(baseVersionConfig, exactSymbols[0]),
			},
			"project2": map[string]interface{}{
				"main.tf": strings.Replace(fmt.Sprintf(baseVersionConfig, exactSymbols[0]), "0.12.8", "0.12.9", -1),
			},
		},
		ModifiedFiles: []string{"project1/main.tf", "project2/main.tf"},
		Exp: map[string][]int{
			"project1": {0, 12, 8},
			"project2": {0, 12, 9},
		},
	}

	logger := logging.NewNoopCtxLogger(t)
	scope, _, _ := metrics.NewLoggingScope(logger, "atlantis")

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			RegisterMockTestingT(t)

			tmpDir, cleanup := DirStructure(t, testCase.DirStructure)

			defer cleanup()
			vcsClient := vcsmocks.NewMockClient()
			When(vcsClient.GetModifiedFiles(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())).ThenReturn(testCase.ModifiedFiles, nil)

			workingDir := mocks.NewMockWorkingDir()
			When(workingDir.Clone(
				matchers.AnyLoggingLogger(),
				matchers.AnyModelsRepo(),
				matchers.AnyModelsPullRequest(),
				AnyString())).ThenReturn(tmpDir, false, nil)

			When(workingDir.GetWorkingDir(
				matchers.AnyModelsRepo(),
				matchers.AnyModelsPullRequest(),
				AnyString())).ThenReturn(tmpDir, nil)

			builder := events.NewProjectCommandBuilder(
				events.NewProjectCommandContextBuilder(&events.CommentParser{}),
				&config.ParserValidator{},
				&events.DefaultProjectFinder{},
				vcsClient,
				workingDir,
				events.NewDefaultWorkingDirLocker(),
				valid.NewGlobalCfg("somedir"),
				&events.DefaultPendingPlanFinder{},
				false,
				"**/*.tf,**/*.tfvars,**/*.tfvars.json,**/terragrunt.hcl",
				logger,
				events.InfiniteProjectsPerPR,
			)

			actCtxs, err := builder.BuildPlanCommands(
				&command.Context{
					RequestCtx: context.TODO(),
					Log:        logger,
					Scope:      scope,
				},
				&command.Comment{
					RepoRelDir: "",
					Flags:      nil,
					Name:       command.Plan,
				})

			Ok(t, err)
			Equals(t, len(testCase.Exp), len(actCtxs))
			for _, actCtx := range actCtxs {
				if testCase.Exp[actCtx.RepoRelDir] != nil {
					Assert(t, actCtx.TerraformVersion != nil, "TerraformVersion is nil.")
					Equals(t, testCase.Exp[actCtx.RepoRelDir], actCtx.TerraformVersion.Segments())
				} else {
					Assert(t, actCtx.TerraformVersion == nil, "TerraformVersion is supposed to be nil.")
				}
			}
		})
	}
}

func TestDefaultProjectCommandBuilder_WithPolicyCheckEnabled_BuildAutoplanCommand(t *testing.T) {
	RegisterMockTestingT(t)
	tmpDir, cleanup := DirStructure(t, map[string]interface{}{
		"main.tf": nil,
	})
	defer cleanup()

	logger := logging.NewNoopCtxLogger(t)
	scope, _, _ := metrics.NewLoggingScope(logger, "atlantis")

	workingDir := mocks.NewMockWorkingDir()
	When(workingDir.Clone(matchers.AnyLoggingLogger(), matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest(), AnyString())).ThenReturn(tmpDir, false, nil)
	vcsClient := vcsmocks.NewMockClient()
	When(vcsClient.GetModifiedFiles(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest())).ThenReturn([]string{"main.tf"}, nil)

	globalCfg := valid.NewGlobalCfg("somedir")
	commentParser := &events.CommentParser{}
	contextBuilder := wrappers.
		WrapProjectContext(events.NewProjectCommandContextBuilder(commentParser)).
		EnablePolicyChecks(commentParser)

	builder := events.NewProjectCommandBuilder(
		contextBuilder,
		&config.ParserValidator{},
		&events.DefaultProjectFinder{},
		vcsClient,
		workingDir,
		events.NewDefaultWorkingDirLocker(),
		globalCfg,
		&events.DefaultPendingPlanFinder{},
		false,
		"**/*.tf,**/*.tfvars,**/*.tfvars.json,**/terragrunt.hcl",
		logger,
		events.InfiniteProjectsPerPR,
	)

	ctxs, err := builder.BuildAutoplanCommands(&command.Context{
		PullRequestStatus: models.PullReqStatus{
			Mergeable: true,
		},
		RequestCtx: context.TODO(),
		Log:        logger,
		Scope:      scope,
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

	logger := logging.NewNoopCtxLogger(t)
	scope := tally.NewTestScope("test", nil)

	builder := events.NewProjectCommandBuilder(
		events.NewProjectCommandContextBuilder(&events.CommentParser{}),
		&config.ParserValidator{},
		&events.DefaultProjectFinder{},
		nil,
		workingDir,
		events.NewDefaultWorkingDirLocker(),
		valid.NewGlobalCfg("somedir"),
		&events.DefaultPendingPlanFinder{},
		false,
		"**/*.tf,**/*.tfvars,**/*.tfvars.json,**/terragrunt.hcl",
		logger,
		events.InfiniteProjectsPerPR,
	)

	ctxs, err := builder.BuildVersionCommands(
		&command.Context{
			RequestCtx: context.TODO(),
			Log:        logger,
			Scope:      scope,
		},
		&command.Comment{
			RepoRelDir:  "",
			Flags:       nil,
			Name:        command.Version,
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

func TestDefaultProjectCommandBuilder_BuildPolicyCheckCommands(t *testing.T) {
	testWorkingDirLocker := mockWorkingDirLocker{}
	tmpDir, cleanup := DirStructure(t, map[string]interface{}{
		"workspace1": map[string]interface{}{
			"project1": map[string]interface{}{
				"main.tf":          nil,
				"workspace.tfplan": nil,
			},
		},
	})
	defer cleanup()
	// Initialize git repos in each workspace so that the .tfplan files get
	// picked up.
	runCmd(t, filepath.Join(tmpDir, "workspace1"), "git", "init")
	testWorkingDir := mockWorkingDir{
		pullDir:    tmpDir,
		workingDir: tmpDir,
	}
	expectedProjects := []command.ProjectContext{
		{
			CommandName: command.PolicyCheck,
		},
	}
	testContextBuilder := mockContextBuilder{
		projects: expectedProjects,
	}
	builder := events.DefaultProjectCommandBuilder{
		ParserValidator:              &config.ParserValidator{},
		WorkingDir:                   testWorkingDir,
		WorkingDirLocker:             testWorkingDirLocker,
		GlobalCfg:                    valid.NewGlobalCfg("somedir"),
		PendingPlanFinder:            &events.DefaultPendingPlanFinder{},
		ProjectCommandContextBuilder: testContextBuilder,
	}
	commandCtx := &command.Context{
		Log:        logging.NewNoopCtxLogger(t),
		Scope:      tally.NewTestScope("atlantis", map[string]string{}),
		RequestCtx: context.Background(),
	}
	projects, err := builder.BuildPolicyCheckCommands(commandCtx)
	assert.NoError(t, err)
	assert.Equal(t, expectedProjects, projects)
}

func TestDefaultProjectCommandBuilder_BuildPolicyCheckCommands_TryLockPullError(t *testing.T) {
	testWorkingDirLocker := mockWorkingDirLocker{
		error: assert.AnError,
	}
	builder := events.DefaultProjectCommandBuilder{
		WorkingDirLocker: testWorkingDirLocker,
	}
	commandCtx := &command.Context{
		Log:        logging.NewNoopCtxLogger(t),
		Scope:      tally.NewTestScope("atlantis", map[string]string{}),
		RequestCtx: context.Background(),
	}
	projects, err := builder.BuildPolicyCheckCommands(commandCtx)
	assert.ErrorIs(t, err, assert.AnError)
	assert.Empty(t, projects)
}

func TestDefaultProjectCommandBuilder_BuildPolicyCheckCommands_GetPullDirError(t *testing.T) {
	testWorkingDir := mockWorkingDir{
		pullErr: assert.AnError,
	}
	testWorkingDirLocker := mockWorkingDirLocker{}
	builder := events.DefaultProjectCommandBuilder{
		WorkingDir:       testWorkingDir,
		WorkingDirLocker: testWorkingDirLocker,
	}
	commandCtx := &command.Context{
		Log:        logging.NewNoopCtxLogger(t),
		Scope:      tally.NewTestScope("atlantis", map[string]string{}),
		RequestCtx: context.Background(),
	}
	projects, err := builder.BuildPolicyCheckCommands(commandCtx)
	assert.ErrorIs(t, err, assert.AnError)
	assert.Empty(t, projects)
}

func TestDefaultProjectCommandBuilder_BuildPolicyCheckCommands_FindError(t *testing.T) {
	testWorkingDir := mockWorkingDir{}
	testWorkingDirLocker := mockWorkingDirLocker{}
	builder := events.DefaultProjectCommandBuilder{
		WorkingDir:        testWorkingDir,
		WorkingDirLocker:  testWorkingDirLocker,
		PendingPlanFinder: &events.DefaultPendingPlanFinder{},
	}
	commandCtx := &command.Context{
		Log:        logging.NewNoopCtxLogger(t),
		Scope:      tally.NewTestScope("atlantis", map[string]string{}),
		RequestCtx: context.Background(),
	}
	projects, err := builder.BuildPolicyCheckCommands(commandCtx)
	assert.ErrorIs(t, err, os.ErrNotExist)
	assert.Empty(t, projects)
}

func TestDefaultProjectCommandBuilder_BuildPolicyCheckCommands_GetWorkingDirErr(t *testing.T) {
	testWorkingDirLocker := mockWorkingDirLocker{}
	tmpDir, cleanup := DirStructure(t, map[string]interface{}{
		"workspace1": map[string]interface{}{
			"project1": map[string]interface{}{
				"main.tf":          nil,
				"workspace.tfplan": nil,
			},
		},
	})
	defer cleanup()
	// Initialize git repos in each workspace so that the .tfplan files get
	// picked up.
	runCmd(t, filepath.Join(tmpDir, "workspace1"), "git", "init")
	testWorkingDir := mockWorkingDir{
		pullDir:       tmpDir,
		workingDirErr: assert.AnError,
	}
	builder := events.DefaultProjectCommandBuilder{
		WorkingDir:        testWorkingDir,
		WorkingDirLocker:  testWorkingDirLocker,
		GlobalCfg:         valid.NewGlobalCfg("somedir"),
		PendingPlanFinder: &events.DefaultPendingPlanFinder{},
	}
	commandCtx := &command.Context{
		Log:        logging.NewNoopCtxLogger(t),
		Scope:      tally.NewTestScope("atlantis", map[string]string{}),
		RequestCtx: context.Background(),
	}
	projects, err := builder.BuildPolicyCheckCommands(commandCtx)
	assert.ErrorIs(t, err, assert.AnError)
	assert.Empty(t, projects)
}

type mockWorkingDirLocker struct {
	error error
}

func (l mockWorkingDirLocker) TryLock(_ string, _ int, _ string) (func(), error) {
	return func() {}, nil
}

func (l mockWorkingDirLocker) TryLockPull(_ string, _ int) (func(), error) {
	if l.error != nil {
		return func() {}, l.error
	}
	return func() {}, nil
}

type mockContextBuilder struct {
	projects []command.ProjectContext
}

func (b mockContextBuilder) BuildProjectContext(_ *command.Context, _ command.Name, _ valid.MergedProjectCfg, _ []string, _ string, _ *command.ContextFlags) []command.ProjectContext {
	return b.projects
}

type mockWorkingDir struct {
	pullDir       string
	workingDir    string
	pullErr       error
	workingDirErr error
}

func (w mockWorkingDir) GetPullDir(_ models.Repo, _ models.PullRequest) (string, error) {
	return w.pullDir, w.pullErr
}

func (w mockWorkingDir) GetWorkingDir(models.Repo, models.PullRequest, string) (string, error) {
	return w.workingDir, w.workingDirErr
}

func (w mockWorkingDir) HasDiverged(logging.Logger, string, models.Repo) bool {
	return false
}

func (w mockWorkingDir) Delete(models.Repo, models.PullRequest) error {
	return nil
}

func (w mockWorkingDir) DeleteForWorkspace(_ models.Repo, _ models.PullRequest, _ string) error {
	return nil
}

func (w mockWorkingDir) Clone(_ logging.Logger, _ models.Repo, _ models.PullRequest, _ string) (string, bool, error) {
	return "", false, nil
}
