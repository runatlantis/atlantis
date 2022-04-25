package apply

import (
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
)

func NewDisabledRunner(commitOutputUpdater events.CommitOutputUpdater) *DisabledRunner {
	return &DisabledRunner{
		commitOutputUpdater: commitOutputUpdater,
	}
}

type DisabledRunner struct {
	commitOutputUpdater events.CommitOutputUpdater
}

func (r *DisabledRunner) Run(ctx *command.Context, cmd *command.Comment) {
	r.commitOutputUpdater.Update(
		ctx,
		cmd,
		command.Result{
			Failure: "Atlantis apply is being deprecated, please merge the PR to apply your changes",
		},
	)
}
