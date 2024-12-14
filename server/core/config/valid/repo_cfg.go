// Package valid contains the structs representing the atlantis.yaml config
// after it's been parsed and validated.
package valid

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	version "github.com/hashicorp/go-version"
)

// RepoCfg is the atlantis.yaml config after it's been parsed and validated.
type RepoCfg struct {
	// Version is the version of the atlantis YAML file.
	Version                    int
	Projects                   []Project
	Workflows                  map[string]Workflow
	PolicySets                 PolicySets
	Automerge                  *bool
	AutoDiscover               *AutoDiscover
	ParallelApply              *bool
	ParallelPlan               *bool
	ParallelPolicyCheck        *bool
	DeleteSourceBranchOnMerge  *bool
	RepoLocks                  *RepoLocks
	CustomPolicyCheck          *bool
	EmojiReaction              string
	AllowedRegexpPrefixes      []string
	AbortOnExcecutionOrderFail bool
	SilencePRComments          []string
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
	// If we found more than one project then we need to make sure that the regex is allowed.
	if len(ps) > 1 && !isRegexAllowed(name, r.AllowedRegexpPrefixes) {
		log.Printf("Found more than one project for regex %q. This regex is not on the allow list.", name)
		return nil
	}
	return ps
}

func isRegexAllowed(name string, allowedRegexpPrefixes []string) bool {
	if len(allowedRegexpPrefixes) == 0 {
		return true
	}
	for _, allowedRegexPrefix := range allowedRegexpPrefixes {
		if strings.HasPrefix(name, allowedRegexPrefix) {
			return true
		}
	}
	return false
}

// This function returns a final true/false decision for whether AutoDiscover is enabled
// for a repo. It takes into account the defaultAutoDiscoverMode when there is no explicit
// repo config. The defaultAutoDiscoverMode param should be understood as the default
// AutoDiscover mode as may be set via CLI params or server side repo config.
func (r RepoCfg) AutoDiscoverEnabled(defaultAutoDiscoverMode AutoDiscoverMode) bool {
	autoDiscoverMode := defaultAutoDiscoverMode
	if r.AutoDiscover != nil {
		autoDiscoverMode = r.AutoDiscover.Mode
	}

	if autoDiscoverMode == AutoDiscoverAutoMode {
		// AutoDiscover is enabled by default when no projects are defined
		return len(r.Projects) == 0
	}

	return autoDiscoverMode == AutoDiscoverEnabledMode
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
	Dir                       string
	BranchRegex               *regexp.Regexp
	Workspace                 string
	Name                      *string
	WorkflowName              *string
	TerraformDistribution     *string
	TerraformVersion          *version.Version
	Autoplan                  Autoplan
	PlanRequirements          []string
	ApplyRequirements         []string
	ImportRequirements        []string
	DependsOn                 []string
	DeleteSourceBranchOnMerge *bool
	RepoLocking               *bool
	RepoLocks                 *RepoLocks
	ExecutionOrderGroup       int
	PolicyCheck               *bool
	CustomPolicyCheck         *bool
	SilencePRComments         []string
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

// PostProcessRunOutputOption is an enum of options for post-processing RunCommand output
type PostProcessRunOutputOption string

const (
	PostProcessRunOutputShow            = "show"
	PostProcessRunOutputHide            = "hide"
	PostProcessRunOutputStripRefreshing = "strip_refreshing"
)

type Stage struct {
	Steps []Step
}

// CommandShell sets up the shell for command execution
type CommandShell struct {
	Shell     string
	ShellArgs []string
}

func (s CommandShell) String() string {
	return fmt.Sprintf("%s %s", s.Shell, strings.Join(s.ShellArgs, " "))
}

type Step struct {
	StepName  string
	ExtraArgs []string
	// RunCommand is either a custom run step or the command to run
	// during an env step to populate the environment variable dynamically.
	RunCommand string
	// Output is option for post-processing a RunCommand output
	Output PostProcessRunOutputOption
	// EnvVarName is the name of the
	// environment variable that should be set by this step.
	EnvVarName string
	// EnvVarValue is the value to set EnvVarName to.
	EnvVarValue string
	// The Shell to use for RunCommand execution.
	RunShell *CommandShell
}

type Workflow struct {
	Name        string
	Apply       Stage
	Plan        Stage
	PolicyCheck Stage
	Import      Stage
	StateRm     Stage
}
