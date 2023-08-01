package events

import (
	"github.com/runatlantis/atlantis/server/events/command"
)

func NewVersionCommandRunner(
	pullUpdater *PullUpdater,
	prjCmdBuilder ProjectVersionCommandBuilder,
	prjCmdRunner ProjectVersionCommandRunner,
	parallelPoolSize int,
	silenceVCSStatusNoProjects bool,
) *VersionCommandRunner {
	return &VersionCommandRunner{
		pullUpdater:                pullUpdater,
		prjCmdBuilder:              prjCmdBuilder,
		prjCmdRunner:               prjCmdRunner,
		parallelPoolSize:           parallelPoolSize,
		silenceVCSStatusNoProjects: silenceVCSStatusNoProjects,
	}
}

type VersionCommandRunner struct {
	pullUpdater      *PullUpdater
	prjCmdBuilder    ProjectVersionCommandBuilder
	prjCmdRunner     ProjectVersionCommandRunner
	parallelPoolSize int
	// SilenceVCSStatusNoProjects is whether any plan should set commit status if no projects
	// are found
	silenceVCSStatusNoProjects bool
}

func (v *VersionCommandRunner) Run(ctx *command.Context, cmd *CommentCommand) {
	var err error
	var projectCmds []command.ProjectContext
	projectCmds, err = v.prjCmdBuilder.BuildVersionCommands(ctx, cmd)
	if err != nil {
		ctx.Log.Warn("Error %s", err)
	}

	if len(projectCmds) == 0 {
		ctx.Log.Info("no projects to run version in")
		return
	}

	// Only run commands in parallel if enabled
	var result command.Result
	if v.isParallelEnabled(projectCmds) {
		ctx.Log.Info("Running version in parallel")
		result = runProjectCmdsParallelGroups(ctx, projectCmds, v.prjCmdRunner.Version, v.parallelPoolSize)
	} else {
		result = runProjectCmds(projectCmds, v.prjCmdRunner.Version)
	}

	v.pullUpdater.updatePull(ctx, cmd, result)
}

func (v *VersionCommandRunner) isParallelEnabled(cmds []command.ProjectContext) bool {
	return len(cmds) > 0 && cmds[0].ParallelPolicyCheckEnabled
}
