package server

import (
	"fmt"
	"os"

	"github.com/pkg/errors"

	"path/filepath"

	version "github.com/hashicorp/go-version"
	"github.com/hootsuite/atlantis/github"
	"github.com/hootsuite/atlantis/locking"
	"github.com/hootsuite/atlantis/models"
	"github.com/hootsuite/atlantis/run"
	"github.com/hootsuite/atlantis/terraform"
)

type ApplyExecutor struct {
	github                *github.Client
	githubStatus          *GithubStatus
	terraform             *terraform.Client
	githubCommentRenderer *GithubCommentRenderer
	lockingClient         *locking.Client
	requireApproval       bool
	run                   *run.Run
	configReader          *ConfigReader
	concurrentRunLocker   *ConcurrentRunLocker
	workspace             *Workspace
}

func (a *ApplyExecutor) execute(ctx *CommandContext) {
	a.githubStatus.Update(ctx.BaseRepo, ctx.Pull, Pending, ApplyStep)
	res := a.setupAndApply(ctx)
	res.Command = Apply
	comment := a.githubCommentRenderer.render(res, ctx.Log.History.String(), ctx.Command.Verbose)
	a.github.CreateComment(ctx.BaseRepo, ctx.Pull, comment)
}

func (a *ApplyExecutor) setupAndApply(ctx *CommandContext) CommandResponse {
	if a.concurrentRunLocker.TryLock(ctx.BaseRepo.FullName, ctx.Command.Environment, ctx.Pull.Num) != true {
		return a.failureResponse(ctx,
			fmt.Sprintf("The %s environment is currently locked by another command that is running for this pull request. Wait until command is complete and try again.", ctx.Command.Environment))
	}
	defer a.concurrentRunLocker.Unlock(ctx.BaseRepo.FullName, ctx.Command.Environment, ctx.Pull.Num)

	if a.requireApproval {
		approved, err := a.github.PullIsApproved(ctx.BaseRepo, ctx.Pull)
		if err != nil {
			return a.errorResponse(ctx, errors.Wrap(err, "checking if pull request was approved"))
		}
		if !approved {
			return a.failureResponse(ctx, "Pull request must be approved before running apply.")
		}
		ctx.Log.Info("confirmed pull request was approved")
	}

	repoDir, err := a.workspace.GetWorkspace(ctx)
	if err != nil {
		return a.failureResponse(ctx, "No workspace found. Did you run plan?")
	}
	ctx.Log.Info("found workspace in %q", repoDir)

	// plans are stored at project roots by their environment names. We just need to find them
	var plans []models.Plan
	filepath.Walk(repoDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// if the plan is for the right env,
		if !info.IsDir() && info.Name() == ctx.Command.Environment+".tfplan" {
			rel, _ := filepath.Rel(repoDir, filepath.Dir(path))
			plans = append(plans, models.Plan{
				Project:   models.NewProject(ctx.BaseRepo.FullName, rel),
				LocalPath: path,
			})
		}
		return nil
	})
	if len(plans) == 0 {
		return a.failureResponse(ctx, "No plans found for that environment.")
	}
	var paths []string
	for _, p := range plans {
		paths = append(paths, p.LocalPath)
	}
	ctx.Log.Info("found %d plan(s) in our workspace: %v", len(plans), paths)

	results := []ProjectResult{}
	for _, plan := range plans {
		ctx.Log.Info("running apply for project at path %q", plan.Project.Path)
		result := a.apply(ctx, repoDir, plan)
		result.Path = plan.LocalPath
		results = append(results, result)
	}
	a.githubStatus.UpdateProjectResult(ctx, results)
	return CommandResponse{ProjectResults: results}
}

func (a *ApplyExecutor) apply(ctx *CommandContext, repoDir string, plan models.Plan) ProjectResult {
	tfEnv := ctx.Command.Environment
	lockAttempt, err := a.lockingClient.TryLock(plan.Project, tfEnv, ctx.Pull, ctx.User)
	if err != nil {
		return ProjectResult{Error: errors.Wrap(err, "acquiring lock")}
	}
	if lockAttempt.LockAcquired != true && lockAttempt.CurrLock.Pull.Num != ctx.Pull.Num {
		return ProjectResult{Failure: fmt.Sprintf(
			"This project is currently locked by #%d. The locking plan must be applied or discarded before future plans can execute.",
			lockAttempt.CurrLock.Pull.Num)}
	}
	ctx.Log.Info("acquired lock with id %q", lockAttempt.LockKey)

	// check if config file is found, if not we continue the run
	absolutePath := filepath.Dir(plan.LocalPath)
	var applyExtraArgs []string
	var config ProjectConfig
	if a.configReader.Exists(absolutePath) {
		config, err = a.configReader.Read(absolutePath)
		if err != nil {
			return ProjectResult{Error: err}
		}
		ctx.Log.Info("parsed atlantis config file in %q", absolutePath)
		applyExtraArgs = config.GetExtraArguments(ctx.Command.Name.String())
	}

	// check if terraform version is >= 0.9.0
	terraformVersion := a.terraform.Version()
	if config.TerraformVersion != nil {
		terraformVersion = config.TerraformVersion
	}
	constraints, _ := version.NewConstraint(">= 0.9.0")
	if constraints.Check(terraformVersion) {
		ctx.Log.Info("determined that we are running terraform with version >= 0.9.0. Running version %s", terraformVersion)
		_, err := a.terraform.RunInitAndEnv(ctx.Log, absolutePath, tfEnv, config.GetExtraArguments("init"), terraformVersion)
		if err != nil {
			return ProjectResult{Error: err}
		}
	}

	// if there are pre apply commands then run them
	if len(config.PreApply.Commands) > 0 {
		_, err := a.run.Execute(ctx.Log, config.PreApply.Commands, absolutePath, tfEnv, terraformVersion, "pre_apply")
		if err != nil {
			return ProjectResult{Error: errors.Wrap(err, "running pre apply commands")}
		}
	}

	tfApplyCmd := append(append([]string{"apply", "-no-color", plan.LocalPath}, applyExtraArgs...), ctx.Command.Flags...)
	output, err := a.terraform.RunCommandWithVersion(ctx.Log, absolutePath, tfApplyCmd, terraformVersion)
	if err != nil {
		return ProjectResult{Error: fmt.Errorf("%s\n%s", err.Error(), output)}
	}
	ctx.Log.Info("apply succeeded")

	// if there are post apply commands then run them
	if len(config.PostApply.Commands) > 0 {
		_, err := a.run.Execute(ctx.Log, config.PostApply.Commands, absolutePath, tfEnv, terraformVersion, "post_apply")
		if err != nil {
			return ProjectResult{Error: errors.Wrap(err, "running post apply commands")}
		}
	}

	return ProjectResult{ApplySuccess: output}
}

func (a *ApplyExecutor) failureResponse(ctx *CommandContext, msg string) CommandResponse {
	ctx.Log.Warn(msg)
	a.githubStatus.Update(ctx.BaseRepo, ctx.Pull, Failure, ApplyStep)
	return CommandResponse{Failure: msg}
}

func (a *ApplyExecutor) errorResponse(ctx *CommandContext, err error) CommandResponse {
	ctx.Log.Err(err.Error())
	a.githubStatus.Update(ctx.BaseRepo, ctx.Pull, Error, ApplyStep)
	return CommandResponse{Error: err}
}
