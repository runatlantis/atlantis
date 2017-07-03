package server

import (
	"fmt"
	"os"

	"github.com/pkg/errors"

	"path/filepath"

	version "github.com/hashicorp/go-version"
	"github.com/hootsuite/atlantis/aws"
	"github.com/hootsuite/atlantis/github"
	"github.com/hootsuite/atlantis/locking"
	"github.com/hootsuite/atlantis/models"
	"github.com/hootsuite/atlantis/prerun"
	"github.com/hootsuite/atlantis/terraform"
)

type ApplyExecutor struct {
	github                *github.Client
	githubStatus          *GithubStatus
	awsConfig             *aws.Config
	terraform             *terraform.Client
	githubCommentRenderer *GithubCommentRenderer
	lockingClient         *locking.Client
	requireApproval       bool
	preRun                *prerun.PreRun
	configReader          *ConfigReader
	concurrentRunLocker   *ConcurrentRunLocker
	workspace             *Workspace
}

/** Result Types **/
type ApplyFailure struct {
	Command      string
	Output       string
	ErrorMessage string
}

func (a ApplyFailure) Template() *CompiledTemplate {
	return ApplyFailureTmpl
}

type ApplySuccess struct {
	Output string
}

func (a ApplySuccess) Template() *CompiledTemplate {
	return ApplySuccessTmpl
}

type PullNotApprovedFailure struct{}

func (p PullNotApprovedFailure) Template() *CompiledTemplate {
	return PullNotApprovedFailureTmpl
}

type NoPlansFailure struct{}

func (n NoPlansFailure) Template() *CompiledTemplate {
	return NoPlansFailureTmpl
}

// todo: why pass githbub.client here, just use the one on the struct
func (a *ApplyExecutor) execute(ctx *CommandContext, github *github.Client) {
	a.githubStatus.Update(ctx.BaseRepo, ctx.Pull, Pending, ApplyStep)
	res := a.setupAndApply(ctx)
	res.Command = Apply
	comment := a.githubCommentRenderer.render(res, ctx.Log.History.String(), ctx.Command.verbose)
	github.CreateComment(ctx.BaseRepo, ctx.Pull, comment)
}

func (a *ApplyExecutor) setupAndApply(ctx *CommandContext) CommandResponse {
	if a.concurrentRunLocker.TryLock(ctx.BaseRepo.FullName, ctx.Command.environment, ctx.Pull.Num) != true {
		return a.failureResponse(ctx,
			fmt.Sprintf("The %s environment is currently locked by another command that is running for this pull request. Wait until command is complete and try again.", ctx.Command.environment))
	}
	defer a.concurrentRunLocker.Unlock(ctx.BaseRepo.FullName, ctx.Command.environment, ctx.Pull.Num)

	if a.requireApproval {
		approved, err := a.github.PullIsApproved(ctx.BaseRepo, ctx.Pull)
		if err != nil {
			return a.errorResponse(ctx, errors.Wrap(err, "checking if pull request was approved"))
		}
		if !approved {
			return a.failureResponse(ctx, "Pull request must be approved before running apply.")
		}
	}

	repoDir, err := a.workspace.GetWorkspace(ctx)
	if err != nil {
		return a.failureResponse(ctx, "No workspace found. Did you run plan?")
	}

	// plans are stored at project roots by their environment names. We just need to find them
	var plans []models.Plan
	filepath.Walk(repoDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// if the plan is for the right env,
		if !info.IsDir() && info.Name() == ctx.Command.environment+".tfplan" {
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

	results := []ProjectResult{}
	for _, plan := range plans {
		result := a.apply(ctx, repoDir, plan)
		result.Path = plan.LocalPath
		results = append(results, result)
	}
	a.githubStatus.UpdatePathResult(ctx, results)
	return CommandResponse{ProjectResults: results}
}

func (a *ApplyExecutor) apply(ctx *CommandContext, repoDir string, plan models.Plan) ProjectResult {
	tfEnv := ctx.Command.environment
	lockAttempt, err := a.lockingClient.TryLock(plan.Project, tfEnv, ctx.Pull, ctx.User)
	if err != nil {
		return ProjectResult{Error: errors.Wrap(err, "acquiring lock")}
	}
	if lockAttempt.LockAcquired != true && lockAttempt.CurrLock.Pull.Num != ctx.Pull.Num {
		return ProjectResult{Failure: fmt.Sprintf(
			"This project is currently locked by #%d. The locking plan must be applied or discarded before future plans can execute.",
			lockAttempt.CurrLock.Pull.Num)}
	}

	// check if config file is found, if not we continue the run
	projectAbsolutePath := filepath.Dir(plan.LocalPath)
	var terraformApplyExtraArgs []string
	var config ProjectConfig
	if a.configReader.Exists(projectAbsolutePath) {
		ctx.Log.Info("Config file found in %s", projectAbsolutePath)
		config, err := a.configReader.Read(projectAbsolutePath)
		if err != nil {
			return ProjectResult{Error: err}
		}

		// add terraform arguments from project config
		terraformApplyExtraArgs = config.GetExtraArguments(ctx.Command.commandType.String())
	}

	// check if terraform version is >= 0.9.0
	terraformVersion := a.terraform.Version()
	if config.TerraformVersion != nil {
		terraformVersion = config.TerraformVersion
	}
	constraints, _ := version.NewConstraint(">= 0.9.0")
	if constraints.Check(terraformVersion) {
		// run terraform init and environment
		outputs, err := a.terraform.RunInitAndEnv(projectAbsolutePath, tfEnv, config.GetExtraArguments("init"))
		if err != nil {
			return ProjectResult{Error: err}
		}
		ctx.Log.Info("terraform init and environment commands ran successfully %s", outputs)
	}

	// if there are pre plan commands then run them
	if len(config.PreApply.Commands) > 0 {
		preRunOutput, err := a.preRun.Start(config.PreApply.Commands, projectAbsolutePath, ctx.Command.environment, config.TerraformVersion)
		if err != nil {
			return ProjectResult{Error: err}
		}
		ctx.Log.Info("Pre run output: \n%s", preRunOutput)
	}

	// need to get auth data from assumed role
	// todo: de-duplicate calls to assumeRole
	a.awsConfig.SessionName = ctx.User.Username
	awsSession, err := a.awsConfig.CreateSession()
	if err != nil {
		return ProjectResult{Error: err}
	}

	credVals, err := awsSession.Config.Credentials.Get()
	if err != nil {
		err = errors.Wrap(err, "getting assumed role credentials")
		ctx.Log.Err(err.Error())
		return ProjectResult{Error: err}
	}

	ctx.Log.Info("running apply from %q", plan.Project.Path)
	tfApplyCmd := []string{"apply", "-no-color", plan.LocalPath}
	// append terraform arguments from config file
	tfApplyCmd = append(tfApplyCmd, terraformApplyExtraArgs...)
	output, err := a.terraform.RunCommand(projectAbsolutePath, tfApplyCmd, []string{
		fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", credVals.AccessKeyID),
		fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", credVals.SecretAccessKey),
		fmt.Sprintf("AWS_SESSION_TOKEN=%s", credVals.SessionToken),
	})
	if err != nil {
		return ProjectResult{Error: fmt.Errorf("%s\n%s", err.Error(), output)}
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
