package apply

import (
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
)

func NewDisabledRunner(outputUpdater events.OutputUpdater) *DisabledRunner {
	return &DisabledRunner{
		pullUpdater: outputUpdater,
	}
}

type DisabledRunner struct {
	pullUpdater events.OutputUpdater
}

func (r *DisabledRunner) Run(ctx *command.Context, cmd *command.Comment) {
	r.pullUpdater.UpdateOutput(
		ctx,
		cmd,
		command.Result{
			Failure: "Atlantis apply is being deprecated, please merge the PR to apply your changes",
		},
	)
}
