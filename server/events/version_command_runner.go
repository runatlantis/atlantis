package events

import "github.com/runatlantis/atlantis/server/events/models"

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

func (v *VersionCommandRunner) Run(ctx *CommandContext, cmd *CommentCommand) {
	var err error
	var projectCmds []models.ProjectCommandContext
	projectCmds, err = v.prjCmdBuilder.BuildVersionCommands(ctx, cmd)
	if err != nil {
		ctx.Log.Warn("Error %s", err)
	}

	if len(projectCmds) == 0 {
		ctx.Log.Info("no projects to run version in")
		return
	}

	// Only run commands in parallel if enabled
	var result CommandResult
	if v.isParallelEnabled(projectCmds) {
		ctx.Log.Info("Running version in parallel")
		result = runProjectCmdsParallel(projectCmds, v.prjCmdRunner.Version, v.parallelPoolSize)
	} else {
		result = runProjectCmds(projectCmds, v.prjCmdRunner.Version)
	}

	v.pullUpdater.updatePull(ctx, cmd, result)
}

func (v *VersionCommandRunner) isParallelEnabled(cmds []models.ProjectCommandContext) bool {
	return len(cmds) > 0 && cmds[0].ParallelPolicyCheckEnabled
}
