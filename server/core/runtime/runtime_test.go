// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package runtime_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/runatlantis/atlantis/server/core/runtime"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

func TestGetPlanFilename(t *testing.T) {
	cases := []struct {
		workspace   string
		projectName string
		exp         string
	}{
		{
			"workspace",
			"",
			"workspace.tfplan",
		},
		{
			"workspace",
			"project",
			"project-workspace.tfplan",
		},
		{
			"workspace",
			"project/with/slash",
			"project::with::slash-workspace.tfplan",
		},
		{
			"workspace",
			"project with space",
			"project with space-workspace.tfplan",
		},
		{
			"workspace😀",
			"project😀",
			"project😀-workspace😀.tfplan",
		},
		// Previously we replaced invalid chars with -'s, however we now
		// rely on validation of the atlantis.yaml file to ensure the name's
		// don't contain chars that need to be url encoded. So now these
		// chars shouldn't get replaced.
		{
			"default",
			`all.invalid.chars \/"*?<>`,
			"all.invalid.chars \\::\"*?<>-default.tfplan",
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			Equals(t, c.exp, runtime.GetPlanFilename(c.workspace, c.projectName))
		})
	}
}

func TestEnsurePlanFileDir(t *testing.T) {
	projectPath := t.TempDir()
	planStoreDir := t.TempDir()
	ctx := command.ProjectContext{
		BaseRepo: models.Repo{
			FullName: "owner/repo",
		},
		LocalPlanStoreDir: planStoreDir,
		Pull: models.PullRequest{
			Num: 2,
		},
		RepoRelDir: "modules/app",
		Workspace:  "default",
	}

	Ok(t, runtime.EnsurePlanFileDir(ctx, projectPath))
	_, err := os.Stat(runtime.GetPlanFileDir(ctx, projectPath))
	Ok(t, err)

	ctx.BaseRepo.FullName = "../../outside"
	ErrContains(t, "plan file path traversal detected", runtime.EnsurePlanFileDir(ctx, projectPath))
}

func TestGetPlanFilePath(t *testing.T) {
	projectPath := filepath.Join("data", "repos", "owner", "repo", "2", "default", "modules", "app")
	ctx := command.ProjectContext{
		BaseRepo: models.Repo{
			FullName: "owner/repo",
		},
		LocalPlanStoreDir: filepath.Join("plans"),
		ProjectName:       "project/name",
		Pull: models.PullRequest{
			Num: 2,
		},
		RepoRelDir: "modules/app",
		Workspace:  "default",
	}

	Equals(t, filepath.Join("plans", "repos", "owner", "repo", "2", "default", "modules", "app", "project::name-default.tfplan"), runtime.GetPlanFilePath(ctx, projectPath))

	ctx.LocalPlanStoreDir = ""
	Equals(t, filepath.Join(projectPath, "project::name-default.tfplan"), runtime.GetPlanFilePath(ctx, projectPath))
func TestIsRemotePlan(t *testing.T) {
	const remotePlanHeader = "Atlantis: this plan was created by remote ops\n"

	cases := []struct {
		name     string
		contents []byte
		exp      bool
	}{
		{
			name: "nil plan",
		},
		{
			name:     "empty plan",
			contents: []byte{},
		},
		{
			name:     "same length without header",
			contents: []byte("Atlantis: this plan was created by local ops\n"),
		},
		{
			name:     "exact header",
			contents: []byte(remotePlanHeader),
			exp:      true,
		},
		{
			name:     "header with plan contents",
			contents: []byte(remotePlanHeader + "plan contents"),
			exp:      true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			Equals(t, c.exp, runtime.IsRemotePlan(c.contents))
		})
	}

	for prefixLen := range len(remotePlanHeader) {
		t.Run(fmt.Sprintf("truncated header length %d", prefixLen), func(t *testing.T) {
			Equals(t, false, runtime.IsRemotePlan([]byte(remotePlanHeader[:prefixLen])))
		})
	}
}

func TestProjectNameFromPlanfile(t *testing.T) {
	cases := []struct {
		workspace string
		filename  string
		exp       string
	}{
		{
			"workspace",
			"workspace.tfplan",
			"",
		},
		{
			"workspace",
			"project-workspace.tfplan",
			"project",
		},
		{
			"workspace",
			"project-workspace-workspace.tfplan",
			"project-workspace",
		},
		{
			"workspace",
			"project::with::slashes::-workspace.tfplan",
			"project/with/slashes/",
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			act, err := runtime.ProjectNameFromPlanfile(c.workspace, c.filename)
			Ok(t, err)
			Equals(t, c.exp, act)
		})
	}
}
