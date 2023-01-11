package events

import (
	"fmt"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

type ProjectPolicyCheckCommandBuilder interface {
	BuildPolicyCheckCommands(ctx *command.Context) ([]command.ProjectContext, error)
}

type PRRPolicyCheckCommandRunner struct {
	PrjCmdBuilder ProjectPolicyCheckCommandBuilder
	*PolicyCheckCommandRunner
}

func (p *PRRPolicyCheckCommandRunner) Run(ctx *command.Context) {
	projectCmds, err := p.PrjCmdBuilder.BuildPolicyCheckCommands(ctx)
	if err != nil {
		if _, statusErr := p.vcsStatusUpdater.UpdateCombined(ctx.RequestCtx, ctx.Pull.BaseRepo, ctx.Pull, models.FailedVCSStatus, command.PolicyCheck, "", ""); statusErr != nil {
			ctx.Log.WarnContext(ctx.RequestCtx, fmt.Sprintf("unable to update commit status: %s", statusErr))
		}
		p.outputUpdater.UpdateOutput(ctx, PolicyCheckCommand{}, command.Result{Error: err})
		return
	}
	p.PolicyCheckCommandRunner.Run(ctx, projectCmds)
}
