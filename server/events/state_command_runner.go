package events

import (
	"fmt"

	"github.com/runatlantis/atlantis/server/events/command"
)

func NewStateCommandRunner(
	pullUpdater *PullUpdater,
	prjCmdBuilder ProjectStateCommandBuilder,
	prjCmdRunner ProjectStateCommandRunner,
) *StateCommandRunner {
	return &StateCommandRunner{
		pullUpdater:   pullUpdater,
		prjCmdBuilder: prjCmdBuilder,
		prjCmdRunner:  prjCmdRunner,
	}
}

type StateCommandRunner struct {
	pullUpdater   *PullUpdater
	prjCmdBuilder ProjectStateCommandBuilder
	prjCmdRunner  ProjectStateCommandRunner
}

func (v *StateCommandRunner) Run(ctx *command.Context, cmd *CommentCommand) {
	var result command.Result
	switch cmd.SubName {
	case "rm":
		result = v.runRm(ctx, cmd)
	default:
		result = command.Result{
			Failure: fmt.Sprintf("unknown state subcommand %s", cmd.SubName),
		}
	}
	v.pullUpdater.updatePull(ctx, cmd, result)
}

func (v *StateCommandRunner) runRm(ctx *command.Context, cmd *CommentCommand) command.Result {
	projectCmds, err := v.prjCmdBuilder.BuildStateRmCommands(ctx, cmd)
	if err != nil {
		ctx.Log.Warn("Error %s", err)
	}
	return runProjectCmds(projectCmds, v.prjCmdRunner.StateRm)
}
