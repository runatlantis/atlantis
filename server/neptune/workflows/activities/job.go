package activities

import (
	"context"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/neptune/logger"
)

type closer interface {
	CloseJob(ctx context.Context, jobID string) error
}

type jobActivities struct {
	StreamCloser closer
}

type CloseJobRequest struct {
	JobID string
}

func (t *jobActivities) CloseJob(ctx context.Context, request CloseJobRequest) error {
	err := t.StreamCloser.CloseJob(ctx, request.JobID)
	if err != nil {
		logger.Error(ctx, errors.Wrapf(err, "closing job").Error())
	}
	return err
}
