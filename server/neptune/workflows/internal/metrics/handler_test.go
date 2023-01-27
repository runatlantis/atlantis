package metrics_test

import (
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/metrics"
	"go.temporal.io/sdk/client"
	"gopkg.in/go-playground/assert.v1"
)

type testHandler struct {
	t                 *testing.T
	called            bool
	expectedTags      map[string]string
	expectedNamespace string
}

func (h testHandler) WithTags(tags map[string]string) client.MetricsHandler {
	assert.Equal(h.t, h.expectedTags, tags)
	return testHandler{
		t:                 h.t,
		expectedNamespace: h.expectedNamespace,
		called:            true,
	}
}
func (h testHandler) Counter(namespace string) client.MetricsCounter {
	assert.Equal(h.t, h.expectedNamespace, namespace)
	return testHandler{
		called: true,
	}
}
func (h testHandler) Gauge(namespace string) client.MetricsGauge {
	assert.Equal(h.t, h.expectedNamespace, namespace)
	return testHandler{
		called: true,
	}
}
func (h testHandler) Timer(namespace string) client.MetricsTimer {
	assert.Equal(h.t, h.expectedNamespace, namespace)
	return testHandler{
		called: true,
	}
}

func (h testHandler) Inc(int64)            {}
func (h testHandler) Update(float64)       {}
func (h testHandler) Record(time.Duration) {}

func TestScope(t *testing.T) {
	t.Run("subscope", func(t *testing.T) {
		handler := testHandler{
			expectedNamespace: "some.namespace.nish.hi",
		}
		_ = metrics.NewScope(handler, "some", "namespace").
			SubScope("nish").
			Counter("hi")
	})
	t.Run("gauge", func(t *testing.T) {
		handler := testHandler{
			expectedNamespace: "nish.hi",
		}
		_ = metrics.NewScope(handler, "nish").
			Gauge("hi")
	})

	t.Run("tags", func(t *testing.T) {
		handler := testHandler{
			expectedNamespace: "nish.hi",
			expectedTags:      map[string]string{"hello": "world"},
		}
		_ = metrics.NewScope(handler, "nish").
			SubScopeWithTags(map[string]string{"hello": "world"}).
			Gauge("hi")
	})
}
