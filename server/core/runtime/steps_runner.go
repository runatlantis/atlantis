package runtime

import (
	"context"
	"strings"

	"github.com/runatlantis/atlantis/server/events/command"
)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_steps_runner.go StepsRunner

// StepsRunner executes steps defined in project config
type StepsRunner interface {
	Run(ctx context.Context, cmdCtx command.ProjectContext, absPath string) (string, error)
}

func NewStepsRunner(
	initStepRunner Runner,
	planStepRunner Runner,
	showStepRunner Runner,
	policyCheckStepRunner Runner,
	applyStepRunner Runner,
	versionStepRunner Runner,
	runStepRunner CustomRunner,
	envStepRunner EnvRunner,
) *stepsRunner {
	stepsRunner := &stepsRunner{}

	stepsRunner.InitRunner = initStepRunner
	stepsRunner.PlanRunner = planStepRunner
	stepsRunner.ShowRunner = showStepRunner
	stepsRunner.PolicyCheckRunner = policyCheckStepRunner
	stepsRunner.ApplyRunner = applyStepRunner
	stepsRunner.VersionRunner = versionStepRunner
	stepsRunner.RunRunner = runStepRunner
	stepsRunner.EnvRunner = envStepRunner

	return stepsRunner
}

func (r *stepsRunner) Run(ctx context.Context, cmdCtx command.ProjectContext, absPath string) (string, error) {
	var outputs []string

	envs := make(map[string]string)
	for _, step := range cmdCtx.Steps {
		var out string
		var err error
		switch step.StepName {
		case "init":
			out, err = r.InitRunner.Run(ctx, cmdCtx, step.ExtraArgs, absPath, envs)
		case "plan":
			out, err = r.PlanRunner.Run(ctx, cmdCtx, step.ExtraArgs, absPath, envs)
		case "show":
			_, err = r.ShowRunner.Run(ctx, cmdCtx, step.ExtraArgs, absPath, envs)
		case "policy_check":
			out, err = r.PolicyCheckRunner.Run(ctx, cmdCtx, step.ExtraArgs, absPath, envs)
		case "apply":
			out, err = r.ApplyRunner.Run(ctx, cmdCtx, step.ExtraArgs, absPath, envs)
		case "version":
			out, err = r.VersionRunner.Run(ctx, cmdCtx, step.ExtraArgs, absPath, envs)
		case "run":
			out, err = r.RunRunner.Run(ctx, cmdCtx, step.RunCommand, absPath, envs)
		case "env":
			out, err = r.EnvRunner.Run(ctx, cmdCtx, step.RunCommand, step.EnvVarValue, absPath, envs)
			envs[step.EnvVarName] = out
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

type stepsRunner struct {
	InitRunner        Runner
	PlanRunner        Runner
	ShowRunner        Runner
	PolicyCheckRunner Runner
	ApplyRunner       Runner
	VersionRunner     Runner
	EnvRunner         EnvRunner
	RunRunner         CustomRunner
}
