// Copyright 2025 The Atlantis Authors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

// TestLocalCmd_Init checks that the local command and its plan sub-command
// are registered correctly.
func TestLocalCmd_Init(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	lc := &LocalCmd{Logger: logger}
	cmd := lc.Init()

	Equals(t, "local", cmd.Use)
	Equals(t, 1, len(cmd.Commands()))
	Equals(t, "plan", cmd.Commands()[0].Use)
}

// TestLocalPlanCmd_Init verifies that all expected flags are present on
// the plan sub-command.
func TestLocalPlanCmd_Init(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	lp := &LocalPlanCmd{Logger: logger}
	cmd := lp.Init()

	for _, flagName := range []string{
		LocalPlanDirFlag,
		LocalPlanProjectFlag,
		LocalPlanWorkspaceFlag,
		LocalPlanVerboseFlag,
	} {
		Assert(t, cmd.Flags().Lookup(flagName) != nil, "expected flag %q to be registered", flagName)
	}
}

// TestLocalPlanCmd_SelectProjects_ByName verifies that when the --project flag
// is supplied, only that project is returned.
func TestLocalPlanCmd_SelectProjects_ByName(t *testing.T) {
	name := "myproject"
	repoCfg := valid.RepoCfg{
		Projects: []valid.Project{
			{Dir: ".", Workspace: "default", Name: &name},
			{Dir: "other", Workspace: "default"},
		},
	}

	logger := logging.NewNoopLogger(t)
	lp := &LocalPlanCmd{Logger: logger}
	projects, err := lp.selectProjects(repoCfg, t.TempDir(), "myproject", "", nil)
	Ok(t, err)
	Equals(t, 1, len(projects))
	Equals(t, name, *projects[0].Name)
}

// TestLocalPlanCmd_SelectProjects_UnknownProject verifies that an error is
// returned when the specified project name does not exist.
func TestLocalPlanCmd_SelectProjects_UnknownProject(t *testing.T) {
	repoCfg := valid.RepoCfg{
		Projects: []valid.Project{
			{Dir: ".", Workspace: "default"},
		},
	}

	logger := logging.NewNoopLogger(t)
	lp := &LocalPlanCmd{Logger: logger}
	_, err := lp.selectProjects(repoCfg, t.TempDir(), "doesnotexist", "", nil)
	Assert(t, err != nil, "expected an error for unknown project name")
}

// TestLocalPlanCmd_SelectProjects_ByWorkspace verifies that when the
// --workspace flag is supplied, only projects in that workspace are returned.
func TestLocalPlanCmd_SelectProjects_ByWorkspace(t *testing.T) {
	// Create a temp dir that represents the repo root with a project sub-dir.
	repoDir := t.TempDir()
	projectDir := filepath.Join(repoDir, "infra")
	Ok(t, os.MkdirAll(projectDir, 0700))

	repoCfg := valid.RepoCfg{
		Projects: []valid.Project{
			{
				Dir:       "infra",
				Workspace: "staging",
				Autoplan:  valid.Autoplan{Enabled: true, WhenModified: []string{"**/*.tf"}},
			},
			{
				Dir:       "infra",
				Workspace: "prod",
				Autoplan:  valid.Autoplan{Enabled: true, WhenModified: []string{"**/*.tf"}},
			},
		},
	}
	modifiedFiles := []string{"infra/main.tf"}

	logger := logging.NewNoopLogger(t)
	lp := &LocalPlanCmd{Logger: logger}
	projects, err := lp.selectProjects(repoCfg, repoDir, "", "staging", modifiedFiles)
	Ok(t, err)
	Equals(t, 1, len(projects))
	Equals(t, "staging", projects[0].Workspace)
}

// TestLocalPlanCmd_SelectProjects_NoMatches verifies that an empty slice is
// returned when no modified files match any project.
func TestLocalPlanCmd_SelectProjects_NoMatches(t *testing.T) {
	repoDir := t.TempDir()
	projectDir := filepath.Join(repoDir, "infra")
	Ok(t, os.MkdirAll(projectDir, 0700))

	repoCfg := valid.RepoCfg{
		Projects: []valid.Project{
			{
				Dir:      "infra",
				Autoplan: valid.Autoplan{Enabled: true, WhenModified: []string{"**/*.tf"}},
			},
		},
	}
	// Only a README was modified, which doesn't match **/*.tf.
	modifiedFiles := []string{"README.md"}

	logger := logging.NewNoopLogger(t)
	lp := &LocalPlanCmd{Logger: logger}
	projects, err := lp.selectProjects(repoCfg, repoDir, "", "", modifiedFiles)
	Ok(t, err)
	Equals(t, 0, len(projects))
}
