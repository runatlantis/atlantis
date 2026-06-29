// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"fmt"
	"strings"

	"github.com/runatlantis/atlantis/server/events/models"
)

func failedPlansApplyBlockFailure(failedProjects []models.ProjectStatus) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Applying all projects is blocked because %d plan(s) failed. Re-run failed plans with `atlantis plan --failed`, or apply a project that planned successfully with `atlantis apply -p <project>`.", len(failedProjects))
	for _, project := range failedProjects {
		b.WriteString("\n")
		b.WriteString(formatProjectStatus(project))
	}
	return b.String()
}

func formatProjectStatus(project models.ProjectStatus) string {
	parts := []string{
		fmt.Sprintf("dir: `%s`", project.RepoRelDir),
		fmt.Sprintf("workspace: `%s`", project.Workspace),
	}
	if project.ProjectName != "" {
		parts = append(parts, fmt.Sprintf("project: `%s`", project.ProjectName))
	}
	return "- " + strings.Join(parts, " ")
}
