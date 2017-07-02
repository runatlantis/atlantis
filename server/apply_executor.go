package server

import (
	"fmt"
	"os"
	"strings"

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

func (a *ApplyExecutor) execute(ctx *CommandContext, github *github.Client) {
	if a.concurrentRunLocker.TryLock(ctx.BaseRepo.FullName, ctx.Command.environment, ctx.Pull.Num) != true {
		ctx.Log.Info("run was locked by a concurrent run")
		github.CreateComment(ctx.BaseRepo, ctx.Pull, "This environment is currently locked by another command that is running for this pull request. Wait until command is complete and try again")
		return
	}
	defer a.concurrentRunLocker.Unlock(ctx.BaseRepo.FullName, ctx.Command.environment, ctx.Pull.Num)

	a.githubStatus.Update(ctx.BaseRepo, ctx.Pull, Pending, ApplyStep)
	res := a.setupAndApply(ctx)
	res.Command = Apply
	comment := a.githubCommentRenderer.render(res, ctx.Log.History.String(), ctx.Command.verbose)
	github.CreateComment(ctx.BaseRepo, ctx.Pull, comment)
}

func (a *ApplyExecutor) setupAndApply(ctx *CommandContext) ExecutionResult {
	if a.requireApproval {
		approved, res := a.isApproved(ctx)
		if !approved {
			return res
		}
	}

	repoDir, err := a.workspace.GetWorkspace(ctx)
	if err != nil {
		ctx.Log.Err(err.Error())
		a.githubStatus.Update(ctx.BaseRepo, ctx.Pull, Error, ApplyStep)
		return ExecutionResult{SetupError: GeneralError{errors.New("Workspace missing, please plan again")}}
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
		failure := "found 0 plans for that environment"
		ctx.Log.Warn(failure)
		a.githubStatus.Update(ctx.BaseRepo, ctx.Pull, Failure, ApplyStep)
		return ExecutionResult{SetupFailure: NoPlansFailure{}}
	}

	applyOutputs := []PathResult{}
	for _, plan := range plans {
		output := a.apply(ctx, repoDir, plan)
		output.Path = plan.LocalPath
		applyOutputs = append(applyOutputs, output)

	}
	a.githubStatus.UpdatePathResult(ctx, applyOutputs)
	return ExecutionResult{PathResults: applyOutputs}
}

func (a *ApplyExecutor) apply(ctx *CommandContext, repoDir string, plan models.Plan) PathResult {
	tfEnv := ctx.Command.environment
	lockAttempt, err := a.lockingClient.TryLock(plan.Project, tfEnv, ctx.Pull, ctx.User)
	if err != nil {
		return PathResult{
			Status: Error,
			Result: GeneralError{errors.Wrap(err, "trying acquire lock")},
		}
	}
	if lockAttempt.LockAcquired != true && lockAttempt.CurrLock.Pull.Num != ctx.Pull.Num {
		return PathResult{
			Status: Error,
			Result: GeneralError{fmt.Errorf("failed to acquire lock: lock held by pull request #%d", lockAttempt.CurrLock.Pull.Num)},
		}
	}

	// check if config file is found, if not we continue the run
	projectAbsolutePath := filepath.Dir(plan.LocalPath)
	var terraformApplyExtraArgs []string
	var config ProjectConfig
	if a.configReader.Exists(projectAbsolutePath) {
		ctx.Log.Info("Config file found in %s", projectAbsolutePath)
		config, err := a.configReader.Read(projectAbsolutePath)
		if err != nil {
			msg := fmt.Sprintf("Error reading config file: %v", err)
			ctx.Log.Err(msg)
			return PathResult{
				Status: Error,
				Result: GeneralError{errors.New(msg)},
			}
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
			msg := fmt.Sprintf("terraform init and environment commands failed. %s %v", outputs, err)
			ctx.Log.Err(msg)
			return PathResult{
				Status: Error,
				Result: GeneralError{errors.New(msg)},
			}
		}
		ctx.Log.Info("terraform init and environment commands ran successfully %s", outputs)
	}

	// if there are pre plan commands then run them
	if len(config.PrePlan.Commands) > 0 {
		preRunOutput, err := a.preRun.Start(config.PreApply.Commands, projectAbsolutePath, ctx.Command.environment, config.TerraformVersion)
		if err != nil {
			msg := fmt.Sprintf("pre run failed: %v", err)
			ctx.Log.Err(msg)
			return PathResult{
				Status: Error,
				Result: GeneralError{errors.New(msg)},
			}
		}
		ctx.Log.Info("Pre run output: \n%s", preRunOutput)
	}

	// need to get auth data from assumed role
	// todo: de-duplicate calls to assumeRole
	a.awsConfig.SessionName = ctx.User.Username
	awsSession, err := a.awsConfig.CreateSession()
	if err != nil {
		ctx.Log.Err(err.Error())
		return PathResult{
			Status: Error,
			Result: GeneralError{err},
		}
	}

	credVals, err := awsSession.Config.Credentials.Get()
	if err != nil {
		msg := fmt.Sprintf("failed to get assumed role credentials: %v", err)
		ctx.Log.Err(msg)
		return PathResult{
			Status: Error,
			Result: GeneralError{errors.New(msg)},
		}
	}

	ctx.Log.Info("running apply from %q", plan.Project.Path)
	tfApplyCmd := []string{"apply", "-no-color", plan.LocalPath}
	// append terraform arguments from config file
	tfApplyCmd = append(tfApplyCmd, terraformApplyExtraArgs...)
	terraformApplyCmdArgs, output, err := a.terraform.RunCommand(projectAbsolutePath, tfApplyCmd, []string{
		fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", credVals.AccessKeyID),
		fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", credVals.SecretAccessKey),
		fmt.Sprintf("AWS_SESSION_TOKEN=%s", credVals.SessionToken),
	})
	if err != nil {
		ctx.Log.Err("failed to apply: %v %s", err, output)
		return PathResult{
			Status: Failure,
			Result: ApplyFailure{Command: strings.Join(terraformApplyCmdArgs, " "), Output: output, ErrorMessage: err.Error()},
		}
	}

	return PathResult{
		Status: Success,
		Result: ApplySuccess{output},
	}
}

func (a *ApplyExecutor) isApproved(ctx *CommandContext) (bool, ExecutionResult) {
	ok, err := a.github.PullIsApproved(ctx.BaseRepo, ctx.Pull)
	if err != nil {
		msg := fmt.Sprintf("failed to determine if pull request was approved: %v", err)
		ctx.Log.Err(msg)
		a.githubStatus.Update(ctx.BaseRepo, ctx.Pull, Error, ApplyStep)
		return false, ExecutionResult{SetupError: GeneralError{errors.New(msg)}}
	}
	if !ok {
		ctx.Log.Info("pull request was not approved")
		a.githubStatus.Update(ctx.BaseRepo, ctx.Pull, Failure, ApplyStep)
		return false, ExecutionResult{SetupFailure: PullNotApprovedFailure{}}
	}
	return true, ExecutionResult{}
}
