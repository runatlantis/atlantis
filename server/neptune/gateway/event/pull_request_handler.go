package event

import (
	"bytes"
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/http"
	"github.com/runatlantis/atlantis/server/logging"
)

// PullRequestEvent is our internal representation of a vcs based pr event
type PullRequest struct {
	Pull              models.PullRequest
	User              models.User
	EventType         models.PullRequestEventType
	Timestamp         time.Time
	InstallationToken int64
}

func NewAutoplannerValidatorProxy(
	autoplanValidator Validator,
	logger logging.Logger,
	workerProxy *PullEventWorkerProxy,
	scheduler scheduler,
) *AutoplannerValidatorProxy {
	return &AutoplannerValidatorProxy{
		autoplanValidator: autoplanValidator,
		workerProxy:       workerProxy,
		logger:            logger,
		scheduler:         scheduler,
	}
}

type Validator interface {
	InstrumentedIsValid(ctx context.Context, logger logging.Logger, baseRepo models.Repo, headRepo models.Repo, pull models.PullRequest, user models.User) bool
}

type Writer interface {
	WriteWithContext(ctx context.Context, payload []byte) error
}

type AutoplannerValidatorProxy struct {
	autoplanValidator Validator
	workerProxy       *PullEventWorkerProxy
	logger            logging.Logger
	scheduler         scheduler
}

func (p *AutoplannerValidatorProxy) handle(ctx context.Context, request *http.BufferedRequest, event PullRequest) error {
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

func (p *AutoplannerValidatorProxy) Handle(ctx context.Context, request *http.BufferedRequest, event PullRequest) error {
	return p.scheduler.Schedule(ctx, func(ctx context.Context) error {
		return p.handle(ctx, request, event)
	})
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

func (p *PullEventWorkerProxy) Handle(ctx context.Context, request *http.BufferedRequest, event PullRequest) error {
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
