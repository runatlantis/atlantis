package metrics

import (
	"strings"

	"go.temporal.io/sdk/client"
)

const (
	SignalNameTag       = "signal_name"
	SignalReceiveMetric = "signal_receive"
)

const delim = "."

// Scope is an interface that attempts to wrap temporal's MetricsHandler with additional
// functionality to build namespaces.
// Namespaces are of the following form:
// <s1>.<s2>.<s3> and can have a number of tags associated with it.
type Scope interface {
	SubScope(namespace ...string) Scope
	SubScopeWithTags(tags map[string]string) Scope
	Counter(name string) client.MetricsCounter
	Gauge(name string) client.MetricsGauge
}

type WorkflowScope struct {
	namespace string
	handler   client.MetricsHandler
}

func NewScope(handler client.MetricsHandler, namespaces ...string) *WorkflowScope {
	return &WorkflowScope{
		namespace: join(namespaces...),
		handler:   handler,
	}
}

// NewNullableScope should only be used for testing purposes since it just drops metrics
func NewNullableScope() *WorkflowScope {
	return &WorkflowScope{
		handler: client.MetricsNopHandler,
	}
}

func (s *WorkflowScope) SubScope(namespaces ...string) Scope {
	return &WorkflowScope{
		namespace: join(s.namespace, join(namespaces...)),
		handler:   s.handler,
	}
}

func (s *WorkflowScope) SubScopeWithTags(tags map[string]string) Scope {
	return &WorkflowScope{
		namespace: s.namespace,
		handler:   s.handler.WithTags(tags),
	}
}

func (s *WorkflowScope) Counter(name string) client.MetricsCounter {
	return s.handler.Counter(join(s.namespace, name))
}

func (s *WorkflowScope) Gauge(name string) client.MetricsGauge {
	return s.handler.Gauge(join(s.namespace, name))
}

func join(s ...string) string {
	return strings.Join(s, delim)
}
