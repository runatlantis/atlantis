package command

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	tally "github.com/uber-go/tally/v4"
)

const (
	planfileSlashReplace = "::"
)

// ProjectContext defines the context for a plan or apply stage that will
// be executed for a project.
type ProjectContext struct {
	CommandName Name
	// ApplyCmd is the command that users should run to apply this plan. If
	// this is an apply then this will be empty.
	ApplyCmd string
	// ApprovePoliciesCmd is the command that users should run to approve policies for this plan. If
	// this is an apply then this will be empty.
	ApprovePoliciesCmd string
	// PlanRequirements is the list of requirements that must be satisfied
	// before we will run the plan stage.
	PlanRequirements []string
	// ApplyRequirements is the list of requirements that must be satisfied
	// before we will run the apply stage.
	ApplyRequirements []string
	// ImportRequirements is the list of requirements that must be satisfied
	// before we will run the import stage.
	ImportRequirements []string
	// AutomergeEnabled is true if automerge is enabled for the repo that this
	// project is in.
	AutomergeEnabled bool
	// ParallelApplyEnabled is true if parallel apply is enabled for this project.
	ParallelApplyEnabled bool
	// ParallelPlanEnabled is true if parallel plan is enabled for this project.
	ParallelPlanEnabled bool
	// ParallelPolicyCheckEnabled is true if parallel policy_check is enabled for this project.
	ParallelPolicyCheckEnabled bool
	// AutoplanEnabled is true if autoplanning is enabled for this project.
	AutoplanEnabled bool
	// BaseRepo is the repository that the pull request will be merged into.
	BaseRepo models.Repo
	// EscapedCommentArgs are the extra arguments that were added to the atlantis
	// command, ex. atlantis plan -- -target=resource. We then escape them
	// by adding a \ before each character so that they can be used within
	// sh -c safely, i.e. sh -c "terraform plan $(touch bad)".
	EscapedCommentArgs []string
	// HeadRepo is the repository that is getting merged into the BaseRepo.
	// If the pull request branch is from the same repository then HeadRepo will
	// be the same as BaseRepo.
	HeadRepo models.Repo
	// Dependencies are a list of project that this project relies on
	// their apply status. These projects must be applied first.
	//
	// Atlantis uses this information to valid the apply
	// orders and to warn the user if they're applying a project that
	// depends on other projects.
	DependsOn []string
	// Log is a logger that's been set up for this context.
	Log logging.SimpleLogging
	// Scope is the scope for reporting stats setup for this context
	Scope tally.Scope
	// PullReqStatus holds state about the PR that requires additional computation outside models.PullRequest
	PullReqStatus models.PullReqStatus
	// CurrentProjectPlanStatus is the status of the current project prior to this command.
	ProjectPlanStatus models.ProjectPlanStatus
	//PullStatus is the status of the current pull request prior to this command.
	PullStatus *models.PullStatus
	// ProjectPolicyStatus is the status of policy sets of the current project prior to this command.
	ProjectPolicyStatus []models.PolicySetStatus

	// Pull is the pull request we're responding to.
	Pull models.PullRequest
	// ProjectName is the name of the project set in atlantis.yaml. If there was
	// no name this will be an empty string.
	ProjectName string
	// RepoConfigVersion is the version of the repo's atlantis.yaml file. If
	// there was no file, this will be 0.
	RepoConfigVersion int
	// RePlanCmd is the command that users should run to re-plan this project.
	// If this is an apply then this will be empty.
	RePlanCmd string
	// RepoRelDir is the directory of this project relative to the repo root.
	RepoRelDir string
	// Steps are the sequence of commands we need to run for this project and this
	// stage.
	Steps []valid.Step
	// TerraformDistribution is the distribution of terraform we should use when
	// executing commands for this project. This can be set to nil in which case
	// we will use the default Atlantis terraform distribution.
	TerraformDistribution *string
	// TerraformVersion is the version of terraform we should use when executing
	// commands for this project. This can be set to nil in which case we will
	// use the default Atlantis terraform version.
	TerraformVersion *version.Version
	// Configuration metadata for a given project.
	User models.User
	// Verbose is true when the user would like verbose output.
	Verbose bool
	// Workspace is the Terraform workspace this project is in. It will always
	// be set.
	Workspace string
	// PolicySets represent the policies that are run on the plan as part of the
	// policy check stage
	PolicySets valid.PolicySets
	// PolicySetTarget describes which policy sets to target on the approve_policies step.
	PolicySetTarget string
	// ClearPolicyApproval determines whether policy counts will be incremented or cleared.
	ClearPolicyApproval bool
	// DeleteSourceBranchOnMerge will attempt to allow a branch to be deleted when merged (AzureDevOps & GitLab Support Only)
	DeleteSourceBranchOnMerge bool
	// Repo locks mode: disabled, on plan or on apply
	RepoLocksMode valid.RepoLocksMode
	// RepoConfigFile
	RepoConfigFile string
	// UUID for atlantis logs
	JobID string
	// The index of order group. Before planning/applying it will use to sort projects. Default is 0.
	ExecutionOrderGroup int
	// If plans/applies should be aborted if any prior plan/apply fails
	AbortOnExcecutionOrderFail bool
	// Allows custom policy check tools outside of Conftest to run in checks
	CustomPolicyCheck bool
	SilencePRComments []string

	// TeamAllowlistChecker is used to check authorization on a project-level
	TeamAllowlistChecker TeamAllowlistChecker
}

// SetProjectScopeTags adds ProjectContext tags to a new returned scope.
func (p ProjectContext) SetProjectScopeTags(scope tally.Scope) tally.Scope {
	v := ""
	if p.TerraformVersion != nil {
		v = p.TerraformVersion.String()
	}

	tags := ProjectScopeTags{
		BaseRepo:         p.BaseRepo.FullName,
		PrNumber:         strconv.Itoa(p.Pull.Num),
		Project:          p.ProjectName,
		ProjectPath:      p.RepoRelDir,
		TerraformVersion: v,
		Workspace:        p.Workspace,
	}

	return scope.Tagged(tags.Loadtags())
}

// GetShowResultFileName returns the filename (not the path) to store the tf show result
func (p ProjectContext) GetShowResultFileName() string {
	if p.ProjectName == "" {
		return fmt.Sprintf("%s.json", p.Workspace)
	}
	projName := strings.Replace(p.ProjectName, "/", planfileSlashReplace, -1)
	return fmt.Sprintf("%s-%s.json", projName, p.Workspace)
}

// GetPolicyCheckResultFileName returns the filename (not the path) to store the result from conftest_client.
func (p ProjectContext) GetPolicyCheckResultFileName() string {
	if p.ProjectName == "" {
		return fmt.Sprintf("%s-policyout.json", p.Workspace)
	}
	projName := strings.Replace(p.ProjectName, "/", planfileSlashReplace, -1)
	return fmt.Sprintf("%s-%s-policyout.json", projName, p.Workspace)
}

// Gets a unique identifier for the current pull request as a single string
func (p ProjectContext) PullInfo() string {
	normalizedOwner := strings.ReplaceAll(p.BaseRepo.Owner, "/", "-")
	normalizedName := strings.ReplaceAll(p.BaseRepo.Name, "/", "-")
	projectRepo := fmt.Sprintf("%s/%s", normalizedOwner, normalizedName)

	return buildPullInfo(projectRepo, p.Pull.Num, p.ProjectName, p.RepoRelDir, p.Workspace)
}
func buildPullInfo(repoName string, pullNum int, projectName string, relDir string, workspace string) string {
	projectIdentifier := getProjectIdentifier(relDir, projectName)
	return fmt.Sprintf("%s/%d/%s/%s", repoName, pullNum, projectIdentifier, workspace)
}

func getProjectIdentifier(relRepoDir string, projectName string) string {
	if projectName != "" {
		return projectName
	}
	// Replace directory separator / with -
	// Replace . with _ to ensure projects with no project name and root dir set to "." have a valid URL
	replacer := strings.NewReplacer("/", "-", ".", "_")
	return replacer.Replace(relRepoDir)
}

// PolicyCleared returns whether all policies are passing or not.
func (p ProjectContext) PolicyCleared() bool {
	passing := true
	for _, psStatus := range p.ProjectPolicyStatus {
		if psStatus.Passed {
			continue
		}
		for _, psCfg := range p.PolicySets.PolicySets {
			if psStatus.PolicySetName == psCfg.Name {
				if psStatus.Approvals != psCfg.ApproveCount {
					passing = false
				}
			}
		}
	}
	return passing
}
