package job

import (
	"strings"

	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/job"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/root"
	"go.temporal.io/sdk/workflow"
)

// stepRunner runs individual run steps
type stepRunner interface {
	Run(executionContext *job.ExecutionContext, localRoot *root.LocalRoot, step job.Step) (string, error)
}

type jobRunner struct {
	EnvStepRunner stepRunner
	CmdStepRunner stepRunner
}

func NewRunner(runStepRunner stepRunner, envStepRunner stepRunner) *jobRunner {
	return &jobRunner{
		CmdStepRunner: runStepRunner,
		EnvStepRunner: envStepRunner,
	}
}

func (r *jobRunner) Run(
	ctx workflow.Context,
	terraformJob job.Job,
	localRoot *root.LocalRoot,
) (string, error) {
	var outputs []string

	// Execution ctx for a job that handles setting up the env vars from the previous steps
	jobExecutionCtx := &job.ExecutionContext{
		Context:   ctx,
		Path:      localRoot.Root.Path,
		Envs:      map[string]string{},
		TfVersion: localRoot.Root.TfVersion,
	}

	for _, step := range terraformJob.Steps {
		var out string
		var err error

		switch step.StepName {
		case "run":
			out, err = r.CmdStepRunner.Run(jobExecutionCtx, localRoot, step)
		case "env":
			out, err = r.EnvStepRunner.Run(jobExecutionCtx, localRoot, step)
			jobExecutionCtx.Envs[step.EnvVarName] = out
			// We reset out to the empty string because we don't want it to
			// be printed to the PR, it's solely to set the environment variable.
			out = ""
		}

		if out != "" {
			outputs = append(outputs, out)
		}
		if err != nil {
			return strings.Join(outputs, "\n"), err
		}
	}
	return strings.Join(outputs, "\n"), nil
}
