package job

import (
	"strings"

	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/execute"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
)

type EnvStepRunner struct {
	CmdStepRunner CmdStepRunner
}

func (e *EnvStepRunner) Run(executionContext *ExecutionContext, localRoot *terraform.LocalRoot, step execute.Step) (string, error) {
	if step.EnvVarValue != "" {
		return step.EnvVarValue, nil
	}

	res, err := e.CmdStepRunner.Run(executionContext, localRoot, step)
	// Trim newline from res to support running `echo env_value` which has
	// a newline. We don't recommend users run echo -n env_value to remove the
	// newline because -n doesn't work in the sh shell which is what we use
	// to run commands.
	return strings.TrimSuffix(res, "\n"), err
}
