package server

import (
	"fmt"
	"github.com/hootsuite/atlantis/locking"
	"github.com/hootsuite/atlantis/logging"
	"github.com/hootsuite/atlantis/models"
	"github.com/hootsuite/atlantis/plan"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

// PlanExecutor handles everything related to running the Terraform plan including integration with S3, Terraform, and GitHub
type PlanExecutor struct {
	github                *GithubClient
	githubStatus          *GithubStatus
	awsConfig             *AWSConfig
	scratchDir            string
	s3Bucket              string
	sshKey                string
	terraform             *TerraformClient
	githubCommentRenderer *GithubCommentRenderer
	lockingClient         *locking.Client
	// DeleteLockURL is a function that given a lock id will return a url for deleting the lock
	DeleteLockURL func(id string) (url string)
	planBackend   plan.Backend
}

/** Result Types **/
type PlanSuccess struct {
	TerraformOutput string
	LockURL         string
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

func (p *PlanExecutor) execute(ctx *CommandContext, github *GithubClient) {
	res := p.setupAndPlan(ctx)
	res.Command = Plan
	comment := p.githubCommentRenderer.render(res, ctx.Log.History.String(), ctx.Command.verbose)
	github.CreateComment(ctx.Repo, ctx.Pull, comment)
}

func (p *PlanExecutor) setupAndPlan(ctx *CommandContext) ExecutionResult {
	p.githubStatus.Update(ctx.Repo, ctx.Pull, Pending, PlanStep)

	// todo: lock when cloning or somehow separate workspaces
	// clean the directory where we're going to clone
	cloneDir := fmt.Sprintf("%s/%s/%d", p.scratchDir, ctx.Repo.FullName, ctx.Pull.Num)
	ctx.Log.Info("cleaning clone directory %q", cloneDir)
	if err := os.RemoveAll(cloneDir); err != nil {
		ctx.Log.Warn("failed to clean dir %q before cloning, attempting to continue: %v", cloneDir, err)
	}

	// create the directory and parents if necessary
	ctx.Log.Info("creating dir %q", cloneDir)
	if err := os.MkdirAll(cloneDir, 0755); err != nil {
		ctx.Log.Warn("failed to create dir %q prior to cloning, attempting to continue: %v", cloneDir, err)
	}

	// Check if ssh key is set and create git ssh wrapper
	cloneCmd := exec.Command("git", "clone", ctx.Repo.SSHURL, cloneDir)
	if p.sshKey != "" {
		err := GenerateSSHWrapper()
		if err != nil {
			return p.setupError(ctx, errors.Wrap(err, "creating git ssh wrapper"))
		}
		cloneCmd.Env = []string{
			fmt.Sprintf("GIT_SSH=%s", defaultSSHWrapper),
			fmt.Sprintf("PKEY=%s", p.sshKey),
		}
	}

	// git clone the repo
	ctx.Log.Info("git cloning %q into %q", ctx.Repo.SSHURL, cloneDir)
	if output, err := cloneCmd.CombinedOutput(); err != nil {
		return p.setupError(ctx, fmt.Errorf("cloning %s: %s: %s", ctx.Repo.SSHURL, err, string(output)))
	}

	// check out the branch for this PR
	ctx.Log.Info("checking out branch %q", ctx.Pull.Branch)
	checkoutCmd := exec.Command("git", "checkout", ctx.Pull.Branch)
	checkoutCmd.Dir = cloneDir
	if err := checkoutCmd.Run(); err != nil {
		return p.setupError(ctx, errors.Wrapf(err, "checking out branch %s", ctx.Pull.Branch))
	}

	ctx.Log.Info("listing modified files from pull request")
	modifiedFiles, err := p.github.GetModifiedFiles(ctx.Repo, ctx.Pull)
	if err != nil {
		return p.setupError(ctx, errors.Wrap(err, "getting modified files"))
	}
	modifiedTerraformFiles := p.filterToTerraform(modifiedFiles)
	if len(modifiedTerraformFiles) == 0 {
		ctx.Log.Info("no modified terraform files found, exiting")
		p.githubStatus.Update(ctx.Repo, ctx.Pull, Failure, PlanStep)
		return ExecutionResult{SetupError: GeneralError{errors.New("Plan Failed: no modified terraform files found")}}
	}
	ctx.Log.Debug("Found %d modified terraform files: %v", len(modifiedTerraformFiles), modifiedTerraformFiles)

	projects := p.ModifiedProjects(ctx.Repo.FullName, modifiedTerraformFiles)
	if len(projects) == 0 {
		ctx.Log.Info("no Terraform projects were modified")
		p.githubStatus.Update(ctx.Repo, ctx.Pull, Failure, PlanStep)
		return ExecutionResult{SetupError: GeneralError{errors.New("Plan Failed: we determined that no terraform projects were modified")}}
	}

	// todo: update how we clean the workspace based on the new way of storing plans
	planFilesPrefix := fmt.Sprintf("%s_%d", strings.Replace(ctx.Repo.FullName, "/", "_", -1), ctx.Pull.Num)
	if err := p.CleanWorkspace(ctx.Log, planFilesPrefix, p.scratchDir, cloneDir, projects); err != nil {
		return p.setupError(ctx, errors.Wrap(err, "cleaning workspace"))
	}

	var config Config
	planOutputs := []PathResult{}
	for _, project := range projects {
		// check if config file is found, if not we continue the run
		absolutePath := filepath.Join(cloneDir, project.Path)
		if config.Exists(absolutePath) {
			ctx.Log.Info("Config file found in %s", absolutePath)
			err := config.Read(absolutePath)
			if err != nil {
				errMsg := fmt.Sprintf("Error reading config file: %v", err)
				ctx.Log.Err(errMsg)
				return ExecutionResult{SetupError: GeneralError{errors.New(errMsg)}}
			}
			// need to use the remote state path and backend to do remote configure
			err = PreRun(&config, ctx.Log, absolutePath, ctx.Command)
			if err != nil {
				errMsg := fmt.Sprintf("pre run failed: %v", err)
				ctx.Log.Err(errMsg)
				return ExecutionResult{SetupError: GeneralError{errors.New(errMsg)}}
			}

			// check if terraform version is specified in config
			if config.TerraformVersion != "" {
				p.terraform.tfExecutableName = "terraform" + config.TerraformVersion
			} else {
				p.terraform.tfExecutableName = "terraform"
			}
		}
		generatePlanResponse := p.plan(ctx, cloneDir, p.scratchDir, project, p.sshKey, config.StashPath)
		generatePlanResponse.Path = project.Path
		planOutputs = append(planOutputs, generatePlanResponse)
	}
	p.githubStatus.UpdatePathResult(ctx, planOutputs)
	return ExecutionResult{PathResults: planOutputs}
}

// plan runs the steps necessary to run `terraform plan`. If there is an error, the error message will be encapsulated in error
// and the GeneratePlanResponse struct will also contain the full log including the error
func (p *PlanExecutor) plan(
	ctx *CommandContext,
	repoDir string,
	planOutDir string,
	project models.Project,
	sshKey string,
	stashPath string) PathResult {
	ctx.Log.Info("generating plan for path %q", project.Path)

	// NOTE: THIS CODE IS TO SUPPORT TERRAFORM PROJECTS THAT AREN'T USING ATLANTIS CONFIG FILE.
	if stashPath == "" {
		_, err := p.terraform.ConfigureRemoteState(ctx.Log, repoDir, project, ctx.Command.environment, sshKey)
		if err != nil {
			return PathResult{
				Status: Error,
				Result: GeneralError{fmt.Errorf("failed to configure remote state: %s", err)},
			}
		}
	}
	// todo: setting environment to default should be done elsewhere
	tfEnv := ctx.Command.environment
	if tfEnv == "" {
		tfEnv = "default"
	}
	lockAttempt, err := p.lockingClient.TryLock(project, ctx.Command.environment, ctx.Pull, ctx.User)
	if err != nil {
		return PathResult{
			Status: Failure,
			Result: GeneralError{fmt.Errorf("failed to lock state: %v", err)},
		}
	}

	// the run is locked unless the locking run is the same pull id as this run
	if lockAttempt.LockAcquired == false && lockAttempt.CurrLock.Pull.Num != ctx.Pull.Num {
		return PathResult{
			Status: Failure,
			Result: RunLockedFailure{lockAttempt.CurrLock.Pull.Num},
		}
	}

	// Run terraform plan
	ctx.Log.Info("running terraform plan in directory %q", project.Path)
	tfPlanCmd := []string{"plan", "-refresh", "-no-color"}
	planFile := filepath.Join(repoDir, project.Path, fmt.Sprintf("%s.tfplan", tfEnv))
	if ctx.Command.environment != "" {
		tfEnvFileName := filepath.Join("env", ctx.Command.environment+".tfvars")
		if _, err := os.Stat(filepath.Join(repoDir, project.Path, tfEnvFileName)); err == nil {
			tfPlanCmd = append(tfPlanCmd, "-var-file", tfEnvFileName, "-out", planFile)
		} else {
			ctx.Log.Err("environment file %q not found", tfEnvFileName)
			return PathResult{
				Status: Failure,
				Result: EnvironmentFileNotFoundFailure{tfEnvFileName},
			}
		}
	} else {
		tfPlanCmd = append(tfPlanCmd, "-out", planFile)
	}

	// set pull request creator as the session name
	p.awsConfig.AWSSessionName = ctx.Pull.Author
	awsSession, err := p.awsConfig.CreateAWSSession()
	if err != nil {
		ctx.Log.Err(err.Error())
		return PathResult{
			Status: Error,
			Result: GeneralError{err},
		}
	}

	credVals, err := awsSession.Config.Credentials.Get()
	if err != nil {
		err = fmt.Errorf("failed to get assumed role credentials: %v", err)
		ctx.Log.Err(err.Error())
		return PathResult{
			Status: Error,
			Result: GeneralError{err},
		}
	}

	terraformPlanCmdArgs, output, err := p.terraform.RunTerraformCommand(filepath.Join(repoDir, project.Path), tfPlanCmd, []string{
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
		ctx.Log.Err("error running terraform plan: %v", output)
		ctx.Log.Info("unlocking state since plan failed")
		if _, err := p.lockingClient.Unlock(lockAttempt.LockKey); err != nil {
			ctx.Log.Err("error unlocking state: %v", err)
		}
		return PathResult{
			Status: Failure,
			Result: err,
		}
	}
	// Save the plan
	if err := p.planBackend.SavePlan(planFile, project, tfEnv, ctx.Pull.Num); err != nil {
		ctx.Log.Err("saving plan: %s", err)
		// there was an error planning so unlock
		if _, err := p.lockingClient.Unlock(lockAttempt.LockKey); err != nil {
			ctx.Log.Err("error unlocking: %v", err)
		}
		return PathResult{
			Status: Error,
			Result: GeneralError{err},
		}
	}
	ctx.Log.Info("saved plan successfully")

	// Delete local plan file
	if err := os.Remove(planFile); err != nil {
		ctx.Log.Err("failed to delete local plan file %q: %s", planFile, err)
		// don't return an error since it should still be fine
	}
	return PathResult{
		Status: Success,
		Result: PlanSuccess{
			TerraformOutput: output,
			LockURL:         p.DeleteLockURL(lockAttempt.LockKey),
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

// CleanWorkspace deletes all .terraform/ folders from the project folders and cleans up any plans in the output directory
func (p *PlanExecutor) CleanWorkspace(log *logging.SimpleLogger, deleteFilesPrefix string, planOutDir string, repoDir string, projects []models.Project) error {
	log.Info("cleaning workspace directory %q", planOutDir)

	// delete .terraform directories
	for _, project := range projects {
		os.RemoveAll(filepath.Join(repoDir, project.Path, ".terraform"))
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

// todo: make OO
func generateStatePath(path string, tfEnvName string) string {
	return strings.Replace(path, "$ENVIRONMENT", tfEnvName, -1)
}

func (p *PlanExecutor) setupError(ctx *CommandContext, err error) ExecutionResult {
	ctx.Log.Err(err.Error())
	p.githubStatus.Update(ctx.Repo, ctx.Pull, Error, PlanStep)
	return ExecutionResult{SetupError: GeneralError{err}}
}
