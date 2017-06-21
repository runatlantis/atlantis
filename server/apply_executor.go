package server

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"path/filepath"

	"github.com/hootsuite/atlantis/locking"
	"github.com/hootsuite/atlantis/plan"
	"strconv"
)

type ApplyExecutor struct {
	github                *GithubClient
	awsConfig             *AWSConfig
	scratchDir            string
	sshKey                string
	terraform             *TerraformClient
	githubCommentRenderer *GithubCommentRenderer
	lockingClient         *locking.Client
	requireApproval       bool
	planStorage           plan.Backend
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

func (a *ApplyExecutor) execute(ctx *CommandContext, github *GithubClient) {
	a.github.UpdateStatus(ctx.Repo, ctx.Pull, Pending, ApplyStep)
	res := a.setupAndApply(ctx)
	res.Command = Apply
	comment := a.githubCommentRenderer.render(res, ctx.Log.History.String(), ctx.Command.verbose)
	github.CreateComment(ctx, comment)
}

func (a *ApplyExecutor) setupAndApply(ctx *CommandContext) ExecutionResult {
	if approved, res := a.isApproved(ctx); !approved {
		return res
	}

	// todo: reclone repo and switch branch, don't assume it's already there
	repoDir := filepath.Join(a.scratchDir, ctx.Repo.FullName, strconv.Itoa(ctx.Pull.Num))
	plans, err := a.planStorage.CopyPlans(repoDir, ctx.Repo.FullName, ctx.Command.environment, ctx.Pull.Num)
	if err != nil {
		errMsg := fmt.Sprintf("failed to get plans: %s", err)
		ctx.Log.Err(errMsg)
		a.github.UpdateStatus(ctx.Repo, ctx.Pull, Error, ApplyStep)
		return ExecutionResult{SetupError: GeneralError{errors.New(errMsg)}}
	}
	if len(plans) == 0 {
		failure := "found 0 plans for this pull request"
		ctx.Log.Warn(failure)
		a.github.UpdateStatus(ctx.Repo, ctx.Pull, Failure, ApplyStep)
		return ExecutionResult{SetupFailure: NoPlansFailure{}}
	}

	applyOutputs := []PathResult{}
	for _, plan := range plans {
		output := a.apply(ctx, repoDir, plan)
		output.Path = plan.LocalPath
		applyOutputs = append(applyOutputs, output)

	}
	a.updateGithubStatus(ctx, applyOutputs)
	return ExecutionResult{PathResults: applyOutputs}
}

func (a *ApplyExecutor) apply(ctx *CommandContext, repoDir string, plan plan.Plan) PathResult {
	var config Config
	var remoteStatePath string
	// check if config file is found, if not we continue the run
	projectAbsolutePath := filepath.Dir(plan.LocalPath)
	if config.Exists(projectAbsolutePath) {
		ctx.Log.Info("Config file found in %s", projectAbsolutePath)
		err := config.Read(projectAbsolutePath)
		if err != nil {
			msg := fmt.Sprintf("Error reading config file: %v", err)
			ctx.Log.Err(msg)
			return PathResult{
				Status: Error,
				Result: GeneralError{errors.New(msg)},
			}
		}
		// need to use the remote state path and backend to do remote configure
		err = PreRun(&config, ctx.Log, projectAbsolutePath, ctx.Command)
		if err != nil {
			msg := fmt.Sprintf("pre run failed: %v", err)
			ctx.Log.Err(msg)
			return PathResult{
				Status: Error,
				Result: GeneralError{errors.New(msg)},
			}
		}
		// check if terraform version is specified in config
		if config.TerraformVersion != "" {
			a.terraform.tfExecutableName = "terraform" + config.TerraformVersion
		} else {
			a.terraform.tfExecutableName = "terraform"
		}
	}

	// NOTE: THIS CODE IS TO SUPPORT TERRAFORM PROJECTS THAT AREN'T USING ATLANTIS CONFIG FILE.
	if config.StashPath == "" {
		// configure remote state
		statePath, err := a.terraform.ConfigureRemoteState(ctx.Log, repoDir, plan.Project, ctx.Command.environment, a.sshKey)
		if err != nil {
			msg := fmt.Sprintf("failed to set up remote state: %v", err)
			ctx.Log.Err(msg)
			return PathResult{
				Status: Error,
				Result: GeneralError{errors.New(msg)},
			}
		}
		remoteStatePath = statePath
	} else {
		// use state path from config file
		remoteStatePath = generateStatePath(config.StashPath, ctx.Command.environment)
	}

	if remoteStatePath != "" {
		tfEnv := ctx.Command.environment
		if tfEnv == "" {
			tfEnv = "default"
		}

		lockAttempt, err := a.lockingClient.TryLock(plan.Project, tfEnv, ctx.Pull, ctx.User)
		if err != nil {
			return PathResult{
				Status: Error,
				Result: GeneralError{fmt.Errorf("failed to acquire lock: %s", err)},
			}
		}
		if lockAttempt.LockAcquired != true && lockAttempt.CurrLock.Pull.Num != ctx.Pull.Num {
			return PathResult{
				Status: Error,
				Result: GeneralError{fmt.Errorf("failed to acquire lock: lock held by pull request #%d", lockAttempt.CurrLock.Pull.Num)},
			}
		}
	}

	// need to get auth data from assumed role
	// todo: de-duplicate calls to assumeRole
	a.awsConfig.AWSSessionName = ctx.User.Username
	awsSession, err := a.awsConfig.CreateAWSSession()
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
	terraformApplyCmdArgs, output, err := a.terraform.RunTerraformCommand(projectAbsolutePath, []string{"apply", "-no-color", plan.LocalPath}, []string{
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

	// clean up, delete local plan file
	os.Remove(plan.LocalPath) // swallow errors, okay if we failed to delete
	return PathResult{
		Status: Success,
		Result: ApplySuccess{output},
	}
}

func (a *ApplyExecutor) updateGithubStatus(ctx *CommandContext, pathResults []PathResult) {
	var statuses []Status
	for _, p := range pathResults {
		statuses = append(statuses, p.Status)
	}
	worst := WorstStatus(statuses)
	a.github.UpdateStatus(ctx.Repo, ctx.Pull, worst, ApplyStep)
}

func (a *ApplyExecutor) isApproved(ctx *CommandContext) (bool, ExecutionResult) {
	if !a.requireApproval {
		return false, ExecutionResult{}
	}

	ok, err := a.github.PullIsApproved(ctx.Repo, ctx.Pull)
	if err != nil {
		msg := fmt.Sprintf("failed to determine if pull request was approved: %v", err)
		ctx.Log.Err(msg)
		a.github.UpdateStatus(ctx.Repo, ctx.Pull, Error, ApplyStep)
		return false, ExecutionResult{SetupError: GeneralError{errors.New(msg)}}
	}
	if !ok {
		ctx.Log.Info("pull request was not approved")
		a.github.UpdateStatus(ctx.Repo, ctx.Pull, Failure, ApplyStep)
		return false, ExecutionResult{SetupFailure: PullNotApprovedFailure{}}
	}
	return true, ExecutionResult{}
}
