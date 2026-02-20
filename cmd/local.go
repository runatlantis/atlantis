// Copyright 2025 The Atlantis Authors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/runatlantis/atlantis/server/core/config"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/spf13/cobra"
)

// LocalCmd is the parent for commands that run Atlantis workflows locally.
type LocalCmd struct {
	Logger logging.SimpleLogging
}

// Init returns the runnable cobra command.
func (l *LocalCmd) Init() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "local",
		Short: "Run Atlantis workflows locally without a server",
	}
	localPlan := &LocalPlanCmd{Logger: l.Logger}
	cmd.AddCommand(localPlan.Init())
	return cmd
}

const (
	// LocalPlanDirFlag is the flag for the directory to run the local plan in.
	LocalPlanDirFlag = "dir"
	// LocalPlanProjectFlag is the flag for the project to plan.
	LocalPlanProjectFlag = "project"
	// LocalPlanWorkspaceFlag is the flag for the workspace to plan.
	LocalPlanWorkspaceFlag = "workspace"
	// LocalPlanVerboseFlag is the flag for verbose output.
	LocalPlanVerboseFlag = "verbose"
)

// LocalPlanCmd runs terraform plan locally using the atlantis.yaml config.
type LocalPlanCmd struct {
	Logger logging.SimpleLogging
}

// Init returns the runnable cobra command.
func (l *LocalPlanCmd) Init() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Run terraform plan locally for projects matching local changes",
		Long: `Parses the local atlantis.yaml config, determines which projects are
affected by local file changes, and runs terraform plan for each affected
project.

This allows you to test your atlantis.yaml configuration and see how
Atlantis would plan your projects without running a server or setting up
a webhook.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, _ := cmd.Flags().GetString(LocalPlanDirFlag)
			project, _ := cmd.Flags().GetString(LocalPlanProjectFlag)
			workspace, _ := cmd.Flags().GetString(LocalPlanWorkspaceFlag)
			verbose, _ := cmd.Flags().GetBool(LocalPlanVerboseFlag)
			return l.run(dir, project, workspace, verbose)
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	cmd.Flags().StringP(LocalPlanDirFlag, "d", "", "Path to the repository root. Defaults to the current directory")
	cmd.Flags().StringP(LocalPlanProjectFlag, "p", "", "Name of a specific atlantis.yaml project to plan")
	cmd.Flags().StringP(LocalPlanWorkspaceFlag, "w", "", "Terraform workspace to plan. Overrides the workspace in atlantis.yaml")
	cmd.Flags().BoolP(LocalPlanVerboseFlag, "v", false, "Print verbose output including each command run")
	return cmd
}

func (l *LocalPlanCmd) run(dir, project, workspace string, verbose bool) error {
	// Resolve the directory.
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("getting current directory: %w", err)
		}
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolving directory %q: %w", dir, err)
	}

	// Find the git root (repo root).
	repoRoot, err := gitRoot(absDir)
	if err != nil {
		return fmt.Errorf("finding git root from %q: %w", absDir, err)
	}

	// Parse the atlantis.yaml repo config.
	parser := &config.ParserValidator{}
	globalCfg := valid.NewGlobalCfgFromArgs(valid.GlobalCfgArgs{AllowAllRepoSettings: true})
	repoCfg, err := parser.ParseRepoCfg(repoRoot, globalCfg, "", "")
	if err != nil {
		return fmt.Errorf("parsing atlantis.yaml: %w", err)
	}

	// Get locally modified files.
	modifiedFiles, err := localModifiedFiles(repoRoot)
	if err != nil {
		return fmt.Errorf("getting modified files: %w", err)
	}
	if verbose {
		fmt.Printf("Modified files: %v\n\n", modifiedFiles)
	}

	// Determine which projects to plan.
	projectsToRun, err := l.selectProjects(repoCfg, repoRoot, project, workspace, modifiedFiles)
	if err != nil {
		return err
	}
	if len(projectsToRun) == 0 {
		fmt.Println("No projects to plan based on the current changes.")
		return nil
	}

	fmt.Printf("Planning %d project(s)...\n\n", len(projectsToRun))

	var planErrs []string
	for _, p := range projectsToRun {
		if err := runLocalProjectPlan(repoRoot, p, verbose); err != nil {
			planErrs = append(planErrs, fmt.Sprintf("project dir=%q workspace=%q: %s", p.Dir, p.Workspace, err))
		}
	}
	if len(planErrs) > 0 {
		return fmt.Errorf("the following projects had errors:\n  %s", strings.Join(planErrs, "\n  "))
	}
	return nil
}

// selectProjects returns the list of projects to plan, filtered by the project
// name and workspace flags and matched against the list of modified files.
func (l *LocalPlanCmd) selectProjects(repoCfg valid.RepoCfg, repoRoot, project, workspace string, modifiedFiles []string) ([]valid.Project, error) {
	if project != "" {
		p := repoCfg.FindProjectByName(project)
		if p == nil {
			return nil, fmt.Errorf("project %q not found in atlantis.yaml", project)
		}
		return []valid.Project{*p}, nil
	}

	finder := &events.DefaultProjectFinder{}
	projects, err := finder.DetermineProjectsViaConfig(l.Logger, modifiedFiles, repoCfg, repoRoot, nil)
	if err != nil {
		return nil, fmt.Errorf("determining affected projects: %w", err)
	}

	if workspace == "" {
		return projects, nil
	}
	// Filter by workspace if the flag was set.
	var filtered []valid.Project
	for _, p := range projects {
		if p.Workspace == workspace {
			filtered = append(filtered, p)
		}
	}
	return filtered, nil
}

// runLocalProjectPlan runs terraform init and plan for a single project.
func runLocalProjectPlan(repoRoot string, project valid.Project, verbose bool) error {
	projectDir := filepath.Join(repoRoot, project.Dir)

	name := project.Dir
	if project.Name != nil {
		name = *project.Name
	}
	fmt.Printf("=== Planning project: %s (dir: %s, workspace: %s) ===\n", name, project.Dir, project.Workspace)

	if err := execCmd(projectDir, verbose, "terraform", "init", "-input=false"); err != nil {
		return fmt.Errorf("terraform init: %w", err)
	}

	if project.Workspace != "" && project.Workspace != "default" {
		if err := execCmd(projectDir, verbose, "terraform", "workspace", "select", "-or-create", project.Workspace); err != nil {
			return fmt.Errorf("terraform workspace select %q: %w", project.Workspace, err)
		}
	}

	if err := execCmd(projectDir, verbose, "terraform", "plan", "-input=false"); err != nil {
		return fmt.Errorf("terraform plan: %w", err)
	}

	fmt.Println()
	return nil
}

// gitRoot returns the root of the git repository containing dir.
func gitRoot(dir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel") // nolint: gosec
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("running git rev-parse: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// localModifiedFiles returns the list of files that have been modified in the
// working tree (staged and unstaged) relative to the repo root.
//
// The -z flag instructs git to output NUL-terminated entries without quoting
// filenames, which correctly handles filenames with spaces or special characters.
func localModifiedFiles(repoRoot string) ([]string, error) {
	cmd := exec.Command("git", "status", "--porcelain", "-z") // nolint: gosec
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("running git status: %w", err)
	}

	var files []string
	// With -z, each entry is NUL-terminated. For renamed files the format is:
	// "XY new_name\0old_name\0"; for all others it is "XY filename\0".
	// We collect every field that looks like a status entry (has the "XY "
	// prefix) and take only the destination name so that renames are correctly
	// represented by their new path.
	for _, entry := range strings.Split(string(out), "\x00") {
		if len(entry) < 4 {
			continue
		}
		// Only process entries that start with the two-character status code
		// followed by a space.
		if entry[2] != ' ' {
			continue
		}
		file := entry[3:]
		if file != "" {
			files = append(files, file)
		}
	}
	return files, nil
}

// execCmd runs a command in the given directory and streams output to stdout/stderr.
func execCmd(dir string, verbose bool, name string, args ...string) error {
	if verbose {
		fmt.Printf("Running: %s %s\n", name, strings.Join(args, " "))
	}
	cmd := exec.Command(name, args...) // nolint: gosec
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
