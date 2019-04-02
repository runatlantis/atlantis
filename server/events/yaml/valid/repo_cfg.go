// Package valid contains the structs representing the atlantis.yaml config
// after it's been parsed and validated.
package valid

import version "github.com/hashicorp/go-version"

// RepoCfg is the atlantis.yaml config after it's been parsed and validated.
type RepoCfg struct {
	// Version is the version of the atlantis YAML file.
	Version   int
	Projects  []Project
	Workflows map[string]Workflow
	Automerge bool
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

type Project struct {
	Dir               string
	Workspace         string
	Name              *string
	WorkflowName      *string
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
	RunCommand string
}

type Workflow struct {
	Name  string
	Apply Stage
	Plan  Stage
}
