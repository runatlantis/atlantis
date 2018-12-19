// Package valid contains the structs representing the atlantis.yaml config
// after it's been parsed and validated.
package valid

import "github.com/hashicorp/go-version"

// Config is the atlantis.yaml config after it's been parsed and validated.
type Config struct {
	// Version is the version of the atlantis YAML file. Will always be equal
	// to 2.
	Version   int
	Projects  []Project
	Workflows map[string]Workflow
}

func (c Config) GetPlanStage(workflowName string) *Stage {
	for name, flow := range c.Workflows {
		if name == workflowName {
			return flow.Plan
		}
	}
	return nil
}

func (c Config) GetApplyStage(workflowName string) *Stage {
	for name, flow := range c.Workflows {
		if name == workflowName {
			return flow.Apply
		}
	}
	return nil
}

func (c Config) FindProjectsByDirWorkspace(dir string, workspace string) []Project {
	var ps []Project
	for _, p := range c.Projects {
		if p.Dir == dir && p.Workspace == workspace {
			ps = append(ps, p)
		}
	}
	return ps
}

// FindProjectsByDir returns all projects that are in dir.
func (c Config) FindProjectsByDir(dir string) []Project {
	var ps []Project
	for _, p := range c.Projects {
		if p.Dir == dir {
			ps = append(ps, p)
		}
	}
	return ps
}

func (c Config) FindProjectByName(name string) *Project {
	for _, p := range c.Projects {
		if p.Name != nil && *p.Name == name {
			return &p
		}
	}
	return nil
}

type Project struct {
	Dir               string
	Workspace         string
	Name              *string
	Workflow          *string
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
	StepName   string
	ExtraArgs  []string
	RunCommand []string
}

type Workflow struct {
	Apply *Stage
	Plan  *Stage
}
