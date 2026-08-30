package container

import (
	"go.etcd.io/bbolt"

	"github.com/runatlantis/atlantis/internal/application/commands"
	"github.com/runatlantis/atlantis/internal/domain/project"
	"github.com/runatlantis/atlantis/internal/infrastructure/persistence"
	"github.com/runatlantis/atlantis/internal/infrastructure/terraform"
	"github.com/runatlantis/atlantis/internal/infrastructure/vcs"
	"github.com/runatlantis/atlantis/internal/presentation/http"
)

// Container holds all application dependencies
type Container struct {
	// Infrastructure
	DB *bbolt.DB

	// Repositories
	ProjectRepo project.Repository

	// Services
	TerraformService commands.TerraformService
	VCSService       commands.VCSService
	ProjectService   project.Service

	// Use Case Handlers
	PlanProjectHandler *commands.PlanProjectHandler

	// HTTP Handlers
	ProjectHandler *http.ProjectHandler
}

// Config holds configuration for the container
type Config struct {
	DatabasePath string
	GitHubToken  string
	TerraformBin string
}

// NewContainer creates and wires up all dependencies
func NewContainer(config Config) (*Container, error) {
	// Initialize database
	db, err := bbolt.Open(config.DatabasePath, 0600, nil)
	if err != nil {
		return nil, err
	}

	// Repositories
	projectRepo := persistence.NewProjectRepository(db)

	// Infrastructure services
	terraformSvc := terraform.NewTerraformService(config.TerraformBin)
	vcsService := vcs.NewGitHubService(config.GitHubToken)

	// Domain services
	projectService := project.NewDomainService(projectRepo)

	// Use case handlers
	planProjectHandler := commands.NewPlanProjectHandler(
		projectRepo,
		projectService,
		terraformSvc,
		vcsService,
	)

	// HTTP handlers
	projectHandler := http.NewProjectHandler(planProjectHandler)

	return &Container{
		DB:                 db,
		ProjectRepo:        projectRepo,
		TerraformService:   terraformSvc,
		VCSService:         vcsService,
		ProjectService:     projectService,
		PlanProjectHandler: planProjectHandler,
		ProjectHandler:     projectHandler,
	}, nil
}

// Close closes all resources
func (c *Container) Close() error {
	if c.DB != nil {
		return c.DB.Close()
	}
	return nil
} 