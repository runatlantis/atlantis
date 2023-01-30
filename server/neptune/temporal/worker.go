package temporal

import (
	"context"

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

type activityOutboundInterceptor struct {
	interceptor.ActivityOutboundInterceptorBase
}

func (a *activityOutboundInterceptor) GetMetricsHandler(ctx context.Context) client.MetricsHandler {
	handler := a.Next.GetMetricsHandler(ctx)
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

type workflowOutboundInterceptor struct {
	interceptor.WorkflowOutboundInterceptorBase
}

func (w *workflowOutboundInterceptor) GetMetricsHandler(ctx workflow.Context) client.MetricsHandler {
	handler := w.Next.GetMetricsHandler(ctx)
	return handler.WithTags(getOptionalTags(ctx))
}
