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
	"github.com/hootsuite/atlantis/models"
	"strconv"
)

type ApplyExecutor struct {
	github                *GithubClient
	awsConfig             *AWSConfig
	scratchDir            string
	s3Bucket              string
	sshKey                string
	terraform             *TerraformClient
	githubCommentRenderer *GithubCommentRenderer
	lockingClient         *locking.Client
	requireApproval       bool
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
	res := a.setupAndApply(ctx)
	res.Command = Apply
	comment := a.githubCommentRenderer.render(res, ctx.Log.History.String(), ctx.Command.verbose)
	github.CreateComment(ctx, comment)
}

func (a *ApplyExecutor) setupAndApply(ctx *CommandContext) ExecutionResult {
	a.github.UpdateStatus(ctx.Repo, ctx.Pull, PendingStatus, "Applying...")

	if a.requireApproval {
		ok, err := a.github.PullIsApproved(ctx.Repo, ctx.Pull)
		if err != nil {
			msg := fmt.Sprintf("failed to determine if pull request was approved: %v", err)
			ctx.Log.Err(msg)
			a.github.UpdateStatus(ctx.Repo, ctx.Pull, ErrorStatus, "Apply Error")
			return ExecutionResult{SetupError: GeneralError{errors.New(msg)}}
		}
		if !ok {
			ctx.Log.Info("pull request was not approved")
			a.github.UpdateStatus(ctx.Repo, ctx.Pull, FailureStatus, "Apply Failed")
			return ExecutionResult{SetupFailure: PullNotApprovedFailure{}}
		}
	}

	planPaths, err := a.downloadPlans(ctx.Repo.FullName, ctx.Pull.Num, ctx.Command.environment, a.scratchDir, a.awsConfig, a.s3Bucket)
	if err != nil {
		errMsg := fmt.Sprintf("failed to download plans: %v", err)
		ctx.Log.Err(errMsg)
		a.github.UpdateStatus(ctx.Repo, ctx.Pull, ErrorStatus, "Apply Error")
		return ExecutionResult{SetupError: GeneralError{errors.New(errMsg)}}
	}

	// If there are no plans found for the pull request
	if len(planPaths) == 0 {
		failure := "found 0 plans for this pull request"
		ctx.Log.Warn(failure)
		a.github.UpdateStatus(ctx.Repo, ctx.Pull, FailureStatus, "Apply Failure")
		return ExecutionResult{SetupFailure: NoPlansFailure{}}
	}

	//runLog = append(runLog, fmt.Sprintf("-> Downloaded plans: %v", planPaths))
	applyOutputs := []PathResult{}
	for _, planPath := range planPaths {
		output := a.apply(ctx, planPath)
		output.Path = planPath
		applyOutputs = append(applyOutputs, output)
	}
	a.updateGithubStatus(ctx, applyOutputs)
	return ExecutionResult{PathResults: applyOutputs}
}

func (a *ApplyExecutor) apply(ctx *CommandContext, planPath string) PathResult {
	planName := path.Base(planPath)
	planSubDir := a.determinePlanSubDir(planName, ctx.Pull.Num)
	// todo: don't assume repo is cloned here
	repoDir := filepath.Join(a.scratchDir, ctx.Repo.FullName, strconv.Itoa(ctx.Pull.Num))
	planDir := filepath.Join(repoDir, planSubDir)
	project := models.NewProject(ctx.Repo.FullName, planSubDir)
	execPath := NewExecutionPath(planDir, planSubDir)
	var config Config
	var remoteStatePath string
	// check if config file is found, if not we continue the run
	if config.Exists(execPath.Absolute) {
		ctx.Log.Info("Config file found in %s", execPath.Absolute)
		err := config.Read(execPath.Absolute)
		if err != nil {
			msg := fmt.Sprintf("Error reading config file: %v", err)
			ctx.Log.Err(msg)
			return PathResult{
				Status: "error",
				Result: GeneralError{errors.New(msg)},
			}
		}
		// need to use the remote state path and backend to do remote configure
		err = PreRun(&config, ctx.Log, execPath.Absolute, ctx.Command)
		if err != nil {
			msg := fmt.Sprintf("pre run failed: %v", err)
			ctx.Log.Err(msg)
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
		statePath, err := a.terraform.ConfigureRemoteState(ctx.Log, repoDir, project, ctx.Command.environment, a.sshKey)
		if err != nil {
			msg := fmt.Sprintf("failed to set up remote state: %v", err)
			ctx.Log.Err(msg)
			return PathResult{
				Status: "error",
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

		lockAttempt, err := a.lockingClient.TryLock(project, tfEnv, ctx.Pull.Num)
		if err != nil {
			return PathResult{
				Status: "error",
				Result: GeneralError{fmt.Errorf("failed to acquire lock: %s", err)},
			}
		}
		if lockAttempt.LockAcquired != true && lockAttempt.LockingPullNum != ctx.Pull.Num {
			return PathResult{
				Status: "error",
				Result: GeneralError{fmt.Errorf("failed to acquire lock: lock held by pull request #%d", lockAttempt.LockingPullNum)},
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
			Status: "error",
			Result: GeneralError{err},
		}
	}

	credVals, err := awsSession.Config.Credentials.Get()
	if err != nil {
		msg := fmt.Sprintf("failed to get assumed role credentials: %v", err)
		ctx.Log.Err(msg)
		return PathResult{
			Status: "error",
			Result: GeneralError{errors.New(msg)},
		}
	}

	ctx.Log.Info("running apply from %q", execPath.Relative)

	terraformApplyCmdArgs, output, err := a.terraform.RunTerraformCommand(execPath.Absolute, []string{"apply", "-no-color", planPath}, []string{
		fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", credVals.AccessKeyID),
		fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", credVals.SecretAccessKey),
		fmt.Sprintf("AWS_SESSION_TOKEN=%s", credVals.SessionToken),
	})
	//runLog = append(runLog, "```\n Apply output:\n", fmt.Sprintf("```bash\n%s\n", string(out[:])))
	if err != nil {
		ctx.Log.Err("failed to apply: %v %s", err, output)
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

func (a *ApplyExecutor) updateGithubStatus(ctx *CommandContext, pathResults []PathResult) {
	// the status will be the worst result
	worstResult := a.worstResult(pathResults)
	if worstResult == "success" {
		a.github.UpdateStatus(ctx.Repo, ctx.Pull, SuccessStatus, "Apply Succeeded")
	} else if worstResult == "failure" {
		a.github.UpdateStatus(ctx.Repo, ctx.Pull, FailureStatus, "Apply Failed")
	} else {
		a.github.UpdateStatus(ctx.Repo, ctx.Pull, ErrorStatus, "Apply Error")
	}
}

func (a *ApplyExecutor) worstResult(results []PathResult) string {
	var worst string = "success"
	for _, result := range results {
		if result.Status == "error" {
			return result.Status
		} else if result.Status == "failure" {
			worst = result.Status
		}
	}
	return worst
}
