package apply

import (
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
)

func NewDisabledRunner(pullUpdater *events.PullUpdater) *DisabledRunner {
	return &DisabledRunner{
		pullUpdater: pullUpdater,
	}
}

type DisabledRunner struct {
	pullUpdater *events.PullUpdater
}

func (r *DisabledRunner) Run(ctx *command.Context, cmd *command.Comment) {
	r.pullUpdater.UpdatePull(
		ctx,
		cmd,
		command.Result{
			Failure: "Atlantis apply is being deprecated, please merge the PR to apply your changes",
		},
	)
}
