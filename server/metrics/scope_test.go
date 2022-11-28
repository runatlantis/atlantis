package metrics_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/metrics"
)

var (
	prometheusConfig = valid.Metrics{
		Prometheus: &valid.Prometheus{
			Endpoint: "/metrics",
		},
	}
)

func TestNewScope_PrometheusTaggingCapabilities(t *testing.T) {
	scope, _, _, err := metrics.NewScope(prometheusConfig, nil, "test")
	if err != nil {
		t.Fatalf("got an error: %s", err.Error())
	}

	scope.Tagged(map[string]string{
		"base_repo": "runatlantis/atlantis",
		"pr_number": "2687",
	})

	want := true
	got := scope.Capabilities().Tagging()
	if want != got {
		t.Errorf("Scope does not have Capability to do Tagging")
	}
}
