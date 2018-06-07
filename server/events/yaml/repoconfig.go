package yaml

import (
	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/logging"
)

type ApplyRequirement interface {
	// IsMet returns true if the requirement is met and false if not.
	// If it returns false, it also returns a string describing why not.
	IsMet() (bool, string)
}

type PlanStage struct {
	Steps []Step
}

type ApplyStage struct {
	Steps             []Step
	ApplyRequirements []ApplyRequirement
}

func (p PlanStage) Run() (string, error) {
	var outputs string
	for _, step := range p.Steps {
		out, err := step.Run()
		if err != nil {
			return outputs, err
		}
		if out != "" {
			// Outputs are separated by newlines.
			outputs += "\n" + out
		}
	}
	return outputs, nil
}

func (a ApplyStage) Run() (string, error) {
	var outputs string
	for _, step := range a.Steps {
		out, err := step.Run()
		if err != nil {
			return outputs, err
		}
		if out != "" {
			// Outputs are separated by newlines.
			outputs += "\n" + out
		}
	}
	return outputs, nil
}

type Step interface {
	// Run runs the step. It returns any output that needs to be commented back
	// onto the pull request and error.
	Run() (string, error)
}

// StepMeta is the data that is available to all steps.
type StepMeta struct {
	Log       *logging.SimpleLogger
	Workspace string
	// AbsolutePath is the path to this project on disk. It's not necessarily
	// the repository root since a project can be in a subdir of the root.
	AbsolutePath string
	// DirRelativeToRepoRoot is the directory for this project relative
	// to the repo root.
	DirRelativeToRepoRoot string
	TerraformVersion      *version.Version
	TerraformExecutor     TerraformExec
	// ExtraCommentArgs are the arguments that may have been appended to the comment.
	// For example 'atlantis plan -- -target=resource'. They are already quoted
	// further up the call tree to mitigate security issues.
	ExtraCommentArgs []string
	// VCS username of who caused this step to be executed. For example the
	// commenter, or who pushed a new commit.
	Username string
}

// MustConstraint returns a constraint. It panics on error.
func MustConstraint(constraint string) version.Constraints {
	c, err := version.NewConstraint(constraint)
	if err != nil {
		panic(err)
	}
	return c
}
