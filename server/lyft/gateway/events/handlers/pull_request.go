package handlers

import (
	"bytes"
	"context"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/http"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/vcs/types/event"
)

func NewAsynchronousAutoplannerWorkerProxy(
	autoplanValidator EventValidator,
	logger logging.Logger,
	workerProxy *PullEventWorkerProxy,
) *AsyncAutoplannerWorkerProxy {
	return &AsyncAutoplannerWorkerProxy{
		proxy: &SynchronousAutoplannerWorkerProxy{
			autoplanValidator: autoplanValidator,
			workerProxy:       workerProxy,
			logger:            logger,
		},
		logger: logger,
	}
}

type EventValidator interface {
	InstrumentedIsValid(ctx context.Context, logger logging.Logger, baseRepo models.Repo, headRepo models.Repo, pull models.PullRequest, user models.User) bool
}

type Writer interface {
	WriteWithContext(ctx context.Context, payload []byte) error
}

type AsyncAutoplannerWorkerProxy struct {
	proxy  *SynchronousAutoplannerWorkerProxy
	logger logging.Logger
}

func (p *AsyncAutoplannerWorkerProxy) Handle(ctx context.Context, request *http.BufferedRequest, event event.PullRequest) error {
	go func() {
		err := p.proxy.Handle(ctx, request, event)

		if err != nil {
			p.logger.ErrorContext(ctx, err.Error())
		}
	}()
	return nil
}

type SynchronousAutoplannerWorkerProxy struct {
	autoplanValidator EventValidator
	workerProxy       *PullEventWorkerProxy
	logger            logging.Logger
}

func (p *SynchronousAutoplannerWorkerProxy) Handle(ctx context.Context, request *http.BufferedRequest, event event.PullRequest) error {
	if ok := p.autoplanValidator.InstrumentedIsValid(
		ctx,
		p.logger,
		event.Pull.BaseRepo,
		event.Pull.HeadRepo,
		event.Pull,
		event.User); ok {

		return p.workerProxy.Handle(ctx, request, event)
	}

	p.logger.WarnContext(ctx, "request isn't valid and will not be proxied to sns")

	return nil
}

func NewPullEventWorkerProxy(
	snsWriter Writer,
	logger logging.Logger,
) *PullEventWorkerProxy {
	return &PullEventWorkerProxy{
		snsWriter: snsWriter,
		logger:    logger,
	}
}

type PullEventWorkerProxy struct {
	snsWriter Writer
	logger    logging.Logger
}

func (p *PullEventWorkerProxy) Handle(ctx context.Context, request *http.BufferedRequest, event event.PullRequest) error {
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
