package server

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"github.com/hootsuite/atlantis/locking"
	"time"
	"github.com/hootsuite/atlantis/logging"
)

// PlanExecutor handles everything related to running the Terraform plan including integration with S3, Terraform, and Github
type PlanExecutor struct {
	BaseExecutor
	// DeleteLockURL is a function that given a lock id will return a url for deleting the lock
	DeleteLockURL func(id string) (url string)
}

/** Result Types **/
type PlanSuccess struct {
	TerraformOutput string
	LockURL   string
}

func (p PlanSuccess) Template() *CompiledTemplate {
	return PlanSuccessTmpl
}

type RunLockedFailure struct {
	LockingPullNum int
}

func (r RunLockedFailure) Template() *CompiledTemplate {
	return RunLockedFailureTmpl
}

type EnvironmentFileNotFoundFailure struct {
	Filename string
}

func (e EnvironmentFileNotFoundFailure) Template() *CompiledTemplate {
	return EnvironmentFileNotFoundFailureTmpl
}

type TerraformFailure struct {
	Command string
	Output  string
}

func (t TerraformFailure) Template() *CompiledTemplate {
	return TerraformFailureTmpl
}

type EnvironmentFailure struct{}

func (e EnvironmentFailure) Template() *CompiledTemplate {
	return EnvironmentErrorTmpl
}

func (p *PlanExecutor) execute(ctx *ExecutionContext, pullCtx *PullRequestContext) ExecutionResult {
	res := p.setupAndPlan(ctx, pullCtx)
	res.Command = Plan
	return res
}

func (p *PlanExecutor) setupAndPlan(ctx *ExecutionContext, pullCtx *PullRequestContext) ExecutionResult {
	p.github.UpdateStatus(pullCtx, "pending", "Planning...")

	// todo: lock when cloning or somehow separate workspaces
	// clean the directory where we're going to clone
	cloneDir := fmt.Sprintf("%s/%s/%d", p.scratchDir, ctx.repoFullName, ctx.pullNum)
	ctx.log.Info("cleaning clone directory %q", cloneDir)
	if err := os.RemoveAll(cloneDir); err != nil {
		ctx.log.Warn("failed to clean dir %q before cloning, attempting to continue: %v", cloneDir, err)
	}

	// create the directory and parents if necessary
	ctx.log.Info("creating dir %q", cloneDir)
	if err := os.MkdirAll(cloneDir, 0755); err != nil {
		ctx.log.Warn("failed to create dir %q prior to cloning, attempting to continue: %v", cloneDir, err)
	}

	// Check if ssh key is set and create git ssh wrapper
	cloneCmd := exec.Command("git", "clone", ctx.repoSSHUrl, cloneDir)
	if p.sshKey != "" {
		err := GenerateSSHWrapper()
		if err != nil {
			errMsg := fmt.Sprintf("failed to create git ssh wrapper: %v", err)
			ctx.log.Err(errMsg)
			p.github.UpdateStatus(pullCtx, ErrorStatus, "Plan Error")
			return ExecutionResult{SetupError: GeneralError{errors.New(errMsg)}}
		}

		cloneCmd.Env = []string{
			fmt.Sprintf("GIT_SSH=%s", defaultSSHWrapper),
			fmt.Sprintf("PKEY=%s", p.sshKey),
		}
	}

	// git clone the repo
	ctx.log.Info("git cloning %q into %q", ctx.repoSSHUrl, cloneDir)
	if output, err := cloneCmd.CombinedOutput(); err != nil {
		errMsg := fmt.Sprintf("failed to clone repository %q: %v: %s", ctx.repoSSHUrl, err, string(output))
		ctx.log.Err(errMsg)
		p.github.UpdateStatus(pullCtx, ErrorStatus, "Plan Error")
		return ExecutionResult{SetupError: GeneralError{errors.New(errMsg)}}
	}

	// check out the branch for this PR
	ctx.log.Info("checking out branch %q", ctx.branch)
	checkoutCmd := exec.Command("git", "checkout", ctx.branch)
	checkoutCmd.Dir = cloneDir
	if err := checkoutCmd.Run(); err != nil {
		errMsg := fmt.Sprintf("failed to git checkout branch %q: %v", ctx.branch, err)
		ctx.log.Err(errMsg)
		p.github.UpdateStatus(pullCtx, ErrorStatus, "Plan Error")
		return ExecutionResult{SetupError: GeneralError{errors.New(errMsg)}}
	}

	ctx.log.Info("listing modified files from pull request")
	modifiedFiles, err := p.github.GetModifiedFiles(pullCtx)
	if err != nil {
		errMsg := fmt.Sprintf("failed to retrieve list of modified files from GitHub: %v", err)
		ctx.log.Err(errMsg)
		p.github.UpdateStatus(pullCtx, ErrorStatus, "Plan Error")
		return ExecutionResult{SetupError: GeneralError{errors.New(errMsg)}}
	}
	modifiedTerraformFiles := p.filterToTerraform(modifiedFiles)
	if len(modifiedTerraformFiles) == 0 {
		ctx.log.Info("no modified terraform files found, exiting")
		p.github.UpdateStatus(pullCtx, FailureStatus, "Plan Failed")
		return ExecutionResult{SetupError: GeneralError{errors.New("Plan Failed: no modified terraform files found")}}
	}
	ctx.log.Debug("Found %d modified terraform files: %v", len(modifiedTerraformFiles), modifiedTerraformFiles)

	execPaths := p.DetermineExecPaths(p.trimSuffix(cloneDir, "/"), modifiedTerraformFiles)
	if len(execPaths) == 0 {
		ctx.log.Info("no exec paths found, exiting")
		p.github.UpdateStatus(pullCtx, FailureStatus, "Plan Failed")
		return ExecutionResult{SetupError: GeneralError{errors.New("Plan Failed: there were no paths to run plan in")}}
	}

	planFilesPrefix := fmt.Sprintf("%s_%d", strings.Replace(ctx.repoFullName, "/", "_", -1), ctx.pullNum)
	if err := p.CleanWorkspace(ctx.log, planFilesPrefix, p.scratchDir, execPaths); err != nil {
		errMsg := fmt.Sprintf("failed to clean workspace, aborting: %v", err)
		ctx.log.Err(errMsg)
		p.github.UpdateStatus(pullCtx, ErrorStatus, "Plan Error")
		return ExecutionResult{SetupError: GeneralError{errors.New(errMsg)}}
	}
	s3Client := NewS3Client(p.awsConfig, p.s3Bucket, "plans")

	var config Config
	// run `terraform plan` in each plan path and collect the results
	planOutputs := []PathResult{}
	for _, path := range execPaths {
		// todo: not sure it makes sense to be generating the output filename and plan name here
		tfPlanFilename := p.GenerateOutputFilename(cloneDir, path, ctx.command.environment)
		tfPlanName := fmt.Sprintf("%s_%d%s", strings.Replace(ctx.repoFullName, "/", "_", -1), ctx.pullNum, tfPlanFilename)
		s3Key := fmt.Sprintf("%s/%s", ctx.repoFullName, tfPlanName)
		// check if config file is found, if not we continue the run
		if config.Exists(path.Absolute) {
			ctx.log.Info("Config file found in %s", path.Absolute)
			err := config.Read(path.Absolute)
			if err != nil {
				errMsg := fmt.Sprintf("Error reading config file: %v", err)
				ctx.log.Err(errMsg)
				return ExecutionResult{SetupError: GeneralError{errors.New(errMsg)}}
			}
			// need to use the remote state path and backend to do remote configure
			err = PreRun(&config, ctx.log, path.Absolute, ctx.command)
			if err != nil {
				errMsg := fmt.Sprintf("pre run failed: %v", err)
				ctx.log.Err(errMsg)
				return ExecutionResult{SetupError: GeneralError{errors.New(errMsg)}}
			}

			// check if terraform version is specified in config
			if config.TerraformVersion != "" {
				p.terraform.tfExecutableName = "terraform" + config.TerraformVersion
			} else {
				p.terraform.tfExecutableName = "terraform"
			}
		}
		generatePlanResponse := p.plan(ctx.log, pullCtx, cloneDir, p.scratchDir, tfPlanName, s3Client, path, ctx.command.environment, s3Key, p.sshKey, ctx.pullCreator, config.StashPath)
		generatePlanResponse.Path = path.Relative
		planOutputs = append(planOutputs, generatePlanResponse)
	}
	p.updateGithubStatus(pullCtx, planOutputs)
	return ExecutionResult{PathResults: planOutputs}
}

// plan runs the steps necessary to run `terraform plan`. If there is an error, the error message will be encapsulated in error
// and the GeneratePlanResponse struct will also contain the full log including the error
func (p *PlanExecutor) plan(log *logging.SimpleLogger,
	pullCtx *PullRequestContext,
	repoDir string,
	planOutDir string,
	tfPlanName string,
	s3Client S3Client,
	path ExecutionPath,
	tfEnvName string,
	s3Key string,
	sshKey string,
	pullRequestCreator string,
	stashPath string) PathResult {
	log.Info("generating plan for path %q", path)
	run := locking.Run{
		RepoFullName:  pullCtx.repoFullName,
		Path:      path.Relative,
		Env:       tfEnvName,
		PullNum:    pullCtx.number,
		User:      pullCtx.terraformApplier,
		Timestamp: time.Now(),
	}

	// NOTE: THIS CODE IS TO SUPPORT TERRAFORM PROJECTS THAT AREN'T USING ATLANTIS CONFIG FILE.
	if stashPath == "" {
		_, err := p.terraform.ConfigureRemoteState(log, path, tfEnvName, sshKey)
		if err != nil {
			return PathResult{
				Status: "error",
				Result: GeneralError{fmt.Errorf("failed to configure remote state: %s", err)},
			}
		}
	}
	lockAttempt, err := p.lockingBackend.TryLock(run)
	if err != nil {
		return PathResult{
			Status:" failure",
			Result: GeneralError{fmt.Errorf("failed to lock state: %v", err)},
		}
	}

	// the run is locked unless the locking run is the same pull id as this run
	if lockAttempt.LockAcquired == false && lockAttempt.LockingRun.PullNum != pullCtx.number {
		return PathResult{
			Status: "failure",
			Result: RunLockedFailure{lockAttempt.LockingRun.PullNum},
		}
	}

	// Run terraform plan
	log.Info("running terraform plan in directory %q", path.Relative)
	tfPlanCmd := []string{"plan", "-refresh", "-no-color"}
	// Generate terraform plan filename
	tfPlanOutputPath := filepath.Join(planOutDir, tfPlanName)
	// Generate terraform plan arguments
	if tfEnvName != "" {
		tfEnvFileName := filepath.Join("env", tfEnvName+".tfvars")
		if _, err := os.Stat(filepath.Join(path.Absolute, tfEnvFileName)); err == nil {
			tfPlanCmd = append(tfPlanCmd, "-var-file", tfEnvFileName, "-out", tfPlanOutputPath)
		} else {
			log.Err("environment file %q not found", tfEnvFileName)
			return PathResult{
				Status: "failure",
				Result: EnvironmentFileNotFoundFailure{tfEnvFileName},
			}
		}
	} else {
		tfPlanCmd = append(tfPlanCmd, "-out", tfPlanOutputPath)
	}

	// set pull request creator as the session name
	p.awsConfig.AWSSessionName = pullRequestCreator
	awsSession, err := p.awsConfig.CreateAWSSession()
	if err != nil {
		log.Err(err.Error())
		return PathResult{
			Status: "error",
			Result: GeneralError{err},
		}
	}

	credVals, err := awsSession.Config.Credentials.Get()
	if err != nil {
		err = fmt.Errorf("failed to get assumed role credentials: %v", err)
		log.Err(err.Error())
		return PathResult{
			Status: "error",
			Result: GeneralError{err},
		}
	}

	terraformPlanCmdArgs, output, err := p.terraform.RunTerraformCommand(path, tfPlanCmd, []string{
		fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", credVals.AccessKeyID),
		fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", credVals.SecretAccessKey),
		fmt.Sprintf("AWS_SESSION_TOKEN=%s", credVals.SessionToken),
	})
	if err != nil {
		if err.Error() != "exit status 1" {
			// if it's not an exit 1 then the details about the failure won't be in the output but in the error itself
			output = err.Error()
		}
		err := TerraformFailure{
			Command: strings.Join(terraformPlanCmdArgs, " "),
			Output:  output,
		}
		log.Err("error running terraform plan: %v", output)
		log.Info("unlocking state since plan failed")
		if err := p.lockingBackend.Unlock(lockAttempt.LockID); err != nil {
			log.Err("error unlocking state: %v", err)
		}
		return PathResult{
			Status: "failure",
			Result: err,
		}
	}
	// Upload plan to S3
	log.Info("uploading plan to S3 with key %q", s3Key)
	if err := UploadPlanFile(s3Client, s3Key, tfPlanOutputPath); err != nil {
		err = fmt.Errorf("failed to upload to S3: %v", err)
		log.Err(err.Error())
		if err := p.lockingBackend.Unlock(lockAttempt.LockID); err != nil {
			log.Err("error unlocking state: %v", err)
		}
		return PathResult{
			Status: "error",
			Result: GeneralError{err},
		}
	}
	// Delete local plan file
	planFilePath := fmt.Sprintf("%s/%s", planOutDir, tfPlanName)
	log.Info("deleting local plan file %q", planFilePath)
	if err := os.Remove(planFilePath); err != nil {
		log.Err("failed to delete local plan file %q", planFilePath, err)
		// todo: return an error
	}
	return PathResult{
		Status: "success",
		Result: PlanSuccess{
			TerraformOutput: output,
			LockURL: p.DeleteLockURL(lockAttempt.LockID),
		},
	}
}

func (p *PlanExecutor) filterToTerraform(files []string) []string {
	var out = []string{}
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

func (p *PlanExecutor) trimSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		s = s[:len(s)-len(suffix)]
	}
	return s
}

func (p *PlanExecutor) removeDuplicates(paths []ExecutionPath) []ExecutionPath {
	deDuped := []ExecutionPath{}
	seen := map[ExecutionPath]bool{}
	for _, path := range paths {
		if _, ok := seen[path]; !ok {
			deDuped = append(deDuped, path)
			seen[path] = true
		}
	}
	return deDuped
}

// DetermineExecPaths returns the list of directories in which we'll need to run `terraform plan`
func (p *PlanExecutor) DetermineExecPaths(repoPath string, modifiedFiles []string) []ExecutionPath {
	var paths []ExecutionPath
	for _, modifiedFile := range modifiedFiles {
		relative := p.getRelativePlanPath(modifiedFile)
		absolute := filepath.Join(repoPath, relative) + "/"
		paths = append(paths, NewExecutionPath(absolute, relative))
	}
	return p.removeDuplicates(paths)
}

func (p *PlanExecutor) getRelativePlanPath(modifiedFilePath string) string {
	parentDir := filepath.Dir(modifiedFilePath)
	if filepath.Base(parentDir) == "env" {
		// if the modified file was inside an env/ directory, we treat this specially and
		// run plan one level up
		return filepath.Dir(parentDir)
	}
	return parentDir
}

// CleanWorkspace deletes all .terraform/ folders from the plan paths and cleans up any plans in the output directory
func (p *PlanExecutor) CleanWorkspace(log *logging.SimpleLogger, deleteFilesPrefix string, planOutDir string, execPaths []ExecutionPath) error {
	log.Info("cleaning workspace directory %q and plan paths %v", planOutDir, execPaths)

	// delete .terraform directories
	for _, path := range execPaths {
		os.RemoveAll(filepath.Join(path.Absolute, ".terraform"))
	}
	// delete old plan files
	files, err := ioutil.ReadDir(planOutDir)
	if err != nil {
		return err
	}
	for _, file := range files {
		if strings.HasPrefix(file.Name(), deleteFilesPrefix) {
			log.Info("deleting file %q", file.Name())
			fullPath := filepath.Join(planOutDir, file.Name())
			if err := os.Remove(fullPath); err != nil {
				log.Warn("failed to remove plan file %q", fullPath)
			}
		}
	}
	return nil
}

func (p *PlanExecutor) DeleteLocalPlanFile(path string) error {
	return os.Remove(path)
}

// GenerateOutputFilename determines the name of the plan that will be stored in s3
// if we're executing inside a sub directory, there will be a leading underscore
func (p *PlanExecutor) GenerateOutputFilename(repoDir string, execPath ExecutionPath, tfEnvName string) string {
	prefix := ""
	if execPath.Relative != "." {
		// If not executing at repo root, need to encode the sub dir in the name of the output file.
		// We do this by substituting / for _
		// We also add an _ because this gets appended to a larger path
		// todo: refactor the path handling so it's all in one place
		prefix = "_" + strings.Replace(execPath.Relative, "/", "_", -1)
	}
	suffix := ""
	if tfEnvName != "" {
		suffix = "." + tfEnvName
	}

	return prefix + ".tfplan" + suffix
}

func generateStatePath(path string, tfEnvName string) string {
	return strings.Replace(path, "$ENVIRONMENT", tfEnvName, -1)
}
