package event

import (
	"context"
	"fmt"
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
	Logger         logging.Logger
	RootDeployer   rootDeployer
	SyncScheduler  scheduler
	AsyncScheduler scheduler
	DeploySignaler deploySignaler
}

func (h *CheckRunHandler) Handle(ctx context.Context, event CheckRun) error {
	// first let's make sure this is an atlantis check run
	if !strings.HasPrefix(event.Name, "atlantis/deploy") {
		h.Logger.DebugContext(ctx, "Ignoring non-atlantis checks event")
		return nil
	}
	matches := checkRunRegex.FindStringSubmatch(event.Name)
	if len(matches) != 2 {
		h.Logger.ErrorContext(ctx, fmt.Sprintf("unable to determine root name: %s", event.Name))
		return fmt.Errorf("unable to determine root name")
	}
	rootName := matches[checkRunRegex.SubexpIndex("name")]
	ctx = context.WithValue(ctx, contextInternal.ProjectKey, rootName)

	// we only handle requested/re-requested action types
	switch event.Action.GetType() {
	case RequestedActionType:
		// wrap this in a scheduler for consistent err handling (ie. logging)
		return h.SyncScheduler.Schedule(ctx, func(ctx context.Context) error {
			return h.handleRequestedAction(ctx, event, rootName)
		})
	case ReRequestedActionType:
		return h.AsyncScheduler.Schedule(ctx, func(ctx context.Context) error {
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
	case "Confirm":
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
	workflowID := buildDeployWorkflowID(event.Repo.FullName, rootName)
	err := h.DeploySignaler.SignalWorkflow(
		ctx,
		// deploy workflow id is repo||root (the name of the check run is the root)
		workflowID,
		// keeping this empty is fine since temporal will find the currently running workflow
		"",
		workflows.DeployUnlockSignalName,
		workflows.DeployUnlockSignalRequest{
			User: event.User.Username,
		})
	if err != nil {
		return errors.Wrapf(err, "signaling workflow with id: %s", workflowID)
	}
	h.Logger.InfoContext(ctx, fmt.Sprintf("Signaled workflow with id %s to unlock", workflowID))
	return nil
}

func (h *CheckRunHandler) buildRoot(ctx context.Context, event CheckRun, rootName string) error {
	deployOptions := RootDeployOptions{
		Repo:              event.Repo,
		Branch:            event.Branch,
		RootNames:         []string{rootName},
		Revision:          event.HeadSha,
		Sender:            event.User,
		InstallationToken: event.InstallationToken,
		Trigger:           workflows.ManualTrigger,
		Rerun:             true,
	}
	return errors.Wrap(h.RootDeployer.Deploy(ctx, deployOptions), "deploying workflow")
}
