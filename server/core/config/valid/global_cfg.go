package valid

import (
	"fmt"
	"regexp"

	"github.com/graymeta/stow"
	stow_s3 "github.com/graymeta/stow/s3"
	version "github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/logging"
)

const MergeableApplyReq = "mergeable"
const ApprovedApplyReq = "approved"
const UnDivergedApplyReq = "undiverged"
const SQUnlockedApplyReq = "unlocked"
const PoliciesPassedApplyReq = "policies_passed"
const ApplyRequirementsKey = "apply_requirements"
const WorkflowKey = "workflow"
const PullRequestWorkflowKey = "pull_request_workflow"
const DeploymentWorkflowKey = "deployment_workflow"
const AllowedOverridesKey = "allowed_overrides"
const AllowCustomWorkflowsKey = "allow_custom_workflows"

const DefaultWorkflowName = "default"

// NonOverrideableApplyReqs will get applied across all "repos" in the server side config.
// If repo config is allowed overrides, they can override this.
// TODO: Make this more customizable, not everyone wants this rigid workflow
// maybe something along the lines of defining overridable/non-overrideable apply
// requirements in the config and removing the flag to enable policy checking.
var NonOverrideableApplyReqs = []string{PoliciesPassedApplyReq}

type WorkflowModeType int

const (
	DefaultWorkflowMode WorkflowModeType = iota
	PlatformWorkflowMode
)

type BackendType string

const (
	S3Backend    BackendType = "s3"
	LocalBackend BackendType = "local"
)

// GlobalCfg is the final parsed version of server-side repo config.
type GlobalCfg struct {
	Repos                []Repo
	Workflows            map[string]Workflow
	PullRequestWorkflows map[string]Workflow
	DeploymentWorkflows  map[string]Workflow
	PolicySets           PolicySets
	Metrics              Metrics
	PersistenceConfig    PersistenceConfig
	TerraformLogFilter   TerraformLogFilters
	Temporal             Temporal
}

type PersistenceConfig struct {
	Deployments StoreConfig
	Jobs        StoreConfig
}

type StoreConfig struct {
	ContainerName string
	Prefix        string
	BackendType   BackendType
	Config        stow.Config
}

// Interface to configure the storage backends
// Additional storage backends will implement this interface
type BackendConfigurer interface {
	GetConfigMap() stow.Config
	GetConfiguredBackend() string
	GetContainerName() string
}

type Jobs struct {
	StorageBackend *StorageBackend
}

type StorageBackend struct {
	BackendConfig BackendConfigurer
}

// S3 implementation for s3 backend storage
type S3 struct {
	BucketName string
}

func (s *S3) GetConfigMap() stow.Config {
	// Only supports Iam auth type for now
	// TODO: Add accesskeys auth type
	return stow.ConfigMap{
		stow_s3.ConfigAuthType: "iam",
	}
}

func (s *S3) GetConfiguredBackend() string {
	return "s3"
}

func (s *S3) GetContainerName() string {
	return s.BucketName
}

type Metrics struct {
	Statsd *Statsd
}

type Statsd struct {
	Port         string
	Host         string
	TagSeparator string
}

type Temporal struct {
	Port               string
	Host               string
	UseSystemCACert    bool
	Namespace          string
	TerraformTaskQueue string
}

type TerraformLogFilters struct {
	Regexes []*regexp.Regexp
}

// TODO: rename project to roots
type MergedProjectCfg struct {
	ApplyRequirements   []string
	Workflow            Workflow
	PullRequestWorkflow Workflow
	DeploymentWorkflow  Workflow
	AllowedWorkflows    []string
	RepoRelDir          string
	Workspace           string
	Name                string
	AutoplanEnabled     bool
	TerraformVersion    *version.Version
	RepoCfgVersion      int
	PolicySets          PolicySets
	Tags                map[string]string
	WorkflowMode        WorkflowModeType
}

// PreWorkflowHook is a map of custom run commands to run before workflows.
type PreWorkflowHook struct {
	StepName   string
	RunCommand string
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

// DefaultLocklessPlanStage is the Atlantis default plan stage for PR workflows in
// platform mode.
var DefaultLocklessPlanStage = Stage{
	Steps: []Step{
		{
			StepName: "init",
		},
		{
			StepName:  "plan",
			ExtraArgs: []string{"-lock=false"},
		},
	},
}

func NewGlobalCfg() GlobalCfg {
	defaultWorkflow := Workflow{
		Name:        DefaultWorkflowName,
		Apply:       DefaultApplyStage,
		Plan:        DefaultPlanStage,
		PolicyCheck: DefaultPolicyCheckStage,
	}

	// defaultPullRequstWorkflow is only used in platform mode. By default it does not
	// support apply stage, and plan stage run with -lock=false flag
	pullRequestWorkflow := Workflow{
		Name:        DefaultWorkflowName,
		Plan:        DefaultLocklessPlanStage,
		PolicyCheck: DefaultPolicyCheckStage,
	}

	deploymentWorkflow := Workflow{
		Name:  DefaultWorkflowName,
		Apply: DefaultApplyStage,
		Plan:  DefaultPlanStage,
	}

	var allowCustomWorkflows bool
	repo := Repo{
		IDRegex:              regexp.MustCompile(".*"),
		BranchRegex:          regexp.MustCompile(".*"),
		Workflow:             &defaultWorkflow,
		DeploymentWorkflow:   &deploymentWorkflow,
		PullRequestWorkflow:  &pullRequestWorkflow,
		AllowedWorkflows:     []string{},
		ApplyRequirements:    []string{},
		AllowCustomWorkflows: &allowCustomWorkflows,
		AllowedOverrides:     []string{},
		CheckoutStrategy:     "branch",
	}

	globalCfg := GlobalCfg{
		Workflows: map[string]Workflow{
			DefaultWorkflowName: defaultWorkflow,
		},
		DeploymentWorkflows: map[string]Workflow{
			DefaultWorkflowName: deploymentWorkflow,
		},
		PullRequestWorkflows: map[string]Workflow{
			DefaultWorkflowName: pullRequestWorkflow,
		},
	}

	globalCfg.Repos = []Repo{repo}

	return globalCfg
}

// MergeProjectCfg merges proj and rCfg with the global config to return a
// final config. It assumes that all configs have been validated.
func (g GlobalCfg) MergeProjectCfg(repoID string, proj Project, rCfg RepoCfg) MergedProjectCfg {
	var applyReqs []string
	var workflow Workflow
	var pullRequestWorkflow Workflow
	var deploymentWorkflow Workflow
	var allowCustomWorkflows bool

	repo := g.foldMatchingRepos(repoID)

	applyReqs = repo.ApplyRequirements
	allowCustomWorkflows = *repo.AllowCustomWorkflows
	workflow = *repo.Workflow

	pullRequestWorkflow = *repo.PullRequestWorkflow
	deploymentWorkflow = *repo.DeploymentWorkflow

	// If repos are allowed to override certain keys then override them.
	for _, key := range repo.AllowedOverrides {
		switch key {
		case ApplyRequirementsKey:
			if proj.ApplyRequirements != nil {
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
				if w, ok := g.Workflows[name]; ok {
					workflow = w
				}

				if w, ok := rCfg.Workflows[name]; allowCustomWorkflows && ok {
					workflow = w
				}
			}
		case PullRequestWorkflowKey:
			if proj.PullRequestWorkflowName != nil {
				name := *proj.PullRequestWorkflowName
				if w, ok := g.PullRequestWorkflows[name]; ok {
					pullRequestWorkflow = w
				}
			}
		case DeploymentWorkflowKey:
			if proj.DeploymentWorkflowName != nil {
				name := *proj.DeploymentWorkflowName
				if w, ok := g.DeploymentWorkflows[name]; ok {
					deploymentWorkflow = w
				}
			}
		}
	}

	return MergedProjectCfg{
		ApplyRequirements:   applyReqs,
		Workflow:            workflow,
		PullRequestWorkflow: pullRequestWorkflow,
		DeploymentWorkflow:  deploymentWorkflow,
		RepoRelDir:          proj.Dir,
		Workspace:           proj.Workspace,
		Name:                proj.GetName(),
		AutoplanEnabled:     proj.Autoplan.Enabled,
		TerraformVersion:    proj.TerraformVersion,
		RepoCfgVersion:      rCfg.Version,
		PolicySets:          g.PolicySets,
		Tags:                proj.Tags,
		WorkflowMode:        rCfg.WorkflowModeType,
	}
}

// DefaultProjCfg returns the default project config for all projects under the
// repo with id repoID. It is used when there is no repo config.
func (g GlobalCfg) DefaultProjCfg(log logging.Logger, repoID string, repoRelDir string, workspace string) MergedProjectCfg {
	repo := g.foldMatchingRepos(repoID)

	mrgPrj := MergedProjectCfg{
		ApplyRequirements: repo.ApplyRequirements,
		Workflow:          *repo.Workflow,
		RepoRelDir:        repoRelDir,
		Workspace:         workspace,
		Name:              "",
		AutoplanEnabled:   DefaultAutoPlanEnabled,
		TerraformVersion:  nil,
		PolicySets:        g.PolicySets,
	}

	return mrgPrj
}

// foldMatchingRepos will return a pseudo repo instance that will iterate over
// the matching repositories and assign relevant fields if they're defined.
// This means returned object will contain the last matching repo's value as a it's fields
func (g GlobalCfg) foldMatchingRepos(repoID string) Repo {
	foldedRepo := Repo{
		AllowedWorkflows:  make([]string, 0),
		AllowedOverrides:  make([]string, 0),
		ApplyRequirements: make([]string, 0),
	}

	for _, repo := range g.Repos {
		if repo.IDMatches(repoID) {
			if repo.ApplyRequirements != nil {
				foldedRepo.ApplyRequirements = repo.ApplyRequirements
			}
			if repo.Workflow != nil {
				foldedRepo.Workflow = repo.Workflow
			}
			if repo.PullRequestWorkflow != nil {
				foldedRepo.PullRequestWorkflow = repo.PullRequestWorkflow
			}
			if repo.DeploymentWorkflow != nil {
				foldedRepo.DeploymentWorkflow = repo.DeploymentWorkflow
			}
			if repo.AllowedWorkflows != nil {
				foldedRepo.AllowedWorkflows = repo.AllowedWorkflows
			}
			if repo.AllowedOverrides != nil {
				foldedRepo.AllowedOverrides = repo.AllowedOverrides
			}
			if repo.AllowCustomWorkflows != nil {
				foldedRepo.AllowCustomWorkflows = repo.AllowCustomWorkflows
			}
		}
	}

	return foldedRepo
}

// ValidateRepoCfg validates that rCfg for repo with id repoID is valid based
// on our global config.
func (g GlobalCfg) ValidateRepoCfg(rCfg RepoCfg, repoID string) error {
	repo := g.foldMatchingRepos(repoID)

	// Check allowed overrides.
	allowedOverrides := repo.AllowedOverrides

	if err := rCfg.ValidateAllowedOverrides(allowedOverrides); err != nil {
		return err
	}

	allowCustomWorkflows := *repo.AllowCustomWorkflows
	// Check custom workflows.
	if len(rCfg.Workflows) > 0 && !allowCustomWorkflows {
		return fmt.Errorf("repo config not allowed to define custom workflows: server-side config needs '%s: true'", AllowCustomWorkflowsKey)
	}

	// Check if the repo has set a workflow name that doesn't exist and if workflow is allowed
	if err := rCfg.ValidateWorkflows(g.Workflows, repo.AllowedWorkflows, allowCustomWorkflows); err != nil {
		return err
	}

	return nil
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
