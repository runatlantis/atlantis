// Package valid contains the structs representing the atlantis.yaml config
// after it's been parsed and validated.
package valid

import (
	"fmt"
	"strings"

	version "github.com/hashicorp/go-version"
)

// RepoCfg is the atlantis.yaml config after it's been parsed and validated.
type RepoCfg struct {
	// Version is the version of the atlantis YAML file.
	Version             int
	Projects            []Project
	Workflows           map[string]Workflow
	PolicySets          PolicySets
	Automerge           bool
	ParallelApply       bool
	ParallelPlan        bool
	ParallelPlans       bool
	ProjectLocks        bool
	ParallelPolicyCheck bool
}

func (r RepoCfg) FindProjectsByDirWorkspace(repoRelDir string, workspace string) []Project {
	var ps []Project
	for _, p := range r.Projects {
		if p.Dir == repoRelDir && p.Workspace == workspace {
			ps = append(ps, p)
		}
	}
	return ps
}

// FindProjectsByDir returns all projects that are in dir.
func (r RepoCfg) FindProjectsByDir(dir string) []Project {
	var ps []Project
	for _, p := range r.Projects {
		if p.Dir == dir {
			ps = append(ps, p)
		}
	}
	return ps
}

func (r RepoCfg) FindProjectByName(name string) *Project {
	for _, p := range r.Projects {
		if p.Name != nil && *p.Name == name {
			return &p
		}
	}
	return nil
}

// validateWorkspaceAllowed returns an error if repoCfg defines projects in
// repoRelDir but none of them use workspace. We want this to be an error
// because if users have gone to the trouble of defining projects in repoRelDir
// then it's likely that if we're running a command for a workspace that isn't
// defined then they probably just typed the workspace name wrong.
func (r RepoCfg) ValidateWorkspaceAllowed(repoRelDir string, workspace string) error {
	projects := r.FindProjectsByDir(repoRelDir)

	// If that directory doesn't have any projects configured then we don't
	// enforce workspace names.
	if len(projects) == 0 {
		return nil
	}

	var configuredSpaces []string
	for _, p := range projects {
		if p.Workspace == workspace {
			return nil
		}
		configuredSpaces = append(configuredSpaces, p.Workspace)
	}

	return fmt.Errorf(
		"running commands in workspace %q is not allowed because this"+
			" directory is only configured for the following workspaces: %s",
		workspace,
		strings.Join(configuredSpaces, ", "),
	)
}

type Project struct {
	Dir               string
	Workspace         string
	Name              *string
	WorkflowName      *string
	TerraformVersion  *version.Version
	Autoplan          Autoplan
	ApplyRequirements []string
}

// GetName returns the name of the project or an empty string if there is no
// project name.
func (p Project) GetName() string {
	if p.Name != nil {
		return *p.Name
	}
	return ""
}

type Autoplan struct {
	WhenModified []string
	Enabled      bool
}

type Stage struct {
	Steps []Step
}

type Step struct {
	StepName  string
	ExtraArgs []string
	// RunCommand is either a custom run step or the command to run
	// during an env step to populate the environment variable dynamically.
	RunCommand string
	// EnvVarName is the name of the
	// environment variable that should be set by this step.
	EnvVarName string
	// EnvVarValue is the value to set EnvVarName to.
	EnvVarValue string
}

type Workflow struct {
	Name        string
	Apply       Stage
	Plan        Stage
	PolicyCheck Stage
}
