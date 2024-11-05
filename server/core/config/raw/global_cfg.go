package raw

import (
	"fmt"
	"regexp"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/utils"
)

// GlobalCfg is the raw schema for server-side repo config.
type GlobalCfg struct {
	Repos      []Repo              `yaml:"repos" json:"repos"`
	Workflows  map[string]Workflow `yaml:"workflows" json:"workflows"`
	PolicySets PolicySets          `yaml:"policies" json:"policies"`
	Metrics    Metrics             `yaml:"metrics" json:"metrics"`
	TeamAuthz  TeamAuthz           `yaml:"team_authz" json:"team_authz"`
}

// Repo is the raw schema for repos in the server-side repo config.
type Repo struct {
	ID                        string         `yaml:"id" json:"id"`
	Branch                    string         `yaml:"branch" json:"branch"`
	RepoConfigFile            string         `yaml:"repo_config_file" json:"repo_config_file"`
	PlanRequirements          []string       `yaml:"plan_requirements" json:"plan_requirements"`
	ApplyRequirements         []string       `yaml:"apply_requirements" json:"apply_requirements"`
	ImportRequirements        []string       `yaml:"import_requirements" json:"import_requirements"`
	PreWorkflowHooks          []WorkflowHook `yaml:"pre_workflow_hooks" json:"pre_workflow_hooks"`
	Workflow                  *string        `yaml:"workflow,omitempty" json:"workflow,omitempty"`
	PostWorkflowHooks         []WorkflowHook `yaml:"post_workflow_hooks" json:"post_workflow_hooks"`
	AllowedWorkflows          []string       `yaml:"allowed_workflows,omitempty" json:"allowed_workflows,omitempty"`
	AllowedOverrides          []string       `yaml:"allowed_overrides" json:"allowed_overrides"`
	AllowCustomWorkflows      *bool          `yaml:"allow_custom_workflows,omitempty" json:"allow_custom_workflows,omitempty"`
	DeleteSourceBranchOnMerge *bool          `yaml:"delete_source_branch_on_merge,omitempty" json:"delete_source_branch_on_merge,omitempty"`
	RepoLocking               *bool          `yaml:"repo_locking,omitempty" json:"repo_locking,omitempty"`
	RepoLocks                 *RepoLocks     `yaml:"repo_locks,omitempty" json:"repo_locks,omitempty"`
	PolicyCheck               *bool          `yaml:"policy_check,omitempty" json:"policy_check,omitempty"`
	CustomPolicyCheck         *bool          `yaml:"custom_policy_check,omitempty" json:"custom_policy_check,omitempty"`
	AutoDiscover              *AutoDiscover  `yaml:"autodiscover,omitempty" json:"autodiscover,omitempty"`
	SilencePRComments         []string       `yaml:"silence_pr_comments,omitempty" json:"silence_pr_comments,omitempty"`
}

func (g GlobalCfg) Validate() error {
	err := validation.ValidateStruct(&g,
		validation.Field(&g.Repos),
		validation.Field(&g.Workflows),
		validation.Field(&g.Metrics),
	)
	if err != nil {
		return err
	}

	// Check that all workflows referenced by repos are actually defined.
	for _, repo := range g.Repos {
		if repo.Workflow == nil {
			continue
		}
		name := *repo.Workflow
		if name == valid.DefaultWorkflowName {
			// The 'default' workflow will always be defined.
			continue
		}
		found := false
		for w := range g.Workflows {
			if w == name {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("workflow %q is not defined", name)
		}
	}

	// Check that all allowed workflows are defined
	for _, repo := range g.Repos {
		if repo.AllowedWorkflows == nil {
			continue
		}
		for _, name := range repo.AllowedWorkflows {
			if name == valid.DefaultWorkflowName {
				// The 'default' workflow will always be defined.
				continue
			}
			found := false
			for w := range g.Workflows {
				if w == name {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("workflow %q is not defined", name)
			}
		}
	}

	// Validate supported SilencePRComments values.
	for _, repo := range g.Repos {
		if repo.SilencePRComments == nil {
			continue
		}
		for _, silenceStage := range repo.SilencePRComments {
			if !utils.SlicesContains(valid.AllowedSilencePRComments, silenceStage) {
				return fmt.Errorf(
					"server-side repo config '%s' key value of '%s' is not supported, supported values are [%s]",
					valid.SilencePRCommentsKey,
					silenceStage,
					strings.Join(valid.AllowedSilencePRComments, ", "),
				)
			}
		}
	}

	return nil
}

func (g GlobalCfg) ToValid(defaultCfg valid.GlobalCfg) valid.GlobalCfg {
	workflows := make(map[string]valid.Workflow)

	// assumes: globalcfg is always initialized with one repo .*
	globalPlanReqs := defaultCfg.Repos[0].PlanRequirements
	applyReqs := defaultCfg.Repos[0].ApplyRequirements
	var globalApplyReqs []string
	for _, req := range applyReqs {
		for _, nonOverrideableReq := range valid.NonOverrideableApplyReqs {
			if req == nonOverrideableReq {
				globalApplyReqs = append(globalApplyReqs, req)
			}
		}
	}
	globalImportReqs := defaultCfg.Repos[0].ImportRequirements

	for k, v := range g.Workflows {
		validatedWorkflow := v.ToValid(k)
		workflows[k] = validatedWorkflow
		if k == valid.DefaultWorkflowName {
			// Handle the special case where they're redefining the default
			// workflow. In this case, our default repo config references
			// the "old" default workflow and so needs to be redefined.
			defaultCfg.Repos[0].Workflow = &validatedWorkflow
		}
	}
	// Merge in defaults without overriding.
	for k, v := range defaultCfg.Workflows {
		if _, ok := workflows[k]; !ok {
			workflows[k] = v
		}
	}

	var repos []valid.Repo
	for _, r := range g.Repos {
		repos = append(repos, r.ToValid(workflows, globalPlanReqs, globalApplyReqs, globalImportReqs))
	}
	repos = append(defaultCfg.Repos, repos...)

	return valid.GlobalCfg{
		Repos:      repos,
		Workflows:  workflows,
		PolicySets: g.PolicySets.ToValid(),
		Metrics:    g.Metrics.ToValid(),
		TeamAuthz:  g.TeamAuthz.ToValid(),
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
		if branch == "" {
			return nil
		}
		if !strings.HasPrefix(branch, "/") || !strings.HasSuffix(branch, "/") {
			return errors.New("regex must begin and end with a slash '/'")
		}
		withoutSlashes := branch[1 : len(branch)-1]
		_, err := regexp.Compile(withoutSlashes)
		return errors.Wrapf(err, "parsing: %s", branch)
	}

	repoConfigFileValid := func(value interface{}) error {
		repoConfigFile := value.(string)
		if repoConfigFile == "" {
			return nil
		}
		if strings.HasPrefix(repoConfigFile, "/") {
			return errors.New("must not starts with a slash '/'")
		}
		if strings.Contains(repoConfigFile, "../") || strings.Contains(repoConfigFile, "..\\") {
			return errors.New("must not contains parent directory path like '../'")
		}
		return nil
	}

	overridesValid := func(value interface{}) error {
		overrides := value.([]string)
		for _, o := range overrides {
			if o != valid.PlanRequirementsKey && o != valid.ApplyRequirementsKey && o != valid.ImportRequirementsKey && o != valid.WorkflowKey && o != valid.DeleteSourceBranchOnMergeKey && o != valid.RepoLockingKey && o != valid.RepoLocksKey && o != valid.PolicyCheckKey && o != valid.CustomPolicyCheckKey && o != valid.SilencePRCommentsKey {
				return fmt.Errorf("%q is not a valid override, only %q, %q, %q, %q, %q, %q, %q, %q, %q, and %q are supported", o, valid.PlanRequirementsKey, valid.ApplyRequirementsKey, valid.ImportRequirementsKey, valid.WorkflowKey, valid.DeleteSourceBranchOnMergeKey, valid.RepoLockingKey, valid.RepoLocksKey, valid.PolicyCheckKey, valid.CustomPolicyCheckKey, valid.SilencePRCommentsKey)
			}
		}
		return nil
	}

	workflowExists := func(value interface{}) error {
		// We validate workflows in ParserValidator.validateRepoWorkflows
		// because we need the list of workflows to validate.
		return nil
	}

	deleteSourceBranchOnMergeValid := func(value interface{}) error {
		//TOBE IMPLEMENTED
		return nil
	}

	autoDiscoverValid := func(value interface{}) error {
		autoDiscover := value.(*AutoDiscover)
		if autoDiscover != nil {
			return autoDiscover.Validate()
		}
		return nil
	}

	repoLocksValid := func(value interface{}) error {
		repoLocks := value.(*RepoLocks)
		if repoLocks != nil {
			return repoLocks.Validate()
		}
		return nil
	}

	return validation.ValidateStruct(&r,
		validation.Field(&r.ID, validation.Required, validation.By(idValid)),
		validation.Field(&r.Branch, validation.By(branchValid)),
		validation.Field(&r.RepoConfigFile, validation.By(repoConfigFileValid)),
		validation.Field(&r.AllowedOverrides, validation.By(overridesValid)),
		validation.Field(&r.PlanRequirements, validation.By(validPlanReq)),
		validation.Field(&r.ApplyRequirements, validation.By(validApplyReq)),
		validation.Field(&r.ImportRequirements, validation.By(validImportReq)),
		validation.Field(&r.Workflow, validation.By(workflowExists)),
		validation.Field(&r.DeleteSourceBranchOnMerge, validation.By(deleteSourceBranchOnMergeValid)),
		validation.Field(&r.AutoDiscover, validation.By(autoDiscoverValid)),
		validation.Field(&r.RepoLocks, validation.By(repoLocksValid)),
	)
}

func (r Repo) ToValid(workflows map[string]valid.Workflow, globalPlanReqs []string, globalApplyReqs []string, globalImportReqs []string) valid.Repo {
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

	var preWorkflowHooks []*valid.WorkflowHook
	if len(r.PreWorkflowHooks) > 0 {
		for _, hook := range r.PreWorkflowHooks {
			preWorkflowHooks = append(preWorkflowHooks, hook.ToValid())
		}
	}

	var postWorkflowHooks []*valid.WorkflowHook
	if len(r.PostWorkflowHooks) > 0 {
		for _, hook := range r.PostWorkflowHooks {
			postWorkflowHooks = append(postWorkflowHooks, hook.ToValid())
		}
	}

	var mergedPlanReqs []string
	mergedPlanReqs = append(mergedPlanReqs, r.PlanRequirements...)
	var mergedApplyReqs []string
	mergedApplyReqs = append(mergedApplyReqs, r.ApplyRequirements...)
	var mergedImportReqs []string
	mergedImportReqs = append(mergedImportReqs, r.ImportRequirements...)

	// only add global reqs if they don't exist already.
OuterGlobalPlanReqs:
	for _, globalReq := range globalPlanReqs {
		for _, currReq := range r.PlanRequirements {
			if globalReq == currReq {
				continue OuterGlobalPlanReqs
			}
		}

		// dont add policy_check step if repo have it explicitly disabled
		if globalReq == valid.PoliciesPassedCommandReq && r.PolicyCheck != nil && !*r.PolicyCheck {
			continue
		}
		mergedPlanReqs = append(mergedPlanReqs, globalReq)
	}
OuterGlobalApplyReqs:
	for _, globalReq := range globalApplyReqs {
		for _, currReq := range r.ApplyRequirements {
			if globalReq == currReq {
				continue OuterGlobalApplyReqs
			}
		}

		// dont add policy_check step if repo have it explicitly disabled
		if globalReq == valid.PoliciesPassedCommandReq && r.PolicyCheck != nil && !*r.PolicyCheck {
			continue
		}
		mergedApplyReqs = append(mergedApplyReqs, globalReq)
	}
OuterGlobalImportReqs:
	for _, globalReq := range globalImportReqs {
		for _, currReq := range r.ImportRequirements {
			if globalReq == currReq {
				continue OuterGlobalImportReqs
			}
		}

		// dont add policy_check step if repo have it explicitly disabled
		if globalReq == valid.PoliciesPassedCommandReq && r.PolicyCheck != nil && !*r.PolicyCheck {
			continue
		}
		mergedImportReqs = append(mergedImportReqs, globalReq)
	}

	var autoDiscover *valid.AutoDiscover
	if r.AutoDiscover != nil {
		autoDiscover = r.AutoDiscover.ToValid()
	}

	var repoLocks *valid.RepoLocks
	if r.RepoLocks != nil {
		repoLocks = r.RepoLocks.ToValid()
	}

	return valid.Repo{
		ID:                        id,
		IDRegex:                   idRegex,
		BranchRegex:               branchRegex,
		RepoConfigFile:            r.RepoConfigFile,
		PlanRequirements:          mergedPlanReqs,
		ApplyRequirements:         mergedApplyReqs,
		ImportRequirements:        mergedImportReqs,
		PreWorkflowHooks:          preWorkflowHooks,
		Workflow:                  workflow,
		PostWorkflowHooks:         postWorkflowHooks,
		AllowedWorkflows:          r.AllowedWorkflows,
		AllowedOverrides:          r.AllowedOverrides,
		AllowCustomWorkflows:      r.AllowCustomWorkflows,
		DeleteSourceBranchOnMerge: r.DeleteSourceBranchOnMerge,
		RepoLocking:               r.RepoLocking,
		RepoLocks:                 repoLocks,
		PolicyCheck:               r.PolicyCheck,
		CustomPolicyCheck:         r.CustomPolicyCheck,
		AutoDiscover:              autoDiscover,
		SilencePRComments:         r.SilencePRComments,
	}
}
