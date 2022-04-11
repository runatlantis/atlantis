package handlers

import (
	"bytes"
	"context"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/http"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/vcs/types/event"
)

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

func (p *CommentEventWorkerProxy) Handle(ctx context.Context, request *http.CloneableRequest, _ event.Comment, _ *command.Comment) error {
	buffer := bytes.NewBuffer([]byte{})

	if err := request.GetRequest().Write(buffer); err != nil {
		return errors.Wrap(err, "writing request to buffer")
	}

	if err := p.snsWriter.WriteWithContext(ctx, buffer.Bytes()); err != nil {
		return errors.Wrap(err, "writing buffer to sns")
	}

	p.logger.InfoContext(ctx, "proxied request to sns")

	return nil
}
