package commands

import (
	"context"
	"fmt"

	"github.com/runatlantis/atlantis/internal/domain/project"
)

// PlanProjectCommand represents the input for planning a project
type PlanProjectCommand struct {
	ProjectID   project.ProjectID
	UserID      string
	PullRequest PullRequestInfo
}

type PullRequestInfo struct {
	Number     int
	Repository string
	HeadSHA    string
}

// PlanProjectResult represents the output of planning a project
type PlanProjectResult struct {
	ProjectID project.ProjectID
	PlanFile  string
	Output    string
	Success   bool
}

// PlanProjectHandler handles the plan project use case
type PlanProjectHandler struct {
	projectRepo    project.Repository
	projectService project.Service
	terraformSvc   TerraformService
	vcsService     VCSService
}

// TerraformService interface for infrastructure concerns
type TerraformService interface {
	Plan(ctx context.Context, request TerraformPlanRequest) (*TerraformPlanResult, error)
}

type TerraformPlanRequest struct {
	WorkingDir string
	Workspace  string
	ExtraArgs  []string
}

type TerraformPlanResult struct {
	PlanFile string
	Output   string
}

// VCSService interface for version control operations
type VCSService interface {
	UpdateCommitStatus(ctx context.Context, repo string, sha string, status CommitStatus) error
}

type CommitStatus struct {
	State       string
	Description string
	Context     string
}

func NewPlanProjectHandler(
	projectRepo project.Repository,
	projectService project.Service,
	terraformSvc TerraformService,
	vcsService VCSService,
) *PlanProjectHandler {
	return &PlanProjectHandler{
		projectRepo:    projectRepo,
		projectService: projectService,
		terraformSvc:   terraformSvc,
		vcsService:     vcsService,
	}
}

func (h *PlanProjectHandler) Handle(ctx context.Context, cmd PlanProjectCommand) (*PlanProjectResult, error) {
	// 1. Retrieve project
	proj, err := h.projectRepo.FindByID(ctx, cmd.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to find project: %w", err)
	}

	// 2. Validate business rules
	if err := proj.CanExecutePlan(); err != nil {
		return nil, fmt.Errorf("cannot execute plan: %w", err)
	}

	// 3. Update commit status to pending
	err = h.vcsService.UpdateCommitStatus(ctx, cmd.PullRequest.Repository, cmd.PullRequest.HeadSHA, CommitStatus{
		State:       "pending",
		Description: "Planning in progress",
		Context:     "atlantis/plan",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update commit status: %w", err)
	}

	// 4. Execute terraform plan
	planResult, err := h.terraformSvc.Plan(ctx, TerraformPlanRequest{
		WorkingDir: proj.Directory(),
		Workspace:  proj.Workspace(),
	})
	if err != nil {
		// Update commit status to failure
		_ = h.vcsService.UpdateCommitStatus(ctx, cmd.PullRequest.Repository, cmd.PullRequest.HeadSHA, CommitStatus{
			State:       "failure",
			Description: "Plan failed",
			Context:     "atlantis/plan",
		})
		return &PlanProjectResult{
			ProjectID: cmd.ProjectID,
			Success:   false,
			Output:    err.Error(),
		}, nil
	}

	// 5. Update project status
	if err := proj.MarkAsPlanned(); err != nil {
		return nil, fmt.Errorf("failed to mark project as planned: %w", err)
	}

	// 6. Save project
	if err := h.projectRepo.Save(ctx, proj); err != nil {
		return nil, fmt.Errorf("failed to save project: %w", err)
	}

	// 7. Update commit status to success
	err = h.vcsService.UpdateCommitStatus(ctx, cmd.PullRequest.Repository, cmd.PullRequest.HeadSHA, CommitStatus{
		State:       "success",
		Description: "Plan completed successfully",
		Context:     "atlantis/plan",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update commit status: %w", err)
	}

	return &PlanProjectResult{
		ProjectID: cmd.ProjectID,
		PlanFile:  planResult.PlanFile,
		Output:    planResult.Output,
		Success:   true,
	}, nil
} 