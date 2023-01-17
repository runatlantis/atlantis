package gate

import (
	"time"

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
}

func (r *Review) Await(ctx workflow.Context, root terraform.Root, planSummary terraform.PlanSummary) PlanStatus {
	waitStartTime := time.Now()
	defer func() {
		r.MetricsHandler.Timer(PlanReviewTimerStat).Record(time.Since(waitStartTime))
	}()

	if root.Plan.Approval.Type == terraform.AutoApproval || planSummary.IsEmpty() {
		return Approved
	}

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

	selector.Select(ctx)

	if timedOut {
		return Rejected
	}

	return planReview.Status
}
