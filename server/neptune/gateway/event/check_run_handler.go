package event

import (
	"context"
	"fmt"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/vcs/provider/github"
	"regexp"
	"strings"

	contextInternal "github.com/runatlantis/atlantis/server/neptune/context"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/neptune/workflows"
)

const (
	RequestedActionType   = "requested_action"
	ReRequestedActionType = "rerequested"
)

var checkRunRegex = regexp.MustCompile("atlantis/deploy: (?P<name>.+)")

type CheckRunAction interface {
	GetType() string
}

type WrappedCheckRunAction string

func (a WrappedCheckRunAction) GetType() string {
	return string(a)
}

type RequestedActionChecksAction struct {
	Identifier string
}

func (a RequestedActionChecksAction) GetType() string {
	return RequestedActionType
}

type CheckRun struct {
	Action            CheckRunAction
	ExternalID        string
	Name              string
	Repo              models.Repo
	User              models.User
	InstallationToken int64
	Branch            string
	HeadSha           string
}

type CheckRunHandler struct {
	Logger            logging.Logger
	RootConfigBuilder rootConfigBuilder
	Scheduler         scheduler
	DeploySignaler    deploySignaler
}

func (h *CheckRunHandler) Handle(ctx context.Context, event CheckRun) error {
	// first let's make sure this is an atlantis check run
	if !strings.HasPrefix(event.Name, "atlantis") {
		h.Logger.DebugContext(ctx, "Ignoring non-atlantis checks event")
		return nil
	}
	matches := checkRunRegex.FindStringSubmatch(event.Name)
	if len(matches) != 2 {
		return fmt.Errorf("unable to determine root name")
	}
	rootName := matches[checkRunRegex.SubexpIndex("name")]
	ctx = context.WithValue(ctx, contextInternal.ProjectKey, rootName)

	// we only handle requested/re-requested action types
	switch event.Action.GetType() {
	case RequestedActionType:
		return h.handleRequestedAction(ctx, event, rootName)
	case ReRequestedActionType:
		return h.Scheduler.Schedule(ctx, func(ctx context.Context) error {
			return h.handleReRequestedRun(ctx, event, rootName)
		})
	}
	h.Logger.DebugContext(ctx, "ignoring checks event that isn't a requested/re-requested action")
	return nil
}

func (h *CheckRunHandler) handleReRequestedRun(ctx context.Context, event CheckRun, rootName string) error {
	// Block force applies
	if event.Branch != event.Repo.DefaultBranch {
		h.Logger.DebugContext(ctx, "dropping event branch unexpected ref")
		return nil
	}
	return h.buildRoot(ctx, event, rootName)
}

func (h *CheckRunHandler) handleRequestedAction(ctx context.Context, event CheckRun, rootName string) error {
	action, ok := event.Action.(RequestedActionChecksAction)
	if !ok {
		return fmt.Errorf("event action type does not match string type.  This is likely a code bug")
	}
	switch action.Identifier {
	case "Unlock":
		return h.signalUnlockWorkflowChannel(ctx, event, rootName)
	case "Approve":
		return h.signalPlanReviewWorkflowChannel(ctx, event, workflows.ApprovedPlanReviewStatus)
	case "Reject":
		return h.signalPlanReviewWorkflowChannel(ctx, event, workflows.RejectedPlanReviewStatus)
	}
	return fmt.Errorf("unknown action id %s", action.Identifier)
}

func (h *CheckRunHandler) signalPlanReviewWorkflowChannel(ctx context.Context, event CheckRun, status workflows.TerraformPlanReviewStatus) error {
	err := h.DeploySignaler.SignalWorkflow(
		ctx,
		// assumed that we're using the check run external id as our workflow id
		event.ExternalID,
		// keeping this empty is fine since temporal will find the currently running workflow
		"",
		workflows.TerraformPlanReviewSignalName,
		workflows.TerraformPlanReviewSignalRequest{
			Status: status,
			User:   event.User.Username,
		})
	if err != nil {
		return errors.Wrapf(err, "signaling workflow with id: %s", event.ExternalID)
	}
	h.Logger.InfoContext(ctx, fmt.Sprintf("Signaled workflow with id %s, review status, %d", event.ExternalID, status))
	return nil
}

func (h *CheckRunHandler) signalUnlockWorkflowChannel(ctx context.Context, event CheckRun, rootName string) error {
	err := h.DeploySignaler.SignalWorkflow(
		ctx,
		// deploy workflow id is repo||root (the name of the check run is the root)
		buildDeployWorkflowID(event.Repo.FullName, rootName),
		// keeping this empty is fine since temporal will find the currently running workflow
		"",
		workflows.DeployUnlockSignalName,
		workflows.DeployUnlockSignalRequest{
			User: event.User.Username,
		})
	if err != nil {
		return errors.Wrapf(err, "signaling workflow with id: %s", event.ExternalID)
	}
	h.Logger.InfoContext(ctx, fmt.Sprintf("Signaled workflow with id %s to unlock", event.ExternalID))
	return nil
}

func (h *CheckRunHandler) buildRoot(ctx context.Context, event CheckRun, rootName string) error {
	builderOptions := BuilderOptions{
		RepoFetcherOptions: github.RepoFetcherOptions{
			ShallowClone: true,
		},
		FileFetcherOptions: github.FileFetcherOptions{
			Sha: event.HeadSha,
		},
	}
	rootCfgs, err := h.RootConfigBuilder.Build(ctx, event.Repo, event.Branch, event.HeadSha, event.InstallationToken, builderOptions)
	if err != nil {
		return errors.Wrap(err, "generating roots")
	}
	for _, rootCfg := range rootCfgs {
		if rootCfg.Name != rootName {
			continue
		}
		c := context.WithValue(ctx, contextInternal.ProjectKey, rootCfg.Name)
		if rootCfg.WorkflowMode != valid.PlatformWorkflowMode {
			h.Logger.DebugContext(c, "root is not configured for platform mode, skipping...")
			continue
		}
		deployOptions := RootDeployOptions{
			Repo:              event.Repo,
			Branch:            event.Branch,
			Revision:          event.HeadSha,
			Sender:            event.User,
			InstallationToken: event.InstallationToken,
			BuilderOptions:    builderOptions,
			Trigger:           workflows.ManualTrigger,
			Rerun:             true,
		}
		run, err := h.DeploySignaler.SignalWithStartWorkflow(c, rootCfg, deployOptions)
		if err != nil {
			return errors.Wrap(err, "signalling workflow")
		}
		h.Logger.InfoContext(c, "Signaled workflow.", map[string]interface{}{
			"workflow-id": run.GetID(), "run-id": run.GetRunID(),
		})
	}
	return nil
}
