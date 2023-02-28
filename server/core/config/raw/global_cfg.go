package raw

import (
	"fmt"
	"regexp"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/config/valid"
)

var validCheckoutStrategies = []interface{}{"merge", "branch"}

// GlobalCfg is the raw schema for server-side repo config.
type GlobalCfg struct {
	Repos                []Repo               `yaml:"repos" json:"repos"`
	Workflows            Workflows            `yaml:"workflows" json:"workflows"`
	PullRequestWorkflows PullRequestWorkflows `yaml:"pull_request_workflows" json:"pull_request_workflows"`
	DeploymentWorkflows  DeploymentWorkflows  `yaml:"deployment_workflows" json:"deployment_workflows"`
	PolicySets           PolicySets           `yaml:"policies" json:"policies"`
	Metrics              Metrics              `yaml:"metrics" json:"metrics"`
	TerraformLogFilters  TerraformLogFilters  `yaml:"terraform_log_filters" json:"terraform_log_filters"`
	Temporal             Temporal             `yaml:"temporal" json:"temporal"`
	Persistence          Persistence          `yaml:"persistence" json:"persistence"`
}

// Repo is the raw schema for repos in the server-side repo config.
type Repo struct {
	ID                          string            `yaml:"id" json:"id"`
	Branch                      string            `yaml:"branch" json:"branch"`
	ApplyRequirements           []string          `yaml:"apply_requirements" json:"apply_requirements"`
	PreWorkflowHooks            []PreWorkflowHook `yaml:"pre_workflow_hooks" json:"pre_workflow_hooks"`
	Workflow                    *string           `yaml:"workflow,omitempty" json:"workflow,omitempty"`
	PullRequestWorkflow         *string           `yaml:"pull_request_workflow,omitempty" json:"pull_request_workflow,omitempty"`
	DeploymentWorkflow          *string           `yaml:"deployment_workflow,omitempty" json:"deployment_workflow,omitempty"`
	AllowedWorkflows            []string          `yaml:"allowed_workflows,omitempty" json:"allowed_workflows,omitempty"`
	AllowedPullRequestWorkflows []string          `yaml:"allowed_pull_request_workflows,omitempty" json:"allowed_pull_request_workflows,omitempty"`
	AllowedDeploymentWorkflows  []string          `yaml:"allowed_deployment_workflows,omitempty" json:"allowed_deployment_workflows,omitempty"`
	AllowedOverrides            []string          `yaml:"allowed_overrides" json:"allowed_overrides"`
	AllowCustomWorkflows        *bool             `yaml:"allow_custom_workflows,omitempty" json:"allow_custom_workflows,omitempty"`
	TemplateOverrides           map[string]string `yaml:"template_overrides,omitempty" json:"template_overrides,omitempty"`
	CheckoutStrategy            string            `yaml:"checkout_strategy,omitempty" json:"checkout_strategy,omitempty"`
	RebaseEnabled               *bool             `yaml:"rebase_enabled,omitempty" json:"rebase_enabled,omitempty"`
}

func (g GlobalCfg) GetWorkflowNames() []string {
	names := make([]string, 0)
	for name := range g.Workflows {
		names = append(names, name)
	}
	return names
}

func (g GlobalCfg) GetPullRequestWorkflowNames() []string {
	names := make([]string, 0)
	for name := range g.PullRequestWorkflows {
		names = append(names, name)
	}
	return names
}

func (g GlobalCfg) GetDeploymentWorkflowNames() []string {
	names := make([]string, 0)
	for name := range g.DeploymentWorkflows {
		names = append(names, name)
	}
	return names
}

func (g GlobalCfg) Validate() error {
	err := validation.ValidateStruct(&g,
		validation.Field(&g.Repos),
		validation.Field(&g.Workflows),
		validation.Field(&g.PullRequestWorkflows),
		validation.Field(&g.DeploymentWorkflows),
		validation.Field(&g.Metrics),
		validation.Field(&g.TerraformLogFilters),
		validation.Field(&g.Persistence),
	)
	if err != nil {
		return err
	}

	// Check that all workflows referenced by repos are actually defined.
	for _, repo := range g.Repos {
		if err := validateWorkflow(repo.Workflow, g.GetWorkflowNames()); err != nil {
			return err
		}

		if err := validateWorkflow(repo.PullRequestWorkflow, g.GetPullRequestWorkflowNames()); err != nil {
			return err
		}

		if err := validateWorkflow(repo.DeploymentWorkflow, g.GetDeploymentWorkflowNames()); err != nil {
			return err
		}
	}

	// Check that all allowed workflows are defined
	for _, repo := range g.Repos {
		if err := validateAllowedWorkflows(repo.AllowedWorkflows, g.GetWorkflowNames()); err != nil {
			return err
		}
		if err := validateAllowedWorkflows(repo.AllowedPullRequestWorkflows, g.GetPullRequestWorkflowNames()); err != nil {
			return err
		}
		if err := validateAllowedWorkflows(repo.AllowedDeploymentWorkflows, g.GetDeploymentWorkflowNames()); err != nil {
			return err
		}
	}

	return nil
}

func (g GlobalCfg) ToValid(defaultCfg valid.GlobalCfg) valid.GlobalCfg {
	var globalApplyReqs []string

	policySets := g.PolicySets.ToValid()
	if policySets.HasPolicies() {
		globalApplyReqs = append(globalApplyReqs, valid.PoliciesPassedApplyReq)
	}

	defaultRepo := &defaultCfg.Repos[0]
	validWorkflows := g.Workflows.ToValid(defaultCfg)
	validPullRequestWorkflows := g.PullRequestWorkflows.ToValid(defaultCfg)
	validDeploymentWorkflows := g.DeploymentWorkflows.ToValid(defaultCfg)

	// Handle the special case where they're redefining the default
	// workflow. In this case, our default repo config references
	// the "old" default workflow and so needs to be redefined.
	if w, ok := validWorkflows[valid.DefaultWorkflowName]; ok {
		defaultRepo.Workflow = &w
	}
	if w, ok := validPullRequestWorkflows[valid.DefaultWorkflowName]; ok {
		defaultRepo.PullRequestWorkflow = &w
	}
	if w, ok := validDeploymentWorkflows[valid.DefaultWorkflowName]; ok {
		defaultRepo.DeploymentWorkflow = &w
	}

	var repos []valid.Repo
	for _, r := range g.Repos {
		validRepo := r.ToValid(
			validWorkflows,
			validPullRequestWorkflows,
			validDeploymentWorkflows,
			globalApplyReqs,
		)

		repos = append(repos, validRepo)
	}
	repos = append(defaultCfg.Repos, repos...)

	return valid.GlobalCfg{
		Repos:                repos,
		Workflows:            validWorkflows,
		PullRequestWorkflows: validPullRequestWorkflows,
		DeploymentWorkflows:  validDeploymentWorkflows,
		PolicySets:           policySets,
		Metrics:              g.Metrics.ToValid(),
		PersistenceConfig:    g.Persistence.ToValid(defaultCfg),
		TerraformLogFilter:   g.TerraformLogFilters.ToValid(),
		Temporal:             g.Temporal.ToValid(),
	}
}

// HasRegexID returns true if r is configured with a regex id instead of an
// exact match id.
func (r Repo) HasRegexID() bool {
	return strings.HasPrefix(r.ID, "/") && strings.HasSuffix(r.ID, "/")
}

// HasRegexBranch returns true if a branch regex was set.
func (r Repo) HasRegexBranch() bool {
	return strings.HasPrefix(r.Branch, "/") && strings.HasSuffix(r.Branch, "/")
}

func (r Repo) Validate() error {
	idValid := func(value interface{}) error {
		id := value.(string)
		if !r.HasRegexID() {
			return nil
		}
		_, err := regexp.Compile(id[1 : len(id)-1])
		return errors.Wrapf(err, "parsing: %s", id)
	}

	branchValid := func(value interface{}) error {
		branch := value.(string)
		if !r.HasRegexBranch() {
			return nil
		}
		_, err := regexp.Compile(branch[1 : len(branch)-1])
		return errors.Wrapf(err, "parsing: %s", branch)
	}

	overridesValid := func(value interface{}) error {
		overrides := value.([]string)
		for _, o := range overrides {
			if o != valid.ApplyRequirementsKey &&
				o != valid.WorkflowKey &&
				o != valid.PullRequestWorkflowKey &&
				o != valid.DeploymentWorkflowKey {
				return fmt.Errorf("%q is not a valid override, only %q and %q are supported", o, valid.ApplyRequirementsKey, valid.WorkflowKey)
			}
		}
		return nil
	}

	workflowExists := func(value interface{}) error {
		// We validate workflows in ParserValidator.validateRepoWorkflows
		// because we need the list of workflows to validate.
		return nil
	}

	return validation.ValidateStruct(&r,
		validation.Field(&r.ID, validation.Required, validation.By(idValid)),
		validation.Field(&r.Branch, validation.By(branchValid)),
		validation.Field(&r.CheckoutStrategy, validation.In(validCheckoutStrategies...)),
		validation.Field(&r.AllowedOverrides, validation.By(overridesValid)),
		validation.Field(&r.ApplyRequirements, validation.By(validApplyReq)),
		validation.Field(&r.Workflow, validation.By(workflowExists)),
		validation.Field(&r.PullRequestWorkflow, validation.By(workflowExists)),
		validation.Field(&r.DeploymentWorkflow, validation.By(workflowExists)),
	)
}

func (r Repo) ToValid(
	workflows map[string]valid.Workflow,
	pullRequestWorkflows map[string]valid.Workflow,
	deploymentWorkflows map[string]valid.Workflow,
	globalApplyReqs []string,
) valid.Repo {
	var id string
	var idRegex *regexp.Regexp
	if r.HasRegexID() {
		withoutSlashes := r.ID[1 : len(r.ID)-1]
		// Safe to use MustCompile because we test it in Validate().
		idRegex = regexp.MustCompile(withoutSlashes)
	} else {
		id = r.ID
	}

	var branchRegex *regexp.Regexp
	if r.HasRegexBranch() {
		withoutSlashes := r.Branch[1 : len(r.Branch)-1]
		// Safe to use MustCompile because we test it in Validate().
		branchRegex = regexp.MustCompile(withoutSlashes)
	}

	var workflow *valid.Workflow
	if r.Workflow != nil {
		// This key is guaranteed to exist because we test for it in
		// ParserValidator.validateRepoWorkflows.
		ptr := workflows[*r.Workflow]
		workflow = &ptr
	}

	var pullRequestWorkflow *valid.Workflow
	if r.PullRequestWorkflow != nil {
		// This key is guaranteed to exist because we test for it in
		// ParserValidator.validateRepoWorkflows.
		ptr := pullRequestWorkflows[*r.PullRequestWorkflow]
		pullRequestWorkflow = &ptr
	}

	var deploymentWorkflow *valid.Workflow
	if r.DeploymentWorkflow != nil {
		// This key is guaranteed to exist because we test for it in
		// ParserValidator.validateRepoWorkflows.
		ptr := deploymentWorkflows[*r.DeploymentWorkflow]
		deploymentWorkflow = &ptr
	}

	var preWorkflowHooks []*valid.PreWorkflowHook
	if len(r.PreWorkflowHooks) > 0 {
		for _, hook := range r.PreWorkflowHooks {
			preWorkflowHooks = append(preWorkflowHooks, hook.ToValid())
		}
	}

	var mergedApplyReqs []string

	mergedApplyReqs = append(mergedApplyReqs, r.ApplyRequirements...)

	// only add global reqs if they don't exist already.
OUTER:
	for _, globalReq := range globalApplyReqs {
		for _, currReq := range r.ApplyRequirements {
			if globalReq == currReq {
				continue OUTER
			}
		}
		mergedApplyReqs = append(mergedApplyReqs, globalReq)
	}

	var checkoutStrategy string
	if r.CheckoutStrategy == "" {
		checkoutStrategy = "branch"
	} else {
		checkoutStrategy = r.CheckoutStrategy
	}

	return valid.Repo{
		ID:                          id,
		IDRegex:                     idRegex,
		BranchRegex:                 branchRegex,
		ApplyRequirements:           mergedApplyReqs,
		PreWorkflowHooks:            preWorkflowHooks,
		Workflow:                    workflow,
		PullRequestWorkflow:         pullRequestWorkflow,
		DeploymentWorkflow:          deploymentWorkflow,
		AllowedWorkflows:            r.AllowedWorkflows,
		AllowedPullRequestWorkflows: r.AllowedPullRequestWorkflows,
		AllowedDeploymentWorkflows:  r.AllowedDeploymentWorkflows,
		AllowedOverrides:            r.AllowedOverrides,
		AllowCustomWorkflows:        r.AllowCustomWorkflows,
		TemplateOverrides:           r.TemplateOverrides,
		CheckoutStrategy:            checkoutStrategy,
		RebaseEnabled:               r.RebaseEnabled,
	}
}

func validateWorkflow(workflow *string, workflowNames []string) error {
	if workflow == nil {
		return nil
	}

	name := *workflow
	if name == valid.DefaultWorkflowName {
		// The 'default' workflow will always be defined.
		return nil
	}

	for _, w := range workflowNames {
		if w == name {
			return nil
		}
	}

	return fmt.Errorf("workflow %q is not defined", name)
}

func validateAllowedWorkflows(allowedWorkflows []string, workflowNames []string) error {
	if allowedWorkflows == nil {
		return nil
	}

	for _, name := range allowedWorkflows {
		if name == valid.DefaultWorkflowName {
			// The 'default' workflow will always be defined.
			continue
		}
		for _, w := range workflowNames {
			if w == name {
				return nil
			}
		}
		return fmt.Errorf("workflow %q is not defined", name)
	}

	return nil
}
