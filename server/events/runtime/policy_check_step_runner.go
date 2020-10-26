package runtime

import "github.com/runatlantis/atlantis/server/events/models"

type PolicyCheckStepRunner struct {
}

func (p *PolicyCheckStepRunner) Run(ctx models.ProjectCommandContext, extraArgs []string, path string, envs map[string]string) (string, error) {
	return "Success!", nil
}
