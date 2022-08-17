package activities

import "context"

type cleanupActivities struct{}

type CleanupRequest struct {
}

func (t *terraformActivities) Cleanup(ctx context.Context, request CleanupRequest) error {
	return nil
}
