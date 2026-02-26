package project

import "context"

// Repository defines the interface for project persistence
type Repository interface {
	Save(ctx context.Context, project *Project) error
	FindByID(ctx context.Context, id ProjectID) (*Project, error)
	FindByRepositoryAndPR(ctx context.Context, repoFullName string, pullNumber int) ([]*Project, error)
	Delete(ctx context.Context, id ProjectID) error
	List(ctx context.Context, filters ListFilters) ([]*Project, error)
}

type ListFilters struct {
	Repository string
	Status     Status
	Workspace  string
}

// DomainService for complex business operations
type Service interface {
	CreateProject(ctx context.Context, request CreateProjectRequest) (*Project, error)
	PlanProject(ctx context.Context, id ProjectID) error
	ApplyProject(ctx context.Context, id ProjectID) error
	ValidateProjectDependencies(ctx context.Context, project *Project) error
}

type CreateProjectRequest struct {
	Name       string
	Directory  string
	Workspace  string
	Repository RepositoryInfo
} 