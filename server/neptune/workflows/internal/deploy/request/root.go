package request

type PlanMode string

const (
	DestroyPlanMode PlanMode = "destroy"
	NormalPlanMode  PlanMode = "normal"
)

type Trigger string

const (
	MergeTrigger  Trigger = "merge"
	ManualTrigger Trigger = "manual"
)

type PlanApprovalType string

type Root struct {
	Name             string
	Apply            Job
	Plan             Job
	RepoRelPath      string
	TfVersion        string
	PlanMode         PlanMode
	PlanApprovalType string
	Trigger          Trigger
	Rerun            bool
}

type Job struct {
	Steps []Step
}

// Step was taken from the Atlantis OG config, we might be able to clean this up/remove it
type Step struct {
	StepName  string
	ExtraArgs []string
	// RunCommand is either a custom run step or the command to run
	// during an env step to populate the environment variable dynamically.
	RunCommand string
	// EnvVarName is the name of the
	// environment variable that should be set by this step.
	EnvVarName string
	// EnvVarValue is the value to set EnvVarName to.
	EnvVarValue string
}
