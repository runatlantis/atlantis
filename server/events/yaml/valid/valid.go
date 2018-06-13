package valid

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

type Project struct {
	Dir               string
	Workspace         string
	Workflow          *string
	TerraformVersion  *string
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
