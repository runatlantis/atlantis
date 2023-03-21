package valid

import (
	"fmt"

	version "github.com/hashicorp/go-version"
)

type workflowType string

const (
	DefaultWorkflowType     workflowType = "workflow"
	PullRequestWorkflowType workflowType = "pull_request_workflow"
	DeploymentWorkflowType  workflowType = "deployment_workflow"
)

// TODO: rename to root
type Project struct {
	Dir                     string
	Workspace               string
	Name                    *string
	WorkflowName            *string
	PullRequestWorkflowName *string
	DeploymentWorkflowName  *string
	TerraformVersion        *version.Version
	Autoplan                Autoplan
	ApplyRequirements       []string
	Tags                    map[string]string
	WorkflowModeType        *string
}

// GetName returns the name of the project or an empty string if there is no
// project name.
func (p Project) GetName() string {
	if p.Name != nil {
		return *p.Name
	}
	// TODO
	// Upstream atlantis only requires project name to be set if there's more than one project
	// with same dir and workspace. If a project name has not been set, we'll use the dir and
	// workspace to build project key.
	// Source: https://www.runatlantis.io/docs/repo-level-atlantis-yaml.html#reference
	return ""
}

func (p Project) ValidateAllowedOverrides(allowedOverrides []string) error {
	if p.WorkflowName != nil && !sliceContains(allowedOverrides, WorkflowKey) {
		return fmt.Errorf("repo config not allowed to set '%s' key: server-side config needs '%s: [%s]'", WorkflowKey, AllowedOverridesKey, WorkflowKey)
	}
	if p.PullRequestWorkflowName != nil && !sliceContains(allowedOverrides, PullRequestWorkflowKey) {
		return fmt.Errorf("repo config not allowed to set '%s' key: server-side config needs '%s: [%s]'", PullRequestWorkflowKey, AllowedOverridesKey, PullRequestWorkflowKey)
	}
	if p.DeploymentWorkflowName != nil && !sliceContains(allowedOverrides, DeploymentWorkflowKey) {
		return fmt.Errorf("repo config not allowed to set '%s' key: server-side config needs '%s: [%s]'", DeploymentWorkflowKey, AllowedOverridesKey, DeploymentWorkflowKey)
	}
	if p.ApplyRequirements != nil && !sliceContains(allowedOverrides, ApplyRequirementsKey) {
		return fmt.Errorf("repo config not allowed to set '%s' key: server-side config needs '%s: [%s]'", ApplyRequirementsKey, AllowedOverridesKey, ApplyRequirementsKey)
	}

	return nil
}

func (p Project) getWorkflowName(workflowType workflowType) *string {
	var name *string
	switch workflowType {
	case DefaultWorkflowType:
		name = p.WorkflowName
	case PullRequestWorkflowType:
		name = p.PullRequestWorkflowName
	case DeploymentWorkflowType:
		name = p.DeploymentWorkflowName
	}
	return name
}

func (p Project) ValidateWorkflow(repoWorkflows map[string]Workflow, globalWorkflows map[string]Workflow) error {
	return p.validateWorkflowForType(DefaultWorkflowType, repoWorkflows, globalWorkflows)
}

func (p Project) ValidatePRWorkflow(globalWorkflows map[string]Workflow) error {
	return p.validateWorkflowForType(PullRequestWorkflowType, map[string]Workflow{}, globalWorkflows)
}

func (p Project) ValidateDeploymentWorkflow(globalWorkflows map[string]Workflow) error {
	return p.validateWorkflowForType(DeploymentWorkflowType, map[string]Workflow{}, globalWorkflows)
}

func (p Project) ValidateWorkflowAllowed(allowedWorkflows []string) error {
	return p.validateWorkflowAllowedForType(DefaultWorkflowType, allowedWorkflows)
}

func (p Project) ValidatePRWorkflowAllowed(allowedWorkflows []string) error {
	return p.validateWorkflowAllowedForType(PullRequestWorkflowType, allowedWorkflows)
}

func (p Project) ValidateDeploymentWorkflowAllowed(allowedWorkflows []string) error {
	return p.validateWorkflowAllowedForType(DeploymentWorkflowType, allowedWorkflows)
}

func (p Project) validateWorkflowForType(
	workflowType workflowType,
	repoWorkflows map[string]Workflow,
	globalWorkflows map[string]Workflow,
) error {
	name := p.getWorkflowName(workflowType)

	if name != nil {
		if !mapContains(repoWorkflows, *name) && !mapContains(globalWorkflows, *name) {
			return fmt.Errorf("%s %q is not defined anywhere", workflowType, *name)
		}
	}

	return nil
}

func (p Project) validateWorkflowAllowedForType(
	workflowType workflowType,
	allowedWorkflows []string,
) error {
	name := p.getWorkflowName(workflowType)

	if name != nil {
		if !sliceContains(allowedWorkflows, *name) {
			return fmt.Errorf("%s %q is not allowed for this repo", workflowType, *name)
		}
	}

	return nil
}

// helper function to check if string is in array
func sliceContains(slc []string, str string) bool {
	for _, s := range slc {
		if s == str {
			return true
		}
	}
	return false
}

// helper function to check if map contains a key
func mapContains(m map[string]Workflow, key string) bool {
	_, ok := m[key]
	return ok
}
