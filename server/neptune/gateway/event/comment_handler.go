package event

import (
	"bytes"
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/http"
	"github.com/runatlantis/atlantis/server/logging"
)

// Comment is our internal representation of a vcs based comment event.
type Comment struct {
	Pull      models.PullRequest
	BaseRepo  models.Repo
	HeadRepo  models.Repo
	User      models.User
	PullNum   int
	Comment   string
	VCSHost   models.VCSHostType
	Timestamp time.Time
}

func NewCommentEventWorkerProxy(
	logger logging.Logger,
	snsWriter Writer,
) *CommentEventWorkerProxy {
	return &CommentEventWorkerProxy{logger: logger, snsWriter: snsWriter}
}

type CommentEventWorkerProxy struct {
	logger    logging.Logger
	snsWriter Writer
}

func (p *CommentEventWorkerProxy) Handle(ctx context.Context, request *http.BufferedRequest, _ Comment, _ *command.Comment) error {
	buffer := bytes.NewBuffer([]byte{})

	if err := request.GetRequestWithContext(ctx).Write(buffer); err != nil {
		return errors.Wrap(err, "writing request to buffer")
	}

	if err := p.snsWriter.WriteWithContext(ctx, buffer.Bytes()); err != nil {
		return errors.Wrap(err, "writing buffer to sns")
	}

	p.logger.InfoContext(ctx, "proxied request to sns")

	return nil
}
