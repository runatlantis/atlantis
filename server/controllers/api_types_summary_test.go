// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers_test

import (
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/controllers"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

func TestDriftSummaryAPICountsOverlappingDriftAndErrors(t *testing.T) {
	projects := []models.ProjectDrift{
		{
			ProjectName: "drift-with-error",
			Drift:       models.DriftSummary{HasDrift: true, ToChange: 1},
			Error:       "policy failed",
		},
		{
			ProjectName: "clean",
			Drift:       models.DriftSummary{HasDrift: false},
		},
	}

	detection := controllers.NewDriftDetectionResultAPI(&models.DriftDetectionResult{
		Repository:        "owner/repo",
		Projects:          projects,
		TotalProjects:     len(projects),
		ProjectsWithDrift: 1,
		DetectedAt:        time.Now(),
	})
	Equals(t, 2, detection.Summary.TotalProjects)
	Equals(t, 1, detection.Summary.ProjectsWithDrift)
	Equals(t, 1, detection.Summary.ProjectsWithErrors)
	Equals(t, 1, detection.Summary.ProjectsWithoutDrift)

	status := controllers.NewDriftStatusAPI(models.DriftStatusResponse{
		Repository:        "owner/repo",
		Projects:          projects,
		TotalProjects:     len(projects),
		ProjectsWithDrift: 1,
		CheckedAt:         time.Now(),
	})
	Equals(t, 2, status.Summary.TotalProjects)
	Equals(t, 1, status.Summary.ProjectsWithDrift)
	Equals(t, 1, status.Summary.ProjectsWithErrors)
	Equals(t, 1, status.Summary.ProjectsWithoutDrift)
}
