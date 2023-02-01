package metrics

import (
	"github.com/runatlantis/atlantis/server/neptune/temporal"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
)

// Scope is an interface that attempts to wrap temporal's MetricsHandler with additional
// functionality to build namespaces using NamespacedMetricsHandler
type Scope interface {
	SubScope(namespace ...string) Scope
	SubScopeWithTags(tags map[string]string) Scope
	Counter(name string) client.MetricsCounter
	Gauge(name string) client.MetricsGauge
}

type WorkflowScope struct {
	handler client.MetricsHandler
}

func NewScopeWithHandler(handler client.MetricsHandler, namespaces ...string) *WorkflowScope {
	return &WorkflowScope{
		handler: toNamespacedHandler(handler, namespaces...),
	}
}

func NewScope(ctx workflow.Context, namespaces ...string) *WorkflowScope {
	return &WorkflowScope{
		handler: toNamespacedHandler(workflow.GetMetricsHandler(ctx), namespaces...),
	}
}

func toNamespacedHandler(handler client.MetricsHandler, namespaces ...string) client.MetricsHandler {
	if len(namespaces) == 0 {
		return handler
	}
	switch h := handler.(type) {
	case *temporal.NamespacedMetricsHandler:
		return h.WithNamespace(namespaces...)
	default:
		return temporal.NewNamespacedMetricsHandler(h, namespaces...)
	}
}

// NewNullableScope should only be used for testing purposes since it just drops metrics
func NewNullableScope() *WorkflowScope {
	return NewScopeWithHandler(client.MetricsNopHandler)
}

func (s *WorkflowScope) SubScope(namespaces ...string) Scope {
	return NewScopeWithHandler(s.handler, namespaces...)
}

func (s *WorkflowScope) SubScopeWithTags(tags map[string]string) Scope {
	return &WorkflowScope{
		handler: s.handler.WithTags(tags),
	}
}

func (s *WorkflowScope) Counter(name string) client.MetricsCounter {
	return s.handler.Counter(name)
}

func (s *WorkflowScope) Gauge(name string) client.MetricsGauge {
	return s.handler.Gauge(name)
}
