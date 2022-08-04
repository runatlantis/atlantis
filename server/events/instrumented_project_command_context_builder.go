package events

import (
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/uber-go/tally/v4"
)

// InstrumentedProjectCommandContextBuilder ensures that project command context contains a scoped stats
// object relevant to the command it applies to.
type InstrumentedProjectCommandContextBuilder struct {
	ProjectCommandContextBuilder
	// Conciously making this global since it gets flushed periodically anyways
	ProjectCounter tally.Counter
}

// BuildProjectContext builds the context and injects the appropriate command level scope after the fact.
func (cb *InstrumentedProjectCommandContextBuilder) BuildProjectContext(
	ctx *command.Context,
	cmdName command.Name,
	prjCfg valid.MergedProjectCfg,
	commentFlags []string,
	repoDir string,
	contextFlags *command.ContextFlags,
) (projectCmds []command.ProjectContext) {
	cb.ProjectCounter.Inc(1)

	cmds := cb.ProjectCommandContextBuilder.BuildProjectContext(
		ctx, cmdName, prjCfg, commentFlags, repoDir, contextFlags,
	)

	projectCmds = []command.ProjectContext{}

	for _, cmd := range cmds {
		// specifically use the command name in the context instead of the arg
		// since we can return multiple commands worth of contexts for a given command name arg
		// to effectively pipeline them.
		cmd.SetScope(cmd.CommandName.String())
		projectCmds = append(projectCmds, cmd)
	}

	return
}
