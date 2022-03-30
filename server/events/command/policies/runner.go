package policies

import (
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
	if err := r.vcsClient.CreateComment(ctx.Pull.BaseRepo, ctx.Pull.Num, "I'm a platform mode approve_policies runner", command.ApprovePolicies.String()); err != nil {
		ctx.Log.Errorf("unable to comment: %s", err)
	}
}
