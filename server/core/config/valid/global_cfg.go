package valid

import (
	"fmt"
	"regexp"
	"strings"

	version "github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/utils"
)

const MergeableCommandReq = "mergeable"
const ApprovedCommandReq = "approved"
const UnDivergedCommandReq = "undiverged"
const PoliciesPassedCommandReq = "policies_passed"
const PlanRequirementsKey = "plan_requirements"
const ApplyRequirementsKey = "apply_requirements"
const ImportRequirementsKey = "import_requirements"
const WorkflowKey = "workflow"
const AllowedOverridesKey = "allowed_overrides"
const AllowCustomWorkflowsKey = "allow_custom_workflows"
const DefaultWorkflowName = "default"
const DeleteSourceBranchOnMergeKey = "delete_source_branch_on_merge"
const RepoLockingKey = "repo_locking"
const RepoLocksKey = "repo_locks"
const PolicyCheckKey = "policy_check"
const CustomPolicyCheckKey = "custom_policy_check"
const AutoDiscoverKey = "autodiscover"
const SilencePRCommentsKey = "silence_pr_comments"

var AllowedSilencePRComments = []string{"plan", "apply"}

// DefaultAtlantisFile is the default name of the config file for each repo.
const DefaultAtlantisFile = "atlantis.yaml"

// NonOverrideableApplyReqs will get applied across all "repos" in the server side config.
// If repo config is allowed overrides, they can override this.
// TODO: Make this more customizable, not everyone wants this rigid workflow
// maybe something along the lines of defining overridable/non-overrideable apply
// requirements in the config and removing the flag to enable policy checking.
var NonOverrideableApplyReqs = []string{PoliciesPassedCommandReq}

// GlobalCfg is the final parsed version of server-side repo config.
type GlobalCfg struct {
	Repos      []Repo
	Workflows  map[string]Workflow
	PolicySets PolicySets
	Metrics    Metrics
	TeamAuthz  TeamAuthz
}

type Metrics struct {
	Statsd     *Statsd
	Prometheus *Prometheus
}

type Statsd struct {
	Port string
	Host string
}

type Prometheus struct {
	Endpoint string
}

// Repo is the final parsed version of server-side repo config.
type Repo struct {
	// ID is the exact match id of this config.
	// If IDRegex is set then this will be empty.
	ID string
	// IDRegex is the regex match for this config.
	// If ID is set then this will be nil.
	IDRegex                   *regexp.Regexp
	BranchRegex               *regexp.Regexp
	RepoConfigFile            string
	PlanRequirements          []string
	ApplyRequirements         []string
	ImportRequirements        []string
	PreWorkflowHooks          []*WorkflowHook
	Workflow                  *Workflow
	PostWorkflowHooks         []*WorkflowHook
	AllowedWorkflows          []string
	AllowedOverrides          []string
	AllowCustomWorkflows      *bool
	DeleteSourceBranchOnMerge *bool
	RepoLocking               *bool
	RepoLocks                 *RepoLocks
	PolicyCheck               *bool
	CustomPolicyCheck         *bool
	AutoDiscover              *AutoDiscover
	SilencePRComments         []string
}

type MergedProjectCfg struct {
	PlanRequirements          []string
	ApplyRequirements         []string
	ImportRequirements        []string
	Workflow                  Workflow
	AllowedWorkflows          []string
	DependsOn                 []string
	RepoRelDir                string
	Workspace                 string
	Name                      string
	AutoplanEnabled           bool
	AutoMergeDisabled         bool
	AutoMergeMethod           string
	TerraformDistribution     *string
	TerraformVersion          *version.Version
	RepoCfgVersion            int
	PolicySets                PolicySets
	DeleteSourceBranchOnMerge bool
	ExecutionOrderGroup       int
	RepoLocks                 RepoLocks
	PolicyCheck               bool
	CustomPolicyCheck         bool
	SilencePRComments         []string
}

// WorkflowHook is a map of custom run commands to run before or after workflows.
type WorkflowHook struct {
	StepName        string
	RunCommand      string
	StepDescription string
	Shell           string
	ShellArgs       string
	Commands        string
}

// DefaultApplyStage is the Atlantis default apply stage.
var DefaultApplyStage = Stage{
	Steps: []Step{
		{
			StepName: "apply",
		},
	},
}

// DefaultPolicyCheckStage is the Atlantis default policy check stage.
var DefaultPolicyCheckStage = Stage{
	Steps: []Step{
		{
			StepName: "show",
		},
		{
			StepName: "policy_check",
		},
	},
}

// DefaultPlanStage is the Atlantis default plan stage.
var DefaultPlanStage = Stage{
	Steps: []Step{
		{
			StepName: "init",
		},
		{
			StepName: "plan",
		},
	},
}

// DefaultImportStage is the Atlantis default import stage.
var DefaultImportStage = Stage{
	Steps: []Step{
		{
			StepName: "init",
		},
		{
			StepName: "import",
		},
	},
}

// DefaultStateRmStage is the Atlantis default state_rm stage.
var DefaultStateRmStage = Stage{
	Steps: []Step{
		{
			StepName: "init",
		},
		{
			StepName: "state_rm",
		},
	},
}

type GlobalCfgArgs struct {
	RepoConfigFile string
	// No longer a user option as of https://github.com/runatlantis/atlantis/pull/3911,
	// but useful for tests to set to true to not require enumeration of allowed settings
	// on the repo side
	AllowAllRepoSettings bool
	PolicyCheckEnabled   bool
	PreWorkflowHooks     []*WorkflowHook
	PostWorkflowHooks    []*WorkflowHook
}

func NewGlobalCfgFromArgs(args GlobalCfgArgs) GlobalCfg {
	defaultWorkflow := Workflow{
		Name:        DefaultWorkflowName,
		Apply:       DefaultApplyStage,
		Plan:        DefaultPlanStage,
		PolicyCheck: DefaultPolicyCheckStage,
		Import:      DefaultImportStage,
		StateRm:     DefaultStateRmStage,
	}
	// Must construct slices here instead of using a `var` declaration because
	// we treat nil slices differently.
	commandReqs := []string{}
	allowedOverrides := []string{}
	allowedWorkflows := []string{}
	policyCheck := false
	if args.PolicyCheckEnabled {
		commandReqs = append(commandReqs, PoliciesPassedCommandReq)
		policyCheck = true
	}

	allowCustomWorkflows := false
	deleteSourceBranchOnMerge := false
	repoLocks := DefaultRepoLocks
	customPolicyCheck := false
	autoDiscover := AutoDiscover{Mode: AutoDiscoverAutoMode}
	var silencePRComments []string
	if args.AllowAllRepoSettings {
		allowedOverrides = []string{PlanRequirementsKey, ApplyRequirementsKey, ImportRequirementsKey, WorkflowKey, DeleteSourceBranchOnMergeKey, RepoLockingKey, RepoLocksKey, PolicyCheckKey, SilencePRCommentsKey}
		allowCustomWorkflows = true
	}

	return GlobalCfg{
		Repos: []Repo{
			{
				IDRegex:                   regexp.MustCompile(".*"),
				BranchRegex:               regexp.MustCompile(".*"),
				RepoConfigFile:            args.RepoConfigFile,
				PlanRequirements:          commandReqs,
				ApplyRequirements:         commandReqs,
				ImportRequirements:        commandReqs,
				PreWorkflowHooks:          args.PreWorkflowHooks,
				Workflow:                  &defaultWorkflow,
				PostWorkflowHooks:         args.PostWorkflowHooks,
				AllowedWorkflows:          allowedWorkflows,
				AllowedOverrides:          allowedOverrides,
				AllowCustomWorkflows:      &allowCustomWorkflows,
				DeleteSourceBranchOnMerge: &deleteSourceBranchOnMerge,
				RepoLocks:                 &repoLocks,
				PolicyCheck:               &policyCheck,
				CustomPolicyCheck:         &customPolicyCheck,
				AutoDiscover:              &autoDiscover,
				SilencePRComments:         silencePRComments,
			},
		},
		Workflows: map[string]Workflow{
			DefaultWorkflowName: defaultWorkflow,
		},
		TeamAuthz: TeamAuthz{
			Args: make([]string, 0),
		},
	}
}

// IDMatches returns true if the repo ID otherID matches this config.
func (r Repo) IDMatches(otherID string) bool {
	if r.ID != "" {
		return r.ID == otherID
	}
	return r.IDRegex.MatchString(otherID)
}

// BranchMatches returns true if the branch other matches a branch regex (if preset).
func (r Repo) BranchMatches(other string) bool {
	if r.BranchRegex == nil {
		return true
	}
	return r.BranchRegex.MatchString(other)
}

// IDString returns a string representation of this config.
func (r Repo) IDString() string {
	if r.ID != "" {
		return r.ID
	}
	return "/" + r.IDRegex.String() + "/"
}

// MergeProjectCfg merges proj and rCfg with the global config to return a
// final config. It assumes that all configs have been validated.
func (g GlobalCfg) MergeProjectCfg(log logging.SimpleLogging, repoID string, proj Project, rCfg RepoCfg) MergedProjectCfg {
	log.Debug("MergeProjectCfg started")
	planReqs, applyReqs, importReqs, workflow, allowedOverrides, allowCustomWorkflows, deleteSourceBranchOnMerge, repoLocks, policyCheck, customPolicyCheck, _, silencePRComments := g.getMatchingCfg(log, repoID)
	// If repos are allowed to override certain keys then override them.
	for _, key := range allowedOverrides {
		switch key {
		case PlanRequirementsKey:
			if proj.PlanRequirements != nil {
				log.Debug("overriding server-defined %s with repo settings: [%s]", PlanRequirementsKey, strings.Join(proj.PlanRequirements, ","))
				planReqs = proj.PlanRequirements
			}
		case ApplyRequirementsKey:
			if proj.ApplyRequirements != nil {
				log.Debug("overriding server-defined %s with repo settings: [%s]", ApplyRequirementsKey, strings.Join(proj.ApplyRequirements, ","))
				applyReqs = proj.ApplyRequirements

				// Preserve policies_passed req if policy check is enabled
				if policyCheck {
					applyReqs = append(applyReqs, PoliciesPassedCommandReq)
				}
			}
		case ImportRequirementsKey:
			if proj.ImportRequirements != nil {
				log.Debug("overriding server-defined %s with repo settings: [%s]", ImportRequirementsKey, strings.Join(proj.ImportRequirements, ","))
				importReqs = proj.ImportRequirements
			}
		case WorkflowKey:
			if proj.WorkflowName != nil {
				// We iterate over the global workflows first and the repo
				// workflows second so that repo workflows override. This is
				// safe because at this point we know if a repo is allowed to
				// define its own workflow. We also know that a workflow will
				// exist with this name due to earlier validation.
				name := *proj.WorkflowName
				for k, v := range g.Workflows {
					if k == name {
						workflow = v
					}
				}
				if allowCustomWorkflows {
					for k, v := range rCfg.Workflows {
						if k == name {
							workflow = v
						}
					}
				}
				log.Debug("overriding server-defined %s with repo-specified workflow: %q", WorkflowKey, workflow.Name)
			}
		case DeleteSourceBranchOnMergeKey:
			//We check whether the server configured value and repo-root level
			//config is different. If it is then we change to the more granular.
			if rCfg.DeleteSourceBranchOnMerge != nil && deleteSourceBranchOnMerge != *rCfg.DeleteSourceBranchOnMerge {
				log.Debug("overriding server-defined %s with repo settings: [%t]", DeleteSourceBranchOnMergeKey, rCfg.DeleteSourceBranchOnMerge)
				deleteSourceBranchOnMerge = *rCfg.DeleteSourceBranchOnMerge
			}
			//Then we check whether the more granular project based config is
			//different. If it is then we set it.
			if proj.DeleteSourceBranchOnMerge != nil && deleteSourceBranchOnMerge != *proj.DeleteSourceBranchOnMerge {
				log.Debug("overriding repo-root-defined %s with repo settings: [%t]", DeleteSourceBranchOnMergeKey, *proj.DeleteSourceBranchOnMerge)
				deleteSourceBranchOnMerge = *proj.DeleteSourceBranchOnMerge
			}
			log.Debug("merged deleteSourceBranchOnMerge: [%t]", deleteSourceBranchOnMerge)
		case RepoLockingKey:
			if proj.RepoLocking != nil {
				log.Debug("overriding server-defined %s with repo settings: [%t]", RepoLockingKey, *proj.RepoLocking)
				if *proj.RepoLocking && repoLocks.Mode == RepoLocksDisabledMode {
					repoLocks.Mode = DefaultRepoLocksMode
				} else if !*proj.RepoLocking {
					repoLocks.Mode = RepoLocksDisabledMode
				}
			}
		case RepoLocksKey:
			//We check whether the server configured value and repo-root level
			//config is different. If it is then we change to the more granular.
			if rCfg.RepoLocks != nil && repoLocks.Mode != rCfg.RepoLocks.Mode {
				log.Debug("overriding server-defined %s with repo settings: [%#v]", RepoLocksKey, rCfg.RepoLocks)
				repoLocks = *rCfg.RepoLocks
			}
			//Then we check whether the more granular project based config is
			//different. If it is then we set it.
			if proj.RepoLocks != nil && repoLocks.Mode != proj.RepoLocks.Mode {
				log.Debug("overriding repo-root-defined %s with repo settings: [%#v]", RepoLocksKey, *proj.RepoLocks)
				repoLocks = *proj.RepoLocks
			}
			log.Debug("merged repoLocks: [%#v]", repoLocks)
		case PolicyCheckKey:
			if proj.PolicyCheck != nil {
				log.Debug("overriding server-defined %s with repo settings: [%t]", PolicyCheckKey, *proj.PolicyCheck)
				policyCheck = *proj.PolicyCheck
			}
		case CustomPolicyCheckKey:
			if proj.CustomPolicyCheck != nil {
				log.Debug("overriding server-defined %s with repo settings: [%t]", CustomPolicyCheckKey, *proj.CustomPolicyCheck)
				customPolicyCheck = *proj.CustomPolicyCheck
			}
		case SilencePRCommentsKey:
			if proj.SilencePRComments != nil {
				log.Debug("overriding repo-root-defined %s with repo settings: [%t]", SilencePRCommentsKey, strings.Join(proj.SilencePRComments, ","))
				silencePRComments = proj.SilencePRComments
			} else if rCfg.SilencePRComments != nil {
				log.Debug("overriding server-defined %s with repo settings: [%s]", SilencePRCommentsKey, strings.Join(rCfg.SilencePRComments, ","))
				silencePRComments = rCfg.SilencePRComments
			}
		}
		log.Debug("MergeProjectCfg completed")
	}

	log.Debug("final settings: %s: [%s], %s: [%s], %s: [%s], %s: %s, %s: %t, %s: %s, %s: %t, %s: %t, %s: [%s]",
		PlanRequirementsKey, strings.Join(planReqs, ","),
		ApplyRequirementsKey, strings.Join(applyReqs, ","),
		ImportRequirementsKey, strings.Join(importReqs, ","),
		WorkflowKey, workflow.Name,
		DeleteSourceBranchOnMergeKey, deleteSourceBranchOnMerge,
		RepoLockingKey, repoLocks.Mode,
		PolicyCheckKey, policyCheck,
		CustomPolicyCheckKey, policyCheck,
		SilencePRCommentsKey, strings.Join(silencePRComments, ","),
	)

	return MergedProjectCfg{
		PlanRequirements:          planReqs,
		ApplyRequirements:         applyReqs,
		ImportRequirements:        importReqs,
		Workflow:                  workflow,
		RepoRelDir:                proj.Dir,
		Workspace:                 proj.Workspace,
		DependsOn:                 proj.DependsOn,
		Name:                      proj.GetName(),
		AutoplanEnabled:           proj.Autoplan.Enabled,
		TerraformDistribution:     proj.TerraformDistribution,
		TerraformVersion:          proj.TerraformVersion,
		RepoCfgVersion:            rCfg.Version,
		PolicySets:                g.PolicySets,
		DeleteSourceBranchOnMerge: deleteSourceBranchOnMerge,
		ExecutionOrderGroup:       proj.ExecutionOrderGroup,
		RepoLocks:                 repoLocks,
		PolicyCheck:               policyCheck,
		CustomPolicyCheck:         customPolicyCheck,
		SilencePRComments:         silencePRComments,
	}
}

// DefaultProjCfg returns the default project config for all projects under the
// repo with id repoID. It is used when there is no repo config.
func (g GlobalCfg) DefaultProjCfg(log logging.SimpleLogging, repoID string, repoRelDir string, workspace string) MergedProjectCfg {
	log.Debug("building config based on server-side config")
	planReqs, applyReqs, importReqs, workflow, _, _, deleteSourceBranchOnMerge, repoLocks, policyCheck, customPolicyCheck, _, silencePRComments := g.getMatchingCfg(log, repoID)
	return MergedProjectCfg{
		PlanRequirements:          planReqs,
		ApplyRequirements:         applyReqs,
		ImportRequirements:        importReqs,
		Workflow:                  workflow,
		RepoRelDir:                repoRelDir,
		Workspace:                 workspace,
		Name:                      "",
		AutoplanEnabled:           DefaultAutoPlanEnabled,
		TerraformDistribution:     nil,
		TerraformVersion:          nil,
		PolicySets:                g.PolicySets,
		DeleteSourceBranchOnMerge: deleteSourceBranchOnMerge,
		RepoLocks:                 repoLocks,
		PolicyCheck:               policyCheck,
		CustomPolicyCheck:         customPolicyCheck,
		SilencePRComments:         silencePRComments,
	}
}

// RepoAutoDiscoverCfg returns the AutoDiscover config from the global config
// for the repo with id repoID. If no matching repo is found or there is no
// AutoDiscover config then this function returns nil.
func (g GlobalCfg) RepoAutoDiscoverCfg(repoID string) *AutoDiscover {
	repo := g.MatchingRepo(repoID)
	if repo != nil {
		return repo.AutoDiscover
	}
	return nil
}

// ValidateRepoCfg validates that rCfg for repo with id repoID is valid based
// on our global config.
func (g GlobalCfg) ValidateRepoCfg(rCfg RepoCfg, repoID string) error {
	mapContainsF := func(m map[string]Workflow, key string) bool {
		for k := range m {
			if k == key {
				return true
			}
		}
		return false
	}

	// Check allowed overrides.
	var allowedOverrides []string
	for _, repo := range g.Repos {
		if repo.IDMatches(repoID) {
			if repo.AllowedOverrides != nil {
				allowedOverrides = repo.AllowedOverrides
			}
		}
	}
	for _, p := range rCfg.Projects {
		if p.WorkflowName != nil && !utils.SlicesContains(allowedOverrides, WorkflowKey) {
			return fmt.Errorf("repo config not allowed to set '%s' key: server-side config needs '%s: [%s]'", WorkflowKey, AllowedOverridesKey, WorkflowKey)
		}
		if p.ApplyRequirements != nil && !utils.SlicesContains(allowedOverrides, ApplyRequirementsKey) {
			return fmt.Errorf("repo config not allowed to set '%s' key: server-side config needs '%s: [%s]'", ApplyRequirementsKey, AllowedOverridesKey, ApplyRequirementsKey)
		}
		if p.PlanRequirements != nil && !utils.SlicesContains(allowedOverrides, PlanRequirementsKey) {
			return fmt.Errorf("repo config not allowed to set '%s' key: server-side config needs '%s: [%s]'", PlanRequirementsKey, AllowedOverridesKey, PlanRequirementsKey)
		}
		if p.ImportRequirements != nil && !utils.SlicesContains(allowedOverrides, ImportRequirementsKey) {
			return fmt.Errorf("repo config not allowed to set '%s' key: server-side config needs '%s: [%s]'", ImportRequirementsKey, AllowedOverridesKey, ImportRequirementsKey)
		}
		if p.DeleteSourceBranchOnMerge != nil && !utils.SlicesContains(allowedOverrides, DeleteSourceBranchOnMergeKey) {
			return fmt.Errorf("repo config not allowed to set '%s' key: server-side config needs '%s: [%s]'", DeleteSourceBranchOnMergeKey, AllowedOverridesKey, DeleteSourceBranchOnMergeKey)
		}
		if p.RepoLocking != nil && !utils.SlicesContains(allowedOverrides, RepoLockingKey) {
			return fmt.Errorf("repo config not allowed to set '%s' key: server-side config needs '%s: [%s]'", RepoLockingKey, AllowedOverridesKey, RepoLockingKey)
		}
		if p.RepoLocks != nil && !utils.SlicesContains(allowedOverrides, RepoLocksKey) {
			return fmt.Errorf("repo config not allowed to set '%s' key: server-side config needs '%s: [%s]'", RepoLocksKey, AllowedOverridesKey, RepoLocksKey)
		}
		if p.CustomPolicyCheck != nil && !utils.SlicesContains(allowedOverrides, CustomPolicyCheckKey) {
			return fmt.Errorf("repo config not allowed to set '%s' key: server-side config needs '%s: [%s]'", CustomPolicyCheckKey, AllowedOverridesKey, CustomPolicyCheckKey)
		}
		if p.SilencePRComments != nil {
			if !utils.SlicesContains(allowedOverrides, SilencePRCommentsKey) {
				return fmt.Errorf(
					"repo config not allowed to set '%s' key: server-side config needs '%s: [%s]'",
					SilencePRCommentsKey,
					AllowedOverridesKey,
					SilencePRCommentsKey,
				)
			}
			for _, silenceStage := range p.SilencePRComments {
				if !utils.SlicesContains(AllowedSilencePRComments, silenceStage) {
					return fmt.Errorf(
						"repo config '%s' key value of '%s' is not supported, supported values are [%s]",
						SilencePRCommentsKey,
						silenceStage,
						strings.Join(AllowedSilencePRComments, ", "),
					)
				}
			}
		}
	}

	// Check custom workflows.
	var allowCustomWorkflows bool
	for _, repo := range g.Repos {
		if repo.IDMatches(repoID) {
			if repo.AllowCustomWorkflows != nil {
				allowCustomWorkflows = *repo.AllowCustomWorkflows
			}
		}
	}

	if len(rCfg.Workflows) > 0 && !allowCustomWorkflows {
		return fmt.Errorf("repo config not allowed to define custom workflows: server-side config needs '%s: true'", AllowCustomWorkflowsKey)
	}

	// Check if the repo has set a workflow name that doesn't exist.
	for _, p := range rCfg.Projects {
		if p.WorkflowName != nil {
			name := *p.WorkflowName
			if !mapContainsF(rCfg.Workflows, name) && !mapContainsF(g.Workflows, name) {
				return fmt.Errorf("workflow %q is not defined anywhere", name)
			}
		}
	}

	// Check workflow is allowed
	var allowedWorkflows []string
	for _, repo := range g.Repos {
		if repo.IDMatches(repoID) {

			if repo.AllowedWorkflows != nil {
				allowedWorkflows = repo.AllowedWorkflows
			}
		}
	}

	for _, p := range rCfg.Projects {
		// default is always allowed
		if p.WorkflowName != nil && len(allowedWorkflows) != 0 {
			name := *p.WorkflowName
			if allowCustomWorkflows {
				// If we allow CustomWorkflows we need to check that workflow name is defined inside repo and not global.
				if mapContainsF(rCfg.Workflows, name) {
					break
				}
			}

			if !utils.SlicesContains(allowedWorkflows, name) {
				return fmt.Errorf("workflow '%s' is not allowed for this repo", name)
			}
		}
	}

	return nil
}

// getMatchingCfg returns the key settings for repoID.
func (g GlobalCfg) getMatchingCfg(log logging.SimpleLogging, repoID string) (planReqs []string, applyReqs []string, importReqs []string, workflow Workflow, allowedOverrides []string, allowCustomWorkflows bool, deleteSourceBranchOnMerge bool, repoLocks RepoLocks, policyCheck bool, customPolicyCheck bool, autoDiscover AutoDiscover, silencePRComments []string) {
	toLog := make(map[string]string)
	traceF := func(repoIdx int, repoID string, key string, val interface{}) string {
		from := "default server config"
		if repoIdx > 0 {
			from = fmt.Sprintf("repos[%d], id: %s", repoIdx, repoID)
		}
		var valStr string
		switch v := val.(type) {
		case string:
			valStr = fmt.Sprintf("%q", v)
		case []string:
			valStr = fmt.Sprintf("[%s]", strings.Join(v, ","))
		case bool:
			valStr = fmt.Sprintf("%t", v)
		default:
			valStr = "this is a bug"
		}

		return fmt.Sprintf("setting %s: %s from %s", key, valStr, from)
	}

	// Can't use raw.DefaultAutoDiscoverMode() because of an import cycle. Should refactor to avoid that.
	autoDiscover = AutoDiscover{Mode: AutoDiscoverAutoMode}
	repoLocking := true
	repoLocks = DefaultRepoLocks

	for _, key := range []string{PlanRequirementsKey, ApplyRequirementsKey, ImportRequirementsKey, WorkflowKey, AllowedOverridesKey, AllowCustomWorkflowsKey, DeleteSourceBranchOnMergeKey, RepoLockingKey, RepoLocksKey, PolicyCheckKey, CustomPolicyCheckKey, SilencePRCommentsKey} {
		for i, repo := range g.Repos {
			if repo.IDMatches(repoID) {
				switch key {
				case PlanRequirementsKey:
					if repo.PlanRequirements != nil {
						toLog[PlanRequirementsKey] = traceF(i, repo.IDString(), PlanRequirementsKey, repo.PlanRequirements)
						planReqs = repo.PlanRequirements
					}
				case ApplyRequirementsKey:
					if repo.ApplyRequirements != nil {
						toLog[ApplyRequirementsKey] = traceF(i, repo.IDString(), ApplyRequirementsKey, repo.ApplyRequirements)
						applyReqs = repo.ApplyRequirements
					}
				case ImportRequirementsKey:
					if repo.ImportRequirements != nil {
						toLog[ImportRequirementsKey] = traceF(i, repo.IDString(), ImportRequirementsKey, repo.ImportRequirements)
						importReqs = repo.ImportRequirements
					}
				case WorkflowKey:
					if repo.Workflow != nil {
						toLog[WorkflowKey] = traceF(i, repo.IDString(), WorkflowKey, repo.Workflow.Name)
						workflow = *repo.Workflow
					}
				case AllowedOverridesKey:
					if repo.AllowedOverrides != nil {
						toLog[AllowedOverridesKey] = traceF(i, repo.IDString(), AllowedOverridesKey, repo.AllowedOverrides)
						allowedOverrides = repo.AllowedOverrides
					}
				case AllowCustomWorkflowsKey:
					if repo.AllowCustomWorkflows != nil {
						toLog[AllowCustomWorkflowsKey] = traceF(i, repo.IDString(), AllowCustomWorkflowsKey, *repo.AllowCustomWorkflows)
						allowCustomWorkflows = *repo.AllowCustomWorkflows
					}
				case DeleteSourceBranchOnMergeKey:
					if repo.DeleteSourceBranchOnMerge != nil {
						toLog[DeleteSourceBranchOnMergeKey] = traceF(i, repo.IDString(), DeleteSourceBranchOnMergeKey, *repo.DeleteSourceBranchOnMerge)
						deleteSourceBranchOnMerge = *repo.DeleteSourceBranchOnMerge
					}
				case RepoLockingKey:
					if repo.RepoLocking != nil {
						toLog[RepoLockingKey] = traceF(i, repo.IDString(), RepoLockingKey, *repo.RepoLocking)
						repoLocking = *repo.RepoLocking
					}
				case RepoLocksKey:
					if repo.RepoLocks != nil {
						toLog[RepoLocksKey] = traceF(i, repo.IDString(), RepoLocksKey, repo.RepoLocks.Mode)
						repoLocks = *repo.RepoLocks
					}
				case PolicyCheckKey:
					if repo.PolicyCheck != nil {
						toLog[PolicyCheckKey] = traceF(i, repo.IDString(), PolicyCheckKey, *repo.PolicyCheck)
						policyCheck = *repo.PolicyCheck
					}
				case CustomPolicyCheckKey:
					if repo.CustomPolicyCheck != nil {
						toLog[CustomPolicyCheckKey] = traceF(i, repo.IDString(), CustomPolicyCheckKey, *repo.CustomPolicyCheck)
						customPolicyCheck = *repo.CustomPolicyCheck
					}
				case AutoDiscoverKey:
					if repo.AutoDiscover != nil {
						toLog[AutoDiscoverKey] = traceF(i, repo.IDString(), AutoDiscoverKey, repo.AutoDiscover.Mode)
						autoDiscover = *repo.AutoDiscover
					}
				case SilencePRCommentsKey:
					if repo.SilencePRComments != nil {
						toLog[SilencePRCommentsKey] = traceF(i, repo.IDString(), SilencePRCommentsKey, repo.SilencePRComments)
						silencePRComments = repo.SilencePRComments
					}
				}
			}
		}
	}
	for _, l := range toLog {
		log.Debug(l)
	}
	// repoLocking is deprecated and enabled by default, disable repo locks if it is explicitly disabled
	if !repoLocking {
		repoLocks.Mode = RepoLocksDisabledMode
	}
	return
}

// MatchingRepo returns an instance of Repo which matches a given repoID.
// If multiple repos match, return the last one for consistency with getMatchingCfg.
func (g GlobalCfg) MatchingRepo(repoID string) *Repo {
	for i := len(g.Repos) - 1; i >= 0; i-- {
		repo := g.Repos[i]
		if repo.IDMatches(repoID) {
			return &repo
		}
	}
	return nil
}

// RepoConfigFile returns a repository specific file path
// If not defined, return atlantis.yaml as default
func (g GlobalCfg) RepoConfigFile(repoID string) string {
	repo := g.MatchingRepo(repoID)
	if repo != nil && repo.RepoConfigFile != "" {
		return repo.RepoConfigFile
	}
	return DefaultAtlantisFile
}
