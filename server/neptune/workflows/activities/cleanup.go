package activities

import (
	"context"
	"os"

	"github.com/pkg/errors"
)

type cleanupActivities struct{}

type CleanupRequest struct {
	DeployDirectory string
}

// Let's start off with an empty struct in case we ever need to add to it
type CleanupResponse struct{}

// TODO: cleanup log streaming resources

func (t *cleanupActivities) Cleanup(ctx context.Context, request CleanupRequest) (CleanupResponse, error) {
	if err := os.RemoveAll(request.DeployDirectory); err != nil {
		return CleanupResponse{}, errors.Wrapf(err, "deleting path: %s", request.DeployDirectory)
	}
	return CleanupResponse{}, nil
}
