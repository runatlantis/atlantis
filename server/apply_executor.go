package server

import (
	"errors"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/hootsuite/atlantis/locking"
	"time"
)

type ApplyExecutor struct {
	BaseExecutor
	requireApproval    bool
	atlantisGithubUser string
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

func (a *ApplyExecutor) execute(ctx *ExecutionContext, pullCtx *PullRequestContext) ExecutionResult {
	res := a.setupAndApply(ctx, pullCtx)
	res.Command = Apply
	return res
}

func (a *ApplyExecutor) setupAndApply(ctx *ExecutionContext, pullCtx *PullRequestContext) ExecutionResult {
	a.github.UpdateStatus(pullCtx, PendingStatus, "Applying...")

	if a.requireApproval {
		ok, err := a.github.PullIsApproved(pullCtx)
		if err != nil {
			msg := fmt.Sprintf("failed to determine if pull request was approved: %v", err)
			ctx.log.Err(msg)
			a.github.UpdateStatus(pullCtx, ErrorStatus, "Apply Error")
			return ExecutionResult{SetupError: GeneralError{errors.New(msg)}}
		}
		if !ok {
			ctx.log.Info("pull request was not approved")
			a.github.UpdateStatus(pullCtx, FailureStatus, "Apply Failed")
			return ExecutionResult{SetupFailure: PullNotApprovedFailure{}}
		}
	}

	planPaths, err := a.downloadPlans(ctx.repoFullName, ctx.pullNum, ctx.command.environment, a.scratchDir, a.awsConfig, a.s3Bucket)
	if err != nil {
		errMsg := fmt.Sprintf("failed to download plans: %v", err)
		ctx.log.Err(errMsg)
		a.github.UpdateStatus(pullCtx, ErrorStatus, "Apply Error")
		return ExecutionResult{SetupError: GeneralError{errors.New(errMsg)}}
	}

	// If there are no plans found for the pull request
	if len(planPaths) == 0 {
		failure := "found 0 plans for this pull request"
		ctx.log.Warn(failure)
		a.github.UpdateStatus(pullCtx, FailureStatus, "Apply Failure")
		return ExecutionResult{SetupFailure: NoPlansFailure{}}
	}

	//runLog = append(runLog, fmt.Sprintf("-> Downloaded plans: %v", planPaths))
	applyOutputs := []PathResult{}
	for _, planPath := range planPaths {
		output := a.apply(ctx, pullCtx, planPath)
		output.Path = planPath
		applyOutputs = append(applyOutputs, output)
	}
	a.updateGithubStatus(pullCtx, applyOutputs)
	return ExecutionResult{PathResults: applyOutputs}
}

func (a *ApplyExecutor) apply(ctx *ExecutionContext, pullCtx *PullRequestContext, planPath string) PathResult {
	//runLog = append(runLog, fmt.Sprintf("-> Running apply %s", planPath))
	planName := path.Base(planPath)
	planSubDir := a.determinePlanSubDir(planName, ctx.pullNum)
	planDir := filepath.Join(a.scratchDir, ctx.repoFullName, fmt.Sprintf("%v", ctx.pullNum), planSubDir)
	execPath := NewExecutionPath(planDir, planSubDir)
	var config Config
	var remoteStatePath string
	// check if config file is found, if not we continue the run
	if config.Exists(execPath.Absolute) {
		ctx.log.Info("Config file found in %s", execPath.Absolute)
		err := config.Read(execPath.Absolute)
		if err != nil {
			msg := fmt.Sprintf("Error reading config file: %v", err)
			ctx.log.Err(msg)
			return PathResult{
				Status: "error",
				Result: GeneralError{errors.New(msg)},
			}
		}
		// need to use the remote state path and backend to do remote configure
		err = PreRun(&config, ctx.log, execPath.Absolute, ctx.command)
		if err != nil {
			msg := fmt.Sprintf("pre run failed: %v", err)
			ctx.log.Err(msg)
			return PathResult{
				Status: "error",
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
		statePath, err := a.terraform.ConfigureRemoteState(ctx.log, execPath, ctx.command.environment, a.sshKey)
		if err != nil {
			msg := fmt.Sprintf("failed to set up remote state: %v", err)
			ctx.log.Err(msg)
			return PathResult{
				Status: "error",
				Result: GeneralError{errors.New(msg)},
			}
		}
		remoteStatePath = statePath
	} else {
		// use state path from config file
		remoteStatePath = generateStatePath(config.StashPath, ctx.command.environment)
	}

	if remoteStatePath != "" {
		tfEnv := ctx.command.environment
		if tfEnv == "" {
			tfEnv = "default"
		}
		run := locking.Run{
			RepoFullName: pullCtx.repoFullName,
			Path: execPath.Relative,
			Env: tfEnv,
			PullNum: pullCtx.number,
			User: pullCtx.terraformApplier,
			Timestamp: time.Now(),
		}

		lockAttempt, err := a.lockingBackend.TryLock(run)
		if err != nil {
			return PathResult{
				Status: "error",
				Result: GeneralError{fmt.Errorf("failed to acquire lock: %s", err)},
			}
		}
		if lockAttempt.LockAcquired != true && lockAttempt.LockingRun.PullNum != pullCtx.number {
			return PathResult{
				Status: "error",
				Result: GeneralError{fmt.Errorf("failed to acquire lock: lock held by pull request #%d", lockAttempt.LockingRun.PullNum)},
			}
		}
	}

	// need to get auth data from assumed role
	// todo: de-duplicate calls to assumeRole
	//runLog = append(runLog, "-> Assuming role prior to running apply")
	a.awsConfig.AWSSessionName = ctx.pullCreator
	awsSession, err := a.awsConfig.CreateAWSSession()
	if err != nil {
		ctx.log.Err(err.Error())
		return PathResult{
			Status: "error",
			Result: GeneralError{err},
		}
	}
	//runLog = append(runLog, "-> Assumed AWS role successfully")

	credVals, err := awsSession.Config.Credentials.Get()
	if err != nil {
		msg := fmt.Sprintf("failed to get assumed role credentials: %v", err)
		ctx.log.Err(msg)
		return PathResult{
			Status: "error",
			Result: GeneralError{errors.New(msg)},
		}
	}

	ctx.log.Info("running apply from %q", execPath.Relative)

	terraformApplyCmdArgs, output, err := a.terraform.RunTerraformCommand(execPath, []string{"apply", "-no-color", planPath}, []string{
		fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", credVals.AccessKeyID),
		fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", credVals.SecretAccessKey),
		fmt.Sprintf("AWS_SESSION_TOKEN=%s", credVals.SessionToken),
	})
	//runLog = append(runLog, "```\n Apply output:\n", fmt.Sprintf("```bash\n%s\n", string(out[:])))
	if err != nil {
		ctx.log.Err("failed to apply: %v %s", err, output)
		return PathResult{
			Status: "failure",
			Result: ApplyFailure{Command: strings.Join(terraformApplyCmdArgs, " "), Output: output, ErrorMessage: err.Error()},
		}
	}

	// clean up, delete local plan file
	os.Remove(execPath.Absolute) // swallow errors, okay if we failed to delete
	return PathResult{
		Status: "success",
		Result: ApplySuccess{output},
	}
}

func (a *ApplyExecutor) downloadPlans(repoFullName string, pullNum int, env string, outputDir string, awsConfig *AWSConfig, s3Bucket string) (planPaths []string, err error) {
	awsSession, err := awsConfig.CreateAWSSession()
	if err != nil {
		return nil, fmt.Errorf("failed to assume role: %v", err)
	}

	// now use the assumed role to download all the plans
	s3Client := s3.New(awsSession)

	// this will be plans/owner/repo/owner_repo_1, may be more than one if there are subdirs or multiple envs
	plansPath := fmt.Sprintf("plans/%s/%s_%d", repoFullName, strings.Replace(repoFullName, "/", "_", -1), pullNum)
	list, err := s3Client.ListObjects(&s3.ListObjectsInput{Bucket: aws.String(s3Bucket), Prefix: &plansPath})
	if err != nil {
		return nil, fmt.Errorf("failed to list plans in path: %v", err)
	}

	for _, obj := range list.Contents {
		planName := path.Base(*obj.Key)
		// filter to plans for the right env, plan names have the format owner_repo_pullNum_optional_sub_dirs.tfvars.env
		if !strings.HasSuffix(planName, env) {
			continue
		}
		// will be something like /tmp/owner_repo_pullNum_optional_sub_dirs.tfvars.env
		outputPath := fmt.Sprintf("%s/%s", outputDir, planName)
		file, err := os.Create(outputPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create file to write plan to: %v", err)
		}
		defer file.Close()

		downloader := s3manager.NewDownloader(awsSession)
		_, err = downloader.Download(file,
			&s3.GetObjectInput{
				Bucket: aws.String(s3Bucket),
				Key:    obj.Key,
			})
		if err != nil {
			return nil, fmt.Errorf("failed to download plan from s3: %v", err)
		}
		planPaths = append(planPaths, outputPath)
	}
	return planPaths, nil
}

func (a *ApplyExecutor) determinePlanSubDir(planName string, pullNum int) string {
	planDirRegex := fmt.Sprintf(`.*_%d_(.*?)\.`, pullNum)
	regex := regexp.MustCompile(planDirRegex) // we assume this will compile
	match := regex.FindStringSubmatch(planName)
	if len(match) <= 1 {
		return "."
	}
	dirsStr := match[1] // in form dir_subdir_subsubdir
	return filepath.Clean(strings.Replace(dirsStr, "_", "/", -1))
}
