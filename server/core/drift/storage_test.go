// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package drift_test

import (
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/core/drift"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

func TestInMemoryStorage_Store(t *testing.T) {
	storage := drift.NewInMemoryStorage()

	projectDrift := models.ProjectDrift{
		ProjectName: "test-project",
		Path:        "modules/vpc",
		Workspace:   "default",
		Ref:         "main",
		Drift: models.DriftSummary{
			HasDrift: true,
			ToAdd:    2,
		},
		LastChecked: time.Now(),
	}

	err := storage.Store("owner/repo", projectDrift)
	Ok(t, err)

	// Verify it was stored
	results, err := storage.Get("owner/repo", drift.GetOptions{})
	Ok(t, err)
	Equals(t, 1, len(results))
	Equals(t, projectDrift.ProjectName, results[0].ProjectName)
}

func TestInMemoryStorage_StoreOverwrite(t *testing.T) {
	storage := drift.NewInMemoryStorage()

	// Store initial drift
	projectDrift1 := models.ProjectDrift{
		ProjectName: "test-project",
		Path:        "modules/vpc",
		Workspace:   "default",
		Drift: models.DriftSummary{
			HasDrift: true,
			ToAdd:    2,
		},
		LastChecked: time.Now(),
	}
	err := storage.Store("owner/repo", projectDrift1)
	Ok(t, err)

	// Store updated drift for same project
	projectDrift2 := models.ProjectDrift{
		ProjectName: "test-project",
		Path:        "modules/vpc",
		Workspace:   "default",
		Drift: models.DriftSummary{
			HasDrift: false,
			ToAdd:    0,
		},
		LastChecked: time.Now().Add(time.Hour),
	}
	err = storage.Store("owner/repo", projectDrift2)
	Ok(t, err)

	// Verify only one entry exists with updated data
	results, err := storage.Get("owner/repo", drift.GetOptions{})
	Ok(t, err)
	Equals(t, 1, len(results))
	Equals(t, false, results[0].Drift.HasDrift)
}

func TestInMemoryStorage_GetEmpty(t *testing.T) {
	storage := drift.NewInMemoryStorage()

	results, err := storage.Get("owner/repo", drift.GetOptions{})
	Ok(t, err)
	Equals(t, 0, len(results))
}

func TestInMemoryStorage_GetByProjectName(t *testing.T) {
	storage := drift.NewInMemoryStorage()

	// Store multiple projects
	storage.Store("owner/repo", models.ProjectDrift{ //nolint:errcheck
		ProjectName: "project1",
		Path:        "path1",
		Workspace:   "default",
		LastChecked: time.Now(),
	})
	storage.Store("owner/repo", models.ProjectDrift{ //nolint:errcheck
		ProjectName: "project2",
		Path:        "path2",
		Workspace:   "default",
		LastChecked: time.Now(),
	})

	// Get by project name
	results, err := storage.Get("owner/repo", drift.GetOptions{ProjectName: "project1"})
	Ok(t, err)
	Equals(t, 1, len(results))
	Equals(t, "project1", results[0].ProjectName)
}

func TestInMemoryStorage_GetByPath(t *testing.T) {
	storage := drift.NewInMemoryStorage()

	// Store multiple projects
	storage.Store("owner/repo", models.ProjectDrift{ //nolint:errcheck
		ProjectName: "project1",
		Path:        "modules/vpc",
		Workspace:   "default",
		LastChecked: time.Now(),
	})
	storage.Store("owner/repo", models.ProjectDrift{ //nolint:errcheck
		ProjectName: "project2",
		Path:        "modules/ec2",
		Workspace:   "default",
		LastChecked: time.Now(),
	})

	// Get by path
	results, err := storage.Get("owner/repo", drift.GetOptions{Path: "modules/vpc"})
	Ok(t, err)
	Equals(t, 1, len(results))
	Equals(t, "modules/vpc", results[0].Path)
}

func TestInMemoryStorage_GetByWorkspace(t *testing.T) {
	storage := drift.NewInMemoryStorage()

	// Store multiple projects
	storage.Store("owner/repo", models.ProjectDrift{ //nolint:errcheck
		ProjectName: "project1",
		Path:        "path1",
		Workspace:   "default",
		LastChecked: time.Now(),
	})
	storage.Store("owner/repo", models.ProjectDrift{ //nolint:errcheck
		ProjectName: "project1",
		Path:        "path1",
		Workspace:   "staging",
		LastChecked: time.Now(),
	})

	// Get by workspace
	results, err := storage.Get("owner/repo", drift.GetOptions{Workspace: "staging"})
	Ok(t, err)
	Equals(t, 1, len(results))
	Equals(t, "staging", results[0].Workspace)
}

func TestInMemoryStorage_GetByMaxAge(t *testing.T) {
	storage := drift.NewInMemoryStorage()

	// Store old drift
	storage.Store("owner/repo", models.ProjectDrift{ //nolint:errcheck
		ProjectName: "old-project",
		Path:        "path1",
		Workspace:   "default",
		LastChecked: time.Now().Add(-2 * time.Hour),
	})
	// Store recent drift
	storage.Store("owner/repo", models.ProjectDrift{ //nolint:errcheck
		ProjectName: "recent-project",
		Path:        "path2",
		Workspace:   "default",
		LastChecked: time.Now(),
	})

	// Get with max age of 1 hour
	results, err := storage.Get("owner/repo", drift.GetOptions{MaxAge: time.Hour})
	Ok(t, err)
	Equals(t, 1, len(results))
	Equals(t, "recent-project", results[0].ProjectName)
}

func TestInMemoryStorage_Delete(t *testing.T) {
	storage := drift.NewInMemoryStorage()

	// Store drift
	storage.Store("owner/repo", models.ProjectDrift{ //nolint:errcheck
		ProjectName: "test-project",
		Path:        "path1",
		Workspace:   "default",
		LastChecked: time.Now(),
	})

	// Delete all drift for repository
	err := storage.Delete("owner/repo", "")
	Ok(t, err)

	// Verify deleted
	results, err := storage.Get("owner/repo", drift.GetOptions{})
	Ok(t, err)
	Equals(t, 0, len(results))
}

func TestInMemoryStorage_DeleteByProject(t *testing.T) {
	storage := drift.NewInMemoryStorage()

	// Store multiple projects
	storage.Store("owner/repo", models.ProjectDrift{ //nolint:errcheck
		ProjectName: "project1",
		Path:        "path1",
		Workspace:   "default",
		LastChecked: time.Now(),
	})
	storage.Store("owner/repo", models.ProjectDrift{ //nolint:errcheck
		ProjectName: "project2",
		Path:        "path2",
		Workspace:   "default",
		LastChecked: time.Now(),
	})

	// Delete specific project
	err := storage.Delete("owner/repo", "project1")
	Ok(t, err)

	// Verify only project2 remains
	results, err := storage.Get("owner/repo", drift.GetOptions{})
	Ok(t, err)
	Equals(t, 1, len(results))
	Equals(t, "project2", results[0].ProjectName)
}

func TestInMemoryStorage_DeleteNonExistent(t *testing.T) {
	storage := drift.NewInMemoryStorage()

	// Delete from non-existent repo should not error
	err := storage.Delete("owner/repo", "")
	Ok(t, err)
}

func TestInMemoryStorage_GetAll(t *testing.T) {
	storage := drift.NewInMemoryStorage()

	// Store drift in multiple repos
	storage.Store("owner/repo1", models.ProjectDrift{ //nolint:errcheck
		ProjectName: "project1",
		Path:        "path1",
		Workspace:   "default",
		LastChecked: time.Now(),
	})
	storage.Store("owner/repo2", models.ProjectDrift{ //nolint:errcheck
		ProjectName: "project2",
		Path:        "path2",
		Workspace:   "default",
		LastChecked: time.Now(),
	})

	// Get all
	results, err := storage.GetAll()
	Ok(t, err)
	Equals(t, 2, len(results))
	Assert(t, len(results["owner/repo1"]) == 1, "should have 1 project in repo1")
	Assert(t, len(results["owner/repo2"]) == 1, "should have 1 project in repo2")
}

func TestInMemoryStorage_GetAllEmpty(t *testing.T) {
	storage := drift.NewInMemoryStorage()

	results, err := storage.GetAll()
	Ok(t, err)
	Equals(t, 0, len(results))
}
