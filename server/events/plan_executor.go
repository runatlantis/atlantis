package events

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	version "github.com/hashicorp/go-version"
	"github.com/hootsuite/atlantis/server/github"
	"github.com/hootsuite/atlantis/server/locking"
	"github.com/hootsuite/atlantis/server/models"
	"github.com/hootsuite/atlantis/server/run"
	"github.com/hootsuite/atlantis/server/terraform"
	"github.com/pkg/errors"
)

//go:generate pegomock generate --use-experimental-model-gen --package mocks -o mocks/mock_lock_url_generator.go LockURLGenerator

type LockURLGenerator interface {
	// SetLockURL takes a function that given a lock id, will return a url
	// to view that lock
	SetLockURL(func(id string) (url string))
}

// atlantisUserTFVar is the name of the variable we execute terraform
// with containing the github username of who is running the command
const atlantisUserTFVar = "atlantis_user"

// PlanExecutor handles everything related to running terraform plan
// including integration with S3, Terraform, and GitHub
type PlanExecutor struct {
	Github       github.Client
	Terraform    *terraform.Client
	Locker       locking.Locker
	LockURL      func(id string) (url string)
	Run          *run.Run
	ConfigReader *ConfigReader
	Workspace    Workspace
}

type PlanSuccess struct {
	TerraformOutput string
	LockURL         string
}

func (p *PlanExecutor) SetLockURL(f func(id string) (url string)) {
	p.LockURL = f
}

func (p *PlanExecutor) Execute(ctx *CommandContext) CommandResponse {
	// figure out what projects have been modified so we know where to run plan
	modifiedFiles, err := p.Github.GetModifiedFiles(ctx.BaseRepo, ctx.Pull)
	if err != nil {
		return CommandResponse{Error: errors.Wrap(err, "getting modified files")}
	}
	ctx.Log.Info("found %d files modified in this pull request", len(modifiedFiles))

	modifiedTerraformFiles := p.filterToTerraform(modifiedFiles)
	if len(modifiedTerraformFiles) == 0 {
		return CommandResponse{Failure: "No Terraform files were modified."}
	}
	ctx.Log.Info("filtered modified files to %d non-module .tf files: %v", len(modifiedTerraformFiles), modifiedTerraformFiles)

	projects := p.ModifiedProjects(ctx.BaseRepo.FullName, modifiedTerraformFiles)
	var paths []string
	for _, p := range projects {
		paths = append(paths, p.Path)
	}
	ctx.Log.Info("based on files modified, determined we have %d modified project(s) at path(s): %v", len(projects), strings.Join(paths, ", "))

	cloneDir, err := p.Workspace.Clone(ctx.Log, ctx.BaseRepo, ctx.HeadRepo, ctx.Pull, ctx.Command.Environment)
	if err != nil {
		return CommandResponse{Error: err}
	}

	results := []ProjectResult{}
	for _, project := range projects {
		ctx.Log.Info("running plan for project at path %q", project.Path)
		result := p.plan(ctx, cloneDir, project)
		result.Path = project.Path
		results = append(results, result)
	}
	return CommandResponse{ProjectResults: results}
}

// plan runs the steps necessary to run `terraform plan`. If there is an error, the error message will be encapsulated in error
// and the GeneratePlanResponse struct will also contain the full log including the error
func (p *PlanExecutor) plan(ctx *CommandContext, repoDir string, project models.Project) ProjectResult {
	tfEnv := ctx.Command.Environment
	lockAttempt, err := p.Locker.TryLock(project, tfEnv, ctx.Pull, ctx.User)
	if err != nil {
		return ProjectResult{Error: errors.Wrap(err, "acquiring lock")}
	}
	if lockAttempt.LockAcquired == false && lockAttempt.CurrLock.Pull.Num != ctx.Pull.Num {
		return ProjectResult{Failure: fmt.Sprintf(
			"This project is currently locked by #%d. The locking plan must be applied or discarded before future plans can execute.",
			lockAttempt.CurrLock.Pull.Num)}
	}
	ctx.Log.Info("acquired lock with id %q", lockAttempt.LockKey)

	// check if config file is found, if not we continue the run
	var config ProjectConfig
	absolutePath := filepath.Join(repoDir, project.Path)
	var planExtraArgs []string
	if p.ConfigReader.Exists(absolutePath) {
		config, err = p.ConfigReader.Read(absolutePath)
		if err != nil {
			return ProjectResult{Error: err}
		}
		ctx.Log.Info("parsed atlantis config file in %q", absolutePath)
		planExtraArgs = config.GetExtraArguments(ctx.Command.Name.String())
	}

	// check if terraform version is >= 0.9.0
	terraformVersion := p.Terraform.Version()
	if config.TerraformVersion != nil {
		terraformVersion = config.TerraformVersion
	}
	constraints, _ := version.NewConstraint(">= 0.9.0")
	if constraints.Check(terraformVersion) {
		ctx.Log.Info("determined that we are running terraform with version >= 0.9.0. Running version %s", terraformVersion)
		_, err := p.Terraform.RunInitAndEnv(ctx.Log, absolutePath, tfEnv, config.GetExtraArguments("init"), terraformVersion)
		if err != nil {
			return ProjectResult{Error: err}
		}
	} else {
		ctx.Log.Info("determined that we are running terraform with version < 0.9.0. Running version %s", terraformVersion)
		terraformGetCmd := append([]string{"get", "-no-color"}, config.GetExtraArguments("get")...)
		_, err := p.Terraform.RunCommandWithVersion(ctx.Log, absolutePath, terraformGetCmd, terraformVersion, tfEnv)
		if err != nil {
			return ProjectResult{Error: err}
		}
	}

	// if there are pre plan commands then run them
	if len(config.PrePlan.Commands) > 0 {
		_, err := p.Run.Execute(ctx.Log, config.PrePlan.Commands, absolutePath, tfEnv, terraformVersion, "pre_plan")
		if err != nil {
			return ProjectResult{Error: errors.Wrap(err, "running pre plan commands")}
		}
	}

	// Run terraform plan
	planFile := filepath.Join(repoDir, project.Path, fmt.Sprintf("%s.tfplan", tfEnv))
	userVar := fmt.Sprintf("%s=%s", atlantisUserTFVar, ctx.User.Username)
	tfPlanCmd := append(append([]string{"plan", "-refresh", "-no-color", "-out", planFile, "-var", userVar}, planExtraArgs...), ctx.Command.Flags...)

	// check if env/{environment}.tfvars exist
	tfEnvFileName := filepath.Join("env", tfEnv+".tfvars")
	if _, err := os.Stat(filepath.Join(repoDir, project.Path, tfEnvFileName)); err == nil {
		tfPlanCmd = append(tfPlanCmd, "-var-file", tfEnvFileName)
	}
	output, err := p.Terraform.RunCommandWithVersion(ctx.Log, filepath.Join(repoDir, project.Path), tfPlanCmd, terraformVersion, tfEnv)
	if err != nil {
		// plan failed so unlock the state
		if _, err := p.Locker.Unlock(lockAttempt.LockKey); err != nil {
			ctx.Log.Err("error unlocking state: %v", err)
		}
		return ProjectResult{Error: fmt.Errorf("%s\n%s", err.Error(), output)}
	}
	ctx.Log.Info("plan succeeded")

	// if there are post plan commands then run them
	if len(config.PostPlan.Commands) > 0 {
		_, err := p.Run.Execute(ctx.Log, config.PostPlan.Commands, absolutePath, tfEnv, terraformVersion, "post_plan")
		if err != nil {
			return ProjectResult{Error: errors.Wrap(err, "running post plan commands")}
		}
	}

	return ProjectResult{
		PlanSuccess: &PlanSuccess{
			TerraformOutput: output,
			LockURL:         p.LockURL(lockAttempt.LockKey),
		},
	}
}

func (p *PlanExecutor) filterToTerraform(files []string) []string {
	var out []string
	for _, fileName := range files {
		if !p.isInExcludeList(fileName) && strings.Contains(fileName, ".tf") {
			out = append(out, fileName)
		}
	}
	return out
}

func (p *PlanExecutor) isInExcludeList(fileName string) bool {
	return strings.Contains(fileName, "terraform.tfstate") || strings.Contains(fileName, "terraform.tfstate.backup") || strings.Contains(fileName, "_modules") || strings.Contains(fileName, "modules")
}

// ModifiedProjects returns the list of Terraform projects that have been changed due to the
// modified files
func (p *PlanExecutor) ModifiedProjects(repoFullName string, modifiedFiles []string) []models.Project {
	var projects []models.Project
	seenPaths := make(map[string]bool)
	for _, modifiedFile := range modifiedFiles {
		path := p.getProjectPath(modifiedFile)
		if _, ok := seenPaths[path]; !ok {
			projects = append(projects, models.NewProject(repoFullName, path))
			seenPaths[path] = true
		}
	}
	return projects
}

// getProjectPath returns the path to the project relative to the repo root
// if the project is at the root returns "."
func (p *PlanExecutor) getProjectPath(modifiedFilePath string) string {
	dir := path.Dir(modifiedFilePath)
	if path.Base(dir) == "env" {
		// if the modified file was inside an env/ directory, we treat this specially and
		// run plan one level up
		return path.Dir(dir)
	}
	return dir
}
