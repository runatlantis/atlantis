package runtime

import (
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
)

type ApprovalOperator struct {
	VCSClient vcs.ClientProxy
}

func (a *ApprovalOperator) IsApproved(baseRepo models.Repo, pull models.PullRequest) (bool, error) {
	approved, err := a.VCSClient.PullIsApproved(baseRepo, pull)
	if err != nil {
		return false, err
	}
	return approved, nil
}
