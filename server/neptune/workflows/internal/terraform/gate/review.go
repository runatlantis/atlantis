package gate

import (
	"time"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/neptune/context"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/config/logger"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/temporal"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
)

const (
	PlanReviewSignalName = "planreview"
	PlanReviewTimerStat  = "workflow.terraform.planreview"
)

type PlanStatus int
type PlanReviewSignalRequest struct {
	Status PlanStatus

	// TODO: Output this info to the checks UI
	User string
}

const (
	Approved PlanStatus = iota
	Rejected
)

// Review waits for a plan review signal or a timeout to occur and returns an associated status.
type Review struct {
	MetricsHandler client.MetricsHandler
	Timeout        time.Duration
	Client         ActionsClient
}

type ActionsClient interface {
	UpdateApprovalActions(approval terraform.PlanApproval) error
}

func (r *Review) Await(ctx workflow.Context, root terraform.Root, planSummary terraform.PlanSummary) (PlanStatus, error) {
	if root.Plan.Approval.Type == terraform.AutoApproval || planSummary.IsEmpty() {
		return Approved, nil
	}

	waitStartTime := time.Now()
	defer func() {
		r.MetricsHandler.Timer(PlanReviewTimerStat).Record(time.Since(waitStartTime))
	}()

	ch := workflow.GetSignalChannel(ctx, PlanReviewSignalName)
	selector := temporal.SelectorWithTimeout{
		Selector: workflow.NewSelector(ctx),
	}

	var planReview PlanReviewSignalRequest
	selector.AddReceive(ch, func(c workflow.ReceiveChannel, more bool) {
		ch.Receive(ctx, &planReview)
	})

	var timedOut bool
	selector.AddTimeout(ctx, r.Timeout, func(f workflow.Future) {
		if err := f.Get(ctx, nil); err != nil {
			logger.Warn(ctx, "Error timing out selector.  This is possibly due to a cancellation signal. ", context.ErrKey, err)
		}
		timedOut = true
	})

	err := r.Client.UpdateApprovalActions(root.Plan.Approval)
	if err != nil {
		return Rejected, errors.Wrap(err, "updating approval actions")
	}

	selector.Select(ctx)

	if timedOut {
		return Rejected, nil
	}

	return planReview.Status, nil
}
