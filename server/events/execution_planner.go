package events

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/runtime"
	"github.com/runatlantis/atlantis/server/events/yaml"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
	"github.com/runatlantis/atlantis/server/logging"
)

const PlanStageName = "plan"
const ApplyStageName = "apply"
const AtlantisYAMLFilename = "atlantis.yaml"

type ExecutionPlanner struct {
	TerraformExecutor TerraformExec
	DefaultTFVersion  *version.Version
	ParserValidator   *yaml.ParserValidator
	ProjectFinder     ProjectFinder
}

type TerraformExec interface {
	RunCommandWithVersion(log *logging.SimpleLogger, path string, args []string, v *version.Version, workspace string) (string, error)
}

func (s *ExecutionPlanner) BuildPlanStage(log *logging.SimpleLogger, repoDir string, workspace string, relProjectPath string, extraCommentArgs []string, username string) (runtime.PlanStage, error) {
	defaults := s.defaultPlanSteps(log, repoDir, workspace, relProjectPath, extraCommentArgs, username)
	steps, err := s.buildStage(PlanStageName, log, repoDir, workspace, relProjectPath, extraCommentArgs, username, defaults)
	if err != nil {
		return runtime.PlanStage{}, err
	}
	return runtime.PlanStage{
		Steps:       steps,
		Workspace:   workspace,
		ProjectPath: relProjectPath,
	}, nil
}

func (s *ExecutionPlanner) BuildAutoplanStages(log *logging.SimpleLogger, repoFullName string, repoDir string, username string, modifiedFiles []string) ([]runtime.PlanStage, error) {
	// If there is an atlantis.yaml
	// -> Get modified files from pull request.
	// -> For each project, if autoplan == true && files match
	// ->->  Build plan stage for that project.
	// Else
	// -> Get modified files
	// -> For each modified project use default plan stage.
	config, err := s.ParserValidator.ReadConfig(repoDir)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	// If there is no config file, then we try to plan for each project that
	// was modified in the pull request.
	if os.IsNotExist(err) {
		projects := s.ProjectFinder.DetermineProjects(log, modifiedFiles, repoFullName, repoDir)
		var stages []runtime.PlanStage
		for _, p := range projects {
			// NOTE: we use the default workspace because we don't know about
			// other workspaces. If users want to plan for other workspaces they
			// need to use a config file.
			steps := s.defaultPlanSteps(log, repoDir, models.DefaultWorkspace, p.Path, nil, username)
			stages = append(stages, runtime.PlanStage{
				Steps:       steps,
				Workspace:   models.DefaultWorkspace,
				ProjectPath: p.Path,
			})
		}
		return stages, nil
	}

	// Else we run plan according to the config file.
	var stages []runtime.PlanStage
	for _, p := range config.Projects {
		if s.shouldAutoplan(p.Autoplan, modifiedFiles) {
			// todo
			stages = append(stages)
		}
	}
	return stages, nil
}

func (s *ExecutionPlanner) shouldAutoplan(autoplan valid.Autoplan, modifiedFiles []string) bool {
	return true
}

func (s *ExecutionPlanner) getSteps() {

}

func (s *ExecutionPlanner) BuildApplyStage(log *logging.SimpleLogger, repoDir string, workspace string, relProjectPath string, extraCommentArgs []string, username string) (*runtime.ApplyStage, error) {
	defaults := s.defaultApplySteps(log, repoDir, workspace, relProjectPath, extraCommentArgs, username)
	steps, err := s.buildStage(ApplyStageName, log, repoDir, workspace, relProjectPath, extraCommentArgs, username, defaults)
	if err != nil {
		return nil, err
	}
	return &runtime.ApplyStage{
		Steps: steps,
	}, nil
}

func (s *ExecutionPlanner) buildStage(stageName string, log *logging.SimpleLogger, repoDir string, workspace string, relProjectPath string, extraCommentArgs []string, username string, defaults []runtime.Step) ([]runtime.Step, error) {
	config, err := s.ParserValidator.ReadConfig(repoDir)

	// If there's no config file, use defaults.
	if os.IsNotExist(err) {
		log.Info("no %s file found––continuing with defaults", AtlantisYAMLFilename)
		return defaults, nil
	}

	if err != nil {
		return nil, err
	}

	// Get this project's configuration.
	for _, p := range config.Projects {
		if p.Dir == relProjectPath && p.Workspace == workspace {
			workflowNamePtr := p.Workflow

			// If they didn't specify a workflow, use the default.
			if workflowNamePtr == nil {
				log.Info("no %s workflow set––continuing with defaults", AtlantisYAMLFilename)
				return defaults, nil
			}

			// If they did specify a workflow, find it.
			workflowName := *workflowNamePtr
			workflow, exists := config.Workflows[workflowName]
			if !exists {
				return nil, fmt.Errorf("no workflow with key %q defined", workflowName)
			}

			// We have a workflow defined, so now we need to build it.
			meta := s.buildMeta(log, repoDir, workspace, relProjectPath, extraCommentArgs, username)
			var steps []runtime.Step
			var stepsConfig []valid.Step
			if stageName == PlanStageName {
				stepsConfig = workflow.Plan.Steps
			} else {
				stepsConfig = workflow.Apply.Steps
			}
			for _, stepConfig := range stepsConfig {
				var step runtime.Step
				switch stepConfig.StepName {
				case "init":
					step = &runtime.InitStep{
						Meta:      meta,
						ExtraArgs: stepConfig.ExtraArgs,
					}
				case "plan":
					step = &runtime.PlanStep{
						Meta:      meta,
						ExtraArgs: stepConfig.ExtraArgs,
					}
				case "apply":
					step = &runtime.ApplyStep{
						Meta:      meta,
						ExtraArgs: stepConfig.ExtraArgs,
					}
				case "run":
					step = &runtime.RunStep{
						Meta:     meta,
						Commands: stepConfig.RunCommand,
					}
				}
				steps = append(steps, step)
			}
			return steps, nil
		}
	}
	// They haven't defined this project, use the default workflow.
	log.Info("no project with dir %q and workspace %q defined; continuing with defaults", relProjectPath, workspace)
	return defaults, nil
}

func (s *ExecutionPlanner) buildMeta(log *logging.SimpleLogger, repoDir string, workspace string, relProjectPath string, extraCommentArgs []string, username string) runtime.StepMeta {
	return runtime.StepMeta{
		Log:                   log,
		Workspace:             workspace,
		AbsolutePath:          filepath.Join(repoDir, relProjectPath),
		DirRelativeToRepoRoot: relProjectPath,
		// If there's no config then we should use the default tf version.
		TerraformVersion:  s.DefaultTFVersion,
		TerraformExecutor: s.TerraformExecutor,
		ExtraCommentArgs:  extraCommentArgs,
		Username:          username,
	}
}

func (s *ExecutionPlanner) defaultPlanSteps(log *logging.SimpleLogger, repoDir string, workspace string, relProjectPath string, extraCommentArgs []string, username string) []runtime.Step {
	meta := s.buildMeta(log, repoDir, workspace, relProjectPath, extraCommentArgs, username)
	return []runtime.Step{
		&runtime.InitStep{
			ExtraArgs: nil,
			Meta:      meta,
		},
		&runtime.PlanStep{
			ExtraArgs: nil,
			Meta:      meta,
		},
	}
}
func (s *ExecutionPlanner) defaultApplySteps(log *logging.SimpleLogger, repoDir string, workspace string, relProjectPath string, extraCommentArgs []string, username string) []runtime.Step {
	meta := s.buildMeta(log, repoDir, workspace, relProjectPath, extraCommentArgs, username)
	return []runtime.Step{
		&runtime.ApplyStep{
			ExtraArgs: nil,
			Meta:      meta,
		},
	}
}
