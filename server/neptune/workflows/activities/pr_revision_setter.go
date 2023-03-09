package activities

import (
	"context"

	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/github"
)

type prRevisionSetterActivities struct{} // nolint: unused

type SetPRRevisionRequest struct {
	Repository  github.Repo
	PullRequest github.PullRequest
}

type SetPRRevisionResponse struct {
}

func (b *prRevisionSetterActivities) SetPRRevision(ctx context.Context, request SetPRRevisionRequest) (SetPRRevisionResponse, error) { // nolint: unused
	return SetPRRevisionResponse{}, nil
}
