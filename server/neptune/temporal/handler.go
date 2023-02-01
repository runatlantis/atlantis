package temporal

import (
	"strings"

	"go.temporal.io/sdk/client"
)

const delim = "."

// NamespacedMetricsHandler caches a given namespace and allows building new objects off that namespace
// Namespaces are of the following form:
// <s1>.<s2>.<s3> and can have a number of tags associated with it.
type NamespacedMetricsHandler struct {
	namespace string
	handler   client.MetricsHandler
}

func NewNamespacedMetricsHandler(handler client.MetricsHandler, namespaces ...string) *NamespacedMetricsHandler {
	return &NamespacedMetricsHandler{
		namespace: join(namespaces...),
		handler:   handler,
	}
}

func (s *NamespacedMetricsHandler) WithTags(tags map[string]string) client.MetricsHandler {
	return &NamespacedMetricsHandler{
		namespace: s.namespace,
		handler:   s.handler.WithTags(tags),
	}
}

func (s *NamespacedMetricsHandler) WithNamespace(namespaces ...string) *NamespacedMetricsHandler {
	return &NamespacedMetricsHandler{
		namespace: join(s.namespace, join(namespaces...)),
		handler:   s.handler,
	}
}

func (s *NamespacedMetricsHandler) Counter(name string) client.MetricsCounter {
	return s.handler.Counter(join(s.namespace, name))
}

func (s *NamespacedMetricsHandler) Gauge(name string) client.MetricsGauge {
	return s.handler.Gauge(join(s.namespace, name))
}

func (s *NamespacedMetricsHandler) Timer(name string) client.MetricsTimer {
	return s.handler.Timer(join(s.namespace, name))
}

func join(s ...string) string {
	return strings.Join(s, delim)
}
