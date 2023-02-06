package job

import (
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/execute"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform/version"
	"go.temporal.io/sdk/workflow"
)

type EnvStepRunner struct {
	CmdStepRunner CmdStepRunner
}

func (e *EnvStepRunner) Run(ctx *ExecutionContext, localRoot *terraform.LocalRoot, step execute.Step) (EnvVar, error) {
	if step.EnvVarValue != "" {
		return NewEnvVarFromString(step.EnvVarName, step.EnvVarValue), nil
	}

	return e.getEnv(ctx, localRoot, step)
}

func (e *EnvStepRunner) getEnv(ctx *ExecutionContext, localRoot *terraform.LocalRoot, step execute.Step) (EnvVar, error) {
	version := workflow.GetVersion(ctx, version.LazyLoadEnvVars, workflow.DefaultVersion, 1)

	if version == workflow.DefaultVersion {
		return e.getLegacyEnvVar(ctx, localRoot, step)
	}

	return NewEnvVarFromCmd(step.EnvVarName, step.RunCommand, ctx.Path, GetDefaultEnvVars(ctx, localRoot)), nil
}

func (e *EnvStepRunner) getLegacyEnvVar(ctx *ExecutionContext, localRoot *terraform.LocalRoot, step execute.Step) (EnvVar, error) {
	res, err := e.CmdStepRunner.Run(ctx, localRoot, step)
	return NewEnvVarFromString(step.EnvVarName, res), err
}

// StringEnvVar is an environment variable who's value is explicltly defined
type StringEnvVar struct {
	name  string
	value string
}

func (v StringEnvVar) ToActivityEnvVar() activities.EnvVar {
	return activities.EnvVar{
		Name:  v.name,
		Value: v.value,
	}
}

// EnvVar's can be serialized to an activities.EnvVar object
// which is used as an activity input.
type EnvVar interface {
	ToActivityEnvVar() activities.EnvVar
}

// CommandEnvVar is an environment variable that is defined by running a shell command
// This command is serialized and lazyily run within associated activities.
type CommandEnvVar struct {
	name           string
	command        string
	dir            string
	additionalEnvs map[string]string
}

func (v CommandEnvVar) ToActivityEnvVar() activities.EnvVar {
	return activities.EnvVar{
		Name: v.name,
		Command: activities.StringCommand{
			Command:        v.command,
			Dir:            v.dir,
			AdditionalEnvs: v.additionalEnvs,
		},
	}
}

func NewEnvVarFromCmd(name string, command string, dir string, additionalEnvs map[string]string) CommandEnvVar {
	return CommandEnvVar{
		name:           name,
		command:        command,
		dir:            dir,
		additionalEnvs: additionalEnvs,
	}
}

func NewEnvVarFromString(name string, value string) StringEnvVar {
	return StringEnvVar{
		name:  name,
		value: value,
	}
}
