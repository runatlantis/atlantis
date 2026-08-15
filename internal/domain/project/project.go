package project

import (
	"errors"
	"fmt"
	"regexp"
)

// Project represents a Terraform project in the domain
type Project struct {
	id          ProjectID
	name        string
	directory   string
	workspace   string
	repository  RepositoryInfo
	status      Status
	requirements Requirements
}

type ProjectID string
type Status string

const (
	StatusPending Status = "pending"
	StatusPlanned Status = "planned"
	StatusApplied Status = "applied"
	StatusErrored Status = "errored"
)

type RepositoryInfo struct {
	FullName string
	Owner    string
	Name     string
}

type Requirements struct {
	PlanRequirements  []string
	ApplyRequirements []string
}

// NewProject creates a new project with validation
func NewProject(id ProjectID, name, directory, workspace string, repo RepositoryInfo) (*Project, error) {
	if err := validateProjectName(name); err != nil {
		return nil, fmt.Errorf("invalid project name: %w", err)
	}

	if err := validateDirectory(directory); err != nil {
		return nil, fmt.Errorf("invalid directory: %w", err)
	}

	return &Project{
		id:         id,
		name:       name,
		directory:  directory,
		workspace:  workspace,
		repository: repo,
		status:     StatusPending,
	}, nil
}

// Business Rules
func (p *Project) CanExecutePlan() error {
	// Implement business rules for plan execution
	return nil
}

func (p *Project) CanExecuteApply() error {
	if p.status != StatusPlanned {
		return errors.New("project must be planned before apply")
	}
	return nil
}

func (p *Project) MarkAsPlanned() error {
	if p.status != StatusPending {
		return errors.New("can only plan pending projects")
	}
	p.status = StatusPlanned
	return nil
}

func (p *Project) MarkAsApplied() error {
	if err := p.CanExecuteApply(); err != nil {
		return err
	}
	p.status = StatusApplied
	return nil
}

// Getters
func (p *Project) ID() ProjectID { return p.id }
func (p *Project) Name() string { return p.name }
func (p *Project) Directory() string { return p.directory }
func (p *Project) Workspace() string { return p.workspace }
func (p *Project) Status() Status { return p.status }

// Validation functions
func validateProjectName(name string) error {
	if name == "" {
		return nil // Name is optional
	}
	
	// URL-safe characters only
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9._-]+$`, name)
	if !matched {
		return errors.New("must contain only URL safe characters")
	}
	return nil
}

func validateDirectory(dir string) error {
	if dir == "" {
		return errors.New("directory is required")
	}
	if regexp.MustCompile(`\.\.`).MatchString(dir) {
		return errors.New("directory cannot contain '..'")
	}
	return nil
} 