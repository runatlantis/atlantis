package runtime

import "github.com/runatlantis/atlantis/server/events/models"

type PolicyRunnerStep struct {
}

func (p *PolicyRunnerStep) Run(ctx models.ProjectCommandContext, extraArgs []string, path string, envs map[string]string) (string, error) {
	return "Success!", nil
}
