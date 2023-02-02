package temporal

import (
	"context"
	"strings"
	"time"

	"github.com/runatlantis/atlantis/server/events/metrics"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/workflow"
)

func NewWorkerInterceptor() *WorkerInterceptor {
	return &WorkerInterceptor{}
}

type WorkerInterceptor struct {
	interceptor.WorkerInterceptorBase
}

func (w *WorkerInterceptor) InterceptActivity(
	ctx context.Context,
	next interceptor.ActivityInboundInterceptor,
) interceptor.ActivityInboundInterceptor {
	i := &activityInboundInterceptor{}
	i.Next = next
	return i
}

func (w *WorkerInterceptor) InterceptWorkflow(
	ctx workflow.Context,
	next interceptor.WorkflowInboundInterceptor,
) interceptor.WorkflowInboundInterceptor {
	i := &workflowInboundInterceptor{}
	i.Next = next
	return i
}

type activityInboundInterceptor struct {
	interceptor.ActivityInboundInterceptorBase
}

func (a *activityInboundInterceptor) Init(outbound interceptor.ActivityOutboundInterceptor) error {
	i := &activityOutboundInterceptor{}
	i.Next = outbound
	return a.Next.Init(i)
}

func (a *activityInboundInterceptor) ExecuteActivity(
	ctx context.Context,
	in *interceptor.ExecuteActivityInput,
) (interface{}, error) {
	handler := activity.GetMetricsHandler(ctx)

	startTime := time.Now()
	timer := handler.Timer(metrics.ActivityExecutionLatency)
	defer func() {
		timer.Record(time.Since(startTime))
	}()

	result, err := a.Next.ExecuteActivity(ctx, in)
	if err != nil {
		handler.Counter(metrics.ActivityExecutionFailure).Inc(1)
		return result, err
	}

	handler.Counter(metrics.ActivityExecutionSuccess).Inc(1)
	return result, err
}

type activityOutboundInterceptor struct {
	interceptor.ActivityOutboundInterceptorBase
}

func (a *activityOutboundInterceptor) GetMetricsHandler(ctx context.Context) client.MetricsHandler {
	info := activity.GetInfo(ctx)
	handler := NewNamespacedMetricsHandler(
		a.Next.GetMetricsHandler(ctx),
		"workflow", strings.ToLower(info.WorkflowType.Name),
	)

	return handler.WithTags(getOptionalTags(ctx))
}

type workflowInboundInterceptor struct {
	interceptor.WorkflowInboundInterceptorBase
}

func (w *workflowInboundInterceptor) Init(outbound interceptor.WorkflowOutboundInterceptor) error {
	i := &workflowOutboundInterceptor{}
	i.Next = outbound
	return w.Next.Init(i)
}

func (w *workflowInboundInterceptor) ExecuteWorkflow(ctx workflow.Context, in *interceptor.ExecuteWorkflowInput) (interface{}, error) {
	handler := workflow.GetMetricsHandler(ctx)

	startTime := time.Now()
	timer := handler.Timer(metrics.WorkflowLatency)
	defer func() {
		timer.Record(time.Since(startTime))
	}()

	result, err := w.Next.ExecuteWorkflow(ctx, in)
	if err != nil {
		handler.Counter(metrics.WorkflowFailure).Inc(1)
		return result, err
	}

	handler.Counter(metrics.WorkflowSuccess).Inc(1)
	return result, err
}

func (w *workflowInboundInterceptor) HandleSignal(ctx workflow.Context, in *interceptor.HandleSignalInput) error {
	handler := workflow.GetMetricsHandler(ctx).WithTags(map[string]string{
		metrics.SignalNameTag: in.SignalName,
	})

	startTime := time.Now()
	timer := handler.Timer(metrics.SignalHandleLatency)
	defer func() {
		timer.Record(time.Since(startTime))
	}()

	err := w.Next.HandleSignal(ctx, in)
	if err != nil {
		handler.Counter(metrics.SignalHandleFailure).Inc(1)
		return err
	}

	handler.Counter(metrics.SignalHandleSuccess).Inc(1)
	return nil
}

type workflowOutboundInterceptor struct {
	interceptor.WorkflowOutboundInterceptorBase
}

func (w *workflowOutboundInterceptor) GetMetricsHandler(ctx workflow.Context) client.MetricsHandler {
	info := workflow.GetInfo(ctx)
	handler := NewNamespacedMetricsHandler(
		w.Next.GetMetricsHandler(ctx),
		"workflow", strings.ToLower(info.WorkflowType.Name),
	)
	return handler.WithTags(getOptionalTags(ctx))
}
