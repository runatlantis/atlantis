package valid

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/logging"
)

const MergeableApplyReq = "mergeable"
const ApprovedApplyReq = "approved"
const ApplyRequirementsKey = "apply_requirements"
const WorkflowKey = "workflow"
const AllowedOverridesKey = "allowed_overrides"
const AllowCustomWorkflowsKey = "allow_custom_workflows"
const DefaultWorkflowName = "default"

type GlobalCfg struct {
	Repos     []Repo
	Workflows map[string]Workflow
}

var DefaultApplyStage = Stage{
	Steps: []Step{
		{
			StepName: "apply",
		},
	},
}

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

func NewGlobalCfg(allowRepoCfg bool, mergeableReq bool, approvedReq bool) GlobalCfg {
	defaultWorkflow := Workflow{
		Name:  DefaultWorkflowName,
		Apply: DefaultApplyStage,
		Plan:  DefaultPlanStage,
	}
	// Must construct slices here instead of using a `var` declaration because
	// we treat nil slices differently.
	applyReqs := []string{}
	allowedOverrides := []string{}
	if mergeableReq {
		applyReqs = append(applyReqs, MergeableApplyReq)
	}
	if approvedReq {
		applyReqs = append(applyReqs, ApprovedApplyReq)
	}

	allowCustomWorkfows := false
	if allowRepoCfg {
		allowedOverrides = []string{ApplyRequirementsKey, WorkflowKey}
		allowCustomWorkfows = true
	}

	return GlobalCfg{
		Repos: []Repo{
			{
				IDRegex:              regexp.MustCompile(".*"),
				ApplyRequirements:    applyReqs,
				Workflow:             &defaultWorkflow,
				AllowedOverrides:     allowedOverrides,
				AllowCustomWorkflows: &allowCustomWorkfows,
			},
		},
		Workflows: map[string]Workflow{
			DefaultWorkflowName: defaultWorkflow,
		},
	}
}

type Repo struct {
	ID                   string
	IDRegex              *regexp.Regexp
	ApplyRequirements    []string
	Workflow             *Workflow
	AllowedOverrides     []string
	AllowCustomWorkflows *bool
}

type MergedProjectCfg struct {
	ApplyRequirements []string
	Workflow          Workflow
	RepoRelDir        string
	Workspace         string
	Name              string
	AutoplanEnabled   bool
	TerraformVersion  *version.Version
}

func (r Repo) IDMatches(otherID string) bool {
	if r.ID != "" {
		return r.ID == otherID
	}
	return r.IDRegex.MatchString(otherID)
}

func (r Repo) IDString() string {
	if r.ID != "" {
		return r.ID
	}
	return "/" + r.IDRegex.String() + "/"
}

func (g GlobalCfg) MergeProjectCfg(log logging.SimpleLogging, repoID string, proj Project, rCfg RepoCfg) MergedProjectCfg {
	applyReqs, workflow, allowedOverrides, allowCustomWorkflows := g.getMatchingCfg(log, repoID)

	// If repos are allowed to override certain keys then override them.
	for _, key := range allowedOverrides {
		switch key {
		case ApplyRequirementsKey:
			if proj.ApplyRequirements != nil {
				log.Debug("overriding global %s with repo settings: [%s]", ApplyRequirementsKey, strings.Join(proj.ApplyRequirements, ","))
				applyReqs = proj.ApplyRequirements
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
				log.Debug("overriding global %s with repo-specified workflow %q", WorkflowKey, workflow.Name)
			}
		}
	}

	log.Debug("final settings for repo %s: %s: [%s], %s: %s",
		repoID, ApplyRequirementsKey, strings.Join(applyReqs, ","), WorkflowKey, workflow.Name)

	return MergedProjectCfg{
		ApplyRequirements: applyReqs,
		Workflow:          workflow,
		RepoRelDir:        proj.Dir,
		Workspace:         proj.Workspace,
		Name:              proj.GetName(),
		AutoplanEnabled:   proj.Autoplan.Enabled,
		TerraformVersion:  proj.TerraformVersion,
	}
}

func (g GlobalCfg) DefaultProjCfg(log logging.SimpleLogging, repoID string, repoRelDir string, workspace string) MergedProjectCfg {
	applyReqs, workflow, _, _ := g.getMatchingCfg(log, repoID)
	return MergedProjectCfg{
		ApplyRequirements: applyReqs,
		Workflow:          workflow,
		RepoRelDir:        repoRelDir,
		Workspace:         workspace,
		Name:              "",
		AutoplanEnabled:   DefaultAutoPlanEnabled,
		TerraformVersion:  nil,
	}
}

func (g GlobalCfg) getMatchingCfg(log logging.SimpleLogging, repoID string) (applyReqs []string, workflow Workflow, allowedOverrides []string, allowCustomWorkflows bool) {
	for _, key := range []string{ApplyRequirementsKey, WorkflowKey, AllowedOverridesKey, AllowCustomWorkflowsKey} {
		for _, repo := range g.Repos {
			if repo.IDMatches(repoID) {
				switch key {
				case ApplyRequirementsKey:
					if repo.ApplyRequirements != nil {
						log.Debug("setting %s: [%s] from repo config %q", ApplyRequirementsKey, strings.Join(repo.ApplyRequirements, ","), repo.IDString())
						applyReqs = repo.ApplyRequirements
					}
				case WorkflowKey:
					if repo.Workflow != nil {
						log.Debug("setting %s %s from repo config %q", WorkflowKey, repo.Workflow.Name, repo.IDString())
						workflow = *repo.Workflow
					}
				case AllowedOverridesKey:
					if repo.AllowedOverrides != nil {
						log.Debug("setting %s: [%s] from repo config %q", AllowedOverridesKey, strings.Join(repo.AllowedOverrides, ","), repo.IDString())
						allowedOverrides = repo.AllowedOverrides
					}
				case AllowCustomWorkflowsKey:
					if repo.AllowCustomWorkflows != nil {
						log.Debug("setting %s: %t from repo config %q", AllowCustomWorkflowsKey, *repo.AllowCustomWorkflows, repo.IDString())
						allowCustomWorkflows = *repo.AllowCustomWorkflows
					}
				}
			}
		}
	}
	return
}

func (g GlobalCfg) ValidateRepoCfg(rCfg RepoCfg, repoID string) error {
	sliceContainsF := func(slc []string, str string) bool {
		for _, s := range slc {
			if s == str {
				return true
			}
		}
		return false
	}
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
		if p.WorkflowName != nil && !sliceContainsF(allowedOverrides, WorkflowKey) {
			return fmt.Errorf("repo config not allowed to set '%s' key: server-side config needs '%s: [%s]'", WorkflowKey, AllowedOverridesKey, WorkflowKey)
		}
		if p.ApplyRequirements != nil && !sliceContainsF(allowedOverrides, ApplyRequirementsKey) {
			return fmt.Errorf("repo config not allowed to set '%s' key: server-side config needs '%s: [%s]'", ApplyRequirementsKey, AllowedOverridesKey, ApplyRequirementsKey)
		}
	}

	// Check custom workflows.
	var allowCustomWorklows bool
	for _, repo := range g.Repos {
		if repo.IDMatches(repoID) {
			if repo.AllowCustomWorkflows != nil {
				allowCustomWorklows = *repo.AllowCustomWorkflows
			}
		}
	}

	if len(rCfg.Workflows) > 0 && !allowCustomWorklows {
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

	return nil
}
