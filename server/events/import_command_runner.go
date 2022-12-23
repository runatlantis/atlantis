package events

import (
	"github.com/runatlantis/atlantis/server/events/command"
)

func NewImportCommandRunner(
	pullUpdater *PullUpdater,
	prjCmdBuilder ProjectImportCommandBuilder,
	prjCmdRunner ProjectImportCommandRunner,
) *ImportCommandRunner {
	return &ImportCommandRunner{
		pullUpdater:   pullUpdater,
		prjCmdBuilder: prjCmdBuilder,
		prjCmdRunner:  prjCmdRunner,
	}
}

type ImportCommandRunner struct {
	pullUpdater   *PullUpdater
	prjCmdBuilder ProjectImportCommandBuilder
	prjCmdRunner  ProjectImportCommandRunner
}

func (v *ImportCommandRunner) Run(ctx *command.Context, cmd *CommentCommand) {
	var err error
	var projectCmds []command.ProjectContext
	projectCmds, err = v.prjCmdBuilder.BuildImportCommands(ctx, cmd)
	if err != nil {
		ctx.Log.Warn("Error %s", err)
	}

	var result command.Result
	if len(projectCmds) > 1 {
		// There is no usecase to kick terraform import into multiple projects.
		// To avoid incorrect import, suppress to execute terraform import in multiple projects.
		result = command.Result{
			Failure: "import cannot run on multiple projects. please specify one project.",
		}
	} else {
		result = runProjectCmds(projectCmds, v.prjCmdRunner.Import)
	}
	v.pullUpdater.updatePull(ctx, cmd, result)
}
