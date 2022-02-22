// Package valid contains the structs representing the atlantis.yaml config
// after it's been parsed and validated.
package valid

import (
	"fmt"
	"regexp"
	"strings"
)

// RepoCfg is the atlantis.yaml config after it's been parsed and validated.
type RepoCfg struct {
	// Version is the version of the atlantis YAML file.
	Version                   int
	Projects                  []Project
	Workflows                 map[string]Workflow
	PullRequestWorkflows      map[string]Workflow
	DeploymentWorkflows       map[string]Workflow
	PolicySets                PolicySets
	Automerge                 bool
	ParallelApply             bool
	ParallelPlan              bool
	ParallelPolicyCheck       bool
	DeleteSourceBranchOnMerge *bool
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

// FindProjectsByName returns all projects that match with name.
func (r RepoCfg) FindProjectsByName(name string) []Project {
	var ps []Project
	sanitizedName := "^" + name + "$"
	for _, p := range r.Projects {
		if p.Name != nil {
			if match, _ := regexp.MatchString(sanitizedName, *p.Name); match {
				ps = append(ps, p)
			}
		}
	}
	return ps
}

// ValidateAllowedOverrides iterates over projects and checks that each project
// is only using allowed overrides defined in global config
func (r RepoCfg) ValidateAllowedOverrides(allowedOverrides []string) error {
	for _, p := range r.Projects {
		if err := p.ValidateAllowedOverrides(allowedOverrides); err != nil {
			return err
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

// ValidateWorkflows ensures that all projects with custom workflow
// names exists either on repo level config or server level config
// Additionally it validates that workflow is allowed to be defined
func (r RepoCfg) ValidateWorkflows(
	globalWorkflows map[string]Workflow,
	allowedWorkflows []string,
	allowCustomWorkflows bool,
) error {
	// Check if the repo has set a workflow name that doesn't exist.
	for _, p := range r.Projects {
		if err := p.ValidateWorkflow(r.Workflows, globalWorkflows); err != nil {
			return err
		}
	}

	if len(allowedWorkflows) == 0 {
		return nil
	}

	// Check workflow is allowed
	for _, p := range r.Projects {
		if allowCustomWorkflows {
			if err := p.ValidateWorkflow(r.Workflows, map[string]Workflow{}); err == nil {
				break
			}
		}

		if err := p.ValidateWorkflowAllowed(allowedWorkflows); err != nil {
			return err
		}
	}

	return nil
}

// ValidatePRWorkflows ensures that all projects with custom
// pull_request workflow names exists either on server level config
// Additionally it validates that workflow is allowed to be defined
func (r RepoCfg) ValidatePRWorkflows(workflows map[string]Workflow, allowedWorkflows []string) error {
	for _, p := range r.Projects {
		if err := p.ValidatePRWorkflow(workflows); err != nil {
			return err
		}
	}
  
	if len(allowedWorkflows) == 0 {
		return nil
	}

	for _, p := range r.Projects {
		if err := p.ValidatePRWorkflowAllowed(allowedWorkflows); err != nil {
			return err
		}
	}
	return nil
}

// ValidateDeploymentWorkflows ensures that all projects with custom
// deployment workflow names exists either on server level config
// Additionally it validates that workflow is allowed to be defined
func (r RepoCfg) ValidateDeploymentWorkflows(workflows map[string]Workflow, allowedWorkflows []string) error {
	for _, p := range r.Projects {
		if err := p.ValidateDeploymentWorkflow(workflows); err != nil {
			return err
		}
	}

	if len(allowedWorkflows) == 0 {
		return nil
	}

	for _, p := range r.Projects {
		if err := p.ValidateDeploymentWorkflowAllowed(allowedWorkflows); err != nil {
			return err
		}
	}
	return nil
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
