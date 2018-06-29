package valid

import "github.com/hashicorp/go-version"

// Spec is the atlantis yaml spec after it's been parsed and validated.
// The raw.Spec is transformed into the ValidSpec which is then used by the
// rest of Atlantis.
type Spec struct {
	// Version is the version of the atlantis YAML file. Will always be equal
	// to 2.
	Version   int
	Projects  []Project
	Workflows map[string]Workflow
}

func (s Spec) GetPlanStage(workflowName string) *Stage {
	for name, flow := range s.Workflows {
		if name == workflowName {
			return flow.Plan
		}
	}
	return nil
}

func (s Spec) GetApplyStage(workflowName string) *Stage {
	for name, flow := range s.Workflows {
		if name == workflowName {
			return flow.Apply
		}
	}
	return nil
}

func (s Spec) FindProjectsByDirWorkspace(dir string, workspace string) []Project {
	var ps []Project
	for _, p := range s.Projects {
		if p.Dir == dir && p.Workspace == workspace {
			ps = append(ps, p)
		}
	}
	return ps
}

func (s Spec) FindProjectByName(name string) *Project {
	for _, p := range s.Projects {
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
