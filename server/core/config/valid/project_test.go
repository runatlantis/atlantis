package valid_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	. "github.com/runatlantis/atlantis/testing"
)

func TestProject_ValidateWorkflow(t *testing.T) {
	defaultWorklfow := valid.Workflow{
		Name: "default",
	}
	customWorkflow := valid.Workflow{
		Name: "custom",
	}
	undefinedWorkflowName := "undefined"
	cases := map[string]struct {
		repoWorflows    map[string]valid.Workflow
		globalWorkflows map[string]valid.Workflow
		project         valid.Project
		expErr          string
	}{
		"project with repo workflows": {
			repoWorflows: map[string]valid.Workflow{
				"custom": customWorkflow,
			},
			project: valid.Project{
				WorkflowName: &customWorkflow.Name,
			},
		},
		"failed validation with undefined workflow": {
			repoWorflows: map[string]valid.Workflow{
				"custom": customWorkflow,
			},
			globalWorkflows: map[string]valid.Workflow{
				"default": defaultWorklfow,
			},
			project: valid.Project{
				WorkflowName: &undefinedWorkflowName,
			},
			expErr: "workflow \"undefined\" is not defined anywhere",
		},
		"workflow defined in global config": {
			repoWorflows: map[string]valid.Workflow{
				"default": defaultWorklfow,
			},
			globalWorkflows: map[string]valid.Workflow{
				"default": defaultWorklfow,
				"custom":  customWorkflow,
			},
			project: valid.Project{
				WorkflowName: &customWorkflow.Name,
			},
		},
		"missing workflow name is valid": {
			repoWorflows: map[string]valid.Workflow{
				"default": defaultWorklfow,
			},
			globalWorkflows: map[string]valid.Workflow{
				"default": defaultWorklfow,
				"custom":  customWorkflow,
			},
			project: valid.Project{
				WorkflowName: nil,
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			actErr := c.project.ValidateWorkflow(c.repoWorflows, c.globalWorkflows)
			if c.expErr == "" {
				Ok(t, actErr)
			} else {
				ErrEquals(t, c.expErr, actErr)
			}
		})
	}
}

func TestProject_ValidateDeploymentWorkflow(t *testing.T) {
	defaultWorklfow := valid.Workflow{
		Name: "default",
	}
	customWorkflow := valid.Workflow{
		Name: "custom",
	}
	undefinedWorkflowName := "undefined"
	cases := map[string]struct {
		globalWorkflows map[string]valid.Workflow
		project         valid.Project
		expErr          string
	}{
		"failed validation with undefined workflow": {
			globalWorkflows: map[string]valid.Workflow{
				"default": defaultWorklfow,
			},
			project: valid.Project{
				DeploymentWorkflowName: &undefinedWorkflowName,
			},
			expErr: "deployment_workflow \"undefined\" is not defined anywhere",
		},
		"workflow defined in global config": {
			globalWorkflows: map[string]valid.Workflow{
				"default": defaultWorklfow,
				"custom":  customWorkflow,
			},
			project: valid.Project{
				DeploymentWorkflowName: &customWorkflow.Name,
			},
		},
		"missing workflow name is valid": {
			globalWorkflows: map[string]valid.Workflow{
				"default": defaultWorklfow,
				"custom":  customWorkflow,
			},
			project: valid.Project{
				DeploymentWorkflowName: nil,
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			actErr := c.project.ValidateDeploymentWorkflow(c.globalWorkflows)
			if c.expErr == "" {
				Ok(t, actErr)
			} else {
				ErrEquals(t, c.expErr, actErr)
			}
		})
	}
}

func TestProject_ValidatePRWorkflow(t *testing.T) {
	defaultWorklfow := valid.Workflow{
		Name: "default",
	}
	customWorkflow := valid.Workflow{
		Name: "custom",
	}
	undefinedWorkflowName := "undefined"

	cases := map[string]struct {
		globalWorkflows map[string]valid.Workflow
		project         valid.Project
		expErr          string
	}{
		"failed validation with undefined workflow": {
			globalWorkflows: map[string]valid.Workflow{
				"default": defaultWorklfow,
			},
			project: valid.Project{
				PullRequestWorkflowName: &undefinedWorkflowName,
			},
			expErr: "pull_request_workflow \"undefined\" is not defined anywhere",
		},
		"workflow defined in global config": {
			globalWorkflows: map[string]valid.Workflow{
				"default": defaultWorklfow,
				"custom":  customWorkflow,
			},
			project: valid.Project{
				PullRequestWorkflowName: &customWorkflow.Name,
			},
		},
		"missing workflow name is valid": {
			globalWorkflows: map[string]valid.Workflow{
				"default": defaultWorklfow,
				"custom":  customWorkflow,
			},
			project: valid.Project{
				PullRequestWorkflowName: nil,
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			actErr := c.project.ValidatePRWorkflow(c.globalWorkflows)
			if c.expErr == "" {
				Ok(t, actErr)
			} else {
				ErrEquals(t, c.expErr, actErr)
			}
		})
	}
}

func TestProject_ValidateWorkflowAllowed(t *testing.T) {
	undefinedWorkflowName := "undefined"
	customWorkflowName := "custom"

	cases := map[string]struct {
		allowedWorkflows []string
		project          valid.Project
		expErr           string
	}{
		"failed validation with undefined workflow": {
			allowedWorkflows: []string{"custom"},
			project: valid.Project{
				WorkflowName: &undefinedWorkflowName,
			},
			expErr: "workflow \"undefined\" is not allowed for this repo",
		},
		"workflow is allowed": {
			allowedWorkflows: []string{"custom"},
			project: valid.Project{
				WorkflowName: &customWorkflowName,
			},
		},
		"missing workflow name is valid": {
			allowedWorkflows: []string{"custom"},
			project: valid.Project{
				WorkflowName: nil,
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			actErr := c.project.ValidateWorkflowAllowed(c.allowedWorkflows)
			if c.expErr == "" {
				Ok(t, actErr)
			} else {
				ErrEquals(t, c.expErr, actErr)
			}
		})
	}
}

func TestProject_ValidatePRWorkflowAllowed(t *testing.T) {
	undefinedWorkflowName := "undefined"
	customWorkflowName := "custom"

	cases := map[string]struct {
		allowedWorkflows []string
		project          valid.Project
		expErr           string
	}{
		"failed validation with undefined workflow": {
			allowedWorkflows: []string{"custom"},
			project: valid.Project{
				PullRequestWorkflowName: &undefinedWorkflowName,
			},
			expErr: "pull_request_workflow \"undefined\" is not allowed for this repo",
		},
		"workflow is allowed": {
			allowedWorkflows: []string{"custom"},
			project: valid.Project{
				PullRequestWorkflowName: &customWorkflowName,
			},
		},
		"missing workflow name is valid": {
			allowedWorkflows: []string{"custom"},
			project: valid.Project{
				PullRequestWorkflowName: nil,
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			actErr := c.project.ValidatePRWorkflowAllowed(c.allowedWorkflows)
			if c.expErr == "" {
				Ok(t, actErr)
			} else {
				ErrEquals(t, c.expErr, actErr)
			}
		})
	}
}

func TestProject_ValidateDeploymentWorkflowAllowed(t *testing.T) {
	undefinedWorkflowName := "undefined"
	customWorkflowName := "custom"

	cases := map[string]struct {
		allowedWorkflows []string
		project          valid.Project
		expErr           string
	}{
		"failed validation with undefined workflow": {
			allowedWorkflows: []string{"custom"},
			project: valid.Project{
				DeploymentWorkflowName: &undefinedWorkflowName,
			},
			expErr: "deployment_workflow \"undefined\" is not allowed for this repo",
		},
		"workflow is allowed": {
			allowedWorkflows: []string{"custom"},
			project: valid.Project{
				DeploymentWorkflowName: &customWorkflowName,
			},
		},
		"missing workflow name is valid": {
			allowedWorkflows: []string{"custom"},
			project: valid.Project{
				DeploymentWorkflowName: nil,
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			actErr := c.project.ValidateDeploymentWorkflowAllowed(c.allowedWorkflows)
			if c.expErr == "" {
				Ok(t, actErr)
			} else {
				ErrEquals(t, c.expErr, actErr)
			}
		})
	}
}

func TestProject_ValidateAllowedOverrides(t *testing.T) {
	workflowName := "custom"
	deleteSourceBranch := true
	cases := map[string]struct {
		allowedOverrides []string
		overrideKey      string
		project          valid.Project
		expErr           string
	}{
		"workflow is not allowed override": {
			allowedOverrides: []string{},
			project: valid.Project{
				WorkflowName: &workflowName,
			},
			expErr: "repo config not allowed to set 'workflow' key: server-side config needs 'allowed_overrides: [workflow]'",
		},
		"pull_request_workflow is not allowed override": {
			allowedOverrides: []string{},
			project: valid.Project{
				PullRequestWorkflowName: &workflowName,
			},
			expErr: "repo config not allowed to set 'pull_request_workflow' key: server-side config needs 'allowed_overrides: [pull_request_workflow]'",
		},
		"deployment_workflow is not allowed override": {
			allowedOverrides: []string{},
			project: valid.Project{
				DeploymentWorkflowName: &workflowName,
			},
			expErr: "repo config not allowed to set 'deployment_workflow' key: server-side config needs 'allowed_overrides: [deployment_workflow]'",
		},
		"apply_requirements is not allowed override": {
			allowedOverrides: []string{},
			project: valid.Project{
				ApplyRequirements: []string{"mergeable"},
			},
			expErr: "repo config not allowed to set 'apply_requirements' key: server-side config needs 'allowed_overrides: [apply_requirements]'",
		},
		"delete_source_branch_on_merge is not allowed override": {
			allowedOverrides: []string{},
			project: valid.Project{
				DeleteSourceBranchOnMerge: &deleteSourceBranch,
			},
			expErr: "repo config not allowed to set 'delete_source_branch_on_merge' key: server-side config needs 'allowed_overrides: [delete_source_branch_on_merge]'",
		},
		"no errors when allowed override": {
			allowedOverrides: []string{"apply_requirements", "deployment_workflow", "pull_request_workflow", "workflow", "delete_source_branch_on_merge"},
			project: valid.Project{
				DeleteSourceBranchOnMerge: &deleteSourceBranch,
				ApplyRequirements:         []string{"mergeable"},
				DeploymentWorkflowName:    &workflowName,
				WorkflowName:              &workflowName,
				PullRequestWorkflowName:   &workflowName,
			},
		},
		"no errors if override attributes nil": {
			allowedOverrides: []string{"apply_requirements", "deployment_workflow", "pull_request_workflow", "workflow", "delete_source_branch_on_merge"},
			project: valid.Project{
				DeleteSourceBranchOnMerge: nil,
				ApplyRequirements:         nil,
				DeploymentWorkflowName:    nil,
				WorkflowName:              nil,
				PullRequestWorkflowName:   nil,
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			actErr := c.project.ValidateAllowedOverrides(c.allowedOverrides)
			if c.expErr == "" {
				Ok(t, actErr)
			} else {
				ErrEquals(t, c.expErr, actErr)
			}
		})
	}
}
