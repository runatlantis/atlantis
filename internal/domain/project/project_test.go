package project_test

import (
	"testing"

	"github.com/runatlantis/atlantis/internal/domain/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProject_ValidInput_CreatesProject(t *testing.T) {
	// Arrange
	id := project.ProjectID("test-project")
	name := "test-project"
	directory := "terraform/dev"
	workspace := "default"
	repo := project.RepositoryInfo{
		FullName: "owner/repo",
		Owner:    "owner",
		Name:     "repo",
	}

	// Act
	proj, err := project.NewProject(id, name, directory, workspace, repo)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, proj)
	assert.Equal(t, id, proj.ID())
	assert.Equal(t, name, proj.Name())
	assert.Equal(t, directory, proj.Directory())
	assert.Equal(t, workspace, proj.Workspace())
	assert.Equal(t, project.StatusPending, proj.Status())
}

func TestNewProject_InvalidProjectName_ReturnsError(t *testing.T) {
	// Arrange
	invalidName := "project with spaces!"
	
	// Act
	proj, err := project.NewProject("id", invalidName, "dir", "workspace", project.RepositoryInfo{})

	// Assert
	assert.Error(t, err)
	assert.Nil(t, proj)
	assert.Contains(t, err.Error(), "invalid project name")
}

func TestNewProject_InvalidDirectory_ReturnsError(t *testing.T) {
	// Arrange
	invalidDirectory := "../../../etc/passwd"
	
	// Act
	proj, err := project.NewProject("id", "name", invalidDirectory, "workspace", project.RepositoryInfo{})

	// Assert
	assert.Error(t, err)
	assert.Nil(t, proj)
	assert.Contains(t, err.Error(), "invalid directory")
}

func TestProject_CanExecutePlan_PendingProject_ReturnsNil(t *testing.T) {
	// Arrange
	proj := createValidProject(t)

	// Act
	err := proj.CanExecutePlan()

	// Assert
	assert.NoError(t, err)
}

func TestProject_MarkAsPlanned_PendingProject_UpdatesStatus(t *testing.T) {
	// Arrange
	proj := createValidProject(t)

	// Act
	err := proj.MarkAsPlanned()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, project.StatusPlanned, proj.Status())
}

func TestProject_MarkAsPlanned_NonPendingProject_ReturnsError(t *testing.T) {
	// Arrange
	proj := createValidProject(t)
	_ = proj.MarkAsPlanned() // First mark as planned

	// Act
	err := proj.MarkAsPlanned() // Try to mark as planned again

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "can only plan pending projects")
}

func TestProject_CanExecuteApply_PlannedProject_ReturnsNil(t *testing.T) {
	// Arrange
	proj := createValidProject(t)
	_ = proj.MarkAsPlanned()

	// Act
	err := proj.CanExecuteApply()

	// Assert
	assert.NoError(t, err)
}

func TestProject_CanExecuteApply_PendingProject_ReturnsError(t *testing.T) {
	// Arrange
	proj := createValidProject(t)

	// Act
	err := proj.CanExecuteApply()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project must be planned before apply")
}

func TestProject_MarkAsApplied_PlannedProject_UpdatesStatus(t *testing.T) {
	// Arrange
	proj := createValidProject(t)
	_ = proj.MarkAsPlanned()

	// Act
	err := proj.MarkAsApplied()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, project.StatusApplied, proj.Status())
}

func TestProject_MarkAsApplied_PendingProject_ReturnsError(t *testing.T) {
	// Arrange
	proj := createValidProject(t)

	// Act
	err := proj.MarkAsApplied()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project must be planned before apply")
}

// Test Helpers

func createValidProject(t *testing.T) *project.Project {
	t.Helper()
	
	proj, err := project.NewProject(
		project.ProjectID("test-project"),
		"test-project",
		"terraform/dev",
		"default",
		project.RepositoryInfo{
			FullName: "owner/repo",
			Owner:    "owner",
			Name:     "repo",
		},
	)
	require.NoError(t, err)
	return proj
} 