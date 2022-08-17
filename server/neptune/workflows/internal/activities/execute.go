package activities

import "context"

type executeCommandActivities struct{}

type ExecuteCommandRequest struct {
}

func (t *terraformActivities) ExecuteCommand(ctx context.Context, request ExecuteCommandRequest) error {
	return nil
}
