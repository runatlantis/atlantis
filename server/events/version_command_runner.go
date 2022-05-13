package events

import (
	"fmt"
	"github.com/runatlantis/atlantis/server/events/command"
)

func NewVersionCommandRunner(
	pullUpdater *PullUpdater,
	prjCmdBuilder ProjectVersionCommandBuilder,
	prjCmdRunner ProjectVersionCommandRunner,
	parallelPoolSize int,
) *VersionCommandRunner {
	return &VersionCommandRunner{
		pullUpdater:      pullUpdater,
		prjCmdBuilder:    prjCmdBuilder,
		prjCmdRunner:     prjCmdRunner,
		parallelPoolSize: parallelPoolSize,
	}
}

type VersionCommandRunner struct {
	pullUpdater      *PullUpdater
	prjCmdBuilder    ProjectVersionCommandBuilder
	prjCmdRunner     ProjectVersionCommandRunner
	parallelPoolSize int
}

func (v *VersionCommandRunner) Run(ctx *command.Context, cmd *command.Comment) {
	var err error
	var projectCmds []command.ProjectContext
	projectCmds, err = v.prjCmdBuilder.BuildVersionCommands(ctx, cmd)
	if err != nil {
		ctx.Log.WarnContext(ctx.RequestCtx, fmt.Sprintf("Error %s", err))
	}

	if len(projectCmds) == 0 {
		ctx.Log.InfoContext(ctx.RequestCtx, "no projects to run version in")
		return
	}

	// Only run commands in parallel if enabled
	var result command.Result
	if v.isParallelEnabled(projectCmds) {
		ctx.Log.InfoContext(ctx.RequestCtx, "Running version in parallel")
		result = runProjectCmdsParallel(projectCmds, v.prjCmdRunner.Version, v.parallelPoolSize)
	} else {
		result = runProjectCmds(projectCmds, v.prjCmdRunner.Version)
	}

	v.pullUpdater.UpdatePull(ctx, cmd, result)
}

func (v *VersionCommandRunner) isParallelEnabled(cmds []command.ProjectContext) bool {
	return len(cmds) > 0 && cmds[0].ParallelPolicyCheckEnabled
}
