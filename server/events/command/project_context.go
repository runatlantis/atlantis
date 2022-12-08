package command

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/uber-go/tally/v4"
)

const (
	planfileSlashReplace = "::"
)

type ContextFlags struct {
	ParallelApply,
	ParallelPlan,
	ForceApply bool
	LogLevel string
}

func NewProjectContext(
	ctx *Context,
	cmd Name,
	applyCmd string,
	planCmd string,
	projCfg valid.MergedProjectCfg,
	steps []valid.Step,
	policySets valid.PolicySets,
	escapedCommentArgs []string,
	contextFlags *ContextFlags,
	scope tally.Scope,
	pullStatus models.PullReqStatus,
) ProjectContext {
	var projectPlanStatus models.ProjectPlanStatus

	if ctx.PullStatus != nil {
		for _, project := range ctx.PullStatus.Projects {
			if project.ProjectName == projCfg.Name {
				projectPlanStatus = project.Status
				break
			}
		}
	}

	return ProjectContext{
		CommandName:          cmd,
		ApplyCmd:             applyCmd,
		BaseRepo:             ctx.Pull.BaseRepo,
		EscapedCommentArgs:   escapedCommentArgs,
		ParallelApplyEnabled: contextFlags.ParallelApply,
		ParallelPlanEnabled:  contextFlags.ParallelPlan,
		AutoplanEnabled:      projCfg.AutoplanEnabled,
		Steps:                steps,
		HeadRepo:             ctx.HeadRepo,
		Log:                  ctx.Log,
		Scope:                scope,
		ProjectPlanStatus:    projectPlanStatus,
		Pull:                 ctx.Pull,
		ProjectName:          projCfg.Name,
		ApplyRequirements:    projCfg.ApplyRequirements,
		RePlanCmd:            planCmd,
		RepoRelDir:           projCfg.RepoRelDir,
		RepoConfigVersion:    projCfg.RepoCfgVersion,
		TerraformVersion:     projCfg.TerraformVersion,
		User:                 ctx.User,
		ForceApply:           contextFlags.ForceApply,
		Workspace:            projCfg.Workspace,
		PolicySets:           policySets,
		Tags:                 projCfg.Tags,
		PullReqStatus:        pullStatus,
		JobID:                uuid.New().String(),
		RequestCtx:           ctx.RequestCtx,
		WorkflowModeType:     projCfg.WorkflowMode,
		InstallationToken:    ctx.InstallationToken,
	}
}

// ProjectContext defines the context for a plan or apply stage that will
// be executed for a project.
type ProjectContext struct {
	CommandName Name
	// ApplyCmd is the command that users should run to apply this plan. If
	// this is an apply then this will be empty.
	ApplyCmd string
	// ApplyRequirements is the list of requirements that must be satisfied
	// before we will run the apply stage.
	ApplyRequirements []string
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
	// Log is a logger that's been set up for this context.
	Log logging.Logger
	// Scope is the scope for reporting stats setup for this context
	Scope tally.Scope
	// PullReqStatus holds state about the PR that requires additional computation outside models.PullRequest
	PullReqStatus models.PullReqStatus
	// CurrentProjectPlanStatus is the status of the current project prior to this command.
	ProjectPlanStatus models.ProjectPlanStatus
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
	// TerraformVersion is the version of terraform we should use when executing
	// commands for this project. This can be set to nil in which case we will
	// use the default Atlantis terraform version.
	TerraformVersion *version.Version
	// Configuration metadata for a given project.
	Tags map[string]string
	// User is the user that triggered this command.
	User models.User
	// ForceApply is true when the apply should ignore apply_requirements.
	ForceApply bool
	// Workspace is the Terraform workspace this project is in. It will always
	// be set.
	Workspace string
	// PolicySets represent the policies that are run on the plan as part of the
	// policy check stage
	PolicySets valid.PolicySets
	// UUID for atlantis logs
	JobID string
	// RequestCtx is the context generated when request is first received from VCS
	RequestCtx context.Context
	// StatusID is used for consecutive status updates in the step runners
	StatusID string

	WorkflowModeType  valid.WorkflowModeType
	InstallationToken int64
}

// ProjectCloneDir creates relative path to clone the repo to. If we are running
// plans and apply in parallel we want to have a directory per project.
func (p ProjectContext) ProjectCloneDir() string {
	if p.ParallelPlanEnabled || p.ParallelApplyEnabled {
		return filepath.Join(p.ProjectName, p.Workspace)
	}

	return p.Workspace
}

// SetScope sets the scope of the stats object field. Note: we deliberately set this on the value
// instead of a pointer since we want scopes to mirror our function stack
func (p ProjectContext) SetScope(scope string) {
	p.Scope = p.Scope.SubScope(scope) //nolint
}

// GetShowResultFileName returns the filename (not the path) to store the tf show result
func (p ProjectContext) GetShowResultFileName() string {
	if p.ProjectName == "" {
		return fmt.Sprintf("%s.json", p.Workspace)
	}
	projName := strings.Replace(p.ProjectName, "/", planfileSlashReplace, -1)
	return fmt.Sprintf("%s-%s.json", projName, p.Workspace)
}

// Gets a unique identifier for the current pull request as a single string
func (p ProjectContext) PullInfof() string {
	return buildPullInfof(p.BaseRepo.FullName, p.Pull.Num, p.ProjectName, p.RepoRelDir, p.Workspace)
}
func buildPullInfof(repoName string, pullNum int, projectName string, relDir string, workspace string) string {
	projectIdentifier := getProjectIdentifier(relDir, projectName)
	return fmt.Sprintf("%s/%d/%s/%s", repoName, pullNum, projectIdentifier, workspace)
}

func getProjectIdentifier(relRepoDir string, projectName string) string {
	if projectName != "" {
		return projectName
	}
	return strings.ReplaceAll(relRepoDir, "/", "-")
}
