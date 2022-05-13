package plan

import (
	"fmt"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/vcs"
)

func NewRunner(vcsClient vcs.Client) *Runner {
	return &Runner{
		vcsClient: vcsClient,
	}
}

type Runner struct {
	vcsClient vcs.Client
}

func (r *Runner) Run(ctx *command.Context, cmd *command.Comment) {
	if err := r.vcsClient.CreateComment(ctx.Pull.BaseRepo, ctx.Pull.Num, "I'm a platform mode plan runner", command.Plan.String()); err != nil {
		ctx.Log.Error(fmt.Sprintf("unable to comment: %s", err))
	}
}
