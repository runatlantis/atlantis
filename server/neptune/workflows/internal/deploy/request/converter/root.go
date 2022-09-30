package converter

import (
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/request"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/job"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/root"
)

func Root(external request.Root) root.Root {
	return root.Root{
		Name: external.Name,
		Apply: job.Terraform{
			Steps: steps(external.Apply.Steps),
		},
		Plan: job.Plan{
			Terraform: job.Terraform{
				Steps: steps(external.Plan.Steps)},
			Mode: mode(external.PlanMode),
		},
		Path:      external.RepoRelPath,
		TfVersion: external.TfVersion,
	}

}

func mode(mode request.PlanMode) *job.PlanMode {
	switch mode {
	case request.DestroyPlanMode:
		return job.NewDestroyPlanMode()
	}

	return nil
}

func steps(requestSteps []request.Step) []job.Step {
	var terraformSteps []job.Step
	for _, step := range requestSteps {
		terraformSteps = append(terraformSteps, job.Step{
			StepName:    step.StepName,
			ExtraArgs:   step.ExtraArgs,
			RunCommand:  step.RunCommand,
			EnvVarName:  step.EnvVarName,
			EnvVarValue: step.EnvVarValue,
		})
	}
	return terraformSteps
}