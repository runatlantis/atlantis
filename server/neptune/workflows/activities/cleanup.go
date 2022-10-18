package activities

import (
	"context"
	"os"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
)

type cleanupActivities struct{}

type CleanupRequest struct {
	LocalRoot *terraform.LocalRoot
}

// Let's start off with an empty struct in case we ever need to add to it
type CleanupResponse struct{}

// TODO: cleanup log streaming resources

func (t *cleanupActivities) Cleanup(ctx context.Context, request CleanupRequest) (CleanupResponse, error) {
	if err := os.RemoveAll(request.LocalRoot.Path); err != nil {
		return CleanupResponse{}, errors.Wrapf(err, "deleting path: %s", request.LocalRoot.Path)
	}
	return CleanupResponse{}, nil
}
