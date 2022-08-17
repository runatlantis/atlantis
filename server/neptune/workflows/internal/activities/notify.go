package activities

import "context"

type TerraformOperation int

const (
	Plan TerraformOperation = iota
	Apply
)

type notifyActivities struct{}

type NotifyRequest struct {
	Operation TerraformOperation
}

type NotifyResponse struct {
}

func (t *terraformActivities) Notify(ctx context.Context, request NotifyRequest) error {
	return nil
}
