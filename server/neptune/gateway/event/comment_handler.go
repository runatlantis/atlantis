package event

import (
	"bytes"
	"context"
	"github.com/runatlantis/atlantis/server/vcs/provider/github"
	"time"

	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/lyft/feature"
	contextInternal "github.com/runatlantis/atlantis/server/neptune/context"
	"github.com/runatlantis/atlantis/server/neptune/workflows"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/http"
	"github.com/runatlantis/atlantis/server/logging"
)

const warningMessage = "⚠️ WARNING ⚠️\n\n You are force applying changes from your PR instead of merging into your default branch 🚀. This can have unpredictable consequences 🙏🏽 and should only be used in an emergency 🆘.\n\n To confirm behavior, review and approve the plan within the generated atlantis/deploy GH check below.\n\n 𝐓𝐡𝐢𝐬 𝐚𝐜𝐭𝐢𝐨𝐧 𝐰𝐢𝐥𝐥 𝐛𝐞 𝐚𝐮𝐝𝐢𝐭𝐞𝐝.\n"

// Comment is our internal representation of a vcs based comment event.
type Comment struct {
	Pull              models.PullRequest
	BaseRepo          models.Repo
	HeadRepo          models.Repo
	User              models.User
	PullNum           int
	Comment           string
	VCSHost           models.VCSHostType
	Timestamp         time.Time
	InstallationToken int64
}

func NewCommentEventWorkerProxy(
	logger logging.Logger,
	snsWriter Writer,
	allocator feature.Allocator,
	rootConfigBuilder rootConfigBuilder,
	scheduler scheduler,
	deploySignaler deploySignaler,
	vcsClient vcs.Client) *CommentEventWorkerProxy {
	return &CommentEventWorkerProxy{
		logger:            logger,
		snsWriter:         snsWriter,
		allocator:         allocator,
		rootConfigBuilder: rootConfigBuilder,
		scheduler:         scheduler,
		deploySignaler:    deploySignaler,
		vcsClient:         vcsClient,
	}
}

type CommentEventWorkerProxy struct {
	logger            logging.Logger
	snsWriter         Writer
	allocator         feature.Allocator
	rootConfigBuilder rootConfigBuilder
	scheduler         scheduler
	deploySignaler    deploySignaler
	vcsClient         vcs.Client
}

func (p *CommentEventWorkerProxy) Handle(ctx context.Context, request *http.BufferedRequest, event Comment, command *command.Comment) error {
	shouldAllocate, err := p.allocator.ShouldAllocate(feature.PlatformMode, feature.FeatureContext{
		RepoName: event.BaseRepo.FullName,
	})

	if err != nil {
		p.logger.ErrorContext(ctx, "unable to allocate platform mode")
		return p.forwardToSns(ctx, request)
	}

	if shouldAllocate && command.ForceApply {
		p.logger.InfoContext(ctx, "running force apply command")
		if commentErr := p.vcsClient.CreateComment(event.BaseRepo, event.PullNum, warningMessage, ""); commentErr != nil {
			p.logger.ErrorContext(ctx, commentErr.Error())
		}
		return p.scheduler.Schedule(ctx, func(ctx context.Context) error {
			return p.forceApply(ctx, event)
		})
	}
	return p.forwardToSns(ctx, request)
}

func (p *CommentEventWorkerProxy) forwardToSns(ctx context.Context, request *http.BufferedRequest) error {
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

func (p *CommentEventWorkerProxy) forceApply(ctx context.Context, event Comment) error {
	fileFetcherOptions := github.FileFetcherOptions{
		PRNum: event.PullNum,
	}
	rootCfgs, err := p.rootConfigBuilder.Build(ctx, event.HeadRepo, event.Pull.HeadBranch, event.Pull.HeadCommit, fileFetcherOptions, event.InstallationToken)
	if err != nil {
		return errors.Wrap(err, "generating roots")
	}
	for _, rootCfg := range rootCfgs {
		ctx = context.WithValue(ctx, contextInternal.ProjectKey, rootCfg.Name)
		run, err := p.deploySignaler.SignalWithStartWorkflow(
			ctx,
			rootCfg,
			event.BaseRepo,
			event.Pull.HeadCommit,
			event.InstallationToken,
			event.Pull.HeadRef,
			event.User,
			workflows.ManualTrigger)
		if err != nil {
			return errors.Wrap(err, "signalling workflow")
		}

		p.logger.InfoContext(ctx, "Signaled workflow.", map[string]interface{}{
			"workflow-id": run.GetID(), "run-id": run.GetRunID(),
		})
	}
	return nil
}
